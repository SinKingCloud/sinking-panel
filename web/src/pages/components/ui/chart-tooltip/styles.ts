import {createStyles} from "antd-style";

export default createStyles(({css, token}, {hasProcess}: {hasProcess: boolean}) => ({
    tooltip: css`
        position: fixed;
        left: 0;
        top: 0;
        will-change: transform;
        z-index: ${token.zIndexPopupBase - 1};
        box-sizing: border-box;
        width: ${hasProcess ? 288 : 220}px;
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

        .tooltip-time { color: ${token.colorTextSecondary}; margin-bottom: 6px; white-space: nowrap; }
        .tooltip-metric { display: flex; align-items: center; gap: 6px; line-height: 21px; }
        .tooltip-metric i { width: 6px; height: 6px; border-radius: 50%; flex: none; }
        .tooltip-metric strong { margin-inline-start: auto; font-weight: 500; }
        .process-section { box-sizing: border-box; height: 172px; border-top: 1px solid ${token.colorBorderSecondary}; margin-top: 8px; padding-top: 8px; }
        .process-heading { display: flex; justify-content: space-between; gap: 8px; font-weight: 500; }
        .process-heading span:last-child { color: ${token.colorTextSecondary}; font-size: 11px; font-weight: 400; }
        .process-time { color: ${token.colorTextTertiary}; font-size: 11px; line-height: 16px; margin-top: 3px; white-space: nowrap; }
        .process-state { padding: 14px 0; color: ${token.colorTextSecondary}; }
        .process-list { display: grid; grid-template-columns: var(--process-pid-width) minmax(0, 1fr) var(--process-value-width); align-items: center; column-gap: 8px; margin: 6px 0 0; padding: 0; list-style: none; }
        .process-row { display: contents; line-height: 24px; }
        .process-pid { color: ${token.colorTextSecondary}; font-size: 11px; white-space: nowrap; }
        .process-name { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
        .process-value { text-align: right; font-weight: 500; white-space: nowrap; }
        &[hidden], [hidden] { display: none !important; }
    `,
}));
