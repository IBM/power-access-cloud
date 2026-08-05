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

// chatSession holds the mutable per-connection state shared across the helper
// functions that make up the WebSocket serve loop.
type chatSession struct {
	conn     *websocket.Conn
	ctx      context.Context
	userID   string
	username string
	logger   *zap.Logger

	conversationID int64
	convEnded      bool
	// history tracks whether the current conversation already has messages.
	// It is non-nil (even if empty) after initChatSession, and is set to a
	// non-empty sentinel after the first user message so the auto-reply is
	// never sent twice for the same conversation.
	history []models.ChatMessage

	// incomingFromAdminCh receives hub messages pushed to this user.
	// unsubFromAdmin must be called on session teardown.
	incomingFromAdminCh chan hubMessage
	unsubFromAdmin      func()
}

// initChatSession loads conversation state from the DB, sends the hello frame,
// and subscribes to the hub.  Returns an error if the session cannot start
// (e.g. DB failure or failed hello write); the caller should return immediately.
func initChatSession(ctx context.Context, conn *websocket.Conn, userID, username string, logger *zap.Logger) (*chatSession, error) {
	conversationID, err := dbCon.GetCurrentConversationID(ctx, userID)
	if err != nil {
		logger.Error("failed to get conversation ID", zap.Error(err))
		_ = wsWrite(ctx, conn, gin.H{"error": "failed to initialize conversation"})
		return nil, err
	}

	convEnded, err := dbCon.IsConversationEnded(ctx, userID, conversationID)
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
		ID             string               `json:"id"`
		ConversationID int64                `json:"conversation_id"`
		Sender         models.MessageSender `json:"sender"`
		Message        string               `json:"message"`
		Timestamp      string               `json:"timestamp"`
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
		"ended":           convEnded,
		"history":         historyFrames,
	}
	if err := wsWrite(ctx, conn, hello); err != nil {
		logger.Error("failed to send hello frame", zap.Error(err))
		return nil, err
	}

	logger.Info("conversation resumed",
		zap.String("userID", userID),
		zap.Int64("conversationID", conversationID),
		zap.Int("history_len", len(history)))

	// incomingFromAdminCh is the channel through which the hub delivers admin
	// and system messages to this session.  AdminReply and AdminEndConversation
	// publish to the hub under the userID key; the hub fans out to this channel
	// so the select loop can push the message to the connected client over
	// WebSocket without polling.  Buffered (16) so a slow WebSocket write does
	// not block the HTTP handler that called hub.publish.
	incomingFromAdminCh := make(chan hubMessage, 16)
	unsubFromAdmin := hub.subscribe(userID, incomingFromAdminCh)

	return &chatSession{
		conn:                conn,
		ctx:                 ctx,
		userID:              userID,
		username:            username,
		logger:              logger,
		conversationID:      conversationID,
		convEnded:           convEnded,
		history:             history,
		incomingFromAdminCh: incomingFromAdminCh,
		unsubFromAdmin:      unsubFromAdmin,
	}, nil
}

// handleEndConversation processes the "end_conversation" action frame.
// Returns true if the WebSocket connection must be closed (fatal write error).
func (s *chatSession) handleEndConversation() (fatal bool) {
	if s.convEnded {
		return false // already ended — ignore duplicate
	}
	s.convEnded = true

	// Persist the ended sentinel so the status survives reconnects.
	if err := dbCon.MarkConversationEnded(s.ctx, s.userID, s.username, s.conversationID); err != nil {
		s.logger.Error("failed to mark conversation ended", zap.Error(err))
	}

	// Notify any admin watching this conversation.
	hub.publish(userToAdminKey(s.userID, s.conversationID), hubMessage{
		ConversationID: s.conversationID,
		Sender:         models.SenderSystem,
		Message:        "conversation_ended",
	})

	frame := map[string]interface{}{
		"type":            "ended",
		"conversation_id": s.conversationID,
	}
	if err := wsWrite(s.ctx, s.conn, frame); err != nil {
		s.logger.Error("failed to send ended frame", zap.Error(err))
		return true
	}
	s.logger.Info("conversation ended by user",
		zap.String("userID", s.userID),
		zap.Int64("conversationID", s.conversationID))
	return false
}

