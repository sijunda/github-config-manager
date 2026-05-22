// Package ssh provides SSH key management for GCM.
package ssh

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github-config-manager/internal/config"
	fileSvc "github-config-manager/internal/service/file"
	"github-config-manager/pkg/logger"
)

// defaultCommandTimeout is applied to external ssh/ssh-add/ssh-keygen calls
// to avoid hangs if the process stalls (e.g. agent prompting for input).
const defaultCommandTimeout = 30 * time.Second

// Test hooks for unreachable OS/IO/crypto error paths.
var (
	sshMkdirFn                    = os.MkdirAll
	sshRandReader       io.Reader = rand.Reader
	sshOpenFileFn                 = os.OpenFile
	sshAbsFn                      = filepath.Abs
	sshLookPathFn                 = exec.LookPath
	sshMarshalPrivKeyFn           = func(key interface{}, comment string) (*pem.Block, error) {
		return ssh.MarshalPrivateKey(key, comment)
	}
	sshMarshalPrivKeyPassFn = func(key interface{}, comment string, passphrase []byte) (*pem.Block, error) {
		return ssh.MarshalPrivateKeyWithPassphrase(key, comment, passphrase)
	}
	sshNewPublicKeyFn = ssh.NewPublicKey
	sshWriteFileFn    = os.WriteFile
	sshFileWriteFn    = func(f *os.File, data []byte) (int, error) { return f.Write(data) }
)

// Manager handles SSH key operations.
type Manager struct {
	cfg *config.Config
	log *logger.Logger
}

// NewManager creates a new SSH manager.
func NewManager(cfg *config.Config, log *logger.Logger) *Manager {
	return &Manager{cfg: cfg, log: log}
}

// GenerateOptions holds parameters for key generation.
type GenerateOptions struct {
	Profile    string
	KeyType    string // ed25519, rsa, ecdsa
	Bits       int    // For RSA, default 4096
	Comment    string
	Passphrase string
}

// KeyInfo holds information about an SSH key.
type KeyInfo struct {
	Path        string
	Type        string
	Fingerprint string
	Comment     string
	PublicKey   string
	InAgent     bool
}

// Generate creates a new SSH key pair.
//
// Key generation is done natively using golang.org/x/crypto/ssh so the
// passphrase never appears on a command line argv (where it would be visible
// to other users via `ps`). The private key is written with 0600 permissions
// and the public key with 0644, inside a directory created with 0700.
func (m *Manager) Generate(opts GenerateOptions) (*KeyInfo, error) {
	if opts.KeyType == "" {
		opts.KeyType = "ed25519"
	}
	if opts.Comment == "" {
		opts.Comment = fmt.Sprintf("gcm-%s", opts.Profile)
	}

	keyPath, err := m.keyPath(opts.Profile, opts.KeyType)
	if err != nil {
		return nil, err
	}

	// Check if key already exists (refuse to overwrite).
	if _, err := os.Stat(keyPath); err == nil {
		return nil, fmt.Errorf("SSH key already exists at %s", keyPath)
	}

	// Ensure .ssh directory exists with restrictive permissions.
	dir := filepath.Dir(keyPath)
	if err := sshMkdirFn(dir, 0o700); err != nil {
		return nil, fmt.Errorf("creating SSH directory: %w", err)
	}

	priv, pub, err := generateKeyPair(opts.KeyType, opts.Bits)
	if err != nil {
		return nil, fmt.Errorf("generating %s key: %w", opts.KeyType, err)
	}

	// Marshal private key in OpenSSH PEM format. If a passphrase is supplied,
	// the key is encrypted at rest with bcrypt-KDF + AES-256-CTR (OpenSSH
	// native format), never exposing the passphrase via argv.
	var pemBlock *pem.Block
	if opts.Passphrase != "" {
		pemBlock, err = sshMarshalPrivKeyPassFn(priv, opts.Comment, []byte(opts.Passphrase))
	} else {
		pemBlock, err = sshMarshalPrivKeyFn(priv, opts.Comment)
	}
	if err != nil {
		return nil, fmt.Errorf("encoding private key: %w", err)
	}

	pubKey, err := sshNewPublicKeyFn(pub)
	if err != nil {
		return nil, fmt.Errorf("deriving public key: %w", err)
	}
	authorizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pubKey)))
	pubLine := fmt.Sprintf("%s %s\n", authorizedKey, opts.Comment)

	if err := writePrivateKey(keyPath, pem.EncodeToMemory(pemBlock)); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath+".pub", []byte(pubLine), 0o644); err != nil {
		// Best-effort cleanup of the private key if public key write fails.
		_ = os.Remove(keyPath)
		return nil, fmt.Errorf("writing public key: %w", err)
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)

	m.log.Debug("SSH key generated",
		logger.F("profile", opts.Profile),
		logger.F("type", opts.KeyType),
		logger.F("path", keyPath))

	return &KeyInfo{
		Path:        keyPath,
		Type:        opts.KeyType,
		Fingerprint: fingerprint,
		Comment:     opts.Comment,
		PublicKey:   strings.TrimSpace(pubLine),
	}, nil
}

