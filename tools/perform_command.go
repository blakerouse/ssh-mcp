package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&PerformCommand{})
}

// PerformCommand is a tool that executes a command on a remote machine.
type PerformCommand struct{}

// Definition returns the mcp.Tool definition.
func (c *PerformCommand) Definition() mcp.Tool {
	return mcp.NewTool("perform_command",
		mcp.WithDescription("SSH into a remote machine and executes a command."),
		mcp.WithArray("name_of_hosts",
			mcp.Required(),
			mcp.Description("Name of the hosts"),
			mcp.WithStringItems(),
		),
		mcp.WithString("command", mcp.Required(), mcp.Description("The command to execute")),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *PerformCommand) Handler(storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sshNameOfHosts, err := request.RequireStringSlice("name_of_hosts")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(sshNameOfHosts) == 0 {
			return mcp.NewToolResultError("no hosts provided"), nil
		}
		commandStr, err := request.RequireString("command")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		found, err := getHostsFromStorage(storageEngine, sshNameOfHosts)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(found) == 0 {
			return mcp.NewToolResultError("no matching hosts found"), nil
		}

		result := performTasksOnHosts(found, func(_ ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
			// sudo is required to update and upgrade
			output, err := sshClient.Exec(commandStr)
			if err != nil {
				return "", fmt.Errorf("failed to execute command: %w", err)
			}
			return string(output), nil
		})

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}
