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

/**
 * 格式化文件大小
 */
export const formatSize = (size: any) => {
    const value = Number(size || 0);
    if (value <= 0) {
        return "0 B";
    }
    const units = ["B", "KB", "MB", "GB", "TB"];
    let nextValue = value;
    let index = 0;
    while (nextValue >= 1024 && index < units.length - 1) {
        nextValue = nextValue / 1024;
        index += 1;
    }
    return `${nextValue.toFixed(index === 0 ? 0 : 2)} ${units[index]}`;
};
