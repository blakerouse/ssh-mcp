package ssh

import (
	"testing"
)

func TestNewClientInfo_ValidConnectionString(t *testing.T) {
	connStr := "ssh://user:pass@host:2222"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Host != "host" {
		t.Errorf("expected host 'host', got '%s'", info.Host)
	}
	if info.Port != "2222" {
		t.Errorf("expected port '2222', got '%s'", info.Port)
	}
	if info.User != "user" {
		t.Errorf("expected user 'user', got '%s'", info.User)
	}
	if info.Pass != "pass" {
		t.Errorf("expected pass 'pass', got '%s'", info.Pass)
	}
	if info.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", info.Name)
	}
}

func TestNewClientInfo_DefaultPort(t *testing.T) {
	connStr := "ssh://user:pass@host"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Port != "22" {
		t.Errorf("expected default port '22', got '%s'", info.Port)
	}
}

func TestNewClientInfo_InvalidScheme(t *testing.T) {
	connStr := "http://user:pass@host:22"
	_, err := NewClientInfo("test", connStr)
	if err == nil || err.Error() != "invalid SSH connection string: not ssh scheme" {
		t.Errorf("expected error for invalid scheme, got %v", err)
	}
}

func TestNewClientInfo_NoUserInfo(t *testing.T) {
	connStr := "ssh://host:22"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error for missing user info (should use defaults), got %v", err)
	}
	if info.User != "" {
		t.Errorf("expected empty user (to be filled by Connect), got '%s'", info.User)
	}
	if info.Pass != "" {
		t.Errorf("expected empty password (will use SSH agent), got '%s'", info.Pass)
	}
}

func TestNewClientInfo_UserOnly(t *testing.T) {
	connStr := "ssh://myuser@host:22"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.User != "myuser" {
		t.Errorf("expected user 'myuser', got '%s'", info.User)
	}
	if info.Pass != "" {
		t.Errorf("expected empty password (will use SSH agent), got '%s'", info.Pass)
	}
}

func TestNewClientInfo_UserWithEmptyPassword(t *testing.T) {
	connStr := "ssh://user:@host:22"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.User != "user" {
		t.Errorf("expected user 'user', got '%s'", info.User)
	}
	if info.Pass != "" {
		t.Errorf("expected empty password, got '%s'", info.Pass)
	}
}

func TestNewClientInfo_MissingHost(t *testing.T) {
	connStr := "ssh://user:pass@:22"
	_, err := NewClientInfo("test", connStr)
	if err == nil || err.Error() != "invalid SSH connection string: missing host" {
		t.Errorf("expected error for missing host, got %v", err)
	}
}

func TestNewClientInfo_InvalidURL(t *testing.T) {
	connStr := "ssh://user:pass@host:badport:extra"
	_, err := NewClientInfo("test", connStr)
	if err == nil {
		t.Errorf("expected error for invalid URL, got nil")
	}
}

func TestNewClientInfo_WithoutScheme(t *testing.T) {
	connStr := "user:pass@host:2222"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Host != "host" {
		t.Errorf("expected host 'host', got '%s'", info.Host)
	}
	if info.Port != "2222" {
		t.Errorf("expected port '2222', got '%s'", info.Port)
	}
	if info.User != "user" {
		t.Errorf("expected user 'user', got '%s'", info.User)
	}
	if info.Pass != "pass" {
		t.Errorf("expected pass 'pass', got '%s'", info.Pass)
	}
}

func TestNewClientInfo_WithoutSchemeHostOnly(t *testing.T) {
	connStr := "host"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Host != "host" {
		t.Errorf("expected host 'host', got '%s'", info.Host)
	}
	if info.Port != "22" {
		t.Errorf("expected default port '22', got '%s'", info.Port)
	}
}

func TestNewClientInfo_WithoutSchemeWithPort(t *testing.T) {
	connStr := "host:2222"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Host != "host" {
		t.Errorf("expected host 'host', got '%s'", info.Host)
	}
	if info.Port != "2222" {
		t.Errorf("expected port '2222', got '%s'", info.Port)
	}
}

func TestNewClientInfo_WithoutSchemeUserAndHost(t *testing.T) {
	connStr := "user@host"
	info, err := NewClientInfo("test", connStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Host != "host" {
		t.Errorf("expected host 'host', got '%s'", info.Host)
	}
	if info.User != "user" {
		t.Errorf("expected user 'user', got '%s'", info.User)
	}
}
