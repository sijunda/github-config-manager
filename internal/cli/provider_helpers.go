package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git-config-manager/internal/profile"
	providerpkg "git-config-manager/internal/provider"
	"git-config-manager/pkg/ui"
)

func providerAccountForProfile(p *profile.Profile, id providerpkg.ProviderID) profile.ProviderAccountConfig {
	if p != nil && p.Providers != nil {
		if account, ok := p.Providers[string(id)]; ok {
			return account
		}
	}
	if id == providerpkg.GitHubID && p != nil && p.GitHub != nil {
		return profile.ProviderAccountConfig{
			Username:   p.GitHub.Username,
			TokenPath:  p.GitHub.TokenPath,
			AuthMethod: providerpkg.AuthMethodLegacy,
			UploadKeys: p.GitHub.UploadKeys,
		}
	}
	return profile.ProviderAccountConfig{}
}

func setProfileProviderAccount(p *profile.Profile, id providerpkg.ProviderID, username, authMethod string) {
	previous := providerAccountForProfile(p, id)
	p.Providers = make(map[string]profile.ProviderAccountConfig)
	account := previous
	account.Username = username
	account.AuthMethod = authMethod
	p.Providers[string(id)] = account

	if id == providerpkg.GitHubID {
		if p.GitHub == nil {
			p.GitHub = &profile.GitHubConfig{}
		}
		p.GitHub.Username = username
	} else {
		p.GitHub = nil
	}
}

func profileProviderID(p *profile.Profile) (providerpkg.ProviderID, bool) {
	if p == nil {
		return "", false
	}
	if profileHasMultipleProviders(p) {
		return "", false
	}
	if len(p.Providers) == 1 {
		for id := range p.Providers {
			return providerpkg.ProviderID(id), true
		}
	}
	if p.GitHub != nil {
		return providerpkg.GitHubID, true
	}
	return "", false
}

func profileProviderDefinition(p *profile.Profile, capability providerpkg.Capability) (providerpkg.Definition, bool) {
	id, ok := profileProviderID(p)
	if !ok || ctr == nil || ctr.ProviderRegistry == nil {
		return providerpkg.Definition{}, false
	}
	def, ok := ctr.ProviderRegistry.Get(id)
	if !ok || !def.Capabilities.Has(capability) {
		return providerpkg.Definition{}, false
	}
	return def, true
}

func profileUsesProvider(p *profile.Profile, id providerpkg.ProviderID) bool {
	profileProviderID, ok := profileProviderID(p)
	return ok && profileProviderID == id
}

func profileHasMultipleProviders(p *profile.Profile) bool {
	if p == nil {
		return false
	}
	if len(p.Providers) > 1 {
		return true
	}
	if len(p.Providers) == 1 && p.GitHub != nil {
		_, hasGitHubProvider := p.Providers[string(providerpkg.GitHubID)]
		return !hasGitHubProvider
	}
	return false
}

func providerTokenKey(profileName string, def providerpkg.Definition, account profile.ProviderAccountConfig) providerpkg.TokenKey {
	return providerpkg.TokenKey{
		Profile:  profileName,
		Provider: def.ID,
		Host:     firstProviderHost(def),
		Account:  account.Account,
	}
}

func firstProviderHost(def providerpkg.Definition) string {
	if len(def.GitHosts) > 0 && def.GitHosts[0] != "" {
		return providerpkg.NormalizeHost(def.GitHosts[0])
	}
	if def.WebURL != "" {
		return providerpkg.NormalizeHost(def.WebURL)
	}
	return providerpkg.NormalizeHost(def.APIURL)
}

func loadProviderToken(profileName string, def providerpkg.Definition, p *profile.Profile) (providerpkg.TokenSet, error) {
	account := providerAccountForProfile(p, def.ID)
	return ctr.TokenStore.LoadTokenSet(providerTokenKey(profileName, def, account))
}

