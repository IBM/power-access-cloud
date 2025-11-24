package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (db *MongoDB) GetFeedbacks(filter models.FeedbacksFilter, startIndex, perPage int64) ([]models.Feedback, int64, error) {

	var totalCount int64
	findOptions := options.Find()
	findOptions.SetSkip(startIndex)
	findOptions.SetLimit(perPage)

	collection := db.Database.Collection("feedbacks")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	cursor, err := collection.Find(ctx, buildFilter(filter), findOptions)
	if err != nil {
		return nil, totalCount, fmt.Errorf("error fetching feedbacks from DB: %w", err)
	}
	defer cursor.Close(ctx)

	feedbacks := []models.Feedback{}
	if err := cursor.All(context.TODO(), &feedbacks); err != nil {
		return nil, totalCount, fmt.Errorf("error getting feedbacks: %w", err)
	}
	totalCount, err = collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, totalCount, fmt.Errorf("error getting total count of events: %w", err)
	}
	return feedbacks, totalCount, nil
}

func buildFilter(filter models.FeedbacksFilter) bson.M {
	bsonFilter := bson.M{}

	// Conditionally add filters based on whether fields are set
	if filter.UserID != "" {
		bsonFilter["user_id"] = filter.UserID
	}

	return bsonFilter
}

// FeedbackAllowed - checks if user has reached the limit of feedback submission.
func (db *MongoDB) FeedbackAllowed(ctx context.Context, userID string) (bool, error) {
	startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)
	collection := db.Database.Collection("feedbacks")
	filter := bson.M{
		"user_id": userID,
		"created_at": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count < services.FeedbackLimit, nil
}
