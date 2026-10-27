import {useLayoutEffect, useRef, useState} from "react";
import {App, Button, Collapse, Form, Grid, Input, InputNumber, Select, Spin, Typography} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import FilePicker from "@/pages/components/file-picker";
import {getDatabaseInfo, testDatabase, updateDatabase} from "@/service/api/database";
import useStyles from "./components/shared/styles/client";
import useFormStyles from "./forms-styles";
import connectionFields from "./connection-fields";

// 根据数据库类型展示连接配置，敏感字段的密文原样回传服务端处理。
export default ({open, record, categories = [], onCancel, onSuccess}: any) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const {styles: formStyles} = useFormStyles();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const [form] = Form.useForm();
    const config = Form.useWatch("config", {form, preserve: true});
    const requestRef = useRef(0);
    const busyRef = useRef(false);
    const testedRef = useRef("");
    const lastType = useRef(record?.type);
    const type = Number(open ? record?.type : lastType.current) || 0;
    const [loadedId, setLoadedId] = useState<number>();
    const [loading, setLoading] = useState(true);
    const [busy, setBusy] = useState("");
    const [testResult, setTestResult] = useState<any>();
    const [advanced, setAdvanced] = useState<string[]>([]);
    const fields = connectionFields[type] || [];
    const waiting = loading || (open && loadedId !== record?.id);
    const unavailable = waiting || !fields.length;

    useLayoutEffect(() => {
        const requestId = ++requestRef.current;
        if (!open || !record?.id) return;
        lastType.current = record.type;
        busyRef.current = false;
        setLoadedId(undefined);
        testedRef.current = "";
        setBusy("");
        setTestResult(undefined);
        setAdvanced([]);
        form.resetFields();
        setLoading(true);
        void (async () => {
            try {
                const response = await getDatabaseInfo({body: {id: record.id}});
                if (requestId !== requestRef.current) return;
                if (response?.code !== 200 || !response?.data) throw new Error(response?.message || "获取实例信息失败");
                const info = response?.data;
                const typeFields = connectionFields[Number(info?.type)];
                if (!typeFields?.length || !info?.config || typeof info?.config !== "object" || Array.isArray(info?.config)) {
                    throw new Error("获取连接配置失败");
                }
                form.setFieldsValue({name: info?.name, type_id: Number(info?.type_id) || 0, config: info?.config});
                setLoadedId(record.id);
            } catch (reason: any) {
                if (requestId !== requestRef.current) return;
                requestRef.current++; onCancel();
                message.error(reason?.message || "获取实例信息失败");
            } finally {
                if (requestId === requestRef.current) setLoading(false);
            }
        })();
        return () => { requestRef.current += 1; };
    }, [open, record?.id]);

    const connectionBody = () => ({
        id: record.id,
        type,
        // 保留接口返回的完整配置，未展示的连接选项不会在保存时丢失。
        config: JSON.stringify(form.getFieldValue("config")),
    });

    const testConnection = async () => {
        if (busyRef.current || unavailable) return;
        const requestId = requestRef.current;
        busyRef.current = true;
        setBusy("test");
        setTestResult(undefined);
        testedRef.current = "";
        try {
            await form.validateFields(fields.map((field) => ["config", field.name]));
            const body = connectionBody();
            const response = await testDatabase({body});
            if (requestId !== requestRef.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "连接失败");
            testedRef.current = JSON.stringify(body);
            setTestResult({success: true, text: "连接成功", version: response?.data?.version});
        } catch (reason: any) {
            if (requestId !== requestRef.current) return;
            if (reason?.errorFields) {
                if (reason?.errorFields.some((item: any) => fields.some((field) => field.advanced && field.name === item.name?.[1]))) setAdvanced(["connection"]);
            } else {
                setTestResult({success: false, text: reason?.message || "连接失败"});
                message.error(reason?.message || "连接失败");
            }
        } finally {
            if (requestId === requestRef.current) { busyRef.current = false; setBusy(""); }
        }
    };

    const submit = async (values: any) => {
        if (busyRef.current || unavailable) return;
        const requestId = requestRef.current;
        busyRef.current = true;
        setBusy("save");
        try {
            const connection = connectionBody();
            const snapshot = JSON.stringify(connection);
            if (testedRef.current !== snapshot) {
                setTestResult(undefined);
                const response = await testDatabase({body: connection});
                if (requestId !== requestRef.current) return;
                if (response?.code !== 200) {
                    setTestResult({success: false, text: response?.message || "连接失败"});
                    message.error(response?.message || "连接失败");
                    return;
                }
                testedRef.current = snapshot;
                setTestResult({success: true, text: "连接成功", version: response?.data?.version});
            }
            const response = await updateDatabase({body: {id: record.id, name: values.name.trim(), type_id: Number(values.type_id) || 0, config: connection.config}});
            if (requestId !== requestRef.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "保存实例失败");
            message.success("连接配置已保存");
            onSuccess({action: "update", id: record.id, name: values.name.trim(), type_id: Number(values.type_id) || 0});
            onCancel();
        } catch (reason: any) {
            if (requestId === requestRef.current) message.error(reason?.message || "保存实例失败");
        } finally {
            if (requestId === requestRef.current) { busyRef.current = false; setBusy(""); }
        }
    };

    const renderField = (field: any) => {
        const placeholder = field.placeholder || `${["select", "boolean"].includes(field.type) ? "请选择" : field.type === "path" ? "请输入或选择" : "请输入"}${field.label}`;
        let input;
        switch (field.type) {
            case "password": input = <Input.Password autoComplete="new-password" placeholder={placeholder}/>; break;
            case "number": input = <InputNumber min={field.min} max={field.max} precision={0} placeholder={placeholder} style={{width: "100%"}}/>; break;
            case "select": input = <Select options={field.options} placeholder={placeholder}/>; break;
            case "boolean": input = <Select placeholder={placeholder} options={[{value: 0, label: "关闭"}, {value: 1, label: "开启"}]}/>; break;
            case "path": input = <FilePicker mode="file" placeholder={placeholder}/>; break;
            default: input = <Input autoComplete="off" placeholder={placeholder}/>;
        }
        return <Form.Item key={field.name} name={["config", field.name]} label={field.label}
            hidden={field.visible ? !field.visible(config) : false}
            className={field.wide || field.type === "path" ? "db-field-wide" : undefined}
            getValueProps={(value) => ({value: field.type === "boolean" ? (value ? 1 : 0) : field.valueAliases?.[value] ?? value})}
            getValueFromEvent={field.type === "boolean" ? (value) => value === 1 : undefined}
            extra={typeof field.description === "function" ? field.description(config) : field.description}
            rules={[{required: Boolean(field.required), message: `请填写${field.label}`}]}>
            {input}
        </Form.Item>;
    };
    const basicFields = fields.filter((field) => !field.advanced);
    const addressFields = basicFields.filter((field) => ["host", "port", "path"].includes(field.name));
    const otherFields = basicFields.filter((field) => !["host", "port", "path"].includes(field.name));
    const moreFields = fields.filter((field) => field.advanced);

    return <ProModal title={<Title>修改信息</Title>}
        onCancel={() => { if (!busyRef.current) onCancel(); }} onOk={form.submit}
        afterClose={() => setAdvanced([])}
        modalProps={{
            open, width: compact ? 520 : 560, zIndex: 1140, rootClassName: styles.modal,
            style: {top: screens.md ? 90 : 24, paddingBottom: screens.md ? 90 : 24, maxWidth: "calc(100vw - 24px)"},
            styles: type === 2 ? {body: {maxHeight: "none", overflow: "visible"}} : undefined,
            closable: !busy, keyboard: !busy, mask: {closable: !busy}, forceRender: true,
            focusable: {focusTriggerAfterClose: false}, okText: "保存", cancelText: "取消",
            confirmLoading: busy === "save", okButtonProps: {disabled: unavailable || busy === "test"},
            cancelButtonProps: {disabled: Boolean(busy)},
        } as any}>
        <Form form={form} layout="vertical" variant="filled"
            className={`${styles.form} ${formStyles.form}`}
            style={{minHeight: type === 2 ? undefined : 285, position: "relative"}} disabled={Boolean(busy) || waiting} onFinish={submit}
            onFinishFailed={({errorFields}) => {
                if (errorFields.some((item) => fields.some((field) => field.advanced && field.name === item.name?.[1]))) setAdvanced(["connection"]);
            }}
            onValuesChange={(changes) => { if (changes.config) { testedRef.current = ""; setTestResult(undefined); } }}>
            {waiting && <div className={styles.loading} style={{position: "absolute", inset: 0}}><Spin/></div>}
            {/* 加载期间保留对应类型的表单尺寸，避免入场动画中途改变高度。 */}
            <div style={{visibility: waiting ? "hidden" : "visible"}} aria-hidden={waiting}>
                <div className="db-form-grid">
                    <Form.Item name="name" label="实例名称" rules={[{required: true, whitespace: true, message: "请输入实例名称"}, {max: 100, message: "名称不能超过 100 个字符"}]}>
                        <Input maxLength={100} placeholder="例如：生产数据库" autoComplete="off"/>
                    </Form.Item>
                    <Form.Item name="type_id" label="分类"><Select placeholder="请选择分类" options={[{value: 0, label: "全部分类"}, ...categories.map((item: any) => ({value: Number(item.id), label: item.name}))]}/></Form.Item>
                </div>
                {addressFields.length > 0 && <div className={`db-form-grid ${type !== 6 ? "db-address-grid" : ""}`}>{addressFields.map(renderField)}</div>}
                {otherFields.length > 0 && <div className="db-form-grid">{otherFields.map(renderField)}</div>}
                {moreFields.length > 0 && <Collapse size="small" bordered={false} className="db-form-more"
                    activeKey={advanced} onChange={(keys) => setAdvanced(Array.isArray(keys) ? keys : [keys])}
                    items={[{key: "connection", label: "更多选项", forceRender: true,
                        children: <div className="db-form-grid">{moreFields.map(renderField)}</div>}]}/>}
                <div className="db-connection-feedback" role="status" aria-live="polite">
                    <Button color="default" variant="filled" loading={busy === "test"} disabled={unavailable || Boolean(busy)}
                        onClick={() => void testConnection()}>测试连接</Button>
                    {testResult && <div className={`db-test-result ${testResult.success ? "success" : "error"}`}>
                        <Icon type={testResult.success ? "CheckCircleOutlined" : "CloseCircleOutlined"}/>
                        <Typography.Text ellipsis title={testResult.text} type={testResult.success ? "success" : "danger"}>{testResult.text}</Typography.Text>
                        {testResult.success && testResult.version && <Typography.Text className="db-test-version" ellipsis title={String(testResult.version)} type="secondary">{String(testResult.version).slice(0, 160)}</Typography.Text>}
                    </div>}
                </div>
            </div>
        </Form>
    </ProModal>;
};
