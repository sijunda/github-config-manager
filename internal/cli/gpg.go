package cli

import (
	"fmt"

	"github-config-manager/internal/audit"
	"github-config-manager/internal/gpg"
	"github-config-manager/internal/profile"
	"github-config-manager/pkg/ui"

	"github.com/spf13/cobra"
)

func newGPGCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gpg",
		Short: "Manage GPG keys",
		RunE: func(_ *cobra.Command, _ []string) error {
			return gpgListRun()
		},
	}

	cmd.AddCommand(newGPGGenerateCmd())
	cmd.AddCommand(newGPGListCmd())
	cmd.AddCommand(newGPGSignCmd())
	cmd.AddCommand(newGPGTestCmd())

	return cmd
}

func newGPGGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate <profile>",
		Short: "Generate a new GPG key for a profile",
		Args:  requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			profileName := args[0]

			p, err := ctr.ProfileManager.Get(profileName)
			if err != nil {
				ui.Error("profile %q not found", profileName)
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}

			sp := ui.NewSpinner("Generating GPG key...")
			sp.Start()

			keyInfo, err := ctr.GPGManager.Generate(gpg.GenerateOptions{
				Name:  p.Git.User.Name,
				Email: p.Git.User.Email,
			})
			if err != nil {
				sp.StopError("Failed to generate GPG key")
				ctr.AuditLogger.Log(audit.ActionGPGGenerate, profileName, nil, err)
				return err
			}

			sp.Stop("GPG key generated!")
			ctr.AuditLogger.Log(audit.ActionGPGGenerate, profileName,
				map[string]string{"key_id": keyInfo.KeyID}, nil)

			ui.Blank()
			ui.Detail("Key ID", keyInfo.KeyID)
			ui.Detail("Fingerprint", keyInfo.Fingerprint)
			ui.Detail("Email", keyInfo.Email)

			// Update profile
			p.GPG = &profile.GPGConfig{KeyID: keyInfo.KeyID}
			p.Git.Commit.GPGSign = profile.BoolPtr(true)
			p.Git.User.SigningKey = keyInfo.KeyID
			_ = ctr.ProfileManager.Update(p)

			ui.NextSteps([]string{
				fmt.Sprintf("Test signing: gcm gpg test %s", profileName),
				fmt.Sprintf("Upload to GitHub: gcm github login %s", profileName),
			})

			return nil
		},
	}

	return cmd
}

func newGPGListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List GPG keys", Aliases: []string{"ls"},
		RunE: func(_ *cobra.Command, _ []string) error {
			return gpgListRun()
		},
	}
}

func gpgListRun() error {
	if !ctr.GPGManager.IsInstalled() {
		ui.Warning("GPG is not installed on your system")
		ui.Info("Install GPG: https://gnupg.org/download/")
		return nil
	}

	keys, err := ctr.GPGManager.List()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		ui.Info("No GPG keys found")
		return nil
	}

	headers := []string{"Key ID", "Name", "Email", "Created", "Trust"}
	var rows [][]string
	for _, k := range keys {
		rows = append(rows, []string{
			k.KeyID, k.Name, k.Email,
			k.Created.Format("2006-01-02"), k.Trust,
		})
	}

	ui.SimpleTable(headers, rows)
	return nil
}

func newGPGSignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Manage commit signing",
	}

	cmd.AddCommand(&cobra.Command{
		Use: "enable <profile>", Short: "Enable commit signing", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			if p.GPG == nil || p.GPG.KeyID == "" {
				return fmt.Errorf("profile %q has no GPG key. Generate one: gcm gpg generate %s", args[0], args[0])
			}
			p.Git.Commit.GPGSign = profile.BoolPtr(true)
			p.Git.User.SigningKey = p.GPG.KeyID
			if err := ctr.ProfileManager.Update(p); err != nil {
				return err
			}
			ui.Success("Commit signing enabled for %q", args[0])
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use: "disable <profile>", Short: "Disable commit signing", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			p.Git.Commit.GPGSign = profile.BoolPtr(false)
			if err := ctr.ProfileManager.Update(p); err != nil {
				return err
			}
			ui.Success("Commit signing disabled for %q", args[0])
			return nil
		},
	})

	return cmd
}

func newGPGTestCmd() *cobra.Command {
	return &cobra.Command{
		Use: "test <profile>", Short: "Test GPG signing", Args: requireArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			if p.GPG == nil || p.GPG.KeyID == "" {
				return fmt.Errorf("profile %q has no GPG key configured", args[0])
			}

			sp := ui.NewSpinner("Testing GPG signing...")
			sp.Start()

			if err := ctr.GPGManager.TestSigning(p.GPG.KeyID); err != nil {
				sp.StopError("GPG signing test failed")
				return err
			}

			sp.Stop("GPG signing works!")
			return nil
		},
	}
}
