<h1 align="center">cb</h1>

<p align="center">Cross-Device Clipboard & Code Snippet Sync</p>

<p align="center">A lightweight CLI + Web tool for developers to sync short text, code snippets, and commands across devices — with end-to-end encryption, local-first storage, and real-time sync.</p>

<p>
  <strong>English | <a href="README_CN.md">中文</a></strong>
</p>

## Why cb?

**SSH'd into production, need a file from your laptop.**
No more opening Slack, finding the chat, copying, pasting into terminal. `cb get deploy-key` — done.

**Wrote a gnarly SQL query at the office, want it at home.**
No more emailing yourself, pasting into a note app, or AirDropping. `cb stash my-query "SELECT ..."` at work, `cb get my-query` at home.

**Colleague needs your nginx config. Now.** (TODO: direct sharing & team space in development)
No more "hold on let me find it" + attaching files in chat. `cb stash nginx-conf < nginx.conf` — they run `cb get nginx-conf`.

**Need to share an API key but it shouldn't live in Slack forever.**
`cb send --encrypt --ttl 30m "sk-xxxx"` — encrypted, auto-destroys in 30 minutes. The server never sees the plaintext.

**Managing 5 servers and every one has different deploy commands.**
`cb save deploy-prod --category ops "kubectl apply -f prod.yaml"` — organize, describe, tag. `cb list --category ops` to see them all.

## Quick Start

```bash
# Install
go install github.com/Morolis/cb@latest

# Login (registers if new)
cb login --user me --api-url http://your-server:8080/v1

# Send text to cloud (syncs across devices)
cb send "hello from my laptop"

# Quick save to cloud with a name
cb stash deploy "kubectl apply -f prod.yaml"

# Get it on another machine
cb get deploy

# Execute directly
cb exec deploy
```

## Installation

### Go install

```bash
go install github.com/Morolis/cb@latest
```

### Download binary

