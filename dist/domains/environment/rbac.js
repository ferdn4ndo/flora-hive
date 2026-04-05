export function canReadRole(role) {
    return role === "viewer" || role === "editor";
}
export function canWriteRole(role) {
    return role === "editor";
}
//# sourceMappingURL=rbac.js.map