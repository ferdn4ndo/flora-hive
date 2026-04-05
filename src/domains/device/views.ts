import type { Device } from "../../db/schema.js";

export function devicePublic(d: Device) {
  return {
    id: d.id,
    path: `environments/${d.environmentId}/devices/${d.deviceId}`,
    environmentId: d.environmentId,
    parentDeviceId: d.parentDeviceId,
    deviceType: d.deviceType,
    deviceId: d.deviceId,
    displayName: d.displayName,
    createdAt: d.createdAt,
    updatedAt: d.updatedAt,
  };
}
