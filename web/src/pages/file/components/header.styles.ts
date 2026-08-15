import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean; dark?: boolean}>(({
    css,
    token,
    isDarkMode,
}, props = {}) => {
    const compact = Boolean(props?.compact);
    const dark = typeof props?.dark === "boolean" ? props.dark : Boolean(isDarkMode);

    return {
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
                max-width: calc(100% - 190px);
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
                letter-spacing: 0;
            }

            .hero-visual {
                position: absolute;
                z-index: 1;
                top: 0;
                right: 88px;
                width: 50%;
                height: 100%;
                opacity: ${dark ? .12 : .23};
                filter: ${dark ? "blur(.8px) saturate(.78)" : "blur(.25px)"};
                -webkit-mask-image: linear-gradient(90deg, transparent 0%, rgba(0, 0, 0, .38) 28%, #000 52%);
                mask-image: linear-gradient(90deg, transparent 0%, rgba(0, 0, 0, .38) 28%, #000 52%);
                pointer-events: none;
            }

            .create-button {
                position: relative;
                z-index: 3;
                min-width: 96px;
                height: ${compact ? 32 : 36}px;
                margin-right: ${compact ? 18 : 24}px;
                padding: 0 ${compact ? 13 : 16}px;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                gap: 6px;
                flex: none;
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
            }

            .create-button:focus-visible {
                box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
            }

            .create-button .switch-arrow {
                font-size: 10px;
                opacity: .72;
                transform: rotate(90deg);
            }

            @container file-workspace (max-width: 600px) {
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
                    right: 28px;
                    width: 68%;
                    opacity: ${dark ? .07 : .14};
                }

                .hero-visual .flow-detail {
                    display: none;
                }
            }

            @container file-workspace (max-width: 390px) {
                .create-button {
                    min-width: 84px;
                    margin-right: 12px;
                    padding-inline: 12px;
                }
            }

            @container file-workspace (max-width: 350px) {
                .hero-visual {
                    display: none;
                }
            }
        `,
        commandBar: css`
            min-width: 0;
            min-height: ${compact ? 48 : 58}px;
            padding: ${compact ? "10px 10px 8px" : "14px 12px 9px"};
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: ${compact ? 12 : 16}px;
            background: ${token.colorBgContainer};

            .command-actions {
                min-width: 0;
                margin-left: auto;
                display: flex;
                flex: 0 1 auto;
                align-items: center;
                justify-content: flex-start;
                gap: 6px;
                overflow-x: auto;
                scrollbar-width: none;
            }

            .command-actions::-webkit-scrollbar {
                display: none;
            }

            @container file-workspace (max-width: 560px) {
                padding: ${compact ? 8 : 10}px;
                flex-wrap: wrap;
                gap: 6px;

                .file-search {
                    width: 100%;
                    max-width: none;
                    min-width: 0;
                    flex: 0 0 100%;
                }

                .command-actions {
                    width: 100%;
                    min-width: 0;
                    margin-left: 0;
                    display: flex;
                    flex: 0 0 100%;
                    justify-content: flex-start;
                    gap: 4px;
                }

                .command-actions > * {
                    width: auto;
                }
            }
        `,
        pathSection: css`
            min-width: 0;
            padding: 0 ${compact ? 10 : 12}px ${compact ? 10 : 12}px;
            background: ${token.colorBgContainer};

            @container file-workspace (max-width: 560px) {
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
                scroll-padding-inline: 6px;
                scrollbar-width: none;
                white-space: nowrap;
                -webkit-overflow-scrolling: touch;
            }

            .path-breadcrumb::-webkit-scrollbar {
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
                line-height: 1;
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

            .path-segment:hover {
                color: ${token.colorPrimary};
            }

            .path-segment:focus-visible {
                color: ${token.colorPrimary};
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            .path-current {
                color: ${token.colorText};
                cursor: text;
                font-weight: 500;
            }

            .path-current:hover {
                color: ${token.colorPrimary};
            }

            .path-current:focus-visible {
                color: ${token.colorPrimary};
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
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
                flex: none;
                display: flex;
                align-items: center;
                gap: 6px;
                overflow-x: auto;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
            }

            .path-actions::-webkit-scrollbar {
                display: none;
            }

            .path-actions > .path-batch,
            .path-actions > .path-paste.ant-btn {
                height: ${compact ? 28 : 30}px;
                flex: none;
            }

            .path-actions > .path-batch {
                width: auto;
                min-width: max-content;
            }

            .path-actions > .path-paste.ant-btn {
                width: auto;
                min-width: max-content;
            }

            .path-separator {
                margin-inline: 1px;
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 8 : 9}px;
            }

            @container file-workspace (max-width: 560px) {
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
        searchBox: css`
            width: ${compact ? 300 : 320}px;
            max-width: 100%;
            min-width: 0;
            flex: 0 1 ${compact ? 300 : 320}px;

            &.ant-input-affix-wrapper {
                height: ${compact ? 30 : 34}px;
                padding: 0 ${compact ? 9 : 12}px;
                border: 1px solid transparent;
                border-radius: ${token.borderRadiusSM}px;
                background: color-mix(in srgb, ${token.colorFillTertiary} 45%, ${token.colorFillQuaternary});
                box-shadow: none;
                transition: border-color .16s ease, background-color .16s ease, box-shadow .16s ease;
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
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
            }

            .ant-input::placeholder {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 11 : 12}px;
            }
        `,
        toolbarTrigger: css`
            width: auto;
            min-width: max-content;
            height: ${compact ? 30 : 34}px;
            padding: 0 ${compact ? 7 : 10}px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 5px;
            border: 1px solid transparent;
            border-radius: ${token.borderRadiusSM}px;
            outline: none;
            appearance: none;
            background: ${token.colorFillQuaternary};
            color: ${token.colorTextSecondary};
            cursor: pointer;
            font: inherit;
            line-height: 1;
            transition: background-color .16s ease;

            .marker {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 11 : 12}px;
            }

            .value {
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
                white-space: nowrap;
            }

            .arrow {
                flex: none;
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 8 : 9}px;
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
                background: ${token.colorFillQuaternary};
                color: ${token.colorTextDisabled};
                cursor: not-allowed;
            }

            &:disabled .marker,
            &:disabled .value,
            &:disabled .arrow {
                color: ${token.colorTextDisabled};
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
                box-shadow: none;
                font-size: ${compact ? 11 : 12}px;
            }

            && .value {
                color: ${token.colorTextSecondary};
                white-space: nowrap;
            }

            &&:hover,
            &&.ant-dropdown-open {
                background: ${token.colorFillSecondary};
                color: ${token.colorTextSecondary};
            }

            &&:focus-visible {
                border-color: ${token.colorPrimaryBorder};
                background: ${token.colorBgContainer};
                color: ${token.colorTextSecondary};
                box-shadow: 0 0 0 3px ${token.colorPrimaryBg};
            }

            &&.ant-btn-loading {
                color: ${token.colorTextTertiary};
            }

            &&:disabled,
            &&:disabled:hover {
                background: ${token.colorFillQuaternary};
                color: ${token.colorTextDisabled};
            }

            &&:disabled .value {
                color: ${token.colorTextDisabled};
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

            && .ant-dropdown-menu-item,
            && .ant-dropdown-menu-submenu-title {
                min-height: ${compact ? 26 : 30}px !important;
                margin: 0 !important;
                padding: 0 7px !important;
                border-radius: ${token.borderRadiusSM}px !important;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px !important;
            }

            && .ant-dropdown-menu-item:hover,
            && .ant-dropdown-menu-submenu-title:hover {
                background: ${token.colorFillQuaternary};
            }

            && .ant-dropdown-menu-title-content {
                min-width: 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            && .ant-dropdown-menu-item-danger:not(.ant-dropdown-menu-item-disabled) {
                color: ${token.colorError};
            }

            && .ant-dropdown-menu-item-danger:not(.ant-dropdown-menu-item-disabled):hover {
                background: ${token.colorErrorBg};
                color: ${token.colorError};
            }
        `,
    };
});

export default useStyles;
