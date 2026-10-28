import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    state: css`
        min-height: 190px;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 12px;
        color: ${token.colorTextSecondary};
        font-size: 12px;
        line-height: 20px;

        .state-icon { color: ${token.colorTextQuaternary}; font-size: 28px; }
        .state-error { max-width: 100%; text-align: center; overflow-wrap: anywhere; }
    `,
    storage: css`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: ${token.marginSM}px;
        margin-bottom: ${token.marginXS}px;
        padding: ${token.padding}px;
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorFillAlter};
        color: ${token.colorTextSecondary};
        font-size: ${token.fontSizeSM}px;
        line-height: ${token.lineHeightSM};
        font-variant-numeric: tabular-nums;

        .storage-label { display: inline-flex; align-items: center; gap: ${token.marginXS}px; min-width: 0; }
        .storage-label .anticon { color: ${token.colorTextTertiary}; }
        .storage-value { color: ${token.colorText}; white-space: nowrap; }
    `,
    warning: css`
        margin-top: 12px;
        color: ${token.colorWarningText};
        font-size: 12px;
        line-height: 20px;
        overflow-wrap: anywhere;
    `,
    error: css`
        margin-top: 12px;
        color: ${token.colorError};
        font-size: 12px;
        line-height: 20px;
        overflow-wrap: anywhere;
    `,
}));
