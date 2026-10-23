import {createStyles} from "antd-style";

export default createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    maintenance: css`
        min-width: 0;
        .redis-maintenance-list { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorBgContainer}; overflow: hidden; }
        .redis-maintenance-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 12px; padding: ${compact ? 10 : 12}px ${compact ? 12 : 14}px; }
        .redis-maintenance-row + .redis-maintenance-row { border-top: 1px solid ${token.colorBorderSecondary}; }
        .redis-maintenance-row > div { display: flex; flex-direction: column; min-width: 0; gap: 2px; }
        .redis-maintenance-title { color: ${token.colorText}; font-size: ${compact ? 12 : 13}px; font-weight: 400; line-height: 1.6; }
        .redis-maintenance-row .ant-typography { color: ${token.colorTextSecondary}; font-size: ${compact ? 11 : 12}px; line-height: 1.6; overflow-wrap: anywhere; }
        .redis-maintenance-row > .ant-btn { min-width: ${compact ? 76 : 84}px; justify-self: end; }
        @media (max-width: 767px) {
            .redis-maintenance-row { padding-inline: 10px; gap: 10px; }
        }
    `,
    confirmation: css`display: grid; gap: 8px; overflow-wrap: anywhere; font-size: ${token.fontSize}px; line-height: 1.6;`,
}));