// generateKeyPair generates an SSH key pair for the given type. Returned
// values are suitable for ssh.MarshalPrivateKey and ssh.NewPublicKey.
func generateKeyPair(keyType string, bits int) (priv any, pub any, err error) {
	switch strings.ToLower(keyType) {
	case "ed25519", "":
		pubK, privK, err := ed25519.GenerateKey(sshRandReader)
		if err != nil {
			return nil, nil, err
		}
		return privK, pubK, nil

	case "rsa":
		if bits == 0 {
			bits = 4096
		}
		if bits < 2048 {
			return nil, nil, fmt.Errorf("RSA key size must be at least 2048 bits")
		}
		// In Go 1.20+ GenerateKey uses an internal CSPRNG and cannot fail.
		k, _ := rsa.GenerateKey(sshRandReader, bits)
		return k, &k.PublicKey, nil

	case "ecdsa":
		var curve elliptic.Curve
		switch bits {
		case 0, 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			return nil, nil, fmt.Errorf("unsupported ECDSA curve size: %d (use 256, 384, or 521)", bits)
		}
		// In Go 1.20+ GenerateKey uses an internal CSPRNG and cannot fail.
		k, _ := ecdsa.GenerateKey(curve, sshRandReader)
		return k, &k.PublicKey, nil

	default:
		return nil, nil, fmt.Errorf("unsupported key type: %s (use ed25519, rsa, or ecdsa)", keyType)
	}
}

// writePrivateKey writes private-key bytes to path with 0600 permissions using
// O_EXCL so we never overwrite an existing file.
func writePrivateKey(path string, data []byte) error {
	f, err := sshOpenFileFn(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("creating private key: %w", err)
	}
	if _, err := sshFileWriteFn(f, data); err != nil {
		f.Close()
		_ = os.Remove(path)
		return fmt.Errorf("writing private key: %w", err)
	}
	return f.Close()
}

// List returns info for all SSH keys in the SSH directory.
func (m *Manager) List() ([]KeyInfo, error) {
	sshDir := m.cfg.SSHDir
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading SSH directory: %w", err)
	}

	agentKeys := m.getAgentKeys()
	var keys []KeyInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}

		pubPath := filepath.Join(sshDir, entry.Name())
		privPath := strings.TrimSuffix(pubPath, ".pub")

		// Read public key
		pubData, err := os.ReadFile(pubPath)
		if err != nil {
			continue
		}

		fingerprint, _ := m.getFingerprint(privPath)

		// Determine key type from public key content
		parts := strings.Fields(string(pubData))
		keyType := ""
		comment := ""
		if len(parts) >= 1 {
			keyType = strings.TrimPrefix(parts[0], "ssh-")
		}
		if len(parts) >= 3 {
			comment = parts[2]
		}

		inAgent := false
		for _, ak := range agentKeys {
			if strings.Contains(ak, fingerprint) || strings.Contains(ak, entry.Name()) {
				inAgent = true
				break
			}
		}

		keys = append(keys, KeyInfo{
			Path:        privPath,
			Type:        keyType,
			Fingerprint: fingerprint,
			Comment:     comment,
			PublicKey:   strings.TrimSpace(string(pubData)),
			InAgent:     inAgent,
		})
	}

	return keys, nil
}

