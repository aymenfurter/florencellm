package main

import (
	"context"
	"github.com/Azure/azure-service-bus-go"
	"log"
	"os"
)

var (
	serviceBusConnString string
	queueName            string
	topicClient          *servicebus.Topic
)

func initServiceBus() {
	var err error
	serviceBusConnString = os.Getenv("AZURE_SERVICE_BUS_CONNECTION_STRING")
	queueName = os.Getenv("QUEUE_NAME")

	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(serviceBusConnString))
	if err != nil {
		log.Fatal(err)
	}

	topicClient, err = ns.NewTopic(queueName)
	if err != nil {
		log.Fatal(err)
	}
}

func sendMessageToServiceBus(repoID string) error {
	ctx := context.Background()
	msg := servicebus.NewMessageFromString(repoID)
	return topicClient.Send(ctx, msg)
}