func saveProviderToken(profileName string, def providerpkg.Definition, p *profile.Profile, token providerpkg.TokenSet) error {
	account := providerAccountForProfile(p, def.ID)
	return ctr.TokenStore.SaveTokenSet(providerTokenKey(profileName, def, account), token)
}

func deleteProviderToken(profileName string, def providerpkg.Definition, p *profile.Profile) error {
	account := providerAccountForProfile(p, def.ID)
	return ctr.TokenStore.DeleteTokenSet(providerTokenKey(profileName, def, account))
}

func configureGitCredentialsForProvider(profileName string, p *profile.Profile, def providerpkg.Definition, token providerpkg.TokenSet) {
	server := def.CredentialServer()
	account := providerAccountForProfile(p, def.ID)
	username := def.CredentialUsername(profileName, account.Username, token)
	clearGitCredentialsForOtherProviders(def)

	if IsCredentialHelperConfiguredFor(server) {
		_ = ctr.GitHubClient.SetGitCredentialUsername(server, username)
		return
	}

	_ = ctr.GitHubClient.ClearGitCredentials(server)
	_ = ctr.GitHubClient.StoreGitCredentials(server, username, token.AccessToken)
	_ = ctr.GitHubClient.SetGitCredentialUsername(server, username)
}

func clearGitCredentialsForOtherProviders(active providerpkg.Definition) {
	if ctr == nil || ctr.ProviderRegistry == nil {
		return
	}
	for _, def := range ctr.ProviderRegistry.All() {
		if def.ID == active.ID || !def.Capabilities.Has(providerpkg.CapabilityCredentialHelper) {
			continue
		}
		_ = ctr.GitHubClient.ClearGitCredentials(def.CredentialServer())
	}
}

func clearProfileProviderAccount(p *profile.Profile, id providerpkg.ProviderID) {
	if p == nil {
		return
	}
	if p.Providers != nil {
		delete(p.Providers, string(id))
		if len(p.Providers) == 0 {
			p.Providers = nil
		}
	}
	if id == providerpkg.GitHubID {
		p.GitHub = nil
	}
}

func clearAllProfileProviderAccounts(p *profile.Profile) {
	if p == nil {
		return
	}
	p.Providers = nil
	p.GitHub = nil
}

func providerDefinitionsWithCapability(capability providerpkg.Capability) []providerpkg.Definition {
	if ctr == nil || ctr.ProviderRegistry == nil {
		return nil
	}
	var defs []providerpkg.Definition
	for _, def := range ctr.ProviderRegistry.All() {
		if def.Capabilities.Has(capability) {
			defs = append(defs, def)
		}
	}
	return defs
}

func selectProfileProviderWithCapability(profileName string, p *profile.Profile, requested string, capability providerpkg.Capability) (providerpkg.Definition, error) {
	def, ok := profileProviderDefinition(p, capability)
	if !ok {
		return providerpkg.Definition{}, fmt.Errorf("profile %q has no provider for this operation; set one with: gcm profile edit %s -i", profileName, profileName)
	}
	if requested == "" {
		return def, nil
	}
	requestedID := normalizeProviderSelection(requested)
	if requestedID != def.ID {
		return providerpkg.Definition{}, fmt.Errorf("profile %q is configured for %s, not %s", profileName, def.DisplayName, requested)
	}
	return def, nil
}

func requireProfileProvider(profileName string, p *profile.Profile, def providerpkg.Definition) error {
	if profileHasMultipleProviders(p) {
		return fmt.Errorf("profile %q has multiple providers configured; choose exactly one with: gcm profile edit %s -i", profileName, profileName)
	}
	if !profileUsesProvider(p, def.ID) {
		return fmt.Errorf("profile %q is not configured for %s; run: gcm %s login %s", profileName, def.DisplayName, def.ID, profileName)
	}
	return nil
}

