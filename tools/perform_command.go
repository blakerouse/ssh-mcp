package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/commands"
	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
	"github.com/blakerouse/ssh-mcp/utils"
)

func init() {
	// register the tool in the registry
	Registry.Register(&PerformCommand{})
}

// PerformCommand is a tool that executes a command on a remote machine.
type PerformCommand struct {
	commandRunner *commands.Runner
}

// SetCommandRunner sets the command runner for background execution
func (c *PerformCommand) SetCommandRunner(runner *commands.Runner) {
	c.commandRunner = runner
}

// Definition returns the mcp.Tool definition.
func (c *PerformCommand) Definition() mcp.Tool {
	return mcp.NewTool("perform_command",
		mcp.WithDescription("SSH into a remote machine and executes a command. You can specify individual hosts or an entire group. Commands that take longer than 30 seconds are automatically moved to background. Use background=true to run immediately in background. For background commands, use get_command_status to poll for progress and see partial output snapshots."),
		mcp.WithString("group",
			mcp.Description("Group name to execute command on all hosts in that group (mutually exclusive with name_of_hosts)"),
		),
		mcp.WithArray("name_of_hosts",
			mcp.Description("Array of host identifiers in format 'group:name' (mutually exclusive with group)"),
			mcp.WithStringItems(),
		),
		mcp.WithString("command", mcp.Required(), mcp.Description("The command to execute")),
		mcp.WithBoolean("background",
			mcp.Description("Run the command in the background immediately and return a command ID (default: false, waits up to 30s before auto-backgrounding)"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *PerformCommand) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(reqCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if c.commandRunner == nil {
			panic("command runner not available")
		}

		commandStr, err := request.RequireString("command")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get hosts either by group or by individual host identifiers
		var found []ssh.ClientInfo
		group := request.GetString("group", "")
		sshNameOfHosts := request.GetStringSlice("name_of_hosts", []string{})
		if group != "" && len(sshNameOfHosts) > 0 {
			return mcp.NewToolResultError("cannot specify both 'group' and 'name_of_hosts'"), nil
		}

		if group != "" {
			found, err = utils.GetHostsFromGroup(storageEngine, group)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		} else if len(sshNameOfHosts) > 0 {
			identifiers, err := utils.ParseHostIdentifiers(sshNameOfHosts)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			found, err = utils.GetHostsFromStorage(storageEngine, identifiers)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		} else {
			return mcp.NewToolResultError("must specify either 'group' or 'name_of_hosts'"), nil
		}

		if len(found) == 0 {
			return mcp.NewToolResultError("no matching hosts found"), nil
		}

		// Create and start the command
		cmd := c.commandRunner.CreateCommand(commandStr, found)
		err = cmd.Start()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to start command: %v", err)), nil
		}

		// If background execution is requested, return immediately
		if request.GetBool("background", false) {
			return mcp.NewToolResultStructured(cmd.ToState(), fmt.Sprintf("Command started in background with ID: %s\nUse get_command_status tool to check progress.", cmd.ID())), nil
		}

		// Wait for command completion with 30 second timeout
		return c.waitForCommandOrBackground(reqCtx, cmd)
	}
}

// waitForCommandOrBackground waits up to 30 seconds for a command to complete.
// If it completes in time, returns the results. Otherwise, returns the command ID for background tracking.
// If the context is cancelled, returns the command ID immediately.
func (c *PerformCommand) waitForCommandOrBackground(ctx context.Context, cmd *commands.Command) (*mcp.CallToolResult, error) {
	const timeout = 30
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return mcp.NewToolResultError("request cancelled"), nil
		case <-ticker.C:
			if cmd.Status() == commands.CommandStatusCompleted ||
				cmd.Status() == commands.CommandStatusFailed ||
				cmd.Status() == commands.CommandStatusCancelled ||
				time.Since(startTime) >= timeout*time.Second {
				return mcp.NewToolResultStructuredOnly(cmd.ToState()), nil
			}
		}
	}
}
