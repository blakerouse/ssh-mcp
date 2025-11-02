package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// Tests for GetOSInfo tool

func TestGetOSInfo_ByGroup(t *testing.T) {
	engine := setupTestStorage(t)

	// Add hosts to groups
	addTestHost(t, engine, "production", "server1", "10.0.1.1")
	addTestHost(t, engine, "production", "server2", "10.0.1.2")
	addTestHost(t, engine, "staging", "server3", "10.0.2.1")

	tool := &GetOSInfo{}
	handler := tool.Handler(context.Background(), engine)

	// Request OS info for production group
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

func TestGetOSInfo_ByHostIdentifiers(t *testing.T) {
	engine := setupTestStorage(t)

	// Add hosts
	addTestHost(t, engine, "production", "server1", "10.0.1.1")
	addTestHost(t, engine, "staging", "server2", "10.0.2.1")

	tool := &GetOSInfo{}
	handler := tool.Handler(context.Background(), engine)

	// Request OS info for specific hosts
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"name_of_hosts": []interface{}{
					"production:server1",
					"staging:server2",
				},
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
}

func TestGetOSInfo_MutuallyExclusiveParams(t *testing.T) {
	engine := setupTestStorage(t)

	addTestHost(t, engine, "production", "server1", "10.0.1.1")

	tool := &GetOSInfo{}
	handler := tool.Handler(context.Background(), engine)

	// Try to specify both group and name_of_hosts
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"group": "production",
				"name_of_hosts": []interface{}{
					"production:server1",
				},
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return an error
	require.True(t, result.IsError)
}

func TestGetOSInfo_NoParams(t *testing.T) {
	engine := setupTestStorage(t)

	tool := &GetOSInfo{}
	handler := tool.Handler(context.Background(), engine)

	// Don't specify either parameter
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return an error requiring at least one parameter
	require.True(t, result.IsError)
}

func TestGetOSInfo_NonexistentHost(t *testing.T) {
	engine := setupTestStorage(t)

	addTestHost(t, engine, "production", "server1", "10.0.1.1")

	tool := &GetOSInfo{}
	handler := tool.Handler(context.Background(), engine)

	// Request OS info for non-existent host
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"name_of_hosts": []interface{}{
					"production:nonexistent",
				},
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return an error
	require.True(t, result.IsError)
}

func TestGetOSInfo_InvalidHostIdentifierFormat(t *testing.T) {
	engine := setupTestStorage(t)

	tool := &GetOSInfo{}
	handler := tool.Handler(context.Background(), engine)

	// Use invalid format (missing group)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"name_of_hosts": []interface{}{
					"server1", // Missing "group:" prefix
				},
			},
		},
	}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return an error about invalid format
	require.True(t, result.IsError)
}

func TestGetOSInfo_EmptyGroup(t *testing.T) {
	engine := setupTestStorage(t)

	addTestHost(t, engine, "production", "server1", "10.0.1.1")

	tool := &GetOSInfo{}
	handler := tool.Handler(context.Background(), engine)

	// Request OS info for empty/nonexistent group
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
	// Should return an error
	require.True(t, result.IsError)
}
