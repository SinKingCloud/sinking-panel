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

/** 读取或修改 Redis 性能配置 POST /redis/performance */
export async function redisPerformance(params: API.RequestParams = {}) {
    return post("/redis/performance", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取或修改 Redis 持久化配置 POST /redis/persistence */
export async function redisPersistence(params: API.RequestParams = {}) {
    return post("/redis/persistence", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取或修改 Redis 配置文件 POST /redis/config */
export async function redisConfig(params: API.RequestParams = {}) {
    return post("/redis/config", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取或清理 Redis 日志 POST /redis/log */
export async function redisLog(params: API.RequestParams = {}) {
    return post("/redis/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
