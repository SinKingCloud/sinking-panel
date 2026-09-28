import {useCallback, useEffect, useRef, useState} from "react";
import {Alert, App, Button, Card, Checkbox, Col, Descriptions, Empty, Form, Input, InputNumber, Row, Select, Space, Spin, Switch, Tabs, Tag, Typography} from "antd";
import {Icon} from "sinking-antd";
import ContainerTerminal from "@/pages/components/container-terminal";
import Log from "@/pages/components/task-log";
import type {LogRef} from "@/pages/components/task-log";

import {getContainerLog} from "@/service/api/container";
import {
    controlDatabase, databaseConfig, databasePerformance, getDatabaseStatus, installDatabase, parseDatabaseImage,
} from "@/service/api/database";
import useEnum from "@/utils/enum";
import {formatSize} from "@/utils/string";
import Records from "./records";
import useStyles from "./styles";

const software: Record<string, any> = {
    mysql: {name: "MySQL", port: 3306, username: "root"},
    pgsql: {name: "PostgreSQL", port: 5432, username: "postgres"},
    mongo: {name: "MongoDB", port: 27017, username: "admin"},
};
// 表单独立加载，状态轮询不会覆盖正在编辑的配置。
const Settings = ({kind, mode, onChange, busy, onBusy}: {
    kind: string; mode: "performance" | "config"; onChange: () => void; busy: boolean; onBusy: (value: boolean) => void;
}) => {
    const {styles} = useStyles();
    const {message} = App.useApp();
    const [form] = Form.useForm();
    const [fields, setFields] = useState<any[]>([]);
    const [content, setContent] = useState("");
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState("");
    const [refresh, setRefresh] = useState(0);
    const mounted = useRef(true);
    const pending = useRef(false);

    useEffect(() => {
        mounted.current = true;
        return () => { mounted.current = false; };
    }, []);

    useEffect(() => {
        let disposed = false;
        setLoading(true);
        setError("");
        const load = async () => {
            try {
                if (mode === "performance") {
                    const response = await databasePerformance(kind, {body: {action: "get"}});
                    if (disposed) return;
                    if (response?.code !== 200 || !response.data) throw new Error(response?.message || "读取配置失败");
                    setFields(response.data.fields || []);
                    form.resetFields();
                    form.setFieldsValue(response.data.values || {});
                } else {
                    const response = await databaseConfig(kind, {body: {action: "get"}});
                    if (disposed) return;
                    if (response?.code !== 200) throw new Error(response?.message || "读取配置失败");
                    setContent(response.data?.content || "");
                }
            } catch (reason) {
                if (!disposed) setError(reason instanceof Error ? reason.message : "读取配置失败");
            } finally { if (!disposed) setLoading(false); }
        };
        void load();
        return () => { disposed = true; };
    }, [kind, mode, refresh, form]);

    const save = async (values?: Record<string, any>) => {
        if (busy || pending.current) return;
        pending.current = true;
        setSaving(true);
        onBusy(true);
        try {
            const response = mode === "performance"
                ? await databasePerformance(kind, {body: {action: "set", config: values || {}}})
                : await databaseConfig(kind, {body: {action: "set", content}});
            if (!mounted.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "保存配置失败");
            message.success(response.data?.restart_required ? "配置已保存，重启后生效" : "配置已保存");
            onChange();
        } catch (reason) {
            if (mounted.current) message.error(reason instanceof Error ? reason.message : "保存配置失败");
        } finally {
            pending.current = false;
            onBusy(false);
            if (mounted.current) setSaving(false);
        }
    };

    if (loading) return <div className={styles.loading}><Spin/></div>;
    if (error) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={error}>
        <Button variant="filled" color="default" onClick={() => setRefresh((value) => value + 1)}>重试</Button>
    </Empty>;

    return mode === "performance" ? <Form form={form} className={styles.form} layout="vertical" variant="filled" onFinish={save} disabled={busy}>
        <Row gutter={[24, 0]}>
            {fields.map((field) => <Col xs={24} sm={12} xl={8} key={field.name}>
                <Form.Item name={field.name} label={`${field.label}${field.unit ? `（${field.unit}）` : ""}`}
                    tooltip={field.tooltip || undefined} valuePropName={field.type === "switch" ? "checked" : "value"}
                    rules={field.type === "switch" ? undefined : [{required: true, message: `请输入${field.label}`}]}>
                    {field.type === "number" ? <InputNumber min={field.min} max={field.max} step={field.step} style={{width: "100%"}}/>
                        : field.type === "select" ? <Select options={field.options}/>
                            : field.type === "switch" ? <Switch/> : <Input/>}
                </Form.Item>
            </Col>)}
        </Row>
        <div className="form-actions"><Button type="primary" htmlType="submit" loading={saving}>保存配置</Button></div>
    </Form> : <div className={styles.config}>
        <Input.TextArea variant="filled" value={content} disabled={busy} onChange={(event) => setContent(event.target.value)} spellCheck={false}/>
        <div className={styles.configActions}>
            <Button color="default" variant="filled" disabled={busy} onClick={() => setRefresh((value) => value + 1)}>重新读取</Button>
            <Button type="primary" loading={saving} disabled={busy && !saving} onClick={() => void save()}>保存配置</Button>
        </div>
    </div>;
};

