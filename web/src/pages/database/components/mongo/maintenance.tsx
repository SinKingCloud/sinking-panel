import {useEffect, useRef, useState} from "react";
import {App, Button, Empty} from "antd";
import {useDatabaseClient} from "../shared/context";

export default ({record, collection, onChange}: any) => {
    const {operate} = useDatabaseClient();
    const {message, modal} = App.useApp();
    const [busy, setBusy] = useState("");
    const live = useRef(true);
    const pending = useRef(false);
    const confirmation = useRef<any>(undefined);
    useEffect(() => { live.current = true; return () => { live.current = false; confirmation.current?.destroy(); }; }, []);
    if (!collection) return <div className="manager-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择集合"/></div>;
    return <div className="maintenance-list">{[
        {action: "validate", name: "检查集合", description: "检查文档与索引的数据完整性。", button: "检查"},
        {action: "compact", name: "整理集合空间", description: "整理集合存储，回收效果取决于存储引擎。", button: "整理"},
        {action: "drop", name: "删除集合", description: "永久删除当前集合及其文档和索引。", button: "删除", danger: true},
    ].map((tool) => <div className="maintenance-row" key={tool.action}><div><strong>{tool.name}</strong><span>{tool.description}</span></div>
        <Button color={tool.danger ? "danger" : "default"} variant="filled" loading={busy === tool.action} disabled={Boolean(busy) && busy !== tool.action} onClick={() => {
            const execute = async () => {
                if (pending.current) return;
                pending.current = true; setBusy(tool.action);
                try {
                    const result = await operate(record, "maintenance", {action: tool.action, collection});
                    if (!live.current) return;
                    const native = result?.result || result;
                    if (native?.valid === false) message.warning("检查完成，集合存在异常，请检查数据库日志");
                    else message.success(result?.message || `${tool.name}已完成`);
                    onChange?.(tool.action);
                } catch (reason: any) { if (live.current) message.error(reason.message || "操作失败"); }
                finally { pending.current = false; if (live.current) setBusy(""); }
            };
            if (tool.action === "validate") { void execute(); return; }
            confirmation.current = modal.confirm({title: `${tool.name}？`, content: `集合：${collection}。${tool.description}`, okText: tool.button, cancelText: "取消", okButtonProps: {danger: tool.danger}, onOk: execute});
        }}>{tool.button}</Button></div>)}</div>;
};
