package commands

import (
	"fmt"

	"github.com/blakerouse/ssh-mcp/ssh"
)

// MockRunner is a mock implementation of Runner for testing purposes
type MockRunner struct {
	Commands          map[string]*Command
	CreateCommandFunc func(commandStr string, hosts []ssh.ClientInfo) *Command
	GetCommandFunc    func(commandID string) (*Command, error)
	GetMostRecentFunc func() (*Command, error)
	ListCommandsFunc  func() []*Command
	CancelAllFunc     func()
}

// NewMockRunner creates a new mock runner
func NewMockRunner() *MockRunner {
	return &MockRunner{
		Commands: make(map[string]*Command),
	}
}

// CreateCommand creates a new command (mock implementation)
func (m *MockRunner) CreateCommand(commandStr string, hosts []ssh.ClientInfo) *Command {
	if m.CreateCommandFunc != nil {
		return m.CreateCommandFunc(commandStr, hosts)
	}
	// Default implementation
	cmd := &Command{
		id:      fmt.Sprintf("mock-cmd-%d", len(m.Commands)),
		status:  CommandStatusPending,
		command: commandStr,
		hosts:   hosts,
		results: make(map[string]CommandResult),
	}
	m.Commands[cmd.id] = cmd
	return cmd
}

// GetCommand retrieves a command by ID (mock implementation)
func (m *MockRunner) GetCommand(commandID string) (*Command, error) {
	if m.GetCommandFunc != nil {
		return m.GetCommandFunc(commandID)
	}
	// Default implementation
	cmd, exists := m.Commands[commandID]
	if !exists {
		return nil, fmt.Errorf("command not found: %s", commandID)
	}
	return cmd, nil
}

// GetMostRecentCommand returns the most recently created command (mock implementation)
func (m *MockRunner) GetMostRecentCommand() (*Command, error) {
	if m.GetMostRecentFunc != nil {
		return m.GetMostRecentFunc()
	}
	// Default implementation
	if len(m.Commands) == 0 {
		return nil, fmt.Errorf("no commands found")
	}
	// Return the first command (simplified mock)
	for _, cmd := range m.Commands {
		return cmd, nil
	}
	return nil, fmt.Errorf("no commands found")
}

// ListCommands returns all commands (mock implementation)
func (m *MockRunner) ListCommands() []*Command {
	if m.ListCommandsFunc != nil {
		return m.ListCommandsFunc()
	}
	// Default implementation
	commands := make([]*Command, 0, len(m.Commands))
	for _, cmd := range m.Commands {
		commands = append(commands, cmd)
	}
	return commands
}

// CancelAllCommands cancels all running commands (mock implementation)
func (m *MockRunner) CancelAllCommands() {
	if m.CancelAllFunc != nil {
		m.CancelAllFunc()
		return
	}
	// Default implementation
	for _, cmd := range m.Commands {
		if cmd.Status() == CommandStatusRunning {
			_ = cmd.Cancel()
		}
	}
}
