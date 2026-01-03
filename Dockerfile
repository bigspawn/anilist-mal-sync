FROM golang:1.23-alpine AS builder

RUN apk --no-cache add ca-certificates tzdata git

RUN adduser -D -u 10001 appuser

WORKDIR /build

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download

COPY . .
RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-w -s" -o anilist-mal-sync .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    mkdir -p /etc/anilist-mal-sync

WORKDIR /app

COPY --from=builder /build/anilist-mal-sync ./anilist-mal-sync
COPY --from=builder /etc/passwd /etc/passwd

USER appuser

LABEL org.opencontainers.image.source="https://github.com/Tareku99/anilist-mal-sync" \
    org.opencontainers.image.description="Synchronization service for AniList and MyAnimeList" \
    org.opencontainers.image.licenses="MIT"

EXPOSE 18080
ENTRYPOINT ["./anilist-mal-sync"]
