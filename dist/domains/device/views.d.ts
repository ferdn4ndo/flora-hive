import type { Device } from "../../db/schema.js";
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