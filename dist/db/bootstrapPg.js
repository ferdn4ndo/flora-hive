/**
 * Idempotent DDL for Flora Hive on PostgreSQL (userver-datamgr / same host pattern as userver-filemgr).
 */
export async function ensurePostgresSchema(pool) {
    await pool.query(`
CREATE TABLE IF NOT EXISTS hive_users (
  id text PRIMARY KEY NOT NULL,
  auth_uuid text NOT NULL,
  username text NOT NULL,
  system_name text NOT NULL,
  updated_at text NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS hive_users_auth_uuid_unique ON hive_users (auth_uuid);

CREATE TABLE IF NOT EXISTS environments (
  id text PRIMARY KEY NOT NULL,
  name text NOT NULL,
  description text,
  created_at text NOT NULL,
  updated_at text NOT NULL
);

CREATE TABLE IF NOT EXISTS environment_members (
  environment_id text NOT NULL,
  user_id text NOT NULL,
  role text NOT NULL,
  PRIMARY KEY (environment_id, user_id),
  CONSTRAINT environment_members_environment_id_fkey
    FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE,
  CONSTRAINT environment_members_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES hive_users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS devices (
  id text PRIMARY KEY NOT NULL,
  environment_id text NOT NULL,
  parent_device_id text,
  device_type text NOT NULL,
  device_id text NOT NULL,
  display_name text,
  created_at text NOT NULL,
  updated_at text NOT NULL,
  CONSTRAINT devices_environment_id_fkey
    FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE
);
  `);
    await pool.query(`
ALTER TABLE environments DROP COLUMN IF EXISTS mqtt_env_id;
DROP INDEX IF EXISTS environments_mqtt_env_id_unique;
DROP INDEX IF EXISTS devices_env_type_device;
CREATE UNIQUE INDEX IF NOT EXISTS devices_env_device_unique ON devices (environment_id, device_id);
  `);
}
//# sourceMappingURL=bootstrapPg.js.map