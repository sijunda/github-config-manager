package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// requireArgs returns a Cobra PositionalArgs validator that gives a clear,
// user-friendly error message when the wrong number of arguments is provided.
// It shows what's expected and gives an example usage.
func requireArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == n {
			return nil
		}
		if len(args) < n {
			// Missing arguments
			return fmt.Errorf("missing required argument\n\n  Usage: %s\n\n  Run 'gcm %s --help' for more information.",
				cmd.UseLine(), cmd.CommandPath()[4:]) // skip "gcm " prefix
		}
		// Too many arguments
		return fmt.Errorf("too many arguments provided\n\n  Usage: %s\n\n  Run 'gcm %s --help' for more information.",
			cmd.UseLine(), cmd.CommandPath()[4:]) // skip "gcm " prefix
	}
}

// formatTimeAgo formats a time.Time as a human-friendly "X ago" string.
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / 24 / 7)
		return fmt.Sprintf("%dw ago", weeks)
	default:
		return t.Format("2006-01-02")
	}
}

// isStdinPiped returns true if stdin is being piped (not a terminal).
func isStdinPiped() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

// readStdinToken reads a single line (token) from stdin, trimming whitespace.
func readStdinToken() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			return "", fmt.Errorf("empty input")
		}
		return token, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no input received")
}
