import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token, isDarkMode}: any, props: any = {}) => {
    const compact = Boolean(props?.isCompactMode);
    const dark = typeof props?.isDarkMode === "boolean" ? props.isDarkMode : Boolean(isDarkMode);
    const tableHighlight = dark
        ? `color-mix(in srgb, #fff 5%, ${token.colorBgContainer})`
        : `color-mix(in srgb, #000 3%, ${token.colorBgContainer})`;

    return {
    page: css`
        width: 100%;
        max-width: 1450px;
        margin: 0 auto;

        > .ant-col {
            min-width: 0;
        }
    `,
    workspace: css`
        container-name: task-workspace;
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

        &::before {
            position: absolute;
            z-index: 0;
            inset: 0;
            background-image:
                linear-gradient(${dark ? "rgba(255,255,255,0.018)" : "rgba(72,132,202,0.035)"} 1px, transparent 1px),
                linear-gradient(90deg, ${dark ? "rgba(255,255,255,0.018)" : "rgba(72,132,202,0.035)"} 1px, transparent 1px);
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
            letter-spacing: 0;
        }

        .eyebrow .status-dot {
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
            letter-spacing: 0;
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
            width: auto;
            min-width: 88px;
            height: ${compact ? 32 : 36}px;
            margin: 0 ${compact ? 18 : 24}px 0 0;
            padding: 0 ${compact ? 14 : 17}px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            flex: none;
            box-sizing: border-box;
            border: 0;
            border-radius: ${token.borderRadius}px;
            outline: none;
            appearance: none;
            background: ${token.colorPrimary};
            color: #fff;
            cursor: pointer;
            font: inherit;
            font-size: ${compact ? 12 : 13}px;
            font-weight: 500;
            line-height: 1;
            white-space: nowrap;
            transition: background-color .18s ease;
        }

        .create-button:hover,
        .create-button:focus-visible {
            background: ${token.colorPrimaryHover};
            color: #fff;
        }

        .create-button:focus-visible {
            box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
        }

        .create-button:active {
            background: ${token.colorPrimaryActive};
        }

        @container task-workspace (max-width: 600px) {
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
                margin-right: 16px;
            }

            .hero-visual {
                right: 42px;
                width: 68%;
                opacity: ${dark ? .07 : .14};
            }

            .hero-visual .flow-secondary,
            .hero-visual .flow-detail {
                display: none;
            }
        }

        @container task-workspace (max-width: 430px) {
            .create-button {
                margin-right: 12px;
                padding: 0 14px;
            }

            .hero-visual {
                right: 16px;
                width: 76%;
                opacity: ${dark ? .045 : .09};
            }

            .hero-visual .flow-node-a {
                display: none;
            }
        }

        @container task-workspace (max-width: 360px) {
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

        .command-actions {
            min-width: max-content;
            margin-left: auto;
            display: flex;
            flex: none;
            align-items: center;
            justify-content: flex-end;
            gap: 6px;
        }

        @container task-workspace (max-width: 680px) {
            min-height: auto;
            padding: ${compact ? 8 : 11}px;
            display: grid;
            grid-template-columns: minmax(0, 1fr);
            gap: 6px;

            > .ant-input-affix-wrapper {
                width: 100%;
                max-width: none;
                min-width: 0;
                flex: none;
            }

            .command-actions {
                width: 100%;
                min-width: 0;
                margin-left: 0;
                display: flex;
                overflow-x: auto;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
                gap: 4px;
            }

            .command-actions::-webkit-scrollbar {
                display: none;
            }

            .command-actions .category-trigger,
            .command-actions .exec-type-trigger,
            .command-actions .status-trigger {
                width: auto;
                min-width: max-content;
                padding: 0 5px;
                gap: 3px;
            }

            .command-actions .value {
                display: block;
            }
        }

        @container task-workspace (max-width: 430px) {
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
            border: 1px solid transparent;
            border-radius: ${token.borderRadiusSM}px;
            background: color-mix(in srgb, ${token.colorFillTertiary} 45%, ${token.colorFillQuaternary});
            box-shadow: none;
            transition: border-color 0.16s ease, background-color 0.16s ease, box-shadow 0.16s ease;
        }

        &.ant-input-affix-wrapper:hover {
            background: color-mix(in srgb, ${token.colorFillTertiary} 65%, ${token.colorFillQuaternary});
        }

        &.ant-input-affix-wrapper-focused {
            border-color: ${token.colorPrimaryBorder};
            background: ${token.colorBgContainer};
            box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
        }

        .ant-input-prefix {
            margin-right: 8px;
            color: ${token.colorTextTertiary};
            font-size: 14px;
        }

        .ant-input {
            background: transparent;
            color: color-mix(in srgb, ${token.colorTextSecondary}, ${token.colorBgContainer} 24%);
            font-size: ${compact ? 11 : 12}px;
        }

        .ant-input::placeholder {
            color: ${token.colorTextQuaternary};
            font-size: ${compact ? 11 : 12}px;
        }

        @container task-workspace (max-width: 820px) {
            flex-basis: ${compact ? 260 : 280}px;
            width: ${compact ? 260 : 280}px;
        }

        @container task-workspace (max-width: 680px) {
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
        height: ${compact ? 30 : 34}px;
        padding: 0 ${compact ? 7 : 10}px;
        border: 1px solid transparent;
        border-radius: ${token.borderRadiusSM}px;
        outline: none;
        appearance: none;
        background: ${token.colorFillQuaternary};
        color: ${token.colorTextSecondary};
        cursor: pointer;
        font: inherit;
        line-height: 1;
        transition: background-color 0.16s ease, color 0.16s ease;

        &.category-trigger,
        &.exec-type-trigger,
        &.status-trigger {
            width: auto;
            min-width: max-content;
        }

        .marker {
            color: ${token.colorTextQuaternary};
            font-size: ${compact ? 11 : 12}px;
        }

        .value {
            flex: none;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
            font-weight: 400;
            white-space: nowrap;
            transition: color 0.16s ease;
        }

        .arrow {
            color: ${token.colorTextQuaternary};
            font-size: ${compact ? 9 : 10}px;
            transition: color 0.16s ease, transform 0.16s ease;
        }

        &:hover,
        &.ant-dropdown-open {
            background: ${token.colorFillSecondary};
        }

        &:hover .value,
        &.ant-dropdown-open .value {
            color: ${token.colorTextSecondary};
        }

        &.ant-dropdown-open .arrow {
            color: ${token.colorTextTertiary};
            transform: rotate(180deg);
        }

        &:focus-visible {
            border-color: ${token.colorPrimaryBorder};
            background: ${token.colorBgContainer};
            box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
        }
    `,
    toolbarDropdown: css`
        && {
            max-width: calc(100vw - 24px);
            box-sizing: border-box;
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

        && .ant-dropdown-menu .ant-dropdown-menu-item {
            min-height: ${compact ? 24 : 30}px !important;
            margin: 0 !important;
            padding: 0 6px !important;
            border-radius: ${token.borderRadiusSM}px !important;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px !important;
            font-weight: 400;
            line-height: ${compact ? 24 : 30}px !important;
        }

        && .ant-dropdown-menu .ant-dropdown-menu-item .ant-dropdown-menu-title-content {
            min-width: 0;
            overflow: hidden;
            font-size: ${compact ? 11 : 12}px !important;
            line-height: ${compact ? 24 : 30}px !important;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        && .ant-dropdown-menu .ant-dropdown-menu-item:hover {
            background: ${token.colorFillQuaternary};
            color: ${token.colorTextSecondary};
        }

        && .ant-dropdown-menu .ant-dropdown-menu-item-selected,
        && .ant-dropdown-menu .ant-dropdown-menu-item-selected:hover {
            background: color-mix(in srgb, ${token.colorPrimary}, ${token.colorBgElevated} 94%);
            color: ${token.colorTextSecondary};
        }
    `,
    dataPanel: css`
        position: relative;
        min-width: 0;
        padding: 0 ${compact ? 10 : 12}px ${compact ? 10 : 12}px;

        @container task-workspace (max-width: 430px) {
            padding: 0 10px 10px;
        }
    `,
    taskTable: css`
        --task-table-text: ${dark ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.65)"};
        min-width: 0;
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
            height: auto;
            padding: ${compact ? 8 : 10}px !important;
            border-bottom: 1px solid ${token.colorBorderSecondary};
            background: ${tableHighlight};
            color: var(--task-table-text);
            font-size: ${token.fontSize}px;
        }

        .ant-table-thead > tr > th.ant-table-column-sort,
        .ant-table-thead > tr > th.ant-table-column-has-sorters:hover {
            background: ${tableHighlight};
        }

        .ant-table-tbody > tr > td.ant-table-column-sort {
            background: ${token.colorBgContainer};
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

        .ant-table-thead > tr > th.action-cell {
            background: ${tableHighlight} !important;
            color: var(--task-table-text) !important;
        }

        .ant-table-thead > tr > .ant-table-cell-fix-end,
        .ant-table-thead > tr > .ant-table-cell-fix-right,
        .ant-table-thead > tr > th[class*="ant-table-cell-fix-end"],
        .ant-table-thead > tr > th[class*="ant-table-cell-fix-right"] {
            background: ${tableHighlight} !important;
            color: var(--task-table-text) !important;
        }

        .ant-table-tbody > tr > .ant-table-cell-fix-end,
        .ant-table-tbody > tr > .ant-table-cell-fix-right,
        .ant-table-tbody > tr > td[class*="ant-table-cell-fix-end"],
        .ant-table-tbody > tr > td[class*="ant-table-cell-fix-right"] {
            background: ${token.colorBgContainer} !important;
        }

        .ant-table-tbody > tr > td.action-cell {
            background: ${token.colorBgContainer} !important;
        }

        .ant-table-tbody > tr.ant-table-row > td {
            height: auto;
            padding: ${compact ? 8 : 10}px !important;
            border-bottom: 1px solid ${token.colorSplit};
            background: ${token.colorBgContainer};
            color: var(--task-table-text);
            font-size: ${token.fontSizeSM + 1}px;
            transition: background-color .16s ease;
        }

        .ant-table-tbody > tr:last-child > td {
            border-bottom: 0;
        }

        .ant-table-placeholder > td {
            height: 112px;
        }

        .ant-table-tbody > tr.ant-table-row:hover > td {
            background: ${tableHighlight};
        }

        .ant-table-tbody > tr.ant-table-row:hover > .ant-table-cell-fix-end,
        .ant-table-tbody > tr.ant-table-row:hover > .ant-table-cell-fix-right,
        .ant-table-tbody > tr.ant-table-row:hover > td[class*="ant-table-cell-fix-end"],
        .ant-table-tbody > tr.ant-table-row:hover > td[class*="ant-table-cell-fix-right"] {
            background: ${tableHighlight} !important;
        }

        .ant-table-tbody > tr.ant-table-row:hover > td.action-cell {
            background: ${tableHighlight} !important;
        }

        .ant-table-tbody > tr.ant-table-row.action-menu-open > td {
            background: ${tableHighlight} !important;
        }

        .ant-table-content {
            overscroll-behavior-x: contain;
            scrollbar-width: thin;
        }

        .task-name {
            min-width: 0;
        }

        .task-name > button {
            display: block;
            max-width: 100%;
            padding: 0;
            overflow: hidden;
            border: 0;
            outline: none;
            background: transparent;
            color: inherit;
            cursor: pointer;
            font: inherit;
            font-weight: 400;
            line-height: 20px;
            text-align: left;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .task-name > button:hover,
        .task-name > button:focus-visible {
            color: ${token.colorPrimary};
        }

        .task-state {
            flex: none;
            display: inline-flex;
            align-items: center;
            gap: 4px;
            color: inherit;
            line-height: 20px;
        }

        .task-state i {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: currentColor;
        }

        .task-state.running {
            color: ${token.colorSuccess};
        }

        .task-state.paused {
            color: ${token.colorWarning};
        }

        .task-category,
        .task-exec-type {
            display: block;
            min-width: 0;
            overflow: hidden;
            color: inherit;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .schedule-value {
            display: block;
            min-width: 0;
            overflow: hidden;
            color: inherit;
            font-weight: 400;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .runtime-value {
            display: block;
            min-width: 0;
            overflow: hidden;
            color: inherit;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .action-cell {
            text-align: center !important;
        }

        .action-cell .ant-dropdown-trigger {
            display: inline-flex;
        }

        .action-cell .ant-btn {
            height: 25px;
            padding: 10px;
            border-radius: ${token.borderRadius}px;
            font-size: ${token.fontSizeSM}px;
        }

    `,
    state: css`
        min-height: ${compact ? 200 : 240}px;
        padding: ${compact ? 18 : 24}px;
        display: flex;
        align-items: center;
        justify-content: center;
    `,
    paginationCard: css`
        container-name: task-pagination;
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
            border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
        }

        @container task-pagination (max-width: 767px) {
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
