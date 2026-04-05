import type { Environment } from "../../db/schema.js";
export declare function environmentPublic(env: Environment, role?: string): {
    role?: string | undefined;
    id: string;
    name: string;
    mqttEnvPath: string;
    description: string | null;
    createdAt: string;
    updatedAt: string;
};
//# sourceMappingURL=views.d.ts.map