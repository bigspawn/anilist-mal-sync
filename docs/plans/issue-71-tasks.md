# Tasks: Cron schedule support for `watch` command

> See [issue-71-fixes.md](./issue-71-fixes.md) for post-review corrections applied after the initial implementation.

Detailed, sequential implementation tasks for [issue-71-cron-schedule.md](./issue-71-cron-schedule.md).
Each task is self-contained and ends with a green quality gate (`make fmt && make lint && make test`).

**Conventions used below:**
- TDD where feasible: write the failing test, then the code.
- Each task lists files touched, the acceptance check, and required tests.
- Tasks are ordered so the tree stays compilable and tests stay green between
  commits.

---

## Phase 1 — Foundation

### Task 1.1 — Add `robfig/cron/v3` dependency ~~DONE~~

**Files:** `go.mod`, `go.sum`, `vendor/`

**Steps:**
1. `go get github.com/robfig/cron/v3@latest`
2. `go mod tidy`
3. `go mod vendor`
4. `make build` — confirm nothing else broke.

**Acceptance:**
- `grep robfig go.mod` shows the dependency.
- `vendor/github.com/robfig/cron/v3/` exists.
- `make build` succeeds.

**Tests:** none (infra step).

---

### Task 1.2 — Extend `WatchConfig` struct with `Schedule` ~~DONE~~

**Files:** `config.go`, `config_test.go`

**Steps:**
1. Add `Schedule string \`yaml:"schedule"\`` to `WatchConfig`.
2. Update `loadConfigFromEnv`: `Schedule: os.Getenv("WATCH_SCHEDULE")`.
3. Update `overrideWatchFromEnv`: `overrideStringFromEnv(&watch.Schedule, "WATCH_SCHEDULE")`.

**Tests (write first):**
- `TestLoadConfigFromEnv_ReadsWatchSchedule` — set `WATCH_SCHEDULE=0 3 * * *`,
  expect `cfg.Watch.Schedule == "0 3 * * *"`.
- `TestOverrideWatchFromEnv_SetsSchedule` — start with `Schedule: ""`, set env,
  expect override applied.
- `TestOverrideWatchFromEnv_ScheduleEmptyEnvKeepsYAMLValue` — start with
  `Schedule: "0 5 * * *"`, no env, expect unchanged.
- `TestLoadConfigFromFile_ParsesSchedule` — YAML fixture with `watch.schedule`
  parses into the field.

**Acceptance:**
- New tests green.
- `make lint` clean.

---

### Task 1.3 — Add `WatchConfig.Validate()` method ~~DONE~~

**Files:** `config.go`, `config_test.go`

**Validate rules:**
1. Both `Interval` and `Schedule` non-empty → `errors.New("watch mode accepts either interval or schedule, not both")` (include the values for debuggability via `%w`-free `fmt.Errorf`).
2. `Interval` non-empty → must parse via `time.ParseDuration` **and** be within
   `[minInterval, maxInterval]` (reuse existing constants — lift bounds check
   from `cmd_watch.go` to keep one source of truth).
3. `Schedule` non-empty → must parse via `cron.ParseStandard`.
4. Both empty → return a sentinel error `ErrWatchConfigMissing` so the CLI layer
   can produce its user-facing help message.

**Tests (write first, one per branch):**
- `TestWatchConfig_Validate_BothSet_ReturnsError`
- `TestWatchConfig_Validate_IntervalOnly_Valid`
- `TestWatchConfig_Validate_IntervalBelowMin_ReturnsError` (e.g. `30m`)
- `TestWatchConfig_Validate_IntervalAboveMax_ReturnsError` (e.g. `200h`)
- `TestWatchConfig_Validate_IntervalMalformed_ReturnsError` (e.g. `xyz`)
- `TestWatchConfig_Validate_ScheduleOnly_Valid` (`0 3 * * *`)
- `TestWatchConfig_Validate_ScheduleMalformed_ReturnsError` (`not a cron`)
- `TestWatchConfig_Validate_ScheduleFieldCount_ReturnsError` (`* * *` — only 3)
- `TestWatchConfig_Validate_BothEmpty_ReturnsMissingSentinel` —
  `errors.Is(err, ErrWatchConfigMissing)`.

**Acceptance:** all tests green, no logic in `cmd_watch.go` changes yet.

---

### Task 1.4 — Add `WatchConfig.ParseSchedule()` helper ~~DONE~~

**Files:** `config.go`, `config_test.go`

**Signature:** `func (w *WatchConfig) ParseSchedule() (cron.Schedule, error)`

**Behavior:** thin wrapper around `cron.ParseStandard(w.Schedule)`; returns
`cron.Schedule` ready for `cron.New().Schedule(...)`.

