import type { Pool } from "pg";
/**
 * Idempotent DDL for Flora Hive on PostgreSQL (userver-datamgr / same host pattern as userver-filemgr).
 */
export declare function ensurePostgresSchema(pool: Pool): Promise<void>;
//# sourceMappingURL=bootstrapPg.d.ts.map