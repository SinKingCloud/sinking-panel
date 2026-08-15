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
            .file-editor-settings {
                width: min(${compact ? 210 : 230}px, calc(100vw - 32px));
                display: flex;
                flex-direction: column;
                gap: ${compact ? 9 : 11}px;
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
        `,
        workspace: css`
            position: relative;
            height: calc(100dvh - 32px);
            min-width: 0;
            min-height: 0;
            display: grid;
            grid-template-columns: ${compact ? 238 : 252}px minmax(0, 1fr);
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

            .file-editor-sidebar {
                min-width: 0;
                min-height: 0;
                overflow: hidden;
                display: flex;
                flex-direction: column;
                border-right: 1px solid ${token.colorSplit};
                background: ${token.colorBgContainer};
                transition: opacity ${token.motionDurationMid}, transform ${token.motionDurationMid};
            }

            &.tree-collapsed .file-editor-sidebar {
                opacity: 0;
                pointer-events: none;
            }

            .file-editor-tree-toolbar {
                min-width: 0;
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

            .file-editor-create-trigger {
                width: auto;
                min-width: max-content;
                height: ${compact ? 30 : 32}px;
                padding: 0 ${compact ? 6 : 8}px;
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
                font-size: ${token.fontSizeSM}px;
                line-height: 1;
                transition: background-color ${token.motionDurationFast};
            }

            .file-editor-create-trigger .marker {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 11 : 12}px;
            }

            .file-editor-create-trigger .arrow {
                color: ${token.colorTextQuaternary};
                font-size: ${compact ? 8 : 9}px;
                transition: transform ${token.motionDurationFast};
            }

            .file-editor-create-trigger:hover,
            .file-editor-create-trigger.ant-dropdown-open {
                background: ${token.colorFillSecondary};
            }

            .file-editor-create-trigger.ant-dropdown-open .arrow {
                transform: rotate(180deg);
            }

            .file-editor-create-trigger:focus-visible {
                border-color: ${token.colorPrimaryBorder};
                box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
            }

            .file-editor-create-trigger:disabled {
                color: ${token.colorTextDisabled};
                cursor: not-allowed;
            }

            .file-editor-tree-actions .ant-btn-text,
            .file-editor-toolbar .ant-btn-text {
                width: ${compact ? 30 : 32}px;
                min-width: ${compact ? 30 : 32}px;
                height: ${compact ? 30 : 32}px;
                padding: 0;
                color: ${token.colorTextSecondary};
                transition: background-color ${token.motionDurationFast}, color ${token.motionDurationFast};
            }

            .file-editor-tree-actions .ant-btn-text:not(:disabled):hover,
            .file-editor-toolbar .ant-btn-text:not(:disabled):hover {
                background: ${token.colorFillSecondary};
                color: ${token.colorText};
            }

            .file-editor-tree-body {
                min-width: 0;
                min-height: 0;
                flex: 1;
                overflow: auto;
                padding: ${compact ? "5px 4px 7px" : "7px 5px 9px"};
                scrollbar-width: thin;
            }

            .file-editor-tree-body > .ant-empty {
                margin-block: 72px 0;
            }

            .file-editor-tree-body .ant-tree {
                min-width: max-content;
                background: transparent;
                color: ${token.colorTextSecondary};
                font-size: ${token.fontSizeSM}px;
            }

            .file-editor-tree-body .ant-tree-treenode {
                width: 100%;
                min-height: ${compact ? 26 : 30}px;
                align-items: center;
            }

            .file-editor-tree-body .ant-tree-switcher {
                height: ${compact ? 24 : 28}px;
                margin-inline-end: 0;
                align-self: center;
                display: inline-flex;
                align-items: center;
                justify-content: flex-end;
                line-height: 1;
            }

            .file-editor-tree-body .ant-tree-switcher::before {
                top: 0;
                width: 100%;
                height: 100%;
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
                min-height: ${compact ? 24 : 28}px;
                padding-inline: 2px ${compact ? 4 : 6}px;
                display: flex;
                align-items: center;
                border-radius: ${token.borderRadiusSM}px;
                line-height: ${compact ? 24 : 28}px;
            }

            .file-editor-tree-body .ant-tree-node-content-wrapper:hover {
                background: ${token.colorFillQuaternary};
            }

            .file-editor-tree-body .ant-tree-node-selected,
            .file-editor-tree-body .ant-tree-node-selected:hover {
                background: ${token.colorPrimaryBg};
                color: ${token.colorText};
            }

            .file-editor-tree-node {
                min-width: 0;
                min-height: ${compact ? 24 : 28}px;
                display: flex;
                align-items: center;
                gap: ${compact ? 6 : 7}px;
                line-height: 1;
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
                max-width: ${compact ? 154 : 170}px;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                line-height: ${compact ? 18 : 20}px;
            }

            .file-editor-pane {
                grid-column: 2;
                min-width: 0;
                min-height: 0;
                display: grid;
                grid-template-rows: auto minmax(0, 1fr) auto;
                background: ${token.colorBgContainer};
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

            .file-editor-toolbar .file-editor-save.ant-btn-text.is-dirty:not(:disabled) {
                background: ${token.colorPrimaryBg};
                color: ${token.colorPrimary};
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

            .file-editor-toolbar .file-editor-close.ant-btn-text:not(:disabled):hover {
                background: ${token.colorErrorBg};
                color: ${token.colorError};
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

            .file-editor-error-retry.ant-btn-text {
                height: ${compact ? 30 : 32}px;
                padding-inline: ${compact ? 10 : 12}px;
                border-radius: ${token.borderRadiusSM}px;
                background: ${token.colorFillQuaternary};
                color: ${token.colorTextSecondary};
                font-size: ${token.fontSizeSM}px;
                font-weight: 500;
            }

            .file-editor-error-retry.ant-btn-text:not(:disabled):hover {
                background: ${token.colorFillSecondary};
                color: ${token.colorPrimary};
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

                .file-editor-tree-actions .ant-btn-text,
                .file-editor-toolbar .ant-btn-text {
                    width: 40px;
                    min-width: 40px;
                    height: 40px;
                }

                .file-editor-create-trigger {
                    height: 40px;
                    padding-inline: 9px;
                }

                .file-editor-tree-toolbar,
                .file-editor-toolbar {
                    min-height: 48px;
                    padding: 0 6px;
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

                .file-editor-toolbar .ant-btn-text {
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
