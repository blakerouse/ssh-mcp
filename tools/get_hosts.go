package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
)

func init() {
	// register the tool in the registry
	Registry.Register(&GetHosts{})
}

// GetHosts is a tool that retrieves the list of hosts from the SSH configuration.
type GetHosts struct{}

// Definition returns the mcp.Tool definition.
func (c *GetHosts) Definition() mcp.Tool {
	return mcp.NewTool("get_hosts",
		mcp.WithDescription("Retrieves the list of hosts from the SSH configuration. Can optionally filter by group."),
		mcp.WithString("group",
			mcp.Description("Optional group name to filter hosts by"),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *GetHosts) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
	return func(reqCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		group := request.GetString("group", "")

		var hosts []ssh.ClientInfo
		var err error

		if group != "" {
			hosts, err = storageEngine.ListGroup(group)
			if err != nil {
				return mcp.NewToolResultError(fmt.Errorf("failed to list hosts in group %s: %w", group, err).Error()), nil
			}
		} else {
			hosts, err = storageEngine.List()
			if err != nil {
				return mcp.NewToolResultError(fmt.Errorf("failed to list hosts: %w", err).Error()), nil
			}
		}

		list := make([]string, 0, len(hosts))
		for _, host := range hosts {
			list = append(list, fmt.Sprintf("%s:%s", host.Group, host.Name))
		}
		return mcp.NewToolResultStructured(hosts, strings.Join(list, ", ")), nil
	}
}
