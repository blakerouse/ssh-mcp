package tools

import (
	"github.com/blakerouse/ssh-mcp/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Tool defines the interface that provides both the definition and the handler for a tool.
type Tool interface {
	Definition() mcp.Tool
	Handler(*storage.Engine) server.ToolHandlerFunc
}
