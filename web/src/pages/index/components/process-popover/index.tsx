import React from "react";
import {Empty} from "antd";
import {createStyles} from "antd-style";
import {Title} from "sinking-antd";
import {formatSize} from "@/utils/string";

const useStyles = createStyles(({css, token}) => ({
    content: css`
        width: min(288px, calc(100vw - 48px));
        color: ${token.colorText};
        font-size: 12px;
        line-height: 18px;
        font-variant-numeric: tabular-nums;

        .process-heading {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
            font-size: ${token.fontSize}px;
            line-height: ${token.lineHeight};
            font-weight: ${token.fontWeightStrong};
        }
        .process-label { color: ${token.colorTextSecondary}; font-size: 11px; font-weight: 400; }
        .process-list {
            display: grid;
            grid-template-columns: max-content minmax(0, 1fr) max-content;
            align-items: center;
            column-gap: 8px;
            margin: 6px 0 0;
            padding: 0;
            list-style: none;
        }
        .process-row { display: contents; line-height: 24px; }
        .process-pid { color: ${token.colorTextSecondary}; font-size: 11px; white-space: nowrap; }
        .process-name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .process-value { text-align: right; font-weight: 500; white-space: nowrap; }
        .ant-empty { margin: 20px 0; font-size: inherit; }
    `,
}));

const numberLabel = new Intl.NumberFormat("zh-CN", {maximumFractionDigits: 2}).format;

const ProcessPopover = React.memo(({processes, metric}: {processes?: any; metric: "cpu" | "memory"}) => {
    const {styles} = useStyles();
    const items = Array.isArray(processes?.[metric]) ? processes[metric].slice(0, 5) : [];

    return <div className={styles.content}>
        <div className="process-heading">
            <Title size="small">{metric === "cpu" ? "CPU占用TOP" : "内存占用TOP"}</Title>
            <span className="process-label">{metric === "cpu" ? "最近采样均值" : "RSS"}</span>
        </div>
        {items.length > 0 ? <ol className="process-list">
            {items.map((item: any) => <li className="process-row" key={`${item?.pid}-${item?.create_time}`}>
                <span className="process-pid">PID {item?.pid ?? "--"}</span>
                <span className="process-name" title={item?.name}>{item?.name || "未知进程"}</span>
                <span className="process-value">{Number.isFinite(item?.[metric]) ? metric === "cpu" ? `${numberLabel(item.cpu)}%` : formatSize(item.memory) : "--"}</span>
            </li>)}
        </ol> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}
    </div>;
});

export default ProcessPopover;
