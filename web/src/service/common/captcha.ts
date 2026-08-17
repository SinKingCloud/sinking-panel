import {get} from "@/utils/request";

/** 获取验证码信息 GET /captcha */
export async function getCaptcha(params: API.RequestParams = {}) {
    return get("/captcha", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
