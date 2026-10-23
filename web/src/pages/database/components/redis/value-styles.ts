import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    editor: css`
        min-width: 0;
        max-width: 100%;
        color: ${token.colorText};
        font-size: ${Math.max(11, token.fontSize - 1)}px;
        .ant-input, .ant-input-number, .ant-table-wrapper, .ant-table { font-size: ${Math.max(11, token.fontSize - 1)}px; }
        &.redis-value-fill { display: flex; flex-direction: column; height: 100%; min-height: 0; }
        &.redis-value-fill .redis-value-text { flex: 1; min-height: 120px; max-height: none; }
        &.redis-value-fill .redis-value-footer { flex: none; margin-top: 4px; }
        .ant-table-wrapper, .ant-spin-nested-loading, .ant-spin-container { min-width: 0; max-width: 100%; }
        .ant-table { color: ${token.colorTextSecondary}; }
        .ant-table-container { border-inline: 1px solid ${token.colorBorderSecondary}; border-bottom: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; overflow: hidden; }
        && .ant-table-tbody > tr:last-child > td,
        && .ant-table-tbody-virtual-holder-inner > .client-last-row > .ant-table-cell { border-bottom: 0; }
        .ant-table-cell { vertical-align: top; }
        .ant-table-thead > tr > th { color: ${token.colorTextSecondary}; font-size: ${token.fontSize}px; font-weight: 500; }
        .ant-table-body, .ant-table-content { overscroll-behavior: contain; }
        .ant-table .ant-input-number { width: 100%; }
        .ant-table .ant-empty { margin-block: 24px; }
        .redis-value-index, .redis-value-score { display: inline-block; padding-top: 5px; font-variant-numeric: tabular-nums; }
        .redis-value-index { color: ${token.colorTextTertiary}; }
        .redis-value-text { display: block; min-height: 120px; max-height: 600px; font-family: ${token.fontFamilyCode}; line-height: 1.65; tab-size: 4; }
        &:not(.redis-value-embedded) .redis-value-text { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorBgContainer}; }
        .redis-value-cell { display: flex; align-items: flex-start; gap: 6px; min-width: 0; }
        .redis-value-cell > .ant-input { flex: 1; min-width: 0; font-family: ${token.fontFamilyCode}; }
        .redis-value-copy { flex: none; width: 16px; padding-top: 5px; line-height: 1.6; }
        .redis-value-copy .ant-typography-copy { margin: 0; }
        .redis-value-preview { margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; font-family: ${token.fontFamilyCode}; line-height: 1.65; }
        .redis-value-preview .ant-typography-expand, .redis-value-preview .ant-typography-collapse { font-family: ${token.fontFamily}; }
        .redis-value-footer { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-top: 10px; }
        .redis-value-footer > span:not(.redis-copy-action) { font-size: ${token.fontSizeSM}px; color: ${token.colorTextSecondary}; }
        &.redis-value-parameters .ant-table-cell { padding-inline: 6px; }
        &.redis-value-parameters .redis-value-footer { margin-top: 8px; }
        &.redis-value-embedded, &.redis-value-embedded .ant-input, &.redis-value-embedded .ant-input-number { font-size: ${token.fontSize}px; }
        .redis-value-entries { display: flex; flex-direction: column; gap: ${token.marginSM}px; min-width: 0; }
        .redis-value-entry { display: flex; flex-direction: column; gap: ${token.marginXS}px; min-width: 0; padding: ${token.paddingSM}px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .redis-value-entry-heading { display: flex; align-items: center; justify-content: space-between; gap: ${token.marginXS}px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .redis-value-entry-field { display: flex; flex-direction: column; gap: ${token.marginXXS}px; min-width: 0; }
        .redis-value-entry-field > span { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .redis-value-entry-field > .ant-input-number { width: 100%; }
        &.redis-value-browser { display: flex; flex-direction: column; min-height: 0; }
        .redis-value-toolbar { display: flex; align-items: center; flex: none; gap: 8px; min-width: 0; margin-bottom: 10px; }
        .redis-value-label { color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; }
        .redis-value-tools { display: flex; align-items: center; flex: none; gap: 6px; margin-left: auto; }
        && .redis-value-search { flex: 0 1 220px; width: 220px; min-width: 0; height: ${compact ? 24 : 28}px; font-size: ${compact ? 11 : 12}px; border-radius: ${token.borderRadiusSM}px; }
        .redis-value-search .ant-input { font-size: inherit; }
        .redis-value-search .ant-input-prefix { margin-inline-end: 6px; color: ${token.colorTextQuaternary}; }
        .redis-value-table-area { flex: 1; min-height: 0; min-width: 0; overflow: hidden; }
        .redis-value-table { width: 100%; }
        && .redis-value-table .ant-table { font-size: ${compact ? 11 : 12}px; }
        && .redis-value-table .ant-table-thead > tr > th { padding: ${compact ? 7 : 8}px 8px; font-size: ${compact ? 11 : 12}px; font-weight: 500; }
        && .redis-value-table :is(.ant-table-tbody, .ant-table-tbody-virtual-holder) .ant-table-cell { padding: 4px 8px; vertical-align: middle; }
        .redis-value-table .ant-table-body, .redis-value-table .ant-table-tbody-virtual-holder { scrollbar-width: thin; }
        .redis-value-number { color: ${token.colorTextTertiary}; font-variant-numeric: tabular-nums; }
        .redis-value-inline { display: flex; align-items: center; height: ${compact ? 34 : 38}px; min-width: 0; }
        .redis-value-display { display: flex; align-items: center; width: 100%; min-width: 0; height: 100%; box-sizing: border-box; padding: 0 8px; border: 1px solid transparent; border-radius: ${token.borderRadiusSM}px; background: transparent; color: ${token.colorTextSecondary}; font: inherit; font-family: ${token.fontFamilyCode}; text-align: left; cursor: pointer; }
        .redis-value-display > span { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; white-space: pre-wrap; overflow-wrap: anywhere; line-height: ${compact ? 16 : 18}px; }
        .redis-value-display em { font-style: normal; color: ${token.colorTextQuaternary}; }
        .redis-value-display:hover:not(:disabled) { background: ${token.colorFillQuaternary}; }
        .redis-value-display:focus-visible { outline: 1px solid ${token.colorPrimaryBorder}; outline-offset: 2px; }
        .redis-value-display:disabled { cursor: default; }
        .redis-value-inline.is-number .redis-value-display { justify-content: flex-end; font-variant-numeric: tabular-nums; }
        && .redis-value-inline > .ant-input, && .redis-value-inline > .ant-input-number { width: 100%; height: ${compact ? 26 : 28}px; min-height: 0; box-sizing: border-box; padding: 4px 8px; border-radius: ${token.borderRadiusSM}px; font: inherit; font-family: ${token.fontFamilyCode}; line-height: ${compact ? 16 : 18}px; }
        && .redis-value-inline > textarea.ant-input { resize: none; }
        && .redis-value-inline > .ant-input-number { padding-block: 0; }
        && .redis-value-inline .ant-input-number-input { height: ${compact ? 24 : 26}px; padding: 0; font: inherit; text-align: right; }
        && .redis-value-table .ant-btn { height: ${compact ? 24 : 28}px; padding-inline: ${compact ? 7 : 9}px; gap: 4px; font-size: ${compact ? 11 : 12}px; border-radius: ${token.borderRadiusSM}px; }
        .redis-value-code { display: flex; flex-direction: column; flex: 1; min-height: 0; min-width: 0; overflow: hidden; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        && .redis-value-code > .redis-value-text { height: 100%; min-height: 0; max-height: none; resize: none; border: 0; border-radius: 0; box-shadow: none; background: ${token.colorBgContainer}; font-size: 12px; }
        @media (max-width: 575px) {
            .redis-value-toolbar { flex-wrap: wrap; gap: 8px; }
            && .redis-value-search { width: 100%; flex-basis: 100%; }
            .redis-value-search + .redis-value-tools { margin-left: 0; }
            .redis-value-table-area { min-height: 140px; }
        }
    `,
}));
