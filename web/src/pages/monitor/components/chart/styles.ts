import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    panel: css`
        container-type: inline-size;
        container-name: monitor-chart;
        min-width: 0;
        height: 100%;
        box-sizing: border-box;
        padding: 16px 18px 12px;
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorBgContainer};
        font-variant-numeric: tabular-nums;

        @media (max-width: 575px) {
            padding: 14px 12px 10px;
        }
    `,
    heading: css`
        display: flex;
        align-items: center;
        justify-content: space-between;
        flex-wrap: wrap;
        gap: 4px 12px;
        min-height: 22px;

        h3 {
            margin: 0;
            color: ${token.colorTextHeading};
            font-size: 14px;
            font-weight: 600;
            line-height: 22px;
        }

        .sample-time {
            color: ${token.colorTextTertiary};
            font-size: 11px;
            line-height: 18px;
            white-space: nowrap;
        }

        .sample-time time { margin-inline-start: 3px; }
    `,
    summary: css`
        display: flex;
        align-items: center;
        gap: 12px;
        min-height: 58px;
        padding-top: 8px;

        .summary-metrics {
            min-width: 0;
            display: grid;
            flex: 1;
            gap: 12px;
        }

        .summary-metric {
            min-width: 0;
            display: flex;
            align-items: baseline;
            flex-wrap: wrap;
            gap: 2px 9px;
        }

        .metric-label {
            min-width: 0;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            color: ${token.colorTextSecondary};
            font-size: 11px;
            line-height: 18px;
            white-space: nowrap;
        }

        .metric-label i {
            width: 6px;
            height: 6px;
            flex: none;
            border-radius: 50%;
            background: var(--series-color);
        }

        .metric-value {
            display: inline-flex;
            align-items: baseline;
            gap: 3px;
            min-width: 0;
            white-space: nowrap;
        }

        .metric-value strong {
            color: ${token.colorTextHeading};
            font-size: ${compact ? 16 : 18}px;
            font-weight: 500;
            line-height: ${compact ? 22 : 24}px;
        }

        .metric-value small {
            color: ${token.colorTextSecondary};
            font-size: 11px;
            line-height: 18px;
        }

        .summary-detail {
            min-width: 0;
            color: ${token.colorTextSecondary};
            font-size: 11px;
            line-height: 18px;
            text-align: right;
        }

        @container monitor-chart (max-width: 440px) {
            .summary-metric { flex-direction: column; align-items: flex-start; gap: 0; }
        }
    `,
    plot: css`
        min-width: 0;
        margin-top: 2px;

        /* 横向手势用于选区，纵向和双指手势交给浏览器；覆盖画布的内联 none。 */
        canvas { touch-action: pan-y pinch-zoom !important; }

        .chart-unit {
            height: 17px;
            color: ${token.colorTextTertiary};
            font-size: 10px;
            line-height: 17px;
            text-align: right;
        }

        .chart-empty { display: flex; align-items: center; justify-content: center; }
        .ant-empty-description { color: ${token.colorTextTertiary}; font-size: 12px; }
    `,
}));
