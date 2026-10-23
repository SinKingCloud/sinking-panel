import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    panel: css`
        display: flex; flex-direction: column; flex: 1; height: 100%; min-height: 0; min-width: 0; position: relative;
        color: ${token.colorText}; font-size: ${token.fontSize}px; line-height: 1.6;
        .redis-navigation { display: flex; justify-content: center; flex: none; min-width: 0; margin-bottom: 16px; }
        .redis-navigation > .ant-segmented { min-width: 0; max-width: 100%; overflow-x: auto; scrollbar-width: none; }
        .redis-navigation .ant-segmented-group { display: flex; width: max-content; min-width: 100%; }
        .redis-navigation .ant-segmented-item { flex: 1 0 auto; }
        .redis-navigation .ant-segmented-item-label { padding-inline: 12px; }
        .redis-tab-body { flex: 1; min-width: 0; min-height: 0; overflow: hidden; }
        .redis-tab-body .ant-input, .redis-tab-body .ant-select { font-size: inherit; }
        .redis-tab-body .ant-btn { height: ${compact ? 24 : 28}px; padding-inline: ${compact ? 7 : 9}px; gap: ${compact ? 4 : 6}px; border-radius: ${token.borderRadiusSM}px; font-size: ${compact ? 11 : 12}px; }
        .redis-tab-body .ant-btn-icon-only { width: ${compact ? 24 : 28}px; padding-inline: 0; }
        .redis-tab-body .ant-btn-color-default:not(:disabled) { color: ${token.colorTextSecondary}; }
        .redis-tab-body .ant-btn .anticon { font-size: inherit; }
        &&& .redis-toolbar .ant-btn {
            height: ${compact ? 24 : 28}px; padding-inline: ${compact ? 7 : 9}px; gap: ${compact ? 4 : 6}px;
            border-radius: ${token.borderRadiusSM}px; font-size: ${compact ? 11 : 12}px;
        }
        &&& .redis-toolbar .ant-btn-icon-only { width: ${compact ? 24 : 28}px; padding-inline: 0; }
        &&& .redis-toolbar .ant-btn-color-default:not(:disabled) { color: ${token.colorTextSecondary}; }
        &&& .redis-toolbar .ant-btn:disabled { border-color: transparent; background: ${token.colorFillTertiary}; }
        && .redis-toolbar .ant-btn .anticon { font-size: inherit; }
        &&& :is(.redis-browser-toolbar, .redis-account-toolbar) .ant-input-affix-wrapper {
            height: ${compact ? 24 : 28}px; padding-inline: ${compact ? 8 : 10}px; border-radius: ${token.borderRadiusSM}px;
            font-size: ${compact ? 11 : 12}px; color: ${token.colorTextSecondary};
        }
        &&& :is(.redis-browser-toolbar, .redis-account-toolbar) .ant-input { font-size: inherit; color: inherit; }
        :is(.redis-browser-toolbar, .redis-account-toolbar) .ant-input::placeholder { color: ${token.colorTextQuaternary}; }
        .redis-account-search .ant-input-prefix { margin-inline-end: 8px; color: ${token.colorTextTertiary}; font-size: ${compact ? 12 : 14}px; }
        .redis-tab-body .redis-copy-action, .redis-tab-body .ant-typography-copy { font-size: ${token.fontSizeSM}px; }
        .redis-tab-body .ant-typography-copy:not(.ant-typography-copy-success) { color: ${token.colorTextSecondary}; }
        .redis-copy-action .ant-typography-copy { display: inline-flex; align-items: center; margin: 0; padding-inline: 4px; }
        .redis-section { min-width: 0; }
        .redis-section[hidden] { display: none; }
        .redis-tool-page { height: 100%; min-height: 0; overflow: auto; padding: 0; box-sizing: border-box; overscroll-behavior: contain; }
        .redis-command-page { overflow: hidden; }
        .redis-command-page > div { height: 100%; min-height: 0; }
        .redis-browser { display: flex; flex-direction: column; position: relative; height: 100%; min-height: 0; min-width: 0; }
        .redis-workbench { display: grid; grid-template-columns: ${compact ? 208 : 224}px minmax(0, 1fr); flex: 1; min-height: 0; min-width: 0; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorBgContainer}; overflow: hidden; }
        .redis-catalog { display: flex; flex-direction: column; min-height: 0; min-width: 0; border-right: 1px solid ${token.colorBorderSecondary}; background: ${token.colorFillQuaternary}; }
        .redis-catalog-actions { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 0 10px 8px; flex: none; }
        .redis-browser-toolbar { display: flex; align-items: center; padding: 10px 10px 8px; min-width: 0; flex: none; }
        .redis-search { display: flex; align-items: center; width: 100%; min-width: 0; height: ${compact ? 24 : 28}px; border: 1px solid transparent; border-radius: ${token.borderRadiusSM}px; background: ${token.colorFillTertiary}; transition: background-color .16s ease; }
        .redis-search:hover { background: ${token.colorFillSecondary}; }
        .redis-search:focus-within { background: ${token.colorFillSecondary}; }
        &&& .redis-browser-toolbar .redis-search > .redis-search-mode { flex: none; height: 100%; padding-inline: ${compact ? 8 : 10}px; border: 0; border-radius: 0; background: transparent; box-shadow: none; }
        &&& .redis-browser-toolbar .redis-search > .ant-input-affix-wrapper { flex: 1; width: 0; min-width: 0; height: calc(100% - 8px); padding-block: 0; border: 0; border-inline-start: 1px solid ${token.colorBorderSecondary}; border-radius: 0; background: transparent; box-shadow: none; outline: 0; }
        .redis-catalog-heading { display: flex; align-items: center; gap: 8px; min-height: 32px; padding-inline: 12px 8px; flex: none; }
        .redis-catalog-heading > span { flex: 1; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .redis-catalog-heading .ant-checkbox-wrapper { flex: none; }
        .redis-catalog-scroll { flex: 1; min-height: 0; overflow: auto; overscroll-behavior: contain; scrollbar-width: thin; padding: 0 6px 8px; }
        .redis-catalog-loading, .redis-catalog-items { min-height: 100%; }
        .redis-key-row { display: flex; align-items: center; gap: 6px; min-height: ${compact ? 30 : 34}px; padding-inline: 7px 2px; border-radius: ${token.borderRadiusSM}px; }
        .redis-key-row + .redis-key-row { margin-top: 2px; }
        .redis-key-row:hover { background: ${token.colorFillTertiary}; }
        .redis-key-row.is-selected { background: ${token.colorPrimaryBg}; }
        .redis-key-row > .ant-checkbox-wrapper { flex: none; }
        .redis-key-item { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; min-height: ${compact ? 30 : 34}px; padding: 0; border: 0; background: transparent; color: ${token.colorTextSecondary}; font: inherit; font-size: ${compact ? 11 : 12}px; text-align: left; cursor: pointer; }
        .redis-key-item > span { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .redis-key-item > small { flex: none; width: 42px; color: ${token.colorTextSecondary}; font-size: ${compact ? 10 : 11}px; text-transform: uppercase; }
        .redis-key-row.is-selected .redis-key-item > span { color: ${token.colorPrimary}; }
        .redis-key-item:focus-visible { outline: 2px solid ${token.colorPrimaryBorder}; outline-offset: 1px; border-radius: 2px; }
        && .redis-key-menu.ant-btn { flex: none; width: ${compact ? 24 : 28}px; height: ${compact ? 24 : 28}px; padding: 0; border-radius: ${token.borderRadiusSM}px; color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; }
        .redis-key-menu .anticon { font-size: inherit; }
        .redis-catalog-footer { display: flex; align-items: center; justify-content: space-between; min-height: ${compact ? 38 : 42}px; gap: 8px; padding-inline: 10px; border-top: 1px solid ${token.colorBorderSecondary}; color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; flex: none; }
        .redis-catalog-pager { display: flex; align-items: center; gap: 1px; }
        .redis-catalog-pager > span { min-width: 18px; text-align: center; font-variant-numeric: tabular-nums; }
        .redis-catalog-note { padding: 0 12px 6px; color: ${token.colorTextTertiary}; }
        .redis-catalog-empty, .redis-inspector-empty, .redis-content-empty { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; min-height: 180px; padding: 16px; overflow-wrap: anywhere; text-align: center; }
        .redis-inspector-empty { height: 100%; }
        .redis-inspector { min-width: 0; min-height: 0; }
        .redis-detail { display: flex; flex-direction: column; height: 100%; min-width: 0; min-height: 0; }
        .redis-detail-header { display: flex; align-items: center; gap: 10px; padding: ${compact ? 10 : 12}px; border-bottom: 1px solid ${token.colorBorderSecondary}; box-sizing: border-box; flex: none; }
        .redis-detail-heading { flex: 1; min-width: 0; }
        .redis-detail-name { display: flex; align-items: center; gap: 8px; min-width: 0; }
        .redis-current-key { flex: 0 1 auto; min-width: 0; width: auto; font-size: ${compact ? 12 : 13}px; color: ${token.colorText}; font-weight: 400; }
        .redis-key-copy { flex: none; }
        .redis-key-copy .ant-typography-copy { display: inline-flex; align-items: center; margin: 0; }
        .redis-detail-actions { display: flex; gap: 6px; align-items: center; flex: none; }
        .redis-dirty-dot { width: 6px; height: 6px; flex: none; border-radius: 50%; background: ${token.colorWarning}; }
        .redis-key-meta { display: flex; align-items: center; flex-wrap: wrap; gap: 4px 10px; margin-top: 6px; color: ${token.colorTextTertiary}; font-size: ${compact ? 11 : 12}px; min-width: 0; font-variant-numeric: tabular-nums; }
        .redis-key-meta > span { white-space: nowrap; }
        .redis-key-meta .ant-btn { height: auto; padding: 0; gap: 5px; color: ${token.colorTextSecondary}; font-size: inherit; }
        .redis-type { padding: 1px 5px; border-radius: ${token.borderRadiusSM}px; background: ${token.colorFillTertiary}; color: ${token.colorTextSecondary}; text-transform: uppercase; font-size: ${compact ? 10 : 11}px; line-height: 1.5; }
        .redis-detail-content { flex: 1; min-height: 0; min-width: 0; overflow: auto; padding: ${compact ? 10 : 12}px; scrollbar-width: thin; }
        .redis-value-spin { height: 100%; min-width: 0; min-height: 0; }
        .redis-value-spin-content { display: flex; flex-direction: column; height: 100%; min-height: 0; min-width: 0; }
        .redis-value-spin-content > .redis-value-fill { flex: 1; height: auto; min-height: 0; }
        .redis-value-spin-content > .redis-content-empty { flex: 1; }
        .redis-readonly-note { flex: none; color: ${token.colorTextTertiary}; margin-bottom: 8px; font-size: ${token.fontSizeSM}px; }
        .redis-member-pager { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 8px ${compact ? 10 : 12}px; flex: none; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .redis-detail-footer { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-height: ${compact ? 42 : 46}px; padding: 8px ${compact ? 10 : 12}px; border-top: 1px solid ${token.colorBorderSecondary}; flex: none; box-sizing: border-box; }
        .redis-save-state { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: ${token.colorTextTertiary}; font-size: ${compact ? 11 : 12}px; }
        .redis-save-state.is-dirty { color: ${token.colorWarningText}; }
        @media (max-width: 767px) {
            .redis-navigation { margin-bottom: 12px; }
            .redis-navigation .ant-segmented-item-label { padding-inline: 7px; }
            .redis-workbench { display: block; }
            .redis-catalog { height: 100%; border-right: 0; }
            .redis-inspector { height: 100%; }
            .redis-detail-header { flex-wrap: wrap; gap: 8px; padding: 10px; }
            .redis-detail-heading { min-width: 0; }
            .redis-detail-actions { margin-left: auto; }
            .redis-detail-content { padding: 10px; }
            .redis-detail-footer { gap: 8px; padding-inline: 10px; }
            .redis-detail-footer > .ant-space { margin-left: auto; }
            .redis-tool-page { padding: 0; }
            .redis-member-pager { padding-inline: 12px; }
            .redis-current-key { max-width: 100%; }
        }
        @media (max-width: 375px) {
            .redis-navigation .ant-segmented-item-label { padding-inline: 4px; }
        }
        && .ant-table-cell:is(.ant-table-cell-fix, .ant-table-cell-fix-start, .ant-table-cell-fix-end, .ant-table-cell-fix-left, .ant-table-cell-fix-right) { transition-property: background-color, color, border-color, box-shadow; }
        @supports (width: 1cqi) {
            .ant-table-has-fix-end :is(.ant-table-header, .ant-table-body, .ant-table-content, .ant-table-tbody-virtual-holder) { container-type: inline-size; }
            && .ant-table-cell-fix-end { inset-inline-end: calc(100% - 100cqi) !important; }
        }
        @media (max-height: 420px) {
            .redis-tab-body { overflow-y: auto; }
            .redis-browser { min-height: 280px; }
        }
    `,
    modal: css`
        .ant-modal-body { min-width: 0; font-size: ${token.fontSize}px; line-height: 1.6; }
        .ant-form-item-label > label, .ant-input, .ant-input-number, .ant-select, .ant-radio-wrapper, .ant-checkbox-wrapper { font-size: ${token.fontSize}px; }
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .ant-form > .ant-form-item:last-child { margin-bottom: 0; }
        .ant-form-item-extra, .ant-form-item-explain { font-size: ${token.fontSizeSM}px; }
        .ant-form-item-control, .ant-form-item-control-input-content { min-width: 0; }
        .redis-create-fields { display: grid; grid-template-columns: minmax(0, 1fr) 160px; gap: 0 12px; }
        .redis-create-fields > .ant-form-item { min-width: 0; margin-bottom: ${token.margin}px; }
        @media (max-width: 575px) {
            .redis-create-fields { grid-template-columns: minmax(0, 1fr); gap: 0; }
        }
    `,
}));
