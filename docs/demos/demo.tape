Output docs/demo.gif

Require gcm

Set Width 1200
Set Height 800
Set Theme "TokyoNight"
Set WindowBar Colorful
Set Padding 30
Set FontSize 16
Set Margin 20
Set MarginFill "#1a1b26"
Set BorderRadius 10
Set TypingSpeed 50ms

# ─────────────────────────────────────────────────────────────────────────────
# Prepare a clean demo environment
# ─────────────────────────────────────────────────────────────────────────────

Hide
Type "export HOME=$(mktemp -d)"
Enter
Type "mkdir -p ~/projects/work-api && cd ~/projects/work-api && git init -q"
Enter
Type "clear"
Enter
Show

# ═══════════════════════════════════════════════════════════════════════════════
# Title
# ═══════════════════════════════════════════════════════════════════════════════

Type "# GCM — Git Config Manager"
Enter
Type "# Manage multiple Git identities, SSH keys, GPG keys & provider accounts"
Enter
Sleep 2s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 1. Version & Help
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 1. Version & Overview ───"
Enter
Sleep 1s

Type "gcm version"
Enter
Sleep 2s

Type "gcm --help"
Enter
Sleep 4s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 2. Profile Creation (non-interactive with flags)
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 2. Create Git Profiles ───"
Enter
Sleep 1s

Type@50ms "gcm profile create work --name 'Jundana' --email 'work@jundana.dev'"
Enter
Sleep 2.5s

Type@50ms "gcm profile create personal --name 'Jundana Al-Basyir' --email 'mail@jundana.dev'"
Enter
Sleep 2.5s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 3. Profile List & Show
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 3. List & Inspect Profiles ───"
Enter
Sleep 1s

Type "gcm profile list"
Enter
Sleep 3s

Type "gcm profile show work"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 4. Profile Activation
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 4. Switch Between Profiles ───"
Enter
Sleep 1s

Type "gcm use work --dry-run"
Enter
Sleep 3s

Type "gcm use work"
Enter
Sleep 2.5s

Type "gcm current"
Enter
Sleep 2.5s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 5. SSH Key Management
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 5. SSH Key Management ───"
Enter
Sleep 1s

Type "gcm ssh generate work --type ed25519"
Enter
Sleep 3s

Type "gcm ssh list"
Enter
Sleep 3s

Type "gcm ssh upload --help"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 6. GPG Signing
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 6. GPG Commit Signing ───"
Enter
Sleep 1s

Type "gcm gpg --help"
Enter
Sleep 3s

Type "gcm gpg upload --help"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 7. Provider Integrations
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 7. Provider Integrations ───"
Enter
Sleep 1s

Type "gcm github --help"
Enter
Sleep 3s

Type "gcm gitlab --help"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 8. Profile Validation
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 8. Validate Profiles ───"
Enter
Sleep 1s

Type "gcm validate"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 9. System Diagnostics
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 9. System Health Check ───"
Enter
Sleep 1s

Type "gcm doctor"
Enter
Sleep 4s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 10. Status Dashboard
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 10. Status Dashboard ───"
Enter
Sleep 1s

Type "gcm status"
Enter
Sleep 4s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 11. Templates
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 11. Configuration Templates ───"
Enter
Sleep 1s

Type "gcm template --help"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 12. Backup & Restore
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 12. Backup & Restore ───"
Enter
Sleep 1s

Type "gcm backup create"
Enter
Sleep 3s

Type "gcm backup list"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 13. Shell Integration
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 13. Shell Integration (auto-switch on cd) ───"
Enter
Sleep 1s

Type "gcm init --help"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# 14. Profile Export & Diff
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ─── 14. Profile Export & Diff ───"
Enter
Sleep 1s

Type "gcm profile export work"
Enter
Sleep 3s

Type "gcm profile diff work personal"
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms

# ═══════════════════════════════════════════════════════════════════════════════
# Outro
# ═══════════════════════════════════════════════════════════════════════════════

Type "# ═══════════════════════════════════════════════════════════════════"
Enter
Type "# GCM — Manage your Git identities with ease"
Enter
Type "#"
Enter
Type "#   Profile Management   •  SSH Key Generation  •  GPG Signing"
Enter
Type "#   GitHub & GitLab      •  Credential Isolation •  Shell Hooks"
Enter
Type "#   Templates            •  Backup & Restore    •  Diagnostics"
Enter
Type "#"
Enter
Type "#   Install:  go install github.com/justjundana/git-config-manager/cmd/gcm@latest"
Enter
Type "#   Start:    gcm setup"
Enter
Type "# ═══════════════════════════════════════════════════════════════════"
Enter
Sleep 5s
