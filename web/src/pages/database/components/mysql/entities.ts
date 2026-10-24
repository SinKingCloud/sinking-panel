import type {ComponentType, CSSProperties} from "react";

export interface DatabasePanelRecord {
    id: number | string;
    name: string;
    database_name?: string;
    server_name?: string;
    [key: string]: any;
}

export interface DatabaseVersion {
    name: string;
    version: string;
    features: Record<string, boolean>;
}

export type MysqlOperation = "version" | "databases" | "database_info" | "charsets" | "create_database" | "update_database" | "drop_database"
    | "tables" | "columns" | "indexes" | "rows" | "ddl" | "execute"
    | "insert" | "update" | "delete" | "create_table" | "rename_table" | "add_column" | "update_column" | "drop_column"
    | "create_index" | "update_index" | "drop_index" | "objects" | "object_definition" | "save_object" | "drop_object"
    | "create_foreign_key" | "update_foreign_key" | "drop_foreign_key" | "users" | "maintenance" | "import" | "export";

export type MysqlRequest = (params: {body: Record<string, any>}) => Promise<{code: number; data?: any; message?: string} | undefined>;
export type MysqlRequests = Record<MysqlOperation, MysqlRequest> & {"system/task": MysqlRequest};

export interface MysqlComponents {
    ScriptTools?: ComponentType<any>;
    FilePicker?: ComponentType<any>;
    ActionDropdown?: ComponentType<any>;
    Editor?: ComponentType<any>;
}

// 完整管理器只依赖实例与独立接口方法，可选接入宿主的文件选择器、菜单和编辑器。
export interface MysqlManagerProps {
    record?: DatabasePanelRecord;
    requests: MysqlRequests;
    open?: boolean;
    onCancel: () => void;
    onConfigure?: () => void;
    refreshKey?: number;
    initialTab?: string;
    components?: MysqlComponents;
    className?: string;
    style?: CSSProperties;
}

export type MysqlApi = {
    [Operation in MysqlOperation]: (record: DatabasePanelRecord, body?: Record<string, any>) => Promise<any>;
} & {
    task: (body: Record<string, any>) => ReturnType<MysqlRequest>;
};

export interface MysqlClientContext {
    ScriptTools?: ComponentType<any>;
    api: MysqlApi;
    version?: DatabaseVersion;
    features: Record<string, boolean>;
    FilePicker: ComponentType<any>;
    ActionDropdown: ComponentType<any>;
    Editor?: ComponentType<any>;
}
