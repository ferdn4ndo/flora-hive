import "dotenv/config";
declare const appConfig: {
    port: number;
    apiKeys: string[];
    postgres: {
        database: string;
        password?: string | undefined;
        host: string;
        port: number;
        user: string;
    };
    mqtt: {
        url: string;
        username: string | undefined;
        password: string | undefined;
        clientId: string;
        defaultQos: number;
    };
    topicPrefix: string;
    devicesSubscribePattern: string;
    deviceHeartbeatTtlSec: number;
    userver: {
        host: string;
        systemName: string;
        systemToken: string;
        configured: boolean;
    };
};
export declare const config: {
    port: number;
    apiKeys: string[];
    postgres: {
        database: string;
        password?: string | undefined;
        host: string;
        port: number;
        user: string;
    };
    mqtt: {
        url: string;
        username: string | undefined;
        password: string | undefined;
        clientId: string;
        defaultQos: number;
    };
    topicPrefix: string;
    devicesSubscribePattern: string;
    deviceHeartbeatTtlSec: number;
    userver: {
        host: string;
        systemName: string;
        systemToken: string;
        configured: boolean;
    };
};
export type AppConfig = typeof appConfig;
export {};
//# sourceMappingURL=config.d.ts.map