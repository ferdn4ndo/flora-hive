import type { Environment } from "../../db/schema.js";

export function environmentPublic(
  env: Environment,
  role?: string
) {
  return {
    id: env.id,
    name: env.name,
    mqttEnvPath: `environments/${env.id}`,
    description: env.description,
    createdAt: env.createdAt,
    updatedAt: env.updatedAt,
    ...(role !== undefined ? { role } : {}),
  };
}
