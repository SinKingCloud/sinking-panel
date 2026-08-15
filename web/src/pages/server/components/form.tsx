import {forwardRef, memo, useCallback, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Col, Form as AntForm, Input, InputNumber, Row, Select} from "antd";
import {ProModal, ProModalRef, Title} from "sinking-antd";
import {createServer, updateServer} from "@/service/api/server";
import type {ServerRecord} from "../hooks/servers";

export interface FormRef {
    open: (record?: ServerRecord) => void;
}

interface FormProps {
    authTypeData: Record<string, string>;
    onSuccess?: (result: FormSuccessResult) => void;
}

export interface FormSuccessResult {
    serverId?: number;
    server?: ServerRecord;
    reconnectRequired: boolean;
}

const fallbackAuthTypes = {
    "0": "密码验证",
    "1": "证书验证",
};

const Form = forwardRef<FormRef, FormProps>(({authTypeData, onSuccess}, ref): any => {
    const {message} = App.useApp();
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const requestRef = useRef(0);
    const submittingRef = useRef(false);
    const [form] = AntForm.useForm();
    const [record, setRecord] = useState<ServerRecord>();
    const [authType, setAuthType] = useState(0);
    const [initialAuthType, setInitialAuthType] = useState(0);
    const [submitting, setSubmitting] = useState(false);
    const editing = record !== undefined;
    const authOptions = useMemo(() => {
        const data = Object.keys(authTypeData || {}).length > 0 ? authTypeData : fallbackAuthTypes;
        return Object.entries(data).map(([value, label]) => ({
            value: Number(value),
            label: String(label),
        }));
    }, [authTypeData]);
    const credentialRequired = !editing || authType !== initialAuthType;

    const reset = useCallback(() => {
        requestRef.current += 1;
        submittingRef.current = false;
        form.resetFields();
        setRecord(undefined);
        setAuthType(0);
        setInitialAuthType(0);
        setSubmitting(false);
    }, [form]);

    const open = useCallback((server?: ServerRecord) => {
        requestRef.current += 1;
        submittingRef.current = false;
        form.resetFields();
        if (server) {
            const currentAuthType = Number(server.auth_type || 0);
            setRecord(server);
            setAuthType(currentAuthType);
            setInitialAuthType(currentAuthType);
            form.setFieldsValue({
                name: server.name,
                ip: server.ip,
                port: server.port || 22,
                user: server.user,
                auth_type: currentAuthType,
                password: "",
            });
        } else {
            setRecord(undefined);
            setAuthType(0);
            setInitialAuthType(0);
            form.setFieldsValue({
                name: "",
                ip: "",
                port: 22,
                user: "",
                auth_type: 0,
                password: "",
            });
        }
        modalRef.current?.show();
    }, [form]);

    useImperativeHandle(ref, () => ({open}), [open]);

    const submit = useCallback(async (values: any) => {
        if (submittingRef.current) {
            return;
        }
        submittingRef.current = true;
        setSubmitting(true);
        const requestId = ++requestRef.current;
        const nextAuthType = Number(values.auth_type || 0);
        const nextPort = Number(values.port || 22);
        const body: any = {
            name: String(values.name || "").trim(),
            ip: String(values.ip || "").trim(),
            port: editing ? String(nextPort) : nextPort,
            user: String(values.user || "").trim(),
        };
        if (!editing || nextAuthType !== Number(record?.auth_type || 0)) {
            body.auth_type = editing ? String(nextAuthType) : nextAuthType;
        }
        const password = String(values.password || "");
        if (password.trim()) {
            body.password = password;
        }
        if (editing) {
            body.ids = [record?.id];
        }

        try {
            const response = await (editing ? updateServer : createServer)({body});
            if (requestRef.current !== requestId || !response) {
                return;
            }
            if (response?.code === 200) {
                const updatedServer = editing && record ? {
                    ...record,
                    name: body.name,
                    ip: body.ip,
                    port: nextPort,
                    user: body.user,
                    auth_type: nextAuthType,
                    searchText: [body.name, body.ip, String(nextPort), `${body.ip}:${nextPort}`]
                        .join("\n")
                        .toLocaleLowerCase(),
                } : undefined;
                const reconnectRequired = Boolean(editing && record && (
                    record.ip !== body.ip ||
                    record.port !== nextPort ||
                    record.user !== body.user ||
                    record.auth_type !== nextAuthType ||
                    password.trim()
                ));
                modalRef.current?.hide();
                message.success(response?.message || (editing ? "修改成功" : "添加成功"));
                onSuccess?.({
                    serverId: editing ? record?.id : undefined,
                    server: updatedServer,
                    reconnectRequired,
                });
            } else {
                message.error(response?.message || (editing ? "修改终端失败" : "添加终端失败"));
            }
        } finally {
            if (requestRef.current === requestId) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    }, [editing, message, onSuccess, record]);

    return (
        <ProModal
            ref={modalRef}
            title={<Title>{editing ? "编辑终端" : "添加终端"}</Title>}
            onOk={form.submit}
            width={440}
            modalProps={{
                confirmLoading: submitting,
                cancelButtonProps: {disabled: submitting},
                closable: !submitting,
                keyboard: !submitting,
                forceRender: true,
                okText: editing ? "保存" : "添加",
                cancelText: "取消",
                style: {top: 100, paddingBottom: 100},
                mask: {closable: !submitting},
                afterClose: reset,
            } as any}>
            <AntForm form={form} layout="vertical" onFinish={submit}>
                <AntForm.Item
                    name="name"
                    label="终端名称"
                    rules={[{required: true, whitespace: true, message: "请输入终端名称"}]}>
                    <Input placeholder="例如：生产终端" autoComplete="off"/>
                </AntForm.Item>
                <Row gutter={[16, 0]}>
                    <Col xs={24} sm={17}>
                        <AntForm.Item
                            name="ip"
                            label="IP 地址"
                            rules={[{required: true, whitespace: true, message: "请输入 IP 地址"}]}>
                            <Input placeholder="例如：192.168.1.10" inputMode="text" autoComplete="off"/>
                        </AntForm.Item>
                    </Col>
                    <Col xs={24} sm={7}>
                        <AntForm.Item
                            name="port"
                            label="SSH 端口"
                            rules={[{required: true, message: "请输入端口"}]}>
                            <InputNumber min={1} max={65535} precision={0} style={{width: "100%"}}/>
                        </AntForm.Item>
                    </Col>
                </Row>
                <Row gutter={[16, 0]}>
                    <Col xs={24} sm={12}>
                        <AntForm.Item
                            name="user"
                            label="登录账号"
                            rules={[{required: true, whitespace: true, message: "请输入登录账号"}]}>
                            <Input placeholder="例如：root" autoComplete="username"/>
                        </AntForm.Item>
                    </Col>
                    <Col xs={24} sm={12}>
                        <AntForm.Item
                            name="auth_type"
                            label="认证方式"
                            rules={[{required: true, message: "请选择认证方式"}]}>
                            <Select
                                options={authOptions}
                                onChange={(value) => {
                                    setAuthType(Number(value));
                                    form.setFieldValue("password", "");
                                }}/>
                        </AntForm.Item>
                    </Col>
                </Row>
                <AntForm.Item
                    name="password"
                    label={authType === 1 ? "私钥" : "登录密码"}
                    rules={[{
                        validator: async (_, value) => {
                            if (credentialRequired && !String(value || "").trim()) {
                                throw new Error(authType === 1 ? "请输入私钥" : "请输入登录密码");
                            }
                        },
                    }]}>
                    {authType === 1 ? (
                        <Input.TextArea
                            autoSize={{minRows: 6, maxRows: 12}}
                            maxLength={128 * 1024}
                            placeholder={editing && !credentialRequired ? "留空则保持不变" : "粘贴 PEM 格式私钥"}
                            autoComplete="off"/>
                    ) : (
                        <Input.Password
                            maxLength={4 * 1024}
                            placeholder={editing && !credentialRequired ? "留空则保持不变" : "请输入登录密码"}
                            autoComplete="new-password"/>
                    )}
                </AntForm.Item>
            </AntForm>
        </ProModal>
    );
});

Form.displayName = "Form";

export default memo(Form);
