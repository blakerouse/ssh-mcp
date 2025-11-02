package commands

import (
	"fmt"
	"sync"
	"time"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/google/uuid"
)

// Runner manages background commands
type Runner struct {
	commands map[string]*Command
	mu       sync.RWMutex
}

// NewRunner creates a new command runner
func NewRunner() *Runner {
	r := &Runner{
		commands: make(map[string]*Command),
	}
	return r
}

// CreateCommand creates a new command and returns it
func (r *Runner) CreateCommand(commandStr string, hosts []ssh.ClientInfo) *Command {
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
func (r *Runner) GetCommand(commandID string) (*Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	command, exists := r.commands[commandID]
	if !exists {
		return nil, fmt.Errorf("command not found: %s", commandID)
	}

	return command, nil
}

// GetMostRecentCommand returns the most recently created command
func (r *Runner) GetMostRecentCommand() (*Command, error) {
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
func (r *Runner) ListCommands() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]*Command, 0, len(r.commands))
	for _, command := range r.commands {
		commands = append(commands, command)
	}

	return commands
}

// CancelAllCommands cancels all running commands
func (r *Runner) CancelAllCommands() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cmd := range r.commands {
		if cmd.Status() == CommandStatusRunning {
			_ = cmd.Cancel()
		}
	}
}
