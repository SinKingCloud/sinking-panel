import {useLayoutEffect, useRef, useState} from "react";
import {App, Form, Input, InputNumber} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import type {DatabasePanelRecord, DatabasePanelRequest} from "./entities";
import useStyles from "./styles/client";
import useFormStyles from "./manager-form-styles";

export default ({open, instance, name, type, request, onCancel, onSuccess}: {
    open: boolean; instance?: DatabasePanelRecord; name?: string; type: number; request: DatabasePanelRequest;
    onCancel: () => void; onSuccess: (change: {name: string}) => void;
}) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const {styles: formStyles} = useFormStyles();
    const theme = useTheme();
    const [form] = Form.useForm();
    const [visible, setVisible] = useState(false);
    const [saving, setSaving] = useState(false);
    const pending = useRef(false);
    const sequence = useRef(0);
    const lastName = useRef(name);
    const currentName = open ? name : lastName.current;
    const editing = currentName !== undefined;
    const width = (editing ? (type === 5 ? 340 : 400) : 450) - (theme?.isCompactTheme?.() ? 30 : 0);
    const label = type === 7 ? "索引" : "数据库";
    useLayoutEffect(() => {
        sequence.current++;
        if (!open) return;
        lastName.current = name;
        setVisible(true); pending.current = false; setSaving(false);
        form.resetFields(); form.setFieldsValue({name: name || "", collection: "", shards: 1, replicas: 0});
        return () => { sequence.current++; };
    }, [open, instance?.id, name]);
    const submit = async (values: any) => {
        if (!open || !instance || pending.current) return;
        const nextName = values.name.trim();
        if (editing && nextName === name) { message.info("没有修改内容"); return; }
        const current = sequence.current;
        pending.current = true; setSaving(true);
        try {
            const body: Record<string, any> = {id: instance.id, name: editing ? name : nextName};
            if (editing) body.new_name = nextName;
            if (!editing && type === 4) body.collection = values.collection.trim();
            if (!editing && type === 7) body.settings = JSON.stringify({number_of_shards: values.shards, number_of_replicas: values.replicas});
            const response = await request(editing ? "rename_database" : type === 7 ? "create_index" : "create_database", body);
            if (current !== sequence.current) return;
            if (response.code !== 200) throw new Error(response.message || "保存失败");
            message.success(editing ? "修改成功" : "创建成功");
            onSuccess({name: nextName}); onCancel();
        } catch (reason: any) { if (current === sequence.current) message.error(reason.message || "保存失败"); }
        finally { if (current === sequence.current) { pending.current = false; setSaving(false); } }
    };
    return <ProModal title={<Title>{editing ? `修改${label}` : `创建${label}`}</Title>} onOk={form.submit}
        onCancel={() => { if (!pending.current) onCancel(); }} modalProps={{
            open, width, zIndex: 1100, rootClassName: styles.modal,
            style: {maxWidth: "calc(100vw - 24px)"}, okText: "保存", cancelText: "取消", confirmLoading: saving,
            cancelButtonProps: {disabled: saving}, closable: !saving, keyboard: !saving, mask: {closable: !saving},
            forceRender: true, afterClose: () => setVisible(false), focusable: {focusTriggerAfterClose: false},
        }}>
        <Form form={form} layout="vertical" variant="filled" onFinish={submit} disabled={saving} className={`${styles.form} ${formStyles.form}`}>
            {visible && <>
                <Form.Item name="name" label={`${label}名称`} rules={[{required: true, whitespace: true, message: `请输入${label}名称`}, {max: 255, message: "名称过长"},
                    ...(type === 7 ? [{pattern: /^(?![._+-])(?!.*[\\/*?"<>| ,#:])[a-z0-9][a-z0-9._-]*$/, message: "请使用小写字母、数字、点、横线或下划线"}] : [])]}>
                    <Input placeholder={`请输入${label}名称`} maxLength={255} autoComplete="off"/>
                </Form.Item>
                {!editing && type === 4 && <Form.Item name="collection" label="初始集合" rules={[{required: true, whitespace: true, message: "请输入初始集合名称"}]}>
                    <Input placeholder="例如 documents" maxLength={255} autoComplete="off"/>
                </Form.Item>}
                {!editing && type === 7 && <>
                    <Form.Item name="shards" label="主分片数" rules={[{required: true, message: "请输入主分片数"}]}>
                        <InputNumber placeholder="请输入主分片数" min={1} max={1024} precision={0} style={{width: "100%"}}/>
                    </Form.Item>
                    <Form.Item name="replicas" label="副本数" rules={[{required: true, message: "请输入副本数"}]}>
                        <InputNumber placeholder="请输入副本数" min={0} max={1024} precision={0} style={{width: "100%"}}/>
                    </Form.Item>
                </>}
            </>}
        </Form>
    </ProModal>;
};
