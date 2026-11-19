import {memo} from "react";
import {App} from "antd";
import Panel from "./panel";

export default memo((props: any) => {
    return <App style={{display: "flex", flex: 1, flexDirection: "column", minHeight: 0}}><Panel {...props}/></App>;
});