const Load = ({status}: {status: any}) => {
    const {styles} = useStyles();
    const previous = useRef<any>(undefined);
    const [cpu, setCpu] = useState<number>();
    const resource = status.resources || status.container;
    useEffect(() => {
        const before = previous.current;
        const elapsed = resource && before ? Date.parse(resource.read_time) - Date.parse(before.read_time) : 0;
        setCpu(resource?.resources_available && before?.resources_available && elapsed > 0
            && resource.generation === before.generation && resource.started_at === before.started_at
            && resource.cpu_usage >= before.cpu_usage
            ? (resource.cpu_usage - before.cpu_usage) / (elapsed * 1000000) * 100 : undefined);
        previous.current = resource;
    }, [resource]);
    const available = resource?.resources_available;
    return <>
        {status.metrics_error && <Alert type="warning" showIcon message={status.metrics_error} style={{marginBottom: 16}}/>}
        <div className={styles.metrics}>
            {[
                ["CPU 使用率", cpu === undefined ? "—" : `${cpu.toFixed(1)}%`],
                ["内存使用", available ? formatSize(resource.memory_usage || 0) : "—"],
                ["进程数量", available ? resource.pids : "—"],
                ["运行时间", status.ready && resource?.uptime ? `${Math.floor(resource.uptime / 3600)} 小时 ${Math.floor(resource.uptime % 3600 / 60)} 分` : "—"],
            ].map(([label, value]) => <div key={String(label)}><span>{label}</span><strong>{value}</strong></div>)}
        </div>
        {available && <Descriptions size="small" column={{xs: 1, sm: 2, lg: 3}} items={[
            {key: "read", label: "累计读取", children: formatSize(resource.io_read_bytes || 0)},
            {key: "write", label: "累计写入", children: formatSize(resource.io_write_bytes || 0)},
            {key: "limit", label: "内存限制", children: resource.memory_limit ? formatSize(resource.memory_limit) : "不限制"},
        ]}/>}
        {status.metrics && <Descriptions size="small" column={{xs: 1, sm: 2, lg: 3}} style={{marginTop: 16}} items={Object.entries({
            connections: "当前连接", max_connections: "最大连接", active_connections: "活动连接", queries: "累计查询",
            slow_queries: "慢查询", threads_running: "运行线程", database_count: "数据库数量",
            transactions_committed: "已提交事务", transactions_rolled_back: "已回滚事务", cache_hits: "缓存命中", disk_reads: "磁盘读取",
            buffer_pool_bytes: "缓冲池大小", buffer_pool_read_requests: "缓冲池读取请求", buffer_pool_reads: "缓冲池磁盘读取",
            resident_memory_mb: "驻留内存（MB）", virtual_memory_mb: "虚拟内存（MB）", operations_insert: "插入操作",
            operations_query: "查询操作", operations_update: "更新操作", operations_delete: "删除操作", bytes_in: "累计接收", bytes_out: "累计发送",
        }).filter(([key]) => status.metrics?.[key] !== undefined).map(([key, label]) => ({
            key, label, children: ["buffer_pool_bytes", "bytes_in", "bytes_out"].includes(key)
                ? formatSize(Number(status.metrics?.[key]) || 0) : String(status.metrics?.[key] ?? "—"),
        }))}/>}
    </>;
};