Download from [GitHub Releases](https://github.com/Morolis/cb/releases):
- Linux: `cb-linux-amd64.tar.gz`, `.deb`
- macOS: `cb-darwin-amd64.tar.gz`, `cb-darwin-arm64.tar.gz`
- Windows: `CBSetup-amd64.exe` (installer) or `cb-windows-amd64.zip`

### Homebrew (macOS)

```bash
brew tap Morolis/cb
brew install cb
```

### Docker (Server)

```bash
# Pre-built image (recommended, supports amd64 and arm64)
docker pull ghcr.io/morolis/cb:latest
docker run -d -p 8080:8080 -v cb-data:/data --name cb --restart always ghcr.io/morolis/cb:latest

# Or build from source
git clone https://github.com/Morolis/cb.git && cd cb
docker build -t cb-server .
docker run -d -p 8080:8080 -v cb-data:/data --name cb cb-server
```

## Commands

### `send` — Cloud clipboard

Send text to the cloud for cross-device sync.

```bash
cb send "kubectl get pods -A"
cb send --alias mylink "https://example.com"
cb send --ttl 1h --encrypt "secret data"
cb send --id 597ebc3e "updated content"       # Update existing snippet (creates version history)
cb send --alias myconfig "v2"                  # Upsert: update if alias exists, create if not
cat config.yaml | cb send --alias myconfig
```

| Flag | Description |
|------|-------------|
| `--alias` | Assign alias (upsert: updates if exists, creates if not) |
| `--id` | Update existing snippet by ID or prefix |
| `--ttl` | Time to live: `30s`, `5m`, `1h`, `1d` |
| `--encrypt` | AES-256-GCM encrypt |
| `--desc` | Description |
| `--var KEY=VALUE` | Variable substitution: replaces `{{.KEY}}` in content |

### `save` — Local-first storage

Save snippet locally. Works offline. Use `--remote` to also sync to cloud.

```bash
cb save mycmd "kubectl get pods -A"
cb save --desc "Get all pods" --category k8s --lang bash mycmd "kubectl get pods -A"
cb save --remote mycmd "ls -la"                # Also sync to cloud
cb save --ttl 7d --encrypt mydb "postgresql://user:pass@host/db"
```

| Flag | Description |
|------|-------------|
| `--remote` | Also push to the remote server |
| `--ttl` | Time to live: `30s`, `5m`, `1h`, `1d` |
| `--encrypt` | AES-256-GCM encrypt |
| `--desc` | Description |
| `--category` | Category for organization |
| `--lang` | Language hint (e.g. `python`, `bash`) |
| `--tags` | Comma-separated tags |

### `stash` — Quick cloud save with alias

Convenience shortcut: save to cloud with an alias name.

```bash
cb stash deploy "kubectl apply -f deploy.yaml"
cb stash --desc "DB dump" db-backup "pg_dump mydb > backup.sql"
```

| Flag | Description |
|------|-------------|
| `--ttl` | Time to live |
| `--encrypt` | AES-256-GCM encrypt |
| `--desc` | Description |

### `get` — Retrieve a snippet

Get a snippet by ID or alias. Checks local first, then remote. No argument returns the most recent.

```bash
cb get mycmd              # By alias
cb get 597ebc3e           # By ID (or prefix)
cb get                    # Most recent snippet
```

### `list` — List all snippets

Merged view of local + cloud snippets.

```bash
cb list
cb list --source local     # Local only
cb list --source remote    # Cloud only
cb list --limit 50
```

Output:
```
SOURCE  ALIAS  DESC  ID        PREVIEW                     CREATED           EXPIRES
------  -----  ----  --        -------                     -------           -------
local   mycmd  -     loc_c8fd  kubectl get pods -A         2024-01-01 10:30  -
remote  -      -     597ebc3e  hello world                 2024-01-01 10:25  -
```

### `exec` — Execute a snippet

Run a saved snippet as a shell command.

```bash
cb exec mycmd             # Execute by alias
cb exec 597ebc3e          # Execute by ID
```

### `rm` — Delete a snippet

Delete by ID or alias. Auto-detects local vs remote.

```bash
cb rm mycmd               # Auto-detect
cb rm loc_c8fdf403a07d    # Local ID
cb rm --source local mycmd
cb rm --source remote mycmd
```

### `history` — Version history

View all past versions of a snippet.

```bash
cb history mycmd
cb history 597ebc3e
```

### `rollback` — Restore a version

Rollback a snippet to a previous version. Current content is saved as a new version before rollback.

```bash
cb rollback mycmd 3           # Rollback to version ID 3
cb rollback 597ebc3e 7
```

### `webhook` — Manage webhooks

Receive HTTP POST notifications when snippets change.

```bash
# Add webhook (default JSON payload)
cb webhook add myhook https://example.com/hook created,updated,deleted

# Add webhook with custom payload template
cb webhook add slack https://hooks.slack.com/xxx created \
  --body '{"text":"[{{.Event}}] {{.Snippet.Content}}"}'

# List webhooks
cb webhook list

# View delivery logs
cb webhook logs <webhook-id>

# Delete
cb webhook rm <webhook-id>
```

Template variables: `{{.Event}}`, `{{.DateTime}}`, `{{.Snippet.ID}}`, `{{.Snippet.Alias}}`, `{{.Snippet.Content}}`, `{{.Snippet.Description}}`, `{{.Snippet.Category}}`, `{{.Snippet.Language}}`, `{{.Snippet.Encrypted}}`, `{{.Snippet.ExpiresAt}}`, `{{.Snippet.CreatedAt}}`, `{{.Snippet.UpdatedAt}}`.

Use `{{json .Snippet.Content}}` to safely escape content for JSON embedding.

### `login` / `logout` — Authentication

```bash
cb login --user myname --api-url http://server:8080/v1
cb logout
```

### `config` — View or modify configuration

```bash
cb config show
cb config set api_url http://your-server:8080/v1
```

Config file: `~/.cb/config.yaml`

### Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `~/.cb/config.yaml` | Config file path |
| `--api-url` | `http://localhost:8080/v1` | Override server API URL |
| `--verbose` | `false` | Debug output |
| `-v, --version` | — | Show version |

## Features

| Feature | Description |
|---------|-------------|
| **Three modes** | `send` (cloud), `save` (local-first), `stash` (quick cloud alias) |
| **Version history** | Every update auto-saves previous versions |
| **End-to-end encryption** | AES-256-GCM, server never sees plaintext |
| **Real-time sync** | WebSocket push across all connected devices |
| **Web UI** | Dashboard with syntax highlighting, editor, version history, webhook management |
| **Shell execution** | `cb exec mycmd` runs saved snippets directly |
| **Auto-expiry** | `--ttl 30m` / `1h` / `1d` — snippets self-destruct |
| **Webhooks** | Notify Slack, DingTalk, WeChat Work, Discord, or any HTTP endpoint |
| **Organization** | Categories, tags, descriptions, aliases |
| **Offline-first** | Local SQLite cache works without network |
| **Pipe support** | `cat file \| cb send`, `cb get deploy \| sh` |

## Server

```bash
# Start server
CB_JWT_SECRET="your-secret" cb-server

# With TLS
CB_TLS_CERT=cert.pem CB_TLS_KEY=key.pem cb-server

# Or auto-generate self-signed cert
CB_TLS_AUTO=true cb-server
```

Open `http://localhost:8080` for the Web UI.

## Documentation

- [API Reference](docs/api.md)
- [Encryption Design](docs/encryption.md)
- [Deployment Guide](docs/deploy.md)

## Contributing

```bash
git clone https://github.com/Morolis/cb.git
cd cb

# CLI
go build -o cb .

# Server (embeds frontend)
cd web && npm install && npm run build && cd ..
go build -o cb-server ./server/main.go

# Frontend dev
cd web && npm run dev
```

## License

[Apache License 2.0](LICENSE)
