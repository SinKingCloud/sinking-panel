import {useLayoutEffect, useMemo, useRef} from "react";
import type {Chart} from "@antv/g2";
import useStyles from "./styles";

type TooltipEvent = {nativeEvent?: boolean; client?: {x: number; y: number}; data?: {items?: {time?: number}[]}};
export type ChartTooltipModel = {
    title: string;
    values: string[];
};

export type ChartTooltipOptions<T extends ChartTooltipModel> = {
    width?: number;
    className?: string;
    // 绑定时创建内容节点，数据点变化时更新；避免悬浮移动触发 React 渲染。
    renderContent?: (container: HTMLElement) => (model: T) => void;
};

function element<K extends keyof HTMLElementTagNameMap>(tag: K, className = "", text?: string) {
    const node = document.createElement(tag);
    node.className = className;
    if (text !== undefined) node.textContent = text;
    return node;
}

function text(node: HTMLElement, value: string) {
    if (node.textContent !== value) node.textContent = value;
}

export default function useChartTooltip<T extends ChartTooltipModel>(models: ReadonlyMap<number, T>, series: {label: string; color: string}[], {width = 220, className = "", renderContent}: ChartTooltipOptions<T> = {}) {
    const {styles} = useStyles();
    const source = useRef(models);
    useLayoutEffect(() => { source.current = models; }, [models]);
    // 数据更新只替换查找表，复用悬浮节点和事件绑定，避免实时图表每秒拆装面板。
    return useMemo(() => {
        let node: HTMLDivElement | undefined;
        let active: T | undefined;
        let pending: T | undefined;
        let frame = 0;
        const hide = () => {
            window.cancelAnimationFrame(frame);
            frame = 0;
            pending = active = undefined;
            if (node) node.hidden = true;
        };
        return {
            hide,
            bind: (chart: Chart, {zIndex}: {zIndex?: number} = {}) => {
                const container = chart.getContainer();
                const root = node = element("div", `${styles.tooltip} ${className}`.trim());
                root.style.width = `${width}px`;
                if (zIndex !== undefined) root.style.zIndex = String(zIndex);
                root.hidden = true;
                const title = element("div", "tooltip-title");
                root.append(title);
                const values = series.map((item) => {
                    const row = element("div", "tooltip-metric");
                    const dot = element("i");
                    dot.style.backgroundColor = item.color;
                    const value = element("strong");
                    row.append(dot, element("span", "", item.label), value);
                    root.append(row);
                    return value;
                });
                let updateContent: ((model: T) => void) | undefined;
                if (renderContent) {
                    const content = element("div");
                    root.append(content);
                    updateContent = renderContent(content);
                }
                document.body.append(root);
                let cursorX = 0;
                let cursorY = 0;
                let maxX = 0;
                let maxY = 0;
                let measuredWidth = 0;
                let flipX = 0;
                let offsetX = 16;
                let offsetY = 16;
                let lastX = -1;
                let lastY = -1;

                const show = (event: TooltipEvent) => {
                    if (!event.nativeEvent || !event.client) return;
                    const model = source.current.get(Number(event.data?.items?.[0]?.time));
                    if (!model) return hide();
                    pending = model;
                    // G2 复用事件坐标对象，立即取数；同帧只处理最后一次移动。
                    cursorX = event.client.x;
                    cursorY = event.client.y;
                    if (!Number.isFinite(cursorX) || !Number.isFinite(cursorY)) return hide();
                    if (frame) return;
                    frame = window.requestAnimationFrame(() => {
                        frame = 0;
                        const model = pending;
                        if (!model) return;
                        if (active !== model) {
                            active = model;
                            text(title, model.title);
                            values.forEach((value, index) => text(value, model.values[index] ?? ""));
                            updateContent?.(model);
                        }
                        const opening = root.hidden;
                        if (opening) {
                            // 打开时缓存尺寸和边界，移动时只做数值判断与 transform 更新。
                            root.style.visibility = "hidden";
                            root.hidden = false;
                            measuredWidth = root.offsetWidth;
                            const height = root.offsetHeight;
                            const bounds = container.getBoundingClientRect();
                            maxX = document.documentElement.clientWidth - measuredWidth - 8;
                            maxY = document.documentElement.clientHeight - height - 8;
                            flipX = Math.min(maxX - 16, Math.max((bounds.left + bounds.right) / 2, bounds.right - measuredWidth - 16));
                            offsetX = cursorX > flipX ? -measuredWidth - 16 : 16;
                            offsetY = cursorY + 16 > maxY ? -height - 16 : 16;
                            lastX = lastY = -1;
                        }
                        // 靠近图表右侧自动切左，退回 24px 再切右，避免边界来回抖动。
                        if (offsetX > 0 && cursorX > flipX) offsetX = -measuredWidth - 16;
                        else if (offsetX < 0 && cursorX < flipX - 24) offsetX = 16;
                        const x = Math.max(8, Math.min(cursorX + offsetX, maxX));
                        const y = Math.max(8, Math.min(cursorY + offsetY, maxY));
                        if (x !== lastX || y !== lastY) {
                            root.style.transform = `translate3d(${x}px, ${y}px, 0)`;
                            lastX = x;
                            lastY = y;
                        }
                        if (opening) root.style.visibility = "visible";
                    });
                };
                const dismiss = () => {
                    hide();
                    chart.emit("tooltip:hide", {nativeEvent: false});
                };
                const outside = (event: PointerEvent) => {
                    if (!(event.target instanceof Node) || !container.contains(event.target)) dismiss();
                };
                chart.on("tooltip:show", show);
                chart.on("tooltip:hide", hide);
                chart.on("beforerender", dismiss);
                chart.on("sliderX:filter", dismiss);
                document.addEventListener("pointerdown", outside, true);
                window.addEventListener("scroll", dismiss, {capture: true, passive: true});
                window.addEventListener("resize", dismiss, {passive: true});
                return () => {
                    dismiss();
                    chart.off("tooltip:show", show);
                    chart.off("tooltip:hide", hide);
                    chart.off("beforerender", dismiss);
                    chart.off("sliderX:filter", dismiss);
                    document.removeEventListener("pointerdown", outside, true);
                    window.removeEventListener("scroll", dismiss, true);
                    window.removeEventListener("resize", dismiss);
                    root.remove();
                    node = undefined;
                };
            },
        };
    }, [series, width, className, renderContent, styles.tooltip]);
}
