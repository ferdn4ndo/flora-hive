/** Normalize publish topic using optional prefix (matches Flora Hive publish rules). */
export declare function normalizeTopic(topic: string, topicPrefix: string): string;
/**
 * Environment id for MQTT ACL: path segment after `environments/` (with optional
 * `floraPrefix/` stripped), e.g. `flora/environments/<uuid>/devices/...` → `<uuid>`.
 */
export declare function parseEnvironmentIdFromTopic(topic: string, floraPrefix: string): string | null;
//# sourceMappingURL=topic.d.ts.map