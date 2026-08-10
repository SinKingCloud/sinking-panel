import {get, post} from "@/utils/request";

/** 获取服务器列表 GET /server/list */
export async function getServerList(params: API.RequestParams = {}) {
    return get("/server/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取服务器详情 GET /server/info */
export async function getServerInfo(params: API.RequestParams = {}) {
    return get("/server/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建服务器 POST /server/create */
export async function createServer(params: API.RequestParams = {}) {
    return post("/server/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改服务器 POST /server/update */
export async function updateServer(params: API.RequestParams = {}) {
    return post("/server/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除服务器 POST /server/delete */
export async function deleteServer(params: API.RequestParams = {}) {
    return post("/server/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
