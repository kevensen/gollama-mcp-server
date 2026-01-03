package mcp

import (
	"github.com/kevensen/gollama-mcp-server/internal/handlers"
	"github.com/mark3labs/mcp-go/server"
)

func init() {
	srv := NewServer()
	httpServer := server.NewStreamableHTTPServer(srv.MCPServer)
	handlers.AddAll("/mcp", srv, httpServer.ServeHTTP)
	srv.ready = true
}
