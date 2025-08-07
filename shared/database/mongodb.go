package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB() *mongo.Database {
	host := getEnv("MONGO_HOST", "localhost")
	port := getEnv("MONGO_PORT", "27017")
	user := getEnv("MONGO_USER", "")
	password := getEnv("MONGO_PASSWORD", "")
	dbname := getEnv("MONGO_DB", "location_service_db")

	// Build connection URI
	var uri string
	if user != "" && password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s", user, password, host, port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s", host, port)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	log.Printf("Connected to MongoDB at %s:%s", host, port)
	return client.Database(dbname)
}
