import { pathParam } from "../../http/params.js";
import { requireAuth } from "../auth/middleware.js";
import { ForbiddenError } from "../environment/services.js";
import { devicePublic } from "./views.js";
import { createDevice, deleteDeviceByEnvAndDeviceId, getDeviceByEnvAndDeviceId, getDeviceRowByEnvAndDeviceId, listDevicesByEnvironmentId, listDevicesInEnvironment, updateDeviceByEnvAndDeviceId, } from "./services.js";
function handleErr(res, next, e) {
    if (e instanceof ForbiddenError) {
        res.status(403).json({ error: "forbidden", message: e.message });
        return;
    }
    if (e instanceof Error && e.message === "Invalid parent device") {
        res.status(400).json({ error: "invalid_request", message: e.message });
        return;
    }
    next(e);
}
export function listDevices(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            const parent = req.query.parent;
            const parentDeviceId = parent === "null" || parent === ""
                ? null
                : parent === undefined
                    ? undefined
                    : parent;
            if (!req.auth)
                return;
            if (req.auth.kind === "api_key") {
                const { getEnvironmentById } = await import("../environment/services.js");
                const env = await getEnvironmentById(environmentId);
                if (!env) {
                    res.status(404).json({ error: "not_found" });
                    return;
                }
                const rows = await listDevicesByEnvironmentId(environmentId, parentDeviceId);
                res.json({ devices: rows.map(devicePublic) });
                return;
            }
            const rows = await listDevicesInEnvironment(environmentId, req.auth.hiveUserId, parentDeviceId);
            res.json({ devices: rows.map(devicePublic) });
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function postDevice(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            if (req.auth?.kind !== "jwt") {
                res.status(403).json({ error: "forbidden", message: "JWT required" });
                return;
            }
            const environmentId = pathParam(req.params.environmentId);
            const { deviceType, deviceId, displayName, parentDeviceId } = req.body || {};
            if (!deviceType || !deviceId) {
                res.status(400).json({
                    error: "invalid_request",
                    message: "deviceType and deviceId required",
                });
                return;
            }
            const row = await createDevice({
                environmentId,
                userId: req.auth.hiveUserId,
                deviceType,
                deviceId,
                displayName: displayName ?? null,
                parentDeviceId: parentDeviceId ?? null,
            });
            res.status(201).json(devicePublic(row));
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function getDeviceById(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            const logicalDeviceId = pathParam(req.params.deviceId);
            if (!req.auth)
                return;
            if (req.auth.kind === "api_key") {
                const row = await getDeviceRowByEnvAndDeviceId(environmentId, logicalDeviceId);
                if (!row) {
                    res.status(404).json({ error: "not_found" });
                    return;
                }
                res.json(devicePublic(row));
                return;
            }
            const row = await getDeviceByEnvAndDeviceId(environmentId, logicalDeviceId, req.auth.hiveUserId);
            if (!row) {
                res.status(404).json({ error: "not_found" });
                return;
            }
            res.json(devicePublic(row));
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function patchDevice(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            if (req.auth?.kind !== "jwt") {
                res.status(403).json({ error: "forbidden", message: "JWT required" });
                return;
            }
            const environmentId = pathParam(req.params.environmentId);
            const logicalDeviceId = pathParam(req.params.deviceId);
            const { deviceType, deviceId, displayName, parentDeviceId } = req.body || {};
            const row = await updateDeviceByEnvAndDeviceId(environmentId, logicalDeviceId, req.auth.hiveUserId, {
                deviceType,
                deviceId,
                displayName,
                parentDeviceId,
            });
            if (!row) {
                res.status(404).json({ error: "not_found" });
                return;
            }
            res.json(devicePublic(row));
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function destroyDevice(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            if (req.auth?.kind !== "jwt") {
                res.status(403).json({ error: "forbidden", message: "JWT required" });
                return;
            }
            const environmentId = pathParam(req.params.environmentId);
            const logicalDeviceId = pathParam(req.params.deviceId);
            const ok = await deleteDeviceByEnvAndDeviceId(environmentId, logicalDeviceId, req.auth.hiveUserId);
            if (!ok) {
                res.status(404).json({ error: "not_found" });
                return;
            }
            res.status(204).send();
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
//# sourceMappingURL=controller.js.map