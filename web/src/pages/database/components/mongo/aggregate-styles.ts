import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    workbench: css`
        display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0;
        .mongo-aggregate-splitter { flex: 1; height: auto; min-height: ${compact ? 440 : 460}px; gap: 0; }
        .mongo-aggregate-splitter > .ant-splitter-panel { overflow: hidden; }
        .mongo-query-input { display: flex; flex-direction: column; height: 100%; min-height: 0; box-sizing: border-box; overflow: hidden; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .mongo-query-toolbar, .mongo-query-footer { display: flex; flex: none; align-items: center; justify-content: space-between; flex-wrap: wrap; background: ${token.colorFillQuaternary}; }
        .mongo-query-toolbar { gap: 8px; padding: 8px 10px; }
        .mongo-query-footer { gap: 4px 8px; padding: 6px 10px; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; }
        .mongo-query-editor { flex: 1; min-height: 0; overflow: hidden; }
        .mongo-query-editor textarea { height: 100%; resize: none; font-family: ${token.fontFamilyCode}; }
        && .mongo-query-editor .ace_placeholder, .mongo-query-editor textarea::placeholder { color: ${token.colorTextQuaternary}; opacity: 1; }
        &&& .mongo-query-resizer::before { display: none; }
        &&& .mongo-query-resizer::after { top: calc(50% - 3px); width: 32px; height: 3px; border-radius: 2px; background: ${token.colorTextQuaternary}; }
        &&& .mongo-query-resizer:is(:hover, .ant-splitter-bar-dragger-active)::after { background: ${token.colorPrimary}; }
        .mongo-query-results { display: flex; flex-direction: column; gap: 12px; padding-top: 12px; }
        .mongo-query-results > * { min-width: 0; }
        .mongo-query-status { display: flex; flex: none; align-items: center; gap: 8px; font-size: ${token.fontSizeSM}px; }
        .mongo-query-status > div { display: flex; align-items: center; flex-wrap: wrap; gap: 4px 10px; min-width: 0; }
        .mongo-query-status > .ant-btn { flex: none; margin-left: auto; }
        .mongo-query-status .ant-btn { height: ${compact ? 24 : 28}px; padding-inline: ${compact ? 8 : 10}px; font-size: ${compact ? 11 : 12}px; color: ${token.colorTextSecondary}; }
        .mongo-query-statistics, .mongo-query-hint { color: ${token.colorTextTertiary}; }
        .mongo-query-output { flex: 1; min-height: 0; margin: 0; padding: ${compact ? 10 : 12}px; overflow: auto; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorFillQuaternary}; color: ${token.colorText}; font: ${token.fontSizeSM}px/1.65 ${token.fontFamilyCode}; white-space: pre-wrap; overflow-wrap: anywhere; }
        .mongo-query-output.is-error { color: ${token.colorError}; }
        .mongo-query-hint { flex: none; font-size: ${token.fontSizeSM}px; overflow-wrap: anywhere; }
        @media (max-width: 767px) {
            .mongo-query-footer { font-size: ${compact ? 10 : 11}px; }
        }
    `,
}));
