# Build stage
FROM golang:latest as builder
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o ./main

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/main ./main
EXPOSE 18080
ENTRYPOINT ./main -c /etc/anilist-mal-sync/config.yaml
