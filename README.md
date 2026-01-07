# anilist-mal-sync [![Build Status](https://github.com/bigspawn/anilist-mal-sync/workflows/go/badge.svg)](https://github.com/bigspawn/anilist-mal-sync/actions)

Program to synchronize your AniList and MyAnimeList accounts.

## Features

- Bidirectional sync between AniList and MyAnimeList (anime and manga)
- OAuth2 authentication
- CLI interface

## Quick Start (Docker - 5 minutes)

### Prerequisites
- Docker
- Accounts on AniList and MyAnimeList

### Step 1: Create OAuth applications

**AniList:**
1. Go to [AniList Developer Settings](https://anilist.co/settings/developer)
2. Click "Create New Client"
3. Set redirect URL: `http://localhost:18080/callback`
4. Save the Client ID and Client Secret

**MyAnimeList:**
1. Go to [MAL API Settings](https://myanimelist.net/apiconfig)
2. Click "Create Application"
3. Set redirect URL: `http://localhost:18080/callback`
4. Save the Client ID and Client Secret

### Step 2: Configure

Create `config.yaml`:
```yaml
anilist:
  client_id: "your_anilist_client_id"
  client_secret: "your_anilist_client_secret"
  username: "your_anilist_username"
myanimelist:
  client_id: "your_mal_client_id"
  client_secret: "your_mal_client_secret"
  username: "your_mal_username"
token_file_path: ""  # Empty = auto-detect (recommended)
```

### Step 3: Authenticate

```bash
docker run --rm -p 18080:18080 \
  -v $(pwd)/config.yaml:/etc/anilist-mal-sync/config.yaml:ro \
  -v $(pwd)/tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest login all
```

Follow the URLs printed in terminal.

### Step 4: Sync

```bash
docker run --rm -p 18080:18080 \
  -v $(pwd)/config.yaml:/etc/anilist-mal-sync/config.yaml:ro \
  -v $(pwd)/tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest sync
```

Done!

## Commands

| Command | Description |
|---------|-------------|
| `login` | Authenticate with services |
| `logout` | Remove authentication tokens |
| `status` | Check authentication status |
| `sync` | Synchronize anime/manga lists |
| `watch` | Run sync on interval |

**Login/Logout options:**
| Short | Long | Description |
|-------|------|-------------|
| `-s` | `--service` | Service: `anilist`, `myanimelist`, `all` (default) |

**Sync options:**
| Short | Long | Description |
|-------|------|-------------|
| `-c` | `--config` | Path to config file (default: `config.yaml`) |
| `-f` | `--force` | Force sync all entries |
| `-d` | `--dry-run` | Dry run without making changes |
| | `--manga` | Sync manga instead of anime |
| | `--all` | Sync both anime and manga |
| | `--verbose` | Enable verbose logging |
| | `--reverse-direction` | Sync from MyAnimeList to AniList |

**Watch options:**
| Short | Long | Description |
|-------|------|-------------|
| `-i` | `--interval` | Sync interval: 1h-168h (overrides config) |
| | `--once` | Sync immediately then start watching |

Interval can be set via `--interval` flag or in `config.yaml` under `watch.interval`.

For backward compatibility, running `anilist-mal-sync [options]` without a command will execute sync.

## Configuration

### Config file

Full `config.yaml` example:

```yaml
oauth:
  port: "18080"
  redirect_uri: "http://localhost:18080/callback"
anilist:
  client_id: "your_client_id"
  client_secret: "your_secret"
  auth_url: "https://anilist.co/api/v2/oauth/authorize"
  token_url: "https://anilist.co/api/v2/oauth/token"
  username: "your_username"
myanimelist:
  client_id: "your_client_id"
  client_secret: "your_secret"
  auth_url: "https://myanimelist.net/v1/oauth2/authorize"
  token_url: "https://myanimelist.net/v1/oauth2/token"
  username: "your_username"
token_file_path: ""  # Leave empty for default: ~/.config/anilist-mal-sync/token.json
watch:
  interval: "24h"  # Sync interval for watch mode (1h-168h), can be overridden with --interval flag
```

### Environment variables (optional)

Override sensitive values:
- `PORT` - OAuth server port (default: 18080)
- `CLIENT_SECRET_ANILIST` - AniList client secret
- `CLIENT_SECRET_MYANIMELIST` - MyAnimeList client secret
- `TOKEN_FILE_PATH` - Token file path (default: `~/.config/anilist-mal-sync/token.json`)

## Advanced

### Install as binary

```bash
go install github.com/bigspawn/anilist-mal-sync@latest
anilist-mal-sync login all
anilist-mal-sync sync
```

### Docker

**Important:** For Docker, set `token_file_path: ""` in config.yaml to use the container's home directory.

#### Volume Permissions

To avoid "permission denied" errors, use one of these approaches:

**Option 1: Named volume (RECOMMENDED)**
```yaml
version: '3'
services:
  sync:
    image: ghcr.io/bigspawn/anilist-mal-sync:latest
    volumes:
      - ./config.yaml:/etc/anilist-mal-sync/config.yaml:ro
      - tokens:/home/appuser/.config/anilist-mal-sync
volumes:
  tokens:
```

**Option 2: Run as host user (for bind mounts)**
```yaml
version: '3'
services:
  sync:
    image: ghcr.io/bigspawn/anilist-mal-sync:latest
    user: "${UID:-1000}:${GID:-1000}"
    volumes:
      - ./config.yaml:/etc/anilist-mal-sync/config.yaml:ro
      - ./tokens:/home/appuser/.config/anilist-mal-sync
```

Then run: `UID=$(id -u) GID=$(id -g) docker-compose up`

#### Pre-built image

```bash
docker pull ghcr.io/bigspawn/anilist-mal-sync:latest

# Login (with named volume)
docker volume create anilist-tokens
docker run --rm \
  -p 18080:18080 \
  -v $(pwd)/config.yaml:/etc/anilist-mal-sync/config.yaml:ro \
  -v anilist-tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest login all

# Or with host user (bind mount)
docker run --rm \
  -p 18080:18080 \
  --user "$(id -u):$(id -g)" \
  -v $(pwd)/config.yaml:/etc/anilist-mal-sync/config.yaml:ro \
  -v $(pwd)/tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest login all
```

### Scheduling

**Built-in watch mode (Docker-friendly):**

Run continuous sync with Docker Compose. Set interval in `config.yaml`:
```yaml
version: '3'
services:
  sync:
    image: ghcr.io/bigspawn/anilist-mal-sync:latest
    command: ["watch"]
    volumes:
      - ./config.yaml:/etc/anilist-mal-sync/config.yaml:ro
      - tokens:/home/appuser/.config/anilist-mal-sync
    restart: unless-stopped
volumes:
  tokens:
```

Or override with CLI flag:
```yaml
command: ["watch", "--interval=12h"]
```

**Interval limits:**
- Minimum: 1 hour (to avoid API rate limits)
- Maximum: 7 days
- Format: `12h`, `24h`, `48h` (hours only)
- Priority: CLI flag > Config file
- Interval must be specified via one of these methods

**Alternative: External schedulers**

For non-Docker setups, use your system's scheduler:
- **Linux/macOS**: cron or systemd timers
- **Windows**: Task Scheduler

Example cron entry (daily at 2 AM):
```bash
0 2 * * * /usr/local/bin/anilist-mal-sync sync
```

## Troubleshooting

**"Config file not found"**
- Ensure `config.yaml` exists in current directory or use `-c /path/to/config.yaml`

**Authentication fails**
- Check redirect URL matches exactly: `http://localhost:18080/callback`
- Verify client ID and secret are correct
- Ensure port 18080 is not already in use

**Sync appears frozen**
- Both services have rate limits. Wait a few minutes and try again
- Use `--verbose` to see progress

**Token expired**
- Run `anilist-mal-sync status` to check
- Run `anilist-mal-sync login all` to reauthenticate

## Disclaimer

This project is not affiliated with AniList or MyAnimeList. Use at your own risk.

## TODO

- [ ] Sync favorites
- [x] Sync MAL to AniList
- [ ] Sync rewatching and rereading
