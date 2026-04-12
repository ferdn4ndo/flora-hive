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
| `go run ./cmd bootstrap:auth` | Optional uServer-Auth bootstrap (`POST /auth/system`, `POST /auth/register`); see [`.env.example`](.env.example) |

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
| `SKIP_CONTAINER_PREPARE` | Set to `1` in Docker to skip the entrypoint prepare step (Postgres bootstrap, `migrate:up`, `bootstrap:auth`). |
| `SKIP_USERVER_AUTH_SETUP` / `SKIP_AUTH_BOOTSTRAP` | Skip `bootstrap:auth` (Docker entrypoint and [`setup.sh`](setup.sh)). |
| `USERVER_AUTH_SYSTEM_CREATION_TOKEN` / `SYSTEM_CREATION_TOKEN` | Bootstrap: `Authorization: Token …` on `POST /auth/system`. |
| `HIVE_SKIP_PERSIST_BOOTSTRAP_ENV` / `FILEMGR_SKIP_PERSIST_BOOTSTRAP_ENV` | Set to `1` to avoid writing bootstrap tokens into `.env`. |
| `HIVE_API_KEYS` | Optional comma-separated API keys (`X-API-Key`) — broad access for automation (see auth model). |
| `FLORA_TOPIC_PREFIX`, `FLORA_DEVICES_SUBSCRIBE_TOPIC`, `FLORA_DEVICE_HEARTBEAT_TTL_SEC` | MQTT topic behavior (default subscribe pattern: `{prefix}/+/heartbeat`; first `+` is catalog `devices.id`). |
| `CORS_ALLOWED_ORIGINS` | Optional comma-separated allowed **`Origin`** values (e.g. `https://flora.sd40.com.br`). If unset, any origin is allowed. CORS runs on the **whole Gin engine** so `OPTIONS` preflights get headers even when no route matches yet. |

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

### Optional: create DB and role (Docker / superuser)

When the app runs **inside the Docker image**, the entrypoint runs [`scripts/docker-bootstrap-postgres.sh`](scripts/docker-bootstrap-postgres.sh) **before** `migrate:up` (unless `SKIP_CONTAINER_PREPARE=1`). If **`POSTGRES_ROOT_USER`** and **`POSTGRES_ROOT_PASS`** are set, it waits for PostgreSQL, then creates **`POSTGRES_DB`** and **`POSTGRES_USER`** with password **`POSTGRES_PASS`** or **`POSTGRES_PASSWORD`** when they are missing—same idea as **userver-filemgr** `setup.sh`. **`POSTGRES_ADMIN_DATABASE`** defaults to **`postgres`**; **`POSTGRES_SSLMODE`** is passed to `psql` via **`PGSSLMODE`**.

Use this when Hive points at a server that only has the default `postgres` superuser (for example first-time **docker compose** against **userver-postgres**): set root credentials to match that instance, and set **`POSTGRES_PASS`** to the password you want for **`POSTGRES_USER`**.

## Docker

The [`Dockerfile`](Dockerfile) uses a **multi-stage** build: the **`build`** stage sits on top of **`test`**, so **`go test`** runs before the binary is compiled. **`make build`** and **`make image`** both execute that path (use **`make test-docker`** to stop after tests). Override the toolchain with a [build-arg](https://docs.docker.com/build/guide/build-args/): `docker build --build-arg GO_VERSION=1.24.5 ...`.

```bash
make test-docker    # only the test stage (faster when you are iterating on tests)
make image          # tests + compile + runtime image flora-hive:local (alias: make docker-build)
```

**Compose — development** ([`docker-compose.yml`](docker-compose.yml)): builds from [`Dockerfile`](Dockerfile) and tags **`flora-hive:local`**.

```bash
docker compose up --build
```

**Compose — production** ([`docker-compose.prod.yml`](docker-compose.prod.yml)): **no `build`**. Defaults to **Docker Hub** [`ferdn4ndo/flora-hive:latest`](https://hub.docker.com/r/ferdn4ndo/flora-hive) (same repository the [release container workflow](.github/workflows/create_release_container.yaml) pushes on each GitHub Release). Pin a version with **`FLORA_HIVE_IMAGE`** (e.g. `ferdn4ndo/flora-hive:0.2.0`) in the shell or `.env`.

```bash
docker compose -f docker-compose.prod.yml pull   # optional
docker compose -f docker-compose.prod.yml up -d
# Pin a release tag (workflow pushes ferdn4ndo/flora-hive:<version> + :latest):
# FLORA_HIVE_IMAGE=ferdn4ndo/flora-hive:0.2.0 docker compose -f docker-compose.prod.yml up -d
```

The runtime image entrypoint runs **`docker-bootstrap-postgres.sh`** (when `POSTGRES_ROOT_*` is set), then **`./flora-hive migrate:up`**, then **`./flora-hive bootstrap:auth`** (non-fatal on failure, same idea as [userver-filemgr](https://github.com/ferdn4ndo/userver-filemgr) `setup.sh`), then **`app:serve`**—unless **`SKIP_CONTAINER_PREPARE=1`** (then only the command you pass runs). Ensure `.env` has `POSTGRES_*`, `MQTT_URL`, and optional auth/MQTT tuning.

### Host setup script

[`setup.sh`](setup.sh) loads `.env` if present, runs optional Postgres bootstrap, **`migrate:up`**, then **`bootstrap:auth`**. Override the binary with **`MIGRATE_BIN`** (default **`./bin/flora-hive`**; run **`make build`** first). Skip auth with **`SKIP_AUTH_BOOTSTRAP=1`** or **`SKIP_USERVER_AUTH_SETUP=1`**.

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
