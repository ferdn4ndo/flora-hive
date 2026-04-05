import mqtt from "mqtt";
import EventEmitter from "node:events";
import type { AppConfig } from "../../config.js";
import type {
  DeviceRegistryEntry,
  PublicMqttDevice,
} from "./types.js";
import { normalizeTopic as normalizePublishTopic } from "./topic.js";

const events = new EventEmitter();

const deviceRegistry = new Map<string, DeviceRegistryEntry>();

let client: mqtt.MqttClient | null = null;
let connected = false;
let lastError: Error | null = null;

let cfgRef: AppConfig | null = null;

export function initMqttService(config: AppConfig) {
  cfgRef = config;
}

function getCfg(): AppConfig {
  if (!cfgRef) throw new Error("MQTT service not initialized");
  return cfgRef;
}

function extractWildcardSegments(
  pattern: string,
  topic: string
): string[] | null {
  const pa = pattern.split("/");
  const ta = topic.split("/");
  if (pa.length !== ta.length) return null;
  const segments: string[] = [];
  for (let i = 0; i < pa.length; i++) {
    if (pa[i] === "+") {
      segments.push(ta[i]);
    } else if (pa[i] !== ta[i]) {
      return null;
    }
  }
  return segments.length > 0 ? segments : null;
}

function compositeDeviceId(segments: string[]) {
  if (segments.length === 1) return segments[0]!;
  return segments.join("/");
}

function hiveIdentityFromSegments(segments: string[]) {
  if (segments.length !== 1) return undefined;
  const deviceRowId = segments[0]!;
  return { deviceRowId };
}

function looksLikeFloraHeartbeatJson(j: unknown): j is Record<string, unknown> {
  return (
    !!j &&
    typeof j === "object" &&
    (typeof (j as { ts?: unknown }).ts === "number" ||
      typeof (j as { ts?: unknown }).ts === "string" ||
      (j as { dht_status?: unknown }).dht_status !== undefined ||
      (j as { registered_at?: unknown }).registered_at !== undefined)
  );
}

function parseDevicePayload(payloadBuf: Buffer): {
  kind: "heartbeat" | "status";
  connected?: boolean;
  meta?: object;
} {
  const s = payloadBuf.toString("utf8").trim();
  if (!s) return { kind: "heartbeat", meta: undefined };
  try {
    const j: unknown = JSON.parse(s);
    if (looksLikeFloraHeartbeatJson(j)) {
      return { kind: "heartbeat", meta: j };
    }
    if (typeof j === "object" && j !== null) {
      const o = j as Record<string, unknown>;
      if (typeof o.connected === "boolean") {
        return { kind: "status", connected: o.connected, meta: o };
      }
      if (typeof o.online === "boolean") {
        return { kind: "status", connected: o.online, meta: o };
      }
      const st = String(o.state || "").toLowerCase();
      if (st === "offline" || st === "disconnected") {
        return { kind: "status", connected: false, meta: o };
      }
      if (st === "online" || st === "connected") {
        return { kind: "status", connected: true, meta: o };
      }
      return { kind: "heartbeat", meta: o };
    }
  } catch {
    /* plain text */
  }

  const lower = s.toLowerCase();
  if (
    lower === "offline" ||
    lower === "false" ||
    lower === "0" ||
    lower === "disconnected"
  ) {
    return { kind: "status", connected: false, meta: undefined };
  }
  if (
    lower === "online" ||
    lower === "true" ||
    lower === "1" ||
    lower === "connected"
  ) {
    return { kind: "status", connected: true, meta: undefined };
  }
  return { kind: "heartbeat", meta: { raw: s } };
}

function millisSinceIso(iso: string) {
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return Infinity;
  return Date.now() - t;
}

function entryIsConnected(entry: DeviceRegistryEntry, ttlSec: number) {
  if (entry.presenceKind === "explicit") {
    return entry.explicitConnected === true;
  }
  return millisSinceIso(entry.lastSeenAt) <= ttlSec * 1000;
}

function toPublicDevice(entry: DeviceRegistryEntry, ttlSec: number): PublicMqttDevice {
  const row: PublicMqttDevice = {
    id: entry.id,
    connected: entryIsConnected(entry, ttlSec),
    lastSeenAt: entry.lastSeenAt,
    lastTopic: entry.lastTopic,
  };
  if (entry.identity) row.identity = entry.identity;
  if (entry.meta !== undefined) row.telemetry = entry.meta;
  return row;
}

function subscribeDeviceTopics() {
  if (!client || !connected) return;
  const pattern = getCfg().devicesSubscribePattern;
  client.subscribe(pattern, { qos: 1 }, (err) => {
    if (err) events.emit("subscribe_error", err);
  });
}

