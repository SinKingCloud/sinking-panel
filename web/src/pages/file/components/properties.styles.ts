import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean}>(({css, token}, props = {}) => {
    const compact = Boolean(props.compact);
    return {
        root: css`
            .ant-modal {
                max-width: calc(100vw - 24px);
            }

            .ant-modal-body {
                max-height: calc(100dvh - 104px);
                overflow-y: auto;
            }

            .file-properties-panel {
                min-width: 0;
            }

            .file-properties-overview {
                min-width: 0;
                padding: ${compact ? "2px 0 12px" : "2px 0 16px"};
                display: flex;
                align-items: flex-start;
                gap: ${compact ? 8 : 10}px;
            }

            .file-properties-icon {
                width: ${compact ? 22 : 24}px;
                height: ${compact ? 22 : 24}px;
                flex: none;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 15 : 16}px;
            }

            .file-properties-icon.is-directory {
                color: ${token.colorPrimary};
            }

            .file-properties-identity {
                min-width: 0;
                flex: 1;
            }

            .file-properties-name-row,
            .file-properties-value-row {
                min-width: 0;
                display: flex;
                gap: ${compact ? 2 : 4}px;
            }

            .file-properties-name-row {
                align-items: center;
            }

            .file-properties-value-row {
                align-items: flex-start;
            }

            .file-properties-name {
                min-width: 0;
                flex: 1;
                display: -webkit-box;
                overflow: hidden;
                overflow-wrap: anywhere;
                word-break: break-word;
                color: ${token.colorTextHeading};
                font-size: ${compact ? 15 : 16}px;
                font-weight: 500;
                line-height: ${compact ? 20 : 22}px;
                user-select: text;
                -webkit-box-orient: vertical;
                -webkit-line-clamp: 2;
            }

            .file-properties-action.ant-btn {
                width: ${compact ? 26 : 28}px;
                min-width: ${compact ? 26 : 28}px;
                height: ${compact ? 26 : 28}px;
                flex: none;
                padding: 0;
                color: ${token.colorTextSecondary};
            }

            .file-properties-action.ant-btn:hover,
            .file-properties-action.ant-btn:focus-visible {
                color: ${token.colorPrimary};
            }

            .file-properties-details {
                min-width: 0;
                display: grid;
                grid-template-columns: repeat(2, minmax(0, 1fr));
                column-gap: ${compact ? 20 : 28}px;
                row-gap: ${compact ? 12 : 16}px;
            }

            .file-properties-item {
                min-width: 0;
            }

            .file-properties-item.is-wide {
                grid-column: 1 / -1;
            }

            .file-properties-label {
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 11 : 12}px;
                line-height: 18px;
            }

            .file-properties-value {
                min-width: 0;
                margin-top: ${compact ? 2 : 4}px;
                overflow-wrap: anywhere;
                color: ${token.colorText};
                font-size: ${compact ? 12 : 13}px;
                font-weight: 500;
                line-height: ${compact ? 18 : 20}px;
            }

            .file-properties-value.is-time {
                font-variant-numeric: tabular-nums;
                white-space: nowrap;
            }

            .file-properties-value.is-permission {
                flex: none;
                font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
                font-variant-numeric: tabular-nums;
            }

            .file-properties-value-row .file-properties-path {
                flex: 1;
            }

            .file-properties-path {
                max-width: 100%;
                max-height: ${compact ? 56 : 60}px;
                overflow-y: auto;
                overflow-wrap: anywhere;
                word-break: break-word;
                color: ${token.colorTextSecondary};
                font-weight: 400;
                white-space: normal;
                user-select: text;
                scrollbar-width: thin;
            }

            @media (min-width: 768px) {
                .ant-modal-body {
                    max-height: calc(100dvh - 256px);
                }
            }

            @media (max-width: 440px) {
                .file-properties-name-row {
                    justify-content: flex-end;
                    flex-wrap: wrap;
                    row-gap: 2px;
                }

                .file-properties-name {
                    flex-basis: 100%;
                    -webkit-line-clamp: 3;
                }

                .file-properties-details {
                    grid-template-columns: minmax(0, 1fr);
                }

                .file-properties-item.is-wide {
                    grid-column: auto;
                }

                .file-properties-value.is-time {
                    white-space: normal;
                }

                .file-properties-path {
                    max-height: 80px;
                }
            }
        `,
    };
});

export default useStyles;