func normalizeProviderSelection(value string) providerpkg.ProviderID {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "gh", "github":
		return providerpkg.GitHubID
	case "gl", "gitlab":
		return providerpkg.GitLabID
	case "bb", "bitbucket":
		return providerpkg.BitbucketID
	default:
		return providerpkg.ProviderID(strings.ToLower(strings.TrimSpace(value)))
	}
}

func providerOption(def providerpkg.Definition) string {
	host := firstProviderHost(def)
	if host == "" {
		return def.DisplayName
	}
	return fmt.Sprintf("%s (%s)", def.DisplayName, host)
}

func setupUploadKeys(ctx context.Context, profileName string) {
	p, err := ctr.ProfileManager.Get(profileName)
	if err != nil || p == nil {
		return
	}
	def, ok := profileProviderDefinition(p, providerpkg.CapabilitySSHKeys)
	if !ok {
		return
	}
	setupUploadKeysForProvider(ctx, profileName, p, def)
}

func setupUploadKeysForProvider(ctx context.Context, profileName string, p *profile.Profile, def providerpkg.Definition) {
	if p == nil || !def.UploadKeys {
		return
	}

	uploaded := false
	if p.SSH != nil && p.SSH.KeyPath != "" {
		pubKey, pubErr := ctr.SSHManager.GetPublicKey(p.SSH.KeyPath)
		if pubErr == nil && pubKey != "" {
			uploaded = setupSSHKeyUploadForProvider(ctx, profileName, p, def, pubKey, string(p.SSH.KeyType))
		}
	}

	if p.GPG != nil && p.GPG.KeyID != "" {
		if !uploaded {
			ui.Blank()
		}
		setupGPGKeyUploadForProvider(ctx, profileName, p, def, p.GPG.KeyID)
	}
}

func setupSSHKeyUploadForProvider(ctx context.Context, profileName string, p *profile.Profile, def providerpkg.Definition, publicKey, keyType string) bool {
	if p == nil || !def.UploadKeys || publicKey == "" {
		return false
	}
	token, err := loadProviderToken(profileName, def, p)
	if err != nil || token.AccessToken == "" {
		return false
	}
	if err := setProviderToken(def, token); err != nil {
		return false
	}

	exists, checkErr := providerSSHKeyExists(ctx, def, publicKey)
	if checkErr == nil && exists {
		ui.Blank()
		ui.Success("SSH key already on %s", def.DisplayName)
		return false
	}
	if checkErr != nil {
		return false
	}

	ui.Blank()
	upload, askErr := ui.AskConfirm(fmt.Sprintf("Upload SSH key to %s?", def.DisplayName), true)
	if askErr != nil || !upload {
		return false
	}
	title := providerResourceName(profileName, def, "ssh", keyType)
	if uploadErr := uploadProviderSSHKey(ctx, def, title, publicKey); uploadErr != nil {
		ui.Warning("Could not upload SSH key to %s: %v", def.DisplayName, uploadErr)
		return false
	}
	ui.Success("SSH key uploaded to %s", def.DisplayName)
	ui.Detail("Title", title)
	return true
}

func setupGPGKeyUploadForProvider(ctx context.Context, profileName string, p *profile.Profile, def providerpkg.Definition, keyID string) bool {
	if p == nil || !def.UploadKeys || keyID == "" {
		return false
	}
	token, err := loadProviderToken(profileName, def, p)
	if err != nil || token.AccessToken == "" {
		return false
	}
	if err := setProviderToken(def, token); err != nil {
		return false
	}

	exists, checkErr := providerGPGKeyExists(ctx, def, keyID)
	if checkErr == nil && exists {
		ui.Success("GPG key already on %s", def.DisplayName)
		return false
	}
	if checkErr != nil {
		return false
	}

	upload, askErr := ui.AskConfirm(fmt.Sprintf("Upload GPG key to %s?", def.DisplayName), true)
	if askErr != nil || !upload {
		return false
	}
	pubKey, gpgErr := ctr.GPGManager.GetPublicKey(keyID)
	if gpgErr != nil {
		ui.Warning("Could not read GPG public key: %v", gpgErr)
		return false
	}
	if uploadErr := uploadProviderGPGKey(ctx, def, pubKey); uploadErr != nil {
		ui.Warning("Could not upload GPG key to %s: %v", def.DisplayName, uploadErr)
		return false
	}
	ui.Success("GPG key uploaded to %s", def.DisplayName)
	return true
}

