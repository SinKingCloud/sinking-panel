import {COMPONENT_CLASS_NAME, MAIN_LAYER_CLASS_NAME, register, select, selectPlotArea} from "@antv/g2";
import type {Path} from "@antv/g";
import type {Axis, Slider} from "@antv/component";

type Range = [number, number];
type Options = {
    getValues: () => Range;
    onChange: (values: Range) => void;
    start: number;
    end: number;
    labelFormatter: (value: number) => string;
};

// 保留原生滑块手势；拖动只改现有路径，松手再更新 G2 的坐标和交互数据。
function rangeInteraction({getValues, onChange, start, end, labelFormatter}: Options) {
    return (context: any, _: unknown, emitter: any) => {
        const {container, setState, update} = context;
        const slider = container.getElementsByClassName("slider")[0] as Slider | undefined;
        const plot = selectPlotArea(container);
        if (!slider || !plot || end <= start) return () => {};
        let view = context.view;
        let disposed = false;
        let dragging = false;
        let updating = false;
        let queued = false;
        let frame = 0;
        let timer = 0;
        let paths: {node: Path; commands: Path["parsedStyle"]["d"]["absolutePath"]}[] = [];
        let axis: Axis | undefined;
        let ticks: Axis["attributes"]["data"] = [];
        let base: Range = [start, end];
        let rendered: Range = base;
        let labels = "";
        let left = 0;
        let width = 0;
        let fallback = false;
        // G 会忽略 undefined 赋值，恢复缺省值时需显式开启命中检测。
        const pointerEvents = plot.style.pointerEvents || "auto";
        const domain = (values: Range): Range => values.map((value) => start + value * (end - start)) as Range;
        const equal = (a: Range, b: Range) => a[0] === b[0] && a[1] === b[1];

        const capture = () => {
            base = view.scale.x.getOptions().domain.map(Number) as Range;
            rendered = base;
            const coordinate = view.coordinate.getOptions();
            left = coordinate.x || 0;
            width = coordinate.width;
            paths = [];
            fallback = false;
            select(plot).selectAll(`.${MAIN_LAYER_CLASS_NAME}`).each((mark: any, _index, layer: any) => {
                if (mark.type !== "line" && mark.type !== "area") {
                    fallback = true;
                    return;
                }
                for (const node of layer.getElementsByTagName("path") as Path[]) {
                    const commands = node.parsedStyle.d.absolutePath;
                    if (!commands.every((command) => ["M", "L", "C", "Z"].includes(command[0]))) {
                        fallback = true;
                        continue;
                    }
                    paths.push({node, commands});
                }
            });
            const axes: Axis[] = [];
            select(container).selectAll(`.${COMPONENT_CLASS_NAME}`).each((component: any, _index, group: any) => {
                if (component.type === "axisX") axes.push(group.children[0] as Axis);
            });
            axis = axes[0];
            ticks = axis?.attributes.data?.map((tick) => ({...tick})) || [];
            labels = ticks.map((tick) => tick.label).join("\n");
        };

        const preview = () => {
            frame = 0;
            if (disposed || updating) return;
            if (fallback) {
                void commit();
                return;
            }
            const next = domain(getValues());
            if (equal(next, rendered)) return;
            const size = next[1] - next[0];
            if (size <= 0) return;
            const scale = (base[1] - base[0]) / size;
            const offset = left * (1 - scale) + width * (base[0] - next[0]) / size;
            // 复用已经算好的贝塞尔曲线，仅映射 x；直接缩放图层会把线宽也放大。
            paths.forEach(({node, commands}) => node.attr("d", commands.map((command) =>
                command.map((value, index) => index % 2 ? Number(value) * scale + offset : value),
            ) as typeof commands));
            const data = ticks.map((tick) => ({...tick, label: labelFormatter(next[0] + tick.value * size)}));
            const nextLabels = data.map((tick) => tick.label).join("\n");
            if (nextLabels !== labels) axis?.update({data}, false);
            labels = nextLabels;
            rendered = next;
        };

        const commit = async () => {
            if (disposed) return;
            if (updating) {
                queued = true;
                return;
            }
            const next = domain(getValues());
            if (equal(next, base)) return;
            updating = true;
            queued = false;
            try {
                setState(slider, (options: any) => ({
                    ...options, clip: true, animate: false,
                    marks: options.marks.map((mark: any) => ({
                        ...mark,
                        scale: {...mark.scale, x: {...mark.scale?.x, domain: next, nice: false}},
                        slider: {...mark.slider, x: {...mark.slider?.x, preserve: true}},
                        animate: false,
                    })),
                }));
                const result = await update();
                if (disposed) return;
                view = result.view;
                capture();
            } finally {
                updating = false;
            }
            if (disposed) return;
            if (!equal(domain(getValues()), base)) {
                // 提交期间仍接收最新位置，避免松手时丢掉最后一次移动。
                if (queued || !dragging || fallback) void commit();
                else preview();
            }
        };

        const finish = () => {
            if (!dragging) return;
            dragging = false;
            window.cancelAnimationFrame(frame);
            frame = 0;
            plot.style.pointerEvents = pointerEvents;
            preview();
            void commit();
        };
        const begin = () => {
            dragging = true;
            window.clearTimeout(timer);
            plot.style.pointerEvents = "none";
            emitter.emit("tooltip:hide", {nativeEvent: false});
        };
        const change = (event: {detail: {value: Range}}) => {
            const values = event.detail.value.map((value) => Math.max(0, Math.min(1, value))) as Range;
            if (!values.every(Number.isFinite) || values[0] >= values[1] || equal(values, getValues())) return;
            onChange(values);
            emitter.emit("sliderX:filter", {nativeEvent: true, data: {selection: [domain(values)]}});
            if (!frame) frame = window.requestAnimationFrame(preview);
            if (!dragging) {
                window.clearTimeout(timer);
                timer = window.setTimeout(() => void commit(), 100);
            }
        };

        capture();
        slider.setValues(getValues(), false);
        slider.addEventListener("pointerdown", begin, true);
        slider.addEventListener("valuechange", change);
        document.addEventListener("pointerup", finish, true);
        document.addEventListener("pointercancel", finish, true);
        if (!equal(domain(getValues()), base)) {
            // 首次刷新先同步呈现保存的选区，不让完整范围闪现一帧。
            preview();
            void commit();
        }
        return () => {
            disposed = true;
            window.cancelAnimationFrame(frame);
            window.clearTimeout(timer);
            slider.removeEventListener("pointerdown", begin, true);
            slider.removeEventListener("valuechange", change);
            document.removeEventListener("pointerup", finish, true);
            document.removeEventListener("pointercancel", finish, true);
            plot.style.pointerEvents = pointerEvents;
        };
    };
}

// 具名注册，避免 Charts 把 Babel 的 PURE 注释误识别成 JSX 并包装交互函数。
register("interaction.monitorTimeRange", rangeInteraction);
