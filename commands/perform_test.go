package commands

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/blakerouse/ssh-mcp/ssh"
)

func TestPerformCommandsOnHosts_EmptyHosts(t *testing.T) {
	hosts := []ssh.ClientInfo{}
	commandCalled := false

	command := func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
		commandCalled = true
		return "test", nil
	}

	results := PerformOnHosts(hosts, command)

	if commandCalled {
		t.Error("command should not be called for empty hosts list")
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestPerformCommandsOnHosts_SingleHost_ConnectionFailure(t *testing.T) {
	// Create a host that will fail to connect (invalid host)
	hosts := []ssh.ClientInfo{
		{
			Name:  "test-host",
			Group: "test",
			Host:  "invalid-host-that-does-not-exist.local",
			Port:  "22",
			User:  "testuser",
		},
	}

	commandCalled := false
	command := func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
		commandCalled = true
		return "should not reach here", nil
	}

	results := PerformOnHosts(hosts, command)

	if commandCalled {
		t.Error("command should not be called when connection fails")
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result, exists := results["test-host"]
	if !exists {
		t.Fatal("expected result for test-host")
	}

	if result.Host != "test-host" {
		t.Errorf("expected host 'test-host', got '%s'", result.Host)
	}

	if result.Err == nil {
		t.Error("expected connection error, got nil")
	}

	if result.Result != "" {
		t.Errorf("expected empty result for failed connection, got '%s'", result.Result)
	}
}

func TestPerformCommandsOnHosts_MultipleHosts_AllConnectionFailures(t *testing.T) {
	hosts := []ssh.ClientInfo{
		{
			Name:  "host1",
			Group: "test",
			Host:  "invalid1.local",
			Port:  "22",
			User:  "testuser",
		},
		{
			Name:  "host2",
			Group: "test",
			Host:  "invalid2.local",
			Port:  "22",
			User:  "testuser",
		},
		{
			Name:  "host3",
			Group: "test",
			Host:  "invalid3.local",
			Port:  "22",
			User:  "testuser",
		},
	}

	command := func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
		return "should not reach here", nil
	}

	results := PerformOnHosts(hosts, command)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify all hosts have error results
	for _, hostName := range []string{"host1", "host2", "host3"} {
		result, exists := results[hostName]
		if !exists {
			t.Errorf("expected result for %s", hostName)
			continue
		}

		if result.Host != hostName {
			t.Errorf("expected host '%s', got '%s'", hostName, result.Host)
		}

		if result.Err == nil {
			t.Errorf("expected error for %s, got nil", hostName)
		}
	}
}

func TestCommandResult_Structure(t *testing.T) {
	// Test that CommandResult has the expected fields
	result := CommandResult{
		Host:   "test-host",
		Result: "test result",
		Err:    errors.New("test error"),
	}

	if result.Host != "test-host" {
		t.Errorf("expected host 'test-host', got '%s'", result.Host)
	}

	if result.Result != "test result" {
		t.Errorf("expected result 'test result', got '%s'", result.Result)
	}

	if result.Err == nil {
		t.Error("expected error, got nil")
	}

	if result.Err.Error() != "test error" {
		t.Errorf("expected error 'test error', got '%s'", result.Err.Error())
	}
}

func TestCommandResult_SuccessCase(t *testing.T) {
	result := CommandResult{
		Host:   "successful-host",
		Result: "operation completed",
		Err:    nil,
	}

	if result.Host != "successful-host" {
		t.Errorf("expected host 'successful-host', got '%s'", result.Host)
	}

	if result.Result != "operation completed" {
		t.Errorf("expected result 'operation completed', got '%s'", result.Result)
	}

	if result.Err != nil {
		t.Errorf("expected no error, got %v", result.Err)
	}
}

func TestPerformCommandsOnHosts_ResultsMapKeys(t *testing.T) {
	// Verify that results are keyed by host name
	hosts := []ssh.ClientInfo{
		{
			Name:  "alpha",
			Group: "test",
			Host:  "invalid.local",
			Port:  "22",
		},
		{
			Name:  "beta",
			Group: "test",
			Host:  "invalid.local",
			Port:  "22",
		},
	}

	command := func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
		return "", nil
	}

	results := PerformOnHosts(hosts, command)

	// Check that results are keyed by the host names
	if _, exists := results["alpha"]; !exists {
		t.Error("expected result keyed by 'alpha'")
	}

	if _, exists := results["beta"]; !exists {
		t.Error("expected result keyed by 'beta'")
	}

	// Ensure Host field matches the key
	if results["alpha"].Host != "alpha" {
		t.Errorf("expected result['alpha'].Host to be 'alpha', got '%s'", results["alpha"].Host)
	}

	if results["beta"].Host != "beta" {
		t.Errorf("expected result['beta'].Host to be 'beta', got '%s'", results["beta"].Host)
	}
}

