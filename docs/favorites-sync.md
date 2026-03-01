# Favorites Synchronization

This document describes the favorites synchronization feature between AniList and MyAnimeList.

## Overview

Favorites sync is an optional feature that synchronizes your favorited anime and manga between AniList and MyAnimeList. It runs as a separate phase after the main status/progress synchronization.

## API Limitations

The feature has different capabilities depending on direction due to API limitations:

| Direction | Read | Write | Behavior |
|-----------|------|-------|----------|
| MAL → AniList | ✅ via Jikan API | ✅ via ToggleFavourite mutation | Full sync (add missing favorites) |
| AniList → MAL | ✅ via isFavourite field | ❌ MAL API v2 has no favorites endpoint | Report only |

## Configuration

### Enable via YAML

Add to your `config.yaml`:

```yaml
favorites:
  enabled: true
```

### Enable via Environment Variable

```bash
export FAVORITES_SYNC_ENABLED=true
```

### Enable via CLI Flag

```bash
anilist-mal-sync sync --favorites
```

The `--favorites` flag automatically enables Jikan API (required for reading MAL favorites).

## Behavior by Direction

### MAL → AniList (with `--reverse-direction`)

- Reads your MAL favorites via Jikan API (public user profile)
- Compares with your AniList list entries
- **Adds** missing favorites on AniList
- **Does not remove** favorites that exist only on AniList (you may have intentionally favorited different items)

Example:
```bash
anilist-mal-sync sync --favorites --reverse-direction
```

Output:
```
★ [Favorites] Added "Cowboy Bebop" to AniList favorites
★ [Favorites] Added "Monster" to AniList favorites
★ Favorites sync complete: +2 added on AniList (15 skipped)
```

### AniList → MAL (default direction)

- Reads your AniList favorites from list entries (via `isFavourite` field)
- Reads your MAL favorites via Jikan API
- Reports differences (cannot write to MAL)

Example:
```bash
anilist-mal-sync sync --favorites
```

Output:
```
★ [Favorites] anime "Cowboy Bebop" is only on AniList
★ [Favorites] manga "Berserk" is only on MAL
★ Favorites: 2 mismatches (AniList→MAL, report only)
```

## Important Notes

### MAL Profile Privacy

Jikan API reads public MAL profiles. If your MAL profile is set to private, favorites sync will fail with a warning:

```
Failed to fetch MAL favorites: user username not found or profile is private (skipping favorites sync)
```

**Solution**: Set your MAL profile to public at https://myanimelist.net/editprofile.php?go=privacy

### Rate Limiting

AniList ToggleFavourite mutations are rate-limited (~90 requests/minute). The tool adds a 700ms delay between favorite toggles to respect this limit.

### Entries Without MAL ID

Anime/manga entries on your AniList list that don't have a corresponding MAL ID are skipped with a debug log:

```
★ [Favorites] Skipping "Some Anime" (no MAL ID)
```

### AniList Favorites Not In Your List

The current implementation only compares favorites for entries that exist on **both** your AniList and MAL lists. If you've favorited an anime on AniList but haven't added it to your list, it won't be included in the sync.

## Examples

### Dry Run

Preview what would change without actually modifying favorites:

```bash
anilist-mal-sync sync --favorites --dry-run
```

Output:
```
★ [Favorites] Would add "Cowboy Bebop" (MAL ID 1, AniList ID 1) to AniList favorites
```

### Sync Anime Only

```bash
anilist-mal-sync sync --favorites --reverse-direction
# Only syncs anime (--manga not specified)
```

### Sync Both Anime and Manga

```bash
anilist-mal-sync sync --favorites --all --reverse-direction
```

### Verbose Output

See detailed logs about skipped entries and ID matching:

```bash
anilist-mal-sync sync --favorites --verbose
```

## Configuration Options Summary

| Option | CLI Flag | Env Var | YAML | Default |
|--------|----------|---------|------|---------|
| Enable favorites | `--favorites` | `FAVORITES_SYNC_ENABLED` | `favorites.enabled` | `false` |
| Jikan API (auto-enabled with `--favorites`) | `--jikan-api` | `JIKAN_API_ENABLED` | `jikan_api.enabled` | `false` |

## Troubleshooting

### "Failed to fetch MAL favorites"

**Cause**: MAL profile is private or username is incorrect.

**Solution**:
1. Verify `MAL_USERNAME` in your config
2. Set your MAL profile to public

### "No favorites added but I know I have favorites"

**Possible causes**:
1. Entries don't have MAL IDs set on AniList
2. AniList and MAL favorites are already in sync
3. Using wrong direction (default is AniList→MAL which is report-only)

Run with `--verbose` to see detailed logs:
```bash
anilist-mal-sync sync --favorites --reverse-direction --verbose
```

### Jikan API Rate Limiting

If you see rate limit errors from Jikan, the tool will automatically retry with exponential backoff. For large favorites lists, consider running sync less frequently.
