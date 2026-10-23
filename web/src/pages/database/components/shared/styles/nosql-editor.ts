import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    editor: css`
        display: flex; flex-direction: column; gap: 8px; min-width: 0; font-size: ${token.fontSize}px;
        .ant-btn, .ant-input, .ant-input-number, .ant-select, .ant-picker { font-size: ${token.fontSize}px; }
        .ant-btn { white-space: nowrap; flex-shrink: 0; }
        .nosql-cell { flex: 1; min-width: 0; }
        .nosql-score { flex: 0 0 100px; width: 100px; }
        .nosql-remove { flex: none; width: 28px; padding-inline: 0; }
        .nosql-add { align-self: flex-start; }
        .nosql-scroll { display: flex; flex-direction: column; gap: 8px; max-height: 330px; overflow: auto; scrollbar-gutter: stable; }
        .nosql-empty { color: ${token.colorTextTertiary}; }
        .nosql-editor-empty { padding: 18px 12px; color: ${token.colorTextSecondary}; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; line-height: 1.65; }
        .nosql-text-value { margin: 0; padding: 10px; min-height: 145px; max-height: 330px; overflow: auto; white-space: pre-wrap; overflow-wrap: anywhere; font-family: ${token.fontFamilyCode}; font-size: inherit; line-height: 1.65; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; color: ${token.colorText}; }
        .nosql-value-copy { align-self: flex-end; font-size: ${token.fontSizeSM}px; color: ${token.colorTextSecondary}; }
        .nosql-member-view { min-width: 0; max-height: 330px; overflow: auto; scrollbar-gutter: stable; }
        .nosql-member-heading, .nosql-member-row { display: grid; grid-template-columns: 28px minmax(0, 1fr); gap: 8px; padding: 9px 7px; min-width: 0; }
        .nosql-member-heading { color: ${token.colorTextSecondary}; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadiusSM}px; font-size: ${token.fontSizeSM}px; }
        .nosql-member-row + .nosql-member-row { border-top: 1px solid ${token.colorBorderSecondary}; }
        .nosql-member-view.hash .nosql-member-heading, .nosql-member-view.hash .nosql-member-row { grid-template-columns: 28px minmax(80px, 1fr) minmax(0, 2fr); }
        .nosql-member-view.sorted .nosql-member-heading, .nosql-member-view.sorted .nosql-member-row { grid-template-columns: 28px minmax(0, 1fr) 80px; }
        .nosql-member-key, .nosql-member-content { min-width: 0; white-space: pre-wrap; overflow-wrap: anywhere; font-family: ${token.fontFamilyCode}; }
        .nosql-member-score { text-align: right; font-variant-numeric: tabular-nums; }
        .nosql-row-number { color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; font-variant-numeric: tabular-nums; }
        .nosql-entry { display: flex; align-items: flex-start; gap: 8px; min-width: 0; }
        .nosql-entry-number { width: 22px; flex: none; text-align: center; padding-top: 6px; color: ${token.colorTextTertiary}; font-variant-numeric: tabular-nums; }
        .nosql-entry-heading { display: flex; align-items: center; gap: 8px; padding-left: 30px; padding-right: 36px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .nosql-entry-heading > span { flex: 1; min-width: 0; }
        .nosql-entry-heading > .nosql-score { flex: 0 0 100px; }
        .nosql-field { padding: 10px; min-width: 0; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .nosql-field.is-readonly { background: ${token.colorFillQuaternary}; border: 0; }
        .mongo-field-line, .mongo-field-heading { display: grid; grid-template-columns: minmax(115px, 1fr) 106px minmax(130px, 1.35fr) 28px; gap: 8px; align-items: start; }
        .mongo-field-heading { padding-inline: 10px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .mongo-field-heading.is-readonly, .is-readonly .mongo-field-line { grid-template-columns: minmax(115px, 1fr) 106px minmax(130px, 1.35fr); }
        .mongo-field-name { display: flex; gap: 6px; align-items: center; min-width: 0; }
        .mongo-field-name > .ant-input { min-width: 0; }
        .mongo-field-name > span:last-child { min-width: 0; overflow-wrap: anywhere; }
        .mongo-field-kind { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; padding-top: 2px; }
        .mongo-field-type { width: 100%; min-width: 0; }
        .mongo-field-content { min-width: 0; }
        .mongo-field-lock { padding-top: 7px; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; }
        .mongo-read-value { display: block; font-family: ${token.fontFamilyCode}; white-space: pre-wrap; overflow-wrap: anywhere; line-height: 1.65; }
        .mongo-null-value { color: ${token.colorTextTertiary}; }
        .mongo-nested-toggle { padding-inline: 0; height: auto; text-align: left; }
        .nosql-nested { padding: 12px 0 0 8px; margin-top: 8px; border-left: 2px solid ${token.colorBorderSecondary}; min-width: 0; }
        .nosql-raw { margin: 0; max-height: 150px; overflow: auto; white-space: pre-wrap; overflow-wrap: anywhere; color: ${token.colorTextSecondary}; font-family: ${token.fontFamilyCode}; font-size: ${token.fontSize}px; }
        .mongo-order-row { display: flex; align-items: center; gap: 8px; }
        .mongo-order-row > .nosql-row-number { width: 16px; flex: none; }
        .mongo-order-direction { width: 112px; flex: none; }
        .mongo-order-heading { display: flex; justify-content: space-between; padding-inline: 24px 36px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .mongo-order-heading > span:last-child:not(:first-child) { width: 112px; }
        @media (max-width: 625px) {
            .mongo-field-line, .mongo-field-heading { grid-template-columns: minmax(0, 1fr) 106px 28px; gap: 8px; }
            .mongo-field-heading > span:nth-child(3) { display: none; }
            .mongo-field-content { grid-column: 1 / -1; grid-row: 2; }
            .mongo-field-line > .nosql-remove, .mongo-field-lock { grid-column: 3; grid-row: 1; }
            .mongo-field-heading.is-readonly, .is-readonly .mongo-field-line { grid-template-columns: minmax(0, 1fr) 106px; }
            .nosql-nested { padding-left: 6px; }
            .nosql-entry-number { width: 18px; }
            .nosql-entry-heading { padding-left: 26px; }
            .nosql-score { flex-basis: 75px; width: 75px; }
            .nosql-entry-heading > .nosql-score { flex-basis: 75px; }
            .mongo-order-direction { width: 100px; }
            .mongo-order-row { gap: 6px; }
            .nosql-member-view.hash .nosql-member-heading, .nosql-member-view.hash .nosql-member-row { grid-template-columns: 22px minmax(65px, 1fr) minmax(0, 1.5fr); }
        }
    `,
    query: css`
        margin-bottom: 12px; font-size: ${token.fontSize}px;
        .nosql-query-summary { display: flex; justify-content: space-between; gap: 10px; align-items: center; }
        .nosql-query-summary > span { color: ${token.colorTextSecondary}; display: flex; align-items: center; gap: 7px; font-size: ${token.fontSizeSM}px; }
        .nosql-query-summary > div { display: flex; align-items: center; gap: 5px; flex: none; }
        .ant-btn { font-size: ${token.fontSize}px; white-space: nowrap; }
        @media (max-width: 575px) { .nosql-query-summary { flex-wrap: wrap; gap: 8px; } }
    `,
    queryModal: css`
        .ant-modal-body { font-size: ${token.fontSize}px; }
        .nosql-query-section-head { display: flex; justify-content: space-between; gap: 10px; align-items: center; margin-bottom: 12px; }
        .nosql-query-section-head strong { font-weight: 500; }
        .nosql-query-section-head .ant-typography { font-size: ${token.fontSizeSM}px; }
        .mongo-filter-card { padding: 12px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        .mongo-filter-heading { display: flex; align-items: center; justify-content: space-between; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; margin-bottom: 8px; }
        .mongo-filter-heading .ant-btn { width: 24px; height: 24px; padding: 0; }
        .mongo-filter-fields { display: grid; grid-template-columns: minmax(0, 1fr) 120px; gap: 8px; margin-bottom: 8px; }
        .mongo-filter-value { display: grid; grid-template-columns: 106px minmax(0, 1fr); gap: 8px; }
        .mongo-filter-value > div { min-width: 0; }
        .mongo-query-options { margin-top: 14px; }
        .mongo-query-options .ant-collapse-header { font-size: ${token.fontSize}px; padding-inline: 0 !important; }
        .mongo-query-options .ant-collapse-body { padding-inline: 0 !important; }
        .mongo-projection-mode { width: 190px; margin-bottom: 12px; }
        @media (max-width: 575px) {
            .nosql-query-section-head { flex-wrap: wrap; gap: 5px; }
            .mongo-filter-card { padding: 10px; }
            .mongo-filter-fields { grid-template-columns: minmax(0, 1fr) 112px; }
            .mongo-filter-value { grid-template-columns: minmax(0, 1fr); }
            .mongo-filter-value > .ant-select { width: 106px; }
        }
    `,
}));
