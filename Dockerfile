FROM golang:1.23-alpine AS builder

RUN apk --no-cache add ca-certificates tzdata git

RUN adduser -D -u 10001 appuser

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

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
