Local testing
=================

Build locally:

```bash
go build -o anilist-mal-sync ./...
```

Run in dry-run mode (no updates):

```bash
./anilist-mal-sync -config config.example.yaml -dry-run
```

Perform OAuth flow (interactive):
- Run the binary and follow printed auth URLs for AniList and MyAnimeList.
- Ensure `token_file_path` in config points to a writable path.

Running tests / mocked APIs:
- This repo currently has no unit tests. For CI-style integration tests, run the binary in `-dry-run` and mock upstream APIs with a local proxy.
