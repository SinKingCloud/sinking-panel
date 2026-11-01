import {forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Form, Grid, InputNumber, Select, Spin} from "antd";
import {ProModal, Title} from "sinking-antd";
import type {ProModalRef} from "sinking-antd";
import {certSettings} from "@/service/api/cert";
import Schedule from "@/pages/task/components/schedule";

export interface CertSettingsRef {
    open: () => void;
    close: () => void;
}

interface Settings {
    enabled: boolean;
    spec: string;
    retries: number;
}

const CertSettings = forwardRef<CertSettingsRef>((_, ref) => {
    const {message} = App.useApp();
    const screens = Grid.useBreakpoint();
    const [form] = Form.useForm<Settings>();
    const modalRef = useRef<ProModalRef>(null);
    const requestRef = useRef(0);
    const savingRef = useRef(false);
    const [settings, setSettings] = useState<Settings>();
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);

    useEffect(() => () => { requestRef.current++; }, []);

    const close = useCallback(() => {
        if (savingRef.current) return;
        requestRef.current++;
        modalRef.current?.hide();
    }, []);

    useImperativeHandle(ref, () => ({
        close,
        open: () => {
            if (savingRef.current) return;
            const request = ++requestRef.current;
            setSettings(undefined);
            setLoading(true);
            modalRef.current?.show();
            void (async () => {
                try {
                    const response = await certSettings({body: {action: "get"}});
                    if (request !== requestRef.current) return;
                    if (response?.code !== 200 || !response?.data) {
                        throw new Error(response?.message || "读取续签设置失败");
                    }
                    setSettings(response.data);
                    setLoading(false);
                } catch (error) {
                    if (request !== requestRef.current) return;
                    close();
                    message.error(error instanceof Error ? error.message : "读取续签设置失败");
                }
            })();
        },
    }), [close, message]);

    const submit = async (values: Settings) => {
        if (loading || !settings || savingRef.current) return;
        const request = requestRef.current;
        savingRef.current = true;
        setSaving(true);
        try {
            const response = await certSettings({body: {...values, spec: values.spec.trim(), action: "set"}});
            if (request !== requestRef.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "保存续签设置失败");
            savingRef.current = false;
            close();
            message.success(response?.message || "续签设置已保存");
        } catch (error) {
            if (request === requestRef.current) {
                message.error(error instanceof Error ? error.message : "保存续签设置失败");
            }
        } finally {
            savingRef.current = false;
            setSaving(false);
        }
    };

    return <ProModal ref={modalRef} title={<Title>续签设置</Title>} width={560}
                     onOk={() => { if (!loading && settings) form.submit(); }} onCancel={close}
                     modalProps={{
                         zIndex: 1030,
                         confirmLoading: saving,
                         okText: "保存",
                         cancelText: "取消",
                         footer: loading || !settings ? null : undefined,
                         cancelButtonProps: {disabled: saving},
                         closable: !saving,
                         keyboard: !saving,
                         mask: {closable: !saving},
                         style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"},
                     }}>
        {loading ? <div style={{height: 240, display: "flex", alignItems: "center", justifyContent: "center"}}
                        role="status" aria-label="正在读取续签设置"><Spin/></div> : settings ?
            <Form form={form} initialValues={settings} clearOnDestroy layout="vertical" variant="filled"
                  disabled={saving} onFinish={submit}>
                <Form.Item name="enabled" label="自动续签"
                           getValueProps={(value) => ({value: value == null ? undefined : Number(value)})}
                           getValueFromEvent={(value) => value === 1}
                           rules={[{required: true, type: "boolean", message: "请选择自动续签状态"}]}>
                    <Select placeholder="请选择自动续签状态" options={[
                        {label: "开启", value: 1},
                        {label: "关闭", value: 0},
                    ]}/>
                </Form.Item>
                <Form.Item name="spec" label="执行周期"
                           rules={[{required: true, whitespace: true, message: "请设置执行周期"}]}>
                    <Schedule/>
                </Form.Item>
                <Form.Item name="retries" label="重试次数" extra="失败后额外重试，0 表示不重试。"
                           rules={[
                               {required: true, message: "请输入重试次数"},
                               {type: "integer", min: 0, max: 10, message: "请输入 0–10 之间的整数"},
                           ]}>
                    <InputNumber min={0} max={10} precision={0} placeholder="请输入重试次数" style={{width: "100%"}}/>
                </Form.Item>
            </Form> : null}
    </ProModal>;
});

CertSettings.displayName = "CertSettings";

export default CertSettings;
