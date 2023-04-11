package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	// Initialize test router
	router := SetupRouter()

	// Test initial mode
	t.Run("Initial mode", func(t *testing.T) {
		input := `{"userMessage": "What is AI?"}`
		request := httptest.NewRequest(http.MethodPost, "/conversation", strings.NewReader(input))
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)

		var responseBody map[string]interface{}
		err := json.Unmarshal(response.Body.Bytes(), &responseBody)
		assert.Nil(t, err)

		assert.Contains(t, responseBody, "response")
		assert.Contains(t, responseBody, "messages")
	})

	// Test continue mode
	t.Run("Continue mode", func(t *testing.T) {
		input := `{"userMessage": "Can you explain more?", "messages": [{"role": "system", "content": "You are Q&A bot. "}, {"role": "user", "content": "What is AI?"}, {"role": "assistant", "content": "AI is artificial intelligence..."}]}`
		request := httptest.NewRequest(http.MethodPost, "/conversation", strings.NewReader(input))
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)

		var responseBody map[string]interface{}
		err := json.Unmarshal(response.Body.Bytes(), &responseBody)
		assert.Nil(t, err)

		assert.Contains(t, responseBody, "response")
		assert.Contains(t, responseBody, "messages")
		assert.Len(t, responseBody["messages"].([]interface{}), 4)
	})
}
