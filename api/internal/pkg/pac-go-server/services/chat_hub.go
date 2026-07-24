package services

import (
	"fmt"
	"sync"
	"sync/atomic"

	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
	"go.uber.org/zap"
)

// chatHub is a simple fan-out hub.  Subscribers register a channel under an
// arbitrary string key and receive every message published to that key.
//
// Two key namespaces are used:
//   - user side:  userID            — admin → user pushes (hubMessage.Sender == "admin")
//   - admin side: "userID:convID"   — user → admin pushes (hubMessage.Sender == "user")
type chatHub struct {
	mu        sync.Mutex
	subs      map[string][]chan hubMessage
	dropCount atomic.Int64 // total messages dropped due to full subscriber buffers
}

type hubMessage struct {
	ConversationID int64  `json:"conversation_id"`
	UserID         string `json:"user_id,omitempty"`
	Sender         string `json:"sender"`
	Message        string `json:"message"`
	Timestamp      string `json:"timestamp"`
}

// adminBroadcastKey is the hub key that every incoming user message is also
// published to, so a single admin WS can receive notifications for all convs.
const adminBroadcastKey = "admin:broadcast"

var hub = &chatHub{
	subs: make(map[string][]chan hubMessage),
}

// userToAdminKey returns the hub key under which an admin watches a specific
// user conversation (user → admin direction: "userID:convID").
func userToAdminKey(userID string, convID int64) string {
	return fmt.Sprintf("%s:%d", userID, convID)
}

// subscribe registers ch under key and returns a cleanup function that removes it.
// The returned function must be called (e.g. via defer) to avoid leaking the
// channel in the hub after the subscriber goroutine exits.
func (h *chatHub) subscribe(key string, ch chan hubMessage) func() {
	h.mu.Lock()
	h.subs[key] = append(h.subs[key], ch)
	h.mu.Unlock()

	// unsubscribe removes ch from the list under key.
	// If the list becomes empty the key is deleted from the map entirely.
	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		list := h.subs[key]
		for i, c := range list {
			if c == ch {
				h.subs[key] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(h.subs[key]) == 0 {
			delete(h.subs, key)
		}
	}
}

// publish sends msg to every channel currently subscribed under key.
// Non-blocking: if a channel's buffer is full the message is dropped for that
// subscriber (WebSocket goroutines are assumed to drain quickly).
//
// Every drop increments dropCount and is logged at Warn level so that
// "hot" conversations that are overwhelming the hub are immediately visible
// in the server logs without any external metrics infrastructure.
func (h *chatHub) publish(key string, msg hubMessage) {
	h.mu.Lock()
	list := make([]chan hubMessage, len(h.subs[key]))
	copy(list, h.subs[key])
	h.mu.Unlock()

	for _, ch := range list {
		select {
		case ch <- msg:
		default:
			// Subscriber channel is full — message dropped.
			// Increment the counter and emit a warning so operators can
			// identify which conversation is generating backpressure.
			total := h.dropCount.Add(1)
			log.GetLogger().Warn("hub: subscriber channel full, message dropped",
				zap.String("key", key),
				zap.String("sender", msg.Sender),
				zap.Int64("total_drops", total),
			)
		}
	}
}
