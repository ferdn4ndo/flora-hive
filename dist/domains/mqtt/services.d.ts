import mqtt from "mqtt";
import type { AppConfig } from "../../config.js";
import type { PublicMqttDevice } from "./types.js";
export declare function initMqttService(config: AppConfig): void;
export declare function listLiveDevices(options: {
    includeOffline?: boolean;
    /** Catalog `devices.id` values the caller may see (API key: null = all). */
    allowedDeviceRowIds: string[] | null;
}): PublicMqttDevice[];
export declare function getMqttState(): {
    connected: boolean;
    clientId: string;
    url: string;
    lastError: string | null;
};
export declare function onMqtt(event: string, fn: (...args: unknown[]) => void): void;
export declare function connectMqtt(): mqtt.MqttClient;
export declare function disconnectMqtt(): Promise<void>;
export declare function publishMqtt(input: {
    topic: string;
    payload?: unknown;
    qos?: number | null;
    retain?: boolean;
}): Promise<{
    topic: string;
    qos: number;
    retain: boolean;
    bytes: number;
}>;
export { parseDeviceRowIdFromTopic, normalizeTopic } from "./topic.js";
//# sourceMappingURL=services.d.ts.map