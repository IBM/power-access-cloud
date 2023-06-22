package mongodb

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

func (db *MongoDB) GetKeyByUserID(id string) ([]models.Key, error) {
	var keys []models.Key

	filter := bson.D{{}}
	if id != "" {
		filter = bson.D{{"user_id", id}}
	}

	collection := db.Database.Collection("keys")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
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
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	_, err := collection.InsertOne(ctx, keys)
	if err != nil {
		return fmt.Errorf("error inserting Key: %w", err)
	}

	return nil
}

// GetRequestByID returns a request by its ID
func (db *MongoDB) GetKeyByID(id string) (*models.Key, error) {
	var key models.Key

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}
	filter := bson.M{"_id": objectId}

	collection := db.Database.Collection("keys")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	err = collection.FindOne(ctx, filter).Decode(&key)
	if err == mongo.ErrNoDocuments {
		log.Println("no documents found")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting request: %w", err)
	}

	return &key, nil
}

func (db *MongoDB) DeleteKey(id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	collection := db.Database.Collection("keys")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	_, err = collection.DeleteOne(ctx, bson.D{{"_id", objectId}})
	if err != nil {
		return fmt.Errorf("error deleting request: %w", err)
	}

	return nil
}
