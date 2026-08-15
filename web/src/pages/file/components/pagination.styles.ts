import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean; dark?: boolean}>(({
    css,
    token,
    isDarkMode,
}, props = {}) => {
    const compact = Boolean(props?.compact);
    const dark = typeof props?.dark === "boolean" ? props.dark : Boolean(isDarkMode);

    return {
        paginationCard: css`
            container-name: file-pagination;
            container-type: inline-size;
            width: 100%;
            margin-top: ${compact ? 10 : 12}px;
            border-radius: ${token.borderRadiusLG}px;
            background: ${token.colorBgContainer};
            box-shadow: ${token.boxShadowTertiary};

            .ant-card-body {
                padding: ${compact ? "10px 12px" : "14px 16px"} !important;
            }
        `,
        pagination: css`
            display: flex;
            align-items: center;
            justify-content: center;
            width: 100%;
            margin: 0 !important;
            background: transparent;

            &.ant-pagination,
            .ant-pagination {
                width: 100%;
                justify-content: center;
            }

            .ant-pagination-total-text {
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 12 : 13}px;
            }

            .ant-pagination-options {
                margin-inline-start: ${compact ? 10 : 12}px;
            }

            .ant-pagination-item,
            .ant-pagination-prev .ant-pagination-item-link,
            .ant-pagination-next .ant-pagination-item-link {
                min-width: ${compact ? 22 : 26}px;
                height: ${compact ? 22 : 26}px;
                margin-inline-end: 4px;
                border: 0;
                border-radius: ${token.borderRadiusSM}px;
                background: transparent;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
                line-height: ${compact ? 22 : 26}px;
            }

            .ant-pagination-item a {
                color: inherit;
            }

            .ant-pagination-item:hover,
            .ant-pagination-prev:hover .ant-pagination-item-link,
            .ant-pagination-next:hover .ant-pagination-item-link {
                background: ${token.colorFillTertiary};
                color: ${token.colorTextSecondary};
            }

            .ant-pagination-item-active,
            .ant-pagination-item-active:hover {
                background: ${token.colorBgContainer};
                box-shadow: inset 0 0 0 1px ${token.colorPrimaryBorder};
                color: ${token.colorPrimary};
            }

            .ant-pagination-item-active a,
            .ant-pagination-item-active:hover a {
                color: ${token.colorPrimary};
            }

            .ant-pagination-options-size-changer.ant-select .ant-select-selector {
                height: ${compact ? 22 : 26}px !important;
                padding-inline: 7px 22px !important;
                border: 0 !important;
                border-radius: ${token.borderRadiusSM}px !important;
                background: ${token.colorBgContainer} !important;
                box-shadow: inset 0 0 0 1px ${token.colorBorderSecondary} !important;
            }

            .ant-pagination-options-size-changer.ant-select,
            .ant-pagination-options-size-changer .ant-select-content,
            .ant-pagination-options-size-changer .ant-select-content-value,
            .ant-pagination-options-size-changer .ant-select-selection-item,
            .ant-pagination-options-quick-jumper {
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
                line-height: ${compact ? 22 : 26}px;
            }

            .ant-pagination-options-quick-jumper input {
                width: ${compact ? 30 : 34}px;
                height: ${compact ? 22 : 26}px;
                margin-inline: 4px;
                border-color: transparent;
                border-radius: ${token.borderRadiusSM}px;
                background: ${dark ? "rgba(255, 255, 255, 0.1)" : "rgba(0, 0, 0, 0.05)"} !important;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
            }

            @container file-pagination (max-width: 767px) {
                .ant-pagination-total-text,
                .ant-pagination-options-quick-jumper {
                    display: none;
                }

                .ant-pagination-options {
                    display: block;
                    margin-inline-start: ${compact ? 8 : 10}px;
                }
            }
        `,
        pageSizeSelect: css`
            && {
                height: ${compact ? 22 : 26}px;
                color: ${token.colorTextSecondary} !important;
                font-size: ${compact ? 11 : 12}px;
            }

            && .ant-select-selector {
                height: ${compact ? 22 : 26}px !important;
                padding-inline: 7px 22px !important;
                border: 0 !important;
                border-radius: ${token.borderRadiusSM}px !important;
                background: ${token.colorBgContainer} !important;
                box-shadow: inset 0 0 0 1px ${token.colorBorderSecondary} !important;
            }

            && .ant-select-content,
            && .ant-select-content-value,
            && .ant-select-selection-item {
                color: ${token.colorTextSecondary} !important;
                font-size: ${compact ? 11 : 12}px !important;
                font-weight: 400 !important;
                line-height: ${compact ? 22 : 26}px !important;
            }

            && .ant-select-arrow {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 9 : 10}px;
            }
        `,
        pageSizeDropdown: css`
            && {
                padding: 2px;
                border-radius: ${token.borderRadius}px;
            }

            && .ant-select-item {
                min-height: ${compact ? 24 : 28}px;
                padding: 3px 6px;
                border-radius: ${token.borderRadiusSM}px;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
                line-height: ${compact ? 18 : 20}px;
            }

            && .ant-select-item-option-content {
                font-size: ${compact ? 11 : 12}px !important;
                font-weight: 400;
            }

            && .ant-select-item-option-selected:not(.ant-select-item-option-disabled) {
                background: ${token.colorFillTertiary};
                color: ${token.colorTextSecondary};
                font-weight: 400;
            }
        `,
    };
});

export default useStyles;
