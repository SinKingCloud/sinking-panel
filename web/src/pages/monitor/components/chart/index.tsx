import React, {useEffect, useMemo, useRef, useState} from "react";
import {Line} from "@ant-design/charts";
import {Grid, Spin, theme as antdTheme} from "antd";
import dayjs from "dayjs";
import {Title, useTheme} from "sinking-antd";
import {formatSize} from "@/utils/string";
import type {MonitorHistory, MonitorSeries, MonitorUnit} from "../../types";
import useTimeRange from "./use-time-range";
import "./range-interaction";
import useTooltip from "./use-tooltip";
import useStyles from "./styles";

type Props = {
    title: string;
    subtitle: string;
    history?: MonitorHistory;
    series: MonitorSeries[];
    iopsSeries?: MonitorSeries[];
    unit: Exclude<MonitorUnit, "iops">;
    rangeKey: string;
    loading: boolean;
};

const units = ["B/s", "KB/s", "MB/s", "GB/s", "TB/s"];
const numberLabel = new Intl.NumberFormat("zh-CN", {maximumFractionDigits: 2}).format;
type ChartDatum = {time: number; value: number | null; series: string};
const metricValue = (value: number | null | undefined, unit: MonitorUnit) => {
    if (typeof value !== "number" || !Number.isFinite(value)) return {value: "--", suffix: ""};
    if (unit === "bytes") {
        const [size, suffix] = formatSize(value).split(" ");
        return {value: size, suffix: `${suffix}/s`};
    }
    if (unit === "iops") return {value: numberLabel(Math.round(value)), suffix: "次/s"};
    return {value: value.toFixed(unit === "percent" ? 1 : 2), suffix: unit === "percent" ? "%" : ""};
};

