import request, {get, post} from "@/utils/request";
import {getHeaders} from "@/utils/auth";

/** 软件列表。 */
export async function getSoftwareList(params: API.RequestParams = {}) {
    return get("/software/list", params.body, params.onSuccess, params.onFail, params.onFinally);
}

/** 软件详情。 */
export async function getSoftwareInfo(params: API.RequestParams = {}) {
    return get("/software/info", params.body, params.onSuccess, params.onFail, params.onFinally);
}

/** 安装服务器上的软件包或源码目录。 */
export async function installSoftware(params: API.RequestParams = {}) {
    return post("/software/install", params.body, params.onSuccess, params.onFail, params.onFinally,
        undefined, {timeout: 660000});
}

/** 修改软件名称。 */
export async function updateSoftware(params: API.RequestParams = {}) {
    return post("/software/update", params.body, params.onSuccess, params.onFail, params.onFinally);
}

/** 卸载软件并等待完成，进度仍可在后台任务查看。 */
export async function uninstallSoftware(params: API.RequestParams = {}) {
    return post("/software/uninstall", params.body, params.onSuccess, params.onFail, params.onFinally,
        undefined, {timeout: 660000});
}

/** 按软件标识调用管理接口，业务参数直接传入请求体。 */
export async function callSoftware(code: string, method: string, params: API.RequestParams = {}) {
    // 镜像导入和首次数据库初始化可能超过普通接口的一分钟限制。
    const timeout = ["setup", "start", "stop", "restart", "reload"].includes(method) ? 660000 : 60000;
    return post(`/software/${encodeURIComponent(code)}/${encodeURIComponent(method)}`, params.body, params.onSuccess, params.onFail, params.onFinally,
        undefined, {timeout});
}

/** 图标走已有资源接口，以请求头鉴权，失败由展示组件回退默认图标。 */
export async function getSoftwareIcon(code: string, icon: string, signal: AbortSignal): Promise<Blob> {
    const path = icon.startsWith("web/") ? icon.slice(4) : "";
    if (!code || !path || path.split("/").some((part) => !part || part === "." || part === "..")) {
        throw new Error("软件图标路径无效");
    }
    const blob = await request<Blob>(`/software/${encodeURIComponent(code)}/page/${path.split("/").map(encodeURIComponent).join("/")}`, {
        method: "GET", headers: getHeaders(), responseType: "blob", signal,
        errorHandler: (error) => { throw error; },
    });
    if (!(blob instanceof Blob) || !blob.type.startsWith("image/")) throw new Error("软件图标内容无效");
    return blob;
}

/** 读取软件前端入口或相对 web 目录的资源，沿用面板鉴权。 */
export async function getSoftwarePage(code: string, path?: string) {
    const url = `/software/${encodeURIComponent(code)}/page${path ? `/${path.split("/").map(encodeURIComponent).join("/")}` : ""}`;
    const result = await request(url, {
        method: "GET", headers: getHeaders(), responseType: "text", getResponse: true,
        errorHandler: (error: any) => { throw error; },
    });
    const response = result?.response;
    const location = response?.headers?.get("Content-Location");
    if (!location) {
        let failure;
        try { failure = JSON.parse(result?.data); } catch {}
        const message = typeof failure?.code === "number" && failure.code !== 200 && typeof failure?.message === "string"
            ? failure.message : "";
        throw new Error(message || "软件资源未返回实际文件路径，请更新面板服务后重试");
    }
    const prefix = `/software/${code}/page/`;
    const resolved = decodeURIComponent(new URL(location, response?.url).pathname);
    if (!resolved.startsWith(prefix)) throw new Error("软件资源未返回实际文件路径，请更新面板服务后重试");
    if (typeof result?.data !== "string") throw new Error("软件资源内容无效");
    return {path: `web/${resolved.slice(prefix.length)}`, source: result?.data};
}
