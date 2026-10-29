import {useMemo} from "react";
import type {Chart} from "@antv/g2";

type Range = [number, number];

export default function useTimeRange(rangeKey: string) {
    return useMemo(() => {
        const range = {
            values: [0, 1] as Range,
            bind: (chart: Chart, start: number, end: number, configuredValues: Range) => {
                const onFilter = (event: {nativeEvent?: boolean; data?: {selection?: Range[]}}) => {
                    if (!event.nativeEvent || !event.data?.selection?.[0]) return;
                    const values = event.data.selection[0].map((value) => Math.max(0, Math.min(1, (Number(value) - start) / (end - start)))) as Range;
                    if (!values.every(Number.isFinite) || values[0] >= values[1]) return;
                    range.values = values;
                };
                const restore = () => {
                    // 新数据的配置已携带当前选区，只有复用旧配置调整尺寸时才需要恢复。
                    if (range.values.every((value, index) => Math.abs(value - configuredValues[index]) < .000001)) return;
                    chart.emit("sliderX:filter", {
                        nativeEvent: false,
                        data: {selection: [range.values.map((value) => start + (end - start) * value)]},
                    });
                };
                chart.on("sliderX:filter", onFilter);
                chart.on("afterrender", restore);
                return () => {
                    chart.off("sliderX:filter", onFilter);
                    chart.off("afterrender", restore);
                };
            },
        };
        return range;
    }, [rangeKey]);
}
