package main

import (
	"context"
	"fmt"

	servicebus "github.com/Azure/azure-service-bus-go"
	"github.com/pinecone-io/go-pinecone/pinecone_grpc"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageHandler struct {
	servicebus.Handler
	repoCol  *mongo.Collection
	pcClient pinecone_grpc.VectorServiceClient
}

func NewMessageHandler(repoCol *mongo.Collection, pcClient pinecone_grpc.VectorServiceClient) *MessageHandler {
	return &MessageHandler{
		repoCol:  repoCol,
		pcClient: pcClient,
	}
}

func (h *MessageHandler) Handle(ctx context.Context, msg *servicebus.Message) error {
	repoID := string(msg.Data)

	err := indexRepository(ctx, repoID, h.repoCol, h.pcClient)
	if err != nil {
		fmt.Println("Error indexing repository:", err)
	} else {
		fmt.Printf("Repository %s indexed successfully\n", repoID)
	}

	return msg.Complete(ctx)
}
