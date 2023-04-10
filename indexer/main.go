package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	servicebus "github.com/Azure/azure-service-bus-go"
)

func main() {
	serviceBusConnectionString := os.Getenv("AZURE_SERVICE_BUS_CONNECTION_STRING")
	queueName := os.Getenv("QUEUE_NAME")

	mongoDBConnectionString := os.Getenv("COSMOS_DB_CONNECTION_STRING")
	clientOptions := options.Client().ApplyURI(mongoDBConnectionString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	repoCol := client.Database("repositoryDB").Collection("repositories")

	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(serviceBusConnectionString))
	if err != nil {
		panic(err)
	}

	queue, err := ns.NewQueue(queueName)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	handler := NewMessageHandler(repoCol)
	err = queue.Receive(ctx, handler)

	defer cancel()

	if err != nil {
		fmt.Println("Error processing messages:", err)
	}
}
