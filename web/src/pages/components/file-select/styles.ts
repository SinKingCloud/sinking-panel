import {createStyles} from "antd-style";

const useStyles = createStyles<{compact: boolean}>(({css, token}, props) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
        }

        .ant-modal-content {
            max-height: calc(100dvh - 24px);
            overflow: hidden;
            display: flex;
            flex-direction: column;
        }

        .ant-modal-body {
            min-height: 0;
            min-width: 0;
            overflow-y: auto;
        }

    `,
    field: css`
        width: 100%;
        min-width: 0;

        .ant-input, .ant-input-affix-wrapper, .ant-select {
            min-width: 0;
            flex: 1;
        }

        > .ant-btn { flex: none; }

        .ant-btn-color-default:not(:disabled) .ant-btn-icon {
            color: ${token.colorTextSecondary};
        }

        && .ant-input:disabled,
        && .ant-input-affix-wrapper-disabled,
        && .ant-input-affix-wrapper-disabled:hover,
        && .ant-select-disabled,
        && .ant-btn:disabled {
            border-color: transparent;
            box-shadow: none;
        }
    `,
    browser: css`
        min-width: 0;
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;
        background: ${token.colorBgContainer};

        .path-select-toolbar {
            display: grid;
            min-width: 0;
            grid-template-columns: auto minmax(0, 1fr) auto;
            align-items: center;
            gap: ${props.compact ? 6 : 8}px;
            padding: ${props.compact ? 7 : 9}px;
            border-bottom: 1px solid ${token.colorSplit};
            background: ${token.colorFillQuaternary};
        }

        .path-select-actions {
            display: flex;
            flex: none;
            align-items: center;
            gap: ${props.compact ? 6 : 8}px;
        }

        .path-select-up,
        .path-select-path,
        .path-select-path-input.ant-input {
            border: 1px solid transparent;
            border-radius: ${token.borderRadius}px;
            background: ${token.colorBgContainer};
        }

        .path-select-up {
            width: ${props.compact ? 28 : 30}px;
            min-width: ${props.compact ? 28 : 30}px;
            height: ${props.compact ? 28 : 30}px;
            padding: 0;
            color: ${token.colorTextSecondary};
        }

        .path-select-up:not(:disabled):hover {
            border-color: ${token.colorBorderSecondary};
            background: ${token.colorFillTertiary};
            color: ${token.colorText};
        }

        .path-select-up.ant-btn:disabled {
            /* 保持边框尺寸和颜色，避免刷新切换禁用状态时闪边。 */
            border: 1px solid transparent;
            background: ${token.colorBgContainer};
            color: ${token.colorTextDisabled};
            box-shadow: none;
        }

        .path-select-path {
            box-sizing: border-box;
            height: ${props.compact ? 30 : 32}px;
            min-width: 0;
            margin: 0;
            padding: 0 ${props.compact ? 6 : 8}px;
            display: flex;
            align-items: center;
            overflow-x: auto;
            overflow-y: hidden;
            list-style: none;
            scrollbar-width: none;

            &::-webkit-scrollbar {
                display: none;
            }
        }

        .path-select-path li {
            min-width: 0;
            display: inline-flex;
            flex: none;
            align-items: center;
        }

        .path-select-path-separator {
            margin-inline: ${props.compact ? 2 : 3}px;
            color: ${token.colorTextQuaternary};
            font-size: ${props.compact ? 8 : 9}px;
        }

        .path-select-path-button,
        .path-select-path-current {
            max-width: ${props.compact ? 150 : 190}px;
            min-width: 0;
            padding: 0 ${props.compact ? 2 : 3}px;
            display: inline-flex;
            align-items: center;
            border: 0;
            border-radius: ${token.borderRadiusSM}px;
            outline: none;
            appearance: none;
            background: transparent;
            color: ${token.colorTextSecondary};
            font: inherit;
            font-size: ${props.compact ? 11 : 12}px;
            line-height: ${props.compact ? 18 : 20}px;
            white-space: nowrap;
            cursor: pointer;
            transition: background-color ${token.motionDurationFast}, color ${token.motionDurationFast};
        }

        .path-select-path-button > span,
        .path-select-path-current > span {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .path-select-path-button:not(:disabled):hover {
            background: ${token.colorFillTertiary};
            color: ${token.colorText};
        }

        .path-select-path-button:disabled {
            color: ${token.colorTextDisabled};
            cursor: not-allowed;
        }

        .path-select-path-button:focus-visible {
            box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
        }

        .path-select-path-current {
            color: ${token.colorText};
            cursor: text;
            font-weight: 500;
        }

        .path-select-path-current:not(:disabled):hover {
            background: ${token.colorFillTertiary};
            color: ${token.colorPrimary};
        }

        .path-select-path-current:focus-visible {
            color: ${token.colorPrimary};
            box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
        }

        .path-select-path-current:disabled {
            color: ${token.colorText};
            cursor: default;
        }

        .path-select-path-input.ant-input {
            width: min(100%, ${props.compact ? 420 : 480}px);
            border-width: 0.5px;
            min-width: 40px;
            max-width: 100%;
            height: ${props.compact ? 30 : 32}px;
            padding: 0 ${props.compact ? 6 : 8}px;
            justify-self: start;
            color: ${token.colorText};
            box-shadow: none;
            font-size: ${props.compact ? 11 : 12}px;
        }

        .path-select-path-input.ant-input:hover {
            border-color: transparent;
        }

        .path-select-path-input.ant-input:focus {
            border-color: ${token.colorBorder};
            outline: none;
            box-shadow: none;
        }

        .path-select-table.ant-table-wrapper {
            --ant-table-header-border-radius: 0px;
        }

        .path-select-table .ant-table,
        .path-select-table .ant-table-container {
            background: transparent;
        }

        .path-select-table .ant-table-thead > tr > th {
            padding: ${props.compact ? "7px 10px" : "9px 12px"} !important;
            border-bottom: 1px solid ${token.colorSplit};
            background: ${token.colorBgContainer};
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;
            font-weight: 500;
        }

        .path-select-table .ant-table-tbody > tr > td {
            padding: ${props.compact ? "7px 10px" : "9px 12px"} !important;
            border-bottom: 1px solid ${token.colorSplit};
            background: ${token.colorBgContainer};
            font-size: ${token.fontSizeSM}px;
            transition: background-color ${token.motionDurationFast};
        }

        .path-select-table .ant-table-tbody > tr.ant-table-measure-row > td {
            padding-block: 0 !important;
            border-block: 0 !important;
        }

        .path-select-action-cell .ant-btn {
            height: 25px;
            padding-inline: 10px;
            border-radius: ${token.borderRadius}px;
            font-size: ${token.fontSizeSM}px;
        }

        .path-select-table .ant-table-tbody > tr:last-child > td {
            border-bottom: 0;
        }

        /* 短列表补末行分隔线；列表填满时由滚动区裁掉，避免与状态栏双线。 */
        .path-select-table .ant-table:not(.ant-table-empty) .ant-table-body > table {
            box-shadow: 0 1px 0 ${token.colorSplit};
        }

        .path-select-table .ant-table-body {
            min-height: ${props.compact
                ? "clamp(180px, 34dvh, 238px)"
                : "clamp(200px, 38dvh, 278px)"};
            scrollbar-width: thin;
        }

        .path-select-table .ant-table-placeholder td {
            height: ${props.compact
                ? "clamp(180px, 34dvh, 238px)"
                : "clamp(200px, 38dvh, 278px)"};
            border-bottom: 0;
        }

        .path-select-name-cell {
            display: flex;
            width: max-content;
            align-items: center;
            gap: ${props.compact ? 7 : 9}px;
        }

        .path-select-icon {
            flex: none;
            color: ${token.colorTextSecondary};
            font-size: ${props.compact ? 14 : 16}px;
        }

        .path-select-directory .path-select-icon {
            color: ${token.colorPrimary};
        }

        .path-select-name {
            flex: none;
            color: ${token.colorText};
            font-weight: 400;
            white-space: nowrap;
        }

        .path-select-row-action {
            width: ${props.compact ? 12 : 14}px;
            flex: none;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            color: ${token.colorTextQuaternary};
            font-size: ${props.compact ? 11 : 12}px;
        }

        .path-select-directory,
        .path-select-selectable {
            cursor: pointer;
        }

        .path-select-rejected {
            cursor: not-allowed;
            .path-select-name, .path-select-icon {
                color: ${token.colorTextDisabled};
            }
        }

        .path-select-directory:hover > td,
        .path-select-selectable:hover > td {
            background: ${token.colorFillQuaternary} !important;
        }

        .path-select-directory:focus-visible > td,
        .path-select-selectable:focus-visible > td {
            background: ${token.colorFillQuaternary} !important;
        }

        .path-select-directory:focus-visible > td:first-child,
        .path-select-selectable:focus-visible > td:first-child {
            box-shadow: inset 3px 0 ${token.colorPrimary};
        }

        .path-select-selected > td {
            background: ${token.colorPrimaryBg} !important;
        }

        .path-select-selected:hover > td,
        .path-select-selected:focus-visible > td {
            background: ${token.colorPrimaryBgHover} !important;
        }

        .path-select-selected .path-select-row-action {
            color: ${token.colorPrimary};
        }

        .path-select-error {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;
        }

        .path-select-status {
            box-sizing: border-box;
            min-height: ${props.compact ? 36 : 40}px;
            padding: ${props.compact ? "5px 9px" : "6px 11px"};
            display: flex;
            min-width: 0;
            align-items: center;
            justify-content: space-between;
            gap: ${props.compact ? 8 : 10}px;
            border-top: 1px solid ${token.colorSplit};
            background: ${token.colorFillQuaternary};
        }

        .path-select-status-main {
            min-width: 0;
            display: flex;
            flex: 1;
            align-items: center;
            gap: ${props.compact ? 6 : 8}px;
            overflow: hidden;
        }

        .path-select-status-main > .ant-checkbox-wrapper {
            flex: none;
            margin: 0;
            font-size: ${token.fontSizeSM}px;
            white-space: nowrap;
        }

        .path-select-status-icon {
            flex: none;
            color: ${token.colorTextTertiary};
        }

        .path-select-status-label {
            flex: none;
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;
        }

        .path-select-status-value {
            min-width: 0;
            flex: 1;
            display: block;
            overflow: hidden;
            color: ${token.colorTextSecondary};
            font-size: ${token.fontSizeSM}px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .path-select-status-meta {
            min-width: 0;
            display: flex;
            flex: none;
            align-items: center;
            flex-wrap: nowrap;
            gap: ${props.compact ? 7 : 9}px;
            white-space: nowrap;
        }

        .path-select-count {
            min-width: 0;
            overflow: hidden;
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .path-select-pagination {
            display: inline-flex;
            flex: none;
            align-items: center;
            margin: 0;
            color: ${token.colorTextSecondary};
            font-size: ${token.fontSizeSM}px;
        }

        .path-select-pagination .ant-pagination-prev,
        .path-select-pagination .ant-pagination-next {
            min-width: ${props.compact ? 24 : 26}px;
            height: ${props.compact ? 24 : 26}px;
            line-height: ${props.compact ? 24 : 26}px;
        }

        .path-select-pagination .ant-pagination-item-link {
            border-radius: ${token.borderRadius}px;
            color: ${token.colorTextSecondary};
        }

        .path-select-pagination .ant-pagination-prev:not(.ant-pagination-disabled):hover .ant-pagination-item-link,
        .path-select-pagination .ant-pagination-next:not(.ant-pagination-disabled):hover .ant-pagination-item-link {
            background: ${token.colorFillTertiary};
        }

        .path-select-pagination .ant-pagination-simple-pager {
            height: ${props.compact ? 24 : 26}px;
            margin-inline: ${props.compact ? 5 : 7}px;
            color: ${token.colorTextTertiary};
            line-height: ${props.compact ? 24 : 26}px;
        }

        .path-select-page-button.ant-btn {
            width: ${props.compact ? 24 : 26}px;
            min-width: ${props.compact ? 24 : 26}px;
            height: ${props.compact ? 24 : 26}px;
            padding: 0;
            border: 0;
            border-radius: ${token.borderRadius}px;
            color: ${token.colorTextSecondary};
            font-size: 12px;
        }

        .path-select-page-button.ant-btn:not(:disabled):hover {
            background: ${token.colorFillTertiary};
            color: ${token.colorTextSecondary};
        }

        .path-select-page-button.ant-btn:disabled {
            border: 0;
            background: transparent;
            color: ${token.colorTextDisabled};
        }

        .path-select-page-number {
            height: ${props.compact ? 24 : 26}px;
            margin-inline: ${props.compact ? 5 : 7}px;
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;
            line-height: ${props.compact ? 24 : 26}px;
        }

        @media (max-width: 520px) {
            .path-select-status {
                min-height: 0;
                padding: 6px 9px 7px;
                align-items: stretch;
                flex-direction: column;
                gap: 4px;
            }

            .path-select-status-main,
            .path-select-status-meta {
                width: 100%;
                min-width: 0;
            }

            .path-select-status-meta {
                justify-content: space-between;
            }
        }

        @media (pointer: coarse) {
            .path-select-up {
                width: 40px;
                min-width: 40px;
                height: 40px;
            }

            .path-select-path {
                height: 40px;
            }

            .path-select-path-button,
            .path-select-path-current {
                display: inline-flex;
                align-items: center;
            }

            .path-select-path-button,
            .path-select-path-current {
                min-height: 40px;
            }

            .path-select-path-input.ant-input {
                height: 40px;
            }

            .path-select-table .ant-table-tbody > tr > td {
                padding-block: 10px !important;
            }
        }
    `,
}));

export default useStyles;
