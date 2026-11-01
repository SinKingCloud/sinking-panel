import {useCallback, useEffect, useRef, useState} from "react";
import {Alert, App, Button, Form, Grid, Input, Space, Spin} from "antd";
import {ProModal, Title} from "sinking-antd";
import {SecretSelect} from "@/pages/components/secret";
import {createOss, getOssInfo, testOss, updateOss} from "@/service/api/oss";
import useStyles from "@/pages/database/forms-styles";
import useModalStyles from "@/pages/database/components/shared/styles/client";

interface ConfigField {
    name: string;
    label: string;
    required?: boolean;
    required_without?: string;
    placeholder?: string;
    help?: string;
    max_length?: number;
}

const configFields: ConfigField[] = [
    {name: "region", label: "地域", required_without: "endpoint", max_length: 100,
        placeholder: "请输入存储桶所在地域", help: "地域与服务地址至少填写一项。"},
    {name: "bucket", label: "存储桶", required: true, max_length: 255,
        placeholder: "请输入已有存储桶名称", help: "填写已创建的存储桶名称。"},
    {name: "endpoint", label: "服务地址", required_without: "region", max_length: 500,
        placeholder: "可选，填写区域服务地址", help: "留空使用该地域的默认地址；地址不包含桶名或路径。"},
];

