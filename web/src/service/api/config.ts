import {get, post} from "@/utils/request";

/** 获取系统配置 GET /config/get */
export async function getConfig(params: any = {}) {
    return get("/config/get", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改系统配置 POST /config/set */
export async function setConfig(params: any = {}) {
    return post("/config/set", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
