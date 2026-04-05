export type DeviceRegistryEntry = {
  id: string;
  lastSeenAt: string;
  lastTopic: string;
  presenceKind: "ttl" | "explicit";
  explicitConnected?: boolean;
  meta?: object;
  identity?: {
    environmentId: string;
    deviceId: string;
    path: string;
  };
};

export type PublicMqttDevice = {
  id: string;
  connected: boolean;
  lastSeenAt: string;
  lastTopic: string;
  identity?: {
    environmentId: string;
    deviceId: string;
    path: string;
  };
  telemetry?: object;
};