**Tests (write first):**
- `TestWatchConfig_ParseSchedule_ValidReturnsSchedule` — non-nil result, `.Next(now)` advances.
- `TestWatchConfig_ParseSchedule_InvalidReturnsError`.
- `TestWatchConfig_ParseSchedule_EmptyReturnsError`.

**Acceptance:** tests green.

---

## Phase 2 — CLI resolution logic

### Task 2.1 — Extract `resolveWatchConfig` from `runWatch` ~~DONE~~

Pure function that merges CLI + already-loaded config into final `interval` /
`schedule` values. Makes the resolution testable without spinning up an app.

**Files:** `cmd_watch.go`, `cmd_watch_test.go`

**Signature:**
```go
func resolveWatchConfig(cmd *cli.Command, cfg WatchConfig) (WatchConfig, error)
```

**Logic:**
- `interval` source precedence: CLI `--interval` → `cfg.Interval` (already carries env override from `loadConfigFromFile` / `loadConfigFromEnv`).
- `schedule` same precedence for CLI `--schedule`.
- Return `WatchConfig{Interval, Schedule}` + call `Validate()` on the result.

**Tests (write first):**
- `TestResolveWatchConfig_CLIIntervalOverridesYAMLInterval`.
- `TestResolveWatchConfig_CLIScheduleOverridesYAMLSchedule`.
- `TestResolveWatchConfig_CLIIntervalAndYAMLSchedule_ReturnsError` — verifies the
  "both effective" rule across layers.
- `TestResolveWatchConfig_CLIScheduleAndYAMLInterval_ReturnsError`.
- `TestResolveWatchConfig_Neither_ReturnsMissingSentinel`.
- `TestResolveWatchConfig_OnlyInterval_OK`.
- `TestResolveWatchConfig_OnlySchedule_OK`.

**Acceptance:**
- Tests green.
- `runWatch` refactored to call this helper; behavior unchanged for interval
  mode.

---

### Task 2.2 — Add `--schedule` CLI flag ~~DONE~~

**Files:** `cmd_watch.go`, `cmd_watch_test.go`

**Steps:**
1. Append `&cli.StringFlag{Name: "schedule", Aliases: []string{"s"}, Usage: "..."}` to `watchFlags`.
2. In `resolveWatchConfig`, read `cmd.String("schedule")`.

**Tests (update + add):**
- Update `TestWatchCommand_HasFlags`: expected count 14 → 15, add `"schedule"`
  to `expectedFlags`.
- New `TestWatchCommand_ScheduleFlag_Exists` — flag is a `*cli.StringFlag`,
  correct alias `-s`, non-empty usage.
- New `TestWatchCommand_ScheduleFlag_NoDefault` — default value `""`.
- Update `TestCLI_HasCommands` if it counts flags transitively (check file).

**Acceptance:**
- Tests green.
- `./anilist-mal-sync watch --help` shows the new flag.

---

## Phase 3 — Runtime refactor

### Task 3.1 — Extract `buildWatchApp` helper ~~DONE~~

**Files:** `cmd_watch.go`, `cmd_watch_test.go`

**Signature:**
```go
func buildWatchApp(ctx context.Context, cmd *cli.Command) (context.Context, *App, Config, error)
```

Encapsulates: load config, apply flags, init logger, wrap context, `NewApp`.

**Tests:** not easily unit-testable (touches OAuth / network via `NewApp`);
covered indirectly by existing integration tests and by the fact that
`runWatch` keeps its current behavior — rely on the existing watch tests to
catch regressions.

**Acceptance:**
- `runWatch` slimmer.
- `make test` still green.

---

### Task 3.2 — Extract `runSyncOnce` helper ~~DONE~~

**Files:** `cmd_watch.go`, `cmd_watch_test.go`

**Signature:**
```go
func runSyncOnce(ctx context.Context, app *App, label string) error
```

`label` is the log prefix ("Initial sync", "Scheduled sync"). Centralises
`app.Refresh` + `app.Run` + error logging.

**Tests:** behavior covered by interval-mode path, unchanged.

**Acceptance:** `runWatch` and both mode functions (next tasks) reuse this.

---

### Task 3.3 — Extract `watchWithInterval` ~~DONE~~

**Files:** `cmd_watch.go`

Move the existing ticker loop out of `runWatch` verbatim; signature:

```go
func watchWithInterval(ctx context.Context, app *App, interval time.Duration, once bool) error
```

**Acceptance:**
- `runWatch` still handles the interval path by delegating.
- All existing watch tests remain green.

---

### Task 3.4 — Implement `watchWithCron` ~~DONE~~

