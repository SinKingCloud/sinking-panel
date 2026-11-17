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

/** 读取网站日志 POST /site/log/list */
export async function getSiteLog(params: API.RequestParams = {}) {
    return post("/site/log/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 清理网站日志 POST /site/log/clear */
export async function clearSiteLog(params: API.RequestParams = {}) {
    return post("/site/log/clear", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站域名 POST /site/domain/list */
export async function getSiteDomainList(params: API.RequestParams = {}) {
    return post("/site/domain/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站域名 POST /site/domain/update */
export async function updateSiteDomain(params: API.RequestParams = {}) {
    return post("/site/domain/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站 SSL 配置 POST /site/ssl/info */
export async function getSiteSSLInfo(params: API.RequestParams = {}) {
    return post("/site/ssl/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站 SSL 配置 POST /site/ssl/update */
export async function updateSiteSSL(params: API.RequestParams = {}) {
    return post("/site/ssl/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站 WAF 配置 POST /site/waf/info */
export async function getSiteWAFInfo(params: API.RequestParams = {}) {
    return post("/site/waf/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站 WAF 配置 POST /site/waf/update */
export async function updateSiteWAF(params: API.RequestParams = {}) {
    return post("/site/waf/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站缓存配置 POST /site/cache/info */
export async function getSiteCacheInfo(params: API.RequestParams = {}) {
    return post("/site/cache/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站缓存配置 POST /site/cache/update */
export async function updateSiteCache(params: API.RequestParams = {}) {
    return post("/site/cache/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 清理网站缓存 POST /site/cache/clear */
export async function clearSiteCache(params: API.RequestParams = {}) {
    return post("/site/cache/clear", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站访问频率限制 POST /site/rate/info */
export async function getSiteRateLimitInfo(params: API.RequestParams = {}) {
    return post("/site/rate/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站访问频率限制 POST /site/rate/update */
export async function updateSiteRateLimit(params: API.RequestParams = {}) {
    return post("/site/rate/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站流量限制 POST /site/traffic/info */
export async function getSiteTrafficLimitInfo(params: API.RequestParams = {}) {
    return post("/site/traffic/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站流量限制 POST /site/traffic/update */
export async function updateSiteTrafficLimit(params: API.RequestParams = {}) {
    return post("/site/traffic/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站请求头和响应头配置 POST /site/header/info */
export async function getSiteHeadersInfo(params: API.RequestParams = {}) {
    return post("/site/header/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站请求头和响应头配置 POST /site/header/update */
export async function updateSiteHeaders(params: API.RequestParams = {}) {
    return post("/site/header/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站响应压缩配置 POST /site/compression/info */
export async function getSiteCompressionInfo(params: API.RequestParams = {}) {
    return post("/site/compression/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站响应压缩配置 POST /site/compression/update */
export async function updateSiteCompression(params: API.RequestParams = {}) {
    return post("/site/compression/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站重定向配置 POST /site/redirect/list */
export async function getSiteRedirectsList(params: API.RequestParams = {}) {
    return post("/site/redirect/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站重定向配置 POST /site/redirect/update */
export async function updateSiteRedirects(params: API.RequestParams = {}) {
    return post("/site/redirect/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站自定义路由 POST /site/route/list */
export async function getSiteRoutesList(params: API.RequestParams = {}) {
    return post("/site/route/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站自定义路由 POST /site/route/update */
export async function updateSiteRoutes(params: API.RequestParams = {}) {
    return post("/site/route/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取静态网站配置 POST /site/static/info */
export async function getSiteStaticInfo(params: API.RequestParams = {}) {
    return post("/site/static/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改静态网站配置 POST /site/static/update */
export async function updateSiteStatic(params: API.RequestParams = {}) {
    return post("/site/static/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站反向代理配置 POST /site/proxy/info */
export async function getSiteProxyInfo(params: API.RequestParams = {}) {
    return post("/site/proxy/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站反向代理配置 POST /site/proxy/update */
export async function updateSiteProxy(params: API.RequestParams = {}) {
    return post("/site/proxy/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站 FastCGI 配置 POST /site/fastcgi/info */
export async function getSiteFastCGIInfo(params: API.RequestParams = {}) {
    return post("/site/fastcgi/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站 FastCGI 配置 POST /site/fastcgi/update */
export async function updateSiteFastCGI(params: API.RequestParams = {}) {
    return post("/site/fastcgi/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站进程配置 POST /site/process/info */
export async function getSiteProcessInfo(params: API.RequestParams = {}) {
    return post("/site/process/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改网站进程配置 POST /site/process/update */
export async function updateSiteProcess(params: API.RequestParams = {}) {
    return post("/site/process/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站进程状态 POST /site/process/status */
export async function getSiteProcessStatus(params: API.RequestParams = {}) {
    return post("/site/process/status", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 启动网站进程 POST /site/process/start */
export async function startSiteProcess(params: API.RequestParams = {}) {
    return post("/site/process/start", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 停止网站进程 POST /site/process/stop */
export async function stopSiteProcess(params: API.RequestParams = {}) {
    return post("/site/process/stop", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 重启网站进程 POST /site/process/restart */
export async function restartSiteProcess(params: API.RequestParams = {}) {
    return post("/site/process/restart", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取网站备份 GET /site/backup/list */
export async function getSiteBackups(params: API.RequestParams = {}) {
    return get("/site/backup/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取备份设置 POST /site/backup/settings/info */
export async function getSiteBackupSettingsInfo(params: API.RequestParams = {}) {
    return post("/site/backup/settings/info", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 修改备份设置 POST /site/backup/settings/update */
export async function updateSiteBackupSettings(params: API.RequestParams = {}) {
    return post("/site/backup/settings/update", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 读取网站自动备份日志 POST /site/backup/log/list */
export async function getSiteBackupLog(params: API.RequestParams = {}) {
    return post("/site/backup/log/list", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 清理网站自动备份日志 POST /site/backup/log/clear */
export async function clearSiteBackupLog(params: API.RequestParams = {}) {
    return post("/site/backup/log/clear", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 提交所选网站的备份任务 POST /site/backup/create */
export async function createSiteBackup(params: API.RequestParams = {}) {
    return post("/site/backup/create", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 删除网站备份 POST /site/backup/delete */
export async function deleteSiteBackups(params: API.RequestParams = {}) {
    return post("/site/backup/delete", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 获取备份的临时下载地址 POST /site/backup/download */
export async function downloadSiteBackup(params: API.RequestParams = {}) {
    return post("/site/backup/download", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}

/** 提交网站恢复 POST /site/backup/restore */
export async function restoreSiteBackup(params: API.RequestParams = {}) {
    return post("/site/backup/restore", params?.body, params?.onSuccess, params?.onFail, params?.onFinally);
}
