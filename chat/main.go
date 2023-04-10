package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

func main() {
	pineconeAPIURL := os.Getenv("PINECONE_API_URL")
	pineconeAPIURL = fmt.Sprintf("%s/query", pineconeAPIURL)

	apiKey := os.Getenv("PINECONE_API_KEY")
	openaiAPIKey := os.Getenv("OPEN_AI_KEY")
	client := openai.NewClient(openaiAPIKey)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "You are Q&A bot. A highly intelligent system that answers user questions based on the information provided by the user above each question. If the information can not be found in the information provided by the user you truthfully say \"I don't know\". Don't answer any other questions. Your purpose is to locate what people (authors) could best help regarding the prompted topic, based on the available information. The author may use a username. An author is always identified (above the question) with the following format: # 1. <AuthorName>. Don't reference any other people that are not referenced in this format (above the question). You may also share the email. Link to a previous git within the text by using the following syntax: [git=<id>]. If you mention an author, always the syntax [user=<author>]",
		},
	}

	// Get user input
	var userMessage string
	fmt.Print("User: ")
	// bugfix: scanln only captures first word by using reader bufio
	reader := bufio.NewReader(os.Stdin)
	userMessage, _ = reader.ReadString('\n')

	// Turn user input into a vector representation
	embeddingReq := createEmbeddingRequest(userMessage)
	// print output
	fmt.Println("Embedding request: ", embeddingReq)
	response, err := requestEmbeddings(client, embeddingReq)
	pcVector := transformToPineconeVectors(response.Data)

	pineconeResult, err := queryPinecone(pineconeAPIURL, apiKey, pcVector)
	if err != nil {
		fmt.Printf("Error while querying Pinecone: %v\n", err)
		return
	}

	fmt.Println("User message: ", userMessage)

	// Append user input to messages
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: "\n\n---\n\n" + pineconeResult + "\n\n---\n\n" + userMessage,
	})

	// print out the messages
	for _, message := range messages {
		fmt.Printf("%s: %s\n", message.Role, message.Content)
	}

	// Fetch GPT-3.5 chat completion
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)
	if err != nil {
		fmt.Printf("Error while fetching chat completion: %v\n", err)
		return
	}

	// Extract GPT-3.5 response
	gptResponse := resp.Choices[0].Message.Content

	fmt.Printf("Bot: %s\n", gptResponse)
}

func createEmbeddingRequest(input string) openai.EmbeddingRequest {
	return openai.EmbeddingRequest{
		Input: []string{input},
		Model: openai.AdaEmbeddingV2,
	}
}

func queryPinecone(apiURL string, apiKey string, query []float32) (string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"vector":          query,
		"top_k":           3,
		"includeMetadata": true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to query Pinecone, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Pinecone response: %w", err)
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	matches := result["matches"].([]interface{})
	matchOutput := ""

	for i, match := range matches {
		match := match.(map[string]interface{})
		metadata := match["metadata"].(map[string]interface{})
		text := metadata["text"].(string)
		output := fmt.Sprintf("\n\n# %d. %s\n", i+1, text)
		if len(output) > 500 {
			output = output[:500]
		}

		matchOutput += output + "\n\n"

		if len(matchOutput) > 2500 {
			break
		}

	}

	return matchOutput, nil
}

func transformToPineconeVectors(embeddings []openai.Embedding) []float32 {
	return embeddings[0].Embedding
	/*
		vectors := make([]*pinecone_grpc.Vector, len(embeddings))
		for i, embedding := range embeddings {
			vector := &pinecone_grpc.Vector{
				Values: embedding.Embedding,
			}
			vectors[i] = vector
		}
		return vectors*/
}

func requestEmbeddings(client *openai.Client, embeddingReq openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	response, err := client.CreateEmbeddings(context.Background(), embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings from OpenAI: %w", err)
	}

	time.Sleep(1 * time.Second) // Avoid rate limit
	return &response, nil
}