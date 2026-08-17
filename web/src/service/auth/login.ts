import {get, post} from "@/utils/request";

/** 账号登录 POST /login */
export async function login(params: API.RequestParams = {}) {
    return post("/login", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 退出登录 GET /logout */
export async function logout(params: API.RequestParams = {}) {
    return get("/logout", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
