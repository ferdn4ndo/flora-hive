import type { PostgresConfig } from "../db/client.js";
/**
 * When POSTGRES_ROOT_USER + POSTGRES_ROOT_PASS are set, ensures POSTGRES_DB and
 * POSTGRES_USER exist (userver-filemgr setup.sh style). No-op otherwise.
 */
export declare function bootstrapPostgresRolesIfConfigured(appPostgres: PostgresConfig): Promise<void>;
//# sourceMappingURL=postgresRoles.d.ts.map