package commands

import (
	"fmt"
	"sync"
	"time"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/google/uuid"
)

// Runner is an interface for managing background commands
type Runner interface {
	CreateCommand(commandStr string, hosts []ssh.ClientInfo) *Command
	GetCommand(commandID string) (*Command, error)
	GetMostRecentCommand() (*Command, error)
	ListCommands() []*Command
	CancelAllCommands()
}

// runner is the implementation of Runner
type runner struct {
	commands map[string]*Command
	mu       sync.RWMutex
}

// NewRunner creates a new command runner
func NewRunner() Runner {
	r := &runner{
		commands: make(map[string]*Command),
	}
	return r
}

// CreateCommand creates a new command and returns it
func (r *runner) CreateCommand(commandStr string, hosts []ssh.ClientInfo) *Command {
	commandID := uuid.New().String()

	cmd := &Command{
		id:        commandID,
		status:    CommandStatusPending,
		command:   commandStr,
		hosts:     hosts,
		results:   make(map[string]CommandResult),
		createdAt: time.Now(),
	}

	r.mu.Lock()
	r.commands[commandID] = cmd
	r.mu.Unlock()

	return cmd
}

// GetCommand retrieves a command by ID
func (r *runner) GetCommand(commandID string) (*Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	command, exists := r.commands[commandID]
	if !exists {
		return nil, fmt.Errorf("command not found: %s", commandID)
	}

	return command, nil
}

// GetMostRecentCommand returns the most recently created command
func (r *runner) GetMostRecentCommand() (*Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.commands) == 0 {
		return nil, fmt.Errorf("no commands found")
	}

	var mostRecent *Command
	for _, cmd := range r.commands {
		if mostRecent == nil || cmd.createdAt.After(mostRecent.createdAt) {
			mostRecent = cmd
		}
	}
	return mostRecent, nil
}

// ListCommands returns all commands
func (r *runner) ListCommands() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]*Command, 0, len(r.commands))
	for _, command := range r.commands {
		commands = append(commands, command)
	}

	return commands
}

// CancelAllCommands cancels all running commands
func (r *runner) CancelAllCommands() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cmd := range r.commands {
		if cmd.Status() == CommandStatusRunning {
			_ = cmd.Cancel()
		}
	}
}
