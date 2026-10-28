import React, {forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Button, Form, InputNumber, Select, Spin} from "antd";
import {Icon, ProModal, Title} from "sinking-antd";
import type {ProModalRef} from "sinking-antd";
import {systemMonitor} from "@/service/api/system";
import {formatSize} from "@/utils/string";
import type {MonitorSettings} from "../../types";
import useStyles from "./styles";

export type MonitorSettingsRef = {open: () => void};

type Props = {
    onLoaded?: (settings: MonitorSettings) => void;
    onSuccess?: (settings: MonitorSettings) => void;
};

type Values = Pick<MonitorSettings, "enabled" | "retention_days">;

const MonitorSettingsModal = forwardRef<MonitorSettingsRef, Props>(({onLoaded, onSuccess}, ref) => {
    const {styles} = useStyles();
    const {message} = App.useApp();
    const [form] = Form.useForm<Values>();
    const modalRef = useRef<ProModalRef>(null);
    const mountedRef = useRef(false);
    const openRef = useRef(false);
    const requestRef = useRef(0);
    const savingRef = useRef(false);
    const loadedRef = useRef(false);
    const callbacksRef = useRef({onLoaded, onSuccess});
    callbacksRef.current = {onLoaded, onSuccess};
    const [settings, setSettings] = useState<MonitorSettings>();
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [loadError, setLoadError] = useState("");
    const [saveError, setSaveError] = useState("");

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            openRef.current = false;
            requestRef.current += 1;
        };
    }, []);

    const load = useCallback(async () => {
        if (!openRef.current || savingRef.current) return;
        const request = ++requestRef.current;
        loadedRef.current = false;
        setLoading(true);
        setLoadError("");
        setSaveError("");
        setSettings(undefined);
        form.resetFields();
        try {
            const response = await systemMonitor({body: {action: "settings"}});
            if (!mountedRef.current || !openRef.current || request !== requestRef.current) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "读取监控设置失败");
            const next = response.data as MonitorSettings;
            setSettings(next);
            form.setFieldsValue({enabled: next.enabled, retention_days: next.retention_days});
            loadedRef.current = true;
            callbacksRef.current.onLoaded?.(next);
        } catch (error) {
            if (mountedRef.current && openRef.current && request === requestRef.current) {
                setLoadError(error instanceof Error ? error.message : "读取监控设置失败");
            }
        } finally {
            if (mountedRef.current && openRef.current && request === requestRef.current) setLoading(false);
        }
    }, [form]);

    const close = useCallback(() => {
        if (savingRef.current) return;
        openRef.current = false;
        loadedRef.current = false;
        requestRef.current += 1;
        modalRef.current?.hide();
    }, []);

    useImperativeHandle(ref, () => ({
        open: () => {
            // 写入进行中时保持同一会话，避免新读取早于上一次保存提交。
            if (!mountedRef.current || savingRef.current) return;
            openRef.current = true;
            modalRef.current?.show();
            void load();
        },
    }), [load]);

    const submit = async (values: Values) => {
        if (savingRef.current || !loadedRef.current || !openRef.current) return;
        const request = requestRef.current;
        savingRef.current = true;
        setSaving(true);
        setSaveError("");
        try {
            const response = await systemMonitor({body: {action: "set", enabled: values.enabled, retention_days: values.retention_days}});
            if (!mountedRef.current || !openRef.current || request !== requestRef.current) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "保存监控设置失败");
            const next = response.data as MonitorSettings;
            savingRef.current = false;
            setSaving(false);
            close();
            message.success(response.message || "监控设置已保存");
            callbacksRef.current.onSuccess?.(next);
        } catch (error) {
            if (mountedRef.current && openRef.current && request === requestRef.current) {
                setSaveError(error instanceof Error ? error.message : "保存监控设置失败");
            }
        } finally {
            if (mountedRef.current && openRef.current && request === requestRef.current) {
                savingRef.current = false;
                setSaving(false);
            }
        }
    };

    return (
        <ProModal ref={modalRef} title={<Title>监控设置</Title>} width={420}
                  onOk={() => { if (loadedRef.current && !savingRef.current) form.submit(); }}
                  onCancel={close}
                  modalProps={{
                      forceRender: true,
                      confirmLoading: saving,
                      okText: "保存",
                      cancelText: "取消",
                      footer: loading || loadError ? null : undefined,
                      okButtonProps: {disabled: loading || !settings || Boolean(loadError)},
                      cancelButtonProps: {disabled: saving},
                      closable: !saving,
                      keyboard: !saving,
                      mask: {closable: !saving},
                      afterClose: () => {
                          if (openRef.current) return;
                          form.resetFields();
                          setSettings(undefined);
                          setLoading(false);
                          setLoadError("");
                          setSaveError("");
                      },
                  } as any}>
            <Form form={form} layout="vertical" variant="filled" requiredMark
                  disabled={loading || saving} onFinish={submit}>
                {loading ? <div className={styles.state} role="status" aria-label="正在读取监控设置"><Spin/></div>
                    : loadError ? <div className={styles.state}>
                        <Icon type="ExclamationCircleOutlined" className="state-icon"/>
                        <span className="state-error" role="alert">{loadError}</span>
                        <Button color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} onClick={() => void load()}>重新加载</Button>
                    </div> : settings ? <>
                        <div className={styles.storage}>
                            <span className="storage-label"><Icon type="DatabaseOutlined"/>历史数据占用</span>
                            <span className="storage-value">{formatSize(settings.storage_bytes)}</span>
                        </div>
                        <Form.Item name="enabled" label="监控采集"
                                   getValueProps={(value) => ({value: value == null ? undefined : Number(value)})}
                                   getValueFromEvent={(value) => value === 1}
                                   rules={[{required: true, type: "boolean", message: "请选择监控采集状态"}]}>
                            <Select placeholder="请选择采集状态" options={[
                                {label: "开启", value: 1},
                                {label: "关闭", value: 0},
                            ]}/>
                        </Form.Item>
                        <Form.Item name="retention_days" label="历史保留时间"
                                   rules={[
                                       {required: true, message: "请输入历史保存天数"},
                                       {type: "integer", min: 1, max: 365, message: "请输入 1–365 之间的整数"},
                                   ]}>
                            <InputNumber min={1} max={365} precision={0} suffix="天" placeholder="请输入保留天数"
                                         aria-label="历史保留时间" style={{width: "100%"}}/>
                        </Form.Item>
                        {settings.last_error && <div className={styles.warning} role="status">{settings.last_error}</div>}
                        {saveError && <div className={styles.error} role="alert">{saveError}</div>}
                    </> : null}
            </Form>
        </ProModal>
    );
});

MonitorSettingsModal.displayName = "MonitorSettingsModal";

export default MonitorSettingsModal;
