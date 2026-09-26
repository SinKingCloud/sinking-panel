import {createStyles} from "antd-style";
import type {CreateStylesUtils} from "antd-style";

export const createTerminalStyles = (
    {css, token}: Pick<CreateStylesUtils, "css" | "token">,
    {compact = false, terminalBackground = "#000000"}: {compact?: boolean; terminalBackground?: string} = {},
) => {
    const panelHeaderHeight = compact ? 46 : 50;
    const mobileHeaderHeight = compact ? 42 : 46;

    return {
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
                grid-row: 3;
                width: 100%;
                max-width: 100%;
                height: auto;
                min-height: max(600px, calc(var(--terminal-page-height) - ${compact ? 220 : 244}px));
                overflow: hidden;
                box-sizing: border-box;
            }

            @media (max-width: 780px) and (max-height: 600px) and (orientation: landscape) {
                min-height: 440px;
            }

            @media (min-width: 781px) and (max-height: 600px) and (orientation: landscape) {
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

            .terminal-error {
                width: min(360px, calc(100% - 32px));
                display: flex;
                flex-direction: column;
                align-items: center;
                gap: 10px;
                color: rgba(255, 255, 255, .86);
                text-align: center;
            }

            .terminal-error-title {
                display: inline-flex;
                align-items: center;
                gap: 7px;
                color: #fff;
                font-size: 14px;
                font-weight: 600;
            }

            .terminal-error-title .anticon {
                color: ${token.colorWarning};
                font-size: 15px;
            }

            .terminal-error-message {
                max-width: 100%;
                color: rgba(255, 255, 255, .58);
                font-size: 12px;
                line-height: 20px;
                overflow-wrap: anywhere;
            }

            .terminal-error-actions {
                display: flex;
                align-items: center;
                justify-content: center;
                gap: 8px;
            }

            .terminal-error-actions .ant-btn {
                height: ${compact ? 28 : 30}px;
                padding-inline: 13px;
            }
        `,
    };
};

const useStyles = createStyles(({css, token, isDarkMode}, props: Parameters<typeof createTerminalStyles>[1] = {}) =>
    createTerminalStyles({css, token}, {
        ...props,
        terminalBackground: props.terminalBackground || (isDarkMode ? "rgb(15, 15, 15)" : "#000000"),
    }),
);

export default useStyles;
