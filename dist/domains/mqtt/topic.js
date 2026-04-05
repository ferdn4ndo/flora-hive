/** Normalize publish topic using optional prefix (matches Flora Hive publish rules). */
export function normalizeTopic(topic, topicPrefix) {
    let t = String(topic).trim();
    if (t.startsWith("/"))
        t = t.slice(1);
    if (!t)
        throw new Error("topic is empty");
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
export function parseDeviceRowIdFromTopic(topic, topicPrefix) {
    let t = topic.replace(/^\/+/, "");
    if (topicPrefix && t.startsWith(`${topicPrefix}/`)) {
        t = t.slice(topicPrefix.length + 1);
    }
    const seg = t.split("/")[0];
    return seg || null;
}
//# sourceMappingURL=topic.js.map