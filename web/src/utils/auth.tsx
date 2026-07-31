const tokenKey = 'token';

export const loginDevice: API.LoginDevice = 'web';

/**
 * 设置登录token
 * @param token
 */
export function setLoginToken(token: string) {
    localStorage.setItem(tokenKey, token);
}

/**
 * 获取登录token
 * @returns {string}
 */
export function getLoginToken() {
    return localStorage.getItem(tokenKey) || '';
}

/**
 * 删除token
 */
export function deleteHeader() {
    localStorage.removeItem(tokenKey);
}

/**
 * 返回请求header
 * @returns {{}}
 */
export function getHeaders() {
    const headers: Record<string, string> = {
        device: loginDevice,
    };
    const token = getLoginToken();
    if (token) {
        headers.token = token;
    }
    return headers;
}
