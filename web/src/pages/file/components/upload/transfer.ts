import defaultSettings from "@/../config/defaultSettings";
import {
    requestFileUpload as requestFileUploadApi,
} from "@/service/api/file";
import {getFileMd5} from "./hash";

const isUploadError = (error: unknown, name: string) => (
    Boolean(error && typeof error === "object" && "name" in error && (error as {name?: unknown}).name === name)
);

const createUploadTransferError = (name: string, message: string) => {
    const error = new Error(message);
    error.name = name;
    return error;
};

const apiUrl = (path: string) => {
    const origin = window.location.origin;
    const gateway = new URL(defaultSettings?.gateway || "/", origin);
    const gatewayPath = gateway.pathname.replace(/\/+$/, "");
    gateway.pathname = `${gatewayPath}/${path.replace(/^\/+/, "")}`.replace(/\/{2,}/g, "/");
    gateway.search = "";
    gateway.hash = "";
    return gateway.toString();
};

const requestFileUpload = <T = any, >(
    action: "upload" | "check" | "merge" | "clear",
    signal: AbortSignal,
    params: {
        body?: FormData;
        query?: Record<string, string>;
        onProgress?: (loaded: number, total: number) => void;
    },
) => requestFileUploadApi<T>(apiUrl("/file/upload"), signal, {
    ...params,
    query: {
        ...params.query,
        action,
        // 分片检查每次都读取服务器的最新状态。
        ...(action === "check" ? {_: `${Date.now()}-${Math.random().toString(36).slice(2)}`} : {}),
    },
});

const FILE_UPLOAD_CHUNK_THRESHOLD = 2 * 1024 * 1024;
const FILE_UPLOAD_CHUNK_SIZE = 8 * 1024 * 1024;
const FILE_UPLOAD_MAX_RETRIES = 3;
const FILE_UPLOAD_CONCURRENCY = 2;
const FILE_UPLOAD_RESUME_STORAGE_KEY = "file-upload-tasks";
const FILE_UPLOAD_RESUME_LEGACY_STORAGE_PREFIX = "file-upload-task:";
const FILE_UPLOAD_RESUME_MAX_AGE = 30 * 24 * 60 * 60 * 1000;
const FILE_UPLOAD_RESUME_CLEANUP_INTERVAL = 60 * 60 * 1000;

let activeFileUploads = 0;
let lastFileUploadResumeCleanup = 0;
const pendingFileUploads: any[] = [];

const runPendingFileUploads = () => {
    while (activeFileUploads < FILE_UPLOAD_CONCURRENCY && pendingFileUploads.length > 0) {
        const item = pendingFileUploads.shift()!;
        if (item.signal.aborted) {
            item.removeAbortListener?.();
            item.reject(createAbortError());
            continue;
        }
        activeFileUploads++;
        item.removeAbortListener?.();
        void item.run().catch(item.reject).finally(() => {
            activeFileUploads--;
            runPendingFileUploads();
        });
    }
};

const scheduleFileUpload = <T, >(run: () => Promise<T>, signal: AbortSignal) => new Promise<T>((resolve, reject) => {
    const item: any = {
        run: async () => resolve(await run()),
        signal,
        reject,
    };
    const handleAbort = () => {
        const index = pendingFileUploads.indexOf(item);
        if (index >= 0) {
            pendingFileUploads.splice(index, 1);
            reject(createAbortError());
        }
    };
    signal.addEventListener("abort", handleAbort, {once: true});
    item.removeAbortListener = () => signal.removeEventListener("abort", handleAbort);
    pendingFileUploads.push(item);
    runPendingFileUploads();
});

class FileUploadBusinessError extends Error {
    response: API.Response<unknown>;

    constructor(response: API.Response<unknown>) {
        super(response.message || "上传失败");
        this.name = "FileUploadBusinessError";
        this.response = response;
    }
}

const createAbortError = () => new DOMException("上传已取消", "AbortError");

