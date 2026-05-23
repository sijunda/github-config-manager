package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"git-config-manager/internal/audit"
	"git-config-manager/internal/profile"
	"git-config-manager/pkg/ui"

	"github.com/spf13/cobra"
)

func newUseCmd() *cobra.Command {
	var (
		global bool
		local  bool
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "use <profile>",
		Short: "Activate a profile",
		Long: `Activate a Git profile for the current session, globally, or locally.

Examples:
  gcm use work           # Session activation
  gcm use work --global  # Set as default
  gcm use work --local   # Pin to current project`,
		Args: requireArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			name := args[0]

			// Validate profile exists (case-sensitive exact match).
			if _, err := ctr.ProfileManager.Get(name); err != nil {
				return fmt.Errorf("profile %q not found\n\n  To see available profiles: gcm profile list\n  To create a new profile:   gcm profile create %s -i", name, name)
			}

			scope := profile.ScopeSession
			if global {
				scope = profile.ScopeGlobal
			} else if local {
				scope = profile.ScopeLocal
			}

			if dryRun {
				p, err := ctr.ProfileManager.Get(name)
				if err != nil {
					return err
				}
				ui.Header("🔍 Dry-run mode: No changes will be made")
				ui.Blank()
				ui.Print("Changes that would be applied:")
				ui.Blank()
				ui.Detail("user.name", p.Git.User.Name)
				ui.Detail("user.email", p.Git.User.Email)
				if p.SSH != nil {
					ui.Detail("SSH key", p.SSH.KeyPath)
				}
				ui.Blank()
				ui.Print("To apply: gcm use %s", name)
				return nil
			}

			// When --global is used, also clear any local override in current dir
			// so the global setting actually takes effect here.
			if scope == profile.ScopeGlobal {
				projectFile := ctr.Config.AutoSwitch.ProjectFile
				if projectFile != "" {
					os.Remove(projectFile)
				}
				ctr.ProfileSwitcher.ClearSession()
			}

			if err := ctr.ProfileSwitcher.Activate(name, scope); err != nil {
				ctr.AuditLogger.Log(audit.ActionProfileActivate, name,
					map[string]string{"scope": scope.String()}, err)
				// Smart fallback: if session fails, use local scope (.gcm-profile in cwd)
				if scope == profile.ScopeSession {
					scope = profile.ScopeLocal
					if err2 := ctr.ProfileSwitcher.Activate(name, scope); err2 != nil {
						return err2
					}
				} else {
					return err
				}
			}
			ctr.AuditLogger.Log(audit.ActionProfileActivate, name,
				map[string]string{"scope": scope.String()}, nil)

			ui.Success("Profile %q activated (%s)", name, scope.String())

			p, _ := ctr.ProfileManager.Get(name)
			if p != nil {
				ui.Detail("User", fmt.Sprintf("%s <%s>", p.Git.User.Name, p.Git.User.Email))
			}

			// Switch git credentials: clear old, store new (if available)
			server := gitServer()

			username := ""
			if p != nil && p.GitHub != nil && p.GitHub.Username != "" {
				username = p.GitHub.Username
			}

			if IsCredentialHelperConfigured() {
				// GCM is the credential helper — git will ask us dynamically.
				// Just pin the username so git knows which account to expect.
				if username == "" {
					username = name
				}
				_ = ctr.GitHubClient.SetGitCredentialUsername(server, username)
			} else {
				// Legacy mode: store credentials in system keychain via git credential approve/reject.
				_ = ctr.GitHubClient.ClearGitCredentials(server)

				if token, err := ctr.GitHubClient.LoadToken(name); err == nil && token != "" {
					if username == "" {
						username = name
					}
					_ = ctr.GitHubClient.StoreGitCredentials(server, username, token)
					_ = ctr.GitHubClient.SetGitCredentialUsername(server, username)
				} else {
					if username == "" {
						username = name
					}
					_ = ctr.GitHubClient.SetGitCredentialUsername(server, username)
				}
			}

			// Proactive token verification (best-effort, short timeout)
			if token, err := ctr.GitHubClient.LoadToken(name); err == nil && token != "" {
				ctr.GitHubClient.SetToken(token)
				verifyCtx, cancel := context.WithTimeout(cobraCmd.Context(), 5*time.Second)
				defer cancel()
				if _, err := ctr.GitHubClient.VerifyToken(verifyCtx); err != nil {
					ui.Blank()
					ui.Warning("GitHub token for %q may be expired or invalid", name)
					ui.Print("  %s Re-authenticate: %s", ui.IconArrow, ui.Cyan(fmt.Sprintf("gcm github login %s", name)))
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&global, "global", "g", false, "Set as global default")
	cmd.Flags().BoolVarP(&local, "local", "l", false, "Pin to current project")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")

	return cmd
}

func newCurrentCmd() *cobra.Command {
	var short bool
	var hideDefault bool

	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the active profile",
		RunE: func(_ *cobra.Command, _ []string) error {
			name, scope, err := ctr.ProfileSwitcher.Current()
			if err != nil {
				if short {
					// Stay silent so shell prompt hooks don't render anything.
					return nil
				}
				ui.Info("No active profile")
				return nil
			}

			if short {
				if hideDefault && name == ctr.Config.DefaultProfile {
					return nil
				}
				fmt.Println(name)
				return nil
			}

			p, getErr := ctr.ProfileManager.Get(name)
			if getErr != nil {
				ui.Success("Currently using profile: %s", name)
				return nil
			}

			ui.Success("Currently using profile: %s", name)
			ui.Blank()
			ui.Detail("User", fmt.Sprintf("%s <%s>", p.Git.User.Name, p.Git.User.Email))
			ui.Detail("Activation", scope.String())

			if p.SSH != nil {
				ui.Detail("SSH Key", p.SSH.KeyPath)
			}
			if p.GPG != nil {
				ui.Detail("GPG Signing", "Enabled ("+p.GPG.KeyID+")")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&short, "short", false, "Print only the profile name")
	cmd.Flags().BoolVar(&hideDefault, "hide-default", false, "Output nothing when active profile is the default (useful for shell prompts)")
	return cmd
}

func newRefreshCmd() *cobra.Command {
	var silent bool

	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Re-evaluate current directory and activate appropriate profile",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := ctr.ProfileSwitcher.Refresh(); err != nil {
				if !silent {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&silent, "silent", false, "Suppress all output")
	return cmd
}