**Files:** `cmd_watch.go`, `cmd_watch_test.go`

**Signature:**
```go
func watchWithCron(ctx context.Context, app *App, schedule string, once bool) error
```

**Actual behavior (differs from original plan):**
1. `watchWithCron` calls `cron.ParseStandard(schedule)` to produce a `cron.Schedule` value.
2. Delegates to `watchWithCronSchedule(ctx, runner, sched, scheduleExpr, once)`.
3. `watchWithCronSchedule` drives ticks with a `time.Timer` loop:
   - Computes `next := sched.Next(time.Now())`, sets `time.NewTimer(time.Until(next))`.
   - On each tick: calls `runSyncOnce(ctx, runner, true)`, resets timer to next tick.
   - `<-ctx.Done()` → return nil.
4. `cron.New` / `c.Start` / `c.Stop` are **not** used — only `cron.ParseStandard` for schedule parsing.

**Testability seam:** `watchWithCronSchedule` accepts a `cronSchedule` interface
(`Next(time.Time) time.Time`), so tests inject a `mockCronSchedule` with
pre-computed fire times instead of waiting real minutes.

**Tests:**
- `TestWatchWithCron_InvalidScheduleReturnsError`
- `TestWatchWithCronSchedule_ContextCancelStopsImmediately`
- `TestWatchWithCronSchedule_OnceTriggersImmediateSync`
- `TestWatchWithCronSchedule_FiresOnSchedule`

**Acceptance:**
- New tests green without sleeping for >2s total.
- `make lint` clean (`funlen` / `gocyclo` respect limits).

---

### Task 3.5 — Wire mode dispatch in `runWatch` ~~DONE~~

**Files:** `cmd_watch.go`

Final shape of `runWatch`:

```go
func runWatch(ctx context.Context, cmd *cli.Command) error {
    ctx, app, cfg, err := buildWatchApp(ctx, cmd)
    if err != nil { return err }

    resolved, err := resolveWatchConfig(cmd, cfg.Watch)
    if err != nil { return err }

    once := cmd.Bool("once")
    if resolved.Schedule != "" {
        return watchWithCron(ctx, app, resolved.Schedule, once)
    }
    interval, _ := resolved.GetInterval() // already validated
    return watchWithInterval(ctx, app, interval, once)
}
```

**Tests:**
- Existing watch tests must stay green.
- `TestRunWatch_BothFlagsSet_ReturnsMutualExclusionError` — both CLI flags set.
- `TestRunWatch_CLIIntervalAndYAMLSchedule_ReturnsError` — CLI interval + YAML schedule.
- `TestRunWatch_NeitherSet_ReturnsMissingError` — neither set, expects `ErrWatchConfigMissing`.

**Acceptance:**
- All tests green.
- Manual smoke: `./anilist-mal-sync watch --schedule "*/2 * * * *"` starts and
  logs next fire time.

---

## Phase 4 — Documentation & examples

### Task 4.1 — Update `config.example.yaml` ~~DONE~~

Add under the `watch:` section:
```yaml
watch:
  interval: "24h"           # existing
  schedule: ""              # Cron expression (5 fields). Mutually exclusive with interval.
                            # Examples:
                            #   "0 3 * * *"    — every day at 03:00
                            #   "*/30 * * * *" — every 30 minutes
                            #   "0 */6 * * *"  — every 6 hours
```

**Acceptance:** file parses; `make test` green (config parse tests should load
the example).

---

### Task 4.2 — Update `docker-compose.example.yaml` ~~DONE~~

Add a commented `WATCH_SCHEDULE=` env alongside `WATCH_INTERVAL=`, with a note
that only one may be set.

**Acceptance:** file lints / composes without errors.

---

### Task 4.3 — Update `README.md` ~~DONE~~

1. In the "Watch mode" section (around line 203 / line 350):
   - Document `--schedule` / `WATCH_SCHEDULE`.
   - Priority table (CLI > ENV > YAML).
   - Mutual-exclusion rule (fails hard).
   - Example docker-compose snippet using `WATCH_SCHEDULE`.
