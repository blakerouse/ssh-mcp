package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/commands"
	"github.com/blakerouse/ssh-mcp/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&CancelCommand{})
}

// CancelCommand is a tool that cancels a running background command.
type CancelCommand struct {
	commandRunner commands.Runner
}

// SetCommandRunner sets the command runner
func (c *CancelCommand) SetCommandRunner(runner commands.Runner) {
	c.commandRunner = runner
}

// Definition returns the mcp.Tool definition.
func (c *CancelCommand) Definition() mcp.Tool {
	return mcp.NewTool("cancel_command",
		mcp.WithDescription("Cancels a running background command by its command ID."),
		mcp.WithString("command_id", mcp.Required(), mcp.Description("The command ID of the running command to cancel")),
	)
}

// Handler is the function that is called when the tool is invoked.
func (c *CancelCommand) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(reqCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if c.commandRunner == nil {
			panic("command runner not available")
		}
		commandID, err := request.RequireString("command_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		cmd, err := c.commandRunner.GetCommand(commandID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		err = cmd.Cancel()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Command %s has been cancelled", commandID)), nil
	}
}
