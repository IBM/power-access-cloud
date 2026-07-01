package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessage represents a single chat message in a conversation
type ChatMessage struct {
	// ID is the unique identifier for the message
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	// ConversationID is the sequential conversation identifier (1, 2, 3...)
	ConversationID int64 `json:"conversation_id" bson:"conversation_id"`
	// UserID is the Keycloak user identifier
	UserID string `json:"user_id" bson:"user_id"`
	// Message is the text content of the message
	Message string `json:"message" bson:"message"`
	// Sender indicates who sent the message ("user" or "admin")
	Sender string `json:"sender" bson:"sender"`
	// Timestamp is when the message was created
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}
