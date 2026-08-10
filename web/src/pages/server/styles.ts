import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token, isDarkMode}: any, props: any = {}) => {
    const compact = Boolean(props?.isCompactMode);
    const dark = typeof props?.isDarkMode === "boolean" ? props.isDarkMode : Boolean(isDarkMode);
    const terminalBackground = "#050505";
    const serverWidth = compact ? 238 : 252;
    const scriptWidth = compact ? 270 : 286;
    const collapsedWidth = compact ? 42 : 46;
    const panelHeaderHeight = compact ? 46 : 50;
    const mobileHeaderHeight = compact ? 42 : 46;
    const serverToolbarHeight = compact ? 46 : 50;

    return {
        page: css`
            --terminal-page-height: calc(100dvh - 92px);
            width: 100%;
            max-width: 1440px;
            min-height: var(--terminal-page-height);
            margin: 0 auto;

            @supports not (height: 100dvh) {
                --terminal-page-height: calc(100vh - 92px);
            }

            @media (max-width: 780px) {
                min-height: 0;
            }
        `,
        workspace: css`
            container-name: terminal-workspace;
            container-type: inline-size;
            width: 100%;
            height: var(--terminal-page-height);
            min-width: 0;
            min-height: 320px;
            overflow: hidden;
            border-radius: ${token.borderRadiusLG}px;
            background: ${token.colorBgContainer};

            > .ant-card-body {
                width: 100%;
                height: 100%;
                min-height: 0;
                padding: 0;
            }

            @media (max-width: 780px) {
                height: auto;
                min-height: var(--terminal-page-height);
            }

            @media (max-height: 600px) and (orientation: landscape) {
                height: auto;
                min-height: 300px;
            }
        `,
        terminalLayout: css`
            width: 100%;
            height: 100%;
            min-width: 0;
            min-height: 0;
            display: grid;
            grid-template-columns: ${serverWidth}px minmax(420px, 1fr) ${scriptWidth}px;
            grid-template-rows: minmax(0, 1fr);

            &.server-collapsed {
                grid-template-columns: ${collapsedWidth}px minmax(420px, 1fr) ${scriptWidth}px;
            }

            &.scripts-collapsed {
                grid-template-columns: ${serverWidth}px minmax(420px, 1fr) ${collapsedWidth}px;
            }

            &.server-collapsed.scripts-collapsed {
                grid-template-columns: ${collapsedWidth}px minmax(420px, 1fr) ${collapsedWidth}px;
            }

            @container terminal-workspace (max-width: 980px) {
                grid-template-columns: ${compact ? 196 : 208}px minmax(300px, 1fr) ${compact ? 220 : 232}px;
                grid-template-rows: minmax(0, 1fr);

                &.scripts-collapsed {
                    grid-template-columns: ${compact ? 196 : 208}px minmax(300px, 1fr) ${collapsedWidth}px;
                }

                &.server-collapsed {
                    grid-template-columns: ${collapsedWidth}px minmax(300px, 1fr) ${compact ? 220 : 232}px;
                }

                &.server-collapsed.scripts-collapsed {
                    grid-template-columns: ${collapsedWidth}px minmax(300px, 1fr) ${collapsedWidth}px;
                }
            }

            @container terminal-workspace (max-width: 760px) {
                height: auto;
                grid-template-columns: 1fr;
                grid-template-rows: auto minmax(430px, auto) auto;

                &.server-collapsed,
                &.scripts-collapsed {
                    grid-template-columns: 1fr;
                }

                &.server-collapsed.scripts-collapsed {
                    grid-template-columns: 1fr;
                }
            }

            @media (min-width: 781px) and (max-height: 600px) and (orientation: landscape) {
                height: auto;
                min-height: 300px;
                grid-template-rows: minmax(300px, 1fr);
            }
        `,
        serverPane: css`
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            border-right: 1px solid ${token.colorSplit};
            background: ${token.colorBgContainer};

            &.collapsed {
                overflow: hidden;
            }

            @container terminal-workspace (max-width: 760px) {
                grid-column: 1;
                grid-row: 1;
                border-right: 0;
                border-bottom: 1px solid ${token.colorSplit};

                &.collapsed {
                    overflow: visible;
                }
            }
        `,
        serverListArea: css`
            position: relative;
            min-width: 0;
            min-height: 0;
            display: flex;
            flex: 1;
            flex-direction: column;

            @container terminal-workspace (max-width: 760px) {
                flex: none;
            }
        `,
        serverLoadingOverlay: css`
            position: absolute;
            z-index: 8;
            inset: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            pointer-events: none;

            .ant-spin {
                line-height: 1;
            }
        `,
        serverHeader: css`
            height: ${panelHeaderHeight}px;
            min-height: ${panelHeaderHeight}px;
            padding: 0 ${compact ? 11 : 13}px;
            display: flex;
            flex: 0 0 ${panelHeaderHeight}px;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
            box-sizing: border-box;
            border-bottom: 1px solid ${token.colorSplit};

            .panel-heading {
                min-width: 0;
                display: flex;
                align-items: center;
                color: ${token.colorTextHeading};
                font-size: ${compact ? 13 : 14}px;
                font-weight: 600;
            }

            .panel-heading > div {
                min-width: 0;
            }

            .ant-btn {
                width: ${compact ? 30 : 32}px;
                height: ${compact ? 30 : 32}px;
                padding: 0;
                flex: none;
                color: ${token.colorTextSecondary};
            }

            .collapsed & {
                padding: 0;
                justify-content: center;
            }

            .collapsed & .panel-heading {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                height: ${mobileHeaderHeight}px;
                min-height: ${mobileHeaderHeight}px;
                flex-basis: ${mobileHeaderHeight}px;
                padding-inline: ${compact ? 9 : 11}px;

                .collapse-trigger {
                    display: none;
                }

                .collapsed & {
                    padding-inline: ${compact ? 9 : 11}px;
                    justify-content: space-between;
                }

                .collapsed & .panel-heading {
                    display: flex;
                }
            }
        `,
        serverToolbar: css`
            min-width: 0;
            height: ${serverToolbarHeight}px;
            min-height: ${serverToolbarHeight}px;
            padding: 8px;
            display: grid;
            flex: 0 0 ${serverToolbarHeight}px;
            grid-template-columns: minmax(0, 1fr) ${compact ? 30 : 34}px;
            align-items: center;
            gap: 6px;
            box-sizing: border-box;

            .collapsed & {
                display: none;
            }

            .ant-input-affix-wrapper {
                height: ${compact ? 30 : 34}px;
                padding-inline: ${compact ? 9 : 10}px;
                border-color: transparent;
                border-radius: ${token.borderRadiusSM}px;
                background: ${token.colorFillQuaternary};
                box-shadow: none;
            }

            .ant-input-affix-wrapper:hover {
                border-color: ${token.colorPrimaryBorder};
            }

            .ant-input-affix-wrapper-focused {
                border-color: ${token.colorPrimary};
                background: ${token.colorBgContainer};
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            .ant-input-prefix {
                color: ${token.colorTextTertiary};
                font-size: 12px;
            }

            .ant-input {
                background: transparent;
                font-size: ${compact ? 11 : 12}px;
            }

            .ant-input::placeholder {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 10 : 11}px;
            }

            .ant-btn {
                width: ${compact ? 30 : 34}px;
                height: ${compact ? 30 : 34}px;
                padding: 0;
                color: ${token.colorTextSecondary};
            }

            @container terminal-workspace (max-width: 760px) {
                min-height: ${compact ? 46 : 50}px;
                padding: 8px;
                grid-template-columns: minmax(0, 1fr) ${compact ? 32 : 34}px;

                .ant-btn {
                    width: ${compact ? 32 : 34}px;
                    height: ${compact ? 32 : 34}px;
                }

                .collapsed & {
                    display: grid;
                }
            }
        `,
        serverList: css`
            min-height: 0;
            padding: 0 ${compact ? 7 : 9}px;
            display: flex;
            flex: 1;
            flex-direction: column;
            gap: 3px;
            overflow-x: hidden;
            overflow-y: auto;
            scrollbar-width: thin;
            scrollbar-color: ${token.colorBorderSecondary} transparent;

            .collapsed & {
                padding: ${compact ? "6px 5px 4px" : "7px 6px 5px"};
                align-items: center;
                gap: 4px;
                overflow-x: hidden;
                overflow-y: auto;
                scrollbar-width: none;
            }

            &::-webkit-scrollbar {
                width: 6px !important;
            }

            &::-webkit-scrollbar-thumb {
                border-radius: ${token.borderRadiusSM}px;
                background: ${token.colorBorderSecondary};
            }

            .collapsed &::-webkit-scrollbar {
                display: none;
            }

            .server-loading,
            .server-empty {
                min-height: 110px;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .server-load-more {
                min-height: 32px;
                display: flex;
                flex: 0 0 auto;
                align-items: center;
                justify-content: center;
            }

            .server-empty .ant-empty {
                margin-block: 12px;
            }

            .server-empty .ant-empty-description {
                color: ${token.colorTextTertiary};
                font-size: 11px;
            }

            @container terminal-workspace (max-width: 760px) {
                min-width: 0;
                padding: 0 ${compact ? 9 : 11}px ${compact ? 9 : 11}px 0;
                flex: none;
                flex-direction: row;
                gap: 6px;
                overflow-x: auto;
                overflow-y: hidden;
                scroll-snap-type: x proximity;
                scroll-padding-left: ${compact ? 12 : 14}px;
                overscroll-behavior-x: contain;
                scrollbar-width: none;
                touch-action: pan-x pan-y;
                -webkit-overflow-scrolling: touch;

                .collapsed & {
                    display: flex;
                    padding: 0 ${compact ? 9 : 11}px ${compact ? 9 : 11}px 0;
                    align-items: stretch;
                    gap: 6px;
                    overflow-x: auto;
                    overflow-y: hidden;
                }

                &::-webkit-scrollbar {
                    display: none;
                }

                .server-empty,
                .server-loading,
                .server-load-more {
                    min-width: 144px;
                    min-height: ${compact ? 48 : 52}px;
                    flex: 0 0 144px;
                }

                .server-empty .ant-empty-image {
                    display: none;
                }
            }
        `,
        serverItem: css`
            position: relative;
            min-width: 0;
            min-height: ${compact ? 50 : 54}px;
            overflow: hidden;
            border: 0;
            border-radius: ${token.borderRadius}px;
            background: transparent;
            color: ${token.colorText};
            content-visibility: auto;
            contain-intrinsic-size: auto ${compact ? 50 : 54}px;
            transition: background-color .16s ease;

            &:hover {
                background: ${token.colorFillQuaternary};
            }

            &.selected {
                background: ${token.colorPrimaryBg};
            }

            .connection-dot {
                position: absolute;
                z-index: 2;
                top: 50%;
                left: ${compact ? 10 : 11}px;
                width: 6px;
                height: 6px;
                border-radius: 50%;
                background: ${token.colorTextQuaternary};
                transform: translateY(-50%);
            }

            .connection-dot.connected {
                background: ${token.colorSuccess};
                box-shadow: 0 0 0 3px ${token.colorSuccessBg};
            }

            .connection-dot.connecting {
                background: ${token.colorWarning};
                box-shadow: 0 0 0 3px ${token.colorWarningBg};
            }

            .connection-dot.error {
                background: ${token.colorError};
                box-shadow: 0 0 0 3px ${token.colorErrorBg};
            }

            .server-icon {
                display: none;
            }

            .server-select {
                width: 100%;
                min-height: ${compact ? 50 : 54}px;
                padding: ${compact ? "5px 38px 5px 26px" : "6px 41px 6px 27px"};
                display: flex;
                align-items: center;
                border: 0;
                border-radius: inherit;
                outline: none;
                background: transparent;
                color: inherit;
                cursor: pointer;
                font: inherit;
                text-align: left;
            }

            .server-select:focus-visible {
                box-shadow: inset 0 0 0 1px ${token.colorPrimaryBorder};
            }

            .server-select:disabled {
                cursor: default;
                opacity: .72;
            }

            .server-copy {
                min-width: 0;
                display: block;
                flex: 1;
            }

            .server-name,
            .server-endpoint {
                display: block;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .server-name {
                color: ${token.colorTextHeading};
                font-size: ${compact ? 12 : 13}px;
                font-weight: 500;
                line-height: ${compact ? 18 : 19}px;
            }

            .server-endpoint {
                color: ${token.colorTextTertiary};
                font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
                font-size: 11px;
                line-height: ${compact ? 15 : 16}px;
            }

            .server-actions {
                position: absolute;
                z-index: 3;
                top: 50%;
                right: ${compact ? 5 : 6}px;
                display: flex;
                align-items: center;
                gap: 1px;
                opacity: 0;
                transform: translateY(-50%);
                transition: opacity .16s ease;
            }

            .server-action {
                width: ${compact ? 28 : 30}px;
                height: ${compact ? 28 : 30}px;
                padding: 0;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                border: 0;
                border-radius: ${token.borderRadiusSM}px;
                outline: none;
                background: transparent;
                color: ${token.colorTextTertiary};
                cursor: pointer;
                transition: background-color .16s ease, color .16s ease;
            }

            &:hover .server-actions,
            &.selected .server-actions,
            .server-actions:focus-within {
                opacity: 1;
            }

            .server-action:hover,
            .server-action:focus-visible {
                background: ${token.colorFillSecondary};
                color: ${token.colorText};
            }

            .server-action:disabled {
                cursor: default;
                opacity: .52;
            }

            .collapsed & {
                width: ${compact ? 30 : 32}px;
                min-height: ${compact ? 30 : 32}px;
                flex: 0 0 ${compact ? 30 : 32}px;
                overflow: visible;
                contain-intrinsic-size: auto ${compact ? 30 : 32}px;
            }

            .collapsed & .server-select {
                width: ${compact ? 30 : 32}px;
                min-height: ${compact ? 30 : 32}px;
                padding: 0;
                justify-content: center;
            }

            .collapsed & .server-icon {
                display: inline-flex;
                align-items: center;
                justify-content: center;
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 14 : 15}px;
            }

            .collapsed &.selected .server-icon {
                color: ${token.colorPrimary};
            }

            .collapsed & .connection-dot {
                top: auto;
                right: 3px;
                bottom: 3px;
                left: auto;
                width: 8px;
                height: 8px;
                border: 1px solid ${token.colorBgContainer};
                box-shadow: none;
                transform: none;
            }

            .collapsed & .server-copy,
            .collapsed & .server-actions {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                min-height: ${compact ? 48 : 52}px;
                flex: 0 0 ${compact ? 160 : 168}px;
                overflow: hidden;
                background: ${token.colorFillQuaternary};
                contain-intrinsic-size: auto ${compact ? 48 : 52}px;
                scroll-snap-align: start;

                .server-select {
                    min-height: ${compact ? 48 : 52}px;
                    padding-right: ${compact ? 43 : 46}px;
                }

                .server-actions {
                    opacity: 1;
                }

                .server-action {
                    width: ${compact ? 32 : 34}px;
                    height: ${compact ? 32 : 34}px;
                }

                .collapsed & {
                    width: auto;
                    min-height: ${compact ? 48 : 52}px;
                    flex: 0 0 ${compact ? 160 : 168}px;
                }

                .collapsed & .server-select {
                    width: 100%;
                    min-height: ${compact ? 48 : 52}px;
                    padding: ${compact ? "5px 43px 5px 26px" : "6px 46px 6px 27px"};
                    justify-content: flex-start;
                }

                .collapsed & .server-icon {
                    display: none;
                }

                .collapsed & .connection-dot {
                    top: 50%;
                    right: auto;
                    bottom: auto;
                    left: ${compact ? 10 : 11}px;
                    width: 6px;
                    height: 6px;
                    border: 0;
                    transform: translateY(-50%);
                }

                .collapsed & .connection-dot.connected {
                    box-shadow: 0 0 0 3px ${token.colorSuccessBg};
                }

                .collapsed & .connection-dot.connecting {
                    box-shadow: 0 0 0 3px ${token.colorWarningBg};
                }

                .collapsed & .connection-dot.error {
                    box-shadow: 0 0 0 3px ${token.colorErrorBg};
                }

                .collapsed & .server-copy {
                    display: block;
                }

                .collapsed & .server-actions {
                    display: flex;
                    opacity: 1;
                }
            }

            @container terminal-workspace (max-width: 430px) {
                flex-basis: ${compact ? 152 : 160}px;

                .collapsed & {
                    flex-basis: ${compact ? 152 : 160}px;
                }
            }
        `,
        serverMenu: css`
            && {
                min-width: ${compact ? 94 : 104}px;
                padding: 0;
                border-radius: ${token.borderRadius}px;
            }

            && .ant-dropdown-menu {
                padding: 3px !important;
                border-radius: ${token.borderRadius}px !important;
                background: ${token.colorBgElevated};
                box-shadow: ${token.boxShadowSecondary};
            }

            && .ant-dropdown-menu-item {
                min-height: ${compact ? 28 : 30}px;
                padding: 0 ${compact ? 7 : 8}px !important;
                border-radius: ${token.borderRadiusSM}px !important;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
            }

            && .ant-dropdown-menu-item-danger {
                color: ${token.colorError};
            }
        `,
        serverFooter: css`
            flex: none;
            padding: ${compact ? "8px 9px 10px" : "9px 10px 11px"};

            .collapsed & {
                padding: ${compact ? "5px 5px 9px" : "6px 6px 10px"};
                display: block;
            }

            .collapsed & > button {
                min-height: ${compact ? 30 : 32}px;
                padding: 0;
                border: 0;
            }

            .collapsed & > button span {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                display: none;
                .collapsed & {
                    display: none;
                }
            }
        `,
        serverAdd: css`
            width: 100%;
            min-height: ${compact ? 38 : 42}px;
            padding: 0 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 7px;
            border: 1px dashed ${token.colorBorder};
            border-radius: ${token.borderRadius}px;
            outline: none;
            appearance: none;
            background: transparent;
            color: ${token.colorTextTertiary};
            cursor: pointer;
            font: inherit;
            font-size: ${compact ? 12 : 13}px;
            transition: border-color .16s ease, background-color .16s ease, color .16s ease;

            &.mobile {
                display: none;
            }

            &:hover,
            &:focus-visible {
                border-color: ${token.colorPrimaryBorder};
                background: ${token.colorFillQuaternary};
                color: ${token.colorPrimary};
            }

            &:focus-visible {
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            @container terminal-workspace (max-width: 760px) {
                min-height: ${compact ? 50 : 54}px;

                &.mobile {
                    width: auto;
                    min-width: ${compact ? 122 : 130}px;
                    margin-left: ${compact ? 12 : 14}px;
                    padding: 0 11px;
                    display: flex;
                    flex: 0 0 auto;
                    white-space: nowrap;
                    scroll-snap-align: start;
                }
            }
        `,
        consolePane: css`
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            background: ${token.colorBgContainer};

            &.session-hidden {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                grid-column: 1;
                grid-row: 2;
                min-height: max(430px, calc(100dvh - 230px));
            }

            @media (max-height: 600px) and (orientation: landscape) {
                min-height: 300px;
            }
        `,
        consoleHeader: css`
            min-width: 0;
            height: ${panelHeaderHeight}px;
            min-height: ${panelHeaderHeight}px;
            padding: 0 ${compact ? 11 : 13}px;
            display: flex;
            flex: 0 0 ${panelHeaderHeight}px;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            box-sizing: border-box;
            border-bottom: 1px solid ${token.colorSplit};

            .console-title {
                min-width: 0;
                display: flex;
                align-items: center;
                gap: 9px;
            }

            .console-title > .status-dot {
                width: 7px;
                height: 7px;
                flex: none;
                border-radius: 50%;
                background: ${token.colorTextQuaternary};
            }

            .console-title > .status-dot.connected {
                background: ${token.colorSuccess};
                box-shadow: 0 0 0 3px ${token.colorSuccessBg};
            }

            .console-title > .status-dot.connecting {
                background: ${token.colorWarning};
                box-shadow: 0 0 0 3px ${token.colorWarningBg};
            }

            .console-title > .status-dot.error {
                background: ${token.colorError};
                box-shadow: 0 0 0 3px ${token.colorErrorBg};
            }

            .console-title > div {
                min-width: 0;
            }

            .console-name,
            .console-endpoint {
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .console-name {
                color: ${token.colorTextHeading};
                font-size: ${compact ? 13 : 14}px;
                font-weight: 600;
                line-height: ${compact ? 19 : 20}px;
            }

            .console-endpoint {
                color: ${token.colorTextTertiary};
                font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
                font-size: ${compact ? 11 : 12}px;
                line-height: ${compact ? 15 : 16}px;
            }

            .console-actions {
                display: flex;
                flex: none;
                align-items: center;
                gap: 2px;
            }

            .console-actions .status-text {
                margin-right: 5px;
                color: ${token.colorTextTertiary};
                font-size: 11px;
                white-space: nowrap;
            }

            .console-actions .status-text.connected {
                color: ${token.colorSuccessText};
            }

            .console-actions .status-text.connecting {
                color: ${token.colorWarningText};
            }

            .console-actions .status-text.error {
                color: ${token.colorErrorText};
            }

            .console-actions .ant-btn {
                width: ${compact ? 30 : 32}px;
                height: ${compact ? 30 : 32}px;
                padding: 0;
                color: ${token.colorTextSecondary};
            }

            @container terminal-workspace (max-width: 520px) {
                .console-endpoint,
                .console-actions .status-text {
                    display: none;
                }
            }

            @container terminal-workspace (max-width: 760px) {
                height: ${mobileHeaderHeight}px;
                min-height: ${mobileHeaderHeight}px;
                flex-basis: ${mobileHeaderHeight}px;
            }
        `,
        terminalBody: css`
            min-width: 0;
            min-height: 0;
            padding: 0;
            display: flex;
            flex: 1;
            border-bottom: ${dark ? `1px solid ${token.colorBorderSecondary}` : "0"};
            background: ${terminalBackground};
        `,
        terminalScreen: css`
            position: relative;
            width: 100%;
            min-width: 0;
            min-height: 0;
            overflow: hidden;
            display: flex;
            flex: 1;
            border: 0;
            border-radius: 0;
            background: ${terminalBackground};

            &:fullscreen {
                width: 100vw;
                height: 100dvh;
                min-width: 0;
                min-height: 0;
                border: 0;
                border-radius: 0;
                background: ${terminalBackground};
            }

            &:fullscreen::backdrop {
                background: ${terminalBackground};
            }

            .fullscreen-actions {
                position: absolute;
                z-index: 3;
                top: 10px;
                right: 12px;
                padding: 3px;
                display: flex;
                align-items: center;
                gap: 2px;
                border: 1px solid rgba(255, 255, 255, .08);
                border-radius: ${token.borderRadius}px;
                background: rgba(18, 18, 18, .82);
                backdrop-filter: blur(8px);
            }

            .fullscreen-actions .ant-btn {
                width: 30px;
                height: 30px;
                padding: 0;
                color: rgba(255, 255, 255, .76);
            }

            .fullscreen-actions .ant-btn:hover,
            .fullscreen-actions .ant-btn:focus-visible {
                background: rgba(255, 255, 255, .1);
                color: #fff;
            }

            @supports not (height: 100dvh) {
                &:fullscreen {
                    height: 100vh;
                }
            }
        `,
        terminalOverlay: css`
            position: absolute;
            z-index: 2;
            inset: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: inherit;
            background: color-mix(in srgb, ${terminalBackground}, transparent 10%);
            pointer-events: auto;

            &.connecting {
                color: #dedede;
            }

            .ant-spin-description {
                color: #a8a8a8;
                font-size: 12px;
            }

            .ant-btn-text:disabled {
                border-color: transparent;
                background: transparent;
                color: rgba(255, 255, 255, .52);
            }
        `,
        scriptPane: css`
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            border-left: 1px solid ${token.colorSplit};
            background: color-mix(in srgb, ${token.colorFillQuaternary}, ${token.colorBgContainer} 62%);

            &.collapsed {
                align-items: stretch;
            }

            @container terminal-workspace (max-width: 760px) {
                grid-column: 1;
                grid-row: 3;
                max-height: 280px;
                border-top: 1px solid ${token.colorSplit};
                border-left: 0;

                &.collapsed {
                    max-height: 280px;
                    flex-direction: column;
                    align-items: stretch;
                }
            }

            @media (max-width: 760px) and (max-height: 600px) and (orientation: landscape) {
                max-height: 150px;

                &.collapsed {
                    max-height: 150px;
                }
            }
        `,
        scriptHeader: css`
            height: ${panelHeaderHeight}px;
            min-height: ${panelHeaderHeight}px;
            padding: 0 ${compact ? 8 : 9}px 0 ${compact ? 11 : 13}px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
            flex: 0 0 ${panelHeaderHeight}px;
            box-sizing: border-box;
            border-bottom: 1px solid ${token.colorSplit};

            .panel-heading {
                min-width: 0;
                display: flex;
                align-items: center;
                gap: 7px;
                color: ${token.colorTextHeading};
                font-size: ${compact ? 13 : 14}px;
                font-weight: 600;
                white-space: nowrap;
            }

            .panel-heading > div {
                min-width: 0;
            }

            .ant-btn {
                width: ${compact ? 30 : 32}px;
                height: ${compact ? 30 : 32}px;
                padding: 0;
                flex: none;
                color: ${token.colorTextSecondary};
            }

            .collapsed & {
                padding: 0;
                justify-content: center;
            }

            .collapsed & .panel-heading {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                height: ${mobileHeaderHeight}px;
                min-height: ${mobileHeaderHeight}px;
                flex-basis: ${mobileHeaderHeight}px;

                .collapse-trigger {
                    display: none;
                }

                .collapsed & {
                    width: auto;
                    min-width: 0;
                    padding: 0 ${compact ? 8 : 9}px 0 ${compact ? 11 : 13}px;
                    flex: none;
                    justify-content: space-between;
                    border-right: 0;
                    border-bottom: 1px solid ${token.colorSplit};
                }

                .collapsed & .panel-heading {
                    display: flex;
                }
            }
        `,
        scriptBody: css`
            min-width: 0;
            min-height: 0;
            padding: ${compact ? "6px 7px 9px" : "7px 8px 10px"};
            flex: 1;
            overflow-y: auto;
            scrollbar-width: thin;
            scrollbar-color: ${token.colorBorderSecondary} transparent;

            .collapsed & {
                display: none;
            }

            .ant-collapse {
                border: 0;
                background: transparent;
            }

            .ant-collapse-item {
                border-bottom-color: ${token.colorSplit};
            }

            .ant-collapse-header {
                min-height: ${compact ? 38 : 42}px;
                padding: ${compact ? "8px 7px" : "9px 8px"} !important;
                align-items: center !important;
                border-radius: ${token.borderRadiusSM}px !important;
                color: ${token.colorTextSecondary} !important;
                font-size: ${compact ? 12 : 13}px;
            }

            .ant-collapse-header:hover {
                background: ${token.colorFillQuaternary};
            }

            .ant-collapse-content-box {
                padding: 2px 0 ${compact ? 5 : 7}px !important;
            }

            .script-group-title {
                min-width: 0;
                display: inline-flex;
                align-items: center;
                gap: 7px;
            }

            .script-group-title .anticon {
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 12 : 13}px;
            }

            .script-list {
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .script-item {
                min-width: 0;
                min-height: ${compact ? 50 : 54}px;
                padding: ${compact ? "3px 3px 3px 8px" : "4px 4px 4px 9px"};
                display: grid;
                grid-template-columns: minmax(0, 1fr) ${compact ? 29 : 32}px;
                align-items: center;
                gap: 3px;
                border-radius: ${token.borderRadiusSM}px;
                transition: background-color .16s ease;
            }

            .script-item:hover {
                background: ${token.colorFillQuaternary};
            }

            .script-copy {
                min-width: 0;
                padding: 0;
                display: flex;
                flex-direction: column;
                align-items: flex-start;
                border: 0;
                outline: none;
                background: transparent;
                cursor: pointer;
                font: inherit;
                text-align: left;
            }

            .script-copy:focus-visible {
                border-radius: ${token.borderRadiusSM}px;
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            .script-name {
                color: ${token.colorText};
                font-size: ${compact ? 12 : 13}px;
                line-height: 18px;
            }

            code {
                width: 100%;
                overflow: hidden;
                color: ${token.colorTextTertiary};
                font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
                font-size: ${compact ? 10 : 11}px;
                line-height: 16px;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .script-item .ant-btn {
                width: ${compact ? 29 : 32}px;
                height: ${compact ? 29 : 32}px;
                padding: 0;
                color: ${token.colorTextTertiary};
            }

            .script-item:hover .ant-btn {
                color: ${token.colorPrimary};
            }

            @container terminal-workspace (max-width: 760px) {
                .collapsed & {
                    display: block;
                }
            }
        `,
        scriptRail: css`
            min-width: 0;
            min-height: 0;
            padding: ${compact ? "6px" : "7px"};
            display: none;
            flex: 1;
            flex-direction: column;
            align-items: center;
            gap: 4px;
            overflow-x: hidden;
            overflow-y: auto;
            scrollbar-width: none;

            .collapsed & {
                display: flex;
            }

            &::-webkit-scrollbar {
                display: none;
            }

            .rail-action {
                width: ${compact ? 30 : 32}px;
                height: ${compact ? 30 : 32}px;
                padding: 0;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                flex: none;
                border: 0;
                border-radius: ${token.borderRadiusSM}px;
                outline: none;
                background: transparent;
                color: ${token.colorTextTertiary};
                cursor: pointer;
                font-size: ${compact ? 14 : 15}px;
                transition: background-color .16s ease, color .16s ease;
            }

            .rail-action:hover,
            .rail-action:focus-visible {
                background: ${token.colorFillSecondary};
                color: ${token.colorPrimary};
            }

            .rail-action:focus-visible {
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            @container terminal-workspace (max-width: 760px) {
                .collapsed & {
                    display: none;
                }
            }
        `,
    };
});

export default useStyles;
