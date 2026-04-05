import express from "express";
import { config } from "./config.js";
import * as authController from "./domains/auth/controller.js";
import { attachAuthOptional } from "./domains/auth/middleware.js";
import * as deviceController from "./domains/device/controller.js";
import * as envController from "./domains/environment/controller.js";
import { getMqttState, listLiveDevices, normalizeTopic, parseEnvironmentIdFromTopic, publishMqtt, } from "./domains/mqtt/services.js";
import { ForbiddenError, getEnvironmentById, listEnvironmentsForUser, requireEnvAccess, } from "./domains/environment/services.js";
import { requireAuth } from "./domains/auth/middleware.js";
async function allowedEnvironmentIdsForRequest(req) {
    if (!req.auth)
        return [];
    if (req.auth.kind === "api_key")
        return null;
    const rows = await listEnvironmentsForUser(req.auth.hiveUserId);
    return rows.map((r) => r.env.id);
}
async function assertMqttPublishAllowed(req, topic) {
    if (req.auth?.kind === "api_key")
        return;
    if (req.auth?.kind !== "jwt") {
        throw new ForbiddenError("Authentication required");
    }
    const normalized = normalizeTopic(topic, config.topicPrefix);
    const environmentId = parseEnvironmentIdFromTopic(normalized, config.topicPrefix);
    if (!environmentId) {
        throw new ForbiddenError("Cannot derive environment from topic");
    }
    const env = await getEnvironmentById(environmentId);
    if (!env) {
        throw new ForbiddenError("Unknown environment for this topic");
    }
    await requireEnvAccess(env.id, req.auth.hiveUserId, true);
}
export function createApp() {
    const app = express();
    app.disable("x-powered-by");
    app.use(express.json({ limit: "512kb" }));
    app.get("/healthz", (_req, res) => {
        res.set("Cache-Control", "no-store");
        res.status(200).json({ status: "ok", service: "flora-hive" });
    });
    app.head("/healthz", (_req, res) => {
        res.set("Cache-Control", "no-store");
        res.status(200).end();
    });
    app.use("/v1", attachAuthOptional);
    const auth = express.Router();
    auth.post("/login", (req, res, next) => {
        void authController.postLogin(req, res).catch(next);
    });
    auth.post("/register", (req, res, next) => {
        void authController.postRegister(req, res).catch(next);
    });
    auth.post("/refresh", (req, res, next) => {
        void authController.postRefresh(req, res).catch(next);
    });
    auth.post("/logout", authController.postLogout);
    auth.get("/me", authController.getMe);
    auth.patch("/password", authController.patchPassword);
    auth.patch("/reset-password", authController.patchPassword);
    app.use("/v1/auth", auth);
    app.get("/v1/mqtt/connection", requireAuth, (_req, res) => {
        res.json(getMqttState());
    });
    app.get("/v1/mqtt/devices", requireAuth, async (req, res, next) => {
        try {
            const raw = req.query.include_offline;
            const includeOffline = raw === "1" || raw === "true" || raw === "yes";
            const allow = await allowedEnvironmentIdsForRequest(req);
            res.json({
                devices: listLiveDevices({ includeOffline, allowedEnvironmentIds: allow }),
            });
        }
        catch (e) {
            next(e);
        }
    });
    app.post("/v1/mqtt/publish", requireAuth, async (req, res, next) => {
        try {
            const { topic, payload, qos, retain } = req.body || {};
            if (topic === undefined || topic === null || String(topic).trim() === "") {
                res.status(400).json({
                    error: "invalid_request",
                    message: "topic is required",
                });
                return;
            }
            await assertMqttPublishAllowed(req, String(topic));
            const result = await publishMqtt({ topic, payload, qos, retain });
            res.status(202).json({ ok: true, ...result });
        }
        catch (err) {
            const msg = err instanceof Error ? err.message : String(err);
            if (err instanceof ForbiddenError) {
                res.status(403).json({ error: "forbidden", message: err.message });
                return;
            }
            if (msg.includes("not connected")) {
                res.status(503).json({ error: "mqtt_unavailable", message: msg });
                return;
            }
            if (msg.includes("qos") || msg.includes("topic")) {
                res.status(400).json({ error: "invalid_request", message: msg });
                return;
            }
            next(err);
        }
    });
    app.get("/v1/environments", envController.listEnvironments);
    app.post("/v1/environments", envController.postEnvironment);
    app.get("/v1/environments/:environmentId", envController.getEnvironment);
    app.patch("/v1/environments/:environmentId", envController.patchEnvironment);
    app.delete("/v1/environments/:environmentId", envController.destroyEnvironment);
    app.get("/v1/environments/:environmentId/members", envController.listEnvironmentMembers);
    app.post("/v1/environments/:environmentId/members", envController.postEnvironmentMember);
    app.patch("/v1/environments/:environmentId/members/:userId", envController.patchEnvironmentMember);
    app.delete("/v1/environments/:environmentId/members/:userId", envController.deleteEnvironmentMember);
    app.get("/v1/environments/:environmentId/devices", deviceController.listDevices);
    app.post("/v1/environments/:environmentId/devices", deviceController.postDevice);
    app.get("/v1/environments/:environmentId/devices/:deviceId", deviceController.getDeviceById);
    app.patch("/v1/environments/:environmentId/devices/:deviceId", deviceController.patchDevice);
    app.delete("/v1/environments/:environmentId/devices/:deviceId", deviceController.destroyDevice);
    app.use((_req, res) => {
        res.status(404).json({ error: "not_found" });
    });
    app.use((err, _req, res, _next) => {
        console.error(err);
        res.status(500).json({ error: "internal_error" });
    });
    return app;
}
//# sourceMappingURL=app.js.map