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
export async function controlDatabase(kind: string, params: API.RequestParams = {}) {
    const {action, ...body} = params?.body || {};
    return post(`/${kind}/${action}`, body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: 180000});
}

/** 获取数据库性能配置 GET /:kind/performance */
export async function getDatabasePerformance(kind: string, params: API.RequestParams = {}) {
    return get(`/${kind}/performance`, {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改数据库性能配置 POST /:kind/performance */
export async function updateDatabasePerformance(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/performance`, {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取数据库配置文件 GET /:kind/config */
export async function getDatabaseConfig(kind: string, params: API.RequestParams = {}) {
    return get(`/${kind}/config`, {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改数据库配置文件 POST /:kind/config */
export async function updateDatabaseConfig(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/config`, {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取数据库列表 GET /:kind/databases */
export async function listDatabases(kind: string, params: API.RequestParams = {}) {
    return get(`/${kind}/databases`, {...params?.body, action: "list"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建或删除数据库 POST /:kind/databases */
export async function updateDatabase(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/databases`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取数据库账户列表 GET /:kind/users */
export async function listDatabaseUsers(kind: string, params: API.RequestParams = {}) {
    return get(`/${kind}/users`, {...params?.body, action: "list"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建、删除数据库账户或修改密码 POST /:kind/users */
export async function updateDatabaseUser(kind: string, params: API.RequestParams = {}) {
    return post(`/${kind}/users`, params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
