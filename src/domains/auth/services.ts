import { config } from "../../config.js";
import type {
  UserverLoginResponse,
  UserverMeResponse,
  UserverRegisterResponse,
} from "./types.js";

function authBase(): string {
  if (!config.userver.configured) {
    throw new Error(
      "uServer-Auth is not configured (USERVER_AUTH_HOST, USERVER_AUTH_SYSTEM_NAME, USERVER_AUTH_SYSTEM_TOKEN)"
    );
  }
  return config.userver.host;
}

async function readJson(res: Response): Promise<unknown> {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { raw: text };
  }
}

export async function userverLogin(body: {
  username: string;
  password: string;
}): Promise<{ ok: true; data: UserverLoginResponse } | { ok: false; status: number; message: string }> {
  const res = await fetch(`${authBase()}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      username: body.username,
      password: body.password,
      system_name: config.userver.systemName,
      system_token: config.userver.systemToken,
    }),
  });
  const data = (await readJson(res)) as Record<string, unknown> | null;
  if (!res.ok) {
    const msg =
      typeof data?.message === "string"
        ? data.message
        : `login failed (${res.status})`;
    return { ok: false, status: res.status, message: msg };
  }
  return { ok: true, data: data as unknown as UserverLoginResponse };
}

export async function userverRegister(body: {
  username: string;
  password: string;
  is_admin?: boolean;
}): Promise<
  { ok: true; data: UserverRegisterResponse } | { ok: false; status: number; message: string }
> {
  const payload: Record<string, unknown> = {
    username: body.username,
    password: body.password,
    system_name: config.userver.systemName,
    system_token: config.userver.systemToken,
  };
  if (body.is_admin !== undefined) payload.is_admin = body.is_admin;
  const res = await fetch(`${authBase()}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const data = (await readJson(res)) as Record<string, unknown> | null;
  if (!res.ok) {
    const msg =
      typeof data?.message === "string"
        ? data.message
        : `register failed (${res.status})`;
    return { ok: false, status: res.status, message: msg };
  }
  return { ok: true, data: data as unknown as UserverRegisterResponse };
}

export async function userverRefresh(refreshToken: string): Promise<
  { ok: true; data: UserverLoginResponse } | { ok: false; status: number; message: string }
> {
  const res = await fetch(`${authBase()}/auth/refresh`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${refreshToken}`,
      "Content-Type": "application/json",
    },
  });
  const data = (await readJson(res)) as Record<string, unknown> | null;
  if (!res.ok) {
    const msg =
      typeof data?.message === "string"
        ? data.message
        : `refresh failed (${res.status})`;
    return { ok: false, status: res.status, message: msg };
  }
  return { ok: true, data: data as unknown as UserverLoginResponse };
}

export async function userverLogout(accessToken: string): Promise<{
  ok: boolean;
  status: number;
}> {
  const res = await fetch(`${authBase()}/auth/logout`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
  return { ok: res.ok || res.status === 204, status: res.status };
}

export async function userverMe(
  accessToken: string
): Promise<
  { ok: true; data: UserverMeResponse } | { ok: false; status: number; message: string }
> {
  const res = await fetch(`${authBase()}/auth/me`, {
    method: "GET",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  const data = (await readJson(res)) as Record<string, unknown> | null;
  if (!res.ok) {
    const msg =
      typeof data?.message === "string"
        ? data.message
        : `me failed (${res.status})`;
    return { ok: false, status: res.status, message: msg };
  }
  if (!data || typeof data.uuid !== "string") {
    return { ok: false, status: 502, message: "Invalid me response" };
  }
  return { ok: true, data: data as unknown as UserverMeResponse };
}

export async function userverChangePassword(
  accessToken: string,
  body: { current_password: string; new_password: string }
): Promise<{ ok: true } | { ok: false; status: number; message: string }> {
  const res = await fetch(`${authBase()}/auth/me/password`, {
    method: "PATCH",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  const data = (await readJson(res)) as Record<string, unknown> | null;
  if (!res.ok) {
    const msg =
      typeof data?.message === "string"
        ? data.message
        : `password update failed (${res.status})`;
    return { ok: false, status: res.status, message: msg };
  }
  return { ok: true };
}
