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

        .ant-input {
            min-width: 0;
        }
    `,
    browser: css`
        min-width: 0;
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;
        background: ${token.colorBgContainer};

        .file-picker-toolbar {
            display: grid;
            min-width: 0;
            grid-template-columns: auto minmax(0, 1fr) auto;
            align-items: center;
            gap: ${props.compact ? 6 : 8}px;
            padding: ${props.compact ? 7 : 9}px;
            border-bottom: 1px solid ${token.colorSplit};
            background: ${token.colorFillQuaternary};
        }

        .file-picker-toolbar:not(.file-picker-toolbar-with-disks) {
            grid-template-columns: auto minmax(0, 1fr);
        }

        .file-picker-up {
            width: ${props.compact ? 28 : 30}px;
            min-width: ${props.compact ? 28 : 30}px;
            height: ${props.compact ? 28 : 30}px;
            padding: 0;
            border: 1px solid transparent;
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorBgContainer};
            color: ${token.colorTextSecondary};
        }

        .file-picker-up:not(:disabled):hover {
            border-color: ${token.colorBorderSecondary};
            background: ${token.colorFillTertiary};
            color: ${token.colorText};
        }

        .file-picker-path {
            box-sizing: border-box;
            height: ${props.compact ? 30 : 32}px;
            min-width: 0;
            margin: 0;
            padding: 0 ${props.compact ? 6 : 8}px;
            display: flex;
            align-items: center;
            overflow-x: auto;
            overflow-y: hidden;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorBgContainer};
            list-style: none;
            scrollbar-width: none;

            &::-webkit-scrollbar {
                display: none;
            }
        }

        .file-picker-path li {
            min-width: 0;
            display: inline-flex;
            flex: none;
            align-items: center;
        }

        .file-picker-path-separator {
            margin-inline: ${props.compact ? 2 : 3}px;
            color: ${token.colorTextQuaternary};
            font-size: ${props.compact ? 8 : 9}px;
        }

        .file-picker-path-button,
        .file-picker-path-current {
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

        .file-picker-path-button > span,
        .file-picker-path-current > span {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .file-picker-path-button:hover {
            background: ${token.colorFillTertiary};
            color: ${token.colorText};
        }

        .file-picker-path-button:disabled {
            color: ${token.colorTextDisabled};
            cursor: not-allowed;
        }

        .file-picker-path-button:focus-visible {
            box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
        }

        .file-picker-path-current {
            color: ${token.colorText};
            cursor: text;
            font-weight: 500;
        }

        .file-picker-path-current:not(:disabled):hover {
            background: ${token.colorFillTertiary};
            color: ${token.colorPrimary};
        }

        .file-picker-path-current:focus-visible {
            color: ${token.colorPrimary};
            box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
        }

        .file-picker-path-current:disabled {
            color: ${token.colorText};
            cursor: default;
        }

        .file-picker-path-input.ant-input {
            width: min(100%, ${props.compact ? 420 : 480}px);
            min-width: 40px;
            max-width: 100%;
            height: ${props.compact ? 30 : 32}px;
            padding: 0 ${props.compact ? 6 : 8}px;
            justify-self: start;
            border-color: ${token.colorPrimaryBorder};
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorBgContainer};
            color: ${token.colorText};
            box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            font-size: ${props.compact ? 11 : 12}px;
        }

        .file-picker-disks {
            width: ${props.compact ? 116 : 132}px;
        }

        .file-picker-table .ant-table,
        .file-picker-table .ant-table-container {
            background: transparent;
        }

        .file-picker-table .ant-table-thead > tr > th {
            padding: ${props.compact ? "7px 10px" : "9px 12px"} !important;
            border-bottom: 1px solid ${token.colorSplit};
            background: ${token.colorBgContainer};
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;
            font-weight: 500;
        }

        .file-picker-table .ant-table-tbody > tr > td {
            padding: ${props.compact ? "7px 10px" : "9px 12px"} !important;
            border-bottom: 1px solid ${token.colorSplit};
            background: ${token.colorBgContainer};
            font-size: ${token.fontSizeSM}px;
            transition: background-color ${token.motionDurationFast};
        }

        .file-picker-table .ant-table-tbody > tr.ant-table-measure-row > td {
            padding-block: 0 !important;
            border-block: 0 !important;
        }

        .file-picker-table .ant-table-tbody > tr:last-child > td {
            border-bottom: 0;
        }

        .file-picker-table .ant-table-body {
            min-height: ${props.compact
                ? "clamp(180px, 34dvh, 238px)"
                : "clamp(200px, 38dvh, 278px)"};
            scrollbar-width: thin;
        }

        .file-picker-table .ant-table-placeholder td {
            height: ${props.compact
                ? "clamp(180px, 34dvh, 238px)"
                : "clamp(200px, 38dvh, 278px)"};
            border-bottom: 0;
        }

        .file-picker-name-cell {
            display: flex;
            width: max-content;
            align-items: center;
            gap: ${props.compact ? 7 : 9}px;
        }

        .file-picker-icon {
            flex: none;
            color: ${token.colorTextSecondary};
            font-size: ${props.compact ? 14 : 16}px;
        }

        .file-picker-directory .file-picker-icon {
            color: ${token.colorPrimary};
        }

        .file-picker-name {
            flex: none;
            color: ${token.colorText};
            white-space: nowrap;
        }

        .file-picker-row-action {
            width: ${props.compact ? 12 : 14}px;
            flex: none;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            color: ${token.colorTextQuaternary};
            font-size: ${props.compact ? 11 : 12}px;
        }

        .file-picker-directory,
        .file-picker-selectable {
            cursor: pointer;
        }

        .file-picker-directory:hover > td,
        .file-picker-selectable:hover > td {
            background: ${token.colorFillQuaternary} !important;
        }

        .file-picker-directory:focus-visible > td,
        .file-picker-selectable:focus-visible > td {
            background: ${token.colorFillQuaternary} !important;
        }

        .file-picker-directory:focus-visible > td:first-child,
        .file-picker-selectable:focus-visible > td:first-child {
            box-shadow: inset 3px 0 ${token.colorPrimary};
        }

        .file-picker-selected > td {
            background: ${token.colorPrimaryBg} !important;
        }

        .file-picker-selected:hover > td,
        .file-picker-selected:focus-visible > td {
            background: ${token.colorPrimaryBgHover} !important;
        }

        .file-picker-selected .file-picker-row-action {
            color: ${token.colorPrimary};
        }

        .file-picker-error {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;
        }

        .file-picker-status {
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

        .file-picker-status-main {
            min-width: 0;
            display: flex;
            flex: 1;
            align-items: center;
            gap: ${props.compact ? 6 : 8}px;
            overflow: hidden;
        }

        .file-picker-status-icon {
            flex: none;
            color: ${token.colorTextTertiary};
        }

        .file-picker-status-label {
            flex: none;
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;
        }

        .file-picker-status-value {
            min-width: 0;
            flex: 1;
            display: block;
            overflow: hidden;
            color: ${token.colorTextSecondary};
            font-size: ${token.fontSizeSM}px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .file-picker-status-meta {
            min-width: 0;
            display: flex;
            flex: none;
            align-items: center;
            flex-wrap: nowrap;
            gap: ${props.compact ? 7 : 9}px;
            white-space: nowrap;
        }

        .file-picker-count {
            min-width: 0;
            overflow: hidden;
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .file-picker-pagination {
            flex: none;
            margin: 0;
            color: ${token.colorTextSecondary};
            font-size: ${token.fontSizeSM}px;
        }

        .file-picker-pagination .ant-pagination-prev,
        .file-picker-pagination .ant-pagination-next {
            min-width: ${props.compact ? 24 : 26}px;
            height: ${props.compact ? 24 : 26}px;
            line-height: ${props.compact ? 24 : 26}px;
        }

        .file-picker-pagination .ant-pagination-item-link {
            border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary};
        }

        .file-picker-pagination .ant-pagination-prev:not(.ant-pagination-disabled):hover .ant-pagination-item-link,
        .file-picker-pagination .ant-pagination-next:not(.ant-pagination-disabled):hover .ant-pagination-item-link {
            background: ${token.colorFillTertiary};
        }

        .file-picker-pagination .ant-pagination-simple-pager {
            height: ${props.compact ? 24 : 26}px;
            margin-inline: ${props.compact ? 5 : 7}px;
            color: ${token.colorTextTertiary};
            line-height: ${props.compact ? 24 : 26}px;
        }

        @media (max-width: 520px) {
            .file-picker-disks {
                width: 104px;
            }

            .file-picker-status {
                min-height: 0;
                padding: 6px 9px 7px;
                align-items: stretch;
                flex-direction: column;
                gap: 4px;
            }

            .file-picker-status-main,
            .file-picker-status-meta {
                width: 100%;
                min-width: 0;
            }

            .file-picker-status-meta {
                justify-content: space-between;
            }
        }

        @media (pointer: coarse) {
            .file-picker-up {
                width: 40px;
                min-width: 40px;
                height: 40px;
            }

            .file-picker-path {
                height: 40px;
            }

            .file-picker-path-button,
            .file-picker-path-current {
                display: inline-flex;
                align-items: center;
            }

            .file-picker-path-button,
            .file-picker-path-current {
                min-height: 40px;
            }

            .file-picker-path-input.ant-input {
                height: 40px;
            }

            .file-picker-disks,
            .file-picker-disks .ant-select-selector {
                height: 40px !important;
            }

            .file-picker-disks .ant-select-selector {
                align-items: center;
            }

            .file-picker-table .ant-table-tbody > tr > td {
                padding-block: 10px !important;
            }
        }
    `,
}));

export default useStyles;
