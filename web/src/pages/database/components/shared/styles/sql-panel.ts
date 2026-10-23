import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    panel: css`
        min-width: 0;
        .sql-detail-navigation { display: flex; align-items: center; gap: 10px; border-bottom: 1px solid ${token.colorBorderSecondary}; margin-bottom: 12px; }
        .sql-detail-navigation .ant-tabs { flex: 1; min-width: 0; }
        .sql-detail-navigation .ant-tabs-nav { margin: 0; }
        .sql-detail-navigation .ant-tabs-nav::before { border: 0; }
        .sql-detail-navigation > .ant-btn { flex: none; margin-bottom: 7px; }
        .sql-stage { display: grid; grid-template-columns: 205px minmax(0, 1fr); gap: 18px; min-height: 445px; align-items: start; }
        .sql-stage-wide { grid-template-columns: minmax(0, 1fr); }
        .sql-content { min-width: 0; }
        .sql-detail-tabs .ant-tabs-nav { margin-bottom: 0; }
        .sql-detail-tabs .ant-tabs-tab { padding-block: 8px; }
        .sql-detail-body { min-height: 375px; min-width: 0; }
        .sql-content .ant-btn { font-size: ${token.fontSize}px; }
        .sql-content .client-footer { padding-top: 10px; }
        .sql-content .client-toolbar .ant-typography { font-size: ${token.fontSizeSM}px; }
        .sql-definition { min-height: 330px; }
        .sql-definition-toolbar { display: flex; justify-content: space-between; align-items: center; gap: 8px; margin-bottom: 10px; }
        .sql-definition-toolbar .ant-typography { font-size: ${token.fontSizeSM}px; }
        .sql-definition-text { min-height: 280px; max-height: 350px; overflow: auto; margin: 0; padding: 12px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; color: ${token.colorText}; font-family: ${token.fontFamilyCode}; font-size: ${token.fontSizeSM}px; line-height: 1.65; white-space: pre-wrap; overflow-wrap: anywhere; }
        @media (max-width: 767px) {
            .sql-stage { grid-template-columns: minmax(0, 1fr); gap: 12px; }
            .sql-detail-body { min-height: 350px; }
            .sql-definition-text { padding: 10px; }
        }
    `,
}));
