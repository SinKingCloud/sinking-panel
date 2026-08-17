import {get, post} from "@/utils/request";

/** 获取常用脚本列表 GET /script/list */
export async function getScriptList(params: API.RequestParams = {}) {
    return get("/script/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取常用脚本详情 GET /script/info */
export async function getScriptInfo(params: API.RequestParams = {}) {
    return get("/script/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建常用脚本 POST /script/create */
export async function createScript(params: API.RequestParams = {}) {
    return post("/script/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改常用脚本 POST /script/update */
export async function updateScript(params: API.RequestParams = {}) {
    return post("/script/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除常用脚本 POST /script/delete */
export async function deleteScript(params: API.RequestParams = {}) {
    return post("/script/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
