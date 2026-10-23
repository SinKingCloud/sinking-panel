import type {ComponentType, CSSProperties} from "react";

export interface RedisRecord {
    id: number | string;
    name: string;
    database_name?: string;
    server_name?: string;
    [key: string]: any;
}

export interface RedisVersion {
    name: string;
    version: string;
    features: Record<string, boolean>;
}

export type RedisOperation = "version" | "databases" | "keys" | "value" | "save" | "delete" | "rename" | "expire" | "command" | "users" | "maintenance" | "info";
export type RedisRequest = (params: {body: Record<string, any>}) => Promise<{code: number; data?: any; message?: string} | undefined>;
export type RedisRequests = Record<RedisOperation, RedisRequest>;
export type RedisApi = {[Name in RedisOperation]: (record: RedisRecord, body?: Record<string, any>) => Promise<any>};

export interface RedisManagerProps {
    record?: RedisRecord;
    requests: RedisRequests;
    open?: boolean;
    onCancel: () => void;
    onConfigure?: () => void;
    refreshKey?: number;
    initialTab?: string;
    components?: {ActionDropdown?: ComponentType<any>; Editor?: ComponentType<any>; [key: string]: any};
    className?: string;
    style?: CSSProperties;
}

export interface RedisClientContext {
    api: RedisApi;
    version?: RedisVersion;
    features: Record<string, boolean>;
    ActionDropdown: ComponentType<any>;
    Editor?: ComponentType<any>;
}
