import "dotenv/config";

function parseApiKeys(raw) {
  if (!raw || typeof raw !== "string") return [];
  return raw
    .split(",")
    .map((k) => k.trim())
    .filter(Boolean);
}

function parseIntEnv(name, fallback) {
  const v = process.env[name];
  if (v === undefined || v === "") return fallback;
  const n = Number.parseInt(v, 10);
  if (Number.isNaN(n)) {
    throw new Error(`Invalid integer for ${name}: ${v}`);
  }
  return n;
}

const apiKeys = parseApiKeys(process.env.HIVE_API_KEYS);
if (apiKeys.length === 0) {
  throw new Error(
    "HIVE_API_KEYS must be set to one or more comma-separated API keys"
  );
}

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

// Matches firmware: flora/{FLORA_ENV_ID}/{FLORA_DEVICE_TYPE}/{FLORA_DEVICE_ID}/heartbeat
const devicesSubscribePattern =
  process.env.FLORA_DEVICES_SUBSCRIBE_TOPIC?.trim() ||
  "flora/+/+/+/heartbeat";

const deviceHeartbeatTtlSec = parseIntEnv("FLORA_DEVICE_HEARTBEAT_TTL_SEC", 180);
if (deviceHeartbeatTtlSec < 10 || deviceHeartbeatTtlSec > 86400) {
  throw new Error("FLORA_DEVICE_HEARTBEAT_TTL_SEC must be between 10 and 86400");
}

export const config = {
  port: parseIntEnv("PORT", 8080),
  apiKeys,
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
};
