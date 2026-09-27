import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {
    Alert,
    App,
    Button,
    Card,
    Col,
    Collapse,
    Empty,
    Form,
    Input,
    InputNumber,
    Row,
    Select,
    Space,
    Spin,
    Switch,
    Tabs,
    Tag,
    Typography,
} from "antd";
import {Icon} from "sinking-antd";
import {
    clearRedisLog,
    getRedisConfig,
    getRedisLog,
    getRedisPersistence,
    getRedisPerformance,
    getRedisStatus,
    installRedis,
    parseRedisImage,
    reloadRedis,
    restartRedis,
    startRedis,
    stopRedis,
    uninstallRedis,
    updateRedisConfig,
    updateRedisPersistence,
    updateRedisPerformance,
} from "@/service/api/redis";
import ContainerTerminal from "@/pages/components/container-terminal";
import useEnum from "@/utils/enum";
import useStyles from "./styles";

const defaultPerformance = {
    bind: "127.0.0.1",
    port: 6379,
    timeout: 0,
    maxclients: 10000,
    databases: 16,
    requirepass: "",
    maxmemory: 0,
};

const defaultPersistence = {
    save: [],
    appendonly: false,
    appendfsync: "everysec",
    auto_aof_rewrite_percentage: 100,
    auto_aof_rewrite_min_size: 64,
};

const responseError = (response: any, fallback: string) => response?.message || fallback;

const displayValue = (value: any, fallback = "-") => {
    if (value === undefined || value === null || value === "") {
        return fallback;
    }
    return String(value);
};

