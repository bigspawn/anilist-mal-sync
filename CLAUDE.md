# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go application that synchronizes anime and manga lists between AniList and MyAnimeList accounts. The application uses OAuth2 authentication to access both services and supports bidirectional synchronization:
- AniList to MyAnimeList (default)
- MyAnimeList to AniList (with `-reverse-direction` flag)

## Architecture

The application follows a modular architecture with the following key components:

- **Main entry point**: `main.go` - handles CLI flags and application initialization
- **Configuration**: `config.go` - manages YAML configuration and environment variables
- **Application core**: `app.go` - coordinates OAuth clients and sync operations
- **API clients**: `anilist.go` and `myanimelist.go` - handle API interactions with respective services
- **Media types**: `anime.go` and `manga.go` - define data structures and transformations
- **Sync logic**: `updater.go` - handles the synchronization process between services
- **OAuth handling**: `oauth.go` - manages OAuth2 authentication flows
- **Statistics**: `statistics.go` - tracks sync operations and results
- **Strategies**: `strategies.go` - implements different sync strategies and entry matching logic

### Core Sync Pattern

The synchronization follows a generic pattern defined in `updater.go`:
- **Source/Target Interface**: Uses `Source` and `Target` interfaces for type-safe operations
- **Updater struct**: Contains function pointers for getting targets by ID/name and updating them
- **Comparison logic**: Implements progress comparison and diff generation between platforms
- **Ignore list**: Supports title-based filtering for entries that don't exist on target platform
- **Force sync**: Optional flag to bypass progress comparison and sync all entries
- **Strategy chain**: Uses configurable strategies for finding and matching entries between platforms

The `App` struct instantiates separate `Updater` instances for anime and manga in both directions (normal and reverse), each with their own function implementations for API operations.

## Common Development Commands

### Building
```bash
# Build the application
make build

# Build using go directly
go build -o anilist-mal-sync

# Build for Docker (used in multi-stage build)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-w -s" -o main
```

Note: The Makefile builds from `./cmd/main.go` but the actual main file is in the root directory as `main.go`.

### Testing
```bash
# Run all tests
make test

# Run tests with verbose output
go test ./... -v

# Test files:
# - anime_test.go: Tests for score normalization/denormalization (anime)
# - manga_test.go: Tests for score normalization/denormalization (manga)
```

### Code Formatting
```bash
# Format code with gofumpt
make fmt

# This ensures code follows gofumpt standards required by CI
```

### Linting
```bash
# Run linter using Docker (recommended)
make lint

# Run linter directly (if golangci-lint is installed locally)
golangci-lint run --new

# This uses golangci-lint v2.2.2 in Docker container
# Note: dupl linter is disabled for test files to allow identical test structures
```

### Full Check (All Git Hooks)
```bash
# Run all checks that Git hooks run (format + imports + lint + vet + test)
make check
```

This runs the same checks as Git hooks:
1. **gofumpt** - Format code
2. **goimports** - Organize imports
3. **go vet** - Static analysis
4. **golangci-lint** - Full lint check
5. **go test** - Run all tests

Use this before pushing or when CI fails locally.

### Development Setup
```bash
# One-command setup - installs all development tools and Git hooks
make install
```

This installs:
- **golangci-lint** v2.2.2 (via brew)
- **gofumpt** (via go install)
- **goimports** (via go install)
- **gci** - import organizer (via go install)
- **govulncheck** - vulnerability scanner (via go install)
- **lefthook** - Git hooks manager (via brew)
- Git hooks configuration

### Git Hooks (Lefthook)
```bash
# Install Git hooks for automatic linting/formatting on commit
make hooks-install

# Uninstall Git hooks
make hooks-uninstall
```

**What is Lefthook?**
Lefthook is a Git hooks manager that automatically runs linting and formatting before commits and pushes. When installed:
- **Pre-commit**: Runs gofumpt, goimports, and golangci-lint on staged files
- **Pre-push**: Runs full lint check and tests

**Note:** `make install` automatically installs and configures Git hooks. Configuration is in `lefthook.yml` at the repository root.

