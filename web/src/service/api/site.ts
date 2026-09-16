import {get, post} from "@/utils/request";

/** 获取网站列表 GET /site/list */
export async function getSiteList(params: API.RequestParams = {}) {
    return get("/site/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站详情 GET /site/info */
export async function getSiteInfo(params: API.RequestParams = {}) {
    return get("/site/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 创建网站 POST /site/create */
export async function createSite(params: API.RequestParams = {}) {
    return post("/site/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站 POST /site/update */
export async function updateSite(params: API.RequestParams = {}) {
    return post("/site/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除网站 POST /site/delete */
export async function deleteSite(params: API.RequestParams = {}) {
    return post("/site/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 启用网站 POST /site/enable */
export async function enableSite(params: API.RequestParams = {}) {
    return post("/site/enable", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 停用网站 POST /site/disable */
export async function disableSite(params: API.RequestParams = {}) {
    return post("/site/disable", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取或清理网站日志 GET|POST /site/log */
export async function getSiteLog(params: API.RequestParams = {}) {
    const request = params?.body?.action === "clear" ? post : get;
    return request("/site/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站域名 GET /site/domain */
export async function getSiteDomain(params: API.RequestParams = {}) {
    return get("/site/domain", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站域名 POST /site/domain */
export async function updateSiteDomain(params: API.RequestParams = {}) {
    return post("/site/domain", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站 SSL 配置 GET /site/ssl */
export async function getSiteSSL(params: API.RequestParams = {}) {
    return get("/site/ssl", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站 SSL 配置 POST /site/ssl */
export async function updateSiteSSL(params: API.RequestParams = {}) {
    return post("/site/ssl", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站 WAF 配置 GET /site/waf */
export async function getSiteWAF(params: API.RequestParams = {}) {
    return get("/site/waf", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站 WAF 配置 POST /site/waf */
export async function updateSiteWAF(params: API.RequestParams = {}) {
    return post("/site/waf", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站缓存配置 GET /site/cache */
export async function getSiteCache(params: API.RequestParams = {}) {
    return get("/site/cache", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站缓存配置 POST /site/cache */
export async function updateSiteCache(params: API.RequestParams = {}) {
    return post("/site/cache", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 清理网站缓存 POST /site/cache */
export async function clearSiteCache(params: API.RequestParams = {}) {
    return post("/site/cache", {...params?.body, action: "clear"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站访问频率限制 GET /site/rate */
export async function getSiteRate(params: API.RequestParams = {}) {
    return get("/site/rate", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站访问频率限制 POST /site/rate */
export async function updateSiteRate(params: API.RequestParams = {}) {
    return post("/site/rate", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站流量限制 GET /site/traffic */
export async function getSiteTraffic(params: API.RequestParams = {}) {
    return get("/site/traffic", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站流量限制 POST /site/traffic */
export async function updateSiteTraffic(params: API.RequestParams = {}) {
    return post("/site/traffic", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站请求头和响应头配置 GET /site/header */
export async function getSiteHeader(params: API.RequestParams = {}) {
    return get("/site/header", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站请求头和响应头配置 POST /site/header */
export async function updateSiteHeader(params: API.RequestParams = {}) {
    return post("/site/header", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站响应压缩配置 GET /site/compression */
export async function getSiteCompression(params: API.RequestParams = {}) {
    return get("/site/compression", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站响应压缩配置 POST /site/compression */
export async function updateSiteCompression(params: API.RequestParams = {}) {
    return post("/site/compression", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站重定向配置 GET /site/redirect */
export async function getSiteRedirect(params: API.RequestParams = {}) {
    return get("/site/redirect", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站重定向配置 POST /site/redirect */
export async function updateSiteRedirect(params: API.RequestParams = {}) {
    return post("/site/redirect", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站自定义路由 GET /site/route */
export async function getSiteRoute(params: API.RequestParams = {}) {
    return get("/site/route", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站自定义路由 POST /site/route */
export async function updateSiteRoute(params: API.RequestParams = {}) {
    return post("/site/route", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取静态网站配置 GET /site/static */
export async function getSiteStatic(params: API.RequestParams = {}) {
    return get("/site/static", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改静态网站配置 POST /site/static */
export async function updateSiteStatic(params: API.RequestParams = {}) {
    return post("/site/static", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站反向代理配置 GET /site/proxy */
export async function getSiteProxy(params: API.RequestParams = {}) {
    return get("/site/proxy", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站反向代理配置 POST /site/proxy */
export async function updateSiteProxy(params: API.RequestParams = {}) {
    return post("/site/proxy", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站 FastCGI 配置 GET /site/fastcgi */
export async function getSiteFastCGI(params: API.RequestParams = {}) {
    return get("/site/fastcgi", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站 FastCGI 配置 POST /site/fastcgi */
export async function updateSiteFastCGI(params: API.RequestParams = {}) {
    return post("/site/fastcgi", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取通用网站进程配置 GET /site/process */
export async function getSiteProcess(params: API.RequestParams = {}) {
    return get("/site/process", {...params?.body, action: "get"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改通用网站进程配置 POST /site/process */
export async function updateSiteProcess(params: API.RequestParams = {}) {
    return post("/site/process", {...params?.body, action: "set"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取通用网站进程状态 GET /site/process */
export async function getSiteProcessStatus(params: API.RequestParams = {}) {
    return get("/site/process", {...params?.body, action: "status"}, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 启动、停止或重启通用网站进程 POST /site/process */
export async function controlSiteProcess(params: API.RequestParams = {}) {
    return post("/site/process", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
