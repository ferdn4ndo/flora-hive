import { relations } from "drizzle-orm";
import { pgTable, text, primaryKey, uniqueIndex, } from "drizzle-orm/pg-core";
export const hiveUsers = pgTable("hive_users", {
    id: text("id").primaryKey(),
    authUuid: text("auth_uuid").notNull().unique(),
    username: text("username").notNull(),
    systemName: text("system_name").notNull(),
    updatedAt: text("updated_at").notNull(),
});
export const environments = pgTable("environments", {
    id: text("id").primaryKey(),
    name: text("name").notNull(),
    description: text("description"),
    createdAt: text("created_at").notNull(),
    updatedAt: text("updated_at").notNull(),
});
export const environmentMembers = pgTable("environment_members", {
    environmentId: text("environment_id")
        .notNull()
        .references(() => environments.id, { onDelete: "cascade" }),
    userId: text("user_id")
        .notNull()
        .references(() => hiveUsers.id, { onDelete: "cascade" }),
    role: text("role", { enum: ["viewer", "editor"] }).notNull(),
}, (t) => ({
    pk: primaryKey({ columns: [t.environmentId, t.userId] }),
}));
export const devices = pgTable("devices", {
    id: text("id").primaryKey(),
    environmentId: text("environment_id")
        .notNull()
        .references(() => environments.id, { onDelete: "cascade" }),
    parentDeviceId: text("parent_device_id"),
    deviceType: text("device_type").notNull(),
    deviceId: text("device_id").notNull(),
    displayName: text("display_name"),
    createdAt: text("created_at").notNull(),
    updatedAt: text("updated_at").notNull(),
}, (t) => ({
    envDeviceUnique: uniqueIndex("devices_env_device_unique").on(t.environmentId, t.deviceId),
}));
export const hiveUsersRelations = relations(hiveUsers, ({ many }) => ({
    memberships: many(environmentMembers),
}));
export const environmentsRelations = relations(environments, ({ many }) => ({
    members: many(environmentMembers),
    devices: many(devices),
}));
export const environmentMembersRelations = relations(environmentMembers, ({ one }) => ({
    environment: one(environments, {
        fields: [environmentMembers.environmentId],
        references: [environments.id],
    }),
    user: one(hiveUsers, {
        fields: [environmentMembers.userId],
        references: [hiveUsers.id],
    }),
}));
export const devicesRelations = relations(devices, ({ one, many }) => ({
    environment: one(environments, {
        fields: [devices.environmentId],
        references: [environments.id],
    }),
    parent: one(devices, {
        fields: [devices.parentDeviceId],
        references: [devices.id],
        relationName: "deviceTree",
    }),
    children: many(devices, { relationName: "deviceTree" }),
}));
//# sourceMappingURL=schema.js.map