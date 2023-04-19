package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

// add main method
func main() {
	router := SetupRouter()
	router.Run()
}

// SetupRouter configures the router for API endpoints
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Initialize OpenAIClient and PineconeClient
	openaiClient := NewOpenAIClient(os.Getenv("OPEN_AI_KEY"))
	pineconeClient := NewPineconeClient(os.Getenv("PINECONE_API_URL"), os.Getenv("PINECONE_API_KEY"))

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))

	router.POST("/api/conversation", func(c *gin.Context) {
		var requestBody ConversationRequest
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := ProcessConversation(openaiClient, pineconeClient, requestBody.UserMessage, requestBody.Messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		messages := requestBody.Messages
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: requestBody.UserMessage,
		}, openai.ChatCompletionMessage{
			Role:    "assistant",
			Content: response,
		})

		c.JSON(http.StatusOK, gin.H{
			"response": response,
			"messages": messages,
		})
	})

	return router
}
