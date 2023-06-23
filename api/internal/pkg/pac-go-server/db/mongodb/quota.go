package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

func (db *MongoDB) GetUserQuota(s string) (models.Quota, error) {
	return models.Quota{
		GroupID: "123",
		Capacity: models.Capacity{
			CPU:    5,
			Memory: 10,
		},
	}, nil
}

func (db *MongoDB) GetQuotaForGroupID(id string) (*models.Quota, error) {
	var quota models.Quota
	logger := log.GetLogger()

	filter := bson.M{"group_id": id}

	collection := db.Database.Collection("quota")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	err := collection.FindOne(ctx, filter).Decode(&quota)
	if err == mongo.ErrNoDocuments {
		logger.Info("No documents found for quota", zap.Error(err))
		return nil, mongo.ErrNoDocuments
	}
	if err != nil {
		return nil, fmt.Errorf("error getting request: %w", err)
	}

	return &quota, nil
}

func (db *MongoDB) NewQuota(quota *models.Quota) error {
	collection := db.Database.Collection("quota")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	_, err := collection.InsertOne(ctx, quota)
	if err != nil {
		return fmt.Errorf("error while adding an entry for quota: %w", err)
	}
	return nil
}

func (db *MongoDB) UpdateQuota(quota *models.Quota) error {
	collection := db.Database.Collection("quota")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	_, err := collection.UpdateOne(ctx, bson.M{"group_id": quota.GroupID}, bson.D{{"$set", bson.D{{Key: "capacity", Value: quota.Capacity}}}})
	if err != nil {
		return fmt.Errorf("error while adding an entry for quota: %w", err)
	}
	return nil
}

func (db *MongoDB) DeleteQuota(id string) error {
	collection := db.Database.Collection("quota")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	filter := bson.M{"group_id": id}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("error deleting quota: %w", err)
	}
	return nil
}

func (db *MongoDB) GetGroupsQuota(groups []string) ([]models.Quota, error) {
	var quota []models.Quota
	if len(groups) == 0 {
		return nil, fmt.Errorf("groups is empty")
	}
	filter := bson.M{"group_id": bson.M{"$in": groups}}

	collection := db.Database.Collection("quota")
	ctx, _ := context.WithTimeout(context.Background(), dbContextTimeout)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error getting requests: %w", err)
	}
	defer cur.Close(ctx)

	if err = cur.All(context.TODO(), &quota); err != nil {
		return nil, fmt.Errorf("error fetching quota: %w", err)
	}
	return quota, nil
}
