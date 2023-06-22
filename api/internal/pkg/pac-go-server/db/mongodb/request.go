package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

func (db *MongoDB) GetRequestsByUserID(id string) ([]models.Request, error) {
	var requests []models.Request

	filter := bson.D{{}}
	if id != "" {
		filter = bson.D{{"user_id", id}}
	}

	collection := db.Database.Collection("requests")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error getting requests: %w", err)
	}
	defer cur.Close(ctx)

	if err = cur.All(context.TODO(), &requests); err != nil {
		return nil, fmt.Errorf("error fetching requests: %w", err)
	}

	return requests, nil
}

func (db *MongoDB) NewRequest(request *models.Request) error {
	collection := db.Database.Collection("requests")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	_, err := collection.InsertOne(ctx, request)
	if err != nil {
		return fmt.Errorf("error inserting request: %w", err)
	}

	return nil
}

func (db *MongoDB) GetRequestByGroupIDAndUserID(groupID string, userID string) ([]models.Request, error) {
	var requests []models.Request

	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"group.group_id", groupID}},
				bson.D{{"user_id", userID}},
			}},
	}

	collection := db.Database.Collection("requests")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error getting request: %w", err)
	}
	defer cur.Close(ctx)
	if err = cur.All(context.TODO(), &requests); err != nil {
		return nil, fmt.Errorf("error fetching requests: %w", err)
	}
	return requests, nil
}

// GetRequestByID returns a request by its ID
func (db *MongoDB) GetRequestByID(id string) (*models.Request, error) {
	var request models.Request

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}
	filter := bson.M{"_id": objectId}

	collection := db.Database.Collection("requests")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	err = collection.FindOne(ctx, filter).Decode(&request)
	if err == mongo.ErrNoDocuments {
		log.Println("no documents found")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting request: %w", err)
	}

	return &request, nil
}

func (db *MongoDB) DeleteRequest(id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	collection := db.Database.Collection("requests")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	_, err = collection.DeleteOne(ctx, bson.D{{"_id", objectId}})
	if err != nil {
		return fmt.Errorf("error deleting request: %w", err)
	}

	return nil
}

func (db *MongoDB) UpdateRequestState(id string, state models.RequestStateType) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	collection := db.Database.Collection("requests")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	_, err = collection.UpdateOne(ctx, bson.D{{"_id", objectId}}, bson.D{{"$set", bson.D{{"state", state}}}})
	if err != nil {
		return fmt.Errorf("error updating request: %w", err)
	}

	return nil
}

func (db *MongoDB) GetRequestByServiceName(serviceName string) ([]models.Request, error) {
	var requests []models.Request
	if serviceName == "" {
		return nil, errors.New("serviceName name is not set")
	}
	filter := bson.D{{
		Key:   "service.name",
		Value: serviceName,
	}}
	collection := db.Database.Collection("requests")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error getting requests: %w", err)
	}
	defer cur.Close(ctx)

	if err = cur.All(context.TODO(), &requests); err != nil {
		return nil, fmt.Errorf("error fetching requests: %w", err)
	}
	return requests, nil
}
