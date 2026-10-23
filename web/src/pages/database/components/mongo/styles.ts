import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    panel: css`
        display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0;
        color: ${token.colorText}; font-size: ${compact ? 12 : 13}px;

        && .ant-table-cell:is(.ant-table-cell-fix, .ant-table-cell-fix-start, .ant-table-cell-fix-end, .ant-table-cell-fix-left, .ant-table-cell-fix-right) { transition-property: background-color, color, border-color, box-shadow; }
        @supports (width: 1cqi) {
            .ant-table-has-fix-end :is(.ant-table-header, .ant-table-body, .ant-table-content, .ant-table-tbody-virtual-holder) { container-type: inline-size; }
            && .ant-table-cell-fix-end { inset-inline-end: calc(100% - 100cqi) !important; }
        }
        .manager-header { display: flex; flex: none; justify-content: center; margin-bottom: 16px; min-width: 0; }
        .manager-navigation { min-width: 0; max-width: 100%; overflow-x: auto; scrollbar-width: none; }
        .manager-navigation::-webkit-scrollbar { display: none; }
        .manager-navigation .ant-segmented-group { display: flex; width: max-content; min-width: 100%; }
        .manager-navigation .ant-segmented-item { flex: 1 0 auto; }
        .manager-navigation .ant-segmented-item-label { padding-inline: 12px; }
        .manager-body, .manager-stage { display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0; }
        .manager-toolbar { display: flex; flex: none; flex-wrap: wrap; align-items: center; gap: 8px; margin-bottom: 12px; min-width: 0; }
        .manager-tools { display: flex; align-items: center; gap: 6px; min-width: 0; margin-left: 0; }
        .manager-tools > *, .mongo-toolbar-actions > * { flex: none; }
        .mongo-toolbar-actions { display: contents; }
        .manager-toolbar .manager-picker, .manager-search { width: 240px; min-width: 0; max-width: 100%; }
        /* 工具栏与 MySQL 工作区共用同一套尺寸与主题色，不影响弹窗表单。 */
        &&& .manager-toolbar :is(.ant-select, .ant-input-affix-wrapper, .ant-input:not(textarea)) {
            height: ${compact ? 24 : 28}px; border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px;
        }
        &&& .manager-toolbar :is(.ant-input-affix-wrapper, .ant-input:not(textarea)) { padding-inline: ${compact ? 8 : 10}px; }
        &&& .manager-toolbar .ant-input-affix-wrapper .ant-input { height: auto; padding: 0; border-radius: 0; font-size: inherit; }
        .manager-toolbar .ant-input::placeholder { color: ${token.colorTextQuaternary}; font-size: inherit; }
        .manager-toolbar .ant-input-prefix { margin-inline-end: 8px; color: ${token.colorTextTertiary}; font-size: ${compact ? 12 : 14}px; }
        .manager-toolbar .ant-select-suffix { color: ${token.colorTextQuaternary}; font-size: ${compact ? 9 : 10}px; }
        &&& .manager-toolbar .manager-picker.ant-select {
            align-items: center; height: ${compact ? 30 : 34}px; padding: 0 ${compact ? 7 : 10}px;
            border-color: transparent; background: ${token.colorFillQuaternary}; box-shadow: none;
            transition: background-color .16s ease, border-color .16s ease;
        }
        &&& .manager-toolbar .manager-picker:not(.ant-select-disabled):hover,
        &&& .manager-toolbar .manager-picker.ant-select-open { background: ${token.colorFillSecondary}; }
        &&& .manager-toolbar .manager-picker.ant-select-focused:not(.ant-select-open) {
            border-color: ${token.colorPrimaryBorder}; background: ${token.colorBgContainer}; box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
        }
        &&& .manager-toolbar .ant-select-disabled { color: ${token.colorTextDisabled}; }
        .manager-picker .ant-select-prefix { margin-inline-end: 5px; color: ${token.colorTextQuaternary}; }
        .manager-picker .ant-select-content { margin-inline-end: 5px; }
        &&& .manager-picker :is(.ant-select-placeholder, .ant-select-input) { color: inherit; }
        &&& :is(.manager-toolbar, .manager-tabs, .maintenance-row, .mongo-query-toolbar) .ant-btn,
        &&& .manager-stage > .ant-btn {
            height: ${compact ? 24 : 28}px; padding-inline: ${compact ? 7 : 9}px; gap: ${compact ? 4 : 6}px;
            border-radius: ${token.borderRadiusSM}px; font-size: ${compact ? 11 : 12}px;
        }
        &&& :is(.manager-toolbar, .manager-tabs, .maintenance-row, .mongo-query-toolbar) .ant-btn-color-default:not(:disabled) { color: ${token.colorTextSecondary}; }
        &&& :is(.manager-toolbar, .manager-tabs, .maintenance-row, .mongo-query-toolbar) .ant-btn:disabled { border-color: transparent; background: ${token.colorFillTertiary}; }
        && :is(.manager-toolbar, .manager-tabs, .maintenance-row, .mongo-query-toolbar) .ant-btn .anticon { font-size: inherit; }
        &&& :is(.manager-toolbar, .manager-tabs, .maintenance-row, .mongo-query-toolbar) .ant-btn-icon-only { width: ${compact ? 24 : 28}px; padding-inline: 0; }
        .manager-toolbar .client-toolbar-dropdown { gap: 5px; }
        .manager-toolbar .client-toolbar-dropdown > .ant-btn-icon { color: ${token.colorTextTertiary}; }
        && .manager-toolbar .client-toolbar-arrow { color: ${token.colorTextQuaternary}; font-size: ${compact ? 9 : 10}px; transition: transform .16s ease; }
        .manager-toolbar .ant-dropdown-open .client-toolbar-arrow { transform: rotate(180deg); }
        .manager-tabs { flex: none; min-width: 0; margin-bottom: ${compact ? 8 : 10}px; padding: 8px 0; }
        .manager-tabs > .ant-tabs-nav { margin: 0; }
        .manager-tabs .ant-tabs-nav::before { display: none; }
        .manager-tabs .ant-tabs-ink-bar { display: none; }
        &&& .manager-tabs .ant-tabs-tab {
            justify-content: center; min-width: ${compact ? 44 : 50}px; height: ${compact ? 26 : 30}px;
            padding: 0 ${compact ? 10 : 12}px; border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; font-weight: 400; line-height: 1.5;
            transition: color .16s ease, background-color .16s ease;
        }
        &&& .manager-tabs .ant-tabs-tab + .ant-tabs-tab { margin-inline-start: 0; }
        &&& .manager-tabs .ant-tabs-tab:not(.ant-tabs-tab-disabled):not(.ant-tabs-tab-active):hover { color: ${token.colorText}; background: ${token.colorFillQuaternary}; }
        &&& .manager-tabs .ant-tabs-tab-active { background: ${token.colorPrimaryBg}; }
        &&& .manager-tabs .ant-tabs-tab-active .ant-tabs-tab-btn { color: ${token.colorPrimary}; font-weight: 400; text-shadow: none; }
        &&& .manager-tabs .ant-tabs-tab-disabled { color: ${token.colorTextDisabled}; }
        .manager-tabs .ant-tabs-tab-btn:focus-visible { border-radius: 2px; outline-offset: -2px; }
        .manager-tabs .ant-tabs-extra-content { display: flex; align-items: center; gap: 6px; }
        .manager-tabs .ant-tabs-content-holder { display: none; }
        .manager-empty, .manager-loading { display: flex; flex: 1; min-height: 240px; align-items: center; justify-content: center; }
        .manager-cell { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .manager-cell-null { color: ${token.colorTextQuaternary}; }
        .manager-table { display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0; }
        .manager-statistics { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: ${token.colorTextTertiary}; font-size: ${compact ? 11 : 12}px; }
        .manager-refresh { margin-left: auto; flex: none; }
        .manager-output pre { margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; }
        .manager-table .ant-table { font-size: ${compact ? 11 : 12}px; color: ${token.colorTextSecondary}; }
        .manager-table .ant-table-thead > tr > th { font-size: ${compact ? 12 : 13}px; font-weight: 500; color: ${token.colorTextSecondary}; }
        .manager-table .ant-table-cell { padding: ${compact ? 6 : 8}px ${compact ? 8 : 10}px !important; white-space: nowrap; }
        .manager-table .ant-table-cell .ant-btn { height: ${compact ? 22 : 24}px; padding-inline: 7px; color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; }
        .manager-table .ant-table-container { border-inline: 1px solid ${token.colorBorderSecondary}; border-bottom: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; overflow: hidden; }
        && .manager-table .ant-table-tbody > tr:last-child > td,
        && .manager-table .ant-table-tbody-virtual-holder-inner > .client-last-row > .ant-table-cell { border-bottom: 0; }
        .manager-table .ant-table-body, .manager-table .ant-table-tbody-virtual-holder { scrollbar-width: none; }
        .manager-table .ant-table-body::-webkit-scrollbar, .manager-table .ant-table-tbody-virtual-scrollbar { display: none; }
        .manager-footer { display: flex; flex: none; align-items: center; gap: 8px; padding-top: 10px; margin-top: auto; min-width: 0; }
        .manager-footer > .ant-typography { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: ${compact ? 11 : 12}px; color: ${token.colorTextTertiary}; }
        .manager-footer .ant-pagination { display: flex; flex-wrap: nowrap; align-items: center; gap: 4px; margin: 0 0 0 auto; font-size: ${compact ? 11 : 12}px; }
        .manager-footer .ant-pagination-item, .manager-footer .ant-pagination-prev, .manager-footer .ant-pagination-next { margin: 0; min-width: ${compact ? 22 : 26}px; height: ${compact ? 22 : 26}px; line-height: ${compact ? 22 : 26}px; border: 0; background: transparent; }
        .manager-footer .ant-pagination-item-active { background: ${token.colorBgContainer}; box-shadow: inset 0 0 0 1px ${token.colorPrimaryBorder}; }
        .manager-footer .ant-pagination-item:not(.ant-pagination-item-active) a, .manager-footer .ant-pagination-item-link { color: ${token.colorTextTertiary}; }
        .manager-footer .ant-pagination-options { display: flex; flex-wrap: nowrap; gap: 8px; margin-left: 4px; }
        .manager-footer .ant-select { height: ${compact ? 22 : 26}px; font-size: inherit; }
        .manager-footer .ant-select-filled.ant-select-disabled { border-color: transparent; }
        .manager-footer .ant-pagination-options-quick-jumper { height: ${compact ? 22 : 26}px; line-height: ${compact ? 22 : 26}px; }
        .manager-footer .ant-pagination-options-quick-jumper input { height: 100%; width: 42px; font-size: inherit; }
        .manager-output { flex: 1; min-height: 140px; margin: 0; padding: 12px; overflow: auto; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorFillQuaternary}; color: ${token.colorText}; font: ${compact ? 11 : 12}px/1.65 ${token.fontFamilyCode}; white-space: pre-wrap; overflow-wrap: anywhere; }
        .manager-output.is-error { color: ${token.colorError}; }
        .maintenance-list { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; overflow: hidden; }
        .maintenance-row { display: grid; grid-template-columns: minmax(0,1fr) auto; align-items: center; gap: 12px; padding: ${compact ? 10 : 12}px ${compact ? 12 : 14}px; }
        .maintenance-row + .maintenance-row { border-top: 1px solid ${token.colorBorderSecondary}; }
        .maintenance-row > div { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
        .maintenance-tool-title { font-size: ${compact ? 12 : 13}px; color: ${token.colorText}; }
        .maintenance-row > div > strong { font-weight: 500; font-size: ${compact ? 12 : 13}px; }
        .maintenance-row > div > span, .maintenance-row .ant-typography { font-size: ${compact ? 11 : 12}px; color: ${token.colorTextSecondary}; }
        @media(min-width:768px) {
            .mongo-main-toolbar > .manager-tools { flex: 1; flex-wrap: wrap; }
            .mongo-main-toolbar .mongo-toolbar-actions > .ant-btn:first-of-type { margin-inline-start: auto; }
            .mongo-main-toolbar > .manager-refresh { margin-left: 0; }
        }
        @media(max-width:767px) {
            .manager-header { margin-bottom: 12px; }
            .manager-navigation { width: 100%; }
            .manager-navigation .ant-segmented-item-label { padding-inline: 6px; }
            .manager-body { min-height: 420px; }
            .manager-tools { max-width: 100%; margin-left: 0; overflow-x: auto; scrollbar-width: none; gap: 6px; }
            .manager-refresh { margin-left: auto; flex: none; }
            .manager-toolbar { gap: 8px; flex-wrap: nowrap; }
            .manager-toolbar > .manager-tools { flex: 1; width: 0; }
            .mongo-main-toolbar { flex-wrap: wrap; }
            .mongo-main-toolbar > .manager-tools { display: contents; }
            .mongo-main-toolbar .mongo-toolbar-actions { display: flex; flex: 1; width: 0; min-width: 0; align-items: center; gap: 6px; overflow-x: auto; scrollbar-width: none; }
            .mongo-main-toolbar .manager-picker, .mongo-main-toolbar .manager-search { flex: 0 0 100%; width: 100%; }
            .manager-footer { flex-wrap: nowrap; gap: 6px; }
            .manager-footer .ant-pagination { gap: 2px; }
            .manager-footer .ant-pagination-options { order: -1; margin-inline: 0 4px; }
            .manager-footer .ant-pagination-simple-pager { margin: 0; display: flex; align-items: center; gap: 4px; height: 26px; }
            .manager-footer .ant-pagination-simple-pager input { width: 34px; height: 24px; padding: 0 4px; margin: 0; border-color: transparent; border-radius: ${token.borderRadiusSM}px; background: ${token.colorFillTertiary}; }
            .manager-footer .ant-pagination-slash { margin: 0; }
        }

    `,
    modal: css`
        .ant-modal-body { min-width: 0; font-size: ${token.fontSize}px; line-height: 1.6; }
        .ant-form-item-label > label, .ant-input, .ant-input-number, .ant-select, .ant-radio-wrapper, .ant-checkbox-wrapper { font-size: ${token.fontSize}px; }
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .ant-form > .ant-form-item:last-child { margin-bottom: 0; }
        .ant-form-item-control, .ant-form-item-control-input-content { min-width: 0; }
        .ant-form-item-extra, .ant-form-item-explain { font-size: ${token.fontSizeSM}px; }
        .nosql-scroll, .nosql-member-view { max-height: none; overflow: visible; }
        .nosql-readonly { display: block; margin-bottom: 12px; }
        .mongo-account-loading { min-height: 180px; display: grid; place-items: center; }
        .mongo-account-grants { display: flex; flex-direction: column; gap: ${token.marginSM}px; min-width: 0; }
        .mongo-account-grant { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; padding: ${token.paddingSM}px; min-width: 0; }
        .mongo-account-grant-heading { display: grid; grid-template-columns: minmax(0, 1fr) ${token.controlHeight}px; gap: ${token.marginXS}px; align-items: start; }
        .mongo-account-grant-heading > .ant-btn { width: ${token.controlHeight}px; height: ${token.controlHeight}px; padding: 0; margin-top: ${token.fontSize * token.lineHeight + token.paddingXS}px; color: ${token.colorTextSecondary}; }
        .mongo-account-grant .ant-form-item { min-width: 0; }
        .mongo-account-grant > .ant-form-item:last-child { margin-bottom: 0; }
        .mongo-account-grant .ant-select { width: 100%; min-width: 0; }
        .mongo-index-options { display: flex; flex-wrap: wrap; gap: 12px; }
        .mongo-document-toolbar { display: grid; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr); align-items: center; gap: 8px; margin-bottom: ${token.margin}px; }
        .mongo-document-toolbar .ant-segmented { grid-column: 2; grid-row: 1; min-width: 0; font-size: ${token.fontSizeSM}px; }
        .mongo-document-toolbar .ant-segmented-item-label { padding-inline: 12px; }
        .mongo-document-toolbar > .ant-btn { grid-column: 3; grid-row: 1; justify-self: end; height: ${token.controlHeightSM}px; font-size: ${token.fontSizeSM}px; }
        .mongo-document-json { width: 100% !important; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; overflow: hidden; }
        textarea.mongo-document-json { font-family: ${token.fontFamilyCode}; resize: vertical; }
    `,
}));
