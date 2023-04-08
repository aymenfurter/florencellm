package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	servicebus "github.com/Azure/azure-service-bus-go"
	openai "github.com/sashabaranov/go-openai"
	schema "github.com/weaviate/weaviate-go-client/v4/schema"
	weaviate "github.com/weaviate/weaviate-go-client/v4/weaviate"
)

func main() {
	// Load environment variables
	serviceBusConnectionString := os.Getenv("AZURE_SERVICE_BUS_CONNECTION_STRING")
	queueName := os.Getenv("QUEUE_NAME")
	weaviateURL := os.Getenv("WEAVIATE_URL")

	// Initialize Weaviate client
	weaviateClient := weaviate.New(weaviateURL)

	// Initialize MongoDB connection
	mongoDBConnectionString := os.Getenv("COSMOS_DB_CONNECTION_STRING")
	clientOptions := options.Client().ApplyURI(mongoDBConnectionString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	// Get the collection for the repositories
	repoCol := client.Database("repositoryDB").Collection("repositories")

	// Create a new Service Bus namespace
	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(serviceBusConnectionString))
	if err != nil {
		panic(err)
	}

	// Create a new Service Bus queue
	queue, err := ns.NewQueue(queueName)
	if err != nil {
		panic(err)
	}

	// Set up a message handler for the Service Bus queue
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err = queue.Receive(ctx, func(ctx context.Context, msg *servicebus.Message) error {
		repoID := string(msg.Data)

		// Process the repository and index it
		err := indexRepository(ctx, repoID, repoCol, weaviateClient)
		if err != nil {
			fmt.Println("Error indexing repository:", err)
		} else {
			fmt.Printf("Repository %s indexed successfully\n", repoID)
		}

		// Mark the message as completed
		return msg.Complete(ctx)
	})

	if err != nil {
		fmt.Println("Error processing messages:", err)
	}
}

func indexRepository(ctx context.Context, repoID string, repoCol *mongo.Collection, weaviateClient *weaviate.Client) error {
	// Retrieve the repository from MongoDB
	repo, err := getRepositoryByID(ctx, repoID, repoCol)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// Clone the Git repository
	r, err := git.PlainCloneContext(ctx, "", true, &git.CloneOptions{
		URL:               repo.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Iterate through the commits and extract commit information
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
			// Process each commit, generate embeddings and store them in Weaviate
			err := processCommit(ctx, commit, weaviateClient)
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

	// Update the repository status in MongoDB
	_, err = repoCol.UpdateOne(ctx, bson.M{"_id": repo.ID}, bson.M{"$set": bson.M{"status": "indexed"}})
	if err != nil {
		return fmt.Errorf("failed to update repository status: %w", err)
	}

	return nil
}

func commitToJSON(username string, email string, diff string) string {
	return fmt.Sprintf(`{"username": "%s", "email": "%s", "diff": "%s"}`, username, email, diff)
}

func getDiff(commit *object.Commit) string {
	previousCommit, err := commit.Parent(0)
	diff, err2 := previousCommit.Patch(commit)

	if (err != nil) || (err2 != nil) {
		return ""
	}

	diffString := diff.String()
	return diffString
}

func processCommit(ctx context.Context, commit *object.Commit, weaviateClient *weaviate.Client) error {
	commitAsJson := commitToJSON(commit.Author.Name, commit.Author.Email, getDiff(commit))
	embeddings, nil := generateEmbeddings(commitAsJson)

	// Store embeddings in Weaviate
	err := storeEmbeddingsInWeaviate(ctx, commit, embeddings, weaviateClient)
	if err != nil {
		return fmt.Errorf("failed to store embeddings in Weaviate: %w", err)
	}

	return nil
}

func generateEmbeddings(message string) ([]float32, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	config := openai.DefaultConfig(apiKey)
	client := openai.NewClientWithConfig(config)
	ctx := context.Background()

	embeddingReq := openai.EmbeddingRequest{
		Input: []string{message},
		Model: openai.DavinciSimilarity, // Choose the desired model
	}

	embeddingResponse, err := client.CreateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if len(embeddingResponse.Embeddings) == 0 {
		return nil, fmt.Errorf("empty embeddings response")
	}

	// Convert the returned embeddings to float32
	embeddings := make([]float32, len(embeddingResponse.Embeddings[0].Embeddings))
	for i, value := range embeddingResponse.Embeddings[0].Embeddings {
		embeddings[i] = float32(value)
	}

	return embeddings, nil
}

func getRepositoryByID(ctx context.Context, repoID string, repoCol *mongo.Collection) (Repository, error) {
	var repo Repository

	// Convert repoID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return repo, fmt.Errorf("failed to convert repoID to ObjectID: %w", err)
	}

	// Find the repository in MongoDB
	filter := bson.M{"_id": objectID}
	err = repoCol.FindOne(ctx, filter).Decode(&repo)
	if err != nil {
		return repo, fmt.Errorf("failed to find repository by ID: %w", err)
	}

	return repo, nil
}

func storeEmbeddingsInWeaviate(ctx context.Context, commit *object.Commit, embeddings []float32, weaviateClient *weaviate.Client) error {
	// Create the commit object
	commitObject := map[string]interface{}{
		"id":         commit.Hash.String(),
		"message":    commit.Message,
		"author":     commit.Author.Name,
		"email":      commit.Author.Email,
		"timestamp":  commit.Author.When,
		"embeddings": schema.Vector(embeddings),
	}

	// Store the commit object in Weaviate
	create := weaviateClient.Objects().Creator()
	create.SetTimeout(30 * time.Second)
	create.SetObject(commitObject)
	create.SetClass("Commit")

	_, err := create.Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to store commit object in Weaviate: %w", err)
	}

	return nil
}

type Repository struct {
	ID             string `json:"id,omitempty" bson:"_id,omitempty"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	IndexingStatus string `json:"indexing_status"`
}
