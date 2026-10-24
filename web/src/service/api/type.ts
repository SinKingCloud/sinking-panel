import {get, post} from "@/utils/request";

/** 获取分类枚举 GET /type/enum */
export async function getTypeEnum(params: API.RequestParams = {}) {
    return get("/type/enum", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取分类列表 GET /type/list */
export async function getTypeList(params: API.RequestParams = {}) {
    return get("/type/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建分类 POST /type/create */
export async function createType(params: API.RequestParams = {}) {
    return post("/type/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改分类 POST /type/update */
export async function updateType(params: API.RequestParams = {}) {
    return post("/type/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除分类 POST /type/delete */
export async function deleteType(params: API.RequestParams = {}) {
    return post("/type/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
