package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pacClient "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/client"
	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
)

// wsWriteTimeout is the maximum time allowed for a single WebSocket write.
// Without a deadline, a zombie client (network dropped but TCP not yet closed)
// will cause wsjson.Write to block indefinitely, leaking the goroutine until
// the OS TCP keepalive eventually fires (minutes to hours).
const wsWriteTimeout = 10 * time.Second

// wsWrite wraps wsjson.Write with a per-call timeout derived from the
// session context.  Using the session ctx (not a fresh Background ctx) means
// that if the session is already cancelled the write returns immediately.
func wsWrite(ctx context.Context, conn *websocket.Conn, v interface{}) error {
	wctx, cancel := context.WithTimeout(ctx, wsWriteTimeout)
	defer cancel()
	return wsjson.Write(wctx, conn, v)
}

// HandleChatWebSocket upgrades the connection to WebSocket and handles real-time
// chat messages for the authenticated user.
func HandleChatWebSocket(c *gin.Context) {
	logger := log.GetLogger()

	// Auth is already done by the middleware chain (InjectTokenFromQuery →
	// ginkeycloak.Auth → RetrospectKeycloakToken). Read userid before the
	// upgrade since c.Request.Context() is only valid until the handler returns.
	userID, _ := c.Request.Context().Value("userid").(string)
	username, _ := c.Request.Context().Value("username").(string)

	// coder/websocket calls WriteHeaderNow() on Gin's writer before Hijack(),
	// which sets Written()=true and causes Gin's Hijack() guard to reject it.
	// Unwrap one level to get the raw net/http ResponseWriter, bypassing that
	// check. coder/websocket's own hijacker() loop still finds the socket.
	type unwrapper interface{ Unwrap() http.ResponseWriter }
	rw := http.ResponseWriter(c.Writer)
	if u, ok := c.Writer.(unwrapper); ok {
		rw = u.Unwrap()
	}
	conn, err := websocket.Accept(rw, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	// Mark Gin's writer as written so Recovery/WriteHeaderNow do not attempt
	// to write on the hijacked connection when the middleware chain unwinds.
	c.Writer.WriteHeaderNow()

	// Run the WebSocket loop in a goroutine and return immediately from the
	// handler. This is the key fix for the EOF-on-first-read bug:
	//
	// RetrospectKeycloakToken calls c.Next() which runs this handler inline.
	// If the handler blocks in the read loop, everything is fine — until the
	// loop exits and the middleware chain unwinds. net/http then calls
	// finishRequest() which closes the TCP socket, killing the connection.
	//
	// By returning immediately, the middleware chain unwinds while the
	// goroutine owns the already-hijacked conn. net/http sees Written()=true
	// and does not attempt to close or write to the connection.
	go func() {
		defer conn.Close(websocket.StatusNormalClosure, "")
		serveChat(conn, userID, username, logger)
	}()
}

// serveChat runs the full WebSocket session for one connected user.
func serveChat(conn *websocket.Conn, userID, username string, logger *zap.Logger) {
	// Use a cancellable context tied to the connection lifetime.
	// When the client disconnects (or the server closes the conn), cancel
	// unblocks any in-flight wsjson.Read/Write immediately — no goroutine leak.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conversationID, err := dbCon.GetCurrentConversationID(ctx, userID)
	if err != nil {
		logger.Error("failed to get conversation ID", zap.Error(err))
		_ = wsWrite(ctx, conn, gin.H{"error": "failed to initialize conversation"})
		return
	}

	// If the current conversation was previously ended, pre-set convEnded so
	// the first new message correctly triggers GetNextConversationID.
	convEndedInit, err := dbCon.IsConversationEnded(ctx, userID, conversationID)
	if err != nil {
		logger.Warn("failed to check conversation ended state", zap.Error(err))
		// non-fatal: treat as not ended
	}

	history, err := dbCon.GetChatMessages(ctx, userID, conversationID)
	if err != nil {
		logger.Error("failed to load chat history", zap.Error(err))
		history = nil
	}

	type msgFrame struct {
		ID             string `json:"id"`
		ConversationID int64  `json:"conversation_id"`
		Sender         string `json:"sender"`
		Message        string `json:"message"`
		Timestamp      string `json:"timestamp"`
	}

	historyFrames := make([]msgFrame, 0, len(history))
	for _, m := range history {
		historyFrames = append(historyFrames, msgFrame{
			ID:             m.ID.Hex(),
			ConversationID: m.ConversationID,
			Sender:         m.Sender,
			Message:        m.Message,
			Timestamp:      m.Timestamp.Format(time.RFC3339),
		})
	}

	hello := map[string]interface{}{
		"type":            "hello",
		"conversation_id": conversationID,
		"history":         historyFrames,
	}
	if err := wsWrite(ctx, conn, hello); err != nil {
		logger.Error("failed to send hello frame", zap.Error(err))
		return
	}

	logger.Info("conversation resumed",
		zap.String("userID", userID),
		zap.Int64("conversationID", conversationID),
		zap.Int("history_len", len(history)))

	// incomingFromAdminCh receives messages pushed by AdminReply to this user.
	// Buffered so a slow WebSocket write does not block the AdminReply HTTP handler.
	// unsubFromAdmin is reassigned when the user starts a new conversation and a new
	// hub subscription is created; the deferred wrapper always calls the current one.
	incomingFromAdminCh := make(chan hubMessage, 16)
	unsubFromAdmin := hub.subscribe(userID, incomingFromAdminCh)
	defer func() { unsubFromAdmin() }()

	// convEnded tracks whether the user has ended the current conversation
	// but not yet sent a message on the next one.  In this state the server
	// keeps the session open and waits: the next real message triggers a lazy
	// GetNextConversationID so that a new conversation ID only enters the DB
	// when there is actual content to attach to it.
	convEnded := convEndedInit

	// A single persistent reader goroutine feeds all incoming client frames
	// into readCh.  Starting it once (rather than once per loop iteration)
	// means there is never more than one outstanding wsjson.Read at a time,
	// so no goroutines are leaked when the adminCh or ctx.Done case fires.
	type readResult struct {
		msg    string
		action string
		err    error
	}
	readCh := make(chan readResult, 1)
	go func() {
		for {
			var incoming struct {
				Message string `json:"message"`
				Action  string `json:"action"`
			}
			err := wsjson.Read(ctx, conn, &incoming)
			readCh <- readResult{msg: incoming.Message, action: incoming.Action, err: err}
			if err != nil {
				return // connection closed; exit the goroutine
			}
		}
	}()

	for {
		select {
		case res := <-readCh:
			if res.err != nil {
				status := websocket.CloseStatus(res.err)
				if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway {
					logger.Info("websocket closed", zap.String("userID", userID))
				} else {
					logger.Error("websocket read error", zap.Error(res.err))
				}
				return
			}

			// "end_conversation" action: mark the conversation ended locally.
			// We do NOT call GetNextConversationID here — no DB write, no new
			// conversation ID until the user actually sends a message.  This
			// means reconnecting before sending anything correctly resumes the
			// same (still highest) conversation ID from the DB.
			if res.action == "end_conversation" {
				if convEnded {
					continue // already ended, ignore duplicate
				}
				convEnded = true

				// Persist the ended sentinel to the DB so the status survives
				// reconnects and page navigations.
				if err := dbCon.MarkConversationEnded(ctx, userID, username, conversationID); err != nil {
					logger.Error("failed to mark conversation ended", zap.Error(err))
				}

				// Notify any admin watching the current conversation.
				hub.publish(userToAdminKey(userID, conversationID), hubMessage{
					ConversationID: conversationID,
					Sender:         models.SenderSystem,
					Message:        "conversation_ended",
				})

				// Tell the client the conversation is ended.
				frame := map[string]interface{}{
					"type":            "ended",
					"conversation_id": conversationID,
				}
				if err := wsWrite(ctx, conn, frame); err != nil {
					logger.Error("failed to send ended frame", zap.Error(err))
					return
				}
				logger.Info("conversation ended by user",
					zap.String("userID", userID),
					zap.Int64("conversationID", conversationID))
				continue
			}

			if res.msg == "" {
				continue
			}

			// If the user ended the previous conversation, lazily assign a new
			// conversation ID now — only because there is real content to save.
			if convEnded {
				newID, err := dbCon.GetNextConversationID(ctx, userID)
				if err != nil {
					logger.Error("failed to get next conversation ID", zap.Error(err))
					_ = wsWrite(ctx, conn, gin.H{"error": "failed to start new conversation"})
					continue
				}
				conversationID = newID
				convEnded = false

				// Re-subscribe under the same userID key for the new conversation.
				unsubFromAdmin()
				incomingFromAdminCh = make(chan hubMessage, 16)
				unsubFromAdmin = hub.subscribe(userID, incomingFromAdminCh)
			}

			userMsg := &models.ChatMessage{
				ConversationID: conversationID,
				UserID:         userID,
				Username:       username,
				Message:        res.msg,
				Sender:         models.SenderUser,
				Timestamp:      time.Now(),
			}
			// DB insert is kept synchronous here: the echo frame confirms
			// persistence, so we must not echo before the write succeeds.
			if err := dbCon.InsertChatMessage(userMsg); err != nil {
				logger.Error("failed to save user message", zap.Error(err))
				_ = wsWrite(ctx, conn, gin.H{"error": "failed to save message"})
				continue
			}

			// Push the new user message to any admin watching this conversation,
			// and also to the broadcast key so the admin overview WS can show
			// unread indicators without polling.
			msg := hubMessage{
				ConversationID: conversationID,
				UserID:         userID,
				Sender:         models.SenderUser,
				Message:        userMsg.Message,
				Timestamp:      userMsg.Timestamp.Format(time.RFC3339),
			}
			hub.publish(userToAdminKey(userID, conversationID), msg)
			hub.publish(adminBroadcastKey, msg)

			echo := map[string]interface{}{
				"conversation_id": conversationID,
				"message":         userMsg.Message,
				"sender":          models.SenderUser,
				"timestamp":       userMsg.Timestamp.Format(time.RFC3339),
			}
			if err := wsWrite(ctx, conn, echo); err != nil {
				logger.Error("failed to send echo", zap.Error(err))
				return
			}

		case adminMsg := <-incomingFromAdminCh:
			// An admin replied while the user is connected — push it live.
			// "conversation_ended" is a system signal, not a display message.
			if adminMsg.Message == "conversation_ended" {
				frame := map[string]interface{}{
					"type":            "ended",
					"conversation_id": adminMsg.ConversationID,
				}
				if err := wsWrite(ctx, conn, frame); err != nil {
					logger.Error("failed to send ended frame to user", zap.Error(err))
					return
				}
				continue
			}
			frame := map[string]interface{}{
				"conversation_id": adminMsg.ConversationID,
				"sender":          adminMsg.Sender,
				"message":         adminMsg.Message,
				"timestamp":       adminMsg.Timestamp,
			}
			if err := wsWrite(ctx, conn, frame); err != nil {
				logger.Error("failed to send admin message", zap.Error(err))
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// GetUserConversations returns the list of the calling user's own conversations.
func GetUserConversations(c *gin.Context) {
	logger := log.GetLogger()
	userID, _ := c.Request.Context().Value("userid").(string)
	summaries, err := dbCon.GetUserConversations(c.Request.Context(), userID)
	if err != nil {
		logger.Error("failed to get user conversations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": summaries})
}

// GetUserConversationMessages returns messages for one of the calling user's own conversations.
func GetUserConversationMessages(c *gin.Context) {
	logger := log.GetLogger()
	userID, _ := c.Request.Context().Value("userid").(string)
	convIDStr := c.Param("conv_id")

	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	messages, err := dbCon.GetChatMessages(c.Request.Context(), userID, convID)
	if err != nil {
		logger.Error("failed to get user conversation messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetAdminConversations returns a list of all user conversations for the admin panel.
// For any conversation whose username was not stored in the DB (legacy data), the
// username is resolved from Keycloak by userID and backfilled in the response.
func GetAdminConversations(c *gin.Context) {
	logger := log.GetLogger()
	summaries, err := dbCon.GetAllConversations(c.Request.Context())
	if err != nil {
		logger.Error("failed to get conversations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Collect unique userIDs that are missing a username (legacy messages).
	missing := map[string]struct{}{}
	for _, s := range summaries {
		if s.Username == "" {
			missing[s.UserID] = struct{}{}
		}
	}

	if len(missing) > 0 {
		config := pacClient.GetConfigFromContext(c.Request.Context())
		kc := pacClient.NewKeyCloakClient(config, c.Request.Context())
		resolved := make(map[string]string, len(missing))
		for uid := range missing {
			user, err := kc.GetUser(uid)
			if err != nil {
				logger.Warn("could not resolve username from Keycloak", zap.String("userID", uid), zap.Error(err))
				resolved[uid] = uid // fall back to raw UUID
				continue
			}
			if user.Username != nil {
				resolved[uid] = *user.Username
			} else {
				resolved[uid] = uid
			}
		}
		for i := range summaries {
			if summaries[i].Username == "" {
				summaries[i].Username = resolved[summaries[i].UserID]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"conversations": summaries})
}

// GetAdminConversationMessages returns all messages for a specific conversation.
func GetAdminConversationMessages(c *gin.Context) {
	logger := log.GetLogger()
	userID := c.Param("user_id")
	convIDStr := c.Param("conv_id")

	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	messages, err := dbCon.GetChatMessages(c.Request.Context(), userID, convID)
	if err != nil {
		logger.Error("failed to get messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// adminReplyWorker is a package-level buffered channel that decouples the
// AdminReply HTTP handler from the MongoDB write.  The handler enqueues the
// fully-constructed ChatMessage and returns HTTP 202 immediately; a single
// background goroutine drains the queue and persists each message.
//
// Buffer size 256: in the worst case (DB temporarily slow) this absorbs a
// burst of 256 admin replies before the handler starts blocking callers.
// If the channel is full the handler falls back to a synchronous write so
// that no reply is ever silently lost.
var adminReplyQueue = make(chan *models.ChatMessage, 256)

// StartAdminReplyWorker drains adminReplyQueue and persists each message to
// MongoDB.  It must be started once from main() before the HTTP server begins
// accepting requests.  It exits when ctx is cancelled (server shutdown).
func StartAdminReplyWorker(ctx context.Context) {
	logger := log.GetLogger()
	for {
		select {
		case msg := <-adminReplyQueue:
			if err := dbCon.AdminReplyToConversation(ctx, msg); err != nil {
				logger.Error("adminReplyWorker: failed to persist reply",
					zap.String("userID", msg.UserID),
					zap.Int64("conversationID", msg.ConversationID),
					zap.Error(err),
				)
			}
		case <-ctx.Done():
			// Drain any remaining items before exiting so in-flight replies
			// are not lost during a graceful shutdown.
			for {
				select {
				case msg := <-adminReplyQueue:
					if err := dbCon.AdminReplyToConversation(context.Background(), msg); err != nil {
						logger.Error("adminReplyWorker: shutdown drain failed",
							zap.String("userID", msg.UserID),
							zap.Error(err),
						)
					}
				default:
					logger.Info("adminReplyWorker: queue drained, exiting")
					return
				}
			}
		}
	}
}

// AdminReply posts an admin reply into an existing conversation.
func AdminReply(c *gin.Context) {
	logger := log.GetLogger()

	adminID, ok := c.Request.Context().Value("userid").(string)
	if !ok || adminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID := c.Param("user_id")
	convIDStr := c.Param("conv_id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var body struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid body: %s", err.Error())})
		return
	}

	msg := &models.ChatMessage{
		ConversationID: convID,
		UserID:         userID,
		Message:        body.Message,
		Sender:         models.SenderAdmin,
		Timestamp:      time.Now(),
	}

	// Attempt to enqueue the DB write for the background worker.
	// If the queue is full (worker backlogged), fall back to a synchronous
	// write so the reply is never silently dropped.
	select {
	case adminReplyQueue <- msg:
		// enqueued — worker will persist asynchronously
	default:
		logger.Warn("adminReplyQueue full, falling back to synchronous write",
			zap.String("userID", userID),
			zap.Int64("conversationID", convID),
		)
		if err := dbCon.AdminReplyToConversation(c.Request.Context(), msg); err != nil {
			logger.Error("failed to save admin reply", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Push the reply to the user's live WebSocket session (if connected).
	hub.publish(userID, hubMessage{
		ConversationID: convID,
		Sender:         models.SenderAdmin,
		Message:        body.Message,
		Timestamp:      msg.Timestamp.Format(time.RFC3339),
	})

	logger.Info("admin reply enqueued",
		zap.String("adminID", adminID),
		zap.String("targetUserID", userID),
		zap.Int64("conversationID", convID))

	c.JSON(http.StatusCreated, gin.H{"message": "reply sent"})
}

// AdminEndConversation lets an admin close a conversation on behalf of the user.
// It notifies the user's live WebSocket session via the hub so the user's UI
// transitions to a fresh conversation immediately.
func AdminEndConversation(c *gin.Context) {
	logger := log.GetLogger()

	userID := c.Param("user_id")
	convIDStr := c.Param("conv_id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	adminUsername, _ := c.Request.Context().Value("username").(string)
	_ = adminUsername // username stored on the user's messages, not the admin's

	// Persist the ended sentinel so the status survives reconnects.
	if err := dbCon.MarkConversationEnded(c.Request.Context(), userID, "", convID); err != nil {
		logger.Error("failed to mark conversation ended in DB", zap.Error(err))
		// Non-fatal: still notify the user and return success.
	}

	// Push an "ended" signal to the user's live WebSocket (if connected).
	hub.publish(userID, hubMessage{
		ConversationID: convID,
		Sender:         models.SenderSystem,
		Message:        "conversation_ended",
	})

	logger.Info("admin ended conversation",
		zap.String("userID", userID),
		zap.Int64("convID", convID))

	c.JSON(http.StatusOK, gin.H{"message": "conversation ended"})
}

// AdminWatchConversation upgrades an admin request to a WebSocket and streams
// new user messages for the specified conversation in real time.
// The admin opens this connection after selecting a conversation; from that
// point on, every message the user sends is forwarded here immediately without
// polling.
func AdminWatchConversation(c *gin.Context) {
	logger := log.GetLogger()

	userID := c.Param("user_id")
	convIDStr := c.Param("conv_id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	type unwrapper interface{ Unwrap() http.ResponseWriter }
	rw := http.ResponseWriter(c.Writer)
	if u, ok := c.Writer.(unwrapper); ok {
		rw = u.Unwrap()
	}
	conn, err := websocket.Accept(rw, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		logger.Error("admin watch: websocket upgrade failed", zap.Error(err))
		return
	}
	c.Writer.WriteHeaderNow()

	go func() {
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// incomingFromUserCh receives messages published by serveChat for this conversation.
		incomingFromUserCh := make(chan hubMessage, 16)
		unsubFromUser := hub.subscribe(userToAdminKey(userID, convID), incomingFromUserCh)
		defer unsubFromUser()

		logger.Info("admin watching conversation",
			zap.String("userID", userID),
			zap.Int64("convID", convID))

		for {
			select {
			case msg := <-incomingFromUserCh:
				var frame map[string]interface{}
				if msg.Sender == models.SenderSystem && msg.Message == "conversation_ended" {
					frame = map[string]interface{}{
						"type":            "ended",
						"conversation_id": msg.ConversationID,
					}
				} else {
					frame = map[string]interface{}{
						"conversation_id": msg.ConversationID,
						"sender":          msg.Sender,
						"message":         msg.Message,
						"timestamp":       msg.Timestamp,
					}
				}
				if err := wsWrite(ctx, conn, frame); err != nil {
					logger.Error("admin watch: write failed", zap.Error(err))
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// AdminWatchAll upgrades an admin request to a WebSocket that receives a
// notification for every incoming user message across ALL conversations.
// The admin panel uses this single connection to drive unread indicators in the
// sidebar without polling.  Each frame contains user_id and conversation_id so
// the frontend can match the right sidebar entry.
func AdminWatchAll(c *gin.Context) {
	logger := log.GetLogger()

	type unwrapper interface{ Unwrap() http.ResponseWriter }
	rw := http.ResponseWriter(c.Writer)
	if u, ok := c.Writer.(unwrapper); ok {
		rw = u.Unwrap()
	}
	conn, err := websocket.Accept(rw, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		logger.Error("admin watch-all: websocket upgrade failed", zap.Error(err))
		return
	}
	c.Writer.WriteHeaderNow()

	go func() {
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// incomingBroadcastCh receives a notification for every user message across all conversations.
		incomingBroadcastCh := make(chan hubMessage, 32)
		unsubFromBroadcast := hub.subscribe(adminBroadcastKey, incomingBroadcastCh)
		defer unsubFromBroadcast()

		logger.Info("admin watching all conversations")

		for {
			select {
			case msg := <-incomingBroadcastCh:
				frame := map[string]interface{}{
					"user_id":         msg.UserID,
					"conversation_id": msg.ConversationID,
					"sender":          msg.Sender,
				}
				if err := wsWrite(ctx, conn, frame); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
