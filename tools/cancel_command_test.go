package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/blakerouse/ssh-mcp/commands"
	"github.com/blakerouse/ssh-mcp/ssh"
)

// TestCancelCommand_Success tests successfully cancelling a running command
func TestCancelCommand_Success(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a running command
	cmd := mock.CreateCommand("sleep 100", hosts)
	cmd.SetStatusForTest(commands.CommandStatusRunning)

	tool := &CancelCommand{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
			},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected successful result")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected text content")
	}

	expectedMsg := "Command " + cmd.ID() + " has been cancelled"
	if textContent.Text != expectedMsg {
		t.Errorf("expected '%s', got '%s'", expectedMsg, textContent.Text)
	}

	// Note: In real execution, the cancellation happens asynchronously via context
	// The test verifies that Cancel() was called successfully without error
}

// TestCancelCommand_MissingCommandID tests error when command_id is not provided
func TestCancelCommand_MissingCommandID(t *testing.T) {
	mock := commands.NewMockRunner()

	tool := &CancelCommand{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error result
	if !result.IsError {
		t.Error("expected error result for missing command_id")
	}
}

// TestCancelCommand_NotFound tests error when command doesn't exist
func TestCancelCommand_NotFound(t *testing.T) {
	mock := commands.NewMockRunner()

	tool := &CancelCommand{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": "nonexistent-id",
			},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error result
	if !result.IsError {
		t.Error("expected error result for nonexistent command")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected text content for error")
	}

	expectedMsg := "command not found: nonexistent-id"
	if textContent.Text != expectedMsg {
		t.Errorf("expected '%s', got '%s'", expectedMsg, textContent.Text)
	}
}

// TestCancelCommand_PendingCommand tests that cancelling a pending command fails
func TestCancelCommand_PendingCommand(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a pending command
	cmd := mock.CreateCommand("echo test", hosts)
	cmd.SetStatusForTest(commands.CommandStatusPending)

	tool := &CancelCommand{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
			},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error result for pending command
	if !result.IsError {
		t.Error("expected error result for pending command")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected text content for error")
	}

	// Error message includes the command ID
	if !strings.Contains(textContent.Text, "is not running") {
		t.Errorf("expected error about command not running, got '%s'", textContent.Text)
	}
}

// TestCancelCommand_AlreadyCompleted tests cancelling a completed command
func TestCancelCommand_AlreadyCompleted(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a completed command
	cmd := mock.CreateCommand("echo done", hosts)
	cmd.SetStatusForTest(commands.CommandStatusCompleted)

	tool := &CancelCommand{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
			},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error result
	if !result.IsError {
		t.Error("expected error result for already completed command")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected text content for error")
	}

	if !strings.Contains(textContent.Text, "is not running") {
		t.Errorf("expected error about command not running, got '%s'", textContent.Text)
	}
}

// TestCancelCommand_AlreadyFailed tests cancelling a failed command
func TestCancelCommand_AlreadyFailed(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a failed command
	cmd := mock.CreateCommand("false", hosts)
	cmd.SetStatusForTest(commands.CommandStatusFailed)

	tool := &CancelCommand{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
			},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error result
	if !result.IsError {
		t.Error("expected error result for already failed command")
	}
}

// TestCancelCommand_AlreadyCancelled tests cancelling an already cancelled command
func TestCancelCommand_AlreadyCancelled(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a cancelled command
	cmd := mock.CreateCommand("sleep 100", hosts)
	cmd.SetStatusForTest(commands.CommandStatusCancelled)

	tool := &CancelCommand{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
			},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error result
	if !result.IsError {
		t.Error("expected error result for already cancelled command")
	}
}

// TestCancelCommand_NilRunner tests panic when runner is not set
func TestCancelCommand_NilRunner(t *testing.T) {
	tool := &CancelCommand{
		commandRunner: nil, // Not set
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": "some-id",
			},
		},
	}

	// Should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when command runner is nil")
		}
	}()

	_, _ = handler(context.Background(), request)
}
