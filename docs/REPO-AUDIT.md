Repo audit summary and recommendations
=====================================

High level recommendations:
- Move `main`-package responsibilities into `cmd/anilist-mal-sync` and create `internal/*` packages for clients and sync logic.
- Reduce globals: pass `Config` and dependencies via an `App` struct.
- Add unit tests for core translation logic (score normalization, matching strategies).

Proposed package layout (minimal):
- `cmd/anilist-mal-sync/main.go` — CLI wiring and flag parsing
- `internal/config` — configuration loader (done)
- `internal/oauth` — OAuth helpers (done)
- `internal/anilist` — AniList client (done)
- `internal/mal` — MAL client (done)
- `internal/sync` — Updater/Strategy logic (TODO)

Next steps to finish refactor:
1. Move `Updater`, `StrategyChain`, and interfaces into `internal/sync`.
2. Replace globals (`dryRun`, `forceSync`, `verbose`) with fields on `App` and pass through.
3. Add unit tests for `normalizeScoreForMAL` and strategy matching.
