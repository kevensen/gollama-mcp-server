package mcp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
)

// OllamaClient is an interface for interacting with the Ollama API
type OllamaClient interface {
	List(ctx context.Context) (*api.ListResponse, error)
	Generate(ctx context.Context, req *api.GenerateRequest, fn func(api.GenerateResponse) error) error
	Chat(ctx context.Context, req *api.ChatRequest, fn func(api.ChatResponse) error) error
	Pull(ctx context.Context, req *api.PullRequest, fn func(api.ProgressResponse) error) error
	Push(ctx context.Context, req *api.PushRequest, fn func(api.ProgressResponse) error) error
	Delete(ctx context.Context, req *api.DeleteRequest) error
	Show(ctx context.Context, req *api.ShowRequest) (*api.ShowResponse, error)
	Copy(ctx context.Context, req *api.CopyRequest) error
	Embeddings(ctx context.Context, req *api.EmbeddingRequest) (*api.EmbeddingResponse, error)
}

// LiveOllamaClient is the production implementation of OllamaClient
type LiveOllamaClient struct {
	client *api.Client
}

// NewOllamaClient creates a new Ollama client using the OLLAMA_HOST environment variable
// or defaulting to http://localhost:11434
func NewOllamaClient() (*LiveOllamaClient, error) {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}

	// Parse and validate the URL
	parsedURL, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("invalid OLLAMA_HOST URL: %w", err)
	}

	client := api.NewClient(parsedURL, http.DefaultClient)

	return &LiveOllamaClient{
		client: client,
	}, nil
}

func (c *LiveOllamaClient) List(ctx context.Context) (*api.ListResponse, error) {
	return c.client.List(ctx)
}

func (c *LiveOllamaClient) Generate(ctx context.Context, req *api.GenerateRequest, fn func(api.GenerateResponse) error) error {
	return c.client.Generate(ctx, req, fn)
}

func (c *LiveOllamaClient) Chat(ctx context.Context, req *api.ChatRequest, fn func(api.ChatResponse) error) error {
	return c.client.Chat(ctx, req, fn)
}

func (c *LiveOllamaClient) Pull(ctx context.Context, req *api.PullRequest, fn func(api.ProgressResponse) error) error {
	return c.client.Pull(ctx, req, fn)
}

func (c *LiveOllamaClient) Push(ctx context.Context, req *api.PushRequest, fn func(api.ProgressResponse) error) error {
	return c.client.Push(ctx, req, fn)
}

func (c *LiveOllamaClient) Delete(ctx context.Context, req *api.DeleteRequest) error {
	return c.client.Delete(ctx, req)
}

func (c *LiveOllamaClient) Show(ctx context.Context, req *api.ShowRequest) (*api.ShowResponse, error) {
	return c.client.Show(ctx, req)
}

func (c *LiveOllamaClient) Copy(ctx context.Context, req *api.CopyRequest) error {
	return c.client.Copy(ctx, req)
}

func (c *LiveOllamaClient) Embeddings(ctx context.Context, req *api.EmbeddingRequest) (*api.EmbeddingResponse, error) {
	return c.client.Embeddings(ctx, req)
}
