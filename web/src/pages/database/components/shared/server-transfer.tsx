import {useEffect, useRef, useState} from "react";
import {App, Button, ConfigProvider, Form, Input, Space, Typography, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {useDatabaseClient} from "./context";
import useStyles from "./styles/client";

// 标准 SQL / BSON 文件只在服务器处理，浏览器仅提交路径和目标对象。
export default ({record, schema, table, collection, disabled, readOnly, exportOnly = false, dropdown = false, exportDatabase = false, onChange}: any) => {
    const {request, FilePicker, ActionDropdown} = useDatabaseClient();
    const {message, modal} = App.useApp();
    const {styles} = useStyles();
    const {token} = antdTheme.useToken();
    const theme = useTheme();
    const [form] = Form.useForm();
    const [dialog, setDialog] = useState<any>();
    const [lastDialog, setLastDialog] = useState<any>();
    const displayDialog = dialog || lastDialog;
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const [unconfirmed, setUnconfirmed] = useState(false);
    const [taskId, setTaskId] = useState("");
    const [tasks, setTasks] = useState<any[]>([]);
    const pending = useRef(false);
    const sequence = useRef(0);
    const live = useRef(true);
    const confirmation = useRef<any>(undefined);
    const accepted = useRef(-1);
    const tracked = useRef<any[]>([]);
    const currentTarget = useRef<any>(undefined);
    currentTarget.current = {record, collection, onChange};
    tracked.current = tasks;
    const task = tasks.find((item) => item.id === taskId);
    const result = task?.result;
    const active = tasks.some((item) => item.status < 2 || item.settling);
    const statusLabels = ["排队中", "处理中", "已完成", "失败", "已取消"];
    const mongo = Number(record.type) === 4;
    const format = mongo ? "BSON" : "SQL";
    useEffect(() => { live.current = true; return () => { live.current = false; sequence.current++; confirmation.current?.destroy(); }; }, []);
    // 这里只跟踪本面板创建的任务；关闭表单不取消任务，离开面板则停止查询。
    useEffect(() => {
        if (!active) return;
        let stopped = false, timer = 0;
        const poll = async () => {
            const pendingTasks = tracked.current.filter((item) => item.status < 2 || item.settling);
            const responses = await Promise.all(pendingTasks.map(async (item) => {
                try { return {item, response: await request("system/task", {action: "info", id: item.id})}; }
                catch { return {item, response: undefined}; }
            }));
            if (stopped) return;
            let next = tracked.current.slice();
            for (const {item, response} of responses) {
                const index = next.findIndex((value) => value.id === item.id);
                if (index < 0 || (next[index].status >= 2 && !next[index].settling)) continue;
                if (item.settling) next[index] = {...next[index], settling: false};
                const valid = response?.code === 200 && response.data && [0, 1, 2, 3, 4].includes(Number(response.data.status));
                if (valid && response) {
                    const status = item.status === 4 ? 4 : Number(response.data.status);
                    // 取消先改变状态，worker 稍后才写入结果；无结果时只补读一次。
                    next[index] = {...item, ...response.data, id: item.id, body: item.body, status, missing: 0,
                        settling: status === 4 && response.data.result == null && item.status < 2};
                } else if (response && String(response.message || "").includes("任务不存在")) {
                    const missing = (item.missing || 0) + 1;
                    next[index] = {...next[index], missing, ...(missing >= 3 ? {status: 4, message: "任务状态不可用，请在系统任务中核对结果"} : {})};
                }
                const updated = next[index], current = currentTarget.current;
                if ((valid || item.settling) && updated.status >= 2 && !updated.settling && item.body.action === "import" && String(current.record.id) === String(item.body.id)
                    && Number(current.record.type) === Number(item.body.type) && (current.record.database_name || current.record.name) === item.body.database_name && (!item.body.collection || current.collection === item.body.collection)) current.onChange?.();
            }
            tracked.current = next; setTasks(next);
            if (next.some((item) => item.status < 2 || item.settling)) timer = window.setTimeout(poll, 2000);
        };
        void poll();
        return () => { stopped = true; window.clearTimeout(timer); };
    }, [active, request, record.id, record.type, record.database_name, record.name]);
    useEffect(() => {
        sequence.current++; confirmation.current?.destroy();
        tracked.current = []; setTasks([]); setTaskId(""); setDialog(undefined); setError(""); setUnconfirmed(false);
    }, [record.id, record.type, record.database_name, record.name]);
    const open = (mode: string, wholeDatabase = false) => {
        if (pending.current) return;
        sequence.current++; setTaskId(""); setUnconfirmed(false); setError("");
        setTasks((items) => items.filter((item) => item.status < 2 || item.settling));
        form.resetFields();
        const targetTable = wholeDatabase ? undefined : table;
        const name = `${(targetTable || collection || record.name).replace(/[\\/:*?"<>|]/g, "_")}.${mongo ? "bson" : "sql"}`;
        form.setFieldsValue({name});
        const next = {mode, record: {...record}, schema, table: targetTable, collection};
        setLastDialog(next); setDialog(next);
    };
    const execute = async (values: any, current: any, requestId: number) => {
        if (pending.current || !live.current || sequence.current !== requestId || accepted.current === requestId) return;
        const directory = String(values.directory || "");
        const separator = directory.includes("\\") && !directory.includes("/") ? "\\" : "/";
        const path = current.mode === "import" ? values.path : `${directory.replace(/[\\/]+$/, "")}${separator}${values.name}`;
        const body = {id: current.record.id, database_name: current.record.database_name || current.record.name, path,
            ...(Number(current.record.type) === 4 ? {collection: current.collection} : current.mode === "export"
                ? {...(current.schema ? {schema: current.schema} : {}), ...(current.table ? {table: current.table} : {})} : {})};
        const taskBody = {...body, type: Number(current.record.type), action: current.mode};
        const existing = tracked.current.find((item) => (item.status < 2 || item.settling) && JSON.stringify(item.body) === JSON.stringify(taskBody));
        if (existing) { accepted.current = requestId; setTaskId(existing.id); message.info("该任务正在后台执行"); return; }
        pending.current = true; setBusy(true); setError(""); setUnconfirmed(false);
        try {
            const response = await request(current.mode, body);
            if (!live.current || String(currentTarget.current.record.id) !== String(current.record.id) || Number(currentTarget.current.record.type) !== Number(current.record.type) || (currentTarget.current.record.database_name || currentTarget.current.record.name) !== (current.record.database_name || current.record.name)) return;
            if (response?.code !== 200 || typeof response.data !== "string" || !response.data) {
                const failure: any = new Error(response?.message || "任务创建结果未确认，请在系统任务中核对后再操作");
                failure.unconfirmed = !response || response.code === 200;
                throw failure;
            }
            accepted.current = requestId;
            const created = {id: response.data, status: 0, progress: 0, message: "等待任务开始", body: taskBody};
            tracked.current = [...tracked.current, created]; setTasks(tracked.current);
            if (sequence.current === requestId) setTaskId(created.id);
            message.success("已加入后台任务");
        } catch (reason: any) {
            if (live.current && sequence.current === requestId) {
                setError(reason.message || "任务创建结果未确认，请在系统任务中核对后再操作");
                setUnconfirmed(reason.unconfirmed !== false);
                if (reason.unconfirmed !== false) accepted.current = requestId;
            }
        } finally { pending.current = false; if (live.current) setBusy(false); }
    };
    const submit = (values: any) => {
        if (!dialog || taskId || unconfirmed || pending.current) return;
        const current = dialog, requestId = sequence.current, submitted = {...values};
        if (current.mode === "export") { void execute(submitted, current, requestId); return; }
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: `确认导入 ${format} 文件？`,
            content: mongo ? "文件中的文档将写入当前集合，已成功写入的内容会保留。" : "SQL 文件将在当前数据库执行，可能创建、修改或删除数据，已成功执行的语句可能无法撤销。",
            okText: "导入", cancelText: "取消", onOk: () => execute(submitted, current, requestId)});
    };
    const close = () => { sequence.current++; confirmation.current?.destroy(); setDialog(undefined); };
    return <>
        {dropdown ? !disabled && !busy && (!mongo || collection) && <ActionDropdown arrow={false} menu={{items: [
            ...(!exportOnly && !readOnly ? [{key: "import", label: `导入 ${format}`, onClick: () => open("import")}] : []),
            {key: "export", label: mongo ? "导出 BSON" : table ? "导出此表" : "导出数据库", onClick: () => open("export")},
            ...(exportDatabase && table && !mongo ? [{key: "export_database", label: "导出数据库", onClick: () => open("export", true)}] : []),
        ]}}><Button className="client-toolbar-dropdown" color="default" variant="filled" icon={<Icon type="SwapOutlined"/>}>
            导入导出<Icon type="DownOutlined" className="client-toolbar-arrow"/>
        </Button></ActionDropdown> : <Space size={8} wrap>
            {!exportOnly && <Button color="default" variant="filled" disabled={disabled || readOnly || busy || (mongo && !collection)} onClick={() => open("import")}>导入 {format}</Button>}
            <Button color="default" variant="filled" disabled={disabled || busy || (mongo && !collection)} onClick={() => open("export")}>{mongo ? "导出 BSON" : table ? "导出此表" : "导出数据库"}</Button>
        </Space>}
        <ProModal title={<Title>{displayDialog?.mode === "import" ? `导入 ${format}` : `导出 ${format}`}</Title>}
            onCancel={close} onOk={form.submit} modalProps={{
                open: Boolean(dialog), width: theme?.isCompactTheme?.() ? 340 : 360,
                className: styles.editor, style: {maxWidth: "calc(100vw - 24px)"}, forceRender: true,
                afterClose: () => setLastDialog(undefined),
                okText: displayDialog?.mode === "import" ? "导入" : "导出", cancelText: "取消", confirmLoading: busy, mask: {closable: true},
                footer: taskId || unconfirmed ? <Button onClick={close}>关闭</Button> : undefined,
            }}>
            <ConfigProvider theme={{components: {Input: {controlHeight: token.controlHeight}}}}>
            <Form form={form} layout="vertical" variant="filled" onFinish={submit} disabled={busy || Boolean(taskId) || unconfirmed}>
                {displayDialog && (displayDialog.mode === "import" ? <Form.Item name="path" label={`${format} 文件`} rules={[{required: true, whitespace: true, message: "请选择服务器上的文件"}]}>
                    <FilePicker mode="file" placeholder={`请选择或输入 ${format} 文件路径`} buttonProps={{color: "default", variant: "filled"}}/>
                </Form.Item> : <>
                    <Form.Item name="directory" label="服务器目录" rules={[{required: true, whitespace: true, message: "请选择保存目录"}]}><FilePicker mode="directory" placeholder="请选择或输入保存目录" buttonProps={{color: "default", variant: "filled"}}/></Form.Item>
                    <Form.Item name="name" label="文件名称" rules={[{required: true, whitespace: true, max: 255, message: "请输入文件名称"},
                        {pattern: /^(?!\.{1,2}$)[^/\\\x00]+$/, message: "文件名称不能包含路径"}]}><Input placeholder="请输入文件名称"/></Form.Item>
                </>)}
            </Form>
            </ConfigProvider>
            {displayDialog && task && <Typography.Text type={task.status === 3 ? "danger" : "secondary"} className="client-hint">
                {statusLabels[task.status]}{task.status < 2 ? ` · ${Math.max(0, Math.min(100, Number(task.progress) || 0)).toFixed(0)}%` : ""}
                {task.message ? ` · ${task.message}` : ""}{task.status < 2 ? "，可关闭窗口，进度和日志可在系统任务中查看。" : ""}
            </Typography.Text>}
            {displayDialog && error && <Typography.Text type="danger" className="client-hint">{error}</Typography.Text>}
            {displayDialog && result?.error && result.error !== task.message && <Typography.Text type="danger" className="client-hint">{result.error}</Typography.Text>}
            {displayDialog && result && <div className="transfer-result">
                <Typography.Text copyable>{result.path}</Typography.Text>
                {typeof result.success === "number" && <span>已处理 <b>{result.success}</b></span>}
                {result.failed > 0 && <span>失败 <b>{result.failed}</b></span>}
                {Boolean(result.unknown) && <span>部分执行结果待核对</span>}
                {result.line > 0 && <span>失败位置：第 {result.line} 行</span>}
                {typeof result.size === "number" && <span>{(result.size / 1024 / 1024).toFixed(2)} MB</span>}
            </div>}
        </ProModal>
    </>;
};
