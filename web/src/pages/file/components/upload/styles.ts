import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean}>(({css, token}, props = {}) => {
    const compact = Boolean(props?.compact);

    return {
        fileUploadModal: css`
            .file-upload-picker:not([hidden]) {
                height: 100%;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .file-upload-picker > .ant-upload-wrapper {
                width: 100%;
                height: 100%;
            }

            .ant-upload-wrapper .ant-upload-drag {
                height: 100%;
                min-height: 0;
                padding: ${compact ? "24px 14px" : "30px 18px"};
                border-color: ${token.colorBorderSecondary};
                border-radius: ${token.borderRadiusLG}px;
                background: ${token.colorBgContainer};
            }

            .ant-upload-wrapper .ant-upload-drag:hover {
                border-color: ${token.colorPrimaryBorder};
            }

            .ant-upload-drag-icon {
                width: ${compact ? 42 : 48}px;
                height: ${compact ? 42 : 48}px;
                margin: 0 auto 12px !important;
                display: flex;
                align-items: center;
                justify-content: center;
                position: relative;
                inset-inline-start: 0;
                color: ${token.colorPrimary};
                font-size: ${compact ? 32 : 38}px;
                line-height: 1;
                transform: translate3d(0, 0, 0);
                animation: none !important;
                transition: none !important;

                .anticon {
                    width: 1em;
                    height: 1em;
                    display: inline-flex;
                    align-items: center;
                    justify-content: center;
                    line-height: 1;
                    vertical-align: middle;
                    transform: none;
                    animation: none !important;
                    transition: none !important;
                }

                .anticon > svg {
                    display: block;
                }
            }

            .ant-upload-text {
                color: ${token.colorText};
                font-size: ${compact ? 13 : 15}px;
                font-weight: 500;
            }

            .ant-upload-hint {
                margin-top: 6px;
                color: ${token.colorTextTertiary} !important;
                font-size: ${compact ? 10 : 11}px !important;
                white-space: normal;
            }

            .file-upload-list {
                min-height: 0;
                max-height: min(420px, calc(100dvh - 300px));
                flex: none;
                padding-right: 2px;
                display: flex;
                flex-direction: column;
                gap: ${compact ? 5 : 7}px;
                overflow-y: auto;
                scrollbar-width: thin;
            }

            .file-upload-tasks {
                min-height: 0;
                display: flex;
                flex-direction: column;
            }

            .file-upload-footer {
                display: flex;
                align-items: center;
                justify-content: flex-end;
                gap: 8px;
            }

            .file-upload-footer-actions {
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .file-upload-item {
                min-width: 0;
                padding: ${compact ? "8px 9px" : "10px 11px"};
                display: grid;
                grid-template-columns: ${compact ? 22 : 24}px minmax(0, 1fr);
                align-items: start;
                gap: ${compact ? 7 : 9}px;
                border-radius: ${token.borderRadius}px;
                background: ${token.colorFillQuaternary};
                transition: background-color ${token.motionDurationFast};
            }

            .file-upload-item:hover {
                background: ${token.colorFillTertiary};
            }

            .upload-file-icon {
                width: ${compact ? 22 : 24}px;
                height: ${compact ? 22 : 24}px;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 15 : 17}px;
            }

            .upload-file-content {
                min-width: 0;
                flex: 1;
            }

            .upload-file-head {
                min-width: 0;
                display: flex;
                align-items: center;
                justify-content: space-between;
                gap: 8px;
            }

            .upload-file-name {
                min-width: 0;
                flex: 1;
                overflow: hidden;
                color: ${token.colorText};
                font-size: ${compact ? 11 : 12}px;
                font-weight: 500;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .upload-file-control {
                width: 96px;
                flex: none;
                display: flex;
                align-items: center;
                justify-content: flex-end;
                gap: 1px;
            }

            .upload-file-state {
                flex: none;
                display: inline-flex;
                align-items: center;
                gap: 5px;
                color: ${token.colorTextTertiary};
                font-variant-numeric: tabular-nums;
            }

            .upload-file-state::before {
                width: 6px;
                height: 6px;
                flex: none;
                border-radius: 50%;
                background: currentColor;
                content: "";
            }

            .upload-file-state.waiting,
            .upload-file-state.canceled {
                color: ${token.colorTextTertiary};
            }

            .upload-file-state.checking,
            .upload-file-state.uploading,
            .upload-file-state.merging {
                color: ${token.colorPrimary};
            }

            .upload-file-state.done {
                color: ${token.colorSuccess};
            }

            .upload-file-state.error {
                color: ${token.colorError};
            }

            .upload-file-control .ant-btn {
                height: ${compact ? 24 : 26}px;
                padding-inline: 6px;
                font-size: ${compact ? 10 : 11}px;
            }

            .upload-file-meta {
                margin-top: 1px;
                min-width: 0;
                display: flex;
                align-items: center;
                gap: 8px;
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 10 : 11}px;
            }

            .file-upload-item .ant-progress {
                margin-top: ${compact ? 4 : 5}px;
                margin-bottom: 0;
            }

            .upload-file-error {
                min-width: 0;
                flex: 1;
                overflow: hidden;
                color: ${token.colorError};
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            @media (max-width: 575px) {
                .file-upload-item {
                    padding: 9px;
                    grid-template-columns: 22px minmax(0, 1fr);
                    gap: 7px;
                }

                .upload-file-control .ant-btn {
                    padding-inline: 5px;
                }

                .upload-file-control {
                    width: 90px;
                }
            }

        `,
    };
});

export default useStyles;
