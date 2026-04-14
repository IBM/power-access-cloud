package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetAllMaintenanceWindows retrieves all non-deleted maintenance windows from the database
func (db *MongoDB) GetAllMaintenanceWindows() ([]models.MaintenanceWindow, error) {
	collection := db.Database.Collection("maintenance_windows")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	// Filter out soft-deleted records
	filter := bson.M{"deleted_at": bson.M{"$exists": false}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error fetching maintenance windows from DB: %w", err)
	}
	defer cursor.Close(ctx)

	var windows []models.MaintenanceWindow
	if err := cursor.All(ctx, &windows); err != nil {
		return nil, fmt.Errorf("error decoding maintenance windows: %w", err)
	}

	return windows, nil
}

// GetMaintenanceWindowByID retrieves a specific non-deleted maintenance window by ID
func (db *MongoDB) GetMaintenanceWindowByID(id string) (*models.MaintenanceWindow, error) {
	collection := db.Database.Collection("maintenance_windows")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid maintenance window ID: %w", err)
	}

	var window models.MaintenanceWindow
	// Filter out soft-deleted records
	filter := bson.M{
		"_id":        objectID,
		"deleted_at": bson.M{"$exists": false},
	}

	err = collection.FindOne(ctx, filter).Decode(&window)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching maintenance window from DB: %w", err)
	}

	return &window, nil
}

// CreateMaintenanceWindow creates a new maintenance window
func (db *MongoDB) CreateMaintenanceWindow(window *models.MaintenanceWindow) error {
	collection := db.Database.Collection("maintenance_windows")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	_, err := collection.InsertOne(ctx, window)
	if err != nil {
		return fmt.Errorf("error creating maintenance window: %w", err)
	}

	return nil
}

// UpdateMaintenanceWindow updates an existing maintenance window
func (db *MongoDB) UpdateMaintenanceWindow(window *models.MaintenanceWindow) error {
	collection := db.Database.Collection("maintenance_windows")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	filter := bson.M{"_id": window.ID}
	update := bson.M{
		"$set": bson.M{
			"enabled":    window.Enabled,
			"start_time": window.StartTime,
			"end_time":   window.EndTime,
			"message":    window.Message,
			"updated_by": window.UpdatedBy,
			"updated_at": window.UpdatedAt,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating maintenance window: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("maintenance window not found")
	}

	return nil
}

// DeleteMaintenanceWindow performs soft delete on a maintenance window by ID
func (db *MongoDB) DeleteMaintenanceWindow(id string, deletedBy string, deletedAt *time.Time) error {
	collection := db.Database.Collection("maintenance_windows")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid maintenance window ID: %w", err)
	}

	// Only soft delete non-deleted records
	filter := bson.M{
		"_id":        objectID,
		"deleted_at": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_by": deletedBy,
			"deleted_at": deletedAt,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error soft deleting maintenance window: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("maintenance window not found or already deleted")
	}

	return nil
}
