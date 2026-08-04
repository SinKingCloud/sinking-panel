import {get, post} from "@/utils/request";

/** 获取账户信息 GET /system/account */
export async function getAccountInfo(params: API.RequestParams<any, API.UserInfo> = {}) {
    return get<API.UserInfo>("/system/account", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改账户 POST /system/account */
export async function updateAccount(params: API.RequestParams<{
    action: 'update';
    account?: string;
    password?: string;
}> = {}) {
    return post("/system/account", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取操作日志 GET /system/log */
export async function getLog(params: API.RequestParams = {}) {
    return get("/system/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
