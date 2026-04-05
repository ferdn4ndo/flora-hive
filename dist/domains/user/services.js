import { eq } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { getDatabase } from "../../db/client.js";
import { hiveUsers } from "../../db/schema.js";
export async function upsertHiveUserFromMe(me) {
    const db = getDatabase();
    const now = new Date().toISOString();
    const existing = await db
        .select()
        .from(hiveUsers)
        .where(eq(hiveUsers.authUuid, me.uuid))
        .limit(1);
    const row = existing[0];
    if (row) {
        await db
            .update(hiveUsers)
            .set({
            username: me.username,
            systemName: me.system_name,
            updatedAt: now,
        })
            .where(eq(hiveUsers.id, row.id));
        return row.id;
    }
    const id = randomUUID();
    await db.insert(hiveUsers).values({
        id,
        authUuid: me.uuid,
        username: me.username,
        systemName: me.system_name,
        updatedAt: now,
    });
    return id;
}
export async function findHiveUserByAuthUuid(authUuid) {
    const db = getDatabase();
    const rows = await db
        .select()
        .from(hiveUsers)
        .where(eq(hiveUsers.authUuid, authUuid))
        .limit(1);
    return rows[0];
}
//# sourceMappingURL=services.js.map