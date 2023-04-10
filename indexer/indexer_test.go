package main

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
)

func TestProcessRepository(t *testing.T) {
	ctx := context.Background()

	// Define a sample repository
	repo := Repository{
		URL: "https://github.com/aymenfurter/todo.js.git",
	}

	err := processRepository(ctx, repo, buildPineconeClient())
	require.NoError(t, err)
}

type GitRepoCloner interface {
	PlainCloneContext(ctx context.Context, path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
}

type defaultGitRepoCloner struct{}

func (d *defaultGitRepoCloner) PlainCloneContext(ctx context.Context, path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainCloneContext(ctx, path, isBare, o)
}
