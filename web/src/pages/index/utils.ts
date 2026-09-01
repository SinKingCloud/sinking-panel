export const gutter = 12;

export const numberValue = (value: any): number => {
    const number = Number(value || 0);
    return Number.isFinite(number) ? number : 0;
};

export const percentValue = (value: any): number => Math.min(100, Math.max(0, numberValue(value)));

export const formatPercent = (value: any): string => `${Number(percentValue(value).toFixed(1))}%`;

const countFormatter = new Intl.NumberFormat("zh-CN", {
    maximumFractionDigits: 1,
});

export const formatCount = (value: any): string => countFormatter.format(numberValue(value));

export const formatUptime = (value: any): string => {
    const seconds = Math.max(0, Math.floor(numberValue(value)));
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor(seconds % 86400 / 3600);
    const minutes = Math.floor(seconds % 3600 / 60);
    if (days > 0) {
        return `${days} 天 ${hours} 小时`;
    }
    if (hours > 0) {
        return `${hours} 小时 ${minutes} 分钟`;
    }
    return `${minutes} 分钟`;
};
