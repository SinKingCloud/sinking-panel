import {get, post} from "@/utils/request";

/** 查询或修改账户 POST /system/account */
export async function systemAccount(params: API.RequestParams = {}) {
    return post("/system/account", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 查询或设置系统时区 POST /system/timezone */
export async function systemTimezone(params: API.RequestParams = {}) {
    return post("/system/timezone", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取操作日志 GET /system/log */
export async function getLog(params: API.RequestParams = {}) {
    return get("/system/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取主机、系统与运行时信息 GET /system/info */
export async function getSystemInfo(params: API.RequestParams = {}) {
    return get("/system/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 重启面板或服务器 POST /system/restart */
export async function restartSystem(params: API.RequestParams = {}) {
    return post("/system/restart", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统资源使用情况 GET /system/statis */
export async function getSystemStatis(params: API.RequestParams = {}) {
    return get("/system/statis", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网络与磁盘 IO 速率、采样历史 GET /system/io */
export async function getSystemIO(params: API.RequestParams = {}) {
    return get("/system/io", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 查询或管理 HTTP 服务、配置及日志 POST /system/http */
export async function systemHTTP(params: API.RequestParams = {}) {
    return post("/system/http", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统枚举 GET /system/enum */
export async function getEnum(params: API.RequestParams = {}) {
    return get("/system/enum", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 查询或管理系统任务及日志 POST /system/task */
export async function systemTask(params: API.RequestParams = {}) {
    return post("/system/task", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
