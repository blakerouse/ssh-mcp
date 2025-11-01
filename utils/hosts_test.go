package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
)

func setupTestStorage(t *testing.T) (*storage.Engine, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "ssh-mcp-utils-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	engine, err := storage.NewEngine(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage engine: %v", err)
	}

	cleanup := func() {
		engine.Close()
		os.RemoveAll(tmpDir)
	}

	return engine, cleanup
}

func addTestHost(t *testing.T, engine *storage.Engine, group, name, host string) {
	t.Helper()
	info := ssh.ClientInfo{
		Name:  name,
		Group: group,
		Host:  host,
		Port:  "22",
		User:  "testuser",
	}
	if err := engine.Set(info); err != nil {
		t.Fatalf("failed to add test host: %v", err)
	}
}

func TestParseHostIdentifiers_Valid(t *testing.T) {
	hostStrings := []string{"prod:web01", "staging:db01", "dev:api01"}
	identifiers, err := ParseHostIdentifiers(hostStrings)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(identifiers) != 3 {
		t.Errorf("expected 3 identifiers, got %d", len(identifiers))
	}

	expected := []HostIdentifier{
		{Group: "prod", Name: "web01"},
		{Group: "staging", Name: "db01"},
		{Group: "dev", Name: "api01"},
	}

	for i, id := range identifiers {
		if id.Group != expected[i].Group {
			t.Errorf("identifier %d: expected group '%s', got '%s'", i, expected[i].Group, id.Group)
		}
		if id.Name != expected[i].Name {
			t.Errorf("identifier %d: expected name '%s', got '%s'", i, expected[i].Name, id.Name)
		}
	}
}

func TestParseHostIdentifiers_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name        string
		hostStrings []string
		expectedErr string
	}{
		{
			name:        "no colon",
			hostStrings: []string{"prod-web01"},
			expectedErr: "invalid host identifier format 'prod-web01', expected 'group:name'",
		},
		{
			name:        "empty string",
			hostStrings: []string{""},
			expectedErr: "invalid host identifier format '', expected 'group:name'",
		},
		{
			name:        "only group",
			hostStrings: []string{"prod"},
			expectedErr: "invalid host identifier format 'prod', expected 'group:name'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseHostIdentifiers(tc.hostStrings)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected error '%s', got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestParseHostIdentifiers_MultipleColons(t *testing.T) {
	// Should handle multiple colons by splitting on first colon only
	hostStrings := []string{"prod:web:01"}
	identifiers, err := ParseHostIdentifiers(hostStrings)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(identifiers) != 1 {
		t.Errorf("expected 1 identifier, got %d", len(identifiers))
	}

	if identifiers[0].Group != "prod" {
		t.Errorf("expected group 'prod', got '%s'", identifiers[0].Group)
	}
	if identifiers[0].Name != "web:01" {
		t.Errorf("expected name 'web:01', got '%s'", identifiers[0].Name)
	}
}

func TestParseHostIdentifiers_EmptyList(t *testing.T) {
	identifiers, err := ParseHostIdentifiers([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(identifiers) != 0 {
		t.Errorf("expected 0 identifiers, got %d", len(identifiers))
	}
}

func TestGetHostsFromStorage_Success(t *testing.T) {
	engine, cleanup := setupTestStorage(t)
	defer cleanup()

	// Add test hosts
	addTestHost(t, engine, "prod", "web01", "10.0.1.1")
	addTestHost(t, engine, "prod", "web02", "10.0.1.2")
	addTestHost(t, engine, "staging", "db01", "10.0.2.1")

	identifiers := []HostIdentifier{
		{Group: "prod", Name: "web01"},
		{Group: "staging", Name: "db01"},
	}

	hosts, err := GetHostsFromStorage(engine, identifiers)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}

	// Verify we got the right hosts
	foundWeb01 := false
	foundDb01 := false
	for _, host := range hosts {
		if host.Group == "prod" && host.Name == "web01" {
			foundWeb01 = true
		}
		if host.Group == "staging" && host.Name == "db01" {
			foundDb01 = true
		}
	}

	if !foundWeb01 {
		t.Error("expected to find prod:web01")
	}
	if !foundDb01 {
		t.Error("expected to find staging:db01")
	}
}

func TestGetHostsFromStorage_NotFound(t *testing.T) {
	engine, cleanup := setupTestStorage(t)
	defer cleanup()

	// Add one host but request a different one
	addTestHost(t, engine, "prod", "web01", "10.0.1.1")

	identifiers := []HostIdentifier{
		{Group: "prod", Name: "web99"},
	}

	_, err := GetHostsFromStorage(engine, identifiers)
	if err == nil {
		t.Error("expected error for non-existent host, got nil")
	}

	expectedErr := "no matching hosts for: prod:web99"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestGetHostsFromStorage_PartialMatch(t *testing.T) {
	engine, cleanup := setupTestStorage(t)
	defer cleanup()

	// Add one host
	addTestHost(t, engine, "prod", "web01", "10.0.1.1")

	// Request one existing and one non-existing host
	identifiers := []HostIdentifier{
		{Group: "prod", Name: "web01"},
		{Group: "prod", Name: "web99"},
	}

	hosts, err := GetHostsFromStorage(engine, identifiers)
	if err != nil {
		t.Fatalf("expected no error for partial match, got %v", err)
	}

	if len(hosts) != 1 {
		t.Errorf("expected 1 host (partial match), got %d", len(hosts))
	}
}

func TestGetHostsFromStorage_EmptyIdentifiers(t *testing.T) {
	engine, cleanup := setupTestStorage(t)
	defer cleanup()

	addTestHost(t, engine, "prod", "web01", "10.0.1.1")

	hosts, err := GetHostsFromStorage(engine, []HostIdentifier{})
	if err == nil {
		t.Error("expected error for empty identifiers, got nil")
	}

	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(hosts))
	}
}

func TestGetHostsFromGroup_Success(t *testing.T) {
	engine, cleanup := setupTestStorage(t)
	defer cleanup()

	// Add multiple hosts to the same group
	addTestHost(t, engine, "prod", "web01", "10.0.1.1")
	addTestHost(t, engine, "prod", "web02", "10.0.1.2")
	addTestHost(t, engine, "prod", "db01", "10.0.1.3")
	addTestHost(t, engine, "staging", "web01", "10.0.2.1")

	hosts, err := GetHostsFromGroup(engine, "prod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(hosts) != 3 {
		t.Errorf("expected 3 hosts in prod group, got %d", len(hosts))
	}

	// Verify all hosts are from the prod group
	for _, host := range hosts {
		if host.Group != "prod" {
			t.Errorf("expected group 'prod', got '%s'", host.Group)
		}
	}
}

func TestGetHostsFromGroup_EmptyGroup(t *testing.T) {
	engine, cleanup := setupTestStorage(t)
	defer cleanup()

	// Add hosts to a different group
	addTestHost(t, engine, "prod", "web01", "10.0.1.1")

	_, err := GetHostsFromGroup(engine, "staging")
	if err == nil {
		t.Error("expected error for empty group, got nil")
	}

	expectedErr := "no hosts found in group: staging"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestGetHostsFromGroup_NonexistentGroup(t *testing.T) {
	engine, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := GetHostsFromGroup(engine, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent group, got nil")
	}
}
