import {useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Input, Space, Splitter, Tooltip} from "antd";
import {CopyOutlined, PlayCircleOutlined} from "@ant-design/icons";
import {useTheme} from "sinking-antd";
import {useDatabaseClient} from "../shared/context";
import {jsonValue, VirtualTable} from "../shared/client-common";
import {copyTextToClipboard} from "../shared/clipboard";
import queryHistory from "../shared/sql/history";
import QueryHistoryModal from "../shared/sql/history-modal";
import {parseMongoShell} from "./shell";
import useStyles from "./aggregate-styles";

export default function MongoAggregate({record, collection, onChange, active = true}: any) {
    const {operate, features = {}, Editor, ScriptTools} = useDatabaseClient();
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const [shell, setShell] = useState("");
    const [selected, setSelected] = useState(false);
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<any>();
    const [error, setError] = useState("");
    const [fallback, setFallback] = useState(!Editor);
    const [historyOpen, setHistoryOpen] = useState(false);
    const [history, setHistory] = useState<any[]>([]);
    const ace = useRef<any>();
    const textarea = useRef<any>();
    const alive = useRef(true);
    const request = useRef(0);
    const busy = useRef(false);
    const confirmation = useRef<any>();
    const runRef = useRef<() => void>(() => {});
    // 集合由语句指定，数据浏览切换集合不会清空当前查询结果。
    const scope = `${record.id}:${record.database_name ?? record.name}`;
    const scopeRef = useRef(scope);
    scopeRef.current = scope;
    const editorCommands = useMemo(() => [{name: "mongo-run", bindKey: {win: "Ctrl-Enter", mac: "Command-Enter"}, exec: () => runRef.current()}], []);
    useEffect(() => { alive.current = true; return () => { alive.current = false; request.current++; confirmation.current?.destroy(); }; }, []);
    useEffect(() => {
        request.current++; busy.current = false; setLoading(false); setResult(undefined); setError("");
        confirmation.current?.destroy(); confirmation.current = undefined;
        setHistoryOpen(false);
    }, [scope]);
    const getSelection = () => {
        if (fallback || !Editor) {
            const input = textarea.current?.resizableTextArea?.textArea;
            return input ? shell.slice(input.selectionStart, input.selectionEnd) : "";
        }
        return ace.current?.getSelectedText?.() || "";
    };
    const run = () => {
        if (busy.current || confirmation.current) return;
        const statement = getSelection().trim() || shell.trim();
        if (!statement) return;
        let parsed: ReturnType<typeof parseMongoShell>;
        try {
            parsed = parseMongoShell(statement);
            if (parsed.operation === "aggregate") {
                if (features.aggregate === false) throw new Error("当前 MongoDB 版本不支持聚合查询");
                const values = jsonValue(parsed.body.pipeline, "array");
                const featureNames: Record<string, string> = {$lookup: "aggregate_32", $sample: "aggregate_32", $addFields: "aggregate_34", $facet: "aggregate_34", $count: "aggregate_34", $set: "aggregate_42", $unset: "aggregate_42", $replaceWith: "aggregate_42", $unionWith: "aggregate_44"};
                values.forEach((stage: any) => {
                    if (!stage || typeof stage !== "object" || Array.isArray(stage) || Object.keys(stage).length !== 1) throw new Error("每个步骤只能包含一个聚合操作");
                    const name = Object.keys(stage)[0];
                    if (featureNames[name] && features[featureNames[name]] === false) throw new Error(`当前 MongoDB 版本不支持 ${name}`);
                });
                parsed.body = {...parsed.body, limit: 500};
            }
        } catch (reason: any) { setResult(undefined); setError(reason?.message || "请检查查询语句"); return; }
        const {operation, body, write, kind} = parsed;
        const targetScope = scope;
        const execute = async () => {
            if (!alive.current || targetScope !== scopeRef.current || busy.current) return;
            const id = ++request.current;
            const time = Date.now();
            const execution = {statement, time, operation: write ? "execute" : "query", success: false, elapsed_ms: 0};
            busy.current = true; setLoading(true); setResult(undefined); setError("");
            try {
                let response = await operate(record, operation, body);
                execution.success = true;
                execution.elapsed_ms = response?.elapsed_ms ?? Date.now() - time;
                if (alive.current && id === request.current && targetScope === scopeRef.current) {
                    if (kind === "countDocuments") {
                        const count = response?.documents?.[0]?.count ?? 0;
                        response = {...response, documents: undefined, result: typeof count === "object" ? count.$numberLong ?? count.$numberInt ?? count : count};
                    } else if (kind === "findOne") response = {...response, documents: undefined, result: response?.documents?.[0] ?? null};
                    else if (kind === "distinct" && Array.isArray(response?.result?.values)) response = {...response, result: response?.result?.values};
                    setResult(response); if (write) onChange?.();
                }
            } catch (reason: any) {
                execution.elapsed_ms = Date.now() - time;
                if (alive.current && id === request.current && targetScope === scopeRef.current) setError(reason?.message || "执行失败，请检查输入内容");
            } finally {
                // 保存实际执行的选区，完整语句过长时不截断写入历史。
                const records = statement.length <= 65536 ? queryHistory(String(record.id), execution) : undefined;
                if (alive.current && id === request.current && targetScope === scopeRef.current) {
                    busy.current = false; setLoading(false);
                    if (records) setHistory(records);
                }
            }
        };
        if (!write) { void execute(); return; }
        const dialog = modal.confirm({title: "执行 MongoDB 命令？", content: "命令将发送到当前数据库，可能修改数据库内容。", okText: "执行", cancelText: "取消",
            afterClose: () => { if (confirmation.current === dialog) confirmation.current = undefined; }, onOk: execute});
        confirmation.current = dialog;
    };
    runRef.current = run;
    const rows = useMemo(() => (Array.isArray(result?.documents) ? result?.documents : []).map((document: any, index: number) => ({key: index, document})), [result?.documents]);
    const columns = useMemo(() => Array.from(new Set<string>(rows.flatMap(({document}: any) => Object.keys(document || {})))).map((name) => ({title: name, key: name, width: name === "_id" ? 200 : 160, ellipsis: true,
        render: (_: any, {document}: any) => {
            const value = document?.[name];
            let text = value === null ? "空值" : value === undefined ? "—" : typeof value === "boolean" ? value ? "是" : "否" : typeof value === "object" ? JSON.stringify(value) : String(value);
            if (value && typeof value === "object" && Object.keys(value).length === 1) {
                const key = Object.keys(value)[0];
                if (["$oid", "$numberInt", "$numberLong", "$numberDouble", "$numberDecimal"].includes(key)) text = String(value[key]);
            }
            return <Tooltip title={text.length < 3000 ? text : undefined}><span className={value === null ? "manager-cell-null" : "manager-cell"}>{text}</span></Tooltip>;
        }})), [rows]);
    const documentResult = Array.isArray(result?.documents);
    const output = useMemo(() => error || (documentResult ? "" : result?.result === undefined ? "执行成功" : typeof result?.result === "string" ? result?.result : JSON.stringify(result?.result, null, 2)), [error, documentResult, result]);
    const placeholder = `请输入 mongosh 语句，例如 db.getCollection(${JSON.stringify(collection || "users")}).aggregate([{ $limit: 20 }])`;
    if (!active) return null;
    return <>
        <div className={styles.workbench}>
            <Splitter className="mongo-aggregate-splitter" orientation="vertical" classNames={{dragger: {default: "mongo-query-resizer"}}}
                styles={{dragger: {default: {height: 16, top: 2}}}}>
                <Splitter.Panel defaultSize={compact ? 180 : 200} min={compact ? 180 : 200}>
                    <section className="mongo-query-input">
                        <div className="mongo-query-toolbar">
                            <Button color="primary" variant="filled" size="small" loading={loading} disabled={!shell.trim()}
                                icon={<PlayCircleOutlined/>} onClick={run}>{selected ? "执行选区" : "执行查询"}</Button>
                            <Space size={8} wrap>
                                {ScriptTools && <ScriptTools instanceId={record.id} value={shell}
                                    disabled={loading} mode="javascript" Editor={Editor} onUse={(statement: string) => {
                                        if (busy.current || confirmation.current) return;
                                        setShell(statement); setSelected(false);
                                    }}/>}
                                <Button color="default" variant="filled" size="small" onClick={() => { setHistory(queryHistory(String(record.id))); setHistoryOpen(true); }}>查询历史</Button>
                            </Space>
                        </div>
                        <div className="mongo-query-editor">
                            {Editor && !fallback ? <Editor value={shell} mode="javascript" height="100%" style={{height: "100%"}} fontSize={12}
                                theme={theme?.isDarkTheme?.() ? "tomorrow_night" : "textmate"} placeholder={placeholder}
                                showPrintMargin={false} wrapEnabled commands={editorCommands} enableSnippets={false} onChange={setShell}
                                onLoad={(editor: any) => { ace.current = editor; setSelected(Boolean(editor.getSelectedText?.().trim())); }} onSelectionChange={() => setSelected(Boolean(getSelection().trim()))}
                                onError={() => { ace.current = undefined; setSelected(false); setFallback(true); }}/>
                                : <Input.TextArea ref={textarea} aria-label="mongosh 查询内容" spellCheck={false} value={shell} placeholder={placeholder}
                                    onChange={(event) => setShell(event.target.value)} onSelect={() => setSelected(Boolean(getSelection().trim()))}
                                    onKeyDown={(event) => { if ((event.ctrlKey || event.metaKey) && event.key === "Enter") { event.preventDefault(); runRef.current(); } }}/>
                            }
                        </div>
                        <div className="mongo-query-footer"><span>{selected ? "已选择部分语句" : "mongosh 编辑器"}</span><span>Ctrl / ⌘ + Enter 执行</span></div>
                    </section>
                </Splitter.Panel>
                <Splitter.Panel className="mongo-query-results" min={220}>
                    {(result !== undefined || error) && <>
                        <div className="mongo-query-status">
                            <div><span>{error ? "执行失败" : "执行结果"}</span><span className="mongo-query-statistics">{documentResult ? `${rows.length} 条记录` : ""}{result?.elapsed_ms !== undefined ? `${documentResult ? " · " : ""}${result?.elapsed_ms} ms` : ""}</span></div>
                            {(rows.length > 0 || output) && <Button color="default" variant="filled" size="small" autoInsertSpace={false} icon={<CopyOutlined/>}
                                onClick={() => void copyTextToClipboard(documentResult && !error ? JSON.stringify(result?.documents, null, 2) : output).then((success) => success ? message.success("复制成功") : message.error("复制失败"))}>复制</Button>}
                        </div>
                        {output && <pre className={`mongo-query-output${error ? " is-error" : ""}`} role={error ? "alert" : "status"}>{output}</pre>}
                        {!error && documentResult && rows.length > 0 && <div className="manager-table"><VirtualTable rowKey="key" size="small" pagination={false} dataSource={rows} columns={columns} scroll={{x: Math.max(500, columns.length * 160)}}/></div>}
                        {!error && (result?.warning || result?.has_more || result?.truncated) && <div className="mongo-query-hint">{result?.warning || "仅显示部分结果，请缩小查询范围。"}</div>}
                    </>}
                </Splitter.Panel>
            </Splitter>
        </div>
        <QueryHistoryModal open={historyOpen} history={history} loading={loading} searchPlaceholder="搜索 mongosh" onCancel={() => setHistoryOpen(false)}
            onUse={(statement: string) => { setShell(statement); setSelected(false); setHistoryOpen(false); }} onRefresh={() => setHistory(queryHistory(String(record.id)))}/>
    </>;
}
