package cli

import (
	"fmt"

	"github-config-manager/internal/audit"
	"github-config-manager/internal/profile"
	"github-config-manager/internal/ssh"
	"github-config-manager/pkg/ui"

	"github.com/spf13/cobra"
)

func newSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Manage SSH keys",
		RunE: func(_ *cobra.Command, _ []string) error {
			return sshListRun()
		},
	}

	cmd.AddCommand(newSSHGenerateCmd())
	cmd.AddCommand(newSSHListCmd())
	cmd.AddCommand(newSSHTestCmd())
	cmd.AddCommand(newSSHCopyCmd())

	return cmd
}

func newSSHGenerateCmd() *cobra.Command {
	var (
		keyType    string
		bits       int
		comment    string
		passphrase string
	)

	cmd := &cobra.Command{
		Use:   "generate <profile>",
		Short: "Generate a new SSH key for a profile",
		Long: `Generate a new SSH key pair and associate it with a profile.

Examples:
  gcm ssh generate work --type ed25519
  gcm ssh generate work --type rsa --bits 4096`,
		Args: requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			if _, err := ctr.ProfileManager.Get(profileName); err != nil {
				ui.Error("profile %q not found", profileName)
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				ui.Print("  To create a new profile:   gcm profile create " + profileName + " -i")
				return nil
			}

			if cmd.Flags().Changed("passphrase") && passphrase != "" {
				ui.Warning("Passphrase provided via --passphrase flag may appear in shell history.")
				ui.Print("  For secure passphrase entry, omit the flag and you will be prompted interactively.")
			}

			sp := ui.NewSpinner("Generating SSH key...")
			sp.Start()

			keyInfo, err := ctr.SSHManager.Generate(ssh.GenerateOptions{
				Profile:    profileName,
				KeyType:    keyType,
				Bits:       bits,
				Comment:    comment,
				Passphrase: passphrase,
			})
			if err != nil {
				sp.StopError("Failed to generate SSH key")
				ctr.AuditLogger.Log(audit.ActionSSHGenerate, profileName,
					map[string]string{"type": keyType}, err)
				return err
			}

			sp.Stop("SSH key generated!")
			ctr.AuditLogger.Log(audit.ActionSSHGenerate, profileName,
				map[string]string{"type": keyInfo.Type, "path": keyInfo.Path}, nil)

			ui.Blank()
			ui.Detail("Path", keyInfo.Path)
			ui.Detail("Type", keyInfo.Type)
			ui.Detail("Fingerprint", keyInfo.Fingerprint)
			ui.Blank()
			ui.Print("Public key:")
			ui.Print(keyInfo.PublicKey)
			ui.NextSteps([]string{
				fmt.Sprintf("Test connection: gcm ssh test %s", profileName),
				fmt.Sprintf("Upload to GitHub: gcm github login %s", profileName),
			})

			// Update profile if it exists
			p, _ := ctr.ProfileManager.Get(profileName)
			if p != nil {
				p.SSH = &profile.SSHConfig{
					KeyPath:     keyInfo.Path,
					KeyType:     profile.KeyType(keyInfo.Type),
					Fingerprint: keyInfo.Fingerprint,
					Comment:     keyInfo.Comment,
				}
				_ = ctr.ProfileManager.Update(p)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&keyType, "type", "t", "ed25519", "Key type (ed25519, rsa, ecdsa)")
	cmd.Flags().IntVarP(&bits, "bits", "b", 4096, "Key bits (RSA only)")
	cmd.Flags().StringVarP(&comment, "comment", "c", "", "Key comment")
	cmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "Key passphrase")

	return cmd
}

func newSSHListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List SSH keys", Aliases: []string{"ls"},
		RunE: func(_ *cobra.Command, _ []string) error {
			return sshListRun()
		},
	}
}

func sshListRun() error {
	keys, err := ctr.SSHManager.List()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		ui.Info("No SSH keys found")
		return nil
	}

	headers := []string{"Key", "Type", "Fingerprint", "Agent"}
	var rows [][]string
	for _, k := range keys {
		agent := ui.Red("✗")
		if k.InAgent {
			agent = ui.Green("✓")
		}
		rows = append(rows, []string{k.Path, k.Type, k.Fingerprint, agent})
	}

	ui.SimpleTable(headers, rows)
	return nil
}

func newSSHTestCmd() *cobra.Command {
	return &cobra.Command{
		Use: "test <profile>", Short: "Test SSH connection to GitHub", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			if p.SSH == nil || p.SSH.KeyPath == "" {
				ui.Error("profile %q has no SSH key configured", args[0])
				ui.Blank()
				ui.Print("  To generate one: gcm ssh generate %s", args[0])
				return nil
			}

			sp := ui.NewSpinner("Testing SSH connection...")
			sp.Start()

			output, testErr := ctr.SSHManager.TestConnection(p.SSH.KeyPath)
			if testErr != nil {
				sp.StopError("SSH test failed")
				ui.Error("%s", output)
				return testErr
			}

			sp.Stop("SSH connection successful!")
			ui.Print(output)
			return nil
		},
	}
}

func newSSHCopyCmd() *cobra.Command {
	return &cobra.Command{
		Use: "copy <profile>", Short: "Show public key", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			if p.SSH == nil || p.SSH.KeyPath == "" {
				ui.Error("profile %q has no SSH key configured", args[0])
				ui.Blank()
				ui.Print("  To generate one: gcm ssh generate %s", args[0])
				return nil
			}
			pubKey, err := ctr.SSHManager.GetPublicKey(p.SSH.KeyPath)
			if err != nil {
				return err
			}
			ui.Print(pubKey)
			return nil
		},
	}
}
