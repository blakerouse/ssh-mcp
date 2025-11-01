package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/blakerouse/ssh-mcp/storage"
	"github.com/blakerouse/ssh-mcp/tools"
)

var rootCmd = &cobra.Command{
	Use:   "ssh-mcp",
	Short: "SSH-MCP is a tool that allows you to manage remote machines over SSH with AI.",
	Run: func(cmd *cobra.Command, args []string) {
		// Start the server
		err := run(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().String("storage", "", "Storage path for hosts")
}

func main() {
	err := rootCmd.Execute()
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	storagePath := cmd.Flag("storage").Value.String()
	if storagePath == "" {
		return errors.New("--storage is required")
	}
	err := os.MkdirAll(path.Dir(storagePath), 0700)
	if err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}
	storageEngine, err := storage.NewEngine(storagePath)
	if err != nil {
		return fmt.Errorf("failed to create storage engine: %w", err)
	}

	s := server.NewMCPServer(
		"SSH",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	for _, tool := range tools.Registry.Tools() {
		s.AddTool(tool.Definition(), tool.Handler(storageEngine))
	}

	// start the stdio server
	stdio := server.NewStdioServer(s)
	return stdio.Listen(ctx, os.Stdin, os.Stdout)
}
