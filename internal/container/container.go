// Package container provides dependency injection for GCM.
package container

import (
	"git-config-manager/internal/audit"
	"git-config-manager/internal/backup"
	"git-config-manager/internal/config"
	"git-config-manager/internal/github"
	"git-config-manager/internal/gpg"
	"git-config-manager/internal/profile"
	cryptoSvc "git-config-manager/internal/service/crypto"
	fileSvc "git-config-manager/internal/service/file"
	"git-config-manager/internal/shell"
	"git-config-manager/internal/ssh"
	"git-config-manager/internal/template"
	"git-config-manager/pkg/logger"
)

// Container holds all application dependencies.
type Container struct {
	Config          *config.Config
	Logger          *logger.Logger
	AuditLogger     *audit.Logger
	FileService     *fileSvc.Service
	CryptoService   *cryptoSvc.Service
	ProfileManager  *profile.Manager
	ProfileSwitcher *profile.Switcher
	SSHManager      *ssh.Manager
	GPGManager      *gpg.Manager
	GitHubClient    *github.Client
	TokenStore      *github.TokenStore
	ShellManager    *shell.Manager
	TemplateManager *template.Manager
	BackupManager   *backup.Manager
}

// SetMasterPasswordPrompt injects the callback used to ask the user for a
// master password when encrypted-file token storage is active. This must be
// called before any Save/Load operation that requires a master password.
func (c *Container) SetMasterPasswordPrompt(fn github.PromptFunc) {
	c.TokenStore.SetPromptFunc(fn)
}

// New creates a fully-wired Container from the loaded configuration.
func New(cfg *config.Config, log *logger.Logger) *Container {
	fs := fileSvc.NewService()
	crypto := cryptoSvc.NewService()
	auditLog := audit.NewLogger(cfg)

	tokenStore := github.NewTokenStore(cfg, crypto, log, nil)
	ghClient := github.NewClient(cfg, log, tokenStore)

	pm := profile.NewManager(cfg, fs, log)
	ps := profile.NewSwitcher(cfg, pm, log)
	sshMgr := ssh.NewManager(cfg, log)
	gpgMgr := gpg.NewManager(cfg, log)
	shellMgr := shell.NewManager(log)
	tmplMgr := template.NewManager(cfg, fs, log)
	bkpMgr := backup.NewManager(cfg, log)

	return &Container{
		Config:          cfg,
		Logger:          log,
		AuditLogger:     auditLog,
		FileService:     fs,
		CryptoService:   crypto,
		ProfileManager:  pm,
		ProfileSwitcher: ps,
		SSHManager:      sshMgr,
		GPGManager:      gpgMgr,
		GitHubClient:    ghClient,
		TokenStore:      tokenStore,
		ShellManager:    shellMgr,
		TemplateManager: tmplMgr,
		BackupManager:   bkpMgr,
	}
}
