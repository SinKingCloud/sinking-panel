import {useMemo} from "react";
import type {Chart} from "@antv/g2";

type Range = [number, number];

export default function useTimeRange(rangeKey: string) {
    return useMemo(() => {
        const range = {
            values: [0, 1] as Range,
            getInitValues: (slider: {setValues: (values: Range, animate: boolean) => void}) => {
                const values = [...range.values] as Range;
                // 刷新和尺寸变化在交互初始化时应用当前选区，避免绘制完成后再跳回。
                slider.setValues(values, false);
                return values[0] === 0 && values[1] === 1 ? undefined : values;
            },
            bind: (chart: Chart, start: number, end: number) => {
                const onFilter = (event: {nativeEvent?: boolean; data?: {selection?: Range[]}}) => {
                    if (!event.nativeEvent || !event.data?.selection?.[0]) return;
                    const values = event.data.selection[0].map((value) => Math.max(0, Math.min(1, (Number(value) - start) / (end - start)))) as Range;
                    if (!values.every(Number.isFinite) || values[0] >= values[1]) return;
                    range.values = values;
                };
                chart.on("sliderX:filter", onFilter);
                return () => { chart.off("sliderX:filter", onFilter); };
            },
        };
        return range;
    }, [rangeKey]);
}
