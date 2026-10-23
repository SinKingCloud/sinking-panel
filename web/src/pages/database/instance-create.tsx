import {useEffect, useRef, useState} from "react";
import {App, Form, Grid, Input, Select} from "antd";
import {ProModal, Title} from "sinking-antd";
import {createDatabase} from "@/service/api/database";
import useStyles from "./components/shared/styles/client";

export default ({open, types, categories = [], defaultType, defaultTypeId, onCancel, onCreated, onSuccess}: any) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const screens = Grid.useBreakpoint();
    const [form] = Form.useForm();
    const [busy, setBusy] = useState(false);
    const busyRef = useRef(false);
    const requestRef = useRef(0);

    useEffect(() => () => { requestRef.current += 1; }, []);
    useEffect(() => {
        requestRef.current += 1;
        busyRef.current = false;
        if (!open) return;
        setBusy(false);
        form.resetFields();
        form.setFieldsValue({type: defaultType || Number(Object.keys(types || {})[0]) || undefined, type_id: Number(defaultTypeId) || 0});
    }, [open]);
    useEffect(() => {
        if (open && !form.getFieldValue("type")) {
            form.setFieldValue("type", defaultType || Number(Object.keys(types || {})[0]) || undefined);
        }
    }, [open, defaultType, types, form]);

    const submit = async (values: any) => {
        if (busyRef.current || !Object.keys(types || {}).length) return;
        const request = requestRef.current;
        const body = {name: values.name.trim(), type: Number(values.type), type_id: Number(values.type_id) || 0};
        busyRef.current = true;
        setBusy(true);
        try {
            const response = await createDatabase({body});
            if (request !== requestRef.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "添加实例失败");
            const id = Number(response.data?.id);
            if (!id) throw new Error("添加成功，但未返回实例信息，请刷新列表");
            message.success("实例已添加，请配置连接信息");
            onSuccess({action: "create", id});
            onCreated({id, ...body});
        } catch (reason: any) {
            if (request === requestRef.current) message.error(reason.message || "添加实例失败");
        } finally {
            if (request === requestRef.current) { busyRef.current = false; setBusy(false); }
        }
    };

    return <ProModal title={<Title>添加实例</Title>}
        onCancel={() => { if (!busyRef.current) onCancel(); }} onOk={form.submit}
        modalProps={{
            open, width: 400, zIndex: 1140, rootClassName: styles.modal,
            style: {top: screens.md ? 90 : 24, paddingBottom: screens.md ? 90 : 24, maxWidth: "calc(100vw - 24px)"},
            closable: !busy, keyboard: !busy, mask: {closable: !busy}, forceRender: true,
            focusable: {focusTriggerAfterClose: false}, okText: "添加", cancelText: "取消",
            confirmLoading: busy, okButtonProps: {disabled: !Object.keys(types || {}).length}, cancelButtonProps: {disabled: busy},
        } as any}>
        <Form form={form} layout="vertical" variant="filled" disabled={busy} onFinish={submit}>
            <Form.Item name="name" label="实例名称" rules={[{required: true, whitespace: true, message: "请输入实例名称"}, {max: 100, message: "名称不能超过 100 个字符"}]}>
                <Input maxLength={100} placeholder="例如：生产数据库" autoComplete="off"/>
            </Form.Item>
            <Form.Item name="type" label="数据库类型" rules={[{required: true, message: "请选择数据库类型"}]}>
                <Select placeholder="请选择数据库类型" options={Object.entries(types || {}).map(([value, label]) => ({value: Number(value), label: String(label)}))}/>
            </Form.Item>
            <Form.Item name="type_id" label="分类" rules={[{required: true, message: "请选择分类"}]}>
                <Select placeholder="请选择分类" options={[{value: 0, label: "全部分类"}, ...categories.map((item: any) => ({value: Number(item.id), label: item.name}))]}/>
            </Form.Item>
        </Form>
    </ProModal>;
};
