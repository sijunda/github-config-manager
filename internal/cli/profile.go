package cli

import (
	"fmt"
	"os"

	"github-config-manager/internal/audit"
	"github-config-manager/internal/gpg"
	"github-config-manager/internal/profile"
	"github-config-manager/internal/ssh"
	"github-config-manager/pkg/ui"

	"github.com/spf13/cobra"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "profile",
		Short:   "Manage Git profiles",
		Long:    "Create, list, show, edit, delete, and export/import Git profiles.",
		Aliases: []string{"p"},
		RunE:    profileListRun,
	}

	cmd.AddCommand(newProfileCreateCmd())
	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileShowCmd())
	cmd.AddCommand(newProfileEditCmd())
	cmd.AddCommand(newProfileDeleteCmd())
	cmd.AddCommand(newProfileExportCmd())
	cmd.AddCommand(newProfileImportCmd())
	cmd.AddCommand(newProfileDiffCmd())

	return cmd
}

func newProfileCreateCmd() *cobra.Command {
	var (
		name         string
		email        string
		interactive  bool
		editor       string
		sshKey       string
		fromTemplate string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new Git profile",
		Long: `Create a new Git profile with your identity information.

Examples:
  gcm profile create work --interactive
  gcm profile create work --name "John Doe" --email "john@company.com"
  gcm profile create work --from-template company-standard -i`,
		Args: requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			if interactive {
				return profileCreateInteractive(profileName, fromTemplate)
			}

			if name == "" || email == "" {
				return fmt.Errorf("--name and --email are required (or use --interactive)")
			}

			p := &profile.Profile{
				Name: profileName,
				Git: profile.GitConfig{
					User: profile.GitUser{Name: name, Email: email},
					Core: profile.GitCore{Editor: editor},
				},
			}

			if sshKey != "" {
				p.SSH = &profile.SSHConfig{KeyPath: sshKey}
			}

			// Apply template settings if specified
			if fromTemplate != "" {
				t, err := ctr.TemplateManager.Get(fromTemplate)
				if err != nil {
					return fmt.Errorf("loading template: %w", err)
				}
				applyTemplateToProfile(t, p)
			}

			if err := ctr.ProfileManager.Create(p); err != nil {
				ctr.AuditLogger.Log(audit.ActionProfileCreate, profileName, nil, err)
				return err
			}
			ctr.AuditLogger.Log(audit.ActionProfileCreate, profileName, nil, nil)

			ui.Success("Profile %q created successfully", profileName)
			ui.Blank()
			ui.Detail("Name", name)
			ui.Detail("Email", email)
			if fromTemplate != "" {
				ui.Detail("Template", fromTemplate)
			}
			ui.NextSteps([]string{
				fmt.Sprintf("Generate SSH key: gcm ssh generate %s", profileName),
				fmt.Sprintf("Activate profile: gcm use %s", profileName),
			})
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Your full name")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Your email address")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive wizard mode")
	cmd.Flags().StringVar(&editor, "editor", "", "Git editor")
	cmd.Flags().StringVar(&sshKey, "ssh-key", "", "Path to existing SSH key")
	cmd.Flags().StringVarP(&fromTemplate, "from-template", "t", "", "Apply settings from a template")

	return cmd
}

