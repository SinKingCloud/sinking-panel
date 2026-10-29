import {useLayoutEffect, useMemo, useRef} from "react";
import type {Chart} from "@antv/g2";
import useStyles from "./styles";

type TooltipEvent = {nativeEvent?: boolean; client?: {x: number; y: number}; data?: {items?: {time?: number}[]}};
export type ChartTooltipModel = {
    title: string;
    values: string[];
    process?: {rows: {name: string; pid: string; value: string}[]; time: string; state: string};
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

function visible(node: HTMLElement, show: boolean) {
    if (node.hidden === show) node.hidden = !show;
}

export default function useChartTooltip(models: ReadonlyMap<number, ChartTooltipModel>, series: {label: string; color: string}[], process?: {title: string; valueLabel: string}) {
    const {styles} = useStyles({hasProcess: Boolean(process)});
    const data = useMemo(() => {
        let pidWidth = 0;
        let valueWidth = 0;
        if (process) models.forEach((model) => model.process?.rows.forEach((row) => {
            pidWidth = Math.max(pidWidth, row.pid.length);
            valueWidth = Math.max(valueWidth, row.value.length);
        }));
        return {models, pidWidth, valueWidth};
    }, [models, process]);
    const source = useRef(data);
    useLayoutEffect(() => { source.current = data; }, [data]);
    // 数据更新只替换查找表，复用悬浮节点和事件绑定，避免实时图表每秒拆装面板。
    return useMemo(() => {
        let node: HTMLDivElement | undefined;
        let active: ChartTooltipModel | undefined;
        let pending: ChartTooltipModel | undefined;
        let frame = 0;
        const hide = () => {
            window.cancelAnimationFrame(frame);
            frame = 0;
            pending = active = undefined;
            if (node) node.hidden = true;
        };
        return {
            hide,
            bind: (chart: Chart) => {
                const container = chart.getContainer();
                const root = node = element("div", styles.tooltip);
                root.hidden = true;
                const title = element("div", "tooltip-time");
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
                const section = element("div", "process-section");
                const heading = element("div", "process-heading");
                if (process) heading.append(element("span", "", process.title), element("span", "", process.valueLabel));
                const time = element("div", "process-time");
                const list = element("ol", "process-list");
                const rows = Array.from({length: process ? 5 : 0}, () => {
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
                if (process) root.append(section);
                document.body.append(root);
                let cursorX = 0;
                let cursorY = 0;
                let maxX = 0;
                let maxY = 0;
                let offsetX = 16;
                let offsetY = 16;
                let lastX = -1;
                let lastY = -1;

                const show = (event: TooltipEvent) => {
                    if (!event.nativeEvent || !event.client) return;
                    const model = source.current.models.get(Number(event.data?.items?.[0]?.time));
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
                            values.forEach((value, index) => text(value, model.values[index]));
                            if (process) {
                                const info = model.process;
                                const hasRows = Boolean(info?.rows.length);
                                visible(time, hasRows);
                                visible(list, hasRows);
                                visible(state, !hasRows);
                                text(hasRows ? time : state, hasRows ? info!.time : info?.state || "该时段暂无进程记录");
                                rows.forEach((row, index) => {
                                    const item = info?.rows[index];
                                    visible(row.node, Boolean(item));
                                    if (!item) return;
                                    text(row.name, item.name);
                                    text(row.pid, item.pid);
                                    text(row.value, item.value);
                                });
                            }
                        }
                        const opening = root.hidden;
                        if (opening) {
                            // 首次显示测量尺寸并选择方向，移动时仅更新 transform，不反复翻转或测量。
                            root.style.setProperty("--process-pid-width", `${source.current.pidWidth + 1}ch`);
                            root.style.setProperty("--process-value-width", `${source.current.valueWidth + 1}ch`);
                            root.style.visibility = "hidden";
                            root.hidden = false;
                            const width = root.offsetWidth;
                            const height = root.offsetHeight;
                            maxX = document.documentElement.clientWidth - width - 8;
                            maxY = document.documentElement.clientHeight - height - 8;
                            offsetX = cursorX + 16 > maxX ? -width - 16 : 16;
                            offsetY = cursorY + 16 > maxY ? -height - 16 : 16;
                            lastX = lastY = -1;
                        }
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
    }, [series, process, styles.tooltip]);
}
