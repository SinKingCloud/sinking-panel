import {deleteHeader, getHeaders} from "@/utils/auth";
import {get, post} from "@/utils/request";
import {historyPush} from "@/utils/route";

export interface FileRecord {
    name: string;
    size: number;
    mode: number;
    is_dir: boolean;
    update_time: number;
    child?: FileRecord[] | null;
}

export type FileOrderField = "name" | "size" | "update_time";
export type FileOrderType = "asc" | "desc";

export interface FileListQuery {
    path: string;
    keyword?: string;
    page?: number;
    page_size?: number;
    order_by_field?: FileOrderField;
    order_by_type?: FileOrderType;
}

export interface FileListData {
    total: number;
    page: number;
    page_size: number;
    list: FileRecord[];
}

export interface FileInfoQuery {
    path: string;
    read?: boolean;
    cursor?: number;
    page_size?: number;
    version?: string;
}

export interface FileInfoData {
    name: string;
    size: number;
    mode: number;
    is_dir: boolean;
    update_time: string;
    content?: string;
    cursor?: number;
    next_cursor?: number;
    eof?: boolean;
    version?: string;
}

export interface FileCreateBody {
    path: string;
    name: string;
}

export interface FileRenameBody {
    path: string;
    name: string;
}

export interface FileUpdateBody {
    path: string;
    name?: string;
    permissions?: string;
    content?: string;
    cursor?: number;
    next_cursor?: number;
    version?: string;
}

export interface FileUpdateResult {
    operations: string[];
    path: string;
    cursor?: number;
    next_cursor?: number;
    eof?: boolean;
    version?: string;
}

export interface FileCountQuery {
    path: string;
}

export interface FileCountData {
    size: number;
    file: number;
    dir: number;
}

export interface FileTransferBody {
    source_paths: string[];
    target_path: string;
}

export interface FileCompressBody {
    paths: string[];
    dir: string;
    format: "zip" | "tar" | "gz" | "tgz";
    name: string;
}

export interface FileExtractBody {
    path: string;
    dir: string;
}

export interface FileRemoteDownloadBody {
    url: string;
    path: string;
    name?: string;
}

export type SystemTaskStatus = 0 | 1 | 2 | 3 | 4;

export interface FileTransferTaskData {
    source_paths: string[];
    target_path: string;
}

export interface FileCompressTaskData {
    paths: string[];
    dest_path: string;
    format: FileCompressBody["format"];
}

export interface FileExtractTaskData {
    path: string;
    dest_dir: string;
    format: FileCompressBody["format"];
}

export interface FileDownloadTaskData {
    url: string;
    target_path: string;
}

export type SystemTaskData = FileTransferTaskData | FileCompressTaskData | FileExtractTaskData | FileDownloadTaskData;

export interface SystemTaskRecord {
    id: string;
    name: string;
    status: SystemTaskStatus;
    progress: number;
    message: string;
    data?: SystemTaskData;
    start_time: number;
    end_time: number;
    create_time: number;
    update_time: number;
}

export interface SystemTaskQuery {
    id: string;
}

export interface FileDeleteBody {
    paths: string[];
    recycle?: boolean;
}

export interface RecycleRecord {
    id: string;
    name: string;
    path: string;
    size: number;
    is_dir: boolean;
    update_time: string;
}

export interface RecycleListQuery {
    page?: number;
    page_size?: number;
    order_by_field?: FileOrderField;
    order_by_type?: FileOrderType;
}

export interface RecycleListData {
    total: number;
    page: number;
    page_size: number;
    list: RecycleRecord[];
}

export interface RecycleItemBody {
    names: string[];
}

export interface RecycleRestoreBody extends RecycleItemBody {
    path?: string;
}

export interface FileUploadData {
    name: string;
    path: string;
    size: number;
    file_hash?: string;
}

export interface FileSignQuery {
    path: string;
    download?: boolean;
}

export class FileUploadRequestError extends Error {
    constructor(message: string) {
        super(message);
        this.name = "FileUploadRequestError";
    }
}

export class FileUploadTransportError extends FileUploadRequestError {
    constructor(message: string) {
        super(message);
        this.name = "FileUploadTransportError";
    }
}

