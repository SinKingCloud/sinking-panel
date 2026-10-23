import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    panel: css`
        display: flex; flex-direction: column; height: 100%; min-height: 0; min-width: 0;
        color: ${token.colorText}; font-size: ${token.fontSize}px; line-height: 1.5;
        .redis-account-toolbar { display: flex; align-items: center; gap: 8px; flex: none; margin-bottom: 12px; }
        .redis-account-search { flex: 0 1 240px; width: 240px; min-width: 0; max-width: 100%; }
        .redis-account-toolbar > .ant-btn { flex: none; }
        .redis-account-toolbar > .ant-btn:not(.redis-account-refresh) { margin-left: auto; }
        .redis-account-table { flex: 1; min-height: 0; }
        .redis-account-table .ant-table { color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; }
        .redis-account-table .ant-table-container { border-inline: 1px solid ${token.colorBorderSecondary}; border-bottom: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; overflow: hidden; }
        && .redis-account-table .ant-table-tbody > tr:last-child > td,
        && .redis-account-table .ant-table-tbody-virtual-holder-inner > .client-last-row > .ant-table-cell { border-bottom: 0; }
        .redis-account-table .ant-table-thead > tr > th { color: ${token.colorTextSecondary}; font-size: ${compact ? 12 : 13}px; font-weight: 500; }
        .redis-account-table .ant-table-cell { padding: ${compact ? 6 : 8}px ${compact ? 8 : 10}px; }
        && .redis-account-table .ant-btn { height: ${compact ? 22 : 24}px; padding-inline: ${compact ? 6 : 7}px; font-size: ${compact ? 11 : 12}px; }
        .redis-account-table .ant-table-body, .redis-account-table .ant-table-tbody-virtual-holder { scrollbar-width: none; }
        .redis-account-table .ant-table-body::-webkit-scrollbar, .redis-account-table .ant-table-tbody-virtual-scrollbar { display: none; }
        .redis-account-name { display: flex; align-items: center; gap: 7px; min-width: 0; }
        .redis-account-name > span { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
        .redis-account-lock { flex: none; color: ${token.colorTextTertiary}; }
        .redis-account-protected { color: ${token.colorTextTertiary}; }
        .redis-account-scope { display: flex; align-items: center; justify-content: space-between; flex: none; gap: 8px; padding-top: 10px; color: ${token.colorTextTertiary}; font-size: ${Math.max(11, token.fontSize - 1)}px; }
        @media (max-width: 767px) {
            .redis-account-toolbar { flex-wrap: wrap; }
            .redis-account-search { flex-basis: 100%; width: 100%; }
            .redis-account-toolbar > .ant-btn:not(.redis-account-refresh) { margin-left: 0; }
            .redis-account-refresh { margin-left: auto; }
        }
    `,
    editor: css`
        .ant-modal-body { min-width: 0; font-size: ${token.fontSize}px; line-height: 1.6; }
        .ant-modal-body .ant-btn, .ant-form-item-label > label, .ant-input, .ant-input-number, .ant-select, .ant-radio-wrapper, .ant-checkbox-wrapper { font-size: ${token.fontSize}px; }
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .ant-form > .ant-form-item:last-child { margin-bottom: 0; }
        .ant-form-item-control, .ant-form-item-control-input-content { min-width: 0; }
        .ant-form-item-extra, .ant-form-item-explain { font-size: ${token.fontSizeSM}px; }
    `,
}));
