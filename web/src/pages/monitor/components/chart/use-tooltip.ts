import {useMemo} from "react";
import dayjs from "dayjs";
import {formatSize} from "@/utils/string";
import useChartTooltip, {ChartTooltipModel} from "@/pages/components/ui/chart-tooltip";
import type {MonitorHistory, MonitorProcessMetric, MonitorSeries} from "../../types";

const labels = {
    cpu: {title: "CPU 进程 Top 5", valueLabel: "整机占比"},
    memory: {title: "内存进程 Top 5", valueLabel: "RSS"},
    disk_io: {title: "磁盘 I/O 进程 Top 5", valueLabel: "读写总速率"},
};
const numberLabel = new Intl.NumberFormat("zh-CN", {maximumFractionDigits: 2}).format;

export default function useTooltip(metric: MonitorProcessMetric | undefined, history: MonitorHistory | undefined, series: MonitorSeries[], unit: "percent" | "bytes" | "load") {
    const models = useMemo(() => {
        const models = new Map<number, ChartTooltipModel>();
        // 随历史数据一次格式化；悬浮时只读取成品，不排序或请求。
        history?.data.forEach((point) => {
            const group = metric && point.samples > 0 ? point.process?.[metric] : undefined;
            const rows = group?.items.slice(0, 5).map((item) => ({
                name: item.name || "未知进程", pid: `PID ${item.pid}`,
                value: !Number.isFinite(item.value) ? "不可用" : metric === "cpu"
                    ? `${numberLabel(item.value)}%` : `${formatSize(item.value)}${metric === "disk_io" ? "/s" : ""}`,
            })) || [];
            models.set(point.time, {
                title: `${dayjs(point.time).format("YYYY-MM-DD HH:mm")}${point.samples === 0 ? "（无采样）" : ""}`,
                values: series.map((item) => {
                    const value = point[item.key];
                    if (!Number.isFinite(value)) return "无数据";
                    return unit === "bytes" ? `${formatSize(value)}/s` : `${numberLabel(value)}${unit === "percent" ? "%" : ""}`;
                }),
                process: metric ? {
                    rows,
                    time: rows.length ? `进程占用高点 · ${group && Number.isFinite(group.time) && group.time > 0 ? dayjs(group.time).format("YYYY-MM-DD HH:mm:ss") : "时间未知"}` : "",
                    state: point.samples === 0 ? "该时段无采样" : history.process_error ? "进程记录暂不可用" : "该时段暂无进程记录",
                } : undefined,
            });
        });
        return models;
    }, [metric, history, series, unit]);
    return useChartTooltip(models, series, metric ? labels[metric] : undefined);
}
