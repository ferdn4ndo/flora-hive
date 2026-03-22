# Flora Hive

**Flora Hive** is a small Node.js service for the Flora system: it maintains an MQTT client to your broker and exposes an **authenticated HTTP API** for publishing messages and listing devices that report over MQTT.

## Requirements

- **Node.js 20** (see `.nvmrc`). [NVM](https://github.com/nvm-sh/nvm) is recommended:

  ```bash
  nvm install
  nvm use
  ```

- A reachable **MQTT broker** (plain TLS, WebSocket, etc. supported via URL scheme).

- **Docker** and **Docker Compose** (optional), if you run Hive in a container.

## Quick start

1. Copy the environment template and edit values (especially `HIVE_API_KEYS` and `MQTT_URL`):

   ```bash
   cp .env.example .env
   ```

2. Install dependencies and start:

   ```bash
   npm install
   npm start
   ```

   For file-watching during development:

   ```bash
   npm run dev
   ```

The HTTP server listens on **`PORT`** (default **8080**).

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HIVE_API_KEYS` | Yes | — | Comma-separated API keys. Any one key grants access. |
| `MQTT_URL` | Yes | — | Broker URL (`mqtt://`, `mqtts://`, `ws://`, `wss://`). |
| `MQTT_USERNAME` | No | — | Broker username. |
| `MQTT_PASSWORD` | No | — | Broker password. |
| `MQTT_CLIENT_ID` | No | `flora-hive` | MQTT client id. |
| `MQTT_DEFAULT_QOS` | No | `1` | Default QoS for `POST /v1/mqtt/publish` when omitted. |
| `PORT` | No | `8080` | HTTP listen port. |
| `FLORA_TOPIC_PREFIX` | No | `flora` | Prefix applied to **publish** topics that are not already under this prefix. If unset in the environment, defaults to `flora`. Set to an **empty** value in `.env` (`FLORA_TOPIC_PREFIX=`) to disable prefixing. |
| `FLORA_DEVICES_SUBSCRIBE_TOPIC` | No | `flora/+/+/+/heartbeat` | MQTT subscription pattern Hive uses to observe devices. Must use `+` segments; see [Device listing](#device-listing). |
| `FLORA_DEVICE_HEARTBEAT_TTL_SEC` | No | `180` | For heartbeat-style messages, seconds after the last message before a device is treated as disconnected (10–86400). |

Generate a strong key, for example:

```bash
openssl rand -hex 32
```

## Authentication

All routes under `/v1/*` require an API key:

- Header **`X-API-Key: <key>`**, or  
- Header **`Authorization: Bearer <key>`**

`GET /healthz` is **not** authenticated (suitable for load balancers and health checks).

## HTTP API

### `GET /healthz`

Public. Returns service liveness.

**Response:** `{ "status": "ok", "service": "flora-hive" }`

---

### `GET /v1/whoami`

**Response:** `{ "ok": true, "role": "hive" }`

---

### `GET /v1/mqtt/connection`

MQTT session summary for the Hive process (broker URL has secrets redacted).

---

### `GET /v1/devices`

Lists known devices derived from subscribed MQTT traffic.

| Query | Meaning |
|-------|---------|
| `include_offline=1` (or `true` / `yes`) | Include devices that are stale or explicitly offline. |
| *(omitted)* | Only devices currently considered **connected**. |

**Response:** `{ "devices": [ … ] }`

Each element typically includes:

- `id` — Composite id from wildcard segments (e.g. `envId/deviceType/deviceId`).
- `identity` — Present for Flora-style **heartbeat** topics with three wildcards: `{ "envId", "deviceType", "deviceId" }`.
- `connected` — Whether the device is considered online (see [Device listing](#device-listing)).
- `lastSeenAt` — ISO timestamp of the last matching message.
- `lastTopic` — Full MQTT topic of that message.
- `telemetry` — Last parsed JSON payload when applicable (e.g. Flora heartbeat JSON).

---

### `POST /v1/mqtt/publish`

Publishes a message to MQTT (Hive must be connected to the broker).

**Body (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `topic` | Yes | Topic string. If it does not already start with `FLORA_TOPIC_PREFIX` (when prefix is non-empty), the prefix is prepended. |
| `payload` | No | String, object (serialized as JSON), or omitted for an empty payload. |
| `qos` | No | `0`, `1`, or `2` (defaults to `MQTT_DEFAULT_QOS`). |
| `retain` | No | Boolean retain flag. |

**Success:** `202` with `{ "ok": true, "topic", "qos", "retain", "bytes" }`.

**Errors:** `400` (validation), `503` (MQTT not connected), `500` (publish failure).

## Device listing

Hive **does not** query the broker for a client list. It infers devices from messages on **`FLORA_DEVICES_SUBSCRIBE_TOPIC`**.

### Default: Flora ESP firmware

The default pattern matches firmware that publishes heartbeats on:

`flora/{FLORA_ENV_ID}/{FLORA_DEVICE_TYPE}/{FLORA_DEVICE_ID}/heartbeat`

with JSON such as `ts`, `dht_status`, `temperature`, `humidity`, `registered_at`, etc.

For those messages, **`connected`** is **time-based**: `true` if a message arrived within **`FLORA_DEVICE_HEARTBEAT_TTL_SEC`** seconds. Tune this to be greater than your worst-case heartbeat interval.

### Optional: explicit presence topics

If you point `FLORA_DEVICES_SUBSCRIBE_TOPIC` at a dedicated status topic (e.g. `flora/+/+/+/status`), payloads can use:

- Plain text: `online`, `offline`, `true`, `false`, etc., or  
- JSON: `connected`, `online`, or `state` (`online` / `offline`, …).

Then **`connected`** follows that explicit signal instead of the heartbeat TTL.

### Custom patterns

The pattern may contain **multiple** `+` wildcards. The listed `id` is those segments joined with `/`. The pattern must align segment-for-segment with real topics (only `+` as a wildcard).

## MQTT → commands (firmware)

Devices subscribe to commands on:

`flora/{FLORA_ENV_ID}/{FLORA_DEVICE_TYPE}/{FLORA_DEVICE_ID}/commands`

Use **`POST /v1/mqtt/publish`** with a full topic or a suffix that your `FLORA_TOPIC_PREFIX` rules expand correctly (for example, with default prefix, `lab/grow/unit1/commands` becomes `flora/lab/grow/unit1/commands`).

## Docker

Build and run with Compose (loads `.env`):

```bash
docker compose build
docker compose up -d
```

Or use npm scripts:

```bash
npm run docker:build
npm run docker:up
```

The image uses **Node 20** (`Dockerfile`), aligned with **`.nvmrc`**. Do not bake secrets into the image; pass them via `.env` or your orchestrator.

## Project layout

```
src/
  index.js      # Process entry: MQTT connect, HTTP server, shutdown
  app.js        # Express routes
  auth.js       # API key middleware
  config.js     # Environment loading and validation
  mqttClient.js # MQTT connection, device registry, publish helpers
```

## License

[MIT](LICENSE)
