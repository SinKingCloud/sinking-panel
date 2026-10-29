import type {FileHashWorkerRequest, FileHashWorkerResponse} from "./hash.worker";

export const getFileMd5 = (
    file: Blob,
    signal?: AbortSignal,
    onProgress?: (loaded: number, total: number) => void,
): Promise<string> => new Promise((resolve, reject) => {
    if (signal?.aborted) {
        reject(new DOMException("上传已取消", "AbortError"));
        return;
    }

    let worker: Worker;
    try {
        worker = new Worker(new URL("./hash.worker.ts", import.meta.url));
    } catch (error) {
        reject(error);
        return;
    }

    let settled = false;
    const cleanup = () => {
        signal?.removeEventListener("abort", abort);
        worker.onmessage = null;
        worker.onerror = null;
        worker.onmessageerror = null;
        worker.terminate();
    };
    const fail = (error: unknown) => {
        if (settled) return;
        settled = true;
        cleanup();
        reject(error);
    };
    // 终止 Worker 会立即中断正在计算的 MD5。
    const abort = () => fail(new DOMException("上传已取消", "AbortError"));
    worker.onmessage = ({data}: MessageEvent<FileHashWorkerResponse>) => {
        if (settled) return;
        if (data.type === "progress") {
            try {
                onProgress?.(data.loaded, data.total);
            } catch (error) {
                fail(error);
            }
        } else if (data.type === "complete") {
            settled = true;
            cleanup();
            resolve(data.hash);
        } else if (data.type === "error") {
            const error = new Error(data.message);
            error.name = data.name;
            fail(error);
        }
    };
    worker.onerror = (event) => {
        event.preventDefault();
        fail(new Error(event.message || "文件检查失败"));
    };
    worker.onmessageerror = () => fail(new Error("读取文件检查结果失败"));
    signal?.addEventListener("abort", abort, {once: true});
    try {
        const request: FileHashWorkerRequest = {file};
        // 传递 Blob，由 Worker 按 2 MiB 分片读取。
        worker.postMessage(request);
    } catch (error) {
        fail(error);
    }
});
