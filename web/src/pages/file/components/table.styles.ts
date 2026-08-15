import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean; dark?: boolean}>(({
    css,
    token,
    isDarkMode,
}, props = {}) => {
    const compact = Boolean(props?.compact);
    const dark = typeof props?.dark === "boolean" ? props.dark : Boolean(isDarkMode);
    const tableHighlight = dark
        ? `color-mix(in srgb, #fff 5%, ${token.colorBgContainer})`
        : `color-mix(in srgb, #000 3%, ${token.colorBgContainer})`;

    return {
        fileTable: css`
            --file-table-text: ${dark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)"};
            min-width: 0;
            overflow: hidden;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusLG}px;
            background: ${token.colorBgContainer};

            .ant-table,
            .ant-table-container {
                background: transparent;
            }

            .ant-table-container {
                overflow: hidden;
                border-radius: inherit;
            }

            .ant-table-thead > tr > th {
                height: auto;
                padding: ${compact ? 8 : 10}px !important;
                border-bottom: 1px solid ${token.colorBorderSecondary};
                background: ${tableHighlight};
                color: var(--file-table-text);
                font-size: ${token.fontSize}px;
            }

            .ant-table-thead > tr > th.ant-table-column-sort,
            .ant-table-thead > tr > th.ant-table-column-has-sorters:hover {
                background: ${tableHighlight};
            }

            .ant-table-tbody > tr > td.ant-table-column-sort {
                background: ${token.colorBgContainer};
            }

            .ant-table-column-sorters {
                justify-content: flex-start;
                gap: 5px;
            }

            .ant-table-column-sorter {
                margin-inline-start: 0;
                color: ${token.colorTextQuaternary};
            }

            .ant-table-column-sorter-up.active,
            .ant-table-column-sorter-down.active {
                color: ${token.colorPrimary};
            }

            .ant-table-thead > tr > th.action-cell {
                background: ${tableHighlight} !important;
                color: var(--file-table-text) !important;
            }

            .ant-table-thead > tr > .ant-table-cell-fix-end,
            .ant-table-thead > tr > .ant-table-cell-fix-right,
            .ant-table-thead > tr > th[class*="ant-table-cell-fix-end"],
            .ant-table-thead > tr > th[class*="ant-table-cell-fix-right"] {
                background: ${tableHighlight} !important;
                color: var(--file-table-text) !important;
            }

            .ant-table-tbody > tr > .ant-table-cell-fix-end,
            .ant-table-tbody > tr > .ant-table-cell-fix-right,
            .ant-table-tbody > tr > td[class*="ant-table-cell-fix-end"],
            .ant-table-tbody > tr > td[class*="ant-table-cell-fix-right"] {
                background: ${token.colorBgContainer} !important;
            }

            .ant-table-tbody > tr > td.action-cell {
                background: ${token.colorBgContainer} !important;
            }

            .ant-table-tbody > tr.ant-table-row > td {
                height: auto;
                padding: ${compact ? 8 : 10}px !important;
                border-bottom: 1px solid ${token.colorSplit};
                background: ${token.colorBgContainer};
                color: var(--file-table-text);
                font-size: ${token.fontSizeSM + 1}px;
                transition: background-color .16s ease;
            }

            .ant-table-tbody > tr:last-child > td {
                border-bottom: 0;
            }

            .ant-table-placeholder > td {
                height: 112px;
            }

            .ant-table-tbody > tr.ant-table-row:hover > td {
                background: ${tableHighlight} !important;
            }

            .ant-table-tbody > tr.ant-table-row:hover > .ant-table-cell-fix-end,
            .ant-table-tbody > tr.ant-table-row:hover > .ant-table-cell-fix-right,
            .ant-table-tbody > tr.ant-table-row:hover > td[class*="ant-table-cell-fix-end"],
            .ant-table-tbody > tr.ant-table-row:hover > td[class*="ant-table-cell-fix-right"] {
                background: ${tableHighlight} !important;
            }

            .ant-table-tbody > tr.ant-table-row:hover > td.action-cell {
                background: ${tableHighlight} !important;
            }

            .ant-table-tbody > tr.ant-table-row-selected > td,
            .ant-table-tbody > tr.ant-table-row-selected > .ant-table-cell-fix-end,
            .ant-table-tbody > tr.ant-table-row-selected > .ant-table-cell-fix-right,
            .ant-table-tbody > tr.ant-table-row-selected > td[class*="ant-table-cell-fix-end"],
            .ant-table-tbody > tr.ant-table-row-selected > td[class*="ant-table-cell-fix-right"],
            .ant-table-tbody > tr.ant-table-row-selected > td.action-cell {
                background: ${token.colorPrimaryBg} !important;
            }

            .ant-table-tbody > tr.ant-table-row-selected:hover > td,
            .ant-table-tbody > tr.ant-table-row-selected:hover > .ant-table-cell-fix-end,
            .ant-table-tbody > tr.ant-table-row-selected:hover > .ant-table-cell-fix-right,
            .ant-table-tbody > tr.ant-table-row-selected:hover > td[class*="ant-table-cell-fix-end"],
            .ant-table-tbody > tr.ant-table-row-selected:hover > td[class*="ant-table-cell-fix-right"],
            .ant-table-tbody > tr.ant-table-row-selected:hover > td.action-cell {
                background: ${token.colorPrimaryBgHover} !important;
            }

            .ant-table-selection-column {
                text-align: center !important;
            }

            .ant-table-tbody > tr.ant-table-row.action-menu-open > td {
                background: ${tableHighlight} !important;
            }

            .ant-table-content {
                overscroll-behavior-x: contain;
                scrollbar-width: thin;
            }

            .action-cell .ant-dropdown-trigger {
                display: inline-flex;
            }

            .action-cell {
                text-align: center !important;
            }

            .action-cell .ant-btn {
                height: 25px;
                padding: 10px;
                border-radius: ${token.borderRadius}px;
                font-size: ${token.fontSizeSM}px;
            }

        `,
        fileNameButton: css`
            width: 100%;
            max-width: 100%;
            padding: 0;
            border: 0;
            outline: none;
            background: transparent;
            color: ${token.colorText};
            cursor: pointer;
            font: inherit;
            text-align: left;

            &:hover .file-name,
            &:focus-visible .file-name {
                color: ${token.colorPrimary};
            }

            &:focus-visible {
                border-radius: ${token.borderRadiusSM}px;
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }
        `,
        fileNameContent: css`
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 8px;
        `,
        fileIcon: css`
            width: 20px;
            height: 20px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            flex: none;
            color: ${token.colorTextTertiary};
            font-size: 14px;

            &.folder {
                color: ${token.colorPrimary};
            }
        `,
        fileName: css`
            min-width: 0;
            overflow: hidden;
            color: ${token.colorText};
            font-size: ${token.fontSizeSM + 1}px;
            font-weight: 400;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
            transition: color .16s ease;
        `,
        fileMeta: css`
            color: inherit;
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: ${token.fontSizeSM}px;
            white-space: nowrap;
        `,
        directorySize: css`
            min-width: 34px;
            min-height: 18px;
            display: inline-flex;
            align-items: center;
            color: inherit;
            border: 0;
            outline: none;
            background: transparent;
            font: inherit;
            line-height: 18px;
            vertical-align: middle;

            &.calculate {
                padding: 0;
                color: ${token.colorTextSecondary};
                cursor: pointer;
                pointer-events: auto;
                text-decoration-line: underline;
                text-decoration-style: dashed;
                text-decoration-color: ${token.colorBorder};
                text-decoration-thickness: 1px;
                text-underline-offset: 2px;
            }

            &.calculate:hover {
                color: ${token.colorText};
                text-decoration-color: currentColor;
            }

            &.calculate:focus-visible {
                border-radius: 2px;
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            &.error {
                color: ${token.colorTextSecondary};
                text-decoration-color: ${token.colorErrorBorder};
            }

            &.calculate.error:hover {
                color: ${token.colorText};
                text-decoration-color: ${token.colorError};
            }

            .ant-spin {
                line-height: 1;
            }

            .ant-spin-dot {
                font-size: 12px;
            }
        `,
        fileMenu: css`
            && .ant-dropdown-menu {
                min-width: 132px;
                max-height: min(420px, calc(100dvh - 24px));
                padding: 3px !important;
                overflow-y: auto;
                border-radius: ${token.borderRadius}px !important;
            }

            && .ant-dropdown-menu-item {
                min-height: ${compact ? 26 : 30}px;
                padding: 0 8px !important;
                border-radius: ${token.borderRadiusSM}px !important;
                font-size: ${compact ? 11 : 12}px;
            }
        `,
    };
});

export default useStyles;
