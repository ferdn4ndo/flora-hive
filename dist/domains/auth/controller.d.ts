import type { Request, Response, NextFunction } from "express";
export declare function postLogin(req: Request, res: Response): Promise<void>;
export declare function postRegister(req: Request, res: Response): Promise<void>;
export declare function postRefresh(req: Request, res: Response): Promise<void>;
export declare function postLogout(req: Request, res: Response, next: NextFunction): Promise<void>;
export declare function getMe(req: Request, res: Response, next: NextFunction): Promise<void>;
export declare function patchPassword(req: Request, res: Response, next: NextFunction): Promise<void>;
//# sourceMappingURL=controller.d.ts.map