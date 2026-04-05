import pg from "pg";
import type { PostgresConfig } from "../db/client.js";

function quoteIdent(ident: string): string {
  return `"${ident.replace(/"/g, '""')}"`;
}

function escapeSqlString(s: string): string {
  return s.replace(/'/g, "''");
}

/**
 * When POSTGRES_ROOT_USER + POSTGRES_ROOT_PASS are set, ensures POSTGRES_DB and
 * POSTGRES_USER exist (userver-filemgr setup.sh style). No-op otherwise.
 */
export async function bootstrapPostgresRolesIfConfigured(
  appPostgres: PostgresConfig
): Promise<void> {
  const rootUser = process.env.POSTGRES_ROOT_USER?.trim();
  const rootPass = process.env.POSTGRES_ROOT_PASS;
  if (!rootUser || rootPass === undefined || rootPass === "") {
    console.log(
      "bootstrap: skipping database/role creation (POSTGRES_ROOT_USER / POSTGRES_ROOT_PASS unset)"
    );
    return;
  }

  const adminDb = process.env.POSTGRES_ADMIN_DATABASE?.trim() || "postgres";
  const {
    host,
    port,
    user: appUser,
    password: appPassOpt,
    database: appDb,
  } = appPostgres;
  const appPass = appPassOpt ?? "";

  const rootPool = new pg.Pool({
    host,
    port,
    user: rootUser,
    password: rootPass,
    database: adminDb,
    max: 1,
  });

  const passLiteral = escapeSqlString(appPass);

  try {
    const dbRow = await rootPool.query(
      "SELECT 1 FROM pg_database WHERE datname = $1",
      [appDb]
    );
    if (dbRow.rowCount === 0) {
      await rootPool.query(`CREATE DATABASE ${quoteIdent(appDb)}`);
      console.log(`bootstrap: created database ${appDb}`);
    } else {
      console.log(`bootstrap: database ${appDb} already exists`);
    }

    const roleRow = await rootPool.query(
      "SELECT 1 FROM pg_roles WHERE rolname = $1",
      [appUser]
    );
    if (roleRow.rowCount === 0) {
      await rootPool.query(
        `CREATE USER ${quoteIdent(appUser)} WITH LOGIN PASSWORD '${passLiteral}'`
      );
      console.log(`bootstrap: created role ${appUser}`);
    } else {
      await rootPool.query(
        `ALTER USER ${quoteIdent(appUser)} WITH PASSWORD '${passLiteral}'`
      );
      console.log(`bootstrap: synced password for role ${appUser}`);
    }

    await rootPool.query(
      `REVOKE ALL PRIVILEGES ON DATABASE ${quoteIdent("postgres")} FROM ${quoteIdent(appUser)}`
    );
    await rootPool.query(`ALTER USER ${quoteIdent(appUser)} CREATEDB`);
    await rootPool.query(
      `GRANT ALL PRIVILEGES ON DATABASE ${quoteIdent(appDb)} TO ${quoteIdent(appUser)}`
    );
  } finally {
    await rootPool.end();
  }

  const onAppDb = new pg.Pool({
    host,
    port,
    user: rootUser,
    password: rootPass,
    database: appDb,
    max: 1,
  });
  try {
    await onAppDb.query(
      `GRANT USAGE, CREATE ON SCHEMA public TO ${quoteIdent(appUser)}`
    );
    console.log(`bootstrap: granted USAGE, CREATE on public to ${appUser}`);
  } finally {
    await onAppDb.end();
  }
}
