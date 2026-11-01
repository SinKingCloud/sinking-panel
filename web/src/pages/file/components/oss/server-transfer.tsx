import {forwardRef, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Button, Checkbox, Form, Input, Progress, Typography} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import FilePicker from "@/pages/components/file-picker";
import {isAbsoluteFilePath, joinFilePath, normalizeFilePath} from "@/pages/file/utils";
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
    const [taskId, setTaskId] = useState("");
    const [tasks, setTasks] = useState<any[]>([]);
    const live = useRef(true);
    const pending = useRef(false);
    const tracked = useRef<any[]>([]);
    const changed = useRef(onChange);
    const busyChanged = useRef(onBusyChange);
    changed.current = onChange;
    busyChanged.current = onBusyChange;
    tracked.current = tasks;
    const task = tasks.find((item) => item.id === taskId);
    const active = tasks.some((item) => item.status < 2);
    const statusLabels = ["排队中", "处理中", "已完成", "失败", "已取消"];

    useEffect(() => { live.current = true; return () => { live.current = false; busyChanged.current?.(false); }; }, []);
    useEffect(() => { onBusyChange?.(busy || active); }, [active, busy, onBusyChange]);
    useImperativeHandle(ref, () => ({
        open: (action: "upload" | "download", key: string) => {
            if (pending.current) return;
            form.resetFields();
            if (action === "download") form.setFieldsValue({key, name: key.split("/").pop(), overwrite: false});
            setTaskId(""); setError(""); setUnconfirmed(false);
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
            for (const {item, response} of responses) {
                const index = next.findIndex((value) => value.id === item.id);
                if (index < 0 || next[index].status >= 2) continue;
                if (response?.code === 200 && [0, 1, 2, 3, 4].includes(Number(response.data?.status))) {
                    next[index] = {...next[index], ...response.data, id: item.id, action: item.action, status: Number(response.data.status), missing: 0};
                    if (next[index].status === 2 && item.action === "upload") changed.current?.();
                } else if (response && String(response.message || "").includes("任务不存在")) {
                    const missing = (next[index].missing || 0) + 1;
                    next[index] = {...next[index], missing, ...(missing >= 3 ? {status: 4, message: "任务状态不可用，请在系统任务中核对结果"} : {})};
                }
            }
            tracked.current = next; setTasks(next);
            if (next.some((item) => item.status < 2)) timer = window.setTimeout(poll, 1500);
        };
        void poll();
        return () => { stopped = true; window.clearTimeout(timer); };
    }, [active]);

    const submit = async (values: any) => {
        if (!dialog || pending.current || taskId || unconfirmed) return;
        const uploading = dialog.action === "upload";
        const path = uploading ? values.path : joinFilePath(values.directory, values.name);
        const key = uploading ? dialog.key + normalizeFilePath(path).split("/").pop() : values.key;
        pending.current = true; setBusy(true); setError("");
        try {
            const response = await (uploading ? ossUpload : ossDownload)({body: {
                id: record.id, action: "server", key, path,
                ...(!uploading ? {overwrite: Boolean(values.overwrite)} : {}),
            }});
            if (!live.current) return;
            if (response?.code !== 200 || typeof response.data !== "string" || !response.data) {
                const unknown = !response || response.code === 200;
                setUnconfirmed(unknown);
                setError(unknown ? "任务提交结果未确认，请在系统任务中核对后再操作" : response.message || "任务创建失败");
                return;
            }
            const created = {id: response.data, action: dialog.action, path, status: 0, progress: 0, message: "等待任务开始"};
            tracked.current = [...tracked.current, created]; setTasks(tracked.current);
            setTaskId(created.id);
            message.success("已加入后台任务");
        } catch {
            if (live.current) { setUnconfirmed(true); setError("任务提交结果未确认，请在系统任务中核对后再操作"); }
        } finally { pending.current = false; if (live.current) setBusy(false); }
    };

    return <ProModal title={<Title>{dialog?.action === "upload" ? "从服务器上传" : "下载到服务器"}</Title>}
        onCancel={() => { if (!pending.current) setDialog(undefined); }} onOk={form.submit}
        modalProps={{open: visible && Boolean(dialog), width: 460, zIndex: 1140,
            style: {maxWidth: "calc(100vw - 24px)"}, forceRender: true,
            okText: dialog?.action === "upload" ? "上传" : "下载", cancelText: "取消", confirmLoading: busy,
            closable: !busy, keyboard: !busy, mask: {closable: !busy}, cancelButtonProps: {disabled: busy},
            footer: taskId || unconfirmed ? <Button onClick={() => setDialog(undefined)}>关闭</Button> : undefined}}>
        {visible && <Form form={form} layout="vertical" variant="filled" disabled={busy || Boolean(taskId) || unconfirmed} onFinish={submit}>
            {dialog?.action === "upload" ? <>
                <Form.Item name="path" label="服务器文件" rules={[
                    {required: true, whitespace: true, max: 4096, message: "请选择或输入服务器文件路径"},
                    {validator: async (_, value) => { if (value && !isAbsoluteFilePath(value)) throw new Error("请输入文件的绝对路径"); }},
                ]}><FilePicker mode="file" maxLength={4096} placeholder="请选择或输入服务器上的文件" buttonProps={{color: "default", variant: "filled"}}/></Form.Item>
                <Typography.Text type="secondary" className={styles.hint}>上传目录：{dialog.key || "/"}，同名文件将直接覆盖。</Typography.Text>
            </> : <>
                <Form.Item name="directory" label="服务器目录" rules={[
                    {required: true, whitespace: true, max: 4096, message: "请选择保存目录"},
                    {validator: async (_, value) => { if (value && !isAbsoluteFilePath(value)) throw new Error("请输入目录的绝对路径"); }},
                ]}><FilePicker mode="directory" maxLength={4096} placeholder="请选择或输入保存目录" buttonProps={{color: "default", variant: "filled"}}/></Form.Item>
                <Form.Item name="name" label="文件名称" rules={[{required: true, whitespace: true, max: 255, message: "请输入文件名称"},
                    {pattern: /^(?!\.{1,2}$)[^/\\\x00]+$/, message: "文件名称不能包含路径"}]}><Input maxLength={255} placeholder="请输入文件名称"/></Form.Item>
                <Form.Item name="key" label="OSS 对象路径" rules={[
                    {required: true, whitespace: true, max: 1024, message: "请输入完整对象路径"},
                    {validator: async (_, value) => { if (value?.endsWith("/")) throw new Error("对象路径必须包含文件名称"); }},
                ]}><Input maxLength={1024} readOnly placeholder="包含目录和文件名"/></Form.Item>
                <Form.Item name="overwrite" valuePropName="checked"><Checkbox>覆盖服务器上的同名文件</Checkbox></Form.Item>
            </>}
        </Form>}
        {task && <>
            <Typography.Text type={task.status === 3 || task.status === 4 ? "danger" : "secondary"} className={styles.hint}>{statusLabels[task.status]}{task.message ? ` · ${task.message}` : ""}</Typography.Text>
            <Progress size="small" percent={Math.max(0, Math.min(100, Number(task.progress) || 0))} status={task.status === 2 ? "success" : task.status >= 3 ? "exception" : "active"}/>
            {task.result?.path && <Typography.Text type="secondary" className={styles.hint}>{task.action === "upload" ? `文件已上传：${task.result.key}` : `文件已保存：${task.result.path}`}</Typography.Text>}
            <Typography.Text type="secondary" className={styles.hint}>可关闭窗口，进度、日志和取消操作可在系统任务中查看。</Typography.Text>
        </>}
        {error && <Typography.Text type="danger" className={styles.hint}>{error}</Typography.Text>}
    </ProModal>;
});

export default ServerTransfer;
