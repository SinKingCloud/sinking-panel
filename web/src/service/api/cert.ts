import {get, post} from "@/utils/request";

/** 获取证书列表 GET /cert/list */
export async function getCertList(params: API.RequestParams = {}) {
    return get("/cert/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取证书详情 GET /cert/info */
export async function getCertInfo(params: API.RequestParams = {}) {
    return get("/cert/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取或清理自动续签日志 POST /cert/log */
export async function getCertLog(params: API.RequestParams = {}) {
    return post("/cert/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 导入证书 POST /cert/create */
export async function createCert(params: API.RequestParams = {}) {
    return post("/cert/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改证书 POST /cert/update */
export async function updateCert(params: API.RequestParams = {}) {
    return post("/cert/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除证书 POST /cert/delete */
export async function deleteCert(params: API.RequestParams = {}) {
    return post("/cert/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 申请证书 POST /cert/obtain */
export async function obtainCert(params: API.RequestParams = {}) {
    return post("/cert/obtain", params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: Number(params?.body?.type) === 1 ? 21 * 60 * 1000 : 35 * 1000});
}

/** 续签证书 POST /cert/renew */
export async function renewCert(params: API.RequestParams = {}) {
    return post("/cert/renew", params?.body, params?.onSuccess, params?.onFail, params?.onFinally, undefined, {timeout: Number(params?.body?.type) === 1 ? 21 * 60 * 1000 : 35 * 1000});
}
