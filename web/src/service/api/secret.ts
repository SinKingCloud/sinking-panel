import {get, post} from "@/utils/request";

export interface SecretRecord {
    id: number;
    name: string;
    provider: number;
    create_time: string;
    update_time: string;
}

export interface SecretInfo extends SecretRecord {
    data: Record<string, string>;
}

export interface SecretEnum {
    provider: Record<string, string>;
    name: Record<string, string>;
}

/** 获取密钥列表 GET /secret/list */
export async function getSecretList(params: API.RequestParams = {}) {
    return get("/secret/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取密钥详情 GET /secret/info */
export async function getSecretInfo(params: API.RequestParams = {}) {
    return get("/secret/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建密钥 POST /secret/create */
export async function createSecret(params: API.RequestParams = {}) {
    return post("/secret/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改密钥 POST /secret/update */
export async function updateSecret(params: API.RequestParams = {}) {
    return post("/secret/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除密钥 POST /secret/delete */
export async function deleteSecret(params: API.RequestParams = {}) {
    return post("/secret/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
