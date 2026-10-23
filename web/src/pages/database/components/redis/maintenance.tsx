import {useEffect, useRef, useState} from "react";
import {App, Button, Typography} from "antd";
import {useTheme} from "sinking-antd";
import {useRedisClient} from "./context";
import useStyles from "./maintenance-styles";

const tools = [
    {action: "save", title: "保存快照", description: "生成整个 Redis 实例的 RDB 快照。", button: "生成快照", danger: false},
    {action: "rewrite", title: "整理持久化日志", description: "重写整个实例的 AOF 文件，减少日志占用。", button: "整理日志", danger: false},
    {action: "flush", title: "清空当前数据库", description: "删除当前数据库的全部键，其他数据库不受影响。", button: "清空数据", danger: true},
];

export default ({record, onChange}: any) => {
    const {api} = useRedisClient();
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const [busy, setBusy] = useState("");
    const live = useRef(true);
    const sequence = useRef(0);
    const pending = useRef(0);
    const confirmation = useRef<any>(undefined);
    const scope = useRef(record.id);
    scope.current = record.id;

    useEffect(() => {
        live.current = true;
        return () => { live.current = false; sequence.current++; confirmation.current?.destroy(); };
    }, []);
    useEffect(() => {
        sequence.current++; pending.current = 0;
        confirmation.current?.destroy(); confirmation.current = undefined;
        setBusy("");
    }, [record.id]);

    const run = (tool: typeof tools[number]) => {
        if (pending.current || confirmation.current) return;
        const target = record.id;
        const revision = sequence.current;
        const dialog = modal.confirm({
            title: `${tool.title}？`,
            content: <div className={styles.confirmation}>
                <span>{tool.action === "flush" ? `当前数据库：DB ${record.name}` : `当前实例：${record.server_name || record.name}`}</span>
                <span>{tool.description}{tool.danger ? "此操作不可撤销。" : "执行期间可能占用数据库资源。"}</span>
            </div>,
            okText: tool.button, cancelText: "取消", okButtonProps: {danger: tool.danger},
            onOk: async () => {
                if (pending.current || !live.current || scope.current !== target || sequence.current !== revision) return;
                const request = ++sequence.current;
                pending.current = request;
                setBusy(tool.action);
                try {
                    const data = await api.maintenance(record, {action: tool.action});
                    if (!live.current || scope.current !== target || sequence.current !== request) return;
                    const text = data?.warning || data?.message || `${tool.title}已完成`;
                    if (data?.warning) message.warning(text); else message.success(text);
                    onChange?.(tool.action);
                } catch (reason: any) {
                    if (live.current && scope.current === target && sequence.current === request) {
                        message.error(reason?.message || "操作失败，请重试");
                    }
                } finally {
                    if (pending.current === request) {
                        pending.current = 0;
                        if (live.current && scope.current === target) setBusy("");
                    }
                }
            },
            afterClose: () => { if (confirmation.current === dialog) confirmation.current = undefined; },
        });
        confirmation.current = dialog;
    };

    return <div className={styles.maintenance}>
        <div className="redis-maintenance-list">{tools.map((tool) => <div className="redis-maintenance-row" key={tool.action}>
            <div>
                <span className="redis-maintenance-title">{tool.title}</span>
                <Typography.Text type="secondary">{tool.description}</Typography.Text>
            </div>
            <Button color={tool.danger ? "danger" : "default"} variant="filled" loading={busy === tool.action}
                disabled={Boolean(busy) && busy !== tool.action} onClick={() => run(tool)}>{tool.button}</Button>
        </div>)}</div>
    </div>;
};
