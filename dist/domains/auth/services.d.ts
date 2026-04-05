import type { UserverLoginResponse, UserverMeResponse, UserverRegisterResponse } from "./types.js";
export declare function userverLogin(body: {
    username: string;
    password: string;
}): Promise<{
    ok: true;
    data: UserverLoginResponse;
} | {
    ok: false;
    status: number;
    message: string;
}>;
export declare function userverRegister(body: {
    username: string;
    password: string;
    is_admin?: boolean;
}): Promise<{
    ok: true;
    data: UserverRegisterResponse;
} | {
    ok: false;
    status: number;
    message: string;
}>;
export declare function userverRefresh(refreshToken: string): Promise<{
    ok: true;
    data: UserverLoginResponse;
} | {
    ok: false;
    status: number;
    message: string;
}>;
export declare function userverLogout(accessToken: string): Promise<{
    ok: boolean;
    status: number;
}>;
export declare function userverMe(accessToken: string): Promise<{
    ok: true;
    data: UserverMeResponse;
} | {
    ok: false;
    status: number;
    message: string;
}>;
export declare function userverChangePassword(accessToken: string, body: {
    current_password: string;
    new_password: string;
}): Promise<{
    ok: true;
} | {
    ok: false;
    status: number;
    message: string;
}>;
//# sourceMappingURL=services.d.ts.map