export default ({kind}: {kind: string}) => {
    const product = software[kind];
    const {styles} = useStyles();
    const {message, modal} = App.useApp();
    const [enumData] = useEnum("container");
    const [form] = Form.useForm();
    const [status, setStatus] = useState<any>();
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [image, setImage] = useState<any>();
    const [parsing, setParsing] = useState(false);
    const [action, setAction] = useState("");
    const [editing, setEditing] = useState(false);
    const [tab, setTab] = useState("databases");
    const [terminal, setTerminal] = useState(false);
    const mounted = useRef(true);
    const actionRef = useRef("");
    const statusRequest = useRef(0);
    const statusPending = useRef(0);
    const parseRequest = useRef(0);
    const logRef = useRef<LogRef>(null);
    const confirmation = useRef<{destroy: () => void} | null>(null);
    const installed = Boolean(status?.installed);
    const retainedImage = status?.data_preserved ? status.image_id : undefined;
    const running = status?.status === "running";
    const busy = Boolean(action) || editing;
    const transitioning = status?.status === "starting" || status?.status === "stopping";
    const locked = useRef(false);
    locked.current = busy;

    const loadStatus = useCallback(async (manual = false) => {
        if (!mounted.current || (!manual && (statusPending.current || actionRef.current))) return;
        const request = ++statusRequest.current;
        statusPending.current = request;
        if (manual) setRefreshing(true);
        try {
            const response = await getDatabaseStatus(kind);
            if (!mounted.current || statusRequest.current !== request) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "获取服务状态失败");
            setStatus(response.data);
            setError("");
        } catch (reason) {
            if (mounted.current && statusRequest.current === request) setError(reason instanceof Error ? reason.message : "获取服务状态失败");
        } finally {
            if (statusPending.current === request) statusPending.current = 0;
            if (mounted.current && statusRequest.current === request) { setLoading(false); setRefreshing(false); }
        }
    }, [kind]);

    useEffect(() => {
        mounted.current = true;
        void loadStatus();
        const timer = window.setInterval(() => {
            if (document.visibilityState === "visible") void loadStatus();
        }, 5000);
        return () => {
            mounted.current = false;
            statusRequest.current += 1;
            statusPending.current = 0;
            parseRequest.current += 1;
            window.clearInterval(timer);
            confirmation.current?.destroy();
        };
    }, [loadStatus]);

    const run = async (name: string, request: () => Promise<any>, success: string) => {
        if (actionRef.current || locked.current || !mounted.current) return false;
        actionRef.current = name;
        statusRequest.current += 1;
        setAction(name);
        try {
            const response = await request();
            if (!mounted.current) return false;
            if (response?.code !== 200) throw new Error(response?.message || "操作失败");
            message.success(response.data?.restart_required ? "配置已保存，重启后生效" : success);
            return true;
        } catch (reason) {
            if (mounted.current) message.error(reason instanceof Error ? reason.message : "操作失败");
            return false;
        } finally {
            actionRef.current = "";
            if (mounted.current) { setAction(""); void loadStatus(true); }
        }
    };

    const parse = async () => {
        if (parsing || busy) return;
        let values;
        try { values = await form.validateFields(["image_path"]); } catch { return; }
        const path = values.image_path.trim();
        const request = ++parseRequest.current;
        setImage(undefined);
        setParsing(true);
        try {
            const response = await parseDatabaseImage(kind, {body: {image_path: path}});
            if (!mounted.current || request !== parseRequest.current) return;
            if (response?.code !== 200 || !response.data?.id) throw new Error(response?.message || "解析镜像失败");
            setImage({...response.data, path});
        } catch (reason) {
            if (mounted.current && request === parseRequest.current) message.error(reason instanceof Error ? reason.message : "解析镜像失败");
        } finally { if (mounted.current && request === parseRequest.current) setParsing(false); }
    };

    const install = async (values: Record<string, any>) => {
        if (retainedImage) {
            await run("install", () => installDatabase(kind, {body: {...values, image_id: retainedImage}}), "安装成功");
            return;
        }
        if (!image || image.path !== values.image_path.trim() || parsing) { message.warning("请先解析当前镜像"); return; }
        await run("install", () => installDatabase(kind, {body: {...values, image_id: image.id, image_path: image.path}}), "安装成功");
    };

    const uninstall = () => {
        let removeData = false;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({
            title: `卸载 ${product.name}`,
            content: <><Typography.Paragraph>停止并删除容器，镜像保留。默认保留数据，重新安装后可以继续使用。</Typography.Paragraph>
                <Checkbox onChange={(event) => { removeData = event.target.checked; }}>同时删除数据库和配置文件</Checkbox></>,
            okText: "卸载", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                const success = await run("uninstall", () => controlDatabase(kind, "uninstall", {body: {remove_data: removeData}}), "卸载成功");
                if (!success) throw new Error("卸载失败");
                if (mounted.current) { setTerminal(false); setImage(undefined); form.resetFields(); }
            },
        });
    };

    if (loading) return <div className={styles.loading}><Spin/></div>;
    if (!status) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={error || "获取服务状态失败"}>
        <Button color="default" variant="filled" loading={refreshing} onClick={() => void loadStatus(true)}>重试</Button>
    </Empty>;

    return <>
        {!installed ? <Card title={`安装 ${product.name}`} className={styles.install}>
            <Form form={form} layout="vertical" variant="filled" className={styles.form} onFinish={install}
                initialValues={{bind: status.bind || "127.0.0.1", port: status.port || product.port, username: status.username || product.username}} disabled={busy}
                onValuesChange={(changed) => {
                    if ("image_path" in changed) { parseRequest.current += 1; setImage(undefined); setParsing(false); }
                }}>
                {retainedImage ? <Alert showIcon type="info" message="检测到已保留的数据"
                    description={`将使用原镜像${status.version ? ` ${status.version}` : ""}重新安装，并保留原有数据库和账户。`} style={{marginBottom: 20}}/>
                    : <Form.Item label="服务器镜像路径">
                    <div className={styles.imageInput}>
                        <Form.Item noStyle name="image_path" rules={[{required: true, whitespace: true, message: "请输入服务器本地镜像路径"}]}>
                            <Input placeholder={`/opt/images/${kind === "pgsql" ? "postgres" : kind}.tar`}/>
                        </Form.Item>
                        <Button color="default" variant="filled" loading={parsing} onClick={() => void parse()}>解析</Button>
                    </div>
                </Form.Item>}
                {image && <div className={styles.imagePreview}>
                    <dl><dt>镜像</dt><dd>{image.name}{image.version && ` · ${image.version}`}</dd>
                        <dt>平台</dt><dd>{image.os} / {image.architecture}</dd>
                        <dt>持久化目录</dt><dd>{image.data_path || "—"}</dd></dl>
                </div>}
                <Row gutter={16}>
                    <Col xs={24} sm={12}><Form.Item name="username" label="管理员账户" rules={[{required: true, whitespace: true, message: "请输入管理员账户"}]}>
                        <Input autoComplete="off" disabled={kind === "mysql" || Boolean(retainedImage) || busy}/>
                    </Form.Item></Col>
                    <Col xs={24} sm={12}><Form.Item name="password" label="管理员密码" rules={[{required: !retainedImage, message: "请输入管理员密码"}]}>
                        <Input.Password autoComplete="new-password" placeholder={retainedImage ? "留空使用原密码" : undefined}/>
                    </Form.Item></Col>
                    <Col xs={24} sm={12}><Form.Item name="bind" label="绑定地址" rules={[{required: true, whitespace: true, message: "请输入绑定地址"}]}>
                        <Input placeholder="127.0.0.1"/>
                    </Form.Item></Col>
                    <Col xs={24} sm={12}><Form.Item name="port" label="端口" rules={[{required: true, message: "请输入端口"}]}>
                        <InputNumber min={1} max={65535} precision={0} style={{width: "100%"}}/>
                    </Form.Item></Col>
                </Row>
                <div className="form-actions"><Button type="primary" htmlType="submit" loading={action === "install"} disabled={(!image && !retainedImage) || parsing}>安装</Button></div>
            </Form>
        </Card> : <Card className={styles.panel}>
            <div className={styles.serviceToolbar}>
                <Space wrap><Typography.Text strong>{product.name}{status.version ? ` ${status.version}` : ""}</Typography.Text>
                    <Tag color={running ? "success" : status.status === "failed" ? "error" : "default"}>{enumData?.status?.[status.status || ""] || status.status || "未知状态"}</Tag>
                    {running && !status.ready && <Typography.Text type="secondary">初始化中</Typography.Text>}
                </Space>
                <Space wrap>
                    <Button color="default" variant="filled" loading={action === "start"} disabled={running || busy || transitioning} onClick={() => void run("start", () => controlDatabase(kind, "start", {body: {}}), "启动成功")}>启动</Button>
                    <Button color="default" variant="filled" loading={action === "stop"} disabled={!running || busy || transitioning} onClick={() => void run("stop", () => controlDatabase(kind, "stop", {body: {}}), "停止成功")}>停止</Button>
                    <Button color="default" variant="filled" loading={action === "restart"} disabled={busy || transitioning} onClick={() => void run("restart", () => controlDatabase(kind, "restart", {body: {}}), "重启成功")}>重启</Button>
                    <Button color="default" variant="filled" disabled={!running || busy} onClick={() => setTerminal(true)}>终端</Button>
                    <Button color="default" variant="filled" onClick={() => logRef.current?.open({id: status.id || kind})}>日志</Button>
                    <Button color="danger" variant="filled" loading={action === "uninstall"} disabled={busy} onClick={uninstall}>卸载</Button>
                    <Button color="default" variant="filled" aria-label="刷新服务状态" icon={<Icon type="ReloadOutlined"/>} loading={refreshing} onClick={() => void loadStatus(true)}/>
                </Space>
            </div>
            <Descriptions size="small" column={{xs: 1, sm: 2, lg: 4}} className={styles.connection} items={[
                {key: "bind", label: "绑定地址", children: status.bind || "—"},
                {key: "port", label: "端口", children: status.port || "—"},
                {key: "username", label: "管理员", children: status.username || product.username},
                {key: "pid", label: "进程 PID", children: status.pid || "—"},
            ]}/>
            {(error || status.error) && <Alert showIcon type="warning" message={error || status.error} style={{marginBottom: 16}}/>}
            {(status.restart_required || status.config_pending) && <Alert showIcon type="info" message="配置已修改，重启服务后生效" style={{marginBottom: 16}}/>}
            <Tabs activeKey={tab} onChange={setTab} items={[
                {key: "databases", label: "数据库", disabled: busy}, {key: "users", label: "账户", disabled: busy},
                {key: "performance", label: "性能配置", disabled: busy}, {key: "load", label: "负载状态", disabled: busy},
                {key: "config", label: "配置文件", disabled: busy},
            ]} tabBarExtraContent={kind === "pgsql" && (tab === "config" || tab === "performance")
                ? <Button color="default" variant="filled" disabled={!running || busy} loading={action === "reload"}
                    onClick={() => void run("reload", () => controlDatabase(kind, "reload", {body: {}}), "配置已重载")}>重载配置</Button> : undefined}/>
            {(tab === "performance" || tab === "config") ? <Settings key={tab} kind={kind} mode={tab} busy={busy}
                onBusy={(value) => { if (mounted.current) setEditing(value); }} onChange={() => void loadStatus(true)}/>
                : tab === "load" ? <Load status={status}/>
                    : <Records key={tab} kind={kind} mode={tab as "databases" | "users"} ready={Boolean(status.ready)} administrator={status.username || product.username}
                        onBusy={(value) => { if (mounted.current) setEditing(value); }}/>} 
        </Card>}
        <Log ref={logRef} request={getContainerLog} body={{action: "read"}} title={`${product.name} 日志`}/>
        <ContainerTerminal id={status.id || kind} open={terminal && installed && running} onClose={() => setTerminal(false)}/>
    </>;
};
