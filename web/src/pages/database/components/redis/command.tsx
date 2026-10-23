import {useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Empty, Input, Spin, Splitter, Typography} from "antd";
import {PlayCircleOutlined} from "@ant-design/icons";
import {useTheme} from "sinking-antd";
import {useRedisClient} from "./context";
import commandHistory from "./history";
import CommandHistoryModal from "./history-modal";
import useStyles from "./command-styles";

export default ({record, onChange}: any) => {
    const {api, Editor: AceEditor} = useRedisClient();
    const {message, modal} = App.useApp();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const live = useRef(true);
    const sequence = useRef(0);
    const busy = useRef(false);
    const confirmation = useRef<any>(undefined);
    const scope = useRef("");
    scope.current = `${record.id}:${record.database_name ?? ""}`;
    const runRef = useRef<() => void>(() => {});
    const commands = useMemo(() => [{name: "executeRedisCommand", bindKey: {win: "Ctrl-Enter", mac: "Command-Enter"}, exec: () => runRef.current()}], []);
    const [command, setCommand] = useState("PING");
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<any>();
    const [error, setError] = useState("");
    const [executed, setExecuted] = useState("");
    const [fallback, setFallback] = useState(false);
    const [historyOpen, setHistoryOpen] = useState(false);
    const [history, setHistory] = useState<any[]>([]);

    useEffect(() => { live.current = true; return () => { live.current = false; confirmation.current?.destroy(); }; }, []);
    useEffect(() => {
        sequence.current++; busy.current = false; confirmation.current?.destroy(); confirmation.current = undefined;
        setCommand("PING"); setResult(undefined); setError(""); setExecuted(""); setLoading(false); setHistoryOpen(false); setHistory([]);
    }, [record.id, record.database_name]);

    const run = () => {
        if (busy.current || confirmation.current) return;
        // 接收完整命令，保留引号中的空格、空字符串和转义字符，再转换为接口参数。
        const argumentsList: string[] = [];
        let value = "", quote = "", started = false;
        const escapes: Record<string, string> = {n: "\n", r: "\r", t: "\t", b: "\b", f: "\f", v: "\v", a: "\x07"};
        for (let i = 0; i < command.length; i++) {
            const character = command[i];
            if (quote === "'") {
                if (character === "\\" && command[i + 1] === "'") { value += "'"; i++; }
                else if (character === "'") quote = "";
                else value += character;
            } else if (character === "\\" && !quote) {
                const next = command[i + 1];
                if (next && (/\s/.test(next) || next === '"' || next === "'")) { value += next; i++; }
                else value += character;
                started = true;
            } else if (character === "\\") {
                if (++i >= command.length) { message.error("命令末尾的转义符不完整"); return; }
                if (command[i] === "x" && /^[\da-f]{2}$/i.test(command.slice(i + 1, i + 3))) {
                    const code = parseInt(command.slice(i + 1, i + 3), 16);
                    if (code > 127) { message.error("当前命令面板不支持非 ASCII 二进制转义，请使用终端执行"); return; }
                    value += String.fromCharCode(code); i += 2;
                } else value += escapes[command[i]] ?? (command[i] === '"' || command[i] === "\\" ? command[i] : `\\${command[i]}`);
                started = true;
            } else if (quote) {
                if (character === quote) quote = "";
                else value += character;
            } else if (character === '"' || character === "'") {
                quote = character; started = true;
            } else if (/\s/.test(character)) {
                if (started) { argumentsList.push(value); value = ""; started = false; }
            } else { value += character; started = true; }
        }
        if (quote) { message.error(`命令中的${quote === '"' ? "双" : "单"}引号未闭合`); return; }
        if (started) argumentsList.push(value);
        if (!argumentsList.length) { message.error("请输入 Redis 命令"); return; }
        if (!/^[a-zA-Z][a-zA-Z0-9_.-]*$/.test(argumentsList[0])) { message.error("命令名称不正确，请检查输入内容"); return; }
        const commandName = argumentsList[0].toUpperCase();
        const target = scope.current;
        const argumentsUpper = argumentsList.map((item) => item.toUpperCase());
        // 认证和改密命令仍可执行，但不将其中的凭据写入浏览器历史。
        const sensitive = commandName === "AUTH" || commandName === "HELLO" && argumentsUpper.includes("AUTH")
            || commandName === "ACL" && argumentsUpper[1] === "SETUSER"
            || commandName === "CONFIG" && argumentsUpper[1] === "SET" && argumentsUpper.some((item, index) => index >= 2 && index % 2 === 0 && /PASS|AUTH|SECRET|TOKEN/.test(item))
            || commandName === "MIGRATE" && argumentsUpper.some((item) => item === "AUTH" || item === "AUTH2");
        const execute = async () => {
            if (!live.current || scope.current !== target || busy.current) return;
            const request = ++sequence.current;
            const execution = {time: Date.now(), command, database: String(record.database_name ?? ""), success: false, sensitive};
            busy.current = true; setLoading(true); setResult(undefined); setError(""); setExecuted(commandName);
            try {
                const data = await api.command(record, {command: JSON.stringify(argumentsList)});
                execution.success = true;
                if (!live.current || scope.current !== target || request !== sequence.current) return;
                setResult(data === undefined ? null : data); onChange?.();
            } catch (reason: any) {
                if (live.current && scope.current === target && request === sequence.current) setError(reason.message || "命令执行失败");
            } finally {
                const records = commandHistory(String(record.id), {...execution, elapsed_ms: Math.max(0, Date.now() - execution.time)});
                if (live.current && scope.current === target && request === sequence.current) { busy.current = false; setLoading(false); setHistory(records); }
            }
        };
        if (commandName === "FLUSHDB" || commandName === "FLUSHALL") {
            const dialog = modal.confirm({title: commandName === "FLUSHALL" ? "清空所有数据库？" : "清空当前数据库？",
                content: "全部数据将被删除，操作无法撤销。", okText: "清空", cancelText: "取消", okButtonProps: {danger: true},
                afterClose: () => { if (confirmation.current === dialog) confirmation.current = undefined; }, onOk: execute});
            confirmation.current = dialog;
        } else void execute();
    };
    runRef.current = run;
    const output = useMemo(() => {
        if (result === undefined) return "";
        // 只展示 Redis 回复；文本和大整数保持原值，列表按顺序分行，嵌套内容缩进。
        const format = (value: any): string => {
            if (value === null || value === undefined) return "（空值）";
            if (typeof value !== "object") return value === "" ? "（空文本）" : String(value);
            const entries = Array.isArray(value) ? value.map((item, index) => [`${index + 1})`, item]) : Object.entries(value).map(([key, item]) => [`${key}：`, item]);
            if (!entries.length) return Array.isArray(value) ? "（空列表）" : "（空对象）";
            return entries.map(([label, item]) => {
                const content = format(item);
                return item !== null && typeof item === "object" ? `${label}\n  ${content.replace(/\n/g, "\n  ")}` : `${label} ${content.replace(/\n/g, "\n   ")}`;
            }).join("\n");
        };
        return format(result?.result);
    }, [result]);
    return <div className={styles.command}>
        <Splitter className="redis-command-workbench" orientation="vertical"
            classNames={{dragger: {default: "redis-command-resizer"}}} styles={{dragger: {default: {height: 16, top: 2}}}}>
            <Splitter.Panel defaultSize={compact ? 180 : 200} min={compact ? 180 : 200}>
            <section className="redis-command-input">
                <div className="redis-command-input-heading redis-toolbar">
                    <Button color="primary" variant="filled" size="small" icon={<PlayCircleOutlined/>} loading={loading} disabled={!command.trim()} onClick={run}>执行命令</Button>
                    <Button color="default" variant="filled" size="small" onClick={() => { setHistory(commandHistory(String(record.id))); setHistoryOpen(true); }}>命令历史</Button>
                </div>
                <div className="redis-command-editor">{AceEditor && !fallback ? <AceEditor value={command} mode="text" height="100%" style={{height: "100%"}} fontSize={12}
                    theme={appearance?.isDarkTheme?.() ? "tomorrow_night" : "textmate"} placeholder={'输入完整命令，例如 SET user:1 "张三"'}
                    readOnly={loading} showPrintMargin={false} wrapEnabled enableSnippets={false} enableBasicAutocompletion={false} enableLiveAutocompletion={false}
                    commands={commands} onChange={setCommand} onError={() => setFallback(true)}/>
                    : <Input.TextArea value={command} variant="filled" placeholder={'输入完整命令，例如 SET user:1 "张三"'} spellCheck={false}
                        aria-label="Redis 命令" readOnly={loading} onChange={(event) => setCommand(event.target.value)}
                        onKeyDown={(event) => { if ((event.ctrlKey || event.metaKey) && event.key === "Enter") { event.preventDefault(); run(); } }}/>}</div>
                <div className="redis-command-controls"><Typography.Text>命令编辑器</Typography.Text><Typography.Text>Ctrl / ⌘ + Enter 执行</Typography.Text></div>
            </section>
            </Splitter.Panel>
            <Splitter.Panel min={220} className="redis-command-results">
            <section className="redis-command-result">
                <div className="redis-command-result-toolbar"><div><strong>{error ? "执行失败" : result !== undefined ? "执行成功" : "执行结果"}</strong>{executed && <span>{executed}</span>}</div>
                    {result !== undefined && !loading && <Typography.Text className="redis-copy-action" copyable={{text: output, icon: ["复制", "已复制"]}}/>}
                </div>
                <Spin spinning={loading} classNames={{root: "redis-command-spin", container: "redis-command-spin-content"}}><div className={`redis-command-output${error ? " has-error" : ""}`}>
                    {loading ? <div className="redis-command-placeholder"/>
                        : error ? <div className="redis-command-error"><p>{error}</p></div>
                            : result === undefined ? <div className="redis-command-placeholder"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/></div>
                                : <>{result?.truncated && <div className="redis-command-truncated">结果过大，以下仅显示部分返回内容。</div>}<pre>{output}</pre></>}
                </div></Spin>
            </section>
            </Splitter.Panel>
        </Splitter>
        <CommandHistoryModal open={historyOpen} history={history} loading={loading} onCancel={() => setHistoryOpen(false)}
            onRefresh={() => setHistory(commandHistory(String(record.id)))} onUse={(item: any) => {
                if (busy.current || confirmation.current) return;
                setCommand(item.command); setResult(undefined); setError(""); setExecuted(""); setHistoryOpen(false);
                if (item.database !== String(record.database_name ?? "")) message.info(`命令已填入，当前数据库为 ${record.database_name || "0"}`);
            }}/>
    </div>;
};
