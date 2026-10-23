import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    editor: css`
        display: flex; flex-direction: column; gap: ${token.marginSM}px; min-width: 0;
        .mongo-filter-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
        .mongo-filter-toolbar > .ant-select { width: 200px; max-width: calc(100% - 64px); }
        .mongo-filter-rows { display: flex; flex-direction: column; gap: ${token.marginSM}px; min-width: 0; }
        .mongo-filter-row { display: flex; flex-direction: column; gap: 8px; padding: ${token.paddingSM}px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; }
        .mongo-filter-main { display: grid; grid-template-columns: minmax(0, 1fr) 128px ${token.controlHeight}px; gap: 8px; align-items: start; }
        .mongo-filter-main > .ant-select { width: 100%; min-width: 0; }
        .mongo-filter-main > .ant-btn { padding-inline: 0; }
        .mongo-filter-value { display: flex; gap: 8px; align-items: start; min-width: 0; }
        .mongo-filter-input { flex: 1; min-width: 0; }
        .mongo-filter-input > :is(.ant-input, .ant-select, .ant-input-number, .ant-picker) { width: 100%; }
        .mongo-filter-value > .ant-select { flex: 0 0 106px; min-width: 0; }
        .mongo-filter-empty { padding: ${token.paddingSM}px 0; color: ${token.colorTextTertiary}; text-align: center; font-size: ${token.fontSizeSM}px; }
        .mongo-filter-add { align-self: stretch; }
        @media (max-width: 575px) {
            .mongo-filter-main { grid-template-columns: minmax(0, 1fr) 112px ${token.controlHeight}px; gap: 6px; }
            .mongo-filter-row { padding: 9px; }
            .mongo-filter-value { gap: 6px; }
            .mongo-filter-value > .ant-select { flex-basis: 86px; }
        }
    `,
    modal: css`
        .mongo-filter-mode { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; margin-bottom: ${token.margin}px; }
        .mongo-filter-mode > div:last-child { display: flex; align-items: center; gap: 6px; margin-left: auto; }
        .mongo-filter-mode .ant-btn { flex: none; height: ${token.controlHeight - 4}px; padding-inline: ${token.paddingSM}px; font-size: ${token.fontSize - 1}px; }
        .mongo-filter-shell { min-width: 0; border: 1px solid ${token.colorBorder}; border-radius: ${token.borderRadius}px; overflow: hidden; }
        .mongo-filter-shell textarea { border: 0; border-radius: 0; font-family: ${token.fontFamilyCode}; resize: vertical; }
        .mongo-filter-options { margin-top: 12px; background: transparent; }
        .mongo-filter-options > .ant-collapse-item { overflow: hidden; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .mongo-filter-options > .ant-collapse-item > .ant-collapse-header { align-items: center; gap: 8px; padding: 9px 12px; background: ${token.colorFillQuaternary}; color: ${token.colorTextSecondary}; }
        .mongo-filter-options .ant-collapse-title { display: flex; align-items: center; min-width: 0; min-height: 20px; margin: 0; }
        .mongo-query-options-label { display: flex; align-items: center; gap: 7px; font-size: ${token.fontSizeSM}px; line-height: 20px; }
        .mongo-query-options-label > .anticon { display: flex; align-items: center; justify-content: center; flex: none; width: 1em; height: 1em; color: ${token.colorTextTertiary}; font-size: ${token.fontSize}px; }
        .mongo-query-options-label > .anticon svg { display: block; }
        .mongo-filter-options .ant-collapse-expand-icon { display: inline-flex; align-items: center; justify-content: center; align-self: center; height: 20px; padding: 0; color: ${token.colorTextTertiary}; }
        .mongo-filter-options .ant-collapse-expand-icon .anticon { font-size: 10px; }
        .mongo-filter-options > .ant-collapse-item > .ant-collapse-panel > .ant-collapse-body { padding: 10px 12px 12px; }
        .mongo-query-options-content { --mongo-option-height: 28px; --mongo-option-font-size: 12px; display: flex; flex-direction: column; gap: 10px; min-width: 0; }
        .mongo-query-options-content[data-compact] { --mongo-option-height: 24px; --mongo-option-font-size: 11px; }
        .mongo-query-option-card { display: flex; flex-direction: column; gap: 10px; min-width: 0; padding: 10px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusSM}px; }
        .mongo-query-option-heading { color: ${token.colorTextSecondary}; font-size: var(--mongo-option-font-size); }
        .mongo-query-options-content :is(.ant-input, .ant-select, .ant-btn) { font-size: var(--mongo-option-font-size); }
        .mongo-query-options-content .mongo-order-row { display: grid; grid-template-columns: minmax(0, 1fr) 76px var(--mongo-option-height); gap: 6px; align-items: start; }
        .mongo-query-options-content .mongo-order-row > * { width: 100%; min-width: 0; }
        .mongo-query-options-content .nosql-remove { height: var(--mongo-option-height); padding: 0; }
        .mongo-query-options-content .nosql-add { margin-top: 2px; }
        .mongo-projection-settings { display: grid; grid-template-columns: 100px minmax(0, 1fr); gap: 10px; align-items: start; }
        .mongo-projection-settings > label { display: flex; flex-direction: column; gap: 4px; min-width: 0; font-size: var(--mongo-option-font-size); }
        .mongo-projection-settings .ant-select { width: 100%; min-width: 0; }
        @media (max-width: 399px) {
            .mongo-query-option-card { padding: 8px; gap: 8px; }
            .mongo-query-options-content .mongo-order-row { grid-template-columns: minmax(0, 1fr) 68px var(--mongo-option-height); }
            .mongo-projection-settings { grid-template-columns: minmax(0, 1fr); gap: 8px; }
        }
    `,
}));
