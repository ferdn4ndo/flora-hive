async function postLogin(base, systemName, systemToken, username, password) {
    const res = await fetch(`${base}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            username,
            password,
            system_name: systemName,
            system_token: systemToken,
        }),
    });
    return res.status;
}
/**
 * Optional uServer-Auth system + admin (same rules as userver-filemgr setup.sh).
 */
export async function ensureUserverAuthBootstrap() {
    const skip = process.env.SKIP_USERVER_AUTH_SETUP?.trim();
    if (skip === "1" || skip?.toLowerCase() === "true") {
        console.log("bootstrap: skipping userver-auth (SKIP_USERVER_AUTH_SETUP)");
        return;
    }
    const host = process.env.USERVER_AUTH_HOST?.replace(/\/$/, "").trim();
    if (!host) {
        console.log("bootstrap: skipping userver-auth (USERVER_AUTH_HOST unset)");
        return;
    }
    const systemName = process.env.USERVER_AUTH_SYSTEM_NAME?.trim() || "";
    const systemToken = process.env.USERVER_AUTH_SYSTEM_TOKEN || "";
    const creationToken = process.env.USERVER_AUTH_SYSTEM_CREATION_TOKEN || "";
    const authUser = process.env.USERVER_AUTH_USER?.trim() || "";
    const authPassword = process.env.USERVER_AUTH_PASSWORD || "";
    if (!systemName || !systemToken) {
        console.log("bootstrap: skipping userver-auth (USERVER_AUTH_SYSTEM_NAME or USERVER_AUTH_SYSTEM_TOKEN unset)");
        return;
    }
    if (!authUser || authPassword === undefined || authPassword === "") {
        console.log("bootstrap: skipping userver-auth bootstrap (USERVER_AUTH_USER / USERVER_AUTH_PASSWORD unset)");
        return;
    }
    console.log("bootstrap: checking userver-auth (login probe)...");
    let loginStatus = await postLogin(host, systemName, systemToken, authUser, authPassword);
    if (loginStatus === 200) {
        console.log("bootstrap: userver-auth credentials OK; skipping system/register.");
        return;
    }
    if (!creationToken) {
        throw new Error(`bootstrap: login failed (HTTP ${loginStatus}) and USERVER_AUTH_SYSTEM_CREATION_TOKEN is empty; cannot create system.`);
    }
    console.log(`bootstrap: userver-auth login returned HTTP ${loginStatus} (expected when bootstrapping).`);
    const sysRes = await fetch(`${host}/auth/system`, {
        method: "POST",
        headers: {
            Authorization: `Token ${creationToken}`,
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: systemName, token: systemToken }),
    });
    const sysBody = await sysRes.text();
    if (sysRes.status !== 201 && sysRes.status !== 409) {
        throw new Error(`bootstrap: POST /auth/system failed HTTP ${sysRes.status}: ${sysBody}`);
    }
    console.log(sysRes.status === 201
        ? "bootstrap: userver-auth system created."
        : "bootstrap: userver-auth system already exists (409), continuing.");
    const regRes = await fetch(`${host}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            username: authUser,
            password: authPassword,
            system_name: systemName,
            system_token: systemToken,
            is_admin: true,
        }),
    });
    const regBody = await regRes.text();
    if (regRes.status !== 201 && regRes.status !== 409) {
        throw new Error(`bootstrap: POST /auth/register failed HTTP ${regRes.status}: ${regBody}`);
    }
    console.log(regRes.status === 201
        ? "bootstrap: userver-auth admin user registered."
        : "bootstrap: userver-auth user already registered (409), continuing.");
    loginStatus = await postLogin(host, systemName, systemToken, authUser, authPassword);
    if (loginStatus !== 200) {
        throw new Error(`bootstrap: login still failing with HTTP ${loginStatus} after bootstrap`);
    }
    console.log("bootstrap: userver-auth login OK.");
}
//# sourceMappingURL=userverAuthBootstrap.js.map