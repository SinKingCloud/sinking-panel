import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}: any) => ({
    table: css`
        .container-text,
        .container-name,
        .container-time {
            display: block;
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .container-name {
            padding: 0;
            border: 0;
            background: transparent;
            color: inherit;
            font: inherit;
            text-align: left;
            cursor: pointer;
        }

        .container-name:hover,
        .container-name:focus-visible { color: ${token.colorPrimary}; }
        .container-name:disabled { color: ${token.colorTextDisabled}; cursor: default; }
        .container-time { color: ${token.colorTextSecondary}; font-variant-numeric: tabular-nums; }
        .container-muted { color: ${token.colorTextTertiary}; }

        .container-state {
            display: inline-flex;
            align-items: center;
            gap: 7px;
            color: ${token.colorTextSecondary};
            white-space: nowrap;
        }

        .container-state::before {
            content: "";
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: currentColor;
        }

        .container-state.running { color: ${token.colorSuccess}; }
        .container-state.failed { color: ${token.colorError}; }
        .container-state.starting,
        .container-state.stopping { color: ${token.colorWarning}; }
    `,
    detail: css`
        display: flex;
        flex-direction: column;
        min-width: 0;
        min-height: ${token.controlHeight * 12}px;
        color: ${token.colorText};
        font-size: 12px;
        line-height: 1.5;
        overflow-wrap: anywhere;

        .detail-loading {
            display: grid;
            min-height: ${token.controlHeight * 12}px;
            place-items: center;
        }

        .detail-data { display: flex; flex-direction: column; flex: 1; min-width: 0; font-size: inherit; line-height: inherit; }
        .detail-overview {
            display: grid;
            grid-template-columns: repeat(2, minmax(0, 1fr));
            align-content: start;
            gap: ${token.marginSM}px ${token.margin}px;
            margin: 0;
            padding: ${token.paddingSM}px;
            background: ${token.colorFillQuaternary};
            border-radius: ${token.borderRadius}px;
        }
        .detail-overview > div { display: grid; align-content: start; gap: 2px; min-width: 0; }
        .detail-overview-wide { grid-column: 1 / -1; }
        .detail-overview dt { color: ${token.colorTextSecondary}; font-size: 11px; }
        .detail-overview dd { min-width: 0; margin: 0; font-variant-numeric: tabular-nums; }
        .detail-overview .ant-tag { margin: 0; font-size: inherit; line-height: 18px; }

        .detail-tags { display: flex; flex-wrap: wrap; gap: 6px; }
        .detail-tags .ant-tag { max-width: 100%; margin: 0; white-space: normal; overflow-wrap: anywhere; }

        .ant-descriptions, .ant-typography, .ant-alert, .ant-alert-title, .ant-alert-description { font-size: inherit; line-height: inherit; font-weight: 400; }
        .ant-typography { display: block; position: relative; padding-inline-end: 24px; font-family: inherit; font-size: inherit; white-space: pre-wrap; overflow-wrap: anywhere; }
        .ant-typography-copy { position: absolute; inset-inline-end: 0; top: 0; color: ${token.colorTextTertiary}; }
        .ant-typography-copy:hover { color: ${token.colorPrimary}; }
        .detail-switch { display: flex; flex-shrink: 0; width: fit-content; max-width: 100%; margin: 0 auto ${token.marginSM}px; font-family: inherit; font-size: ${Math.max(token.fontSizeSM, 12)}px; font-weight: 400; }
        .detail-content { display: flex; flex-direction: column; flex: 1 1 auto; min-width: 0; min-height: ${token.controlHeight * 9}px; font-size: 12px; line-height: 18px; }

        .detail-config { display: grid; gap: ${token.marginXS}px; margin: 0; }
        .ant-descriptions { padding: ${token.paddingXS}px ${token.paddingSM}px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        .ant-descriptions-row > th { padding-bottom: 4px; }
        .ant-descriptions-row:last-child > td { padding-bottom: 0; }
        .detail-label, .ant-descriptions-item-label { color: ${token.colorTextSecondary}; font-size: 11px; line-height: 16px; }
        .ant-descriptions-item-content { min-width: 0; font-size: 12px; line-height: 18px; font-variant-numeric: tabular-nums; white-space: pre-wrap; overflow-wrap: anywhere; }
        .ant-descriptions .ant-tag { margin: 0; font-size: inherit; }

        .detail-field { display: grid; align-content: start; gap: 4px; min-width: 0; }
        .detail-code { min-width: 0; margin: 0; padding: ${token.paddingXS}px ${token.paddingSM}px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        .detail-command .detail-code { height: ${token.controlHeight * 3}px; box-sizing: border-box; overflow-x: hidden; overflow-y: auto; scrollbar-width: thin; }
        .detail-field-heading { display: flex; align-items: center; justify-content: space-between; gap: ${token.marginXS}px; }
        .detail-muted { color: ${token.colorTextTertiary}; font-size: 11px; line-height: 16px; }
        .detail-mount { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: ${token.marginXS}px; margin: 0; padding: ${token.paddingXS}px ${token.paddingSM}px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
        .detail-mount dd { min-width: 0; margin: 0; }
        .detail-mounts { display: grid; gap: ${token.marginXS}px; margin: 0; }
        .ant-alert { margin-bottom: ${token.marginSM}px; }

        @media (max-width: 575px) {
            .detail-overview, .detail-mount { grid-template-columns: minmax(0, 1fr); }
            .detail-overview { row-gap: ${token.marginXS}px; }
            .detail-overview > div, .detail-mount > .detail-field, .detail-command {
                grid-template-columns: 76px minmax(0, 1fr);
                align-items: baseline;
                column-gap: 8px;
            }
            .detail-overview dt { white-space: nowrap; }
            .detail-mount .detail-field-heading { flex-wrap: wrap; justify-content: flex-start; gap: 2px 6px; }
            .detail-command { align-items: start; padding: ${token.paddingXS}px ${token.paddingSM}px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; }
            .detail-command .detail-code { padding: 0; background: transparent; }
            .detail-switch .ant-segmented-item-label { padding-inline: 6px; }
        }
    `,
    imageDetail: css`
        min-height: ${token.controlHeight * 8}px;
        .detail-content { min-height: ${token.controlHeight * 6}px; }
        .detail-loading { min-height: ${token.controlHeight * 8}px; }
        .detail-command .detail-code { height: ${token.controlHeight * 2}px; }
    `,
}));

export default useStyles;
