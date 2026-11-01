import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    pagination: css`
        container-name: oss-pagination;
        container-type: inline-size;
        width: 100%;
        box-sizing: border-box;
        padding: ${compact ? "10px 12px" : "14px 16px"};
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorFillQuaternary};
        color: ${token.colorTextSecondary};
        font-size: ${compact ? 11 : 12}px;
        font-variant-numeric: tabular-nums;
        white-space: nowrap;

        .pagination-count {
            margin-inline-end: 8px;
        }

        .pagination-content,
        .pagination-pages {
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .pagination-content {
            flex-wrap: wrap;
            gap: 6px 0;
        }

        .pagination-pages {
            flex: none;
            gap: 4px;
        }

        .pagination-pages .ant-btn,
        .pagination-current {
            box-sizing: border-box;
            min-width: ${compact ? 22 : 26}px;
            height: ${compact ? 22 : 26}px;
            padding-inline: 6px;
            border-radius: ${token.borderRadiusSM}px;
            font-size: inherit;
        }

        .pagination-pages .ant-btn:not(:disabled) {
            color: ${token.colorTextTertiary};
        }

        .pagination-pages .ant-btn:not(:disabled):hover {
            background: ${token.colorFillTertiary};
        }

        .pagination-current {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            background: ${token.colorBgContainer};
            box-shadow: inset 0 0 0 1px ${token.colorPrimaryBorder};
            color: ${token.colorPrimary};
        }

        @container oss-pagination (max-width: 767px) {
            .pagination-count {
                display: none;
            }

            .pagination-content > .ant-select {
                margin-inline-start: ${compact ? 8 : 10}px;
            }
        }
    `,
    // 与 ModalTable 内的 PageTable 分页下拉保持相同尺寸和选中样式。
    pageSizeSelect: css`
        && {
            width: ${compact ? 96 : 104}px;
            height: ${compact ? 22 : 26}px;
            margin-inline-start: ${compact ? 10 : 12}px;
            color: ${token.colorTextSecondary} !important;
            font-size: ${compact ? 11 : 12}px;
            text-align: left;
        }

        && .ant-select-suffix {
            color: ${token.colorTextQuaternary};
            font-size: ${compact ? 9 : 10}px;
        }
    `,
    pageSizeDropdown: css`
        && {
            padding: 2px;
            border-radius: ${token.borderRadius}px;
        }

        && .ant-select-item {
            min-height: ${compact ? 24 : 28}px;
            padding: 3px 6px;
            border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
        }

        && .ant-select-item-option-selected:not(.ant-select-item-option-disabled) {
            background: ${token.colorFillTertiary};
            color: ${token.colorTextSecondary};
            font-weight: 400;
        }
    `,
    hint: css`
        display: block;
        margin-top: 8px;
    `,
}));

export default useStyles;
