import React, {useCallback, useEffect, useRef, useState} from "react";
import {App, Button, Form, Input, Select, Space, Tooltip} from "antd";
import {Icon} from "sinking-antd";
import {systemTimezone} from "@/service/api/system";
import Actions from "../actions";
import Error from "../error";
import Loading from "../loading";

export default ({styles}: any): React.ReactNode => {
    const {message} = App.useApp();
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(true);
    const [pending, setPending] = useState({set: false, sync: false});
    const [error, setError] = useState("");
    const [timezone, setTimezone] = useState("");
    const [timezones, setTimezones] = useState<string[]>([]);
    const [serverTime, setServerTime] = useState("");
    const mountedRef = useRef(true);
    const pendingRef = useRef({set: false, sync: false});
    const requestRef = useRef(0);
    const actionRequestRef = useRef({set: 0, sync: 0});

    const load = useCallback(async (preserveSelection = false) => {
        const requestId = ++requestRef.current;
        if (!preserveSelection) {
            setLoading(true);
            setError("");
        }
        try {
            const response = await systemTimezone({body: {action: "get"}});
            if (!mountedRef.current || requestId !== requestRef.current) {
                return;
            }
            if (response?.code !== 200) {
                const text = response?.message || "时区信息加载失败";
                if (preserveSelection) message.error(text);
                else setError(text);
                return;
            }
            setError("");
            const currentTimezone = String(response?.data?.timezone || "");
            const availableTimezones: string[] = Array.isArray(response?.data?.timezones) ? response?.data?.timezones.filter((value: unknown): value is string => typeof value === "string") : [];
            setTimezone(currentTimezone);
            setTimezones(currentTimezone && !availableTimezones.includes(currentTimezone)
                ? [currentTimezone, ...availableTimezones]
                : availableTimezones);
            // 保留服务器当地时间及其 UTC 偏移量，不转换到浏览器所在时区。
            setServerTime(String(response?.data?.time || "").replace("T", " "));
            if (!preserveSelection) {
                form.setFieldsValue({timezone: currentTimezone || undefined});
            }
        } catch (error: any) {
            if (mountedRef.current && requestId === requestRef.current) {
                const text = error?.message || "时区信息加载失败";
                if (preserveSelection) message.error(text);
                else setError(text);
            }
        } finally {
            if (mountedRef.current && requestId === requestRef.current) {
                setLoading(false);
            }
        }
    }, [form, message]);

    useEffect(() => {
        mountedRef.current = true;
        load();
        return () => {
            mountedRef.current = false;
            pendingRef.current = {set: false, sync: false};
            requestRef.current += 1;
            actionRequestRef.current.set += 1;
            actionRequestRef.current.sync += 1;
        };
    }, [load]);

    const reset = useCallback(() => {
        form.setFieldsValue({timezone: timezone || undefined});
    }, [form, timezone]);

    const submit = useCallback(async (action: "set" | "sync", nextTimezone?: string) => {
        if (!mountedRef.current || pendingRef.current[action]) {
            return;
        }
        if (action === "set" && nextTimezone === timezone) {
            message.info("时区没有变化");
            return;
        }
        pendingRef.current[action] = true;
        const requestId = ++actionRequestRef.current[action];
        const failureMessage = action === "sync" ? "时间同步失败" : "时区保存失败";
        setPending((current) => ({...current, [action]: true}));
        try {
            const response = await systemTimezone({body: action === "sync" ? {action} : {action, timezone: nextTimezone}});
            if (!mountedRef.current || requestId !== actionRequestRef.current[action]) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || failureMessage);
                return;
            }
            message.success(response?.message || (action === "sync" ? "时间同步成功" : "时区保存成功"));
            await load(action === "sync");
        } catch (error: any) {
            if (mountedRef.current && requestId === actionRequestRef.current[action]) {
                message.error(error?.message || failureMessage);
            }
        } finally {
            if (mountedRef.current && requestId === actionRequestRef.current[action]) {
                pendingRef.current[action] = false;
                setPending((current) => ({...current, [action]: false}));
            }
        }
    }, [load, message, timezone]);

    return (
        <Form form={form} layout="vertical" variant="filled" disabled={pending.set} onFinish={(values) => submit("set", values?.timezone)}>
            {loading ? <Loading styles={styles}/> : error ? (
                <Error styles={styles} message={error} onRetry={() => load()}/>
            ) : <>
                <Form.Item label="当前时区" className={styles.formField}>
                    <Input value={timezone || "—"} readOnly/>
                </Form.Item>
                <Form.Item label="服务器时间" className={styles.formField}>
                    <Space.Compact block>
                        <Input value={serverTime || "—"} readOnly/>
                        <Tooltip title="从云端同步时间">
                            <Button color="default" variant="filled" aria-label="从云端同步时间"
                                icon={<Icon type="SyncOutlined"/>} loading={pending.sync} onClick={() => submit("sync")}/>
                        </Tooltip>
                    </Space.Compact>
                </Form.Item>
                <Form.Item
                    name="timezone"
                    label="时区"
                    className={styles.formField}
                    rules={[{required: true, message: "请选择时区"}]}>
                    <Select
                        showSearch={{optionFilterProp: "label"}}
                        placeholder="搜索并选择时区"
                        options={timezones.map((value) => ({label: value, value}))}
                    />
                </Form.Item>
                <Actions styles={styles} saving={pending.set} onReset={reset}/>
            </>}
        </Form>
    );
};
