import type { Request, Response, NextFunction } from "express";
export declare function attachAuthOptional(req: Request, _res: Response, next: NextFunction): Promise<void>;
export declare function requireAuth(req: Request, res: Response, next: NextFunction): void;
//# sourceMappingURL=middleware.d.ts.map