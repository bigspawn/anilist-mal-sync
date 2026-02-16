# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go CLI application that synchronizes anime/manga lists bidirectionally between AniList and MyAnimeList. Uses OAuth2 authentication and supports both YAML config and environment variables.

## Build & Development Commands

```bash
make install    # Install all dev tools (golangci-lint, gofumpt, goimports, gci, govulncheck, lefthook)
make build      # Build binary: go build -o anilist-mal-sync ./cmd/main.go
make test       # Run all tests: go test ./... -v
make generate   # Generate mocks using mockgen (run before tests if interfaces change)
make fmt        # Format with gofumpt
make lint       # Run golangci-lint (new issues only)
make check      # Run all checks: generate + format + imports + vet + lint + test
make clean      # Remove artifacts and test cache
```

### Running Single Tests

```bash
go test -run TestFunctionName ./...
go test -run TestConfig ./...  # Run all tests matching "TestConfig"
```

## Architecture

### Key Abstractions

The codebase uses a **strategy pattern** for matching entries between services:

- `Source` / `Target` interfaces - abstract data from source/target services
- `StrategyChain` - chains multiple matching strategies (varies by direction and media type):
  - **Forward Anime**: Manual → ID → OfflineDB → HatoAPI → ARM → Title → APISearch
  - **Forward Manga**: Manual → ID → HatoAPI → Title → Jikan → APISearch
  - **Reverse Anime**: Manual → ID → OfflineDB → HatoAPI → ARM → Title → MALID → APISearch
  - **Reverse Manga**: Manual → ID → HatoAPI → Title → Jikan → MALID → APISearch
- `Updater` - generic orchestrator with 3-phase pipeline (resolve → deduplicate → process) that uses strategies to match and update entries

### Main Components

| File | Purpose |
|------|---------|
| `app.go` | App structure & sync orchestration |
| `cli.go` | CLI interface (urfave/cli/v3) with 6 commands: login, logout, status, sync, watch, unmapped |
| `config.go` | Config loading (env vars take precedence over YAML) |
| `oauth.go` | Token management & OAuth2 flow |
| `anilist.go` | AniList GraphQL client (via verniy library) |
| `myanimelist.go` | MAL REST API client (via go-myanimelist) |
| `anime.go` / `manga.go` | Domain models implementing Source/Target interfaces |
| `strategies.go` | Matching strategy implementations |
| `arm_api.go` | ARM API client for online ID mapping |
| `hato_api.go` / `hato_cache.go` | Hato API client for ID mapping with response caching |
| `jikan_api.go` / `jikan_cache.go` | Jikan API client for manga ID mapping with response caching |
| `offline_database.go` | Offline database using anime-offline-database |
| `updater.go` | Generic 3-phase update orchestration (resolve, deduplicate, process) |
| `service.go` | MediaService interface and implementations |
| `mappings.go` | Manual AniList↔MAL mappings and ignore rules (YAML) |
| `unmapped.go` | Unmapped entries state persistence (JSON) |
| `cmd_sync.go` / `cmd_watch.go` | Sync and watch command implementations |
| `cmd_login.go` / `cmd_logout.go` / `cmd_status.go` | Auth and status commands |
| `cmd_unmapped.go` | CLI command for managing unmapped entries |
| `report.go` | Sync report: warnings, unmapped items, duplicate conflicts |
| `statistics.go` | Sync statistics tracking and summary output |
| `logger.go` | Leveled logger with color support, context-based logging |
| `logging.go` | HTTP round-tripper debug logging |
| `http_retry.go` | Exponential backoff retry logic |

### Sync Flow

1. Load config (env vars or YAML)
2. Load manual mappings and ignore rules from `mappings.yaml`
3. Get OAuth tokens for both services
4. Fetch lists from source and target
5. **Resolve**: match entries using strategy chain (see Key Abstractions for per-direction chains)
6. **Deduplicate**: detect N:1 conflicts (multiple sources → same target), resolve by strategy priority
7. **Process**: update target service with changes
8. Save unmapped entries state for `unmapped` command
9. Print statistics and sync report

### Sync Directions

- **Default**: AniList → MyAnimeList
- **Reverse** (`--reverse-direction`): MyAnimeList → AniList

## Code Quality Rules

The `.golangci.yml` enforces strict limits to prevent overly complex code:

- **funlen**: 100 lines max, 50 statements max
- **gocyclo**: 15 complexity max
- **cyclop**: 25 complexity max
- **nestif**: 4 depth max
- **lll**: 140 characters max line length

Test files are exempt from complexity checks.

### Git Hooks (Lefthook)

Pre-commit runs: gofumpt, goimports, golangci-lint (new issues only), go vet
Pre-push runs: full golangci-lint, complete test suite

## Logging

The codebase uses a `Logger` struct (`logger.go`) with 4 levels: Error, Warn, Info, Debug.

Context-based free functions (require `context.Context` with logger):
- `LogDebug(ctx, format, args...)` — verbose mode only
- `LogWarn(ctx, format, args...)` — always shown
- `LogInfo(ctx, format, args...)` — normal mode
- `LogStage(ctx, format, args...)` — section headers
- `LogProgress(ctx, current, total, status, title)` — progress bars
- `LogInfoSuccess(ctx, format, args...)`, `LogInfoUpdate(ctx, ...)`, `LogInfoDryRun(ctx, ...)`

**Never** use `log.Printf` directly or `DPrintf` (deprecated no-op in `updater.go`):
```go
// ❌ Bad - raw log or deprecated DPrintf
log.Printf("[DEBUG] message")
DPrintf("[DEBUG] message")

// ✅ Good - use context-based logging
LogDebug(ctx, "Processing item %d", id)
```

## Dependencies

- `github.com/rl404/verniy` - AniList GraphQL client
- `github.com/nstratos/go-myanimelist` - MAL API client
- `github.com/Sethispr/jikanGo` - Jikan API client for manga
- `github.com/urfave/cli/v3` - CLI framework
- `github.com/cenkalti/backoff/v4` - Retry logic
- `gopkg.in/yaml.v3` - YAML marshaling with comments
- `go.uber.org/mock` - Mock generation for tests
- `anime-offline-database` - Offline ID mapping (downloaded from GitHub releases)

## Testing Notes

Test files follow the pattern `*_test.go` in the root directory. Main test areas:
- CLI structure and flags (`cli_test.go`)
- Config loading from env vars (`config_test.go`)
- OAuth flows (`oauth_test.go`)
- Domain logic: anime, manga, strategies, score normalization
- Mappings: load/save, manual mapping, ignore rules (`mappings_test.go`)
- Unmapped state: save/load, JSON round-trip (`unmapped_test.go`)
- Updater: deduplication, duplicate target detection (`updater_test.go`)
