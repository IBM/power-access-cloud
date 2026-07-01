package services

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now - should be restricted in production
		return true
	},
}

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
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}
	defer conn.Close()
	
	logger.Info("WebSocket connection established", zap.String("userID", userIDStr))
	
	// Get or create conversation ID for this user
	ctx := context.Background()
	conversationID, err := dbCon.GetNextConversationID(ctx, userIDStr)
	if err != nil {
		logger.Error("failed to get conversation ID", zap.Error(err))
		conn.WriteJSON(gin.H{"error": "Failed to initialize conversation"})
		return
	}
	
	logger.Info("Conversation initialized", 
		zap.String("userID", userIDStr), 
		zap.Int64("conversationID", conversationID))
	
	// Handle incoming messages
	for {
		var incomingMsg models.ChatMessage
		err := conn.ReadJSON(&incomingMsg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("websocket error", zap.Error(err))
			} else {
				logger.Info("websocket connection closed", zap.String("userID", userIDStr))
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
			conn.WriteJSON(gin.H{"error": "Failed to save message"})
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
			conn.WriteJSON(gin.H{"error": "Failed to save response"})
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
		
		if err := conn.WriteJSON(response); err != nil {
			logger.Error("failed to send response", zap.Error(err))
			break
		}
		
		logger.Info("Response sent to client", zap.String("userID", userIDStr))
	}
}
