import type {ComponentType, CSSProperties} from "react";
import type {ButtonProps, InputProps} from "antd";

// 每种数据库组件固定自己的类型，调用方只需要提供目标数据库信息。
export interface DatabasePanelRecord {
    id: number | string;
    name: string;
    database_name?: string;
    server_name?: string;
    [key: string]: any;
}

// operation 为数据库原生操作名或系统任务查询 system/task。
// 原生操作的 params 包含实例 ID 和目标 database_name；import / export + 服务器文件 path 返回后台任务 ID。
// system/task 由宿主映射到系统任务接口，action=info 返回 status、progress、message、result。
export type DatabasePanelRequest = (operation: string, params: Record<string, any>) => Promise<{
    code: number;
    data?: any;
    message?: string;
}>;

// 功能标识由服务端根据版本确定；未返回的标识表示尚未确认。
export interface DatabaseVersion {
    name: string;
    version: string;
    features: Record<string, boolean>;
}

export interface DatabasePanelProps {
    record: DatabasePanelRecord;
    initialTab?: string;
    onTabChange?: (tab: string) => void;
    request: DatabasePanelRequest;
    // 宿主已获取版本时直接复用，独立使用组件时仍会自动读取。
    version?: DatabaseVersion;
    components?: DatabasePanelComponents;
    className?: string;
    style?: CSSProperties;
}

export type DatabaseRequest = (params: {body: Record<string, any>}) => Promise<{code: number; data?: any; message?: string} | undefined>;

// 完整管理器入口；请求、文件选择器和编辑器由使用方注入。
export interface DatabaseManagerProps {
    record?: DatabasePanelRecord;
    requests: Record<string, DatabaseRequest>;
    open?: boolean;
    onCancel: () => void;
    onConfigure?: () => void;
    refreshKey?: number;
    initialTab?: string;
    components?: DatabasePanelComponents;
    className?: string;
    style?: CSSProperties;
}

// 宿主文件选择器在单选、多选时均接收并返回路径数组。
export interface FileSelectProps extends Omit<InputProps, "value" | "defaultValue" | "onChange" | "onSelect" | "multiple"> {
    value?: string[];
    defaultValue?: string[];
    onChange?: (paths: string[]) => void;
    onSelect?: (paths: string[]) => boolean | void;
    mode?: "file" | "directory";
    multiple?: boolean;
    initialPath?: string;
    buttonProps?: ButtonProps;
}

// 使用方可接入自己的服务器文件选择器、操作菜单和代码编辑器。
export interface DatabasePanelComponents {
    ScriptTools?: ComponentType<any>;
    FileSelect?: ComponentType<FileSelectProps>;
    ActionDropdown?: ComponentType<any>;
    Editor?: ComponentType<any>;
}

export interface DatabaseClientContext {
    ScriptTools?: ComponentType<any>;
    request: DatabasePanelRequest;
    version?: DatabaseVersion;
    features: Record<string, boolean>;
    operate: (record: DatabasePanelRecord, operation: string, body?: Record<string, any>) => Promise<any>;
    api: Record<string, (record: DatabasePanelRecord, body?: Record<string, any>) => Promise<any>> & {
        task: (body: Record<string, any>) => Promise<any>;
    };
    FileSelect: ComponentType<FileSelectProps>;
    ActionDropdown: ComponentType<any>;
    Editor?: ComponentType<any>;
}

export type DatabaseApi = DatabaseClientContext["api"];
