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

func ProcessConversation(openaiClient *OpenAIClient, pineconeClient *PineconeClient, userMessage string, messagesIn []openai.ChatCompletionMessage) (string, error) {
	messages := []openai.ChatCompletionMessage{}
	prePrompt := "Always start a sentence with 'I would recommend to'  You are Q&A bot. You must always elobrate / explain your memory in great details (in your own words!), you will find it above the question 🕵️. You are a highly intelligent system that locates people (authors) that could best help regarding a certain topic or question using your memory 🔎. Your personal memory is provided provided above each question. If the answer can not be found in the your personal memory you truthfully say \"I don't know\". Don't answer any other questions. The author may use a username. An author is provided (above the question) with the following format: # 1. <AuthorName>. Don't reference any other people or information that is not mentioned above the question. Always share the email address (if available) in this format: [foobar@example.com] (foobar@example.com). Please always link the to relevant commit (e.g. [https://github.com/aymenfurter/x/commit/64e49e60dc41ecd1d6c5a5aebdc5b66e2275c41f](https://github.com/aymenfurter/x/commit/64e49e60dc41ecd1d6c5a5aebdc5b66e2275c41f)). If you mention an author, always the syntax [user](user@example.com) \n Do you understand? "
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "system",
		Content: prePrompt,
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "assistant",
		Content: "Yes, I understand. I will start a message with 'I would recommend to' ✋",
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: "Author: Sam Alias Jones \nRepoURL: https://github.com/MicrosoftDocs/azure-docs.git \nCommit-Message: Uploaded file articles/virtual-machines-sharepoint-farm-azure-preview.md Email: sam.alias.jones@microsoft.com CommitId: 04b6e1dcac6376ddc7bb05f892f4d1b6028e224b Diff: diff --git a/articles/virtual-machines-sharepoint-farm-azure-preview.md b/articles/virtual-machines-sharepoint-farm-azure-preview.md index a/articles/virtual-machines-sharepoint-farm-azure-preview.md +++ b/articles/virtual-machines-sharepoint-farm-azure-preview.md @@ -1,6 +1,23 @@ +<properties + pageTitle=SharePoint Server Farm Describes the new SharePoint Server Farm feature available in the Azure Preview Portal servicesdocumentationCenter= With SharePoint Server Farm, the Microsoft Azure Preview Portal automatically creates a pre-configured SharePoint Server 2013\nAuthor: Jane Doe Smith \nRepoURL: https://github.com/MicrosoftDocs/azure-docs.git \nCommit-Message: Added high availability guide to articles/virtual-machines-sharepoint-farm-azure-preview.md Email: jane.doe.smith@microsoft.com CommitId: f7d42ec13b8465dac3b97c547293db7c62a9d91c Diff: diff --git a/articles/virtual-machines-sharepoint-farm-azure-preview.md b/articles/virtual-machines-sharepoint-farm-azure-preview.md index a/articles/virtual-machines-sharepoint-farm-azure-preview.md +++ b/articles/virtual-machines-sharepoint-farm-azure-preview.md @@ -2,12 +2,35 @@ +<properties + pageTitle=SharePoint Server Farm High Availability Guide Provides a step-by-step guide on configuring high availability for SharePoint Server Farm in Azure Preview Portal servicesdocumentationCenter= This guide focuses on achieving high availability and fault tolerance for your SharePoint Server Farm in the Azure Preview Portal\n\n	\n	Who can help me with with high availability and learning what server farm features are available with Sharepoint Server?",
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "assistant",
		Content: "Hey there! I recommend to take a look at this: [SharePoint Server Farm](https://github.com/MicrosoftDocs/azure-docs/commit/04b6e1dcac6376ddc7bb05f892f4d1b6028e224b) within Azure Documentation 👀. \nI'd recommend reaching out to Sam Alias Jones (sam.alias.jones@microsoft.com) as he's was contributing in this area and Jane Doe Smith (jane.doe.smith@microsoft.com) for her expertise in high availability and fault tolerance. \nI hope you find this information helpful! 			\n\n			Here are the  references 📙\n			1. [SharePoint Server Farm](https://github.com/MicrosoftDocs/azure-docs/commit/04b6e1dcac6376ddc7bb05f892f4d1b6028e224b) by Sam Alias Jones (sam.alias.jones@microsoft.com)			2. [SharePoint Server Farm High Availability Guide](https://github.com/MicrosoftDocs/azure-docs/commit/f7d42ec13b8465dac3b97c547293db7c62a9d91c) by Jane Doe Smith (jane.doe.smith@microsoft.com)			\n\n			Do you have any follow up questions? 🤗",
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: "OK - Next topic. \n\n\n",
	})

	embeddingMsgContext := ""
	for _, message := range messagesIn {
		if message.Role == "user" {
			embeddingMsgContext += message.Content
		}
		embeddingMsgContext += "\n\n---\n\n"
	}
	embeddingMsgContext += userMessage

	embeddingReq := createEmbeddingRequest(embeddingMsgContext)
	response, err := openaiClient.RequestEmbeddings(embeddingReq)
	if err != nil {
		return "", fmt.Errorf("Error generating embeddings: %v", err)
	}
	pcVector := transformToPineconeVectors(response.Data)

	pineconeResult, err := pineconeClient.QueryPinecone(pcVector)
	if err != nil {
		return "", fmt.Errorf("Error while querying Pinecone: %v", err)
	}

	for _, message := range messagesIn {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: "\n\n-----BEGIN YOUR PERSONAL BOT MEMORY -----n\n" + pineconeResult + "\n\n-----END YOUR PERSONAL BOT MEMORY -----n\n" + userMessage,
	})
	fmt.Println("Messages: ", messages)

	resp, err := openaiClient.CreateChatCompletion(messages)
	if err != nil {
		return "", fmt.Errorf("Error while fetching chat completion: %v", err)
	}

	gptResponse := resp.Choices[0].Message.Content

	return gptResponse, nil
}
