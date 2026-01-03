package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ollama/ollama/api"

	mcp_go "github.com/mark3labs/mcp-go/mcp"
)

// mockOllamaClient is a mock implementation of OllamaClient for testing
type mockOllamaClient struct {
	listResponse       *api.ListResponse
	listError          error
	pullError          error
	pullProgress       []api.ProgressResponse
	pushError          error
	pushProgress       []api.ProgressResponse
	deleteError        error
	showResponse       *api.ShowResponse
	showError          error
	copyError          error
	embeddingsResponse *api.EmbeddingResponse
	embeddingsError    error
}

func (m *mockOllamaClient) List(ctx context.Context) (*api.ListResponse, error) {
	return m.listResponse, m.listError
}

func (m *mockOllamaClient) Generate(ctx context.Context, req *api.GenerateRequest, fn func(api.GenerateResponse) error) error {
	return nil
}

func (m *mockOllamaClient) Chat(ctx context.Context, req *api.ChatRequest, fn func(api.ChatResponse) error) error {
	return nil
}

func (m *mockOllamaClient) Pull(ctx context.Context, req *api.PullRequest, fn func(api.ProgressResponse) error) error {
	if m.pullError != nil {
		return m.pullError
	}
	for _, progress := range m.pullProgress {
		if err := fn(progress); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockOllamaClient) Push(ctx context.Context, req *api.PushRequest, fn func(api.ProgressResponse) error) error {
	if m.pushError != nil {
		return m.pushError
	}
	for _, progress := range m.pushProgress {
		if err := fn(progress); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockOllamaClient) Delete(ctx context.Context, req *api.DeleteRequest) error {
	return m.deleteError
}

func (m *mockOllamaClient) Show(ctx context.Context, req *api.ShowRequest) (*api.ShowResponse, error) {
	return m.showResponse, m.showError
}

func (m *mockOllamaClient) Copy(ctx context.Context, req *api.CopyRequest) error {
	return m.copyError
}

func (m *mockOllamaClient) Embeddings(ctx context.Context, req *api.EmbeddingRequest) (*api.EmbeddingResponse, error) {
	return m.embeddingsResponse, m.embeddingsError
}

func TestListModels(t *testing.T) {
	testCases := []struct {
		desc         string
		mockClient   *mockOllamaClient
		wantErr      bool
		validateResp func(*testing.T, *mcp_go.CallToolResult)
	}{
		{
			desc: "Successfully list models",
			mockClient: &mockOllamaClient{
				listResponse: &api.ListResponse{
					Models: []api.ListModelResponse{
						{
							Name:   "llama2:latest",
							Size:   3825819519,
							Digest: "abc123",
							Details: api.ModelDetails{
								Format:            "gguf",
								Family:            "llama",
								ParameterSize:     "7B",
								QuantizationLevel: "Q4_0",
							},
						},
						{
							Name:   "mistral:latest",
							Size:   4109865159,
							Digest: "def456",
							Details: api.ModelDetails{
								Format:            "gguf",
								Family:            "mistral",
								ParameterSize:     "7B",
								QuantizationLevel: "Q4_0",
							},
						},
					},
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, result *mcp_go.CallToolResult) {
				if result == nil || len(result.Content) == 0 {
					t.Fatal("Expected non-empty result")
				}
				textContent, ok := result.Content[0].(mcp_go.TextContent)
				if !ok {
					t.Fatalf("Expected TextContent, got %T", result.Content[0])
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if count, ok := response["count"].(float64); !ok || count != 2 {
					t.Errorf("Expected count 2, got %v", response["count"])
				}
			},
		},
		{
			desc: "Ollama client error",
			mockClient: &mockOllamaClient{
				listError: errors.New("connection refused"),
			},
			wantErr: true,
		},
		{
			desc: "Empty model list",
			mockClient: &mockOllamaClient{
				listResponse: &api.ListResponse{
					Models: []api.ListModelResponse{},
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, result *mcp_go.CallToolResult) {
				textContent, ok := result.Content[0].(mcp_go.TextContent)
				if !ok {
					t.Fatalf("Expected TextContent")
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if count, ok := response["count"].(float64); !ok || count != 0 {
					t.Errorf("Expected count 0, got %v", response["count"])
				}
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				OllamaClient: tc.mockClient,
			}

			request := mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{},
			}

			result, err := s.ListModels(ctx, request)
			if (err != nil) != tc.wantErr {
				t.Errorf("ListModels() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && tc.validateResp != nil {
				tc.validateResp(t, result)
			}
		})
	}
}

func TestPullModel(t *testing.T) {
	testCases := []struct {
		desc         string
		request      mcp_go.CallToolRequest
		mockClient   *mockOllamaClient
		wantErr      bool
		validateResp func(*testing.T, *mcp_go.CallToolResult)
	}{
		{
			desc: "Successfully pull model",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "llama2:latest",
					},
				},
			},
			mockClient: &mockOllamaClient{
				pullProgress: []api.ProgressResponse{
					{Status: "pulling manifest", Completed: 0, Total: 100},
					{Status: "downloading", Completed: 50, Total: 100},
					{Status: "downloading", Completed: 100, Total: 100},
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, result *mcp_go.CallToolResult) {
				textContent, ok := result.Content[0].(mcp_go.TextContent)
				if !ok {
					t.Fatal("Expected TextContent")
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if status := response["status"]; status != "success" {
					t.Errorf("Expected status 'success', got %v", status)
				}
			},
		},
		{
			desc: "Missing model parameter",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    true,
		},
		{
			desc: "Pull error",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "llama2:latest",
					},
				},
			},
			mockClient: &mockOllamaClient{
				pullError: errors.New("model not found"),
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				OllamaClient: tc.mockClient,
			}

			result, err := s.PullModel(ctx, tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("PullModel() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && tc.validateResp != nil {
				tc.validateResp(t, result)
			}
		})
	}
}

func TestPushModel(t *testing.T) {
	testCases := []struct {
		desc         string
		request      mcp_go.CallToolRequest
		mockClient   *mockOllamaClient
		wantErr      bool
		validateResp func(*testing.T, *mcp_go.CallToolResult)
	}{
		{
			desc: "Successfully push model",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "myuser/llama2:custom",
					},
				},
			},
			mockClient: &mockOllamaClient{
				pushProgress: []api.ProgressResponse{
					{Status: "preparing", Completed: 0, Total: 100},
					{Status: "uploading", Completed: 100, Total: 100},
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, result *mcp_go.CallToolResult) {
				textContent, ok := result.Content[0].(mcp_go.TextContent)
				if !ok {
					t.Fatal("Expected TextContent")
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if status := response["status"]; status != "success" {
					t.Errorf("Expected status 'success', got %v", status)
				}
			},
		},
		{
			desc: "Missing model parameter",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    true,
		},
		{
			desc: "Push error",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "myuser/llama2:custom",
					},
				},
			},
			mockClient: &mockOllamaClient{
				pushError: errors.New("unauthorized"),
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				OllamaClient: tc.mockClient,
			}

			result, err := s.PushModel(ctx, tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("PushModel() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && tc.validateResp != nil {
				tc.validateResp(t, result)
			}
		})
	}
}

