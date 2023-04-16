package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
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
		"top_k":           15,
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
	authorRE := regexp.MustCompile(`Author:\s(.+)`)

	// TODO - Index all blocked users and expose this as a config
	blockedUsers := []string{
		"huypub",
		"prmerger-automator",
	}

	for i, match := range matches {
		match := match.(map[string]interface{})
		metadata := match["metadata"].(map[string]interface{})
		text := metadata["text"].(string)

		validUser := true
		author := authorRE.FindStringSubmatch(text)[1]
		if author != "" {
			for _, blockedUser := range blockedUsers {
				if author == blockedUser {
					fmt.Printf("Blocked user: %s", author)
					validUser = false
				}
			}
		}

		fmt.Printf("Allowed user: %s", author)

		//if author != "" {
		//	url := "https://github.com/" + author

		//	resp, err := http.Get(url)
		//	if err != nil {
		//		log.Fatalf("Error making request: %v", err)
		//	}
		//	defer resp.Body.Close()

		//	if resp.StatusCode == http.StatusOK {
		//		doc, err := goquery.NewDocumentFromReader(resp.Body)
		//		if err != nil {
		//			log.Fatalf("Error parsing HTML: %v", err)
		//		}

		//		userProfileBio := doc.Find(".user-profile-bio").Text()
		//		for _, word := range blocklist {
		//			if strings.Contains(userProfileBio, word) {
		//				fmt.Println("Blocked" + author)
		//				continue
		//			}
		//		}
		//	} else {
		//		fmt.Println("Not checked " + author)
		//	}

		//}

		output := fmt.Sprintf("\n\n# %d. %s\n", i+1, text)
		if len(output) > 500 {
			output = output[:500]
		}

		if validUser {
			matchOutput += output + "\n\n"
		}

		if len(matchOutput) > 4500 {
			break
		}

	}

	return matchOutput, nil
}
