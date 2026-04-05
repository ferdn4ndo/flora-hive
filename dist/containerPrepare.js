/**
 * Runs before `index.ts` in Docker: optional Postgres DB/role bootstrap, Hive DDL,
 * optional uServer-Auth bootstrap. Idempotent.
 */
import "dotenv/config";
import { config } from "./config.js";
import { closeDatabase, initDatabase } from "./db/client.js";
import { bootstrapPostgresRolesIfConfigured } from "./bootstrap/postgresRoles.js";
import { ensureUserverAuthBootstrap } from "./bootstrap/userverAuthBootstrap.js";
if (!process.env.MQTT_URL) {
    process.env.MQTT_URL = "mqtt://127.0.0.1:1883";
}
async function main() {
    await bootstrapPostgresRolesIfConfigured(config.postgres);
    await initDatabase(config.postgres);
    await closeDatabase();
    await ensureUserverAuthBootstrap();
    console.log("container-prepare: ok");
}
main().catch((err) => {
    console.error(err);
    const e = err;
    if (e.code === "28P01") {
        console.error("PostgreSQL password authentication failed. After POSTGRES_ROOT_* creates the role, POSTGRES_PASS must match. Set POSTGRES_PASS or POSTGRES_PASSWORD.");
    }
    process.exit(1);
});
//# sourceMappingURL=containerPrepare.js.map