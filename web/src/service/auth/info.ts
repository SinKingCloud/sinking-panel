import {get} from "@/utils/request";

/** 获取当前网站信息 GET /info */
export async function getWebInfo(params: API.RequestParams<any, API.WebInfo> = {}) {
    return get<API.WebInfo>("/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
