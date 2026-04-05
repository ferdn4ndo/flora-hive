import type { Request, Response, NextFunction } from "express";
import { config } from "../../config.js";
import { upsertHiveUserFromMe } from "../user/services.js";
import { userverMe } from "./services.js";
import type { AuthPrincipal } from "./types.js";

function extractBearer(header: string | undefined): string | null {
  if (!header) return null;
  const m = header.match(/^Bearer\s+(.+)$/i);
  return m ? m[1].trim() : null;
}

export async function attachAuthOptional(
  req: Request,
  _res: Response,
  next: NextFunction
): Promise<void> {
  try {
    const xApi = req.get("X-API-Key")?.trim();
    if (xApi && config.apiKeys.includes(xApi)) {
      (req as Request & { auth?: AuthPrincipal }).auth = { kind: "api_key" };
      next();
      return;
    }

    const bearer = extractBearer(req.get("Authorization"));
    if (!bearer || !config.userver.configured) {
      next();
      return;
    }

    const me = await userverMe(bearer);
    if (!me.ok) {
      next();
      return;
    }

    const hiveUserId = await upsertHiveUserFromMe(me.data);
    (req as Request & { auth?: AuthPrincipal }).auth = {
      kind: "jwt",
      accessToken: bearer,
      me: me.data,
      hiveUserId,
    };
    next();
  } catch (e) {
    next(e);
  }
}

export function requireAuth(req: Request, res: Response, next: NextFunction) {
  const auth = (req as Request & { auth?: AuthPrincipal }).auth;
  if (!auth) {
    res.status(401).json({
      error: "unauthorized",
      message:
        "Authentication required: Bearer access token (uServer-Auth) or X-API-Key",
    });
    return;
  }
  next();
}
