import {useEffect, useRef, useState} from "react";
import {App, AutoComplete, Button, Col, ConfigProvider, Form, Input, InputNumber, Row, Segmented, Select, Spin, Typography, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {useDatabaseClient} from "../shared/context";
import useStyles from "./styles";

const fieldTypes = [{value: "string", label: "文本"}, {value: "number", label: "数字"}, {value: "boolean", label: "布尔"},
    {value: "json", label: "对象 / 数组"}, {value: "null", label: "空值"}];

// 默认逐字段编辑，嵌套内容仍可使用 JSON，不要求用户拆分完整文档。
export default ({record, editor, onCancel, onSuccess}: any) => {
    const {message} = App.useApp();
    const {operate, features, Editor: AceEditor} = useDatabaseClient();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const {token} = antdTheme.useToken();
    const [form] = Form.useForm();
    const fields = Form.useWatch("fields", form) || [];
    const documentType = Form.useWatch("document_type", form);
    const [mode, setMode] = useState("fields");
    const [saving, setSaving] = useState(false);
    const [documentTypes, setDocumentTypes] = useState<string[]>([]);
    const [mapping, setMapping] = useState<any>(null);
    const routingRule = (features.typed_documents && documentTypes.length
        ? mapping?.[documentType || documentTypes[0]]?._routing?.required : mapping?._routing?.required) ?? mapping?._default_?._routing?.required;
    const routingRequired = routingRule === true || routingRule === "true";
    const live = useRef(true);
    const pending = useRef(false);
    const sequence = useRef(0);
    const lastEditor = useRef(editor);
    const displayEditor = editor || lastEditor.current;
    const readonly = displayEditor?.mode === "view";
    const [rawSource, setRawSource] = useState("{}");
    const [fallback, setFallback] = useState(false);
    const [reading, setReading] = useState(false);
    const snapshot = useRef<Record<string, string>>({});
    const drafts = useRef({fields: "", document: ""});
    const hasUnsafeNumber = (value: any): boolean => typeof value === "number"
        ? !Number.isFinite(value) || Number.isInteger(value) && !Number.isSafeInteger(value)
        : value !== null && typeof value === "object" && Object.values(value).some(hasUnsafeNumber);
    const document = (values: any, currentMode = mode): string => {
        if (currentMode === "json") {
            const text = values.document?.trim() || "{}";
            const value = JSON.parse(text);
            if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("文档内容必须是 JSON 对象");
            return text;
        }
        const names = new Set<string>();
        const fields: string[] = [];
        for (const field of values.fields || []) {
            const name = String(field.name ?? "");
            if (!name.trim()) throw new Error("请输入字段名称");
            if (names.has(name)) throw new Error(`字段“${name}”重复`);
            names.add(name);
            let value: string;
            if (field.type === "json") {
                value = String(field.value || "").trim();
                JSON.parse(value);
            } else if (field.type === "number") {
                value = String(field.value ?? "");
                let number;
                try { number = JSON.parse(value); } catch { throw new Error(`字段“${name}”需要有效数字`); }
                if (typeof number !== "number" || !Number.isFinite(number)) throw new Error(`字段“${name}”需要有效数字`);
            } else value = field.type === "null" ? "null" : field.type === "boolean" ? String(Boolean(field.value)) : JSON.stringify(String(field.value ?? ""));
            // 字段名转义，数字和嵌套 JSON 保留原文，避免经过 Number 后丢失精度。
            fields.push(`${JSON.stringify(name)}: ${value}`);
        }
        return `{\n  ${fields.join(",\n  ")}\n}`;
    };
    const populate = (raw: string) => {
        const source = JSON.parse(raw);
        if (!source || typeof source !== "object" || Array.isArray(source)) throw new Error("文档内容必须是 JSON 对象");
        const values: any[] = [];
        let depth = 0, name = "", start = -1;
        // 仅切分顶层字段，数字和嵌套 JSON 使用原文，切换模式不会丢失大整数精度。
        for (const match of raw.matchAll(/"(?:\\[\s\S]|[^"\\])*"|[{}\[\],:]/g)) {
            const token = match[0], index = match.index!;
            if (depth === 1 && start >= 0 && (token === "," || token === "}")) {
                const text = raw.slice(start, index).trim(), value = JSON.parse(text);
                const type = value === null ? "null" : typeof value === "object" ? "json" : typeof value;
                values.push({name, type, value: type === "number" || type === "json" ? text : value});
                start = -1;
            }
            if (token === "{" || token === "[") depth++;
            else if (token === "}" || token === "]") depth--;
            else if (depth === 1 && start < 0) {
                if (token.startsWith('"')) name = JSON.parse(token);
                else if (token === ":") start = index + 1;
            }
        }
        return values;
    };
    useEffect(() => { live.current = true; return () => { live.current = false; sequence.current++; }; }, []);
    useEffect(() => {
        const request = ++sequence.current;
        if (!editor) return;
        lastEditor.current = editor;
        pending.current = false; setSaving(false); setMode("fields"); setFallback(false); setDocumentTypes([]); setMapping(null); setReading(true);
        snapshot.current = {};
        const populateForm = (source: any, raw = JSON.stringify(source, null, 2)) => {
            setRawSource(raw); setMode("fields");
            if (editor.mode === "view") return;
            form.resetFields();
            const values = populate(raw);
            const fields = values.length ? values : [{name: "", type: "string", value: ""}];
            form.setFieldsValue({document_id: editor.item?._id || "", document_type: editor.item?._type || "", routing: snapshot.current.routing || "",
                fields, document: raw});
            drafts.current = {fields: JSON.stringify(fields), document: raw};
        };
        if (editor.item) {
            // 列表允许投影字段，编辑前按 ID 读取完整文档，避免覆盖未显示的数据。
            const routing = String(editor.item._routing ?? editor.item.fields?._routing?.[0] ?? "");
            const query = new URLSearchParams();
            if (routing) query.set("routing", routing);
            const path = `/${encodeURIComponent(record.name)}/${encodeURIComponent(features.typed_documents ? editor.item._type || "_doc" : "_doc")}/${encodeURIComponent(editor.item._id)}${routing ? `?${query}` : ""}`;
            void operate(record, "request", {method: "GET", path}).then((response) => {
                if (!live.current || request !== sequence.current) return;
                if (response?.data?.found === false || !response?.data?._source) throw new Error("文档不存在或未启用原始内容存储");
                if (hasUnsafeNumber(response?.data?._source) && typeof response?.data?._source_json !== "string") throw new Error("文档包含大整数，需要更新服务端后再编辑");
                const current = response?.data;
                snapshot.current = {
                    routing: String(current?._routing ?? current?.fields?._routing?.[0] ?? routing),
                    ...(current?._seq_no !== undefined && current?._primary_term !== undefined ? {seq_no: String(current?._seq_no), primary_term: String(current?._primary_term)} : {}),
                    ...(current?._version !== undefined ? {version: String(current?._version)} : {}),
                };
                populateForm(current?._source, current?._source_json);
            }).catch((reason) => { if (live.current && request === sequence.current) { message.error(reason?.message || "读取文档失败"); onCancel(); } })
                .finally(() => { if (live.current && request === sequence.current) setReading(false); });
        } else {
            populateForm({});
            void operate(record, "mapping").then((data) => {
                if (!live.current || request !== sequence.current) return;
                const mapping = data?.[record.name]?.mappings || data?.mappings || data;
                if (!mapping || typeof mapping !== "object" || Array.isArray(mapping)) throw new Error("无法读取索引映射，请重试");
                const directMapping = ["properties", "dynamic", "_source", "_meta", "_routing", "_all", "dynamic_templates", "date_detection", "numeric_detection", "dynamic_date_formats"].some((key) => Object.prototype.hasOwnProperty.call(mapping, key));
                const types = features.typed_documents && !directMapping ? Object.keys(mapping).filter((name) => name !== "_default_" && mapping[name] && typeof mapping[name] === "object") : [];
                setMapping(mapping);
                setDocumentTypes(types);
                if (types.length === 1 && !form.getFieldValue("document_type")) form.setFieldValue("document_type", types[0]);
            }).catch((reason) => { if (live.current && request === sequence.current) { message.error(reason?.message || "读取索引映射失败"); onCancel(); } })
                .finally(() => { if (live.current && request === sequence.current) setReading(false); });
        }
    }, [editor, form]);
    const switchMode = (value: string) => {
        if (value === mode || pending.current || reading) return;
        const values = form.getFieldsValue(true);
        try {
            if (value === "json" && JSON.stringify(values.fields) !== drafts.current.fields) {
                // 空白占位行留在字段草稿里，不作为文档字段；未完成的输入不阻止切换。
                const fields = (values.fields || []).filter((field: any) => field.name || field.type !== "string" || field.value);
                form.setFieldValue("document", document({fields}, "fields"));
            } else if (value === "fields" && values.document !== drafts.current.document) {
                form.setFieldValue("fields", populate(values.document?.trim() || "{}"));
            }
        } catch {
            message.info(value === "json" ? "未填完的字段已保留，当前显示 JSON 草稿" : "JSON 尚未完成，当前显示字段草稿");
        }
        drafts.current = {fields: JSON.stringify(form.getFieldValue("fields")), document: form.getFieldValue("document")};
        setMode(value);
    };
    const save = async (values: any) => {
        if (pending.current || reading || !editor || readonly || !editor.item && mapping === null) return;
        let content;
        try { content = document(values); }
        catch (reason: any) { message.error(reason?.message || "文档内容不合法"); return; }
        const request = sequence.current;
        pending.current = true; setSaving(true);
        try {
            await operate(record, editor.item ? "update" : "insert", {document: content, document_id: editor.item ? editor.item._id : values.document_id || "",
                routing: editor.item ? snapshot.current.routing || "" : values.routing || "",
                ...(editor.item ? snapshot.current : {}),
                ...(features.typed_documents ? {document_type: values.document_type?.trim()} : {})});
            if (!live.current || request !== sequence.current) return;
            message.success("保存成功"); onSuccess?.(); onCancel();
        } catch (reason: any) { if (live.current && request === sequence.current) message.error(reason?.message || "保存失败"); }
        finally { if (live.current && request === sequence.current) { pending.current = false; setSaving(false); } }
    };
    return <ProModal title={<Title>{readonly ? "查看文档" : displayEditor?.item ? "编辑文档" : "添加文档"}</Title>}
        onCancel={() => { if (!pending.current) onCancel(); }} onOk={() => readonly ? onCancel() : form.submit()} modalProps={{
            open: Boolean(editor), width: readonly ? (theme?.isCompactTheme?.() ? 480 : 520) : (theme?.isCompactTheme?.() ? 480 : 540), rootClassName: styles.modal,
            styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}},
            style: {maxWidth: "calc(100vw - 24px)"}, zIndex: 1140, okText: readonly ? "确认" : "保存", cancelText: "取消", okButtonProps: {disabled: reading},
            confirmLoading: saving, closable: !saving, keyboard: !saving, mask: {closable: !saving}, destroyOnHidden: true, forceRender: true,
            cancelButtonProps: {disabled: saving, style: readonly ? {display: "none"} : undefined},
        }}>
        <Spin spinning={reading}>{readonly ? <><div className="es-copy"><Typography.Text copyable={{text: rawSource}}>复制</Typography.Text></div>
            <pre className="es-code">{rawSource}</pre></>
            : <Form form={form} layout="vertical" variant="filled" onFinish={save} disabled={saving || reading}>
                <ConfigProvider theme={{components: {Input: {controlHeight: token.controlHeight}, Select: {controlHeight: token.controlHeight}}}}>
                {!displayEditor?.item && <Row gutter={12}>
                    <Col xs={24} sm={12}><Form.Item name="document_id" label="文档 ID">
                        <Input placeholder="留空自动生成" autoComplete="off"/>
                    </Form.Item></Col>
                    <Col xs={24} sm={12}><Form.Item name="routing" label="路由值" rules={routingRequired ? [{required: true, whitespace: true, message: "该索引要求填写路由值"}] : undefined}>
                        <Input placeholder={routingRequired ? "请输入路由值" : "留空使用默认路由"} autoComplete="off"/>
                    </Form.Item></Col>
                </Row>}
                {features.typed_documents && <Form.Item name="document_type" label="文档类型" rules={documentTypes.length > 1 ? [{required: true, whitespace: true, message: "请选择文档类型"}] : undefined}>
                    <AutoComplete disabled={Boolean(displayEditor?.item)} options={documentTypes.map((value) => ({value}))} placeholder="选择或填写，留空自动确定"/>
                </Form.Item>}
                </ConfigProvider>
                <div className="es-document-mode"><span>文档内容</span><Segmented value={mode} options={[{value: "fields", label: "字段编辑"}, {value: "json", label: "JSON 编辑"}]} disabled={saving || reading} onChange={switchMode}/></div>
                {mode === "json" ? <div className="es-document-json"><Form.Item name="document" rules={[{required: true, message: "请输入文档内容"}, {validator: async (_, value) => {
                    try { const data = JSON.parse(value || "{}"); if (!data || typeof data !== "object" || Array.isArray(data)) throw new Error(); }
                    catch { throw new Error("请输入有效的 JSON 对象"); }
                }}]}>{AceEditor && !fallback ? <AceEditor className="es-document-ace" mode="json" height={compact ? 220 : 240} fontSize={compact ? 12 : 13} tabSize={2}
                    theme={theme?.isDarkTheme?.() ? "tomorrow_night" : "textmate"} readOnly={saving || reading}
                    showPrintMargin={false} wrapEnabled enableSnippets={false} onError={() => setFallback(true)}
                    placeholder={'请输入 JSON 文档，例如 {"title":"示例"}'}/>
                    : <Input.TextArea className="es-json-editor es-document-json-fallback" spellCheck={false} placeholder={'请输入 JSON 文档，例如 {"title":"示例"}'}/>}
                </Form.Item></div>
                    : <Form.List name="fields">{(items, {add, remove}) => <>
                        <div className="es-document-fields">{items.map((field) => {
                            const type = fields[field.name]?.type || "string";
                            return <div className="es-field" key={field.key}>
                                <div className="es-field-heading"><span>字段 {field.name + 1}</span><Button type="text" size="small" disabled={saving} icon={<Icon type="CloseOutlined"/>} aria-label="删除字段" onClick={() => remove(field.name)}/></div>
                                <div className="es-field-inputs">
                                <Form.Item name={[field.name, "name"]} label="字段名称" rules={[{required: true, whitespace: true, message: "请输入字段名称"}]}><Input placeholder="例如 title"/></Form.Item>
                                <Form.Item name={[field.name, "type"]} label="数据类型"><Select placeholder="请选择类型" options={fieldTypes} onChange={(type) => form.setFieldValue(["fields", field.name, "value"], type === "boolean" ? false : type === "number" ? undefined : type === "json" ? "{}" : "")}/></Form.Item>
                                <Form.Item className="es-field-value" label="字段值" name={type === "null" ? undefined : [field.name, "value"]} rules={type === "json" ? [{validator: async (_, value) => { try { JSON.parse(value); } catch { throw new Error("请输入有效的 JSON"); } }}] : type === "number" ? [{required: true, message: "请输入数字"}] : undefined}>
                                    {type === "null" ? <Input value="NULL" disabled/>
                                        : type === "boolean" ? <Select placeholder="请选择布尔值" options={[{value: true, label: "true"}, {value: false, label: "false"}]}/>
                                        : type === "number" ? <InputNumber stringMode placeholder="请输入数字"/> : type === "json" ? <Input.TextArea className="es-json-editor" rows={3} spellCheck={false} placeholder="请输入 JSON 对象或数组"/> : <Input.TextArea autoSize={{minRows: 1, maxRows: 4}} placeholder="字段值"/>}
                                </Form.Item>
                                </div>
                            </div>;
                        })}</div>
                        <ConfigProvider wave={{disabled: true}}><Button block type="dashed" icon={<Icon type="PlusOutlined"/>} disabled={saving} onClick={() => add({name: "", type: "string", value: ""})}>添加字段</Button></ConfigProvider>
                    </>}</Form.List>}
            </Form>}</Spin>
    </ProModal>;
};
