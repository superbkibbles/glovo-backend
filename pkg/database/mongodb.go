package database

import (
	"context"
	"time"

	"github.com/mendmzury/food-delivery/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB holds the MongoDB client and database
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	logger.Info().Str("database", dbName).Msg("Connected to MongoDB")

	return &MongoDB{
		Client:   client,
		Database: client.Database(dbName),
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	logger.Info().Msg("Closing MongoDB connection")
	return m.Client.Disconnect(ctx)
}

// Collection returns a collection from the database
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// Collections for food delivery
const (
	CollectionNotifications      = "notifications"
	CollectionDeviceTokens      = "device_tokens"
	CollectionFiles             = "files"
	CollectionSharedDriveFolders = "shared_drive_folders"
	CollectionSettings           = "settings"
	CollectionOperatingAreas     = "operating_areas"
)

// CreateIndexes creates necessary indexes for food delivery collections
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// Notifications indexes
	notifIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "read", Value: 1}}},
	}
	if _, err := m.Collection(CollectionNotifications).Indexes().CreateMany(ctx, notifIndexes); err != nil {
		return err
	}

	// Device tokens indexes
	deviceIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "token", Value: 1}}},
	}
	if _, err := m.Collection(CollectionDeviceTokens).Indexes().CreateMany(ctx, deviceIndexes); err != nil {
		return err
	}

	// Settings indexes
	settingsIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "key", Value: 1}}},
	}
	if _, err := m.Collection(CollectionSettings).Indexes().CreateMany(ctx, settingsIndexes); err != nil {
		return err
	}

	// Operating areas indexes
	areaIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "active", Value: 1}}},
	}
	if _, err := m.Collection(CollectionOperatingAreas).Indexes().CreateMany(ctx, areaIndexes); err != nil {
		return err
	}

	logger.Info().Msg("Created MongoDB indexes")
	return nil
}
