import {useMemo} from "react";
import {createStyles} from "antd-style";
import dayjs from "dayjs";
import {formatSize} from "@/utils/string";
import useChartTooltip from "@/pages/components/ui/chart-tooltip";
import type {ChartTooltipModel, ChartTooltipOptions} from "@/pages/components/ui/chart-tooltip";
import type {MonitorHistory, MonitorProcessMetric, MonitorSeries} from "../../types";

type MonitorTooltipModel = ChartTooltipModel & {
    process?: {
        rows: {name: string; pid: string; value: string}[];
        time: string;
        state: string;
        widths: {pid: number; value: number};
    };
};

const useStyles = createStyles(({css, token}) => ({
    tooltip: css`
        .process-section { box-sizing: border-box; height: 172px; border-top: 1px solid ${token.colorBorderSecondary}; margin-top: 8px; padding-top: 8px; }
        .process-heading { display: flex; justify-content: space-between; gap: 8px; font-weight: 500; }
        .process-heading span:last-child { color: ${token.colorTextSecondary}; font-size: 11px; font-weight: 400; }
        .process-time { color: ${token.colorTextTertiary}; font-size: 11px; line-height: 16px; margin-top: 3px; white-space: nowrap; }
        .process-state { padding: 14px 0; color: ${token.colorTextSecondary}; }
        .process-list { display: grid; grid-template-columns: var(--process-pid-width) minmax(0, 1fr) var(--process-value-width); align-items: center; column-gap: 8px; margin: 6px 0 0; padding: 0; list-style: none; }
        .process-row { display: contents; line-height: 24px; }
        .process-pid { color: ${token.colorTextSecondary}; font-size: 11px; white-space: nowrap; }
        .process-name { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
        .process-value { text-align: right; font-weight: 500; white-space: nowrap; }
    `,
}));

const labels = {
    cpu: {title: "CPU 进程 Top 5", valueLabel: "整机占比"},
    memory: {title: "内存进程 Top 5", valueLabel: "RSS"},
    disk_io: {title: "磁盘 I/O 进程 Top 5", valueLabel: "读写总速率"},
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

export default function useTooltip(metric: MonitorProcessMetric | undefined, history: MonitorHistory | undefined, series: MonitorSeries[], unit: "percent" | "bytes" | "load") {
    const {styles} = useStyles();
    const models = useMemo(() => {
        const models = new Map<number, MonitorTooltipModel>();
        const widths = {pid: 0, value: 0};
        // 随历史数据一次格式化；悬浮时只读取成品，不排序或请求。
        history?.data.forEach((point) => {
            const group = metric && point.samples > 0 ? point.process?.[metric] : undefined;
            // 旧记录没有配对高点；内存和 I/O 曲线为区间均值，只有 CPU 与峰值曲线校验。
            const matched = typeof group?.peak === "number" && Number.isFinite(group.peak) && (metric !== "cpu" || group.peak === point.cpu_max);
            const rows = (matched ? group?.items : undefined)?.slice(0, 5).map((item) => ({
                name: item.name || "未知进程", pid: `PID ${item.pid}`,
                value: !Number.isFinite(item.value) ? "不可用" : metric === "cpu"
                    ? `${numberLabel(item.value)}%` : `${formatSize(item.value)}${metric === "disk_io" ? "/s" : ""}`,
            })) || [];
            rows.forEach((row) => {
                widths.pid = Math.max(widths.pid, row.pid.length);
                widths.value = Math.max(widths.value, row.value.length);
            });
            const highPoint = metric === "cpu" ? "峰值" : "高点";
            const peakLabel = rows.length && group ? metric === "cpu" ? "峰值时刻"
                : `高点 ${metric === "memory" ? `${numberLabel(group.peak!)}%` : `${formatSize(group.peak!)}/s`}` : "";
            models.set(point.time, {
                title: `${dayjs(point.time).format("YYYY-MM-DD HH:mm")}${point.samples === 0 ? "（无采样）" : ""}`,
                values: series.map((item) => {
                    const value = point[item.key];
                    if (!Number.isFinite(value)) return "无数据";
                    return unit === "bytes" ? `${formatSize(value)}/s` : `${numberLabel(value)}${unit === "percent" ? "%" : ""}`;
                }),
                process: metric ? {
                    rows,
                    widths,
                    time: rows.length ? `${peakLabel} · ${group && Number.isFinite(group.time) && group.time > 0 ? dayjs(group.time).format(metric === "cpu" ? "YYYY-MM-DD HH:mm:ss" : "HH:mm:ss") : "时间未知"}${group?.partial ? "（部分进程）" : ""}` : "",
                    state: point.samples === 0 ? "该时段无采样" : history.process_error ? "进程记录暂不可用"
                        : matched ? `${highPoint}时刻的进程信息不可用` : `该时段未记录${highPoint}对应的进程`,
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
                const heading = element("div", "process-heading");
                heading.append(element("span", "", label.title), element("span", "", label.valueLabel));
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
                const state = element("div", "process-state");
                section.append(heading, time, list, state);
                container.append(section);
                let widths: {pid: number; value: number} | undefined;
                return (model) => {
                    const info = model.process;
                    const hasRows = Boolean(info?.rows.length);
                    visible(time, hasRows);
                    visible(list, hasRows);
                    visible(state, !hasRows);
                    text(hasRows ? time : state, hasRows ? info!.time : info?.state || "该时段暂无进程记录");
                    if (info && widths !== info.widths) {
                        widths = info.widths;
                        section.style.setProperty("--process-pid-width", `${widths.pid + 1}ch`);
                        section.style.setProperty("--process-value-width", `${widths.value + 1}ch`);
                    }
                    rows.forEach((row, index) => {
                        const item = info?.rows[index];
                        visible(row.node, Boolean(item));
                        if (!item) return;
                        text(row.name, item.name);
                        text(row.pid, item.pid);
                        text(row.value, item.value);
                    });
                };
            },
        };
    }, [metric, styles.tooltip]);
    return useChartTooltip(models, series, options);
}