func TestDeleteModel(t *testing.T) {
	testCases := []struct {
		desc         string
		request      mcp_go.CallToolRequest
		mockClient   *mockOllamaClient
		wantErr      bool
		validateResp func(*testing.T, *mcp_go.CallToolResult)
	}{
		{
			desc: "Successfully delete model",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "llama2:latest",
					},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    false,
			validateResp: func(t *testing.T, result *mcp_go.CallToolResult) {
				textContent, ok := result.Content[0].(mcp_go.TextContent)
				if !ok {
					t.Fatal("Expected TextContent")
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if status := response["status"]; status != "deleted" {
					t.Errorf("Expected status 'deleted', got %v", status)
				}
			},
		},
		{
			desc: "Missing model parameter",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    true,
		},
		{
			desc: "Delete error",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "llama2:latest",
					},
				},
			},
			mockClient: &mockOllamaClient{
				deleteError: errors.New("model not found"),
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				OllamaClient: tc.mockClient,
			}

			result, err := s.DeleteModel(ctx, tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteModel() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && tc.validateResp != nil {
				tc.validateResp(t, result)
			}
		})
	}
}

func TestShowModel(t *testing.T) {
	testCases := []struct {
		desc         string
		request      mcp_go.CallToolRequest
		mockClient   *mockOllamaClient
		wantErr      bool
		validateResp func(*testing.T, *mcp_go.CallToolResult)
	}{
		{
			desc: "Successfully show model info",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "llama2:latest",
					},
				},
			},
			mockClient: &mockOllamaClient{
				showResponse: &api.ShowResponse{
					Modelfile:  "FROM llama2:latest",
					Parameters: "parameter stop \"<|im_end|>\"",
					Template:   "{{ .System }}\n{{ .Prompt }}",
					Details: api.ModelDetails{
						Format:            "gguf",
						Family:            "llama",
						ParameterSize:     "7B",
						QuantizationLevel: "Q4_0",
					},
					ModifiedAt: time.Now(),
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, result *mcp_go.CallToolResult) {
				textContent, ok := result.Content[0].(mcp_go.TextContent)
				if !ok {
					t.Fatal("Expected TextContent")
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if modelfile := response["modelfile"]; modelfile != "FROM llama2:latest" {
					t.Errorf("Expected modelfile, got %v", modelfile)
				}
			},
		},
		{
			desc: "Missing model parameter",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    true,
		},
		{
			desc: "Show error",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "nonexistent:latest",
					},
				},
			},
			mockClient: &mockOllamaClient{
				showError: errors.New("model not found"),
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				OllamaClient: tc.mockClient,
			}

			result, err := s.ShowModel(ctx, tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("ShowModel() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && tc.validateResp != nil {
				tc.validateResp(t, result)
			}
		})
	}
}

