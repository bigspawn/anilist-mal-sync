FROM golang:1.22-alpine AS build
WORKDIR /src

# install git for `go get` if needed and ca-certificates
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/anilist-mal-sync ./...

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /out/anilist-mal-sync /usr/local/bin/anilist-mal-sync
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/anilist-mal-sync"]
FROM golang:1.23-alpine AS builder

RUN apk --no-cache add ca-certificates tzdata git

RUN adduser -D -u 10001 appuser

WORKDIR /build

COPY . .

RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-w -s" -o main

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    mkdir -p /etc/anilist-mal-sync

WORKDIR /app
COPY --from=builder /build/main ./main
COPY --from=builder /etc/passwd /etc/passwd

USER appuser

LABEL org.opencontainers.image.source="https://github.com/username/anilist-mal-sync" \
    org.opencontainers.image.description="Synchronization service for AniList and MyAnimeList" \
    org.opencontainers.image.licenses="MIT"

EXPOSE 18080
ENTRYPOINT ["./main", "-c", "/etc/anilist-mal-sync/config.yaml"]
