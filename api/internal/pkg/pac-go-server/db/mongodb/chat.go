package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const chatCollection = "chat_messages"

// BSON field name constants for the chat_messages collection.
// Using constants prevents silent query failures from typos in field name strings.
const (
	fieldUserID         = "user_id"
	fieldConversationID = "conversation_id"
	fieldSender         = "sender"
	fieldMessage        = "message"
	fieldTimestamp      = "timestamp"
	fieldUsername       = "username"
)

// MarkConversationEnded writes a system sentinel message so the ended state
// survives server restarts and page navigations.
func (db *MongoDB) MarkConversationEnded(ctx context.Context, userID, username string, conversationID int64) error {
	sentinel := &models.ChatMessage{
		ConversationID: conversationID,
		UserID:         userID,
		Username:       username,
		Message:        EndedSentinel,
		Sender:         models.SenderSystem,
		Timestamp:      time.Now(),
	}
	return db.InsertChatMessage(ctx, sentinel)
}

// IsConversationEnded returns true if the given conversation has an ended sentinel.
func (db *MongoDB) IsConversationEnded(ctx context.Context, userID string, conversationID int64) (bool, error) {
	collection := db.Database.Collection(chatCollection)
	filter := bson.M{
		fieldUserID:         userID,
		fieldConversationID: conversationID,
		fieldSender:         models.SenderSystem,
		fieldMessage:        EndedSentinel,
	}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("error checking conversation ended: %w", err)
	}
	return count > 0, nil
}

// InsertChatMessage saves a single chat message to the DB.
// The caller-supplied ctx is used as the parent for the write timeout so that
// request cancellations (e.g. client disconnect) propagate correctly.
func (db *MongoDB) InsertChatMessage(ctx context.Context, msg *models.ChatMessage) error {
	collection := db.Database.Collection(chatCollection)
	ctxT, cancel := context.WithTimeout(ctx, dbContextTimeout)
	defer cancel()

	if _, err := collection.InsertOne(ctxT, msg); err != nil {
		return fmt.Errorf("error inserting chat message: %w", err)
	}
	return nil
}

// GetCurrentConversationID returns the latest conversation ID for a user,
// or 1 if the user has no messages yet. Unlike GetNextConversationID it does
// NOT increment — reconnecting always resumes the same conversation.
//
// All message senders (user, admin, system) are considered so that a new
// conversation that only has a system sentinel so far (e.g. the user started
// a new conv but the server restart happened before the first user message)
// is still correctly identified as the current conversation.
func (db *MongoDB) GetCurrentConversationID(ctx context.Context, userID string) (int64, error) {
	collection := db.Database.Collection(chatCollection)

	opts := options.FindOne().SetSort(bson.D{{Key: fieldConversationID, Value: -1}})
	filter := bson.M{fieldUserID: userID}

	var result models.ChatMessage
	err := collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 1, nil
		}
		return 0, fmt.Errorf("error finding current conversation_id: %w", err)
	}
	return result.ConversationID, nil
}

// GetNextConversationID returns a brand-new conversation ID (latest + 1).
// Use this only when the user explicitly starts a new conversation.
func (db *MongoDB) GetNextConversationID(ctx context.Context, userID string) (int64, error) {
	collection := db.Database.Collection(chatCollection)

	opts := options.FindOne().SetSort(bson.D{{Key: fieldConversationID, Value: -1}})
	filter := bson.M{fieldUserID: userID}

	var result models.ChatMessage
	err := collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 1, nil
		}
		return 0, fmt.Errorf("error finding max conversation_id: %w", err)
	}
	return result.ConversationID + 1, nil
}

// GetChatMessages returns all displayable messages for a user's conversation, oldest first.
// System sentinel messages (conversation_ended) are excluded; other system messages (e.g. auto-reply) are included.
func (db *MongoDB) GetChatMessages(ctx context.Context, userID string, conversationID int64) ([]models.ChatMessage, error) {
	collection := db.Database.Collection(chatCollection)

	filter := bson.M{
		fieldUserID:         userID,
		fieldConversationID: conversationID,
		// Exclude only the ended sentinel; keep system auto-reply messages.
		"$nor": bson.A{
			bson.M{fieldSender: models.SenderSystem, fieldMessage: EndedSentinel},
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: fieldTimestamp, Value: 1}})

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

// EndedSentinel is the message text written to the DB when a conversation ends.
// It is used to detect ended conversations without a separate collection.
const EndedSentinel = "conversation_ended"

