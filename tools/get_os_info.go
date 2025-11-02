package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
	"github.com/blakerouse/ssh-mcp/utils"
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
		mcp.WithDescription("Retrieves the cached operating system information for Linux and Windows hosts. You can specify individual hosts or an entire group."),
		mcp.WithString("group",
			mcp.Description("Group name to get OS info for all hosts in that group (mutually exclusive with name_of_hosts)"),
		),
		mcp.WithArray("name_of_hosts",
			mcp.Description("Array of host identifiers in format 'group:name' (mutually exclusive with group)"),
			mcp.WithStringItems(),
		),
	)
}

// Handle is the function that is called when the tool is invoked.
func (c *GetOSInfo) Handler(ctx context.Context, storageEngine *storage.Engine) server.ToolHandlerFunc {
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

		return mcp.NewToolResultStructuredOnly(found), nil
	}
}
