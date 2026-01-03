package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ollama/ollama/api"

	mcp_go "github.com/mark3labs/mcp-go/mcp"
)

// ListModels lists all available models in Ollama
func (s *Server) ListModels(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	resp, err := s.OllamaClient.List(ctx)
	if err != nil {
		return nil, NewOllamaClientError("list", err)
	}

	var modelNames []string
	var modelDetails []map[string]interface{}

	for _, model := range resp.Models {
		modelNames = append(modelNames, model.Name)
		modelDetails = append(modelDetails, map[string]interface{}{
			"name":     model.Name,
			"size":     model.Size,
			"digest":   model.Digest,
			"modified": model.ModifiedAt,
			"details":  model.Details,
		})
	}

	result, err := json.MarshalIndent(map[string]interface{}{
		"models": modelDetails,
		"count":  len(modelNames),
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp_go.NewToolResultText(string(result)), nil
}

// PullModel pulls a model from the Ollama library
func (s *Server) PullModel(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	model := request.GetString("model", "")
	if model == "" {
		return nil, NewMissingParameterError("model")
	}

	insecure := request.GetBool("insecure", false)
	stream := request.GetBool("stream", true)

	req := &api.PullRequest{
		Model:    model,
		Insecure: insecure,
		Stream:   &stream,
	}

	var statusMessages []string
	var lastStatus string

	err := s.OllamaClient.Pull(ctx, req, func(resp api.ProgressResponse) error {
		if stream && resp.Status != lastStatus {
			slog.InfoContext(ctx, "Pull progress", "status", resp.Status, "completed", resp.Completed, "total", resp.Total)
			lastStatus = resp.Status
			statusMessages = append(statusMessages, fmt.Sprintf("%s: %d/%d", resp.Status, resp.Completed, resp.Total))
		}
		return nil
	})

	if err != nil {
		return nil, NewOllamaClientError("pull", err)
	}

	result, err := json.MarshalIndent(map[string]interface{}{
		"model":    model,
		"status":   "success",
		"progress": statusMessages,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp_go.NewToolResultText(string(result)), nil
}

// PushModel pushes a model to a registry
func (s *Server) PushModel(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	model := request.GetString("model", "")
	if model == "" {
		return nil, NewMissingParameterError("model")
	}

	insecure := request.GetBool("insecure", false)
	stream := request.GetBool("stream", true)

	req := &api.PushRequest{
		Model:    model,
		Insecure: insecure,
		Stream:   &stream,
	}

	var statusMessages []string
	var lastStatus string

	err := s.OllamaClient.Push(ctx, req, func(resp api.ProgressResponse) error {
		if stream && resp.Status != lastStatus {
			slog.InfoContext(ctx, "Push progress", "status", resp.Status, "completed", resp.Completed, "total", resp.Total)
			lastStatus = resp.Status
			statusMessages = append(statusMessages, fmt.Sprintf("%s: %d/%d", resp.Status, resp.Completed, resp.Total))
		}
		return nil
	})

	if err != nil {
		return nil, NewOllamaClientError("push", err)
	}

	result, err := json.MarshalIndent(map[string]interface{}{
		"model":    model,
		"status":   "success",
		"progress": statusMessages,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp_go.NewToolResultText(string(result)), nil
}

// DeleteModel deletes a model from the local Ollama instance
func (s *Server) DeleteModel(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	model := request.GetString("model", "")
	if model == "" {
		return nil, NewMissingParameterError("model")
	}

	req := &api.DeleteRequest{
		Model: model,
	}

	err := s.OllamaClient.Delete(ctx, req)
	if err != nil {
		return nil, NewOllamaClientError("delete", err)
	}

	result, err := json.MarshalIndent(map[string]interface{}{
		"model":  model,
		"status": "deleted",
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp_go.NewToolResultText(string(result)), nil
}

// ShowModel shows detailed information about a model
func (s *Server) ShowModel(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	model := request.GetString("model", "")
	if model == "" {
		return nil, NewMissingParameterError("model")
	}

	req := &api.ShowRequest{
		Model: model,
	}

	resp, err := s.OllamaClient.Show(ctx, req)
	if err != nil {
		return nil, NewOllamaClientError("show", err)
	}

	result, err := json.MarshalIndent(map[string]interface{}{
		"model":      model,
		"modelfile":  resp.Modelfile,
		"parameters": resp.Parameters,
		"template":   resp.Template,
		"details":    resp.Details,
		"modified":   resp.ModifiedAt,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp_go.NewToolResultText(string(result)), nil
}

// CopyModel copies a model to create a new model with a different name
func (s *Server) CopyModel(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	source := request.GetString("source", "")
	if source == "" {
		return nil, NewMissingParameterError("source")
	}

	destination := request.GetString("destination", "")
	if destination == "" {
		return nil, NewMissingParameterError("destination")
	}

	req := &api.CopyRequest{
		Source:      source,
		Destination: destination,
	}

	err := s.OllamaClient.Copy(ctx, req)
	if err != nil {
		return nil, NewOllamaClientError("copy", err)
	}

	result, err := json.MarshalIndent(map[string]interface{}{
		"source":      source,
		"destination": destination,
		"status":      "copied",
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp_go.NewToolResultText(string(result)), nil
}

// Embeddings generates embeddings for the given text
func (s *Server) Embeddings(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	model := request.GetString("model", "")
	if model == "" {
		return nil, NewMissingParameterError("model")
	}

	input := request.GetString("input", "")
	if input == "" {
		return nil, NewMissingParameterError("input")
	}

	req := &api.EmbeddingRequest{
		Model:  model,
		Prompt: input,
	}

	resp, err := s.OllamaClient.Embeddings(ctx, req)
	if err != nil {
		return nil, NewOllamaClientError("embeddings", err)
	}

	result, err := json.MarshalIndent(map[string]interface{}{
		"model":      model,
		"embeddings": resp.Embedding,
		"dimensions": len(resp.Embedding),
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp_go.NewToolResultText(string(result)), nil
}
