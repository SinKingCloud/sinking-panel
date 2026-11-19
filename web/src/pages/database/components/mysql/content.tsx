import Panel from "./panel";
import {memo} from "react";
import {App} from "antd";
import useStyles from "./component-styles/client";
import type {DatabasePanelRecord} from "./entities";
import {useTheme} from "sinking-antd";

const engine = {
    query: "SELECT VERSION();",
    objects: [{key: "view", label: "视图", icon: "EyeOutlined"}, {key: "routine", label: "过程与函数", icon: "FunctionOutlined"}, {key: "trigger", label: "触发器", icon: "ThunderboltOutlined"}],
};

export default memo(({onTabChange, ...props}: {record: DatabasePanelRecord; initialTab?: string; onTabChange?: (tab: string) => void}) => {
    const {styles} = useStyles();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());

    return <App className={`${styles.panel} ${styles.workspace}`}><Panel {...props} engine={engine} compact={compact} onTabChange={onTabChange}/></App>;
});
