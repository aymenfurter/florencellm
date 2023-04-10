package main

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
)

func TestProcessRepository(t *testing.T) {
	ctx := context.Background()

	repo := Repository{
		URL: "https://github.com/MicrosoftDocs/architecture-center.git",
	}

	err := processRepository(ctx, repo)
	require.NoError(t, err)
}

type GitRepoCloner interface {
	PlainCloneContext(ctx context.Context, path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
}

type defaultGitRepoCloner struct{}

func (d *defaultGitRepoCloner) PlainCloneContext(ctx context.Context, path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainCloneContext(ctx, path, isBare, o)
}
