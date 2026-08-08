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
