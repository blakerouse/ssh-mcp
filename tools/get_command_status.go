package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/commands"
	"github.com/blakerouse/ssh-mcp/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&GetCommandStatus{})
}

// GetCommandStatus is a tool that retrieves the status and results of a background command.
type GetCommandStatus struct {
	commandRunner *commands.Runner
}

// SetCommandRunner sets the command runner
func (g *GetCommandStatus) SetCommandRunner(runner *commands.Runner) {
	g.commandRunner = runner
}

// Definition returns the mcp.Tool definition.
func (g *GetCommandStatus) Definition() mcp.Tool {
	return mcp.NewTool("get_command_status",
		mcp.WithDescription("Retrieves the status and results of a background command by its command ID. For running commands, returns a snapshot of the partial output captured so far. Poll this repeatedly to monitor progress of long-running commands. If no ID is provided, returns the most recent command."),
		mcp.WithString("command_id", mcp.Description("The command ID returned when starting a background command (optional - defaults to most recent command)")),
	)
}

// Handler is the function that is called when the tool is invoked.
func (g *GetCommandStatus) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(reqCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if g.commandRunner == nil {
			panic("command runner not available")
		}

		var cmd *commands.Command
		var err error

		commandID := request.GetString("command_id", "")
		if commandID == "" {
			cmd, err = g.commandRunner.GetMostRecentCommand()
		} else {
			cmd, err = g.commandRunner.GetCommand(commandID)
		}
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultStructuredOnly(cmd.ToState()), nil
	}
}
