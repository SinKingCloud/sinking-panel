import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact = false}: {compact?: boolean} = {}) => ({
    datePopup: css`
        @media (max-width: 575px) {
            max-width: calc(100vw - 24px);
            .ant-picker-panel-container { max-height: calc(100dvh - 24px); overflow-y: auto; }
            .ant-picker-datetime-panel { flex-direction: column; }
            .ant-picker-datetime-panel .ant-picker-time-panel { border-inline-start: 0; border-top: 1px solid ${token.colorBorderSecondary}; }
            .ant-picker-datetime-panel .ant-picker-time-panel-column { height: 144px; flex: 1; }
        }
    `,
    form: css`
        min-width: 0;
        .es-filter-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: ${token.margin}px; }
        .es-filter-heading .ant-segmented { min-width: 0; }
        .es-filter-heading .ant-btn { flex: none; height: ${token.controlHeight - 4}px; padding-inline: ${token.paddingSM}px; font-size: ${token.fontSize - 1}px; }
        .es-filter-logic { width: 100%; margin-bottom: 14px; }
        .es-filter-labels, .es-filter-row { display: grid; grid-template-columns: minmax(0, 1fr) 112px minmax(0, 1.1fr) ${compact ? 28 : 32}px; gap: 8px; }
        .es-filter-labels { margin-bottom: 6px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .es-filter-row { align-items: start; margin-bottom: 10px; }
        .es-filter-row > *, .es-filter-value > * { width: 100%; min-width: 0; }
        .es-filter-row .ant-select-content { min-width: 0; }
        .es-filter-row .ant-select-selection-item { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .es-filter-value { display: flex; flex-direction: column; gap: 6px; }
        .es-filter-remove { width: ${compact ? 28 : 32}px; padding-inline: 0; }
        .es-filter-date { display: flex; align-items: stretch; gap: 4px; }
        .es-filter-date > .ant-input { flex: 1; width: 0; min-width: 0; }
        .es-filter-date > .ant-picker { flex: none; width: ${compact ? 28 : 32}px; padding-inline: 7px; }
        .es-filter-empty { padding: 18px 8px 22px; color: ${token.colorTextTertiary}; text-align: center; font-size: ${token.fontSizeSM}px; }
        .es-filter-hint { margin-top: 10px; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; }
        .es-filter-preview { margin-top: ${token.margin}px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; overflow: hidden; }
        .es-filter-preview-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 7px 10px; background: ${token.colorFillQuaternary}; }
        .es-filter-preview-toggle { padding: 0; border: 0; background: transparent; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; cursor: pointer; }
        .es-filter-preview-heading .ant-btn { height: 24px; font-size: ${token.fontSizeSM}px; }
        .es-filter-preview pre { margin: 0; padding: 12px; white-space: pre-wrap; overflow-wrap: anywhere; color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; line-height: 1.7; }
        .es-filter-preview-error { padding: 12px; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; }
        .es-filter-dsl { margin-bottom: 0; }
        .es-filter-dsl .ant-form-item-label > label { width: 100%; height: auto; }
        .es-filter-dsl-heading { display: flex; align-items: center; justify-content: space-between; width: 100%; gap: 8px; }
        .es-filter-dsl-heading .ant-btn { height: 24px; font-size: ${token.fontSizeSM}px; color: ${token.colorTextSecondary}; }
        .es-filter-dsl-editor { height: ${compact ? 220 : 240}px; min-width: 0; overflow: hidden; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .es-filter-dsl-fallback { height: 100%; border: 0; border-radius: 0; resize: none; font-family: ${token.fontFamilyCode}; font-size: ${compact ? 12 : 13}px; }
        @media (max-width: 575px) {
            .es-filter-labels { display: none; }
            .es-filter-row { grid-template-columns: minmax(0, 1fr) 108px ${compact ? 28 : 32}px; padding-bottom: 12px; margin-bottom: 12px; border-bottom: 1px solid ${token.colorBorderSecondary}; }
            .es-filter-value { grid-column: 1 / 3; grid-row: 2; }
            .es-filter-remove { grid-column: 3; grid-row: 1; }
            .es-filter-heading { gap: 8px; }
        }
    `,
}));
