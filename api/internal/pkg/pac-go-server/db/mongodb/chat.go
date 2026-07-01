package mongodb

import (
	"context"
	"fmt"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertChatMessage - insert chat message into the DB
func (db *MongoDB) InsertChatMessage(message *models.ChatMessage) error {
	collection := db.Database.Collection("chat_messages")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	if _, err := collection.InsertOne(ctx, message); err != nil {
		return fmt.Errorf("error inserting chat message: %w", err)
	}

	return nil
}

// GetNextConversationID - get the next available conversation ID for a user
func (db *MongoDB) GetNextConversationID(ctx context.Context, userID string) (int64, error) {
	collection := db.Database.Collection("chat_messages")
	
	// Find the highest conversation_id for this user
	opts := options.FindOne().SetSort(bson.D{{Key: "conversation_id", Value: -1}})
	filter := bson.M{"user_id": userID}
	
	var result models.ChatMessage
	err := collection.FindOne(ctx, filter, opts).Decode(&result)
	
	if err != nil {
		// If no documents found, this is the first conversation
		if err.Error() == "mongo: no documents in result" {
			return 1, nil
		}
		return 0, fmt.Errorf("error finding max conversation_id: %w", err)
	}
	
	// Return next conversation ID
	return result.ConversationID + 1, nil
}

// GetChatMessages - get all messages for a specific conversation
func (db *MongoDB) GetChatMessages(ctx context.Context, userID string, conversationID int64) ([]models.ChatMessage, error) {
	collection := db.Database.Collection("chat_messages")
	
	filter := bson.M{
		"user_id":         userID,
		"conversation_id": conversationID,
	}
	
	// Sort by timestamp ascending (oldest first)
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching chat messages: %w", err)
	}
	defer cursor.Close(ctx)
	
	var messages []models.ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, fmt.Errorf("error decoding chat messages: %w", err)
	}
	
	return messages, nil
}