// TestPerformCommandsOnHosts_Concurrency verifies that tasks run concurrently
// by checking that all hosts complete within a reasonable timeframe
func TestPerformCommandsOnHosts_Concurrency(t *testing.T) {
	// Create multiple hosts that will all fail quickly
	hosts := make([]ssh.ClientInfo, 5)
	for i := 0; i < 5; i++ {
		hosts[i] = ssh.ClientInfo{
			Name:  string(rune('a' + i)),
			Group: "test",
			Host:  "nonexistent.local",
			Port:  "22",
		}
	}

	command := func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
		return "", nil
	}

	results := PerformOnHosts(hosts, command)

	// All 5 connection attempts should complete
	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	// All should have errors (connection failures)
	for key, result := range results {
		if result.Err == nil {
			t.Errorf("expected error for host %s, got nil", key)
		}
	}
}

// TestCommandResult_MarshalJSON_WithError tests JSON marshaling when error is present
func TestCommandResult_MarshalJSON_WithError(t *testing.T) {
	result := CommandResult{
		Host:   "test-host",
		Result: "some output",
		Err:    errors.New("connection failed"),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal CommandResult: %v", err)
	}

	// Unmarshal to verify structure
	var unmarshaled map[string]any
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if unmarshaled["host"] != "test-host" {
		t.Errorf("expected host 'test-host', got '%v'", unmarshaled["host"])
	}

	if unmarshaled["result"] != "some output" {
		t.Errorf("expected result 'some output', got '%v'", unmarshaled["result"])
	}

	if unmarshaled["error"] != "connection failed" {
		t.Errorf("expected error 'connection failed', got '%v'", unmarshaled["error"])
	}

	// Verify the JSON string contains expected fields
	jsonStr := string(jsonData)
	expectedFields := []string{`"host":"test-host"`, `"result":"some output"`, `"error":"connection failed"`}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("expected JSON to contain %s, got: %s", field, jsonStr)
		}
	}
}

// TestCommandResult_MarshalJSON_WithoutError tests JSON marshaling when no error
func TestCommandResult_MarshalJSON_WithoutError(t *testing.T) {
	result := CommandResult{
		Host:   "success-host",
		Result: "command completed successfully",
		Err:    nil,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal CommandResult: %v", err)
	}

	// Unmarshal to verify structure
	var unmarshaled map[string]any
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if unmarshaled["host"] != "success-host" {
		t.Errorf("expected host 'success-host', got '%v'", unmarshaled["host"])
	}

	if unmarshaled["result"] != "command completed successfully" {
		t.Errorf("expected result 'command completed successfully', got '%v'", unmarshaled["result"])
	}

	// Error field should be omitted when empty (omitempty)
	if _, exists := unmarshaled["error"]; exists {
		t.Errorf("expected error field to be omitted, but it exists with value: %v", unmarshaled["error"])
	}

	// Verify JSON string doesn't contain error field
	jsonStr := string(jsonData)
	if strings.Contains(jsonStr, "error") {
		t.Errorf("expected JSON to not contain 'error' field when Err is nil, got: %s", jsonStr)
	}
}

// TestCommandResult_MarshalJSON_EmptyResult tests JSON marshaling with empty result
func TestCommandResult_MarshalJSON_EmptyResult(t *testing.T) {
	result := CommandResult{
		Host:   "empty-host",
		Result: "",
		Err:    errors.New("failed before output"),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal CommandResult: %v", err)
	}

	var unmarshaled map[string]any
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if unmarshaled["host"] != "empty-host" {
		t.Errorf("expected host 'empty-host', got '%v'", unmarshaled["host"])
	}

	if unmarshaled["result"] != "" {
		t.Errorf("expected empty result, got '%v'", unmarshaled["result"])
	}

	if unmarshaled["error"] != "failed before output" {
		t.Errorf("expected error 'failed before output', got '%v'", unmarshaled["error"])
	}
}