export function listLiveDevices(options: {
  includeOffline?: boolean;
  /** Catalog `devices.id` values the caller may see (API key: null = all). */
  allowedDeviceRowIds: string[] | null;
}): PublicMqttDevice[] {
  const config = getCfg();
  const ttl = config.deviceHeartbeatTtlSec;
  const rows = [...deviceRegistry.values()].map((e) => toPublicDevice(e, ttl));
  let filtered = options.includeOffline
    ? rows
    : rows.filter((d) => d.connected);

  if (options.allowedDeviceRowIds !== null) {
    const allow = new Set(options.allowedDeviceRowIds);
    filtered = filtered.filter((d) => {
      const rowId = d.identity?.deviceRowId ?? d.id;
      return allow.has(rowId);
    });
  }

  filtered.sort((a, b) => a.id.localeCompare(b.id));
  return filtered;
}

export function getMqttState() {
  const config = getCfg();
  return {
    connected,
    clientId: config.mqtt.clientId,
    url: redactUrl(config.mqtt.url),
    lastError: lastError ? String(lastError.message || lastError) : null,
  };
}

function redactUrl(url: string) {
  try {
    const u = new URL(url);
    if (u.password) u.password = "***";
    if (u.username) u.username = u.username ? "***" : "";
    return u.toString();
  } catch {
    return "(invalid url)";
  }
}

export function onMqtt(event: string, fn: (...args: unknown[]) => void) {
  events.on(event, fn);
}

export function connectMqtt() {
  const config = getCfg();
  if (client) return client;

  const opts: mqtt.IClientOptions = {
    clientId: config.mqtt.clientId,
    reconnectPeriod: 5_000,
    connectTimeout: 30_000,
  };
  if (config.mqtt.username) opts.username = config.mqtt.username;
  if (config.mqtt.password) opts.password = config.mqtt.password;

  client = mqtt.connect(config.mqtt.url, opts);

  client.on("connect", () => {
    connected = true;
    lastError = null;
    subscribeDeviceTopics();
    events.emit("connect");
  });

  client.on("message", (topic, payload) => {
    const segments = extractWildcardSegments(
      config.devicesSubscribePattern,
      topic
    );
    if (!segments) return;
    const id = compositeDeviceId(segments);
    const parsed = parseDevicePayload(payload);
    const identity = hiveIdentityFromSegments(segments);
    const now = new Date().toISOString();

    if (parsed.kind === "heartbeat") {
      deviceRegistry.set(id, {
        id,
        lastSeenAt: now,
        lastTopic: topic,
        presenceKind: "ttl",
        ...(identity ? { identity } : {}),
        ...(parsed.meta !== undefined ? { meta: parsed.meta } : {}),
      });
      return;
    }

    deviceRegistry.set(id, {
      id,
      lastSeenAt: now,
      lastTopic: topic,
      presenceKind: "explicit",
      explicitConnected: parsed.connected,
      ...(identity ? { identity } : {}),
      ...(parsed.meta !== undefined ? { meta: parsed.meta } : {}),
    });
  });

  client.on("reconnect", () => {
    connected = false;
    events.emit("reconnect");
  });

  client.on("close", () => {
    connected = false;
    events.emit("close");
  });

  client.on("error", (err) => {
    lastError = err;
    connected = false;
    events.emit("error", err);
  });

  client.on("offline", () => {
    connected = false;
    events.emit("offline");
  });

  return client;
}

export function disconnectMqtt(): Promise<void> {
  return new Promise((resolve) => {
    if (!client) {
      resolve();
      return;
    }
    const c = client;
    client = null;
    connected = false;
    deviceRegistry.clear();
    c.end(false, {}, () => resolve());
  });
}

function encodePayload(payload: unknown): Buffer {
  if (payload === null || payload === undefined) return Buffer.alloc(0);
  if (Buffer.isBuffer(payload)) return payload;
  if (typeof payload === "string") return Buffer.from(payload, "utf8");
  return Buffer.from(JSON.stringify(payload), "utf8");
}

export async function publishMqtt(input: {
  topic: string;
  payload?: unknown;
  qos?: number | null;
  retain?: boolean;
}): Promise<{ topic: string; qos: number; retain: boolean; bytes: number }> {
  const config = getCfg();
  if (!client) throw new Error("MQTT client not initialized");
  if (!connected) throw new Error("MQTT not connected");

  const t = normalizePublishTopic(input.topic, config.topicPrefix);
  const q =
    input.qos !== undefined && input.qos !== null
      ? Number(input.qos)
      : config.mqtt.defaultQos;
  if (q < 0 || q > 2 || Number.isNaN(q)) {
    throw new Error("qos must be 0, 1, or 2");
  }
  const r = Boolean(input.retain);
  const buf = encodePayload(input.payload);

  return new Promise((resolve, reject) => {
    client!.publish(t, buf, { qos: q as 0 | 1 | 2, retain: r }, (err) => {
      if (err) reject(err);
      else resolve({ topic: t, qos: q, retain: r, bytes: buf.length });
    });
  });
}

export { parseDeviceRowIdFromTopic, normalizeTopic } from "./topic.js";
