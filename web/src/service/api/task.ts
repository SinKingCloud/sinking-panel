import {get, post} from "@/utils/request";

/** 获取计划任务列表 GET /task/list */
export async function getTaskList(params: API.RequestParams = {}) {
    return get("/task/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取计划任务详情 GET /task/info */
export async function getTaskInfo(params: API.RequestParams = {}) {
    return get("/task/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取或清理计划任务日志 GET|POST /task/log */
export async function getTaskLog(params: API.RequestParams = {}) {
    const request = params?.body?.action === "clear" ? post : get;
    return request("/task/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建计划任务 POST /task/create */
export async function createTask(params: API.RequestParams = {}) {
    return post("/task/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改计划任务 POST /task/update */
export async function updateTask(params: API.RequestParams = {}) {
    return post("/task/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除计划任务 POST /task/delete */
export async function deleteTask(params: API.RequestParams = {}) {
    return post("/task/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 立即执行计划任务 POST /task/run */
export async function runTask(params: API.RequestParams = {}) {
    return post("/task/run", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 暂停计划任务 POST /task/stop */
export async function stopTask(params: API.RequestParams = {}) {
    return post("/task/stop", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 恢复计划任务 POST /task/restore */
export async function restoreTask(params: API.RequestParams = {}) {
    return post("/task/restore", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
