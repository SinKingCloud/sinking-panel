import {Line} from "@ant-design/charts";
import {Grid, theme as antdTheme} from "antd";
import React, {useEffect, useMemo, useRef, useState} from "react";
import useChartTooltip from "@/pages/components/ui/chart-tooltip";
import type {ChartTooltipModel} from "@/pages/components/ui/chart-tooltip";

const byteUnits = ["B/s", "KB/s", "MB/s", "GB/s", "TB/s", "PB/s"];
const rateFormatters = [0, 1, 2].map((maximumFractionDigits) => new Intl.NumberFormat("zh-CN", {maximumFractionDigits}));

const numberValue = (value: any): number => {
    const number = Number(value);
    return Number.isFinite(number) ? Math.max(0, number) : 0;
};

const formatByteRate = (value: any): string => {
    let rate = numberValue(value);
    let unitIndex = 0;
    while (rate >= 1024 && unitIndex < byteUnits.length - 1) {
        rate /= 1024;
        unitIndex += 1;
    }

    const maximumFractionDigits = rate >= 100 || unitIndex === 0 ? 0 : rate >= 10 ? 1 : 2;
    return `${rateFormatters[maximumFractionDigits].format(rate)} ${byteUnits[unitIndex]}`;
};

const formatTime = (value: any): string => {
    const numeric = Number(value);
    const date = value instanceof Date
        ? value
        : new Date(Number.isFinite(numeric) ? numeric : value);
    if (Number.isNaN(date.getTime())) {
        return "--:--:--";
    }

    const pad = (part: number) => String(part).padStart(2, "0");
    return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
};