const ensureUploadNotAborted = (signal: AbortSignal) => {
    if (signal.aborted) {
        throw createAbortError();
    }
};

const appendUploadMeta = (body: FormData, meta: any) => {
    body.append("path", meta.path);
    body.append("file_name", meta.fileName);
    body.append("total_size", String(meta.totalSize));
    body.append("chunk_size", String(meta.chunkSize));
    body.append("total_chunks", String(meta.totalChunks));
    if (meta.fileHash) {
        body.append("file_hash", meta.fileHash);
    }
};

const isValidFileUploadResumeItem = (value: unknown, now: number): value is any => {
    if (!value || typeof value !== "object") {
        return false;
    }
    const item = value as any;
    return typeof item.uploadId === "string"
        && typeof item.updateTime === "number"
        && Number.isFinite(item.updateTime)
        && now - item.updateTime <= FILE_UPLOAD_RESUME_MAX_AGE;
};

const readFileUploadResumeItems = (now = Date.now()) => {
    const items: Record<string, any> = {};
    const raw = localStorage.getItem(FILE_UPLOAD_RESUME_STORAGE_KEY);
    if (!raw) {
        return {items, changed: false};
    }
    try {
        const stored = JSON.parse(raw) as Record<string, unknown> | null;
        if (!stored || typeof stored !== "object" || Array.isArray(stored)) {
            return {items, changed: true};
        }
        let changed = false;
        Object.entries(stored).forEach(([key, value]) => {
            if (isValidFileUploadResumeItem(value, now)) {
                items[key] = value;
            } else {
                changed = true;
            }
        });
        return {items, changed};
    } catch {
        return {items, changed: true};
    }
};

const writeFileUploadResumeItems = (items: Record<string, any>) => {
    try {
        if (Object.keys(items).length === 0) {
            localStorage.removeItem(FILE_UPLOAD_RESUME_STORAGE_KEY);
        } else {
            localStorage.setItem(FILE_UPLOAD_RESUME_STORAGE_KEY, JSON.stringify(items));
        }
        return true;
    } catch {
        return false;
    }
};

const cleanupExpiredFileUploadResumeItems = () => {
    const now = Date.now();
    if (now - lastFileUploadResumeCleanup < FILE_UPLOAD_RESUME_CLEANUP_INTERVAL) {
        return;
    }
    lastFileUploadResumeCleanup = now;
    try {
        const stored = readFileUploadResumeItems(now);
        const items = stored.items;
        let changed = stored.changed;
        const legacyStorageKeys: string[] = [];
        for (let index = localStorage.length - 1; index >= 0; index--) {
            const storageKey = localStorage.key(index);
            if (!storageKey?.startsWith(FILE_UPLOAD_RESUME_LEGACY_STORAGE_PREFIX)) {
                continue;
            }
            legacyStorageKeys.push(storageKey);
            try {
                const item = JSON.parse(localStorage.getItem(storageKey) || "null") as any;
                if (!isValidFileUploadResumeItem(item, now)) {
                    continue;
                }
                const key = storageKey.slice(FILE_UPLOAD_RESUME_LEGACY_STORAGE_PREFIX.length);
                if (!items[key] || items[key].updateTime < item.updateTime) {
                    items[key] = item;
                    changed = true;
                }
            } catch {
            }
        }
        if ((!changed || writeFileUploadResumeItems(items))) {
            legacyStorageKeys.forEach((storageKey) => localStorage.removeItem(storageKey));
        }
    } catch {
    }
};

const getFileUploadResumeId = (key: string) => {
    try {
        const stored = readFileUploadResumeItems();
        if (stored.changed) {
            writeFileUploadResumeItems(stored.items);
        }
        return stored.items[key]?.uploadId;
    } catch {
    }
    return undefined;
};

const setFileUploadResumeId = (key: string, uploadId: string) => {
    try {
        const stored = readFileUploadResumeItems();
        stored.items[key] = {uploadId, updateTime: Date.now()};
        writeFileUploadResumeItems(stored.items);
    } catch {
    }
};