### Running
```bash
# Run with default config
go run .

# Run with custom config file
go run . -c myconfig.yaml

# Run with dry run mode (no actual updates)
go run . -d

# Run with verbose logging
go run . -verbose

# Sync manga instead of anime
go run . -manga

# Sync both anime and manga
go run . -all

# Force sync all entries
go run . -f

# Reverse sync direction (MyAnimeList to AniList)
go run . -reverse-direction
```

### Cleanup
```bash
# Clean build artifacts and test cache
make clean
```

### Docker Development
```bash
# Build Docker image
docker build -t anilist-mal-sync .

# Run with environment variables (recommended)
docker run -p 18080:18080 \
  -e PUID=$(id -u) \
  -e PGID=$(id -g) \
  -e ANILIST_CLIENT_ID=your_id \
  -e ANILIST_CLIENT_SECRET=your_secret \
  -e ANILIST_USERNAME=your_username \
  -e MAL_CLIENT_ID=your_id \
  -e MAL_CLIENT_SECRET=your_secret \
  -v tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest

# Or with config file (optional)
docker run -p 18080:18080 \
  -e PUID=$(id -u) \
  -e PGID=$(id -g) \
  -v /path/to/config.yaml:/etc/anilist-mal-sync/config.yaml:ro \
  -v /path/to/tokens:/home/appuser/.config/anilist-mal-sync \
  ghcr.io/bigspawn/anilist-mal-sync:latest -c /etc/anilist-mal-sync/config.yaml
```

**Volume Permissions:**
The image implements LinuxServer.io-style PUID/PGID environment variables to handle volume permissions:
- Set `PUID` and `PGID` to your user's UID/GID (from `id -u` / `id -g`)
- Entrypoint script (`docker-entrypoint.sh`) adjusts container user to match
- Files created by the container will have correct ownership on the host
- No manual `chown` required

## Configuration

The application can be configured via environment variables (recommended for Docker) or YAML config file.

**Required environment variables:**
- `ANILIST_CLIENT_ID` - AniList client ID
- `ANILIST_CLIENT_SECRET` - AniList client secret
- `ANILIST_USERNAME` - AniList username
- `MAL_CLIENT_ID` - MyAnimeList client ID
- `MAL_CLIENT_SECRET` - MyAnimeList client secret

**Optional environment variables:**
- `WATCH_INTERVAL` - Sync interval for watch mode (e.g., `12h`)
- `OAUTH_PORT` - OAuth server port (default: `18080`)
- `OAUTH_REDIRECT_URI` - OAuth redirect URI
- `TOKEN_FILE_PATH` - Token file path for persistent authentication
- `PUID` / `PGID` - User/Group ID for Docker volume permissions

**Config file:** Use `-c config.yaml` flag to load from YAML file instead of environment variables.

## Key Dependencies

- `golang.org/x/oauth2` - OAuth2 client implementation
- `gopkg.in/yaml.v2` - YAML configuration parsing
- `github.com/nstratos/go-myanimelist` - MyAnimeList API client
- `github.com/rl404/verniy` - AniList API client
- `github.com/cenkalti/backoff/v4` - Exponential backoff for API retry logic

### Important: Dependency Management
**Always run `go mod vendor` after `go get` to keep the vendor directory in sync with go.mod changes.**

## Authentication Flow

The application implements OAuth2 authentication for both services:
1. Starts a local server on port 18080 (configurable)
2. Opens browser to service authorization URL
3. Handles callback and exchanges code for access token
4. Stores tokens in `~/.config/anilist-mal-sync/token.json`

## Sync Process

1. Fetches user's anime/manga list from AniList
2. Fetches user's anime/manga list from MyAnimeList
3. Normalizes AniList scores to 0-10 format (see Score Normalization below)
4. Compares entries and identifies differences
5. Updates MyAnimeList entries to match AniList status
6. In reverse sync, denormalizes scores back to user's AniList format
7. Provides statistics on sync operations

### Score Normalization

The application handles score format differences between AniList and MyAnimeList:

**AniList Score Formats**:
- `POINT_100` (0-100) - e.g., 85/100
- `POINT_10_DECIMAL` (0-10.0) - e.g., 8.5/10
- `POINT_10` (0-10) - e.g., 8/10
- `POINT_5` (0-5) - e.g., 4/5
- `POINT_3` (1-3) - e.g., 2/3

**MyAnimeList Score Format**:
- Integer 0-10 only

