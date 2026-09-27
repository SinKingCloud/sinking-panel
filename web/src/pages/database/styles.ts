import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}: any) => ({
    page: css`
        width: 100%;
        min-width: 0;
        max-width: 1450px;
        margin: 0 auto;
    `,
    heading: css`
        padding: 4px 0 8px;
        h1 { margin: 0; color: ${token.colorTextHeading}; font-size: ${token.fontSizeHeading4}px; font-weight: 600; }
    `,
    loading: css`min-height: 300px; display: grid; place-items: center;`,
    install: css`max-width: 620px;`,
    serviceToolbar: css`
        display: flex;
        align-items: center;
        justify-content: space-between;
        flex-wrap: wrap;
        gap: 12px;
        margin-bottom: 20px;
        .ant-tag { margin-inline-end: 0; }
    `,
    connection: css`margin-bottom: 12px;`,
    configActions: css`display: flex; justify-content: flex-end; gap: 8px; margin-top: 16px;`,
    metrics: css`
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 12px;
        margin-bottom: 20px;
        > div { min-width: 0; padding: 14px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        span { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; display: block; }
        strong { display: block; margin-top: 6px; color: ${token.colorText}; font-size: ${token.fontSizeLG}px; font-weight: 500; overflow-wrap: anywhere; }
        @media (max-width: 680px) { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    `,
    content: css`
        margin-top: 0;
    `,
    panel: css`
        height: 100%;

        .ant-card-head {
            min-height: 48px;
        }

        .ant-card-body {
            min-width: 0;
        }
    `,
    status: css`
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 8px;
        margin-bottom: 14px;

        .status-item {
            min-width: 0;
            padding: 11px 12px;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadius}px;
            background: ${token.colorFillQuaternary};
        }

        .label {
            display: block;
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;
            line-height: 18px;
        }

        .value {
            display: block;
            margin-top: 3px;
            overflow: hidden;
            color: ${token.colorTextHeading};
            font-size: 16px;
            font-weight: 600;
            line-height: 22px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @media (max-width: 680px) {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    `,
    form: css`
        .ant-form-item {
            margin-bottom: 14px;
        }

        .form-actions {
            display: flex;
            align-items: center;
            justify-content: flex-end;
            gap: 8px;
            margin-top: 4px;
        }
    `,
    imageHint: css`
        font-size: ${token.fontSizeSM}px;
        line-height: 1.7;
    `,
    imageInput: css`
        display: flex;
        align-items: flex-start;
        gap: 8px;

        > .ant-input {
            min-width: 0;
            flex: 1;
        }

        > .ant-btn {
            flex: none;
        }
    `,
    imagePreview: css`
        margin-bottom: 14px;
        padding: 12px;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};
        font-size: ${token.fontSizeSM}px;

        dl {
            display: grid;
            grid-template-columns: auto minmax(0, 1fr);
            gap: 7px 12px;
            margin: 0;
        }

        dt {
            color: ${token.colorTextTertiary};
        }

        dd {
            min-width: 0;
            margin: 0;
            color: ${token.colorText};
            overflow-wrap: anywhere;
        }

        .code {
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
        }

        .ant-collapse {
            margin-top: 8px;
            font-size: inherit;
        }

        .ant-collapse > .ant-collapse-item > .ant-collapse-header {
            padding-inline: 0;
        }

        .ant-collapse-content > .ant-collapse-content-box {
            padding-inline: 0;
            padding-bottom: 0;
        }
    `,
    imageSummary: css`
        display: flex;
        align-items: baseline;
        gap: 12px;
        margin-bottom: 14px;
        font-size: ${token.fontSizeSM}px;
        color: ${token.colorTextSecondary};

        > :first-child {
            flex: none;
            color: ${token.colorTextTertiary};
        }

        > :last-child {
            min-width: 0;
            overflow-wrap: anywhere;
        }
    `,
    rule: css`
        display: flex;
        align-items: flex-start;
        gap: 8px;
        margin-bottom: 8px;

        .ant-form-item {
            min-width: 0;
            flex: 1;
            margin-bottom: 0;
        }

        .remove {
            flex: none;
            margin-top: 2px;
        }

        @media (max-width: 500px) {
            align-items: stretch;

            .ant-input-number {
                width: 100%;
            }
        }
    `,
    config: css`
        .ant-input {
            min-height: 340px;
            resize: vertical;
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: 12px;
            line-height: 1.6;
        }
    `,
    log: css`
        min-height: 340px;
        max-height: 520px;
        overflow: auto;
        padding: 14px;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};
        color: ${token.colorTextSecondary};
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
        font-size: 12px;
        line-height: 1.65;
        white-space: pre-wrap;
        overflow-wrap: anywhere;
    `,
    runtime: css`
        max-height: 300px;
        margin: 0;
        overflow: auto;
        padding: 10px;
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};
        color: ${token.colorTextSecondary};
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
        font-size: 11px;
        line-height: 1.55;
        white-space: pre-wrap;
        overflow-wrap: anywhere;
    `,
}));

export default useStyles;
