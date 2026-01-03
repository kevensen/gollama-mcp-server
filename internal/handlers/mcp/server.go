package mcp

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	mcp_go "github.com/mark3labs/mcp-go/mcp"
	mcp_go_server "github.com/mark3labs/mcp-go/server"
)

type Server struct {
	*mcp_go_server.MCPServer
	OllamaClient OllamaClient
	ready        bool
}

func NewServer() *Server {
	hooks := &mcp_go_server.Hooks{}

	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		slog.InfoContext(ctx, "beforeCallTool", slog.Any("id", id), slog.Any("message", message))
	})
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		slog.InfoContext(ctx, "afterCallTool", slog.Any("id", id), slog.Any("message", message), slog.Any("result", result))
	})

	client, err := NewOllamaClient()
	if err != nil {
		slog.Error("Failed to create Ollama client", "error", err)
		return nil
	}

	s := &Server{
		MCPServer: mcp_go_server.NewMCPServer(
			"github.com/kevensen/gollama-mcp-server",
			"1.0.0",
			mcp_go_server.WithToolCapabilities(true),
			mcp_go_server.WithLogging(),
			mcp_go_server.WithHooks(hooks),
		),
		OllamaClient: client,
	}

	// Register tools
	s.registerTools()

	return s
}

func (s *Server) registerTools() {
	// List models
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"listModels",
			mcp_go.WithDescription("List all available models in Ollama."),
			mcp_go.WithReadOnlyHintAnnotation(true),
		),
		s.ListModels)

	// Pull model
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"pullModel",
			mcp_go.WithDescription("Pull a model from the Ollama library. This operation may take some time depending on the model size."),
			mcp_go.WithString("model", mcp_go.Required()),
			mcp_go.WithBoolean("insecure"),
			mcp_go.WithBoolean("stream"),
			mcp_go.WithOpenWorldHintAnnotation(true),
		),
		s.PullModel)

	// Push model
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"pushModel",
			mcp_go.WithDescription("Push a model to a registry."),
			mcp_go.WithString("model", mcp_go.Required()),
			mcp_go.WithBoolean("insecure"),
			mcp_go.WithBoolean("stream"),
			mcp_go.WithOpenWorldHintAnnotation(true),
		),
		s.PushModel)

	// Delete model
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"deleteModel",
			mcp_go.WithDescription("Delete a model from the local Ollama instance."),
			mcp_go.WithString("model", mcp_go.Required()),
			mcp_go.WithOpenWorldHintAnnotation(true),
			mcp_go.WithDestructiveHintAnnotation(true),
		),
		s.DeleteModel)

	// Show model info
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"showModel",
			mcp_go.WithDescription("Show detailed information about a model including modelfile, parameters, template, and more."),
			mcp_go.WithString("model", mcp_go.Required()),
			mcp_go.WithReadOnlyHintAnnotation(true),
		),
		s.ShowModel)

	// Copy model
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"copyModel",
			mcp_go.WithDescription("Copy a model to create a new model with a different name."),
			mcp_go.WithString("source", mcp_go.Required()),
			mcp_go.WithString("destination", mcp_go.Required()),
		),
		s.CopyModel)

	// Generate embeddings
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"embeddings",
			mcp_go.WithDescription("Generate embeddings for the given text using a specified model."),
			mcp_go.WithString("model", mcp_go.Required()),
			mcp_go.WithString("input", mcp_go.Required()),
			mcp_go.WithReadOnlyHintAnnotation(true),
		),
		s.Embeddings)
}

func (s *Server) Ready() bool {
	return s.ready
}
