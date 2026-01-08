# anilist-mal-sync [![Build Status](https://github.com/bigspawn/anilist-mal-sync/workflows/go/badge.svg)](https://github.com/bigspawn/anilist-mal-sync/actions)

Program to synchronize your AniList and MyAnimeList accounts.

## Features

- Bidirectional sync between AniList and MyAnimeList (anime and manga)
- OAuth2 authentication
- CLI interface

## Quick Start (Docker)

### Step 1: Create OAuth applications

**AniList:**
1. Go to [AniList Developer Settings](https://anilist.co/settings/developer)
2. Create New Client with redirect URL: `http://localhost:18080/callback`
3. Save Client ID and Client Secret

**MyAnimeList:**
1. Go to [MAL API Settings](https://myanimelist.net/apiconfig)
2. Create Application with redirect URL: `http://localhost:18080/callback`
3. Save Client ID and Client Secret

### Step 2: Configure

Download [`docker-compose.example.yaml`](docker-compose.example.yaml) and edit credentials:

```yaml
version: "3"
services:
  sync:
    image: ghcr.io/bigspawn/anilist-mal-sync:latest
    ports:
      - "18080:18080"
    environment:
      - PUID=1000  # Your UID: id -u
      - PGID=1000  # Your GID: id -g
      - ANILIST_CLIENT_ID=your_anilist_client_id
      - ANILIST_CLIENT_SECRET=your_anilist_secret
      - ANILIST_USERNAME=your_anilist_username
      - MAL_CLIENT_ID=your_mal_client_id
      - MAL_CLIENT_SECRET=your_mal_secret
      - MAL_USERNAME=your_mal_username
      # - WATCH_INTERVAL=12h  # Optional: enable watch mode
    volumes:
      - tokens:/home/appuser/.config/anilist-mal-sync
    restart: unless-stopped

volumes:
  tokens:
```

### Step 3: Authenticate

```bash
docker-compose run --rm --service-ports sync login all
```

Follow the URLs printed in terminal.

### Step 4: Run

```bash
docker-compose up -d
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
| `-c` | `--config` | Path to config file (optional, uses env vars if not specified) |
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

### Environment variables

Configuration can be provided entirely via environment variables (recommended for Docker):

**Required:**
- `ANILIST_CLIENT_ID` - AniList Client ID
- `ANILIST_CLIENT_SECRET` - AniList Client Secret
- `ANILIST_USERNAME` - AniList username
- `MAL_CLIENT_ID` - MyAnimeList Client ID
- `MAL_CLIENT_SECRET` - MyAnimeList Client Secret
- `MAL_USERNAME` - MyAnimeList username

**Optional:**
- `WATCH_INTERVAL` - Sync interval for watch mode (e.g., `12h`, `24h`)
- `OAUTH_PORT` - OAuth server port (default: `18080`)
- `OAUTH_REDIRECT_URI` - OAuth redirect URI (default: `http://localhost:18080/callback`)
- `TOKEN_FILE_PATH` - Token file path (default: `~/.config/anilist-mal-sync/token.json`)
- `PUID` / `PGID` - User/Group ID for Docker volume permissions

## Advanced

### Install as binary

```bash
go install github.com/bigspawn/anilist-mal-sync@latest
anilist-mal-sync login all
anilist-mal-sync sync
```

### Docker

See [Quick Start](#quick-start-docker) for the recommended setup.

**Using config file instead of environment variables:**

```bash
docker run --rm -p 18080:18080 \
  -e PUID=$(id -u) -e PGID=$(id -g) \
  -v $(pwd)/config.yaml:/etc/anilist-mal-sync/config.yaml:ro \
  -v $(pwd)/tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest -c /etc/anilist-mal-sync/config.yaml sync
```

### Watch mode

Enable continuous sync by setting `WATCH_INTERVAL` environment variable:

```yaml
environment:
  - WATCH_INTERVAL=12h  # Sync every 12 hours
```

Or run watch command manually:
```bash
docker-compose run --rm sync watch --interval=12h
```

**Interval limits:** 1h - 168h (7 days)

### Scheduling (non-Docker)

Use your system's scheduler for periodic sync:

```bash
# Linux/macOS cron (daily at 2 AM)
0 2 * * * /usr/local/bin/anilist-mal-sync sync
```

## Troubleshooting

**"Required environment variables not set"**
- Set required env vars: `ANILIST_CLIENT_ID`, `ANILIST_CLIENT_SECRET`, `ANILIST_USERNAME`, `MAL_CLIENT_ID`, `MAL_CLIENT_SECRET`, `MAL_USERNAME`
- Or use config file with `-c /path/to/config.yaml`

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
