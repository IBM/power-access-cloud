package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/db"
)

var (
	_     db.DB = &MongoDB{}
	PacDB       = "pac"
)

const dbContextTimeout = 10 * time.Second

type MongoDB struct {
	URI      string
	client   *mongo.Client
	Database *mongo.Database
}

func New() db.DB {
	uri, found := os.LookupEnv("MONGODB_URI")
	if !found {
		log.Fatal("MONGODB_URI environment variable not set")
	}

	return &MongoDB{
		URI: uri,
	}
}

func (db *MongoDB) Connect() error {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(db.URI).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return fmt.Errorf("error connecting to MongoDB: %w", err)
	}
	db.client = client

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("error pinging MongoDB: %w", err)
	}
	log.Println("Connected to MongoDB")
	db.Database = client.Database(PacDB)

	return nil
}

func (db *MongoDB) Disconnect() error {
	return db.client.Disconnect(context.Background())
}

func (db *MongoDB) CollectionExists(name string) (bool, error) {
	colNames, err := db.Database.ListCollectionNames(context.Background(), bson.M{"name": name})
	if err != nil {
		return false, fmt.Errorf("failed to list collections from DB: %w", err)
	}

	if len(colNames) == 0 {
		// collection does not exist
		return false, nil
	}

	// collection exists
	return true, nil
}
