import {Line} from "@ant-design/charts";
import {theme as antdTheme} from "antd";
import React, {useMemo} from "react";

const byteUnits = ["B/s", "KB/s", "MB/s", "GB/s", "TB/s", "PB/s"];

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
    return `${rate.toLocaleString("zh-CN", {maximumFractionDigits})} ${byteUnits[unitIndex]}`;
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
    const firstLabel = first?.label || "系列一";
    const secondLabel = second?.label || "系列二";
    const firstColor = first?.color || "#1677ff";
    const secondColor = second?.color || "#13c2c2";
    const chart = useMemo<any>(() => {
        const gridColor = isDarkMode ? "rgba(255,255,255,0.08)" : "rgba(5,5,5,0.06)";
        const source = Array.isArray(data)
            ? [...data].sort((left: any, right: any) => numberValue(left?.time) - numberValue(right?.time))
            : [];
        const chartData = source.flatMap((item: any) => {
            const time = numberValue(item?.time);
            return [
                {time, value: numberValue(item?.first), series: firstLabel},
                {time, value: numberValue(item?.second), series: secondLabel},
            ];
        });
        const maxValue = chartData.reduce((max: number, item: any) => Math.max(max, item.value), 0);
        let axisUnitIndex = 0;
        while (axisUnitIndex < byteUnits.length - 1 && maxValue >= 1024 ** (axisUnitIndex + 1)) {
            axisUnitIndex += 1;
        }
        const axisDivisor = 1024 ** axisUnitIndex;
        const formatAxisRate = (value: any): string => {
            const rate = numberValue(value) / axisDivisor;
            const maximumFractionDigits = rate >= 100 || axisUnitIndex === 0 ? 0 : rate >= 10 ? 1 : 2;
            return rate.toLocaleString("zh-CN", {maximumFractionDigits});
        };

        return {
            unit: byteUnits[axisUnitIndex],
            config: {
                data: chartData,
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
                        domainMax: maxValue > 0 ? undefined : 1,
                        nice: true,
                    },
                },
                axis: {
                    x: {
                        title: false,
                        line: false,
                        tick: false,
                        tickCount: 5,
                        labelFill: token.colorTextQuaternary,
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
                        labelFill: token.colorTextQuaternary,
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
                    title: (datum: any) => formatTime(datum?.time),
                    items: [
                        (datum: any) => ({
                            name: datum?.series || "--",
                            color: datum?.series === firstLabel ? firstColor : secondColor,
                            value: formatByteRate(datum?.value),
                        }),
                    ],
                },
                interaction: {
                    tooltip: {
                        shared: true,
                        series: true,
                        crosshairs: true,
                    },
                },
                theme: {type: isDarkMode ? "classicDark" : "classic"},
                autoFit: true,
                style: {
                    lineWidth: 2,
                    lineCap: "round",
                    lineJoin: "round",
                },
                animate: {
                    enter: {type: "fadeIn", duration: 280},
                    update: {type: null},
                },
            },
        };
    }, [data, firstColor, firstLabel, isDarkMode, secondColor, secondLabel, token.colorTextQuaternary]);

    return (
        <div className={className}>
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
            <Line {...chart.config} className="telemetry-plot"/>
        </div>
    );
});

TelemetryChart.displayName = "TelemetryChart";

export default TelemetryChart;
