package mongodb

import (
	"context"
	"fmt"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"go.mongodb.org/mongo-driver/bson"
)

// InsertFeedback - insert feedback record into the DB
func (db *MongoDB) InsertFeedback(feedback *models.Feedback) error {
	collection := db.Database.Collection("feedbacks")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	if _, err := collection.InsertOne(ctx, feedback); err != nil {
		return fmt.Errorf("error inserting feedback: %w", err)
	}

	return nil
}

// GetFeedbacks - returns list of feedbacks based on the provided filter
func (db *MongoDB) GetFeedbacks(filter models.FeedbacksFilter) ([]models.Feedback, error) {
	collection := db.Database.Collection("feedbacks")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	cursor, err := collection.Find(ctx, buildFilter(filter))
	if err != nil {
		return nil, fmt.Errorf("error fetching feedbacks from DB: %w", err)
	}
	defer cursor.Close(ctx)

	feedbacks := []models.Feedback{}
	if err := cursor.All(context.TODO(), &feedbacks); err != nil {
		return nil, fmt.Errorf("error getting feedbacks: %w", err)
	}

	return feedbacks, nil
}

func buildFilter(filter models.FeedbacksFilter) bson.M {
	bsonFilter := bson.M{}

	// Conditionally add filters based on whether fields are set
	if filter.UserID != "" {
		bsonFilter["user_id"] = filter.UserID
	}

	return bsonFilter
}
