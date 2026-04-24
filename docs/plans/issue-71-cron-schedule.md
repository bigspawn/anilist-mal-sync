# Plan: Cron schedule support for `watch` command

**Issue:** [#71](https://github.com/bigspawn/anilist-mal-sync/issues/71) — Support Cron expressions for watch interval

## Goal

Allow users to specify a cron expression for the `watch` command in addition to the
existing duration-based interval, so the container can run syncs at precise times
(e.g. `0 3 * * *` for daily at 03:00) without relying on an external cron.

## Priority and mutual exclusion

Each of `interval` and `schedule` is resolved from sources in this order
(highest wins):

1. CLI flag (`--interval`, `--schedule`)
2. Environment variable (`WATCH_INTERVAL`, `WATCH_SCHEDULE`)
3. YAML config (`watch.interval`, `watch.schedule`)

After resolution, if **both** `interval` and `schedule` are non-empty → fail hard
with:

```
watch mode accepts either --interval or --schedule, not both (got interval=%v, schedule=%q)
```

Rationale: mixing a CLI flag of one kind with a YAML value of the other usually
means the user forgot to remove the old setting. Failing loudly is safer than
silently picking one.

If neither is set — keep current behavior (error asking for `--interval` or
config value).

## Scope decisions (fixed)

- **No frequency limits for cron** — the `1h`–`168h` bounds apply only to
  `interval`. Cron expressions are accepted as-is.
- **Timezone:** `time.Local` (no new `WATCH_TIMEZONE` option for v1).
- **Cron format:** standard 5-field cron (`minute hour dom month dow`), parsed via
  `cron.ParseStandard`. No seconds field, no custom descriptors beyond what
  `robfig/cron/v3` provides out of the box.

## Implementation steps

### 1. Dependency

```bash
go get github.com/robfig/cron/v3
go mod tidy
go mod vendor   # project uses vendor/
```

### 2. Config (`config.go`)

- Extend struct:
  ```go
  type WatchConfig struct {
      Interval string `yaml:"interval"`
      Schedule string `yaml:"schedule"` // cron expression
  }
  ```
- `loadConfigFromEnv`: read `WATCH_SCHEDULE`.
- `overrideWatchFromEnv`: add override for `Schedule`.
- New method `(w *WatchConfig) Validate() error`:
  - if both `Interval` and `Schedule` non-empty → error.
  - if `Schedule` non-empty → validate via `cron.ParseStandard`.
- New method `(w *WatchConfig) ParseSchedule() (cron.Schedule, error)` (used by
  the watch command when building the cron runner).

### 3. CLI (`cmd_watch.go`)

- New flag:
  ```go
  &cli.StringFlag{
      Name:    "schedule",
      Aliases: []string{"s"},
      Usage:   "Cron expression (e.g. '0 3 * * *'). Mutually exclusive with --interval",
  }
  ```
- Resolve the effective `interval` and `schedule` values (CLI > env already in
  config > YAML).
- Call `WatchConfig.Validate()` on the resolved values → return error on
  conflict.
- Mode selection:
  - `schedule != ""` → cron mode
  - `interval != 0` → interval mode (existing behavior, unchanged)
  - otherwise → existing "interval required" error, updated to also mention
    `--schedule`.

### 4. Cron runtime

> **Note (post-implementation):** The actual implementation differs from the
> original design. `cron.New` / `c.Start` / `c.Stop` are **not** used.
> Instead:

- `cron.ParseStandard(expr)` produces a `cron.Schedule` value.
- A manual `time.Timer` loop drives ticks: `next := sched.Next(time.Now())`,
  `timer.Reset(time.Until(next))`.
- On each tick: call `runSyncOnce(ctx, runner, true)` (Refresh + Run), compute
  next tick, reset timer.
- `<-ctx.Done()` → return nil.
- `--once` flag: call `runSyncOnce(ctx, runner, false)` (Run only, no Refresh)
  before entering the timer loop.
- The internal `watchWithCronSchedule` helper accepts a `cronSchedule` interface
  (only `Next(time.Time) time.Time` needed), enabling tests to inject
  mock schedules with pre-computed fire times.

### 5. Refactor

Extract shared setup from `runWatch` so we don't blow past `funlen` (100 lines /
50 stmts):

- `buildWatchApp(ctx, cmd) (*App, Config, error)` — config load + logger + app.
- `runSyncOnce(ctx, app) error` — single iteration used by both modes and by
  `--once`.
- `watchWithInterval(ctx, app, interval)` — current ticker loop.
- `watchWithCron(ctx, app, expr)` — new cron loop.

`runWatch` becomes a thin dispatcher: build app → resolve values → validate →
branch.

### 6. Tests

- `cmd_watch_test.go`:
  - Bump expected flag count 14 → 15 in `TestWatchCommand_HasFlags`.
  - Add `"schedule"` to `expectedFlags`.
- New tests:
  - `TestWatchConfig_Validate_BothSet` — error when both fields non-empty.
  - `TestWatchConfig_Validate_InvalidCron` — error on malformed expression.
  - `TestWatchConfig_Validate_ValidCron` — no error.
  - `TestOverrideWatchFromEnv_Schedule` — `WATCH_SCHEDULE` overrides YAML.
  - `TestWatch_ResolveMode_*` — resolution precedence (CLI > env > YAML) for
    each field. Test the pure resolver function, not the goroutine loop.

### 7. Documentation

- `config.example.yaml`: add `schedule: ""` with examples in the comment
  (`"0 3 * * *"` daily at 03:00, `"*/30 * * * *"` every 30 min).
- `README.md`:
  - Update watch section: new `--schedule` flag, `WATCH_SCHEDULE` env.
  - Document mutual-exclusion rule.
  - Add a docker-compose snippet using `WATCH_SCHEDULE`.
- `CLAUDE.md`: note the new option in the `cmd_watch.go` row and any new files.
- `docker-compose.example.yaml`: commented `WATCH_SCHEDULE` example.

### 8. Quality gates

Run in order:

```bash
make fmt
make lint
make test
make check
```

Watch for `funlen` on `runWatch` after refactor; split further if needed.

## Out of scope (possible follow-ups)

- `WATCH_TIMEZONE` / `watch.timezone` to override `time.Local`.
- Minimum-frequency warning for sub-hour cron (to avoid API rate limits).
- 6-field cron with seconds.
