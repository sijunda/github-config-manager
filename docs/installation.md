# Installation

## Requirements

| Tool    | Required | Notes                                 |
| ------- | -------- | ------------------------------------- |
| **Go**  | Yes      | 1.26+ (only for building from source) |
| **Git** | Yes      | 2.20+ recommended                     |
| OpenSSH | Yes      | `ssh`, `ssh-add` on `PATH`            |
| GPG     | Optional | Only needed for commit signing        |

## From Source (recommended)

```bash
git clone https://github.com/justjundana/github-config-manager.git
cd github-config-manager
make build          # produces ./bin/gcm
make install        # installs to $(go env GOPATH)/bin/gcm (no sudo needed)
```

Make sure `$(go env GOPATH)/bin` is on your `PATH` (typically `~/go/bin`).

To install system-wide instead:

```bash
make install-system   # installs to /usr/local/bin/gcm (needs sudo)
```

## Via `go install`

```bash
go install github.com/justjundana/github-config-manager/cmd/gcm@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`.

## Prebuilt Binaries

> **Coming soon.** Homebrew tap and prebuilt release binaries are planned. For now, build from source.

When available, download the archive for your platform, extract, and move the binary onto your `PATH`:

```bash
tar -xzf gcm_<version>_<os>_<arch>.tar.gz
sudo mv gcm /usr/local/bin/
chmod +x /usr/local/bin/gcm
```

## Cross-Compilation

Build for all supported platforms at once:

```bash
make build-all
```

This produces binaries in `./bin/` for:
- `darwin/amd64`, `darwin/arm64` (macOS)
- `linux/amd64`, `linux/arm64`, `linux/arm`
- `windows/amd64`

## Verify Installation

```bash
gcm version     # show version, commit, build date
gcm doctor      # check Git, SSH, GPG, and shell integration
```

## Next Steps

After installing, continue with:

1. [Quick Start](quick-start.md) — up and running in 5 minutes
2. [Getting Started](getting-started.md) — create your first profile
3. [Shell Integration](shell-integration.md) — set up auto-switching on `cd`
4. [Commands Reference](commands.md) — explore every command and flag
