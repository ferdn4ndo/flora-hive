import { drizzle } from "drizzle-orm/node-postgres";
import { Pool } from "pg";
import * as schema from "./schema.js";
import { ensurePostgresSchema } from "./bootstrapPg.js";
let _pool = null;
let _db = null;
export function getPool(cfg) {
    if (_pool)
        return _pool;
    const config = {
        host: cfg.host,
        port: cfg.port,
        user: cfg.user,
        database: cfg.database,
        max: 10,
        ...(cfg.password !== undefined && cfg.password !== ""
            ? { password: cfg.password }
            : {}),
    };
    _pool = new Pool(config);
    return _pool;
}
export function getDatabase() {
    if (!_pool) {
        throw new Error("Database pool not initialized; call getPool() first");
    }
    if (!_db) {
        _db = drizzle(_pool, { schema });
    }
    return _db;
}
export async function initDatabase(cfg) {
    const pool = getPool(cfg);
    await ensurePostgresSchema(pool);
    getDatabase();
}
export async function closeDatabase() {
    _db = null;
    if (_pool) {
        await _pool.end();
        _pool = null;
    }
}
//# sourceMappingURL=client.js.map