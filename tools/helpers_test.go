package tools

import (
	"path/filepath"
	"testing"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/storage"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary storage engine for testing
func setupTestStorage(t *testing.T) *storage.Engine {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_db")
	engine, err := storage.NewEngine(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		engine.Close()
	})
	return engine
}

// Helper function to add test hosts to storage
func addTestHost(t *testing.T, engine *storage.Engine, group, name, host string) {
	info := ssh.ClientInfo{
		Name:  name,
		Group: group,
		Host:  host,
		Port:  "22",
		User:  "testuser",
		Pass:  "testpass",
		OS: ssh.OSInfo{
			OSRelease: "Ubuntu 22.04",
			Uname:     "Linux test 5.15.0",
		},
	}
	err := engine.Set(info)
	require.NoError(t, err)
}
