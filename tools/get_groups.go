package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&GetGroups{})
}

// GetGroups is a tool that retrieves the list of groups from the SSH configuration.
type GetGroups struct{}

// Definition returns the mcp.Tool definition.
func (c *GetGroups) Definition() mcp.Tool {
	return mcp.NewTool("get_groups",
		mcp.WithDescription("Retrieves the list of all groups from the SSH configuration."),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *GetGroups) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(reqCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		groups, err := storageEngine.ListGroups()
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("failed to list groups: %w", err).Error()), nil
		}

		return mcp.NewToolResultStructured(map[string]any{"groups": groups}, strings.Join(groups, ", ")), nil
	}
}