/** 获取磁盘挂载路径 GET /file/disk */
export async function getFileDisks(params: API.RequestParams<never, string[]> = {}) {
    return get<string[]>("/file/disk", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取文件列表 GET /file/list */
export async function getFileList(params: API.RequestParams<FileListQuery, FileListData> = {}) {
    return get<FileListData>("/file/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取文件信息 GET /file/info */
export async function getFileInfo(params: API.RequestParams<FileInfoQuery, FileInfoData> = {}) {
    return get<FileInfoData>("/file/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取文件访问签名 GET /file/sign */
export async function getFileSign(params: API.RequestParams<FileSignQuery, string> = {}) {
    return get<string>("/file/sign", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建文件或目录 POST /file/create */
export async function createFile(params: API.RequestParams<FileCreateBody> = {}) {
    return post("/file/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 重命名文件或目录 POST /file/update */
export async function renameFile(params: API.RequestParams<FileRenameBody> = {}) {
    return post("/file/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 更新文件名称、权限或内容 POST /file/update */
export async function updateFile(params: API.RequestParams<FileUpdateBody, FileUpdateResult> = {}) {
    return post<FileUpdateResult>("/file/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 统计文件或目录 GET /file/count */
export async function countFile(params: API.RequestParams<FileCountQuery, FileCountData> = {}) {
    return get<FileCountData>("/file/count", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 复制文件或目录 POST /file/copy */
export async function copyFile(params: API.RequestParams<FileTransferBody, string> = {}) {
    return post<string>("/file/copy", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 移动文件或目录 POST /file/move */
export async function moveFile(params: API.RequestParams<FileTransferBody, string> = {}) {
    return post<string>("/file/move", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 压缩文件或目录 POST /file/compress */
export async function compressFile(params: API.RequestParams<FileCompressBody, string> = {}) {
    return post<string>("/file/compress", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 解压文件 POST /file/extract */
export async function extractFile(params: API.RequestParams<FileExtractBody, string> = {}) {
    return post<string>("/file/extract", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 下载远程文件到服务器 POST /file/download */
export async function remoteDownloadFile(params: API.RequestParams<FileRemoteDownloadBody, string> = {}) {
    return post<string>("/file/download", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除文件或目录 POST /file/delete */
export async function deleteFile(params: API.RequestParams<FileDeleteBody> = {}) {
    return post("/file/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取回收站列表 GET /recycle/list */
export async function getRecycleList(params: API.RequestParams<RecycleListQuery, RecycleListData> = {}) {
    return get<RecycleListData>("/recycle/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取回收站统计 GET /recycle/count */
export async function getRecycleCount(params: API.RequestParams<never, FileCountData> = {}) {
    return get<FileCountData>("/recycle/count", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 恢复回收站文件 POST /recycle/restore */
export async function restoreRecycleItem(params: API.RequestParams<RecycleRestoreBody> = {}) {
    return post("/recycle/restore", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 彻底删除回收站文件 POST /recycle/delete */
export async function deleteRecycleItem(params: API.RequestParams<RecycleItemBody> = {}) {
    return post("/recycle/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 清空回收站 POST /recycle/clear */
export async function clearRecycle(params: API.RequestParams = {}) {
    return post("/recycle/clear", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取系统任务 GET /system/task */
export async function getSystemTask(params: API.RequestParams<SystemTaskQuery, SystemTaskRecord> = {}) {
    return get<SystemTaskRecord>("/system/task", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 取消系统任务 POST /system/task */
export async function cancelSystemTask(params: API.RequestParams<SystemTaskQuery> = {}) {
    return post("/system/task", {...params?.body, action: "cancel"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 文件上传请求 GET|POST /file/upload */
export function requestFileUpload<T>(
    endpoint: string,
    action: "upload" | "check" | "merge" | "clear",
    signal: AbortSignal,
    params: {
        body?: FormData;
        query?: Record<string, string>;
        onProgress?: (loaded: number, total: number) => void;
    },
) {
    return new Promise<API.Response<T>>((resolve, reject) => {
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
                reject(new FileUploadTransportError(`上传失败（HTTP ${xhr.status}）`));
                return;
            }
            try {
                const response = JSON.parse(xhr.responseText) as API.Response<T>;
                if (!response || typeof response.code !== "number") {
                    reject(new FileUploadRequestError("服务器返回的数据格式不正确"));
                    return;
                }
                if (response.code === 403 || response.code === 503) {
                    deleteHeader();
                    historyPush("login");
                }
                resolve(response);
            } catch {
                reject(new FileUploadRequestError("服务器返回的数据格式不正确"));
            }
        };
        xhr.onerror = () => {
            signal.removeEventListener("abort", abortRequest);
            reject(new FileUploadTransportError("上传文件失败，请检查网络连接"));
        };
        xhr.ontimeout = () => {
            signal.removeEventListener("abort", abortRequest);
            reject(new FileUploadTransportError("上传文件超时"));
        };
        xhr.onabort = () => {
            signal.removeEventListener("abort", abortRequest);
            reject(new DOMException("上传已取消", "AbortError"));
        };
        signal.addEventListener("abort", abortRequest, {once: true});
        xhr.send(params.body);
    });
}
