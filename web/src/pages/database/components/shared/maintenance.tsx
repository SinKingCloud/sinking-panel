import {useEffect, useRef, useState} from "react";
import {App, Button, Collapse, Empty, Typography} from "antd";
import {useDatabaseClient} from "./context";

const tools: Record<number, any[]> = {
    1: [
        {action: "save", title: "保存快照", description: "生成整个 Redis 实例的 RDB 备份快照。", button: "生成快照"},
        {action: "rewrite", title: "整理持久化日志", description: "重写整个实例的 AOF 文件，减少日志占用。", button: "整理日志"},
        {action: "flush", title: "清空当前数据库", description: "删除当前逻辑库的全部键，其他逻辑库不受影响。", button: "清空数据", danger: true},
    ],
    3: [
        {action: "analyze", title: "更新查询统计", description: "更新当前表的统计信息以帮助优化查询。", button: "更新统计"},
        {action: "optimize", title: "清理失效记录", description: "执行 VACUUM ANALYZE，清理失效记录并更新统计。", button: "开始清理"},
    ],
    4: [
        {action: "validate", title: "检查集合", description: "检查当前集合的文档与索引是否完整。", button: "开始检查"},
        {action: "compact", title: "整理集合空间", description: "整理集合存储；回收效果取决于版本和存储引擎。", button: "整理空间"},
        {action: "drop", title: "删除集合", description: "永久删除当前集合及其文档和索引。", button: "删除集合", danger: true},
    ],
    5: [
        {action: "check", title: "检查数据表", description: "检查表和索引的数据一致性。", button: "开始检查"},
        {action: "analyze", title: "更新查询统计", description: "更新当前表的统计信息以帮助优化查询。", button: "更新统计"},
        {action: "optimize", title: "重建表索引", description: "重建当前表的索引，减少索引碎片。", button: "重建索引"},
    ],
    6: [
        {action: "check", title: "检查数据完整性", description: "检查所选表及其索引是否存在损坏。", button: "开始检查"},
        {action: "analyze", title: "更新查询统计", description: "更新所选表的统计信息以帮助优化查询。", button: "更新统计"},
        {action: "optimize", title: "回收数据库空间", description: "整理整个 SQLite 文件，释放删除数据后留下的空闲空间。", button: "整理空间"},
    ],
};

export default ({record, schema, table, collection, isView, isMaterializedView = false, disabled = false, showTarget = false, collapsible = false, onChange}: any) => {
    const {operate} = useDatabaseClient();
    const {message, modal} = App.useApp();
    const [busy, setBusy] = useState("");
    const pending = useRef(false);
    const live = useRef(true);
    const sequence = useRef(0);
    const confirmation = useRef<any>(undefined);
    const type = Number(record.type);
    const target = collection || table;
    const actions = [...(tools[type] || []), ...([3, 5, 6].includes(type) ? [
        {action: "truncate", title: "清空数据表", description: "删除全部记录，保留表结构。", button: "清空数据", danger: true},
        {action: "drop", title: isMaterializedView ? "删除物化视图" : isView ? "删除视图" : "删除数据表", description: isView ? "删除当前视图，保留原始数据表。" : "永久删除当前表及其数据和索引。", button: isView ? "删除视图" : "删除表", danger: true},
    ] : [])].filter((tool) => isMaterializedView ? tool.action !== "truncate" : !isView || tool.action === "drop");

    useEffect(() => { live.current = true; return () => { live.current = false; sequence.current++; confirmation.current?.destroy(); }; }, []);
    useEffect(() => {
        sequence.current++; confirmation.current?.destroy();
    }, [record.id, schema, table, collection]);

    const run = (tool: any) => {
        if (disabled) return;
        const requestId = sequence.current;
        const execute = async () => {
            if (pending.current || !live.current || requestId !== sequence.current) return;
            pending.current = true; setBusy(tool.action);
            try {
                const data = await operate(record, "maintenance", {action: tool.action, schema, table, collection});
                if (!live.current || requestId !== sequence.current) return;
                if (data?.warning) message.warning(data.warning);
                else message.success(data?.message || `${tool.title}已完成`);
                onChange?.(tool.action);
            } catch (reason: any) {
                if (live.current && requestId === sequence.current) {
                    const description = reason?.message || "操作失败，请检查数据库状态";
                    message.error(description);
                }
                throw reason;
            } finally { pending.current = false; if (live.current) setBusy(""); }
        };
        if (!tool.danger && ["check", "validate", "analyze"].includes(tool.action)) { void execute().catch(() => {}); return; }
        confirmation.current?.destroy();
        const scope = type === 1 && tool.action !== "flush" ? `实例 ${record.server_name || record.name}` : target || `DB ${record.name}`;
        confirmation.current = modal.confirm({title: `${tool.title}？`,
            content: <div style={{display: "grid", gap: 8, overflowWrap: "anywhere"}}><strong>{scope}</strong><span>{tool.description}{tool.danger ? "此操作不可撤销。" : "执行期间可能占用数据库资源。"}</span></div>,
            okText: tool.button, cancelText: "取消", okButtonProps: {danger: tool.danger},
            onOk: execute});
        void Promise.resolve(confirmation.current);
    };
    const row = (tool: any) => <div className={`maintenance-row${tool.danger ? " maintenance-row-danger" : ""}`} key={tool.action}>
        <div><span className="maintenance-tool-title">{tool.title}</span><Typography.Text type="secondary">{tool.description}</Typography.Text></div>
        <Button color={tool.danger ? "danger" : "default"} variant="filled" loading={busy === tool.action}
            disabled={disabled || Boolean(busy) && busy !== tool.action} onClick={() => run(tool)}>{tool.button}</Button>
    </div>;

    if (type !== 1 && !target) return <div className="client-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={type === 4 ? "先在文档页选择集合，再使用维护工具" : "先在数据页选择数据表，再使用维护工具"}/></div>;
    return <div className="maintenance-panel">
        {showTarget && <div className="maintenance-target"><span>{type === 1 ? "当前逻辑库" : type === 4 ? "当前集合" : isMaterializedView ? "当前物化视图" : isView ? "当前视图" : "当前数据表"}</span>
            <strong>{type === 1 ? `DB ${record.name}` : `${schema ? `${schema}.` : ""}${target}`}</strong></div>}
        {(!collapsible || actions.some((tool) => !tool.danger && !tool.advanced)) && <div className="maintenance-list">{actions.filter((tool) => !collapsible || !tool.danger && !tool.advanced).map(row)}</div>}
        {collapsible && actions.some((tool) => tool.danger || tool.advanced) && <Collapse ghost className="maintenance-more" items={[{
            key: "advanced", label: "其他维护与数据清理", children: <div className="maintenance-list">{actions.filter((tool) => tool.danger || tool.advanced).map(row)}</div>,
        }]}/>}

    </div>;
};
