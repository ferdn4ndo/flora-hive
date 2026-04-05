export type UserverMeResponse = {
    uuid: string;
    system_name: string;
    username: string;
    registered_at: string;
    last_activity_at: string;
    is_admin: boolean;
    token: {
        issued_at: string;
        expires_at: string;
    };
};
export type UserverLoginResponse = {
    access_token: string;
    access_token_exp: string;
    refresh_token: string;
    refresh_token_exp: string;
};
export type UserverRegisterResponse = {
    username: string;
    system_name: string;
    is_admin: boolean;
    auth: UserverLoginResponse;
};
export type AuthPrincipal = {
    kind: "jwt";
    accessToken: string;
    me: UserverMeResponse;
    hiveUserId: string;
} | {
    kind: "api_key";
};
//# sourceMappingURL=types.d.ts.map