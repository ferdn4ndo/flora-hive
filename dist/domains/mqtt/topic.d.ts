/** Normalize publish topic using optional prefix (matches Flora Hive publish rules). */
export declare function normalizeTopic(topic: string, topicPrefix: string): string;
/**
 * Device row id for MQTT ACL: first path segment after optional `topicPrefix/`
 * (the catalog `devices.id` UUID). E.g. `flora/<uuid>/heartbeat` → `<uuid>`.
 */
export declare function parseDeviceRowIdFromTopic(topic: string, topicPrefix: string): string | null;
//# sourceMappingURL=topic.d.ts.map