func TestCopyModel(t *testing.T) {
	testCases := []struct {
		desc         string
		request      mcp_go.CallToolRequest
		mockClient   *mockOllamaClient
		wantErr      bool
		validateResp func(*testing.T, *mcp_go.CallToolResult)
	}{
		{
			desc: "Successfully copy model",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"source":      "llama2:latest",
						"destination": "llama2:custom",
					},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    false,
			validateResp: func(t *testing.T, result *mcp_go.CallToolResult) {
				textContent, ok := result.Content[0].(mcp_go.TextContent)
				if !ok {
					t.Fatal("Expected TextContent")
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if status := response["status"]; status != "copied" {
					t.Errorf("Expected status 'copied', got %v", status)
				}
				if source := response["source"]; source != "llama2:latest" {
					t.Errorf("Expected source 'llama2:latest', got %v", source)
				}
			},
		},
		{
			desc: "Missing source parameter",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"destination": "llama2:custom",
					},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    true,
		},
		{
			desc: "Missing destination parameter",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"source": "llama2:latest",
					},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    true,
		},
		{
			desc: "Copy error",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"source":      "nonexistent:latest",
						"destination": "llama2:custom",
					},
				},
			},
			mockClient: &mockOllamaClient{
				copyError: errors.New("source model not found"),
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				OllamaClient: tc.mockClient,
			}

			result, err := s.CopyModel(ctx, tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("CopyModel() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && tc.validateResp != nil {
				tc.validateResp(t, result)
			}
		})
	}
}

func TestEmbeddings(t *testing.T) {
	testCases := []struct {
		desc         string
		request      mcp_go.CallToolRequest
		mockClient   *mockOllamaClient
		wantErr      bool
		validateResp func(*testing.T, *mcp_go.CallToolResult)
	}{
		{
			desc: "Successfully generate embeddings",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "llama2:latest",
						"input": "Hello, world!",
					},
				},
			},
			mockClient: &mockOllamaClient{
				embeddingsResponse: &api.EmbeddingResponse{
					Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, result *mcp_go.CallToolResult) {
				textContent, ok := result.Content[0].(mcp_go.TextContent)
				if !ok {
					t.Fatal("Expected TextContent")
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if dimensions, ok := response["dimensions"].(float64); !ok || dimensions != 5 {
					t.Errorf("Expected dimensions 5, got %v", response["dimensions"])
				}
			},
		},
		{
			desc: "Missing model parameter",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"input": "Hello, world!",
					},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    true,
		},
		{
			desc: "Missing input parameter",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "llama2:latest",
					},
				},
			},
			mockClient: &mockOllamaClient{},
			wantErr:    true,
		},
		{
			desc: "Embeddings error",
			request: mcp_go.CallToolRequest{
				Params: mcp_go.CallToolParams{
					Arguments: map[string]interface{}{
						"model": "llama2:latest",
						"input": "Hello, world!",
					},
				},
			},
			mockClient: &mockOllamaClient{
				embeddingsError: errors.New("model does not support embeddings"),
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				OllamaClient: tc.mockClient,
			}

			result, err := s.Embeddings(ctx, tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("Embeddings() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && tc.validateResp != nil {
				tc.validateResp(t, result)
			}
		})
	}
}
