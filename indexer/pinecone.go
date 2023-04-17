package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pinecone-io/go-pinecone/pinecone_grpc"
	"github.com/sashabaranov/go-openai"
)

func storeEmbeddings(commitId string, embeddings []*pinecone_grpc.Vector) error {
	pineconeAPIURL := os.Getenv("PINECONE_API_URL")
	pineconeAPIURL = fmt.Sprintf("%s/vectors/upsert", pineconeAPIURL)
	apiKey := os.Getenv("PINECONE_API_KEY")

	requestBody, err := json.Marshal(map[string]interface{}{
		"vectors": embeddings,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", pineconeAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to upsert embeddings to Pinecone, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func transformToPineconeVectors(commitId string, embeddings []openai.Embedding) []*pinecone_grpc.Vector {
	vectors := make([]*pinecone_grpc.Vector, len(embeddings))
	for i, embedding := range embeddings {
		embeddingsId := commitId
		if i != 0 {
			embeddingsId = fmt.Sprintf("%s-%d", commitId, i)
		}

		vector := &pinecone_grpc.Vector{
			Id:     embeddingsId,
			Values: embedding.Embedding,
		}
		vectors[i] = vector
	}
	return vectors
}
