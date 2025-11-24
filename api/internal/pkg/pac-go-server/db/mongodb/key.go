package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
)

func (db *MongoDB) GetKeyByUserID(id string) ([]models.Key, error) {
	keys := []models.Key{}

	filter := bson.D{{}}
	if id != "" {
		filter = bson.D{{Key: "user_id", Value: id}}
	}

	collection := db.Database.Collection("keys")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error getting keys: %w", err)
	}
	defer cur.Close(ctx)

	if err = cur.All(context.TODO(), &keys); err != nil {
		return nil, fmt.Errorf("error fetching keys: %w", err)
	}

	return keys, nil
}

func (db *MongoDB) CreateKey(keys *models.Key) error {
	collection := db.Database.Collection("keys")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	_, err := collection.InsertOne(ctx, keys)
	if err != nil {
		return fmt.Errorf("error inserting Key: %w", err)
	}

	return nil
}

// GetKeyByID returns a key by its ID
func (db *MongoDB) GetKeyByID(id string) (*models.Key, error) {
	var key models.Key

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}
	filter := bson.M{"_id": objectId}

	collection := db.Database.Collection("keys")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	err = collection.FindOne(ctx, filter).Decode(&key)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, utils.ErrResourceNotFound
		}
		return nil, fmt.Errorf("error getting key: %w", err)
	}

	return &key, nil
}

func (db *MongoDB) DeleteKey(id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	collection := db.Database.Collection("keys")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	_, err = collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: objectId}})
	if err != nil {
		return fmt.Errorf("error deleting key: %w", err)
	}

	return nil
}
