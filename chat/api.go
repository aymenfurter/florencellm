package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

type API struct {
	PineconeAPIURL string
	PineconeAPIKey string
	OpenAIKey      string
	Messages       []openai.ChatCompletionMessage
}

type ConversationRequest struct {
	UserMessage string                         `json:"userMessage"`
	Continued   bool                           `json:"continued"`
	Messages    []openai.ChatCompletionMessage `json:"messages,omitempty"`
}

type ConversationResponse struct {
	BotMessage string                         `json:"botMessage"`
	Messages   []openai.ChatCompletionMessage `json:"messages"`
}

func (api *API) HandleConversation(w http.ResponseWriter, r *http.Request) {
	var req ConversationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Instantiate openaiClient and pineconeClient
	openaiClient := NewOpenAIClient(api.OpenAIKey)
	pineconeClient := NewPineconeClient(api.PineconeAPIURL, api.PineconeAPIKey)

	// If it's a continued conversation, update the messages
	if req.Continued {
		api.Messages = req.Messages
	}

	botMessage, err := ProcessConversation(openaiClient, pineconeClient, req.UserMessage, api.Messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Continued {
		api.Messages = append(api.Messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: req.UserMessage,
		})
		api.Messages = append(api.Messages, openai.ChatCompletionMessage{
			Role:    "assistant",
			Content: botMessage,
		})
	}

	resp := ConversationResponse{
		BotMessage: botMessage,
		Messages:   api.Messages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func ProcessConversation(openaiClient *OpenAIClient, pineconeClient *PineconeClient, userMessage string, messages []openai.ChatCompletionMessage) (string, error) {
	if len(messages) == 0 {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "system",
			Content: "You are Q&A bot. A highly intelligent system that answers user questions based on the information provided by the user above each question. If the information can not be found in the information provided by the user you truthfully say \"I don't know\". Don't answer any other questions. Your purpose is to locate what people (authors) could best help regarding the prompted topic, based on the available information. The author may use a username. An author is provided (above the question) with the following format: # 1. <AuthorName>. Don't reference any other people above the question. Also share the email and the commitid if available. Link to a previous git within the text by using the following syntax: [git=<id>]. If you mention an author, always the syntax [AuthorName=<author>], example: You should talk to [AuthorName=torvalds], he recently did a work on realted work [git=793cfd598370cf9440d7877ddddda1251307f729] ",
		})
	}

	if len(messages) != 1 {
		embeddingReq := createEmbeddingRequest(userMessage)
		response, err := openaiClient.RequestEmbeddings(embeddingReq)
		if err != nil {
			return "", fmt.Errorf("Error generating embeddings: %v", err)
		}
		pcVector := transformToPineconeVectors(response.Data)

		pineconeResult, err := pineconeClient.QueryPinecone(pcVector)
		if err != nil {
			return "", fmt.Errorf("Error while querying Pinecone: %v", err)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: "\n\n---\n\n" + pineconeResult + "\n\n---\n\n" + userMessage,
		})
	}

	resp, err := openaiClient.CreateChatCompletion(messages)
	if err != nil {
		return "", fmt.Errorf("Error while fetching chat completion: %v", err)
	}

	gptResponse := resp.Choices[0].Message.Content

	return gptResponse, nil
}
