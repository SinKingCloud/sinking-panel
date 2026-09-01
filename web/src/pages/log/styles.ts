import {createStyles} from "antd-style";

const useStyles = createStyles(({css}: any) => ({
        dateFilter: css`
            position: relative;
            display: inline-flex;
            flex: none;
        `,
        datePickerAnchor: css`
            && {
                position: absolute;
                right: 0;
                bottom: 0;
                width: 1px;
                min-width: 0;
                height: 1px;
                padding: 0;
                overflow: hidden;
                border: 0;
                opacity: 0;
                pointer-events: none;
            }
        `,
        datePickerPopup: css`
            && .ant-picker-range-arrow {
                right: 16px;
                left: auto !important;
            }

            && .ant-picker-panels > :last-child:not(:first-child) {
                display: none;
            }

            && .ant-picker-panels > :first-child .ant-picker-header-next-btn,
            && .ant-picker-panels > :first-child .ant-picker-header-super-next-btn {
                visibility: visible !important;
            }
        `,
        logTable: css`
            min-width: 0;

            .log-copy {
                max-width: 100%;
                margin: 0;
            }

            .log-type {
                max-width: 100%;
                margin: 0;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .log-text,
            .log-time {
                display: block;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        `,
}));

export default useStyles;
