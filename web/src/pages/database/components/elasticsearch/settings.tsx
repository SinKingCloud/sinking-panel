import {useLayoutEffect, useRef, useState} from "react";
import {App, Form, Input, InputNumber, Spin} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import type {DatabasePanelRecord, DatabasePanelRequest, DatabaseVersion} from "../shared/entities";
import useStyles from "../shared/styles/client";
import useFormStyles from "../shared/manager-form-styles";

export default ({open, instance, name, request, version, onCancel, onSuccess}: {
    open: boolean; instance?: DatabasePanelRecord; name?: string; request: DatabasePanelRequest; version?: DatabaseVersion;
    onCancel: () => void; onSuccess?: () => void;
}) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const {styles: formStyles} = useFormStyles();
    const theme = useTheme();
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [resultWindow, setResultWindow] = useState(true);
    const sequence = useRef(0);
    const pending = useRef(false);

    useLayoutEffect(() => {
        const current = ++sequence.current;
        if (!open) return;
        pending.current = false;
        setSaving(false);
        setLoading(true);
        form.resetFields();
        const parts = (version?.version || "").split(".").map(Number);
        setResultWindow(version?.features?.max_result_window ?? (!version?.version || parts[0] > 2 || parts[0] === 2 && parts[1] >= 1));
        void (async () => {
            try {
                if (!instance?.id || !name) throw new Error("请选择索引");
                const response = await request("settings", {id: instance.id, database_name: name});
                if (current !== sequence.current) return;
                if (response?.code !== 200 || !response.data || typeof response.data !== "object" || Array.isArray(response.data)) {
                    throw new Error(response?.message || "读取索引设置失败");
                }
                const data = response.data[name] || response.data;
                const settings = data.settings?.index || data.settings || data.index || data;
                form.setFieldsValue({
                    shards: settings.number_of_shards ?? settings["index.number_of_shards"] ?? "—",
                    replicas: Number(settings.number_of_replicas ?? settings["index.number_of_replicas"] ?? 0),
                    refresh_interval: settings.refresh_interval ?? settings["index.refresh_interval"] ?? "1s",
                    max_result_window: Number(settings.max_result_window ?? settings["index.max_result_window"] ?? 10000),
                });
                setLoading(false);
            } catch (reason: any) {
                if (current !== sequence.current) return;
                onCancel();
                message.error(reason.message || "读取索引设置失败");
            }
        })();
        return () => { sequence.current++; };
    }, [open, instance?.id, name]);

    const submit = async (values: any) => {
        if (!open || !instance?.id || !name || loading || pending.current) return;
        const current = sequence.current;
        pending.current = true;
        setSaving(true);
        try {
            const settings: Record<string, any> = {number_of_replicas: values.replicas, refresh_interval: values.refresh_interval.trim()};
            if (resultWindow) settings.max_result_window = values.max_result_window;
            const response = await request("update_settings", {id: instance.id, database_name: name, settings: JSON.stringify({index: settings})});
            if (current !== sequence.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "保存失败");
            message.success("保存成功");
            onSuccess?.();
            onCancel();
        } catch (reason: any) {
            if (current === sequence.current) message.error(reason.message || "保存失败");
        } finally {
            if (current === sequence.current) { pending.current = false; setSaving(false); }
        }
    };

    return <ProModal title={<Title>修改索引</Title>} onOk={form.submit} onCancel={() => { if (!pending.current) onCancel(); }} modalProps={{
        open, width: theme?.isCompactTheme?.() ? 400 : 440, zIndex: 1100, rootClassName: styles.modal,
        style: {maxWidth: "calc(100vw - 24px)"}, styles: {body: {maxHeight: "none", overflow: "visible"}},
        okText: "保存", cancelText: "取消", confirmLoading: saving, okButtonProps: {disabled: loading},
        cancelButtonProps: {disabled: saving}, closable: !saving, keyboard: !saving, mask: {closable: !saving},
        forceRender: true, focusable: {focusTriggerAfterClose: false},
    }}>
        <Form form={form} layout="vertical" variant="filled" onFinish={submit} disabled={saving || loading} className={`${styles.form} ${formStyles.form}`}>
            {loading ? <div className={styles.loading} style={{paddingBlock: 48}}><Spin/></div> : <div className="db-form-grid">
                <Form.Item name="shards" label="主分片数"><Input readOnly placeholder="主分片数"/></Form.Item>
                <Form.Item name="replicas" label="副本数" rules={[{required: true, message: "请输入副本数"}]}>
                    <InputNumber min={0} max={1024} precision={0} placeholder="请输入副本数" style={{width: "100%"}}/>
                </Form.Item>
                <Form.Item name="refresh_interval" label="刷新间隔" rules={[{required: true, whitespace: true, message: "请输入刷新间隔"},
                    {pattern: /^(?:-1|\d+(?:ms|s|m|h|d))$/, message: "例如 1s、30s，-1 表示关闭"}]}>
                    <Input placeholder="例如 1s，-1 表示关闭"/>
                </Form.Item>
                {resultWindow && <Form.Item name="max_result_window" label="最大分页结果" rules={[{required: true, message: "请输入结果上限"}]}>
                    <InputNumber min={1} max={1000000} precision={0} placeholder="请输入结果上限" style={{width: "100%"}}/>
                </Form.Item>}
            </div>}
        </Form>
    </ProModal>;
};
