package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// Tests for GetGroups tool

func TestGetGroups_EmptyStorage(t *testing.T) {
	engine := setupTestStorage(t)
	tool := &GetGroups{}
	handler := tool.Handler(engine)

	request := mcp.CallToolRequest{}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return empty list, not an error
	require.False(t, result.IsError)
}

func TestGetGroups_MultipleGroups(t *testing.T) {
	engine := setupTestStorage(t)

	// Add hosts to different groups
	addTestHost(t, engine, "production", "server1", "10.0.1.1")
	addTestHost(t, engine, "production", "server2", "10.0.1.2")
	addTestHost(t, engine, "staging", "server3", "10.0.2.1")
	addTestHost(t, engine, "development", "server4", "10.0.3.1")

	tool := &GetGroups{}
	handler := tool.Handler(engine)

	request := mcp.CallToolRequest{}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)

	// Verify the result contains the groups
	content := result.Content
	require.NotEmpty(t, content)
}

func TestGetGroups_SingleGroup(t *testing.T) {
	engine := setupTestStorage(t)

	// Add multiple hosts to same group
	addTestHost(t, engine, "production", "server1", "10.0.1.1")
	addTestHost(t, engine, "production", "server2", "10.0.1.2")

	tool := &GetGroups{}
	handler := tool.Handler(engine)

	request := mcp.CallToolRequest{}
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
}
