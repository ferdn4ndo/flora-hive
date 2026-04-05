import type { Device } from "../../db/schema.js";
/** `id` is the catalog row UUID — firmware publishes `{prefix}/<id>/…` on MQTT. */
export declare function devicePublic(d: Device): {
    id: string;
    path: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
    createdAt: string;
    updatedAt: string;
};
//# sourceMappingURL=views.d.ts.map