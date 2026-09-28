import React, {useCallback, useEffect, useRef, useState} from "react";
import {App, Form, Input, Select} from "antd";
import {systemTimezone} from "@/service/api/system";
import Actions from "../actions";
import Error from "../error";
import Loading from "../loading";

export default ({styles}: any): React.ReactNode => {
    const {message} = App.useApp();
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState("");
    const [timezone, setTimezone] = useState("");
    const [timezones, setTimezones] = useState<string[]>([]);
    const [serverTime, setServerTime] = useState("");
    const mountedRef = useRef(true);
    const savingRef = useRef(false);
    const requestRef = useRef(0);
    const saveRequestRef = useRef(0);

    const load = useCallback(async () => {
        const requestId = ++requestRef.current;
        setLoading(true);
        setError("");
        const response = await systemTimezone({body: {action: "get"}});
        if (!mountedRef.current || requestId !== requestRef.current) {
            return;
        }
        if (response?.code !== 200) {
            setError(response?.message || "时区信息加载失败");
            setLoading(false);
            return;
        }
        const currentTimezone = String(response?.data?.timezone || "");
        const availableTimezones: string[] = response?.data?.timezones || [];
        setTimezone(currentTimezone);
        setTimezones(currentTimezone && !availableTimezones.includes(currentTimezone)
            ? [currentTimezone, ...availableTimezones]
            : availableTimezones);
        // 保留服务器当地时间及其 UTC 偏移量，不转换到浏览器所在时区。
        setServerTime(String(response?.data?.time || "").replace("T", " "));
        form.setFieldsValue({timezone: currentTimezone || undefined});
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
        form.setFieldsValue({timezone: timezone || undefined});
    }, [form, timezone]);

    const save = useCallback(async (values: any) => {
        if (savingRef.current) {
            return;
        }
        if (values.timezone === timezone) {
            message.info("时区没有变化");
            return;
        }
        savingRef.current = true;
        const requestId = ++saveRequestRef.current;
        setSaving(true);
        const response = await systemTimezone({body: {action: "set", timezone: values.timezone}});
        if (!mountedRef.current || requestId !== saveRequestRef.current) {
            return;
        }
        if (response?.code !== 200) {
            message.error(response?.message || "时区保存失败");
            savingRef.current = false;
            setSaving(false);
            return;
        }
        message.success(response?.message || "时区保存成功");
        await load();
        if (mountedRef.current && requestId === saveRequestRef.current) {
            savingRef.current = false;
            setSaving(false);
        }
    }, [load, message, timezone]);

    return (
        <Form form={form} layout="vertical" variant="filled" disabled={saving} onFinish={save}>
            {loading ? <Loading styles={styles}/> : error ? (
                <Error styles={styles} message={error} onRetry={load}/>
            ) : <>
                <Form.Item label="当前时区" className={styles.formField}>
                    <Input value={timezone || "—"} readOnly/>
                </Form.Item>
                <Form.Item label="服务器时间" className={styles.formField}>
                    <Input value={serverTime || "—"} readOnly/>
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
                <Actions styles={styles} saving={saving} onReset={reset}/>
            </>}
        </Form>
    );
};
