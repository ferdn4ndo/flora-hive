export declare function listDevicesByEnvironmentId(environmentId: string, parentDeviceId: string | null | undefined): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
}[]>;
export declare function listDevicesInEnvironment(environmentId: string, userId: string, parentDeviceId: string | null | undefined): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
}[]>;
export declare function createDevice(input: {
    environmentId: string;
    userId: string;
    deviceType: string;
    deviceId: string;
    displayName?: string | null;
    parentDeviceId?: string | null;
}): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
}>;
export declare function getDeviceRowById(deviceRowId: string): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
}>;
export declare function getDeviceRowByEnvAndDeviceId(environmentId: string, deviceId: string): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
}>;
export declare function getDevice(deviceRowId: string, userId: string): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
} | null>;
export declare function getDeviceByEnvAndDeviceId(environmentId: string, logicalDeviceId: string, userId: string): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
} | null>;
export declare function updateDevice(deviceRowId: string, userId: string, patch: {
    deviceType?: string;
    deviceId?: string;
    displayName?: string | null;
    parentDeviceId?: string | null;
}): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
} | null>;
export declare function deleteDevice(deviceRowId: string, userId: string): Promise<boolean>;
export declare function updateDeviceByEnvAndDeviceId(environmentId: string, logicalDeviceId: string, userId: string, patch: {
    deviceType?: string;
    deviceId?: string;
    displayName?: string | null;
    parentDeviceId?: string | null;
}): Promise<{
    id: string;
    updatedAt: string;
    createdAt: string;
    environmentId: string;
    parentDeviceId: string | null;
    deviceType: string;
    deviceId: string;
    displayName: string | null;
} | null>;
export declare function deleteDeviceByEnvAndDeviceId(environmentId: string, logicalDeviceId: string, userId: string): Promise<boolean>;
//# sourceMappingURL=services.d.ts.map