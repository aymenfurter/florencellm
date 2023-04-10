package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sashabaranov/go-openai"
)

func generateEmbeddings(author object.Signature, email string, diffString string) ([]openai.Embedding, error) {
	client := openai.NewClient(os.Getenv("OPEN_AI_KEY"))

	input := fmt.Sprintf("Author: %s\nEmail: %s\nDiff: %s", author.Name, email, diffString)

	embeddingReq := openai.EmbeddingRequest{
		Input: []string{input},
		Model: openai.AdaEmbeddingV2,
	}
	response, err := client.CreateEmbeddings(context.Background(), embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings from OpenAI: %w", err)
	}

	embeddings := response.Data

	return embeddings, nil
}
