import { createServer } from "node:http";
import { config } from "./config.js";
import { createApp } from "./app.js";
import { connectMqtt, disconnectMqtt, onMqtt } from "./mqttClient.js";

const app = createApp();
const server = createServer(app);

function log(level, msg, extra) {
  const line = { ts: new Date().toISOString(), level, msg, ...extra };
  console.log(JSON.stringify(line));
}

connectMqtt();
onMqtt("connect", () => log("info", "mqtt_connected"));
onMqtt("error", (err) => log("error", "mqtt_error", { err: String(err) }));
onMqtt("reconnect", () => log("warn", "mqtt_reconnecting"));

server.listen(config.port, () => {
  log("info", "hive_listening", { port: config.port });
});

async function shutdown(signal) {
  log("info", "shutdown", { signal });
  await new Promise((resolve) => server.close(resolve));
  await disconnectMqtt();
  process.exit(0);
}

process.on("SIGINT", () => shutdown("SIGINT"));
process.on("SIGTERM", () => shutdown("SIGTERM"));
