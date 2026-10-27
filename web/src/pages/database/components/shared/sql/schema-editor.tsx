import {useLayoutEffect, useRef, useState} from "react";
import {App, AutoComplete, Button, Collapse, ConfigProvider, Form, Input, Select, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {useDatabaseClient} from "../context";
import useStyles from "../styles/sql-schema";
import SqlForeignKeyFields, {validateForeignKeys} from "./foreign-key-fields";

const types: Record<number, string[]> = {
    3: ["BIGINT", "INTEGER", "SMALLINT", "NUMERIC(10,2)", "DOUBLE PRECISION", "VARCHAR(255)", "TEXT", "BOOLEAN", "DATE", "TIMESTAMP", "TIMESTAMPTZ", "UUID", "JSON", "JSONB", "BYTEA"],
    5: ["BIGINT", "INT", "SMALLINT", "TINYINT", "DECIMAL(10,2)", "FLOAT", "NVARCHAR(255)", "NVARCHAR(MAX)", "VARCHAR(255)", "BIT", "DATE", "DATETIME", "DATETIME2", "UNIQUEIDENTIFIER", "VARBINARY(MAX)"],
    6: ["INTEGER", "TEXT", "REAL", "NUMERIC", "BLOB", "VARCHAR(255)", "BOOLEAN", "DATETIME"],
};

// 弹窗只负责可视化设计与参数编码；SQL 方言和能力校验由各数据库操作类处理。
export default ({record, action, schema, table, columns = [], indexes = [], tables = [], schemas = [], isMaterializedView = false, onCancel, onSuccess}: any) => {
    const {operate, features = {}, ActionDropdown} = useDatabaseClient();
    const {message} = App.useApp();
    const {styles} = useStyles();
    const {token} = antdTheme.useToken();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const fieldFontSize = compact ? 11 : 12;
    const fieldControlHeight = compact ? 24 : 28;
    const fieldTheme = {
        token: {fontSize: fieldFontSize, fontSizeSM: fieldFontSize, controlHeight: fieldControlHeight, controlHeightSM: fieldControlHeight},
        components: {
            Input: {inputFontSize: fieldFontSize, inputFontSizeSM: fieldFontSize, controlHeight: fieldControlHeight, controlHeightSM: fieldControlHeight},
            Select: {optionFontSize: fieldFontSize, optionHeight: fieldControlHeight, controlHeight: fieldControlHeight, controlHeightSM: fieldControlHeight},
            Form: {labelFontSize: fieldFontSize, controlHeight: fieldControlHeight, controlHeightSM: fieldControlHeight},
            Button: {contentFontSize: fieldFontSize, contentFontSizeSM: fieldFontSize, controlHeight: fieldControlHeight, controlHeightSM: fieldControlHeight},
        },
    };
    const [form] = Form.useForm();
    const fields = Form.useWatch("columns", form) || [];
    const tableName = Form.useWatch("table", form);
    const [saving, setSaving] = useState(false);
    const [advanced, setAdvanced] = useState<string[]>([]);
    const request = useRef(0);
    const pending = useRef(false);
    const lastAction = useRef(action);
    const activeAction = useRef(action);
    activeAction.current = action;
    const type = Number(record.type);
    const typeFeatures: Record<string, string> = {JSON: "json_type", JSONB: "jsonb_type", DATE: "date_type", DATETIME2: "datetime2_type", UUID: "uuid_type"};
    // 关闭动画期间保留操作与表单，避免内容提前清空、弹窗缩小。
    const mode = (action || lastAction.current)?.mode;
    const current = (action || lastAction.current)?.column;
    const creatingTable = mode === "create_table";
    const columnMode = creatingTable || mode === "add_column" || mode === "update_column";
    const renameOnly = mode === "update_column" && type === 5 && Boolean(current?.generated || current?.auto_increment || current?.computed);
    const title = ({create_table: "新建数据表", rename_table: "重命名", add_column: "添加字段", update_column: "修改字段", create_index: "创建索引"} as any)[mode];
    const typeOptions = (types[type] || []).filter((value) => features[typeFeatures[value]] !== false).map((value) => {
        const description = /char|text/i.test(value) ? "文本" : /int|numeric|decimal|float|double|real/i.test(value) ? "数字"
            : /date|time/i.test(value) ? "日期时间" : /bool|bit/i.test(value) ? "是或否" : /blob|binary|bytea/i.test(value) ? "二进制内容" : "其他类型";
        return {value, search: `${value} ${description}`, label: creatingTable
            ? <span className={styles.typeOption}><span>{value}</span><small>{description}</small></span>
            : `${value} · ${description}`};
    });

    useLayoutEffect(() => {
        ++request.current;
        if (!action?.mode) return;
        lastAction.current = action;
        pending.current = false;
        setSaving(false); setAdvanced([]);
        form.resetFields();
        if (columnMode) {
            const value = current ? {
                name: current.name,
                type: current.type,
                nullable: Boolean(current.nullable),
                primary_key: Boolean(current.primary_key),
                auto_increment: Boolean(current.auto_increment),
                default_kind: current.default == null ? "none" : current.default_expression ? "expression" : "value",
                default: current.default == null ? "" : String(current.default),
            } : {name: creatingTable ? "id" : "", type: type === 6 ? "INTEGER" : "BIGINT", nullable: !creatingTable,
                primary_key: creatingTable, auto_increment: creatingTable, default_kind: "none", default: ""};
            form.setFieldsValue({columns: [value], table: "", indexes: [], foreign_keys: []});
        } else if (mode === "rename_table") {
            form.setFieldsValue({new_name: table});
        } else if (mode === "create_index") {
            form.setFieldsValue({name: "", columns: [{name: columns[0]?.name, direction: "asc"}], index_type: "normal"});
        }
        return () => { request.current += 1; };
    }, [action, record.id, schema, table]);

    const save = async (values: any) => {
        if (pending.current || !action?.mode) return;
        let body: any = {schema, table};
        try {
            if (columnMode) {
                const names = new Set<string>();
                const definitions = (values.columns || []).map((field: any) => {
                    const name = String(field.name || "").trim();
                    if (!name || names.has(name)) throw new Error("字段名称不能为空或重复");
                    names.add(name);
                    if (mode === "add_column" && columns.some((column: any) => column.name === name)) throw new Error("此字段已存在");
                    if (mode === "update_column" && name !== current.name && columns.some((column: any) => column.name === name)) throw new Error("此字段已存在");
                    // SQL Server 生成字段只能改名，结构参数始终使用原始元数据。
                    if (renameOnly) return {name, type: current.type, nullable: Boolean(current.nullable),
                        default: current.default == null ? null : String(current.default), default_expression: Boolean(current.default_expression),
                        auto_increment: Boolean(current.auto_increment)};
                    const source = field;
                    return {name, type: String(source.type || "").trim(), nullable: Boolean(source.nullable) && !(creatingTable && field.primary_key) && !source.auto_increment,
                        default: source.default_kind === "none" ? null : String(source.default ?? ""),
                        default_expression: source.default_kind === "expression", auto_increment: Boolean(source.auto_increment)};
                });
                if (!definitions.length) throw new Error("请至少添加一个字段");
                if (type !== 6 && definitions.some((column: any) => features[typeFeatures[column.type.toUpperCase().replace(/\(.*/, "").trim()]] === false)) throw new Error("当前版本不支持所选字段类型");
                if (type === 3 && features.smallserial === false && definitions.some((column: any) => column.auto_increment && /^(SMALLINT|INT2|SMALLSERIAL|SERIAL2)$/i.test(column.type))) throw new Error("当前版本的自动编号请使用 INTEGER 或 BIGINT");
                if (definitions.some((column: any) => !column.type)) throw new Error("请填写字段类型");
                if (definitions.some((column: any) => column.default_expression && !column.default.trim())) throw new Error("请填写默认值表达式");
                if (type !== 3 && definitions.filter((column: any) => column.auto_increment).length > 1) throw new Error("仅支持一个自增字段");
                const keys = (values.columns || []).filter((field: any) => field.primary_key).map((field: any) => String(field.name).trim());
                if (creatingTable && type === 6 && definitions.some((column: any) => column.auto_increment && (column.type.toUpperCase() !== "INTEGER" || keys.length !== 1 || keys[0] !== column.name))) {
                    throw new Error("SQLite 自增字段必须是 INTEGER 类型的单字段主键");
                }
                if (creatingTable) {
                    const foreignKeys = values.foreign_keys || [];
                    if (foreignKeys.length && features.foreign_keys === false) throw new Error("当前数据库不支持外键约束");
                    validateForeignKeys(foreignKeys);
                    body = {...body, table: values.table.trim(), columns: JSON.stringify(definitions), primary_keys: JSON.stringify(keys), indexes: JSON.stringify((values.indexes || []).map((index: any) => ({name: index.name.trim(), unique: index.type === "unique",
                        columns: index.columns.map((column: any) => ({name: column.name, desc: column.direction === "desc"}))}))), foreign_keys: JSON.stringify(foreignKeys)};
                }
                else body = {...body, column: JSON.stringify(definitions[0]), ...(mode === "update_column" ? {column_name: current.name} : {})};
            } else if (mode === "create_index") {
                const fields = values.columns || [];
                const primary = values.index_type === "primary";
                if (!["normal", "unique", "primary"].includes(values.index_type)) throw new Error("请选择索引类型");
                if (!fields.length || new Set(fields.map((field: any) => field.name)).size !== fields.length) throw new Error("请选择不重复的索引字段");
                if (fields.some((field: any) => !columns.some((column: any) => column.name === field.name))) throw new Error("请选择有效的索引字段");
                if (primary && isMaterializedView) throw new Error("物化视图不支持主键索引");
                if (primary && (type === 6 || indexes.some((index: any) => index.primary))) throw new Error("当前对象不支持新增主键");
                body = {...body, name: values.name.trim(), columns: JSON.stringify(fields.map((field: any) => ({name: field.name, desc: field.direction === "desc"}))),
                    unique: values.index_type === "unique" || primary, primary};
            } else if (mode === "rename_table") {
                if (values.new_name.trim() === table) { message.info("名称没有变动"); return; }
                body = {...body, new_name: values.new_name.trim()};
            }
        } catch (reason: any) { message.error(reason?.message); return; }
        const id = request.current;
        pending.current = true; setSaving(true);
        try {
            await operate(record, mode, body);
            if (id !== request.current) return;
            message.success("操作成功");
            onSuccess?.({mode, table: mode === "rename_table" ? body.new_name : body.table});
        } catch (reason: any) { if (id === request.current) message.error(reason?.message); }
        finally { if (id === request.current) { pending.current = false; setSaving(false); } }
    };

    const fieldEditor = (field: any, index: number, {remove, move}: any, count: number) => {
        const value = fields[index] || {};
        const nameField = <Form.Item name={[field.name, "name"]} label="字段名称" rules={[{required: true, whitespace: true, message: "请输入字段名称"}]}>
            <Input placeholder="例如 username" autoComplete="off" aria-label={`字段 ${index + 1} 名称`}/>
        </Form.Item>;
        const typeField = <Form.Item name={[field.name, "type"]} label="数据类型" rules={[{required: true, whitespace: true, message: "请选择或输入类型"}]}>
            <AutoComplete placeholder="选择或输入类型" disabled={saving || renameOnly} options={typeOptions} popupMatchSelectWidth={creatingTable ? 260 : undefined}
                filterOption={(input, option) => String(option?.search || option?.value || "").toLowerCase().includes(input.toLowerCase())}>
                <Input disabled={saving || renameOnly} placeholder="选择或输入类型" aria-label={`字段 ${index + 1} 类型`}/>
            </AutoComplete>
        </Form.Item>;
        const options = creatingTable ? <div className="schema-field-options">
                <Form.Item name={[field.name, "primary_key"]} label="主键" getValueProps={(value) => ({value: value ? 1 : 0})} normalize={(value) => value === 1}>
                    <Select placeholder="请选择是否为主键" disabled={saving || value.auto_increment} options={[{value: 0, label: "否"}, {value: 1, label: "是"}]}
                        onChange={(value) => { if (value === 1) form.setFieldValue(["columns", field.name, "nullable"], false); }}/>
                </Form.Item>
                <Form.Item name={[field.name, "auto_increment"]} label="自增" getValueProps={(value) => ({value: value ? 1 : 0})} normalize={(value) => value === 1}>
                    <Select placeholder="请选择是否自增" disabled={saving || (type !== 3 && fields.some((item: any, position: number) => position !== index && item?.auto_increment))}
                        options={[{value: 0, label: "关闭"}, {value: 1, label: "开启"}]} onChange={(value) => {
                        if (value === 1) {
                            form.setFieldValue(["columns", field.name, "nullable"], false);
                            form.setFieldValue(["columns", field.name, "default_kind"], "none");
                            if (type === 6) form.setFieldValue(["columns", field.name, "primary_key"], true);
                        }
                    }}/>
                </Form.Item>
                <Form.Item name={[field.name, "nullable"]} label="允许空值" required getValueProps={(value) => ({value: value ? 1 : 0})} normalize={(value) => value === 1}>
                    <Select placeholder="请选择是否允许空值" disabled={saving || value.primary_key || value.auto_increment} options={[{value: 0, label: "不允许"}, {value: 1, label: "允许"}]}/>
                </Form.Item>
            </div> : <div className="schema-column-options">
                <Form.Item name={[field.name, "nullable"]} label="允许空值" required getValueProps={(value) => ({value: value ? 1 : 0})} normalize={(value) => value === 1}>
                    <Select placeholder="请选择是否允许空值" disabled={saving || renameOnly || value.primary_key || value.auto_increment} options={[{value: 1, label: "允许"}, {value: 0, label: "不允许"}]}/>
                </Form.Item>
                <Form.Item name={[field.name, "auto_increment"]} label="自增" getValueProps={(value) => ({value: value ? 1 : 0})} normalize={(value) => value === 1}>
                    <Select placeholder="请选择是否自增" disabled={saving || renameOnly || (mode === "update_column" && [5,6].includes(type)) || (type === 6 && !creatingTable)} options={[{value: 0, label: "关闭"}, {value: 1, label: "开启"}]} onChange={(value) => {
                        if (value === 1) {
                            form.setFieldValue(["columns", field.name, "nullable"], false);
                            form.setFieldValue(["columns", field.name, "default_kind"], "none");
                        }
                    }}/>
                </Form.Item>
            </div>;
        const defaultKind = <Form.Item name={[field.name, "default_kind"]} noStyle={!creatingTable}>
            <Select placeholder="请选择默认值方式" aria-label={`${value.name || "字段"} 默认值方式`} disabled={saving || renameOnly || value.auto_increment}
                options={[{value: "none", label: "无默认值"}, {value: "value", label: "固定值"}, {value: "expression", label: features.default_expression === false ? "时间或常量" : "表达式"}]}/>
        </Form.Item>;
        const defaultValue = <Form.Item name={[field.name, "default"]} dependencies={[["columns", field.name, "default_kind"]]} rules={[{validator: async (_, value) => {
            if (form.getFieldValue(["columns", field.name, "default_kind"]) === "expression" && !String(value || "").trim()) {
                throw new Error("请输入表达式");
            }
        }}]}><Input aria-label={`${value.name || "字段"} 默认值`} disabled={saving || renameOnly || value.default_kind === "none" || value.auto_increment}
            placeholder={value.default_kind === "expression" ? "例如 CURRENT_TIMESTAMP" : value.default_kind === "value" ? "留空表示空字符串" : "由数据库决定"}/></Form.Item>;
        const hasDefault = !value.auto_increment && (value.default_kind === "value" || value.default_kind === "expression");
        const defaultEditor = creatingTable ? <div className="schema-default" data-has-value={hasDefault || undefined}>{defaultKind}{hasDefault && defaultValue}</div>
            : <Form.Item label="默认值"><div className="schema-column-default">
                {defaultKind}{hasDefault && defaultValue}
            </div></Form.Item>;
        if (creatingTable) return <div key={field.key} className="schema-field-row">
            <div className="schema-field-heading"><span>字段 {index + 1}</span><div className="schema-field-tools">
                <Button type="text" disabled={saving || index === 0} aria-label={`上移字段 ${index + 1}`} title="上移"
                    icon={<Icon type="ArrowUpOutlined"/>} onClick={() => move(field.name, field.name - 1)}/>
                <Button type="text" disabled={saving || index === count - 1} aria-label={`下移字段 ${index + 1}`} title="下移"
                    icon={<Icon type="ArrowDownOutlined"/>} onClick={() => move(field.name, field.name + 1)}/>
                <Button type="text" danger disabled={saving || count < 2} aria-label={`删除字段 ${index + 1}`} title="删除"
                    icon={<Icon type="DeleteOutlined"/>} onClick={() => remove(field.name)}/>
            </div></div>
            <div className="schema-field-main">{nameField}{typeField}</div>
            {options}
            <div className="schema-field-details"><Form.Item label="默认值">{defaultEditor}</Form.Item></div>
        </div>;
        return <div key={field.key}>{nameField}{typeField}{options}{defaultEditor}</div>;
    };

    return <ProModal title={<Title>{title || "结构管理"}</Title>}
        onCancel={() => { if (!saving) { request.current++; activeAction.current = undefined; onCancel(); } }} onOk={form.submit} okText={creatingTable ? "创建" : "保存"}
        modalProps={{
            open: Boolean(action?.mode), className: !creatingTable && (columnMode || mode === "create_index") ? styles.formEditor : styles.editor,
            width: mode === "rename_table" ? compact ? 320 : 340 : creatingTable ? compact ? 560 : 600 : compact ? 440 : 480,
            styles: {wrapper: {paddingInline: 12}}, style: {maxWidth: "calc(100vw - 24px)"},
            cancelText: "取消", confirmLoading: saving, mask: {closable: !saving}, keyboard: !saving, closable: !saving, destroyOnHidden: true,
        }}>
        <ConfigProvider theme={{components: {
            Input: {controlHeight: token.controlHeight},
            Select: {controlHeight: token.controlHeight},
        }}}>
        <Form form={form} layout="vertical" variant="filled" onFinish={save} disabled={saving}
            onFinishFailed={({errorFields}) => {
                const sections = errorFields.flatMap((field) => field.name[0] === "indexes" ? ["indexes"] : field.name[0] === "foreign_keys" ? ["foreign-keys"] : []);
                if (sections.length) setAdvanced((current) => Array.from(new Set([...current, ...sections])));
                if (!creatingTable) return;
                const field = errorFields.find((field) => field.name[0] === "columns");
                const index = errorFields.find((field) => field.name[0] === "indexes");
                if (field) {
                    message.error({key: `sql-schema-field-error-${record.id}`, content: `第 ${Number(field.name[1]) + 1} 个字段：${field.errors[0]}`});
                    form.scrollToField(field.name, {block: "nearest", focus: true});
                } else if (index) {
                    message.error({key: `sql-schema-field-error-${record.id}`, content: `${index.name.length > 1 ? `索引 ${Number(index.name[1]) + 1}：` : ""}${index.errors[0]}`});
                }
            }}>
            {creatingTable && <Form.Item name="table" label="数据表名称" rules={[{required: true, whitespace: true, message: "请输入名称"}]}><Input autoComplete="off" placeholder="例如 users"/></Form.Item>}
            {columnMode && <>
                <Form.List name="columns">{(values, operations) => {
                    const editor = values.map((field, index) => fieldEditor(field, index, operations, values.length));
                    return creatingTable ? <Form.Item label="字段结构" required style={{marginBottom: 0}}>
                        <ConfigProvider componentSize="small" theme={fieldTheme}>
                            <div className="schema-field-editor" data-compact={compact || undefined}>
                                <div className="schema-field-list">{editor}</div>
                            </div>
                        </ConfigProvider>
                        <ConfigProvider wave={{disabled: true}}>
                            <Button type="dashed" block className="schema-field-add" icon={<Icon type="PlusOutlined"/>} disabled={saving || values.length >= 1024}
                                onClick={() => operations.add({name: "", type: type === 5 ? "NVARCHAR(255)" : type === 6 ? "TEXT" : "VARCHAR(255)", nullable: true, primary_key: false, auto_increment: false, default_kind: "none", default: ""})}>添加字段</Button>
                        </ConfigProvider>
                    </Form.Item> : editor;
                }}</Form.List>
                {creatingTable && <Collapse ghost className="schema-advanced" activeKey={advanced} expandIconPlacement="end"
                    expandIcon={({isActive}) => <Icon type="DownOutlined" rotate={isActive ? 180 : 0}/>}
                    onChange={(keys) => setAdvanced(Array.isArray(keys) ? keys.map(String) : [String(keys)])} items={[
                    {key: "indexes", forceRender: true, label: <span className="schema-advanced-label"><Icon type="OrderedListOutlined"/>索引配置</span>, children:
                        <ConfigProvider componentSize="small" theme={fieldTheme}><div className="schema-field-editor schema-create-indexes" data-compact={compact || undefined}>
                            <Form.List name="indexes">{(items, {add, remove}) => <>
                                {items.map((item, position) => <div className="schema-create-index" key={item.key}>
                                    <div className="schema-index-heading"><span>索引 {position + 1}</span>
                                        <Button type="text" danger aria-label={`删除索引 ${position + 1}`} icon={<Icon type="DeleteOutlined"/>} onClick={() => remove(item.name)}/></div>
                                    <div className="schema-index-settings">
                                        <Form.Item name={[item.name, "name"]} label="索引名称" rules={[{required: true, whitespace: true, message: "请输入索引名称"}, {validator: async (_, value) => {
                                            const name = String(value || "").trim().toLowerCase();
                                            if (name && (form.getFieldValue("indexes") || []).some((index: any, indexPosition: number) => indexPosition !== position && String(index?.name || "").trim().toLowerCase() === name)) throw new Error("索引名称不能重复");
                                        }}]}><Input autoComplete="off" placeholder="例如 idx_username"/></Form.Item>
                                        <Form.Item name={[item.name, "type"]} label="索引类型"><Select placeholder="请选择索引类型" options={[{value: "normal", label: "普通索引"}, {value: "unique", label: "唯一索引"}]}/></Form.Item>
                                    </div>
                                    <Form.Item label="索引字段" required className="schema-index-columns">
                                        <Form.List name={[item.name, "columns"]} rules={[{validator: async (_, columns) => {
                                            if (!columns?.length) throw new Error("请选择索引字段");
                                            const names = columns.map((column: any) => String(column?.name || "").toLowerCase()).filter(Boolean);
                                            if (new Set(names).size !== names.length) throw new Error("索引字段不能重复");
                                        }}]}>{(columns, {add: addColumn, remove: removeColumn}) => <>
                                            <div className="schema-index-list">{columns.map((column, columnPosition) => <div className="schema-index-row" key={column.key}>
                                                <Form.Item name={[column.name, "name"]} rules={[{required: true, message: "请选择索引字段"}, {validator: async (_, value) => {
                                                    if (value && !fields.some((field: any) => String(field?.name || "").trim() === value)) throw new Error("索引字段已修改或删除，请重新选择");
                                                }}]}><Select aria-label={`索引 ${position + 1} 字段 ${columnPosition + 1}`} placeholder="选择字段" showSearch={{optionFilterProp: "label"}}
                                                    options={fields.map((field: any) => String(field?.name || "").trim()).filter(Boolean).map((name: string) => ({value: name, label: name}))}/></Form.Item>
                                                <Form.Item name={[column.name, "direction"]}><Select placeholder="请选择排序方式" aria-label={`索引 ${position + 1} 字段 ${columnPosition + 1} 排序`} options={[{value: "asc", label: "升序"}, {value: "desc", label: "降序"}]}/></Form.Item>
                                                <Button type="text" danger disabled={saving || columns.length < 2} aria-label={`删除索引 ${position + 1} 字段 ${columnPosition + 1}`} icon={<Icon type="CloseOutlined"/>} onClick={() => removeColumn(column.name)}/>
                                            </div>)}</div>
                                            <Button className="schema-index-column-add" color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} disabled={saving || columns.length >= fields.length} onClick={() => addColumn({direction: "asc"})}>添加字段</Button>
                                        </>}</Form.List>
                                    </Form.Item>
                                </div>)}
                                <Button block color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} disabled={saving || items.length >= 64}
                                    onClick={() => add({name: "", type: "normal", columns: [{name: undefined, direction: "asc"}]})}>添加索引</Button>
                            </>}</Form.List>
                        </div></ConfigProvider>},
                    ...(features.foreign_keys !== false ? [{key: "foreign-keys", forceRender: true, label: <span className="schema-advanced-label"><Icon type="LinkOutlined"/>外键约束</span>, children:
                    <Form.Item name="foreign_keys" rules={[{validator: async (_, value) => { validateForeignKeys(value || []); }}]}>
                        <SqlForeignKeyFields record={record} table={String(tableName || "").trim()} columns={fields.map((field: any) => String(field?.name || "").trim()).filter(Boolean)}
                            tables={tables} schemas={schemas} schema={schema} disabled={saving} active={action || false}
                            onError={(reason: any) => {
                                if (!action || activeAction.current !== action) return;
                                activeAction.current = undefined;
                                request.current++;
                                onCancel();
                                message.error({key: `sql-schema-definition-${record.id}`, content: reason?.message || "获取引用字段失败"});
                            }}/>
                    </Form.Item>}] : [])]} />}
                {renameOnly && <div className="schema-note">此字段由数据库生成，仅支持修改名称。</div>}
            </>}
            {mode === "rename_table" && <Form.Item name="new_name" label="新名称" rules={[{required: true, whitespace: true, message: "请输入新名称"}]}><Input autoComplete="off" placeholder="请输入新名称"/></Form.Item>}
            {mode === "create_index" && <>
                <Form.Item name="name" label="索引名称" rules={[{required: true, whitespace: true, message: "请输入索引名称"}]}><Input autoComplete="off" placeholder="例如 idx_username"/></Form.Item>
                <Form.Item name="index_type" label="索引类型" rules={[{required: true, message: "请选择索引类型"}, {validator: async (_, value) => {
                    if (value !== "primary") return;
                    if (isMaterializedView) throw new Error("物化视图不支持主键索引");
                    if (type === 6 || indexes.some((index: any) => index.primary)) throw new Error("当前对象不支持新增主键");
                }}]}>
                    <Select placeholder="请选择索引类型" options={[
                        {value: "normal", label: "普通索引"}, {value: "unique", label: "唯一索引"},
                        ...(type !== 6 && !isMaterializedView && !indexes.some((index: any) => index.primary) ? [{value: "primary", label: "主键索引"}] : []),
                    ]}/>
                </Form.Item>
                <Form.Item label="索引字段" required>
                    <ConfigProvider theme={{components: {Select: {controlHeight: fieldControlHeight}}}}>
                    <Form.List name="columns">{(values, {add, remove, move}) => <>
                        <div className="schema-index-fields">{values.map((field, index) => <div className="schema-index-field" key={field.key}>
                            <Form.Item name={[field.name, "name"]} rules={[{required: true, message: "请选择字段"}]}>
                                <Select placeholder="选择字段" aria-label={`索引字段 ${index + 1}`} showSearch optionFilterProp="label"
                                    options={columns.filter((column: any) => !fields.some((item: any, position: number) => position !== field.name && item?.name === column.name))
                                        .map((column: any) => ({value: column.name, label: column.name}))}/>
                            </Form.Item>
                            <Form.Item name={[field.name, "direction"]}><Select placeholder="排序" aria-label={`字段 ${index + 1} 排序`}
                                options={[{value: "asc", label: "升序"}, {value: "desc", label: "降序"}]}/></Form.Item>
                            {!saving && values.length > 1 && <ActionDropdown arrow={false} menu={{items: [
                                ...(index > 0 ? [{key: "up", label: "上移"}] : []),
                                ...(index < values.length - 1 ? [{key: "down", label: "下移"}] : []),
                                {type: "divider"}, {key: "delete", label: "移除", danger: true},
                            ], onClick: ({key}: {key: string}) => {
                                if (saving || pending.current || !action?.mode) return;
                                if (key === "up") move(field.name, field.name - 1);
                                if (key === "down") move(field.name, field.name + 1);
                                if (key === "delete") remove(field.name);
                            }}}><Button color="default" variant="filled" aria-label={`索引字段 ${index + 1} 操作`} icon={<Icon type="MoreOutlined"/>}/></ActionDropdown>}
                        </div>)}</div>
                        <ConfigProvider wave={{disabled: true}}>
                            <Button block type="dashed" className="schema-field-add" icon={<Icon type="PlusOutlined"/>} disabled={saving || values.length >= columns.length}
                                onClick={() => add({name: columns.find((column: any) => !fields.some((field: any) => field?.name === column.name))?.name, direction: "asc"})}>添加字段</Button>
                        </ConfigProvider>
                    </>}</Form.List>
                    </ConfigProvider>
                </Form.Item>
            </>}
        </Form>
        </ConfigProvider>
    </ProModal>;
};
