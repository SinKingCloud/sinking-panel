import {get, post} from "@/utils/request";

/** 获取分类列表 GET /type/list */
export async function getTypeList(params: API.RequestParams = {}) {
    return get("/type/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取指定模块的全部分类 */
export async function getAllTypes(module: string) {
    const items: any[] = [];
    let page = 1;
    let lastResponse: any;
    while (true) {
        const response = await getTypeList({
            body: {
                module,
                page,
                page_size: 1000,
                order_by_field: "sort",
                order_by_type: "asc",
            },
        });
        if (!response || response.code !== 200) {
            return response;
        }
        lastResponse = response;
        const batch = Array.isArray(response.data?.list) ? response.data.list : [];
        items.push(...batch);
        const total = Number(response.data?.total || 0);
        if (batch.length === 0 || items.length >= total) {
            break;
        }
        page += 1;
    }
    return {
        ...lastResponse,
        data: {
            ...lastResponse?.data,
            total: items.length,
            page: 1,
            page_size: items.length,
            list: items,
        },
    } as any;
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
