package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	client *mongo.Client
}

func NewDatabase(uri string) *Database {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return &Database{client}
}

func (db *Database) StoreLinks(url string, links []string) {
	collection := db.client.Database("crawler").Collection("urls")

	for _, link := range links {
		_, err := collection.InsertOne(context.Background(), bson.M{"url": url, "link": link})
		if err != nil {
			log.Printf("Error storing link for %s: %v", url, err)
		}
	}
}
