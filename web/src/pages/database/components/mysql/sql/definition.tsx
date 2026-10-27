import {useEffect, useRef, useState} from "react";
import {createPortal} from "react-dom";
import {App, Button, Space, Spin, Typography} from "antd";
import {Icon} from "sinking-antd";
import {copyTextToClipboard} from "../clipboard";
import {useMysqlClient} from "../context";

export default ({record, schema, table, version, toolbarContainer}: any) => {
    const {api} = useMysqlClient();
    const {message} = App.useApp();
    const [value, setValue] = useState("");
    const [loading, setLoading] = useState(false);
    const request = useRef(0);
    useEffect(() => {
        const id = ++request.current;
        setValue(""); setLoading(true);
        void api.ddl(record, {schema, table}).then((data) => {
            if (request.current === id) setValue(typeof data === "string" ? data : "");
        }).catch((reason) => {
            if (request.current !== id) return;
            message.error({key: `mysql-definition-${record.id}`, content: reason?.message});
        })
            .finally(() => { if (request.current === id) setLoading(false); });
        return () => { request.current++; };
    }, [record.id, schema, table, version]);
    const actions = <>
        <Button color="default" variant="filled" size="small" autoInsertSpace={false} icon={<Icon type="CopyOutlined"/>}
            disabled={!value || loading} onClick={async () => {
                if (await copyTextToClipboard(value)) message.success("复制成功");
                else message.error("复制失败，请手动复制");
            }}>复制</Button>
    </>;
    return <div className="sql-definition">
        {toolbarContainer === undefined ? <div className="sql-definition-toolbar"><Typography.Text type="secondary">SQL 定义</Typography.Text>
            <Space size={10}>{actions}</Space></div> : toolbarContainer && createPortal(actions, toolbarContainer)}
        <Spin spinning={loading}><pre className="sql-definition-text">{value}</pre></Spin>
    </div>;
};
