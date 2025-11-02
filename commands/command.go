package commands

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/utils"
)

// CommandStatus represents the current state of a command
type CommandStatus string

const (
	CommandStatusPending   CommandStatus = "pending"
	CommandStatusRunning   CommandStatus = "running"
	CommandStatusCompleted CommandStatus = "completed"
	CommandStatusFailed    CommandStatus = "failed"
	CommandStatusCancelled CommandStatus = "cancelled"
)

// Command represents a background command
type Command struct {
	id        string
	status    CommandStatus
	command   string
	hosts     []ssh.ClientInfo
	results   map[string]CommandResult
	createdAt time.Time
	startedAt *time.Time
	endedAt   *time.Time
	err       error
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// CommandState represents the serializable state of a Command
type CommandState struct {
	ID        string                         `json:"id"`
	Status    CommandStatus                  `json:"status"`
	Command   string                         `json:"command"`
	Hosts     []utils.HostIdentifier         `json:"hosts"`
	Results   map[string]CommandResult `json:"results"`
	CreatedAt time.Time                      `json:"created_at"`
	StartedAt *time.Time                     `json:"started_at,omitempty"`
	EndedAt   *time.Time                     `json:"ended_at,omitempty"`
	Error     string                         `json:"error,omitempty"`
}

// Start starts executing the command in the background
func (c *Command) Start() error {
	c.mu.Lock()
	if c.status != CommandStatusPending {
		c.mu.Unlock()
		return fmt.Errorf("command %s is not in pending state", c.id)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.status = CommandStatusRunning
	now := time.Now()
	c.startedAt = &now
	c.mu.Unlock()

	// Run the command in a goroutine
	go func() {
		results := PerformOnHosts(c.hosts, func(host ssh.ClientInfo, sshClient *ssh.Client) (string, error) {
			// Check if context is cancelled before executing
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("command cancelled")
			default:
			}

			output, err := sshClient.Exec(c.command)
			if err != nil {
				return "", fmt.Errorf("failed to execute command: %w", err)
			}
			return string(output), nil
		})

		c.mu.Lock()
		defer c.mu.Unlock()

		c.results = results
		now := time.Now()
		c.endedAt = &now

		// Check if any results have errors
		hasErrors := false
		for _, result := range results {
			if result.Err != nil {
				hasErrors = true
				break
			}
		}

		// Check if command was cancelled
		select {
		case <-ctx.Done():
			c.status = CommandStatusCancelled
		default:
			if hasErrors {
				c.status = CommandStatusFailed
			} else {
				c.status = CommandStatusCompleted
			}
		}
	}()

	return nil
}

// Cancel cancels the running command
func (c *Command) Cancel() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status != CommandStatusRunning {
		return fmt.Errorf("command %s is not running", c.id)
	}

	if c.cancel != nil {
		c.cancel()
	}

	return nil
}

// ID returns the command's unique identifier
func (c *Command) ID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.id
}

// Status returns the command's current status
func (c *Command) Status() CommandStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// CreatedAt returns the command's creation time
func (c *Command) CreatedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.createdAt
}

// ToState returns a safe copy of the command state for serialization
func (c *Command) ToState() *CommandState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Convert hosts to simplified identifiers
	hosts := make([]utils.HostIdentifier, len(c.hosts))
	for i, h := range c.hosts {
		hosts[i] = utils.HostIdentifier{
			Group: h.Group,
			Name:  h.Name,
		}
	}

	// Copy results
	results := make(map[string]CommandResult, len(c.results))
	maps.Copy(results, c.results)

	// Convert error to string
	errStr := ""
	if c.err != nil {
		errStr = c.err.Error()
	}

	return &CommandState{
		ID:        c.id,
		Status:    c.status,
		Command:   c.command,
		Hosts:     hosts,
		Results:   results,
		CreatedAt: c.createdAt,
		StartedAt: c.startedAt,
		EndedAt:   c.endedAt,
		Error:     errStr,
	}
}
