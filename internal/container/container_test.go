package container

import (
	"os"
	"testing"

	"github-config-manager/internal/config"
	"github-config-manager/pkg/logger"
)

func TestNew(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := config.DefaultConfig()
	cfg.ProfilesDir = tmp + "/profiles"
	cfg.TemplatesDir = tmp + "/templates"
	cfg.CacheDir = tmp + "/cache"
	cfg.SSHDir = tmp + "/ssh"
	cfg.GPGHome = tmp + "/gpg"

	os.MkdirAll(cfg.ProfilesDir, 0o755)
	os.MkdirAll(cfg.TemplatesDir, 0o755)

	log := logger.New(logger.LevelError, os.Stderr)
	ctr := New(cfg, log)

	if ctr == nil {
		t.Fatal("expected non-nil container")
	}
	if ctr.Config != cfg {
		t.Error("config not set")
	}
	if ctr.Logger != log {
		t.Error("logger not set")
	}
	if ctr.AuditLogger == nil {
		t.Error("audit logger not initialized")
	}
	if ctr.FileService == nil {
		t.Error("file service not initialized")
	}
	if ctr.CryptoService == nil {
		t.Error("crypto service not initialized")
	}
	if ctr.ProfileManager == nil {
		t.Error("profile manager not initialized")
	}
	if ctr.ProfileSwitcher == nil {
		t.Error("profile switcher not initialized")
	}
	if ctr.SSHManager == nil {
		t.Error("SSH manager not initialized")
	}
	if ctr.GPGManager == nil {
		t.Error("GPG manager not initialized")
	}
	if ctr.GitHubClient == nil {
		t.Error("GitHub client not initialized")
	}
	if ctr.ShellManager == nil {
		t.Error("shell manager not initialized")
	}
	if ctr.TemplateManager == nil {
		t.Error("template manager not initialized")
	}
	if ctr.BackupManager == nil {
		t.Error("backup manager not initialized")
	}
}

func TestSetMasterPasswordPrompt(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := config.DefaultConfig()
	cfg.ProfilesDir = tmp + "/profiles"
	cfg.TemplatesDir = tmp + "/templates"
	cfg.CacheDir = tmp + "/cache"
	cfg.SSHDir = tmp + "/ssh"
	cfg.GPGHome = tmp + "/gpg"

	os.MkdirAll(cfg.ProfilesDir, 0o755)
	os.MkdirAll(cfg.TemplatesDir, 0o755)

	log := logger.New(logger.LevelError, os.Stderr)
	ctr := New(cfg, log)

	// Verify SetMasterPasswordPrompt doesn't panic and sets the function
	ctr.SetMasterPasswordPrompt(func(_ string) (string, error) {
		return "secret", nil
	})
}