const TelemetryChart = React.memo(({className, data, first, second, isDarkMode}: any) => {
    const {token} = antdTheme.useToken();
    const containerRef = useRef<HTMLDivElement | null>(null);
    const chartRef = useRef<any>(null);
    const appliedRef = useRef<{instance: any; config: any; data: any[]} | undefined>(undefined);
    const sizeRef = useRef({width: 0, height: 0});
    const [inView, setInView] = useState(false);
    const [pageVisible, setPageVisible] = useState(() => !document.hidden);
    const [view, setView] = useState<any[]>();
    const mobile = Grid.useBreakpoint().md === false;
    const active = inView && pageVisible;
    const firstLabel = first?.label || "系列一";
    const secondLabel = second?.label || "系列二";
    const firstColor = first?.color || "#1677ff";
    const secondColor = second?.color || "#13c2c2";
    const series = useMemo(() => [{label: firstLabel, color: firstColor}, {label: secondLabel, color: secondColor}], [firstLabel, secondLabel, firstColor, secondColor]);

    useEffect(() => {
        const updateVisibility = () => setPageVisible(!document.hidden);
        document.addEventListener("visibilitychange", updateVisibility);
        const observer = typeof IntersectionObserver === "undefined" ? undefined : new IntersectionObserver(([entry]) => setInView(entry.isIntersecting));
        if (observer && containerRef.current) observer.observe(containerRef.current);
        else setInView(true);
        return () => {
            observer?.disconnect();
            document.removeEventListener("visibilitychange", updateVisibility);
        };
    }, []);

    useEffect(() => {
        if (active) setView(data);
    }, [active, data]);

    useEffect(() => {
        const container = containerRef.current;
        if (!active || !view || !container || typeof ResizeObserver === "undefined") {
            return;
        }
        let timer = 0;
        const observer = new ResizeObserver(([entry]) => {
            const nextWidth = Math.round(entry.contentRect.width);
            const nextHeight = Math.round(entry.contentRect.height);
            const previous = sizeRef.current;
            if (nextWidth <= 0 || (nextWidth === previous.width && nextHeight === previous.height)) {
                return;
            }
            sizeRef.current = {width: nextWidth, height: nextHeight};
            // 首次绘制已按容器布局，重新入屏时才补上离屏期间发生的尺寸变化。
            if (!previous.width) return;
            window.clearTimeout(timer);
            timer = window.setTimeout(() => {
                const chart = chartRef.current?.chart;
                if (chart?.forceFit) {
                    void chart.forceFit();
                }
            }, 80);
        });
        observer.observe(container);
        return () => {
            window.clearTimeout(timer);
            observer.disconnect();
        };
    }, [active, Boolean(view)]);

    const sampleCache = useMemo(() => new WeakMap<object, {data: any[]; model: ChartTooltipModel; maximum: number}>(), [firstLabel, secondLabel]);
    const prepared = useMemo(() => {
        const chartData: any[] = [];
        const models = new Map<number, ChartTooltipModel>();
        let maxValue = 0;
        // 上游窗口已按时间排序；保留点复用折线对象与已格式化的悬浮内容。
        view?.forEach((item) => {
            let sample = sampleCache.get(item);
            if (!sample) {
                const time = numberValue(item.time);
                const first = numberValue(item.first);
                const second = numberValue(item.second);
                sample = {
                    data: [{time, value: first, series: firstLabel}, {time, value: second, series: secondLabel}],
                    model: {title: formatTime(time), values: [formatByteRate(first), formatByteRate(second)]},
                    maximum: Math.max(first, second),
                };
                sampleCache.set(item, sample);
            }
            chartData.push(...sample.data);
            models.set(sample.data[0].time, sample.model);
            maxValue = Math.max(maxValue, sample.maximum);
        });
        return {chartData, models, maxValue};
    }, [view, sampleCache, firstLabel, secondLabel]);
    const tooltip = useChartTooltip(prepared.models, series);
    useEffect(() => {
        if (!active) tooltip.hide();
    }, [active, tooltip]);
    useEffect(() => {
        if (!view || !chartRef.current?.chart) return;
        return tooltip.bind(chartRef.current.chart);
    }, [tooltip, Boolean(view)]);

    let axisUnitIndex = 0;
    while (axisUnitIndex < byteUnits.length - 1 && prepared.maxValue >= 1024 ** (axisUnitIndex + 1)) axisUnitIndex += 1;
    const hasValues = prepared.maxValue > 0;
    const chart = useMemo<any>(() => {
        const gridColor = isDarkMode ? "rgba(255,255,255,0.08)" : "rgba(5,5,5,0.06)";
        const axisDivisor = 1024 ** axisUnitIndex;
        const formatAxisRate = (value: any): string => {
            const rate = numberValue(value) / axisDivisor;
            const maximumFractionDigits = rate >= 100 || axisUnitIndex === 0 ? 0 : rate >= 10 ? 1 : 2;
            return rateFormatters[maximumFractionDigits].format(rate);
        };

        return {
            unit: byteUnits[axisUnitIndex],
            config: {
                // 配置仅在坐标或样式变化时携带当时数据；普通采样通过 changeData 提交。
                data: prepared.chartData,
                xField: "time",
                yField: "value",
                colorField: "series",
                seriesField: "series",
                keyField: "series",
                shape: "smooth",
                paddingTop: 8,
                insetBottom: 3,
                scale: {
                    x: {
                        type: "time",
                        nice: false,
                    },
                    color: {
                        domain: [firstLabel, secondLabel],
                        range: [firstColor, secondColor],
                    },
                    y: {
                        domainMin: 0,
                        domainMax: hasValues ? undefined : 1,
                        nice: true,
                    },
                },
                axis: {
                    x: {
                        title: false,
                        line: false,
                        tick: false,
                        tickCount: 5,
                        labelFill: token.colorTextSecondary,
                        labelFontSize: 10,
                        labelAutoHide: true,
                        labelAutoRotate: false,
                        labelFormatter: formatTime,
                        grid: false,
                    },
                    y: {
                        title: false,
                        line: false,
                        tick: false,
                        tickCount: 4,
                        labelFill: token.colorTextSecondary,
                        labelFontSize: 10,
                        labelFormatter: formatAxisRate,
                        grid: true,
                        gridLineDash: [4, 4],
                        gridLineWidth: 1,
                        gridStroke: gridColor,
                        gridStrokeOpacity: 1,
                    },
                },
                legend: false,
                tooltip: {
                    title: false,
                    items: [
                        (datum: any) => ({
                            name: datum?.series || "--",
                            color: datum?.series === firstLabel ? firstColor : secondColor,
                            value: datum?.value,
                            time: datum?.time,
                        }),
                    ],
                },
                interaction: {
                    tooltip: {
                        shared: true,
                        series: true,
                        crosshairs: true,
                        body: false,
                        wait: 0,
                        disableAutoHide: mobile,
                    },
                },
                theme: {type: isDarkMode ? "classicDark" : "classic"},
                autoFit: true,
                style: {
                    lineWidth: 2,
                    lineCap: "round",
                    lineJoin: "round",
                },
                animate: false,
            },
        };
    }, [axisUnitIndex, hasValues, Boolean(view), firstColor, firstLabel, isDarkMode, secondColor, secondLabel, token.colorTextSecondary, mobile]);

    useEffect(() => {
        const instance = chartRef.current?.chart;
        if (!view || !instance) return;
        const previous = appliedRef.current;
        appliedRef.current = {instance, config: chart.config, data: prepared.chartData};
        // Line 已负责首次/配置变更绘制，不再追加一次相同数据的 render。
        if (!previous || previous.instance !== instance || previous.config !== chart.config || previous.data === prepared.chartData) return;
        // G2 同步更新自身及子 mark 的 data，后续 forceFit 仍使用最新窗口。
        void instance.changeData(prepared.chartData);
    }, [chart.config, prepared.chartData, Boolean(view)]);
    const plot = useMemo(() => <Line ref={chartRef} {...chart.config} className="telemetry-plot"/>, [chart.config]);

    return (
        <div ref={containerRef} className={className}>
            <div className="telemetry-toolbar">
                <div className="telemetry-legends">
                    <span className="telemetry-legend" style={{"--legend-color": firstColor} as React.CSSProperties}>
                        <i/>{firstLabel}
                    </span>
                    <span className="telemetry-legend" style={{"--legend-color": secondColor} as React.CSSProperties}>
                        <i/>{secondLabel}
                    </span>
                </div>
                <span className="telemetry-unit">{chart.unit}</span>
            </div>
            {view ? plot : <div className="telemetry-plot"/>}
        </div>
    );
});

TelemetryChart.displayName = "TelemetryChart";

export default TelemetryChart;