// handleUserMessage saves the user message, publishes it to the hub, echoes it
// back, and fires the one-time auto-reply on the first message of each conversation.
// Returns true if the WebSocket connection must be closed (fatal write error).
// Returns (false, true) as (fatal, skip) when a non-fatal error means the
// message should be skipped (e.g. DB failure getting next conv ID).
func (s *chatSession) handleUserMessage(msg string) (fatal bool, skip bool) {
	// If the previous conversation was ended, lazily promote to a new conv ID
	// now that there is real content to save.
	isNewConv := false
	if s.convEnded {
		newID, err := dbCon.GetNextConversationID(s.ctx, s.userID)
		if err != nil {
			s.logger.Error("failed to get next conversation ID", zap.Error(err))
			_ = wsWrite(s.ctx, s.conn, gin.H{"error": "failed to start new conversation"})
			return false, true
		}
		s.conversationID = newID
		s.convEnded = false
		isNewConv = true

		// Re-subscribe so hub messages for the new conv reach this session.
		s.unsubFromAdmin()
		s.incomingFromAdminCh = make(chan hubMessage, 16)
		s.unsubFromAdmin = hub.subscribe(s.userID, s.incomingFromAdminCh)
	}

	// isFirstMessage is true for the very first message of each conversation:
	// either a brand-new conv (isNewConv) or the opening message of conv 1.
	isFirstMessage := isNewConv || len(s.history) == 0

	userMsg := &models.ChatMessage{
		ConversationID: s.conversationID,
		UserID:         s.userID,
		Username:       s.username,
		Message:        msg,
		Sender:         models.SenderUser,
		Timestamp:      time.Now(),
	}
	// DB insert is synchronous: we echo only after persistence is confirmed.
	if err := dbCon.InsertChatMessage(s.ctx, userMsg); err != nil {
		s.logger.Error("failed to save user message", zap.Error(err))
		_ = wsWrite(s.ctx, s.conn, gin.H{"error": "failed to save message"})
		return false, true
	}
	// Mark history non-empty so the auto-reply is not re-triggered.
	if isFirstMessage {
		s.history = []models.ChatMessage{{}}
	}

	// Publish to any admin watching this conversation and to the broadcast key.
	hubMsg := hubMessage{
		ConversationID: s.conversationID,
		UserID:         s.userID,
		Sender:         models.SenderUser,
		Message:        userMsg.Message,
		Timestamp:      userMsg.Timestamp.Format(time.RFC3339),
	}
	hub.publish(userToAdminKey(s.userID, s.conversationID), hubMsg)
	hub.publish(adminBroadcastKey, hubMsg)

	echo := map[string]interface{}{
		"conversation_id": s.conversationID,
		"message":         userMsg.Message,
		"sender":          models.SenderUser,
		"timestamp":       userMsg.Timestamp.Format(time.RFC3339),
	}
	if err := wsWrite(s.ctx, s.conn, echo); err != nil {
		s.logger.Error("failed to send echo", zap.Error(err))
		return true, false
	}

	if isFirstMessage {
		if fatal := s.sendAutoReply(); fatal {
			return true, false
		}
	}
	return false, false
}

