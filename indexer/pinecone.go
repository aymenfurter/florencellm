package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pinecone-io/go-pinecone/pinecone_grpc"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func buildPineconeClient() pinecone_grpc.VectorServiceClient {
	apiKey := os.Getenv("PINECONE_API_KEY")
	indexName := os.Getenv("PINECONE_INDEX_NAME")
	projectName := os.Getenv("PINECONE_PROJECT_NAME")
	pineconeEnv := os.Getenv("PINECONE_ENV")

	certPool, err := x509.SystemCertPool()
	creds := credentials.NewClientTLSFromCert(certPool, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "api-key", apiKey)
	target := fmt.Sprintf("%s-%s.svc.%s.pinecone.io:443", indexName, projectName, pineconeEnv)
	log.Printf("connecting to %v", target)
	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(creds),
		grpc.WithAuthority(target),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := pinecone_grpc.NewVectorServiceClient(conn)

	return client
}

func storeEmbeddings(commitId string, embeddings []openai.Embedding, client pinecone_grpc.VectorServiceClient) error {
	namespace := os.Getenv("PINECONE_NAMESPACE")

	vector := transformToPineconeVectors(commitId, embeddings)

	_, upsertErr := client.Upsert(context.Background(), &pinecone_grpc.UpsertRequest{
		Vectors:   vector,
		Namespace: namespace,
	})
	if upsertErr != nil {
		return fmt.Errorf("failed to upsert embeddings to Pinecone: %w", upsertErr)
	}

	return nil
}

func transformToPineconeVectors(commitId string, embeddings []openai.Embedding) []*pinecone_grpc.Vector {
	vectors := make([]*pinecone_grpc.Vector, len(embeddings))
	for i, embedding := range embeddings {
		vector := &pinecone_grpc.Vector{
			Id:     commitId,
			Values: embedding.Embedding,
		}
		vectors[i] = vector
	}
	return vectors
}
