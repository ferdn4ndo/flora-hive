import express from "express";
import { requireApiKey } from "./auth.js";
import { getMqttState, listDevices, publishMqtt } from "./mqttClient.js";

export function createApp() {
  const app = express();
  app.disable("x-powered-by");
  app.use(express.json({ limit: "512kb" }));

  app.get("/healthz", (_req, res) => {
    res.json({ status: "ok", service: "flora-hive" });
  });

  const v1 = express.Router();
  v1.use(requireApiKey);

  v1.get("/whoami", (_req, res) => {
    res.json({ ok: true, role: "hive" });
  });

  v1.get("/mqtt/connection", (_req, res) => {
    res.json(getMqttState());
  });

  v1.get("/devices", (req, res) => {
    const raw = req.query.include_offline;
    const includeOffline =
      raw === "1" ||
      raw === "true" ||
      raw === "yes";
    res.json({ devices: listDevices({ includeOffline }) });
  });

  v1.post("/mqtt/publish", async (req, res) => {
    const { topic, payload, qos, retain } = req.body || {};
    if (topic === undefined || topic === null || String(topic).trim() === "") {
      res.status(400).json({ error: "invalid_request", message: "topic is required" });
      return;
    }
    try {
      const result = await publishMqtt({ topic, payload, qos, retain });
      res.status(202).json({ ok: true, ...result });
    } catch (err) {
      const msg = err?.message || String(err);
      if (msg.includes("not connected")) {
        res.status(503).json({
          error: "mqtt_unavailable",
          message: msg,
        });
        return;
      }
      if (msg.includes("qos") || msg.includes("topic")) {
        res.status(400).json({ error: "invalid_request", message: msg });
        return;
      }
      res.status(500).json({ error: "publish_failed", message: msg });
    }
  });

  app.use("/v1", v1);

  app.use((_req, res) => {
    res.status(404).json({ error: "not_found" });
  });

  return app;
}