// GetUserConversations returns conversation summaries for a single user, ordered by most recent.
// Each summary includes Ended (true if a system sentinel exists) and FirstMessage (first user text).
func (db *MongoDB) GetUserConversations(ctx context.Context, userID string) ([]models.ConversationSummary, error) {
	collection := db.Database.Collection(chatCollection)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{fieldUserID: userID}}},
		{
			{Key: "$sort", Value: bson.D{{Key: fieldTimestamp, Value: 1}}},
		},
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$" + fieldConversationID},
				{Key: "message_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "last_message_at", Value: bson.D{{Key: "$max", Value: "$" + fieldTimestamp}}},
				{Key: fieldUsername, Value: bson.D{{Key: "$first", Value: "$" + fieldUsername}}},
				// Collect all (sender, message) pairs to derive ended + first_message.
				{Key: "msgs", Value: bson.D{{Key: "$push", Value: bson.D{
					{Key: fieldSender, Value: "$" + fieldSender},
					{Key: fieldMessage, Value: "$" + fieldMessage},
				}}}},
			}},
		},
		{{Key: "$sort", Value: bson.D{{Key: "last_message_at", Value: -1}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating user conversations: %w", err)
	}
	defer cursor.Close(ctx)

	type msgPair struct {
		Sender  models.MessageSender `bson:"sender"`
		Message string               `bson:"message"`
	}
	type rawSummary struct {
		ConversationID int64     `bson:"_id"`
		MessageCount   int64     `bson:"message_count"`
		LastMessageAt  time.Time `bson:"last_message_at"`
		Username       string    `bson:"username"`
		Msgs           []msgPair `bson:"msgs"`
	}

	var raw []rawSummary
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, fmt.Errorf("error decoding user conversations: %w", err)
	}

	summaries := make([]models.ConversationSummary, 0, len(raw))
	for _, r := range raw {
		ended := false
		firstMsg := ""
		for _, m := range r.Msgs {
			if m.Sender == models.SenderSystem && m.Message == EndedSentinel {
				ended = true
			}
			if firstMsg == "" && m.Sender == models.SenderUser {
				firstMsg = m.Message
			}
		}
		summaries = append(summaries, models.ConversationSummary{
			ConversationID: r.ConversationID,
			UserID:         userID,
			Username:       r.Username,
			MessageCount:   r.MessageCount,
			LastMessageAt:  r.LastMessageAt,
			Ended:          ended,
			FirstMessage:   firstMsg,
		})
	}
	return summaries, nil
}

// GetAllConversations returns one summary row per unique (user_id, conversation_id) pair,
// ordered by the most recent activity, including ended status.
func (db *MongoDB) GetAllConversations(ctx context.Context) ([]models.ConversationSummary, error) {
	collection := db.Database.Collection(chatCollection)

	// Pipeline overview — runs entirely inside MongoDB, no per-row round-trips:
	//
	// Stage 1 — $group by (user_id, conversation_id):
	//   • message_count  — total number of messages in the conversation.
	//   • last_message_at — timestamp of the most recent message; used for sort.
	//   • username       — first username seen (all rows share the same value).
	//   • senders        — distinct (sender, message) pairs collected via
	//                      $addToSet.  This is NOT the full message history;
	//                      it is a small deduplicated set used only to check
	//                      whether the ended sentinel is present.  Message text
	//                      is loaded separately by GetAdminConversationMessages
	//                      when the admin opens a specific conversation.
	//
	// Stage 2 — $sort by last_message_at descending (newest conversation first).
	pipeline := mongo.Pipeline{
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: fieldUserID, Value: "$" + fieldUserID},
					{Key: fieldConversationID, Value: "$" + fieldConversationID},
				}},
				{Key: "message_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "last_message_at", Value: bson.D{{Key: "$max", Value: "$" + fieldTimestamp}}},
				{Key: fieldUsername, Value: bson.D{{Key: "$first", Value: "$" + fieldUsername}}},
				{Key: "senders", Value: bson.D{{Key: "$addToSet", Value: bson.D{
					{Key: fieldSender, Value: "$" + fieldSender},
					{Key: fieldMessage, Value: "$" + fieldMessage},
				}}}},
			}},
		},
		{{Key: "$sort", Value: bson.D{{Key: "last_message_at", Value: -1}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating conversations: %w", err)
	}
	defer cursor.Close(ctx)

	type senderPair struct {
		Sender  models.MessageSender `bson:"sender"`
		Message string               `bson:"message"`
	}
	type rawSummary struct {
		ID struct {
			UserID         string `bson:"user_id"`
			ConversationID int64  `bson:"conversation_id"`
		} `bson:"_id"`
		MessageCount  int64        `bson:"message_count"`
		LastMessageAt time.Time    `bson:"last_message_at"`
		Username      string       `bson:"username"`
		Senders       []senderPair `bson:"senders"`
	}

	var raw []rawSummary
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, fmt.Errorf("error decoding conversations: %w", err)
	}

	summaries := make([]models.ConversationSummary, 0, len(raw))
	for _, r := range raw {
		ended := false
		for _, s := range r.Senders {
			if s.Sender == models.SenderSystem && s.Message == EndedSentinel {
				ended = true
				break
			}
		}
		summaries = append(summaries, models.ConversationSummary{
			ConversationID: r.ID.ConversationID,
			UserID:         r.ID.UserID,
			Username:       r.Username,
			MessageCount:   r.MessageCount,
			LastMessageAt:  r.LastMessageAt,
			Ended:          ended,
		})
	}
	return summaries, nil
}
