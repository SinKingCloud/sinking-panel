import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean; dark?: boolean}>(({css, token, isDarkMode}, props = {}) => {
    const compact = Boolean(props.compact);
    const dark = typeof props.dark === "boolean" ? props.dark : Boolean(isDarkMode);
    const chromeBackground = dark
        ? `color-mix(in srgb, #fff 5%, ${token.colorBgContainer})`
        : token.colorBgContainer;
    return {
        modal: css`
            .ant-modal {
                max-width: calc(100vw - 24px);
                margin-inline: auto;
            }

            .ant-modal-container {
                max-height: calc(100dvh - 16px);
                padding: 0;
                overflow: hidden;
                display: flex;
                flex-direction: column;
            }

            .ant-modal-header {
                position: absolute;
                width: 1px;
                height: 1px;
                margin: -1px;
                padding: 0;
                overflow: hidden;
                clip: rect(0 0 0 0);
                clip-path: inset(50%);
                white-space: nowrap;
                border: 0;
            }

            .ant-modal-body {
                min-width: 0;
                min-height: 0;
                overflow: hidden;
                padding: 0;
            }
        `,
        settingsPopup: css`
            position: absolute;
            z-index: ${token.zIndexPopupBase + 30};
            inset-block-start: calc(100% + ${compact ? 5 : 7}px);
            inset-inline-end: 0;
            --file-editor-settings-arrow-x: calc(100% - 88px);
            width: min(${compact ? 232 : 254}px, calc(100vw - 32px));
            box-sizing: border-box;
            padding: ${compact ? 10 : 12}px;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusLG}px;
            opacity: 0;
            pointer-events: none;
            transform: translateY(-5px) scale(0.985);
            transform-origin: var(--file-editor-settings-arrow-x) top;
            background: ${token.colorBgElevated};
            box-shadow: ${token.boxShadowSecondary};
            transition:
                opacity ${token.motionDurationFast},
                transform ${token.motionDurationFast};

            &::before,
            &::after {
                content: "";
                position: absolute;
                z-index: 1;
                inset-block-start: -7px;
                left: var(--file-editor-settings-arrow-x);
                width: 0;
                height: 0;
                border-right: 7px solid transparent;
                border-bottom: 7px solid ${token.colorBorderSecondary};
                border-left: 7px solid transparent;
                pointer-events: none;
                transform: translateX(-50%);
            }

            &::after {
                inset-block-start: -6px;
                border-right-width: 6px;
                border-bottom: 6px solid ${token.colorBgElevated};
                border-left-width: 6px;
            }

            &.is-open {
                opacity: 1;
                pointer-events: auto;
                transform: translateY(0) scale(1);
                animation: file-editor-settings-enter ${token.motionDurationFast};
            }

            @keyframes file-editor-settings-enter {
                from {
                    opacity: 0;
                    transform: translateY(-5px) scale(0.985);
                }
                to {
                    opacity: 1;
                    transform: translateY(0) scale(1);
                }
            }

            .file-editor-settings {
                width: 100%;
                max-height: calc(100dvh - ${compact ? 84 : 88}px);
                display: flex;
                flex-direction: column;
                gap: ${compact ? 9 : 11}px;
                overflow-y: auto;
                scrollbar-width: thin;
            }

            .file-editor-setting-row {
                display: grid;
                grid-template-columns: minmax(72px, 1fr) minmax(100px, 1.25fr);
                align-items: center;
                gap: 12px;
                color: ${token.colorTextSecondary};
                font-size: ${token.fontSizeSM}px;
            }

            .file-editor-setting-row .ant-select,
            .file-editor-setting-row .ant-input-number,
            .file-editor-setting-row .ant-segmented {
                width: 100%;
            }

            .file-editor-setting-row .ant-switch {
                justify-self: end;
            }

            @media (prefers-reduced-motion: reduce) {
                transition: none;
                animation: none;
            }
        `,
        treeMenu: css`
            && {
                max-width: calc(100vw - 24px);
                box-sizing: border-box;
                padding: 0;
                border-radius: ${token.borderRadius}px;
            }

            && .ant-dropdown-menu {
                min-width: 132px;
                padding: 2px !important;
                border-radius: ${token.borderRadius}px !important;
                background: ${token.colorBgElevated};
                box-shadow: ${token.boxShadowSecondary};
            }

            && .ant-dropdown-menu-item,
            && .ant-dropdown-menu-submenu-title {
                min-height: ${compact ? 26 : 30}px !important;
                margin: 0 !important;
                padding: 0 8px !important;
                border-radius: ${token.borderRadiusSM}px !important;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px !important;
            }

            && .ant-dropdown-menu-item:hover,
            && .ant-dropdown-menu-submenu-title:hover {
                background: ${token.colorFillQuaternary};
            }

            && .ant-dropdown-menu-item-danger:not(.ant-dropdown-menu-item-disabled) {
                color: ${token.colorError};
            }

            && .ant-dropdown-menu-item-danger:not(.ant-dropdown-menu-item-disabled):hover {
                background: ${token.colorErrorBg};
                color: ${token.colorError};
            }
        `,
        workspace: css`
            --file-editor-tree-width: ${compact ? 238 : 252}px;
            position: relative;
            height: calc(100dvh - 32px);
            min-width: 0;
            min-height: 0;
            display: grid;
            grid-template-columns: var(--file-editor-tree-width) minmax(0, 1fr);
            overflow: hidden;
            background: ${token.colorBgContainer};
            transition: grid-template-columns ${token.motionDurationMid};

            &:fullscreen {
                width: 100vw;
                height: 100dvh;
                max-width: none;
                min-height: 0;
                border: 0;
                border-radius: 0;
                background: ${token.colorBgContainer};
            }

            &:fullscreen::backdrop {
                background: ${token.colorBgContainer};
            }

            &.tree-collapsed {
                grid-template-columns: 0 minmax(0, 1fr);
            }

            &.tree-resizing {
                transition: none;
                cursor: col-resize;
                user-select: none;
                -webkit-user-select: none;
            }

            &.tree-resizing .file-editor-sidebar,
            &.tree-resizing .file-editor-pane {
                pointer-events: none;
            }

            .file-editor-tree-resize-handle {
                position: absolute;
                z-index: 5;
                inset-block: 0;
                inset-inline-start: calc(var(--file-editor-tree-width) - 3px);
                width: 6px;
                outline: none;
                cursor: col-resize;
                touch-action: none;
                user-select: none;
                -webkit-user-select: none;
            }

            .file-editor-tree-resize-handle::before {
                content: "";
                position: absolute;
                inset-block: 0;
                inset-inline-start: 2px;
                width: 1px;
                background: ${token.colorPrimary};
                opacity: 0;
                transition: opacity ${token.motionDurationFast};
                pointer-events: none;
            }

            .file-editor-tree-resize-handle:hover::before,
            .file-editor-tree-resize-handle:focus-visible::before,
            .file-editor-tree-resize-handle.is-dragging::before {
                opacity: 1;
            }

            .file-editor-sidebar {
                min-width: 0;
                min-height: 0;
                overflow: hidden;
                display: flex;
                flex-direction: column;
                border-right: 1px solid ${token.colorSplit};
                background: ${token.colorBgContainer};
                contain: layout;
                transition: opacity ${token.motionDurationMid}, transform ${token.motionDurationMid};
            }

            &.tree-collapsed .file-editor-sidebar {
                opacity: 0;
                pointer-events: none;
            }

            .file-editor-tree-toolbar {
                width: var(--file-editor-tree-width);
                min-width: 0;
                box-sizing: border-box;
                min-height: ${compact ? 42 : 46}px;
                padding: 0 ${compact ? 9 : 11}px;
                display: flex;
                align-items: center;
                justify-content: space-between;
                gap: 8px;
                border-bottom: 1px solid ${token.colorSplit};
                background: ${chromeBackground};
            }

            .file-editor-tree-target {
                min-width: 0;
                flex: 1;
                display: flex;
                align-items: center;
                gap: 7px;
                color: ${token.colorTextSecondary};
                font-size: ${token.fontSizeSM}px;
                font-weight: 500;
                line-height: 20px;
                user-select: none;
                -webkit-user-select: none;
                cursor: copy;
            }

            .file-editor-tree-target > .anticon {
                flex: none;
                color: ${token.colorPrimary};
            }

            .file-editor-tree-target > span {
                min-width: 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .file-editor-tree-actions {
                flex: none;
                display: flex;
                align-items: center;
                gap: 2px;
            }

            .file-editor-tree-actions .ant-btn,
            .file-editor-toolbar .ant-btn {
                width: ${compact ? 30 : 32}px;
                min-width: ${compact ? 30 : 32}px;
                height: ${compact ? 30 : 32}px;
                padding: 0;
            }

            .file-editor-tree-body {
                width: var(--file-editor-tree-width);
                min-width: 0;
                min-height: 0;
                box-sizing: border-box;
                flex: 1;
                overflow: hidden;
                padding: ${compact ? "5px 4px 7px" : "7px 5px 9px"};
            }

            .file-editor-tree-body > .ant-empty {
                margin-block: 72px 0;
            }

            .file-editor-tree-loading {
                min-height: 180px;
                display: flex;
                align-items: center;
                justify-content: center;
                color: ${token.colorTextTertiary};
            }

            .file-editor-tree-body .ant-tree {
                width: 100%;
                min-width: 0;
                background: transparent;
                color: ${token.colorTextSecondary};
                font-size: ${token.fontSizeSM}px;
            }

            .file-editor-tree-body .ant-tree-treenode {
                --file-editor-tree-row-background: transparent;
                width: 100%;
                height: ${compact ? 26 : 30}px;
                min-height: ${compact ? 26 : 30}px;
                margin-block: 0 !important;
                padding-block: 0 !important;
                align-items: center;
                position: relative;
                border-radius: ${token.borderRadiusSM}px;
                background: var(--file-editor-tree-row-background);
            }

            .file-editor-tree-body .ant-tree-list-holder {
                overflow-x: hidden !important;
                overscroll-behavior: contain;
                scrollbar-width: thin;
            }

            .file-editor-tree-body .ant-tree-list-holder-inner {
                box-sizing: border-box;
                /* Keep the virtual scrollbar's 8px track clear, including when hidden. */
                padding-inline-end: 8px;
            }

            .file-editor-tree-body .ant-tree-treenode:hover {
                --file-editor-tree-row-background: ${token.colorFillQuaternary};
            }

            .file-editor-tree-body .ant-tree-treenode.ant-tree-treenode-selected,
            .file-editor-tree-body .ant-tree-treenode.ant-tree-treenode-selected:hover {
                --file-editor-tree-row-background: ${token.colorPrimaryBg};
            }

            .file-editor-tree-body .ant-tree-switcher {
                width: ${compact ? 22 : 24}px;
                min-width: ${compact ? 22 : 24}px;
                height: ${compact ? 26 : 30}px;
                margin-inline-end: 0;
                align-self: center;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                line-height: 1;
                position: relative;
                z-index: 1;
                background: transparent !important;
            }

            .file-editor-tree-body .ant-tree-switcher::before {
                top: 0;
                width: 100%;
                height: 100%;
                background: transparent !important;
            }

            .file-editor-tree-body .ant-tree-switcher-icon {
                width: 14px;
                height: 14px;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                line-height: 1;
                vertical-align: middle;
            }

            .file-editor-tree-body .ant-tree-switcher-icon > svg {
                display: block;
            }

            .file-editor-tree-body .ant-tree-node-content-wrapper {
                min-width: 0;
                height: ${compact ? 26 : 30}px;
                min-height: ${compact ? 26 : 30}px;
                padding-inline: 2px 0;
                display: flex;
                align-items: center;
                border-radius: ${token.borderRadiusSM}px;
                line-height: ${compact ? 26 : 30}px;
                position: relative;
                z-index: 1;
                background: transparent !important;
            }

            .file-editor-tree-body .ant-tree-title {
                min-width: 0;
                flex: 1;
                user-select: none;
                -webkit-user-select: none;
            }

            .file-editor-tree-body .ant-tree-title > .ant-dropdown-trigger {
                display: block;
                width: 100%;
            }

            .file-editor-tree-body .ant-tree-node-selected,
            .file-editor-tree-body .ant-tree-node-selected:hover {
                background: transparent !important;
                color: ${token.colorText};
            }

            .file-editor-tree-node {
                width: 100%;
                min-width: 0;
                height: ${compact ? 26 : 30}px;
                min-height: ${compact ? 26 : 30}px;
                display: flex;
                align-items: center;
                gap: ${compact ? 6 : 7}px;
                line-height: 1;
                user-select: none;
                -webkit-user-select: none;
            }

            .file-editor-tree-node-shell {
                width: 100%;
                min-width: 0;
                height: ${compact ? 26 : 30}px;
                min-height: ${compact ? 26 : 30}px;
                display: flex;
                align-items: center;
                position: relative;
                user-select: none;
                -webkit-user-select: none;
            }

            .file-editor-tree-node-shell > .file-editor-tree-node {
                flex: 1;
                width: auto;
            }

            .file-editor-tree-node-actions {
                position: absolute;
                inset-inline-end: 0;
                inset-block: 0;
                z-index: 1;
                display: flex;
                align-items: center;
                background: linear-gradient(var(--file-editor-tree-row-background), var(--file-editor-tree-row-background)), ${token.colorBgContainer};
                opacity: 0;
                pointer-events: none;
            }

            .file-editor-tree-node-more {
                width: ${compact ? 26 : 28}px;
                min-width: ${compact ? 26 : 28}px;
                height: ${compact ? 26 : 28}px;
                flex: none;
                padding: 0;
                color: ${token.colorTextTertiary};
                transition: background-color ${token.motionDurationFast}, color ${token.motionDurationFast};
            }

            .file-editor-tree-body .ant-tree-treenode:hover .file-editor-tree-node-actions,
            .file-editor-tree-node-shell:focus-within .file-editor-tree-node-actions,
            .file-editor-tree-node-actions.is-open {
                opacity: 1;
                pointer-events: auto;
            }

            .file-editor-tree-node-more:hover,
            .file-editor-tree-node-more:focus-visible {
                background: ${token.colorFillSecondary};
                color: ${token.colorText};
            }

            .file-editor-tree-node-more:disabled {
                cursor: not-allowed;
                opacity: .5;
            }

            .file-editor-tree-node > .anticon,
            .file-editor-tree-node > .ant-spin {
                width: 15px;
                height: 15px;
                flex: none;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 13 : 14}px;
                line-height: 1;
            }

            .file-editor-tree-node > .anticon > svg {
                display: block;
            }

            .file-editor-tree-node > .is-directory {
                color: ${token.colorPrimary};
            }

            .file-editor-tree-node.is-more {
                color: ${token.colorTextTertiary};
                font-style: normal;
            }

            .file-editor-tree-name {
                min-width: 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                line-height: ${compact ? 18 : 20}px;
                user-select: none;
                -webkit-user-select: none;
            }

            @media (pointer: coarse) {
                .file-editor-tree-node-shell > .file-editor-tree-node {
                    padding-inline-end: ${compact ? 26 : 28}px;
                }

                .file-editor-tree-node-actions {
                    opacity: 1;
                    pointer-events: auto;
                }
            }

            .file-editor-pane {
                grid-column: 2;
                min-width: 0;
                min-height: 0;
                display: grid;
                grid-template-rows: auto minmax(0, 1fr) auto;
                background: ${token.colorBgContainer};
                contain: layout paint;
            }

            .file-editor-toolbar {
                min-width: 0;
                min-height: ${compact ? 42 : 46}px;
                padding: 0 ${compact ? 7 : 9}px;
                display: flex;
                align-items: center;
                gap: ${compact ? 4 : 6}px;
                border-bottom: 1px solid ${token.colorSplit};
                background: ${chromeBackground};
            }

            .file-editor-tabs,
            .file-editor-tabs-empty {
                min-width: 0;
                flex: 1;
                display: flex;
                align-items: center;
                align-self: stretch;
            }

            .file-editor-tabs {
                gap: 2px;
                padding-block: 4px;
                overflow-x: auto;
                overflow-y: hidden;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
                touch-action: pan-x;
                -webkit-overflow-scrolling: touch;
            }

            .file-editor-tab,
            .file-editor-tab-main {
                touch-action: pan-x;
            }

            .file-editor-tabs::-webkit-scrollbar {
                display: none;
            }

            .file-editor-tabs-empty {
                padding-inline: 8px;
                color: ${token.colorTextTertiary};
                font-size: ${token.fontSizeSM}px;
            }

            .file-editor-tab {
                position: relative;
                height: ${compact ? 30 : 34}px;
                min-width: ${compact ? 108 : 120}px;
                max-width: ${compact ? 172 : 196}px;
                flex: 0 1 auto;
                display: flex;
                align-items: center;
                border-radius: ${token.borderRadiusSM}px;
                background: transparent;
                color: ${token.colorTextSecondary};
                transition: background-color ${token.motionDurationFast}, color ${token.motionDurationFast};
            }

            .file-editor-tab:hover {
                background: ${token.colorFillQuaternary};
            }

            .file-editor-tab::after {
                position: absolute;
                right: 0;
                bottom: 0;
                left: 0;
                height: 2px;
                border-radius: 2px 2px 0 0;
                background: transparent;
                content: "";
            }

            .file-editor-tab.is-active {
                background: ${token.colorFillSecondary};
                color: ${token.colorText};
            }

            .file-editor-tab.is-active::after {
                background: ${token.colorPrimary};
            }

            .file-editor-tab.has-error .file-editor-tab-main > .anticon {
                color: ${token.colorError};
            }

            .file-editor-tab-main {
                min-width: 0;
                height: 100%;
                flex: 1;
                padding: 0 2px 0 ${compact ? 8 : 10}px;
                display: flex;
                align-items: center;
                gap: ${compact ? 5 : 6}px;
                overflow: hidden;
                border: 0;
                outline: none;
                background: transparent;
                color: inherit;
                cursor: pointer;
                font: inherit;
                text-align: start;
            }

            .file-editor-tab-main:focus-visible {
                box-shadow: inset 0 0 0 2px ${token.colorPrimaryBorder};
            }

            .file-editor-tab-main > .anticon {
                flex: none;
                color: ${token.colorTextTertiary};
                font-size: ${compact ? 12 : 13}px;
            }

            .file-editor-tab-main > span {
                min-width: 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .file-editor-tab-main > i {
                width: 6px;
                height: 6px;
                flex: none;
                border-radius: 50%;
                background: ${token.colorWarning};
            }

            .file-editor-tab-close.ant-btn {
                width: ${compact ? 24 : 26}px;
                min-width: ${compact ? 24 : 26}px;
                height: ${compact ? 24 : 26}px;
                margin-inline-end: 2px;
                opacity: 0;
                pointer-events: none;
                color: ${token.colorTextTertiary};
            }

            .file-editor-tab:hover .file-editor-tab-close,
            .file-editor-tab.is-active .file-editor-tab-close,
            .file-editor-tab-close:focus-visible {
                opacity: 1;
                pointer-events: auto;
            }

            .file-editor-tab-close.ant-btn:not(:disabled):hover {
                background: ${token.colorFillTertiary};
                color: ${token.colorText};
            }

            .file-editor-toolbar-actions {
                position: relative;
                display: flex;
                flex: none;
                align-items: center;
                gap: 2px;
                margin-inline-start: auto;
                overflow: visible;
            }

            .file-editor-settings-trigger {
                position: static;
                display: inline-flex;
                flex: none;
            }

            .file-editor-toolbar-divider {
                width: 1px;
                height: 16px;
                margin-inline: 4px 2px;
                background: ${token.colorSplit};
            }

            .file-editor-close-wrap {
                display: inline-flex;
            }

            .file-editor-canvas {
                position: relative;
                min-width: 0;
                min-height: 0;
                overflow: hidden;
                background: ${token.colorBgContainer};
            }

            .file-editor-ace {
                width: 100%;
                height: 100%;
                overflow: hidden;

                /* 只恢复容器内的纵向滚动条，横向仍可滚动但不显示滚动条。 */
                & ::-webkit-scrollbar {
                    width: 8px !important;
                    height: 0 !important;
                }
            }

            .file-editor-loading,
            .file-editor-empty,
            .file-editor-error {
                position: absolute;
                inset: 0;
                display: flex;
                align-items: center;
                justify-content: center;
                background: ${token.colorBgContainer};
            }

            .file-editor-error {
                padding: 24px;
            }

            .file-editor-error .ant-empty-description {
                max-width: 420px;
                overflow-wrap: anywhere;
                color: ${token.colorTextSecondary};
                line-height: 1.6;
            }

            .file-editor-error .ant-empty-footer {
                margin-top: ${compact ? 10 : 12}px;
            }

            .file-editor-status {
                min-width: 0;
                min-height: ${compact ? 25 : 28}px;
                padding-inline: ${compact ? 8 : 10}px;
                display: flex;
                align-items: center;
                justify-content: space-between;
                gap: 12px;
                border-top: 1px solid ${token.colorSplit};
                background: ${token.colorFillQuaternary};
                color: ${token.colorTextTertiary};
                font-size: ${token.fontSizeSM}px;
                font-variant-numeric: tabular-nums;
            }

            .file-editor-status-path,
            .file-editor-status-meta {
                min-width: 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .file-editor-status-path {
                flex: 1;
                font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            }

            .file-editor-status-meta {
                display: flex;
                flex: none;
                align-items: center;
                gap: 5px;
            }

            .file-editor-status-state::before {
                content: "· ";
            }

            @media (max-width: 767px) {
                &:not(:fullscreen) {
                    height: calc(100dvh - 16px);
                    min-height: 0;
                }
            }

            @media (max-width: 720px) {
                grid-template-columns: minmax(0, 1fr);

                &.tree-collapsed {
                    grid-template-columns: minmax(0, 1fr);
                }

                .file-editor-tree-resize-handle {
                    display: none;
                }

                .file-editor-sidebar {
                    position: absolute;
                    z-index: 6;
                    inset: 48px 0 0;
                    width: 100%;
                    border-right: 0;
                    box-shadow: none;
                }

                .file-editor-pane {
                    width: 100%;
                    grid-column: 1;
                }

                &.tree-collapsed .file-editor-sidebar {
                    opacity: 0;
                    transform: translateX(-100%);
                }

                .file-editor-tree-actions .ant-btn,
                .file-editor-toolbar .ant-btn {
                    width: 40px;
                    min-width: 40px;
                    height: 40px;
                }

                .file-editor-tree-toolbar,
                .file-editor-toolbar {
                    min-height: 48px;
                    padding: 0 6px;
                }

                .file-editor-tree-toolbar,
                .file-editor-tree-body {
                    width: 100%;
                }

                .file-editor-tab {
                    width: max-content;
                    min-width: 132px;
                    max-width: none;
                    flex: 0 0 auto;
                }

                .file-editor-tab-main {
                    min-width: max-content;
                    flex: none;
                    overflow: visible;
                }

                .file-editor-tab-main > span {
                    overflow: visible;
                    text-overflow: clip;
                }

                .file-editor-tab-close.ant-btn {
                    width: 32px;
                    min-width: 32px;
                    height: 32px;
                    opacity: 1;
                    pointer-events: auto;
                }
            }

            @media (max-width: 520px) {
                .file-editor-toolbar {
                    gap: 2px;
                }

                .file-editor-toolbar .ant-btn {
                    width: 40px;
                    min-width: 40px;
                    height: 40px;
                }

                .file-editor-toolbar .file-editor-tab-close.ant-btn {
                    width: 32px;
                    min-width: 32px;
                    height: 32px;
                }

                .file-editor-toolbar .file-editor-reload {
                    display: none;
                }

                .file-editor-status-path,
                .file-editor-status-detail {
                    display: none;
                }

                .file-editor-status-meta {
                    margin-inline-start: auto;
                }

                .file-editor-status-state::before {
                    content: "";
                }
            }

            @media (max-width: 360px) {
                &:not(:fullscreen) .file-editor-fullscreen {
                    display: none;
                }

                .file-editor-toolbar .ant-btn {
                    width: 36px;
                    min-width: 36px;
                    height: 36px;
                }

                .file-editor-toolbar .file-editor-tab-close.ant-btn {
                    width: 28px;
                    min-width: 28px;
                    height: 28px;
                }

                .file-editor-toolbar-divider {
                    display: none;
                }

            }

        `,
    };
});

export default useStyles;
