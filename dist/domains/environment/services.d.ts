import { type MemberRole } from "../../db/schema.js";
import { canReadRole, canWriteRole } from "./rbac.js";
export { canReadRole, canWriteRole };
export declare class ForbiddenError extends Error {
    constructor(message?: string);
}
export declare class NotFoundError extends Error {
    constructor(message?: string);
}
export declare function listAllEnvironments(): Promise<{
    id: string;
    updatedAt: string;
    name: string;
    description: string | null;
    createdAt: string;
}[]>;
export declare function listEnvironmentsForUser(userId: string): Promise<{
    env: {
        id: string;
        updatedAt: string;
        name: string;
        description: string | null;
        createdAt: string;
    };
    role: "viewer" | "editor";
}[]>;
export declare function getMembership(environmentId: string, userId: string): Promise<{
    environmentId: string;
    userId: string;
    role: "viewer" | "editor";
}>;
export declare function requireEnvAccess(environmentId: string, userId: string, needWrite: boolean): Promise<{
    environmentId: string;
    userId: string;
    role: "viewer" | "editor";
}>;
export declare function createEnvironment(input: {
    name: string;
    description?: string | null;
    creatorUserId: string;
}): Promise<{
    id: string;
    updatedAt: string;
    name: string;
    description: string | null;
    createdAt: string;
}>;
export declare function getEnvironmentById(id: string): Promise<{
    id: string;
    updatedAt: string;
    name: string;
    description: string | null;
    createdAt: string;
}>;
export declare function updateEnvironment(id: string, patch: {
    name?: string;
    description?: string | null;
}): Promise<{
    id: string;
    updatedAt: string;
    name: string;
    description: string | null;
    createdAt: string;
}>;
export declare function deleteEnvironment(id: string): Promise<void>;
export declare function listMembers(environmentId: string): Promise<{
    userId: string;
    role: "viewer" | "editor";
    username: string;
    authUuid: string;
}[]>;
export declare function upsertMember(input: {
    environmentId: string;
    userId: string;
    role: MemberRole;
}): Promise<void>;
export declare function removeMember(environmentId: string, userId: string): Promise<void>;
export declare function getHiveUserById(userId: string): Promise<{
    id: string;
    authUuid: string;
    username: string;
    systemName: string;
    updatedAt: string;
}>;
//# sourceMappingURL=services.d.ts.map