/** `id` is the catalog row UUID — firmware publishes `{prefix}/<id>/…` on MQTT. */
export function devicePublic(d) {
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
//# sourceMappingURL=views.js.map