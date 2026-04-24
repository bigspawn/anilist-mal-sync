# Fixes: outstanding gaps in issue-71 implementation

Follow-up task list for discrepancies found during the review of branch
`feat/issue-71-cron-schedule`. Each task is self-contained, sequential, and
ends with a green `make fmt && make lint && make test`.

Tasks are grouped by severity. Do the 🔴 ones first — they fix actual code
gaps. 🟡 are cleanup, 🟢 are docs/plan alignment.

---

## 🔴 F1 — Extract `buildWatchApp` helper ~~DONE~~

**Why:** Task 3.1 of `issue-71-tasks.md` is marked DONE, but `runWatch`
(`cmd_watch.go:72-105`) still inlines config load + logger init + `NewApp`.
The planned "thin dispatcher" shape never happened.

**Files:** `cmd_watch.go`

**Steps:**
1. Create:
   ```go
   func buildWatchApp(ctx context.Context, cmd *cli.Command) (context.Context, *App, Config, error)
   ```
2. Move into it: `getSyncFlagsFromCmd`, `loadConfigFromFile`,
   `applySyncFlagsToConfig`, `NewLogger`, `logger.WithContext`, `NewApp`.
3. `runWatch` becomes: call `buildWatchApp` → `resolveWatchConfig` → dispatch.

**Tests:** covered indirectly by the existing watch tests (interval path must
keep working). No new unit tests — the helper touches OAuth/network via
`NewApp`, not easily unit-testable in isolation.

**Acceptance:**
- `runWatch` body shrinks to ≤20 lines.
- `make test` green.

---

## 🔴 F2 — Extract `runSyncOnce` helper and deduplicate sync bodies ~~DONE~~

**Why:** Task 3.2 is marked DONE, but the helper does not exist.
`--once` logic and "scheduled tick" logic are duplicated across
`watchWithInterval` (`cmd_watch.go:108-140`) and `watchWithCronSchedule`
(`cmd_watch.go:152-184`).

**Files:** `cmd_watch.go`

**Steps:**
1. Add:
   ```go
   func runSyncOnce(ctx context.Context, runner syncRunner, label string) error
   ```
   Body: `runner.Refresh(ctx)` + `runner.Run(ctx)` + `log.Printf("%s ...", label)`.
   Decide: the `--once` initial call historically does **not** call `Refresh`;
   scheduled ticks do. Either pass a `refresh bool` parameter, or keep two
   helpers (`runInitialSync` / `runScheduledSync`). Prefer a single function
   with a `refresh bool` to stay minimal.
2. Replace both duplicated blocks in `watchWithInterval` and
   `watchWithCronSchedule` with calls to the helper.
3. Keep error-return semantics identical (initial sync returns error; scheduled
   tick logs and continues).

**Tests (new):**
- `TestRunSyncOnce_CallsRunWithoutRefresh` — verify Refresh is not called when
  `refresh=false`; Run is called once.
- `TestRunSyncOnce_CallsRefreshAndRun` — both are called when `refresh=true`.
- `TestRunSyncOnce_ReturnsRunError` — propagates Run's error.

Existing `TestWatchWithCronSchedule_*` and interval tests must stay green.

**Acceptance:**
- No duplicated `Refresh`/`Run`/log block remains.
- New tests green.

---

## 🔴 F3 — Add missing dispatch-level tests ~~DONE~~

**Why:** Task 3.5 listed `TestRunWatch_BothConfigured_ReturnsError` and
`TestRunWatch_NeitherConfigured_ReturnsHelpError`. Neither exists. Current
coverage stops at `resolveWatchConfig` / `Validate`; the full `runWatch` path
with real config loading is untested.

**Files:** `cmd_watch_test.go`

**Constraint:** `runWatch` calls `NewApp` which performs OAuth. To test the
dispatch error-before-app-build paths we need the validation/resolution to
fail **before** `NewApp`. Currently `resolveWatchConfig` runs **after**
`NewApp` (see `cmd_watch.go:86-94`) — so a mutual-exclusion error still costs
an OAuth flow in tests.

**Steps:**
1. Move `resolveWatchConfig` call **before** `NewApp` in `runWatch` (cheap fix,
   and arguably more correct — fail fast on bad flags without touching
   credentials). Do this as part of F1 if F1 is done in the same PR.
