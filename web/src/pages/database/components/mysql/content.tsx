import Panel from "./panel";
import {memo} from "react";
import {App, ConfigProvider, theme} from "antd";
import useStyles from "./component-styles/client";
import type {DatabasePanelRecord} from "./entities";
import {useTheme} from "sinking-antd";

const engine = {
    query: "SELECT VERSION();",
    objects: [{key: "view", label: "视图", icon: "EyeOutlined"}, {key: "routine", label: "过程与函数", icon: "FunctionOutlined"}, {key: "trigger", label: "触发器", icon: "ThunderboltOutlined"}],
};

export default memo(({onTabChange, ...props}: {record: DatabasePanelRecord; initialTab?: string; onTabChange?: (tab: string) => void}) => {
    const {styles} = useStyles();
    const {token} = theme.useToken();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());
    const fontSize = compact ? 12 : 13;
    const buttonFontSize = compact ? 11 : 12;
    const controlHeight = compact ? 24 : 28;

    return <ConfigProvider theme={{
        token: {fontSize, fontSizeSM: fontSize},
        components: {
            // 弹窗标题和按钮沿用外层主题字号，工作区的小按钮仅在内容区设置样式。
            Button: {contentFontSize: token.fontSize, contentFontSizeSM: token.fontSize},
            Modal: {titleFontSize: token.fontSizeHeading5, titleLineHeight: token.lineHeightHeading5},
            Input: {controlHeight, inputFontSize: fontSize, inputFontSizeSM: fontSize},
            Select: {controlHeight, optionFontSize: fontSize},
            Table: {fontFamily: token.fontFamily, lineHeight: token.lineHeight, colorText: token.colorTextSecondary, headerColor: token.colorTextSecondary,
                cellFontSize: buttonFontSize, cellFontSizeMD: buttonFontSize, cellFontSizeSM: buttonFontSize,
                cellPaddingBlock: compact ? 6 : 8, cellPaddingBlockMD: compact ? 6 : 8, cellPaddingBlockSM: compact ? 6 : 8,
                cellPaddingInline: compact ? 8 : 10, cellPaddingInlineMD: compact ? 8 : 10, cellPaddingInlineSM: compact ? 8 : 10},
            Segmented: {controlHeight: compact ? 28 : 32, ...(!compact && {controlHeightSM: 28}), itemColor: token.colorTextSecondary},
        },
    }}>
        <App className={`${styles.panel} ${styles.workspace}`}><Panel {...props} engine={engine} compact={compact} onTabChange={onTabChange}/></App>
    </ConfigProvider>;
});
