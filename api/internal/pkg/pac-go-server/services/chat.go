package services

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
)

// HandleChatWebSocket handles WebSocket connections for chat
func HandleChatWebSocket(c *gin.Context) {
	logger := log.GetLogger()
	
	// Get user ID from context (set by Keycloak middleware)
	userID, exists := c.Get("userid")
	if !exists {
		logger.Error("user ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userIDStr := userID.(string)
	logger.Info("WebSocket connection request", zap.String("userID", userIDStr))
	
	// Upgrade HTTP connection to WebSocket
	// coder/websocket uses Accept instead of Upgrade
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // Allow all origins for now - should be restricted in production
	})
	if err != nil {
		logger.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}
	defer conn.Close(websocket.StatusInternalError, "internal error")
	
	logger.Info("WebSocket connection established", zap.String("userID", userIDStr))
	
	// Get or create conversation ID for this user
	ctx := context.Background()
	conversationID, err := dbCon.GetNextConversationID(ctx, userIDStr)
	if err != nil {
		logger.Error("failed to get conversation ID", zap.Error(err))
		conn.Close(websocket.StatusInternalError, "Failed to initialize conversation")
		return
	}
	
	logger.Info("Conversation initialized",
		zap.String("userID", userIDStr),
		zap.Int64("conversationID", conversationID))
	
	// Handle incoming messages
	for {
		var incomingMsg models.ChatMessage
		err := wsjson.Read(ctx, conn, &incomingMsg)
		if err != nil {
			// Check if it's a normal close
			closeStatus := websocket.CloseStatus(err)
			if closeStatus == websocket.StatusNormalClosure || closeStatus == websocket.StatusGoingAway {
				logger.Info("websocket connection closed normally", zap.String("userID", userIDStr))
			} else if closeStatus != -1 {
				logger.Error("websocket error", zap.Error(err), zap.Int("closeStatus", int(closeStatus)))
			} else {
				logger.Error("websocket read error", zap.Error(err))
			}
			break
		}
		
		logger.Info("Received message",
			zap.String("userID", userIDStr),
			zap.String("message", incomingMsg.Message))
		
		// Create user message
		userMessage := &models.ChatMessage{
			ConversationID: conversationID,
			UserID:         userIDStr,
			Message:        incomingMsg.Message,
			Sender:         "user",
			Timestamp:      time.Now(),
		}
		
		// Save user message to MongoDB
		if err := dbCon.InsertChatMessage(userMessage); err != nil {
			logger.Error("failed to save user message", zap.Error(err))
			wsjson.Write(ctx, conn, gin.H{"error": "Failed to save message"})
			continue
		}
		
		logger.Info("User message saved to database")
		
		// Send automated "test" response
		adminMessage := &models.ChatMessage{
			ConversationID: conversationID,
			UserID:         userIDStr,
			Message:        "test",
			Sender:         "admin",
			Timestamp:      time.Now(),
		}
		
		// Save admin response to MongoDB
		if err := dbCon.InsertChatMessage(adminMessage); err != nil {
			logger.Error("failed to save admin message", zap.Error(err))
			wsjson.Write(ctx, conn, gin.H{"error": "Failed to save response"})
			continue
		}
		
		logger.Info("Admin response saved to database")
		
		// Send response back to client
		response := map[string]interface{}{
			"conversation_id": conversationID,
			"message":         "test",
			"sender":          "admin",
			"timestamp":       adminMessage.Timestamp.Format(time.RFC3339),
		}
		
		if err := wsjson.Write(ctx, conn, response); err != nil {
			logger.Error("failed to send response", zap.Error(err))
			break
		}
		
		logger.Info("Response sent to client", zap.String("userID", userIDStr))
	}
	
	conn.Close(websocket.StatusNormalClosure, "")
}