2. Add tests:
   - `TestRunWatch_BothFlagsSet_ReturnsMutualExclusionError` — build a
     `cli.Command` with both `--interval 2h` and `--schedule "0 3 * * *"`,
     point `--config` at a minimal valid YAML, expect the mutual-exclusion
     error string. Confirm no OAuth attempt (assert by observing that no
     token file was read — use a temp `TOKEN_FILE_PATH`).
   - `TestRunWatch_CLIIntervalAndYAMLSchedule_ReturnsError` — only CLI
     `--interval`, YAML has `watch.schedule` — same result.
   - `TestRunWatch_NeitherSet_ReturnsMissingError` — no flags, YAML without
     `watch.interval` or `watch.schedule` — expect
     `errors.Is(err, ErrWatchConfigMissing)`.

**Acceptance:**
- New tests green without network.
- `resolveWatchConfig` runs before any auth / `NewApp` call.

---

## 🟡 F4 — Remove dead `validateInterval` test helper ~~DONE~~

**Why:** `cmd_watch_test.go:190-198` defines `validateInterval` that
duplicates the bounds check now living in `WatchConfig.Validate()`. The plan
explicitly asked for one source of truth.

**Files:** `cmd_watch_test.go`

**Steps:**
1. Rewrite `TestValidateInterval_ValidIntervals` to call
   `WatchConfig{Interval: d.String()}.Validate()` instead of `validateInterval`.
   Rename to `TestWatchConfig_Validate_IntervalTable` (table-driven, one place
   for all bounds cases).
