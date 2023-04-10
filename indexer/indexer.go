package main

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pinecone-io/go-pinecone/pinecone_grpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func indexRepository(ctx context.Context, repoID string, repoCol *mongo.Collection, pcClient pinecone_grpc.VectorServiceClient) error {
	repo, err := getRepositoryByID(ctx, repoID, repoCol)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	err = processRepository(ctx, repo, pcClient)
	if err != nil {
		return fmt.Errorf("failed to process repository: %w", err)
	}

	_, err = repoCol.UpdateOne(ctx, bson.M{"_id": repo.ID}, bson.M{"$set": bson.M{"status": "indexed"}})
	if err != nil {
		return fmt.Errorf("failed to update repository status: %w", err)
	}

	return nil
}

func processRepository(ctx context.Context, repo Repository, pcClient pinecone_grpc.VectorServiceClient) error {
	storer := memory.NewStorage()

	r, err := git.Clone(storer, nil, &git.CloneOptions{
		URL:               repo.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	refIter, err := r.Branches()
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	err = refIter.ForEach(func(ref *plumbing.Reference) error {
		iter, err := r.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			return fmt.Errorf("failed to get log: %w", err)
		}

		err = iter.ForEach(func(commit *object.Commit) error {
			err := processCommit(ctx, commit, pcClient)
			if err != nil {
				return fmt.Errorf("failed to process commit: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to iterate through commits: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to iterate through branches: %w", err)
	}
	return nil
}

func commitToJSON(username string, email string, diff string) string {
	return fmt.Sprintf(`{"username": "%s", "email": "%s", "diff": "%s"}`, username, email, diff)
}

func getDiff(commit *object.Commit) string {
	if commit.NumParents() == 0 {
		return ""
	}
	previousCommit, err := commit.Parent(0)
	diff, err2 := previousCommit.Patch(commit)

	if (err != nil) || (err2 != nil) {
		return ""
	}

	diffString := diff.String()
	return diffString
}

func processCommit(ctx context.Context, commit *object.Commit, pcClient pinecone_grpc.VectorServiceClient) error {
	author := commit.Author
	email := author.Email
	diffString := getDiff(commit)
	commitId := commit.Hash.String()
	embeddings, err := generateEmbeddings(author, email, diffString)

	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	fmt.Println(embeddings)
	if pcClient == nil {
		return nil
	}

	err = storeEmbeddings(commitId, embeddings, pcClient)
	if err != nil {
		return fmt.Errorf("failed to store embeddings in Pinecone: %w", err)
	}

	return nil
}
