FROM golang:1.25-alpine AS builder

RUN apk --no-cache add ca-certificates tzdata git

RUN adduser -D -u 10001 appuser

WORKDIR /build

COPY . .

RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-w -s" -o main

FROM alpine:3.19

# Install su-exec for user switching and shadow for usermod/groupmod
RUN apk --no-cache add ca-certificates tzdata su-exec shadow && \
    mkdir -p /etc/anilist-mal-sync && \
    adduser -D -u 10001 appuser && \
    mkdir -p /home/appuser/.config/anilist-mal-sync && \
    chown -R appuser:appuser /home/appuser

WORKDIR /app
COPY --from=builder /build/main ./main
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

LABEL org.opencontainers.image.source="https://github.com/username/anilist-mal-sync" \
    org.opencontainers.image.description="Synchronization service for AniList and MyAnimeList" \
    org.opencontainers.image.licenses="MIT"

EXPOSE 18080
ENTRYPOINT ["docker-entrypoint.sh"]
CMD []
