package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func tempFilePath(t *testing.T) string {
	dir := t.TempDir()
	return filepath.Join(dir, "engine_test.yaml")
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

func TestNewEngine_FileNotExist(t *testing.T) {
	path := tempFilePath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)
	require.NotNil(t, e)
	require.Empty(t, e.hosts)
}

func TestNewEngine_FileExists(t *testing.T) {
	path := tempFilePath(t)
	hosts := map[string]ssh.ClientInfo{
		"host1": dummyClientInfo("host1"),
	}
	data, err := yaml.Marshal(hosts)
	require.NoError(t, err)
	err = os.WriteFile(path, data, 0644)
	require.NoError(t, err)

	e, err := NewEngine(path)
	require.NoError(t, err)
	require.Equal(t, hosts, e.hosts)
}

func TestEngine_SetAndGet(t *testing.T) {
	path := tempFilePath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)

	info := dummyClientInfo("host1")
	err = e.Set(info)
	require.NoError(t, err)

	got, ok := e.Get("host1")
	require.True(t, ok)
	require.Equal(t, info, got)
}

func TestEngine_Get_NotFound(t *testing.T) {
	path := tempFilePath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)

	_, ok := e.Get("missing")
	require.False(t, ok)
}

func TestEngine_Delete(t *testing.T) {
	path := tempFilePath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)

	info := dummyClientInfo("host1")
	require.NoError(t, e.Set(info))

	err = e.Delete("host1")
	require.NoError(t, err)

	_, ok := e.Get("host1")
	require.False(t, ok)
}

func TestEngine_List(t *testing.T) {
	path := tempFilePath(t)
	e, err := NewEngine(path)
	require.NoError(t, err)

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

func TestEngine_load_InvalidFile(t *testing.T) {
	path := tempFilePath(t)
	err := os.WriteFile(path, []byte("invalid_yaml: [:"), 0644)
	require.NoError(t, err)

	e := &Engine{path: path}
	loadErr := e.load()
	require.Error(t, loadErr)
}

func TestEngine_save_Error(t *testing.T) {
	e := &Engine{
		hosts: map[string]ssh.ClientInfo{
			"host1": dummyClientInfo("host1"),
		},
		path: "/invalid/path/engine.yaml",
	}
	err := e.save()
	require.Error(t, err)
}
