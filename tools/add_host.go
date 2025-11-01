package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
	"github.com/blakerouse/ssh-mcp/utils"
)

func init() {
	// register the tool in the registry
	Registry.Register(&AddHost{})
}

// AddHost is a tool that adds a new host to the SSH configuration.
type AddHost struct{}

// Definition returns the mcp.Tool definition.
func (c *AddHost) Definition() mcp.Tool {
	return mcp.NewTool("add_host",
		mcp.WithDescription("Adds a new Linux or Windows host to the SSH configuration with automatic OS detection. Username and password are optional in the connection string - if not provided, the current user and SSH agent will be used for authentication."),
		mcp.WithString("group",
			mcp.Required(),
			mcp.Description("Group that the host belongs to"),
		),
		mcp.WithString("ssh_connection_string",
			mcp.Required(),
			mcp.Description("SSH connection string in format: [user[:password]@]host[:port]. The 'ssh://' prefix is optional. Examples: server.com, user@server.com, user:pass@server.com:2222, ssh://user@server.com:2222"),
		),
		mcp.WithString("name_of_host",
			mcp.Description("Name of the host (optional, defaults to hostname)"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *AddHost) Handler(storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		group, err := request.RequireString("group")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Validate that group is not empty
		if group == "" {
			return mcp.NewToolResultError("group cannot be empty"), nil
		}

		sshConnectionString, err := request.RequireString("ssh_connection_string")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sshNameOfHost := request.GetString("name_of_host", "")

		clientInfo, err := ssh.NewClientInfo(sshNameOfHost, sshConnectionString)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Set the group
		clientInfo.Group = group

		sshClient := ssh.NewClient(clientInfo)

		// connect over ssh
		err = sshClient.Connect()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer sshClient.Close()

		// Detect OS and gather system information (supports Linux and Windows)
		osRelease, uname, err := utils.GatherOSInfo(sshClient)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to gather OS information: %w", err).Error()), nil
		}

		// set the OS info and store it for usage later
		clientInfo.OS.OSRelease = osRelease
		clientInfo.OS.Uname = uname
		err = storageEngine.Set(*clientInfo)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to add host to storage: %w", err).Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("successfully added %s to group %s", clientInfo.Name, group)), nil
	}
}
