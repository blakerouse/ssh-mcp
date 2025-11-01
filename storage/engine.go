package storage

import (
	"errors"
	"fmt"
	"os"

	"github.com/blakerouse/ssh-mcp/ssh"
	"gopkg.in/yaml.v3"
)

// Engine is the storage engine for SSH connections.
type Engine struct {
	hosts map[string]ssh.ClientInfo

	// path to store the state
	path string
}

// NewEngine creates a new storage Engine instance.
func NewEngine(path string) (*Engine, error) {
	e := &Engine{
		path: path,
	}
	err := e.load()
	if err != nil {
		return nil, err
	}
	return e, nil
}

// Get retrieves the SSH client information for a host.
func (e *Engine) Get(host string) (ssh.ClientInfo, bool) {
	info, ok := e.hosts[host]
	if !ok {
		return ssh.ClientInfo{}, false
	}
	return info, true
}

// Set saves the SSH client information for a host.
func (e *Engine) Set(info ssh.ClientInfo) error {
	e.hosts[info.Name] = info
	return e.save()
}

// Delete removes the SSH client information for a host.
func (e *Engine) Delete(host string) error {
	delete(e.hosts, host)
	return e.save()
}

// List retrieves the names of all hosts.
func (e *Engine) List() ([]ssh.ClientInfo, error) {
	var hosts []ssh.ClientInfo
	for _, info := range e.hosts {
		hosts = append(hosts, info)
	}
	return hosts, nil
}

func (e *Engine) load() error {
	data, err := os.ReadFile(e.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// doesn't exist yet
			e.hosts = make(map[string]ssh.ClientInfo)
			return nil
		}
		return fmt.Errorf("failed to read storage file: %w", err)
	}
	err = yaml.Unmarshal(data, &e.hosts)
	if err != nil {
		return fmt.Errorf("failed to unmarshal storage file: %w", err)
	}
	return nil
}

func (e *Engine) save() error {
	data, err := yaml.Marshal(e.hosts)
	if err != nil {
		return fmt.Errorf("failed to marshal storage data: %w", err)
	}
	err = os.WriteFile(e.path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}
	return nil
}