func profileCreateInteractive(profileName string, fromTemplate string) error {
	ui.Header("%s Create New Profile: %s", ui.IconProfile, profileName)

	// If template specified, show what will be applied
	if fromTemplate != "" {
		t, err := ctr.TemplateManager.Get(fromTemplate)
		if err != nil {
			return fmt.Errorf("loading template: %w", err)
		}
		ui.Print("  Using template: %s", ui.Cyan(fromTemplate))
		if t.Description != "" {
			ui.Print("  %s", ui.Dim(t.Description))
		}
		ui.Blank()
	}

	ui.SubHeader("Step 1/4: Basic Information")
	name, err := ui.AskString("Your full name:", "")
	if err != nil {
		return err
	}
	email, err := ui.AskString("Your email:", "")
	if err != nil {
		return err
	}
	editor, err := ui.AskString("Git editor (leave empty for default):", "")
	if err != nil {
		return err
	}

	ui.SubHeader("Step 2/4: SSH Configuration")
	generateSSH, err := ui.AskConfirm("Generate a new SSH key?", true)
	if err != nil {
		return err
	}

	var sshConfig *profile.SSHConfig
	if generateSSH {
		keyType, askErr := ui.AskSelect("SSH key type:", []string{"ed25519 (recommended)", "rsa (4096 bits)", "ecdsa"})
		if askErr != nil {
			return askErr
		}
		keyTypeClean := "ed25519"
		if keyType == "rsa (4096 bits)" {
			keyTypeClean = "rsa"
		} else if keyType == "ecdsa" {
			keyTypeClean = "ecdsa"
		}

		sp := ui.NewSpinner("Generating SSH key...")
		sp.Start()
		keyInfo, genErr := ctr.SSHManager.Generate(ssh.GenerateOptions{
			Profile: profileName,
			KeyType: keyTypeClean,
		})
		if genErr != nil {
			sp.StopError("Failed to generate SSH key")
			ui.Warning("SSH key generation failed: %v", genErr)
		} else {
			sp.Stop("SSH key generated")
			sshConfig = &profile.SSHConfig{
				KeyPath:     keyInfo.Path,
				KeyType:     profile.KeyType(keyInfo.Type),
				Fingerprint: keyInfo.Fingerprint,
			}
		}
	}

	ui.SubHeader("Step 3/4: GPG Signing")
	enableGPG, err := ui.AskConfirm("Enable commit signing?", false)
	if err != nil {
		return err
	}

	var gpgConfig *profile.GPGConfig
	if enableGPG {
		sp := ui.NewSpinner("Generating GPG key...")
		sp.Start()
		keyInfo, genErr := ctr.GPGManager.Generate(gpg.GenerateOptions{
			Name: name, Email: email,
		})
		if genErr != nil {
			sp.StopError("Failed to generate GPG key")
			ui.Warning("GPG key generation failed: %v", genErr)
		} else {
			sp.Stop("GPG key generated")
			gpgConfig = &profile.GPGConfig{KeyID: keyInfo.KeyID}
		}
	}

	ui.SubHeader("Step 4/4: GitHub (Optional)")
	githubUser, err := ui.AskString("GitHub username (leave empty to skip):", "")
	if err != nil {
		return err
	}

	var ghConfig *profile.GitHubConfig
	if githubUser != "" {
		ghConfig = &profile.GitHubConfig{Username: githubUser}
	}

	p := &profile.Profile{
		Name: profileName,
		Git: profile.GitConfig{
			User: profile.GitUser{Name: name, Email: email},
			Core: profile.GitCore{Editor: editor},
		},
		SSH: sshConfig, GPG: gpgConfig, GitHub: ghConfig,
	}

	if gpgConfig != nil {
		p.Git.Commit.GPGSign = profile.BoolPtr(true)
		p.Git.User.SigningKey = gpgConfig.KeyID
	}

	// Apply template settings if specified
	if fromTemplate != "" {
		t, _ := ctr.TemplateManager.Get(fromTemplate)
		if t != nil {
			applyTemplateToProfile(t, p)
		}
	}

	if err := ctr.ProfileManager.Create(p); err != nil {
		ctr.AuditLogger.Log(audit.ActionProfileCreate, profileName, nil, err)
		return err
	}
	ctr.AuditLogger.Log(audit.ActionProfileCreate, profileName, nil, nil)

	ui.Blank()
	ui.Success("Profile %q created and configured!", profileName)
	ui.NextSteps([]string{
		fmt.Sprintf("Activate it now: gcm use %s", profileName),
		"Create more profiles: gcm profile create <name> -i",
	})

	return nil
}

func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List all profiles", Aliases: []string{"ls"},
		RunE: profileListRun,
	}
}

func profileListRun(_ *cobra.Command, _ []string) error {
	profiles, err := ctr.ProfileManager.List()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		ui.Info("No profiles found. Create one with: gcm profile create <name> -i")
		return nil
	}

	currentName, _, _ := ctr.ProfileSwitcher.Current()

	headers := []string{"Profile", "Status", "Email", "Signing", "Last Used"}
	var rows [][]string

	for _, p := range profiles {
		status := ""
		if p.Name == currentName && p.Name == ctr.Config.DefaultProfile {
			status = ui.Green("●") + " " + ui.Cyan("default")
		} else if p.Name == currentName {
			status = ui.Green("●")
		} else if p.Name == ctr.Config.DefaultProfile {
			status = ui.Cyan("default")
		}

		signing := ui.Red("✗")
		if p.Git.Commit.GPGSign != nil && *p.Git.Commit.GPGSign {
			signing = ui.Green("✓")
		}

		lastUsed := ui.Dim("never")
		if p.Metadata.LastUsed != nil {
			lastUsed = formatTimeAgo(*p.Metadata.LastUsed)
		}

		rows = append(rows, []string{p.Name, status, p.Git.User.Email, signing, lastUsed})
	}

	ui.SimpleTable(headers, rows)
	ui.Blank()
	ui.Print("%d profiles", len(profiles))
	return nil
}

func newProfileShowCmd() *cobra.Command {
	return &cobra.Command{
		Use: "show <name>", Short: "Show profile details", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			ui.Header("Profile: %s", p.Name)
			ui.SubHeader("Git Configuration")
			ui.Detail("Name", p.Git.User.Name)
			ui.Detail("Email", p.Git.User.Email)
			if p.Git.Core.Editor != "" {
				ui.Detail("Editor", p.Git.Core.Editor)
			}
			if p.SSH != nil {
				ui.SubHeader("SSH")
				ui.Detail("Key", p.SSH.KeyPath)
				ui.Detail("Type", string(p.SSH.KeyType))
				if p.SSH.Fingerprint != "" {
					ui.Detail("Fingerprint", p.SSH.Fingerprint)
				}
			}
			if p.GPG != nil {
				ui.SubHeader("GPG")
				ui.Detail("Key ID", p.GPG.KeyID)
			}
			if p.GitHub != nil {
				ui.SubHeader("GitHub")
				ui.Detail("Username", p.GitHub.Username)
			}
			ui.SubHeader("Metadata")
			ui.Detail("Created", p.Metadata.Created.Format("2006-01-02 15:04:05"))
			ui.Detail("Usage", fmt.Sprintf("%d", p.Metadata.UsageCount))
			return nil
		},
	}
}