2. Delete the `validateInterval` helper.
3. Collapse the individual `TestWatchConfig_Validate_Interval*` tests from
   `config_test.go:1206-1237` into the same table if convenient (optional,
   only if it simplifies — don't lose coverage).

**Acceptance:** No function `validateInterval` remains; bounds check covered
by one table-driven test.

---

## 🟡 F5 — Remove unused `defaultInterval` constant ~~DONE~~

**Why:** `cmd_watch.go:12` — `defaultInterval = 24 * time.Hour` is unused in
production; only `TestValidateInterval_ValidIntervals` references it. Once F4
lands, nothing will reference it.

**Files:** `cmd_watch.go`, `cmd_watch_test.go`

**Steps:**
1. Confirm no production reference: `rg "defaultInterval" --type go`.
2. Delete the constant.
3. If any test still needed a literal, use `24 * time.Hour` inline or a
   test-local const.

**Acceptance:** `rg defaultInterval` returns zero matches.

---

## 🟡 F6 — Decide on the stray vendor diff ~~DONE~~

**Why:** `git status` shows ~23 modified files in `vendor/gopkg.in/yaml.v{2,3}`,
`vendor/github.com/davecgh/go-spew`, `vendor/github.com/pmezard/go-difflib`
— comment reformatting (`////` → `// //`) produced by `go mod vendor` on a
different Go toolchain than whatever produced the original vendor tree.
Unrelated to issue #71.

**Options (pick one):**

**A — revert** (recommended if CI still passes):
```bash
git checkout -- vendor/gopkg.in vendor/github.com/davecgh vendor/github.com/pmezard
```
Then rerun `make test` to confirm nothing in the yaml/diff paths actually
needed the new formatting.

**B — commit as a separate chore:** if the local toolchain is the one we want
going forward, make a dedicated commit
`chore(vendor): revendor stdlib-mirror packages with current toolchain`
and land it before the feature PR, so the feature diff stays small.

**Acceptance:** `git status vendor/` is clean (either reverted or committed).
The feature PR's vendor diff contains **only** `vendor/github.com/robfig/cron/v3/`.

---

## 🟢 F7 — Finish `config.example.yaml` comment ~~DONE~~

**Why:** Plan (Task 4.1) asked for three cron examples in the comment.
Current file (`config.example.yaml:20`) has only one short comment.

**Files:** `config.example.yaml`

**Steps:** replace the single-line comment with:
```yaml
watch:
  interval: "24h"   # Sync interval for watch mode (1h-168h). Mutually exclusive with schedule.
  schedule: ""      # Cron expression (5 fields). Mutually exclusive with interval.
                    # Examples:
                    #   "0 3 * * *"    — every day at 03:00
                    #   "*/30 * * * *" — every 30 minutes
                    #   "0 */6 * * *"  — every 6 hours
```

**Acceptance:** file still parses (the existing `TestLoadConfigFromFile_*`
tests cover this indirectly; add a quick `TestConfigExample_Parses` if one
doesn't exist).

---

## 🟢 F8 — Remove phantom test from plan matrix ~~DONE~~

**Why:** Plan lists `TestResolveWatchConfig_CLIIntervalClearsYAMLSchedule_ReturnsError`
as a distinct case, but it's semantically identical to
`TestResolveWatchConfig_CLIIntervalAndYAMLSchedule_ReturnsError` (which
exists). The matrix overstates coverage.

**Files:** `docs/plans/issue-71-tasks.md`

**Steps:**
1. Remove the duplicate bullet from Task 2.1.
2. Remove the duplicate row from the coverage matrix.

**Acceptance:** plan matches reality; no test name appears in the plan
without a matching function in the codebase.

---

## 🟢 F9 — Align plan description with actual cron runtime ~~DONE~~

**Why:** Plan (Task 3.4) described `cron.New(cron.WithLocation(time.Local))`
+ `AddFunc` + `c.Start` + drain via `c.Stop`. Actual implementation
(`cmd_watch.go:143-184`) uses `cron.ParseStandard` + a manual `time.Timer`
loop. Both work; plan is stale.

**Files:** `docs/plans/issue-71-tasks.md`, `docs/plans/issue-71-cron-schedule.md`

**Steps:**
1. Update Task 3.4 description to match: we only use
   `cron.ParseStandard(...)` to produce a `cron.Schedule`, then drive ticks
   with `time.Timer` against `sched.Next(time.Now())`.
2. Note the testability seam (`cronSchedule` interface, `watchWithCronSchedule`
   with an injectable schedule) that the real code uses.
3. In the main plan doc, drop the `cron.New` / `c.Start` reference.

**Acceptance:** someone reading only the plan can reproduce the current code
structure without confusion.

---

## 🟢 F10 — Document timezone choice explicitly ~~DONE~~

**Why:** We committed to `time.Local`, but the current code relies on the
parser/`time.Now()` defaults rather than setting a location explicitly.
There's no test asserting the contract.

**Files:** `cmd_watch.go` (optional), `cmd_watch_test.go`, `README.md`

**Steps:**
1. README: in the watch section, add "Cron expressions are evaluated in the
   container / host local timezone (`time.Local`). Set `TZ` env var on the
   container to change it."
2. (Optional) Test: `TestWatchWithCron_UsesLocalTimezone` — parse
   `"0 12 * * *"`, compute `sched.Next(time.Now())`, assert that the returned
   time's `Location()` is `time.Local` (or equivalent via
   `.Location().String() == time.Local.String()`).

**Acceptance:**
- README has the timezone note.
- If test added, it passes.

---

## 🟢 F11 — Update task checkboxes in `issue-71-tasks.md` ~~DONE~~

**Why:** Tasks 3.1, 3.2 are marked `~~DONE~~` but F1 and F2 above prove they
aren't done. Task 3.5 is marked DONE but is missing two tests (F3).

**Files:** `docs/plans/issue-71-tasks.md`

**Steps:**
1. Flip 3.1, 3.2, 3.5 back to open (remove `~~DONE~~`).
2. Once F1/F2/F3 land, flip them to DONE again.
3. Add a pointer at the top of the file:
   `> See [issue-71-fixes.md](./issue-71-fixes.md) for post-review corrections.`

**Acceptance:** plan reflects reality.

---

## Execution order

Do them in this order to minimise rework:

1. **F6** (vendor diff) — isolated, fixes the dirty tree.
2. **F1** → **F2** → **F3** — the real code fixes, together, one PR.
3. **F4** → **F5** — cleanup, can ride with F1–F3.
4. **F7** — docs, same PR.
5. **F8** → **F9** → **F10** → **F11** — plan hygiene, same PR.

One PR title suggestion:
```
fix(watch): address review findings from issue-71 implementation
```

## Quality gates (run after each phase)

```bash
make fmt
make lint
make test
```

Final before opening the PR:
```bash
make check
git diff --stat main..HEAD   # verify no out-of-scope files
```

---

## Summary table

| ID  | Severity | Kind          | Touches                                      |
|-----|----------|---------------|----------------------------------------------|
| F1  | 🔴       | Code          | `cmd_watch.go`                               |
| F2  | 🔴       | Code + tests  | `cmd_watch.go`, `cmd_watch_test.go`          |
| F3  | 🔴       | Tests         | `cmd_watch.go` (reorder), `cmd_watch_test.go`|
| F4  | 🟡       | Tests cleanup | `cmd_watch_test.go`                          |
| F5  | 🟡       | Code cleanup  | `cmd_watch.go`, `cmd_watch_test.go`          |
| F6  | 🟡       | Infra         | `vendor/`                                    |
| F7  | 🟢       | Docs          | `config.example.yaml`                        |
| F8  | 🟢       | Plan          | `docs/plans/issue-71-tasks.md`               |
| F9  | 🟢       | Plan          | `docs/plans/issue-71-*.md`                   |
| F10 | 🟢       | Docs + test   | `README.md`, optional test                   |
| F11 | 🟢       | Plan          | `docs/plans/issue-71-tasks.md`               |
