import {useMemo} from "react";
import {createStyles} from "antd-style";
import useChartTooltip from "@/pages/components/ui/chart-tooltip";
import type {ChartTooltipModel, ChartTooltipOptions} from "@/pages/components/ui/chart-tooltip";

export type TelemetryTooltipModel = ChartTooltipModel & {
    iops?: {read: string; write: string};
    ranking?: {
        title: string;
        valueLabel: string;
        rows: {name: string; pid: string; value: string}[];
    };
};

const useStyles = createStyles(({css, token}) => ({
    tooltip: css`
        .ranking-section { border-top: 1px solid ${token.colorBorderSecondary}; margin-top: 8px; padding-top: 8px; }
        .ranking-heading { display: flex; justify-content: space-between; gap: 8px; font-weight: 500; }
        .ranking-heading span:last-child { flex: none; color: ${token.colorTextSecondary}; font-size: 11px; font-weight: 400; }
        .ranking-list { display: grid; grid-template-columns: 9ch minmax(0, 1fr) 12ch; align-items: center; column-gap: 8px; margin: 6px 0 0; padding: 0; list-style: none; }
        .ranking-list[data-network="true"] { grid-template-columns: minmax(0, 1fr) 12ch; }
        .ranking-row { display: contents; line-height: 24px; }
        .ranking-pid { color: ${token.colorTextSecondary}; font-size: 11px; white-space: nowrap; }
        .ranking-name { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
        .ranking-value { text-align: right; font-weight: 500; white-space: nowrap; }
    `,
}));

export default function useTooltip(models: ReadonlyMap<number, TelemetryTooltipModel>, series: {label: string; color: string}[]) {
    const {styles} = useStyles();
    const options = useMemo<ChartTooltipOptions<TelemetryTooltipModel>>(() => ({
        width: 304,
        className: styles.tooltip,
        renderContent: (container) => {
            const create = (tag: "div" | "span" | "ol" | "li" | "i" | "strong", className: string) => {
                const node = document.createElement(tag);
                node.className = className;
                return node;
            };
            const iopsSection = create("div", "");
            iopsSection.hidden = true;
            const iopsValues = ["读取 IOPS", "写入 IOPS"].map((label, index) => {
                const row = create("div", "tooltip-metric");
                const dot = create("i", "");
                dot.style.backgroundColor = series[index]?.color || "currentColor";
                const name = create("span", "");
                name.textContent = label;
                const value = create("strong", "");
                row.append(dot, name, value);
                iopsSection.append(row);
                return value;
            });
            const section = create("div", "ranking-section");
            section.hidden = true;
            const heading = create("div", "ranking-heading");
            const title = create("span", "");
            const valueLabel = create("span", "");
            heading.append(title, valueLabel);
            const list = create("ol", "ranking-list");
            const rows = Array.from({length: 5}, () => {
                const node = create("li", "ranking-row");
                const pid = create("span", "ranking-pid");
                const name = create("span", "ranking-name");
                const value = create("span", "ranking-value");
                node.append(pid, name, value);
                list.append(node);
                return {node, pid, name, value};
            });
            section.append(heading, list);
            container.append(iopsSection, section);
            let previousIOPS = false;
            let previousCount = 0;
            let previousHeading = "";
            const text = (node: HTMLElement, value: string) => {
                if (node.textContent !== value) node.textContent = value;
            };
            // 同一套节点复用所有历史点，仅修改文本；采样更新不会拆装 tooltip。
            return ({ranking, iops}) => {
                const count = ranking?.rows.length || 0;
                const label = ranking?.valueLabel || "";
                const currentHeading = `${ranking?.title || ""}/${label}`;
                const hasIOPS = Boolean(iops);
                const changed = previousCount !== count || previousHeading !== currentHeading || previousIOPS !== hasIOPS;
                previousIOPS = hasIOPS;
                previousCount = count;
                previousHeading = currentHeading;
                iopsSection.hidden = !hasIOPS;
                if (iops) {
                    text(iopsValues[0], iops.read);
                    text(iopsValues[1], iops.write);
                }
                section.hidden = !count;
                if (!ranking || !count) return changed;
                text(title, ranking.title);
                text(valueLabel, label);
                list.dataset.network = String(!ranking.rows[0]?.pid);
                rows.forEach((row, index) => {
                    const item = ranking.rows[index];
                    row.node.hidden = !item;
                    if (!item) return;
                    row.pid.hidden = !item.pid;
                    text(row.pid, item.pid);
                    text(row.name, item.name);
                    text(row.value, item.value);
                });
                return changed;
            };
        },
    }), [styles.tooltip, series]);
    return useChartTooltip(models, series, options);
}
