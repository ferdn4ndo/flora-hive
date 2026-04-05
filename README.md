# Flora Hive

**Flora Hive** is a Node.js **TypeScript** service for the Flora stack: it connects to **MQTT**, stores **environments** and **nested device** metadata in **PostgreSQL** (via [Drizzle ORM](https://orm.drizzle.team/)), and exposes an HTTP API. Authentication follows the same **uServer-Auth** HTTP pattern as [userver-filemgr](https://github.com/ferdn4ndo/userver-filemgr) (Bearer JWT from the Flask auth service, validated with `GET /auth/me`).

## Requirements

- **Node.js 20** (`.nvmrc`) — use [NVM](https://github.com/nvm-sh/nvm): `nvm install && nvm use`
- **PostgreSQL** — same env convention as [userver-filemgr](https://github.com/ferdn4ndo/userver-filemgr): `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASS`. A typical stack DB is [userver-datamgr](https://github.com/ferdn4ndo/userver-datamgr) (`userver-postgres` when sharing the Compose network).
- **MQTT broker**
- **uServer-Auth** (optional but required for user login/register/JWT flows): base URL + system name + system token
- **Docker** (optional)

## Quick start

```bash
cp .env.example .env
# Set MQTT_URL, USERVER_AUTH_* , and optionally HIVE_API_KEYS
npm install
npm run dev
```

Production:

```bash
npm run build
npm start
```

## Configuration

See [`.env.example`](.env.example). Important variables:

| Variable | Purpose |
|----------|---------|
| `MQTT_URL` | Broker URL (required). With userver-eventmgr: **`mqtt://userver-mosquitto:1883`** (native MQTT on 1883). Use **`ws://userver-mosquitto:9001`** only for WebSockets — not `ws://…:1883` (that causes immediate disconnect / “socket hang up”). |
| `MQTT_USERNAME` / `MQTT_PASSWORD` | Often required; eventmgr’s Mosquitto uses `allow_anonymous false` and a password file. |
| `POSTGRES_HOST` / `POSTGRES_PORT` / `POSTGRES_DB` / `POSTGRES_USER` | App database (required); matches userver-filemgr / Django `DATABASES`. |
| `POSTGRES_PASS` or `POSTGRES_PASSWORD` | Password for `POSTGRES_USER` (either name works; empty only if the role has no password). |
| `USERVER_AUTH_HOST` | Base URL of uServer-Auth (no trailing slash). |
| `USERVER_AUTH_SYSTEM_NAME` / `USERVER_AUTH_SYSTEM_TOKEN` | Forwarded on login/register (server-side); not exposed to clients on those routes. |
| `USERVER_AUTH_SYSTEM_CREATION_TOKEN` | Optional; matches uServer-Auth `SYSTEM_CREATION_TOKEN`. Required for `db:prepare` to create the system when login fails. |
| `USERVER_AUTH_USER` / `USERVER_AUTH_PASSWORD` | Optional; admin user for `db:prepare` to probe login and register if missing. |
| `SKIP_USERVER_AUTH_SETUP` | Set to `1` to skip auth bootstrap in `db:prepare` / container entrypoint. |
| `SKIP_CONTAINER_PREPARE` | Set to `1` in Docker to skip entrypoint DB + auth bootstrap (app only). |
| `HIVE_API_KEYS` | Optional comma-separated API keys (`X-API-Key`) for automation — **full** access to MQTT + all environments. |
| `FLORA_TOPIC_PREFIX`, `FLORA_DEVICES_SUBSCRIBE_TOPIC`, `FLORA_DEVICE_HEARTBEAT_TTL_SEC` | MQTT topic behaviour (default subscribe pattern: `{prefix}/environments/+/devices/+/heartbeat`). |

## Authentication model

1. **Users (JWT)** — `Authorization: Bearer <access_token>` from uServer-Auth. Hive calls `GET {USERVER_AUTH_HOST}/auth/me` to validate and sync a local `hive_users` row.
2. **Service (API key)** — `X-API-Key: <key>` when `HIVE_API_KEYS` is set. Intended for trusted backends; bypasses per-environment membership for MQTT listing/publish rules that use “all envs”.

### Auth routes (proxied / local)

| Method | Path | Notes |
|--------|------|--------|
| POST | `/v1/auth/login` | Body: `{ "username", "password" }` — Hive adds `system_name` / `system_token` from env. |
| POST | `/v1/auth/register` | Body: `{ "username", "password", "is_admin"? }`. |
| POST | `/v1/auth/refresh` | Body: `{ "refresh_token" }`. |
| POST | `/v1/auth/logout` | Bearer access token. |
| GET | `/v1/auth/me` | Bearer — returns uServer `me` + Hive user profile. |
| PATCH | `/v1/auth/password` or `/v1/auth/reset-password` | Bearer — `{ "current_password", "new_password" }` → uServer `/auth/me/password`. |

## Domain model

- **Environment** — `name`, optional `description`. MQTT path prefix for the environment is `environments/<environment_id>` (same as the environment row `id`).
- **Membership** — each user is **viewer** (read) or **editor** (read/write) on an environment.
- **Device (catalog)** — logical devices under an environment: `deviceType`, `deviceId`, optional `parentDeviceId` for nesting, optional `displayName`. MQTT device paths use `environments/<environment_id>/devices/<device_id>`. Distinct from **live MQTT devices** below.

## HTTP API overview

### Health

- `GET /healthz` — public.

### MQTT

- `GET /v1/mqtt/connection` — broker connection status.
- `GET /v1/mqtt/devices` — live devices from MQTT heartbeats, **filtered to environments the JWT user can access**. API key sees all. Query `include_offline=1` to include stale rows.
- `POST /v1/mqtt/publish` — **editor** on the target environment (or API key). Normalized topic must include `environments/<environment_id>/…` for a known environment id.

### Environments CRUD

- `GET /v1/environments` — JWT: only member environments; API key: all.
- `POST /v1/environments` — JWT only (creator becomes **editor**). Body `{ "name", "description"? }`.
- `GET|PATCH|DELETE /v1/environments/:environmentId` — membership required; PATCH/DELETE need **editor**.

### Members CRUD

- `GET /v1/environments/:environmentId/members`
- `POST /v1/environments/:environmentId/members` — body `{ "authUserUuid", "role": "viewer"|"editor" }` (target user must have called `GET /v1/auth/me` once so Hive has a row).
- `PATCH /v1/environments/:environmentId/members/:userId` — body `{ "role" }`.
- `DELETE /v1/environments/:environmentId/members/:userId`

### Device catalog CRUD

- `GET /v1/environments/:environmentId/devices` — query `parent` omitted = all; `parent=null` roots only; `parent=<uuid>` children (internal device row id).
- `POST /v1/environments/:environmentId/devices` — editor only.
- `GET|PATCH|DELETE /v1/environments/:environmentId/devices/:deviceId` — JWT membership; API key may GET any device (`deviceId` is the catalog logical id, not the internal row UUID).

## Project layout (domain-based)

```
src/
  app.ts                 # Express wiring
  index.ts               # Entry, MQTT + HTTP + shutdown
  config.ts
  db/                    # PostgreSQL bootstrap SQL + Drizzle schema
  http/params.ts
  domains/
    auth/                # userver HTTP client, middleware, controller, types
    user/                # services, views
    environment/         # rbac, services, views, controller
    device/              # services, views, controller
    mqtt/                # topic helpers, MQTT client service, types
  types/express.d.ts
```

## Database provisioning

Create a database and role on your PostgreSQL instance (same patterns as userver-filemgr: PostgreSQL 15+ may need `GRANT USAGE, CREATE ON SCHEMA public` for the app user). On first start, Hive runs **idempotent** DDL (`src/db/bootstrapPg.ts`) to create tables and indexes.

For **drizzle-kit** introspection or migrations: `npm run db:generate` with `POSTGRES_*` set (see `drizzle.config.ts`).

## Docker

```bash
docker compose up --build
```

Point `.env` at your Postgres host (e.g. `POSTGRES_HOST=userver-postgres` on the userver-datamgr Docker network). The image **entrypoint** runs `dist/containerPrepare.js` first (Postgres DB/role when `POSTGRES_ROOT_*` is set, Hive DDL, optional uServer-Auth bootstrap), then **`node dist/index.js`**. Set `SKIP_CONTAINER_PREPARE=1` to skip that step.

## Scripts

| Script | Purpose |
|--------|---------|
| `npm run dev` | `tsx watch src/index.ts` |
| `npm run build` | `tsc` → `dist/` |
| `npm start` | `node dist/index.js` |
| `npm test` | Vitest unit tests |
| `npm run db:generate` | Drizzle Kit (optional migrations) |
| `npm run db:prepare` | Postgres bootstrap (optional), Hive DDL, then uServer-Auth system + admin (optional; see `.env.example`) |

## uServer-Auth reference

Same API surface as in userver-filemgr’s `UServerAuthenticationService`: `POST /auth/login`, `POST /auth/register`, `GET /auth/me` with `Authorization: Bearer`, etc. See [userver-auth](https://github.com/ferdn4ndo/userver-auth) for the canonical contract.

## License

[MIT](LICENSE)
