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
	Registry.Register(&AddHost{})
}

// AddHost is a tool that adds a new host to the SSH configuration.
type AddHost struct{}

// Definition returns the mcp.Tool definition.
func (c *AddHost) Definition() mcp.Tool {
	return mcp.NewTool("add_host",
		mcp.WithDescription("Adds a new host to the SSH configuration."),
		mcp.WithString("ssh_connection_string",
			mcp.Required(),
			mcp.Description("SSH connection string"),
		),
		mcp.WithString("name_of_host",
			mcp.Description("Name of the host"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *AddHost) Handler(storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sshConnectionString, err := request.RequireString("ssh_connection_string")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sshNameOfHost := request.GetString("name_of_host", "")

		clientInfo, err := ssh.NewClientInfo(sshNameOfHost, sshConnectionString)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sshClient := ssh.NewClient(clientInfo)

		// connect over ssh
		err = sshClient.Connect()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer sshClient.Close()

		// from this point forward it is very much assuming linux
		// this really should be improved to do more checks to see if this macOS or Windows

		osRelease, err := sshClient.Exec("cat /etc/os-release")
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to get output of /etc/os-release: %w", err).Error()), nil
		}
		uname, err := sshClient.Exec("uname -a")
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to get output of uname -a: %w", err).Error()), nil
		}

		// set the OS info and store it for usage later
		clientInfo.OS.OSRelease = string(osRelease)
		clientInfo.OS.Uname = string(uname)
		err = storageEngine.Set(*clientInfo)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to add host to storage: %w", err).Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("successfully added %s", clientInfo.Name)), nil
	}
}
