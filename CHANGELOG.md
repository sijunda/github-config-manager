# Changelog

All notable changes to GCM will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- **Built-in credential helper** — GCM registers itself as git's credential helper for github.com (`gcm credential-helper`). Git push/pull/clone reads tokens directly from GCM's encrypted store, bypassing the system keychain entirely. External credential store changes (VS Code logout, browser session clear, etc.) no longer break git authentication
- **Git credential isolation** — `gcm use` now clears previous credentials and stores the new profile's token via `git credential approve/reject`, preventing credential bleed between profiles
- **Credential username pinning** — sets `credential.https://github.com.username` in global git config so git only uses credentials matching the active profile
- **Smart scope fallback** — `gcm use <name>` works anywhere: inside a git repo → session scope, outside → local scope (writes `.gcm-profile`). No more "not in a git repository" errors
- **`--global` clears local overrides** — `gcm use <name> --global` now removes any `.gcm-profile` file and session marker in the current directory
- **`--hide-default` flag on `gcm current`** — outputs nothing when the active profile is the default; ideal for shell prompts that only show an indicator when you've explicitly switched
- **`--clear-credentials` flag on `gcm github logout`** — clears git credentials via `git credential reject` (default: true)
- **Login credential isolation** — `gcm github login*` commands only store git credentials if the profile being logged in is currently active; prevents credential bleed from non-active logins
- **Shell prompt improvements** — uses `precmd`/`PROMPT_COMMAND` hook with a `$_GCM_PROMPT` variable approach (idempotent, no subshell on every keystroke, hides when default is active)
- Profile management (create, list, show, edit, delete, export, import, diff)
- Profile activation with session, global, and local scopes
- Dry-run mode for profile activation
- Session marker file (`.git/gcm-session`) for reliable session detection independent of git config
- Session-aware profile detection: `gcm current` checks session marker → local marker → email matching → global default
- SSH key generation (ed25519, RSA, ECDSA)
- SSH key listing, connection testing, and clipboard copy
- GPG key generation and commit signing management
- GitHub OAuth device flow authentication (`gcm github login-oauth`)
- GitHub Personal Access Token (PAT) authentication (`gcm github login`)
- GitHub CLI token import (`gcm github login-gh`)
- GitHub authentication status overview (`gcm github status`)
- Encrypted token storage (AES-256-GCM)
- Shell integration for bash, zsh, fish, and powershell
- Auto-profile switching on directory change via `.gcm-profile`
- Shell prompt indicator for active profile
- Configuration template management (create, list, show, delete, import/export, apply)
- Backup and restore with tar.gz archives
- Backup pruning with retention policy
- Profile validation (basic and deep filesystem checks)
- System health check (`gcm doctor`)
- Cache cleaning utility
- Audit logging (JSONL format)
- Responsive table output (auto-adapts to terminal width: truncate → hide columns → vertical cards)
- Cross-platform support (macOS, Linux, Windows)
- Comprehensive CLI with colors, spinners, and interactive prompts
- GoReleaser configuration for automated releases
- Makefile with build, test, lint, and release targets
- Unit tests for core packages (crypto, file service, logger, profile, version)

### Fixed
- **`gcm use` not switching profiles** — session scope (`.git/gcm-session`) now takes priority over local (`.gcm-profile`), so `gcm use <name>` always reflects immediately in `gcm current`
- **Logout credential bleed** — `gcm github logout personal` while on `work` profile no longer clears work's git credentials
- **Profile delete without safeguard** — deleting the active profile now warns and requires extra confirmation
- **Raw error messages everywhere** — all commands that take a profile name now show a clear `✗ profile "x" not found` with actionable suggestions instead of internal filesystem errors
- **`gcm ssh test/copy`** — shows `✗ no SSH key configured` with suggestion to generate one, instead of raw error
- **`gcm backup restore`** — validates file exists before prompting for confirmation
- **`gcm profile import`** — shows clear message when file not found
- **`gcm validate`** — shows clean "not found" instead of raw YAML path error
- **Missing argument messages** — all commands now show usage and help command instead of `"accepts N arg(s), received 0"`

### Changed
- **Profile detection priority** — `gcm current` now checks: session → local → global (was: local → session → global). Session represents the user's most recent explicit `gcm use` action
- **Log verbosity** — all internal operation logs moved from INFO to DEBUG; only shown with `-v` flag. No more `[INFO]` noise in normal usage
- **GitHub OAuth error messages** — connection failures, rejected requests, and missing device codes now show user-friendly messages instead of raw HTTP/JSON errors
