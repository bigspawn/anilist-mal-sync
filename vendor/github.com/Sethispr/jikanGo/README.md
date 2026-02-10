# Jikan Go

[![Go Report Card](https://goreportcard.com/badge/github.com/Sethispr/jikanGo)](https://goreportcard.com/report/github.com/Sethispr/jikanGo)

A Go client for the [Jikan API](https://jikan.moe) that retries on 429s, caches responses, and actually validates IDs before burning a network request. Saves you from writing the same HTTP wrapper again.

## Install

```bash
go get github.com/Sethispr/jikanGo
```

### Quick Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Sethispr/jikanGo"
)

func main() {
    client := jikan.New()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    char, err := client.Character.ByID(ctx, 1) // Spike Spiegel's ID is 1
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("%s (%s)\n", char.Name, char.NameKanji)
    fmt.Printf("Known as: %v\n", char.Nicknames)
}
```

## Features

- **Complete API**: Characters, anime, manga, clubs, genres, people, seasons, producers, etc
- **Strongly Typed**: Access `char.NameKanji`, `char.Nicknames`, etc directly, no type assertions
- **Retries**: Handles 429s and other errors with exponential backoff
- **Caching**: Optional response caching with TTL
- **Rate Limiting**: Optional built in request throttling
- **Pagination**: Returns `*Pagination` with `LastPage` and `HasNext`
- **Context Support**: All methods accept `context.Context` for timeouts

## Rate Limiting

Jikan allows [3 requests per second](https://docs.api.jikan.moe/). Turn on the built in limiter to stay under this limit automatically:

```go
// Default: 3 requests per second (Jikan's limit)
client := jikan.New(jikan.WithRateLimit(3))
```

Customize the rate:
```go
// 5 requests per second (only if you have special access)
client := jikan.New(jikan.WithRateLimit(5))
```

Advanced configuration (custom burst size):
```go
import "golang.org/x/time/rate"

// Allow bursts of 5, then limit to 3 per second
lim := rate.NewLimiter(rate.Every(333*time.Millisecond), 5)
client := jikan.New(jikan.WithRateLimiter(lim))
```

The rate limiter respects context cancellation. If your context times out while waiting for the rate limiter, it returns immediately with the context error.

## Caching

Avoid hitting the API twice for the same data. The client accepts any cache implementing the `Cache` interface.

In memory cache (included):
```go
cache := jikan.NewMemoryCache()
defer cache.Stop()

client := jikan.New(
    jikan.WithCache(cache, 5*time.Minute),
)
```

Bring your own (Redis, etc):
```go
type RedisCache struct { client *redis.Client }
func (r *RedisCache) Get(ctx context.Context, key string, dst interface{}) error { ... }
func (r *RedisCache) Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error { ... }
func (r *RedisCache) Delete(ctx context.Context, key string) error { ... }

client := jikan.New(jikan.WithCache(&RedisCache{redisClient}, time.Hour))
```

Skip cache for specific requests:
```go
ctx := jikan.NoCache(context.Background())
char, _ := client.Character.ByID(ctx, 1) // Forces API call
```

---

## Examples

Get combined data (one call instead of four):
```go
full, err := client.Character.Full(ctx, 1)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Appears in %d anime and %d manga\n", len(full.Anime), len(full.Manga))
```

Search with pagination:
```go
results, pg, err := client.Character.Search(ctx, "Spike", 1)
if err != nil {
    log.Fatal(err)
}
if pg.HasNext {
    fmt.Println("More pages available")
}
```

Filter genres safely:
```go
themes, _, err := client.Genre.Anime(ctx, jikan.GenreThemes, 1, 25)
if err != nil {
    log.Fatal(err)
}
for _, t := range themes {
    fmt.Println(t.Name)
}
```

See `examples/` folder for working CLIs.
