import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    editor: css`
        .ant-modal-body { font-size: ${token.fontSize}px; }
        .ant-modal-body .ant-btn, .ant-form-item-label > label, .ant-input, .ant-select-selector, .ant-select-selection-item, .ant-checkbox-wrapper { font-size: ${token.fontSize}px; }
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .schema-field-editor {
            --schema-field-font-size: 12px;
            --schema-field-control-height: 28px;
            font-size: var(--schema-field-font-size);
        }
        .schema-field-editor[data-compact] {
            --schema-field-font-size: 11px;
            --schema-field-control-height: 24px;
        }
        .schema-field-editor :is(.ant-btn, .ant-form-item-label > label, .ant-input, .ant-select-selector, .ant-select-selection-item, .ant-select-placeholder, .ant-checkbox-wrapper, .ant-form-item-explain) {
            font-size: var(--schema-field-font-size);
        }
        .schema-field-editor .ant-form-item-control-input { min-height: var(--schema-field-control-height); }
        .schema-field-editor .ant-form-item-label { padding-bottom: 4px; }
        /* 字段行保留错误高亮，校验文字由提交提示显示，避免改变行高。 */
        .schema-field-editor .ant-form-item-additional { display: none; }
        .schema-field-list { display: flex; flex-direction: column; gap: 10px; }
        .schema-field-add { margin-top: ${token.marginSM}px; }
        .schema-field-row {
            display: flex; flex-direction: column; gap: 10px; min-width: 0; padding: 10px;
            border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusSM}px;
        }
        .schema-field-row .ant-form-item { min-width: 0; margin: 0; }
        .schema-field-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; color: ${token.colorTextSecondary}; }
        .schema-field-tools { display: inline-flex; align-items: center; flex: none; gap: 2px; }
        .schema-field-tools .ant-btn { width: var(--schema-field-control-height); height: var(--schema-field-control-height); padding: 0; }
        .schema-field-main, .schema-field-options {
            display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; gap: 10px;
        }
        .schema-field-options { grid-template-columns: repeat(3, minmax(0, 1fr)); }
        .schema-field-row .ant-select { width: 100%; }
        .schema-field-details { display: grid; grid-template-columns: minmax(0, 1fr); align-items: start; gap: 10px; }
        .schema-field-details > * { min-width: 0; }
        .schema-field-details > :only-child { grid-column: 1 / -1; }
        .schema-field-details .ant-form-item { margin: 0; }
        .schema-field-details .schema-default { grid-template-columns: minmax(0, 1fr); gap: 6px; margin-top: 0; }
        .schema-field-details .schema-default[data-has-value] { grid-template-columns: 94px minmax(0, 1fr); }
        .schema-field-editor[data-compact] .schema-field-list { gap: 8px; }
        .schema-field-editor[data-compact] .schema-field-row { padding: 8px; gap: 8px; }
        .schema-field-editor[data-compact] :is(.schema-field-main, .schema-field-options, .schema-field-details) { gap: 8px; }
        .schema-default > *, .schema-index-row > * { min-width: 0; }
        .schema-default { display: grid; grid-template-columns: 116px minmax(0, 1fr); gap: 10px; margin-top: 10px; }
        .schema-default .ant-form-item { margin-bottom: 0; }
        .schema-advanced { display: flex; flex-direction: column; gap: 8px; margin-top: 12px; background: transparent; }
        .schema-advanced > .ant-collapse-item { overflow: hidden; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .schema-advanced > .ant-collapse-item > .ant-collapse-header {
            align-items: center; gap: 8px; padding: 9px 12px; background: ${token.colorFillQuaternary}; color: ${token.colorTextSecondary};
        }
        .schema-advanced > .ant-collapse-item > .ant-collapse-header > .ant-collapse-title {
            display: flex; align-items: center; min-width: 0; min-height: 20px; margin: 0;
        }
        .schema-advanced-label { display: flex; align-items: center; gap: 7px; font-size: ${token.fontSizeSM}px; line-height: 20px; }
        .schema-advanced-label > .anticon {
            display: flex; align-items: center; justify-content: center; flex: none; width: 1em; height: 1em;
            color: ${token.colorTextTertiary}; font-size: ${token.fontSize}px; line-height: 1;
        }
        .schema-advanced-label > .anticon svg { display: block; }
        .schema-advanced > .ant-collapse-item > .ant-collapse-header > .ant-collapse-expand-icon {
            display: inline-flex; align-items: center; justify-content: center; align-self: center; height: 20px; padding: 0; color: ${token.colorTextTertiary};
        }
        .schema-advanced > .ant-collapse-item > .ant-collapse-header > .ant-collapse-expand-icon .anticon { font-size: 10px; }
        .schema-advanced > .ant-collapse-item > .ant-collapse-panel > .ant-collapse-body { padding: 10px 12px 12px; }
        .schema-advanced > .ant-collapse-item > .ant-collapse-panel > .ant-collapse-body > .ant-form-item { margin-bottom: 0; }
        .schema-create-indexes { display: flex; flex-direction: column; gap: 10px; }
        .schema-create-index { display: flex; flex-direction: column; gap: 10px; min-width: 0; padding: 10px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusSM}px; }
        .schema-create-index .ant-form-item { margin: 0; }
        .schema-index-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; color: ${token.colorTextSecondary}; }
        .schema-index-heading .ant-btn { width: var(--schema-field-control-height); height: var(--schema-field-control-height); padding: 0; }
        .schema-index-settings { display: grid; grid-template-columns: minmax(0, 1fr) 116px; gap: 10px; }
        .schema-index-settings > *, .schema-index-columns { min-width: 0; }
        .schema-create-index .schema-index-list { gap: 6px; }
        .schema-create-index .schema-index-row { grid-template-columns: minmax(0, 1fr) 76px var(--schema-field-control-height); gap: 6px; }
        .schema-create-index .schema-index-row .ant-form-item { margin: 0; }
        .schema-create-index .schema-index-row .ant-select { width: 100%; }
        .schema-index-column-add { margin-top: 6px; }
        .schema-index-row { display: grid; grid-template-columns: minmax(0, 1fr) 102px 28px; align-items: start; gap: 8px; }
        .schema-index-row .ant-form-item { margin-bottom: 0; }
        .schema-index-list { display: flex; flex-direction: column; gap: 10px; }
        @media (max-width: 575px) {
            .schema-default { grid-template-columns: 98px minmax(0, 1fr); gap: 8px; }
            .schema-index-row { grid-template-columns: minmax(0, 1fr) 84px 28px; gap: 6px; }
        }
        @media (max-width: 399px) {
            .schema-field-main { grid-template-columns: minmax(0, 1fr); }
            .schema-field-options { gap: 6px; }
            .schema-index-settings { grid-template-columns: minmax(0, 1fr) 94px; gap: 8px; }
            .schema-create-index .schema-index-row { grid-template-columns: minmax(0, 1fr) 68px var(--schema-field-control-height); }
            .schema-create-index .schema-index-row > .ant-form-item:nth-child(2) { grid-column: 2; grid-row: 1; }
            .schema-create-index .schema-index-row > .ant-btn { grid-column: 3; grid-row: 1; }
            .schema-index-row { grid-template-columns: minmax(0, 1fr) 28px; }
            .schema-index-row > .ant-form-item:nth-child(2) { grid-column: 1; grid-row: 2; }
            .schema-index-row > .ant-btn { grid-column: 2; grid-row: 1; }
        }
    `,
    formEditor: css`
        .schema-note { margin-top: 10px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; line-height: 1.65; }
        .schema-column-options { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; gap: ${token.marginSM}px; }
        .schema-column-options > .ant-form-item { min-width: 0; }
        .schema-column-default { display: flex; flex-direction: column; gap: ${token.marginXS}px; }
        .schema-column-default > .ant-form-item { margin-bottom: 0; }
        .schema-index-fields { display: flex; flex-direction: column; gap: ${token.marginSM}px; }
        .schema-index-field { display: grid; grid-template-columns: minmax(0, 1fr) 88px ${token.controlHeight}px; align-items: start; gap: ${token.marginXS}px; }
        .schema-index-field .ant-form-item { min-width: 0; margin: 0; }
        .schema-index-field .ant-select { width: 100%; }
        .schema-index-field > .ant-btn { width: ${token.controlHeight}px; padding-inline: 0; }
        .schema-field-add { margin-top: ${token.marginSM}px; }
    `,
    typeOption: css`
        display: flex; align-items: baseline; justify-content: space-between; gap: 12px; min-width: 0;
        > span { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        > small { flex: none; color: ${token.colorTextTertiary}; font-size: inherit; }
    `,
}));
