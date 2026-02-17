# Date Synchronization

This document describes how start/end dates are synchronized between AniList and MyAnimeList.

## Overview

Both AniList and MAL track two dates per entry:
- **Start date** — when the user started watching/reading
- **End date** — when the user finished watching/reading

Dates are synced in both directions:
- **Forward** (AniList -> MAL): dates flow via `GetUpdateOptions()` using `mal.StartDate` / `mal.FinishDate`
- **Reverse** (MAL -> AniList): dates flow via GraphQL `SaveMediaListEntry` mutation using `FuzzyDateInput`

## Date Handling Rules

| Source date | Target date | Action                          |
|-------------|-------------|---------------------------------|
| nil         | nil         | Nothing to do                   |
| nil         | set         | Skip — don't clear target date  |
| set         | nil         | Sync date to target             |
| same        | same        | No update needed                |
| different   | different   | Update target with source date  |

Key principle: **nil source dates are never sent** to avoid clearing manually-set dates on the target service.

## Conditional End Date

The end date (`FinishedAt` / `completedAt`) is **only sent when status is COMPLETED**. This prevents prematurely setting a finish date for entries that are still in progress.

| Status      | startedAt sent? | completedAt sent? |
|-------------|-----------------|-------------------|
| WATCHING    | yes (if set)    | no                |
| COMPLETED   | yes (if set)    | yes (if set)      |
| ON_HOLD     | yes (if set)    | no                |
| DROPPED     | yes (if set)    | no                |
| PLAN_TO_*   | yes (if set)    | no                |

## Date Formats

### AniList GraphQL — FuzzyDateInput

```graphql
input FuzzyDateInput {
  year: Int
  month: Int
  day: Int
}
```

Dates are sent as a map: `{"year": 2023, "month": 6, "day": 15}`. When a date is nil, the variable is omitted from the mutation entirely (not sent as null).

### MAL REST API

Dates are formatted as `YYYY-MM-DD` strings (e.g., `"2023-06-15"`).

### Internal Representation

Internally, dates are stored as `*time.Time` in UTC with day-level precision. The `sameDates()` function compares only year, month, and day — ignoring time of day.

## Comparison Logic

The `sameDates(source, target)` function determines if dates need syncing:

```
sameDates(nil, nil)   = true   // no dates, nothing to do
sameDates(nil, set)   = true   // don't clear existing target date
sameDates(set, nil)   = false  // need to sync date to target
sameDates(same, same) = true   // already in sync
sameDates(diff, diff) = false  // need to update target
```

Date comparison is integrated into `SameProgressWithTarget()` for both Anime and Manga, so date differences trigger sync updates automatically.

## Edge Cases

- **Partial dates**: AniList may return dates with only year/month set (day=nil). The `convertFuzzyDateToTimeOrNow` function returns nil if any component is missing.
- **Timezone**: All dates are stored in UTC. Day-level comparison ignores time of day.
- **First sync**: Entries with date differences will be updated on the first run after enabling date sync.
