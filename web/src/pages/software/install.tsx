import {useEffect, useRef, useState} from "react";
import {App, Form, Grid} from "antd";
import {ProModal, Title} from "sinking-antd";
import FilePicker from "@/pages/components/file-picker";
import {installSoftware} from "@/service/api/software";

export default function InstallSoftware({open, onCancel, onSuccess}: any) {
    const {message} = App.useApp();
    const screens = Grid.useBreakpoint();
    const [form] = Form.useForm();
    const [busy, setBusy] = useState(false);
    const pending = useRef(false);
    const active = useRef(true);
    useEffect(() => { active.current = true; return () => { active.current = false; }; }, []);

    const submit = async ({path}: {path: string}) => {
        if (pending.current) return;
        pending.current = true; setBusy(true);
        try {
            const response = await installSoftware({body: {path: path.trim()}});
            if (!active.current) return;
            if (response?.code !== 200 || !response?.data) throw new Error(response?.message || "安装软件失败");
            if (Number(response?.data?.status) !== 2) throw new Error("软件未安装成功，请重新安装");
            onSuccess(response?.data);
        } catch (reason: any) { if (active.current) message.error(reason?.message || "安装软件失败"); }
        finally { pending.current = false; if (active.current) setBusy(false); }
    };

    return <ProModal title={<Title>添加软件</Title>} onOk={form.submit} onCancel={() => { if (!pending.current) onCancel(); }}
        modalProps={{open, forceRender: true, width: 360, zIndex: 1100,
            style: {top: screens.md ? 90 : 24, paddingBottom: screens.md ? 90 : 24, maxWidth: "calc(100vw - 24px)"},
            okText: "安装", cancelText: "取消", confirmLoading: busy, closable: !busy, keyboard: !busy,
            mask: {closable: !busy}, cancelButtonProps: {disabled: busy}, afterClose: () => form.resetFields()} as any}>
        <Form form={form} layout="vertical" variant="filled" disabled={busy} onFinish={submit}>
            <Form.Item name="path" label="软件路径" rules={[
                {required: true, whitespace: true, message: "请输入软件路径"}, {max: 4096, message: "路径最多 4096 个字符"},
            ]}>
                <FilePicker mode="file" placeholder="请输入或选择路径" maxLength={4096}/>
            </Form.Item>
        </Form>
    </ProModal>;
}
