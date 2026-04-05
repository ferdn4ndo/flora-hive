export function environmentPublic(env, role) {
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
//# sourceMappingURL=views.js.map