2. Add cron syntax pointer (link to `robfig/cron/v3` CRON_TZ section is not
   relevant — we fix `time.Local`; just link to the crontab.guru or the
   library's README for syntax).

**Acceptance:** README renders; no broken links.

---

### Task 4.4 — Update `CLAUDE.md` ~~DONE~~

- In the component table, note `cmd_watch.go` handles both interval and cron.
- In the "Sync Flow" / environment variable list, mention `WATCH_SCHEDULE`.

**Acceptance:** file updated in the same PR.

---

## Phase 5 — Quality gates & release

### Task 5.1 — Full quality gate ~~DONE~~

```bash
make fmt
make lint
make test
make check
```

All green. Investigate any `funlen` / `gocyclo` warnings on `runWatch` and
split further if needed.

### Task 5.2 — Manual smoke tests

Run locally (document findings in PR):

1. `anilist-mal-sync watch --schedule "*/1 * * * *"` → observe two scheduled
   syncs, ctrl-C terminates gracefully.
2. `anilist-mal-sync watch --schedule "*/1 * * * *" --once` → observe immediate
   sync, then scheduled.
3. `anilist-mal-sync watch --interval 1h --schedule "0 3 * * *"` → expect
   mutual-exclusion error.
4. `WATCH_SCHEDULE="0 3 * * *" anilist-mal-sync watch` → cron mode via env.
5. YAML with `watch.schedule: "0 3 * * *"`, no CLI, no env → cron mode.
6. YAML `schedule` + CLI `--interval 2h` → mutual-exclusion error.
7. `anilist-mal-sync watch --schedule "bad cron"` → validation error before
   starting.

### Task 5.3 — PR

- Title: `feat(watch): support cron expressions via --schedule / WATCH_SCHEDULE (#71)`.
- Body: summary + link to issue + priority table + mutual-exclusion rule +
  smoke-test checklist from 5.2.

---

## Test coverage matrix

| Area                              | Test(s)                                                         | Phase/Task |
|-----------------------------------|-----------------------------------------------------------------|------------|
| ENV → config (`WATCH_SCHEDULE`)   | `TestLoadConfigFromEnv_ReadsWatchSchedule`, `TestOverrideWatchFromEnv_SetsSchedule`, `TestOverrideWatchFromEnv_ScheduleEmptyEnvKeepsYAMLValue` | 1.2        |
| YAML → config                     | `TestLoadConfigFromFile_ParsesSchedule`                         | 1.2        |
| Validate: mutual exclusion        | `TestWatchConfig_Validate_BothSet_ReturnsError`                 | 1.3        |
| Validate: interval bounds         | `..._IntervalBelowMin`, `..._IntervalAboveMax`, `..._IntervalMalformed`, `..._IntervalOnly_Valid` | 1.3 |
| Validate: schedule syntax         | `..._ScheduleOnly_Valid`, `..._ScheduleMalformed`, `..._ScheduleFieldCount` | 1.3 |
| Validate: neither set             | `..._BothEmpty_ReturnsMissingSentinel`                          | 1.3        |
| ParseSchedule                     | `..._ValidReturnsSchedule`, `..._Invalid`, `..._Empty`          | 1.4        |
| CLI flag presence & shape         | Updated `TestWatchCommand_HasFlags`, `TestWatchCommand_ScheduleFlag_Exists`, `..._NoDefault` | 2.2 |
| Resolution precedence             | `TestResolveWatchConfig_CLIIntervalOverridesYAMLInterval`, `..._CLIScheduleOverridesYAMLSchedule`, `..._OnlyInterval_OK`, `..._OnlySchedule_OK` | 2.1 |
| Resolution conflict               | `..._CLIIntervalAndYAMLSchedule_ReturnsError`, `..._CLIScheduleAndYAMLInterval_ReturnsError` | 2.1 |
| Resolution missing                | `..._Neither_ReturnsMissingSentinel`                            | 2.1        |
| Cron runtime: invalid expr        | `TestWatchWithCron_InvalidScheduleReturnsError`                 | 3.4        |
| Cron runtime: cancellation        | `TestWatchWithCron_ContextCancelStopsImmediately`               | 3.4        |
| Cron runtime: --once              | `TestWatchWithCron_OnceTriggersImmediateSync`                   | 3.4        |
| Cron runtime: fires on schedule   | `TestWatchWithCron_FiresOnSchedule` (using `ConstantDelaySchedule` seam) | 3.4 |
| Run dispatch: both set            | `TestRunWatch_BothFlagsSet_ReturnsMutualExclusionError`, `TestRunWatch_CLIIntervalAndYAMLSchedule_ReturnsError` | 3.5 |
| Run dispatch: neither             | `TestRunWatch_NeitherSet_ReturnsMissingError`                   | 3.5        |
| Backward-compat interval mode     | all existing `cmd_watch_test.go` tests kept green               | 3.3, 3.5   |
| End-to-end manual                 | Scenarios 1–7 in 5.2                                            | 5.2        |

---

## Rollback plan

Feature is additive — if a critical issue surfaces post-merge:

1. Revert the PR; `interval`-mode watch continues unchanged.
2. Users who only set `watch.schedule` get the pre-existing
   "interval required" error — no data at risk, no migration needed.
