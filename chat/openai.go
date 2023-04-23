package main

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	Client *openai.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		Client: openai.NewClient(apiKey),
	}
}

func (client *OpenAIClient) RequestEmbeddings(embeddingReq openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	response, err := client.Client.CreateEmbeddings(context.Background(), embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings from OpenAI: %w", err)
	}

	//time.Sleep(1 * time.Second) // Avoid rate limit
	return &response, nil
}

func (c *OpenAIClient) CreateChatCompletion(messages []openai.ChatCompletionMessage) (*openai.ChatCompletionResponse, error) {
	resp, err := c.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			Messages:    messages,
			MaxTokens:   1024,
			Temperature: 0.3,
			N:           1,
		},
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func transformToPineconeVectors(embeddings []openai.Embedding) []float32 {
	return embeddings[0].Embedding
}

func createEmbeddingRequest(input string) openai.EmbeddingRequest {
	return openai.EmbeddingRequest{
		Input: []string{input},
		Model: openai.AdaEmbeddingV2,
	}
}
