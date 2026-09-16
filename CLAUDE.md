# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go CLI application that synchronizes anime/manga lists bidirectionally between AniList and MyAnimeList. Uses OAuth2 authentication and supports both YAML config and environment variables.

## Build & Development Commands

```bash
make install    # Install all dev tools (golangci-lint, gofumpt, goimports, gci, govulncheck)
make build      # Build binary: go build -o anilist-mal-sync .
make test       # Run all tests: go test ./... -v
make generate   # Generate mocks using mockgen (run before tests if interfaces change)
make fmt        # Format with gofumpt
make lint       # Run golangci-lint (new issues only, fast feedback)
make lint-all   # Run full golangci-lint (all issues, used by CI gate)
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
  - **Forward Manga**: Manual → ID → HatoAPI → MangaBakaAPI → Title → Jikan → APISearch
  - **Reverse Anime**: Manual → ID → OfflineDB → HatoAPI → ARM → Title → MALID → APISearch
  - **Reverse Manga**: Manual → ID → HatoAPI → MangaBakaAPI → Title → Jikan → MALID → APISearch
- `Updater` - generic orchestrator with 3-phase pipeline (resolve → deduplicate → process) that uses strategies to match and update entries
- `FavoritesSync` - separate phase for favorites synchronization that runs after the main sync, using already-resolved AniList↔MAL ID mappings

### Main Components

| File | Purpose |
|------|---------|
| `app.go` | App structure & sync orchestration |
| `cli.go` | CLI interface (urfave/cli/v3) with 6 commands: login, logout, status, sync, watch, unmapped |
| `config.go` | Config loading (env vars take precedence over YAML) |
| `mapping_sources.go` | Registry of ARM/Hato/MangaBaka/Jikan CLI flags, env prefixes, and defaults |
| `oauth.go` | Token management & OAuth2 flow |
| `anilist.go` | AniList GraphQL client (via verniy library) with ToggleFavourite mutation |
| `myanimelist.go` | MAL REST API client (via go-myanimelist) |
| `anime.go` / `manga.go` | Domain models implementing Source/Target interfaces with IsFavourite field |
| `strategies.go` | Matching strategy implementations |
| `api_source.go` | Shared HTTP client + circuit breaker embedded by Hato and MangaBaka clients |
| `arm_api.go` | ARM API client for online ID mapping |
| `hato_api.go` / `hato_cache.go` | Hato API client for ID mapping with response caching |
| `mangabaka_api.go` / `mangabaka_cache.go` | MangaBaka API client for manga ID mapping, fallback after Hato |
| `mapping_cache.go` | Generic persistent mapping cache shared by Hato, Jikan, and MangaBaka |
| `jikan_api.go` / `jikan_cache.go` | Jikan API client for manga ID mapping with response caching and GetUserFavorites |
| `offline_database.go` | Offline database using anime-offline-database |
| `updater.go` | Generic 3-phase update orchestration (resolve, deduplicate, process) |
| `service.go` | MediaService interface and implementations |
| `favorites.go` | Favorites synchronization logic (MAL→AniList sync, AniList→MAL report only) |
| `mappings.go` | Manual AniList↔MAL mappings and ignore rules (YAML) |
| `unmapped.go` | Unmapped entries state persistence (JSON) |
| `cmd_sync.go` / `cmd_watch.go` | Sync and watch command implementations (watch supports both interval and cron schedule modes) |
| `cmd_login.go` / `cmd_logout.go` / `cmd_status.go` | Auth and status commands |
| `cmd_unmapped.go` | CLI command for managing unmapped entries |
| `report.go` | Sync report: warnings, unmapped items, duplicate conflicts, favorites mismatches |
| `statistics.go` | Sync statistics tracking and summary output with favorites results |
| `logger.go` | Leveled logger with color support, context-based logging |
| `logging.go` | HTTP round-tripper debug logging |
| `http_retry.go` | Exponential backoff retry logic |
| `docs/date-sync.md` | Documentation: date synchronization logic and behavior tables |

### Watch Modes

The `watch` command supports two mutually exclusive scheduling modes:

- **Interval mode**: fixed duration between syncs (`--interval` / `WATCH_INTERVAL` / `watch.interval`, range 1h–168h)
- **Cron mode**: sync at specific times via cron expression (`--schedule` / `WATCH_SCHEDULE` / `watch.schedule`, standard 5-field syntax)

Priority: CLI flag > env var > config YAML. Setting both modes is a validation error.

Both modes support `--once` for an immediate first sync. Cron mode uses `time.Local` timezone.

### Sync Flow

1. Load config (env vars or YAML)
2. Load manual mappings and ignore rules from `mappings.yaml`
3. Get OAuth tokens for both services
4. Fetch lists from source and target
5. **Resolve**: match entries using strategy chain (see Key Abstractions for per-direction chains)
6. **Deduplicate**: detect N:1 conflicts (multiple sources → same target), resolve by strategy priority
7. **Process**: update target service with changes
8. **Favorites sync** (if enabled with `--favorites`): synchronize favorites between services
   - MAL→AniList: add missing favorites on AniList (does not remove)
   - AniList→MAL: report mismatches only (MAL API does not support favorites write)
9. Save unmapped entries state for `unmapped` command
10. Print statistics and sync report

### Sync Directions

- **Default**: AniList → MyAnimeList
- **Reverse** (`--reverse-direction`): MyAnimeList → AniList

Start/end dates are preserved in both directions. See [docs/date-sync.md](docs/date-sync.md) for details.

### Favorites Synchronization

Favorites sync is an optional feature enabled via `--favorites` CLI flag or `FAVORITES_SYNC_ENABLED` env var. It runs as a separate phase after the main sync.

- **MAL → AniList**: Adds favorites from MAL to AniList. Does not remove favorites that exist only on AniList (user may have intentionally favorited different items on each service).
- **AniList → MAL**: Reports mismatches only. MAL API v2 does not support favorites read/write, so MAL favorites are fetched via Jikan API but cannot be modified.

Note: Favorites sync requires Jikan API to be enabled (automatically enabled when `--favorites` is used).

### ID Mapping Direction (read before adding a mapping source)

**An ID-mapping strategy must ask only for the ID the current sync direction needs.**
Never query "MAL→AniList first, then AniList→MAL" as a fixed order.

- forward sync (AniList→MAL): targets are MAL entries keyed by MAL ID → ask for the **MAL** ID
- reverse sync (MAL→AniList): targets are AniList entries keyed by AniList ID → ask for the **AniList** ID

The direction is on the entry itself: `Manga.isReverse` / `Anime.isReverse`, set by the
services when the list is fetched. Do not infer it from which ID happens to be populated —
an AniList entry usually carries its MAL ID too, so both are non-zero and inference picks wrong.

Use `directionalIDLookup` (`strategies.go`) rather than hand-rolling the branching.

**Why this matters more than a wasted request.** The resolved ID goes to
`checkExistingTarget`, which looks it up in a map keyed by the *target* service's IDs.
AniList and MAL ID ranges overlap heavily (both run well into the 100000s), so an ID
fetched in the wrong direction does not simply miss — it can land on an unrelated entry
and sync the wrong title. This was a real bug in Hato, MangaBaka and ARM: a live dry run
had 4 of 22 MangaBaka lookups going the wrong way.

Tests covering a strategy must set `isReverse` to match the data they use: a source with
only `IDMal` and AniList-keyed targets is a reverse sync and has to say so.

## Configuration

**Every option must be settable all three ways: CLI flag, env var, and config YAML.**
A new option is not finished until all three exist — adding only an env var leaves the
feature unreachable for anyone driving the tool from the command line, and the asymmetry
is invisible until a user hits it.

Priority is always: **CLI flag > env var > config YAML > built-in default.**

For a generic option (not an ID-mapping source), wiring it still touches four places:

| Where | What to add |
|-------|-------------|
| `config.go` | field with a `yaml:"..."` tag, default via `getEnvOrDefault`/`getEnvBoolOrDefault`, and an `override*FromEnv` entry |
| `cli.go` | flag in `syncFlags` **and** in the root command's flag list in `NewCLI` (they are separate lists), plus an `if cmd.IsSet(...)` branch in `applySyncFlagsToConfig` |
| `README.md` | row in the CLI flag table and an entry in the env var list |
| `docker-compose.example.yaml` | commented env var line |

Both flag lists in `cli.go` must be kept in sync: `syncFlags` serves the `sync` and `watch`
subcommands, while `NewCLI` builds its own list for the root command, which runs sync when
no subcommand is given.

For a boolean whose default is `true`, the flag needs `Value: true`; for a default of
`false`, omit `Value`. Because the branches are guarded by `cmd.IsSet`, an untouched flag
never overrides an env var.

### ID-mapping sources

ARM, Hato, MangaBaka and Jikan share one config shape (`MappingSourceConfig`: Enabled,
BaseURL, CacheDir, CacheMaxAge) and one descriptor table, `mappingSources()` in
`mapping_sources.go`. That table builds both `cli.go` flag lists, and drives `config.go`'s
env-var loading and overrides — so for these four sources the "two flag lists that must be
kept in sync by hand" trap above no longer applies. Adding one of these sources means:

1. a new entry in `mappingSources()` (flag names, env prefix, defaults)
2. a client (`*_api.go`, following the `HatoClient`/`MangaBakaClient` shape — embed
   `apiSource` for the shared HTTP client and circuit breaker if the source can go down)
3. a strategy, wired into `loadIDMappingStrategies` and the relevant chain(s) in `app.go`
4. a row in `README.md`'s flag table and env var list, and a commented line in
   `docker-compose.example.yaml`
5. a section in both YAML examples — README's `config.yaml` block and `config.example.yaml` —
   plus the source in README's strategy chains with the right "enabled/disabled by default"

`docs_test.go` checks steps 4 and 5 against `mappingSources()`.

Offline database is not in this table: it has its own config shape (`AutoUpdate`,
`ForceRefresh`) and no base-URL flag, so it stays wired the generic way above.

Defaults (see `mappingSources()` for the authoritative table):

| Source | CLI flag | Env var | Default |
|--------|----------|---------|---------|
| Offline DB | `--offline-db` | `OFFLINE_DATABASE_ENABLED` | `true` |
| ARM | `--arm-api` | `ARM_API_ENABLED` | `false` |
| Hato | `--hato-api` | `HATO_API_ENABLED` | `true` |
| MangaBaka | `--mangabaka-api` | `MANGABAKA_API_ENABLED` | `false` |
| Jikan | `--jikan-api` | `JIKAN_API_ENABLED` | `false` |

## Code Quality & Linting

### Linter Configuration

`make lint` runs `golangci-lint run --new` for fast feedback on new issues only.
`make lint-all` runs the full linter (`golangci-lint run --timeout=5m`), which CI gates on.

The `.golangci.yml` enforces strict limits to prevent overly complex code:

- **funlen**: 100 lines max, 50 statements max
- **gocyclo**: 15 complexity max
- **cyclop**: 25 complexity max
- **nestif**: 4 depth max
- **lll**: 140 characters max line length

Test files are exempt from:
- **funlen**, **gocyclo**, **cyclop**, **nestif** — test fixtures with many cases need flexibility
- **goconst** — test fixture strings like "test", "interval", "Some Title" should not become named constants (fixtures are not tied to production values and extra constants reduce readability)
- **gosec G101** — test files contain obvious fixture values like `test_client_secret` and example.com URLs that are not real credentials (same reasoning: fixtures, not security issues)

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
- Jikan API — custom HTTP client in `jikan_api.go` (no external library)
- `github.com/urfave/cli/v3` - CLI framework
- HTTP retry — custom exponential backoff in `http_retry.go` (no external library)
- `gopkg.in/yaml.v2` - Config file parsing (`config.go`)
- `gopkg.in/yaml.v3` - Mappings file parsing with comments (`mappings.go`)
- `go.uber.org/mock` - Mock generation for tests
- `github.com/robfig/cron/v3` - Cron expression parsing for watch schedule
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
