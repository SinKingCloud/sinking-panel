import {memo} from "react";
import {App, ConfigProvider, theme} from "antd";
import {useTheme} from "sinking-antd";
import Panel from "./panel";

export default memo((props: any) => {
    const {token} = theme.useToken();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());
    const fontSize = compact ? 12 : 13;
    const buttonFontSize = compact ? 11 : 12;
    const controlHeight = compact ? 24 : 28;
    return <ConfigProvider theme={{token: {fontSize, fontSizeSM: fontSize}, components: {
        Button: {contentFontSize: token.fontSize, contentFontSizeSM: token.fontSize},
        Modal: {titleFontSize: token.fontSizeHeading5, titleLineHeight: token.lineHeightHeading5},
        Input: {controlHeight, inputFontSize: fontSize, inputFontSizeSM: fontSize},
        Select: {controlHeight, optionFontSize: fontSize},
        Table: {fontFamily: token.fontFamily, lineHeight: token.lineHeight, colorText: token.colorTextSecondary, headerColor: token.colorTextSecondary, cellFontSize: buttonFontSize, cellFontSizeMD: buttonFontSize, cellFontSizeSM: buttonFontSize,
            cellPaddingBlock: compact ? 6 : 8, cellPaddingBlockMD: compact ? 6 : 8, cellPaddingBlockSM: compact ? 6 : 8,
            cellPaddingInline: compact ? 8 : 10, cellPaddingInlineMD: compact ? 8 : 10, cellPaddingInlineSM: compact ? 8 : 10},
        Segmented: {controlHeight: compact ? 28 : 32, ...(!compact && {controlHeightSM: 28}), itemColor: token.colorTextSecondary},
    }}}><App style={{display: "flex", flex: 1, flexDirection: "column", minHeight: 0}}><Panel {...props}/></App></ConfigProvider>;
});
