import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact = false}: {compact?: boolean} = {}) => ({
    panel: css`
        display:flex; flex-direction:column; flex:1; min-height:0;
        .account-toolbar { flex:none; }
        .client-footer { flex:none; }

        min-width: 0;
        font-size: ${token.fontSize}px;
        .account-toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; min-width: 0; margin-bottom: 12px; }
        .account-search { flex: 0 1 240px; width: 240px; min-width: 0; max-width: 100%; }
        &&& .account-search.ant-input-affix-wrapper {
            height: ${compact ? 24 : 28}px;
            padding-inline: ${compact ? 8 : 10}px;
            border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
        }
        &&& .account-search .ant-input { color: ${token.colorTextSecondary}; font-size: inherit; }
        .account-search .ant-input::placeholder { color: ${token.colorTextQuaternary}; }
        .account-search .ant-input-prefix { margin-inline-end: 8px; color: ${token.colorTextTertiary}; font-size: ${compact ? 12 : 14}px; }
        .account-database { min-width: 145px; max-width: 240px; }
        &&& .account-database.ant-select { height: ${compact ? 24 : 28}px; border-radius: ${token.borderRadiusSM}px; color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; }
        .account-database .ant-select-suffix { color: ${token.colorTextQuaternary}; font-size: ${compact ? 9 : 10}px; }
        .account-tools { display: flex; flex: none; flex-wrap: wrap; align-items: center; gap: 6px; max-width: 100%; margin-left: auto; }
        .account-tools > * { flex: none; }
        .account-context { display: flex; justify-content: space-between; flex-wrap: wrap; gap: 6px; margin-bottom: 10px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .account-name { display: flex; align-items: center; gap: 7px; min-width: 0; }
        .account-name > span:first-child { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .account-protected { flex: none; color: ${token.colorTextTertiary}; }
        .account-permission { display: flex; align-items: center; gap: 7px; }
        .account-readonly { color: ${token.colorTextSecondary}; }
        .account-empty { min-height: 280px; display: grid; place-items: center; }
        .account-notice { margin-bottom: 12px; }
        .account-notice .ant-alert-title { font-size: ${token.fontSize}px; }
        .account-notice .ant-alert-description { font-size: ${token.fontSizeSM}px; }
        .ant-table, .ant-btn { font-size: ${token.fontSize}px; }
        @media (max-width: 575px) {
            .account-search { flex: 1; min-width: 110px; }
            .account-database { min-width: 130px; max-width: 48%; }
            .account-tools { width: 100%; }
            .account-tools > .ant-btn:last-child { margin-left: auto; }
            .account-empty { min-height: 230px; }
        }
    `,
    editor: css`
        .ant-modal-body { min-width: 0; font-size: ${token.fontSize}px; }
        .ant-modal-body .ant-btn, .ant-form-item-label > label, .ant-input, .ant-input-number, .ant-select, .ant-radio-wrapper { font-size: ${token.fontSize}px; }
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .ant-form-item-control, .ant-form-item-control-input-content { min-width: 0; }
        .ant-form-item-extra, .ant-form-item-explain { font-size: ${token.fontSizeSM}px; }
        .account-loading { display: flex; align-items: center; justify-content: center; padding: ${token.paddingLG}px 0; }
        .account-form-intro { display: flex; flex-wrap: wrap; gap: 5px 8px; padding: 9px 11px; margin-bottom: ${token.margin}px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; font-size: ${token.fontSizeSM}px; color: ${token.colorTextSecondary}; }
        .account-form-intro b { font-weight: 500; color: ${token.colorText}; overflow-wrap: anywhere; }
        .account-hint { margin-top: -${token.marginXS}px; margin-bottom: ${token.margin}px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .account-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 12px; }
        .account-grid > .ant-form-item { min-width: 0; }
        .account-presets .ant-select-item-option-content { white-space: normal; }
        .account-network, .account-custom { background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        .account-network .ant-collapse-header, .account-custom .ant-collapse-header { font-size: ${token.fontSizeSM}px; }
        .account-network .ant-collapse-body, .account-custom .ant-collapse-body { padding-top: 0; }
        .account-network-list { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 6px 12px; margin-bottom: 9px; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .account-network-list dt { color: ${token.colorTextSecondary}; }
        .account-network-list dd { margin: 0; color: ${token.colorText}; overflow-wrap: anywhere; }
        .account-network-note { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .account-custom .ant-form-item { margin-bottom: 0; }
        .account-repair { margin-bottom: 16px; }
        .account-repair .ant-alert-title { font-size: ${token.fontSize}px; }
        .account-repair .ant-alert-description { font-size: ${token.fontSizeSM}px; }
        @media (max-width: 575px) { .account-grid { grid-template-columns: 1fr; } }
    `,
}));
