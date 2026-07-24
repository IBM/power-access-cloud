package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Sender constants for ChatMessage.Sender.
const (
	SenderUser   = "user"
	SenderAdmin  = "admin"
	SenderSystem = "system"
)

// ChatMessage represents a single message in a support conversation.
type ChatMessage struct {
	ID             bson.ObjectID `json:"id" bson:"_id,omitempty"`
	ConversationID int64              `json:"conversation_id" bson:"conversation_id"`
	UserID         string             `json:"user_id" bson:"user_id"`
	Username       string             `json:"username" bson:"username"` // preferred_username from Keycloak
	Message        string             `json:"message" bson:"message"`
	// Sender is one of SenderUser, SenderAdmin, or SenderSystem.
	Sender    string    `json:"sender" bson:"sender"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

// ConversationSummary is returned to list conversations.
type ConversationSummary struct {
	ConversationID int64     `json:"conversation_id" bson:"conversation_id"`
	UserID         string    `json:"user_id" bson:"user_id"`
	Username       string    `json:"username"` // preferred_username, for display
	MessageCount   int64     `json:"message_count"`
	LastMessageAt  time.Time `json:"last_message_at" bson:"last_message_at"`
	Ended          bool      `json:"ended"`
	FirstMessage   string    `json:"first_message"` // preview label for the user sidebar
}
