const numberPart = (value: string, min: number, max: number) => {
    if (!/^\d+$/.test(value)) {
        return undefined;
    }
    const number = Number(value);
    return number >= min && number <= max ? number : undefined;
};

const stepPart = (value: string, max: number) => {
    const matched = value.match(/^\*\/([1-9]\d*)$/);
    if (!matched) {
        return undefined;
    }
    const number = Number(matched[1]);
    return number <= max ? number : undefined;
};

export const parseSchedule = (value: any): any => {
    const parts = String(value || "").trim().split(/\s+/);
    if (parts.length !== 6) {
        return {mode: "custom"};
    }

    const [second, minute, hour, day, month, weekday] = parts;
    const secondStep = second === "*" ? 1 : stepPart(second, 59);
    if (secondStep && minute === "*" && hour === "*" && day === "*" && month === "*" && weekday === "*") {
        return {mode: "second", interval: secondStep};
    }

    const minuteStep = minute === "*" ? 1 : stepPart(minute, 59);
    if (second === "0" && minuteStep && hour === "*" && day === "*" && month === "*" && weekday === "*") {
        return {mode: "minute", interval: minuteStep};
    }

    const minuteNumber = numberPart(minute, 0, 59);
    const hourStep = hour === "*" ? 1 : stepPart(hour, 23);
    if (second === "0" && minuteNumber !== undefined && hourStep && day === "*" && month === "*" && weekday === "*") {
        return {mode: "hour", interval: hourStep, minute: minuteNumber};
    }

    const hourNumber = numberPart(hour, 0, 23);
    if (second !== "0" || minuteNumber === undefined || hourNumber === undefined || month !== "*") {
        return {mode: "custom"};
    }
    if (day === "*" && weekday === "*") {
        return {mode: "day", minute: minuteNumber, hour: hourNumber};
    }

    const weekdayNumber = numberPart(weekday, 0, 6);
    if (day === "*" && weekdayNumber !== undefined) {
        return {mode: "week", minute: minuteNumber, hour: hourNumber, weekday: weekdayNumber};
    }

    const dayNumber = numberPart(day, 1, 31);
    if (dayNumber !== undefined && weekday === "*") {
        return {mode: "month", minute: minuteNumber, hour: hourNumber, day: dayNumber};
    }
    return {mode: "custom"};
};

const weekdayNames = ["周日", "周一", "周二", "周三", "周四", "周五", "周六"];
const pad = (value: any) => String(Number(value || 0)).padStart(2, "0");

export const describeSchedule = (schedule: any): string => {
    const time = `${pad(schedule.hour)}:${pad(schedule.minute)}`;
    switch (schedule.mode) {
        case "second":
            return `每 ${schedule.interval} 秒执行一次`;
        case "minute":
            return `每 ${schedule.interval} 分钟执行一次`;
        case "hour":
            return schedule.minute
                ? `每 ${schedule.interval} 小时的第 ${schedule.minute} 分钟执行`
                : `每 ${schedule.interval} 小时执行一次`;
        case "day":
            return `每天 ${time} 执行`;
        case "week":
            return `每${weekdayNames[schedule.weekday || 0]} ${time} 执行`;
        case "month":
            return `每月 ${schedule.day} 日 ${time} 执行`;
        default:
            return "自定义执行周期";
    }
};

export const describeTaskSchedule = (value: any) => describeSchedule(parseSchedule(value));

export const parseRequest = (value: any): any => {
    let data = value;
    if (typeof value === "string") {
        try {
            data = JSON.parse(value);
        } catch {
            data = {url: value};
        }
    }
    if (!data || typeof data !== "object" || Array.isArray(data)) {
        data = {};
    }
    const headers = data.headers && typeof data.headers === "object" && !Array.isArray(data.headers)
        ? Object.entries(data.headers).map(([name, headerValue]) => ({name, value: String(headerValue ?? "")}))
        : [];
    return {
        method: String(data.method || "GET").trim().toUpperCase(),
        url: String(data.url || ""),
        headers,
        body: String(data.body || ""),
    };
};

export const stringifyRequest = (value: any): string => {
    const headers: any = {};
    const names = new Set<string>();
    (value?.headers || []).forEach((header: any) => {
        const name = String(header?.name || "").trim();
        const key = name.toLowerCase();
        if (!name) {
            throw new Error("请求头名称不能为空");
        }
        if (names.has(key)) {
            throw new Error("请求头名称不能重复");
        }
        names.add(key);
        const headerValue = String(header?.value ?? "");
        if (key === "content-length" || key === "transfer-encoding" || key === "trailer") {
            throw new Error("不支持自定义传输层请求头");
        }
        if (key === "host" && !headerValue.trim()) {
            throw new Error("Host请求头不能为空");
        }
        if (/\r|\n/.test(headerValue)) {
            throw new Error("请求头值不能换行");
        }
        if (/^[^\u0000-\u0008\u000A-\u001F\u007F]*$/.test(headerValue) === false) {
            throw new Error("请求头值包含非法字符");
        }
        headers[name] = headerValue;
    });
    return JSON.stringify({
        method: String(value?.method || "GET").trim().toUpperCase(),
        url: String(value?.url || "").trim(),
        headers,
        body: String(value?.body ?? ""),
    });
};
