import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}) => ({
    taskTable: css`
        .task-name {
            min-width: 0;
        }

        .task-name > button {
            display: block;
            max-width: 100%;
            padding: 0;
            overflow: hidden;
            border: 0;
            outline: none;
            background: transparent;
            color: inherit;
            cursor: pointer;
            font: inherit;
            line-height: 20px;
            text-align: left;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .task-name > button:hover,
        .task-name > button:focus-visible {
            color: ${token.colorPrimary};
        }

        .task-state {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            line-height: 20px;
        }

        .task-state i {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: currentColor;
        }

        .task-state.running {
            color: ${token.colorSuccess};
        }

        .task-state.paused {
            color: ${token.colorWarning};
        }

        .task-category,
        .task-exec-type,
        .schedule-value,
        .runtime-value {
            display: block;
            min-width: 0;
            overflow: hidden;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
    `,
}));

export default useStyles;
