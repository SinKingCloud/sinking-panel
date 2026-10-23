import {useState} from "react";
import {Segmented} from "antd";
import {useTheme} from "sinking-antd";
import Documents from "./documents";
import Metadata from "./metadata";
import Status from "./status";
import Request from "./request";
import useStyles from "./styles";

const tabs = [{value: "documents", label: "数据浏览"}, {value: "mapping", label: "索引结构"},
    {value: "status", label: "集群状态"}, {value: "request", label: "请求执行"}];

export default ({record, initialTab, onTabChange}: any) => {
    const appearance = useTheme();
    const {styles} = useStyles({compact: Boolean(appearance?.isCompactTheme?.())});
    const [tab, setTab] = useState(tabs.some((item) => item.value === initialTab) ? initialTab : record.name ? "documents" : "status");
    const [requestOpened, setRequestOpened] = useState(initialTab === "request");
    return <div className={styles.panel}>
        <header className="manager-header"><Segmented className="manager-navigation" aria-label="Elasticsearch 管理功能" value={tab} options={tabs} onChange={(value) => { setTab(value); if (value === "request") setRequestOpened(true); onTabChange?.(value); }}/></header>
        <div className="manager-body">
            {tab === "documents" ? <Documents record={record}/>
                : tab === "mapping" ? <Metadata record={record}/>
                    : tab === "status" ? <Status record={record}/> : null}
            {requestOpened && <div className="manager-stage" style={tab === "request" ? undefined : {display: "none"}}><Request record={record}/></div>}
        </div>
    </div>;
};
