import {get, post} from "@/utils/request";
import {getHeaders} from "@/utils/auth";
import defaultSettings from "@/../config/defaultSettings";

/** OSS 配置列表。 */
export const getOssList = (params: API.RequestParams = {}) => get("/oss/list", params.body, params.onSuccess, params.onFail, params.onFinally);
/** 按 ID 获取 OSS 详情，config 返回解析后的配置对象。 */
export const getOssInfo = (params: API.RequestParams = {}) => get("/oss/info", params.body, params.onSuccess, params.onFail, params.onFinally);
/** 新建 OSS 配置。 */
export const createOss = (params: API.RequestParams = {}) => post("/oss/create", params.body, params.onSuccess, params.onFail, params.onFinally);
/** 修改 OSS 配置。 */
export const updateOss = (params: API.RequestParams = {}) => post("/oss/update", params.body, params.onSuccess, params.onFail, params.onFinally);
/** 删除面板中的 OSS 配置。 */
export const deleteOss = (params: API.RequestParams = {}) => post("/oss/delete", params.body, params.onSuccess, params.onFail, params.onFinally);
/** 使用云密钥 ID 和当前连接配置测试存储桶，无需先保存。 */
export const testOss = (params: API.RequestParams = {}) => post("/oss/test", params.body, params.onSuccess, params.onFail, params.onFinally);
/** 按系统游标分页读取文件和目录，返回 list、page_size、has_more 和 next_cursor_id。 */
export const getOssFiles = (params: API.RequestParams = {}) => post("/oss/files", params.body, params.onSuccess, params.onFail, params.onFinally);
/** 文件信息、目录创建、删除、复制、移动、重命名及元数据操作。 */
export const ossFile = (params: API.RequestParams = {}) => post("/oss/file", params.body, params.onSuccess, params.onFail, params.onFinally, undefined, ["delete", "copy", "move", "rename", "metadata"].includes(params.body?.action) ? {timeout: 15 * 60 * 1000} : undefined);
/** 管理分片会话或创建服务器上传任务，action 通过查询参数选择。 */
export const ossUpload = (params: API.RequestParams = {}) => post(`/oss/upload?action=${encodeURIComponent(params.body?.action || "file")}`, params.body, params.onSuccess, params.onFail, params.onFinally, undefined, params.body?.action === "complete" ? {timeout: 15 * 60 * 1000} : undefined);
/** 创建服务器下载任务，action=server 时返回系统任务编号。 */
export const ossDownload = (params: API.RequestParams = {}) => post(`/oss/download?action=${encodeURIComponent(params.body?.action || "file")}`, params.body, params.onSuccess, params.onFail, params.onFinally);
/** 生成文件临时访问或附件下载签名。 */
export const ossSign = (params: API.RequestParams = {}) => post("/oss/sign", params.body, params.onSuccess, params.onFail, params.onFinally);

/** 发送一个文件或分片，浏览器直接发送 Blob，不将文件读入 JS 内存。 */
const sendOssBody = (query: Record<string, string>, body: Blob, signal: AbortSignal, progress: (loaded: number) => void): Promise<any> => new Promise((resolve, reject) => {
    if (signal.aborted) { reject(new DOMException("上传已取消", "AbortError")); return; }
    const url = new URL(`${defaultSettings.gateway || ""}/oss/upload`, window.location.href);
    Object.entries(query).forEach(([key, value]) => url.searchParams.set(key, value));
    const xhr = new XMLHttpRequest();
    const abort = () => xhr.abort();
    const finish = () => signal.removeEventListener("abort", abort);
    xhr.open("POST", url.toString());
    xhr.timeout = 15 * 60 * 1000;
    Object.entries(getHeaders()).forEach(([key, value]) => xhr.setRequestHeader(key, value));
    xhr.setRequestHeader("Content-Type", body.type || "application/octet-stream");
    xhr.upload.onprogress = (event) => progress(event.loaded);
    xhr.onload = () => {
        finish();
        try {
            if (xhr.status < 200 || xhr.status >= 300) throw new Error(`上传失败（HTTP ${xhr.status}）`);
            const response = JSON.parse(xhr.responseText);
            if (response?.code !== 200) throw new Error(response?.message || "上传失败");
            resolve(response.data);
        } catch (error) { reject(error); }
    };
    xhr.onerror = () => { finish(); reject(new Error("上传连接失败")); };
    xhr.ontimeout = () => { finish(); reject(new Error("上传超时")); };
    xhr.onabort = () => { finish(); reject(new DOMException("上传已取消", "AbortError")); };
    signal.addEventListener("abort", abort, {once: true});
    xhr.send(body);
});

/** 小文件直接上传，大文件顺序分片；失败或取消时清理云端分片。 */
export async function uploadOssFile(id: number, key: string, file: File, signal: AbortSignal, progress: (loaded: number, finalizing?: boolean) => void) {
    const check = () => { if (signal.aborted) throw new DOMException("上传已取消", "AbortError"); };
    check();
    if (file.size <= 16 * 1024 * 1024) return sendOssBody({action: "file", id: String(id), key}, file, signal, progress);
    const response = await ossUpload({body: {id, key, action: "init", options: {content_type: file.type || "application/octet-stream"}}});
    if (response?.code !== 200 || !response.data?.upload_id) throw new Error(response?.message || "初始化分片上传失败");
    const upload_id = response.data.upload_id;
    try {
        const parts: any[] = [];
        const partSize = Math.max(16 * 1024 * 1024, Math.ceil(file.size / 10000));
        for (let offset = 0, number = 1; offset < file.size; offset += partSize, number++) {
            check();
            const part = await sendOssBody({action: "part", id: String(id), key, upload_id, number: String(number)}, file.slice(offset, Math.min(offset + partSize, file.size)), signal, (loaded) => progress(offset + loaded));
            parts.push(part);
        }
        check();
        progress(file.size, true);
        const complete = await ossUpload({body: {id, key, upload_id, action: "complete", parts}});
        if (complete?.code !== 200) throw new Error(complete?.message || "合并分片失败");
        progress(file.size);
        return complete.data;
    } catch (error: any) {
        let aborted;
        try { aborted = await ossUpload({body: {id, key, upload_id, action: "abort"}}); } catch { /* 保留原始错误并提示未清理的上传 ID。 */ }
        if (aborted?.code !== 200) throw new Error(`${error?.message || "上传失败"}；清理分片失败，请在云端检查未完成上传 ${upload_id}`);
        throw error;
    }
}
