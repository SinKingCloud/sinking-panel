import {useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Empty, Input, Select, Table, Tooltip, Typography} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {useRedisClient} from "./context";
import useStyles from "./status-styles";

const groupNames: Record<string, string> = {
    Server: "服务信息", Clients: "客户端连接", Memory: "内存统计", Persistence: "数据持久化", Stats: "运行统计",
    Replication: "复制状态", CPU: "CPU 统计", Modules: "模块信息", Cluster: "集群状态", Keyspace: "数据库统计",
    Commandstats: "命令统计", Errorstats: "错误统计", Latencystats: "延迟统计", Sentinel: "哨兵状态",
};
const labels: Record<string, string> = {
    redis_version: "Redis 版本", redis_mode: "运行模式", role: "实例角色", tcp_port: "连接端口", os: "操作系统",
    arch_bits: "系统位数", process_id: "进程 ID", uptime_in_seconds: "运行时长", uptime_in_days: "运行天数", config_file: "配置文件",
    used_memory: "内存用量（字节）", used_memory_human: "内存用量", used_memory_rss: "物理内存（字节）", used_memory_rss_human: "物理内存",
    used_memory_peak: "内存峰值（字节）", used_memory_peak_human: "内存峰值", maxmemory: "内存上限（字节）", maxmemory_human: "内存上限",
    maxmemory_policy: "内存淘汰策略", mem_fragmentation_ratio: "内存碎片率", connected_clients: "客户端连接", blocked_clients: "阻塞连接",
    maxclients: "最大连接数", rejected_connections: "拒绝连接", total_connections_received: "累计连接", total_commands_processed: "累计命令",
    instantaneous_ops_per_sec: "每秒操作", total_net_input_bytes: "累计接收（字节）", total_net_output_bytes: "累计发送（字节）",
    instantaneous_input_kbps: "接收速率（KB/s）", instantaneous_output_kbps: "发送速率（KB/s）", expired_keys: "过期键数", evicted_keys: "淘汰键数",
    keyspace_hits: "命中次数", keyspace_misses: "未命中次数", rdb_last_bgsave_status: "最近快照", rdb_last_save_time: "快照时间",
    rdb_changes_since_last_save: "待持久化变更", rdb_bgsave_in_progress: "快照任务", aof_enabled: "AOF 日志", aof_rewrite_in_progress: "AOF 重写任务",
    aof_last_bgrewrite_status: "最近重写", aof_last_write_status: "最近 AOF 写入", connected_slaves: "从节点数量", master_host: "主节点地址",
    master_port: "主节点端口", master_link_status: "复制连接", master_last_io_seconds_ago: "距上次复制（秒）", cluster_enabled: "集群模式",
};