func authenticatedProvidersForProfile(profileName string, p *profile.Profile, capability providerpkg.Capability) []providerpkg.Definition {
	def, ok := profileProviderDefinition(p, capability)
	if !ok {
		return nil
	}
	token, err := loadProviderToken(profileName, def, p)
	if err != nil || token.AccessToken == "" {
		return nil
	}
	return []providerpkg.Definition{def}
}

func setProviderToken(def providerpkg.Definition, token providerpkg.TokenSet) error {
	switch def.ID {
	case providerpkg.GitHubID:
		ctr.GitHubClient.SetToken(token.AccessToken)
		return nil
	case providerpkg.GitLabID:
		ctr.GitLabClient.SetTokenSet(token)
		return nil
	default:
		return fmt.Errorf("provider %q is not implemented", def.ID)
	}
}

func providerSSHKeyExists(ctx context.Context, def providerpkg.Definition, publicKey string) (bool, error) {
	switch def.ID {
	case providerpkg.GitHubID:
		return ctr.GitHubClient.SSHKeyExists(ctx, publicKey)
	case providerpkg.GitLabID:
		return ctr.GitLabClient.SSHKeyExists(ctx, publicKey)
	default:
		return false, fmt.Errorf("provider %q does not support SSH key upload yet", def.ID)
	}
}

func uploadProviderSSHKey(ctx context.Context, def providerpkg.Definition, title, publicKey string) error {
	switch def.ID {
	case providerpkg.GitHubID:
		return ctr.GitHubClient.UploadSSHKey(ctx, title, publicKey)
	case providerpkg.GitLabID:
		return ctr.GitLabClient.UploadSSHKey(ctx, title, publicKey)
	default:
		return fmt.Errorf("provider %q does not support SSH key upload yet", def.ID)
	}
}

func deleteProviderSSHKey(ctx context.Context, def providerpkg.Definition, publicKey string) (bool, error) {
	switch def.ID {
	case providerpkg.GitHubID:
		return ctr.GitHubClient.DeleteSSHKey(ctx, publicKey)
	case providerpkg.GitLabID:
		return ctr.GitLabClient.DeleteSSHKey(ctx, publicKey)
	default:
		return false, fmt.Errorf("provider %q does not support SSH key deletion yet", def.ID)
	}
}

func providerGPGKeyExists(ctx context.Context, def providerpkg.Definition, keyID string) (bool, error) {
	switch def.ID {
	case providerpkg.GitHubID:
		return ctr.GitHubClient.GPGKeyExists(ctx, keyID)
	case providerpkg.GitLabID:
		return ctr.GitLabClient.GPGKeyExists(ctx, keyID)
	default:
		return false, fmt.Errorf("provider %q does not support GPG key upload yet", def.ID)
	}
}

func uploadProviderGPGKey(ctx context.Context, def providerpkg.Definition, armoredKey string) error {
	switch def.ID {
	case providerpkg.GitHubID:
		return ctr.GitHubClient.UploadGPGKey(ctx, armoredKey)
	case providerpkg.GitLabID:
		return ctr.GitLabClient.UploadGPGKey(ctx, armoredKey)
	default:
		return fmt.Errorf("provider %q does not support GPG key upload yet", def.ID)
	}
}

func deleteProviderGPGKey(ctx context.Context, def providerpkg.Definition, keyID string) (bool, error) {
	switch def.ID {
	case providerpkg.GitHubID:
		return ctr.GitHubClient.DeleteGPGKey(ctx, keyID)
	case providerpkg.GitLabID:
		return ctr.GitLabClient.DeleteGPGKey(ctx, keyID)
	default:
		return false, fmt.Errorf("provider %q does not support GPG key deletion yet", def.ID)
	}
}

