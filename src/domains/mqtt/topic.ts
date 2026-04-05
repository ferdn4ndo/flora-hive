/** Normalize publish topic using optional prefix (matches Flora Hive publish rules). */
export function normalizeTopic(topic: string, topicPrefix: string): string {
  let t = String(topic).trim();
  if (t.startsWith("/")) t = t.slice(1);
  if (!t) throw new Error("topic is empty");
  const p = topicPrefix;
  if (p && !t.startsWith(`${p}/`) && t !== p) {
    t = `${p}/${t}`;
  }
  return t;
}

/**
 * Device row id for MQTT ACL: first path segment after optional `topicPrefix/`
 * (the catalog `devices.id` UUID). E.g. `flora/<uuid>/heartbeat` → `<uuid>`.
 */
export function parseDeviceRowIdFromTopic(
  topic: string,
  topicPrefix: string
): string | null {
  let t = topic.replace(/^\/+/, "");
  if (topicPrefix && t.startsWith(`${topicPrefix}/`)) {
    t = t.slice(topicPrefix.length + 1);
  }
  const seg = t.split("/")[0];
  return seg || null;
}
