# Toolchain version: bump GO_VERSION when you adopt a new Go release (keep in sync with go.mod minimum).
ARG GO_VERSION=1.24.4

# --- Shared toolchain (all compile/test steps use this image only) ---
FROM golang:${GO_VERSION}-alpine AS gobase
WORKDIR /src
RUN apk add --no-cache git ca-certificates

FROM gobase AS deps
COPY go.mod go.sum ./
RUN go mod download

# --- Unit tests (same Go as the compile stage) ---
FROM deps AS test
COPY . .
RUN go mod verify \
  && go test -count=1 ./...

# --- Static binary (runs only after tests succeed) ---
FROM test AS build
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/flora-hive ./cmd

# --- Runtime ---
FROM alpine:3.23

RUN apk add --no-cache ca-certificates wget

WORKDIR /app
COPY --from=build /out/flora-hive /app/flora-hive
COPY migrations ./migrations
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh /app/flora-hive

ENV GIN_MODE=release
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=45s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["/app/flora-hive", "app:serve"]
