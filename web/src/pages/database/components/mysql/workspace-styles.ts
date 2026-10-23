import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    workspace: css`
        .ant-modal-body { max-height: calc(100dvh - 126px); overflow-y: auto; overscroll-behavior: contain; }
        .database-browser { display: grid; grid-template-columns: 208px minmax(0, 1fr); gap: 18px; min-height: 520px; }
        .database-catalog { min-width: 0; display: flex; flex-direction: column; align-self: start; position: sticky; top: 0; max-height: calc(100dvh - 165px); padding: 10px; border-radius: ${token.borderRadiusLG}px; background: ${token.colorFillQuaternary}; }
        .database-catalog-toolbar { display: flex; flex-direction: column; flex: none; min-width: 0; gap: ${compact ? 6 : 8}px; margin-bottom: 10px; }
        .database-catalog-search { width: 100%; min-width: 0; height: ${compact ? 28 : 30}px; padding-inline: 8px; font-size: ${compact ? 12 : 13}px; }
        .database-catalog-search .ant-input { font-size: inherit; }
        .database-catalog-search .ant-input-prefix { color: ${token.colorTextTertiary}; }
        .database-catalog-actions { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-width: 0; }
        .database-catalog-actions > .ant-space { min-width: 0; overflow-x: auto; scrollbar-width: none; }
        .database-catalog-actions .ant-btn { flex: none; width: ${compact ? 28 : 30}px; height: ${compact ? 28 : 30}px; padding: 0; font-size: ${compact ? 12 : 13}px; }
        .database-catalog-actions .ant-btn-color-default:not(:disabled) { color: ${token.colorTextSecondary}; }
        .database-catalog-actions .ant-btn .anticon { font-size: ${compact ? 13 : 14}px; }
        .database-catalog-list { display: flex; flex-direction: column; min-height: 160px; overflow-y: auto; overscroll-behavior: contain; }
        .database-catalog-more { display: flex; flex: none; align-items: center; justify-content: center; min-height: 36px; padding: 8px 0; color: ${token.colorTextSecondary}; }
        .database-catalog-item { display: flex; flex: none; align-items: center; min-width: 0; gap: 3px; padding: 3px 4px; border-radius: ${token.borderRadius}px; }
        .database-catalog-item + .database-catalog-item { margin-top: 3px; }
        .database-catalog-item:hover { background: ${token.colorFillTertiary}; }
        .database-catalog-item.active { background: ${token.colorPrimaryBg}; }
        .database-catalog-select { display: flex; align-items: center; flex: 1; min-width: 0; gap: 8px; padding: 6px 4px; background: none; border: 0; text-align: left; color: ${token.colorText}; font: inherit; cursor: pointer; }
        .database-catalog-select > span:last-child { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
        .database-catalog-select > .anticon { flex: none; color: ${token.colorTextTertiary}; font-size: ${token.fontSize}px; }
        .database-catalog-item.active .database-catalog-select, .database-catalog-item.active .database-catalog-select > .anticon { color: ${token.colorPrimary}; }
        .database-catalog-item > .ant-btn { flex: none; }
        .database-catalog-select:focus-visible { outline: 2px solid ${token.colorPrimaryBorder}; outline-offset: -1px; border-radius: ${token.borderRadius}px; }
        .database-browser-content { min-width: 0; }
        .database-browser-header { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-bottom: 12px; }
        .database-browser-heading { display: flex; align-items: center; gap: 10px; min-width: 0; }
        .database-browser-heading > strong { font-size: ${token.fontSize}px; font-weight: 500; color: ${token.colorText}; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .database-browser-heading > span { flex: none; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; }
        .database-browser-header .ant-btn { flex: none; }
        .database-catalog-state { display: flex; flex: 1; flex-direction: column; align-items: center; justify-content: center; min-height: 0; padding: 12px; text-align: center; overflow-wrap: anywhere; }
        .database-catalog-state .ant-typography { font-size: ${token.fontSizeSM}px; }
        .database-browser-empty { display: flex; align-items: center; justify-content: center; min-height: 460px; }
        @media (max-width: 767px) {
            .ant-modal-body { max-height: calc(100dvh - 112px); }
            .database-browser { grid-template-columns: minmax(0, 1fr); gap: 0; min-height: 440px; }
            .database-catalog { position: static; max-height: none; }
            .database-catalog-list { max-height: 55dvh; }
            .database-browser-header { flex-wrap: wrap; }
            .database-browser-heading { max-width: 100%; }
            .database-browser-empty { min-height: 360px; }
        }
    `,
    mysqlWorkspace: css`
        .ant-modal-container { padding: 0; overflow: hidden; }
        .ant-modal-header { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
        && .ant-modal-body { padding: 0; min-height: 0; max-height: none; overflow: hidden; }
        .mysql-window {
            display: flex; flex-direction: column; min-width: 0;
            height: ${compact ? 680 : 700}px; min-height: ${compact ? 680 : 700}px;
            background: ${token.colorBgContainer}; color: ${token.colorText}; font-size: ${token.fontSize}px;
        }
        .mysql-window-header { display: flex; align-items: center; flex: none; min-height: ${token.controlHeight + 20}px; gap: 16px; padding: 0 16px; border-bottom: 1px solid ${token.colorBorderSecondary}; }
        .mysql-window-title { display: flex; align-items: center; gap: 9px; min-width: 0; }
        .mysql-window-label { white-space: nowrap; font-weight: 500; }
        .mysql-catalog-toggle.ant-btn { flex: none; width: 32px; height: 32px; padding: 0; color: ${token.colorTextSecondary}; }
        .mysql-window-actions { display: flex; align-items: center; flex: none; gap: 2px; margin-left: auto; }
        .mysql-window-actions .ant-btn { color: ${token.colorTextSecondary}; }
        .mysql-connection-state { flex: 1; min-height: 0; display: flex; align-items: center; justify-content: center; overflow-y: auto; padding: 32px 20px; }
        .database-browser {
            position: relative;
            flex: 1;
            min-height: 0;
            grid-template-columns: ${compact ? 196 : 208}px minmax(0, 1fr);
            grid-template-rows: minmax(0, 1fr);
            gap: 0;
            overflow: hidden;
        }
        .database-catalog {
            position: static;
            align-self: stretch;
            min-height: 0;
            max-height: none;
            padding: 14px 10px;
            border-right: 1px solid ${token.colorBorderSecondary};
            border-radius: 0;
            background: ${token.colorFillQuaternary};
        }
        .database-catalog-item > .ant-btn { width: ${compact ? 28 : 30}px; height: ${compact ? 28 : 30}px; }
        .database-catalog-list { flex: 1; min-height: 0; max-height: none; margin-inline: -10px; padding-inline: 10px; scrollbar-width: thin; }
        .database-catalog-select { gap: 7px; padding-block: ${compact ? 6 : 7}px; font-size: ${compact ? 12 : 13}px; }
        .database-browser-content { display: flex; flex-direction: column; min-height: 0; padding: 16px 18px; overflow-y: auto; overscroll-behavior: contain; scrollbar-width: thin; }
        .database-browser-content > .ant-app { display: flex; flex: 1; flex-direction: column; min-height: 0; }
        .database-browser-content .sql-workbench > .sql-stage { min-height: 0; }
        .database-browser-empty { min-height: 0; height: 100%; }
        @media (max-width: 767px) {
            .mysql-window { height: calc(100dvh - 48px); min-height: 0; }
            .mysql-window-header { gap: 8px; padding-inline: 12px; }
            .mysql-window-actions { gap: 0; }
            .database-browser { grid-template-columns: minmax(0, 1fr); }
            .database-browser-content { padding: 12px; }
        }
        &.is-fullscreen { margin: 0; }
        &.is-fullscreen .ant-modal-container { border-radius: 0; }
        &.is-fullscreen .mysql-window { height: 100dvh; min-height: 0; }
    `,
    mysqlInstance: css`
        flex: none;
        display: flex;
        flex-direction: column;
        gap: 3px;
        min-width: 0;
        padding: 12px 6px 0;
        border-top: 1px solid ${token.colorBorderSecondary};
        .mysql-instance-name { color: ${token.colorText}; font-size: ${compact ? 12 : 13}px; line-height: 20px; }
        .mysql-instance-version { color: ${token.colorTextTertiary}; font-size: ${compact ? 11 : 12}px; line-height: 18px; }
    `,
    mysqlCatalogDrawer: css`
        .ant-drawer-body { display: flex; min-height: 0; padding: 0; overflow: hidden; }
        && .database-catalog { flex: 1; min-height: 0; max-height: none; padding: 14px; border: 0; border-radius: 0; box-sizing: border-box; }
        && .database-catalog-list { flex: 1; min-height: 0; max-height: none; margin-inline: -14px; padding-inline: 14px; }
        .database-catalog-select { font-size: ${compact ? 12 : 13}px; }
    `,
}));
