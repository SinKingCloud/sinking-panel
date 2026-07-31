/**
 * 获取随机字符
 * @param length 长度
 */
export function getRandStr(length = 16) {
    let str = 'abcdefghijklmnopqrstuvwxyz';
    str += str.toUpperCase();
    str += '0123456789'
    let _str = '';
    for (let i = 0; i < length; i++) {
        const rand = Math.floor(Math.random() * str.length);
        _str += str[rand];
    }
    return _str
}
