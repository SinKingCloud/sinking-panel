import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    command: css`
        display: flex;
        flex-direction: column;
        height: 100%;
        min-height: 0;
        min-width: 0;
        overflow-x: hidden;
        overflow-y: auto;
        overscroll-behavior: contain;
        scrollbar-width: thin;
        color: ${token.colorText};
        font-size: ${token.fontSize}px;
        line-height: 1.5;
        .redis-command-workbench { flex: 1; min-height: ${compact ? 440 : 460}px; height: auto; }
        .redis-command-input, .redis-command-result { min-width: 0; }
        .redis-command-input { display: flex; flex-direction: column; height: 100%; min-height: 0; box-sizing: border-box; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; overflow: hidden; }
        .redis-command-input-heading { display: flex; align-items: center; justify-content: space-between; flex: none; flex-wrap: wrap; gap: 8px; padding: 8px 10px; background: ${token.colorFillQuaternary}; }
        .redis-command-editor { flex: 1; min-width: 0; min-height: 0; overflow: hidden; }
        .redis-command-input textarea.ant-input { height: 100%; min-height: 0; resize: none; border: 0; border-radius: 0; background: ${token.colorBgContainer}; box-shadow: none; padding: 10px; font: 12px/1.65 ${token.fontFamilyCode}; }
        .redis-command-controls { display: flex; align-items: center; justify-content: space-between; flex: none; flex-wrap: wrap; gap: 4px 8px; padding: 6px 10px; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; background: ${token.colorFillQuaternary}; }
        .redis-command-controls .ant-typography { font-size: inherit; color: inherit; }
        &&& .redis-command-resizer::before { display: none; }
        &&& .redis-command-resizer::after { top: calc(50% - 3px); width: 32px; height: 3px; border-radius: 2px; background: ${token.colorTextQuaternary}; }
        &&& .redis-command-resizer:is(:hover, .ant-splitter-bar-dragger-active)::after { background: ${token.colorPrimary}; }
        .redis-command-results { display: flex; flex-direction: column; min-width: 0; padding-top: 12px; }
        .redis-command-result { display: grid; grid-template-rows: auto minmax(0, 1fr); flex: 1; min-height: 0; }
        .redis-command-result-toolbar { display: flex; align-items: center; justify-content: space-between; gap: ${token.marginSM}px; min-height: ${Math.max(26, token.controlHeight - 4)}px; margin-bottom: 8px; }
        .redis-command-result-toolbar strong { color: ${token.colorTextSecondary}; font-size: ${Math.max(11, token.fontSize - 1)}px; font-weight: 400; }
        .redis-command-result-toolbar > div:first-child { display: flex; align-items: baseline; flex-wrap: wrap; gap: 8px; min-width: 0; }
        .redis-command-result-toolbar > div:first-child > span { color: ${token.colorTextSecondary}; font-family: ${token.fontFamilyCode}; overflow-wrap: anywhere; }
        .redis-command-result-toolbar > .redis-copy-action { flex-shrink: 0; white-space: nowrap; }
        .redis-command-spin, .redis-command-spin-content { height: 100%; min-height: 0; min-width: 0; }
        .redis-command-output { height: 100%; min-height: 0; box-sizing: border-box; overflow: auto; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorBgContainer}; scrollbar-width: thin; }
        .redis-command-output pre { margin: 0; padding: ${token.paddingSM}px; font-family: ${token.fontFamilyCode}; font-size: ${Math.max(11, token.fontSize - 1)}px; line-height: 1.65; white-space: pre-wrap; overflow-wrap: anywhere; }
        .redis-command-truncated { padding: 8px ${token.paddingSM}px; border-bottom: 1px solid ${token.colorBorderSecondary}; background: ${token.colorFillQuaternary}; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .redis-command-placeholder { height: 100%; min-height: 60px; display: grid; place-items: center; color: ${token.colorTextSecondary}; }
        .redis-command-placeholder .ant-empty-description { font-size: inherit; }
        .redis-command-error { padding: ${token.paddingSM}px; }
        .redis-command-error > p { margin: 0; color: ${token.colorError}; font-size: ${Math.max(11, token.fontSize - 1)}px; line-height: 1.65; overflow-wrap: anywhere; }
        @media (max-width: 767px) {
            .redis-command-input-heading { gap: 6px; padding: 8px; }
        }
    `,
}));
