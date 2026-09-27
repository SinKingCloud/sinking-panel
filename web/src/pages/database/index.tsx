import {useEffect, useState} from "react";
import {Button, Empty, Spin, Tabs} from "antd";
import {Body} from "sinking-antd";
import {getSystemInfo} from "@/service/api/system";
import Redis from "./redis";
import Software from "./software";
import useStyles from "./styles";

export default () => {
    const {styles} = useStyles();
    const [tab, setTab] = useState("redis");
    const [platform, setPlatform] = useState<string>();
    const [error, setError] = useState("");
    const [refresh, setRefresh] = useState(0);

    useEffect(() => {
        let disposed = false;
        setError("");
        void getSystemInfo().then((response) => {
            if (disposed) return;
            if (response?.code !== 200 || !response.data?.os) {
                setError(response?.message || "获取服务器信息失败");
                return;
            }
            setPlatform(String(response.data.os));
        }).catch(() => { if (!disposed) setError("获取服务器信息失败"); });
        return () => { disposed = true; };
    }, [refresh]);

    return <Body>
        <main className={styles.page}>
            <div className={styles.heading}><h1>数据管理</h1></div>
            {error ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={error}>
                <Button variant="filled" color="default" onClick={() => setRefresh((value) => value + 1)}>重试</Button>
            </Empty> : !platform ? <div className={styles.loading}><Spin/></div>
                : platform !== "linux" ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="数据服务目前仅支持 Linux 服务器"/>
                    : <>
                        <Tabs activeKey={tab} onChange={setTab} items={[
                            {key: "redis", label: "Redis"},
                            {key: "mysql", label: "MySQL"},
                            {key: "pgsql", label: "PostgreSQL"},
                            {key: "mongo", label: "MongoDB"},
                        ]}/>
                        {tab === "redis" ? <Redis/> : <Software key={tab} kind={tab}/>}
                    </>}
        </main>
    </Body>;
};