// TestConnection tests SSH connectivity to GitHub.
func (m *Manager) TestConnection(keyPath string) (string, error) {
	expanded := expandPath(keyPath)

	// Best-effort: load key into agent first (short timeout).
	if _, err := exec.LookPath("ssh-add"); err == nil {
		addCtx, addCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer addCancel()
		_ = exec.CommandContext(addCtx, "ssh-add", expanded).Run()
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, m.cfg.Advanced.SSHCommand,
		"-T", "-i", expanded,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		"-o", "IdentitiesOnly=yes",
		"git@github.com",
	)

	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	// GitHub returns exit code 1 even on success with a greeting
	if strings.Contains(output, "successfully authenticated") ||
		strings.Contains(output, "Hi ") {
		return output, nil
	}

	if err != nil {
		return output, fmt.Errorf("SSH test failed: %s", output)
	}

	return output, nil
}

// AddToAgent loads an SSH key into the ssh-agent.
func (m *Manager) AddToAgent(keyPath string) error {
	if _, err := exec.LookPath("ssh-add"); err != nil {
		return fmt.Errorf("ssh-add not found — SSH agent is not available on this system")
	}
	expanded := expandPath(keyPath)
	ctx, cancel := context.WithTimeout(context.Background(), defaultCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ssh-add", expanded)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ssh-add failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// RemoveFromAgent removes an SSH key from the ssh-agent.
func (m *Manager) RemoveFromAgent(keyPath string) error {
	if _, err := exec.LookPath("ssh-add"); err != nil {
		return nil // ssh-agent not available, nothing to do
	}
	expanded := expandPath(keyPath)
	ctx, cancel := context.WithTimeout(context.Background(), defaultCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ssh-add", "-d", expanded)
	_ = cmd.Run() // ignore error — key might not be loaded
	return nil
}

// GetPublicKey reads and returns the public key content.
func (m *Manager) GetPublicKey(keyPath string) (string, error) {
	expanded := expandPath(keyPath)
	pubPath := expanded + ".pub"
	data, err := os.ReadFile(pubPath)
	if err != nil {
		return "", fmt.Errorf("reading public key: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func (m *Manager) keyPath(profile, keyType string) (string, error) {
	if strings.ContainsAny(profile, `/\`) || strings.Contains(profile, "..") || profile == "" {
		return "", fmt.Errorf("invalid profile name %q", profile)
	}
	filename := fmt.Sprintf("id_%s_%s", keyType, profile)
	full := filepath.Join(m.cfg.SSHDir, filename)
	// Verify the result is still under SSHDir.
	sshAbs, err := sshAbsFn(m.cfg.SSHDir)
	if err != nil {
		return "", fmt.Errorf("resolving SSH dir: %w", err)
	}
	rel, err := filepath.Rel(sshAbs, filepath.Clean(full))
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid profile name %q", profile)
	}
	return full, nil
}

func (m *Manager) getFingerprint(keyPath string) (string, error) {
	// Prefer native fingerprint calculation from the public key bytes so we
	// don't depend on an external ssh-keygen being installed. Fall back to
	// ssh-keygen if the public key can't be parsed.
	pubBytes, err := os.ReadFile(keyPath + ".pub")
	if err == nil {
		if pk, _, _, _, parseErr := ssh.ParseAuthorizedKey(pubBytes); parseErr == nil {
			return ssh.FingerprintSHA256(pk), nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ssh-keygen", "-lf", keyPath+".pub")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	parts := strings.Fields(string(out))
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return "", fmt.Errorf("unexpected fingerprint output")
}

func (m *Manager) getAgentKeys() []string {
	if _, err := sshLookPathFn("ssh-add"); err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ssh-add", "-l")
	out, _ := cmd.Output()
	return strings.Split(string(out), "\n")
}

func expandPath(path string) string {
	return fileSvc.ExpandPath(path)
}
