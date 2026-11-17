import {get, post} from "@/utils/request";

/** 获取证书列表 GET /site/cert/list */
export async function getCertList(params: API.RequestParams = {}) {
    return get("/site/cert/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取证书详情 GET /site/cert/info */
export async function getCertInfo(params: API.RequestParams = {}) {
    return get("/site/cert/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取自动续签日志 POST /site/cert/log/list */
export async function getCertLog(params: API.RequestParams = {}) {
    return post("/site/cert/log/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 清理自动续签日志 POST /site/cert/log/clear */
export async function clearCertLog(params: API.RequestParams = {}) {
    return post("/site/cert/log/clear", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取自动续签设置 POST /site/cert/settings/info */
export async function getCertSettingsInfo(params: API.RequestParams = {}) {
    return post("/site/cert/settings/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改自动续签设置 POST /site/cert/settings/update */
export async function updateCertSettings(params: API.RequestParams = {}) {
    return post("/site/cert/settings/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 导入证书 POST /site/cert/create */
export async function createCert(params: API.RequestParams = {}) {
    return post("/site/cert/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改证书 POST /site/cert/update */
export async function updateCert(params: API.RequestParams = {}) {
    return post("/site/cert/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除证书 POST /site/cert/delete */
export async function deleteCert(params: API.RequestParams = {}) {
    return post("/site/cert/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 申请证书 POST /site/cert/obtain */
export async function obtainCert(params: API.RequestParams = {}) {
    return post("/site/cert/obtain", params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: Number(params?.body?.type) === 1 ? 21 * 60 * 1000 : 50 * 1000});
}

/** 续签证书 POST /site/cert/renew */
export async function renewCert(params: API.RequestParams = {}) {
    return post("/site/cert/renew", params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: Number(params?.body?.type) === 1 ? 21 * 60 * 1000 : 50 * 1000});
}
