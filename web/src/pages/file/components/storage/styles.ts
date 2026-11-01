import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    table: css`
        .storage-name {
            display: block;
            max-width: 100%;
            padding: 0;
            overflow: hidden;
            border: 0;
            background: transparent;
            color: inherit;
            font: inherit;
            line-height: 20px;
            text-align: left;
            text-overflow: ellipsis;
            white-space: nowrap;
            cursor: pointer;
        }

        .storage-name:hover,
        .storage-name:focus-visible {
            color: ${token.colorPrimary};
        }

        .storage-name:disabled {
            color: ${token.colorTextDisabled};
            cursor: not-allowed;
        }
    `,
}));
