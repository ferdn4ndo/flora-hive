export function pathParam(v) {
    if (v === undefined)
        return "";
    return Array.isArray(v) ? (v[0] ?? "") : v;
}
//# sourceMappingURL=params.js.map