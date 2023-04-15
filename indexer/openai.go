package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pinecone-io/go-pinecone/pinecone_grpc"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/types/known/structpb"
)

const maxDiffStringLength = 8000
const chunkCutoffThreshold = 3

func generateEmbeddings(commitMsg string, author object.Signature, email, diffString string, commitId string, repoURL string) ([]*pinecone_grpc.Vector, error) {
	client := newOpenAIClient()

	if len(diffString) < maxDiffStringLength {
		return generateSingleEmbedding(client, commitMsg, author, email, diffString, commitId, repoURL)
	}

	return generateChunkedEmbeddings(client, commitMsg, author, email, diffString, commitId, repoURL)
}

func newOpenAIClient() *openai.Client {
	return openai.NewClient(os.Getenv("OPEN_AI_KEY"))
}

func generateSingleEmbedding(client *openai.Client, commitMsg string, author object.Signature, email, diffString string, commitId string, repoURL string) ([]*pinecone_grpc.Vector, error) {
	embeddings := make([]*pinecone_grpc.Vector, 0)
	input := fmt.Sprintf("Author: %s\nRepoURL:\n%s\nCommit-Message:\n%s\nEmail: %s\nCommitId: \n%s\nDiff: %s", author.Name, repoURL, commitMsg, email, commitId, diffString)
	metadata := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"text": &structpb.Value{
				Kind: &structpb.Value_StringValue{
					StringValue: input,
				},
			},
		},
	}

	embeddingReq := createEmbeddingRequest(input)

	response, err := requestEmbeddings(client, embeddingReq)

	if err != nil {
		return nil, err
	}

	embeddingsId := commitId

	embedding := &pinecone_grpc.Vector{
		Id:       embeddingsId,
		Values:   response.Data[0].Embedding,
		Metadata: metadata,
	}

	embeddings = append(embeddings, embedding)

	return embeddings, nil
}

func generateChunkedEmbeddings(client *openai.Client, commitMsg string, author object.Signature, email, diffString string, commitId string, repoURL string) ([]*pinecone_grpc.Vector, error) {
	chunks := chunkString(diffString, maxDiffStringLength)
	embeddings := make([]*pinecone_grpc.Vector, 0)

	for i, chunk := range chunks {
		if i <= chunkCutoffThreshold {
			input := fmt.Sprintf("Author: %s\nRepo-URL:\n%s\nCommit-Message:%s\nEmail: %s\nChunk:\n%d\nCommitId: \n%s\nDiff: %s", author.Name, repoURL, commitMsg, email, i, commitId, chunk)
			embeddingReq := createEmbeddingRequest(input)

			response, err := requestEmbeddings(client, embeddingReq)
			if err != nil {
				return nil, err
			}

			embeddingsId := commitId
			if i != 0 {
				embeddingsId = fmt.Sprintf("%s-%d", commitId, i)
			}

			metadata := &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"text": &structpb.Value{
						Kind: &structpb.Value_StringValue{
							StringValue: input,
						},
					},
				},
			}

			embedding := &pinecone_grpc.Vector{
				Id:       embeddingsId,
				Values:   response.Data[0].Embedding,
				Metadata: metadata,
			}

			embeddings = append(embeddings, embedding)
		}
	}

	return embeddings, nil
}

func createEmbeddingRequest(input string) openai.EmbeddingRequest {
	return openai.EmbeddingRequest{
		Input: []string{input},
		Model: openai.AdaEmbeddingV2,
	}
}

func requestEmbeddings(client *openai.Client, embeddingReq openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	response, err := client.CreateEmbeddings(context.Background(), embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings from OpenAI: %w", err)
	}

	//time.Sleep(1 * time.Second) // rate limit
	return &response, nil
}

func chunkString(str string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(str); i += chunkSize {
		end := i + chunkSize
		if end > len(str) {
			end = len(str)
		}
		chunks = append(chunks, str[i:end])
	}
	return chunks
}
