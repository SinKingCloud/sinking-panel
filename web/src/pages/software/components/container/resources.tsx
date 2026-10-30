import {memo, useEffect, useMemo, useRef, useState} from "react";
import {Line} from "@ant-design/charts";
import {Alert, Grid, Spin, Tooltip, theme as antdTheme} from "antd";
import {createStyles} from "antd-style";
import {useTheme} from "sinking-antd";
import dayjs from "dayjs";
import {getContainerStatus} from "@/service/api/container";
import useChartTooltip from "@/pages/components/ui/chart-tooltip";
import type {ChartTooltipModel} from "@/pages/components/ui/chart-tooltip";

import {formatSize} from "@/utils/string";

const SAMPLE_INTERVAL = 1000;
const MAX_SAMPLES = 500;
const byteUnits = ["B", "KB", "MB", "GB", "TB", "PB"];
const validNumber = (value: unknown): value is number => typeof value === "number" && Number.isFinite(value) && value >= 0;
const timeText = (value: unknown) => {
    const numeric = Number(value);
    const date = new Date(Number.isFinite(numeric) ? numeric : String(value));
    return Number.isFinite(date.getTime()) ? dayjs(date).format("HH:mm:ss") : "—";
};
const secondsText = (value: number) => `${Math.floor(value / 3600)} 小时 ${Math.floor(value % 3600 / 60)} 分 ${Math.floor(value % 60)} 秒`;

const useStyles = createStyles(({css, token}: any) => ({
    resources: css`
        min-width: 0;
        font-size: 12px;
        line-height: 1.5;
        .resource-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 6px; }
        .resource-chart { min-width: 0; padding: 6px 8px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; }
        .resource-chart-wide { grid-column: 1 / -1; }
        .resource-chart-header { display: flex; align-items: baseline; justify-content: space-between; flex-wrap: wrap; gap: 2px 8px; margin-bottom: 2px; }
        .resource-chart-title { margin: 0; color: ${token.colorTextSecondary}; font-size: 11px; font-weight: 400; }
        .resource-readings { display: flex; align-items: center; flex-wrap: wrap; gap: 2px 10px; }
        .resource-reading { display: inline-flex; align-items: center; gap: 4px; color: ${token.colorTextSecondary}; font-size: inherit; }
        .resource-reading i { display: inline-block; width: 6px; height: 6px; flex: none; border-radius: 50%; }
        .resource-reading strong { color: ${token.colorText}; font-size: inherit; font-weight: 400; font-variant-numeric: tabular-nums; }
        .resource-plot { width: 100%; min-width: 0; }
        .resource-empty { display: flex; align-items: center; justify-content: center; gap: ${token.marginXS}px; color: ${token.colorTextTertiary}; font-size: 11px; }
        .resource-details { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; gap: 6px; margin-top: 6px; }
        .resource-metric-group { min-width: 0; padding: 6px 8px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .resource-metric-title { margin: 0 0 4px; color: ${token.colorTextSecondary}; font-size: 11px; font-weight: 400; }
        .resource-metrics { display: grid; gap: 2px; margin: 0; }
        .resource-metric { display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: baseline; gap: 8px; min-width: 0; line-height: 16px; }
        .resource-metric dt { min-width: 0; color: ${token.colorTextSecondary}; font-size: 11px; }
        .resource-metric dd { min-width: 0; max-width: 100%; margin: 0; color: ${token.colorText}; text-align: right; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
        @media (max-width: 575px) {
            .resource-grid, .resource-details { grid-template-columns: minmax(0, 1fr); }
        }
    `,
}));

