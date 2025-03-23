# Verniy

[![Go Report Card](https://goreportcard.com/badge/github.com/rl404/verniy)](https://goreportcard.com/report/github.com/rl404/verniy)
![License: MIT](https://img.shields.io/github/license/rl404/verniy.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/rl404/verniy.svg)](https://pkg.go.dev/github.com/rl404/verniy)

_Verniy_ is an unofficial [Anilist](https://anilist.co) GraphQL API library using [Go](https://golang.org/). Verniy will generate graphql query string
according to your request, make request to Anilist, parse reponse body,
and convert to a struct.

The goal of this library is to make a flexible and easy to call Anilist API.
This library may not cover all available API but you are always welcome to
make a pull request.

This library is built using [Anilist official docs](https://github.com/AniList/ApiV2-GraphQL-Docs) especially the graphql [reference](https://anilist.github.io/ApiV2-GraphQL-Docs/).

## Features

- Get anime data (details, characters, staff, stats)
- Get manga data (details, characters, staff, stats)
- Get character data (details, anime roles, manga roles)
- Get voice actor/mangaka/staff data (details, character roles, anime roles, manga roles)
- Get studio list and their produced anime
- Get genre list
- Get tag list
- Get user data (details, favorite, anime list, manga list)
- Get user data with OAuth2 Access Token for private profiles
- Search anime
- Search manga
- Search character
- Search voice actor/mangaka/staff
- Built-in request rate limiting

## Installation

```
go get github.com/rl404/verniy
```

## Quick Start

```go
package main

import (
	"github.com/rl404/verniy"
)

func main() {
    v := verniy.New()

    // Get anime One Piece data (with default fields).
    data, err := v.GetAnime(21)

    // Get anime One Piece data (with custom fields).
    data, err := v.GetAnime(21,
        verniy.MediaFieldID,
        verniy.MediaFieldTitle(
            verniy.MediaTitleFieldRomaji,
            verniy.MediaTitleFieldEnglish,
            verniy.MediaTitleFieldNative),
        verniy.MediaFieldType,
        verniy.MediaFieldFormat,
        verniy.MediaFieldStatusV2,
        verniy.MediaFieldDescription,
        verniy.MediaFieldStartDate,
        verniy.MediaFieldEndDate,
        verniy.MediaFieldSeason,
        verniy.MediaFieldSeasonYear)
}
```

*For more usage, please go to the [documentation](https://pkg.go.dev/github.com/rl404/verniy) or `example` folder.*

### Functions

There are alot of functions in verniy package but
most of the time, you just need functions in `Client` struct.
And it's recommended to use them to make your life easier.
If you want to make a custom request, you can use function
`MakeRequest()` (go to the function for more details).

### Parameters

Yes, there are tons of custom types and functions but let
your IDE auto-complete helps you choose the params for the
functions. Not only constant value but also function (like
`MediaFieldTitle` in the example) to fill the params.
But, most of the functions have variadic parameters,
so you don't need to actually fill it. There is already default
value and you can read the code yourself to see the default
value of the variadic parameters.

### Response

Most of the functions in `Client` struct return a struct and error.
Yes, most of the fields in the returned struct are pointers because,
like any graphql, Anilist will not return fields that we didn't
request. It is recommended to make your own pointer handler when
using the field. For example.

```go
// Handle *string to string.
func ptrToString(str *string) string {
    if str == nil {
        return ""
    }
    return *str
}
```

### Rate Limit

Anilist has default rate limit **90 requests per minute**. If you go over
the rate limit you'll receive a 1-minute timeout. But, verniy has
a built-in rate limiter to prevent the requests going over the limit.
It will put your request on hold until the limit is available again.

### OAuth2 Authentication

To access private user data, you need to use OAuth2 Access Token.

Read the [Anilist OAuth2 documentation](https://anilist.gitbook.io/anilist-apiv2-docs/overview/oauth/getting-started) for more details.

The OAuth2 Access Token can be set in the `Client` struct.

```go
c := verniy.New()
c.AccessToken = "your-access-token"
```

This token could also be used to update user data in the future.

## Trivia

[Verniy](https://en.wikipedia.org/wiki/Japanese_destroyer_Hibiki_(1932))'s name is taken from japanese
destroyer. She is the 24th ship of the Fubuki class or the 2nd ship of the Akatsuki class. Originally,
named Hibiki but later on, transferred to the Soviet Union and renamed to Верный (Verniy). Also,
[exists](https://en.kancollewiki.net/Hibiki) in Kantai Collection games and anime.

## License

MIT License

Copyright (c) 2021 Axel
