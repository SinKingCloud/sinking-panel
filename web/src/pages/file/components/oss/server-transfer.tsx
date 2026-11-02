import {forwardRef, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Button, Form, Input, Progress, Typography} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import FilePicker from "@/pages/components/file-picker";
import {comparableFilePath, isAbsoluteFilePath, joinFilePath, normalizeFilePath} from "@/pages/file/utils";
import {ossDownload, ossUpload} from "@/service/api/oss";
import {systemTask} from "@/service/api/system";
import useStyles from "./styles";

// ServerTransfer 只提交服务器文件路径，传输和取消由系统任务管理。
const ServerTransfer = forwardRef<any, any>(({record, active: visible = true, onChange, onBusyChange}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const [form] = Form.useForm();
    const [dialog, setDialog] = useState<{action: "upload" | "download"; key: string}>();
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const [unconfirmed, setUnconfirmed] = useState(false);
    const [taskIds, setTaskIds] = useState<string[]>([]);
    const [tasks, setTasks] = useState<any[]>([]);
    const live = useRef(true);
    const pending = useRef(false);
    const modalWidth = useRef(420);
    const tracked = useRef<any[]>([]);
    const changed = useRef(onChange);
    const busyChanged = useRef(onBusyChange);
    changed.current = onChange;
    busyChanged.current = onBusyChange;
    tracked.current = tasks;
    const currentTasks = tasks.filter((item) => taskIds.includes(item.id));
    const active = tasks.some((item) => item.status < 2);
    const statusLabels = ["排队中", "处理中", "已完成", "失败", "已取消"];

    useEffect(() => { live.current = true; return () => { live.current = false; busyChanged.current?.(false); }; }, []);
    useEffect(() => { onBusyChange?.(busy || active); }, [active, busy, onBusyChange]);
    useImperativeHandle(ref, () => ({
        open: (action: "upload" | "download", key: string) => {
            if (pending.current) return;
            modalWidth.current = action === "upload" ? 380 : 420;
            form.resetFields();
            if (action === "download") form.setFieldsValue({name: key.split("/").pop()});
            setTaskIds([]); setError(""); setUnconfirmed(false);
            setTasks((items) => items.filter((item) => item.status < 2));
            setDialog({action, key});
        },
    }), [form]);

    // 关闭表单后继续查询已提交的任务，离开 OSS 页面仅停止查询，不取消传输。
    useEffect(() => {
        if (!active) return;
        let stopped = false, timer = 0;
        const poll = async () => {
            const responses = await Promise.all(tracked.current.filter((item) => item.status < 2).map(async (item) => {
                try { return {item, response: await systemTask({body: {id: item.id, action: "info"}})}; }
                catch { return {item, response: undefined}; }
            }));
            if (stopped) return;
            const next = tracked.current.slice();
            let uploaded = false;
            for (const {item, response} of responses) {
                const index = next.findIndex((value) => value.id === item.id);
                if (index < 0 || next[index].status >= 2) continue;
                if (response?.code === 200 && [0, 1, 2, 3, 4].includes(Number(response.data?.status))) {
                    next[index] = {...next[index], ...response.data, id: item.id, action: item.action, status: Number(response.data.status), missing: 0};
                    if (next[index].status === 2 && item.action === "upload") uploaded = true;
                } else if (response && String(response.message || "").includes("任务不存在")) {
                    const missing = (next[index].missing || 0) + 1;
                    next[index] = {...next[index], missing, ...(missing >= 3 ? {status: 4, message: "任务状态不可用，请在系统任务中核对结果"} : {})};
                }
            }
            tracked.current = next; setTasks(next);
            if (uploaded) changed.current?.();
            if (next.some((item) => item.status < 2)) timer = window.setTimeout(poll, 1500);
        };
        void poll();
        return () => { stopped = true; window.clearTimeout(timer); };
    }, [active]);

    const submit = async (values: any) => {
        if (!dialog || pending.current || taskIds.length || unconfirmed) return;
        const uploading = dialog.action === "upload";
        const paths = uploading
            ? [...new Map((values.paths as string[]).map((path) => [comparableFilePath(path), path])).values()]
            : [joinFilePath(values.directory, values.name)];
        const createdIds: string[] = [];
        pending.current = true; setBusy(true); setError("");
        try {
            // 顺序入队，出现失败即停止；已创建的任务继续执行，避免重复提交整批文件。
            for (const [index, path] of paths.entries()) {
                if (!live.current) return;
                const key = uploading ? dialog.key + normalizeFilePath(path).split("/").pop() : dialog.key;
                let response;
                try {
                    response = await (uploading ? ossUpload : ossDownload)({body: {
                        id: record.id, action: "server", key, path,
                        ...(!uploading ? {overwrite: true} : {}),
                    }});
                } catch { /* 无响应时无法确认服务端是否已创建任务，不自动重试。 */ }
                if (!live.current) return;
                if (response?.code !== 200 || typeof response.data !== "string" || !response.data) {
                    const unknown = !response || response.code === 200;
                    setUnconfirmed(unknown);
                    const reason = unknown ? "任务提交结果未确认，请在系统任务中核对后再操作" : response.message || "任务创建失败";
                    setError(paths.length > 1 ? `${path}：${reason}。已提交 ${createdIds.length} 个文件，后续 ${paths.length - index - 1} 个文件未提交。` : reason);
                    return;
                }
                const created = {id: response.data, action: dialog.action, path, status: 0, progress: 0, message: "等待任务开始"};
                tracked.current = [...tracked.current, created]; setTasks(tracked.current);
                createdIds.push(created.id); setTaskIds([...createdIds]);
            }
            message.success(paths.length > 1 ? `已加入 ${createdIds.length} 个后台任务` : "已加入后台任务");
        } finally { pending.current = false; if (live.current) setBusy(false); }
    };

    return <ProModal title={<Title>{dialog?.action === "upload" ? "从服务器上传" : "下载到服务器"}</Title>}
        onCancel={() => { if (!pending.current) setDialog(undefined); }} onOk={form.submit}
        modalProps={{open: visible && Boolean(dialog), width: modalWidth.current, zIndex: 1140,
            style: {maxWidth: "calc(100vw - 24px)"}, forceRender: true,
            okText: dialog?.action === "upload" ? "上传" : "下载", cancelText: "取消", confirmLoading: busy,
            closable: !busy, keyboard: !busy, mask: {closable: !busy}, cancelButtonProps: {disabled: busy},
            footer: taskIds.length || unconfirmed ? <Button disabled={busy} onClick={() => setDialog(undefined)}>关闭</Button> : undefined}}>
        {visible && <Form form={form} layout="vertical" variant="filled" disabled={busy || Boolean(taskIds.length) || unconfirmed} onFinish={submit}>
            {dialog?.action === "upload" ? <>
                <Form.Item name="paths" label="服务器文件" rules={[
                    {required: true, type: "array", min: 1, message: "请选择服务器上的文件"},
                    {validator: async (_, paths?: string[]) => {
                        const names = new Map<string, string>();
                        for (const path of paths || []) {
                            if (typeof path !== "string" || !isAbsoluteFilePath(path) || path.length > 4096) throw new Error("请输入有效的文件绝对路径，长度不能超过 4096 个字符");
                            const name = normalizeFilePath(path).split("/").pop()!;
                            if (names.has(name) && names.get(name) !== comparableFilePath(path)) throw new Error(`所选文件包含同名文件 ${name}，请分批上传`);
                            names.set(name, comparableFilePath(path));
                        }
                    }},
                ]}><FilePicker mode="file" multiple maxLength={4096} placeholder="请选择服务器上的文件" buttonProps={{color: "default", variant: "filled"}}/></Form.Item>
            </> : <>
                <Form.Item name="directory" label="服务器目录" rules={[
                    {required: true, whitespace: true, max: 4096, message: "请选择保存目录"},
                    {validator: async (_, value) => { if (value && !isAbsoluteFilePath(value)) throw new Error("请输入目录的绝对路径"); }},
                ]}><FilePicker mode="directory" maxLength={4096} placeholder="请选择或输入保存目录" buttonProps={{color: "default", variant: "filled"}}/></Form.Item>
                <Form.Item name="name" label="文件名称" rules={[{required: true, whitespace: true, max: 255, message: "请输入文件名称"},
                    {pattern: /^(?!\.{1,2}$)[^/\\\x00]+$/, message: "文件名称不能包含路径"}]}><Input maxLength={255} placeholder="请输入文件名称"/></Form.Item>
            </>}
        </Form>}
        {currentTasks.length > 0 && <>
            <div className={styles.transferTasks}>{currentTasks.map((task) => <div key={task.id}>
                {currentTasks.length > 1 && <Typography.Text ellipsis title={task.path} className={styles.hint}>{normalizeFilePath(task.path).split("/").pop()}</Typography.Text>}
                <Typography.Text type={task.status === 3 || task.status === 4 ? "danger" : "secondary"} className={styles.hint}>{statusLabels[task.status]}{task.message ? ` · ${task.message}` : ""}</Typography.Text>
                <Progress size="small" percent={Math.max(0, Math.min(100, Number(task.progress) || 0))} status={task.status === 2 ? "success" : task.status >= 3 ? "exception" : "active"}/>
                {task.result?.path && currentTasks.length === 1 && <Typography.Text type="secondary" className={styles.hint}>{task.action === "upload" ? `文件已上传：${task.result.key}` : `文件已保存：${task.result.path}`}</Typography.Text>}
            </div>)}</div>
            <Typography.Text type="secondary" className={styles.hint}>可关闭窗口，进度、日志和取消操作可在系统任务中查看。</Typography.Text>
        </>}
        {error && <Typography.Text type="danger" className={styles.hint}>{error}</Typography.Text>}
    </ProModal>;
});

export default ServerTransfer;
