import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    status: css`
        display: flex; flex-direction: column; height: 100%; min-height: 0; min-width: 0;
        color: ${token.colorText}; font-size: ${compact ? 12 : 13}px; line-height: 1.6;
        .redis-status-toolbar { display: flex; flex: none; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; min-width: 0; }
        .redis-status-tools { display: flex; flex: none; align-items: center; gap: 6px; margin-left: auto; }
        .redis-status-tools > * { flex: none; }
        .redis-status-search {
            display: flex; align-items: center; flex: 0 1 240px; width: 240px; max-width: 100%; min-width: 0;
            height: ${compact ? 24 : 28}px; border: 1px solid transparent; border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorFillTertiary}; transition: background-color .16s ease, border-color .16s ease;
        }
        .redis-status-search:hover { background: ${token.colorFillSecondary}; }
        .redis-status-search:focus-within { border-color: ${token.colorPrimary}; background: ${token.colorBgContainer}; }
        &&& .redis-status-search > :is(.ant-select, .ant-input-affix-wrapper) {
            border: 0; border-radius: 0; background: transparent; box-shadow: none;
            color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px;
        }
        &&& .redis-status-search > .ant-select {
            flex: none; align-items: center; width: max-content; min-width: ${compact ? 54 : 60}px;
            max-width: min(46%, 120px); height: 100%; padding-block: 0; padding-inline: ${compact ? 7 : 8}px 5px;
        }
        .redis-status-search-label { display: block; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .redis-status-search .ant-select-suffix { color: ${token.colorTextQuaternary}; font-size: ${compact ? 9 : 10}px; }
        &&& .redis-status-search > .ant-input-affix-wrapper {
            flex: 1; width: 0; min-width: 0; height: calc(100% - 8px); padding-block: 0; padding-inline: ${compact ? 8 : 10}px;
            border-inline-start: 1px solid ${token.colorBorderSecondary};
        }
        &&& .redis-status-search .ant-input { height: auto; padding: 0; font-size: inherit; color: inherit; }
        .redis-status-search .ant-input::placeholder { color: ${token.colorTextQuaternary}; }
        .redis-status-search .ant-input-suffix { margin-inline-start: 4px; }
        .redis-status-summary { display: grid; flex: none; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; margin-bottom: ${compact ? 14 : 18}px; }
        .redis-status-summary > div { display: flex; flex-direction: column; min-width: 0; gap: 6px; padding: ${compact ? 10 : 12}px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; }
        .redis-status-summary small { color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; }
        .redis-status-summary strong { font-size: ${compact ? 15 : 17}px; font-weight: 500; overflow-wrap: anywhere; font-variant-numeric: tabular-nums; }
        .redis-status-note { flex: none; margin-bottom: 12px; color: ${token.colorTextTertiary}; font-size: ${compact ? 11 : 12}px; overflow-wrap: anywhere; }
        .redis-status-table { flex: 1; min-height: 140px; min-width: 0; }
        .redis-status-table .ant-table { font-size: ${compact ? 11 : 12}px; color: ${token.colorTextSecondary}; }
        .redis-status-table .ant-table-thead > tr > th { font-size: ${compact ? 12 : 13}px; font-weight: 500; color: ${token.colorTextSecondary}; }
        .redis-status-table .ant-table-cell { padding: ${compact ? 6 : 8}px ${compact ? 8 : 10}px; white-space: nowrap; }
        .redis-status-table .ant-table-container { border-inline: 1px solid ${token.colorBorderSecondary}; border-bottom: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; overflow: hidden; }
        && .redis-status-table .ant-table-tbody > tr:last-child > td,
        && .redis-status-table .ant-table-tbody-virtual-holder-inner > .client-last-row > .ant-table-cell { border-bottom: 0; }
        .redis-status-table .ant-table-body, .redis-status-table .ant-table-tbody-virtual-holder { scrollbar-width: none; }
        .redis-status-table .ant-table-body::-webkit-scrollbar, .redis-status-table .ant-table-tbody-virtual-scrollbar { display: none; }
        .redis-status-field { display: grid; min-width: 0; gap: 2px; }
        .redis-status-field > span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .redis-status-name { color: ${token.colorTextTertiary}; font-size: ${compact ? 10 : 11}px; }
        .redis-status-value { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-family: ${token.fontFamilyCode}; font-variant-numeric: tabular-nums; }
        @media (max-width: 767px) {
            .redis-status-search { flex-basis: 100%; width: 100%; }
            .redis-status-tools { flex: 0 0 100%; width: 100%; margin-left: 0; }
            .redis-status-refresh { margin-left: auto; }
            .redis-status-summary { grid-template-columns: repeat(2, minmax(0, 1fr)); }
        }
    `,
    detail: css`
        .ant-modal-body { min-width: 0; font-size: ${token.fontSize}px; line-height: 1.6; }
        .redis-status-copy { display: flex; justify-content: flex-end; margin-bottom: 8px; }
        .redis-status-copy .ant-typography-copy { display: inline-flex; align-items: center; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        pre { margin: 0; padding: ${token.padding}px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; color: ${token.colorText}; font: ${token.fontSizeSM}px/1.8 ${token.fontFamilyCode}; white-space: pre-wrap; overflow-wrap: anywhere; tab-size: 4; }
    `,
}));
