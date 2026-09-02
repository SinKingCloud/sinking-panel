import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}, props: {compact?: boolean} = {}) => {
    const compact = Boolean(props.compact);
    return {
        pathSection: css`
            min-width: 0;
            padding: 0 ${compact ? 10 : 12}px ${compact ? 10 : 12}px;
            background: ${token.colorBgContainer};

            @container data-table-workspace (max-width: 560px) {
                padding: 0 10px 10px;
            }
        `,
        pathBar: css`
            min-width: 0;
            min-height: ${compact ? 28 : 32}px;
            display: flex;
            align-items: center;
            gap: 6px;

            > .path-back {
                width: ${compact ? 28 : 30}px;
                height: ${compact ? 28 : 30}px;
                padding: 0;
                flex: none;
                border: 1px solid transparent;
                border-radius: ${token.borderRadiusSM}px;
                background: transparent;
                color: ${token.colorTextTertiary};
            }

            > .path-back:not(:disabled):hover {
                background: ${token.colorFillQuaternary};
                color: ${token.colorPrimary};
            }

            > .path-back:focus-visible {
                border-color: ${token.colorPrimaryBorder};
                color: ${token.colorPrimary};
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            > .path-back:disabled {
                background: transparent;
                color: ${token.colorTextQuaternary};
            }

            .path-breadcrumb {
                min-width: 40px;
                margin: 0;
                padding: 0 0 0 8px;
                display: flex;
                flex: 1;
                align-items: center;
                overflow-x: auto;
                overflow-y: hidden;
                border-inline-start: 1px solid ${token.colorSplit};
                list-style: none;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
                white-space: nowrap;
                -webkit-overflow-scrolling: touch;
            }

            .path-breadcrumb::-webkit-scrollbar,
            .path-actions::-webkit-scrollbar {
                display: none;
            }

            .path-breadcrumb > li {
                min-width: 0;
                display: inline-flex;
                flex: 0 0 auto;
                align-items: center;
            }

            .path-segment,
            .path-current {
                min-width: 0;
                max-width: 220px;
                height: ${compact ? 28 : 30}px;
                padding: 0 6px;
                display: inline-flex;
                align-items: center;
                border: 0;
                border-radius: ${token.borderRadiusSM}px;
                outline: none;
                background: transparent;
                color: ${token.colorTextSecondary};
                font: inherit;
                font-size: ${compact ? 11 : 12}px;
            }

            .path-segment > span,
            .path-current > span {
                min-width: 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .path-segment {
                cursor: pointer;
            }

            .path-segment:hover,
            .path-current:hover,
            .path-segment:focus-visible,
            .path-current:focus-visible {
                color: ${token.colorPrimary};
            }

            .path-segment:focus-visible,
            .path-current:focus-visible {
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            .path-current {
                color: ${token.colorText};
                cursor: text;
                font-weight: 500;
            }

            .path-current:disabled {
                color: ${token.colorText};
                cursor: default;
            }

            > .path-input.ant-input {
                width: min(100%, ${compact ? 560 : 640}px);
                min-width: 40px;
                max-width: 100%;
                height: ${compact ? 28 : 30}px;
                padding: 0 7px;
                flex: 0 1 ${compact ? 560 : 640}px;
                border: 1px solid ${token.colorPrimaryBorder};
                border-radius: ${token.borderRadiusSM}px;
                background: ${token.colorBgContainer};
                color: ${token.colorText};
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
                font-size: ${compact ? 11 : 12}px;
            }

            .path-actions {
                max-width: min(58%, 230px);
                min-width: 0;
                margin-inline-start: auto;
                display: flex;
                flex: none;
                align-items: center;
                gap: 6px;
                overflow-x: auto;
                scrollbar-width: none;
            }

            .path-actions > .path-editor.ant-btn,
            .path-actions > .path-batch,
            .path-actions > .path-paste.ant-btn {
                min-width: max-content;
                height: ${compact ? 28 : 30}px;
                flex: none;
            }

            .path-separator {
                margin-inline: 1px;
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 8 : 9}px;
            }

            @container data-table-workspace (max-width: 560px) {
                gap: 4px;

                .path-segment,
                .path-current {
                    max-width: 140px;
                }

                .path-current {
                    max-width: min(55cqw, 220px);
                }

                .path-actions {
                    max-width: min(64%, 206px);
                    gap: 4px;
                }
            }
        `,
    };
});

export default useStyles;
