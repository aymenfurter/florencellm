package main

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
)

func TestProcessRepository(t *testing.T) {
	// until bd674acdbc7907ca30352081bc7137511b2e79c7
	ctx := context.Background()

	repo := Repository{
		//URL: "https://github.com/MicrosoftDocs/architecture-center.git",
		URL: "https://github.com/MicrosoftDocs/azure-docs.git",
	}

	err := processRepository(ctx, repo, "bd674acdbc7907ca30352081bc7137511b2e79c7")
	require.NoError(t, err)
}

type GitRepoCloner interface {
	PlainCloneContext(ctx context.Context, path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
}

type defaultGitRepoCloner struct{}

func (d *defaultGitRepoCloner) PlainCloneContext(ctx context.Context, path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainCloneContext(ctx, path, isBare, o)
}
