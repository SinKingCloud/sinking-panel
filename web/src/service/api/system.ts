import {get, post} from "@/utils/request";

export interface SystemEnumQuery {
    name: string;
}

export type SystemEnumData = Record<string, Record<string, string>>;

export type SystemTaskStatus = 0 | 1 | 2 | 3 | 4;

export interface FileTransferTaskData {
    source_paths: string[];
    target_path: string;
}

export interface FileCompressTaskData {
    paths: string[];
    dest_path: string;
    format: "zip" | "tar" | "gz" | "tgz";
}

export interface FileExtractTaskData {
    path: string;
    dest_dir: string;
    format: "zip" | "tar" | "gz" | "tgz";
}

export interface FileDownloadTaskData {
    url: string;
    target_path: string;
}

export type SystemTaskData = FileTransferTaskData | FileCompressTaskData | FileExtractTaskData | FileDownloadTaskData;

export interface SystemTaskRecord {
    id: string;
    name: string;
    status: SystemTaskStatus;
    progress: number;
    message: string;
    data?: SystemTaskData;
    start_time: number;
    end_time: number;
    create_time: number;
    update_time: number;
}

export interface SystemTaskQuery {
    id: string;
}

/** 获取账户信息 GET /system/account */
export async function getAccountInfo(params: API.RequestParams<any, API.UserInfo> = {}) {
    return get<API.UserInfo>("/system/account", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改账户 POST /system/account */
export async function updateAccount(params: API.RequestParams<{
    action: 'update';
    account?: string;
    password?: string;
}> = {}) {
    return post("/system/account", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取操作日志 GET /system/log */
export async function getLog(params: API.RequestParams = {}) {
    return get("/system/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统信息 GET /system/info */
export async function getSystemInfo(params: API.RequestParams = {}) {
    return get("/system/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统状态 GET /system/status */
export async function getSystemStatus(params: API.RequestParams = {}) {
    return get("/system/status", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统枚举 GET /system/enum */
export async function getEnum(params: API.RequestParams<SystemEnumQuery, SystemEnumData> = {}) {
    return get<SystemEnumData>("/system/enum", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统任务 GET /system/task */
export async function getSystemTask(params: API.RequestParams<SystemTaskQuery, SystemTaskRecord> = {}) {
    return get<SystemTaskRecord>("/system/task", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统任务列表 GET /system/task */
export async function getSystemTaskList(params: API.RequestParams<never, SystemTaskRecord[]> = {}) {
    return get<SystemTaskRecord[]>("/system/task", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 取消系统任务 POST /system/task */
export async function cancelSystemTask(params: API.RequestParams<SystemTaskQuery> = {}) {
    return post("/system/task", {...params?.body, action: "cancel"}, params?.onSuccess, params?.onFail, params?.onFinally);
}
