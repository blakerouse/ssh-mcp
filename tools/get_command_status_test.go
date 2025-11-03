package tools

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/blakerouse/ssh-mcp/commands"
	"github.com/blakerouse/ssh-mcp/ssh"
)

// TestGetCommandStatus_ByID tests retrieving a specific command by ID
func TestGetCommandStatus_ByID(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a test command
	cmd := mock.CreateCommand("echo test", hosts)
	cmd.SetStatusForTest(commands.CommandStatusCompleted)

	tool := &GetCommandStatus{
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

	// Should return structured content with command state
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

// TestGetCommandStatus_MostRecent tests getting the most recent command
func TestGetCommandStatus_MostRecent(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create multiple commands
	cmd1 := mock.CreateCommand("echo first", hosts)
	cmd1.SetStatusForTest(commands.CommandStatusCompleted)

	time.Sleep(10 * time.Millisecond)

	cmd2 := mock.CreateCommand("echo second", hosts)
	cmd2.SetStatusForTest(commands.CommandStatusRunning)

	tool := &GetCommandStatus{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	// No command_id provided - should get most recent
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected successful result")
	}
}

// TestGetCommandStatus_NotFound tests error when command doesn't exist
func TestGetCommandStatus_NotFound(t *testing.T) {
	mock := commands.NewMockRunner()

	tool := &GetCommandStatus{
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
}

// TestGetCommandStatus_NoCommands tests error when no commands exist
func TestGetCommandStatus_NoCommands(t *testing.T) {
	mock := commands.NewMockRunner()

	tool := &GetCommandStatus{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	// No command_id and no commands exist
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
		t.Error("expected error result when no commands exist")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected text content for error")
	}

	expectedMsg := "no commands found"
	if textContent.Text != expectedMsg {
		t.Errorf("expected '%s', got '%s'", expectedMsg, textContent.Text)
	}
}

// TestGetCommandStatus_RunningCommand tests getting status of running command
func TestGetCommandStatus_RunningCommand(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a running command
	cmd := mock.CreateCommand("sleep 10", hosts)
	cmd.SetStatusForTest(commands.CommandStatusRunning)

	tool := &GetCommandStatus{
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

	// Should return partial results for running command
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

// TestGetCommandStatus_Wait_AlreadyCompleted tests wait parameter with completed command
func TestGetCommandStatus_Wait_AlreadyCompleted(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a completed command
	cmd := mock.CreateCommand("echo done", hosts)
	cmd.SetStatusForTest(commands.CommandStatusCompleted)

	tool := &GetCommandStatus{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
				"wait":       true,
			},
		},
	}

	start := time.Now()
	result, err := handler(context.Background(), request)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected successful result")
	}

	// Should return immediately since already completed
	if elapsed > 1*time.Second {
		t.Errorf("expected quick return for completed command, took %v", elapsed)
	}
}

// TestGetCommandStatus_Wait_CompletesQuickly tests wait parameter when command completes
func TestGetCommandStatus_Wait_CompletesQuickly(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a running command that will complete soon
	cmd := mock.CreateCommand("echo test", hosts)
	cmd.SetStatusForTest(commands.CommandStatusRunning)

	// Set up to complete the command after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cmd.SetStatusForTest(commands.CommandStatusCompleted)
	}()

	tool := &GetCommandStatus{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
				"wait":       true,
			},
		},
	}

	start := time.Now()
	result, err := handler(context.Background(), request)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected successful result")
	}

	// Should wait for completion but not the full 30 seconds
	if elapsed < 100*time.Millisecond {
		t.Error("expected to wait for command completion")
	}
	if elapsed > 2*time.Second {
		t.Errorf("expected to return quickly after completion, took %v", elapsed)
	}
}

// TestGetCommandStatus_Wait_Timeout tests wait parameter with 30s timeout
func TestGetCommandStatus_Wait_Timeout(t *testing.T) {
	// Skip this test in short mode as it takes time
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a command that stays running
	cmd := mock.CreateCommand("sleep infinity", hosts)
	cmd.SetStatusForTest(commands.CommandStatusRunning)

	tool := &GetCommandStatus{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
				"wait":       true,
			},
		},
	}

	start := time.Now()
	result, err := handler(context.Background(), request)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected successful result even after timeout")
	}

	// Should timeout after 30 seconds
	if elapsed < 29*time.Second || elapsed > 31*time.Second {
		t.Errorf("expected ~30s timeout, took %v", elapsed)
	}
}

// TestGetCommandStatus_Wait_Failed tests wait parameter with failed command
func TestGetCommandStatus_Wait_Failed(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a running command that will fail
	cmd := mock.CreateCommand("false", hosts)
	cmd.SetStatusForTest(commands.CommandStatusRunning)

	// Set up to fail the command after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cmd.SetStatusForTest(commands.CommandStatusFailed)
	}()

	tool := &GetCommandStatus{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
				"wait":       true,
			},
		},
	}

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return successful result even though command failed
	if result.IsError {
		t.Error("expected successful result (command failure is not a tool error)")
	}
}

// TestGetCommandStatus_Wait_Cancelled tests wait parameter with cancelled command
func TestGetCommandStatus_Wait_Cancelled(t *testing.T) {
	mock := commands.NewMockRunner()

	hosts := []ssh.ClientInfo{
		{Name: "host1", Host: "example.com", Port: "22", Group: "prod"},
	}

	// Create a running command that will be cancelled
	cmd := mock.CreateCommand("sleep 100", hosts)
	cmd.SetStatusForTest(commands.CommandStatusRunning)

	// Set up to cancel the command after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cmd.SetStatusForTest(commands.CommandStatusCancelled)
	}()

	tool := &GetCommandStatus{
		commandRunner: mock,
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"command_id": cmd.ID(),
				"wait":       true,
			},
		},
	}

	start := time.Now()
	result, err := handler(context.Background(), request)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected successful result")
	}

	// Should return quickly after cancellation
	if elapsed > 2*time.Second {
		t.Errorf("expected to return quickly after cancellation, took %v", elapsed)
	}
}

// TestGetCommandStatus_NilRunner tests panic when runner is not set
func TestGetCommandStatus_NilRunner(t *testing.T) {
	tool := &GetCommandStatus{
		commandRunner: nil, // Not set
	}

	storageEngine := createTestStorage(t)
	defer storageEngine.Close()

	handler := tool.Handler(context.Background(), storageEngine)
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
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
