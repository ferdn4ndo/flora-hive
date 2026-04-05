import type { HiveUser } from "../../db/schema.js";

export function hiveUserPublic(u: HiveUser) {
  return {
    id: u.id,
    authUuid: u.authUuid,
    username: u.username,
    systemName: u.systemName,
    updatedAt: u.updatedAt,
  };
}
