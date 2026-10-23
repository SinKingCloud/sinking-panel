import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    editor: css`
        display: flex; flex-direction: column; gap: ${token.marginSM}px; min-width: 0;
        .mongo-document-field { min-width: 0; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorBgContainer}; }
        .mongo-document-field-header { display: grid; grid-template-columns: minmax(0, 1fr) 126px ${token.controlHeight}px; align-items: center; gap: 8px; padding: ${token.paddingSM}px; padding-bottom: 0; }
        .mongo-document-field-name { display: flex; align-items: center; gap: 8px; min-width: 0; }
        .mongo-document-field-name > .ant-input { min-width: 0; flex: 1; }
        .mongo-document-field-name > span:not(.mongo-document-field-number) { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .mongo-document-field-number { flex: none; min-width: 14px; text-align: center; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; font-variant-numeric: tabular-nums; }
        .mongo-document-field-type { width: 100%; min-width: 0; }
        .mongo-document-field-remove { width: ${token.controlHeight}px; padding: 0; color: ${token.colorTextTertiary}; }
        .mongo-document-field-remove:hover { color: ${token.colorError} !important; background: ${token.colorErrorBg} !important; }
        .mongo-document-field-lock { color: ${token.colorTextTertiary}; text-align: center; font-size: ${token.fontSizeSM}px; white-space: nowrap; }
        .mongo-document-field-content { min-width: 0; padding: ${token.paddingSM}px; padding-top: 8px; }
        .mongo-document-element-name { color: ${token.colorTextSecondary}; }
        .mongo-document-nested-toggle { display: flex; align-items: center; width: 100%; padding-inline: 8px; background: ${token.colorFillQuaternary}; color: ${token.colorTextSecondary}; text-align: left; }
        .mongo-document-nested-toggle .ant-btn-icon { font-size: ${token.fontSizeSM}px; }
        .mongo-document-child-count { margin-left: auto; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; }
        .mongo-document-nested { padding-top: ${token.paddingSM}px; }
        .mongo-document-field.is-readonly > .mongo-document-field-header { grid-template-columns: minmax(0, 1fr) auto; gap: 12px; }
        .mongo-document-field-kind { color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; white-space: nowrap; }
        .mongo-document-editor-empty { padding-block: ${token.paddingSM}px; color: ${token.colorTextTertiary}; text-align: center; font-size: ${token.fontSizeSM}px; }
        .mongo-document-field-add { width: 100%; }
        @media (max-width: 575px) {
            .mongo-document-field-header { grid-template-columns: minmax(0, 1fr) 106px ${token.controlHeight}px; gap: 6px; padding: 10px; padding-bottom: 0; }
            .mongo-document-field-name { gap: 6px; }
            .mongo-document-field-content { padding: 10px; padding-top: 8px; }
            .mongo-document-field-number { display: none; }
            .mongo-document-nested { padding-top: 8px; }
            .mongo-document-nested .mongo-document-field-header { grid-template-columns: minmax(0, 1fr) 98px ${token.controlHeight}px; }
        }
        @media (max-width: 359px) {
            .mongo-document-field-header { grid-template-columns: minmax(0, 1fr) ${token.controlHeight}px; }
            .mongo-document-field-name { grid-column: 1; }
            .mongo-document-field-remove, .mongo-document-field-lock { grid-column: 2; grid-row: 1; }
            .mongo-document-field-type, .mongo-document-field-type-label { grid-column: 1 / -1; }
            .mongo-document-nested .mongo-document-field-header { grid-template-columns: minmax(0, 1fr) ${token.controlHeight}px; }
        }
    `,
    value: css`
        min-width: 0; width: 100%;
        > :is(.ant-input, .ant-select, .ant-input-number, .ant-picker) { width: 100%; min-width: 0; }
        .ant-picker-input { min-width: 0; }
        .mongo-document-null, .mongo-document-read-value, .mongo-document-raw { margin: 0; padding: ${token.paddingXS}px ${token.paddingSM}px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; color: ${token.colorText}; line-height: ${token.lineHeight}; overflow-wrap: anywhere; }
        .mongo-document-null { display: flex; align-items: center; gap: 8px; min-height: ${token.controlHeight}px; box-sizing: border-box; color: ${token.colorTextSecondary}; }
        .mongo-document-null > span, .mongo-document-empty-value { color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; }
        .mongo-document-read-value { min-height: ${token.controlHeight}px; box-sizing: border-box; white-space: pre-wrap; }
        .mongo-document-raw { font-family: ${token.fontFamilyCode}; font-size: ${token.fontSizeSM}px; white-space: pre-wrap; }
    `,
}));
