import type { UserverMeResponse } from "../auth/types.js";
export declare function upsertHiveUserFromMe(me: UserverMeResponse): Promise<string>;
export declare function findHiveUserByAuthUuid(authUuid: string): Promise<{
    id: string;
    authUuid: string;
    username: string;
    systemName: string;
    updatedAt: string;
}>;
//# sourceMappingURL=services.d.ts.map