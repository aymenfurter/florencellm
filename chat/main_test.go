package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestChat(t *testing.T) {
	pineconeAPIURL := os.Getenv("PINECONE_API_URL")
	pineconeAPIKey := os.Getenv("PINECONE_API_KEY")
	openaiAPIKey := os.Getenv("OPEN_AI_KEY")

	openaiClient := NewOpenAIClient(openaiAPIKey)
	pineconeClient := NewPineconeClient(pineconeAPIURL, pineconeAPIKey)

	userMessage := "Who can help me with AKS?"
	messages := []openai.ChatCompletionMessage{}

	gptResponse, err := ProcessConversation(openaiClient, pineconeClient, userMessage, messages)
	if err != nil {
		fmt.Printf("Error processing conversation: %v\n", err)
		return
	}

	fmt.Printf("Bot: %s\n", gptResponse)
}
