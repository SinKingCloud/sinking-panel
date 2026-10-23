import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    host: css`
        display: flex;
        flex: 1;
        flex-direction: column;
        min-width: 0;
        min-height: 0;
        container: mysql-content / inline-size;
        color: ${token.colorText};
        font-size: ${token.fontSize}px;

        /* 共享工作区有固定字号，保持这里的文字与原生组件主题一致。 */
        &&& .ant-input, &&& .ant-input-affix-wrapper, &&& .ant-select {
            font-size: ${token.fontSize}px;
        }
        /* 仅调整工作区按钮，避免主题配置传递到弹窗底部。 */
        &&& .ant-btn {
            height: ${compact ? 24 : 28}px;
            padding-inline: ${compact ? 7 : 9}px;
            gap: ${compact ? 4 : 6}px;
            border-radius: ${token.borderRadiusSM}px;
            font-size: ${compact ? 11 : 12}px;
        }
        &&& .ant-btn-icon-only { width: ${compact ? 24 : 28}px; padding-inline: 0; }
        &&& .ant-btn-color-default:not(:disabled) { color: ${token.colorTextSecondary}; }
        &&& .ant-btn:disabled { border-color: transparent; background: ${token.colorFillTertiary}; }
        && .ant-btn .anticon { font-size: inherit; }
        &&& .ant-table .ant-btn-sm { height: ${compact ? 22 : 24}px; padding-inline: ${compact ? 6 : 7}px; }
        /* 固定列重新计算位置时直接就位，只保留颜色和阴影过渡。 */
        && .ant-table-cell:is(.ant-table-cell-fix, .ant-table-cell-fix-start, .ant-table-cell-fix-end, .ant-table-cell-fix-left, .ant-table-cell-fix-right) {
            transition-property: background-color, color, border-color, box-shadow;
        }
        /* Safari 缩放时会用变换后的滚动宽度定位 sticky；与实际容器宽度相减，保持末尾操作列贴边。 */
        @supports (width: 1cqi) {
            .ant-table-has-fix-end :is(.ant-table-header, .ant-table-body, .ant-table-content, .ant-table-tbody-virtual-holder) {
                container-type: inline-size;
            }
            && .ant-table-cell-fix-end {
                inset-inline-end: calc(100% - 100cqi) !important;
            }
        }

        .mysql-header { display: grid; flex: none; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr); align-items: center; gap: 16px; min-width: 0; margin-bottom: 16px; }
        .mysql-navigation { grid-column: 2; justify-self: center; min-width: 0; max-width: 100%; overflow-x: auto; scrollbar-width: none; }
        .mysql-navigation::-webkit-scrollbar { display: none; }
        .mysql-navigation .ant-segmented-group { display: flex; width: max-content; min-width: 100%; }
        .mysql-navigation .ant-segmented-item { flex: 1 0 auto; }
        .mysql-navigation .ant-segmented-item-label { padding-inline: 12px; }

        .sql-workbench { display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0; }
        .sql-workbench > .sql-stage {
            flex: 1;
            grid-template-columns: minmax(0, 1fr);
            gap: 0;
            min-height: ${compact ? 500 : 520}px;
            align-items: stretch;
            border: 0;
            border-radius: 0;
        }

        /* 表名是当前文档，常用操作与表内视图分成两行。 */
        .mysql-browser-header { flex: none; min-width: 0; }
        .mysql-table-toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: 8px 16px; min-width: 0; padding: 0 0 ${compact ? 10 : 12}px; }
        .mysql-browse-target { display: flex; flex: 1 1 180px; align-items: center; gap: 8px; min-width: 0; }
        .mysql-browse-target > .ant-select { min-width: 100px; max-width: 100%; }
        .mysql-browse-target > .mysql-table-picker { flex: 0 1 240px; width: 240px; min-width: 0; max-width: 100%; }
        &&& .mysql-table-picker.ant-select {
            align-items: center;
            height: ${compact ? 30 : 34}px;
            padding: 0 ${compact ? 7 : 10}px;
            border-color: transparent;
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorFillQuaternary};
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
            box-shadow: none;
            transition: background-color .16s ease, border-color .16s ease;
        }
        &&& .mysql-table-picker:not(.ant-select-disabled):hover,
        &&& .mysql-table-picker.ant-select-open { background: ${token.colorFillSecondary}; }
        &&& .mysql-table-picker.ant-select-focused:not(.ant-select-open) {
            border-color: ${token.colorPrimaryBorder};
            background: ${token.colorBgContainer};
            box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
        }
        &&& .mysql-table-picker.ant-select-disabled { color: ${token.colorTextDisabled}; }
        .mysql-table-picker .ant-select-prefix { margin-inline-end: 5px; color: ${token.colorTextQuaternary}; }
        .mysql-table-picker .ant-select-content { margin-inline-end: 5px; }
        &&& .mysql-table-picker .ant-select-placeholder,
        &&& .mysql-table-picker .ant-select-input { color: inherit; }
        .mysql-table-picker .ant-select-suffix { color: ${token.colorTextQuaternary}; font-size: ${compact ? 9 : 10}px; }
        .mysql-browse-target > .ant-select:not(.mysql-table-picker) { flex: 0 1 140px; }
        .mysql-table-tools { display: flex; flex: none; align-items: center; gap: 6px; min-width: 0; margin-left: auto; }
        .mysql-table-operations { display: flex; align-items: center; gap: 6px; min-width: 0; }
        .mysql-table-operations > .ant-btn { flex: none; }
        .mysql-table-tools > .ant-btn { flex: none; }
        .mysql-table-tools .client-toolbar-dropdown { gap: 5px; }
        .mysql-table-tools .client-toolbar-dropdown > .ant-btn-icon { color: ${token.colorTextTertiary}; }
        && .mysql-table-tools .client-toolbar-dropdown .client-toolbar-arrow {
            color: ${token.colorTextQuaternary};
            font-size: ${compact ? 9 : 10}px;
            transition: transform .16s ease;
        }
        .mysql-table-tools .ant-dropdown-open .client-toolbar-arrow { transform: rotate(180deg); }

        .sql-workbench .sql-content {
            display: flex;
            flex-direction: column;
            min-width: 0;
            min-height: 0;
            padding: 0;
            container: mysql-detail / inline-size;
        }

        /* 表内页签与当前视图操作同排，窄屏时操作自然换行。 */
        .sql-workbench .sql-detail-navigation {
            display: flex;
            flex: none;
            flex-wrap: wrap;
            align-items: center;
            gap: 8px 12px;
            margin: 0 0 ${compact ? 8 : 10}px;
            padding: 8px 0;
            border: 0;
        }
        .sql-workbench .sql-detail-navigation > .sql-detail-tabs { flex: 1 1 260px; min-width: 0; max-width: 100%; }
        .sql-workbench .sql-detail-tabs .ant-tabs-nav { margin: 0; }
        &&& .sql-detail-tabs .ant-tabs-tab {
            justify-content: center;
            min-width: ${compact ? 44 : 50}px;
            height: ${compact ? 26 : 30}px;
            padding: 0 ${compact ? 10 : 12}px;
            border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
            font-weight: 400;
            line-height: 1.5;
            transition: color .16s ease, background-color .16s ease;
        }
        &&& .sql-detail-tabs .ant-tabs-tab + .ant-tabs-tab { margin-inline-start: 0; }
        &&& .sql-detail-tabs .ant-tabs-tab:not(.ant-tabs-tab-disabled):not(.ant-tabs-tab-active):hover { color: ${token.colorText}; background: ${token.colorFillQuaternary}; }
        &&& .sql-detail-tabs .ant-tabs-tab-active { background: ${token.colorPrimaryBg}; }
        &&& .sql-detail-tabs .ant-tabs-tab-active .ant-tabs-tab-btn { color: ${token.colorPrimary}; font-weight: 400; text-shadow: none; }
        &&& .sql-detail-tabs .ant-tabs-tab-disabled { color: ${token.colorTextDisabled}; }
        .sql-workbench .sql-detail-tabs .ant-tabs-tab-btn:focus-visible { border-radius: 2px; outline-offset: -2px; }
        .sql-workbench .sql-detail-tabs .ant-tabs-nav::before { border: 0; }
        .sql-workbench .sql-detail-tabs .ant-tabs-ink-bar { display: none; }
        .mysql-detail-actions { display: flex; flex: 0 1 auto; flex-wrap: wrap; justify-content: flex-end; align-items: center; gap: 8px; min-width: 0; max-width: 100%; margin-left: auto; }
        .sql-workbench .mysql-detail-actions .client-toolbar { justify-content: flex-end; gap: 6px; margin: 0; }
        .mysql-detail-actions .client-toolbar-spacer { display: none; }
        .sql-workbench .mysql-detail-actions .client-toolbar-actions { margin: 0; gap: 6px; }
        .mysql-detail-actions .client-toolbar > .ant-typography { color: ${token.colorTextTertiary}; white-space: nowrap; }
        .sql-workbench .sql-detail-body { display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0; overflow-y: auto; }
        .sql-workbench .sql-detail-body > .client-toolbar { flex: none; }
        .sql-workbench .sql-detail-body > .client-hint { flex: none; margin: 8px 0 0; }
        /* 少量数据按自然高度展示，超出可用高度时仅滚动表体。 */
        .mysql-fill-table,
        .mysql-fill-table > .ant-spin,
        .mysql-fill-table > .ant-spin > .ant-spin-container,
        .mysql-fill-table .ant-table,
        .mysql-fill-table .ant-table-container {
            display: flex;
            flex: 0 1 auto;
            flex-direction: column;
            min-width: 0;
            min-height: 0;
        }
        .mysql-fill-table .ant-table-header, .mysql-fill-table .ant-table-pagination { flex: none; }
        .mysql-fill-table .ant-table-body { flex: 0 1 auto; min-height: 0; }
        .sql-workbench .client-toolbar { gap: 8px; margin-bottom: 12px; }
        .mysql-object-tools { display: flex; flex: none; align-items: center; gap: 8px; margin-inline-start: auto; }
        &&& .sql-object-search {
            flex: 0 1 240px;
            width: 240px;
            min-width: 0;
            max-width: 100%;
            height: ${compact ? 24 : 28}px;
            padding-inline: ${compact ? 8 : 10}px;
            border-radius: ${token.borderRadiusSM}px;
            font-size: ${compact ? 11 : 12}px;
        }
        &&& .sql-object-search .ant-input { color: ${token.colorTextSecondary}; font-size: inherit; }
        .sql-object-search .ant-input::placeholder { color: ${token.colorTextQuaternary}; font-size: inherit; }
        .sql-object-search .ant-input-prefix { margin-inline-end: 8px; color: ${token.colorTextTertiary}; font-size: ${compact ? 12 : 14}px; }
        .sql-workbench .client-toolbar-actions { gap: 8px; margin-left: auto; }
        .sql-workbench .client-footer { flex: none; margin-top: auto; padding-top: 10px; }
        .sql-workbench .client-pagination { display: flex; flex: 0 1 auto; flex-wrap: wrap; align-items: center; justify-content: flex-end; gap: 6px 4px; min-width: 0; max-width: 100%; margin: 0 0 0 auto; }
        &&& .client-pagination { font-size: ${compact ? 11 : 12}px; }
        &&& .client-pagination .ant-pagination-item,
        &&& .client-pagination .ant-pagination-prev,
        &&& .client-pagination .ant-pagination-next,
        &&& .client-pagination .ant-pagination-jump-prev,
        &&& .client-pagination .ant-pagination-jump-next {
            min-width: ${compact ? 22 : 26}px;
            height: ${compact ? 22 : 26}px;
            margin: 0;
            border: 0;
            border-radius: ${token.borderRadiusSM}px;
            background: transparent;
            font-size: inherit;
            line-height: ${compact ? 22 : 26}px;
        }
        &&& .client-pagination .ant-pagination-item a,
        &&& .client-pagination .ant-pagination-item-link { color: ${token.colorTextTertiary}; }
        &&& .client-pagination .ant-pagination-item:not(.ant-pagination-item-active):hover,
        &&& .client-pagination .ant-pagination-prev:not(.ant-pagination-disabled):hover,
        &&& .client-pagination .ant-pagination-next:not(.ant-pagination-disabled):hover { background: ${token.colorFillTertiary}; }
        &&& .client-pagination .ant-pagination-item-active,
        &&& .client-pagination.ant-pagination-disabled .ant-pagination-item-active {
            &, &:hover, &:active { background: ${token.colorBgContainer}; box-shadow: inset 0 0 0 1px ${token.colorPrimaryBorder}; }
            a, &:hover a, &:active a { color: ${token.colorPrimary}; }
        }
        &&& .client-pagination .ant-pagination-disabled .ant-pagination-item-link { color: ${token.colorTextDisabled}; }
        &&& .client-pagination .ant-pagination-options { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; margin-inline-start: 6px; }
        &&& .client-pagination .ant-select { height: ${compact ? 22 : 26}px; color: ${token.colorTextSecondary}; font-size: inherit; }
        /* 加载时会禁用分页，填充式容量选择仍保持透明边框。 */
        &&& .client-pagination .ant-select-filled.ant-select-disabled { border-color: transparent; }
        .client-pagination .ant-select-suffix { color: ${token.colorTextQuaternary}; font-size: ${compact ? 9 : 10}px; }
        &&& .client-pagination .ant-pagination-options-quick-jumper { height: auto; margin: 0; color: ${token.colorTextSecondary}; font-size: inherit; line-height: ${compact ? 22 : 26}px; }
        &&& .client-pagination .ant-pagination-options-quick-jumper input { width: ${compact ? 38 : 42}px; height: ${compact ? 22 : 26}px; margin-inline: 4px; border-radius: ${token.borderRadiusSM}px; font-size: inherit; }
        .sql-workbench .client-footer .ant-typography, .sql-workbench .client-hint { font-size: ${token.fontSizeSM}px; }
        .sql-workbench .client-empty { flex: 1; min-height: 280px; }
        .sql-workbench .client-hint { margin-top: 8px; }
        &&& .ant-table { font-size: ${compact ? 11 : 12}px; }
        &&& .ant-table:not(.ant-table-bordered) > .ant-table-container {
            overflow: hidden;
            border-inline: 1px solid ${token.colorBorderSecondary};
            border-bottom: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusLG}px;
        }
        &&& .ant-table-tbody > tr:last-child > td,
        &&& .ant-table-tbody-virtual-holder-inner > .client-last-row > .ant-table-cell { border-bottom: 0; }
        &&& .ant-table .client-cell, &&& .ant-table .client-row-number { font-size: inherit; }
        .sql-workbench .client-result-status .ant-typography { font-size: ${token.fontSizeSM}px; }
        .sql-workbench .ant-table-cell { white-space: nowrap; }
        &&& .ant-table-thead > tr > th { font-size: ${compact ? 12 : 13}px; font-weight: 500; }

        /* 小窗口保留两区最低空间，超出时由工作区滚动，避免压没编辑内容或结果。 */
        .sql-workbench .client-query { flex: 1; height: auto; min-height: ${compact ? 440 : 460}px; gap: 0; }
        /* 拖动柄与编辑器底边重合，热区略向下偏移，避开底部文字。 */
        &&& .client-query-resizer::before { display: none; }
        &&& .client-query-resizer::after {
            top: calc(50% - 3px);
            width: 32px;
            height: 3px;
            border-radius: 2px;
            background: ${token.colorTextQuaternary};
        }
        &&& .client-query-resizer:is(:hover, .ant-splitter-bar-dragger-active)::after { background: ${token.colorPrimary}; }
        .sql-workbench .client-query-results { display: flex; flex-direction: column; gap: 12px; padding-top: 12px; }
        .sql-workbench .client-query-results > * { min-width: 0; }
        .sql-workbench .client-query-results > .client-hint { flex: none; margin: 0; }
        .sql-workbench .client-result-status { flex: none; margin-bottom: 0; }
        .sql-workbench .client-result-actions { flex-wrap: wrap; margin-left: auto; }
        .sql-workbench .client-query-output {
            flex: 1;
            min-height: 0;
            margin: 0;
            padding: ${compact ? 10 : 12}px;
            overflow: auto;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusLG}px;
            background: ${token.colorFillQuaternary};
            color: ${token.colorText};
            font-family: ${token.fontFamilyCode};
            font-size: ${token.fontSizeSM}px;
            line-height: 1.65;
            white-space: pre-wrap;
            overflow-wrap: anywhere;
        }
        .sql-workbench .client-query-error { color: ${token.colorError}; }
        .sql-workbench .query-editor-toolbar, .sql-workbench .query-editor-footer { padding-inline: 10px; }
        .sql-workbench .query-editor-toolbar > .ant-space { flex-wrap: wrap; max-width: 100%; }
        .sql-workbench .query-editor-toolbar > .ant-space:last-child { margin-left: auto; }
        .sql-workbench .query-editor-footer { flex-wrap: wrap; }
        .sql-workbench .query-editor-footer, .sql-workbench .query-editor-toolbar .ant-typography { font-size: ${token.fontSizeSM}px; }

        .mysql-section-loading { min-height: 320px; display: grid; place-items: center; }
        .mysql-maintenance { display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0; }
        .mysql-maintenance-target { display: flex; flex: none; align-items: center; gap: 8px; margin-bottom: 12px; }
        .mysql-maintenance-target > .mysql-browse-target { flex: 1; }
        .mysql-maintenance-target > .ant-btn { flex: none; margin-left: auto; }
        .mysql-maintenance > .maintenance-panel { min-width: 0; }
        .mysql-maintenance > .client-empty { flex: 1; }
        .mysql-maintenance .maintenance-list { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorBgContainer}; overflow: hidden; }
        .mysql-maintenance .maintenance-row {
            display: grid;
            grid-template-columns: minmax(0, 1fr) auto;
            align-items: center;
            gap: 12px;
            padding: ${compact ? 10 : 12}px ${compact ? 12 : 14}px;
            border: 0;
            background: none;
        }
        .mysql-maintenance .maintenance-row + .maintenance-row { border-top: 1px solid ${token.colorBorderSecondary}; }
        .mysql-maintenance .maintenance-tool-title { color: ${token.colorText}; font-size: ${compact ? 12 : 13}px; font-weight: 400; line-height: 1.6; }
        .mysql-maintenance .maintenance-row .ant-typography { color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; line-height: 1.6; }
        .mysql-maintenance .maintenance-row > div { min-width: 0; gap: 2px; }
        .mysql-maintenance .maintenance-row > .ant-btn { min-width: ${compact ? 76 : 84}px; justify-self: end; }

        && .ant-table-body, && .ant-table-content, && .ant-table-tbody-virtual-holder { scrollbar-width: none; }
        .ant-table-body::-webkit-scrollbar, .ant-table-content::-webkit-scrollbar, .ant-table-tbody-virtual-holder::-webkit-scrollbar { display: none; }
        .ant-table-tbody-virtual-scrollbar { display: none; }
        .sql-definition-text { scrollbar-width: thin; scrollbar-color: ${token.colorFill} transparent; }
        .sql-definition-text::-webkit-scrollbar { width: 6px; height: 6px; display: block; }
        .sql-definition-text::-webkit-scrollbar-thumb { border-radius: 3px; background: ${token.colorFill}; }

        @container mysql-detail (max-width: 540px) {
            .mysql-browse-target { flex-basis: 100%; }
            .mysql-table-tools { width: 100%; max-width: 100%; margin-left: 0; }
            .mysql-table-operations { flex: 1; overflow-x: auto; scrollbar-width: none; }
            .mysql-table-operations::-webkit-scrollbar { display: none; }
            .sql-workbench .sql-detail-navigation > .sql-detail-tabs { flex-basis: 100%; }
            .mysql-detail-actions { flex-basis: 100%; justify-content: flex-start; margin-left: 0; }
            .sql-workbench .mysql-detail-actions .client-toolbar { justify-content: flex-start; }
            .sql-workbench .mysql-detail-actions .client-toolbar-actions { margin-left: auto; }
            .sql-workbench .sql-detail-body > .client-toolbar > .client-toolbar-spacer { display: none; }
            .sql-workbench .client-query-results { gap: 10px; }
            .sql-workbench .query-editor-toolbar { gap: 6px; padding: 8px; }
            .sql-workbench .query-editor-toolbar > .ant-space:last-child { margin-left: 0; }
            .sql-workbench .client-result-status > .ant-space:first-child { flex-basis: 100%; }
            .sql-workbench .client-result-actions { margin-left: 0; }
        }
        @container mysql-content (max-width: 640px) {
            .mysql-header { grid-template-columns: minmax(0, 1fr); gap: 10px; }
            .mysql-navigation { grid-column: 1; }
            .mysql-maintenance-target > .mysql-browse-target { flex-basis: 0; }
            .mysql-maintenance .maintenance-row { padding-inline: 10px; gap: 10px; }
        }
        @media (max-width: 767px) {
            .mysql-header { margin-bottom: 12px; }
            .mysql-navigation { width: 100%; }
            .mysql-navigation .ant-segmented-item-label { padding-inline: 6px; }
            .sql-workbench > .sql-stage { min-height: 460px; }
            .mysql-table-toolbar { gap: 8px; padding-bottom: 10px; }
            .mysql-table-toolbar > .mysql-browse-target { flex-basis: 100%; width: 100%; }
            .mysql-table-toolbar > .mysql-browse-target > .mysql-table-picker { flex: 1 1 0; width: 0; }
            .mysql-table-tools { width: 100%; }
            .mysql-table-operations { flex: 1; overflow-x: auto; scrollbar-width: none; }
            .mysql-table-operations::-webkit-scrollbar { display: none; }
            .mysql-detail-actions { width: 100%; flex-wrap: nowrap; }
            .sql-workbench .mysql-detail-actions .client-toolbar { flex: 1; flex-wrap: nowrap; min-width: 0; overflow-x: auto; scrollbar-width: none; }
            .sql-workbench .mysql-detail-actions .client-toolbar::-webkit-scrollbar { display: none; }
            .mysql-detail-actions .client-toolbar > .ant-btn { flex: none; }
            &&& .sql-object-search { flex: 1 1 100%; width: 100%; }
            .mysql-object-tools { width: 100%; gap: 6px; }
            .sql-workbench .mysql-toolbar-refresh { flex: none; margin-inline-start: auto; }
            .sql-workbench .sql-detail-navigation { column-gap: 8px; }
            .sql-workbench .client-footer { flex-wrap: nowrap; gap: 6px; }
            .sql-workbench .client-footer > .ant-typography {
                flex: 0 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
                font-size: ${compact ? 11 : 12}px; font-variant-numeric: tabular-nums;
            }
            &&& .client-pagination { flex: none; width: auto; flex-wrap: nowrap; gap: 2px; white-space: nowrap; }
            &&& .client-pagination .ant-pagination-options { order: -1; flex: none; flex-wrap: nowrap; margin-inline: 0 4px; }
            &&& .client-pagination .ant-pagination-simple-pager {
                display: flex; align-items: center; gap: 5px; height: ${compact ? 22 : 26}px; margin: 0;
                color: ${token.colorTextSecondary}; font-variant-numeric: tabular-nums;
            }
            &&& .client-pagination .ant-pagination-simple-pager input {
                width: ${compact ? 34 : 38}px; height: 100%; margin: 0; padding: 0 4px;
                border: 1px solid transparent; border-radius: ${token.borderRadiusSM}px;
                background: ${token.colorFillTertiary}; color: inherit; font: inherit; text-align: center;
            }
            &&& .client-pagination .ant-pagination-simple-pager input:focus { border-color: ${token.colorPrimaryBorder}; }
            &&& .client-pagination .ant-pagination-slash { margin: 0; color: ${token.colorTextQuaternary}; }
            .sql-workbench .client-toolbar, .sql-workbench .client-toolbar-actions { gap: 6px; }
        }
    `,
}));
