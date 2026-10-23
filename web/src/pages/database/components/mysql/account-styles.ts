import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    panel: css`
        display: flex;
        flex: 1;
        flex-direction: column;
        min-width: 0;
        min-height: 0;
        container: mysql-accounts / inline-size;

        .mysql-account-toolbar,
        .mysql-account-tools {
            display: flex;
            align-items: center;
            gap: 8px;
            min-width: 0;
        }
        .mysql-account-toolbar { flex: none; flex-wrap: wrap; margin-bottom: 12px; }
        .mysql-account-search { flex: 0 1 240px; width: 240px; min-width: 0; max-width: 100%; }
        &&& .mysql-account-search.ant-input-affix-wrapper {
            height: ${compact ? 24 : 28}px;
            padding-inline: ${compact ? 8 : 10}px;
            border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
        }
        &&& .mysql-account-search .ant-input { color: ${token.colorTextSecondary}; font-size: inherit; }
        .mysql-account-search .ant-input::placeholder { color: ${token.colorTextQuaternary}; }
        .mysql-account-search .ant-input-prefix { margin-inline-end: 8px; color: ${token.colorTextTertiary}; font-size: ${compact ? 12 : 14}px; }
        .mysql-account-tools { flex: none; flex-wrap: wrap; gap: 6px; max-width: 100%; margin-left: auto; }
        .mysql-account-tools > * { flex: none; }
        .mysql-account-table { min-width: 0; }
        .mysql-account-name { display: flex; align-items: center; gap: 7px; min-width: 0; }
        .mysql-account-name > span:first-child {
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        .mysql-account-protected { flex: none; color: ${token.colorTextTertiary}; }
        .mysql-account-source {
            display: block;
            overflow: hidden;
            color: ${token.colorTextSecondary};
            font-size: inherit;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        .mysql-account-grants {
            display: block; width: fit-content; max-width: 100%; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
            border: 0; padding: 0; background: transparent; font: inherit; color: inherit; text-align: left; cursor: pointer;
        }
        .mysql-account-grants:hover { color: ${token.colorPrimary}; }
        .mysql-account-grants:focus-visible { outline: 2px solid ${token.colorPrimaryBorder}; outline-offset: 2px; border-radius: 3px; }
        .mysql-account-no-grant { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: ${token.colorTextTertiary}; }
        .mysql-account-table .action-menu-open > .ant-table-cell { background: ${token.colorFillQuaternary}; }
        .client-footer { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; }
        .client-footer > .ant-typography { min-width: 0; overflow-wrap: anywhere; font-variant-numeric: tabular-nums; }

        @container mysql-accounts (max-width: 560px) {
            .mysql-account-tools { width: 100%; }
            .mysql-account-tools > :last-child { margin-left: auto; }
            .client-footer { gap: 8px; }
        }
        @media (max-width: 767px) {
            .mysql-account-search { flex: 1 1 100%; width: 100%; }
            .mysql-account-tools { width: 100%; }
            .mysql-account-tools > :last-child { margin-inline-start: auto; }
        }
    `,
    grantsTooltip: css`
        max-height: min(300px, 50dvh); overflow-y: auto; overflow-wrap: anywhere;
        overscroll-behavior: contain; scrollbar-width: thin;
        > div + div { margin-top: 4px; }
    `,
    permissions: css`
        .ant-modal-body { min-width: 0; font-size: ${compact ? 12 : 13}px; }
        .mysql-permission-section { padding: 11px 13px; margin-bottom: 8px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; }
        .mysql-permission-section:last-child { margin-bottom: 0; }
        .mysql-permission-heading { display: flex; align-items: baseline; flex-wrap: wrap; gap: 4px 12px; }
        .mysql-permission-heading > span { min-width: 0; flex: 1; color: ${token.colorText}; overflow-wrap: anywhere; }
        .mysql-permission-heading small { color: ${token.colorTextTertiary}; font-size: ${compact ? 10 : 11}px; }
        .mysql-permission-section p { margin: 7px 0 0; color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; line-height: 1.7; overflow-wrap: anywhere; }
        .mysql-permission-object { margin-top: 6px; color: ${token.colorTextSecondary}; font-family: ${token.fontFamilyCode}; font-size: ${compact ? 11 : 12}px; overflow-wrap: anywhere; }
        @media (max-width: 575px) {
            .mysql-permission-section { padding: 10px 11px; }
            .mysql-permission-heading > span { flex-basis: 100%; }
        }
    `,
}));
