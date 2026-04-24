# anilist-mal-sync [![Build Status](https://github.com/bigspawn/anilist-mal-sync/workflows/go/badge.svg)](https://github.com/bigspawn/anilist-mal-sync/actions) [![codecov](https://codecov.io/gh/bigspawn/anilist-mal-sync/branch/main/graph/badge.svg)](https://codecov.io/gh/bigspawn/anilist-mal-sync)

> **Note:** This project is under development. Feedback, suggestions, and issues are highly appreciated!

Program to synchronize your AniList and MyAnimeList accounts.

## Features

- Bidirectional sync between AniList and MyAnimeList (anime and manga)
- Favorites synchronization (MAL → AniList with add-only, AniList → MAL report-only)
- OAuth2 authentication
- CLI interface
- Manual ID mappings and ignore rules via `mappings.yaml`
- Duplicate target detection with automatic conflict resolution
- Unmapped entries tracking with interactive management (`unmapped` command)
- Offline ID mapping using anime-offline-database (prevents incorrect season matches)
- Optional ARM API integration for online ID lookups

## What gets synced

For each entry in your list the following fields are synchronized from source to target:

| Field | Synced |
|-------|--------|
| Status (watching / completed / on-hold / dropped / plan to watch) | ✅ |
| Score (automatically normalized between AniList and MAL score formats) | ✅ |
| Progress (episodes watched / chapters + volumes read) | ✅ |
| Start date | ✅ (nil source date never overwrites a set target date) |
| Finish date | ✅ (only when status is Completed) |
| Favorites | ✅ optional, via `--favorites` flag |

**Conflict rule:** the source service always wins. In the default direction (AniList → MAL) your AniList data overwrites MAL. Use `--reverse-direction` to flip.

See [docs/date-sync.md](docs/date-sync.md) for detailed date synchronization rules.

## Prerequisites

