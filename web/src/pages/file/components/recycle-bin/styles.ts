import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}, {compact = false}: {compact?: boolean} = {}) => ({
    summary: css`
        display: flex;
        width: max-content;
        min-height: ${compact ? 30 : 34}px;
        align-items: center;
        gap: ${compact ? 12 : 16}px;
        color: ${token.colorTextTertiary};
        font-size: ${compact ? 11 : 12}px;
        font-variant-numeric: tabular-nums;
        white-space: nowrap;

        .space-used,
        .entry-count {
            display: inline-flex;
            flex: none;
            align-items: center;
            gap: 6px;
        }

        .space-used strong {
            color: ${token.colorText};
            font-size: ${compact ? 12 : 13}px;
            font-weight: 600;
        }

        .entry-count {
            gap: 4px;
        }

        .entry-count strong {
            color: ${token.colorTextSecondary};
            font-weight: 500;
        }
    `,
}));

export default useStyles;
