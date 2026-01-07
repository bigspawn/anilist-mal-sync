# anilist-mal-sync [![Build Status](https://github.com/bigspawn/anilist-mal-sync/workflows/go/badge.svg)](https://github.com/bigspawn/anilist-mal-sync/actions)

Program to synchronize your AniList and MyAnimeList accounts.

## Features

- Bidirectional sync between AniList and MyAnimeList (anime and manga)
- OAuth2 authentication
- CLI interface

## Quick Start (5 minutes)

### Prerequisites
- Go 1.22+ OR Docker
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

### Step 2: Install and configure

```bash
# Download
git clone https://github.com/bigspawn/anilist-mal-sync.git
cd anilist-mal-sync

# Create config
cp config.example.yaml config.yaml

# Edit config.yaml
nano config.yaml
```

Edit `config.yaml` with your credentials:
```yaml
anilist:
  client_id: "your_anilist_client_id"
  client_secret: "your_anilist_client_secret"
  username: "your_anilist_username"
myanimelist:
  client_id: "your_mal_client_id"
  client_secret: "your_mal_client_secret"
  username: "your_mal_username"
token_file_path: ""  # Leave empty for default location
```

### Step 3: Authenticate

```bash
go run . login all
```

Follow the URLs printed in terminal to authorize both services.

### Step 4: Sync

```bash
go run . sync
```

Your lists are now synchronized!

## Commands

| Command | Description |
|---------|-------------|
| `login` | Authenticate with services |
| `logout` | Remove authentication tokens |
| `status` | Check authentication status |
| `sync` | Synchronize anime/manga lists |

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
```

### Environment variables (optional)

Override sensitive values:
- `PORT` - OAuth server port (default: 18080)
- `CLIENT_SECRET_ANILIST` - AniList client secret
- `CLIENT_SECRET_MYANIMELIST` - MyAnimeList client secret

## Advanced

### Install as binary

```bash
go install github.com/bigspawn/anilist-mal-sync@latest
anilist-mal-sync login all
anilist-mal-sync sync
```

### Docker

**Important:** For Docker, set `token_file_path: ""` in config.yaml to use the container's home directory.

**Pre-built image:**
```bash
docker pull ghcr.io/bigspawn/anilist-mal-sync:latest

# Create tokens directory
mkdir -p tokens

# Login
docker run --rm \
  -p 18080:18080 \
  -v $(pwd)/config.yaml:/etc/anilist-mal-sync/config.yaml:ro \
  -v $(pwd)/tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest login all

# Sync
docker run --rm \
  -p 18080:18080 \
  -v $(pwd)/config.yaml:/etc/anilist-mal-sync/config.yaml:ro \
  -v $(pwd)/tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest sync
```

**Docker Compose:**
```yaml
version: '3'
services:
  anilist-mal-sync:
    image: ghcr.io/bigspawn/anilist-mal-sync:latest
    ports:
      - "18080:18080"
    volumes:
      - ./config.yaml:/etc/anilist-mal-sync/config.yaml:ro
      - ./tokens:/home/appuser/.config/anilist-mal-sync
```

### Scheduling

This tool does not include built-in scheduling. Use your system's scheduler:

- **Linux/macOS**: cron or systemd timers
- **Windows**: Task Scheduler
- **Docker**: host cron or external orchestrator

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
