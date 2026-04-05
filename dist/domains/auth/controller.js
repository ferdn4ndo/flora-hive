import { config } from "../../config.js";
import { findHiveUserByAuthUuid, upsertHiveUserFromMe, } from "../user/services.js";
import { hiveUserPublic } from "../user/views.js";
import { requireAuth } from "./middleware.js";
import { userverChangePassword, userverLogin, userverLogout, userverMe, userverRefresh, userverRegister, } from "./services.js";
export async function postLogin(req, res) {
    if (!config.userver.configured) {
        res.status(503).json({ error: "auth_unavailable", message: "uServer-Auth not configured" });
        return;
    }
    const { username, password } = req.body || {};
    if (!username || !password) {
        res.status(400).json({ error: "invalid_request", message: "username and password required" });
        return;
    }
    const result = await userverLogin({ username, password });
    if (!result.ok) {
        res.status(result.status).json({ error: "login_failed", message: result.message });
        return;
    }
    res.json(result.data);
}
export async function postRegister(req, res) {
    if (!config.userver.configured) {
        res.status(503).json({ error: "auth_unavailable", message: "uServer-Auth not configured" });
        return;
    }
    const { username, password, is_admin } = req.body || {};
    if (!username || !password) {
        res.status(400).json({ error: "invalid_request", message: "username and password required" });
        return;
    }
    const result = await userverRegister({
        username,
        password,
        ...(typeof is_admin === "boolean" ? { is_admin } : {}),
    });
    if (!result.ok) {
        res.status(result.status).json({ error: "register_failed", message: result.message });
        return;
    }
    res.status(201).json(result.data);
}
export async function postRefresh(req, res) {
    if (!config.userver.configured) {
        res.status(503).json({ error: "auth_unavailable", message: "uServer-Auth not configured" });
        return;
    }
    const { refresh_token } = req.body || {};
    if (!refresh_token || typeof refresh_token !== "string") {
        res.status(400).json({ error: "invalid_request", message: "refresh_token required" });
        return;
    }
    const result = await userverRefresh(refresh_token);
    if (!result.ok) {
        res.status(result.status).json({ error: "refresh_failed", message: result.message });
        return;
    }
    res.json(result.data);
}
export async function postLogout(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            if (req.auth?.kind !== "jwt") {
                res.status(400).json({ error: "invalid_request", message: "JWT session required" });
                return;
            }
            await userverLogout(req.auth.accessToken);
            res.status(204).send();
        }
        catch (e) {
            next(e);
        }
    });
}
export async function getMe(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            if (req.auth?.kind === "api_key") {
                res.json({ kind: "api_key", role: "service" });
                return;
            }
            if (req.auth?.kind !== "jwt") {
                res.status(401).json({ error: "unauthorized" });
                return;
            }
            const me = await userverMe(req.auth.accessToken);
            if (!me.ok) {
                res.status(me.status).json({ error: "auth_invalid", message: me.message });
                return;
            }
            await upsertHiveUserFromMe(me.data);
            const dbUser = await findHiveUserByAuthUuid(me.data.uuid);
            if (!dbUser) {
                res.status(500).json({ error: "internal_error", message: "Hive user sync failed" });
                return;
            }
            res.json({
                userver: me.data,
                hiveUser: hiveUserPublic(dbUser),
            });
        }
        catch (e) {
            next(e);
        }
    });
}
export async function patchPassword(req, res, next) {
    requireAuth(req, res, async () => {
        try {
            if (req.auth?.kind !== "jwt") {
                res.status(400).json({ error: "invalid_request", message: "Bearer access token required" });
                return;
            }
            const { current_password, new_password } = req.body || {};
            if (!current_password || !new_password) {
                res.status(400).json({
                    error: "invalid_request",
                    message: "current_password and new_password required",
                });
                return;
            }
            const result = await userverChangePassword(req.auth.accessToken, {
                current_password,
                new_password,
            });
            if (!result.ok) {
                res.status(result.status).json({ error: "password_change_failed", message: result.message });
                return;
            }
            res.json({ ok: true, message: "Password updated." });
        }
        catch (e) {
            next(e);
        }
    });
}
//# sourceMappingURL=controller.js.map