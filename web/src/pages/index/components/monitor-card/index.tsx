import React from "react";
import {Card, Select} from "antd";
import {createStyles} from "antd-style";
import {Title} from "sinking-antd";
import TelemetryChart from "../telemetry-chart";

const useStyles = createStyles(({css, token, isDarkMode}: any) => ({
    monitorCard: css`
        height: 100%;
        overflow: hidden;
        border-radius: ${token.borderRadiusLG}px;

        .ant-card-head {
            min-height: 48px;
            margin-bottom: 0;
            border-bottom: 0;
            box-shadow: inset 0 -1px 0 ${isDarkMode ? "rgba(255,255,255,0.06)" : "rgba(5,5,5,0.06)"};
        }

        .ant-card-head-title,
        .ant-card-extra {
            min-width: 0;
        }

        .ant-card-extra {
            flex: none;
            margin-inline-start: 12px;
        }

        .ant-card-body {
            padding: 0;
            overflow: hidden;
        }
    `,
    deviceSelect: css`
        && {
            width: 86px;
            height: 28px;
            color: ${token.colorTextSecondary};
            font-size: 12px;
        }

        && .ant-select-selector {
            height: 28px !important;
            padding-inline: 9px 26px !important;
            border: 0 !important;
            border-radius: ${token.borderRadiusSM}px !important;
            background: ${token.colorFillQuaternary} !important;
            box-shadow: inset 0 0 0 1px transparent !important;
            transition: background-color .2s ease, box-shadow .2s ease;
        }

        && .ant-select-content,
        && .ant-select-content-value,
        && .ant-select-selection-item {
            min-width: 0;
            color: ${token.colorTextSecondary} !important;
            font-size: 12px !important;
            font-weight: 500 !important;
            line-height: 28px !important;
        }

        && .ant-select-arrow {
            color: ${token.colorTextTertiary};
            font-size: 10px;
        }

        &&:hover .ant-select-selector {
            background: ${token.colorFillTertiary} !important;
            box-shadow: inset 0 0 0 1px ${token.colorBorderSecondary} !important;
        }

        &&.ant-select-focused .ant-select-selector,
        &&.ant-select-open .ant-select-selector {
            background: ${token.colorPrimaryBg} !important;
            box-shadow: inset 0 0 0 1px color-mix(in srgb, ${token.colorPrimary} 32%, transparent) !important;
        }
    `,
    deviceDropdown: css`
        && {
            padding: 3px;
            border-radius: ${token.borderRadius}px;
        }

        && .ant-select-item {
            min-height: 30px;
            padding: 5px 9px;
            border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary};
            font-size: 12px;
            line-height: 20px;
        }

        && .ant-select-item-option-content {
            overflow: hidden;
            font-size: 12px;
            font-weight: 400;
            text-overflow: ellipsis;
        }

        && .ant-select-item-option-active:not(.ant-select-item-option-disabled) {
            background: ${token.colorFillQuaternary};
        }

        && .ant-select-item-option-selected:not(.ant-select-item-option-disabled) {
            background: ${token.colorFillTertiary};
            color: ${token.colorText};
            font-weight: 500;
        }
    `,
    monitorBody: css`
        width: 100%;
        min-width: 0;
        display: grid;
        grid-template-columns: minmax(0, 1fr);
        align-items: stretch;
    `,
    throughputStats: css`
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        align-content: center;
        padding: 10px 14px 0;

        .throughput-stat {
            position: relative;
            min-width: 0;
            min-height: 66px;
            padding: 8px 14px;
            display: flex;
            flex-direction: column;
            justify-content: center;
            box-sizing: border-box;
        }

        .throughput-stat + .throughput-stat::before {
            content: "";
            position: absolute;
            inset-inline-start: 0;
            top: 16px;
            bottom: 16px;
            width: 1px;
            background: ${isDarkMode ? "rgba(255,255,255,0.08)" : "rgba(5,5,5,0.08)"};
        }

        .stat-label {
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 7px;
            overflow: hidden;
            color: ${token.colorTextTertiary};
            font-size: 11px;
            font-weight: 500;
            line-height: 16px;
            white-space: nowrap;
            text-overflow: ellipsis;
            letter-spacing: 0;
        }

        .stat-label::before {
            content: "";
            width: 3px;
            height: 12px;
            flex: none;
            border-radius: 2px;
            background: var(--stat-color);
        }

        .stat-value {
            min-width: 0;
            margin-top: 5px;
            display: flex;
            align-items: baseline;
            gap: 3px;
            overflow: hidden;
            white-space: nowrap;
        }

        .stat-value strong {
            min-width: 0;
            overflow: hidden;
            color: ${token.colorTextHeading};
            font-size: 17px;
            font-weight: 600;
            font-variant-numeric: tabular-nums;
            line-height: 24px;
            text-overflow: ellipsis;
            letter-spacing: 0;
        }

        .stat-value small {
            flex: none;
            color: ${token.colorTextTertiary};
            font-size: 10px;
            font-weight: 500;
            line-height: 16px;
        }

        @media (max-width: 575px) {
            grid-template-columns: repeat(2, minmax(0, 1fr));

            .throughput-stat:nth-child(odd)::before {
                display: none;
            }

            .throughput-stat:nth-child(n + 3)::after {
                content: "";
                position: absolute;
                top: 0;
                inset-inline: 14px;
                height: 1px;
                display: block;
                background: ${isDarkMode ? "rgba(255,255,255,0.08)" : "rgba(5,5,5,0.08)"};
            }
        }
    `,
    throughputChart: css`
        width: 100%;
        min-width: 0;
        height: 314px;
        margin: 8px 0 12px;
        padding: 0 14px;
        display: flex;
        flex-direction: column;
        overflow: hidden;

        .telemetry-toolbar {
            height: 26px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            flex: none;
            padding: 0 2px;
        }

        .telemetry-legends {
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 14px;
        }

        .telemetry-legend {
            min-width: 0;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            color: ${token.colorTextTertiary};
            font-size: 10px;
            font-weight: 500;
            line-height: 16px;
            white-space: nowrap;
        }

        .telemetry-legend i {
            width: 7px;
            height: 7px;
            flex: none;
            border-radius: 50%;
            background: var(--legend-color);
        }

        .telemetry-unit {
            flex: none;
            color: ${token.colorTextQuaternary};
            font-size: 10px;
            font-weight: 500;
            line-height: 16px;
        }

        .telemetry-plot {
            width: 100%;
            min-height: 0;
            flex: 1;
        }

        @media (max-width: 576px) {
            height: 276px;
            margin-bottom: 10px;
            padding: 0 10px;
        }
    `,
}));

