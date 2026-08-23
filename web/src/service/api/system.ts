import {get, post} from "@/utils/request";

/** 获取账户信息 GET /system/account */
export async function getAccountInfo(params: API.RequestParams = {}) {
    return get("/system/account", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改账户 POST /system/account */
export async function updateAccount(params: API.RequestParams = {}) {
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

/** 获取 HTTP 服务状态或配置 GET /system/http */
export async function getHTTPService(params: API.RequestParams = {}) {
    return get("/system/http", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 管理 HTTP 服务或修改配置 POST /system/http */
export async function updateHTTPService(params: API.RequestParams = {}) {
    return post("/system/http", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统枚举 GET /system/enum */
export async function getEnum(params: API.RequestParams = {}) {
    return get("/system/enum", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统任务 GET /system/task */
export async function getSystemTask(params: API.RequestParams = {}) {
    return get("/system/task", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统任务列表 GET /system/task */
export async function getSystemTaskList(params: API.RequestParams = {}) {
    return get("/system/task", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 取消系统任务 POST /system/task */
export async function cancelSystemTask(params: API.RequestParams = {}) {
    return post("/system/task", {...params?.body, action: "cancel"}, params?.onSuccess, params?.onFail, params?.onFinally);
}
