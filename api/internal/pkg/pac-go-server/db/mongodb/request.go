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

func (db *MongoDB) GetRequestsByUserID(id, requestType string) ([]models.Request, error) {
	var requests []models.Request
	filter := bson.D{}
	if requestType != "" {
		filter = append(filter, bson.E{Key: "type", Value: requestType})
	}
	if id != "" {
		filter = append(filter, bson.E{Key: "user_id", Value: id})
	}

	collection := db.Database.Collection("requests")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
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

func (db *MongoDB) NewRequest(request *models.Request) (string, error) {
	collection := db.Database.Collection("requests")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	res, err := collection.InsertOne(ctx, request)
	if err != nil {
		return "", fmt.Errorf("error inserting request: %w", err)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("unable to get inserted id for request")
	}

	return oid.Hex(), nil
}

func (db *MongoDB) GetRequestByGroupIDAndUserID(groupID string, userID string) ([]models.Request, error) {
	var requests []models.Request

	filter := bson.D{
		{Key: "$and",
			Value: bson.A{
				bson.D{{Key: "group.group_id", Value: groupID}},
				bson.D{{Key: "user_id", Value: userID}},
			}},
	}

	collection := db.Database.Collection("requests")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	err = collection.FindOne(ctx, filter).Decode(&request)
	if err == mongo.ErrNoDocuments {
		log.Println("no documents found")
		return nil, err
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
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	_, err = collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: objectId}})
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
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	_, err = collection.UpdateOne(ctx, bson.D{{Key: "_id", Value: objectId}}, bson.D{{Key: "$set", Value: bson.D{{Key: "state", Value: state}}}})
	if err != nil {
		return fmt.Errorf("error updating request: %w", err)
	}

	return nil
}

func (db *MongoDB) UpdateRequestStateWithComment(id string, state models.RequestStateType, comment string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	collection := db.Database.Collection("requests")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	_, err = collection.UpdateOne(ctx, bson.D{{Key: "_id", Value: objectId}}, bson.D{{Key: "$set", Value: bson.D{{Key: "state", Value: state}, {Key: "comment", Value: comment}}}})
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
	filter := bson.D{{Key: "service.name", Value: serviceName}}
	collection := db.Database.Collection("requests")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
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
