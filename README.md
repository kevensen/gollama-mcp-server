# gollama-mcp-server

An MCP (Model Context Protocol) server written in Golang to interact with the Ollama API. This server provides tools for AI assistants to interact with local Ollama models through the MCP protocol.

## Features

This MCP server provides the following tools for interacting with Ollama:

- **listModels** - List all available models in your Ollama instance
- **pullModel** - Download models from the Ollama library
- **pushModel** - Push models to a registry
- **deleteModel** - Remove models from your local instance
- **showModel** - View detailed model information including modelfile, parameters, and template
- **copyModel** - Create a copy of a model with a different name
- **embeddings** - Generate vector embeddings for text

## Prerequisites

- Go 1.24.4 or later
- Ollama installed and running locally (or accessible via network)

## Configuration

The server reads the Ollama host from the `OLLAMA_HOST` environment variable. If not set, it defaults to `http://localhost:11434`.

Example values:
```bash
export OLLAMA_HOST=http://localhost:11434
export OLLAMA_HOST=http://192.168.0.185:11434
export OLLAMA_HOST=http://spark.lan:11434
```

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd gollama-mcp-server

# Build the binary
go build -o bin/gollama-mcp-server ./cmd/main.go
```

## Usage

### STDIO Mode (for MCP clients)

Run the server in STDIO mode for use with MCP-compatible clients:

```bash
./bin/gollama-mcp-server
```

### HTTP Mode (Streamable HTTP)

Run the server with HTTP transport on a specific port:

```bash
./bin/gollama-mcp-server -port 8080
```

You can also specify a custom host:

```bash
./bin/gollama-mcp-server -host 127.0.0.1 -port 8080
```

### Docker

#### Building the Docker Image

Build the Docker image from the repository root:

```bash
docker build -t gollama-mcp-server .
```

#### Running with Docker

Run the server in HTTP mode (port 8080):

```bash
docker run -p 8080:8080 gollama-mcp-server
```

If your Ollama instance is running on the host machine, use host networking:

```bash
# Linux
docker run --network host gollama-mcp-server

# macOS/Windows - use host.docker.internal
docker run -p 8080:8080 -e OLLAMA_HOST=http://host.docker.internal:11434 gollama-mcp-server
```

To connect to a remote Ollama instance:

```bash
docker run -p 8080:8080 -e OLLAMA_HOST=http://your-ollama-host:11434 gollama-mcp-server
```

Run in detached mode with automatic restart:

```bash
docker run -d --restart unless-stopped \
  -p 8080:8080 \
  -e OLLAMA_HOST=http://host.docker.internal:11434 \
  --name gollama-mcp \
  gollama-mcp-server
```

## Tool Usage Examples

### List Models
```json
{
  "name": "listModels"
}
```

### Generate Completion
```json
{
  "name": "generate",
  "arguments": {
    "model": "llama3.2",
    "prompt": "Write a haiku about programming",
    "temperature": 0.7,
    "stream": false
  }
}
```

### Chat
```json
{
  "name": "chat",
  "arguments": {
    "model": "llama3.2",
    "messages": [
      {"role": "user", "content": "Hello! How are you?"}
    ],
    "temperature": 0.8,
    "stream": false
  }
}
```

### Pull Model
```json
{
  "name": "pullModel",
  "arguments": {
    "model": "llama3.2",
    "stream": true
  }
}
```

### Generate Embeddings
```json
{
  "name": "embeddings",
  "arguments": {
    "model": "nomic-embed-text",
    "input": "The quick brown fox jumps over the lazy dog"
  }
}
```

## Development

### Project Structure

```
gollama-mcp-server/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   └── handlers/
│       ├── handlers.go      # HTTP routing and handlers
│       └── mcp/
│           ├── client.go    # Ollama client wrapper
│           ├── errors.go    # Error types
│           ├── mcp.go       # MCP initialization
│           ├── server.go    # MCP server setup and tool registration
│           └── tools.go     # Tool implementations
├── bin/                     # Compiled binaries
├── go.mod
├── go.work
└── README.md
```

### Running Tests

```bash
go test ./...
```

### Adding New Tools

1. Add the tool registration in [server.go](internal/handlers/mcp/server.go) in the `registerTools()` method
2. Implement the tool handler in [tools.go](internal/handlers/mcp/tools.go)
3. Follow the existing patterns for error handling and response formatting

## Dependencies

- [github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP protocol implementation
- [github.com/ollama/ollama](https://github.com/ollama/ollama) - Ollama API client
- [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync) - Synchronization utilities

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
