import {useMemo} from "react";
import type {Chart} from "@antv/g2";

type Range = [number, number];

export default function useTimeRange(rangeKey: string) {
    return useMemo(() => {
        const range = {
            values: [0, 1] as Range,
            getValues: (): Range => range.values,
            setValues: (values: Range) => { range.values = values; },
            bind: (chart: Chart) => {
                const container = chart.getContainer();
                let pointer: PointerEvent | undefined;
                const finishTouch = () => {
                    if (!pointer) return;
                    const {target, pointerId, pointerType, clientX, clientY} = pointer;
                    pointer = undefined;
                    // AntV 滑块只在 pointerup 清理文档监听，滚动取消及卸载也需要结束拖动。
                    (target instanceof Element && target.isConnected ? target : document).dispatchEvent(
                        new PointerEvent("pointerup", {bubbles: true, pointerId, pointerType, clientX, clientY}),
                    );
                };
                const onPointer = (event: PointerEvent) => {
                    if (event.pointerType !== "touch") return;
                    if (event.type === "pointerdown") pointer ||= event;
                    else if (event.pointerId === pointer?.pointerId) {
                        if (event.type === "pointercancel") finishTouch();
                        else pointer = undefined;
                    }
                };
                const onTouchStart = (event: TouchEvent) => {
                    if (event.touches.length !== 1) {
                        finishTouch();
                        event.stopPropagation();
                        return;
                    }
                    const touch = event.changedTouches[0];
                    if (!touch || "clientX" in event) return;
                    // AntV 的滑块从 nativeEvent 读取鼠标坐标，但触屏的 nativeEvent 是 TouchEvent。
                    Object.defineProperties(event, {
                        clientX: {value: touch.clientX},
                        clientY: {value: touch.clientY},
                    });
                };
                container.addEventListener("touchstart", onTouchStart, {capture: true, passive: true});
                container.addEventListener("pointerdown", onPointer, true);
                const endEvents = ["pointerup", "pointercancel"] as const;
                endEvents.forEach((type) => window.addEventListener(type, onPointer, true));
                return () => {
                    finishTouch();
                    container.removeEventListener("touchstart", onTouchStart, true);
                    container.removeEventListener("pointerdown", onPointer, true);
                    endEvents.forEach((type) => window.removeEventListener(type, onPointer, true));
                };
            },
        };
        return range;
    }, [rangeKey]);
}
