package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&GetOSInfo{})
}

// GetOSInfo is a tool that retrieves the operating system information from a remote machine.
type GetOSInfo struct{}

// Definition returns the mcp.Tool definition.
func (c *GetOSInfo) Definition() mcp.Tool {
	return mcp.NewTool("get_os_info",
		mcp.WithDescription("Retrieves the operating system information."),
		mcp.WithArray("name_of_hosts",
			mcp.Required(),
			mcp.Description("Name of the hosts"),
			mcp.WithStringItems(),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *GetOSInfo) Handler(storageEngine *storage.Engine) server.ToolHandlerFunc {
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

		return mcp.NewToolResultStructuredOnly(found), nil
	}
}