export default (): React.ReactNode => {
    const {styles} = useStyles();
    const {message, modal} = App.useApp();
    const [enumData] = useEnum("container");
    const mountedRef = useRef(true);
    const actionRef = useRef("");
    const imageRequestRef = useRef(0);
    const parsingPathRef = useRef("");
    const confirmRef = useRef<{destroy: () => void} | null>(null);
    const [status, setStatus] = useState<any>();
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [action, setAction] = useState("");
    const [image, setImage] = useState<any>();
    const [imageParsing, setImageParsing] = useState(false);
    const [config, setConfig] = useState("");
    const [configLoaded, setConfigLoaded] = useState(false);
    const [configLoading, setConfigLoading] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);
    const [logLoading, setLogLoading] = useState(false);
    const [logCursor, setLogCursor] = useState(0);
    const [terminalOpen, setTerminalOpen] = useState(false);
    const [installForm] = Form.useForm();
    const [performanceForm] = Form.useForm();
    const [persistenceForm] = Form.useForm();
    const installed = Boolean(status?.installed);

    const loadSettings = useCallback(async () => {
        const [performanceResponse, persistenceResponse] = await Promise.all([
            getRedisPerformance(),
            getRedisPersistence(),
        ]);
        if (!mountedRef.current) {
            return;
        }
        if (performanceResponse?.code === 200) {
            performanceForm.setFieldsValue({...defaultPerformance, ...(performanceResponse.data || {})});
        }
        if (persistenceResponse?.code === 200) {
            persistenceForm.setFieldsValue({...defaultPersistence, ...(persistenceResponse.data || {})});
        }
    }, [performanceForm, persistenceForm]);

    const loadStatus = useCallback(async (manual = false) => {
        if (manual) {
            setRefreshing(true);
        } else {
            setLoading(true);
        }
        try {
            const response = await getRedisStatus();
            if (!mountedRef.current) {
                return;
            }
            if (response?.code !== 200) {
                message.error(responseError(response, "获取 Redis 状态失败"));
                return;
            }
            const nextStatus = response.data || {};
            setStatus(nextStatus);
            if (nextStatus.installed) {
                await loadSettings();
            }
        } catch {
            if (mountedRef.current) {
                message.error("获取 Redis 状态失败");
            }
        } finally {
            if (mountedRef.current) {
                setLoading(false);
                setRefreshing(false);
            }
        }
    }, [loadSettings, message]);

    useEffect(() => {
        mountedRef.current = true;
        void loadStatus();
        return () => {
            mountedRef.current = false;
            imageRequestRef.current += 1;
            parsingPathRef.current = "";
            confirmRef.current?.destroy();
        };
    }, [loadStatus]);

    const runAction = useCallback(async (name: string, request: () => Promise<any>, fallback: string) => {
        if (actionRef.current) {
            return;
        }
        actionRef.current = name;
        setAction(name);
        try {
            const response = await request();
            if (!mountedRef.current) return;
            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(responseError(response, fallback));
                return;
            }
            message.success(response.message || fallback);
            await loadStatus(true);
        } catch {
            if (mountedRef.current) message.error(fallback);
        } finally {
            actionRef.current = "";
            if (mountedRef.current) {
                setAction("");
            }
        }
    }, [loadStatus, message]);

    const parseImage = useCallback(async () => {
        let values;
        try {
            values = await installForm.validateFields(["image_path"]);
        } catch {
            return;
        }
        const path = values.image_path.trim();
        if (actionRef.current || parsingPathRef.current === path || !mountedRef.current) {
            return;
        }
        const requestId = ++imageRequestRef.current;
        parsingPathRef.current = path;
        setImage(undefined);
        setImageParsing(true);
        try {
            const response = await parseRedisImage({body: {image_path: path}});
            if (!mountedRef.current || imageRequestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200 || !response.data?.id) {
                message.error(responseError(response, "解析 Redis 镜像失败"));
                return;
            }
            setImage({...response.data, path});
        } catch {
            if (mountedRef.current && imageRequestRef.current === requestId) {
                message.error("解析 Redis 镜像失败");
            }
        } finally {
            if (mountedRef.current && imageRequestRef.current === requestId) {
                parsingPathRef.current = "";
                setImageParsing(false);
            }
        }
    }, [installForm, message]);

    const submitInstall = useCallback(async (values: any) => {
        if (!image || image.path !== values.image_path.trim() || parsingPathRef.current) {
            message.warning("请先解析当前镜像");
            return;
        }
        await runAction("install", () => installRedis({body: {image_id: image.id}}), "安装 Redis 失败");
    }, [image, message, runAction]);

    const uninstall = useCallback(async () => {
        setTerminalOpen(false);
        imageRequestRef.current += 1;
        parsingPathRef.current = "";
        setImage(undefined);
        setImageParsing(false);
        await runAction("uninstall", () => uninstallRedis(), "卸载 Redis 失败");
        if (mountedRef.current) {
            setConfig("");
            setConfigLoaded(false);
            setLogs([]);
            setLogCursor(0);
        }
    }, [runAction]);

    const savePerformance = useCallback(async (values: any) => {
        await runAction("performance", () => updateRedisPerformance({body: values}), "保存性能配置失败");
    }, [runAction]);

    const savePersistence = useCallback(async (values: any) => {
        await runAction("persistence", () => updateRedisPersistence({body: values}), "保存持久化配置失败");
    }, [runAction]);

    const loadConfig = useCallback(async () => {
        setConfigLoading(true);
        try {
            const response = await getRedisConfig();
            if (!mountedRef.current) return;
            if (response?.code !== 200) {
                message.error(responseError(response, "获取 Redis 配置失败"));
                return;
            }
            setConfig(String(response.data?.content || ""));
            setConfigLoaded(true);
        } catch {
            if (mountedRef.current) message.error("获取 Redis 配置失败");
        } finally {
            if (mountedRef.current) {
                setConfigLoading(false);
            }
        }
    }, [message]);

    const saveConfig = useCallback(async () => {
        if (!config.trim()) {
            message.warning("配置内容不能为空");
            return;
        }
        await runAction("config", () => updateRedisConfig({body: {content: config}}), "保存 Redis 配置失败");
    }, [config, message, runAction]);

    const loadLogs = useCallback(async (after = 0) => {
        setLogLoading(true);
        try {
            const response = await getRedisLog({body: {after}});
            if (!mountedRef.current) return;
            if (response?.code !== 200) {
                message.error(responseError(response, "获取 Redis 日志失败"));
                return;
            }
            const data = response.data || {};
            const nextLines = Array.isArray(data.lines) ? data.lines.map((line: any) => String(line)) : [];
            setLogs((current) => after > 0 ? [...current, ...nextLines] : nextLines);
            setLogCursor(Number(data.cursor || 0));
        } catch {
            if (mountedRef.current) message.error("获取 Redis 日志失败");
        } finally {
            if (mountedRef.current) {
                setLogLoading(false);
            }
        }
    }, [message]);

    const clearLogs = useCallback(async () => {
        await runAction("clear-log", async () => {
            const response = await clearRedisLog();
            if (mountedRef.current && response?.code === 200) {
                setLogs([]);
                setLogCursor(0);
            }
            return response;
        }, "清理 Redis 日志失败");
    }, [runAction]);

    const tabItems = useMemo(() => [
        {
            key: "performance",
            label: "性能配置",
            children: (
                <Form form={performanceForm} className={styles.form} layout="vertical" variant="filled" onFinish={savePerformance}>
                    <Row gutter={[16, 0]}>
                        <Col xs={24} sm={12}>
                            <Form.Item name="bind" label="绑定 IP" tooltip="修改绑定 IP 可能会影响访问安全">
                                <Input placeholder="例如 127.0.0.1"/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="port" label="绑定端口">
                                <InputNumber min={1} max={65535} style={{width: "100%"}}/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="timeout" label="空闲连接超时（秒）">
                                <InputNumber min={0} style={{width: "100%"}}/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="maxclients" label="最大连接数">
                                <InputNumber min={1} style={{width: "100%"}}/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="databases" label="数据库数量">
                                <InputNumber min={1} max={2147483647} style={{width: "100%"}}/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="maxmemory" label="最大使用内存（MB）">
                                <InputNumber min={0} style={{width: "100%"}}/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="requirepass" label="Redis 密码">
                                <Input.Password placeholder="留空表示不设置密码"/>
                            </Form.Item>
                        </Col>
                    </Row>
                    <div className="form-actions">
                        <Button htmlType="submit" type="primary" loading={action === "performance"}>保存配置</Button>
                    </div>
                </Form>
            ),
        },
        {
            key: "persistence",
            label: "持久化",
            children: (
                <Form form={persistenceForm} className={styles.form} layout="vertical" variant="filled" onFinish={savePersistence}>
                    <Row gutter={[16, 0]}>
                        <Col xs={24} sm={12}>
                            <Form.Item name="appendonly" label="AOF 持久化" valuePropName="checked">
                                <Switch checkedChildren="开启" unCheckedChildren="关闭"/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="appendfsync" label="AOF 同步策略">
                                <Select options={[
                                    {label: "每秒同步", value: "everysec"},
                                    {label: "每次写入同步", value: "always"},
                                    {label: "由系统决定", value: "no"},
                                ]}/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="auto_aof_rewrite_percentage" label="AOF 自动重写比例（%）">
                                <InputNumber min={0} style={{width: "100%"}}/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="auto_aof_rewrite_min_size" label="AOF 自动重写最小值（MB）">
                                <InputNumber min={0} style={{width: "100%"}}/>
                            </Form.Item>
                        </Col>
                    </Row>
                    <Form.Item label="RDB 自动快照规则">
                        <Form.List name="save">
                            {(fields, {add, remove}) => <>
                                {fields.map(({key, name, ...restField}) => <div className={styles.rule} key={key}>
                                    <Form.Item {...restField} name={[name, "seconds"]} rules={[{required: true, message: "请输入秒数"}]}>
                                        <InputNumber min={1} style={{width: "100%"}} placeholder="间隔秒数"/>
                                    </Form.Item>
                                    <Form.Item {...restField} name={[name, "changes"]} rules={[{required: true, message: "请输入次数"}]}>
                                        <InputNumber min={1} style={{width: "100%"}} placeholder="修改次数"/>
                                    </Form.Item>
                                    <Button className="remove" color="danger" variant="filled" aria-label="删除规则" icon={<Icon type="DeleteOutlined"/>} onClick={() => remove(name)}/>
                                </div>)}
                                <Button color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} onClick={() => add({seconds: 3600, changes: 1})}>添加规则</Button>
                            </>}
                        </Form.List>
                    </Form.Item>
                    <div className="form-actions">
                        <Button htmlType="submit" type="primary" loading={action === "persistence"}>保存配置</Button>
                    </div>
                </Form>
            ),
        },
        {
            key: "config",
            label: "配置文件",
            children: (
                <div className={styles.config}>
                    <Space style={{marginBottom: 12}} wrap>
                        <Button color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} loading={configLoading} onClick={() => void loadConfig()}>读取配置</Button>
                        <Button type="primary" loading={action === "config"} disabled={!configLoaded} onClick={() => void saveConfig()}>保存配置</Button>
                        <Button color="default" variant="filled" loading={action === "reload"} disabled={!configLoaded} onClick={() => void runAction("reload", () => reloadRedis(), "重载 Redis 配置失败")}>重载配置</Button>
                    </Space>
                    {configLoaded ? <Input.TextArea value={config} onChange={(event) => setConfig(event.target.value)} spellCheck={false}/> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请先读取配置文件"/>}
                </div>
            ),
        },
        {
            key: "log",
            label: "运行日志",
            children: (
                <div>
                    <Space style={{marginBottom: 12}} wrap>
                        <Button color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} loading={logLoading} onClick={() => void loadLogs()}>刷新日志</Button>
                        <Button color="danger" variant="filled" icon={<Icon type="ClearOutlined"/>} loading={action === "clear-log"} onClick={() => void clearLogs()}>清理日志</Button>
                        {logCursor > 0 && <Button color="default" variant="filled" disabled={logLoading} onClick={() => void loadLogs(logCursor)}>读取新增</Button>}
                    </Space>
                    {logs.length > 0 ? <pre className={styles.log}>{logs.join("\n")}</pre> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无日志"/>}
                </div>
            ),
        },
    ], [action, clearLogs, config, configLoaded, configLoading, loadConfig, loadLogs, logCursor, logLoading, performanceForm, persistenceForm, runAction, saveConfig, savePerformance, savePersistence, styles.config, styles.form, styles.log, styles.rule, styles.runtime]);

    const statusLabel = enumData?.status?.[status?.status] || displayValue(status?.status, "未安装");
    const statusColor = status?.status === "running" ? "success" : status?.status === "failed" ? "error" : "default";

    return (
        <>
                <div className={styles.content}>
                    <Row gutter={[12, 12]}>
                        <Col xs={24} lg={8}>
                            <Card className={styles.panel} title="Redis 服务" extra={<Button color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} loading={refreshing} onClick={() => void loadStatus(true)}>刷新</Button>}>
                                {loading ? <div style={{padding: 36, textAlign: "center"}}><Spin/></div> : <>
                                    <div className={styles.status}>
                                        <div className="status-item"><span className="label">状态</span><span className="value"><Tag color={statusColor}>{statusLabel}</Tag></span></div>
                                        <div className="status-item"><span className="label">就绪</span><span className="value">{status?.ready ? "是" : "否"}</span></div>
                                        <div className="status-item"><span className="label">进程 PID</span><span className="value">{displayValue(status?.pid, "0")}</span></div>
                                        <div className="status-item"><span className="label">端口</span><span className="value">{displayValue(status?.port, "-")}</span></div>
                                    </div>
                                    {status?.error && <Alert type="warning" showIcon message={status.error} style={{marginBottom: 14}}/>}
                                    {status?.restart_required && <Alert type="info" showIcon message="配置已保存，重启 Redis 后生效" style={{marginBottom: 14}}/>}
                                    {installed && (status?.image_version || status?.image_name) && <div className={styles.imageSummary}>
                                        <span>镜像</span><span>{status.image_name || "Redis"}{status.image_version ? ` · ${status.image_version}` : ""}</span>
                                    </div>}
                                    {!installed ? <Form form={installForm} className={styles.form} layout="vertical" variant="filled" onFinish={submitInstall} onValuesChange={() => {
                                        imageRequestRef.current += 1;
                                        parsingPathRef.current = "";
                                        setImage(undefined);
                                        setImageParsing(false);
                                    }}>
                                        <Typography.Paragraph type="secondary" className={styles.imageHint}>支持官方 Redis 镜像，自动适配版本配置。请选择服务器本地镜像文件，解析后安装。</Typography.Paragraph>
                                        <Form.Item label="Redis 镜像路径">
                                            <div className={styles.imageInput}>
                                                <Form.Item name="image_path" noStyle rules={[{required: true, whitespace: true, message: "请输入镜像绝对路径"}]}>
                                                    <Input placeholder="例如 /opt/images/redis.tar" disabled={Boolean(action)}/>
                                                </Form.Item>
                                                <Button color="default" variant="filled" loading={imageParsing} disabled={Boolean(action)} onClick={() => void parseImage()}>解析</Button>
                                            </div>
                                        </Form.Item>
                                        {image && <div className={styles.imagePreview}>
                                            <dl>
                                                <dt>版本</dt><dd>{image.version}</dd>
                                                <dt>平台</dt><dd>{image.os} / {image.architecture}</dd>
                                                <dt>运行用户</dt><dd>{image.user}</dd>
                                                <dt>启动命令</dt><dd className="code">{[...(image.entrypoint || []), ...(image.command || [])].join(" ")}</dd>
                                                <dt>持久化路径</dt><dd className="code">{image.data_path}</dd>
                                                <dt>配置路径</dt><dd className="code">{image.config_path}</dd>
                                            </dl>
                                            <Collapse ghost size="small" items={[{
                                                key: "config",
                                                label: "默认配置",
                                                children: <pre className={styles.runtime}>{image.config}</pre>,
                                            }]}/>
                                        </div>}
                                        <div className="form-actions"><Button type="primary" htmlType="submit" disabled={!image || imageParsing || (Boolean(action) && action !== "install")} loading={action === "install"}>安装 Redis</Button></div>
                                    </Form> : <Space wrap>
                                        <Button type="primary" disabled={status?.status === "running" || Boolean(action)} loading={action === "start"} onClick={() => void runAction("start", () => startRedis(), "启动 Redis 失败")}>启动</Button>
                                        <Button color="default" variant="filled" disabled={status?.status !== "running" || Boolean(action)} loading={action === "stop"} onClick={() => void runAction("stop", () => stopRedis(), "停止 Redis 失败")}>停止</Button>
                                        <Button color="default" variant="filled" disabled={Boolean(action)} loading={action === "restart"} onClick={() => void runAction("restart", () => restartRedis(), "重启 Redis 失败")}>重启</Button>
                                        <Button color="default" variant="filled" disabled={status?.status !== "running" || Boolean(action)} onClick={() => setTerminalOpen(true)}>终端</Button>
                                        <Button color="danger" variant="filled" disabled={Boolean(action)} loading={action === "uninstall"} onClick={() => { confirmRef.current = modal.confirm({
                                            title: "卸载 Redis",
                                            content: "卸载会停止并删除 Redis 容器、配置和持久化数据，通用镜像会保留。确定继续吗？",
                                            okText: "卸载",
                                            cancelText: "取消",
                                            okButtonProps: {danger: true},
                                            onOk: uninstall,
                                        }); }}>卸载</Button>
                                    </Space>}
                                    {installed && (status?.info || status?.resources) && <Collapse ghost items={[{
                                        key: "runtime",
                                        label: "运行信息",
                                        children: <pre className={styles.runtime}>{JSON.stringify({resources: status.resources, info: status.info}, null, 2)}</pre>,
                                    }]}/>}
                                </>}
                            </Card>
                        </Col>
                        <Col xs={24} lg={16}>
                            <Card className={styles.panel} title="Redis 管理" bordered={false}>
                                {installed ? <Tabs items={tabItems} onChange={(key) => {
                                    if (key === "config" && !configLoaded) {
                                        void loadConfig();
                                    }
                                    if (key === "log" && logs.length === 0) {
                                        void loadLogs();
                                    }
                                }}/> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="安装 Redis 后可管理配置和日志"/>}
                            </Card>
                        </Col>
                    </Row>
                </div>
            <ContainerTerminal id="redis" open={terminalOpen && installed} onClose={() => setTerminalOpen(false)}/>
        </>
    );
};
