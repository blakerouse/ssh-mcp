package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/commands"
	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
	"github.com/blakerouse/ssh-mcp/utils"
)

func init() {
	// register the tool in the registry
	Registry.Register(&UpdateOSInfo{})
}

// UpdateOSInfo is a tool that updates the operating system information on a remote machine.
type UpdateOSInfo struct{}

// Definition returns the mcp.Tool definition.
func (c *UpdateOSInfo) Definition() mcp.Tool {
	return mcp.NewTool("update_os_info",
		mcp.WithDescription("Updates the cached operating system information for Linux and Windows hosts. You can specify individual hosts or an entire group."),
		mcp.WithString("group",
			mcp.Description("Group name to update OS info for all hosts in that group (mutually exclusive with name_of_hosts)"),
		),
		mcp.WithArray("name_of_hosts",
			mcp.Description("Array of host identifiers in format 'group:name' (mutually exclusive with group)"),
			mcp.WithStringItems(),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *UpdateOSInfo) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(reqCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get hosts either by group or by individual host identifiers
		var found []ssh.ClientInfo
		var err error
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

		// Detect OS and gather system information (supports Linux and Windows)
		result := commands.PerformOnHosts(found, func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
			osRelease, uname, err := utils.GatherOSInfo(sshClient)
			if err != nil {
				return "", fmt.Errorf("failed to gather OS information: %w", err)
			}

			// set the OS info and store it for usage later
			host.OS.OSRelease = osRelease
			host.OS.Uname = uname
			err = storageEngine.Set(host)
			if err != nil {
				return "", fmt.Errorf("failed to add host to storage: %w", err)
			}
			return fmt.Sprintf("successfully updated %s", host.Name), nil
		})

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}