export default ({record, active = true}: any) => {
    const {api} = useRedisClient();
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const sequence = useRef(0);
    const tableContainer = useRef<HTMLDivElement>(null);
    const [info, setInfo] = useState<any>();
    const [loading, setLoading] = useState(active);
    const [failed, setFailed] = useState(false);
    const [refresh, setRefresh] = useState(0);
    const [rawOpen, setRawOpen] = useState(false);
    const [group, setGroup] = useState("all");
    const [keyword, setKeyword] = useState("");
    const [tableHeight, setTableHeight] = useState(260);

    useEffect(() => { setInfo(undefined); setGroup("all"); setKeyword(""); setRawOpen(false); }, [record.id, record.name]);
    useEffect(() => {
        const request = ++sequence.current;
        if (!active) { setLoading(false); return; }
        setLoading(true); setFailed(false);
        void api.info(record).then((data) => {
            if (request === sequence.current) setInfo(data);
        }).catch((reason) => {
            if (request === sequence.current) { setFailed(true); message.error(reason?.message || "读取实例信息失败"); }
        }).finally(() => { if (request === sequence.current) setLoading(false); });
        return () => { sequence.current++; };
    }, [record.id, record.name, active, refresh, api]);
    useEffect(() => {
        const node = tableContainer.current;
        if (!node) return;
        // 按剩余布局高度计算，不读取弹窗缩放动画中的视觉尺寸。
        const observer = new ResizeObserver(([entry]) => {
            if (entry.contentRect.height <= 0) return;
            const header = node.querySelector<HTMLElement>(".ant-table-header");
            setTableHeight(Math.max(80, Math.floor(entry.contentRect.height - (header?.offsetHeight || 40) - 2)));
        });
        observer.observe(node);
        return () => observer.disconnect();
    }, []);

    const raw = typeof info === "string" ? info : typeof info?.info === "string" ? info?.info : "";
    const rawContent = useMemo(() => raw || JSON.stringify(info, null, 2), [raw, info]);
    const {rows, metrics} = useMemo(() => {
        const entries: {key: string; group: string; title: string; name: string; label: string; value: string; display: string}[] = [];
        const values: Record<string, string> = {};
        let section = "Server";
        // 根据 INFO 的分组展开全部参数，新旧版本和模块新增字段均可显示。
        for (const line of raw.split(/\r?\n/)) {
            if (line.startsWith("#")) { section = line.slice(1).trim() || "Server"; continue; }
            const split = line.indexOf(":");
            if (split <= 0) continue;
            const name = line.slice(0, split), value = line.slice(split + 1);
            values[name] = value;
            entries.push({key: `${section}:${name}:${entries.length}`, group: section, title: groupNames[section] || section,
                name, label: labels[name] || name, value, display: value});
        }
        const format = (name: string, value?: string) => {
            if (value === undefined) return "—";
            if (value === "") return '""';
            if ((name === "maxmemory" || name === "maxmemory_human") && values.maxmemory === "0") return "不限制";
            if (name === "role") return ({master: "主节点", slave: "从节点", replica: "从节点"} as Record<string, string>)[value] || value;
            if (name === "redis_mode") return ({standalone: "单机", cluster: "集群", sentinel: "哨兵"} as Record<string, string>)[value] || value;
            if (name === "aof_enabled" || name === "cluster_enabled") return value === "1" ? "已开启" : value === "0" ? "未开启" : value;
            if (name.endsWith("_status")) return ({ok: "成功", err: "失败", up: "已连接", down: "已断开"} as Record<string, string>)[value] || value;
            if (name === "rdb_bgsave_in_progress" || name === "aof_rewrite_in_progress") return value === "1" ? "正在执行" : value === "0" ? "空闲" : value;
            if (name === "uptime_in_seconds") {
                const seconds = Number(value);
                if (!Number.isFinite(seconds) || seconds < 0) return value;
                if (seconds >= 86400) return `${Math.floor(seconds / 86400)} 天 ${Math.floor(seconds % 86400 / 3600)} 小时`;
                if (seconds >= 3600) return `${Math.floor(seconds / 3600)} 小时 ${Math.floor(seconds % 3600 / 60)} 分钟`;
                return seconds >= 60 ? `${Math.floor(seconds / 60)} 分钟` : `${seconds} 秒`;
            }
            if (name === "rdb_last_save_time") {
                const date = new Date(Number(value) * 1000);
                return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
            }
            if (/^\d+$/.test(value) && Number.isSafeInteger(Number(value))) return Number(value).toLocaleString();
            return value;
        };
        entries.forEach((item) => { item.display = format(item.name, item.value); });
        const databaseName = info?.database ?? record.name;
        const keyspace = values[`db${databaseName}`];
        const empty = keyspace === undefined && raw.includes("# Keyspace");
        const keys = keyspace?.match(/(?:^|,)keys=(\d+)/)?.[1];
        const expires = keyspace?.match(/(?:^|,)expires=(\d+)/)?.[1];
        const averageTTL = keyspace?.match(/(?:^|,)avg_ttl=(\d+)/)?.[1];
        const database = info === undefined ? [] : [
            ["database", "当前数据库", `DB ${databaseName}`], ["instance", "所属实例", record.server_name || "—"],
            ["expires", "过期键数", expires ?? (empty ? "0" : "—")],
            ["average_ttl", "平均有效期", averageTTL !== undefined ? `${Math.round(Number(averageTTL) / 1000)} 秒` : empty ? "0 秒" : "—"],
            ["id", "实例 ID", record.id], ["create_time", "添加时间", record.create_time],
            ["update_time", "更新时间", record.update_time], ["remark", "备注", record.remark],
        ].filter(([, , value]) => value !== undefined && value !== "").map(([name, label, value]) => ({
            key: `database:${name}`, group: "database", title: "当前数据库", name: String(name), label: String(label), value: String(value), display: String(value),
        }));
        return {rows: [...database, ...entries], metrics: [
            {label: "内存用量", value: format("used_memory_human", values.used_memory_human)},
            {label: "客户端连接", value: format("connected_clients", values.connected_clients)},
            {label: "每秒操作", value: format("instantaneous_ops_per_sec", values.instantaneous_ops_per_sec)},
            {label: "当前数据库键数", value: format("keys", keys ?? (empty ? "0" : undefined))},
        ]};
    }, [raw, info, record.id, record.name, record.server_name, record.create_time, record.update_time, record.remark]);
    const groups = useMemo(() => [{value: "all", label: "全部信息"},
        ...Array.from(new Map(rows.map((item) => [item.group, item.title])), ([value, label]) => ({value, label}))], [rows]);
    const currentGroup = groups.some((item) => item.value === group) ? group : "all";
    const filtered = useMemo(() => {
        const search = keyword.trim().toLowerCase();
        return rows.filter((item) => (currentGroup === "all" || item.group === currentGroup)
            && (!search || `${item.title} ${item.name} ${item.label} ${item.value} ${item.display}`.toLowerCase().includes(search)));
    }, [rows, currentGroup, keyword]);
    const columns = useMemo(() => [{title: "分类", dataIndex: "title", width: 120, ellipsis: true}, {
        title: "参数", key: "field", width: 280,
        render: (_: unknown, item: any) => <div className="redis-status-field"><span title={item.label}>{item.label}</span>
            {item.name !== item.label && <span className="redis-status-name" title={item.name}>{item.name}</span>}</div>,
    }, {
        title: "值", key: "value", width: 280,
        render: (_: unknown, item: any) => <Tooltip title={<span style={{whiteSpace: "pre-wrap", overflowWrap: "anywhere"}}>{item.value || '""'}</span>}
            styles={{body: {maxHeight: 320, overflowY: "auto"}}}><span className="redis-status-value">{item.display}</span></Tooltip>,
    }], []);

    return <div className={styles.status}>
        <div className="redis-status-toolbar redis-toolbar">
            <div className="redis-status-search">
                <Select variant="borderless" styles={{root: {outline: "none", boxShadow: "none", border: 0}}}
                    aria-label="信息分类" placeholder="选择分类" value={currentGroup} options={groups} onChange={setGroup}
                    popupMatchSelectWidth={false} labelRender={({label}) => <span className="redis-status-search-label">{label}</span>}/>
                <Input variant="borderless" styles={{root: {outline: "none", boxShadow: "none"}, input: {outline: "none", boxShadow: "none"}}}
                    aria-label="搜索实例信息" placeholder="搜索参数或值" allowClear value={keyword} onChange={(event) => setKeyword(event.target.value)}/>
            </div>
            <div className="redis-status-tools">
                <Button color="default" variant="filled" icon={<Icon type="FileTextOutlined"/>} disabled={info === undefined} onClick={() => setRawOpen(true)}>详细信息</Button>
                <Button className="redis-status-refresh" color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} loading={loading} disabled={!active}
                    aria-label="刷新实例信息" onClick={() => setRefresh((value) => value + 1)}/>
            </div>
        </div>
        {failed && info !== undefined && <div className="redis-status-note" role="status">本次刷新失败，当前显示上次读取的实例信息。</div>}
        {info !== undefined && <div className="redis-status-summary">{metrics.map((item) => <div key={item.label}><small>{item.label}</small><strong>{item.value}</strong></div>)}</div>}
        <div className="redis-status-table" ref={tableContainer}>
            <Table aria-label="Redis 实例信息" rowKey="key" size="small" columns={columns} dataSource={filtered} loading={loading} pagination={false} virtual tableLayout="fixed"
                rowClassName={(_, index) => index === filtered.length - 1 ? "client-last-row" : ""}
                locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
                scroll={{x: 680, y: tableHeight, scrollToFirstRowOnChange: false}}/>
        </div>
        <ProModal title={<Title>详细信息</Title>} onCancel={() => setRawOpen(false)} onOk={() => setRawOpen(false)} modalProps={{
            open: rawOpen, width: compact ? 480 : 520, rootClassName: styles.detail,
            styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}}, style: {maxWidth: "calc(100vw - 24px)"},
            zIndex: 1140, destroyOnHidden: true, mask: {closable: true}, okText: "确认", cancelButtonProps: {style: {display: "none"}},
        }}>
            <div className="redis-status-copy"><Typography.Text copyable={{text: rawContent, icon: ["复制", "已复制"]}}/></div>
            <pre>{rawContent}</pre>
        </ProModal>
    </div>;
};
