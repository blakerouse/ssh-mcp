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

func dummyClientInfo(group, name string) ssh.ClientInfo {
	return ssh.ClientInfo{
		Name:  name,
		Group: group,
		Host:  "127.0.0.1",
		Port:  "22",
		User:  "testuser",
		Pass:  "testpass",
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
	info := dummyClientInfo("production", "host1")
	require.NoError(t, e1.Set(info))
	e1.Close()

	// Reopen and verify data persists
	e2, err := NewEngine(path)
	require.NoError(t, err)
	defer e2.Close()

	got, ok := e2.Get("production", "host1")
	require.True(t, ok)
	require.Equal(t, info, got)
}

func TestEngine_SetAndGet(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	info := dummyClientInfo("staging", "host1")
	err = e.Set(info)
	require.NoError(t, err)

	got, ok := e.Get("staging", "host1")
	require.True(t, ok)
	require.Equal(t, info, got)
}

func TestEngine_Get_NotFound(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	_, ok := e.Get("production", "missing")
	require.False(t, ok)
}

func TestEngine_Delete(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	info := dummyClientInfo("development", "host1")
	require.NoError(t, e.Set(info))

	err = e.Delete("development", "host1")
	require.NoError(t, err)

	_, ok := e.Get("development", "host1")
	require.False(t, ok)
}

func TestEngine_List(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	info1 := dummyClientInfo("production", "host1")
	info2 := dummyClientInfo("staging", "host2")
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

func TestEngine_ListGroup(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	// Add hosts to different groups
	prodHost1 := dummyClientInfo("production", "host1")
	prodHost2 := dummyClientInfo("production", "host2")
	stagingHost := dummyClientInfo("staging", "host3")

	require.NoError(t, e.Set(prodHost1))
	require.NoError(t, e.Set(prodHost2))
	require.NoError(t, e.Set(stagingHost))

	// List production group
	prodList, err := e.ListGroup("production")
	require.NoError(t, err)
	require.Len(t, prodList, 2)
	require.Contains(t, prodList, prodHost1)
	require.Contains(t, prodList, prodHost2)

	// List staging group
	stagingList, err := e.ListGroup("staging")
	require.NoError(t, err)
	require.Len(t, stagingList, 1)
	require.Contains(t, stagingList, stagingHost)

	// List non-existent group
	emptyList, err := e.ListGroup("nonexistent")
	require.NoError(t, err)
	require.Empty(t, emptyList)
}

func TestEngine_ListGroups(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	// Add hosts to different groups
	require.NoError(t, e.Set(dummyClientInfo("production", "host1")))
	require.NoError(t, e.Set(dummyClientInfo("production", "host2")))
	require.NoError(t, e.Set(dummyClientInfo("staging", "host3")))
	require.NoError(t, e.Set(dummyClientInfo("development", "host4")))

	// List all groups
	groups, err := e.ListGroups()
	require.NoError(t, err)
	require.Len(t, groups, 3)
	require.Contains(t, groups, "production")
	require.Contains(t, groups, "staging")
	require.Contains(t, groups, "development")
}

func TestEngine_Set_EmptyGroup(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	info := ssh.ClientInfo{
		Name:  "host1",
		Group: "", // Empty group
		Host:  "127.0.0.1",
		Port:  "22",
		User:  "testuser",
		Pass:  "testpass",
	}

	err = e.Set(info)
	require.Error(t, err)
	require.Contains(t, err.Error(), "group cannot be empty")
}

func TestEngine_Set_EmptyName(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	info := ssh.ClientInfo{
		Name:  "", // Empty name
		Group: "production",
		Host:  "127.0.0.1",
		Port:  "22",
		User:  "testuser",
		Pass:  "testpass",
	}

	err = e.Set(info)
	require.Error(t, err)
	require.Contains(t, err.Error(), "name cannot be empty")
}

func TestEngine_GroupIsolation(t *testing.T) {
	path := tempDBPath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	defer e.Close()

	// Add hosts with same name but different groups
	prodHost := dummyClientInfo("production", "server1")
	stagingHost := dummyClientInfo("staging", "server1")

	require.NoError(t, e.Set(prodHost))
	require.NoError(t, e.Set(stagingHost))

	// Verify they are stored separately
	gotProd, ok := e.Get("production", "server1")
	require.True(t, ok)
	require.Equal(t, prodHost, gotProd)

	gotStaging, ok := e.Get("staging", "server1")
	require.True(t, ok)
	require.Equal(t, stagingHost, gotStaging)

	// Delete production host
	require.NoError(t, e.Delete("production", "server1"))

	// Verify staging host still exists
	gotStaging, ok = e.Get("staging", "server1")
	require.True(t, ok)
	require.Equal(t, stagingHost, gotStaging)

	// Verify production host is deleted
	_, ok = e.Get("production", "server1")
	require.False(t, ok)
}
