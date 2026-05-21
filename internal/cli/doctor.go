package cli

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github-config-manager/internal/profile"
	"github-config-manager/pkg/ui"

	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [profile]",
		Short: "Validate a profile configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Validate all profiles
				profiles, err := ctr.ProfileManager.List()
				if err != nil {
					return err
				}
				for _, p := range profiles {
					validateAndPrint(p)
				}
				return nil
			}

			p, err := ctr.ProfileManager.Get(args[0])
			if err != nil {
				ui.Error("profile %q not found", args[0])
				ui.Blank()
				ui.Print("  To see available profiles: gcm profile list")
				return nil
			}
			validateAndPrint(p)
			return nil
		},
	}
}

func validateAndPrint(p *profile.Profile) {
	result := profile.ValidateDeep(p)

	icon := ui.Green(ui.IconSuccess)
	if !result.IsValid() {
		icon = ui.Red(ui.IconError)
	}

	ui.Print("\n%s Profile: %s", icon, ui.Bold(p.Name))

	for _, issue := range result.Info {
		ui.Print("  %s %s: %s", ui.Green(ui.IconSuccess), issue.Category, issue.Message)
	}
	for _, issue := range result.Warnings {
		ui.Print("  %s %s: %s", ui.Yellow(ui.IconWarning), issue.Category, issue.Message)
		if issue.Suggestion != "" {
			ui.Print("      %s", ui.Dim(issue.Suggestion))
		}
	}
	for _, issue := range result.Errors {
		ui.Print("  %s %s: %s", ui.Red(ui.IconError), issue.Category, issue.Message)
		if issue.Suggestion != "" {
			ui.Print("      %s", ui.Dim(issue.Suggestion))
		}
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system health and dependencies",
		RunE: func(_ *cobra.Command, _ []string) error {
			ui.Header("%s GCM System Health Check", ui.IconDoctor)
			ui.Blank()

			// System info
			ui.SubHeader("System")
			ui.Print("  %s OS: %s/%s", ui.Green(ui.IconSuccess), runtime.GOOS, runtime.GOARCH)
			ui.Print("  %s Go: %s", ui.Green(ui.IconSuccess), runtime.Version())

			// Git
			ui.SubHeader("Dependencies")
			checkCommand("Git", "git", "--version")
			checkCommand("SSH", "ssh", "-V")
			checkCommand("GPG", ctr.Config.Advanced.GPGCommand, "--version")

			// SSH agent
			ui.SubHeader("Services")
			checkSSHAgent()

			// Config
			ui.SubHeader("Configuration")
			configPath := "~/.gcm/config.yaml"
			ui.Print("  %s Config: %s", ui.Green(ui.IconSuccess), configPath)

			profiles, err := ctr.ProfileManager.List()
			if err != nil {
				ui.Print("  %s Profiles: error reading", ui.Red(ui.IconError))
			} else {
				ui.Print("  %s Profiles: %d found", ui.Green(ui.IconSuccess), len(profiles))
			}

			currentName, currentScope, _ := ctr.ProfileSwitcher.Current()
			if currentName != "" {
				ui.Print("  %s Active: %s (%s)", ui.Green(ui.IconSuccess), currentName, currentScope.String())
			} else {
				ui.Print("  %s No active profile", ui.Yellow(ui.IconWarning))
			}

			// Shell
			ui.SubHeader("Shell")
			shellType := ctr.ShellManager.DetectShell()
			ui.Print("  %s Detected: %s", ui.Green(ui.IconSuccess), string(shellType))

			// Credential Helper
			ui.SubHeader("Credential Helper")
			if IsCredentialHelperConfigured() {
				ui.Print("  %s GCM registered as git credential helper for github.com", ui.Green(ui.IconSuccess))
				ui.Print("    Credentials are served from GCM's encrypted store (immune to external logout)")
			} else {
				ui.Print("  %s GCM is NOT the credential helper for github.com", ui.Yellow(ui.IconWarning))
				ui.Print("    Git credentials use the system keychain (can be affected by VS Code logout, etc.)")
				ui.Print("    Fix: run %s to register GCM as the credential helper", ui.Cyan("gcm init"))
			}

			ui.Blank()
			ui.Success("Health check complete")
			return nil
		},
	}
}

func checkCommand(label, cmd string, args ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fullCmd := exec.CommandContext(ctx, cmd, args...)
	out, err := fullCmd.Output()
	if err != nil {
		ui.Print("  %s %s: not installed", ui.Red(ui.IconError), label)
		return
	}
	ver := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	ui.Print("  %s %s: %s", ui.Green(ui.IconSuccess), label, ver)
}

func checkSSHAgent() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ssh-add", "-l")
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		if strings.Contains(output, "Could not open") || strings.Contains(output, "not running") {
			ui.Print("  %s SSH Agent: not running", ui.Red(ui.IconError))
			return
		}
	}

	if strings.Contains(output, "no identities") {
		ui.Print("  %s SSH Agent: running (no keys loaded)", ui.Yellow(ui.IconWarning))
	} else {
		lines := strings.Split(output, "\n")
		ui.Print("  %s SSH Agent: running (%d keys)", ui.Green(ui.IconSuccess), len(lines))
	}
}
