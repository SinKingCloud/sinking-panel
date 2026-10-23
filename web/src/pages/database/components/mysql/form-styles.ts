import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    form: css`
        min-height: 0;
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .ant-form-item-extra { font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .db-form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; gap: 0 ${token.margin}px; }
        .db-form-grid .ant-form-item { margin-bottom: ${token.margin}px; }
        .db-form-grid .db-field-wide { grid-column: 1 / -1; }
        .db-address-grid { grid-template-columns: minmax(0, 1fr) 128px; }
        .db-form-more { margin-top: 2px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; }
        .db-form-more .ant-collapse-item { border: 0; }
        .db-form-more .ant-collapse-header { align-items: center !important; padding: ${token.paddingSM}px !important; font-size: ${token.fontSize}px; }
        .db-form-more .ant-collapse-content { background: transparent; border: 0; }
        .db-form-more .ant-collapse-content-box { padding: ${token.paddingXS}px ${token.paddingSM}px ${token.paddingSM}px !important; }
        .db-form-more .db-form-grid { row-gap: ${token.margin}px; }
        .db-form-more .ant-form-item { margin-bottom: 0; }
        .db-connection-feedback { display: flex; align-items: center; gap: ${token.marginSM}px; min-height: ${token.controlHeight}px; margin-top: ${token.margin}px; font-size: ${token.fontSizeSM}px; }
        .db-connection-feedback > .ant-btn { flex: none; }
        .db-test-result { display: flex; align-items: center; gap: 6px; min-width: 0; flex: 1; }
        .db-test-result > .anticon { flex: none; }
        .db-test-result.success { color: ${token.colorSuccess}; }
        .db-test-result.error { color: ${token.colorError}; }
        .db-connection-feedback .ant-typography { font-size: inherit; min-width: 0; }
        .db-test-version { flex: 1; }
        .db-form-empty { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 16px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; text-align: center; }
        .db-form-empty .ant-empty { margin-block: 0; }
        .db-note-input { resize: none; }
        @media (max-width: 575px) {
            .db-form-grid { grid-template-columns: minmax(0, 1fr); gap: 0; }
            .db-address-grid { grid-template-columns: minmax(0, 1fr) 100px; gap: 0 12px; }
            .db-form-grid .db-field-wide { grid-column: 1 / -1; }
            .db-connection-feedback { flex-wrap: wrap; }
            .db-test-result { flex-basis: 180px; }
        }
    `,
    detail: css`
        .db-detail-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 18px; }
        .db-detail-heading > span:first-child { font-size: ${token.fontSize}px; font-weight: 500; overflow-wrap: anywhere; }
        .db-detail-heading .ant-tag { flex: none; margin-inline-end: 0; }
        .db-detail-metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(108px, 1fr)); gap: 8px; margin-bottom: 20px; }
        .db-detail-metric { display: flex; flex-direction: column; gap: 6px; padding: 12px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; }
        .db-detail-metric > span { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .db-detail-metric > strong { font-size: ${token.fontSize}px; font-weight: 500; color: ${token.colorText}; font-variant-numeric: tabular-nums; }
        .db-detail-group + .db-detail-group { margin-top: 20px; }
        .db-detail-group > h3 { margin: 0 0 12px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; font-weight: 500; }
        .db-detail-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px 20px; margin: 0; }
        .db-detail-grid > div { display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: start; gap: 10px; font-size: ${token.fontSizeSM}px; line-height: 1.6; }
        .db-detail-grid dt { color: ${token.colorTextSecondary}; }
        .db-detail-grid dd { margin: 0; color: ${token.colorText}; overflow-wrap: anywhere; font-variant-numeric: tabular-nums; }
        .db-detail-grid dd .ant-typography { font-size: inherit; }
        .db-detail-grid .db-detail-wide { grid-column: 1 / -1; }
        @media (max-width: 575px) {
            .db-detail-grid { grid-template-columns: minmax(0, 1fr); gap: 10px; }
            .db-detail-grid > div { grid-template-columns: 84px minmax(0, 1fr); }
        }
    `,
}));
