package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/go-git/go-git/v5/storage/filesystem"
	//"github.com/go-git/go-git/v5/storage/memory"
	"github.com/go-git/go-billy/v5/osfs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/semaphore"
)

func indexRepository(ctx context.Context, repoID string, repoCol *mongo.Collection) error {
	repo, err := getRepositoryByID(ctx, repoID, repoCol)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	err = processRepository(ctx, repo)
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

func getNextCommit(iter object.CommitIter) (*object.Commit, error) {
	mutex.Lock()
	commit, err := iter.Next()
	mutex.Unlock()
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func processRepository(ctx context.Context, repo Repository) error {
	//storer := memory.NewStorage()

	if repo.URL == "" {
		return fmt.Errorf("repository URL is empty")
	}

	folderName := repo.URL[strings.LastIndex(repo.URL, "/")+1:]

	storer := filesystem.NewStorage(
		osfs.New("/tmp/"+folderName+"/"),
		cache.NewObjectLRUDefault())

	fmt.Println("Cloning repository.. ", repo.URL)

	r := &git.Repository{}

	_, err := storer.Reference(plumbing.ReferenceName("refs/heads/main"))
	if err == nil {
		fmt.Println("Repository already exists.. ", repo.URL)
		r, err = git.Open(storer, osfs.New("/tmp/"+folderName+"/"))
	} else {
		fmt.Println("Repository does not exist.. ", repo.URL)
		r, err = git.Clone(storer, osfs.New("/tmp/"+folderName+"/"), &git.CloneOptions{
			URL:           repo.URL,
			ReferenceName: plumbing.ReferenceName("refs/heads/main"),
			SingleBranch:  true,
			Depth:         20000,
		})
	}

	fmt.Println("Cloning completed.. ", repo.URL)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	refIter, err := r.Branches()
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(50)
	branchesRemaining := true

	for branchesRemaining {
		ref, err := getNextReference(refIter)
		if err != nil || ref == nil {
			branchesRemaining = false
			break
		}
		if (ref.Name().Short() != "master") && (ref.Name().Short() != "main") {
			return nil
		}
		fmt.Println("Processing branch: ", ref.Name().Short())
		mutex.Lock()
		iter, err := r.Log(&git.LogOptions{From: ref.Hash()})
		mutex.Unlock()

		if err != nil {
			return fmt.Errorf("failed to get log: %w", err)
		}

		commitRemaining := true
		for commitRemaining {
			commit, err := getNextCommit(iter)
			if err != nil || ref == nil {
				commitRemaining = false
				break
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %w", err)
			}

			wg.Add(1)
			go func() {
				defer sem.Release(1)
				defer wg.Done()

				fmt.Println("Processing commit: ", commit.Hash.String())
				err := processCommit(ctx, commit, repo.URL)
				if err != nil {
					fmt.Println("Warning: failed to process commit: ", err.Error())
				}
			}()
		}
		if err != nil {
			fmt.Println("Warning: failed to iterate through commits: ", err.Error())
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to iterate through branches: %w", err)
	}

	wg.Wait()
	return nil
}

func commitToJSON(username string, email string, diff string) string {
	return fmt.Sprintf(`{"username": "%s", "email": "%s", "diff": "%s"}`, username, email, diff)
}

func getDiff(commit *object.Commit) string {
	if commit == nil || commit.NumParents() == 0 {
		return ""
	}

	previousCommit, err := commit.Parent(0)
	if err != nil {
		fmt.Println("Error getting parent commit:", err)
		return ""
	}

	if previousCommit == nil {
		fmt.Println("Error: previousCommit is nil")
		return ""
	}

	diff, err2 := previousCommit.Patch(commit)
	if err2 != nil {
		fmt.Println("Error getting diff:", err2)
		return ""
	}

	diffString := diff.String()
	if len(diffString) > 32000 {
		diffString = diffString[:32000]
	}
	return diffString
}

func processCommit(ctx context.Context, commit *object.Commit, repoUrl string) error {
	mutex.Lock()
	author := commit.Author
	email := author.Email
	diffString := getDiff(commit)
	commitId := commit.Hash.String()
	commitMsg := commit.Message
	mutex.Unlock()
	embeddings, err := generateEmbeddings(commitMsg, author, email, diffString, commitId, repoUrl)

	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	err = storeEmbeddings(commitId, embeddings)
	if err != nil {
		return fmt.Errorf("failed to store embeddings in Pinecone: %w", err)
	}

	return nil
}
