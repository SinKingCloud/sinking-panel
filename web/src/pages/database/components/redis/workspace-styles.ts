import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    icon: css`
        width: ${token.fontSize + 2}px; height: ${token.fontSize + 2}px; flex: none; display: inline-block; background: #d82c20;
        mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath d='m12 2 11 5-11 5L1 7l11-5Zm-9 8 9 4 9-4 2 1-11 5L1 11l2-1Zm0 4 9 4 9-4 2 1-11 5L1 15l2-1Zm0 4 9 4 9-4 2 1-11 5L1 19l2-1Z'/%3E%3C/svg%3E") center / contain no-repeat;
    `,
    catalogBody: css`
        && { padding: 0; min-height: 0; overflow: hidden; display: flex; flex-direction: column; }
        .redis-mobile-database { display: flex; align-items: center; gap: 8px; padding: 14px; border-bottom: 1px solid ${token.colorBorderSecondary}; flex: none; }
        .redis-mobile-database .ant-select { flex: 1; min-width: 0; height: ${compact ? 28 : 30}px; color: ${token.colorTextSecondary}; font-size: ${compact ? 12 : 13}px; }
        .redis-mobile-database .ant-btn { flex: none; width: ${compact ? 28 : 30}px; height: ${compact ? 28 : 30}px; padding: 0; color: ${token.colorTextSecondary}; }
        .redis-mobile-database .ant-btn .anticon { font-size: ${compact ? 13 : 14}px; }
        .redis-mobile-key-host { display: flex; flex: 1; min-width: 0; min-height: 0; }
        .redis-mobile-key-host .redis-catalog { border: 0; border-radius: 0; }
        .redis-mobile-key-host > div { flex: 1; }
        && .redis-instance-footer { display: flex; flex: none; flex-direction: column; gap: 3px; padding: 12px 14px; border-top: 1px solid ${token.colorBorderSecondary}; }
        .redis-instance-name { font-size: ${compact ? 12 : 13}px; color: ${token.colorText}; }
        .redis-instance-version { font-size: ${compact ? 11 : 12}px; color: ${token.colorTextTertiary}; }
    `,
    modal: css`
        .ant-modal-container { padding: 0; overflow: hidden; }
        .ant-modal-header { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
        && .ant-modal-body { padding: 0; min-height: 0; max-height: none; overflow: hidden; }
        &.is-fullscreen { margin: 0; }
        &.is-fullscreen .ant-modal-container { border-radius: 0; }
        &.is-fullscreen > .ant-modal-container > .ant-modal-body > div { height: 100dvh; min-height: 0; }
    `,
    workspace: css`
        height: ${compact ? 680 : 700}px; min-height: ${compact ? 680 : 700}px; min-width: 0; display: flex; flex-direction: column;
        background: ${token.colorBgContainer}; color: ${token.colorText}; font-size: ${token.fontSize}px;
        .redis-window-header { display: flex; align-items: center; flex: none; min-height: ${token.controlHeight + 20}px; gap: 16px; padding: 0 16px; border-bottom: 1px solid ${token.colorBorderSecondary}; }
        .redis-catalog-toggle.ant-btn { flex: none; width: 32px; height: 32px; padding: 0; color: ${token.colorTextSecondary}; }
        .redis-window-title { display: flex; align-items: center; gap: 9px; min-width: 0; font-weight: 500; white-space: nowrap; }
        .redis-window-actions { display: flex; align-items: center; gap: 2px; margin-left: auto; }
        .redis-window-actions .ant-btn { color: ${token.colorTextSecondary}; }
        .redis-window-body { display: flex; flex: 1; min-width: 0; min-height: 0; position: relative; overflow: hidden; }
        .redis-window-main { display: flex; flex-direction: column; flex: 1; min-width: 0; min-height: 0; padding: 16px 18px; overflow: hidden; }
        .redis-database-catalog { display: flex; flex-direction: column; width: ${compact ? 196 : 208}px; min-height: 0; min-width: 0; flex: none; padding: 14px 10px; box-sizing: border-box; border-right: 1px solid ${token.colorBorderSecondary}; background: ${token.colorFillQuaternary}; }
        .redis-database-toolbar { display: flex; align-items: center; gap: ${compact ? 6 : 8}px; margin-bottom: 10px; flex: none; min-width: 0; }
        .redis-database-toolbar .ant-input-affix-wrapper { flex: 1; width: 0; min-width: 0; height: ${compact ? 28 : 30}px; padding-inline: 8px; font-size: ${compact ? 12 : 13}px; }
        .redis-database-toolbar .ant-input { font-size: inherit; }
        .redis-database-toolbar .ant-input-prefix { color: ${token.colorTextQuaternary}; }
        .redis-database-toolbar .ant-btn { flex: none; width: ${compact ? 28 : 30}px; height: ${compact ? 28 : 30}px; padding: 0; color: ${token.colorTextSecondary}; }
        .redis-database-toolbar .ant-btn .anticon { font-size: ${compact ? 13 : 14}px; }
        .redis-database-list { display: flex; flex-direction: column; flex: 1; min-height: 0; overflow: auto; scrollbar-width: thin; margin-inline: -10px; padding: 0 10px 10px; }
        .redis-database-list > .redis-catalog-empty { flex: 1; min-height: 0; }
        .redis-catalog-more { display: flex; flex: none; justify-content: center; padding: 10px 0; }
        .redis-database-item { display: flex; flex: none; align-items: center; gap: 7px; width: 100%; min-height: ${compact ? 36 : 40}px; padding: ${compact ? 9 : 10}px 8px; border: 0; border-radius: ${token.borderRadius}px; background: transparent; color: ${token.colorText}; font-family: inherit; font-size: ${compact ? 12 : 13}px; line-height: 1.5; text-align: left; cursor: pointer; }
        .redis-database-item .anticon { color: ${token.colorTextTertiary}; }
        .redis-database-item + .redis-database-item { margin-top: 3px; }
        .redis-database-item > span:last-child { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .redis-database-item:hover { background: ${token.colorFillTertiary}; }
        .redis-database-item.is-selected { background: ${token.colorPrimaryBg}; color: ${token.colorPrimary}; }
        .redis-database-item.is-selected .anticon { color: inherit; }
        .redis-database-item:focus-visible { outline: 2px solid ${token.colorPrimaryBorder}; outline-offset: -2px; }
        .redis-instance-footer { display: flex; flex: none; flex-direction: column; gap: 3px; min-width: 0; padding: 12px 6px 0; border-top: 1px solid ${token.colorBorderSecondary}; box-sizing: border-box; }
        .redis-instance-name { font-size: ${compact ? 12 : 13}px; line-height: 20px; color: ${token.colorText}; }
        .redis-instance-version { font-size: ${compact ? 11 : 12}px; color: ${token.colorTextTertiary}; font-variant-numeric: tabular-nums; line-height: 18px; }
        .redis-connection-state { display: flex; flex: 1; flex-direction: column; justify-content: center; align-items: center; gap: 12px; padding: 24px; }
        .redis-catalog-empty { display: flex; min-height: 180px; align-items: center; justify-content: center; }
        @media (max-width: 767px) {
            height: calc(100dvh - 48px); min-height: 0;
            .redis-window-main { padding: 12px; }
            .redis-window-header { gap: 8px; padding-inline: 12px; }
            .redis-window-title { flex: 1; gap: 8px; }
            .redis-window-actions { gap: 0; }
        }
    `,
}));
