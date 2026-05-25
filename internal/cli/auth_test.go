package cli

import (
	"context"
	"fmt"
	"strings"
	"testing"

	authsvc "github.com/sijunda/git-config-manager/internal/auth"
	"github.com/sijunda/git-config-manager/internal/profile"
	providerpkg "github.com/sijunda/git-config-manager/internal/provider"
	"github.com/sijunda/git-config-manager/internal/providerclient"
)

type cliAuthFakeVerifier map[string]providerclient.AuthenticatedUser

func (f cliAuthFakeVerifier) VerifyPAT(_ context.Context, _ providerpkg.Definition, token string) (providerclient.AuthenticatedUser, error) {
	if user, ok := f[token]; ok {
		return user, nil
	}
	return providerclient.AuthenticatedUser{}, fmt.Errorf("invalid token")
}

type cliAuthFakeInspector struct {
	inspection GitCredentialInspection
	rejected   bool
}

type GitCredentialInspection = authsvc.GitCredentialInspection

func (f *cliAuthFakeInspector) InspectGitCredential(context.Context, providerpkg.Definition, string) (authsvc.GitCredentialInspection, error) {
	return f.inspection, nil
}

func (f *cliAuthFakeInspector) RejectGitCredential(context.Context, providerpkg.Definition, string) error {
	f.rejected = true
	return nil
}

func withAuthManagerFactory(t *testing.T, manager *authsvc.Manager) {
	t.Helper()
	original := authManagerFactory
	authManagerFactory = func() *authsvc.Manager { return manager }
	t.Cleanup(func() { authManagerFactory = original })
}

func createAuthTestProfile(t *testing.T, name string, providerID providerpkg.ProviderID, username string) *profile.Profile {
	t.Helper()
	p := repairTestProfile(name)
	if providerID != "" {
		p.Providers = map[string]profile.ProviderAccountConfig{string(providerID): {Username: username}}
	}
	if err := ctr.ProfileManager.Create(p); err != nil {
		t.Fatalf("create profile: %v", err)
	}
	return p
}

func TestAuthCommandRegistration(t *testing.T) {
	root := NewRootCmd()
	for _, path := range [][]string{{"auth"}, {"auth", "status"}, {"auth", "inspect"}, {"auth", "adopt"}, {"auth", "logout"}, {"auth", "doctor"}, {"auth", "repair"}} {
		if _, _, err := root.Find(path); err != nil {
			t.Fatalf("missing command %v: %v", path, err)
		}
	}
}

func TestRunAuthStatusUsesSourceAwareResolver(t *testing.T) {
	c := withRepairTestContainer(t)
	p := createAuthTestProfile(t, "work", providerpkg.GitHubID, "octo")
	def, ok := c.ProviderRegistry.Get(providerpkg.GitHubID)
	if !ok {
		t.Fatal("github provider missing")
	}
	if err := saveProviderToken("work", def, p, providerpkg.TokenSet{AccessToken: "gcm-token", AuthMethod: providerpkg.AuthMethodPAT}); err != nil {
		t.Fatalf("save token: %v", err)
	}
	manager := authsvc.NewManager(c.TokenStore, cliAuthFakeVerifier{"gcm-token": {Username: "octo"}})
	manager.ExternalInspector = &cliAuthFakeInspector{}
	withAuthManagerFactory(t, manager)

	output := captureStdout(t, func() {
		if err := runAuthStatus(context.Background(), "work", authStatusOptions{}); err != nil {
			t.Fatalf("runAuthStatus: %v", err)
		}
	})
	if !strings.Contains(output, "authenticated:gcm") || !strings.Contains(output, "gcm-store") {
		t.Fatalf("status output missing source-aware state:\n%s", output)
	}
}

func TestRunAuthAdoptDryRunDoesNotMutateProfile(t *testing.T) {
	c := withRepairTestContainer(t)
	createAuthTestProfile(t, "work", "", "")
	manager := authsvc.NewManager(c.TokenStore, cliAuthFakeVerifier{"external-token": {Username: "octo"}})
	manager.ExternalInspector = &cliAuthFakeInspector{inspection: authsvc.GitCredentialInspection{Credential: authsvc.CredentialStatus{
		Type: "https", Source: authsvc.SourceOSXKeychain, Ownership: authsvc.OwnershipExternal, State: authsvc.StateAuthenticatedExternal, Present: true, Secret: "external-token", Username: "octo",
	}}}
	withAuthManagerFactory(t, manager)

	output := captureStdout(t, func() {
		if err := runAuthAdopt(context.Background(), "work", authAdoptOptions{provider: "github", dryRun: true}); err != nil {
			t.Fatalf("runAuthAdopt: %v", err)
		}
	})
	updated, err := c.ProfileManager.Get("work")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if profile.UsesProvider(updated, providerpkg.GitHubID) {
		t.Fatal("dry-run adopt should not set provider")
	}
	if !strings.Contains(output, "Auth Adopt Dry Run") {
		t.Fatalf("missing dry-run output:\n%s", output)
	}
}

func TestRunAuthLogoutDryRunDoesNotRejectExternalCredential(t *testing.T) {
	c := withRepairTestContainer(t)
	createAuthTestProfile(t, "work", providerpkg.GitHubID, "octo")
	inspector := &cliAuthFakeInspector{inspection: authsvc.GitCredentialInspection{Credential: authsvc.CredentialStatus{
		Type: "https", Source: authsvc.SourceOSXKeychain, Ownership: authsvc.OwnershipExternal, State: authsvc.StateAuthenticatedExternal, Present: true, Secret: "external-token", Username: "octo",
	}}}
	manager := authsvc.NewManager(c.TokenStore, cliAuthFakeVerifier{"external-token": {Username: "octo"}})
	manager.ExternalInspector = inspector
	withAuthManagerFactory(t, manager)

	output := captureStdout(t, func() {
		if err := runAuthLogout(context.Background(), "work", authLogoutOptions{scope: "external", dryRun: true}); err != nil {
			t.Fatalf("runAuthLogout: %v", err)
		}
	})
	if inspector.rejected {
		t.Fatal("dry-run logout should not reject external credential")
	}
	if !strings.Contains(output, "Auth Logout Dry Run") || !strings.Contains(output, "osxkeychain") {
		t.Fatalf("missing logout dry-run details:\n%s", output)
	}
}