// sendAutoReply saves and echoes the one-time automated system reply sent at
// the start of every new conversation.
// Returns true if the WebSocket connection must be closed (fatal write error).
func (s *chatSession) sendAutoReply() (fatal bool) {
	const autoReplyText = "Thanks for reaching out! An admin will reply within 48 hours. " +
		"In the meantime, feel free to browse our FAQ for quick answers: " +
		"https://github.com/IBM/power-access-cloud/blob/main/support/docs/FAQ.md"

	autoReply := &models.ChatMessage{
		ConversationID: s.conversationID,
		UserID:         s.userID,
		Username:       s.username,
		Message:        autoReplyText,
		Sender:         models.SenderSystem,
		Timestamp:      time.Now(),
	}
	if err := dbCon.InsertChatMessage(s.ctx, autoReply); err != nil {
		s.logger.Error("failed to save auto-reply", zap.Error(err))
		return false // non-fatal: user message was already saved and echoed
	}
	frame := map[string]interface{}{
		"conversation_id": s.conversationID,
		"message":         autoReply.Message,
		"sender":          models.SenderSystem,
		"timestamp":       autoReply.Timestamp.Format(time.RFC3339),
	}
	if err := wsWrite(s.ctx, s.conn, frame); err != nil {
		s.logger.Error("failed to send auto-reply frame", zap.Error(err))
		return true
	}
	return false
}

// handleAdminMessage pushes an incoming hub message (admin reply or remote end)
// to the connected user.
// Returns true if the WebSocket connection must be closed (fatal write error).
func (s *chatSession) handleAdminMessage(adminMsg hubMessage) (fatal bool) {
	// "conversation_ended" is a system signal, not a display message.
	if adminMsg.Message == "conversation_ended" {
		// Mark the session as ended so the next user message correctly
		// triggers a new conversation ID instead of saving to the old one.
		s.convEnded = true
		frame := map[string]interface{}{
			"type":            "ended",
			"conversation_id": adminMsg.ConversationID,
		}
		if err := wsWrite(s.ctx, s.conn, frame); err != nil {
			s.logger.Error("failed to send ended frame to user", zap.Error(err))
			return true
		}
		return false
	}
	frame := map[string]interface{}{
		"conversation_id": adminMsg.ConversationID,
		"sender":          adminMsg.Sender,
		"message":         adminMsg.Message,
		"timestamp":       adminMsg.Timestamp,
	}
	if err := wsWrite(s.ctx, s.conn, frame); err != nil {
		s.logger.Error("failed to send admin message", zap.Error(err))
		return true
	}
	return false
}

// serveChat runs the full WebSocket session for one connected user.
func serveChat(conn *websocket.Conn, userID, username string, logger *zap.Logger) {
	// Use a cancellable context tied to the connection lifetime.
	// When the client disconnects (or the server closes the conn), cancel
	// unblocks any in-flight wsjson.Read/Write immediately — no goroutine leak.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, err := initChatSession(ctx, conn, userID, username, logger)
	if err != nil {
		return
	}
	defer session.unsubFromAdmin()

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
				return
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

			if res.action == "end_conversation" {
				if fatal := session.handleEndConversation(); fatal {
					return
				}
				continue
			}

			if res.msg == "" {
				continue
			}

			if fatal, skip := session.handleUserMessage(res.msg); fatal || skip {
				if fatal {
					return
				}
				continue
			}

		case adminMsg := <-session.incomingFromAdminCh:
			if fatal := session.handleAdminMessage(adminMsg); fatal {
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
			if err := dbCon.InsertChatMessage(ctx, msg); err != nil {
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
					if err := dbCon.InsertChatMessage(context.Background(), msg); err != nil {
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
		if err := dbCon.InsertChatMessage(c.Request.Context(), msg); err != nil {
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

	// Also push to every admin watching this conversation so all admin UIs
	// see the new message in real time (including the sender's own tab).
	hub.publish(userToAdminKey(userID, convID), hubMessage{
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

		// Drain the WebSocket read side in a separate goroutine so that a
		// client disconnect (tab close / network drop) is detected immediately
		// and cancels ctx, unblocking the select below.  Without this, the
		// goroutine leaks until the next write attempt.
		go func() {
			defer cancel()
			for {
				if _, _, err := conn.Read(ctx); err != nil {
					return
				}
			}
		}()

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

		// Same disconnect-detection pattern as AdminWatchConversation: drain
		// the read side so a client close/drop cancels ctx immediately.
		go func() {
			defer cancel()
			for {
				if _, _, err := conn.Read(ctx); err != nil {
					return
				}
			}
		}()

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
