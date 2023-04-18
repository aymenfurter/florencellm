package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"golang.org/x/sync/semaphore"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func indexRepository(ctx context.Context, repoID string, repoCol *mongo.Collection) error {
	repo, err := getRepositoryByID(ctx, repoID, repoCol)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	err = processRepository(ctx, repo, "")
	if err != nil {
		return fmt.Errorf("failed to process repository: %w", err)
	}

	_, err = repoCol.UpdateOne(ctx, bson.M{"_id": repo.ID}, bson.M{"$set": bson.M{"status": "indexed"}})
	if err != nil {
		return fmt.Errorf("failed to update repository status: %w", err)
	}

	return nil
}

var mutex = &sync.Mutex{}

func getNextReference(refIter storer.ReferenceIter) (*plumbing.Reference, error) {
	mutex.Lock()
	ref, err := refIter.Next()
	mutex.Unlock()
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func gitClone(repoURL, folderName, referenceName string, depth int) error {
	tmpDir := filepath.Join(tempDir(), folderName)
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "clone", "--branch", referenceName, "--single-branch", "--depth", fmt.Sprint(depth), repoURL, tmpDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return err
	}

	return nil
}

func tempDir() string {
	tmpFolder := os.Getenv("TEMP_FOLDER")
	if tmpFolder == "" {
		tmpFolder = os.TempDir()
	}
	return tmpFolder
}

func processRepository(ctx context.Context, repo Repository, lastCommit string) error {
	if repo.URL == "" {
		return fmt.Errorf("repository URL is empty")
	}

	folderName := extractFolderName(repo.URL)
	fmt.Printf("Cloning repository: %s\n", repo.URL)

	r, err := openOrCloneRepo(repo.URL, folderName, "main", 20000)
	if err != nil {
		return err
	}

	fmt.Printf("Cloning completed: %s\n", repo.URL)
	return processBranches(ctx, r, lastCommit, repo)
}

func extractFolderName(url string) string {
	folderName := url[strings.LastIndex(url, "/")+1:]
	return folderName[:len(folderName)-4]
}

func openOrCloneRepo(url, folderName, branch string, depth int) (*git.Repository, error) {
	tempFolderPath := filepath.Join(tempDir(), folderName)

	if _, err := os.Stat(tempFolderPath); os.IsNotExist(err) {
		err = gitClone(url, folderName, branch, depth)
		if err != nil {
			return nil, err
		}
	}

	return git.PlainOpen(tempFolderPath)
}

func processBranches(ctx context.Context, r *git.Repository, lastCommit string, repo Repository) error {
	refIter, err := r.Branches()
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(1)
	branchesRemaining := true

	for branchesRemaining {
		ref, err := getNextReference(refIter)
		if err != nil || ref == nil {
			branchesRemaining = false
			break
		}

		if !isMasterOrMainBranch(ref.Name().Short()) {
			continue
		}

		fmt.Printf("Processing branch: %s\n", ref.Name().Short())
		mutex.Lock()
		iter, err := r.Log(&git.LogOptions{From: ref.Hash()})
		mutex.Unlock()

		commitRemaining := true
		reachedCheckpoint := lastCommit == ""

		for commitRemaining {
			commit, err := getNextCommit(iter)
			commitRemaining, reachedCheckpoint, err = handleCommit(ctx, commit, err, r, &wg, sem, repo.URL, lastCommit, reachedCheckpoint)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getNextCommit(iter object.CommitIter) (*object.Commit, error) {
	mutex.Lock()
	commit, err := iter.Next()
	mutex.Unlock()
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func handleCommit(ctx context.Context, commit *object.Commit, err error, r *git.Repository, wg *sync.WaitGroup, sem *semaphore.Weighted, repoURL, lastCommit string, reachedCheckpoint bool) (bool, bool, error) {
	if reachedCheckpoint == false && commit != nil && commit.Hash.String() != lastCommit {
		fmt.Printf("Skipping commit: %s\n", commit.Hash.String())
	} else {
		reachedCheckpoint = true
		if err != nil || commit == nil {
			return false, reachedCheckpoint, err
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			return false, reachedCheckpoint, fmt.Errorf("failed to acquire semaphore: %w", err)
		}

		wg.Add(1)
		go func() {
			defer sem.Release(50)
			defer wg.Done()

			fmt.Printf("Processing commit: %s\n", commit.Hash.String())
			err := processCommit(ctx, commit, repoURL)
			if err != nil {
				fmt.Printf("Warning: failed to process commit: %s\n", err.Error())
			}
		}()
	}
	return true, reachedCheckpoint, nil
}

func isMasterOrMainBranch(branch string) bool {
	return branch == "master" || branch == "main"
}

func commitToJSON(username string, email string, diff string) string {
	return fmt.Sprintf(`{"username": "%s", "email": "%s", "diff": "%s"}`, username, email, diff)
}

func getDiff(commit *object.Commit) (string, error) {
	if commit == nil || commit.NumParents() == 0 {
		return "", nil
	}

	previousCommit, err := commit.Parent(0)
	if err != nil {
		return "", fmt.Errorf("error getting parent commit: %w", err)
	}

	if previousCommit == nil {
		return "", errors.New("previousCommit is nil")
	}

	diff, err := previousCommit.Patch(commit)
	if err != nil {
		return "", fmt.Errorf("error getting diff: %w", err)
	}

	diffString := diff.String()
	if len(diffString) > 32000 {
		diffString = diffString[:32000]
	}
	return diffString, nil
}

func processCommit(ctx context.Context, commit *object.Commit, repoURL string) error {
	author := commit.Author
	email := author.Email
	diffString, err := getDiff(commit)
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	commitID := commit.Hash.String()
	commitMsg := commit.Message
	embeddings, err := generateEmbeddings(commitMsg, author, email, diffString, commitID, repoURL)

	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	err = storeEmbeddings(commitID, embeddings)
	if err != nil {
		return fmt.Errorf("failed to store embeddings in Pinecone: %w", err)
	}

	return nil
}
