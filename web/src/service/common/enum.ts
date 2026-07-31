import {get} from "@/utils/request";

/** 获取枚举数据 */
export async function getEnum(params: API.RequestParams<{ name: string }, Record<string, Record<string, string>>> = {}) {
    return get<Record<string, Record<string, string>>>(
        "/system/enum",
        params?.body,
        params?.onSuccess,
        params?.onFail,
        params?.onFinally,
    );
}
