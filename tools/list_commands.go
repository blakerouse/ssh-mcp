package tools

import (
	"context"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/commands"
	"github.com/blakerouse/ssh-mcp/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&ListCommands{})
}

// ListCommands is a tool that lists all background commands.
type ListCommands struct {
	commandRunner *commands.Runner
}

// SetCommandRunner sets the command runner
func (l *ListCommands) SetCommandRunner(runner *commands.Runner) {
	l.commandRunner = runner
}

// Definition returns the mcp.Tool definition.
func (l *ListCommands) Definition() mcp.Tool {
	return mcp.NewTool("list_commands",
		mcp.WithDescription("Lists all background commands with their current status (id, status, command, hosts, created_at, started_at, ended_at). Use get_command_status to see detailed results for a specific command."),
	)
}

// Handler is the function that is called when the tool is invoked.
func (l *ListCommands) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(reqCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if l.commandRunner == nil {
			panic("command runner not available")
		}

		allCommands := l.commandRunner.ListCommands()
		if len(allCommands) == 0 {
			return mcp.NewToolResultText("No commands found"), nil
		}

		// Convert to list items (without results)
		commandList := make([]*commands.CommandListItem, len(allCommands))
		for i, cmd := range allCommands {
			commandList[i] = cmd.ToListItem()
		}

		// Sort commands by creation time (newest first)
		sort.Slice(commandList, func(i, j int) bool {
			return commandList[i].CreatedAt.After(commandList[j].CreatedAt)
		})

		return mcp.NewToolResultStructuredOnly(commandList), nil
	}
}
