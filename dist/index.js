import { createServer } from "node:http";
import { config } from "./config.js";
import { createApp } from "./app.js";
import { closeDatabase, initDatabase } from "./db/client.js";
import { connectMqtt, disconnectMqtt, getMqttState, initMqttService, onMqtt, } from "./domains/mqtt/services.js";
async function main() {
    await initDatabase(config.postgres);
    initMqttService(config);
    const app = createApp();
    const server = createServer(app);
    function log(level, msg, extra) {
        console.log(JSON.stringify({
            ts: new Date().toISOString(),
            level,
            msg,
            ...extra,
        }));
    }
    connectMqtt();
    onMqtt("connect", () => log("info", "mqtt_connected"));
    onMqtt("error", (err) => log("error", "mqtt_error", { err: String(err) }));
    onMqtt("reconnect", () => {
        const st = getMqttState();
        log("warn", "mqtt_reconnecting", {
            broker: st.url,
            lastError: st.lastError,
        });
    });
    server.listen(config.port, () => {
        log("info", "hive_listening", { port: config.port });
    });
    async function shutdown(signal) {
        log("info", "shutdown", { signal });
        await new Promise((resolve) => server.close(() => resolve()));
        await disconnectMqtt();
        await closeDatabase();
        process.exit(0);
    }
    process.on("SIGINT", () => void shutdown("SIGINT"));
    process.on("SIGTERM", () => void shutdown("SIGTERM"));
}
main().catch((err) => {
    console.error(err);
    const e = err;
    if (e.code === "28P01") {
        console.error("PostgreSQL password authentication failed. Set POSTGRES_PASS (or POSTGRES_PASSWORD) in .env to the same password as the database role POSTGRES_USER, or leave both unset only if that role has no password.");
    }
    process.exit(1);
});
//# sourceMappingURL=index.js.map