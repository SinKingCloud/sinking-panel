import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}) => ({
    siteTable: css`
        .root-cell {
            width: 220px;
            max-width: 220px;
            min-width: 0;
        }

        .site-name,
        .site-name > button,
        .site-category,
        .site-type,
        .site-root,
        .update-time {
            min-width: 0;
        }

        .site-name > button {
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

        .site-name > button:hover,
        .site-name > button:focus-visible {
            color: ${token.colorPrimary};
        }

        .site-category,
        .site-type,
        .site-root,
        .update-time {
            display: block;
            overflow: hidden;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .site-root {
            width: 100%;
            color: ${token.colorTextSecondary};
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: ${token.fontSizeSM}px;
        }

        .site-state {
            display: inline-flex;
            align-items: center;
            gap: 5px;
            line-height: 20px;
            white-space: nowrap;
        }

        .site-state i {
            width: 6px;
            height: 6px;
            flex: none;
            border-radius: 50%;
            background: currentColor;
        }

        .site-state.running {
            color: ${token.colorSuccess};
        }

        .site-state.stopped {
            color: ${token.colorWarning};
        }
    `,
}));

export default useStyles;
