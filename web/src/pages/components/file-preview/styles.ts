import {createStyles} from "antd-style";

const useStyles = createStyles<{compact?: boolean}>(({css, token}, props = {}) => {
    const compact = Boolean(props.compact);
    return {
        root: css`
            .ant-modal {
                max-width: calc(100vw - 24px);
            }

            .ant-modal-content {
                overflow: hidden;
            }

            .ant-modal-body {
                min-width: 0;
                padding-top: ${compact ? 8 : 12}px;
            }

            .file-preview {
                min-width: 0;
                display: flex;
                flex-direction: column;
                gap: ${compact ? 8 : 10}px;
            }

            .file-preview-meta {
                min-width: 0;
                display: flex;
                align-items: center;
                gap: ${compact ? 7 : 9}px;
            }

            .file-preview-meta-icon {
                width: ${compact ? 22 : 24}px;
                height: ${compact ? 22 : 24}px;
                flex: none;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 15 : 16}px;
            }

            .file-preview-name {
                min-width: 0;
                flex: 1;
                overflow: hidden;
                color: ${token.colorText};
                font-size: ${token.fontSize}px;
                font-weight: 500;
                line-height: ${compact ? 20 : 22}px;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .file-preview-position {
                flex: none;
                color: ${token.colorTextTertiary};
                font-size: ${token.fontSizeSM}px;
                font-variant-numeric: tabular-nums;
                white-space: nowrap;
            }

            .file-preview-stage {
                position: relative;
                height: clamp(280px, 56dvh, 560px);
                min-width: 0;
                overflow: hidden;
                border-radius: ${token.borderRadius}px;
                background: ${token.colorFillQuaternary};
            }

            .file-preview-stage.is-embedded {
                width: 100%;
                height: 100%;
                border-radius: 0;
            }

            .file-preview-stage:fullscreen {
                width: 100vw;
                height: 100vh;
                border-radius: 0;
                background: #000;
            }

            .file-preview-media {
                width: 100%;
                height: 100%;
                display: block;
            }

            .file-preview-image,
            .file-preview-video {
                object-fit: contain;
            }

            .file-preview-video {
                background: #000;
            }

            .file-preview-audio {
                width: 100%;
                height: 100%;
                padding: 24px;
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                gap: 18px;
            }

            .file-preview-audio-icon {
                color: ${token.colorTextTertiary};
                font-size: 36px;
                line-height: 1;
            }

            .file-preview-audio audio {
                width: min(100%, 560px);
            }

            .file-preview-loading,
            .file-preview-error {
                position: absolute;
                inset: 0;
                z-index: 2;
                display: flex;
                align-items: center;
                justify-content: center;
                background: ${token.colorFillQuaternary};
            }

            .file-preview-error {
                padding: 24px;
            }

            .file-preview-fullscreen.ant-btn {
                position: absolute;
                top: ${compact ? 10 : 14}px;
                right: ${compact ? 10 : 14}px;
                z-index: 4;
                width: ${compact ? 36 : 40}px;
                min-width: ${compact ? 36 : 40}px;
                height: ${compact ? 36 : 40}px;
                padding: 0;
                border: 0;
                border-radius: 50%;
                background: ${token.colorBgElevated};
                color: ${token.colorTextSecondary};
                box-shadow: ${token.boxShadowSecondary};
                transition: opacity ${token.motionDurationFast}, background-color ${token.motionDurationFast}, color ${token.motionDurationFast};
            }

            .file-preview-fullscreen.ant-btn:hover,
            .file-preview-fullscreen.ant-btn:focus-visible {
                background: ${token.colorBgElevated};
                color: ${token.colorPrimary};
            }

            .file-preview-nav.ant-btn {
                position: absolute;
                top: 50%;
                z-index: 3;
                width: ${compact ? 38 : 42}px;
                min-width: ${compact ? 38 : 42}px;
                height: ${compact ? 38 : 42}px;
                padding: 0;
                border: 0;
                border-radius: 50%;
                background: ${token.colorBgElevated};
                color: ${token.colorTextSecondary};
                box-shadow: ${token.boxShadowSecondary};
                transform: translateY(-50%);
                transition: opacity ${token.motionDurationFast}, background-color ${token.motionDurationFast}, color ${token.motionDurationFast};
            }

            .file-preview-stage.file-preview-controls-hidden .file-preview-fullscreen.ant-btn,
            .file-preview-stage.file-preview-controls-hidden .file-preview-nav.ant-btn {
                opacity: 0;
                pointer-events: none;
            }

            .file-preview-stage.file-preview-controls-hidden .file-preview-fullscreen.ant-btn:focus-visible,
            .file-preview-stage.file-preview-controls-hidden .file-preview-nav.ant-btn:focus-visible {
                opacity: 1;
                pointer-events: auto;
            }

            .file-preview-nav.ant-btn:not(:disabled):hover,
            .file-preview-nav.ant-btn:not(:disabled):focus-visible {
                background: ${token.colorBgElevated};
                color: ${token.colorPrimary};
            }

            .file-preview-stage:fullscreen .file-preview-fullscreen.ant-btn,
            .file-preview-stage:fullscreen .file-preview-nav.ant-btn {
                background: rgba(24, 24, 24, .46);
                color: rgba(255, 255, 255, .9);
                -webkit-backdrop-filter: blur(12px);
                backdrop-filter: blur(12px);
            }

            .file-preview-stage:fullscreen .file-preview-fullscreen.ant-btn:hover,
            .file-preview-stage:fullscreen .file-preview-fullscreen.ant-btn:focus-visible,
            .file-preview-stage:fullscreen .file-preview-nav.ant-btn:not(:disabled):hover,
            .file-preview-stage:fullscreen .file-preview-nav.ant-btn:not(:disabled):focus-visible {
                background: rgba(48, 48, 48, .62);
                color: #fff;
            }

            .file-preview-nav.is-previous {
                left: ${compact ? 10 : 14}px;
            }

            .file-preview-nav.is-next {
                right: ${compact ? 10 : 14}px;
            }

            @media (max-width: 600px) {
                .ant-modal-body {
                    padding-inline: 12px;
                }

                .file-preview-stage {
                    height: clamp(240px, 48dvh, 420px);
                }

                .file-preview-stage.is-embedded {
                    height: 100%;
                }

                .file-preview-nav.ant-btn {
                    width: 44px;
                    min-width: 44px;
                    height: 44px;
                }

                .file-preview-fullscreen.ant-btn {
                    top: 8px;
                    right: 8px;
                    width: 44px;
                    min-width: 44px;
                    height: 44px;
                }

                .file-preview-nav.is-previous {
                    left: 8px;
                }

                .file-preview-nav.is-next {
                    right: 8px;
                }
            }
        `,
    };
});

export default useStyles;