const ResourceChart = memo(({title, data, series, readings, kind, dark, wide, emptyText, loading, height, active}: {
    title: string; data: {time: number; value: number; series: string}[]; series: {label: string; color: string}[]; readings: string[];
    kind: "percent" | "bytes" | "rate"; dark: boolean; wide?: boolean; emptyText: string; loading?: boolean; height: number;
    active: boolean;
}) => {
    const {token} = antdTheme.useToken();
    const mobile = Grid.useBreakpoint().md === false;
    const containerRef = useRef<HTMLDivElement>(null);
    const chartRef = useRef<any>(null);
    const appliedRef = useRef<{instance: any; config: any; data: typeof data} | undefined>(undefined);
    const hasData = data.length > 0;
    const singlePoint = data.length <= series.length;
    const prepared = useMemo(() => {
        const models = new Map<number, ChartTooltipModel>();
        const indexes = new Map(series.map((item, index) => [item.label, index]));
        let maximum = 0;
        data.forEach((item) => {
            maximum = Math.max(maximum, item.value);
            let model = models.get(item.time);
            if (!model) {
                model = {title: timeText(item.time), values: series.map(() => "—")};
                models.set(item.time, model);
            }
            model.values[indexes.get(item.series)!] = kind === "percent" ? `${item.value.toFixed(1)}%`
                : `${formatSize(item.value)}${kind === "rate" ? "/s" : ""}`;
        });
        return {models, maximum};
    }, [data, kind, series]);
    const tooltip = useChartTooltip(prepared.models, series);
    useEffect(() => {
        if (!active || !hasData || !chartRef.current?.chart) return;
        return tooltip.bind(chartRef.current.chart, {zIndex: token.zIndexPopupBase + 100});
    }, [active, hasData, tooltip, token.zIndexPopupBase]);

    useEffect(() => {
        const container = containerRef.current;
        if (!container || typeof ResizeObserver === "undefined") return;
        let width = Math.round(container.getBoundingClientRect().width);
        let timer = 0;
        const observer = new ResizeObserver(([entry]) => {
            const nextWidth = Math.round(entry.contentRect.width);
            if (nextWidth <= 0 || nextWidth === width) return;
            width = nextWidth;
            window.clearTimeout(timer);
            timer = window.setTimeout(() => { void chartRef.current?.chart?.forceFit?.(); }, 80);
        });
        observer.observe(container);
        return () => { window.clearTimeout(timer); observer.disconnect(); };
    }, []);

    let unitIndex = 0;
    while (kind !== "percent" && unitIndex < byteUnits.length - 1 && prepared.maximum >= 1024 ** (unitIndex + 1)) unitIndex += 1;
    const domainMax = prepared.maximum <= (kind === "percent" ? 1 : 0) ? 1 : undefined;
    const config = useMemo<any>(() => {
        const unit = kind === "percent" ? "%" : `${byteUnits[unitIndex]}${kind === "rate" ? "/s" : ""}`;
        return {
            data,
            height,
            xField: "time",
            yField: "value",
            colorField: "series",
            seriesField: "series",
            keyField: "series",
            shape: "smooth",
            margin: 0,
            paddingTop: 4,
            paddingRight: 8,
            paddingBottom: 22,
            paddingLeft: 45,
            inset: 0,
            scale: {
                x: {type: "time", nice: false},
                color: {domain: series.map((item) => item.label), range: series.map((item) => item.color)},
                y: {domainMin: 0, domainMax, nice: true},
            },
            axis: {
                x: {
                    title: false, line: false, tick: false, tickCount: wide ? 5 : 3,
                    labelFill: token.colorTextTertiary, labelFontSize: 10, labelSpacing: 4,
                    labelAutoHide: true, labelAutoRotate: false, labelFormatter: timeText, grid: false,
                },
                y: {
                    title: false, line: false, tick: false, tickCount: 3,
                    labelFill: token.colorTextTertiary, labelFontSize: 10, labelSpacing: 4,
                    labelFormatter: (value: unknown) => `${(Number(value) / (kind === "percent" ? 1 : 1024 ** unitIndex)).toLocaleString("zh-CN", {maximumFractionDigits: 1})}${kind === "percent" ? "" : " "}${unit}`,
                    grid: true, gridLineDash: [4, 4], gridLineWidth: 1,
                    gridStroke: token.colorBorderSecondary, gridStrokeOpacity: 1,
                },
            },
            legend: false,
            tooltip: {
                title: false,
                items: [(datum: any) => ({
                    name: datum?.series,
                    value: datum?.value,
                    time: datum?.time,
                })],
            },
            interaction: {tooltip: {
                shared: true, series: true, crosshairs: true,
                body: false, wait: 0, disableAutoHide: mobile,
            }},
            theme: {type: dark ? "classicDark" : "classic"},
            autoFit: true,
            style: {lineWidth: 1.5, lineCap: "round", lineJoin: "round"},
            point: singlePoint ? {sizeField: 2, shapeField: "circle"} : undefined,
            animate: false,
        };
    }, [hasData, dark, height, kind, series, unitIndex, domainMax, singlePoint, mobile, token.colorBorderSecondary, token.colorTextTertiary, wide]);
    useEffect(() => {
        const instance = chartRef.current?.chart;
        if (!hasData || !instance) return;
        const previous = appliedRef.current;
        appliedRef.current = {instance, config, data};
        // 和首页一致，普通采样仅更新数据，复用图表配置及悬浮事件。
        if (!previous || previous.instance !== instance || previous.config !== config || previous.data === data) return;
        void instance.changeData(data);
    }, [config, data, hasData]);
    const plot = useMemo(() => <Line ref={chartRef} {...config}/>, [config]);

    return <section ref={containerRef} className={`resource-chart${wide ? " resource-chart-wide" : ""}`} aria-label={`${title}趋势`}>
        <div className="resource-chart-header">
            <h3 className="resource-chart-title"><Tooltip title={kind === "percent" ? "100% 表示占用 1 个 CPU 核心，多核使用率可超过 100%" : undefined}>
                <span tabIndex={kind === "percent" ? 0 : undefined}>{title}</span>
            </Tooltip></h3>
            <div className="resource-readings">{series.map((item, index) => <span className="resource-reading" key={item.label}>
                <i style={{background: item.color}}/>{series.length > 1 && item.label}
                <strong>{readings[index]}</strong>
            </span>)}</div>
        </div>
        {hasData ? <div className="resource-plot" style={{height}}>{plot}</div> : <div className="resource-empty" style={{height}}>
            {loading ? <><Spin size="small"/><span>等待采样</span></> : emptyText}
        </div>}
    </section>;
});

