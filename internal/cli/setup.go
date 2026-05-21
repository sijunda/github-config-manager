package cli

import (
	"context"
	"fmt"

	"github-config-manager/internal/audit"
	"github-config-manager/internal/gpg"
	"github-config-manager/internal/profile"
	"github-config-manager/internal/shell"
	"github-config-manager/internal/ssh"
	"github-config-manager/pkg/ui"

	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Guided first-time setup wizard",
		Long: `Interactive wizard that walks you through the complete GCM setup.

This command will guide you through:
  1. Shell integration (auto-switching, prompt)
  2. Creating your first profile (name, email)
  3. SSH key generation
  4. GPG signing (optional)
  5. GitHub authentication
  6. Activating your profile

Perfect for first-time users. Run this once and you're fully set up.`,
		Aliases: []string{"quickstart"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSetup(cmd.Context())
		},
	}
}

func runSetup(ctx context.Context) error {
	ui.Header("%s Welcome to GCM — Let's get you set up!", ui.IconRocket)
	ui.Blank()
	ui.Print("This wizard will guide you through the complete setup.")
	ui.Print("It takes about 2 minutes. You can skip any step.")
	ui.Blank()
	ui.Divider()

	// ═══════════════════════════════════════════════
	// Step 1: Shell Integration
	// ═══════════════════════════════════════════════
	ui.Header("Step 1/6: Shell Integration")
	ui.Print("Shell hooks enable auto-switching when you cd into projects.")
	ui.Blank()

	installShell, err := ui.AskConfirm("Install shell integration?", true)
	if err != nil {
		return err
	}

	if installShell {
		shellType := ctr.ShellManager.DetectShell()
		if shellType == shell.ShellUnknown {
			ui.Warning("Could not detect your shell — skipping")
		} else {
			configFile, shellErr := ctr.ShellManager.Install(shellType)
			if shellErr != nil {
				ui.Warning("Shell integration: %v", shellErr)
			} else {
				ui.Success("Shell integration installed for %s", string(shellType))
				ui.Detail("Config", configFile)
				ctr.AuditLogger.Log(audit.ActionShellInit, "",
					map[string]string{"shell": string(shellType), "config": configFile}, nil)
			}
		}
	} else {
		ui.Info("Skipped — you can run %s later", ui.Cyan("gcm init"))
	}

	// Register GCM as credential helper (always — this protects against
	// external credential store changes like VS Code logout).
	if !IsCredentialHelperConfigured() {
		if err := RegisterCredentialHelper(); err != nil {
			ui.Warning("Could not register credential helper: %v", err)
		} else {
			ui.Success("Git credential helper registered for github.com")
		}
	}

	ui.Blank()
	ui.Divider()

	// ═══════════════════════════════════════════════
	// Step 2: Create Profile
	// ═══════════════════════════════════════════════
	ui.Header("Step 2/6: Create Your First Profile")
	ui.Print("A profile holds your Git identity (name, email, keys).")
	ui.Print("Most people have 2: %s and %s", ui.Cyan("work"), ui.Cyan("personal"))
	ui.Blank()

	profileName, err := ui.AskString("Profile name:", "work")
	if err != nil {
		return err
	}

	// Check if it already exists
	existing, _ := ctr.ProfileManager.Get(profileName)
	if existing != nil {
		ui.Success("Profile %q already exists — using it", profileName)
	} else {
		fullName, nameErr := ui.AskString("Your full name:", "")
		if nameErr != nil {
			return nameErr
		}
		email, emailErr := ui.AskString("Your email:", "")
		if emailErr != nil {
			return emailErr
		}

		p := &profile.Profile{
			Name: profileName,
			Git: profile.GitConfig{
				User: profile.GitUser{Name: fullName, Email: email},
			},
		}

		if createErr := ctr.ProfileManager.Create(p); createErr != nil {
			ui.Error("Could not create profile: %v", createErr)
			return createErr
		}
		ctr.AuditLogger.Log(audit.ActionProfileCreate, profileName, nil, nil)
		ui.Success("Profile %q created!", profileName)
	}

	ui.Blank()
	ui.Divider()

	// ═══════════════════════════════════════════════
	// Step 3: SSH Key
	// ═══════════════════════════════════════════════
	ui.Header("Step 3/6: SSH Key")
	ui.Print("SSH keys let you push/pull without passwords.")
	ui.Blank()

	genSSH, err := ui.AskConfirm("Generate an SSH key for this profile?", true)
	if err != nil {
		return err
	}

	if genSSH {
		sp := ui.NewSpinner("Generating ed25519 SSH key...")
		sp.Start()
		keyInfo, genErr := ctr.SSHManager.Generate(ssh.GenerateOptions{
			Profile: profileName,
			KeyType: "ed25519",
		})
		if genErr != nil {
			sp.StopError("SSH key generation failed")
			ui.Warning("%v", genErr)
		} else {
			sp.Stop("SSH key generated!")
			ui.Detail("Path", keyInfo.Path)
			ui.Detail("Fingerprint", keyInfo.Fingerprint)
			ctr.AuditLogger.Log(audit.ActionSSHGenerate, profileName,
				map[string]string{"type": keyInfo.Type, "path": keyInfo.Path}, nil)

			// Update profile with SSH info
			p, _ := ctr.ProfileManager.Get(profileName)
			if p != nil {
				p.SSH = &profile.SSHConfig{
					KeyPath:     keyInfo.Path,
					KeyType:     profile.KeyType(keyInfo.Type),
					Fingerprint: keyInfo.Fingerprint,
				}
				_ = ctr.ProfileManager.Update(p)
			}

			ui.Blank()
			ui.Print("Public key (add to GitHub → Settings → SSH keys):")
			ui.Print("  %s", ui.Dim(keyInfo.PublicKey))
		}
	} else {
		ui.Info("Skipped — you can run %s later", ui.Cyan(fmt.Sprintf("gcm ssh generate %s", profileName)))
	}

	ui.Blank()
	ui.Divider()

	// ═══════════════════════════════════════════════
	// Step 4: GPG Signing (optional)
	// ═══════════════════════════════════════════════
	ui.Header("Step 4/6: Commit Signing (Optional)")
	ui.Print("GPG signing proves commits came from you (shows 'Verified' badge).")
	ui.Blank()

	enableGPG, err := ui.AskConfirm("Enable commit signing?", false)
	if err != nil {
		return err
	}

	if enableGPG {
		p, _ := ctr.ProfileManager.Get(profileName)
		name := profileName
		email := ""
		if p != nil {
			name = p.Git.User.Name
			email = p.Git.User.Email
		}

		sp := ui.NewSpinner("Generating GPG key...")
		sp.Start()
		keyInfo, genErr := ctr.GPGManager.Generate(gpg.GenerateOptions{
			Name: name, Email: email,
		})
		if genErr != nil {
			sp.StopError("GPG key generation failed")
			ui.Warning("%v", genErr)
		} else {
			sp.Stop("GPG key generated!")
			ui.Detail("Key ID", keyInfo.KeyID)

			if p != nil {
				p.GPG = &profile.GPGConfig{KeyID: keyInfo.KeyID}
				p.Git.Commit.GPGSign = profile.BoolPtr(true)
				p.Git.User.SigningKey = keyInfo.KeyID
				_ = ctr.ProfileManager.Update(p)
			}
		}
	} else {
		ui.Info("Skipped — you can enable this later with %s", ui.Cyan("gcm gpg generate "+profileName))
	}

	ui.Blank()
	ui.Divider()

	// ═══════════════════════════════════════════════
	// Step 5: GitHub Authentication
	// ═══════════════════════════════════════════════
	ui.Header("Step 5/6: GitHub Authentication")
	ui.Print("Connecting GitHub lets GCM manage your git credentials automatically.")
	ui.Blank()

	loginGH, err := ui.AskConfirm("Authenticate with GitHub now?", true)
	if err != nil {
		return err
	}

	if loginGH {
		method, methodErr := ui.AskSelect("Authentication method:", []string{
			"Personal Access Token (paste a token)",
			"OAuth Device Flow (browser-based)",
		})
		if methodErr != nil {
			return methodErr
		}

		if method == "Personal Access Token (paste a token)" {
			ui.Blank()
			ui.Print("Get a token at: %s", ui.Cyan("https://github.com/settings/tokens"))
			ui.Print("Scopes needed: repo, admin:public_key, admin:gpg_key")
			ui.Blank()

			token, tokenErr := ui.AskPassword("Paste your token")
			if tokenErr != nil {
				return tokenErr
			}

			if token != "" {
				// Verify before saving
				ctr.GitHubClient.SetToken(token)
				user, verifyErr := ctr.GitHubClient.VerifyToken(ctx)
				if verifyErr != nil {
					ui.Error("Token is invalid: %v", verifyErr)
					ui.Print("  %s You can try again later: %s", ui.IconArrow, ui.Cyan(fmt.Sprintf("gcm github login %s", profileName)))
				} else {
					if saveErr := ctr.GitHubClient.SaveToken(profileName, token); saveErr != nil {
						ui.Error("Could not save token: %v", saveErr)
					} else {
						ui.Success("Authenticated as @%s", user.Login)
						// Update profile with GitHub username
						p, _ := ctr.ProfileManager.Get(profileName)
						if p != nil {
							if p.GitHub == nil {
								p.GitHub = &profile.GitHubConfig{}
							}
							p.GitHub.Username = user.Login
							_ = ctr.ProfileManager.Update(p)
						}
					}
				}
			}
		} else {
			// OAuth device flow
			ui.Blank()
			dcr, flowErr := ctr.GitHubClient.InitiateDeviceFlow()
			if flowErr != nil {
				ui.Error("Could not start device flow: %v", flowErr)
				ui.Print("  %s Try PAT instead: %s", ui.IconArrow, ui.Cyan(fmt.Sprintf("gcm github login %s", profileName)))
			} else {
				ui.Print("Open this URL in your browser:")
				ui.Print("  %s", ui.Cyan(dcr.VerificationURI))
				ui.Blank()
				ui.Print("Enter this code: %s", ui.Bold(dcr.UserCode))
				ui.Blank()

				sp := ui.NewSpinner("Waiting for authorization...")
				sp.Start()

				token, pollErr := ctr.GitHubClient.PollForToken(
					ctx, dcr.DeviceCode, dcr.Interval)
				if pollErr != nil {
					sp.StopError("Authorization failed")
					ui.Error("%v", pollErr)
				} else {
					sp.Stop("Authorized!")
					ctr.GitHubClient.SetToken(token)
					user, _ := ctr.GitHubClient.VerifyToken(ctx)
					if saveErr := ctr.GitHubClient.SaveToken(profileName, token); saveErr != nil {
						ui.Error("Could not save token: %v", saveErr)
					} else {
						login := profileName
						if user != nil {
							login = user.Login
						}
						ui.Success("Authenticated as @%s", login)

						p, _ := ctr.ProfileManager.Get(profileName)
						if p != nil {
							if p.GitHub == nil {
								p.GitHub = &profile.GitHubConfig{}
							}
							if user != nil {
								p.GitHub.Username = user.Login
							}
							_ = ctr.ProfileManager.Update(p)
						}
					}
				}
			}
		}
	} else {
		ui.Info("Skipped — you can run %s later", ui.Cyan(fmt.Sprintf("gcm github login %s", profileName)))
	}

	ui.Blank()
	ui.Divider()

	// ═══════════════════════════════════════════════
	// Step 6: Activate
	// ═══════════════════════════════════════════════
	ui.Header("Step 6/6: Activate Profile")
	ui.Blank()

	activate, err := ui.AskConfirm(fmt.Sprintf("Activate profile %q now?", profileName), true)
	if err != nil {
		return err
	}

	if activate {
		if actErr := ctr.ProfileSwitcher.Activate(profileName, profile.ScopeSession); actErr != nil {
			// Fallback to local scope
			if actErr2 := ctr.ProfileSwitcher.Activate(profileName, profile.ScopeLocal); actErr2 != nil {
				ui.Warning("Could not activate: %v", actErr2)
			} else {
				ui.Success("Profile %q activated (local)", profileName)
			}
		} else {
			ui.Success("Profile %q activated (session)", profileName)
		}
		ctr.AuditLogger.Log(audit.ActionProfileActivate, profileName,
			map[string]string{"scope": "session"}, nil)
	}

	// ═══════════════════════════════════════════════
	// Done!
	// ═══════════════════════════════════════════════
	ui.Blank()
	ui.Divider()
	ui.Header("%s You're all set!", ui.IconCheck)
	ui.Blank()
	ui.Print("Your GCM setup is complete. Here's what you can do now:")
	ui.Blank()
	ui.NextSteps([]string{
		fmt.Sprintf("Check status anytime: %s", ui.Cyan("gcm status")),
		fmt.Sprintf("Create another profile: %s", ui.Cyan("gcm profile create <name> -i")),
		fmt.Sprintf("Switch profiles: %s", ui.Cyan("gcm use <profile>")),
		fmt.Sprintf("View all commands: %s", ui.Cyan("gcm --help")),
	})
	ui.Blank()

	return nil
}
