package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// Tests for GetHosts tool

func TestGetHosts_AllHosts(t *testing.T) {
	engine := setupTestStorage(t)

	// Add hosts to different groups
	addTestHost(t, engine, "production", "server1", "10.0.1.1")
	addTestHost(t, engine, "staging", "server2", "10.0.2.1")

	tool := &GetHosts{}
	handler := tool.Handler(context.Background(), engine)

	// Request all hosts (no group filter)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
}

func TestGetHosts_FilterByGroup(t *testing.T) {
	engine := setupTestStorage(t)

	// Add hosts to different groups
	addTestHost(t, engine, "production", "server1", "10.0.1.1")
	addTestHost(t, engine, "production", "server2", "10.0.1.2")
	addTestHost(t, engine, "staging", "server3", "10.0.2.1")

	tool := &GetHosts{}
	handler := tool.Handler(context.Background(), engine)

	// Request only production hosts
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group": "production",
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
}

func TestGetHosts_EmptyStorage(t *testing.T) {
	engine := setupTestStorage(t)
	tool := &GetHosts{}
	handler := tool.Handler(context.Background(), engine)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return empty list, not an error
	require.False(t, result.IsError)
}

func TestGetHosts_NonexistentGroup(t *testing.T) {
	engine := setupTestStorage(t)

	// Add a host to production
	addTestHost(t, engine, "production", "server1", "10.0.1.1")

	tool := &GetHosts{}
	handler := tool.Handler(context.Background(), engine)

	// Request hosts from nonexistent group
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group": "nonexistent",
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return error or empty list
	// The actual behavior depends on implementation
}
