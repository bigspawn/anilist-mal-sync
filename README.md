# anilist-mal-sync [![Build Status](https://github.com/bigspawn/anilist-mal-sync/workflows/go/badge.svg)](https://github.com/bigspawn/anilist-mal-sync/actions)

Program to synchronize your AniList and MyAnimeList accounts.

## Features

- Sync AniList to MyAnimeList (anime and manga)
- Sync MyAnimeList to AniList (anime and manga) using `-reverse-direction` flag
- OAuth2 authentication with AniList and MyAnimeList
- CLI interface
- Configurable by environment variables and config file

## Usage

### Authentication

Configure your accounts in the respective sites, then authenticate using the CLI:

```bash
anilist-mal-sync login anilist
anilist-mal-sync login myanimelist
```

The program will print the URL to visit in your browser. After authorization, you'll be redirected and the token will be saved automatically.

Tokens are stored in `~/.config/anilist-mal-sync/token.json` and reused. To reauthenticate, run:
```bash
anilist-mal-sync logout <service>
```

#### AniList

1. Go to [AniList settings](https://anilist.co/settings/developer) (Settings -> Apps -> Developer)
2. Create a new client
3. Set the redirect URL to `http://localhost:18080/callback`
4. Set the client ID and client secret in the config file or as environment variables

#### MyAnimeList

1. Go to [MyAnimeList API Settings](https://myanimelist.net/apiconfig) (Profile -> Account Settings -> API)
2. Create a new application
3. Set the redirect URL to `http://localhost:18080/callback`
3. Set the client ID and client secret in the config file or as environment variables

### Configuration

#### Config file

```yaml
oauth:
  port: "18080" # Port for OAuth server to listen on (default: 18080).
  redirect_uri: "http://localhost:18080/callback" # Redirect URI for OAuth server (default: http://localhost:18080/callback).
anilist:
  client_id: "1" # AniList client ID.
  client_secret: "secret" # AniList client secret.
  auth_url: "https://anilist.co/api/v2/oauth/authorize"
  token_url: "https://anilist.co/api/v2/oauth/token"
  username: "username" # Your AniList username.
myanimelist:
  client_id: "1" # MyAnimeList client ID.
  client_secret: "secret" # MyAnimeList client secret.
  auth_url: "https://myanimelist.net/v1/oauth2/authorize"
  token_url: "https://myanimelist.net/v1/oauth2/token"
  username: "username" # Your MyAnimeList username.
token_file_path: "" # Absolute path to token file, empty string use default path `$HOME/.config/anilist-mal-sync/token.json`
```

#### Environment variables

- `PORT` - Port for OAuth server to listen on (default: 18080).
- `CLIENT_SECRET_ANILIST` - AniList client secret.
- `CLIENT_SECRET_MYANIMELIST` - MyAnimeList client secret.

### Commands

#### login
Authenticate with a service:
```bash
anilist-mal-sync login <service>
```
Services: `anilist`, `myanimelist`, `all`

#### logout
Logout from a service:
```bash
anilist-mal-sync logout <service>
```

#### status
Check authentication status:
```bash
anilist-mal-sync status
```

#### sync
Synchronize anime/manga lists:
```bash
anilist-mal-sync sync [options]
```

Options:
- `-c, --config` - Path to config file (default: `config.yaml`)
- `-f, --force` - Force sync all entries
- `-d, --dry-run` - Dry run without making changes
- `--manga` - Sync manga instead of anime
- `--all` - Sync both anime and manga
- `--verbose` - Enable verbose logging
- `--reverse-direction` - Sync from MyAnimeList to AniList

For backward compatibility, running `anilist-mal-sync [options]` without a command will execute sync.


### How to run

Requirements: Go 1.22 or later

Build and run from source:
```bash
git clone https://github.com/bigspawn/anilist-mal-sync.git
cd anilist-mal-sync
cp config.example.yaml config.yaml
# Edit config.yaml with your credentials
go run . login anilist
go run . login myanimelist
go run . sync
```

Or install the program:

```bash
go install github.com/bigspawn/anilist-mal-sync@latest
anilist-mal-sync
```

### Running with Docker

You can also run the application using Docker.

#### Using pre-built image

The image is available on GitHub Container Registry:

```bash
docker pull ghcr.io/bigspawn/anilist-mal-sync:latest
```

Run the container:

```bash
docker run \
    -p 18080:18080 \
    -v /path/to/your/config.yaml:/etc/anilist-mal-sync/config.yaml \
    -v /path/to/token/directory:/home/appuser/.config/anilist-mal-sync \
    ghcr.io/bigspawn/anilist-mal-sync:latest
```

#### Building your own image

1. Clone the repository: `git clone https://github.com/bigspawn/anilist-mal-sync.git`
2. Change directory: `cd anilist-mal-sync`
3. Build the Docker image: `docker build -t anilist-mal-sync .`
4. Run the container

#### Using Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3'

services:
  anilist-mal-sync:
    image: ghcr.io/bigspawn/anilist-mal-sync:latest
    ports:
      - "18080:18080"
    volumes:
      - ./config.yaml:/etc/anilist-mal-sync/config.yaml
      - ./tokens:/root/.config/anilist-mal-sync # it must be a directory
    environment:
      - CLIENT_SECRET_ANILIST=your_secret_here  # Optional
      - CLIENT_SECRET_MYANIMELIST=your_secret_here  # Optional
      - PORT=18080  # Optional
```

Run with Docker Compose:

```bash
docker-compose up
```

Note: When running in Docker, the browser authentication flow requires that port 18080 is exposed and accessible from your host machine. Also ensure that your token storage directory is mounted as a volume to preserve authentication between runs.

## Disclaimer

This project is not affiliated with AniList or MyAnimeList. Use at your own risk.
Both services have rate limits and the program can look like it's frozen or stop by timeout.
Just stop it and wait for a while and run again.

## TODO

- [ ] Sync favorites
- [x] Sync MAL to AniList
- [ ] Sync rewatching and rereading
