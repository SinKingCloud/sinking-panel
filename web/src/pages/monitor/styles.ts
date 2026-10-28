import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    page: css`
        font-variant-numeric: tabular-nums;

        @container data-table-workspace (max-width: 560px) {
            .ui-table-toolbar.with-content { flex-wrap: wrap; gap: 8px; }
            .ui-table-toolbar .command-content { flex-basis: 112px; }
            .ui-table-toolbar.with-content .command-actions { margin-inline-start: auto; }
        }
    `,
    timeFilter: css`
        position: relative;
        display: inline-flex;
        align-items: center;
        max-width: 100%;

        .ant-segmented { font-size: 12px; }
        .ant-segmented-item-label { padding-inline: 12px; }

        @container data-table-workspace (max-width: 380px) {
            .ant-segmented-item-label { padding-inline: 7px; }
        }
    `,
    mobile: css`
        .ui-table-toolbar .command-actions-scroll { position: relative; }
        .ui-table-toolbar .command-actions-scroll > .ant-btn:last-child { margin-inline-start: auto; }
        .ui-table-toolbar .command-actions-scroll > .ant-picker {
            inset-inline-start: 0;
            inset-inline-end: auto;
            width: 112px;
        }
    `,
    dateAnchor: css`
        && {
            position: absolute;
            inset: 0 0 0 auto;
            width: 72px;
            min-width: 0;
            height: 100%;
            padding: 0;
            overflow: hidden;
            border: 0;
            opacity: 0;
            pointer-events: none;
        }
    `,
    datePopup: css`
        && .ant-picker-range-arrow {
            left: clamp(${token.borderRadiusLG}px, calc(var(--arrow-x, 50%) - ${token.sizePopupArrow / 2}px), calc(100% - ${token.borderRadiusLG + token.sizePopupArrow}px)) !important;
            right: auto;
            padding-inline: 0;
            transition: none;
            &::before { inset-inline-start: 0; }
        }

        && .ant-picker-panels > :last-child:not(:first-child) { display: none; }
        && .ant-picker-panels > :first-child .ant-picker-header-next-btn,
        && .ant-picker-panels > :first-child .ant-picker-header-super-next-btn { visibility: visible !important; }
    `,
    problem: css`
        margin-inline-start: 5px;
        color: ${token.colorWarningText};
    `,
    charts: css`
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 16px;
        min-width: 0;

        @container data-table-workspace (max-width: 860px) {
            grid-template-columns: minmax(0, 1fr);
            gap: 12px;
        }
    `,
}));
