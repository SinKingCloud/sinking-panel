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

/** 网站日志 POST /site/log */
export async function getSiteLog(params: API.RequestParams = {}) {
    return post("/site/log", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站域名 POST /site/domain */
export async function siteDomain(params: API.RequestParams = {}) {
    return post("/site/domain", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站 SSL 配置 POST /site/ssl */
export async function siteSSL(params: API.RequestParams = {}) {
    return post("/site/ssl", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站 WAF 配置 POST /site/waf */
export async function siteWAF(params: API.RequestParams = {}) {
    return post("/site/waf", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站缓存配置 POST /site/cache */
export async function siteCache(params: API.RequestParams = {}) {
    return post("/site/cache", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站访问频率限制 POST /site/rate */
export async function siteRateLimit(params: API.RequestParams = {}) {
    return post("/site/rate", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站流量限制 POST /site/traffic */
export async function siteTrafficLimit(params: API.RequestParams = {}) {
    return post("/site/traffic", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站请求头和响应头配置 POST /site/header */
export async function siteHeaders(params: API.RequestParams = {}) {
    return post("/site/header", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站响应压缩配置 POST /site/compression */
export async function siteCompression(params: API.RequestParams = {}) {
    return post("/site/compression", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站重定向配置 POST /site/redirect */
export async function siteRedirects(params: API.RequestParams = {}) {
    return post("/site/redirect", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站自定义路由 POST /site/route */
export async function siteRoutes(params: API.RequestParams = {}) {
    return post("/site/route", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 静态网站配置 POST /site/static */
export async function siteStatic(params: API.RequestParams = {}) {
    return post("/site/static", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站反向代理配置 POST /site/proxy */
export async function siteProxy(params: API.RequestParams = {}) {
    return post("/site/proxy", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 网站 FastCGI 配置 POST /site/fastcgi */
export async function siteFastCGI(params: API.RequestParams = {}) {
    return post("/site/fastcgi", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 通用网站进程 POST /site/process */
export async function siteProcess(params: API.RequestParams = {}) {
    return post("/site/process", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
