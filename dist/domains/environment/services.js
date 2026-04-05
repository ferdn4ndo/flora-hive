import { and, eq } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { getDatabase } from "../../db/client.js";
import { environmentMembers, environments, hiveUsers, } from "../../db/schema.js";
import { canReadRole, canWriteRole } from "./rbac.js";
export { canReadRole, canWriteRole };
export class ForbiddenError extends Error {
    constructor(message = "Forbidden") {
        super(message);
        this.name = "ForbiddenError";
    }
}
export class NotFoundError extends Error {
    constructor(message = "Not found") {
        super(message);
        this.name = "NotFoundError";
    }
}
export async function listAllEnvironments() {
    const db = getDatabase();
    return db.select().from(environments);
}
export async function listEnvironmentsForUser(userId) {
    const db = getDatabase();
    return db
        .select({ env: environments, role: environmentMembers.role })
        .from(environmentMembers)
        .innerJoin(environments, eq(environmentMembers.environmentId, environments.id))
        .where(eq(environmentMembers.userId, userId));
}
export async function getMembership(environmentId, userId) {
    const db = getDatabase();
    const rows = await db
        .select()
        .from(environmentMembers)
        .where(and(eq(environmentMembers.environmentId, environmentId), eq(environmentMembers.userId, userId)))
        .limit(1);
    return rows[0] ?? null;
}
export async function requireEnvAccess(environmentId, userId, needWrite) {
    const m = await getMembership(environmentId, userId);
    if (!m || !canReadRole(m.role)) {
        throw new ForbiddenError("No access to this environment");
    }
    if (needWrite && !canWriteRole(m.role)) {
        throw new ForbiddenError("Editor role required");
    }
    return m;
}
export async function createEnvironment(input) {
    const db = getDatabase();
    const now = new Date().toISOString();
    const id = randomUUID();
    await db.insert(environments).values({
        id,
        name: input.name,
        description: input.description ?? null,
        createdAt: now,
        updatedAt: now,
    });
    await db.insert(environmentMembers).values({
        environmentId: id,
        userId: input.creatorUserId,
        role: "editor",
    });
    return getEnvironmentById(id);
}
export async function getEnvironmentById(id) {
    const db = getDatabase();
    const rows = await db
        .select()
        .from(environments)
        .where(eq(environments.id, id))
        .limit(1);
    return rows[0] ?? null;
}
export async function updateEnvironment(id, patch) {
    const db = getDatabase();
    const now = new Date().toISOString();
    await db
        .update(environments)
        .set({
        ...(patch.name !== undefined ? { name: patch.name } : {}),
        ...(patch.description !== undefined
            ? { description: patch.description }
            : {}),
        updatedAt: now,
    })
        .where(eq(environments.id, id));
    return getEnvironmentById(id);
}
export async function deleteEnvironment(id) {
    const db = getDatabase();
    await db.delete(environments).where(eq(environments.id, id));
}
export async function listMembers(environmentId) {
    const db = getDatabase();
    return db
        .select({
        userId: environmentMembers.userId,
        role: environmentMembers.role,
        username: hiveUsers.username,
        authUuid: hiveUsers.authUuid,
    })
        .from(environmentMembers)
        .innerJoin(hiveUsers, eq(environmentMembers.userId, hiveUsers.id))
        .where(eq(environmentMembers.environmentId, environmentId));
}
export async function upsertMember(input) {
    const db = getDatabase();
    await db
        .insert(environmentMembers)
        .values({
        environmentId: input.environmentId,
        userId: input.userId,
        role: input.role,
    })
        .onConflictDoUpdate({
        target: [
            environmentMembers.environmentId,
            environmentMembers.userId,
        ],
        set: { role: input.role },
    });
}
export async function removeMember(environmentId, userId) {
    const db = getDatabase();
    await db
        .delete(environmentMembers)
        .where(and(eq(environmentMembers.environmentId, environmentId), eq(environmentMembers.userId, userId)));
}
export async function getHiveUserById(userId) {
    const db = getDatabase();
    const rows = await db
        .select()
        .from(hiveUsers)
        .where(eq(hiveUsers.id, userId))
        .limit(1);
    return rows[0] ?? null;
}
//# sourceMappingURL=services.js.map