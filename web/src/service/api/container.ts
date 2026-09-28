import {get, post} from "@/utils/request";

/** 获取容器列表 GET /container/list */
export async function getContainerList(params: API.RequestParams = {}) {
    return get("/container/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取容器详情 GET /container/info */
export async function getContainerInfo(params: API.RequestParams = {}) {
    return get("/container/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取容器运行状态 GET /container/status */
export async function getContainerStatus(params: API.RequestParams = {}) {
    return get("/container/status", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建容器 POST /container/run；本机镜像解包使用独立超时。 */
export async function runContainer(params: API.RequestParams = {}) {
    return post("/container/run", params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: 300000});
}

/** 启动容器 POST /container/start */
export async function startContainer(params: API.RequestParams = {}) {
    return post("/container/start", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 停止容器 POST /container/stop */
export async function stopContainer(params: API.RequestParams = {}) {
    return post("/container/stop", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 重启容器 POST /container/restart */
export async function restartContainer(params: API.RequestParams = {}) {
    return post("/container/restart", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 批量删除容器 POST /container/remove，body.ids 为容器 ID 列表。 */
export async function removeContainer(params: API.RequestParams = {}) {
    return post("/container/remove", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 镜像管理 POST /container/images；导入本机镜像使用独立超时。 */
export async function containerImages(params: API.RequestParams = {}) {
    return post("/container/images", params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: 300000});
}

/** 读取或清理容器日志 POST /container/log */
export async function getContainerLog(params: API.RequestParams = {}) {
    return post("/container/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 执行容器命令 POST /container/exec */
export async function execContainer(params: API.RequestParams = {}) {
    const timeout = Math.min(3600, Math.max(1, Number(params?.body?.timeout) || 30));
    return post("/container/exec", params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: (timeout + 15) * 1000});
}
