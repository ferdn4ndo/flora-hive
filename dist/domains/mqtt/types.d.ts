export type DeviceRegistryEntry = {
    id: string;
    lastSeenAt: string;
    lastTopic: string;
    presenceKind: "ttl" | "explicit";
    explicitConnected?: boolean;
    meta?: object;
    /** Catalog `devices.id` — same UUID used as the first MQTT segment under the prefix. */
    identity?: {
        deviceRowId: string;
    };
};
export type PublicMqttDevice = {
    id: string;
    connected: boolean;
    lastSeenAt: string;
    lastTopic: string;
    identity?: {
        deviceRowId: string;
    };
    telemetry?: object;
};
//# sourceMappingURL=types.d.ts.map