**Implementation**:
- Scores are stored internally as `int` in normalized 0-10 format
- When reading from AniList: scores are normalized to 0-10
- When writing to AniList: scores are denormalized back to user's format
- When reading/writing MAL: no conversion needed (already 0-10)
- Functions: `normalizeScoreForMAL()` and `denormalizeScoreForAniList()` in `anime.go` and `manga.go`
- User's score format is retrieved once at startup via `GetUserScoreFormat()` in `anilist.go`

This architecture prevents MAL API errors when AniList scores exceed 10 (e.g., `400 invalid score bad_request`)

## Docker Support

The application includes Docker support with:
- Multi-stage build using Alpine base image
- Non-root user execution (appuser with UID 10001)
- Volume mounts for configuration and token storage
- Port exposure for OAuth callback (18080)
- Vendored dependencies for faster builds

## Development Notes

### Rate Limiting
Both AniList and MyAnimeList have rate limits. The application may appear to freeze due to API timeouts. If this occurs, stop the application and wait before retrying.

### Entry Matching
The sync process uses two methods to match entries:
1. **By ID**: When AniList entry has a MAL ID reference
2. **By Title**: When no ID is available, searches MAL by title and matches by media type

### Debug Output
Use the `-verbose` flag to enable detailed logging of the sync process, including API calls and comparison results.

### Ignore List
Hard-coded ignore lists in `app.go` skip entries that don't exist on the target platform (e.g., "Scott Pilgrim Takes Off" is not available on MAL).

## AI Agent Development Rules

### Code Quality Requirements

**CRITICAL**: All code generated or modified MUST pass linting without introducing new warnings.

#### Linting Workflow
1. Run `make check` after significant code changes (runs all Git hooks checks)
2. Run `golangci-lint run --new` for quick feedback during development
3. Run `make fmt` (gofumpt) on all Go files after changes
4. Fix all linting issues before considering work complete
5. If linting fails, fix the issues - do not bypass

**Before committing**: Always run `make check` to ensure all checks pass.

#### Code Complexity Limits
- **Function length**: Maximum 60 lines
- **Function statements**: Maximum 40 statements
- **Cyclomatic complexity**: Maximum 15
- **If nesting**: Maximum 4 levels
- **Maintainability index**: Minimum 20/100

#### Code Style Rules
- Preallocate slices with known capacity using `make([]T, 0, capacity)`
- Use explicit struct types instead of `map[string]interface{}` where possible
- Avoid capturing loop variables in goroutines (use `copyloopvar` or explicit copies)
- Never return `(nil, nil)` - use proper error handling
- Add `json`, `yaml` tags for exported struct fields used in serialization

#### Error Handling
- Always check errors returned by functions
- Use `errcheck` compliance - no unchecked errors allowed
- Prefer wrapped errors with context: `fmt.Errorf("context: %w", err)`

#### Import Organization
- Imports are automatically organized by `gci` formatter (standard, project, third-party)
- Run `make fmt` to format imports after changes

#### Testing
- Write tests for new functionality
- Test files (`_test.go`) have relaxed `dupl` rules for identical test structures
- Maintain existing test coverage

#### Dependencies
- Always run `go mod vendor` after `go get`
- Keep vendor directory in sync with go.mod

## Additional Tooling

### IDE Integration
For optimal development experience with AI assistants:

1. **gopls** (Go Language Server)
   - Built-in analysis: unusedparams, shadow, nilness, fieldalignment
   - Provides real-time feedback in your IDE

2. **govulncheck**
   - Official Go vulnerability scanner
   - Run: `go install golang.org/x/vuln/cmd/govulncheck@latest`
   - Usage: `govulncheck ./...`

3. **nilaway** (Optional)
   - Uber's nil-safety analyzer for compile-time nil pointer detection
   - Catches nil pointer dereferences before runtime

### CI/CD Integration
- Use GitHub Actions with `only-new-issues: true` for PRs
- Cache `~/.cache/golangci-lint` using go.sum as cache key (3-5x speedup)
- Configure timeout to prevent CI blocking (currently 5m)
- Reject commits that fail linting

### Performance Optimization
The `.golangci.yml` includes performance optimizations:
- Build cache enabled
- Readonly module download mode
- 4-thread concurrency for parallel linting
- Specific file exclusions (e.g., `.pb.go` generated files)