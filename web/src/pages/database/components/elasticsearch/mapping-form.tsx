import {useLayoutEffect, useRef} from "react";
import {App, Button, ConfigProvider, Form, Input, Select, Spin, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import useStyles from "./styles";
import useFormStyles from "./mapping-form-styles";

export default ({open, editor, types, legacyStrings, multiFields, canUpdateSearch, analyzers, loading, saving, onCancel, onSubmit}: any) => {
    const {message} = App.useApp();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());
    const {token} = antdTheme.useToken();
    const {styles} = useStyles({compact});
    const {styles: formStyles} = useFormStyles({compact});
    const [form] = Form.useForm();
    const retained = useRef({editor, loading});
    if (open && editor) retained.current = {editor, loading};
    const current = retained.current.editor;
    const reading = open ? loading : retained.current.loading;
    const mode = current?.mode || "create";
    const definition = current?.item?.definition || {};
    const editing = mode === "edit";
    const subfieldsOnly = mode === "subfield";
    const type = Form.useWatch("type", {form, preserve: true});
    const analyzed = ["text", "string"].includes(type) && !(editing && type === "string" && (definition.index === "not_analyzed" || definition.index === false));
    const allowSubfields = multiFields && !current?.item?.isSubfield && (subfieldsOnly || ["text", "keyword", "string"].includes(type));
    const subfieldTypes = legacyStrings ? ["string"] : ["text", "keyword"];

    useLayoutEffect(() => {
        if (!open || !editor) return;
        const field = editor.item?.definition || {};
        form.resetFields();
        form.setFieldsValue({
            name: editor.mode === "create" ? "" : editor.item?.name,
            type: editor.mode === "create" ? (legacyStrings ? "string" : "text") : field.type || (field.properties ? "object" : undefined),
            analyzer: field.analyzer,
            search_analyzer: field.search_analyzer,
            format: field.format,
            fields: editor.mode === "subfield" ? [{name: "", type: legacyStrings ? "string" : "text"}] : [],
        });
    }, [open, editor, legacyStrings, form]);

    const submit = async (values: any) => {
        if (!open || loading || saving) return;
        try { await onSubmit(values); }
        catch (reason: any) { message.error(reason?.message || "保存失败"); }
    };

    return <ProModal title={<Title>{subfieldsOnly ? "添加子字段" : editing ? "编辑字段" : "添加字段"}</Title>}
        onOk={form.submit} onCancel={() => { if (!saving) onCancel(); }} modalProps={{
            open, width: compact ? 480 : 520, rootClassName: `${styles.modal} ${formStyles.modal}`,
            styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}},
            style: {maxWidth: "calc(100vw - 24px)"}, zIndex: 1140, okText: "保存", cancelText: "取消",
            confirmLoading: saving, okButtonProps: {disabled: loading}, cancelButtonProps: {disabled: saving},
            closable: !saving, keyboard: !saving, mask: {closable: !saving}, forceRender: true,
            focusable: {focusTriggerAfterClose: false},
        }}>
        <ConfigProvider theme={{components: {Input: {controlHeight: token.controlHeight}, Select: {controlHeight: token.controlHeight}}}}>
            <Form form={form} layout="vertical" variant="filled" className={formStyles.form} onFinish={submit} disabled={saving || reading}>
                {reading ? <div className="es-mapping-loading"><Spin/></div> : current && <>
                    {subfieldsOnly ? <Form.Item label="所属字段"><Input readOnly value={current.item?.name} placeholder="所属字段"/></Form.Item> : <>
                        <Form.Item name="name" label="字段名称" rules={editing ? undefined : [
                            {required: true, whitespace: true, message: "请输入字段名称"},
                            {pattern: /^[^\s.]+$/, message: "字段名称不能包含空格或点"},
                            {validator: async (_, value) => {
                                if (value && Object.prototype.hasOwnProperty.call(current.properties || {}, value.trim())) throw new Error("字段名称已存在");
                            }},
                        ]}><Input readOnly={editing} autoComplete="off" placeholder="例如 title"/></Form.Item>
                        <Form.Item name="type" label="字段类型" rules={[{required: true, message: "请选择字段类型"}]}>
                            {editing ? <Input readOnly placeholder="字段类型"/> : <Select placeholder="请选择字段类型" options={(types || []).map((value: string) => ({value, label: value}))}/>}
                        </Form.Item>
                        {type === "date" && <Form.Item name="format" label="日期格式"><Input readOnly={editing} placeholder="留空使用默认格式，例如 yyyy-MM-dd"/></Form.Item>}
                        {analyzed && <>
                            <Form.Item name="analyzer" label="写入分析器">
                                {editing ? <Input readOnly placeholder="默认分析器"/> : <Select allowClear showSearch optionFilterProp="label" options={analyzers} placeholder="默认分析器"/>}
                            </Form.Item>
                            {(!editing || canUpdateSearch) && <Form.Item name="search_analyzer" label="搜索分析器">
                                <Select allowClear={!editing} showSearch optionFilterProp="label" options={analyzers} placeholder="默认使用写入分析器"/>
                            </Form.Item>}
                            {editing && <div className="es-mapping-note">已存在字段的写入分析器不能直接更改。</div>}
                        </>}
                    </>}
                    {allowSubfields && <Form.Item label="子字段" required={subfieldsOnly} className="es-mapping-fields">
                        {mode !== "create" && <div className="es-mapping-note">新增子字段需更新历史文档。</div>}
                        <Form.List name="fields" rules={[{validator: async (_, values) => {
                            if (subfieldsOnly && !values?.length) throw new Error("请至少添加一个子字段");
                        }}]}>{(fields, {add, remove}, {errors}) => <>
                            <div className="es-mapping-field-list">{fields.map((field, index) => <div className="es-mapping-field" key={field.key}>
                                <div className="es-mapping-field-heading"><span>子字段 {index + 1}</span>
                                    <Button color="default" variant="filled" size="small" danger icon={<Icon type="CloseOutlined"/>}
                                        aria-label={`删除子字段 ${index + 1}`} disabled={saving || subfieldsOnly && fields.length === 1} onClick={() => remove(field.name)}/>
                                </div>
                                <div className="es-mapping-grid">
                                    <Form.Item name={[field.name, "name"]} label="字段名称" rules={[
                                        {required: true, whitespace: true, message: "请输入子字段名称"},
                                        {pattern: /^[^\s.]+$/, message: "名称不能包含空格或点"},
                                        {validator: async (_, value) => {
                                            const name = value?.trim();
                                            if (!name) return;
                                            if (Object.prototype.hasOwnProperty.call(definition.fields || {}, name)) throw new Error("子字段名称已存在");
                                            if ((form.getFieldValue("fields") || []).some((item: any, index: number) => index !== field.name && item?.name?.trim() === name)) throw new Error("子字段名称不能重复");
                                        }},
                                    ]}><Input placeholder="例如 pinyin、zh" autoComplete="off"/></Form.Item>
                                    <Form.Item name={[field.name, "type"]} label="字段类型" rules={[{required: true, message: "请选择字段类型"}]}>
                                        <Select options={subfieldTypes.map((value) => ({value, label: value}))} placeholder="请选择字段类型"/>
                                    </Form.Item>
                                    <Form.Item noStyle shouldUpdate={(before, after) => before.fields?.[field.name]?.type !== after.fields?.[field.name]?.type}>
                                        {({getFieldValue}) => ["text", "string"].includes(getFieldValue(["fields", field.name, "type"])) && <>
                                            <Form.Item name={[field.name, "analyzer"]} label="写入分析器"><Select allowClear showSearch optionFilterProp="label" options={analyzers} placeholder="默认分析器"/></Form.Item>
                                            <Form.Item name={[field.name, "search_analyzer"]} label="搜索分析器"><Select allowClear showSearch optionFilterProp="label" options={analyzers} placeholder="默认使用写入分析器"/></Form.Item>
                                        </>}
                                    </Form.Item>
                                </div>
                            </div>)}</div>
                            <Button block type="dashed" icon={<Icon type="PlusOutlined"/>} onClick={() => add({name: "", type: legacyStrings ? "string" : "text"})}>新增子字段</Button>
                            <Form.ErrorList errors={errors}/>
                        </>}</Form.List>
                    </Form.Item>}
                </>}
            </Form>
        </ConfigProvider>
    </ProModal>;
};
