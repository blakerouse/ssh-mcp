package commands

import (
	"context"
	"time"
)

// SetStatusForTest is a helper method for testing to set the command status
// This should only be used in tests
func (c *Command) SetStatusForTest(status CommandStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.status = status

	// Set timestamps based on status
	now := time.Now()
	switch status {
	case CommandStatusRunning:
		if c.startedAt == nil {
			c.startedAt = &now
		}
		// Set up a cancel function for running commands
		if c.cancel == nil {
			_, cancel := context.WithCancel(context.Background())
			c.cancel = cancel
		}
	case CommandStatusCompleted, CommandStatusFailed, CommandStatusCancelled:
		if c.startedAt == nil {
			c.startedAt = &now
		}
		if c.endedAt == nil {
			c.endedAt = &now
		}
	}
}

// SimulateCancellationForTest simulates the cancellation process for testing
// This sets the command to cancelled status after Cancel() is called
func (c *Command) SimulateCancellationForTest() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status == CommandStatusRunning || c.status == CommandStatusPending {
		c.status = CommandStatusCancelled
		now := time.Now()
		if c.startedAt == nil {
			c.startedAt = &now
		}
		if c.endedAt == nil {
			c.endedAt = &now
		}
	}
}
