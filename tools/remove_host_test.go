package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// Tests for RemoveHost tool

func TestRemoveHost_Success(t *testing.T) {
	engine := setupTestStorage(t)

	// Add a host first
	addTestHost(t, engine, "production", "server1", "10.0.1.1")

	tool := &RemoveHost{}
	handler := tool.Handler(engine)

	// Remove the host
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group":        "production",
				"name_of_host": "server1",
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)

	// Verify the host was actually removed
	_, ok := engine.Get("production", "server1")
	require.False(t, ok)
}

func TestRemoveHost_NonexistentHost(t *testing.T) {
	engine := setupTestStorage(t)

	tool := &RemoveHost{}
	handler := tool.Handler(engine)

	// Try to remove a host that doesn't exist
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group":        "production",
				"name_of_host": "nonexistent",
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should not be an error, just indicates host not found
	require.False(t, result.IsError)
}

func TestRemoveHost_EmptyGroup(t *testing.T) {
	engine := setupTestStorage(t)

	tool := &RemoveHost{}
	handler := tool.Handler(engine)

	// Try to remove with empty group
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group":        "",
				"name_of_host": "server1",
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return an error
	require.True(t, result.IsError)
}

func TestRemoveHost_EmptyName(t *testing.T) {
	engine := setupTestStorage(t)

	tool := &RemoveHost{}
	handler := tool.Handler(engine)

	// Try to remove with empty name
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group":        "production",
				"name_of_host": "",
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return an error
	require.True(t, result.IsError)
}

func TestRemoveHost_MissingGroup(t *testing.T) {
	engine := setupTestStorage(t)

	tool := &RemoveHost{}
	handler := tool.Handler(engine)

	// Try to remove without specifying group
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"name_of_host": "server1",
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return an error
	require.True(t, result.IsError)
}

func TestRemoveHost_MissingName(t *testing.T) {
	engine := setupTestStorage(t)

	tool := &RemoveHost{}
	handler := tool.Handler(engine)

	// Try to remove without specifying name
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
	// Should return an error
	require.True(t, result.IsError)
}

func TestRemoveHost_GroupIsolation(t *testing.T) {
	engine := setupTestStorage(t)

	// Add hosts with same name in different groups
	addTestHost(t, engine, "production", "server1", "10.0.1.1")
	addTestHost(t, engine, "staging", "server1", "10.0.2.1")

	tool := &RemoveHost{}
	handler := tool.Handler(engine)

	// Remove from production only
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group":        "production",
				"name_of_host": "server1",
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)

	// Verify production server1 is removed
	_, ok := engine.Get("production", "server1")
	require.False(t, ok)

	// Verify staging server1 still exists
	_, ok = engine.Get("staging", "server1")
	require.True(t, ok)
}

func TestRemoveHost_MultipleRemoves(t *testing.T) {
	engine := setupTestStorage(t)

	// Add multiple hosts
	addTestHost(t, engine, "production", "server1", "10.0.1.1")
	addTestHost(t, engine, "production", "server2", "10.0.1.2")
	addTestHost(t, engine, "production", "server3", "10.0.1.3")

	tool := &RemoveHost{}
	handler := tool.Handler(engine)

	// Remove server1
	request1 := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group":        "production",
				"name_of_host": "server1",
			},
		},
	}
	result1, err := handler(context.Background(), request1)
	require.NoError(t, err)
	require.False(t, result1.IsError)

	// Remove server2
	request2 := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group":        "production",
				"name_of_host": "server2",
			},
		},
	}
	result2, err := handler(context.Background(), request2)
	require.NoError(t, err)
	require.False(t, result2.IsError)

	// Verify only server3 remains
	_, ok := engine.Get("production", "server1")
	require.False(t, ok)
	_, ok = engine.Get("production", "server2")
	require.False(t, ok)
	_, ok = engine.Get("production", "server3")
	require.True(t, ok)
}
