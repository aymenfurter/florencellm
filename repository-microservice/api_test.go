package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRepositories(t *testing.T) {
	// Set up a request and response recorder
	req, _ := http.NewRequest("GET", "/api/repository", nil)
	rr := httptest.NewRecorder()

	// Call the repositoryHandler with the request and response recorder
	repositoryHandler(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code, "Status code should be 200")

	// Check if the response is a JSON array
	var repositories []Repository
	err := json.Unmarshal(rr.Body.Bytes(), &repositories)
	assert.NoError(t, err, "Response should be a valid JSON array")
}

func TestAddRepository(t *testing.T) {
	// Create a new repository object
	repo := Repository{
		Name:           "Test Repo",
		URL:            "https://github.com/example/test-repo.git",
		IndexingStatus: "pending",
	}

	// Marshal the repository object into JSON
	repoJSON, _ := json.Marshal(repo)

	// Set up a request and response recorder
	req, _ := http.NewRequest("PUT", "/api/repository", bytes.NewBuffer(repoJSON))
	rr := httptest.NewRecorder()

	// Call the repositoryHandler with the request and response recorder
	repositoryHandler(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusCreated, rr.Code, "Status code should be 201")

	// Check if the response is a JSON object
	var createdRepo Repository
	err := json.Unmarshal(rr.Body.Bytes(), &createdRepo)
	assert.NoError(t, err, "Response should be a valid JSON object")

	// Check if the createdRepo object matches the original repo object
	assert.Equal(t, repo.Name, createdRepo.Name, "Repository names should match")
	assert.Equal(t, repo.URL, createdRepo.URL, "Repository URLs should match")
	assert.Equal(t, repo.IndexingStatus, createdRepo.IndexingStatus, "Repository indexing status should match")
}

