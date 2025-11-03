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
	commandRunner commands.Runner
}

// SetCommandRunner sets the command runner
func (l *ListCommands) SetCommandRunner(runner commands.Runner) {
	l.commandRunner = runner
}

// Definition returns the mcp.Tool definition.
func (l *ListCommands) Definition() mcp.Tool {
	return mcp.NewTool("list_commands",
		mcp.WithDescription("Lists all background commands with their current status (id, status, command, hosts, created_at, started_at, ended_at). Use get_command_status to see detailed results for a specific command."),
		mcp.WithString("status", mcp.Description("Optional filter by command status (pending, running, completed, failed, cancelled)")),
	)
}

// Handler is the function that is called when the tool is invoked.
func (l *ListCommands) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(reqCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if l.commandRunner == nil {
			panic("command runner not available")
		}

		// Get optional status filter
		statusFilter := request.GetString("status", "")
		var filterStatus commands.CommandStatus
		if statusFilter != "" {
			// Validate the status filter
			filterStatus = commands.CommandStatus(statusFilter)
			switch filterStatus {
			case commands.CommandStatusPending, commands.CommandStatusRunning,
				commands.CommandStatusCompleted, commands.CommandStatusFailed,
				commands.CommandStatusCancelled:
				// Valid status
			default:
				return mcp.NewToolResultError("invalid status filter: must be one of pending, running, completed, failed, cancelled"), nil
			}
		}

		allCommands := l.commandRunner.ListCommands()
		if len(allCommands) == 0 {
			return mcp.NewToolResultText("No commands found"), nil
		}

		// Convert to list items (without results) and apply filter
		commandList := make([]*commands.CommandListItem, 0, len(allCommands))
		for _, cmd := range allCommands {
			listItem := cmd.ToListItem()

			// Apply status filter if provided
			if statusFilter == "" || listItem.Status == filterStatus {
				commandList = append(commandList, listItem)
			}
		}

		// Check if any commands match the filter
		if len(commandList) == 0 {
			if statusFilter != "" {
				return mcp.NewToolResultText("No commands found with status: " + statusFilter), nil
			}
			return mcp.NewToolResultText("No commands found"), nil
		}

		// Sort commands by creation time (newest first)
		sort.Slice(commandList, func(i, j int) bool {
			return commandList[i].CreatedAt.After(commandList[j].CreatedAt)
		})

		return mcp.NewToolResultStructuredOnly(map[string]any{"commands": commandList}), nil
	}
}
