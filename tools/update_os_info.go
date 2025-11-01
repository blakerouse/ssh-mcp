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
	Registry.Register(&UpdateOSInfo{})
}

// UpdateOSInfo is a tool that updates the operating system information on a remote machine.
type UpdateOSInfo struct{}

// Definition returns the mcp.Tool definition.
func (c *UpdateOSInfo) Definition() mcp.Tool {
	return mcp.NewTool("update_os_info",
		mcp.WithDescription("Updates the cached operating system information."),
		mcp.WithArray("name_of_hosts",
			mcp.Required(),
			mcp.Description("Name of the hosts"),
			mcp.WithStringItems(),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *UpdateOSInfo) Handler(storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sshNameOfHosts, err := request.RequireStringSlice("name_of_hosts")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(sshNameOfHosts) == 0 {
			return mcp.NewToolResultError("no hosts provided"), nil
		}

		found, err := getHostsFromStorage(storageEngine, sshNameOfHosts)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(found) == 0 {
			return mcp.NewToolResultError("no matching hosts found"), nil
		}

		// from this point forward it is very much assuming linux
		// this really should be improved to do more checks to see if this macOS or Windows

		result := performTasksOnHosts(found, func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
			osRelease, err := sshClient.Exec("cat /etc/os-release")
			if err != nil {
				return "", fmt.Errorf("failed to get output of /etc/os-release: %w", err)
			}
			uname, err := sshClient.Exec("uname -a")
			if err != nil {
				return "", fmt.Errorf("failed to get output of uname -a: %w", err)
			}

			// set the OS info and store it for usage later
			host.OS.OSRelease = string(osRelease)
			host.OS.Uname = string(uname)
			err = storageEngine.Set(host)
			if err != nil {
				return "", fmt.Errorf("failed to add host to storage: %w", err)
			}
			return fmt.Sprintf("successfully updated %s", host.Name), nil
		})

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}
