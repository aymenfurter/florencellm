package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type Repository struct {
	ID             string `json:"id,omitempty" bson:"_id,omitempty"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	IndexingStatus string `json:"indexing_status"`
}

func repositoryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		listRepositories(w, r)
	case "PUT":
		addRepository(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func listRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	// get all documents
	cursor, err := repoCol.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Error fetching repositories", http.StatusInternalServerError)
		log.Printf("Error fetching repositories: %v", err)
		return
	}

	var repositories []Repository
	if err := cursor.All(ctx, &repositories); err != nil {
		http.Error(w, "Error decoding repositories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repositories)
}

func generateSHAHashFromURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

func addRepository(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var repo Repository
	err = json.Unmarshal(body, &repo)
	repo.ID = generateSHAHashFromURL(repo.URL)
	if err != nil {
		http.Error(w, "Error decoding request body ", http.StatusBadRequest)
		return
	}

	repo.IndexingStatus = "pending"

	ctx := context.Background()
	insertResult, err := repoCol.InsertOne(ctx, repo)
	if err != nil {
		if strings.Contains(err.Error(), "unique index constraint") {
			http.Error(w, "Repository already exists", http.StatusConflict)
			return
		}

		http.Error(w, "Error inserting repository", http.StatusInternalServerError)
		return
	}

	repo.ID = insertResult.InsertedID.(string)
	if err := sendMessageToServiceBus(repo.ID); err != nil {
		http.Error(w, "Error sending message to service bus", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(repo)
}
