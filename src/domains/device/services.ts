import { and, eq, inArray, isNull } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { getDatabase } from "../../db/client.js";
import { devices } from "../../db/schema.js";
import { requireEnvAccess } from "../environment/services.js";

export async function listDevicesByEnvironmentId(
  environmentId: string,
  parentDeviceId: string | null | undefined
) {
  const db = getDatabase();
  if (parentDeviceId === undefined) {
    return db
      .select()
      .from(devices)
      .where(eq(devices.environmentId, environmentId));
  }
  if (parentDeviceId === null) {
    return db
      .select()
      .from(devices)
      .where(
        and(
          eq(devices.environmentId, environmentId),
          isNull(devices.parentDeviceId)
        )
      );
  }
  return db
    .select()
    .from(devices)
    .where(
      and(
        eq(devices.environmentId, environmentId),
        eq(devices.parentDeviceId, parentDeviceId)
      )
    );
}

export async function listDevicesInEnvironment(
  environmentId: string,
  userId: string,
  parentDeviceId: string | null | undefined
) {
  await requireEnvAccess(environmentId, userId, false);
  return listDevicesByEnvironmentId(environmentId, parentDeviceId);
}

/** Catalog row ids for MQTT filtering (`devices.id` = first topic segment under prefix). */
export async function listDeviceRowIdsForEnvironments(
  environmentIds: string[]
): Promise<string[]> {
  if (environmentIds.length === 0) return [];
  const db = getDatabase();
  const rows = await db
    .select({ id: devices.id })
    .from(devices)
    .where(inArray(devices.environmentId, environmentIds));
  return rows.map((r) => r.id);
}

export async function createDevice(input: {
  environmentId: string;
  userId: string;
  deviceType: string;
  deviceId: string;
  displayName?: string | null;
  parentDeviceId?: string | null;
}) {
  await requireEnvAccess(input.environmentId, input.userId, true);
  const db = getDatabase();
  if (input.parentDeviceId) {
    const parents = await db
      .select()
      .from(devices)
      .where(eq(devices.id, input.parentDeviceId))
      .limit(1);
    const parent = parents[0];
    if (!parent || parent.environmentId !== input.environmentId) {
      throw new Error("Invalid parent device");
    }
  }
  const now = new Date().toISOString();
  const id = randomUUID();
  await db.insert(devices).values({
    id,
    environmentId: input.environmentId,
    parentDeviceId: input.parentDeviceId ?? null,
    deviceType: input.deviceType,
    deviceId: input.deviceId,
    displayName: input.displayName ?? null,
    createdAt: now,
    updatedAt: now,
  });
  return getDeviceRowById(id);
}

export async function getDeviceRowById(deviceRowId: string) {
  const db = getDatabase();
  const rows = await db
    .select()
    .from(devices)
    .where(eq(devices.id, deviceRowId))
    .limit(1);
  return rows[0] ?? null;
}

export async function getDeviceRowByEnvAndDeviceId(
  environmentId: string,
  deviceId: string
) {
  const db = getDatabase();
  const rows = await db
    .select()
    .from(devices)
    .where(
      and(
        eq(devices.environmentId, environmentId),
        eq(devices.deviceId, deviceId)
      )
    )
    .limit(1);
  return rows[0] ?? null;
}

export async function getDevice(deviceRowId: string, userId: string) {
  const row = await getDeviceRowById(deviceRowId);
  if (!row) return null;
  await requireEnvAccess(row.environmentId, userId, false);
  return row;
}

export async function getDeviceByEnvAndDeviceId(
  environmentId: string,
  logicalDeviceId: string,
  userId: string
) {
  const row = await getDeviceRowByEnvAndDeviceId(environmentId, logicalDeviceId);
  if (!row) return null;
  await requireEnvAccess(row.environmentId, userId, false);
  return row;
}

export async function updateDevice(
  deviceRowId: string,
  userId: string,
  patch: {
    deviceType?: string;
    deviceId?: string;
    displayName?: string | null;
    parentDeviceId?: string | null;
  }
) {
  const db = getDatabase();
  const row = await getDeviceRowById(deviceRowId);
  if (!row) return null;
  await requireEnvAccess(row.environmentId, userId, true);
  const now = new Date().toISOString();
  await db
    .update(devices)
    .set({
      ...(patch.deviceType !== undefined ? { deviceType: patch.deviceType } : {}),
      ...(patch.deviceId !== undefined ? { deviceId: patch.deviceId } : {}),
      ...(patch.displayName !== undefined ? { displayName: patch.displayName } : {}),
      ...(patch.parentDeviceId !== undefined
        ? { parentDeviceId: patch.parentDeviceId }
        : {}),
      updatedAt: now,
    })
    .where(eq(devices.id, deviceRowId));
  return getDeviceRowById(deviceRowId);
}

export async function deleteDevice(deviceRowId: string, userId: string) {
  const row = await getDeviceRowById(deviceRowId);
  if (!row) return false;
  await requireEnvAccess(row.environmentId, userId, true);
  const db = getDatabase();
  await db.delete(devices).where(eq(devices.id, deviceRowId));
  return true;
}

export async function updateDeviceByEnvAndDeviceId(
  environmentId: string,
  logicalDeviceId: string,
  userId: string,
  patch: {
    deviceType?: string;
    deviceId?: string;
    displayName?: string | null;
    parentDeviceId?: string | null;
  }
) {
  const row = await getDeviceRowByEnvAndDeviceId(environmentId, logicalDeviceId);
  if (!row) return null;
  return updateDevice(row.id, userId, patch);
}

export async function deleteDeviceByEnvAndDeviceId(
  environmentId: string,
  logicalDeviceId: string,
  userId: string
) {
  const row = await getDeviceRowByEnvAndDeviceId(environmentId, logicalDeviceId);
  if (!row) return false;
  return deleteDevice(row.id, userId);
}