func newProfileEditCmd() *cobra.Command {
	var (
		name  string
		email string
	)
	cmd := &cobra.Command{
		Use: "edit <profile>", Short: "Edit a profile", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			if name != "" {
				p.Git.User.Name = name
			}
			if email != "" {
				p.Git.User.Email = email
			}
			if err := ctr.ProfileManager.Update(p); err != nil {
				return err
			}
			ui.Success("Profile %q updated", args[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Update name")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Update email")
	return cmd
}

func newProfileDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use: "delete <name>", Short: "Delete a profile", Aliases: []string{"rm"}, Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			profileName := args[0]

			// Load profile before deletion to get SSH key path
			p, err := ctr.ProfileManager.Get(profileName)
			if err != nil {
				return fmt.Errorf("profile %q not found\n\n  To see available profiles: gcm profile list", profileName)
			}

			// Check if trying to delete the active profile
			if isActiveProfile(profileName) {
				ui.Warning("Profile %q is currently active.", profileName)
				ui.Print("  Switch to another profile first: gcm use <other-profile>")
				ui.Blank()
				if !yes {
					confirm, err := ui.AskConfirm("Delete the active profile anyway?", false)
					if err != nil || !confirm {
						ui.Info("Cancelled")
						return nil
					}
				}
			} else if !yes {
				confirm, err := ui.AskConfirm(fmt.Sprintf("Delete profile %q?", profileName), false)
				if err != nil || !confirm {
					ui.Info("Cancelled")
					return nil
				}
			}

			if err := ctr.ProfileManager.Delete(profileName); err != nil {
				ctr.AuditLogger.Log(audit.ActionProfileDelete, profileName, nil, err)
				return fmt.Errorf("could not delete profile %q\n\n  Make sure the profile exists: gcm profile list", profileName)
			}
			ctr.AuditLogger.Log(audit.ActionProfileDelete, profileName, nil, nil)
			ui.Success("Profile %q deleted", profileName)

			// Clean up associated SSH key files
			if p.SSH != nil && p.SSH.KeyPath != "" {
				privKey := p.SSH.KeyPath
				pubKey := privKey + ".pub"
				removedAny := false
				if err := os.Remove(privKey); err == nil {
					removedAny = true
				}
				if err := os.Remove(pubKey); err == nil {
					removedAny = true
				}
				if removedAny {
					ui.Success("SSH key files removed")
					ui.Detail("Removed", privKey)
				}
			}

			// Clean up GitHub token
			if delErr := ctr.GitHubClient.DeleteToken(profileName); delErr == nil {
				ui.Success("GitHub token removed")
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

func newProfileExportCmd() *cobra.Command {
	return &cobra.Command{
		Use: "export <name>", Short: "Export profile (YAML)", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			data, err := ctr.ProfileManager.Export(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			fmt.Fprint(os.Stdout, string(data))
			return nil
		},
	}
}

func newProfileImportCmd() *cobra.Command {
	return &cobra.Command{
		Use: "import <file>", Short: "Import profile from file", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				ui.Error("could not read file: %s", args[0])
				ui.Blank()
				ui.Print("  Make sure the file path is correct.")
				ui.Print("  To export a profile: gcm profile export <name> > file.yaml")
				return nil
			}
			p, err := ctr.ProfileManager.Import(data)
			if err != nil {
				return err
			}
			ui.Success("Profile %q imported", p.Name)
			return nil
		},
	}
}

func newProfileDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use: "diff <profile1> <profile2>", Short: "Compare two profiles", Args: requireArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			p1, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			p2, err := ctr.ProfileManager.Get(args[1])
			if err != nil {
				ui.Error("profile %q not found", args[1])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			ui.Header("Comparing: %s vs %s", args[0], args[1])
			diffField("Name", p1.Git.User.Name, p2.Git.User.Name, args[0], args[1])
			diffField("Email", p1.Git.User.Email, p2.Git.User.Email, args[0], args[1])
			sshP1, sshP2 := "", ""
			if p1.SSH != nil {
				sshP1 = p1.SSH.KeyPath
			}
			if p2.SSH != nil {
				sshP2 = p2.SSH.KeyPath
			}
			diffField("SSH Key", sshP1, sshP2, args[0], args[1])
			return nil
		},
	}
}

func diffField(label, v1, v2, n1, n2 string) {
	if v1 != v2 {
		ui.Print("\n%s:", label)
		ui.Detail(n1, v1)
		ui.Detail(n2, v2)
	}
}