const removeFileUploadResumeId = (key: string, expectedUploadId: string) => {
    try {
        const stored = readFileUploadResumeItems();
        if (stored.items[key]?.uploadId === expectedUploadId) {
            delete stored.items[key];
            writeFileUploadResumeItems(stored.items);
        } else if (stored.changed) {
            writeFileUploadResumeItems(stored.items);
        }
    } catch {
    }
};

const createFileUploadId = () => {
    if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
        return crypto.randomUUID();
    }
    return `${Date.now()}-${Math.random().toString(36).slice(2)}-${Math.random().toString(36).slice(2)}`;
};

const buildFileUploadResumeKey = (options: any, fileName: string, chunkSize: number, fileHash: string) => JSON.stringify({
    gateway: apiUrl("/file/upload"),
    path: options.path,
    name: fileName,
    size: options.file.size,
    lastModified: (options.file as File).lastModified || 0,
    chunkSize,
    fileHash,
});

const requestFileUploadWithRetry = async <T, >(
    action: "check" | "merge",
    signal: AbortSignal,
    buildParams: () => {body?: FormData; query?: Record<string, string>},
) => {
    for (let attempt = 1; attempt <= FILE_UPLOAD_MAX_RETRIES + 1; attempt++) {
        try {
            return await requestFileUpload<T>(action, signal, buildParams());
        } catch (error) {
            if ((!isUploadError(error, "FileUploadRequestError") && !isUploadError(error, "FileUploadTransportError")) || attempt > FILE_UPLOAD_MAX_RETRIES || signal.aborted) {
                throw error;
            }
            await waitForFileUploadRetry(attempt, signal);
        }
    }
    throw createUploadTransferError("FileUploadTransportError", "上传文件失败，请检查网络连接");
};

const waitForFileUploadRetry = (attempt: number, signal: AbortSignal) => new Promise<void>((resolve, reject) => {
    if (signal.aborted) {
        reject(createAbortError());
        return;
    }
    const timeout = window.setTimeout(() => {
        signal.removeEventListener("abort", cancel);
        resolve();
    }, 500 * 2 ** (attempt - 1));
    const cancel = () => {
        window.clearTimeout(timeout);
        reject(createAbortError());
    };
    signal.addEventListener("abort", cancel, {once: true});
});

const getFileUploadChunkSize = (index: number, meta: any) => {
    const start = index * meta.chunkSize;
    return Math.max(0, Math.min(meta.chunkSize, meta.totalSize - start));
};

const uploadFileDirectly = async (
    options: any,
    fileName: string,
    signal: AbortSignal,
) => {
    const meta: any = {
        path: options.path,
        fileName,
        totalSize: options.file.size,
        chunkSize: options.file.size,
        totalChunks: 1,
    };
    const body = new FormData();
    appendUploadMeta(body, meta);
    body.append("file", options.file, fileName);
    const response = await requestFileUpload<any>("upload", signal, {
        body,
        onProgress: (loaded, total) => {
            const current = total > 0 ? Math.min(meta.totalSize, loaded / total * meta.totalSize) : Math.min(loaded, meta.totalSize);
            options.onProgress?.({
                loaded: current,
                total: meta.totalSize,
                percent: meta.totalSize > 0 ? current / meta.totalSize * 100 : 0,
                phase: "uploading",
                chunkIndex: 0,
                totalChunks: 1,
            });
        },
    });
    if (response.code === 200) {
        options.onProgress?.({loaded: meta.totalSize, total: meta.totalSize, percent: 100, phase: "complete"});
    }
    return response;
};

