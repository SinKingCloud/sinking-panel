import {useEffect, useRef, useState} from "react";
import {App, Button, Typography} from "antd";
import {useMysqlClient} from "./context";

const tools: any[] = [
    {action: "check", title: "检查数据表", description: "检查表和索引是否存在错误。", button: "开始检查"},
    {action: "analyze", title: "更新查询统计", description: "数据变化较多时，更新统计信息以帮助优化查询。", button: "更新统计"},
    {action: "optimize", title: "整理表空间", description: "整理表和索引，回收可释放的空间。", button: "整理空间"},
    {action: "repair", title: "修复数据表", description: "修复 MyISAM 数据表中的错误。", button: "修复数据"},
];

export default ({record, schema, table, isView, storageEngine, disabled = false, onChange}: any) => {
    const {api} = useMysqlClient();
    const {modal, message} = App.useApp();
    const [busy, setBusy] = useState("");
    const pending = useRef(false);
    const live = useRef(true);
    const sequence = useRef(0);
    const confirmation = useRef<any>(undefined);
    const actions = [...tools,
        {action: "truncate", title: "清空数据表", description: "删除全部记录，保留表结构。", button: "清空数据", danger: true},
        {action: "drop", title: isView ? "删除视图" : "删除数据表", description: isView ? "删除当前视图，保留原始数据表。" : "永久删除当前表及其数据和索引。", button: isView ? "删除视图" : "删除表", danger: true},
    ].filter((tool) => (!isView || tool.action === "drop") && (tool.action !== "repair" || String(storageEngine || "").toLowerCase() === "myisam"));

    useEffect(() => { live.current = true; return () => { live.current = false; sequence.current++; confirmation.current?.destroy(); }; }, []);
    useEffect(() => {
        sequence.current++; confirmation.current?.destroy();
    }, [record.id, record.database_name, record.name, schema, table]);

    const run = (tool: any) => {
        if (disabled) return;
        const requestId = sequence.current;
        const execute = async () => {
            if (pending.current || !live.current || requestId !== sequence.current) return;
            pending.current = true; setBusy(tool.action);
            try {
                const data = await api.maintenance(record, {action: tool.action, schema, table});
                if (!live.current || requestId !== sequence.current) return;
                const columns = (Array.isArray(data?.columns) ? data?.columns : []).map((name: string) => String(name ?? "").toLowerCase());
                const typeIndex = columns.indexOf("msg_type"), textIndex = columns.indexOf("msg_text");
                const results = (Array.isArray(data?.rows) ? data?.rows : []).map((row: any[]) => ({type: String(row?.[typeIndex] || "").toLowerCase(), text: String(row?.[textIndex] || "")}));
                const error = results.filter((item: any) => item.type === "error").map((item: any) => item.text).join("；");
                if (error) throw new Error(error);
                const warning = [data?.warning, ...results.filter((item: any) => item.type === "warning").map((item: any) => item.text)].filter(Boolean).join("；");
                const note = results.filter((item: any) => item.type === "note").map((item: any) => item.text).join("；");
                if (warning) message.warning(warning);
                else if (note && !results.some((item: any) => item.type === "status" && item.text.trim().toUpperCase() === "OK")) message.info(note);
                else message.success(data?.message || `${tool.title}完成`);
                onChange?.(tool.action);
            } catch (reason: any) {
                if (live.current && requestId === sequence.current) {
                    message.error(reason?.message || "操作失败，请检查数据库状态");
                }
            } finally { pending.current = false; if (live.current) setBusy(""); }
        };
        if (!tool.danger && ["check", "analyze"].includes(tool.action)) { void execute(); return; }
        confirmation.current?.destroy();
        const scope = `${schema ? `${schema}.` : ""}${table}`;
        confirmation.current = modal.confirm({title: `${tool.title}？`,
            content: <div style={{display: "grid", gap: 8, overflowWrap: "anywhere"}}><strong>{scope}</strong><span>{tool.description}{tool.danger ? "此操作不可撤销。" : "执行期间可能占用数据库资源。"}</span></div>,
            okText: tool.button, cancelText: "取消", okButtonProps: {danger: tool.danger},
            onOk: () => { void execute(); }});
    };
    const row = (tool: any) => <div className={`maintenance-row${tool.danger ? " maintenance-row-danger" : ""}`} key={tool.action}>
        <div><span className="maintenance-tool-title">{tool.title}</span><Typography.Text type="secondary">{tool.description}</Typography.Text></div>
        <Button color={tool.danger ? "danger" : "default"} variant="filled" loading={busy === tool.action}
            disabled={disabled || Boolean(busy) && busy !== tool.action} onClick={() => run(tool)}>{tool.button}</Button>
    </div>;

    return <div className="maintenance-panel"><div className="maintenance-list">{actions.map(row)}</div></div>;
};
