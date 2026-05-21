# Quick Start

Get started with GCM in under 5 minutes.

## 1. Install

```bash
# Clone and build
git clone https://github.com/sijunda/github-config-manager.git
cd github-config-manager
make build && make install

# Or via go install
go install github.com/sijunda/github-config-manager/cmd/gcm@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH` (typically `~/go/bin`).

## 2. Verify

```bash
gcm version
gcm doctor
```

## 3. Set Up Shell Integration

```bash
gcm init
```

This installs shell hooks (auto-switch, prompt indicator) and registers the built-in credential helper for `github.com`.

Restart your terminal:

```bash
# Bash
source ~/.bashrc

# Zsh
source ~/.zshrc

# Fish
source ~/.config/fish/config.fish

# PowerShell
. $PROFILE
```

## 4. Create Your First Profile

```bash
gcm profile create work --interactive
```

The wizard prompts for:
1. Name and email
2. SSH key generation
3. GPG signing setup
4. GitHub username

## 5. Activate It

```bash
gcm use work
```

## 6. Pin to a Project

```bash
cd ~/projects/work-repo
gcm use work --local
```

Now GCM auto-switches to `work` every time you `cd` into that directory.

## 7. Create Another Profile

```bash
gcm profile create personal -i
cd ~/projects/personal-repo
gcm use personal --local
```

## Verify Everything

```bash
gcm current         # show active profile
gcm profile list    # list all profiles
gcm validate        # validate all profiles
```

## Common Commands

```bash
gcm profile create <name> -i       # Create profile (interactive)
gcm profile list                    # List all profiles
gcm use <name>                      # Activate for session
gcm use <name> --global             # Set as default
gcm use <name> --local              # Pin to current directory
gcm ssh generate <name>             # Generate SSH key
gcm github login <name>             # Authenticate with GitHub
gcm backup create                   # Back up your config
gcm doctor                          # Health check
```

## Next Steps

- **[Getting Started](getting-started.md)** — Detailed first-time walkthrough
- **[Commands Reference](commands.md)** — Every command and flag
- **[Shell Integration](shell-integration.md)** — Master auto-switching
- **[Examples](examples.md)** — Real-world workflows
- **[Configuration](configuration.md)** — Customize GCM