export default ({open, record, onCancel, onCreated, onSuccess}: any) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const {styles: modalStyles} = useModalStyles();
    const screens = Grid.useBreakpoint();
    const [form] = Form.useForm();
    const request = useRef(0);
    const submitting = useRef(false);
    const [fields, setFields] = useState<ConfigField[]>([]);
    const [loading, setLoading] = useState(false);
    const [busy, setBusy] = useState(false);
    const [testing, setTesting] = useState(false);
    const [error, setError] = useState("");
    const editing = Boolean(record?.id);
    const blocked = busy || testing;
    // load 按配置对象的实际字段生成表单，切换记录时清空旧值。
    const load = useCallback(async () => {
        const current = ++request.current;
        submitting.current = false;
        setLoading(true); setError(""); setBusy(false); setTesting(false); setFields([]); form.resetFields();
        try {
            if (!record?.id) return;
            const response = await getOssInfo({body: {id: record.id}});
            if (current !== request.current) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "获取 OSS 配置失败");
            const values = response.data;
            const config = values.config;
            if (!config || typeof config !== "object" || Array.isArray(config) || Object.keys(config).length === 0 ||
                Object.entries(config).some(([key, value]) => !key.trim() || typeof value !== "string")) throw new Error("OSS 配置格式错误");
            form.setFieldsValue(values);
            setFields(Object.keys(config).map((name) => configFields.find((field) => field.name === name) || {name, label: name}));
        } catch (reason: any) { if (current === request.current) setError(reason?.message || "加载失败"); }
        finally { if (current === request.current) setLoading(false); }
    }, [form, record?.id]);
    useEffect(() => { if (open) void load(); return () => { request.current++; }; }, [open, load]);
    const save = async (values: any) => {
        if (submitting.current || loading || error || (editing && fields.length === 0)) return;
        const current = request.current;
        submitting.current = true;
        setBusy(true);
        try {
            if (editing && (!values.config || typeof values.config !== "object" || Array.isArray(values.config) ||
                fields.some((field) => typeof values.config[field.name] !== "string"))) throw new Error("OSS 配置格式错误");
            const body = {
                name: values.name.trim(),
                ...(editing ? {
                    id: record.id,
                    config: JSON.stringify(Object.fromEntries(fields.map((field) => [field.name, values.config[field.name]]))),
                } : {secret_id: Number(values.secret_id)}),
            };
            const response = await (editing ? updateOss : createOss)({body});
            if (current !== request.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "保存失败");
            if (editing) {
                message.success("OSS 配置已保存"); onSuccess(record.id, true); onCancel();
            } else {
                const id = Number(response.data?.id);
                if (!Number.isSafeInteger(id) || id <= 0) throw new Error("服务器未返回 OSS 配置 ID");
                message.success("OSS 已添加"); onCreated(id); onSuccess(id, false);
            }
        } catch (reason: any) { if (current === request.current) message.error(reason?.message || "保存失败"); }
        finally { if (current === request.current) { submitting.current = false; setBusy(false); } }
    };
    const test = async () => {
        if (!editing || submitting.current || loading || error || fields.length === 0) return;
        const current = request.current;
        submitting.current = true;
        setTesting(true);
        try {
            await form.validateFields(fields.map((field) => ["config", field.name]));
            if (current !== request.current) return;
            const secretId = Number(form.getFieldValue("secret_id"));
            if (!Number.isInteger(secretId) || secretId <= 0) throw new Error("请选择已有云密钥");
            const values = form.getFieldValue("config");
            const response = await testOss({body: {
                secret_id: secretId,
                config: JSON.stringify(Object.fromEntries(fields.map((field) => [field.name, values?.[field.name] || ""]))),
            }});
            if (current !== request.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "连接失败");
            message.success("存储桶连接正常");
        } catch (reason: any) {
            if (current === request.current && !reason?.errorFields) message.error(reason?.message || "连接失败");
        } finally {
            if (current === request.current) { submitting.current = false; setTesting(false); }
        }
    };
    return <ProModal title={<Title>{editing ? "修改 OSS 配置" : "添加 OSS 配置"}</Title>} onOk={form.submit} onCancel={() => { if (!blocked) onCancel(); }} modalProps={{open, width: 460, zIndex: 1140, rootClassName: modalStyles.modal, style: {top: screens.md ? 90 : 24, paddingBottom: screens.md ? 90 : 24, maxWidth: "calc(100vw - 24px)"}, confirmLoading: busy, okText: editing ? "保存" : "添加", cancelText: "取消", okButtonProps: {disabled: blocked || loading || Boolean(error) || (editing && fields.length === 0)}, cancelButtonProps: {disabled: blocked}, closable: !blocked, keyboard: !blocked, mask: {closable: !blocked},
        footer: loading ? null : !editing ? undefined : (_: any, {OkBtn, CancelBtn}: any) => <div style={{display: "flex", alignItems: "center", justifyContent: "space-between", gap: 12}}>
            <Button loading={testing} disabled={busy || Boolean(error) || fields.length === 0} onClick={() => void test()}>测试连接</Button>
            <Space><CancelBtn/><OkBtn/></Space>
        </div>,
    } as any}>
        {loading ? <Spin/> : error ? <Alert type="error" showIcon title={error} action={<Button onClick={load}>重试</Button>}/> : <Form form={form} layout="vertical" variant="filled" disabled={blocked} className={styles.form} onFinish={save}>
            <Form.Item name="name" label="名称" rules={[{required: true, whitespace: true, message: "请输入名称"}]}><Input maxLength={100} placeholder="例如：网站备份存储"/></Form.Item>
            {!editing && <Form.Item name="secret_id" label="云密钥" rules={[{required: true, message: "请选择已有云密钥"}]}>
                <SecretSelect placeholder="请选择云密钥" zIndex={1160}/>
            </Form.Item>}
            {editing && fields.map((field) => <Form.Item key={field.name}
                name={["config", field.name]} label={field.label} extra={field.help}
                dependencies={field.required_without ? [["config", field.required_without]] : []}
                rules={[
                    ...(field.required ? [{required: true, whitespace: true, message: `请输入${field.label}`}] : []),
                    ...(field.max_length ? [{max: field.max_length, message: `${field.label}不能超过 ${field.max_length} 个字符`}] : []),
                    {validator: async (_, value) => {
                        if (field.required_without && !String(value || "").trim() && !String(form.getFieldValue(["config", field.required_without]) || "").trim()) {
                            const other = fields.find((item) => item.name === field.required_without);
                            throw new Error(`${field.label}与${other?.label || field.required_without}至少填写一项`);
                        }
                    }},
                ]}><Input maxLength={field.max_length} placeholder={field.placeholder} autoComplete="off"/></Form.Item>)}
        </Form>}
    </ProModal>;
};
