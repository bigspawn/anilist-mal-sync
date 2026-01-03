# anilist-mal-sync [![Build Status](https://github.com/Tareku99/anilist-mal-sync/workflows/go/badge.svg)](https://github.com/Tareku99/anilist-mal-sync/actions)

Program to synchronize your AniList and MyAnimeList accounts.

## Features

- Sync AniList to MyAnimeList (anime and manga) - default direction
- Sync MyAnimeList to AniList (anime and manga) using `-direction=mal-to-anilist`
- Bidirectional sync: sync both directions in a single run using `-direction=both`
- OAuth2 authentication with AniList and MyAnimeList
- CLI interface
- Configurable by environment variables and config file

## Usage

### Authentication

First configurate your accounts in the site.
Then run the program and follow the instructions to authenticate.
It prints the URL to visit in the browser.
Then you will be redirected to the callback URL and save the token.
After that you will go same steps for MyAnimeList.

Token will be saved in the `~/.config/anilist-mal-sync/token.json` file and reused then.
You can change the path in the config file.
If you want to reauthenticate, just delete the file.

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
score_normalization: true # Enable score normalization (AniList 100-point -> MAL 10-point scale). Default is true.
ignore_anime_titles: [] # List of anime titles to skip during sync (case-insensitive). Example: ["scott pilgrim takes off", "bocchi the rock! recap part 2"]
ignore_manga_titles: [] # List of manga titles to skip during sync (case-insensitive)
```

#### Environment variables

- `PORT` - Port for OAuth server to listen on (default: 18080).
- `CLIENT_SECRET_ANILIST` - AniList client secret.
- `CLIENT_SECRET_MYANIMELIST` - MyAnimeList client secret.

#### Configuration options

- `score_normalization` (boolean, default: `true`) - When enabled, automatically normalizes AniList scores > 10 to MAL's 0-10 integer scale. This prevents sync failures when AniList users use 100-point scoring systems. **Important:** MAL only accepts integer scores from 0-10, so this should remain enabled for AniList → MAL syncs.
- `ignore_anime_titles` (array of strings) - List of anime titles (case-insensitive) to skip during sync. Useful for titles that don't exist in the target service. Example: `["scott pilgrim takes off", "bocchi the rock! recap part 2"]`
- `ignore_manga_titles` (array of strings) - List of manga titles (case-insensitive) to skip during sync. Useful for titles that don't exist in the target service.

### Options

Program supports the following command-line options:

- `-c` - Path to the config file. Default is `config.yaml`.
- `-f` - Force sync (sync all entries, not just the ones that have changed). Default is false.
- `-d` - Dry run (do not make any changes). Default is false.
- `-manga` - Sync manga instead of anime. Default is anime.
- `-all` - Sync both anime and manga. Default is anime.
- `-verbose` - Print debug messages. Default is false.
- `-direction` - Sync direction: `anilist-to-mal` (AniList → MAL, default), `mal-to-anilist` (MAL → AniList), or `both` (bidirectional). Default is `anilist-to-mal`.
- `-h` - Print help message.


### How to run

Requirements:

- Go 1.22 or later

Build and run the program from source:

1. Clone the repository: `git clone https://github.com/Tareku99/anilist-mal-sync.git`
2. Change directory: `cd anilist-mal-sync`
3. Configure the program: `cp config.example.yaml config.yaml` and fill in the necessary fields
4. Run the program: `go run .`

Or install the program:

```bash
go install github.com/Tareku99/anilist-mal-sync@latest
anilist-mal-sync
```

### Running with Docker

You can also run the application using Docker.

#### Using pre-built image

The image is available on GitHub Container Registry:

```bash
docker pull ghcr.io/Tareku99/anilist-mal-sync:latest
```

Run the container:

```bash
docker run \
    -p 18080:18080 \
    -v /path/to/your/config.yaml:/etc/anilist-mal-sync/config.yaml \
    -v /path/to/token/directory:/home/appuser/.config/anilist-mal-sync \
    ghcr.io/Tareku99/anilist-mal-sync:latest
```

#### Building your own image

1. Clone the repository: `git clone https://github.com/Tareku99/anilist-mal-sync.git`
2. Change directory: `cd anilist-mal-sync`
3. Build the Docker image: `docker build -t anilist-mal-sync .`
4. Run the container

#### Using Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3'

services:
  anilist-mal-sync:
    image: ghcr.io/Tareku99/anilist-mal-sync:latest
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

## Vendor

This repository vendors third-party dependencies in the `vendor/` directory so the project can be built in offline or restricted environments (for example, some Unraid setups).

To update vendored dependencies locally:

```bash
go mod tidy
go mod vendor
```

If you update dependencies, commit the `vendor/` changes. CI validates that `vendor/` is in sync with `go.mod`.

## Disclaimer

This project is not affiliated with AniList or MyAnimeList. Use at your own risk.
Both services have rate limits and the program can look like it's frozen or stop by timeout.
Just stop it and wait for a while and run again.

## TODO

- [ ] Sync favorites
- [x] Sync MAL to AniList
- [ ] Sync rewatching and rereading
