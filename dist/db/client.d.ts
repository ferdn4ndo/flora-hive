import { type NodePgDatabase } from "drizzle-orm/node-postgres";
import { Pool } from "pg";
import * as schema from "./schema.js";
export type PostgresConfig = {
    host: string;
    port: number;
    user: string;
    /** When unset, the driver omits `password` (trust / no-password roles). */
    password?: string;
    database: string;
};
export declare function getPool(cfg: PostgresConfig): Pool;
export declare function getDatabase(): NodePgDatabase<typeof schema>;
export declare function initDatabase(cfg: PostgresConfig): Promise<void>;
export declare function closeDatabase(): Promise<void>;
//# sourceMappingURL=client.d.ts.map