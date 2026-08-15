import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean}>(({css, token}, props = {}) => {
    const compact = Boolean(props.compact);
    return {
        root: css`
            .ant-drawer-body {
                padding: ${compact ? 10 : 14}px;
            }

            .file-task-list {
                display: flex;
                flex-direction: column;
                gap: ${compact ? 5 : 7}px;
            }

            .file-task-item {
                min-width: 0;
                padding: ${compact ? "8px 9px" : "10px 11px"};
                border-radius: ${token.borderRadius}px;
                background: ${token.colorFillQuaternary};
            }

            .file-task-head {
                min-width: 0;
                display: flex;
                align-items: center;
                justify-content: space-between;
                gap: 8px;
            }

            .file-task-title {
                min-width: 0;
                overflow: hidden;
                color: ${token.colorText};
                font-size: ${compact ? 11 : 12}px;
                font-weight: 500;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .file-task-state {
                display: inline-flex;
                align-items: center;
                gap: 5px;
                flex: none;
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 10 : 11}px;
            }

            .file-task-state::before {
                width: 6px;
                height: 6px;
                border-radius: 50%;
                background: currentColor;
                content: "";
            }

            .file-task-state.running {
                color: ${token.colorPrimary};
            }

            .file-task-state.completed {
                color: ${token.colorSuccess};
            }

            .file-task-state.failed {
                color: ${token.colorError};
            }

            .file-task-state.canceled {
                color: ${token.colorTextQuaternary};
            }

            .file-task-message {
                min-width: 0;
                margin-top: 5px;
                overflow: hidden;
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 10 : 11}px;
                line-height: 18px;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .file-task-progress {
                margin-top: 5px;
                margin-bottom: 0;
            }

            .file-task-actions {
                margin-top: 4px;
                display: flex;
                justify-content: flex-end;
            }

            .file-task-actions .ant-btn {
                height: ${compact ? 24 : 28}px;
                padding-inline: 7px;
                font-size: ${compact ? 10 : 11}px;
            }
        `,
    };
});

export default useStyles;
