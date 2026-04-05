import type { MemberRole } from "../../db/schema.js";

export function canReadRole(role: MemberRole | undefined): boolean {
  return role === "viewer" || role === "editor";
}

export function canWriteRole(role: MemberRole | undefined): boolean {
  return role === "editor";
}
