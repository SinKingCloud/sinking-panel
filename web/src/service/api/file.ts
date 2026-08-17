import {getHeaders} from "@/utils/auth";
import {get, post} from "@/utils/request";

const createUploadError = (name: string, message: string) => {
    const error = new Error(message);
    error.name = name;
    return error;
};

/** 获取磁盘挂载路径 GET /file/disk */
export async function getFileDisks(params: API.RequestParams = {}) {
    return get("/file/disk", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取文件列表 GET /file/list */
export async function getFileList(params: API.RequestParams = {}) {
    return get("/file/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取文件信息 GET /file/info */
export async function getFileInfo(params: API.RequestParams = {}) {
    return get("/file/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取文件访问签名 GET /file/sign */
export async function getFileSign(params: API.RequestParams = {}) {
    return get("/file/sign", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建文件或目录 POST /file/create */
export async function createFile(params: API.RequestParams = {}) {
    return post("/file/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 重命名文件或目录 POST /file/update */
export async function renameFile(params: API.RequestParams = {}) {
    return post("/file/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 更新文件名称、权限或内容 POST /file/update */
export async function updateFile(params: API.RequestParams = {}) {
    return post("/file/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 统计文件或目录 GET /file/count */
export async function countFile(params: API.RequestParams = {}) {
    return get("/file/count", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 复制文件或目录 POST /file/copy */
export async function copyFile(params: API.RequestParams = {}) {
    return post("/file/copy", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 移动文件或目录 POST /file/move */
export async function moveFile(params: API.RequestParams = {}) {
    return post("/file/move", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 压缩文件或目录 POST /file/compress */
export async function compressFile(params: API.RequestParams = {}) {
    return post("/file/compress", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 解压文件 POST /file/extract */
export async function extractFile(params: API.RequestParams = {}) {
    return post("/file/extract", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 下载远程文件到服务器 POST /file/download */
export async function remoteDownloadFile(params: API.RequestParams = {}) {
    return post("/file/download", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除文件或目录 POST /file/delete */
export async function deleteFile(params: API.RequestParams = {}) {
    return post("/file/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取回收站列表 GET /recycle/list */
export async function getRecycleList(params: API.RequestParams = {}) {
    return get("/recycle/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取回收站统计 GET /recycle/count */
export async function getRecycleCount(params: API.RequestParams = {}) {
    return get("/recycle/count", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 恢复回收站文件 POST /recycle/restore */
export async function restoreRecycleItem(params: API.RequestParams = {}) {
    return post("/recycle/restore", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 彻底删除回收站文件 POST /recycle/delete */
export async function deleteRecycleItem(params: API.RequestParams = {}) {
    return post("/recycle/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 清空回收站 POST /recycle/clear */
export async function clearRecycle(params: API.RequestParams = {}) {
    return post("/recycle/clear", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 文件上传请求 GET|POST /file/upload */
export function requestFileUpload(
    endpoint: string,
    action: "upload" | "check" | "merge" | "clear",
    signal: AbortSignal,
    params: {
        body?: FormData;
        query?: Record<string, string>;
        onProgress?: (loaded: number, total: number) => void;
    },
) {
    return new Promise<API.Response<any>>((resolve, reject) => {
        if (signal.aborted) {
            reject(new DOMException("上传已取消", "AbortError"));
            return;
        }
        const xhr = new XMLHttpRequest();
        const url = new URL(endpoint);
        url.searchParams.set("action", action);
        Object.entries(params.query || {}).forEach(([key, value]) => url.searchParams.set(key, value));
        const abortRequest = () => xhr.abort();

        xhr.open(params.body ? "POST" : "GET", url.toString());
        Object.entries(getHeaders()).forEach(([key, value]) => xhr.setRequestHeader(key, value));
        xhr.upload.onprogress = (event) => params.onProgress?.(event.loaded, event.total);
        xhr.onload = () => {
            signal.removeEventListener("abort", abortRequest);
            if (xhr.status < 200 || xhr.status >= 300) {
                reject(createUploadError("FileUploadTransportError", `上传失败（HTTP ${xhr.status}）`));
                return;
            }
            try {
                const response = JSON.parse(xhr.responseText) as API.Response<any>;
                if (!response || typeof response.code !== "number") {
                    reject(createUploadError("FileUploadRequestError", "服务器返回的数据格式不正确"));
                    return;
                }
                resolve(response);
            } catch {
                reject(createUploadError("FileUploadRequestError", "服务器返回的数据格式不正确"));
            }
        };
        xhr.onerror = () => {
            signal.removeEventListener("abort", abortRequest);
            reject(createUploadError("FileUploadTransportError", "上传文件失败，请检查网络连接"));
        };
        xhr.ontimeout = () => {
            signal.removeEventListener("abort", abortRequest);
            reject(createUploadError("FileUploadTransportError", "上传文件超时"));
        };
        xhr.onabort = () => {
            signal.removeEventListener("abort", abortRequest);
            reject(new DOMException("上传已取消", "AbortError"));
        };
        signal.addEventListener("abort", abortRequest, {once: true});
        xhr.send(params.body);
    });
}
