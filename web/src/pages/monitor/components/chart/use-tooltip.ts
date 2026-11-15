import {useMemo} from "react";
import {createStyles} from "antd-style";
import dayjs from "dayjs";
import {formatSize} from "@/utils/string";
import useChartTooltip from "@/pages/components/ui/chart-tooltip";
import type {ChartTooltipModel, ChartTooltipOptions} from "@/pages/components/ui/chart-tooltip";
import type {MonitorHistory, MonitorProcessMetric, MonitorSeries, MonitorUnit} from "../../types";

type MonitorTooltipModel = ChartTooltipModel & {
    process?: {
        title: string;
        valueLabel: string;
        rows: {name: string; pid: string; value: string}[];
        time: string;
        widths: {pid: number; value: number};
    };
};

const useStyles = createStyles(({css, token}) => ({
    tooltip: css`
        .process-section { box-sizing: border-box; height: 172px; border-top: 1px solid ${token.colorBorderSecondary}; margin-top: 8px; padding-top: 8px; }
        .process-heading { display: flex; justify-content: space-between; gap: 8px; font-weight: 500; }
        .process-heading span:last-child { color: ${token.colorTextSecondary}; font-size: 11px; font-weight: 400; }
        .process-time { color: ${token.colorTextTertiary}; font-size: 11px; line-height: 16px; margin-top: 3px; white-space: nowrap; }
        .process-list { display: grid; grid-template-columns: var(--process-pid-width) minmax(0, 1fr) var(--process-value-width); align-items: center; column-gap: 8px; margin: 6px 0 0; padding: 0; list-style: none; }
        .process-row { display: contents; line-height: 24px; }
        .process-pid { color: ${token.colorTextSecondary}; font-size: 11px; white-space: nowrap; }
        .process-name { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
        .process-value { text-align: right; font-weight: 500; white-space: nowrap; }
    `,
}));

const labels = {
    cpu: {title: "CPU 进程", valueLabel: "整机占比 · 均值"},
    memory: {title: "内存进程", valueLabel: "RSS"},
    disk_io: {title: "磁盘 I/O 进程", valueLabel: "读写均值"},
    process_io: {title: "进程总 I/O", valueLabel: "总 I/O 均值"},
};
const numberLabel = new Intl.NumberFormat("zh-CN", {maximumFractionDigits: 2}).format;

function element<K extends keyof HTMLElementTagNameMap>(tag: K, className = "", value?: string) {
    const node = document.createElement(tag);
    node.className = className;
    if (value !== undefined) node.textContent = value;
    return node;
}

function text(node: HTMLElement, value: string) {
    if (node.textContent !== value) node.textContent = value;
}

function visible(node: HTMLElement, show: boolean) {
    if (node.hidden === show) node.hidden = !show;
}

export default function useTooltip(metric: MonitorProcessMetric | undefined, history: MonitorHistory | undefined, series: MonitorSeries[], unit: MonitorUnit) {
    const {styles} = useStyles();
    const models = useMemo(() => {
        const models = new Map<number, MonitorTooltipModel>();
        const widths = {pid: 0, value: 0};
        // 随历史数据一次格式化；悬浮时只读取成品，不排序或请求。
        history?.data.forEach((point) => {
            const group = metric && point.samples > 0 ? point.process?.[metric] : undefined;
            // 进程低频采样，CPU/I/O 展示自身窗口的均值，不与每秒整机峰值混用。
            const rows = (Array.isArray(group?.items) ? group.items : []).slice(0, 5).map((item) => ({
                name: item.name || "未知进程", pid: `PID ${item.pid}`,
                value: !Number.isFinite(item.value) ? "不可用" : metric === "cpu"
                    ? `${numberLabel(item.value)}%` : `${formatSize(item.value)}${metric === "disk_io" ? "/s" : ""}`,
            }));
            rows.forEach((row) => {
                widths.pid = Math.max(widths.pid, row.pid.length);
                widths.value = Math.max(widths.value, row.value.length);
            });
            const peakValue = typeof group?.peak === "number" && Number.isFinite(group.peak)
                ? metric === "disk_io" ? `${formatSize(group.peak)}/s` : `${numberLabel(group.peak)}%` : "";
            const peakLabel = metric === "memory" ? "采样时刻" : metric === "disk_io" && group?.io_scope === "process" ? "磁盘采样均值" : "采样均值";
            models.set(point.time, {
                title: dayjs(point.time).format("YYYY-MM-DD HH:mm"),
                values: series.map((item) => {
                    const value = point[item.key];
                    if (typeof value !== "number" || !Number.isFinite(value)) return "--";
                    const metricUnit = item.unit || unit;
                    if (metricUnit === "iops") return `${numberLabel(Math.round(value))} 次/s`;
                    return metricUnit === "bytes" ? `${formatSize(value)}/s` : `${numberLabel(value)}${metricUnit === "percent" ? "%" : ""}`;
                }),
                process: metric && rows.length ? {
                    ...labels[metric === "disk_io" && group?.io_scope === "process" ? "process_io" : metric],
                    rows,
                    widths,
                    time: rows.length ? `${peakLabel}${peakValue ? ` ${peakValue}` : ""} · ${group && Number.isFinite(group.time) && group.time > 0 ? dayjs(group.time).format("HH:mm:ss") : "时间未知"}` : "",
                } : undefined,
            });
        });
        return models;
    }, [metric, history, series, unit]);
    const options = useMemo<ChartTooltipOptions<MonitorTooltipModel> | undefined>(() => {
        if (!metric) return;
        const label = labels[metric];
        return {
            width: 288,
            className: styles.tooltip,
            // 只在绑定时创建节点；换采样点仅更新文本，刷新数据也不重建渲染器。
            renderContent: (container) => {
                const section = element("div", "process-section");
                section.hidden = true;
                const heading = element("div", "process-heading");
                const title = element("span", "", label.title);
                const valueLabel = element("span", "", label.valueLabel);
                heading.append(title, valueLabel);
                const time = element("div", "process-time");
                const list = element("ol", "process-list");
                const rows = Array.from({length: 5}, () => {
                    const node = element("li", "process-row");
                    const pid = element("span", "process-pid");
                    const name = element("span", "process-name");
                    const value = element("span", "process-value");
                    node.append(pid, name, value);
                    list.append(node);
                    return {node, name, pid, value};
                });
                section.append(heading, time, list);
                container.append(section);
                let widths: {pid: number; value: number} | undefined;
                return (model) => {
                    const info = model.process;
                    const sizeChanged = section.hidden === Boolean(info);
                    visible(section, Boolean(info));
                    if (!info) return sizeChanged;
                    text(title, info.title);
                    text(valueLabel, info.valueLabel);
                    text(time, info.time);
                    if (widths !== info.widths) {
                        widths = info.widths;
                        section.style.setProperty("--process-pid-width", `${widths.pid + 1}ch`);
                        section.style.setProperty("--process-value-width", `${widths.value + 1}ch`);
                    }
                    rows.forEach((row, index) => {
                        const item = info.rows[index];
                        visible(row.node, Boolean(item));
                        if (!item) return;
                        text(row.name, item.name);
                        text(row.pid, item.pid);
                        text(row.value, item.value);
                    });
                    return sizeChanged;
                };
            },
        };
    }, [metric, styles.tooltip]);
    return useChartTooltip(models, series, options);
}
