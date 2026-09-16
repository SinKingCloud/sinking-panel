// 根据目标地址识别上游协议，UNIX 套接字统一转换为服务端使用的监听地址。
export const parseProxyTarget = (value: unknown, defaultScheme = "http") => {
    const target = String(value || "").trim();
    if (!target || /[\r\n\0]/.test(target)) throw new Error("请输入有效的目标地址");
    const unix = /^unix(?:\+h2c)?\//i.test(target) || target.startsWith("/");
    if (unix) {
        const path = target.replace(/^unix(?:\+h2c)?\//i, "");
        if (!path.startsWith("/") || path === "/") throw new Error("UNIX 套接字必须填写绝对文件路径，例如 /tmp/panel.sock");
        return {dial: `unix/${path}`, scheme: "http", hostname: "", h2c: /^unix\+h2c\//i.test(target)};
    }
    try {
        const address = target.includes("://") ? target : `${defaultScheme}://${target}`;
        if (!/^https?:\/\/[^/\\]/i.test(address) || target.includes("\\")) throw new Error();
        const url = new URL(address);
        if (!url.hostname || url.username || url.password || url.pathname !== "/" || url.search || url.hash) throw new Error();
        return {dial: `${url.protocol}//${url.host}`, scheme: url.protocol.slice(0, -1), hostname: url.hostname, h2c: false};
    } catch {
        throw new Error("请输入 HTTP、HTTPS 地址或 UNIX 套接字路径，网址不能包含路径和参数");
    }
};
