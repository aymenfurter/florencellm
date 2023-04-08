package main

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client  *mongo.Client
	repoDB  *mongo.Database
	repoCol *mongo.Collection
)

func initDatabase() {
	var err error
	connString := os.Getenv("COSMOS_DB_CONNECTION_STRING")
	client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(connString))
	if err != nil {
		log.Fatal(err)
	}
	repoDB = client.Database("repositoryDB")
	repoCol = repoDB.Collection("repositories")
}
