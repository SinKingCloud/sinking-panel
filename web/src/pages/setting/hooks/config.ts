import {useCallback, useEffect, useRef, useState} from "react";
import {App, Form} from "antd";
import {useModel} from "umi";
import {getConfig, setConfig} from "@/service/api/config";

const serialize = (value: any) => {
    if (typeof value === "boolean") {
        return value ? "1" : "0";
    }
    return String(value ?? "");
};

export default (group: string, defaults: any, normalize: (values: any) => any = (values) => values) => {
    const {message} = App.useApp();
    const web = useModel("web");
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState("");
    const mountedRef = useRef(true);
    const savingRef = useRef(false);
    const requestRef = useRef(0);
    const saveRequestRef = useRef(0);
    const initialRef = useRef<any>({});
    const defaultsRef = useRef(defaults);
    const normalizeRef = useRef(normalize);

    defaultsRef.current = defaults;
    normalizeRef.current = normalize;

    const load = useCallback(async () => {
        const requestId = ++requestRef.current;
        setLoading(true);
        setError("");
        const response = await getConfig({body: {group}});
        if (!mountedRef.current || requestId !== requestRef.current) {
            return;
        }

        if (response?.code !== 200) {
            setError(response?.message || "配置加载失败");
            setLoading(false);
            return;
        }
        const source: any = response?.data || {};
        const values = Object.keys(defaultsRef.current).reduce((result: any, key) => {
            const configKey = `${group}.${key}`;
            const value = source[configKey];
            result[key] = value === undefined ? defaultsRef.current[key] : value;
            return result;
        }, {});
        const nextValues = normalizeRef.current(values);
        initialRef.current = nextValues;
        form.setFieldsValue(nextValues);
        setLoading(false);
    }, [form, group]);

    useEffect(() => {
        mountedRef.current = true;
        load();
        return () => {
            mountedRef.current = false;
            requestRef.current += 1;
            saveRequestRef.current += 1;
        };
    }, [load]);

    const save = useCallback(async (values: any) => {
        if (savingRef.current) {
            return false;
        }
        savingRef.current = true;
        const requestId = ++saveRequestRef.current;
        setSaving(true);
        const nextValues = normalizeRef.current(values);
        const response = await setConfig({
            body: {
                configs: Object.entries(nextValues).map(([key, value]) => ({
                    key: `${group}.${key}`,
                    value: serialize(value),
                })),
            },
        });
        if (!mountedRef.current || requestId !== saveRequestRef.current) {
            return false;
        }
        if (response?.code !== 200) {
            message.error(response?.message || "配置保存失败");
            savingRef.current = false;
            setSaving(false);
            return false;
        }

        message.success(response?.message || "配置保存成功");
        await load();
        const latest = await web?.getInfo?.();
        if (mountedRef.current && requestId === saveRequestRef.current && latest) {
            web?.setInfo?.(latest);
        }
        if (mountedRef.current && requestId === saveRequestRef.current) {
            savingRef.current = false;
            setSaving(false);
        }
        return true;
    }, [group, load, message, web]);

    const reset = useCallback(() => {
        form.setFieldsValue(initialRef.current);
    }, [form]);

    return {form, loading, saving, error, load, save, reset};
};
