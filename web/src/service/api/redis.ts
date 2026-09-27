import {get, post} from "@/utils/request";

/** 获取 Redis 容器和服务状态 GET /redis/status */
export async function getRedisStatus(params: API.RequestParams = {}) {
    return get("/redis/status", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 解析本机镜像并生成默认配置，不创建或启动 Redis 实例。 */
export async function parseRedisImage(params: API.RequestParams = {}) {
    return post("/redis/image", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 导入本机镜像并创建 Redis 实例 POST /redis/install */
export async function installRedis(params: API.RequestParams = {}) {
    return post("/redis/install", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 卸载 Redis 实例 POST /redis/uninstall */
export async function uninstallRedis(params: API.RequestParams = {}) {
    return post("/redis/uninstall", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 启动 Redis POST /redis/start */
export async function startRedis(params: API.RequestParams = {}) {
    return post("/redis/start", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 停止 Redis POST /redis/stop */
export async function stopRedis(params: API.RequestParams = {}) {
    return post("/redis/stop", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 重启 Redis POST /redis/restart */
export async function restartRedis(params: API.RequestParams = {}) {
    return post("/redis/restart", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 重载 Redis 配置 POST /redis/reload */
export async function reloadRedis(params: API.RequestParams = {}) {
    return post("/redis/reload", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取 Redis 性能配置 GET /redis/performance */
export async function getRedisPerformance(params: API.RequestParams = {}) {
    return get("/redis/performance", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改 Redis 性能配置 POST /redis/performance */
export async function updateRedisPerformance(params: API.RequestParams = {}) {
    const body = params?.body || {};
    return post("/redis/performance", {action: "set", config: body.config ?? body}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取 Redis 持久化配置 GET /redis/persistence */
export async function getRedisPersistence(params: API.RequestParams = {}) {
    return get("/redis/persistence", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改 Redis 持久化配置 POST /redis/persistence */
export async function updateRedisPersistence(params: API.RequestParams = {}) {
    const body = params?.body || {};
    return post("/redis/persistence", {action: "set", config: body.config ?? body}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取 Redis 配置文件 GET /redis/config */
export async function getRedisConfig(params: API.RequestParams = {}) {
    return get("/redis/config", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改 Redis 配置文件 POST /redis/config */
export async function updateRedisConfig(params: API.RequestParams = {}) {
    return post("/redis/config", {action: "set", content: params?.body?.content || ""}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取 Redis 日志 GET /redis/log */
export async function getRedisLog(params: API.RequestParams = {}) {
    return get("/redis/log", {action: "read", page_size: 500, ...params?.body}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 清理 Redis 日志 POST /redis/log */
export async function clearRedisLog(params: API.RequestParams = {}) {
    return post("/redis/log", {action: "clear"}, params?.onSuccess, params?.onFail, params?.onFinally);
}
