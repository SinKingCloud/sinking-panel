import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Button, Form, Grid, Input, Select, Space, Spin, Typography} from "antd";
import {createStyles} from "antd-style";
import {ProModal, Title, useTheme} from "sinking-antd";
import type {ProModalRef} from "sinking-antd";
import {createSecret, getSecretInfo, updateSecret} from "@/service/api/secret";
import {getEnum as getSystemEnum} from "@/service/api/system";

export interface SecretChange {
    action: "create" | "update" | "delete";
    id: number;
}

export interface SecretFormRef {
    open: (record?: any) => void;
    close: () => void;
}

interface SecretFormProps {
    zIndex: number;
    onSuccess: (change: SecretChange) => void;
}


const useStyles = createStyles(({css}: any) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-body {
            max-height: calc(100dvh - 290px);
            overflow-y: auto;
            overscroll-behavior: contain;
        }

        @supports not (height: 100dvh) {
            .ant-modal-body {
                max-height: calc(100vh - 290px);
            }
        }
        @media (max-width: 767px) {
            .ant-modal-body {
                max-height: calc(100dvh - 138px);
            }

            @supports not (height: 100dvh) {
                .ant-modal-body {
                    max-height: calc(100vh - 138px);
                }
            }
        }
    `,
    loading: css`
        min-height: 240px;
        display: flex;
        align-items: center;
        justify-content: center;
    `,
}));

const SecretForm = forwardRef<SecretFormRef, SecretFormProps>(({zIndex, onSuccess}, ref): any => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles();
    const modalRef = useRef<ProModalRef | null>(null);
    const requestRef = useRef(0);
    const mountedRef = useRef(true);
    const openRef = useRef(false);
    const submittingRef = useRef(false);
    const originalCredentialsRef = useRef<Record<string, string>>({});
    const [form] = Form.useForm<any>();
    const [record, setRecord] = useState<any>();
    const [providerNames, setProviderNames] = useState<Record<string, string>>({});
    const [fields, setFields] = useState<string[]>([]);
    const [active, setActive] = useState(false);
    const [loading, setLoading] = useState(false);
    const [loadError, setLoadError] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const editing = record !== undefined;
    const providerOptions = useMemo(() => Object.entries(providerNames).map(([value, label]) => ({
        value: Number(value),
        label
    })), [providerNames]);

    const close = useCallback(() => {
        requestRef.current += 1;
        openRef.current = false;
        submittingRef.current = false;
        originalCredentialsRef.current = {};
        modalRef.current?.hide();
    }, []);

    const open = useCallback((record?: any) => {
        const requestId = ++requestRef.current;
        openRef.current = true;
        submittingRef.current = false;
        originalCredentialsRef.current = {};
        form.resetFields();
        setRecord(record);
        setProviderNames({});
        setFields([]);
        setActive(true);
        setLoading(true);
        setLoadError("");
        setSubmitting(false);
        modalRef.current?.show();
        void (async () => {
            try {
                const [enumResponse, infoResponse] = await Promise.all([
                    getSystemEnum({body: {name: "secret"}}),
                    record ? getSecretInfo({body: {id: record.id}}) : Promise.resolve(undefined),
                ]);
                if (requestRef.current !== requestId || !openRef.current) {
                    return;
                }
                if (enumResponse?.code !== 200 || !enumResponse?.data?.provider) {
                    throw new Error(enumResponse?.message || "获取密钥配置失败");
                }
                if (record && (infoResponse?.code !== 200 || !infoResponse?.data)) {
                    throw new Error(infoResponse?.message || "获取密钥详情失败");
                }
                const enums = enumResponse?.data;
                const info = infoResponse?.data;
                const nextProvider = info ? Number(info?.provider) : Number(Object.keys(enums?.provider || {})[0]);
                const credentials = info ? info?.data : {};
                if (!enums?.provider?.[String(nextProvider)] || !credentials || typeof credentials !== "object"
                    || Array.isArray(credentials) || (info && Object.keys(credentials).length === 0)
                    || Object.values(credentials).some((value) => typeof value !== "string")) {
                    throw new Error("密钥配置不可用，请重新加载");
                }
                originalCredentialsRef.current = info ? {...credentials} : {};
                setProviderNames(enums?.provider || {});
                setFields(Object.keys(credentials));
                form.setFieldsValue({
                    name: info?.name || "",
                    provider: nextProvider,
                    credentials: {...credentials},
                });
            } catch (error) {
                if (requestRef.current === requestId && openRef.current) {
                    setLoadError(error instanceof Error ? error.message : "获取密钥配置失败");
                }
            } finally {
                if (requestRef.current === requestId) {
                    setLoading(false);
                }
            }
        })();
    }, [form]);

    useImperativeHandle(ref, () => ({open, close}), [close, open]);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            requestRef.current += 1;
            openRef.current = false;
            originalCredentialsRef.current = {};
        };
    }, []);

    const submit = useCallback(async (values: any) => {
        if (!openRef.current || submittingRef.current || loading || loadError || (editing && fields.length === 0)) {
            return;
        }
        const selectedProvider = Number(values.provider);
        if (!editing && !providerNames[String(selectedProvider)]) {
            message.error("请选择密钥厂商");
            return;
        }
        const requestId = ++requestRef.current;
        const credentials: Record<string, string> = {};
        if (editing) {
            fields.forEach((field) => {
                const value = String(values.credentials?.[field] || "").trim();
                if (value !== originalCredentialsRef.current[field]) {
                    credentials[field] = value;
                }
            });
        }
        const body = {
            name: values.name.trim(),
            ...(!editing ? {provider: selectedProvider} : {}),
            ...(Object.keys(credentials).length > 0 ? {data: JSON.stringify(credentials)} : {}),
            ...(record ? {id: record.id} : {}),
        };
        submittingRef.current = true;
        setSubmitting(true);
        try {
            const response = await (editing ? updateSecret : createSecret)({body});
            if (!mountedRef.current) {
                return;
            }
            if (response?.code !== 200) {
                if (requestRef.current === requestId) {
                    message.error(response?.message || "保存密钥失败");
                }
                return;
            }
            if (requestRef.current === requestId) {
                message.success(response?.message || "保存成功");
                if (record) {
                    close();
                } else {
                    open({id: Number(response?.data?.id)});
                }
            }
            // 已完成的保存仍需通知列表和调用方，不影响后来打开的表单。
            onSuccess({action: editing ? "update" : "create", id: record?.id ?? Number(response?.data?.id)});
        } catch {
            if (requestRef.current === requestId && openRef.current) {
                message.error("保存密钥失败");
            }
        } finally {
            if (requestRef.current === requestId) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    }, [close, editing, fields, loadError, loading, message, onSuccess, open, providerNames, record]);

    return <ProModal
        ref={modalRef}
        title={<Title>{editing ? "编辑密钥" : "添加密钥"}</Title>}
        width={editing ? (compact ? 380 : 400) : 350}
        onOk={form.submit}
        onCancel={close}
        modalProps={{
            rootClassName: styles.modal,
            zIndex,
            forceRender: true,
            confirmLoading: submitting,
            okButtonProps: {disabled: loading || Boolean(loadError)},
            okText: editing ? "保存" : "添加",
            cancelText: "取消",
            footer: loading || loadError ? null : undefined,
            mask: {closable: true},
            style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
            focusable: {focusTriggerAfterClose: false},
            afterClose: () => {
                if (openRef.current) return;
                form.resetFields();
                setRecord(undefined);
                setProviderNames({});
                setFields([]);
                setActive(false);
                setLoading(false);
                setLoadError("");
                setSubmitting(false);
            },
        } as any}>
        <Form form={form} layout="vertical" variant="filled" onFinish={submit}>
            {active && (loading ? <div className={styles.loading}><Spin/></div> : loadError ?
                <div className={styles.loading}>
                    <Space orientation="vertical" align="center">
                        <Typography.Text type="danger">{loadError}</Typography.Text>
                        <Button color="default" variant="filled" onClick={() => open(record)}>重新加载</Button>
                    </Space>
                </div> : <>
                    <Form.Item name="name" label="密钥名称" rules={[
                        {required: true, whitespace: true, message: "请输入密钥名称"},
                        {max: 100, message: "密钥名称不能超过 100 个字符"},
                    ]}>
                        <Input maxLength={100} placeholder="请输入密钥名称" autoComplete="off"/>
                    </Form.Item>
                    {!editing && <Form.Item name="provider" label="密钥厂商"
                                            rules={[{required: true, message: "请选择密钥厂商"}]}>
                        <Select options={providerOptions} placeholder="请选择密钥厂商" disabled={submitting}/>
                    </Form.Item>}
                    {editing && fields.map((field) => <Form.Item
                        key={field}
                        preserve={false}
                        name={["credentials", field]}
                        label={field}
                        rules={[{required: true, whitespace: true, message: `请输入 ${field}`}]}>
                        <Input placeholder={`请输入 ${field}`} autoComplete="off"/>
                    </Form.Item>)}
                </>)}
        </Form>
    </ProModal>;
});

SecretForm.displayName = "SecretForm";

export default memo(SecretForm);
