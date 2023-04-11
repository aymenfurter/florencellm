package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type PineconeClient struct {
	APIURL string
	APIKey string
}

func NewPineconeClient(apiURL, apiKey string) *PineconeClient {
	return &PineconeClient{
		APIURL: apiURL,
		APIKey: apiKey,
	}
}

func (client *PineconeClient) QueryPinecone(query []float32) (string, error) {
	apiURL := fmt.Sprintf("%s/query", client.APIURL)

	requestBody, err := json.Marshal(map[string]interface{}{
		"vector":          query,
		"top_k":           10,
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
	req.Header.Set("Api-Key", client.APIKey)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
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
