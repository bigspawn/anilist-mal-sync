# anilist-mal-sync

Programm to synchronize your AniList and MyAnimeList accounts.

## Features

- Sync AniList to MyAnimeList (currently only syncing anime)
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
token_file_path: "" # Absolute path to token file, empty string use default path.
```

#### Environment variables

- `PORT` - Port for OAuth server to listen on (default: 18080).
- `CLIENT_SECRET_ANILIST` - AniList client secret.
- `CLIENT_SECRET_MYANIMELIST` - MyAnimeList client secret.

### Options

Program supports the following command-line options:

- `-c` - Path to the config file. Default is `config.yaml`.
- `-f` - Force sync (sync all entries, not just the ones that have changed). Default is false.
- `-d` - Dry run (do not make any changes to MyAnimeList). Default is false.

### How to run

1. Clone the repository: `git clone https://github.com/bigspawn/anilist-mal-sync.git`
2. Change directory: `cd anilist-mal-sync`
3. Configure the program: `cp config.example.yaml config.yaml` and fill in the necessary fields
4. Run the program: `go run .`

## Disclaimer

This project is not affiliated with AniList or MyAnimeList. Use at your own risk.
Both services have rate limits and programm can be looks like freezed or stop by timeout.
Just stop it and wait for a while and run again.

## TODO

- [ ] Add documentation
- [ ] Sync favorites
- [ ] Sync manga
- [ ] Sync MAL to AniList
