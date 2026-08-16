import {createStyles} from "antd-style";
import Settings from "@/../config/defaultSettings";

const backgroundUrl = `${String(Settings?.basePath || "/").replace(/\/?$/, "/")}images/background.svg`;

export default createStyles(({css, token, isDarkMode}: any, props: any = {}) => {
    const compact = Boolean(props?.isCompactMode);
    const dark = typeof props?.isDarkMode === "boolean" ? props.isDarkMode : Boolean(isDarkMode);

    return {
        body: css`
            min-height: 100dvh;
            padding: 0 !important;
            background: ${token.colorBgContainer};
        `,
        screen: css`
            position: relative;
            width: 100%;
            min-height: 100dvh;
            padding: ${compact ? 32 : 40}px 24px;
            display: flex;
            align-items: center;
            justify-content: center;
            overflow-x: hidden;
            overflow-y: auto;
            background: ${token.colorBgContainer};

            @media (max-width: 640px) {
                padding: 24px 12px;
            }

            @media (max-width: 640px) and (max-height: 520px) {
                padding-top: 24px;
                align-items: flex-start;
            }

            @media (min-width: 641px) and (max-height: 620px) {
                padding-top: 28px;
                align-items: flex-start;
            }
        `,
        backdrop: css`
            position: absolute;
            z-index: 0;
            inset: 0;
            pointer-events: none;
            background-image: url('${backgroundUrl}');
            background-repeat: no-repeat;
            background-position: center 72px;
            background-size: min(1361px, 100%) auto;
            opacity: ${dark ? .16 : 1};

            @media (max-width: 640px) {
                background-position: center 8px;
                background-size: 100% auto;
                opacity: ${dark ? .20 : .90};
            }
        `,
        loginPanel: css`
            position: relative;
            z-index: 1;
            width: 100%;
            max-width: ${compact ? 392 : 424}px;
            padding: ${compact ? "26px 32px 30px" : "30px 36px 34px"};
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusLG}px;
            background: ${token.colorBgElevated};
            box-shadow: ${token.boxShadowTertiary};
            transform: translateY(${compact ? -36 : -44}px);

            .brand {
                display: flex;
                align-items: center;
                justify-content: center;
                gap: ${compact ? 10 : 12}px;

                > .anticon {
                    flex: none;
                    color: ${token.colorPrimary};
                    font-size: ${compact ? 29 : 34}px;
                }

                h1 {
                    min-width: 0;
                    margin: 0;
                    overflow: hidden;
                    color: ${token.colorTextHeading};
                    font-size: ${compact ? 24 : 28}px;
                    font-weight: 700;
                    line-height: ${compact ? 33 : 38}px;
                    letter-spacing: 0;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                }
            }

            .product {
                margin: ${compact ? 5 : 7}px 0 0;
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 12 : token.fontSize}px;
                line-height: 22px;
                text-align: center;
            }

            @media (max-width: 640px) {
                max-width: ${compact ? 340 : 360}px;
                padding: ${compact ? "22px 20px 24px" : "24px 20px 26px"};

                .brand > .anticon {
                    font-size: ${compact ? 26 : 30}px;
                }

                .brand h1 {
                    font-size: ${compact ? 24 : 28}px;
                    font-weight: 800;
                    line-height: ${compact ? 32 : 36}px;
                }
            }

            @media (max-width: 360px) {
                padding: ${compact ? "20px 18px 22px" : "22px 18px 24px"};
            }

            @media (max-width: 640px) and (max-height: 520px) {
                transform: none;
            }

            @media (min-width: 641px) and (max-height: 620px) {
                transform: none;
            }
        `,
        form: css`
            margin-top: ${compact ? 20 : 24}px;
            display: flex;
            flex-direction: column;

            .field-item {
                margin-bottom: 0;
            }

            .field-item .ant-form-item-control {
                min-height: ${compact ? 60 : 66}px;
            }

            .ant-form-item-additional {
                min-height: 20px;
            }

            .ant-form-item-explain-error {
                font-size: 12px;
                line-height: 20px;
            }

            .submit-item {
                margin-top: ${compact ? 4 : 6}px;
                margin-bottom: 0;
            }
        `,
        input: css`
            &.ant-input-affix-wrapper,
            &.ant-input {
                height: ${compact ? 40 : 44}px;
                padding-inline: ${compact ? 11 : 13}px;
                border-color: transparent;
                border-radius: ${token.borderRadius}px;
                background: ${token.colorFillQuaternary};
                box-shadow: none;
                transition: border-color .18s ease, background-color .18s ease, box-shadow .18s ease;

                > .anticon {
                    color: ${token.colorTextTertiary};
                    font-size: ${compact ? 14 : 15}px;
                }

                &:hover {
                    border-color: ${token.colorPrimaryBorderHover};
                }

                &.ant-input-affix-wrapper-focused,
                &:focus {
                    border-color: ${token.colorPrimary};
                    background: ${token.colorBgContainer};
                    box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
                }

                .ant-input {
                    background: transparent;
                    font-size: ${token.fontSize}px;
                }
            }
        `,
        submit: css`
            height: ${compact ? 40 : 44}px;
            border-radius: ${token.borderRadius}px;
            font-size: ${token.fontSize}px;
            font-weight: 600;
            box-shadow: none;
        `,
    };
});
