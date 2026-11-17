import {useLayoutEffect, useRef, useState} from "react";
import {App, Button, Collapse, Form, Grid, Input, InputNumber, Select, Space, Spin} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import FileSelect from "@/pages/components/file-select";
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
    // useWatch 的通知会延后；订阅字段变化，但布局始终读取本次已写入的配置。
    Form.useWatch("config", {form, preserve: true});
    const requestRef = useRef(0);
    const busyRef = useRef(false);
    const testedRef = useRef("");
    const lastType = useRef(record?.type);
    if (open && record) lastType.current = record.type;
    const type = Number(lastType.current) || 0;
    const [loadedId, setLoadedId] = useState<number>();
    const [entered, setEntered] = useState(false);
    const config = entered ? form.getFieldValue("config") : undefined;
    const [loading, setLoading] = useState(true);
    const [busy, setBusy] = useState("");
    const [advanced, setAdvanced] = useState<string[]>([]);
    const fields = connectionFields[type] || [];
    const waiting = loading || (open && (!entered || loadedId !== record?.id));
    const unavailable = waiting || !fields.length;

    useLayoutEffect(() => {
        const requestId = ++requestRef.current;
        if (!open) { setEntered(false); return; }
        // 每次打开先清除上一轮的就绪状态，避免入场结束时短暂露出旧表单。
        setLoadedId(undefined);
        setLoading(true);
        // 先显示弹窗和 loading，再初始化表单，避免重置字段阻塞入场绘制。
        if (!entered || !record?.id) return;
        busyRef.current = false;
        testedRef.current = "";
        setBusy("");
        form.resetFields();
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
    }, [open, entered, record?.id]);

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
        testedRef.current = "";
        try {
            await form.validateFields(fields.map((field) => ["config", field.name]));
            if (requestId !== requestRef.current) return;
            const body = connectionBody();
            const response = await testDatabase({body});
            if (requestId !== requestRef.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "连接失败");
            testedRef.current = JSON.stringify(body);
            message.success("连接成功");
        } catch (reason: any) {
            if (requestId !== requestRef.current) return;
            if (reason?.errorFields) {
                if (reason?.errorFields.some((item: any) => fields.some((field) => field.advanced && field.name === item.name?.[1]))) setAdvanced(["connection"]);
            } else {
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
                const response = await testDatabase({body: connection});
                if (requestId !== requestRef.current) return;
                if (response?.code !== 200) {
                    message.error(response?.message || "连接失败");
                    return;
                }
                testedRef.current = snapshot;
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
            case "path": input = <FileSelect type="file" placeholder={placeholder}/>; break;
            default: input = <Input autoComplete="off" placeholder={placeholder}/>;
        }
        return <Form.Item key={field.name} name={["config", field.name]} label={field.label}
            hidden={field.visible ? !field.visible(config) : false}
            className={field.wide || field.type === "path" ? "db-field-wide" : undefined}
            getValueProps={(value) => ({value: field.type === "boolean" ? (value ? 1 : 0) : field.type === "path" ? (value ? [value] : []) : field.valueAliases?.[value] ?? value})}
            getValueFromEvent={field.type === "boolean" ? (value) => value === 1 : field.type === "path" ? (paths: string[]) => paths[0] || "" : undefined}
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
            // 关闭动画结束后销毁折叠面板，避免下次打开补播上次的收起动画。
            destroyOnHidden: true,
            afterOpenChange: setEntered,
            style: {top: screens.md ? 90 : 24, paddingBottom: screens.md ? 90 : 24, maxWidth: "calc(100vw - 24px)"},
            styles: type === 2 ? {body: {maxHeight: "none", overflow: "visible"}} : undefined,
            closable: !busy, keyboard: !busy, mask: {closable: !busy},
            focusable: {focusTriggerAfterClose: false}, okText: "保存", cancelText: "取消",
            confirmLoading: busy === "save", okButtonProps: {disabled: !fields.length || busy === "test"},
            // 加载时只隐藏底部，避免显示时从禁用样式过渡到正常样式。
            footer: (_: any, {OkBtn, CancelBtn}: any) => <div aria-hidden={waiting}
                style={{display: "flex", alignItems: "center", justifyContent: "space-between", gap: 12, visibility: waiting ? "hidden" : "visible"}}>
                <Button loading={busy === "test"} disabled={!fields.length || Boolean(busy)}
                    onClick={() => void testConnection()}>测试连接</Button>
                <Space><CancelBtn/><OkBtn/></Space>
            </div>,
            cancelButtonProps: {disabled: Boolean(busy)},
        } as any}>
        <Form form={form} layout="vertical" variant="filled"
            className={formStyles.form}
            style={{position: "relative"}} disabled={Boolean(busy)} onFinish={submit}
            onFinishFailed={({errorFields}) => {
                if (errorFields.some((item) => fields.some((field) => field.advanced && field.name === item.name?.[1]))) setAdvanced(["connection"]);
            }}
            onValuesChange={(changes) => { if (changes.config) testedRef.current = ""; }}>
            {waiting && <div className={styles.loading} style={{position: "absolute", inset: 0}}><Spin/></div>}
            {/* 加载时隐藏字段以保留自然高度，不切换禁用态，避免显示时边框过渡闪烁。 */}
            <div className="db-form-fields" style={{visibility: waiting ? "hidden" : "visible"}} aria-hidden={waiting}>
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
            </div>
        </Form>
    </ProModal>;
};
