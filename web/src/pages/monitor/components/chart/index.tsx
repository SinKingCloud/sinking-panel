import React, {useEffect, useMemo, useRef, useState} from "react";
import {Line} from "@ant-design/charts";
import {Empty, Grid, Spin, theme as antdTheme} from "antd";
import dayjs from "dayjs";
import {Title, useTheme} from "sinking-antd";
import {formatSize} from "@/utils/string";
import type {MonitorHistory, MonitorSeries} from "../../types";
import useTimeRange from "./use-time-range";
import useStyles from "./styles";

type Props = {
    title: string;
    subtitle: string;
    history?: MonitorHistory;
    series: MonitorSeries[];
    unit: "percent" | "bytes" | "load";
    loading: boolean;
    rangeKey: string;
};

const units = ["B/s", "KB/s", "MB/s", "GB/s", "TB/s"];
const numberLabel = new Intl.NumberFormat("zh-CN", {maximumFractionDigits: 2}).format;
type ChartDatum = {time: number; value: number | null; series: string; samples: number};
const metricValue = (value: number | undefined, unit: Props["unit"]) => {
    if (typeof value !== "number" || !Number.isFinite(value)) return {value: "--", suffix: ""};
    if (unit === "bytes") {
        const [size, suffix] = formatSize(value).split(" ");
        return {value: size, suffix: `${suffix}/s`};
    }
    return {value: value.toFixed(unit === "percent" ? 1 : 2), suffix: unit === "percent" ? "%" : ""};
};

const MonitorChart = React.memo(({title, subtitle, history, series, unit, loading, rangeKey}: Props) => {
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const {token} = antdTheme.useToken();
    const screens = Grid.useBreakpoint();
    const chartHeight = screens.sm === false ? 228 : 248;
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const containerRef = useRef<HTMLDivElement | null>(null);
    const chartRef = useRef<React.ComponentRef<typeof Line>>(null);
    const [inView, setInView] = useState(false);
    const [view, setView] = useState<{history?: MonitorHistory; rangeKey: string}>();
    const timeRange = useTimeRange(view?.rangeKey || "");
    const points = view?.history?.points;
    const start = view?.history?.start;
    const end = view?.history?.end;
    const pointCount = points?.length || 0;
    const latest = history?.points[history.points.length - 1];

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
        // 直接展示后端从首个真实采样开始的时间桶，不在页面补点。
        points?.forEach((point) => {
            series.forEach((item) => {
                const value = point[item.key];
                const valid = Number.isFinite(value);
                if (valid) maximum = Math.max(maximum, value);
                const datum = {time: point.time, value: valid ? value : null, series: item.label, samples: point.samples};
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
        const formatValue = (value: number | null | undefined) => value == null || !Number.isFinite(value)
            ? "无数据"
            : unit === "bytes" ? `${formatSize(value)}/s` : `${numberLabel(value)}${unit === "percent" ? "%" : ""}`;
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
                title: (datum: any) => `${dayjs(datum?.time).format("YYYY-MM-DD HH:mm")}${datum?.samples === 0 ? "（无采样）" : ""}`,
                items: [(datum: any) => ({
                    name: datum?.series,
                    color: series.find((item) => item.label === datum?.series)?.color,
                    value: formatValue(datum?.value),
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
                tooltip: {shared: true, series: true, crosshairs: true},
                sliderFilter: {adaptiveMode: false, wait: 32, leading: true, trailing: true},
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
    }, [chartData, start, end, pointCount, series, unit, dark, timeRange, token.colorTextSecondary, token.colorBorderSecondary, token.colorPrimary, token.colorFillTertiary, token.colorBgContainer]);

    useEffect(() => {
        if (!chartRef.current?.chart || !chart.config.slider || start === undefined || end === undefined) return;
        return timeRange.bind(chartRef.current.chart, start, end, chart.config.slider.x.values);
    }, [timeRange, chart.config, start, end]);

    // 加载状态只更新遮罩，避免图表组件为相同配置重复复制数据并重新绘制。
    const plot = useMemo(() => (
        <Line ref={chartRef} {...chart.config} height={chartHeight}
              containerStyle={{height: chartHeight, width: "100%"}}/>
    ), [chart.config, chartHeight]);

    return (
        <section className={styles.panel} aria-label={title}>
            <div className={styles.heading}>
                <Title size="small"><h3>{title}</h3></Title>
                <span className="sample-time" title={latest ? `${latest.samples > 0 ? "所选范围末个时段" : "当前时段无采样"}：${dayjs(latest.time).format("YYYY-MM-DD HH:mm")}` : undefined}>
                    {latest ? <>{latest.samples > 0 ? "统计时段" : "无采样"} <time dateTime={new Date(latest.time).toISOString()}>{dayjs(latest.time).format("MM-DD HH:mm")}</time></> : "暂无采样"}
                </span>
            </div>
            <div className={styles.summary}>
                <div className="summary-metrics" style={{gridTemplateColumns: `repeat(${series.length}, minmax(0, 1fr))`}}>
                    {series.map((item) => {
                        const value = metricValue(latest?.[item.key], unit);
                        return <div className="summary-metric" key={item.key} style={{"--series-color": item.color} as React.CSSProperties}>
                            <span className="metric-label"><i/>{series.length === 1 ? "使用率" : item.label}</span>
                            <span className="metric-value"><strong>{value.value}</strong>{value.suffix && <small>{value.suffix}</small>}</span>
                        </div>;
                    })}
                </div>
                {subtitle && <span className="summary-detail">{subtitle}</span>}
            </div>
            <div ref={containerRef} className={styles.plot}>
                <div className="chart-unit">{chart.displayUnit}</div>
                <Spin spinning={loading}>
                    {!view ? <div style={{height: chartHeight}}/> : points?.length
                        ? plot
                        : <div className="chart-empty" style={{height: chartHeight}}><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={loading ? "正在加载监控数据" : "所选时间暂无监控数据"}/></div>}
                </Spin>
            </div>
        </section>
    );
});

MonitorChart.displayName = "MonitorChart";

export default MonitorChart;
