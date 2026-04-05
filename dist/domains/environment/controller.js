import { pathParam } from "../../http/params.js";
import { requireAuth } from "../auth/middleware.js";
import { findHiveUserByAuthUuid } from "../user/services.js";
import { environmentPublic } from "./views.js";
import { ForbiddenError, createEnvironment, deleteEnvironment, getEnvironmentById, getMembership, listAllEnvironments, listEnvironmentsForUser, listMembers, removeMember, requireEnvAccess, updateEnvironment, upsertMember, } from "./services.js";
function handleErr(res, next, e) {
    if (e instanceof ForbiddenError) {
        res.status(403).json({ error: "forbidden", message: e.message });
        return;
    }
    next(e);
}
export function listEnvironments(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            if (!req.auth)
                return;
            if (req.auth.kind === "api_key") {
                const all = await listAllEnvironments();
                res.json({
                    environments: all.map((e) => environmentPublic(e)),
                });
                return;
            }
            const rows = await listEnvironmentsForUser(req.auth.hiveUserId);
            res.json({
                environments: rows.map((r) => environmentPublic(r.env, r.role)),
            });
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function postEnvironment(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            if (!req.auth)
                return;
            if (req.auth.kind === "api_key") {
                res.status(403).json({
                    error: "forbidden",
                    message: "Create environment with a user JWT, not an API key",
                });
                return;
            }
            const { name, description } = req.body || {};
            if (!name) {
                res.status(400).json({
                    error: "invalid_request",
                    message: "name required",
                });
                return;
            }
            const env = await createEnvironment({
                name,
                description: description ?? null,
                creatorUserId: req.auth.hiveUserId,
            });
            res.status(201).json(environmentPublic(env));
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function getEnvironment(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            const row = await getEnvironmentById(environmentId);
            if (!row) {
                res.status(404).json({ error: "not_found" });
                return;
            }
            if (req.auth?.kind === "api_key") {
                res.json(environmentPublic(row));
                return;
            }
            if (req.auth?.kind !== "jwt")
                return;
            await requireEnvAccess(environmentId, req.auth.hiveUserId, false);
            const m = await getMembership(environmentId, req.auth.hiveUserId);
            res.json(environmentPublic(row, m?.role));
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function patchEnvironment(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            if (req.auth?.kind !== "jwt") {
                res.status(403).json({ error: "forbidden", message: "JWT required" });
                return;
            }
            await requireEnvAccess(environmentId, req.auth.hiveUserId, true);
            const { name, description } = req.body || {};
            const row = await getEnvironmentById(environmentId);
            if (!row) {
                res.status(404).json({ error: "not_found" });
                return;
            }
            const updated = await updateEnvironment(environmentId, {
                name,
                description,
            });
            res.json(environmentPublic(updated));
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function destroyEnvironment(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            if (req.auth?.kind !== "jwt") {
                res.status(403).json({ error: "forbidden", message: "JWT required" });
                return;
            }
            await requireEnvAccess(environmentId, req.auth.hiveUserId, true);
            const row = await getEnvironmentById(environmentId);
            if (!row) {
                res.status(404).json({ error: "not_found" });
                return;
            }
            await deleteEnvironment(environmentId);
            res.status(204).send();
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function listEnvironmentMembers(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            if (req.auth?.kind === "api_key") {
                const row = await getEnvironmentById(environmentId);
                if (!row) {
                    res.status(404).json({ error: "not_found" });
                    return;
                }
                res.json({ members: await listMembers(environmentId) });
                return;
            }
            if (req.auth?.kind !== "jwt")
                return;
            await requireEnvAccess(environmentId, req.auth.hiveUserId, false);
            res.json({ members: await listMembers(environmentId) });
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function postEnvironmentMember(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            if (req.auth?.kind !== "jwt") {
                res.status(403).json({ error: "forbidden", message: "JWT required" });
                return;
            }
            await requireEnvAccess(environmentId, req.auth.hiveUserId, true);
            const { authUserUuid, role } = req.body || {};
            if (!authUserUuid || (role !== "viewer" && role !== "editor")) {
                res.status(400).json({
                    error: "invalid_request",
                    message: "authUserUuid and role (viewer|editor) required",
                });
                return;
            }
            const target = await findHiveUserByAuthUuid(authUserUuid);
            if (!target) {
                res.status(404).json({
                    error: "not_found",
                    message: "Hive user not found for that auth UUID; the user must call GET /v1/auth/me once after registering.",
                });
                return;
            }
            await upsertMember({
                environmentId,
                userId: target.id,
                role: role,
            });
            res.status(201).json({
                environmentId,
                userId: target.id,
                role,
            });
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function patchEnvironmentMember(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            const userId = pathParam(req.params.userId);
            if (req.auth?.kind !== "jwt") {
                res.status(403).json({ error: "forbidden", message: "JWT required" });
                return;
            }
            await requireEnvAccess(environmentId, req.auth.hiveUserId, true);
            const { role } = req.body || {};
            if (role !== "viewer" && role !== "editor") {
                res.status(400).json({ error: "invalid_request", message: "role required" });
                return;
            }
            await upsertMember({
                environmentId,
                userId,
                role: role,
            });
            res.json({ environmentId, userId, role });
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
export function deleteEnvironmentMember(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            const environmentId = pathParam(req.params.environmentId);
            const userId = pathParam(req.params.userId);
            if (req.auth?.kind !== "jwt") {
                res.status(403).json({ error: "forbidden", message: "JWT required" });
                return;
            }
            await requireEnvAccess(environmentId, req.auth.hiveUserId, true);
            await removeMember(environmentId, userId);
            res.status(204).send();
        }
        catch (e) {
            handleErr(res, next, e);
        }
    });
}
//# sourceMappingURL=controller.js.map