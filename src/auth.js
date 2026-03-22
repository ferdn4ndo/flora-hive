import { config } from "./config.js";

const keySet = new Set(config.apiKeys);

function extractBearer(authHeader) {
  if (!authHeader || typeof authHeader !== "string") return null;
  const m = authHeader.match(/^Bearer\s+(.+)$/i);
  return m ? m[1].trim() : null;
}

export function requireApiKey(req, res, next) {
  const fromHeader = req.get("X-API-Key");
  const fromBearer = extractBearer(req.get("Authorization"));
  const key = (fromHeader || fromBearer || "").trim();

  if (!key || !keySet.has(key)) {
    res.status(401).json({
      error: "unauthorized",
      message: "Valid API key required (X-API-Key or Authorization: Bearer)",
    });
    return;
  }

  next();
}
