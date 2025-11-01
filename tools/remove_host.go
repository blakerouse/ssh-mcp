package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&RemoveHost{})
}

// RemoveHost is a tool that removes a host from the SSH configuration.
type RemoveHost struct{}

// Definition returns the mcp.Tool definition.
func (c *RemoveHost) Definition() mcp.Tool {
	return mcp.NewTool("remove_host",
		mcp.WithDescription("Removes a host from the SSH configuration."),
		mcp.WithString("name_of_host",
			mcp.Required(),
			mcp.Description("Name of the host"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *RemoveHost) Handler(storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sshNameOfHost, err := request.RequireString("name_of_host")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// check if its existed first so we change change the resulting output depending
		// on its existance
		_, ok := storageEngine.Get(sshNameOfHost)
		err = storageEngine.Delete(sshNameOfHost)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to remove host from storage: %w", err).Error()), nil
		}

		if ok {
			return mcp.NewToolResultText(fmt.Sprintf("successfully removed %s", sshNameOfHost)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("host %s not found", sshNameOfHost)), nil
	}
}