const MonitorChart = React.memo(({title, subtitle, history, series, iopsSeries, unit, rangeKey, loading}: Props) => {
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const {token} = antdTheme.useToken();
    const screens = Grid.useBreakpoint();
    const mobile = screens.md === false;
    const chartHeight = screens.sm === false ? 228 : 248;
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const containerRef = useRef<HTMLDivElement | null>(null);
    const chartRef = useRef<React.ComponentRef<typeof Line>>(null);
    const [inView, setInView] = useState(false);
    const [view, setView] = useState<{history?: MonitorHistory; rangeKey: string}>();
    const summarySeries = useMemo(() => [
        ...series.map((item) => ({...item, unit})),
        ...(iopsSeries || []).map((item) => ({...item, unit: "iops" as const})),
    ], [series, iopsSeries, unit]);
    const timeRange = useTimeRange(view?.rangeKey || "");
    const points = view?.history?.data;
    const start = view?.history?.start;
    const end = view?.history?.end;
    const pointCount = points?.length || 0;
    const latest = history?.data[history.data.length - 1];
    const processMetric = series[0]?.key === "cpu" ? "cpu" : series[0]?.key === "memory" ? "memory" : series[0]?.key === "disk_read" ? "disk_io" : undefined;
    const tooltip = useTooltip(processMetric, view?.history, summarySeries, unit);

    useEffect(() => {
        if (!inView || loading) tooltip.hide();
    }, [inView, loading, tooltip]);

    // 离屏保留实例和选区，暂停接收绘图数据；重新进入视口时一次更新到最新结果。
    useEffect(() => {
        if (inView) setView({history, rangeKey});
    }, [inView, history, rangeKey]);

    useEffect(() => {
        if (typeof IntersectionObserver === "undefined") {
            setInView(true);
            return;
        }
        const observer = new IntersectionObserver(([entry]) => {
            setInView(entry.isIntersecting);
        });
        if (containerRef.current) observer.observe(containerRef.current);
        return () => observer.disconnect();
    }, []);

    useEffect(() => {
        if (!inView || pointCount === 0 || !containerRef.current || typeof ResizeObserver === "undefined") return;
        let width = 0;
        let timer = 0;
        const observer = new ResizeObserver(([entry]) => {
            const nextWidth = Math.round(entry.contentRect.width);
            if (nextWidth <= 0 || nextWidth === width) return;
            width = nextWidth;
            window.clearTimeout(timer);
            timer = window.setTimeout(() => chartRef.current?.chart?.forceFit?.(), 80);
        });
        observer.observe(containerRef.current);
        return () => {
            observer.disconnect();
            window.clearTimeout(timer);
        };
    }, [inView, pointCount > 0]);

    const chartData = useMemo(() => {
        const data: ChartDatum[] = [];
        const areaData: ChartDatum[] = [];
        let maximum = 0;
        // 直接展示后端按完整查询范围补齐的时间桶，不在页面补点。
        points?.forEach((point) => {
            series.forEach((item) => {
                const value = point[item.key];
                const validValue = typeof value === "number" && Number.isFinite(value) ? value : null;
                if (validValue !== null) maximum = Math.max(maximum, validValue);
                const datum = {time: point.time, value: validValue, series: item.label};
                data.push(datum);
                if (item === series[0]) areaData.push(datum);
            });
        });
        // 缩略图只保留最多 120 个桶的峰值，主曲线和提示仍使用后端聚合后的全部点。
        const overview: number[] = [];
        const step = Math.max(1, Math.ceil(areaData.length / 120));
        for (let index = 0; index < areaData.length; index += step) {
            let peak = 0;
            for (let offset = index; offset < Math.min(index + step, areaData.length); offset++) {
                peak = Math.max(peak, areaData[offset].value ?? 0);
            }
            overview.push(peak);
        }
        return {data, areaData, maximum, overview};
    }, [points, series]);

    const chart = useMemo(() => {
        const {data, areaData, maximum, overview} = chartData;
        let unitIndex = 0;
        if (unit === "bytes") {
            while (unitIndex < units.length - 1 && maximum >= 1024 ** (unitIndex + 1)) unitIndex += 1;
        }
        const displayUnit = unit === "percent" ? "%" : unit === "load" ? "负载" : units[unitIndex];
        const labelTime = (value: any) => {
            const date = dayjs(value instanceof Date ? value : Number(value));
            return date.format((end || 0) - (start || 0) > 86400000 ? "MM-DD HH:mm" : "HH:mm");
        };
        const config: any = {
            data,
            xField: "time",
            yField: "value",
            colorField: "series",
            seriesField: "series",
            shape: "smooth",
            connectNulls: false,
            clip: true,
            margin: 0,
            marginBottom: 4,
            paddingRight: 8,
            paddingTop: 6,
            insetBottom: 3,
            scale: {
                // 单个采样可能恰好等于结束时间，只扩展坐标轴，绝不伪造数据点。
                x: {type: "time", domain: start !== undefined && end !== undefined ? [start, Math.max(end, start + 1)] : undefined, nice: false},
                y: {domainMin: 0, domainMax: unit === "percent" ? 100 : maximum > 0 ? undefined : 1, nice: unit !== "percent"},
                color: {domain: series.map((item) => item.label), range: series.map((item) => item.color)},
            },
            axis: {
                x: {
                    title: false, line: false, tick: false, tickCount: 4,
                    tickMethod: (min: number, max: number) => Array.from({length: 4}, (_, index) => min + (max - min) * index / 3),
                    labelSpacing: 2,
                    labelFill: token.colorTextSecondary, labelFontSize: 11, labelOpacity: 1,
                    labelTextAlign: ({value}: {value: number}) => value < .15 ? "left" : value > .85 ? "right" : "center",
                    labelAutoHide: true, labelAutoRotate: false, labelFormatter: labelTime, grid: false,
                },
                y: {
                    title: false, line: false, tick: false, tickCount: 4,
                    labelSpacing: 2,
                    labelFill: token.colorTextSecondary, labelFontSize: 11, labelOpacity: 1,
                    labelFormatter: (value: any) => numberLabel(Number(value) / (unit === "bytes" ? 1024 ** unitIndex : 1)),
                    grid: true, gridLineDash: [3, 4], gridLineWidth: 1,
                    gridStroke: token.colorBorderSecondary, gridStrokeOpacity: .8,
                },
            },
            legend: false,
            tooltip: {
                title: false,
                items: [(datum: any) => ({
                    name: datum?.series,
                    color: series.find((item) => item.label === datum?.series)?.color,
                    value: datum?.value,
                    time: datum?.time,
                })],
            },
            slider: pointCount > 1 && start !== undefined && end !== undefined && end > start ? {
                x: {
                    values: [...timeRange.values],
                    size: 28,
                    trackSize: 22,
                    crossPadding: 6,
                    showLabel: false,
                    showLabelOnInteraction: true,
                    labelFormatter: labelTime,
                    sparklineData: overview,
                    sparklineAreaFill: token.colorPrimary,
                    sparklineAreaFillOpacity: .12,
                    sparklineLineStroke: token.colorPrimary,
                    sparklineLineStrokeOpacity: .35,
                    trackFill: token.colorFillTertiary,
                    trackFillOpacity: 1,
                    selectionFill: token.colorPrimary,
                    selectionFillOpacity: .12,
                    handleIconSize: 12,
                    handleIconFill: token.colorBgContainer,
                    handleIconStroke: token.colorPrimary,
                    handleIconStrokeOpacity: .6,
                    handleLabelFill: token.colorTextSecondary,
                    handleLabelFillOpacity: 1,
                    handleLabelFontSize: 11,
                },
            } : false,
            interaction: {
                tooltip: {
                    shared: true, series: true, crosshairs: true,
                    body: false, disableAutoHide: mobile, wait: 0,
                },
                sliderFilter: false,
                monitorTimeRange: {getValues: timeRange.getValues, onChange: timeRange.setValues, start, end, labelFormatter: labelTime},
            },
            point: pointCount <= 2 ? {size: 3, shape: "point"} : false,
            // 主系列的淡色面积与折线使用相同的后端数据。
            area: {
                data: areaData,
                shape: "smooth",
                connectNulls: false,
                style: {fillOpacity: dark ? .1 : .055, lineWidth: 0},
                tooltip: false,
                zIndex: -1,
            },
            style: {lineWidth: 1.8, lineCap: "round", lineJoin: "round"},
            theme: {type: dark ? "classicDark" : "classic"},
            autoFit: true,
            animate: false,
        };
        return {config, displayUnit};
    }, [chartData, start, end, pointCount, series, unit, dark, mobile, timeRange, token.colorTextSecondary, token.colorBorderSecondary, token.colorPrimary, token.colorFillTertiary, token.colorBgContainer]);

    useEffect(() => {
        if (!chartRef.current?.chart || !chart.config.slider || start === undefined || end === undefined) return;
        return timeRange.bind(chartRef.current.chart);
    }, [timeRange, chart.config, start, end]);

    useEffect(() => {
        if (!chartRef.current?.chart || !pointCount) return;
        return tooltip.bind(chartRef.current.chart);
    }, [tooltip, pointCount > 0]);

    // 加载时保留图表实例和高度，收到新数据后再更新。
    const plot = useMemo(() => (
        <Line ref={chartRef} {...chart.config} height={chartHeight}
              containerStyle={{height: chartHeight, width: "100%"}}/>
    ), [chart.config, chartHeight]);

    return (
        <section className={styles.panel} aria-label={title} aria-busy={loading}>
            <div className={styles.heading}>
                <Title size="small"><h3>{title}</h3></Title>
                {latest && <span className="sample-time" title={`所选范围末个时段：${dayjs(latest.time).format("YYYY-MM-DD HH:mm")}`}>
                    统计时段 <time dateTime={new Date(latest.time).toISOString()}>{dayjs(latest.time).format("MM-DD HH:mm")}</time>
                </span>}
            </div>
            <div className={styles.summary}>
                <div className={`summary-metrics${iopsSeries?.length ? " io-metrics" : ""}`} style={{gridTemplateColumns: `repeat(${series.length}, minmax(0, 1fr))`}}>
                    {series.map((item, index) => {
                        const value = metricValue(latest?.[item.key], unit);
                        const iops = iopsSeries?.[index];
                        const count = iops ? metricValue(latest?.[iops.key], "iops") : undefined;
                        return <div className="summary-metric" key={item.key} style={{"--series-color": item.color} as React.CSSProperties}>
                            <span className="metric-label"><i/>{series.length === 1 ? "使用率" : item.label}</span>
                            <span className="metric-value" title={`${value.value}${value.suffix ? ` ${value.suffix}` : ""}`}><strong>{value.value}</strong>{value.suffix && <small>{value.suffix}</small>}</span>
                            {count && <span className="io-count" title={`${count.value} IOPS`}><strong>{count.value}</strong><small>IOPS</small></span>}
                        </div>;
                    })}
                </div>
                {subtitle && <span className="summary-detail">{subtitle}</span>}
            </div>
            <div ref={containerRef} className={styles.plot}>
                <Spin spinning={loading}>
                    <div className="chart-unit">{chart.displayUnit}</div>
                    {points?.length ? plot : <div style={{height: chartHeight}}/>}
                </Spin>
            </div>
        </section>
    );
});

MonitorChart.displayName = "MonitorChart";

export default MonitorChart;
