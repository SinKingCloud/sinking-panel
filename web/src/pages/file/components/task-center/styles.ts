import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean}>(({css, token}, props = {}) => {
    const compact = Boolean(props.compact);
    return {
        root: css`
            position: fixed;
            z-index: ${token.zIndexPopupBase + 40};
            left: 50%;
            top: 50%;
            width: min(440px, calc(100vw - 32px));
            max-height: min(520px, calc(100dvh - 32px));
            overflow: hidden;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusLG + 2}px;
            background: ${token.colorBgContainer};
            box-shadow: ${token.boxShadowSecondary};
            color: ${token.colorText};
            pointer-events: auto;
            transform: translate(-50%, -50%);
            animation: file-task-progress-in ${token.motionDurationMid};

            @keyframes file-task-progress-in {
                from {
                    opacity: 0;
                    transform: translate(-50%, calc(-50% + 8px));
                }
                to {
                    opacity: 1;
                    transform: translate(-50%, -50%);
                }
            }

            .file-task-progress-panel-header {
                padding: ${compact ? 9 : 11}px ${compact ? 12 : 14}px;
                display: flex;
                align-items: center;
                justify-content: space-between;
                border-bottom: 1px solid ${token.colorSplit};
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
                font-weight: 500;
                line-height: 20px;
                cursor: grab;
                user-select: none;
                touch-action: none;
            }

            .file-task-progress-panel-header.is-dragging {
                cursor: grabbing;
            }

            .file-task-progress-panel-title {
                min-width: 0;
                flex: 1;
            }

            .file-task-progress-panel-title > div {
                min-width: 0;
            }

            .file-task-progress-panel-close.ant-btn {
                flex: none;
                width: 24px;
                min-width: 24px;
                height: 24px;
                padding: 0;
                color: ${token.colorTextQuaternary};
            }

            .file-task-progress-panel-close.ant-btn:hover {
                color: ${token.colorText};
                background: ${token.colorFillTertiary};
            }

            .file-task-progress-list {
                max-height: min(468px, calc(100dvh - 84px));
                padding: ${compact ? 5 : 7}px;
                overflow-y: auto;
            }

            .file-task-progress-item {
                min-width: 0;
                padding: ${compact ? 8 : 10}px ${compact ? 7 : 9}px;
                border-radius: ${token.borderRadius}px;
                background: ${token.colorFillQuaternary};
            }

            .file-task-progress-item + .file-task-progress-item {
                margin-top: ${compact ? 4 : 5}px;
            }

            .file-task-progress-header {
                min-width: 0;
                display: flex;
                align-items: center;
                gap: 7px;
            }

            .file-task-progress-status-dot {
                flex: none;
                width: 7px;
                height: 7px;
                border-radius: 50%;
                background: ${token.colorPrimary};
            }

            &.file-task-progress-completed .file-task-progress-status-dot {
                background: ${token.colorSuccess};
            }

            &.file-task-progress-failed .file-task-progress-status-dot {
                background: ${token.colorError};
            }

            &.file-task-progress-canceled .file-task-progress-status-dot {
                background: ${token.colorTextQuaternary};
            }

            .file-task-progress-name {
                min-width: 0;
                flex: 1;
                overflow: hidden;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
                font-weight: 500;
                line-height: 22px;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .file-task-progress-value {
                flex: none;
                color: ${token.colorPrimary};
                font-size: ${compact ? 11 : 12}px;
                font-variant-numeric: tabular-nums;
                font-weight: 500;
                line-height: 22px;
            }

            .file-task-progress-state {
                margin-bottom: 5px;
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 10 : 11}px;
                line-height: 16px;
            }

            .file-task-progress-state.completed {
                color: ${token.colorSuccess};
            }

            .file-task-progress-state.failed {
                color: ${token.colorError};
            }

            .file-task-progress-state.canceled {
                color: ${token.colorTextQuaternary};
            }

            .file-task-progress-bar {
                width: 100%;
                margin: 4px 0 0;
            }

            .file-task-progress-message {
                max-height: ${compact ? 60 : 72}px;
                margin-top: 8px;
                padding: 6px 8px;
                overflow: hidden;
                border-radius: ${token.borderRadiusSM}px;
                background: ${token.colorFillQuaternary};
                color: ${token.colorTextTertiary};
                font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
                font-size: ${compact ? 10 : 11}px;
                line-height: 16px;
                white-space: normal;
                overflow-wrap: anywhere;
                word-break: break-word;
                display: -webkit-box;
                -webkit-box-orient: vertical;
                -webkit-line-clamp: 4;
            }

            .file-task-progress-cancel.ant-btn {
                flex: none;
                width: 28px;
                min-width: 28px;
                height: 26px;
                padding: 0;
                color: ${token.colorTextQuaternary};
                border-radius: ${token.borderRadiusSM}px;
            }

            .file-task-progress-cancel.ant-btn:hover {
                color: ${token.colorError};
                background: ${token.colorErrorBg};
            }

            @media (max-width: 520px) {
                left: 50%;
                top: 50%;
                width: min(440px, calc(100vw - 24px));
            }
        `,
    };
});

export default useStyles;
