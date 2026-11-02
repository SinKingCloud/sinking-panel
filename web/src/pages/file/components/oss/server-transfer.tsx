import {forwardRef, useEffect, useImperativeHandle, useRef, useState} from "react";
import {Button, Form, Input, Typography} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import {useOutletContext} from "umi";
import type {LayoutContext} from "@/layouts/context";
import FilePicker from "@/pages/components/file-picker";
import {comparableFilePath, isAbsoluteFilePath, joinFilePath, normalizeFilePath} from "@/pages/file/utils";
import {ossDownload, ossUpload} from "@/service/api/oss";
import useStyles from "./styles";

// ServerTransfer 只提交服务器文件路径，传输和取消由系统任务管理。
const ServerTransfer = forwardRef<any, any>(({record, active: visible = true, onBusyChange}, ref) => {
    const {task} = useOutletContext<LayoutContext>();
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const [form] = Form.useForm();
    const [dialog, setDialog] = useState<{action: "upload" | "download"; key: string}>();
    const [open, setOpen] = useState(false);
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const [unconfirmed, setUnconfirmed] = useState(false);
    const [submitted, setSubmitted] = useState(0);
    const live = useRef(true);
    const pending = useRef(false);
    const modalWidth = useRef(420);
    const busyChanged = useRef(onBusyChange);
    busyChanged.current = onBusyChange;
    const locked = submitted > 0 || unconfirmed;

    useEffect(() => { live.current = true; return () => { live.current = false; busyChanged.current?.(false); }; }, []);
    useEffect(() => { onBusyChange?.(busy); }, [busy, onBusyChange]);
    useEffect(() => { if (!visible) setOpen(false); }, [visible]);
    useImperativeHandle(ref, () => ({
        open: (action: "upload" | "download", key: string) => {
            if (pending.current) return;
            modalWidth.current = action === "upload" ? 380 : 420;
            form.resetFields();
            if (action === "download") form.setFieldsValue({name: key.split("/").pop()});
            setSubmitted(0); setError(""); setUnconfirmed(false);
            setDialog({action, key});
            setOpen(true);
        },
    }), [form]);

    const submit = async (values: any) => {
        if (!open || !dialog || pending.current || locked) return;
        const uploading = dialog.action === "upload";
        const paths = uploading
            ? [...new Map((values.paths as string[]).map((path) => [comparableFilePath(path), path])).values()]
            : [joinFilePath(values.directory, values.name)];
        let created = 0;
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
                    setError(paths.length > 1 ? `${path}：${reason}。已提交 ${created} 个文件，后续 ${paths.length - index - 1} 个文件未提交。` : reason);
                    if (created > 0) task?.open();
                    return;
                }
                setSubmitted(++created);
            }
            setOpen(false);
            task?.open();
        } finally { pending.current = false; if (live.current) setBusy(false); }
    };

    return <ProModal title={<Title>{dialog?.action === "upload" ? "从服务器上传" : "下载到服务器"}</Title>}
        onCancel={() => { if (!pending.current) setOpen(false); }} onOk={form.submit}
        modalProps={{open: visible && open, width: modalWidth.current, zIndex: 1140,
            style: {maxWidth: "calc(100vw - 24px)"}, forceRender: true,
            okText: dialog?.action === "upload" ? "上传" : "下载", cancelText: "取消", confirmLoading: busy,
            closable: !busy, keyboard: !busy, mask: {closable: !busy}, cancelButtonProps: {disabled: busy},
            footer: locked ? <>
                <Button disabled={busy} onClick={() => setOpen(false)}>关闭</Button>
                <Button disabled={busy} type="primary" onClick={() => task?.open()}>查看系统任务</Button>
            </> : undefined}}>
        {visible && <Form form={form} layout="vertical" variant="filled" disabled={busy || locked} onFinish={submit}>
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
        {error && <Typography.Text type="danger" className={styles.hint}>{error}</Typography.Text>}
    </ProModal>;
});

export default ServerTransfer;
