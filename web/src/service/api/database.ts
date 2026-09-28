import {get, post} from "@/utils/request";

/** 获取数据库服务状态 GET /:kind/status */
export async function getDatabaseStatus(kind: string, params: API.RequestParams = {}) {
    return get(`/${kind}/status`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 解析数据库镜像 POST /:kind/image */
export async function parseDatabaseImage(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/image`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: 300000});
}

/** 安装数据库 POST /:kind/install */
export async function installDatabase(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/install`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: 300000});
}

/** 启动、停止、重启、重载或卸载数据库 POST /:kind/:action */
export async function controlDatabase(kind: string, action: string, params: API.RequestParams = {}) {
    return post(`/${kind}/${action}`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: 180000});
}

/** 读取或修改数据库性能配置 POST /:kind/performance */
export async function databasePerformance(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/performance`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取或修改数据库配置文件 POST /:kind/config */
export async function databaseConfig(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/config`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 查询、创建或删除数据库 POST /:kind/databases */
export async function databaseDatabases(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/databases`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 查询、创建、删除数据库账户或修改密码 POST /:kind/users */
export async function databaseUsers(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/users`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
