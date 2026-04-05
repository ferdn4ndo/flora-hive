/**
 * Verifies PostgreSQL connectivity and idempotent schema bootstrap (CI / local).
 * Requires the same POSTGRES_* env vars as the app (and MQTT_URL for config load).
 */
import "dotenv/config";
import { config } from "../src/config.js";
import { closeDatabase, initDatabase } from "../src/db/client.js";

if (!process.env.MQTT_URL) {
  process.env.MQTT_URL = "mqtt://127.0.0.1:1883";
}

await initDatabase(config.postgres);
await closeDatabase();
console.log("db-smoke: ok");
