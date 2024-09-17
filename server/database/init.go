package database

import (
	"context"
	"fmt"
	"log"
	"server/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitDb initializes the database connection to MongoDB.
// This will proactively panic if any step of the connection protocol breaks
func InitDb() *mongo.Database {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(config.GetEnv("MONGODB_URI")).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to server
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err.Error())
	}

	// Send ping to confirm successful connections
	var result bson.M
	if err := client.Database("admin").RunCommand(context.Background(), bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		log.Fatalf("Error pinging database: %s\n", err.Error())
	}
	fmt.Println("Successfully connected to database!")

	// Return the "jury" database
	db := client.Database("jury")

	// Create indexes for token_set and judges tables/collections
	tokenSetIndexModel := mongo.IndexModel{Keys: bson.D{{"user_id", true}}}
	db.Collection("token_set").Indexes().CreateOne(context.Background(), tokenSetIndexModel)

	judgesIndexModel := mongo.IndexModel{Keys: bson.D{{"keycloak_user_id", true}}}
	db.Collection("judges").Indexes().CreateOne(context.Background(), judgesIndexModel)

	return db
}
