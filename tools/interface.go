package tools

import (
	"context"

	"github.com/blakerouse/ssh-mcp/commands"
	"github.com/blakerouse/ssh-mcp/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Tool defines the interface that provides both the definition and the handler for a tool.
type Tool interface {
	Definition() mcp.Tool
	Handler(ctx context.Context, engine *storage.Engine) server.ToolHandlerFunc
}

// CommandRunnerAware is an optional interface that tools can implement to support background command execution.
type CommandRunnerAware interface {
	Tool

	// SetCommandRunner sets the command runner for background execution.
	SetCommandRunner(runner commands.Runner)
}
