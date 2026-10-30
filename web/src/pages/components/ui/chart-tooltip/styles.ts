import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    tooltip: css`
        position: fixed;
        left: 0;
        top: 0;
        will-change: transform;
        z-index: ${token.zIndexPopupBase - 1};
        box-sizing: border-box;
        max-width: calc(100vw - 16px);
        padding: 10px 12px;
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorBgElevated};
        box-shadow: ${token.boxShadowSecondary};
        pointer-events: none;
        contain: layout style;
        color: ${token.colorText};
        font-family: ${token.fontFamily};
        font-size: 12px;
        line-height: 18px;
        font-variant-numeric: tabular-nums;

        .tooltip-title { color: ${token.colorTextSecondary}; margin-bottom: 6px; white-space: nowrap; }
        .tooltip-metric { display: flex; align-items: center; gap: 6px; line-height: 21px; }
        .tooltip-metric i { width: 6px; height: 6px; border-radius: 50%; flex: none; }
        .tooltip-metric strong { margin-inline-start: auto; font-weight: 500; }
        &[hidden], [hidden] { display: none !important; }
    `,
}));
