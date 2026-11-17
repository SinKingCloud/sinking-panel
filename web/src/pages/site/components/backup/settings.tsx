import {forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Col, Form, Grid, InputNumber, Row, Select, Spin} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import {getSiteBackupSettingsInfo, updateSiteBackupSettings} from "@/service/api/site";
import Schedule from "@/pages/task/components/schedule";
import OssSelect from "@/pages/components/oss-select";

export interface BackupSettingsRef {
    open: () => void;
}

const SettingsForm = ({settings, saving, onFinish, form}: any) => {
    const upload = Form.useWatch("upload_oss", form) ?? Boolean(settings.upload_oss);
    const ossId = Form.useWatch("oss_id", form) ?? settings.oss_id;
    return <Form form={form} initialValues={settings} layout="vertical" variant="filled"
                 disabled={saving} clearOnDestroy onFinish={onFinish}>
        <Row gutter={16}>
            <Col xs={24} sm={12}>
                <Form.Item name="enabled" label="自动备份"
                           getValueProps={(value) => ({value: value == null ? undefined : Number(value)})}
                           getValueFromEvent={(value) => value === 1}
                           rules={[{required: true, type: "boolean", message: "请选择自动备份状态"}]}>
                    <Select placeholder="请选择自动备份状态" options={[
                        {label: "开启", value: 1},
                        {label: "关闭", value: 0},
                    ]}/>
                </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
                <Form.Item name="retries" label="重试次数" rules={[{required: true, message: "请选择重试次数"}]}>
                    <Select placeholder="请选择重试次数" options={Array.from({length: 11}, (_, value) => ({
                        value, label: value === 0 ? "不重试" : `${value} 次`,
                    }))}/>
                </Form.Item>
            </Col>
        </Row>
        <Row gutter={16}>
            <Col xs={24} sm={12}>
                <Form.Item name="local_keep" label="本地保留副本" dependencies={["upload_oss"]}
                           rules={[{required: true, type: "integer", min: upload ? 0 : 1, max: 1000,
                               message: upload ? "请输入 0 到 1000 份" : "请输入 1 到 1000 份"}]}>
                    <InputNumber min={upload ? 0 : 1} max={1000} precision={0} suffix="份"
                                 placeholder="输入保留份数" style={{width: "100%"}}/>
                </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
                <Form.Item name="upload_oss" label="云端存储" rules={[{required: true, type: "boolean", message: "请选择是否上传"}]}
                           getValueProps={(value) => ({value: value == null ? undefined : Number(value)})}
                           getValueFromEvent={(value) => value === 1}>
                    <Select placeholder="请选择是否上传" options={[{label: "不上传", value: 0}, {label: "上传至云端", value: 1}]}
                            onChange={(value) => {
                                if (value === 0 && !form.getFieldValue("local_keep")) form.setFieldValue("local_keep", 3);
                            }}/>
                </Form.Item>
            </Col>
        </Row>
        {upload && <>
            <Form.Item name="oss_id" label="存储实例" rules={[{required: true, type: "number", min: 1, message: "请选择存储实例"}]}>
                <OssSelect emptyValue={0} placeholder="选择存储实例" zIndex={1160}
                           onChange={(id) => { if (String(id) !== String(ossId)) form.setFieldValue("prefix", "/backups"); }}/>
            </Form.Item>
            <Form.Item name="prefix" label="备份目录" rules={[
                {required: true, transform: (value) => value === "" ? "/" : value, message: "请选择备份目录"},
                {max: 512, message: "备份目录不能超过 512 个字符"},
            ]} getValueProps={(value) => ({value: value == null ? [] : [String(value).replace(/^\/+/, "")]})}
                       getValueFromEvent={(paths: string[]) => paths.length ? `/${paths[0]}` : undefined}>
                <OssSelect mode="directory" ossId={ossId} placeholder={Number(ossId) > 0 ? "/backups" : "请先选择存储实例"}
                           maxLength={512} zIndex={1160}/>
            </Form.Item>
            <Form.Item name="oss_keep" label="云端保留副本" rules={[
                {required: true, type: "integer", min: 1, max: 1000, message: "请输入 1 到 1000 份"},
            ]}>
                <InputNumber min={1} max={1000} precision={0} suffix="份" placeholder="输入保留份数" style={{width: "100%"}}/>
            </Form.Item>
        </>}
        <Form.Item name="spec" label="执行周期" rules={[{required: true, whitespace: true, message: "请设置执行周期"}]}>
            <Schedule/>
        </Form.Item>
    </Form>;
};

const BackupSettings = forwardRef<BackupSettingsRef>((_, ref) => {
    const {message} = App.useApp();
    const screens = Grid.useBreakpoint();
    const compact = Boolean(useTheme()?.isCompactTheme?.());
    const [form] = Form.useForm();
    const requestRef = useRef(0);
    const savingRef = useRef(false);
    const [open, setOpen] = useState(false);
    const [settings, setSettings] = useState<any>();
    const [saving, setSaving] = useState(false);

    useEffect(() => () => { requestRef.current++; }, []);
    const close = useCallback(() => {
        if (savingRef.current) return;
        requestRef.current++;
        setOpen(false);
    }, []);
    useImperativeHandle(ref, () => ({open: () => {
        if (open || savingRef.current) return;
        requestRef.current++;
        setSettings(undefined);
        setOpen(true);
    }}), [open]);

    const submit = async (values: any) => {
        if (savingRef.current || !settings) return;
        const request = requestRef.current;
        savingRef.current = true;
        setSaving(true);
        try {
            const response = await updateSiteBackupSettings({body: {...values, spec: values.spec.trim()}});
            if (request !== requestRef.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "保存备份设置失败");
            savingRef.current = false;
            close();
            message.success(response?.message || "备份设置已保存");
        } catch (error) {
            if (request === requestRef.current) message.error(error instanceof Error ? error.message : "保存备份设置失败");
        } finally {
            savingRef.current = false;
            setSaving(false);
        }
    };

    return <ProModal title={<Title>备份设置</Title>} width={compact ? 430 : 460}
                     onOk={() => { if (settings) form.submit(); }} onCancel={close} modalProps={{
        open,
        afterOpenChange: async (visible: boolean) => {
            if (!visible || !open) return;
            const request = requestRef.current;
            try {
                const response = await getSiteBackupSettingsInfo();
                if (request !== requestRef.current) return;
                if (response?.code !== 200 || !response?.data) throw new Error(response?.message || "读取备份设置失败");
                setSettings({
                    ...response.data,
                    enabled: response?.data?.enabled ?? true,
                    upload_oss: Boolean(response.data.upload_oss),
                });
            } catch (error) {
                if (request !== requestRef.current) return;
                close();
                message.error(error instanceof Error ? error.message : "读取备份设置失败");
            }
        },
        zIndex: 1100,
        destroyOnHidden: true,
        confirmLoading: saving,
        okText: "保存",
        cancelText: "取消",
        footer: settings ? undefined : null,
        cancelButtonProps: {disabled: saving},
        closable: !saving,
        keyboard: !saving,
        mask: {closable: !saving},
        style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"},
    } as any}>
        {settings ? <SettingsForm form={form} settings={settings} saving={saving} onFinish={submit}/>
            : <div style={{height: 240, display: "flex", alignItems: "center", justifyContent: "center"}}
                   role="status" aria-label="正在读取备份设置"><Spin/></div>}
    </ProModal>;
});

BackupSettings.displayName = "BackupSettings";
export default BackupSettings;
