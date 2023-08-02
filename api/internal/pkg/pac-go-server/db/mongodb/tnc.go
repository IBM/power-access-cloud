package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

// GetTermsAndConditionsByUserID gets the terms and conditions status for a user.
func (db *MongoDB) GetTermsAndConditionsByUserID(id string) (*models.TermsAndConditions, error) {
	var terms models.TermsAndConditions

	if id == "" {
		return nil, fmt.Errorf("user id is required")
	}

	filter := bson.D{{Key: "user_id", Value: id}}

	collection := db.Database.Collection("tnc")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	if err := collection.FindOne(ctx, filter).Decode(&terms); err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("error getting terms and coditions entries: %w", err)
	} else if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("no terms and coditions entries found for user id: %s, err: %w", id, err)
	}

	return &terms, nil
}

// AcceptTermsAndConditions updates the terms and conditions for a user. Creates a new entry if one does not exist or updates the existing one.
func (db *MongoDB) AcceptTermsAndConditions(terms *models.TermsAndConditions) error {
	collection := db.Database.Collection("tnc")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	filter := bson.D{{"user_id", terms.UserID}}
	update := bson.D{{"$set", bson.D{{"user_id", terms.UserID}, {"accepted", terms.Accepted}, {"accepted_at", terms.AcceptedAt}}}}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("error updating terms and coditions: %w", err)
	}

	return nil
}
