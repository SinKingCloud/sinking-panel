import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact = false}: {compact?: boolean} = {}) => ({
    request: css`
        .es-request-form { display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0; }
        .es-request-splitter { flex: 1; height: auto; min-height: ${compact ? 380 : 400}px; }
        .es-request-editor {
            display: flex; flex-direction: column; height: 100%; min-height: 0; overflow: hidden; box-sizing: border-box;
            border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px;
        }
        .es-request-content { flex: 1; min-width: 0; min-height: 0; overflow: hidden; }
        .es-request-content > textarea.ant-input {
            height: 100%; min-height: 0; border: 0; border-radius: 0; resize: none;
            font: ${compact ? 12 : 13}px/1.6 ${token.fontFamilyCode};
        }
        .manager-query-actions { flex: none; background: ${token.colorFillQuaternary}; }
        /* 拖动热区贴合编辑器底边，上下两区由 Splitter 保留最低空间。 */
        &&& .es-request-resizer::before { display: none; }
        &&& .es-request-resizer::after {
            top: calc(50% - 3px); width: 32px; height: 3px; border-radius: 2px; background: ${token.colorTextQuaternary};
        }
        &&& .es-request-resizer:is(:hover, .ant-splitter-bar-dragger-active)::after { background: ${token.colorPrimary}; }
        .es-request-results { display: flex; flex-direction: column; min-width: 0; padding-top: 12px; }
        .es-request-results .manager-tools { margin-left: auto; }
        .es-request-results .manager-output, .es-request-results .manager-empty { min-height: 0; }
        .es-request-results .manager-output > .ant-spin-nested-loading,
        .es-request-results .manager-output .ant-spin-container { min-height: 100%; }
    `,
}));
