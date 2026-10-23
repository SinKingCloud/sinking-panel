import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact = false}: {compact?: boolean} = {}) => ({
    modal: css`
        .ant-modal-body { overflow: visible; }
    `,
    form: css`
        min-width: 0;
        .es-mapping-loading { display: flex; align-items: center; justify-content: center; padding-block: 64px; }
        .es-mapping-note { margin-bottom: ${token.margin}px; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .es-mapping-fields { margin-bottom: 0; }
        .es-mapping-fields .es-mapping-note { margin-bottom: ${token.marginSM}px; }
        .es-mapping-field-list { display: grid; gap: ${token.marginSM}px; }
        .es-mapping-field-list:not(:empty) { margin-bottom: ${token.marginSM}px; }
        .es-mapping-field { min-width: 0; padding: ${compact ? 10 : 12}px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; }
        .es-mapping-field-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: ${token.marginSM}px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .es-mapping-field-heading .ant-btn { flex: none; }
        .es-mapping-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: ${token.marginSM}px 12px; }
        .es-mapping-grid .ant-form-item { min-width: 0; margin-bottom: 0; }
        .es-mapping-grid .ant-select { width: 100%; }
        .es-mapping-grid .ant-select-content { min-width: 0; }
        .es-mapping-grid .ant-select-selection-item { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        @media (max-width: 575px) {
            .es-mapping-grid { grid-template-columns: minmax(0, 1fr); }
        }
    `,
}));
