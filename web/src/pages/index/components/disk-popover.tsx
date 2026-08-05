import React from "react";
import {Empty, Tooltip, theme as antdTheme} from "antd";
import {createStyles} from "antd-style";
import {formatSize} from "@/utils/string";
import {formatPercent, numberValue, percentValue} from "./helper";

const useStyles = createStyles(({css, token, isDarkMode}: any) => ({
    diskPopover: css`
        width: min(350px, calc(100vw - 28px));

        .disk-overview {
            --disk-color: ${token.colorPrimary};
            padding: 10px 12px;
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorFillQuaternary};
        }

        .disk-overview-title,
        .disk-overview-head,
        .disk-overview-meta,
        .disk-item-head,
        .disk-item-meta {
            min-width: 0;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .disk-overview-title {
            margin-bottom: 7px;
            color: ${token.colorTextTertiary};
            font-size: 10px;
            line-height: 16px;
        }

        .disk-overview-title strong {
            color: ${token.colorTextSecondary};
            font-size: 11px;
            font-weight: 600;
        }

        .disk-overview-head {
            gap: 16px;
        }

        .disk-overview-capacity,
        .disk-overview-value {
            min-width: 0;
        }

        .disk-overview-capacity span,
        .disk-overview-value span {
            display: block;
            color: ${token.colorTextTertiary};
            font-size: 10px;
            line-height: 15px;
        }

        .disk-overview-capacity strong,
        .disk-overview-value strong {
            display: block;
            margin-top: 2px;
            font-size: 18px;
            font-weight: 600;
            line-height: 25px;
            white-space: nowrap;
        }

        .disk-overview-capacity strong {
            color: ${token.colorTextHeading};
        }

        .disk-overview-value {
            text-align: right;
        }

        .disk-overview-value strong {
            color: var(--disk-color);
        }

        .disk-overview-track,
        .disk-track {
            overflow: hidden;
            border-radius: 999px;
            background: color-mix(in srgb, var(--disk-color) 10%, ${token.colorFillSecondary});
        }

        .disk-overview-track {
            height: 5px;
            margin-top: 9px;
        }

        .disk-overview-track i,
        .disk-track i {
            display: block;
            width: var(--disk-percent);
            height: 100%;
            border-radius: inherit;
            background: var(--disk-color);
            transition: width .3s ease;
        }

        .disk-overview-meta {
            justify-content: flex-start;
            gap: 16px;
            margin-top: 7px;
            color: ${token.colorTextTertiary};
            font-size: 10px;
            line-height: 15px;
        }

        .disk-overview-meta b {
            margin-left: 4px;
            color: ${token.colorTextSecondary};
            font-weight: 500;
        }

        .disk-list {
            max-height: 230px;
            display: flex;
            flex-direction: column;
            gap: 7px;
            margin-top: 9px;
            padding: 1px 1px 2px;
            overflow-y: auto;
        }

        .disk-item {
            --disk-color: ${token.colorPrimary};
            padding: 9px 10px;
            border: 1px solid transparent;
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorFillQuaternary};
            transition: border-color .2s ease, background-color .2s ease;
        }

        @media (hover: hover) {
            .disk-item:hover {
                border-color: color-mix(in srgb, var(--disk-color) 28%, ${token.colorBorderSecondary});
                background: color-mix(in srgb, var(--disk-color) ${isDarkMode ? 6 : 4}%, ${token.colorFillQuaternary});
            }
        }

        .disk-item-head {
            gap: 14px;
        }

        .disk-item-path {
            min-width: 0;
            flex: 1;
            display: block;
            color: ${token.colorText};
            font-size: 12px;
            font-weight: 600;
            line-height: 19px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .disk-item-percent {
            flex: none;
            color: var(--disk-color);
            font-size: 12px;
            font-weight: 600;
            line-height: 19px;
        }

        .disk-item-meta {
            gap: 10px;
            margin-top: 3px;
            color: ${token.colorTextQuaternary};
            font-size: 9px;
            line-height: 14px;
        }

        .disk-item-meta span {
            min-width: 0;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .disk-item-meta span:last-child {
            flex: none;
            color: ${token.colorTextTertiary};
        }

        .disk-item-meta b {
            color: ${token.colorTextSecondary};
            font-weight: 500;
        }

        .disk-track {
            height: 3px;
            margin-top: 7px;
        }

        @media (max-width: 575px) {
            width: min(320px, calc(100vw - 24px));

            .disk-item-head {
                gap: 10px;
            }
        }
    `,
    empty: css`
        padding: 38px 0 30px;

        .ant-empty-description {
            color: ${token.colorTextTertiary};
            font-size: 12px;
        }
    `,
}));

const DiskPopover = React.memo(({
    diskFree,
    diskPercent,
    diskSummary,
    summaryDisks,
}: any) => {
    const {styles} = useStyles();
    const {token} = antdTheme.useToken();
    const overviewPercent = percentValue(diskPercent);
    const overviewColor = overviewPercent >= 90
        ? token.colorError
        : (overviewPercent >= 75 ? token.colorWarning : token.colorPrimary);

    return (
        <div className={styles.diskPopover}>
            <div
                className="disk-overview"
                style={{
                    "--disk-color": overviewColor,
                    "--disk-percent": `${overviewPercent}%`,
                } as React.CSSProperties}
            >
                <div className="disk-overview-title">
                    <strong>容量概览</strong>
                    <span>{summaryDisks.length} 个挂载点</span>
                </div>
                <div className="disk-overview-head">
                    <div className="disk-overview-capacity">
                        <span>总容量</span>
                        <strong>{formatSize(diskSummary.total)}</strong>
                    </div>
                    <div className="disk-overview-value">
                        <span>整体使用率</span>
                        <strong>{formatPercent(overviewPercent)}</strong>
                    </div>
                </div>
                <div className="disk-overview-track"><i/></div>
                <div className="disk-overview-meta">
                    <span>已使用<b>{formatSize(diskSummary.used)}</b></span>
                    <span>可用空间<b>{formatSize(diskFree)}</b></span>
                </div>
            </div>
            {summaryDisks.length > 0 ? (
                <div className="disk-list">
                    {summaryDisks.map((item: any, index: number) => {
                        const value = percentValue(item?.size?.percentage);
                        const total = numberValue(item?.size?.total);
                        const used = numberValue(item?.size?.used);
                        const color = value >= 90
                            ? token.colorError
                            : (value >= 75 ? token.colorWarning : token.colorPrimary);
                        return (
                            <div
                                className="disk-item"
                                key={`${item.path}-${item.filesystem}-${index}`}
                                style={{
                                    "--disk-color": color,
                                    "--disk-percent": `${value}%`,
                                } as React.CSSProperties}
                            >
                                <div className="disk-item-head">
                                    <Tooltip title={item.path} placement="topLeft">
                                        <strong className="disk-item-path">{item.path || "未知挂载点"}</strong>
                                    </Tooltip>
                                    <span className="disk-item-percent">{formatPercent(value)}</span>
                                </div>
                                <div className="disk-item-meta">
                                    <span>{item.filesystem || "--"} · {item.type || "未知类型"}</span>
                                    <span>已用 <b>{formatSize(used)}</b> / {formatSize(total)}</span>
                                </div>
                                <div className="disk-track"><i/></div>
                            </div>
                        );
                    })}
                </div>
            ) : <Empty className={styles.empty} image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无磁盘信息"/>}
        </div>
    );
});

export default DiskPopover;
