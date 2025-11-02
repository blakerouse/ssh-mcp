package commands

import (
	"context"
	"fmt"
	"io"
	"maps"
	"sync"
	"time"

	"github.com/blakerouse/ssh-mcp/ssh"
	"github.com/blakerouse/ssh-mcp/utils"
	gossh "golang.org/x/crypto/ssh"
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

	// Run the command on all hosts in parallel
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(c.hosts))

		for _, host := range c.hosts {
			go func(host ssh.ClientInfo) {
				defer wg.Done()

				// Check if context is cancelled before starting
				select {
				case <-ctx.Done():
					c.mu.Lock()
					c.results[host.Name] = CommandResult{
						Host: host.Name,
						Err:  fmt.Errorf("command cancelled"),
					}
					c.mu.Unlock()
					return
				default:
				}

				// Connect to the host
				sshClient := ssh.NewClient(&host)
				err := sshClient.Connect()
				if err != nil {
					c.mu.Lock()
					c.results[host.Name] = CommandResult{
						Host: host.Name,
						Err:  fmt.Errorf("failed to connect: %w", err),
					}
					c.mu.Unlock()
					return
				}
				defer sshClient.Close()

				// Execute command with streaming output
				c.executeWithStreaming(ctx, sshClient, host.Name)
			}(host)
		}

		wg.Wait()

		// Update final status
		c.mu.Lock()
		defer c.mu.Unlock()

		now := time.Now()
		c.endedAt = &now

		// Check if command was cancelled
		select {
		case <-ctx.Done():
			c.status = CommandStatusCancelled
			return
		default:
		}

		// Check if any results have errors
		hasErrors := false
		for _, result := range c.results {
			if result.Err != nil {
				hasErrors = true
				break
			}
		}

		if hasErrors {
			c.status = CommandStatusFailed
		} else {
			c.status = CommandStatusCompleted
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

// executeWithStreaming executes a command with streaming stdout/stderr capture
func (c *Command) executeWithStreaming(ctx context.Context, sshClient *ssh.Client, hostName string) {
	// Create SSH session
	session, err := sshClient.NewSession()
	if err != nil {
		c.mu.Lock()
		c.results[hostName] = CommandResult{
			Host: hostName,
			Err:  fmt.Errorf("failed to create session: %w", err),
		}
		c.mu.Unlock()
		return
	}
	defer session.Close()

	// Create a pipe for stdout and stderr
	stdout, err := session.StdoutPipe()
	if err != nil {
		c.mu.Lock()
		c.results[hostName] = CommandResult{
			Host: hostName,
			Err:  fmt.Errorf("failed to create stdout pipe: %w", err),
		}
		c.mu.Unlock()
		return
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		c.mu.Lock()
		c.results[hostName] = CommandResult{
			Host: hostName,
			Err:  fmt.Errorf("failed to create stderr pipe: %w", err),
		}
		c.mu.Unlock()
		return
	}

	// Start the command
	if err := session.Start(c.command); err != nil {
		c.mu.Lock()
		c.results[hostName] = CommandResult{
			Host: hostName,
			Err:  fmt.Errorf("failed to start command: %w", err),
		}
		c.mu.Unlock()
		return
	}

	// Read output in real-time and update results
	var output []byte
	done := make(chan error, 1)

	go func() {
		// Read from stdout and stderr concurrently
		var stdoutBuf, stderrBuf []byte
		var bufMu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(2)

		// Helper function to read from a pipe and update the buffer
		readPipe := func(pipe io.Reader, buf *[]byte) {
			defer wg.Done()
			readBuf := make([]byte, 4096)
			for {
				n, err := pipe.Read(readBuf)
				if n > 0 {
					bufMu.Lock()
					*buf = append(*buf, readBuf[:n]...)
					// Update the result with partial output
					combined := string(append(stdoutBuf, stderrBuf...))
					bufMu.Unlock()

					c.mu.Lock()
					if result, exists := c.results[hostName]; exists {
						result.Result = combined
						c.results[hostName] = result
					} else {
						c.results[hostName] = CommandResult{
							Host:   hostName,
							Result: combined,
						}
					}
					c.mu.Unlock()
				}
				if err != nil {
					break
				}
			}
		}

		go readPipe(stdout, &stdoutBuf)
		go readPipe(stderr, &stderrBuf)

		wg.Wait()
		output = append(stdoutBuf, stderrBuf...)
		done <- session.Wait()
	}()

	// Wait for command to complete or context to be cancelled
	select {
	case <-ctx.Done():
		// Try to terminate the session gracefully
		_ = session.Signal(gossh.SIGTERM)
		session.Close()
		c.mu.Lock()
		c.results[hostName] = CommandResult{
			Host:   hostName,
			Result: string(output),
			Err:    fmt.Errorf("command cancelled"),
		}
		c.mu.Unlock()
	case err := <-done:
		c.mu.Lock()
		if err != nil {
			c.results[hostName] = CommandResult{
				Host:   hostName,
				Result: string(output),
				Err:    fmt.Errorf("command failed: %w", err),
			}
		} else {
			c.results[hostName] = CommandResult{
				Host:   hostName,
				Result: string(output),
			}
		}
		c.mu.Unlock()
	}
}
