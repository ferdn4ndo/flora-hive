import { drizzle, type NodePgDatabase } from "drizzle-orm/node-postgres";
import { Pool, type PoolConfig } from "pg";
import * as schema from "./schema.js";
import { ensurePostgresSchema } from "./bootstrapPg.js";

export type PostgresConfig = {
  host: string;
  port: number;
  user: string;
  /** When unset, the driver omits `password` (trust / no-password roles). */
  password?: string;
  database: string;
};

let _pool: Pool | null = null;
let _db: NodePgDatabase<typeof schema> | null = null;

export function getPool(cfg: PostgresConfig): Pool {
  if (_pool) return _pool;
  const config: PoolConfig = {
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

export function getDatabase(): NodePgDatabase<typeof schema> {
  if (!_pool) {
    throw new Error("Database pool not initialized; call getPool() first");
  }
  if (!_db) {
    _db = drizzle(_pool, { schema });
  }
  return _db;
}

export async function initDatabase(cfg: PostgresConfig): Promise<void> {
  const pool = getPool(cfg);
  await ensurePostgresSchema(pool);
  getDatabase();
}

export async function closeDatabase(): Promise<void> {
  _db = null;
  if (_pool) {
    await _pool.end();
    _pool = null;
  }
}