const uploadFileChunk = async (
    options: any,
    fileName: string,
    uploadId: string,
    chunkIndex: number,
    uploadedBytes: number,
    meta: any,
    signal: AbortSignal,
    report: (loaded: number, attempt: number) => void,
) => {
    const start = chunkIndex * meta.chunkSize;
    const end = Math.min(start + meta.chunkSize, meta.totalSize);
    const chunkSize = end - start;
    for (let attempt = 1; attempt <= FILE_UPLOAD_MAX_RETRIES + 1; attempt++) {
        ensureUploadNotAborted(signal);
        const body = new FormData();
        appendUploadMeta(body, meta);
        body.append("upload_id", uploadId);
        body.append("chunk_index", String(chunkIndex));
        body.append("file", options.file.slice(start, end), fileName);
        try {
            const response = await requestFileUpload("upload", signal, {
                body,
                onProgress: (loaded, total) => {
                    const current = total > 0 ? loaded / total * chunkSize : Math.min(loaded, chunkSize);
                    report(uploadedBytes + Math.min(current, chunkSize), attempt);
                },
            });
            if (response.code !== 200) {
                throw new FileUploadBusinessError(response);
            }
            const data = response.data;
            const completed = data?.name === fileName && data?.size === meta.totalSize && data?.file_hash === meta.fileHash;
            if (!completed && (data?.upload_id !== uploadId || data?.chunk_index !== chunkIndex)) {
                throw createUploadTransferError("FileUploadRequestError", "服务器返回的分片信息不一致");
            }
            return;
        } catch (error) {
            if ((!isUploadError(error, "FileUploadRequestError") && !isUploadError(error, "FileUploadTransportError")) || attempt > FILE_UPLOAD_MAX_RETRIES || signal.aborted) {
                throw error;
            }
            await waitForFileUploadRetry(attempt, signal);
        }
    }
};

