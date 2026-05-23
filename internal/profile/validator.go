package profile

import (
	"fmt"
	"net/mail"
	"os"
	"regexp"
	"time"

	fileSvc "git-config-manager/internal/service/file"
)

// ValidateProfile validates a profile configuration.
func ValidateProfile(p *Profile) error {
	if p.Name == "" {
		return ErrProfileNameEmpty()
	}

	if !isValidProfileName(p.Name) {
		return &ProfileError{
			Code:       ErrCodeInvalid,
			Message:    fmt.Sprintf("invalid profile name %q", p.Name),
			Profile:    p.Name,
			Suggestion: "Profile names must be alphanumeric with hyphens or underscores",
		}
	}

	if err := validateGitConfig(&p.Git); err != nil {
		return err
	}

	if p.SSH != nil {
		if err := validateSSHConfig(p.SSH); err != nil {
			return err
		}
	}

	if p.GPG != nil {
		if err := validateGPGConfig(p.GPG); err != nil {
			return err
		}
	}

	return nil
}

func validateGitConfig(g *GitConfig) error {
	if g.User.Name == "" {
		return ErrGitUserNameEmpty()
	}
	if g.User.Email == "" {
		return ErrGitUserEmailEmpty()
	}
	if !isValidEmail(g.User.Email) {
		return ErrGitUserEmailInvalid()
	}
	return nil
}

func validateSSHConfig(s *SSHConfig) error {
	if s.KeyPath == "" {
		return &ProfileError{
			Code:       ErrCodeSSHKeyPath,
			Message:    "SSH key path cannot be empty",
			Suggestion: "Provide a key path, e.g.: ~/.ssh/id_ed25519_work",
		}
	}
	return nil
}

func validateGPGConfig(g *GPGConfig) error {
	if g.KeyID == "" {
		return &ProfileError{
			Code:       ErrCodeGPGKeyID,
			Message:    "GPG key ID cannot be empty",
			Suggestion: "Provide a GPG key ID or generate one: gcm gpg generate <profile>",
		}
	}
	return nil
}

// ValidationResult holds validation results for a profile.
type ValidationResult struct {
	Profile  string
	Errors   []ValidationIssue
	Warnings []ValidationIssue
	Info     []ValidationIssue
}

// ValidationIssue represents a single validation finding.
type ValidationIssue struct {
	Category   string
	Message    string
	Suggestion string
}

// IsValid returns true if there are no errors.
func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// ValidateDeep performs comprehensive validation including file system checks.
func ValidateDeep(p *Profile) *ValidationResult {
	result := &ValidationResult{Profile: p.Name}

	// Git configuration
	if p.Git.User.Name != "" {
		result.Info = append(result.Info, ValidationIssue{
			Category: "Git", Message: fmt.Sprintf("user.name is set: %s", p.Git.User.Name),
		})
	} else {
		result.Errors = append(result.Errors, ValidationIssue{
			Category: "Git", Message: "user.name is not set",
			Suggestion: "Set it with: gcm profile edit " + p.Name + " --name \"Your Name\"",
		})
	}

	if p.Git.User.Email != "" && isValidEmail(p.Git.User.Email) {
		result.Info = append(result.Info, ValidationIssue{
			Category: "Git", Message: fmt.Sprintf("user.email is valid: %s", p.Git.User.Email),
		})
	} else if p.Git.User.Email != "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Category: "Git", Message: "user.email has invalid format",
			Suggestion: "Update email: gcm profile edit " + p.Name + " --email user@example.com",
		})
	} else {
		result.Errors = append(result.Errors, ValidationIssue{
			Category: "Git", Message: "user.email is not set",
		})
	}

	// SSH configuration
	if p.SSH != nil && p.SSH.KeyPath != "" {
		expanded := expandPath(p.SSH.KeyPath)
		if _, err := os.Stat(expanded); err == nil {
			result.Info = append(result.Info, ValidationIssue{
				Category: "SSH", Message: "SSH key exists",
			})
			// Check permissions
			info, _ := os.Stat(expanded)
			if info != nil && info.Mode().Perm()&0077 != 0 {
				result.Warnings = append(result.Warnings, ValidationIssue{
					Category: "SSH", Message: "SSH key has overly permissive permissions",
					Suggestion: fmt.Sprintf("Fix: chmod 600 %s", p.SSH.KeyPath),
				})
			}
		} else {
			result.Errors = append(result.Errors, ValidationIssue{
				Category: "SSH", Message: "SSH key file not found",
				Suggestion: fmt.Sprintf("Generate: gcm ssh generate %s", p.Name),
			})
		}
	}

	// GPG configuration
	if p.GPG != nil && p.GPG.KeyID != "" {
		result.Info = append(result.Info, ValidationIssue{
			Category: "GPG", Message: fmt.Sprintf("GPG key ID set: %s", p.GPG.KeyID),
		})
		if p.GPG.ExpiresAt != nil {
			if time.Now().After(*p.GPG.ExpiresAt) {
				result.Errors = append(result.Errors, ValidationIssue{
					Category:   "GPG",
					Message:    fmt.Sprintf("GPG key expired on %s", p.GPG.ExpiresAt.Format("2006-01-02")),
					Suggestion: fmt.Sprintf("Generate a new key: gcm gpg generate %s", p.Name),
				})
			} else if time.Until(*p.GPG.ExpiresAt) < 30*24*time.Hour {
				result.Warnings = append(result.Warnings, ValidationIssue{
					Category:   "GPG",
					Message:    fmt.Sprintf("GPG key expires soon: %s", p.GPG.ExpiresAt.Format("2006-01-02")),
					Suggestion: fmt.Sprintf("Consider renewing: gcm gpg generate %s", p.Name),
				})
			} else {
				result.Info = append(result.Info, ValidationIssue{
					Category: "GPG", Message: fmt.Sprintf("Expires: %s", p.GPG.ExpiresAt.Format("2006-01-02")),
				})
			}
		}
	}

	return result
}

var profileNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

func isValidProfileName(name string) bool {
	return profileNameRegex.MatchString(name)
}

func isValidEmail(email string) bool {
	// Guard against pathological input that can trigger quadratic behaviour
	// in net/mail.ParseAddress (GO-2026-4986, GO-2026-4977). RFC 5321 limits
	// the total address to 254 octets.
	if len(email) > 254 {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func expandPath(path string) string {
	return fileSvc.ExpandPath(path)
}
