// GCM is the GitHub Config Manager CLI entry point.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github-config-manager/internal/cli"
	"github-config-manager/internal/config"
	"github-config-manager/internal/container"
	"github-config-manager/pkg/logger"
	"github-config-manager/pkg/ui"

	"github.com/spf13/cobra"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Ensure data directories exist
	if err := config.EnsureDirs(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(logger.LevelInfo, os.Stderr)

	// Handle UI settings
	if !cfg.UI.Color {
		ui.DisableColor()
	}
	if cfg.UI.Verbose {
		log.SetVerbose(true)
	}
	if cfg.UI.Quiet {
		log.SetQuiet(true)
	}

	// Create dependency container
	ctr := container.New(cfg, log)

	// Wire the master-password prompt so encrypted token storage can ask
	// the user interactively. Uses the ui.AskPassword helper which
	// suppresses echo when stdin is a terminal.
	ctr.SetMasterPasswordPrompt(func(msg string) (string, error) {
		return ui.AskPassword(msg)
	})

	// Wire CLI commands
	cli.SetContainer(ctr)
	rootCmd := cli.NewRootCmd()

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("verbose"); v {
			log.SetVerbose(true)
		}
		if nc, _ := cmd.Flags().GetBool("no-color"); nc {
			ui.DisableColor()
		}
		if q, _ := cmd.Flags().GetBool("quiet"); q {
			log.SetQuiet(true)
		}
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Cancel in-flight API calls on SIGINT/SIGTERM so resources are freed
	// promptly on Ctrl-C rather than waiting for the HTTP timeout.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		ui.Error("%v", err)
		ctr.TokenStore.ZeroPassword()
		os.Exit(1)
	}

	// Zero sensitive material from memory before process exit.
	ctr.TokenStore.ZeroPassword()
}
