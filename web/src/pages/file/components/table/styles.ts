import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}: any, props: {compact?: boolean} = {}) => {
    const compact = Boolean(props.compact);
    return {
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

            &:disabled,
            &:disabled .file-name,
            &:disabled:hover .file-name {
                color: ${token.colorTextDisabled};
                cursor: not-allowed;
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
        fileMetaButton: css`
            max-width: 100%;
            padding: 0;
            display: inline-flex;
            align-items: center;
            border: 0;
            outline: none;
            background: transparent;
            color: inherit;
            cursor: pointer;
            font: inherit;

            &:hover .file-meta,
            &:focus-visible .file-meta {
                color: ${token.colorPrimary};
            }

            &:focus-visible {
                border-radius: ${token.borderRadiusSM}px;
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            &:disabled,
            &:disabled .file-meta,
            &:disabled:hover .file-meta {
                color: ${token.colorTextDisabled};
                cursor: not-allowed;
            }
        `,
        directorySize: css`
            min-width: 34px;
            min-height: 18px;
            display: inline-flex;
            align-items: center;
            border: 0;
            outline: none;
            background: transparent;
            color: inherit;
            font: inherit;
            line-height: 18px;
            vertical-align: middle;

            &.calculate {
                padding: 0;
                color: ${token.colorTextSecondary};
                cursor: pointer;
                text-decoration-line: underline;
                text-decoration-style: dashed;
                text-decoration-color: ${token.colorBorder};
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
