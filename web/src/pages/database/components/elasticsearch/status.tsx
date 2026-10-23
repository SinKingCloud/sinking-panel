import {useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Input, Select, Tooltip} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {useDatabaseClient} from "../shared/context";
import useStyles from "./styles";
import {VirtualTable} from "../shared/client-common";
import statusData from "./status-data";

export default ({record}: any) => {
    const {operate} = useDatabaseClient();
    const {message} = App.useApp();
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const [data, setData] = useState<any>();
    const [info, setInfo] = useState<any>();
    const [loading, setLoading] = useState(true);
    const [refresh, setRefresh] = useState(0);
    const [raw, setRaw] = useState(false);
    const [group, setGroup] = useState("all");
    const [keyword, setKeyword] = useState("");
    const [refreshFailed, setRefreshFailed] = useState(false);
    const [infoFailed, setInfoFailed] = useState(false);
    const sequence = useRef(0);
    const previousRecord = useRef({id: record.id, name: record.name});
    useEffect(() => {
        const request = ++sequence.current;
        if (previousRecord.current.id !== record.id) { setData(undefined); setGroup("all"); setKeyword(""); }
        if (previousRecord.current.id !== record.id || previousRecord.current.name !== record.name) setInfo(undefined);
        previousRecord.current = {id: record.id, name: record.name};
        setLoading(true);
        setRefreshFailed(false);
        setInfoFailed(false);
        // 索引详情不可用时仍保留集群状态，刷新期间沿用已有数据。
        void Promise.allSettled([operate(record, "status"), record.name ? operate(record, "info") : Promise.resolve(undefined)]).then(([statusResult, infoResult]) => {
            if (request !== sequence.current) return;
            if (statusResult.status === "fulfilled") setData(statusResult.value);
            else setRefreshFailed(true);
            if (infoResult.status === "fulfilled") setInfo(infoResult.value);
            else { setInfo(undefined); setInfoFailed(true); }
            if (statusResult.status === "rejected" || infoResult.status === "rejected") {
                const reason = statusResult.status === "rejected" ? statusResult.reason : infoResult.status === "rejected" ? infoResult.reason : undefined;
                message.error(reason?.message || (statusResult.status === "rejected" ? "读取集群状态失败" : "读取索引详情失败"));
            }
            setLoading(false);
        });
        return () => { sequence.current++; };
    }, [record.id, record.name, operate, refresh]);

    const health = data?.health || {};
    const stats = data?.stats || {};
    // 根据实际响应展开，旧版本和新增字段无需再维护展示白名单。
    const rows = useMemo(() => statusData(data, info), [data, info]);
    const rawContent = useMemo(() => JSON.stringify({...data, ...(info ? {index: info} : {})}, null, 2), [data, info]);
    const groups = useMemo(() => [{value: "all", label: "全部信息"},
        ...Array.from(new Map<string, string>(rows.map((item) => [item.group, item.title])), ([value, label]) => ({value, label}))], [rows]);
    const currentGroup = groups.some((item) => item.value === group) ? group : "all";
    const filtered = useMemo(() => {
        const search = keyword.trim().toLowerCase();
        return rows.filter((item) => (currentGroup === "all" || item.group === currentGroup)
            && (!search || `${item.title} ${item.path} ${item.label} ${typeof item.value === "object" ? JSON.stringify(item.value) : item.value}`.toLowerCase().includes(search)));
    }, [rows, currentGroup, keyword]);
    const bytes = (value: unknown) => {
        const size = typeof value === "number" ? value : typeof value === "string" && value.trim() ? Number(value) : NaN;
        if (!Number.isFinite(size) || size < 0) return "—";
        const unit = size === 0 ? 0 : Math.max(0, Math.min(5, Math.floor(Math.log(size) / Math.log(1024))));
        return `${(size / 1024 ** unit).toLocaleString(undefined, {maximumFractionDigits: unit ? 2 : 0})} ${["B", "KB", "MB", "GB", "TB", "PB"][unit]}`;
    };
    const columns = useMemo(() => [{
        title: "分类", dataIndex: "title", width: 120, ellipsis: true,
    }, {
        title: "参数", key: "field", width: 280,
        render: (_: unknown, item: any) => <div className="es-status-field">
            <span title={item.label}>{item.label}</span>{item.path !== item.label && <span className="es-status-path" title={item.path}>{item.path}</span>}
        </div>,
    }, {
        title: "值", key: "value", width: 280,
        render: (_: unknown, item: any) => {
            const value = typeof item.value === "string" ? item.value === "" ? '""' : item.value : JSON.stringify(item.value);
            const display = typeof item.value === "number" && item.value >= 0 && /(?:_in_bytes|_bytes)$/.test(item.name) ? bytes(item.value)
                : typeof item.value === "number" && /_millis$/.test(item.name) ? `${value} ms` : value;
            // 保留原始值，长文本可悬浮查看；配置数值不做隐式转换。
            return <Tooltip title={<span style={{whiteSpace: "pre-wrap", overflowWrap: "anywhere"}}>{JSON.stringify(item.value)}</span>}
                styles={{body: {maxHeight: 320, overflowY: "auto"}}}>
                <span className="manager-cell es-status-value">{display}</span>
            </Tooltip>;
        },
    }], []);

    return <div className="manager-stage">
        <div className="manager-toolbar es-status-toolbar">
            <div className="es-search">
                <Select variant="borderless" styles={{root: {outline: "none", boxShadow: "none", border: 0}}}
                    aria-label="信息分类" placeholder="选择分类" value={currentGroup} options={groups} onChange={setGroup}
                    popupMatchSelectWidth={false} labelRender={({label}) => <span className="es-search-label">{label}</span>}/>
                <Input variant="borderless" styles={{root: {outline: "none", boxShadow: "none"}, input: {outline: "none", boxShadow: "none"}}}
                    aria-label="搜索统计和配置" placeholder="搜索参数或值" allowClear value={keyword} onChange={(event) => setKeyword(event.target.value)}/>
            </div>
            <div className="manager-tools">
                <Button color="default" variant="filled" icon={<Icon type="FileTextOutlined"/>} disabled={!data && !info} onClick={() => setRaw(true)}>详细信息</Button>
                <Button className="manager-refresh" color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} aria-label="刷新集群状态" loading={loading} onClick={() => setRefresh((value) => value + 1)}/>
            </div>
        </div>
        {refreshFailed && data && <div className="es-status-note" role="status">本次刷新失败，当前显示上次读取的集群状态。</div>}
        {(data?.stats_error || infoFailed || info?.stats_error) && <div className="es-status-note" role="status">
            {[data?.stats_error && "集群统计暂不可用", infoFailed ? "索引详情暂不可用" : info?.stats_error && "索引统计暂不可用"].filter(Boolean).join("；")}，其余信息仍可查看。
        </div>}
        {data && <div className="es-summary">{[["索引数量", stats.indices?.count], ["文档数量", stats.indices?.docs?.count],
            ["索引存储含副本", bytes(stats.indices?.store?.size_in_bytes)], ["活动分片", health.active_shards]].map(([label, value]) =>
            <div key={label}><small>{label}</small><strong>{typeof value === "number" ? value.toLocaleString() : value ?? "—"}</strong></div>)}</div>}
        <div className="manager-table es-status-table">
            <VirtualTable rowKey="key" size="small" columns={columns} dataSource={filtered} loading={loading} pagination={false}
                scroll={{x: 680, scrollToFirstRowOnChange: false}}/>
        </div>
        <ProModal title={<Title>详细信息</Title>} onCancel={() => setRaw(false)} onOk={() => setRaw(false)} modalProps={{
            open: raw, width: theme?.isCompactTheme?.() ? 480 : 520, rootClassName: styles.modal,
            styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}}, style: {maxWidth: "calc(100vw - 24px)"},
            zIndex: 1140, destroyOnHidden: true, mask: {closable: true}, okText: "确认", cancelButtonProps: {style: {display: "none"}},
        }}>
            <pre className="es-code">{rawContent}</pre>
        </ProModal>
    </div>;
};