const MonitorCard = React.memo(({
    title,
    device,
    options,
    onDeviceChange,
    metrics,
    data,
    first,
    second,
    isDarkMode,
}: any) => {
    const {styles} = useStyles();

    return (
        <Card
            className={styles.monitorCard}
            variant="borderless"
            title={<Title>{title}</Title>}
            extra={(
                <Select
                    className={styles.deviceSelect}
                    size="small"
                    value={device}
                    options={options}
                    variant="filled"
                    classNames={{popup: {root: styles.deviceDropdown}}}
                    onChange={onDeviceChange}
                />
            )}
        >
            <div className={styles.monitorBody}>
                <div className={styles.throughputStats}>
                    {metrics.map((item: any) => (
                        <div
                            className="throughput-stat"
                            key={item.label}
                            style={{"--stat-color": item.color} as React.CSSProperties}
                        >
                            <span className="stat-label">{item.label}</span>
                            <span className="stat-value">
                                <strong>{item.value}</strong>
                                {item.suffix && <small>{item.suffix}</small>}
                            </span>
                        </div>
                    ))}
                </div>
                <TelemetryChart
                    className={styles.throughputChart}
                    data={data}
                    first={first}
                    second={second}
                    isDarkMode={isDarkMode}
                />
            </div>
        </Card>
    );
});

MonitorCard.displayName = "MonitorCard";

export default MonitorCard;