const uploadFileByChunks = async (
    options: any,
    fileName: string,
    signal: AbortSignal,
): Promise<API.Response<any>> => {
    ensureUploadNotAborted(signal);
    options.onProgress?.({
        loaded: 0,
        total: options.file.size,
        percent: 0,
        phase: "checking",
    });
    const fileHash = (await getFileMd5(options.file, signal)).toLowerCase();
    const meta: any = {
        path: options.path,
        fileName,
        totalSize: options.file.size,
        chunkSize: FILE_UPLOAD_CHUNK_SIZE,
        totalChunks: Math.ceil(options.file.size / FILE_UPLOAD_CHUNK_SIZE),
        fileHash,
    };
    const resumeKey = buildFileUploadResumeKey(options, fileName, meta.chunkSize, fileHash);
    let uploadId = getFileUploadResumeId(resumeKey) || createFileUploadId();
    setFileUploadResumeId(resumeKey, uploadId);
    ensureUploadNotAborted(signal);
    options.onProgress?.({
        loaded: 0,
        total: meta.totalSize,
        percent: 0,
        phase: "checking",
        totalChunks: meta.totalChunks,
    });
    const buildCheckQuery = (): Record<string, string> => ({
        path: meta.path,
        file_name: meta.fileName,
        total_size: String(meta.totalSize),
        chunk_size: String(meta.chunkSize),
        total_chunks: String(meta.totalChunks),
        upload_id: uploadId,
        file_hash: fileHash,
    });
    let displayedBytes = 0;
    const report = (loaded: number, chunkIndex?: number, attempt?: number) => {
        displayedBytes = Math.max(displayedBytes, Math.min(loaded, meta.totalSize));
        options.onProgress?.({
            loaded: displayedBytes,
            total: meta.totalSize,
            percent: displayedBytes / meta.totalSize * 100,
            phase: "uploading",
            chunkIndex,
            totalChunks: meta.totalChunks,
            attempt,
        });
    };
    // 上传或合并时任务失效、分片丢失都重新检查，最多补传三轮，复用已计算的哈希。
    for (let repair = 0; repair <= FILE_UPLOAD_MAX_RETRIES; repair++) {
        ensureUploadNotAborted(signal);
        try {
            const checkResponse = await requestFileUploadWithRetry<any>("check", signal, () => ({query: buildCheckQuery()}));
            if (checkResponse.code !== 200) {
                throw new FileUploadBusinessError(checkResponse);
            }
            if (checkResponse.data?.completed && checkResponse.data.file) {
                removeFileUploadResumeId(resumeKey, uploadId);
                options.onProgress?.({loaded: meta.totalSize, total: meta.totalSize, percent: 100, phase: "complete"});
                return {...checkResponse, data: checkResponse.data.file};
            }
            const uploadedChunks = new Set<number>(
                (checkResponse.data?.uploaded_chunks || []).filter((index: number) => Number.isInteger(index) && index >= 0 && index < meta.totalChunks),
            );
            let uploadedBytes = Array.from(uploadedChunks).reduce(
                (total, index) => total + getFileUploadChunkSize(index, meta),
                0,
            );
            report(uploadedBytes);
            for (let chunkIndex = 0; chunkIndex < meta.totalChunks; chunkIndex++) {
                if (uploadedChunks.has(chunkIndex)) {
                    continue;
                }
                await uploadFileChunk(options, fileName, uploadId, chunkIndex, uploadedBytes, meta, signal, (loaded, attempt) => {
                    report(loaded, chunkIndex, attempt);
                });
                uploadedBytes += getFileUploadChunkSize(chunkIndex, meta);
                report(uploadedBytes, chunkIndex);
            }
            ensureUploadNotAborted(signal);
            options.onProgress?.({
                loaded: meta.totalSize,
                total: meta.totalSize,
                percent: 100,
                phase: "merging",
                totalChunks: meta.totalChunks,
            });
            const mergeResponse = await requestFileUploadWithRetry<any>("merge", signal, () => {
                const mergeBody = new FormData();
                appendUploadMeta(mergeBody, meta);
                mergeBody.append("upload_id", uploadId);
                return {body: mergeBody};
            });
            if (mergeResponse.code !== 200) {
                throw new FileUploadBusinessError(mergeResponse);
            }
            removeFileUploadResumeId(resumeKey, uploadId);
            options.onProgress?.({
                loaded: meta.totalSize,
                total: meta.totalSize,
                percent: 100,
                phase: "complete",
                totalChunks: meta.totalChunks,
            });
            return mergeResponse;
        } catch (error) {
            ensureUploadNotAborted(signal);
            if (!(error instanceof FileUploadBusinessError)) {
                throw error;
            }
            const message = error.message;
            if (message === "合并文件校验失败，请重新上传") {
                removeFileUploadResumeId(resumeKey, uploadId);
            }
            const newSession = ["上传任务信息不一致", "上传完成记录不合法", "已上传文件不存在或不完整", "已上传文件校验失败"].includes(message);
            if (repair >= FILE_UPLOAD_MAX_RETRIES || (!newSession && message !== "上传任务不存在，请重新检查"
                && !/^分片 \d+ 不存在或不完整$/.test(message))) {
                throw error;
            }
            if (newSession) {
                // 其他标签页可能仍在使用旧任务，只更换当前断点，不清理服务器上的任务。
                uploadId = createFileUploadId();
                setFileUploadResumeId(resumeKey, uploadId);
            }
        }
    }
    throw createUploadTransferError("FileUploadRequestError", "分片补传失败，请重试");
};

/** 上传文件 POST /file/upload */
export function uploadFile(options: any): any {
    cleanupExpiredFileUploadResumeItems();
    const controller = new AbortController();
    let settled = false;

    const abort = () => {
        if (!settled) {
            controller.abort();
        }
    };

    const fileName = options.fileName || (options.file as File)?.name || "upload";
    const promise = scheduleFileUpload(
        () => options.file.size > FILE_UPLOAD_CHUNK_THRESHOLD
            ? uploadFileByChunks(options, fileName, controller.signal)
            : uploadFileDirectly(options, fileName, controller.signal),
        controller.signal,
    ).catch((error) => {
        if (error instanceof FileUploadBusinessError) {
            return error.response as API.Response<any>;
        }
        throw error;
    }).finally(() => {
        settled = true;
    });

    return {promise, abort};
}
