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
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return &Database{client}
}

func (db *Database) StoreLinks(url string, links []string) {
	collection := db.client.Database("crawler").Collection("urls")

	var operations []mongo.WriteModel
	for _, link := range links {
		op := mongo.NewInsertOneModel().SetDocument(bson.M{"url": url, "link": link})
		operations = append(operations, op)
	}

	_, err := collection.BulkWrite(context.Background(), operations)
	if err != nil {
		log.Printf("Bulk insert failed for %s: %v", url, err)
	}
}
