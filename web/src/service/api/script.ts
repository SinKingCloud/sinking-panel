import {get, post} from "@/utils/request";

export interface ScriptRecord {
    id: number;
    type_id: number;
    name: string;
    create_time?: string;
    update_time?: string;
}

export interface ScriptInfo extends ScriptRecord {
    script: string;
}

export interface ScriptListData {
    page_size: number;
    list: ScriptRecord[];
    has_more?: boolean;
    next_cursor_id?: string;
}

export interface ScriptListQuery {
    type_id?: string;
    keyword?: string;
    page_size?: number;
    order_by_field?: "id";
    order_by_type?: "asc" | "desc";
    cursor_id?: string;
}

export interface ScriptCreateBody {
    type_id: number;
    name: string;
    script: string;
}

export interface ScriptUpdateBody {
    ids: number[];
    type_id: string;
    name: string;
    script: string;
}

/** 获取常用脚本列表 GET /script/list */
export async function getScriptList(params: API.RequestParams<ScriptListQuery, ScriptListData> = {}) {
    return get<ScriptListData>("/script/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取常用脚本详情 GET /script/info */
export async function getScriptInfo(params: API.RequestParams<{id: number}, ScriptInfo> = {}) {
    return get<ScriptInfo>("/script/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建常用脚本 POST /script/create */
export async function createScript(params: API.RequestParams<ScriptCreateBody> = {}) {
    return post("/script/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改常用脚本 POST /script/update */
export async function updateScript(params: API.RequestParams<ScriptUpdateBody> = {}) {
    return post("/script/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除常用脚本 POST /script/delete */
export async function deleteScript(params: API.RequestParams<{ids: number[]}> = {}) {
    return post("/script/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