ResourceChart.displayName = "ResourceChart";

const ContainerResources = ({id, active = true, visible = true, onStatus}: {
    id: string; active?: boolean; visible?: boolean; onStatus?: (status: any) => void;
}) => {
    const {styles} = useStyles();
    const {token} = antdTheme.useToken();
    const theme = useTheme();
    const dark = Boolean(theme?.isDarkTheme?.());
    const chartHeight = theme?.isCompactTheme?.() ? 100 : 112;
    const [status, setStatus] = useState<any>();
    const [rates, setRates] = useState<any>();
    const [history, setHistory] = useState<any[]>([]);
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(true);
    const callbackRef = useRef(onStatus);
    const requestRef = useRef<ReturnType<typeof getContainerStatus> | null>(null);
    callbackRef.current = onStatus;

    useEffect(() => {
        if (!active) return;
        let disposed = false;
        let timer = 0;
        let polling = false;
        let previous: any;
        const reset = () => { previous = undefined; setRates(undefined); setHistory([]); };
        setStatus(undefined);
        setError("");
        setLoading(true);
        reset();

        const poll = async () => {
            if (disposed || polling) return;
            polling = true;
            let requestStartedAt = performance.now();
            try {
                // 同一组件切换目标或重新挂载 effect 时，先等旧请求结束，避免并发采样。
                if (requestRef.current) await requestRef.current.catch(() => undefined);
                if (disposed) return;
                requestStartedAt = performance.now();
                const pending = getContainerStatus({body: {id}});
                requestRef.current = pending;
                let response;
                try { response = await pending; }
                finally { if (requestRef.current === pending) requestRef.current = null; }
                if (disposed) return;
                if (response?.code !== 200 || !response?.data) throw new Error(response?.message || "获取资源状态失败");
                const next = response?.data;
                setStatus(next);
                callbackRef.current?.(next);
                const time = Date.parse(next?.read_time);
                if (!Number.isFinite(time) || time <= 0) throw new Error("采样时间无效，等待下一次采样");
                if (![next?.cpu_usage, next?.io_read_bytes, next?.io_write_bytes, next?.memory_usage].every(validNumber)) {
                    throw new Error("资源采样数据无效，等待下一次采样");
                }
                const elapsed = previous ? time - Date.parse(previous?.read_time) : 0;
                const continuous = !!previous && elapsed > 0
                    && next?.generation === previous?.generation && next?.started_at === previous?.started_at
                    && next?.cpu_usage >= previous?.cpu_usage && next?.io_read_bytes >= previous?.io_read_bytes
                    && next?.io_write_bytes >= previous?.io_write_bytes;
                const nextRates = continuous && previous ? {
                    cpu: (next?.cpu_usage - previous?.cpu_usage) / (elapsed * 1000000) * 100,
                    read: (next?.io_read_bytes - previous?.io_read_bytes) / elapsed * 1000,
                    write: (next?.io_write_bytes - previous?.io_write_bytes) / elapsed * 1000,
                } : undefined;
                // 重启、计数器回退及采样时间倒退时，从新基线重新绘图。
                setHistory((items) => [...(continuous ? items : []), {time, memory: next?.memory_usage, rates: nextRates}].slice(-MAX_SAMPLES));
                setRates(nextRates);
                previous = next;
                setError("");
            } catch (cause) {
                if (!disposed) {
                    reset();
                    setError(cause instanceof Error ? cause.message : "获取资源状态失败");
                }
            } finally {
                polling = false;
                if (!disposed) {
                    setLoading(false);
                    const delay = Math.max(0, SAMPLE_INTERVAL - (performance.now() - requestStartedAt));
                    timer = window.setTimeout(() => void poll(), delay);
                }
            }
        };
        void poll();
        return () => {
            disposed = true;
            window.clearTimeout(timer);
        };
    }, [id, active]);

    // 三张图统一等到第二次有效采样、能够计算速率后再显示。
    const plots = useMemo(() => !visible || !history[history.length - 1]?.rates ? {cpu: [], memory: [], io: []} : ({
        cpu: history.flatMap((item) => item.rates ? [{time: item.time, value: item.rates.cpu, series: "CPU"}] : []),
        memory: history.map((item) => ({time: item.time, value: item.memory, series: "内存"})),
        io: history.flatMap((item) => item.rates ? [
            {time: item.time, value: item.rates.read, series: "读取"},
            {time: item.time, value: item.rates.write, series: "写入"},
        ] : []),
    }), [history, visible]);
    const series = useMemo(() => ({
        cpu: [{label: "CPU", color: token.colorPrimary}],
        memory: [{label: "内存", color: token.colorPrimary}],
        io: [{label: "读取", color: token.colorPrimary}, {label: "写入", color: token.colorSuccess}],
    }), [token.colorPrimary, token.colorSuccess]);
    // 切换详情分段只卸载图表，采样与历史数据保留到弹窗关闭。
    if (!visible) return null;

    const available = active && history.length > 0 && !error;
    const displayBytes = (value?: number) => available && validNumber(value) ? formatSize(value) : "—";
    const currentRates = available ? rates : undefined;
    const emptyText = !active ? "" : error ? "等待采样恢复" : "等待有效采样";
    const lastSample = history[history.length - 1];
    const rateEmptyText = available && lastSample ? "等待下一次采样计算速率" : emptyText;
    const waiting = active && !error && (loading || !rates);
    const displayCount = (value?: number) => available && validNumber(value) ? value.toLocaleString() : "—";
    const displaySeconds = (value?: number) => available && validNumber(value) ? `${(value / 1e9).toFixed(2)} 秒` : "—";
    const metricGroups = [
        {title: "运行状态", items: [
            ["本次运行时长", status && validNumber(status.uptime) ? secondsText(status.uptime) : "—"],
            ["采样时间", lastSample ? dayjs(lastSample.time).format("YYYY-MM-DD HH:mm:ss") : "—"],
            ["进程与线程数", displayCount(status?.pids)],
            ["进程数上限", available && validNumber(status?.pids_limit) ? status.pids_limit > 0 ? displayCount(status.pids_limit) : "不限制" : "—"],
        ]},
        {title: "磁盘 I/O", items: [
            ["累计读取", displayBytes(status?.io_read_bytes)],
            ["累计写入", displayBytes(status?.io_write_bytes)],
            ["累计读取次数", displayCount(status?.io_read_operations)],
            ["累计写入次数", displayCount(status?.io_write_operations)],
        ]},
        {title: "内存", items: [
            ["内存上限", available && validNumber(status?.memory_limit) ? status.memory_limit > 0 ? displayBytes(status.memory_limit) : "不限制" : "—"],
            ["内存峰值", displayBytes(status?.memory_peak)],
            ["常驻内存", displayBytes(status?.memory_rss)],
            ["文件缓存", displayBytes(status?.memory_cache)],
            ["活跃内存", displayBytes(status?.memory_working_set)],
            ["交换空间", displayBytes(status?.swap_usage)],
        ]},
        {title: "CPU", items: [
            ["累计 CPU 时间", displaySeconds(status?.cpu_usage)],
            ["CPU 调度周期", displayCount(status?.cpu_periods)],
            ["用户态 CPU 时间", displaySeconds(status?.cpu_user_usage)],
            ["内核态 CPU 时间", displaySeconds(status?.cpu_system_usage)],
            ["CPU 限流次数", displayCount(status?.cpu_throttled_periods)],
            ["CPU 限流时间", displaySeconds(status?.cpu_throttled_time)],
        ]},
    ];

    return <div className={styles.resources}>
        {error && <Alert type="warning" showIcon title={error}/>}
        <div className="resource-grid">
            <ResourceChart title="CPU 使用率" data={plots.cpu} series={series.cpu} readings={[currentRates ? `${currentRates.cpu.toFixed(1)}%` : "—"]}
                           kind="percent" dark={dark} emptyText={rateEmptyText} loading={waiting} height={chartHeight} active={active}/>
            <ResourceChart title="内存用量" data={plots.memory} series={series.memory} readings={[currentRates ? displayBytes(status?.memory_usage) : "—"]}
                           kind="bytes" dark={dark} emptyText={rateEmptyText} loading={waiting} height={chartHeight} active={active}/>
            <ResourceChart title="磁盘 I/O" data={plots.io} series={series.io}
                           readings={[currentRates ? `${formatSize(currentRates.read)}/s` : "—", currentRates ? `${formatSize(currentRates.write)}/s` : "—"]}
                           kind="rate" dark={dark} wide emptyText={rateEmptyText} loading={waiting} height={chartHeight} active={active}/>
        </div>
        <div className="resource-details">{metricGroups.map((group) => <section className="resource-metric-group" key={group.title} aria-label={`${group.title}指标`}>
            <h3 className="resource-metric-title">{group.title}</h3>
            <dl className="resource-metrics">{group.items.map(([label, value]) => <div className="resource-metric" key={label}>
                <dt>{label}</dt><dd>{value}</dd>
            </div>)}</dl>
        </section>)}</div>
    </div>;
};

export default ContainerResources;
