import type { AuthPrincipal } from "../domains/auth/types.js";

declare global {
  namespace Express {
    interface Request {
      auth?: AuthPrincipal;
    }
  }
}

export {};
