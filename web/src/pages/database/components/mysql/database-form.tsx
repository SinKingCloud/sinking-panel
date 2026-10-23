import {useLayoutEffect, useRef, useState} from "react";
import {App, Form, Input, Select, Spin} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import {useMysqlClient} from "./context";
import useStyles from "./component-styles/client";
import useFormStyles from "./form-styles";

// 创建数据库或调整默认字符集，直接操作当前 MySQL 实例。
export default ({open, instance, name, onCancel, onSuccess}: any) => {
    const {api} = useMysqlClient();
    const {message} = App.useApp();
    const {styles} = useStyles();
    const {styles: formStyles} = useFormStyles();
    const theme = useTheme();
    const [form] = Form.useForm();
    const lastName = useRef(name);
    const editing = (open ? name : lastName.current) !== undefined;
    const sequence = useRef(0);
    const pending = useRef(false);
    const original = useRef<any>({});
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [visible, setVisible] = useState(false);
    const [options, setOptions] = useState<any>();
    const charset = Form.useWatch("charset", {form, preserve: true});
    const characterSets = options?.charsets || [];
    const collations = characterSets.find((item: any) => item.name === charset)?.collations || [];
    // 显示前重置加载状态，重新打开同一个数据库时也不会闪现上次的表单。
    useLayoutEffect(() => {
        const request = ++sequence.current;
        if (!open || !instance) return;
        setVisible(true);
        lastName.current = name;
        pending.current = false; setSaving(false); form.resetFields(); original.current = {};
        setOptions(undefined); setLoading(true);
        void Promise.all([
            api.charsets(instance),
            editing ? api.database_info(instance, {database_name: name}) : Promise.resolve(undefined),
        ]).then(([options, data]) => {
            if (request !== sequence.current) return;
            const current = editing ? data : options;
            if (!current || !options?.charsets?.length) throw new Error("获取数据库字符集和排序规则失败");
            const selected = options.charsets.find((item: any) => item.name === current.charset)
                || options.charsets.find((item: any) => item.collations.includes(current.collation));
            if (!selected?.collations.includes(current.collation)) throw new Error("未找到数据库当前使用的字符集和排序规则");
            original.current = {charset: selected.name, collation: current.collation};
            setOptions(options);
            form.setFieldsValue({...original.current, ...(!editing ? {name: ""} : {})});
        }).catch((reason) => {
            if (request !== sequence.current) return;
            sequence.current++; onCancel();
            message.error(reason.message || "获取数据库字符集和排序规则失败");
        })
            .finally(() => { if (request === sequence.current) setLoading(false); });
        return () => { sequence.current++; };
    }, [open, instance?.id, name]);
    const submit = async (values: any) => {
        if (!open || pending.current || loading || !instance) return;
        const request = sequence.current;
        pending.current = true; setSaving(true);
        try {
            const body = {name: editing ? name : values.name.trim(), charset: values.charset, collation: values.collation};
            if (editing && body.charset === original.current.charset && body.collation === original.current.collation) { message.info("没有修改内容"); return; }
            if (editing) await api.update_database(instance, body);
            else await api.create_database(instance, body);
            if (request !== sequence.current) return;
            message.success(editing ? "修改成功" : "创建成功");
            onSuccess?.({name: body.name, action: editing ? "update" : "create"}); onCancel();
        } catch (reason: any) { if (request === sequence.current) message.error(reason.message || "保存数据库失败"); }
        finally { if (request === sequence.current) { pending.current = false; setSaving(false); } }
    };
    return <ProModal title={<Title>{editing ? "编辑数据库" : "创建数据库"}</Title>}
        onCancel={() => { if (!pending.current) onCancel(); }} onOk={form.submit}
        modalProps={{
            open, width: editing ? (theme?.isCompactTheme?.() ? 370 : 400) : (theme?.isCompactTheme?.() ? 420 : 450), zIndex: 1100, rootClassName: styles.modal,
            style: {maxWidth: "calc(100vw - 24px)"},
            okText: editing ? "保存" : "创建", cancelText: "取消", confirmLoading: saving,
            okButtonProps: {disabled: loading}, cancelButtonProps: {disabled: saving},
            // 空 Form 保持实例连接，输入项随弹窗打开挂载，退出动画后释放。
            closable: !saving, keyboard: !saving, mask: {closable: !saving}, forceRender: true,
            afterClose: () => { setVisible(false); setOptions(undefined); },
            focusable: {focusTriggerAfterClose: false},
        } as any}>
        <Form form={form} layout="vertical" variant="filled" className={`${styles.form} ${formStyles.form}`} onFinish={submit} disabled={saving}>
            {visible && (loading ? <div className={styles.loading}><Spin/></div> : <>
                {!editing && <Form.Item name="name" label="数据库名称" rules={[{required: true, whitespace: true, message: "请输入数据库名称"}, {max: 64, message: "数据库名称最多允许 64 个字符"}]}>
                    <Input maxLength={64} autoComplete="off" placeholder="请输入数据库名称"/>
                </Form.Item>}
                <Form.Item name="charset" label="字符集" rules={[{required: true, message: "请选择字符集"}]}>
                    <Select showSearch optionFilterProp="label" placeholder="请选择字符集"
                        options={characterSets.map((item: any) => ({label: item.name, value: item.name}))}
                        onChange={(value) => form.setFieldValue("collation", characterSets.find((item: any) => item.name === value)?.default_collation)}/>
                </Form.Item>
                <Form.Item name="collation" label="排序规则" rules={[{required: true, message: "请选择排序规则"}]}>
                    <Select showSearch optionFilterProp="label" placeholder="请选择排序规则" disabled={saving || !charset}
                        options={collations.map((value: string) => ({label: value, value}))}/>
                </Form.Item>
            </>)}
        </Form>
    </ProModal>;
};
