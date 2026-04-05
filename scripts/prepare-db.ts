/**
 * Local / CI: same steps as Docker entrypoint (`src/containerPrepare.ts`).
 * Ensures Postgres (optional root bootstrap), Hive DDL, optional uServer-Auth.
 */
import "dotenv/config";
import { config } from "../src/config.js";
import { closeDatabase, initDatabase } from "../src/db/client.js";
import { bootstrapPostgresRolesIfConfigured } from "../src/bootstrap/postgresRoles.js";
import { ensureUserverAuthBootstrap } from "../src/bootstrap/userverAuthBootstrap.js";

if (!process.env.MQTT_URL) {
  process.env.MQTT_URL = "mqtt://127.0.0.1:1883";
}

async function main(): Promise<void> {
  await bootstrapPostgresRolesIfConfigured(config.postgres);
  await initDatabase(config.postgres);
  await closeDatabase();
  await ensureUserverAuthBootstrap();
  console.log("prepare-db: ok");
}

main().catch((err: unknown) => {
  console.error(err);
  process.exit(1);
});