| Deployment | Requirements |
|---|---|
| **Docker** | [Docker](https://docs.docker.com/get-docker/) + [docker compose](https://docs.docker.com/compose/install/) (v2, built-in) or legacy `docker-compose` (v1) |
| **Binary (go install)** | [Go 1.25+](https://go.dev/dl/) |
| **Local build** | [Go 1.25+](https://go.dev/dl/), git |

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

Download and rename the example compose file, then fill in your credentials:

```bash
cp docker-compose.example.yaml docker-compose.yaml
```

Edit `docker-compose.yaml` with your credentials:

```yaml
services:
  sync:
    image: ghcr.io/bigspawn/anilist-mal-sync:latest
    command: ["watch", "--once"]
    ports:
      - "18080:18080"
    environment:
      # User/Group ID for volume permissions (run: id -u / id -g)
      - PUID=1000
      - PGID=1000
      # Required: API credentials
      - ANILIST_CLIENT_ID=your_anilist_client_id
      - ANILIST_CLIENT_SECRET=your_anilist_secret
      - ANILIST_USERNAME=your_anilist_username
      - MAL_CLIENT_ID=your_mal_client_id
      - MAL_CLIENT_SECRET=your_mal_secret
      - MAL_USERNAME=your_mal_username
      # Watch mode: use either interval OR schedule (not both)
      - WATCH_INTERVAL=12h
      # - WATCH_SCHEDULE=0 3 * * *  # Cron expression (e.g. daily at 03:00). Mutually exclusive with WATCH_INTERVAL.
      # Optional: Manual mappings file path
      # - MAPPINGS_FILE_PATH=/home/appuser/.config/anilist-mal-sync/mappings.yaml
      # Optional: ID Mapping settings
      # - OFFLINE_DATABASE_ENABLED=true  # Enable offline DB (default: true)
      # - HATO_API_ENABLED=true  # Enable Hato API (default: true)
      # - HATO_API_URL=https://hato.malupdaterosx.moe  # Hato API base URL
      # - HATO_API_CACHE_DIR=/home/appuser/.config/anilist-mal-sync/hato-cache  # Cache directory
      # - HATO_API_CACHE_MAX_AGE=720h  # Cache max age (default: 720h / 30 days)
      # - ARM_API_ENABLED=false  # Enable ARM API (default: false)
      # - ARM_API_URL=https://arm.haglund.dev  # ARM API base URL
      # - JIKAN_API_ENABLED=false  # Enable Jikan API for manga ID mapping
      # - JIKAN_API_CACHE_DIR=/home/appuser/.config/anilist-mal-sync/jikan-cache
      # - JIKAN_API_CACHE_MAX_AGE=168h
      # - FAVORITES_SYNC_ENABLED=false  # Enable favorites sync (requires Jikan API)
    volumes:
      - tokens:/home/appuser/.config/anilist-mal-sync
    restart: unless-stopped

volumes:
  tokens:
```

### Step 3: Authenticate

```bash
docker-compose run --rm --service-ports sync login
```

The tool will print two URLs — one for MyAnimeList and one for AniList. For each:
1. Copy the URL and open it in your browser
2. Authorize the application on the website
3. Your browser will redirect to `http://localhost:18080/callback` — the tool captures this automatically
4. Repeat for both services (MAL first, then AniList)

> **Note:** The `--service-ports` flag is required here so that the OAuth redirect to port 18080 reaches the container. Make sure port 18080 is free on your host.

Tokens are saved into the `tokens` Docker volume and persist across restarts.

### Step 4: Run

#### Step 4.1: Preview changes (dry run)

**Recommended before the first real sync.** On a large list the tool may update hundreds of
entries at once — dry run lets you see exactly what will change without touching anything.

```bash
docker-compose run --rm sync sync --dry-run --all
```

#### Step 4.2: Start the sync daemon

`docker-compose up -d` starts the container in **watch mode** (`--once` flag causes an
immediate first sync, then it repeats every `WATCH_INTERVAL` hours in the background).

```bash
docker-compose up -d
```

Done! The service will sync your lists every 12 hours (or whatever you set in `WATCH_INTERVAL`).

To check that the service started correctly and view sync output:

```bash
docker-compose logs -f sync
```

> **Note:** `WATCH_INTERVAL` accepts values between `1h` and `168h` (7 days).
> To run a one-off sync instead of the daemon use:
> ```bash
> docker-compose run --rm sync sync
> ```

## Commands

| Command | Description |
|---------|-------------|
| `login` | Authenticate with services |
| `logout` | Remove authentication tokens |
| `status` | Check authentication status |
| `sync` | Synchronize anime/manga lists |
| `watch` | Run sync on interval or cron schedule |
| `unmapped` | Show and manage unmapped entries from last sync |

**Global options** (available for all commands):
| Short | Long | Description |
|-------|------|-------------|
| `-c` | `--config` | Path to config file (optional, uses env vars if not specified) |

**Login/Logout options:**
| Short | Long | Description |
|-------|------|-------------|
| `-s` | `--service` | Service: `anilist`, `myanimelist`, `all` (default: `all`) |

**Sync options:**
| Short | Long | Description |
|-------|------|-------------|
| `-f` | `--force` | Force sync all entries |
| `-d` | `--dry-run` | Dry run without making changes |
| | `--manga` | Sync manga instead of anime |
| | `--all` | Sync both anime and manga |
| | `--verbose` | Enable verbose logging |
| | `--reverse-direction` | Sync from MyAnimeList to AniList |
| | `--offline-db` | Enable offline database for anime ID mapping (default: `true`, ignored for `--manga`) |
| | `--offline-db-force-refresh` | Force re-download offline database |
| | `--arm-api` | Enable ARM API for anime ID mapping (default: `false`, ignored for `--manga`) |
| | `--arm-api-url` | ARM API base URL |
| | `--jikan-api` | Enable Jikan API for manga ID mapping (default: `false`, ignored for anime) |
| | `--favorites` | Sync favorites between services (requires Jikan API for MAL favorites) |

**Watch options:**
| Short | Long | Description |
|-------|------|-------------|
| `-i` | `--interval` | Sync interval: 1h-168h (overrides config). Mutually exclusive with `--schedule` |
| `-s` | `--schedule` | Cron expression (e.g. `0 3 * * *` for daily at 03:00). Mutually exclusive with `--interval` |
| | `--once` | Run one sync immediately, then start the interval/cron loop. **Without this flag the first sync is delayed by the full interval or until the next cron tick.** |

Watch mode requires either `--interval` or `--schedule` (or the corresponding config/env var). Setting both is an error.

**Priority:** CLI flag > environment variable > `config.yaml`.

| Source | Interval | Schedule |
|--------|----------|----------|
| CLI flag | `--interval 12h` | `--schedule "0 3 * * *"` |
| Environment | `WATCH_INTERVAL=12h` | `WATCH_SCHEDULE=0 3 * * *` |
| Config YAML | `watch.interval: "12h"` | `watch.schedule: "0 3 * * *"` |

**Unmapped options:**
| Short | Long | Description |
|-------|------|-------------|
| | `--fix` | Interactively fix unmapped entries (ignore or map to MAL ID) |
| | `--ignore-all` | Add all unmapped entries to ignore list |

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
mappings_file_path: ""  # Leave empty for default: ~/.config/anilist-mal-sync/mappings.yaml
watch:
  interval: "24h"  # Sync interval for watch mode (1h-168h), can be overridden with --interval flag
  # schedule: "0 3 * * *"  # Cron expression (mutually exclusive with interval)
http_timeout: "30s"  # HTTP client timeout for API requests (default: 30s)
offline_database:
  enabled: true
  cache_dir: ""  # Default: ~/.config/anilist-mal-sync/aod-cache
  auto_update: true
arm_api:
  enabled: false
  base_url: "https://arm.haglund.dev" # Default: https://arm.haglund.dev
hato_api:
  enabled: true  # Enable Hato API for ID mapping (default: true)
  base_url: "https://hato.malupdaterosx.moe"  # Hato API base URL
  cache_dir: ""  # Leave empty for default: ~/.config/anilist-mal-sync/hato-cache
  cache_max_age: "720h"  # Cache max age (default: 720h / 30 days)
jikan_api:
  enabled: false  # Enable Jikan API for manga ID mapping (default: false)
  cache_dir: ""  # Default: ~/.config/anilist-mal-sync/jikan-cache
  cache_max_age: "168h"  # Cache max age (default: 168h / 7 days)
favorites:
  enabled: false  # Enable favorites synchronization (default: false)
```

## ID Mapping Strategies

The tool uses different ID mapping strategies for anime and manga, and the chain differs by direction.

### Forward direction (AniList → MAL, default)

**Anime** (`sync` or `sync --all`):
1. **Manual Mapping** - User-defined AniList↔MAL mappings from `mappings.yaml`
2. **Direct ID lookup** - If the entry already exists in your target list
3. **Offline Database** (optional, enabled by default) - Local database from [anime-offline-database](https://github.com/manami-project/anime-offline-database)
4. **Hato API** (optional, enabled by default) - Online API for anime/manga ID mapping
5. **ARM API** (optional, disabled by default) - Online fallback to [arm-server](https://arm.haglund.dev)
6. **Title matching** - Match by title similarity
7. **API search** - Search the MAL API

**Manga** (`sync --manga` or `sync --all`):
1. **Manual Mapping** - User-defined AniList↔MAL mappings from `mappings.yaml`
2. **Direct ID lookup** - If the entry already exists in your target list
3. **Hato API** (optional, enabled by default) - Online API for manga ID mapping
4. **Title matching** - Match by title similarity
5. **Jikan API** (optional, disabled by default) - Online API for manga ID mapping via [Jikan](https://jikan.moe/) (unofficial MAL API)
6. **API search** - Search the MAL API

### Reverse direction (MAL → AniList, `--reverse-direction`)

**Anime** (`sync --reverse-direction`):
1. **Manual Mapping** - User-defined AniList↔MAL mappings from `mappings.yaml`
2. **Direct ID lookup** - If the entry already exists in your target list
3. **Offline Database** (optional, enabled by default)
4. **Hato API** (optional, enabled by default)
5. **ARM API** (optional, disabled by default)
6. **Title matching**
7. **MAL ID lookup** - Find AniList entry by MAL ID directly
8. **API search** - Search the AniList API

**Manga** (`sync --manga --reverse-direction`):
1. **Manual Mapping**
2. **Direct ID lookup**
3. **Hato API** (optional, enabled by default)
4. **Title matching**
5. **Jikan API** (optional, disabled by default)
6. **MAL ID lookup** - Find AniList entry by MAL ID directly
7. **API search** - Search the AniList API

**Notes:**
- The offline database and ARM API are anime-only and automatically disabled when using `--manga` flag (without `--all`) to improve startup performance.
- Hato API supports both anime and manga and is enabled by default.

### Manual Mappings & Ignore Rules

You can define manual ID mappings and ignore rules in a YAML file (`mappings.yaml`):

```yaml
manual_mappings:
  - anilist_id: 12345
    mal_id: 67890
    comment: "Season 2 mapped manually"
ignore:
  anilist_ids:
    - 99999 # Title Name : reason for ignoring
  titles:
    - "Some Title to Ignore"
```

Default location: `~/.config/anilist-mal-sync/mappings.yaml`

You can also manage ignore rules interactively:
```bash
# Show unmapped entries from last sync
anilist-mal-sync unmapped

# Interactively fix unmapped entries (ignore or map to MAL ID)
anilist-mal-sync unmapped --fix

# Add all unmapped entries to ignore list
anilist-mal-sync unmapped --ignore-all
```

### Environment variables

Configuration can be provided entirely via environment variables (recommended for Docker):

**Required:**
- `ANILIST_CLIENT_ID` - AniList Client ID
- `ANILIST_CLIENT_SECRET` - AniList Client Secret (also accepts `CLIENT_SECRET_ANILIST`)
- `ANILIST_USERNAME` - AniList username
- `MAL_CLIENT_ID` - MyAnimeList Client ID
- `MAL_CLIENT_SECRET` - MyAnimeList Client Secret (also accepts `CLIENT_SECRET_MYANIMELIST`)
- `MAL_USERNAME` - MyAnimeList username

**Required for `watch` mode:**
- `WATCH_INTERVAL` - Sync interval (e.g., `12h`, `24h`); range `1h`–`168h`. Mutually exclusive with `WATCH_SCHEDULE`.
- `WATCH_SCHEDULE` - Cron expression for scheduled sync (e.g., `0 3 * * *` for daily at 03:00). Mutually exclusive with `WATCH_INTERVAL`.

One of `WATCH_INTERVAL` or `WATCH_SCHEDULE` (or their CLI flag equivalents) is required for watch mode. Setting both is an error.

**Optional:**
- `HTTP_TIMEOUT` - HTTP client timeout for API requests (default: `30s`, e.g., `10s`, `1m`)
- `OAUTH_PORT` - OAuth server port (default: `18080`)
- `OAUTH_REDIRECT_URI` - OAuth redirect URI (default: `http://localhost:18080/callback`)
- `TOKEN_FILE_PATH` - Token file path (default: `~/.config/anilist-mal-sync/token.json`)
- `MAPPINGS_FILE_PATH` - Path to manual mappings YAML file (default: `~/.config/anilist-mal-sync/mappings.yaml`)
- `PUID` / `PGID` - User/Group ID for Docker volume permissions
- `OFFLINE_DATABASE_ENABLED` - Enable offline database for anime ID mapping (default: `true`, not used for manga-only sync)
- `OFFLINE_DATABASE_CACHE_DIR` - Cache directory (default: `~/.config/anilist-mal-sync/aod-cache`)
- `OFFLINE_DATABASE_AUTO_UPDATE` - Auto-update database (default: `true`)
- `HATO_API_ENABLED` - Enable Hato API for ID mapping (default: `true`, supports both anime and manga)
- `HATO_API_URL` - Hato API base URL (default: `https://hato.malupdaterosx.moe`)
- `HATO_API_CACHE_DIR` - Hato API cache directory (default: `~/.config/anilist-mal-sync/hato-cache`)
- `HATO_API_CACHE_MAX_AGE` - Hato API cache max age (default: `720h` / 30 days)
- `ARM_API_ENABLED` - Enable ARM API for anime ID mapping (default: `false`, not used for manga-only sync)
- `ARM_API_URL` - ARM API base URL (default: `https://arm.haglund.dev`)
- `JIKAN_API_ENABLED` - Enable Jikan API for manga ID mapping (default: `false`, not used for anime sync)
- `JIKAN_API_CACHE_DIR` - Jikan API cache directory (default: `~/.config/anilist-mal-sync/jikan-cache`)
- `JIKAN_API_CACHE_MAX_AGE` - Jikan API cache max age (default: `168h` / 7 days)
- `FAVORITES_SYNC_ENABLED` - Enable favorites synchronization (default: `false`)

## Favorites Synchronization

Favorites sync is an optional feature that synchronizes your favorited anime and manga between AniList and MyAnimeList. It runs as a separate phase after the main status/progress synchronization.

### API Limitations

| Direction | Read | Write | Behavior |
|-----------|------|-------|----------|
| MAL → AniList | ✅ via Jikan API | ✅ via ToggleFavourite mutation | Full sync (add missing favorites) |
| AniList → MAL | ✅ via isFavourite field | ❌ MAL API v2 has no favorites endpoint | Report only |

### Enabling Favorites Sync

**Via CLI flag:**
```bash
# Sync with favorites enabled
anilist-mal-sync sync --favorites

# Reverse sync (MAL → AniList) with favorites
anilist-mal-sync sync --favorites --reverse-direction
```

**Via environment variable:**
```bash
export FAVORITES_SYNC_ENABLED=true
anilist-mal-sync sync
```

**Via config file:**
```yaml
favorites:
  enabled: true
```

**Via Docker:**
```yaml
environment:
  - FAVORITES_SYNC_ENABLED=true
```

Note: The `--favorites` flag automatically enables Jikan API (required for reading MAL favorites).

### Behavior

#### MAL → AniList (with `--reverse-direction`)
- Reads your MAL favorites via Jikan API (public user profile)
- Compares with your AniList list entries
- **Adds** missing favorites on AniList
- **Does not remove** favorites that exist only on AniList (you may have intentionally favorited different items)

Example output:
```
★ [Favorites] Added "Cowboy Bebop" to AniList favorites
★ [Favorites] Added "Monster" to AniList favorites
★ Favorites sync complete: +2 added on AniList (15 skipped)
```

#### AniList → MAL (default direction)
- Reads your AniList favorites from list entries (via `isFavourite` field)
- Reads your MAL favorites via Jikan API
- Reports differences (cannot write to MAL)

Example output:
```
★ [Favorites] anime "Cowboy Bebop" is only on AniList
★ [Favorites] manga "Berserk" is only on MAL
★ Favorites: 2 mismatches (AniList→MAL, report only)
```

For detailed documentation, see [docs/favorites-sync.md](docs/favorites-sync.md).

## Advanced

### Install as binary (without Docker)

Requires **Go 1.25+** ([download](https://go.dev/dl/)).

**Option A — install from registry:**
```bash
go install github.com/bigspawn/anilist-mal-sync@latest
```

**Option B — clone and build locally:**
```bash
git clone https://github.com/bigspawn/anilist-mal-sync.git
cd anilist-mal-sync
go build -o anilist-mal-sync .
```

**First run:**
```bash
# 1. Create config file
cp config.example.yaml config.yaml
# Edit config.yaml with your AniList and MAL credentials

# 2. Authenticate (opens OAuth flow on port 18080)
anilist-mal-sync -c config.yaml login

# 3. Preview changes before syncing
anilist-mal-sync -c config.yaml sync --dry-run --all

# 4. Run sync
anilist-mal-sync -c config.yaml sync
```

Tokens are saved to `~/.config/anilist-mal-sync/token.json` by default.

### Docker

> **docker compose vs docker-compose:** Examples use `docker-compose` (CLI v1). If your system has Docker Compose v2 (bundled with modern Docker Desktop / Engine), replace `docker-compose` with `docker compose` (no hyphen).

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

Enable continuous sync using either a fixed interval or a cron schedule.

**Interval mode** (fixed interval between syncs):

```yaml
environment:
  - WATCH_INTERVAL=12h  # Sync every 12 hours
```

Or run watch command manually:
```bash
docker-compose run --rm sync watch --interval=12h
```

**Cron mode** (sync at specific times):

```yaml
environment:
  - WATCH_SCHEDULE=0 3 * * *  # Sync daily at 03:00
```

Or run watch command manually:
```bash
docker-compose run --rm sync watch --schedule "0 3 * * *"
```

**Cron expression examples:**

| Expression | Meaning |
|------------|---------|
| `0 3 * * *` | Every day at 03:00 |
| `*/30 * * * *` | Every 30 minutes |
| `0 */6 * * *` | Every 6 hours |
| `0 0 * * 1` | Every Monday at midnight |

**Interval limits (interval mode only):** 1h - 168h (7 days). `WATCH_INTERVAL` (or `--interval`) is **required** for watch mode when not using `--schedule`. Setting both `--interval` and `--schedule` (or their env/config equivalents) is an error.

### Scheduling (non-Docker)

Use the built-in cron schedule to sync at specific times:

```bash
# Sync daily at 3 AM using built-in cron support
anilist-mal-sync watch --schedule "0 3 * * *"

# With immediate first sync
anilist-mal-sync watch --schedule "0 3 * * *" --once
```

Cron expressions use standard 5-field syntax (`minute hour dom month dow`).
Expressions are evaluated in the **host/container local timezone** (`time.Local`).
Set the `TZ` environment variable on the container to change it (e.g. `TZ=Europe/Berlin`).

Or use your system's scheduler for one-off syncs:

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
- Run `anilist-mal-sync login` to reauthenticate (re-authenticates both services)

## Disclaimer

This project is not affiliated with AniList or MyAnimeList. Use at your own risk.

## Roadmap

- [ ] Sync rewatching and rereading counts

## Credits

- [anime-offline-database](https://github.com/manami-project/anime-offline-database) for JSON based anime dataset
- [arm-server](https://github.com/BeeeQueue/arm-server) for API anime dataset
- [Hato](https://github.com/Atelier-Shiori/Hato) for JSON API anime and manga
- [Jikan](https://jikan.moe/) for unofficial MyAnimeList API
