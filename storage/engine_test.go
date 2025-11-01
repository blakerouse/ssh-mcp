package storage

import (
	"path/filepath"
	"testing"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/stretchr/testify/require"
)

func tempDBPath(t *testing.T) string {
	dir := t.TempDir()
	return filepath.Join(dir, "badger_test")
}

func dummyClientInfo(name string) ssh.ClientInfo {
	return ssh.ClientInfo{
		Name: name,
		Host: "127.0.0.1",
		Port: "22",
		User: "testuser",
		Pass: "testpass",
	}
}

func TestNewEngine_DBNotExist(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	require.NotNil(t, e)
	defer e.Close()

	// Verify empty database
	list, err := e.List()
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestNewEngine_DBExists(t *testing.T) {
	path := tempDBPath(t)

	// Create and populate database
	e1, err := NewEngine(path)
	require.NoError(t, err)
	info := dummyClientInfo("host1")
	require.NoError(t, e1.Set(info))
	e1.Close()

	// Reopen and verify data persists
	e2, err := NewEngine(path)
	require.NoError(t, err)
	defer e2.Close()

	got, ok := e2.Get("host1")
	require.True(t, ok)
	require.Equal(t, info, got)
}

func TestEngine_SetAndGet(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	info := dummyClientInfo("host1")
	err = e.Set(info)
	require.NoError(t, err)

	got, ok := e.Get("host1")
	require.True(t, ok)
	require.Equal(t, info, got)
}

func TestEngine_Get_NotFound(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	_, ok := e.Get("missing")
	require.False(t, ok)
}

func TestEngine_Delete(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	info := dummyClientInfo("host1")
	require.NoError(t, e.Set(info))

	err = e.Delete("host1")
	require.NoError(t, err)

	_, ok := e.Get("host1")
	require.False(t, ok)
}

func TestEngine_List(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	info1 := dummyClientInfo("host1")
	info2 := dummyClientInfo("host2")
	require.NoError(t, e.Set(info1))
	require.NoError(t, e.Set(info2))

	list, err := e.List()
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Contains(t, list, info1)
	require.Contains(t, list, info2)
}

func TestEngine_InvalidPath(t *testing.T) {
	// Test that opening a database at an invalid path fails
	_, err := NewEngine("/dev/null/invalid/path")
	require.Error(t, err)
}

func TestEngine_Close(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)

	err = e.Close()
	require.NoError(t, err)

	// Calling Close again should be safe
	err = e.Close()
	require.NoError(t, err)
}
