import React, {useCallback, useEffect, useRef, useState} from "react";
import {App, Form, Input} from "antd";
import {useModel} from "umi";
import {getAccountInfo, updateAccount} from "@/service/api/system";
import Actions from "./actions";
import Error from "./error";
import Loading from "./loading";

const fieldStyle = {width: "100%", maxWidth: 350};

export default ({styles}: any): React.ReactNode => {
    const {message} = App.useApp();
    const user = useModel("user");
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState("");
    const mountedRef = useRef(true);
    const savingRef = useRef(false);
    const requestRef = useRef(0);
    const saveRequestRef = useRef(0);
    const accountRef = useRef("");

    const load = useCallback(async () => {
        const requestId = ++requestRef.current;
        setLoading(true);
        setError("");
        const response = await getAccountInfo();
        if (!mountedRef.current || requestId !== requestRef.current) {
            return;
        }
        if (response?.code !== 200) {
            setError(response?.message || "登录信息加载失败");
            setLoading(false);
            return;
        }
        const account = String(response?.data?.account || "");
        accountRef.current = account;
        form.setFieldsValue({account, password: "", confirm: ""});
        setLoading(false);
    }, [form]);

    useEffect(() => {
        mountedRef.current = true;
        load();
        return () => {
            mountedRef.current = false;
            requestRef.current += 1;
            saveRequestRef.current += 1;
        };
    }, [load]);

    const reset = useCallback(() => {
        form.setFieldsValue({account: accountRef.current, password: "", confirm: ""});
    }, [form]);

    const save = useCallback(async (values: any) => {
        if (savingRef.current) {
            return;
        }
        const accountValue = String(values.account || "");
        const account = accountValue.trim();
        const password = String(values.password || "");
        const body: any = {action: "update"};
        if (accountValue !== accountRef.current) {
            body.account = account;
        }
        if (password) {
            body.password = password;
        }
        if (!body.account && !body.password) {
            message.info("登录信息没有变化");
            return;
        }

        savingRef.current = true;
        const requestId = ++saveRequestRef.current;
        setSaving(true);
        const response = await updateAccount({body});
        if (!mountedRef.current || requestId !== saveRequestRef.current) {
            return;
        }
        if (response?.code !== 200) {
            message.error(response?.message || "登录信息保存失败");
            savingRef.current = false;
            setSaving(false);
            return;
        }
        const nextAccount = body.account || accountRef.current;
        accountRef.current = nextAccount;
        form.setFieldsValue({account: nextAccount, password: "", confirm: ""});
        user?.refreshWebUser?.();
        message.success(response?.message || "登录信息保存成功");
        savingRef.current = false;
        setSaving(false);
    }, [form, message, user]);

    return (
        <Form form={form} layout="vertical" onFinish={save}>
            {loading ? <Loading styles={styles}/> : error ? (
                <Error styles={styles} message={error} onRetry={load}/>
            ) : <>
                <Form.Item
                    name="account"
                    label="登录账号"
                    tooltip="用于登录面板的账号"
                    style={fieldStyle}
                    rules={[
                        {required: true, whitespace: true, message: "请输入登录账号"},
                        {
                            validator: (_, value) => {
                                if (!String(value ?? "").trim()) {
                                    return Promise.resolve();
                                }
                                if (value === accountRef.current) {
                                    return Promise.resolve();
                                }
                                if (!/^[A-Za-z0-9]+$/.test(value || "")) {
                                    return Promise.reject(new Error("登录账号只能包含字母和数字"));
                                }
                                if (String(value).length > 20) {
                                    return Promise.reject(new Error("登录账号不能超过20个字符"));
                                }
                                return Promise.resolve();
                            },
                        },
                    ]}>
                    <Input placeholder="请输入登录账号" maxLength={20} autoComplete="username"/>
                </Form.Item>
                <Form.Item
                    name="password"
                    label="新密码"
                    tooltip="不修改密码时请留空"
                    style={fieldStyle}
                    rules={[
                        {min: 6, message: "密码至少6个字符"},
                        {max: 20, message: "密码不能超过20个字符"},
                    ]}>
                    <Input.Password placeholder="不修改请留空" maxLength={20} autoComplete="new-password"/>
                </Form.Item>
                <Form.Item
                    name="confirm"
                    label="确认密码"
                    tooltip="再次输入新密码"
                    style={fieldStyle}
                    dependencies={["password"] as any}
                    rules={[
                        ({getFieldValue}) => ({
                            validator(_, value) {
                                const password = getFieldValue("password");
                                if (!password && !value) {
                                    return Promise.resolve();
                                }
                                if (!password) {
                                    return Promise.reject(new Error("请先输入新密码"));
                                }
                                if (value === password) {
                                    return Promise.resolve();
                                }
                                return Promise.reject(new Error("两次输入的密码不一致"));
                            },
                        }),
                    ]}>
                    <Input.Password placeholder="再次输入新密码" maxLength={20} autoComplete="new-password"/>
                </Form.Item>
                <Actions styles={styles} saving={saving} onReset={reset}/>
            </>}
        </Form>
    );
};
