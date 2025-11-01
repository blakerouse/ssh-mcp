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
		mcp.WithDescription("SSH into a remote machine and executes a command. You can specify individual hosts or an entire group."),
		mcp.WithString("group",
			mcp.Description("Group name to execute command on all hosts in that group (mutually exclusive with name_of_hosts)"),
		),
		mcp.WithArray("name_of_hosts",
			mcp.Description("Array of host identifiers in format 'group:name' (mutually exclusive with group)"),
			mcp.WithStringItems(),
		),
		mcp.WithString("command", mcp.Required(), mcp.Description("The command to execute")),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *PerformCommand) Handler(storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			found, err = getHostsFromGroup(storageEngine, group)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		} else if len(sshNameOfHosts) > 0 {
			identifiers, err := parseHostIdentifiers(sshNameOfHosts)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			found, err = getHostsFromStorage(storageEngine, identifiers)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		} else {
			return mcp.NewToolResultError("must specify either 'group' or 'name_of_hosts'"), nil
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
