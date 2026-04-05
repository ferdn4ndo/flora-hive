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
 * Environment id for MQTT ACL: path segment after `environments/` (with optional
 * `floraPrefix/` stripped), e.g. `flora/environments/<uuid>/devices/...` → `<uuid>`.
 */
export function parseEnvironmentIdFromTopic(
  topic: string,
  floraPrefix: string
): string | null {
  let t = topic.replace(/^\/+/, "");
  if (floraPrefix && t.startsWith(`${floraPrefix}/`)) {
    t = t.slice(floraPrefix.length + 1);
  }
  const m = t.match(/^environments\/([^/]+)/);
  return m?.[1] ?? null;
}
