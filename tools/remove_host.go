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
		mcp.WithString("group",
			mcp.Required(),
			mcp.Description("Group that the host belongs to"),
		),
		mcp.WithString("name_of_host",
			mcp.Required(),
			mcp.Description("Name of the host"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *RemoveHost) Handler(storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		group, err := request.RequireString("group")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Validate that group is not empty
		if group == "" {
			return mcp.NewToolResultError("group cannot be empty"), nil
		}

		sshNameOfHost, err := request.RequireString("name_of_host")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Validate that name is not empty
		if sshNameOfHost == "" {
			return mcp.NewToolResultError("name_of_host cannot be empty"), nil
		}

		// check if its existed first so we change change the resulting output depending
		// on its existance
		_, ok := storageEngine.Get(group, sshNameOfHost)
		err = storageEngine.Delete(group, sshNameOfHost)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to remove host from storage: %w", err).Error()), nil
		}

		if ok {
			return mcp.NewToolResultText(fmt.Sprintf("successfully removed %s from group %s", sshNameOfHost, group)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("host %s in group %s not found", sshNameOfHost, group)), nil
	}
}
