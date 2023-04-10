package main

import (
	"context"
	fmt "fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	ID             string `json:"id,omitempty" bson:"_id,omitempty"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	IndexingStatus string `json:"indexing_status"`
}

func getRepositoryByID(ctx context.Context, repoID string, repoCol *mongo.Collection) (Repository, error) {
	var repo Repository
	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return repo, fmt.Errorf("failed to convert repoID to ObjectID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	err = repoCol.FindOne(ctx, filter).Decode(&repo)
	if err != nil {
		return repo, fmt.Errorf("failed to find repository by ID: %w", err)
	}

	return repo, nil
}

func updateRepositoryStatus(ctx context.Context, repo Repository, status string, repoCol *mongo.Collection) error {
	_, err := repoCol.UpdateOne(ctx, bson.M{"_id": repo.ID}, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		return fmt.Errorf("failed to update repository status: %w", err)
	}

	return bson.ErrDecodeToNil
}
