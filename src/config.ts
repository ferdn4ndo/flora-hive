import "dotenv/config";

function parseApiKeys(raw: string | undefined): string[] {
  if (!raw || typeof raw !== "string") return [];
  return raw
    .split(",")
    .map((k) => k.trim())
    .filter(Boolean);
}

function parseIntEnv(name: string, fallback: number): number {
  const v = process.env[name];
  if (v === undefined || v === "") return fallback;
  const n = Number.parseInt(v, 10);
  if (Number.isNaN(n)) {
    throw new Error(`Invalid integer for ${name}: ${v}`);
  }
  return n;
}

const apiKeys = parseApiKeys(process.env.HIVE_API_KEYS);

const mqttUrl = process.env.MQTT_URL;
if (!mqttUrl) {
  throw new Error("MQTT_URL is required");
}

const defaultQos = parseIntEnv("MQTT_DEFAULT_QOS", 1);
if (defaultQos < 0 || defaultQos > 2) {
  throw new Error("MQTT_DEFAULT_QOS must be 0, 1, or 2");
}

const topicPrefix =
  process.env.FLORA_TOPIC_PREFIX === undefined
    ? "flora"
    : String(process.env.FLORA_TOPIC_PREFIX).replace(/\/$/, "");

function defaultDevicesSubscribePattern(prefix: string): string {
  const p = prefix.trim();
  return p
    ? `${p}/environments/+/devices/+/heartbeat`
    : "environments/+/devices/+/heartbeat";
}

const devicesSubscribePattern =
  process.env.FLORA_DEVICES_SUBSCRIBE_TOPIC?.trim() ||
  defaultDevicesSubscribePattern(topicPrefix);

const deviceHeartbeatTtlSec = parseIntEnv(
  "FLORA_DEVICE_HEARTBEAT_TTL_SEC",
  180
);
if (deviceHeartbeatTtlSec < 10 || deviceHeartbeatTtlSec > 86400) {
  throw new Error(
    "FLORA_DEVICE_HEARTBEAT_TTL_SEC must be between 10 and 86400"
  );
}

const userverAuthHost = process.env.USERVER_AUTH_HOST?.replace(/\/$/, "") || "";
const userverSystemName = process.env.USERVER_AUTH_SYSTEM_NAME || "";
const userverSystemToken = process.env.USERVER_AUTH_SYSTEM_TOKEN || "";

/** Same variables as userver-filemgr / Django `DATABASES` (PostgreSQL from userver-datamgr). */
const postgresHost = process.env.POSTGRES_HOST;
const postgresDb = process.env.POSTGRES_DB;
const postgresUser = process.env.POSTGRES_USER;
const postgresPort = parseIntEnv("POSTGRES_PORT", 5432);

if (!postgresHost || !postgresDb || !postgresUser) {
  throw new Error(
    "POSTGRES_HOST, POSTGRES_DB, and POSTGRES_USER are required (PostgreSQL from userver-datamgr)"
  );
}
/**
 * App password for `POSTGRES_USER` (userver-filemgr uses `POSTGRES_PASS`).
 * `POSTGRES_PASSWORD` is accepted as a fallback (same name as the official Postgres Docker image).
 * Omit both or leave empty only when the role truly has no password / uses trust.
 */
const resolvedPostgresPassRaw =
  process.env.POSTGRES_PASS?.trim() ||
  process.env.POSTGRES_PASSWORD?.trim() ||
  "";
const resolvedPostgresPass =
  resolvedPostgresPassRaw === "" ? undefined : resolvedPostgresPassRaw;

const appConfig = {
  port: parseIntEnv("PORT", 8080),
  apiKeys,
  postgres: {
    host: postgresHost,
    port: postgresPort,
    user: postgresUser,
    ...(resolvedPostgresPass !== undefined
      ? { password: resolvedPostgresPass }
      : {}),
    database: postgresDb,
  },
  mqtt: {
    url: mqttUrl,
    username: process.env.MQTT_USERNAME || undefined,
    password: process.env.MQTT_PASSWORD || undefined,
    clientId: process.env.MQTT_CLIENT_ID || "flora-hive",
    defaultQos,
  },
  topicPrefix,
  devicesSubscribePattern,
  deviceHeartbeatTtlSec,
  userver: {
    host: userverAuthHost,
    systemName: userverSystemName,
    systemToken: userverSystemToken,
    configured: Boolean(
      userverAuthHost && userverSystemName && userverSystemToken
    ),
  },
};

export const config = appConfig;
export type AppConfig = typeof appConfig;
