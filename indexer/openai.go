package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sashabaranov/go-openai"
)

const maxDiffStringLength = 8000
const chunkCutoffThreshold = 1

func generateEmbeddings(commitMsg string, author object.Signature, email, diffString string) ([]openai.Embedding, error) {
	client := newOpenAIClient()

	if len(diffString) < maxDiffStringLength {
		return generateSingleEmbedding(client, commitMsg, author, email, diffString)
	}

	return generateChunkedEmbeddings(client, commitMsg, author, email, diffString)
}

func newOpenAIClient() *openai.Client {
	return openai.NewClient(os.Getenv("OPEN_AI_KEY"))
}

func generateSingleEmbedding(client *openai.Client, commitMsg string, author object.Signature, email, diffString string) ([]openai.Embedding, error) {
	input := fmt.Sprintf("Author: %s\nCommit-Message:\n%s\nEmail: %s\nDiff: %s", author.Name, commitMsg, email, diffString)
	embeddingReq := createEmbeddingRequest(input)

	response, err := requestEmbeddings(client, embeddingReq)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func generateChunkedEmbeddings(client *openai.Client, commitMsg string, author object.Signature, email, diffString string) ([]openai.Embedding, error) {
	chunks := chunkString(diffString, maxDiffStringLength)
	embeddings := make([]openai.Embedding, 0)

	for i, chunk := range chunks {
		if i <= chunkCutoffThreshold {
			input := fmt.Sprintf("Author: %s\nCommit-Message:%s\nEmail: %s\nChunk:\n%d\nDiff: %s", author.Name, commitMsg, email, i, chunk)
			embeddingReq := createEmbeddingRequest(input)

			response, err := requestEmbeddings(client, embeddingReq)
			if err != nil {
				return nil, err
			}

			embeddings = append(embeddings, response.Data...)
		}
	}

	return embeddings, nil
}

func createEmbeddingRequest(input string) openai.EmbeddingRequest {
	return openai.EmbeddingRequest{
		Input: []string{input},
		Model: openai.AdaEmbeddingV2,
	}
}

func requestEmbeddings(client *openai.Client, embeddingReq openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	response, err := client.CreateEmbeddings(context.Background(), embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings from OpenAI: %w", err)
	}

	time.Sleep(1 * time.Second) // Avoid rate limit
	return &response, nil
}

func chunkString(str string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(str); i += chunkSize {
		end := i + chunkSize
		if end > len(str) {
			end = len(str)
		}
		chunks = append(chunks, str[i:end])
	}
	return chunks
}
