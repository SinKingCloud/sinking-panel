import {useEffect, useMemo, useRef, useState} from "react";
import {Button, Input, Space, Typography} from "antd";
import {createStyles} from "antd-style";
import {Icon, useTheme} from "sinking-antd";
import {useMysqlClient} from "../context";
import queryHistory from "./history";
import useSqlCompletion from "./use-completion";
import QueryHistoryModal from "./history-modal";

const useStyles = createStyles(({css, token}) => ({
    editor: css`
        display: flex; flex-direction: column; height: 100%; min-height: 0; box-sizing: border-box;
        overflow: hidden; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px;
        .query-editor-toolbar, .query-editor-footer { flex: none; }
        .query-editor-content { flex: 1; min-height: 0; overflow: hidden; }
        .query-editor-toolbar { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; padding: 8px 10px; background: ${token.colorFillQuaternary}; }
        .query-editor-toolbar .ant-btn { font-size: ${token.fontSize}px; }
        .query-editor-toolbar .ant-typography { font-size: ${token.fontSizeSM}px; }
        .query-editor-footer { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 4px 8px; padding: 6px 10px; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; background: ${token.colorFillQuaternary}; }
        .query-editor-footer .ant-typography { font-size: inherit; color: inherit; }
        .ant-input { font-family: ${token.fontFamilyCode}; }
    `,
}));

export default ({initialValue, onChange, onRun, loading, record, schema, table, tables = [], columns = [], completionKey, execution}: any) => {
    const {api, Editor: AceEditor, ScriptTools} = useMysqlClient();
    const theme = useTheme();
    const {styles} = useStyles();
    const [value, setValue] = useState(initialValue || "");
    const change = (next: string) => { setValue(next); onChange(next); };
    const ace = useRef<any>();
    const attachCompletion = useSqlCompletion({api, record, schema, table, tables, columns, revision: completionKey});
    const textarea = useRef<any>();
    const current = useRef<any>();
    current.current = {value, onRun, loading};
    const [fallback, setFallback] = useState(!AceEditor);
    const [selected, setSelected] = useState(false);
    const [historyOpen, setHistoryOpen] = useState(false);
    const [history, setHistory] = useState<any[]>([]);
    const executionTime = useRef(0);
    const historyId = String(record.id);
    const getSelection = () => {
        if (fallback) {
            const input = textarea.current?.resizableTextArea?.textArea;
            return input ? String(current.current.value).slice(input.selectionStart, input.selectionEnd) : "";
        }
        return ace.current?.getSelectedText?.() || "";
    };
    const run = () => {
        if (!current.current.loading) current.current.onRun(getSelection().trim() || undefined);
    };
    const commandRef = useRef<any>();
    commandRef.current = run;
    const commands = useMemo(() => [{name: "runSqlQuery", bindKey: {win: "Ctrl-Enter", mac: "Command-Enter"}, exec: () => commandRef.current()}], []);
    useEffect(() => {
        executionTime.current = 0;
        setHistory(queryHistory(historyId));
    }, [historyId]);
    useEffect(() => {
        if (!execution || execution.time === executionTime.current) return;
        executionTime.current = execution.time;
        // 不保存被截断的 SQL，避免重新填入时执行了不同语句。
        if (typeof execution.statement !== "string" || execution.statement.length > 65536) return;
        setHistory(queryHistory(historyId, execution));
    }, [execution, historyId]);
    return <>
        <div className={styles.editor}>
            <div className="query-editor-toolbar">
                <Button color="primary" variant="filled" size="small" loading={loading} disabled={!value?.trim()} icon={<Icon type="PlayCircleOutlined"/>} onClick={run}>{selected ? "执行选区" : "执行SQL"}</Button>
                <Space size={8} wrap>
                    {ScriptTools && <ScriptTools instanceId={record.id} value={value}
                        disabled={loading} mode="sql" Editor={AceEditor} onUse={(statement: string) => { change(statement); setSelected(false); }}/>}
                    <Button color="default" variant="filled" size="small" onClick={() => { setHistory(queryHistory(historyId)); setHistoryOpen(true); }}>查询历史</Button>
                </Space>
            </div>
            <div className="query-editor-content">
                {fallback || !AceEditor ? <Input.TextArea ref={textarea} placeholder="请输入 SQL 语句" value={value} onChange={(event) => change(event.target.value)} style={{height: "100%", resize: "none"}}
                    onSelect={() => setSelected(Boolean(getSelection().trim()))} onKeyDown={(event) => { if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) { event.preventDefault(); commandRef.current(); } }}/>
                    : <AceEditor placeholder="请输入 SQL 语句" value={value} onChange={change} mode="sql" height="100%" style={{height: "100%"}} fontSize={12}
                        theme={theme?.isDarkTheme?.() ? "tomorrow_night" : "textmate"} showPrintMargin={false} wrapEnabled commands={commands} enableSnippets={false}
                        onLoad={(editor: any) => { ace.current = editor; attachCompletion(editor); }} onSelectionChange={() => setSelected(Boolean(getSelection().trim()))}
                        onError={() => { ace.current = undefined; setSelected(false); setFallback(true); }}/>}
            </div>
            <div className="query-editor-footer"><Typography.Text>{selected ? "已选择部分 SQL" : "SQL 编辑器"}</Typography.Text><Typography.Text>{!fallback && AceEditor ? "Alt / ⌥ + / 补全 · " : ""}Ctrl / ⌘ + Enter 执行</Typography.Text></div>
        </div>
        <QueryHistoryModal open={historyOpen} history={history} loading={loading} onCancel={() => setHistoryOpen(false)}
            onUse={(statement: string) => { change(statement); setHistoryOpen(false); }} onRefresh={() => setHistory(queryHistory(historyId))}/>
    </>;
};
