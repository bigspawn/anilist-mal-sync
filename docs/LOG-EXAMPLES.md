# Example Log Output

This document shows what the log output looks like with the improved statistics and ID-based matching.

## Example 1: Reverse Sync (MAL → AniList) - Mixed Results

```
2026/01/03 06:27:36 [MAL to AniList Anime] Fetching MAL...
2026/01/03 06:27:40 [MAL to AniList Anime] Fetching AniList...
2026/01/03 06:27:45 [MAL to AniList Anime] Got 52 from MAL
2026/01/03 06:27:45 [MAL to AniList Anime] Got 54 from AniList
2026/01/03 06:27:45 [MAL to AniList Anime] Processing for status: completed
2026/01/03 06:27:45 [MAL to AniList Anime] Searching AniList by MAL ID: 1234
2026/01/03 06:27:45 [MAL to AniList Anime] Found target by MAL ID 1234 and title match: Attack on Titan
2026/01/03 06:27:45 [MAL to AniList Anime] Src title: Attack on Titan
2026/01/03 06:27:45 [MAL to AniList Anime] Tgt title: Attack on Titan
2026/01/03 06:27:45 [MAL to AniList Anime] Progress is not same, need to update: Diff{Status: watching != completed, Score: 8.5 != 9.0, Progress: 25 != 75}
2026/01/03 06:27:45 [MAL to AniList Anime] Updated Attack on Titan
2026/01/03 06:27:45 [MAL to AniList Anime] Searching AniList by MAL ID: 5678
2026/01/03 06:27:45 [MAL to AniList Anime] Found target by MAL ID 5678 (single result): One Piece
2026/01/03 06:27:45 [MAL to AniList Anime] Src title: One Piece
2026/01/03 06:27:45 [MAL to AniList Anime] Tgt title: One Piece
2026/01/03 06:27:45 [MAL to AniList Anime] Updated One Piece
2026/01/03 06:27:45 [MAL to AniList Anime] Searching AniList by MAL ID: 9999
2026/01/03 06:27:45 [MAL to AniList Anime] Error searching by MAL ID 9999: no results found, falling back to name search
2026/01/03 06:27:45 [MAL to AniList Anime] No ID available, searching by name: Capeta
2026/01/03 06:27:45 [MAL to AniList Anime] No targets found by name: Capeta
2026/01/03 06:27:45 [MAL to AniList Anime] Error finding target: no target found for source: Capeta
2026/01/03 06:27:45 [MAL to AniList Anime] Processing for status: plan_to_watch
2026/01/03 06:27:45 [MAL to AniList Anime] Processing for status: watching
2026/01/03 06:27:45 [MAL to AniList Anime] Updated 2 out of 52
2026/01/03 06:27:45 [MAL to AniList Anime] Skipped 48 (already in sync)
2026/01/03 06:27:45 [MAL to AniList Anime] Not found 2 (could not match in target service)
```

## Example 2: Forward Sync (AniList → MAL) - All Found

```
2026/01/03 06:30:00 [AniList to MAL Anime] Fetching AniList...
2026/01/03 06:30:05 [AniList to MAL Anime] Fetching MAL...
2026/01/03 06:30:10 [AniList to MAL Anime] Got 54 from AniList
2026/01/03 06:30:10 [AniList to MAL Anime] Got 52 from MAL
2026/01/03 06:30:10 [AniList to MAL Anime] Processing for status: completed
2026/01/03 06:30:10 [AniList to MAL Anime] Finding target by API ID: 1234
2026/01/03 06:30:10 [AniList to MAL Anime] Found existing user target by ID: 1234
2026/01/03 06:30:10 [AniList to MAL Anime] Src title: Attack on Titan
2026/01/03 06:30:10 [AniList to MAL Anime] Tgt title: Attack on Titan
2026/01/03 06:30:10 [AniList to MAL Anime] Progress is not same, need to update: Diff{Status: completed != watching, Score: 9.0 != 8.5, Progress: 75 != 25}
2026/01/03 06:30:10 [AniList to MAL Anime] Updated Attack on Titan
2026/01/03 06:30:10 [AniList to MAL Anime] Finding target by API ID: 5678
2026/01/03 06:30:10 [AniList to MAL Anime] Found target by API ID (not in user list): 5678
2026/01/03 06:30:10 [AniList to MAL Anime] Updated One Piece
2026/01/03 06:30:10 [AniList to MAL Anime] Processing for status: plan_to_watch
2026/01/03 06:30:10 [AniList to MAL Anime] Processing for status: watching
2026/01/03 06:30:10 [AniList to MAL Anime] Updated 2 out of 54
2026/01/03 06:30:10 [AniList to MAL Anime] Skipped 52 (already in sync)
```

## Example 3: Bidirectional Sync

```
2026/01/03 06:35:00 [AniList to MAL Anime] Fetching AniList...
2026/01/03 06:35:05 [AniList to MAL Anime] Fetching MAL...
2026/01/03 06:35:10 [AniList to MAL Anime] Got 54 from AniList
2026/01/03 06:35:10 [AniList to MAL Anime] Got 52 from MAL
2026/01/03 06:35:10 [AniList to MAL Anime] Processing for status: completed
2026/01/03 06:35:10 [AniList to MAL Anime] Updated 5 out of 54
2026/01/03 06:35:10 [AniList to MAL Anime] Skipped 49 (already in sync)
2026/01/03 06:35:15 [MAL to AniList Anime] Fetching MAL...
2026/01/03 06:35:20 [MAL to AniList Anime] Fetching AniList...
2026/01/03 06:35:25 [MAL to AniList Anime] Got 52 from MAL
2026/01/03 06:35:25 [MAL to AniList Anime] Got 54 from AniList
2026/01/03 06:35:25 [MAL to AniList Anime] Processing for status: completed
2026/01/03 06:35:25 [MAL to AniList Anime] Searching AniList by MAL ID: 1234
2026/01/03 06:35:25 [MAL to AniList Anime] Found target by MAL ID 1234 and title match: Attack on Titan
2026/01/03 06:35:25 [MAL to AniList Anime] Updated 3 out of 52
2026/01/03 06:35:25 [MAL to AniList Anime] Skipped 47 (already in sync)
2026/01/03 06:35:25 [MAL to AniList Anime] Not found 2 (could not match in target service)
```

## Key Differences from Before

### Before:
- All failures were counted as "Skipped", making it unclear if items were:
  - Already in sync (no update needed) ✅
  - Not found (couldn't match) ❌

### After:
- **Updated X out of Y**: Items that were successfully updated
- **Skipped X (already in sync)**: Items found but already matching (no update needed)
- **Not found X (could not match in target service)**: Items that couldn't be matched/found

This makes it much clearer what happened during sync!
