# Flora Hive

**Flora Hive** is a **Go** service for the Flora stack: it connects to **MQTT**, stores **environments** and **nested device** metadata in **PostgreSQL** (via **sqlx** and **golang-migrate**), and exposes a JSON HTTP API. Authentication matches the **uServer-Auth** HTTP pattern used by [userver-filemgr](https://github.com/ferdn4ndo/userver-filemgr): Bearer JWT validated with `GET /auth/me`, plus optional `X-API-Key` for trusted automation.

## Requirements

- **Docker** — required for **`make build`** and **`make test-docker`** (the compiler version is fixed by the `GO_VERSION` build-arg in the [`Dockerfile`](Dockerfile), default `1.24.4`).
- **Go 1.24+** (optional) — only if you use **`make test`**, **`make test-race`**, **`make lint`**, or **`go run ./cmd`** on the host; align with `go.mod`.
- **PostgreSQL** — same env convention as userver-filemgr: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASS` (or `POSTGRES_PASSWORD`). A typical stack DB is [userver-datamgr](https://github.com/ferdn4ndo/userver-datamgr) (`userver-postgres` when sharing the Compose network).
- **MQTT broker**
- **uServer-Auth** (optional but required for user login/register/JWT flows): base URL + system name + system token

## Quick start

```bash
cp .env.example .env
# Set MQTT_URL, POSTGRES_*, USERVER_AUTH_* as needed, and optionally HIVE_API_KEYS
make build          # compiles inside Docker; writes bin/flora-hive
./bin/flora-hive app:serve
```

With a local Go toolchain you can skip Docker for the binary: `go run ./cmd app:serve`.

## CLI

| Command | Purpose |
|--------|---------|
| `go run ./cmd app:serve` | HTTP server (`PORT`, default `8080`) |
| `go run ./cmd migrate:up` | Apply SQL migrations from `./migrations` |
| `go run ./cmd migrate:down` | Roll back one migration step |

## Configuration

See [`.env.example`](.env.example). Important variables:

| Variable | Purpose |
|----------|---------|
| `MQTT_URL` | Broker URL (required). With userver-eventmgr: **`mqtt://userver-mosquitto:1883`**. Use **`ws://…:9001`** only for WebSockets — not `ws://…:1883`. |
| `MQTT_USERNAME` / `MQTT_PASSWORD` | Often required when the broker disallows anonymous clients. |
| `POSTGRES_HOST` / `POSTGRES_PORT` / `POSTGRES_DB` / `POSTGRES_USER` | App database (required). |
| `POSTGRES_PASS` or `POSTGRES_PASSWORD` | Password for `POSTGRES_USER` (either name works; omit only if the role has no password). |
| `POSTGRES_SSLMODE` | Default `disable`; set e.g. `require` for TLS to the server. |
| `ENVIRONMENT` | `local` (default) relaxes some production-only behavior (e.g. Gin dev mode, Sentry skipped). |
| `USERVER_AUTH_HOST` | Base URL of uServer-Auth (no trailing slash). |
| `USERVER_AUTH_SYSTEM_NAME` / `USERVER_AUTH_SYSTEM_TOKEN` | Forwarded on login/register server-side. |
| `SKIP_CONTAINER_PREPARE` | Set to `1` in Docker to skip the entrypoint migration step. |
| `HIVE_API_KEYS` | Optional comma-separated API keys (`X-API-Key`) — broad access for automation (see auth model). |
| `FLORA_TOPIC_PREFIX`, `FLORA_DEVICES_SUBSCRIBE_TOPIC`, `FLORA_DEVICE_HEARTBEAT_TTL_SEC` | MQTT topic behavior (default subscribe pattern: `{prefix}/+/heartbeat`; first `+` is catalog `devices.id`). |

## Authentication model

1. **Users (JWT)** — `Authorization: Bearer <access_token>`. Hive calls `{USERVER_AUTH_HOST}/auth/me` and syncs `hive_users`.
2. **Service (API key)** — `X-API-Key` when `HIVE_API_KEYS` is set. Intended for trusted backends; bypasses per-environment membership so automation can see all environments.

### Auth routes

| Method | Path | Notes |
|--------|------|--------|
| POST | `/v1/auth/login` | Body: `username`, `password` — Hive adds `system_name` / `system_token`. |
| POST | `/v1/auth/register` | Body: `username`, `password`, optional `is_admin`. |
| POST | `/v1/auth/refresh` | Body: `refresh_token`. |
| POST | `/v1/auth/logout` | Bearer access token. |
| GET | `/v1/auth/me` | Bearer — uServer `me` + Hive user profile. |
| PATCH | `/v1/auth/password` or `/v1/auth/reset-password` | Bearer — `current_password`, `new_password`. |

## Domain model

- **Environment** — `name`, optional `description`. Path id is the row UUID.
- **Membership** — **viewer** (read) or **editor** (read/write).
- **Device (catalog)** — `deviceType`, `deviceId`, optional `parentDeviceId`, optional `displayName`. HTTP uses `/v1/environments/.../devices/...`. MQTT uses catalog row UUID `devices.id` as the first segment after the flora prefix (e.g. `{prefix}/<devices.id>/heartbeat`).

## HTTP API overview

### Health

- `GET` / `HEAD` **`/healthz`** — public.

### MQTT

- `GET /v1/mqtt/connection` — broker connection status (redacted URL).
- `GET /v1/mqtt/devices` — live registry; JWT users see devices in their environments; API key sees all. `include_offline=1|true|yes`.
- `POST /v1/mqtt/publish` — editor on the device’s environment (or API key). Topic normalized with `FLORA_TOPIC_PREFIX`; first segment after prefix must match a catalog `devices.id`.

### Environments, members, devices

CRUD for environments, members, and nested devices under **`/v1/...`**. Add OpenAPI later if you want a generated contract.

## Project layout (Go)

```
cmd/                     # main, cobra commands (serve, migrate)
lib/                     # env, db, gin handler, logging
migrations/              # golang-migrate SQL
internal/
  domain/                # models, ports, pure helpers (e.g. mqtttopic)
  infrastructure/        # repos, userver HTTP client, MQTT service
  services/              # environment, device, user orchestration
  controllers/           # Gin routes, auth middleware
```

## Database

Schema is applied with **golang-migrate** (`migrate:up`). SQL lives in `migrations/` (`hive_users`, `environments`, `environment_members`, `devices`). Migrations use `IF NOT EXISTS` where appropriate when applying to an existing database.

## Docker

The [`Dockerfile`](Dockerfile) uses a **multi-stage** build: the **`build`** stage sits on top of **`test`**, so **`go test`** runs before the binary is compiled. **`make build`** and **`make image`** both execute that path (use **`make test-docker`** to stop after tests). Override the toolchain with a [build-arg](https://docs.docker.com/build/guide/build-args/): `docker build --build-arg GO_VERSION=1.24.5 ...`.

```bash
make test-docker    # only the test stage (faster when you are iterating on tests)
make image          # tests + compile + runtime image flora-hive:local (alias: make docker-build)
docker compose up --build
```

The runtime image entrypoint runs **`./flora-hive migrate:up`** unless `SKIP_CONTAINER_PREPARE=1`, then starts **`app:serve`**. Ensure `.env` has `POSTGRES_*`, `MQTT_URL`, and optional auth/MQTT tuning.

## Development

| Make target | Purpose |
|-------------|---------|
| `make build` | **`docker build --target build`**: runs **`go test` in-image**, then compiles and copies `bin/flora-hive` to the host |
| `make test-docker` | **`docker build --target test`** — same pinned Go, stops after tests (skips compile) |
| `make test` | Host `go test ./...` (fast when you have Go installed) |
| `make test-race` | Host `go test -race -count=1 ./...` (closest to CI race job) |
| `make image` / `make docker-build` | Full runtime image `flora-hive:local` |
| `make lint` | `golangci-lint run` (requires host binary; see `.golangci.yaml`) |
| `make migrate-up` | `make build` then `./bin/flora-hive migrate:up` |

Set **`GO_VERSION`** (Makefile and `docker build --build-arg`) to match the [`Dockerfile`](Dockerfile) default when bumping the compiler.

## uServer-Auth reference

Same HTTP surface as in userver-filemgr: `POST /auth/login`, `POST /auth/register`, `GET /auth/me`, etc. See [userver-auth](https://github.com/ferdn4ndo/userver-auth) for the canonical contract.

## License

[MIT](LICENSE)
