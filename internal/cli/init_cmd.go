package cli

import (
	"github-config-manager/internal/audit"
	"github-config-manager/internal/shell"
	"github-config-manager/pkg/ui"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up shell integration and credential helper",
		Long: `Install shell hooks for auto-switching and prompt integration,
and register GCM as the git credential helper for GitHub.

Examples:
  gcm init            # Auto-detect shell
  gcm init --shell zsh`,
		RunE: func(_ *cobra.Command, _ []string) error {
			shellType := ctr.ShellManager.DetectShell()
			if shellType == shell.ShellUnknown {
				ui.Error("Could not detect your shell")
				ui.Info("Set SHELL environment variable and retry: SHELL=/bin/zsh gcm init")
				return nil
			}

			ui.Header("%s Setting up GCM for %s", ui.IconRocket, string(shellType))
			ui.Blank()

			configFile, err := ctr.ShellManager.Install(shellType)
			if err != nil {
				if configFile != "" {
					ui.Warning("%v", err)
				} else {
					return err
				}
				return nil
			}

			ui.Success("Shell integration installed!")
			ctr.AuditLogger.Log(audit.ActionShellInit, "",
				map[string]string{"shell": string(shellType), "config": configFile}, nil)
			ui.Detail("Shell", string(shellType))
			ui.Detail("Config", configFile)

			// Register GCM as credential helper for GitHub
			ui.Blank()
			if err := RegisterCredentialHelper(); err != nil {
				ui.Warning("Could not register credential helper: %v", err)
				ui.Print("  Git will fall back to the system keychain for credentials.")
			} else {
				ui.Success("Git credential helper registered!")
				ui.Detail("Scope", "github.com")
				ui.Print("  Git push/pull/clone will use GCM's encrypted token store directly.")
				ui.Print("  External logout (VS Code, browser) will no longer affect git operations.")
			}

			ui.Blank()
			ui.Info("Restart your shell or run: source %s", configFile)

			// Clear global git identity if no profile is set as default.
			// This ensures git won't use stale identity values.
			if ctr.Config.DefaultProfile == "" {
				_ = ctr.ProfileSwitcher.ClearGlobalIdentity()
				ui.Blank()
				ui.Info("Global git identity cleared — activate a profile to set your identity:")
				ui.Print("  gcm setup          (guided wizard)")
				ui.Print("  gcm use <profile>  (if you already have profiles)")
			}

			return nil
		},
	}

	return cmd
}
