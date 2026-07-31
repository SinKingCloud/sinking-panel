export function ago(string: string) {
    if (!string) {
        return;
    }
    const f = string.split(' ', 2);
    const d = (f?.[0] || '').replace(/\//g, '-').split('-', 3);
    const t = (f?.[1] || '').split(':', 3);
    const dateTimeStamp = new Date(
        parseInt(d[0], 10) || 0,
        (parseInt(d[1], 10) || 1) - 1,
        parseInt(d[2], 10) || 0,
        parseInt(t[0], 10) || 0,
        parseInt(t[1], 10) || 0,
        parseInt(t[2], 10) || 0,
    ).getTime();
    const minute = 1000 * 60;
    const hour = minute * 60;
    const day = hour * 24;
    const week = day * 7;
    const month = day * 30;
    if (!Number.isFinite(dateTimeStamp)) {
        return;
    }
    const diffValue = Date.now() - dateTimeStamp;
    const minC = diffValue / minute;
    const hourC = diffValue / hour;
    const dayC = diffValue / day;
    const weekC = diffValue / week;
    const monthC = diffValue / month;
    let result;
    if (diffValue <= minute) {
        result = "刚刚"
    } else if (minC < 60) {
        result = parseInt(String(minC)) + "分钟前"
    } else if (hourC < 24) {
        result = parseInt(String(hourC)) + "小时前"
    } else if (dayC < 7) {
        result = parseInt(String(dayC)) + "天前"
    } else if (dayC < 30) {
        result = parseInt(String(weekC)) + "周前"
    } else if (monthC < 6) {
        result = parseInt(String(monthC)) + "月前"
    } else {
        const datetime = new Date();
        datetime.setTime(dateTimeStamp);
        const Nyear = datetime.getFullYear();
        const Nmonth = datetime.getMonth() + 1 < 10 ? "0" + (datetime.getMonth() + 1) : datetime.getMonth() + 1;
        const Ndate = datetime.getDate() < 10 ? "0" + datetime.getDate() : datetime.getDate();
        const Nhour = datetime.getHours() < 10 ? "0" + datetime.getHours() : datetime.getHours();
        const Nminute = datetime.getMinutes() < 10 ? "0" + datetime.getMinutes() : datetime.getMinutes();
        const Nsecond = datetime.getSeconds() < 10 ? "0" + datetime.getSeconds() : datetime.getSeconds();
        result = Nyear + "-" + Nmonth + "-" + Ndate + ' ' + Nhour + ":" + Nminute + ":" + Nsecond;
    }
    return result;
}
