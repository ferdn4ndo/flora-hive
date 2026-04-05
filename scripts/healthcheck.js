/**
 * Container / orchestrator probe: GET /healthz on the app port (default 8080).
 * Uses Node only (no curl/wget) so the runtime image stays minimal.
 */
import http from "node:http";

const port = Number(process.env.PORT) || 8080;
const timeoutMs = 4000;

const req = http.get(
  `http://127.0.0.1:${port}/healthz`,
  (res) => {
    res.resume();
    res.on("end", () => {
      process.exit(res.statusCode === 200 ? 0 : 1);
    });
  }
);

req.on("error", () => process.exit(1));
req.setTimeout(timeoutMs, () => {
  req.destroy();
  process.exit(1);
});
