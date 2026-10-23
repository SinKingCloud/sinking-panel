import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    panel: css`
        min-width: 0;
        font-size: ${token.fontSize}px;
        color: ${token.colorText};
        .ant-btn, .ant-input, .ant-input-number, .ant-select, .ant-form-item-label > label, .ant-checkbox-wrapper { font-size: ${token.fontSize}px; }
        .ant-btn { white-space: nowrap; flex-shrink: 0; }
        .ant-form-item-extra, .ant-form-item-explain { font-size: ${token.fontSizeSM}px; }
        .nosql-topbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding-bottom: 12px; margin-bottom: 16px; border-bottom: 1px solid ${token.colorBorderSecondary}; }
        .nosql-sections { display: flex; align-items: center; gap: 4px; min-width: 0; overflow-x: auto; }
        .nosql-sections .ant-btn { flex: none; color: ${token.colorTextSecondary}; }
        .nosql-sections .ant-btn.active { color: ${token.colorPrimary}; background: ${token.colorPrimaryBg}; font-weight: 500; }
        .nosql-advanced { flex: none; }
        .nosql-body { display: grid; grid-template-columns: 212px minmax(0, 1fr); gap: 16px; min-height: 420px; }
        .nosql-body.full { display: block; }
        .nosql-catalog { display: flex; flex-direction: column; gap: 10px; min-width: 0; align-self: start; padding: 12px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        .nosql-catalog-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; font-weight: 500; }
        .nosql-catalog-head > span { display: flex; align-items: center; gap: 7px; }
        .nosql-catalog > .client-navigation { min-height: 210px; max-height: 410px; }
        .nosql-catalog .client-nav-item { padding: 9px 7px; color: ${token.colorText}; font-size: ${token.fontSize}px; }
        .nosql-catalog .client-nav-item.active { color: ${token.colorPrimary}; }
        .nosql-kind { flex: none; font-size: ${token.fontSizeSM}px; color: ${token.colorTextSecondary}; background: ${token.colorFillTertiary}; border-radius: ${token.borderRadiusSM}px; padding: 2px 6px; white-space: nowrap; }
        .nosql-key-name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .nosql-catalog-footer { display: flex; justify-content: space-between; align-items: center; gap: 5px; padding-top: 7px; font-variant-numeric: tabular-nums; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .nosql-catalog-footer .ant-btn { padding-inline: 7px; }
        .nosql-caption { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .nosql-main { min-width: 0; }
        .nosql-heading { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; min-height: 32px; margin-bottom: 12px; }
        .nosql-heading-main { display: flex; align-items: center; gap: 8px; min-width: 0; flex: 1; }
        .nosql-heading-main > strong { font-weight: 500; color: ${token.colorText}; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .nosql-actions { display: flex; align-items: center; justify-content: flex-end; gap: 4px; flex-wrap: wrap; }
        .nosql-data-card { overflow: hidden; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .nosql-object-bar { display: flex; align-items: flex-start; justify-content: space-between; gap: 10px; padding: 12px; background: ${token.colorFillQuaternary}; }
        .nosql-object-title { min-width: 0; flex: 1; }
        .nosql-object-title > .ant-typography { margin: 0; font-weight: 500; display: block; max-width: 100%; overflow-wrap: anywhere; font-size: ${token.fontSize}px; }
        .nosql-object-meta { display: flex; flex-wrap: wrap; gap: 5px 12px; margin-top: 6px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .nosql-data-content { padding: 12px; }
        .nosql-content-heading { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-bottom: 12px; }
        .nosql-content-heading > span:first-child { font-weight: 500; }
        .nosql-editing-label { color: ${token.colorPrimary}; font-size: ${token.fontSizeSM}px; }
        .nosql-value { min-height: 160px; }
        .nosql-data-footer { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; padding: 12px; border-top: 1px solid ${token.colorBorderSecondary}; }
        .nosql-readonly { display: block; color: ${token.colorTextSecondary}; margin-bottom: 10px; font-size: ${token.fontSizeSM}px; }
        .nosql-empty { display: grid; place-items: center; min-height: 260px; }
        .nosql-transfer { margin-top: 10px; }
        .nosql-transfer .ant-collapse-header { padding: 8px 0 !important; font-size: ${token.fontSize}px; color: ${token.colorTextSecondary} !important; }
        .nosql-transfer .ant-collapse-body { padding: 4px 0 8px !important; }
        .nosql-query-form { padding: 16px; margin-bottom: 16px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        .nosql-query-form .ant-form-item { margin-bottom: 14px; }
        .nosql-query-form .ant-form-item:last-child { margin-bottom: 0; }
        .nosql-query-submit { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: 14px; color: ${token.colorTextSecondary}; }
        .nosql-query-submit > span { font-size: ${token.fontSizeSM}px; }
        .nosql-result-heading { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-bottom: 10px; color: ${token.colorTextSecondary}; }
        .nosql-output { margin: 0; padding: 12px; min-height: 165px; max-height: 340px; overflow: auto; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; font-family: ${token.fontFamilyCode}; font-size: ${token.fontSize}px; line-height: 1.65; color: ${token.colorText}; white-space: pre-wrap; overflow-wrap: anywhere; }
        .nosql-stats { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; margin-bottom: 16px; }
        .nosql-stat { padding: 14px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; min-width: 0; }
        .nosql-stat > span { display: block; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; margin-bottom: 6px; }
        .nosql-stat > strong { display: block; font-weight: 500; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
        .nosql-info-sections { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
        .nosql-info-group { min-width: 0; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; overflow: hidden; }
        .nosql-info-group > strong { display: block; font-weight: 500; padding: 10px 12px; background: ${token.colorFillQuaternary}; }
        .nosql-info-row { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; padding: 9px 12px; }
        .nosql-info-row + .nosql-info-row { border-top: 1px solid ${token.colorBorderSecondary}; }
        .nosql-info-row > span { color: ${token.colorTextSecondary}; overflow-wrap: anywhere; }
        .nosql-info-row > b { font-weight: 400; text-align: right; overflow-wrap: anywhere; }
        .nosql-raw-info { margin-top: 14px; }
        .nosql-raw-info .ant-collapse-header { font-size: ${token.fontSize}px; }
        .nosql-query-mode { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
        .nosql-mode-control { display: flex; align-items: center; gap: 8px; }
        .nosql-mode-control > span { color: ${token.colorTextSecondary}; }
        .nosql-expiry-fields.is-timed { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
        .nosql-content-tabs { margin-bottom: 12px; }
        .mongo-document-list { display: flex; flex-direction: column; gap: 10px; max-height: 490px; min-height: 220px; overflow-y: auto; scrollbar-gutter: stable; padding-right: 3px; }
        .mongo-document-card { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; background: ${token.colorBgContainer}; }
        .mongo-document-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 9px 12px; background: ${token.colorFillQuaternary}; }
        .mongo-document-id { display: flex; align-items: center; gap: 8px; min-width: 0; flex: 1; }
        .mongo-document-id > span:first-child { flex: none; font-size: ${token.fontSizeSM}px; color: ${token.colorTextSecondary}; }
        .mongo-document-id > .ant-typography { font-family: ${token.fontFamilyCode}; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin: 0; }
        .mongo-document-actions { display: flex; align-items: center; gap: 3px; flex: none; }
        .mongo-document-actions .ant-btn { padding-inline: 7px; }
        .mongo-document-preview { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 9px 16px; padding: 12px; margin: 0; }
        .mongo-document-preview > div { display: grid; grid-template-columns: minmax(0, 85px) minmax(0, 1fr); gap: 8px; min-width: 0; }
        .mongo-document-preview dt { margin: 0; color: ${token.colorTextSecondary}; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .mongo-document-preview dd { margin: 0; min-width: 0; max-height: 3.3em; overflow: hidden; overflow-wrap: anywhere; white-space: pre-wrap; font-family: ${token.fontFamilyCode}; line-height: 1.65; }
        .mongo-document-preview .is-null { color: ${token.colorTextTertiary}; }
        .mongo-document-more { margin: 0 5px 5px; color: ${token.colorTextSecondary}; }
        .mongo-document-empty { padding: 12px; color: ${token.colorTextTertiary}; }
        .mongo-index-list { display: flex; flex-direction: column; min-height: 240px; }
        .mongo-index-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; padding: 14px 0; border-bottom: 1px solid ${token.colorBorderSecondary}; }
        .mongo-index-row:first-child { padding-top: 5px; }
        .mongo-index-content { min-width: 0; flex: 1; }
        .mongo-index-title { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
        .mongo-index-title strong { font-weight: 500; overflow-wrap: anywhere; }
        .mongo-index-fields { display: flex; gap: 8px; flex-wrap: wrap; margin: 9px 0; }
        .mongo-index-fields > span { display: flex; gap: 7px; padding: 4px 7px; border-radius: ${token.borderRadiusSM}px; background: ${token.colorFillQuaternary}; overflow-wrap: anywhere; }
        .mongo-index-fields small { font-size: ${token.fontSizeSM}px; color: ${token.colorTextSecondary}; white-space: nowrap; }
        .mongo-index-options { padding: 12px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; margin-bottom: 16px; }
        .mongo-index-options .ant-form-item { margin-bottom: 8px; }
        .mongo-index-options .ant-form-item:last-child { margin-bottom: 0; }
        .nosql-advanced-target { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
        .nosql-advanced-target > .ant-select { min-width: 0; width: 260px; max-width: 75%; }
        @media (max-width: 767px) {
            .nosql-topbar { gap: 8px; margin-bottom: 12px; }
            .nosql-sections { gap: 2px; }
            .nosql-sections .ant-btn { padding-inline: 10px; }
            .nosql-body { grid-template-columns: minmax(0, 1fr); gap: 14px; min-height: 360px; }
            .nosql-catalog { padding: 10px; gap: 8px; }
            .nosql-catalog > .client-navigation { min-height: 65px; max-height: 130px; }
            .nosql-catalog-footer { justify-content: flex-end; gap: 10px; }
            .nosql-empty { min-height: 200px; }
            .nosql-query-form { padding: 12px; }
            .mongo-document-preview { grid-template-columns: minmax(0, 1fr); gap: 8px; }
            .mongo-document-head { padding: 8px; }
            .mongo-document-actions { gap: 0; }
            .mongo-document-actions .ant-btn { padding-inline: 5px; }
            .mongo-document-id { gap: 5px; }
            .mongo-document-id > span:first-child { display: none; }
            .mongo-document-list { max-height: 440px; }
        }
        @media (max-width: 575px) {
            .nosql-stats { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; }
            .nosql-info-sections { grid-template-columns: minmax(0, 1fr); }
            .nosql-heading-main { flex-wrap: wrap; }
            .nosql-object-meta { gap: 5px 9px; }
            .nosql-expiry-fields.is-timed { grid-template-columns: minmax(0, 1fr); gap: 0; }
            .nosql-result-heading { flex-wrap: wrap; }
            .nosql-query-mode { align-items: flex-start; }
            .nosql-advanced { padding-inline: 9px; }
        }
    `,
}));
