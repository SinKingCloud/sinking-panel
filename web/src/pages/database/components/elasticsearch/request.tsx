import {useRef, useState, useEffect, useMemo} from "react";
import {App, Button, Empty, Form, Input, Select, Space, Spin, Splitter} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {useDatabaseClient} from "../shared/context";
import {copyTextToClipboard} from "../shared/clipboard";
import requestHistory from "./history";
import RequestHistoryModal from "./history-modal";
import useStyles from "./request-styles";

export default ({record}: any) => {
    const {modal, message} = App.useApp();
    const {operate, Editor: AceEditor, ScriptTools} = useDatabaseClient();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const [form] = Form.useForm();
    const values = Form.useWatch([], form);
    const [fallback, setFallback] = useState(false);
    const [result, setResult] = useState<any>();
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);
    const [historyOpen, setHistoryOpen] = useState(false);
    const [history, setHistory] = useState<any[]>([]);
    const live = useRef(true);
    const pending = useRef(false);
    const confirmation = useRef<any>(undefined);
    const commands = useMemo(() => [{name: "executeRequest", bindKey: {win: "Ctrl-Enter", mac: "Command-Enter"}, exec: () => form.submit()}], [form]);
    useEffect(() => { live.current = true; return () => { live.current = false; confirmation.current?.destroy(); }; }, []);
    const execute = async (values: any) => {
        if (pending.current || !live.current) return;
        pending.current = true; setLoading(true); setError(""); setResult(undefined);
        const params = {method: values.method, path: values.path.trim(), body: values.body?.trim() || ""};
        const execution: any = {...params, time: Date.now(), success: false};
        try {
            const data = await operate(record, "request", params);
            execution.success = true;
            execution.status = data?.status;
            if (live.current) setResult(data);
        } catch (reason: any) { if (live.current) setError(reason?.message || "请求失败"); }
        finally {
            // 只记录实际执行的请求；关闭工作区后完成的请求也保留历史。
            const records = requestHistory(String(record.id), {...execution, elapsed_ms: Math.max(0, Date.now() - execution.time)});
            pending.current = false;
            if (live.current) { setHistory(records); setLoading(false); }
        }
    };
    const output = useMemo(() => error || (result == null ? "" : typeof result?.data === "string" ? result?.data : (JSON.stringify(result?.data, null, 2) ?? "")), [error, result]);
    return <div className={`manager-stage ${styles.request}`}>
        <Form form={form} className="es-request-form" layout="vertical" variant="filled" initialValues={{method: "GET", path: record.name ? `/${record.name}/_search` : "/_cluster/health", body: ""}}
            onFinishFailed={({errorFields}) => message.error(errorFields[0]?.errors?.[0] || "请检查请求参数")}
            onFinish={(values) => {
                if (pending.current) return;
                if (values.method !== "DELETE") { void execute(values); return; }
                confirmation.current?.destroy();
                confirmation.current = modal.confirm({title: "执行删除请求", content: `确定向 ${values.path} 发送 DELETE 请求？`, okText: "执行", cancelText: "取消", okButtonProps: {danger: true}, onOk: () => execute(values)});
            }}>
            <div className="manager-toolbar"><div className="manager-tools" style={{flex: 1}}>
                <Form.Item name="method" noStyle><Select style={{width: compact ? 82 : 88}} popupMatchSelectWidth={false} placeholder="请求方法" options={["GET", "POST", "PUT", "DELETE", "HEAD"].map((value) => ({value, label: value}))}/></Form.Item>
                <Form.Item name="path" noStyle rules={[{required: true, whitespace: true, message: "请输入请求路径"}]}><Input style={{flex: 1, minWidth: 120}} aria-label="请求路径" placeholder="例如 /_cluster/health"/></Form.Item>
                <Button color="default" variant="filled" htmlType="submit" icon={<Icon type="CaretRightOutlined"/>} loading={loading}>执行请求</Button>
            </div></div>
            <Splitter className="es-request-splitter" orientation="vertical"
                classNames={{dragger: {default: "es-request-resizer"}}} styles={{dragger: {default: {height: 16, top: 2}}}}>
                <Splitter.Panel defaultSize={compact ? 160 : 180} min={compact ? 160 : 180}>
                    <div className="es-request-editor">
                        <div className="es-request-content"><Form.Item name="body" noStyle rules={[{validator: async (_, value) => {
                            if (!value?.trim()) return;
                            try { const data = JSON.parse(value); if (!data || typeof data !== "object" || Array.isArray(data)) throw new Error(); }
                            catch { throw new Error("请输入有效的 JSON 对象"); }
                        }}]}>{AceEditor && !fallback ? <AceEditor mode="json" height="100%" fontSize={compact ? 12 : 13}
                            placeholder="请输入 JSON 请求内容，无请求内容时可留空" theme={appearance?.isDarkTheme?.() ? "tomorrow_night" : "textmate"}
                            showPrintMargin={false} wrapEnabled enableSnippets={false} commands={commands} onError={() => setFallback(true)}/>
                            : <Input.TextArea spellCheck={false} placeholder="请输入 JSON 请求内容，无请求内容时可留空"
                                onKeyDown={(event) => { if ((event.ctrlKey || event.metaKey) && event.key === "Enter") { event.preventDefault(); form.submit(); } }}/>}</Form.Item></div>
                        <div className="manager-query-actions"><span>Ctrl / ⌘ + Enter 执行</span>
                            <Space size={8} wrap>
                                {ScriptTools && <ScriptTools instanceId={record.id}
                                    value={`${values?.method || "GET"} ${values?.path ?? (record.name ? `/${record.name}/_search` : "/_cluster/health")}\n\n${values?.body || ""}`}
                                    disabled={loading} mode="text" Editor={AceEditor} onUse={(script: string) => {
                                        if (pending.current) return;
                                        const request = /^(GET|POST|PUT|DELETE|HEAD) +([^\r\n]+)(?:\r?\n(?:\r?\n)?([\s\S]*))?$/i.exec(script);
                                        if (!request) throw new Error("请求脚本格式应为：请求方法 路径，空行后填写请求内容");
                                        form.setFieldsValue({method: request[1].toUpperCase(), path: request[2].trim(), body: request[3] || ""});
                                        setResult(undefined); setError("");
                                    }}/>}
                                <Button color="default" variant="filled" onClick={() => { setHistory(requestHistory(String(record.id))); setHistoryOpen(true); }}>请求历史</Button>
                            </Space>
                        </div>
                    </div>
                </Splitter.Panel>
                <Splitter.Panel className="es-request-results" min={160}>
                    {(output || loading) ? <><div className="manager-toolbar"><span className="manager-statistics">{error ? "执行失败" : result ? `响应状态 ${result?.status}` : "响应结果"}</span>
                        <div className="manager-tools"><Button color="default" variant="filled" icon={<Icon type="CopyOutlined"/>} disabled={!output} onClick={() => copyTextToClipboard(output).then((success) => success ? message.success("复制成功") : message.error("复制失败"))}>复制</Button></div></div>
                        <div className={`manager-output${error ? " is-error" : ""}`}><Spin spinning={loading}><pre>{output}</pre></Spin></div></>
                        : <div className="manager-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="执行后显示结果"/></div>}
                </Splitter.Panel>
            </Splitter>
        </Form>
        <RequestHistoryModal open={historyOpen} history={history} loading={loading} onCancel={() => setHistoryOpen(false)}
            onRefresh={() => setHistory(requestHistory(String(record.id)))} onUse={(item: any) => {
                if (pending.current) return;
                form.setFieldsValue({method: item.method, path: item.path, body: item.body});
                setResult(undefined); setError(""); setHistoryOpen(false);
            }}/>
    </div>;
};