func providerResourceName(profileName string, def providerpkg.Definition, parts ...string) string {
	components := []string{"gcm", safeProviderNameComponent(profileName), safeProviderNameComponent(string(def.ID))}
	for _, part := range parts {
		if cleaned := safeProviderNameComponent(part); cleaned != "" {
			components = append(components, cleaned)
		}
	}
	return strings.Join(components, "-")
}

func safeProviderNameComponent(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		allowed := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if allowed {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func sshKeyProfileName(profileName string, p *profile.Profile) string {
	id, ok := profileProviderID(p)
	if !ok {
		return profileName
	}
	suffix := safeProviderNameComponent(string(id))
	if suffix == "" {
		return profileName
	}
	return fmt.Sprintf("%s_%s", profileName, suffix)
}

func migrateProfileSSHKeyPathToProvider(profileName string, p *profile.Profile) (bool, error) {
	if p == nil || p.SSH == nil || p.SSH.KeyPath == "" {
		return false, nil
	}

	targetProfileName := sshKeyProfileName(profileName, p)
	if targetProfileName == profileName {
		return false, nil
	}

	keyType := string(p.SSH.KeyType)
	if keyType == "" {
		keyType = inferSSHKeyTypeFromPath(p.SSH.KeyPath)
	}
	if keyType == "" {
		return false, nil
	}

	currentPriv := p.SSH.KeyPath
	legacyName := fmt.Sprintf("id_%s_%s", keyType, profileName)
	if filepath.Base(currentPriv) != legacyName {
		return false, nil
	}

	targetPriv := filepath.Join(filepath.Dir(currentPriv), fmt.Sprintf("id_%s_%s", keyType, targetProfileName))
	if targetPriv == currentPriv {
		return false, nil
	}

	if _, err := os.Stat(targetPriv); err == nil {
		return false, fmt.Errorf("target SSH key already exists: %s", targetPriv)
	}

	currentPub := currentPriv + ".pub"
	targetPub := targetPriv + ".pub"

	if err := os.Rename(currentPriv, targetPriv); err != nil {
		return false, fmt.Errorf("renaming SSH key: %w", err)
	}

	pubExists := false
	if _, err := os.Stat(currentPub); err == nil {
		pubExists = true
		if err := os.Rename(currentPub, targetPub); err != nil {
			_ = os.Rename(targetPriv, currentPriv)
			return false, fmt.Errorf("renaming SSH public key: %w", err)
		}
	}

	originalPath := p.SSH.KeyPath
	p.SSH.KeyPath = targetPriv
	if err := ctr.ProfileManager.Update(p); err != nil {
		if pubExists {
			_ = os.Rename(targetPub, currentPub)
		}
		_ = os.Rename(targetPriv, currentPriv)
		p.SSH.KeyPath = originalPath
		return false, fmt.Errorf("updating profile after SSH key rename: %w", err)
	}

	return true, nil
}

func inferSSHKeyTypeFromPath(keyPath string) string {
	base := filepath.Base(strings.TrimSpace(keyPath))
	if !strings.HasPrefix(base, "id_") {
		return ""
	}
	rest := strings.TrimPrefix(base, "id_")
	idx := strings.Index(rest, "_")
	if idx <= 0 {
		return ""
	}
	return rest[:idx]
}

func providerManualKeyURL(def providerpkg.Definition, kind string) string {
	webURL := strings.TrimRight(def.WebURL, "/")
	if webURL == "" {
		webURL = strings.TrimRight(def.CredentialServer(), "/")
	}
	switch def.ID {
	case providerpkg.GitHubID:
		return webURL + "/settings/keys"
	case providerpkg.GitLabID:
		if kind == "gpg" {
			return webURL + "/-/user_settings/gpg_keys"
		}
		return webURL + "/-/user_settings/ssh_keys"
	default:
		return webURL
	}
}
