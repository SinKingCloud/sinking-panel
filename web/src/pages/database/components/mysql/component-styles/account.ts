import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    editor: css`
        .ant-modal-body { min-width: 0; font-size: ${token.fontSize}px; }
        .ant-modal-body .ant-btn, .ant-form-item-label > label, .ant-input, .ant-input-number, .ant-select, .ant-radio-wrapper { font-size: ${token.fontSize}px; }
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .ant-form-item-control, .ant-form-item-control-input-content { min-width: 0; }
        .ant-form-item-extra, .ant-form-item-explain { font-size: ${token.fontSizeSM}px; }
        .account-hint { margin-top: -${token.marginXS}px; margin-bottom: ${token.margin}px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .account-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 12px; }
        .account-grid > .ant-form-item { min-width: 0; }
        .account-presets .ant-select-item-option-content { white-space: normal; }
        .account-network { background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        .account-network .ant-collapse-header { font-size: ${token.fontSizeSM}px; }
        .account-network .ant-collapse-body { padding-top: 0; }
        .account-network-list { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 6px 12px; margin-bottom: 9px; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .account-network-list dt { color: ${token.colorTextSecondary}; }
        .account-network-list dd { margin: 0; color: ${token.colorText}; overflow-wrap: anywhere; }
        .account-network-note { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .account-permission-fields { display: flex; flex-direction: column; gap: ${token.marginXS}px; min-width: 0; }
        .account-permission-select .ant-select-selection-overflow { flex-wrap: nowrap; }
        .account-permission-summary { display: block; box-sizing: border-box; min-width: 0; max-width: 100%; padding-inline: ${token.paddingXS}px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .account-permission-description { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .account-permission-list { display: flex; flex-direction: column; gap: ${token.marginSM}px; min-width: 0; }
        .account-permission-list + .account-hint { margin-top: ${token.marginSM}px; margin-bottom: 0; }
        .account-loading { display: flex; align-items: center; justify-content: center; padding: ${token.paddingLG}px 0; }
        .account-permission-heading, .account-permission-item { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) ${token.controlHeight}px; align-items: start; gap: ${token.marginXS}px; min-width: 0; }
        .account-permission-heading { color: ${token.colorTextSecondary}; font-size: ${token.fontSize}px; }
        .account-permission-item > .ant-form-item { min-width: 0; margin-bottom: 0; }
        .account-permission-item > .ant-btn { width: ${token.controlHeight}px; height: ${token.controlHeight}px; padding: 0; color: ${token.colorTextSecondary}; }
        .account-permission-item .ant-select { width: 100%; min-width: 0; }
        .account-permission-item .ant-select-selection-item { max-width: 100%; }
        .account-repair { margin-bottom: 16px; }
        .account-repair .ant-alert-title { font-size: ${token.fontSize}px; }
        .account-repair .ant-alert-description { font-size: ${token.fontSizeSM}px; }
        @media (max-width: 575px) { .account-grid { grid-template-columns: 1fr; } }
        @media (max-width: 399px) {
            .account-permission-heading, .account-permission-item { grid-template-columns: minmax(0, 1fr) minmax(110px, 1fr) ${token.controlHeight}px; column-gap: ${token.marginXXS}px; }
        }
    `,
    permissionPopup: css`
        max-width: calc(100vw - 32px);
        .account-permission-presets { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: ${token.marginXXS}px; padding: ${token.paddingXXS}px; margin-bottom: ${token.marginXXS}px; }
        .account-permission-presets:has(> .ant-btn:nth-child(4)) { grid-template-columns: repeat(2, minmax(0, 1fr)); }
        .account-permission-presets .ant-btn { padding-inline: ${token.paddingXXS}px; font-size: ${token.fontSizeSM}px; }
    `,
}));
