import {createStyles} from "antd-style";

const useStyles = createStyles<{isCompactMode?: boolean}>(({css, token}, props = {}) => {
    const compact = Boolean(props?.isCompactMode);

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
            container-name: file-workspace;
            container-type: inline-size;
            min-width: 0;
            overflow: hidden;
            border-radius: ${token.borderRadiusLG}px;
            background: ${token.colorBgContainer};
            box-shadow: ${token.boxShadowTertiary};
        `,
        dataPanel: css`
            position: relative;
            min-width: 0;
            padding: 0 ${compact ? 10 : 12}px ${compact ? 10 : 12}px;

            @container file-workspace (max-width: 430px) {
                padding: 0 10px 10px;
            }
        `,
        state: css`
            min-height: ${compact ? 200 : 240}px;
            padding: ${compact ? 18 : 24}px;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: 8px;
            color: ${token.colorTextTertiary};
            font-size: ${token.fontSizeSM}px;

            &.error-state .ant-empty-description {
                color: ${token.colorTextSecondary};
                font-weight: 500;
            }

            &.error-state .ant-btn {
                height: ${compact ? 27 : 30}px;
                border-color: ${token.colorBorderSecondary};
                background: ${token.colorBgContainer};
                color: ${token.colorTextSecondary};
                font-size: ${compact ? 11 : 12}px;
            }

            &.error-state .ant-btn:hover {
                border-color: ${token.colorPrimaryBorder};
                background: ${token.colorFillQuaternary};
                color: ${token.colorPrimary};
            }
        `,
    };
});

export default useStyles;
