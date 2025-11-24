package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
)

func (db *MongoDB) NewEvent(e *models.Event) error {
	collection := db.Database.Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	_, err := collection.InsertOne(ctx, e)
	if err != nil {
		return fmt.Errorf("error inserting Event: %w", err)
	}

	return nil
}

// GetEventsByUserID returns all events by user id
func (db *MongoDB) GetEventsByUserID(id string, startIndex, perPage int64) ([]models.Event, int64, error) {
	events := []models.Event{}
	var totalCount int64

	filter := bson.D{{}}
	if id != "" {
		filter = bson.D{{Key: "user_id", Value: id}}
	}

	findOptions := options.Find()
	findOptions.SetSkip(startIndex)
	findOptions.SetLimit(perPage)

	collection := db.Database.Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	cur, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, totalCount, fmt.Errorf("error getting events: %w", err)
	}
	defer cur.Close(ctx)

	if err = cur.All(context.Background(), &events); err != nil {
		return nil, totalCount, fmt.Errorf("error fetching events: %w", err)
	}
	// Get the total number of events from the database
	totalCount, err = collection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return nil, totalCount, fmt.Errorf("error getting total count of events: %w", err)
	}

	return events, totalCount, nil
}

// GetEventsByType returns all events of specified type within specified duration
func (db *MongoDB) GetEventsByType(eventType models.EventType, duration uint) ([]models.Event, int64, error) {
	events := []models.Event{}
	var totalCount int64

	startTime := time.Now().Add(-time.Duration(duration) * time.Hour)

	filter := bson.D{{}}
	if eventType != "" {
		filter = bson.D{
			{Key: "type", Value: eventType},
			{Key: "created_at", Value: bson.D{{Key: "$gt", Value: startTime}}},
		}
	}
	findOptions := options.Find()
	collection := db.Database.Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	cur, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, totalCount, fmt.Errorf("error getting events: %w", err)
	}
	defer cur.Close(ctx)

	if err = cur.All(context.Background(), &events); err != nil {
		return nil, totalCount, fmt.Errorf("error fetching events: %w", err)
	}
	return events, totalCount, nil
}

// WatchEvents watches the events collection for changes
func (db *MongoDB) WatchEvents(eventCh chan<- *models.Event) error {
	collection := db.Database.Collection("events")
	cursorOptions := options.Find().SetCursorType(options.TailableAwait).SetMaxAwaitTime(2 * time.Second)

	for {
		// Create a new cursor for each iteration to handle possible cursor timeouts
		cursor, err := collection.Find(context.Background(), bson.D{}, cursorOptions)
		if err != nil {
			return fmt.Errorf("error creating cursor: %w", err)
		}

		// Process the change events
		for cursor.Next(context.Background()) {
			var event models.Event
			if err := cursor.Decode(&event); err != nil {
				log.Println("Error decoding event:", err)
				continue
			}

			eventCh <- &event
		}

		if err := cursor.Err(); err != nil {
			return fmt.Errorf("error reading cursor: %w", err)
		}

		// Close the cursor
		if err := cursor.Close(context.Background()); err != nil {
			return fmt.Errorf("error closing cursor: %w", err)
		}

		// Sleep for a while before checking for new changes
		time.Sleep(5 * time.Second)
	}
}

// MarkEventAsNotified marks the event as notified
func (db *MongoDB) MarkEventAsNotified(id string) error {
	event := models.Event{}
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	filter := bson.M{"_id": objectId}

	collection := db.Database.Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	if err := collection.FindOne(ctx, filter).Decode(&event); err != nil && err != mongo.ErrNoDocuments {
		return fmt.Errorf("error getting event: %w", err)
	} else if err == mongo.ErrNoDocuments {
		return fmt.Errorf("event not found with id: %s, err : %w", id, err)
	}

	if _, err := collection.UpdateOne(ctx, bson.M{"_id": objectId},
		bson.D{{Key: "$set", Value: bson.D{{Key: "notified", Value: true}}}}); err != nil {
		return fmt.Errorf("error while updating event: %w", err)
	}

	return nil
}

// SetEventCapping caps event to given value
func (db *MongoDB) SetEventCapping(cap int64) error {
	opts := options.CreateCollection()
	opts.SetCapped(true)
	opts.SetSizeInBytes(cap)
	log.Println("Invoked SetEventCapping...")

	ctx, cancel := context.WithTimeout(context.Background(), dbContextTimeout)
	defer cancel()
	return db.Database.CreateCollection(ctx, "events", opts)
}
