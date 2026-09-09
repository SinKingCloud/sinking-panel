import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token, isDarkMode}: any, props: any = {}) => {
    const compact = Boolean(props?.compact);
    const dark = typeof props?.dark === "boolean" ? props.dark : Boolean(isDarkMode);
    const highlight = dark
        ? `color-mix(in srgb, #fff 5%, ${token.colorBgContainer})`
        : `color-mix(in srgb, #000 3%, ${token.colorBgContainer})`;

    return {
        page: css`
            width: 100%;
            max-width: 1450px;
            margin: 0 auto;
        `,
        workspace: css`
            container-name: data-table-workspace;
            container-type: inline-size;
            min-width: 0;
            overflow: hidden;
            border-radius: ${token.borderRadiusLG}px;
            background: ${token.colorBgContainer};
            box-shadow: ${token.boxShadowTertiary};
        `,
        hero: css`
            position: relative;
            min-height: ${compact ? 104 : 118}px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            overflow: hidden;
            isolation: isolate;
            background: linear-gradient(
                118deg,
                color-mix(in srgb, ${token.colorPrimary}, ${token.colorBgContainer} ${dark ? 94 : 93}%) 0%,
                ${token.colorBgContainer} 58%,
                color-mix(in srgb, #13c2c2, ${token.colorBgContainer} ${dark ? 97 : 96}%) 100%
            );

            &.plain-background { background: ${token.colorBgContainer}; }
            &.plain-background::before, &.plain-background::after { display: none; }
            &.without-action .hero-copy { max-width: 100%; }

            &::before {
                position: absolute;
                z-index: 0;
                inset: 0;
                background-image:
                    linear-gradient(${dark ? "rgba(255,255,255,.018)" : "rgba(72,132,202,.035)"} 1px, transparent 1px),
                    linear-gradient(90deg, ${dark ? "rgba(255,255,255,.018)" : "rgba(72,132,202,.035)"} 1px, transparent 1px);
                background-size: 24px 24px;
                content: "";
                -webkit-mask-image: linear-gradient(90deg, transparent 36%, rgba(0, 0, 0, .2) 58%, #000 100%);
                mask-image: linear-gradient(90deg, transparent 36%, rgba(0, 0, 0, .2) 58%, #000 100%);
                pointer-events: none;
            }

            &::after {
                position: absolute;
                z-index: 0;
                inset: -45% -10% -55% 48%;
                background:
                    radial-gradient(ellipse at 48% 32%, color-mix(in srgb, ${token.colorPrimary}, transparent 84%) 0%, transparent 60%),
                    radial-gradient(ellipse at 76% 72%, color-mix(in srgb, #13c2c2, transparent 89%) 0%, transparent 58%);
                content: "";
                filter: blur(28px);
                opacity: ${dark ? .34 : .52};
                pointer-events: none;
            }

            .hero-copy {
                position: relative;
                z-index: 2;
                min-width: 0;
                max-width: calc(100% - 178px);
                padding: ${compact ? "14px 18px" : "18px 22px"};
            }

            .eyebrow {
                display: flex;
                align-items: center;
                gap: 7px;
                color: color-mix(in srgb, ${token.colorPrimary}, ${token.colorTextTertiary} 62%);
                font-size: 9px;
                font-weight: 600;
                line-height: 16px;
            }

            .status-dot {
                width: 5px;
                height: 5px;
                flex: none;
                border-radius: 50%;
                background: color-mix(in srgb, ${token.colorPrimary}, #fff 26%);
                box-shadow: 0 0 0 4px color-mix(in srgb, ${token.colorPrimary}, transparent 90%);
            }

            h1 {
                margin: 3px 0 0;
                color: ${token.colorTextHeading};
                font-size: ${compact ? 19 : 20}px;
                font-weight: 600;
                line-height: ${compact ? 26 : 28}px;
            }

            .hero-visual {
                position: absolute;
                z-index: 1;
                top: 0;
                right: 82px;
                width: 52%;
                height: 100%;
                opacity: ${dark ? .12 : .25};
                filter: ${dark ? "blur(.8px) saturate(.78)" : "blur(.25px)"};
                -webkit-mask-image: linear-gradient(90deg, transparent 0%, rgba(0, 0, 0, .38) 28%, #000 52%);
                mask-image: linear-gradient(90deg, transparent 0%, rgba(0, 0, 0, .38) 28%, #000 52%);
                pointer-events: none;
            }

            .create-button {
                position: relative;
                z-index: 3;
                min-width: 88px;
                height: ${compact ? 32 : 36}px;
                margin: 0 ${compact ? 18 : 24}px 0 0;
                padding: 0 ${compact ? 14 : 17}px;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                gap: 6px;
                flex: none;
                border: 0;
                border-radius: ${token.borderRadius}px;
                outline: none;
                background: ${token.colorPrimary};
                color: #fff;
                cursor: pointer;
                font: inherit;
                font-size: ${compact ? 12 : 13}px;
                font-weight: 500;
                white-space: nowrap;
                transition: background-color .18s ease;
            }

            .create-button:hover,
            .create-button:focus-visible {
                background: ${token.colorPrimaryHover};
            }

            .create-button:focus-visible {
                box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
            }

            .create-button:active {
                background: ${token.colorPrimaryActive};
            }

            .create-button:disabled {
                background: ${token.colorFill};
                color: ${token.colorTextDisabled};
                cursor: not-allowed;
            }

            .create-button .suffix-icon {
                font-size: 10px;
                opacity: .72;
                transform: rotate(90deg);
            }

            @container data-table-workspace (max-width: 600px) {
                min-height: ${compact ? 104 : 116}px;

                .hero-copy {
                    max-width: calc(100% - 145px);
                    padding: 16px 14px;
                }

                h1 {
                    font-size: 18px;
                    line-height: 25px;
                }

                .create-button {
                    margin-right: 14px;
                }

                .hero-visual {
                    right: 32px;
                    width: 68%;
                    opacity: ${dark ? .07 : .14};
                }

                .hero-visual .flow-secondary,
                .hero-visual .flow-detail {
                    display: none;
                }
            }

            @container data-table-workspace (max-width: 430px) {
                .create-button {
                    margin-right: 12px;
                    padding-inline: 12px;
                }

                .hero-visual {
                    right: 16px;
                    width: 76%;
                    opacity: ${dark ? .045 : .09};
                }
            }

            @container data-table-workspace (max-width: 350px) {
                .hero-visual {
                    display: none;
                }
            }
        `,
        commandBar: css`
            min-width: 0;
            min-height: ${compact ? 56 : 66}px;
            padding: ${compact ? "12px 10px 10px" : "17px 12px 13px"};
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: ${compact ? 12 : 16}px;
            background: ${token.colorBgContainer};

            &.without-search .command-actions { width: 100%; margin-left: 0; }
            &.without-search .command-actions-scroll:empty { flex: 0; }
            &.without-search .command-actions > .ant-tooltip-open:last-child,
            &.without-search .command-actions > .ant-btn:last-child { margin-left: auto; }
            .table-selection-action { flex: none; }

            .command-content {
                min-width: 0;
                flex: 1;
                overflow-x: auto;
                overflow-y: hidden;
                scrollbar-width: none;
                white-space: nowrap;
            }
            .command-content::-webkit-scrollbar { display: none; }
            &.with-content .command-actions { flex: none; max-width: 100%; }

            .command-actions {
                min-width: 0;
                margin-left: auto;
                display: flex;
                flex: 0 1 auto;
                align-items: center;
                gap: 6px;
            }

            .command-actions-scroll {
                min-width: 0;
                display: flex;
                flex: 1 1 auto;
                align-items: center;
                gap: 6px;
                overflow-x: auto;
                overflow-y: hidden;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
                -webkit-overflow-scrolling: touch;
            }

            .command-actions-scroll::-webkit-scrollbar {
                display: none;
            }

            .command-actions-scroll > * {
                flex: none;
            }

            @container data-table-workspace (max-width: 680px) {
                min-height: auto;
                padding: ${compact ? 8 : 11}px;
                display: grid;
                grid-template-columns: minmax(0, 1fr);
                gap: 6px;

                > .ant-input-affix-wrapper {
                    width: 100%;
                    max-width: none;
                }

                .command-actions {
                    width: 100%;
                    margin-left: 0;
                    gap: 4px;
                }

                .command-actions-scroll {
                    gap: 4px;
                }

                &.with-content { display: flex; }
                &.with-content .command-actions { width: auto; margin-left: auto; }
            }

            @container data-table-workspace (max-width: 430px) {
                padding: ${compact ? 8 : 10}px;
            }
        `,
        searchBox: css`
            flex: 0 1 ${compact ? 300 : 320}px;
            width: ${compact ? 300 : 320}px;
            max-width: 100%;
            min-width: 0;

            &.ant-input-affix-wrapper {
                height: ${compact ? 30 : 34}px;
                padding: 0 ${compact ? 9 : 12}px;
                border-radius: ${token.borderRadiusSM}px;
            }

            .ant-input-prefix {
                margin-right: 8px;
                color: ${token.colorTextTertiary};
                font-size: 14px;
            }

            .ant-input,
            .ant-input::placeholder {
                font-size: ${compact ? 11 : 12}px;
            }

            .ant-input {
                color: ${token.colorTextSecondary};
            }

            .ant-input::placeholder {
                color: ${token.colorTextQuaternary};
            }

            @container data-table-workspace (max-width: 820px) {
                flex-basis: ${compact ? 260 : 280}px;
                width: ${compact ? 260 : 280}px;
            }

            @container data-table-workspace (max-width: 680px) {
                width: 100%;
                max-width: none;
                flex: none;
            }
        `,
        toolbarTrigger: css`
            display: inline-flex;
            flex: none;
            align-items: center;
            justify-content: center;
            gap: 5px;
            width: auto;
            min-width: max-content;
            height: ${compact ? 30 : 34}px;
            padding: 0 ${compact ? 7 : 10}px;
            border: 1px solid transparent;
            border-radius: ${token.borderRadiusSM}px;
            outline: none;
            background: ${token.colorFillQuaternary};
            color: ${token.colorTextSecondary};
            cursor: pointer;
            font: inherit;
            transition: background-color .16s ease, color .16s ease;

            .marker {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 11 : 12}px;
            }

            .value {
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
                white-space: nowrap;
            }

            .value-stack {
                min-width: 0;
                max-width: 160px;
                display: grid;
            }

            .value-stack > .value {
                min-width: 0;
                grid-area: 1 / 1;
            }

            .value-stack > .measure {
                visibility: hidden;
                pointer-events: none;
            }

            .value-stack > .active {
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .arrow {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 9 : 10}px;
                transition: transform .16s ease;
            }

            &:hover,
            &.ant-dropdown-open {
                background: ${token.colorFillSecondary};
            }

            &.ant-dropdown-open .arrow {
                transform: rotate(180deg);
            }

            &:focus-visible {
                border-color: ${token.colorPrimaryBorder};
                background: ${token.colorBgContainer};
                box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
            }

            &:disabled {
                color: ${token.colorTextDisabled};
                cursor: not-allowed;
            }
        `,
        toolbarAction: css`
            && {
                width: auto;
                min-width: max-content;
                height: ${compact ? 30 : 34}px;
                padding: 0 ${compact ? 7 : 10}px;
                flex: none;
                gap: 5px;
                border: 1px solid transparent;
                border-radius: ${token.borderRadiusSM}px;
                background: ${token.colorFillQuaternary};
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 11 : 12}px;
                box-shadow: none;
            }

            && .value {
                color: ${token.colorTextSecondary};
                white-space: nowrap;
            }

            && .arrow {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 9 : 10}px;
                transition: transform .16s ease;
            }

            &&:not(:disabled):not(.ant-btn-disabled):is(:hover, :active, .ant-dropdown-open) {
                border-color: transparent;
                background: ${token.colorFillSecondary};
                color: ${token.colorTextTertiary};

                &.ant-btn-dangerous,
                &.ant-btn-color-dangerous {
                    color: ${token.colorError};
                }
            }

            &&.ant-dropdown-open .arrow {
                transform: rotate(180deg);
            }

            &&:focus-visible {
                border-color: ${token.colorPrimaryBorder};
                background: ${token.colorBgContainer};
                box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
            }

            &&:is(.ant-btn-dangerous, .ant-btn-color-dangerous),
            &&:is(.ant-btn-dangerous, .ant-btn-color-dangerous) .value { color: ${token.colorError}; }

            &&:is(:disabled, .ant-btn-disabled) {
                background: ${token.colorFillQuaternary};
                color: ${token.colorTextDisabled};
            }
        `,
        toolbarIconButton: css`
            && {
                width: ${compact ? 30 : 34}px;
                min-width: ${compact ? 30 : 34}px;
                height: ${compact ? 30 : 34}px;
                padding: 0;
                flex: none;
                border: 1px solid transparent;
                border-radius: ${token.borderRadiusSM}px;
                background: ${token.colorFillQuaternary};
                color: ${token.colorTextSecondary};
                box-shadow: none;
            }

            && .anticon {
                font-size: ${compact ? 12 : 13}px;
            }

            &&:is(.ant-btn-dangerous, .ant-btn-color-dangerous) {
                color: ${token.colorError};
            }

            &&:disabled,
            &&.ant-btn-disabled {
                color: ${token.colorTextDisabled};
            }

            &&:not(:disabled):not(.ant-btn-disabled):is(:hover, :focus, :active, .ant-dropdown-open) {
                border-color: transparent;
                background: ${token.colorFillSecondary};
                color: ${token.colorTextSecondary};
                box-shadow: none;

                &.ant-btn-dangerous,
                &.ant-btn-color-dangerous {
                    color: ${token.colorError};
                }
            }
        `,
        toolbarDropdown: css`
            && {
                max-width: calc(100vw - 24px);
                padding: 0;
                border-radius: ${token.borderRadius}px;
            }

            && .ant-dropdown-menu {
                width: 100%;
                min-width: 0 !important;
                padding: 2px !important;
                border-radius: ${token.borderRadius}px !important;
                background: ${token.colorBgElevated};
                box-shadow: ${token.boxShadowSecondary};
            }

            && .ant-dropdown-menu-item {
                min-height: ${compact ? 24 : 30}px !important;
                margin: 0 !important;
                padding: 0 6px !important;
                border-radius: ${token.borderRadiusSM}px !important;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px !important;
                line-height: ${compact ? 24 : 30}px !important;
            }

            && .ant-dropdown-menu-title-content {
                min-width: 0;
                overflow: hidden;
                font-size: inherit !important;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            && .ant-dropdown-menu-item:hover {
                background: ${token.colorFillQuaternary};
            }

            && .ant-dropdown-menu-item-selected,
            && .ant-dropdown-menu-item-selected:hover {
                background: color-mix(in srgb, ${token.colorPrimary}, ${token.colorBgElevated} 94%);
                color: ${token.colorTextSecondary};
            }
        `,
        contentBar: css`
            box-sizing: content-box;
            min-width: 0;
            min-height: ${compact ? 30 : 34}px;
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 0 ${compact ? 10 : 12}px ${compact ? 10 : 12}px;

            .table-bar-content {
                min-width: 0;
                flex: 1;
                overflow-x: auto;
                scrollbar-width: none;
            }
            .table-bar-content::-webkit-scrollbar { display: none; }
            .table-bar-actions {
                min-width: 0;
                max-width: min(58%, 300px);
                margin-left: auto;
                display: flex;
                flex: none;
                align-items: center;
                gap: 6px;
            }
            .table-bar-actions-scroll {
                min-width: 0;
                display: flex;
                flex: 1 1 auto;
                align-items: center;
                gap: 6px;
                overflow-x: auto;
                overflow-y: hidden;
                scrollbar-width: none;
                overscroll-behavior-inline: contain;
            }
            .table-bar-actions-scroll::-webkit-scrollbar { display: none; }
            .table-bar-actions-scroll > * { flex: none; }
            .table-bar-actions .table-selection-action {
                min-width: ${compact ? 104 : 116}px;
                font-variant-numeric: tabular-nums;
            }
            @container data-table-workspace (max-width: 560px) {
                padding: 0 10px 10px;
                gap: 4px;
                .table-bar-actions { max-width: min(64%, 230px); gap: 4px; }
            }
        `,
        dateFilter: css`
            position: relative;
            display: inline-flex;
            flex: none;
        `,
        datePickerAnchor: css`
            && {
                position: absolute;
                inset: 0;
                width: 100%;
                min-width: 0;
                height: 100%;
                padding: 0;
                overflow: hidden;
                border: 0;
                opacity: 0;
                pointer-events: none;
            }
        `,
        datePickerPopup: css`
            /* 与弹层共用避让后的箭头坐标，不使用隐藏日期输入框的偏移。 */
            && .ant-picker-range-arrow {
                left: clamp(${token.borderRadiusLG}px, calc(var(--arrow-x, 50%) - ${token.sizePopupArrow / 2}px), calc(100% - ${token.borderRadiusLG + token.sizePopupArrow}px)) !important;
                right: auto;
                padding-inline: 0;
                transition: none;

                &::before { inset-inline-start: 0; }
            }
            && .ant-picker-panels > :last-child:not(:first-child) { display: none; }
            && .ant-picker-panels > :first-child .ant-picker-header-next-btn,
            && .ant-picker-panels > :first-child .ant-picker-header-super-next-btn { visibility: visible !important; }
        `,
        dataPanel: css`
            position: relative;
            min-width: 0;
            padding: 0 ${compact ? 10 : 12}px ${compact ? 10 : 12}px;

            @container data-table-workspace (max-width: 430px) {
                padding: 0 10px 10px;
            }
        `,
        state: css`
            min-height: ${compact ? 200 : 240}px;
            padding: ${compact ? 18 : 24}px;
            display: flex;
            align-items: center;
            justify-content: center;

            .ant-btn {
                height: ${compact ? 27 : 30}px;
                border-color: ${token.colorBorderSecondary};
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
            }
        `,
        table: css`
            --page-table-text: ${dark ? "rgba(255,255,255,.65)" : "rgba(0,0,0,.65)"};
            min-width: 0;
            flex: none;
            overflow: hidden;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusLG}px;
            background: ${token.colorBgContainer};

            .ant-table,
            .ant-table-container {
                background: transparent;
            }

            .ant-table-container {
                overflow: hidden;
                border-radius: inherit;
            }

            .ant-table-thead > tr > th {
                padding: ${compact ? 8 : 10}px !important;
                border-bottom: 1px solid ${token.colorBorderSecondary};
                background: ${highlight};
                color: var(--page-table-text);
                font-size: ${token.fontSize}px;
            }

            .ant-table-thead > tr > th.ant-table-column-sort,
            .ant-table-thead > tr > th.ant-table-column-has-sorters:hover,
            .ant-table-thead > tr > th.action-cell,
            .ant-table-thead > tr > .ant-table-cell-fix-end,
            .ant-table-thead > tr > .ant-table-cell-fix-right {
                background: ${highlight} !important;
            }

            .ant-table-column-sorters {
                justify-content: flex-start;
                gap: 5px;
            }

            .ant-table-column-sorter {
                margin-inline-start: 0;
                color: ${token.colorTextQuaternary};
            }

            .ant-table-column-sorter-up.active,
            .ant-table-column-sorter-down.active {
                color: ${token.colorPrimary};
            }

            .ant-table-tbody > tr.ant-table-row > td {
                padding: ${compact ? 8 : 10}px !important;
                border-bottom: 1px solid ${token.colorSplit};
                background: ${token.colorBgContainer};
                color: var(--page-table-text);
                font-size: ${token.fontSizeSM + 1}px;
                transition: background-color .16s ease;
            }

            .ant-table-tbody > tr:last-child > td {
                border-bottom: 0;
            }

            .ant-table-tbody > tr > .ant-table-cell-fix-end,
            .ant-table-tbody > tr > .ant-table-cell-fix-right,
            .ant-table-tbody > tr > td.action-cell {
                background: ${token.colorBgContainer} !important;
            }

            .ant-table-tbody > tr.ant-table-row:hover > td,
            .ant-table-tbody > tr.ant-table-row.action-menu-open > td,
            .ant-table-tbody > tr.ant-table-row:hover > .ant-table-cell-fix-end,
            .ant-table-tbody > tr.ant-table-row:hover > .ant-table-cell-fix-right,
            .ant-table-tbody > tr.ant-table-row:hover > td.action-cell {
                background: ${highlight} !important;
            }

            .ant-table-tbody > tr.ant-table-row-selected > td,
            .ant-table-tbody > tr.ant-table-row-selected > .ant-table-cell-fix-end,
            .ant-table-tbody > tr.ant-table-row-selected > .ant-table-cell-fix-right,
            .ant-table-tbody > tr.ant-table-row-selected > td.action-cell {
                background: ${token.colorPrimaryBg} !important;
            }

            .ant-table-tbody > tr.ant-table-row-selected:hover > td,
            .ant-table-tbody > tr.ant-table-row-selected:hover > .ant-table-cell-fix-end,
            .ant-table-tbody > tr.ant-table-row-selected:hover > .ant-table-cell-fix-right,
            .ant-table-tbody > tr.ant-table-row-selected:hover > td.action-cell {
                background: ${token.colorPrimaryBgHover} !important;
            }

            .ant-table-placeholder > td {
                height: 112px;
            }

            .ant-table-content {
                overscroll-behavior-x: contain;
                scrollbar-width: thin;
                -webkit-overflow-scrolling: touch;
            }

            .ant-table-selection-column,
            .action-cell {
                text-align: center !important;
            }

            .action-cell .ant-dropdown-trigger {
                display: inline-flex;
            }

            .action-cell .ant-btn {
                height: 25px;
                padding-inline: 10px;
                border-radius: ${token.borderRadius}px;
                font-size: ${token.fontSizeSM}px;
            }
        `,
        paginationCard: css`
            container-name: data-table-pagination;
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
            width: 100%;
            margin: 0 !important;
            display: flex;
            align-items: center;
            justify-content: center;
            background: transparent;

            .ant-pagination-total-text,
            .ant-pagination-options-quick-jumper {
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
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

            .ant-pagination-item:hover,
            .ant-pagination-prev:hover .ant-pagination-item-link,
            .ant-pagination-next:hover .ant-pagination-item-link {
                background: ${token.colorFillTertiary};
            }

            /* 刷新时会临时禁用分页，当前页保持同一套颜色，避免闪色。 */
            && .ant-pagination-item-active,
            &&.ant-pagination-disabled .ant-pagination-item-active {
                &, &:hover, &:active {
                    background: ${token.colorBgContainer};
                    box-shadow: inset 0 0 0 1px ${token.colorPrimaryBorder};
                }

                a, &:hover a, &:active a {
                    color: ${token.colorPrimary};
                }
            }

            .ant-pagination-options-quick-jumper input {
                width: ${compact ? 30 : 34}px;
                height: ${compact ? 22 : 26}px;
                margin-inline: 4px;
                border-radius: ${token.borderRadiusSM}px;
                font-size: ${compact ? 11 : 12}px;
            }

            @container data-table-pagination (max-width: 767px) {
                .ant-pagination-total-text,
                .ant-pagination-options-quick-jumper {
                    display: none;
                }

                .ant-pagination-options {
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

            && .ant-select-suffix {
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
