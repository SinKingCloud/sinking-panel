import {useEffect, useMemo, useRef, useState} from "react";
import {createPortal} from "react-dom";
import {App, Button, ConfigProvider, Empty, Form, Grid, Input, Pagination, Select, Space, Spin, Tooltip, Typography, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {VirtualTable} from "../client-common";
import {useMysqlClient} from "../context";
import SqlForeignKeyFields, {validateForeignKeys} from "./foreign-key-fields";
import useStyles from "../component-styles/client";
import {copyTextToClipboard} from "../clipboard";

const labels: Record<string, string> = {view: "视图", routine: "存储过程与函数", trigger: "触发器", foreign_key: "外键关系"};

export default ({record, schema, table, columns = [], tables = [], schemas = [], onChange, kind: controlledKind, tableOnly = false, onBrowse, reloadKey = 0, toolbarContainer}: any) => {
    const {api, ActionDropdown, Editor: AceEditor, features = {}} = useMysqlClient();
    const {message, modal} = App.useApp();
    const {styles} = useStyles();
    const theme = useTheme();
    const screens = Grid.useBreakpoint();
    const {token} = antdTheme.useToken();
    const [form] = Form.useForm();
    const [selectedKind, setKind] = useState("view");
    const kindFeatures: Record<string, string> = {view: "views", routine: "routines", trigger: "triggers", foreign_key: "foreign_keys"};
    const requestedKind = controlledKind || selectedKind;
    const kind = features[kindFeatures[requestedKind]] !== false ? requestedKind : Object.keys(labels).find((value) => features[kindFeatures[value]] !== false) || "";
    const [keyword, setKeyword] = useState("");
    const [query, setQuery] = useState({keyword: "", page: 1, pageSize: 20});
    const [pagination, setPagination] = useState<{total: number; page: number; pageSize: number}>();
    const [rows, setRows] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [version, setVersion] = useState(0);
    const [editor, setEditor] = useState<any>();
    const lastEditor = useRef<any>();
    const displayedEditor = editor || lastEditor.current;
    const [reading, setReading] = useState(false);
    const [referencesLoading, setReferencesLoading] = useState(false);
    const initializing = reading || referencesLoading;
    const [saving, setSaving] = useState(false);
    const [keyColumns, setKeyColumns] = useState<any[]>([]);
    const live = useRef(true);
    const sequence = useRef(0);
    const editorSequence = useRef(0);
    const pending = useRef(false);
    const confirm = useRef<any>(undefined);
    const serverPagination = ["view", "routine", "trigger"].includes(kind);
    const name = Form.useWatch("name", {form, preserve: true});
    const parentTable = Form.useWatch("table", {form, preserve: true});
    const objectType = Form.useWatch("object_type", {form, preserve: true});
    const definition = Form.useWatch("definition", {form, preserve: true}) || "";
    const objectLabel = displayedEditor?.kind === "routine" ? objectType === "function" ? "函数" : "存储过程" : labels[displayedEditor?.kind] || "对象";
    const definitionPlaceholder = displayedEditor?.readonly ? undefined : displayedEditor?.item ? "请输入 SQL 定义" : "请输入 SQL 定义，或点击“生成模板”";
    const routineTypes = [{value: "procedure", label: "存储过程"}, {value: "function", label: "函数"}].filter((item) => features[item.value === "procedure" ? "procedures" : "functions"] !== false);
    const quote = (name: string) => `\`${name.replace(/`/g, "``")}\``;
    const qualified = (name: string) => `${quote(schema || record.name)}.${quote(name)}`;
    useEffect(() => {
        live.current = true;
        return () => { live.current = false; sequence.current++; editorSequence.current++; confirm.current?.destroy(); };
    }, []);
    useEffect(() => {
        const request = ++sequence.current;
        setError("");
        if (!serverPagination) setRows([]);
        if (!kind) { setRows([]); setLoading(false); return; }
        setLoading(true);
        void api.objects(record, {schema, kind, ...(serverPagination ? {keyword: query.keyword, page: query.page, page_size: query.pageSize} : {})}).then((data) => {
            if (!live.current || sequence.current !== request) return;
            if (serverPagination) {
                if (!Array.isArray(data?.list) || !Number.isFinite(data?.total) || !Number.isInteger(data?.page) || data?.page < 1 || !Number.isInteger(data?.page_size) || data?.page_size < 1) {
                    throw new Error("对象列表返回的分页数据不完整");
                }
                setPagination({total: data?.total, page: data?.page, pageSize: data?.page_size});
            }
            setRows((Array.isArray(data) ? data : Array.isArray(data?.list) ? data?.list : []).filter((item: any) => item && typeof item === "object"));
        }).catch((reason) => {
            if (!live.current || sequence.current !== request) return;
            setError(reason?.message);
            message.error({key: `mysql-objects-${record.id}`, content: reason?.message});
        })
            .finally(() => { if (live.current && sequence.current === request) setLoading(false); });
        return () => { sequence.current++; };
    }, [record.id, record.database_name, record.name, schema, kind, version, reloadKey, serverPagination, query]);
    const search = (value = keyword) => {
        if (!serverPagination) return;
        const keyword = value.trim();
        setQuery((current) => current.keyword === keyword ? current : {...current, keyword, page: 1});
    };
    const reload = (notify = true) => {
        if (serverPagination && pagination && !loading && !error) {
            setQuery((current) => current.keyword !== query.keyword || current.page === pagination.page ? current : {...current, page: pagination.page});
        }
        setVersion((value) => value + 1);
        if (notify) onChange?.();
    };
    const close = () => { if (pending.current) return; editorSequence.current++; setEditor(undefined); };
    const open = async (item?: any, readonly = false) => {
        if (!kind || features[kindFeatures[kind]] === false) return;
        const request = ++editorSequence.current;
        form.resetFields();
        lastEditor.current = {item, kind, readonly, schema, table: item?.table || table, request};
        setReferencesLoading(kind === "foreign_key" && Boolean(item));
        setEditor(lastEditor.current);
        form.setFieldsValue({name: item?.name || "", table: item?.table || table || undefined,
            keys: kind === "foreign_key" ? [item ? {...item} : {name: "", columns: [], reference_schema: schema, reference_columns: [], on_delete: "NO ACTION", on_update: "NO ACTION"}] : [],
            object_type: kind === "routine" ? routineTypes[0]?.value : "view", definition: ""});
        if (kind === "foreign_key") {
            setKeyColumns(item?.table && item.table !== table ? [] : columns);
            if (!item?.table || item.table === table) { setReading(false); return; }
            setReading(true);
            try {
                const data = await api.columns(record, {schema, table: item.table});
                if (live.current && editorSequence.current === request) setKeyColumns(Array.isArray(data) ? data.filter((item: any) => typeof item?.name === "string") : []);
            } catch (reason: any) {
                if (live.current && editorSequence.current === request) {
                    editorSequence.current++;
                    setEditor(undefined);
                    message.error({key: `mysql-object-definition-${record.id}`, content: reason?.message || "获取字段失败"});
                }
            }
            finally { if (live.current && editorSequence.current === request) setReading(false); }
            return;
        }
        if (!item) { setReading(false); return; }
        setReading(true);
        try {
            const data = await api.object_definition(record, {schema, kind, name: item.identity || item.name});
            if (live.current && editorSequence.current === request) form.setFieldValue("definition", typeof data === "string" ? data : data?.definition || "");
        } catch (reason: any) {
            if (live.current && editorSequence.current === request) {
                editorSequence.current++;
                setEditor(undefined);
                message.error({key: `mysql-object-definition-${record.id}`, content: reason?.message || "获取对象定义失败"});
            }
        }
        finally { if (live.current && editorSequence.current === request) setReading(false); }
    };
    const template = () => {
        if (!name?.trim()) { message.info("请先填写对象名称"); return; }
        if (editor.kind === "routine" && !routineTypes.some((item) => item.value === objectType)) { message.info("当前版本不支持这个对象类型"); return; }
        const target = qualified(name.trim());
        if (editor.kind === "view") form.setFieldValue("definition", `CREATE VIEW ${target} AS\nSELECT 1 AS value;`);
        else if (editor.kind === "routine") form.setFieldValue("definition", objectType === "function"
            ? `CREATE FUNCTION ${target}() RETURNS INT\nDETERMINISTIC\nRETURN 1;`
            : `CREATE PROCEDURE ${target}()\nBEGIN\n    SELECT 1;\nEND`);
        else {
            if (!parentTable) { message.info("请选择触发器所属的数据表"); return; }
            const targetTable = qualified(parentTable);
            form.setFieldValue("definition", `CREATE TRIGGER ${target} BEFORE INSERT ON ${targetTable}\nFOR EACH ROW\nBEGIN\n    -- 在此填写触发器逻辑\nEND`);
        }
    };
    const save = async (values: any) => {
        if (!editor || editor.readonly || pending.current || initializing || features[kindFeatures[editor?.kind]] === false) return;
        if (!editor.item && (editor.kind === "routine" && !routineTypes.some((item) => item.value === values.object_type))) { message.error("当前版本不支持这个对象类型"); return; }
        if (editor.kind === "foreign_key") {
            try { validateForeignKeys(values.keys); if (values.keys.length !== 1) throw new Error("请配置一个外键"); }
            catch (reason: any) { message.error(reason?.message); return; }
        }
        const current = editor;
        const request = editorSequence.current;
        const execute = async () => {
            pending.current = true; setSaving(true);
            try {
                await (current.kind === "foreign_key" ? current.item ? api.update_foreign_key : api.create_foreign_key : api.save_object)(record, current.kind === "foreign_key"
                    ? {schema: current.schema, table: current.table, ...(current.item ? {name: current.item.name} : {}), foreign_key: JSON.stringify(values.keys[0])}
                    : {schema: current.schema, kind: current.kind, name: current.item?.identity || current.item?.name || values.name.trim(), definition: values.definition, replace: Boolean(current.item)});
                if (live.current) { message.success("保存成功"); if (request === editorSequence.current) setEditor(undefined); reload(); }
            } catch (reason: any) { if (live.current) message.error(reason?.message); throw reason; }
            finally { pending.current = false; if (live.current) setSaving(false); }
        };
        confirm.current?.destroy();
        confirm.current = modal.confirm({title: current.item ? "保存对象修改？" : `创建${labels[current.kind]}？`,
            content: "对象定义将直接提交给数据库执行，请确认当前数据库与 SQL 内容。", okText: "保存", cancelText: "取消", onOk: execute});
        // 启用 AntD Promise 模式：失败保留弹窗重试，避免再次抛出未处理的拒绝。
        void Promise.resolve(confirm.current);
    };
    const remove = (item: any) => {
        confirm.current?.destroy();
        confirm.current = modal.confirm({title: `删除“${item.name}”？`, content: "对象和依赖关系将按数据库规则删除，此操作不可撤销。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true}, onOk: async () => {
                if (pending.current) return;
                pending.current = true; setSaving(true);
                try {
                    await (kind === "foreign_key" ? api.drop_foreign_key : api.drop_object)(record, kind === "foreign_key"
                        ? {schema, table: item.table, name: item.name} : {schema, kind, name: item.identity || item.name});
                    if (live.current) { message.success("删除成功"); reload(); }
                } catch (reason: any) { if (live.current) message.error(reason?.message); throw reason; }
                finally { pending.current = false; if (live.current) setSaving(false); }
            }});
        // 启用 AntD Promise 模式：失败保留弹窗重试，避免再次抛出未处理的拒绝。
        void Promise.resolve(confirm.current);
    };
    const visible = useMemo(() => serverPagination ? rows : rows.filter((item) => (!tableOnly || item.table === table) && String(item.name).toLowerCase().includes(keyword.trim().toLowerCase())), [rows, keyword, table, tableOnly, serverPagination]);
    const canAdd = Boolean(kind) && (kind !== "routine" || routineTypes.length > 0) && (kind !== "foreign_key" || (table && !tables.some((item: any) => item.name === table && String(item.type).includes("VIEW"))));
    const addButton = <Button color="default" variant="filled" size={toolbarContainer !== undefined ? "small" : undefined}
        autoInsertSpace={toolbarContainer !== undefined ? false : undefined} icon={<Icon type="PlusOutlined"/>}
        disabled={saving || !canAdd} onClick={() => void open()}>新增</Button>;
    return <>
        {toolbarContainer === undefined ? <div className="client-toolbar">{!controlledKind && <Select variant="filled" placeholder="请选择对象类型" aria-label="对象类型" value={kind} disabled={saving} style={{minWidth: 140}}
            options={Object.entries(labels).filter(([value]) => features[kindFeatures[value]] !== false).map(([value, label]) => ({value, label}))} onChange={(value) => {
                setKind(value); setKeyword(""); setRows([]); setQuery((current) => ({...current, keyword: "", page: 1}));
                setPagination(undefined);
            }}/>}
            <Input className="sql-object-search" variant="filled" allowClear prefix={<Icon type="SearchOutlined"/>}
                aria-label={kind === "foreign_key" ? "搜索外键名称" : "搜索对象名称"} placeholder={kind === "foreign_key" ? "搜索外键名称" : "搜索对象名称"}
                value={keyword} onChange={(event) => {
                    const value = event.target.value; setKeyword(value); if (!value) search("");
                }} onPressEnter={() => search()} onBlur={() => search()}/>
            <div className="mysql-object-tools">{addButton}
                <Button className="mysql-toolbar-refresh" color="default" variant="filled" loading={loading} disabled={saving} aria-label="刷新对象" icon={<Icon type="ReloadOutlined"/>} onClick={() => reload(false)}/>
            </div></div>
            : toolbarContainer && createPortal(addButton, toolbarContainer)}
        {kind === "foreign_key" && !tableOnly && <Typography.Text type="secondary" className="client-hint">{`新建外键应用于左侧选中的数据表${table ? `“${table}”` : "，请先选择数据表"}。`}</Typography.Text>}
        <VirtualTable className="mysql-fill-table" size="small" loading={loading} rowKey={(item) => JSON.stringify([item.schema, item.table, item.type, item.identity || item.name])}
            dataSource={visible} rowClassName={(item, index) => item === visible[visible.length - 1] || !serverPagination && index === 19 ? "client-last-row" : ""} scroll={{x: serverPagination ? 280 + 180 + 140 + 84 : 600, scrollToFirstRowOnChange: false}} pagination={serverPagination ? false : {pageSize: 20, simple: !screens.md, showSizeChanger: false, hideOnSinglePage: true}}
            locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
            columns={[
                {title: "名称", dataIndex: "name", width: serverPagination ? 280 : 200, ellipsis: true},
                {title: "所属对象", render: (_, item) => item.table || item.schema || schema || "—", width: serverPagination ? 180 : 140, ellipsis: true},
                {title: kind === "foreign_key" ? "关联关系" : "类型", width: serverPagination ? 140 : undefined, render: (_, item) => kind === "foreign_key"
                    ? `${(item.columns || []).join(", ")} → ${item.reference_table || ""} (${(item.reference_columns || []).join(", ")})` : item.type || labels[kind], ellipsis: true},
                {title: "操作", key: "action", width: 84, align: "center", fixed: "right", render: (_, item) => saving || loading ? null : <ActionDropdown menu={{items: kind === "foreign_key"
                    ? [{key: "edit", label: "编辑", onClick: () => void open(item)},
                        {key: "drop", label: "删除", danger: true, onClick: () => remove(item)}]
                    : [...(kind === "view" && onBrowse ? [{key: "browse", label: "浏览", onClick: () => onBrowse(item.name)}] : []), {key: "view", label: "详情", onClick: () => void open(item, true)}, {key: "edit", label: "编辑", onClick: () => void open(item)},
                        {key: "drop", label: "删除", danger: true, onClick: () => remove(item)}]}}><Button size="small" color="default" variant="filled" autoInsertSpace={false}>操作</Button></ActionDropdown>},
            ]}/>
        {serverPagination && <div className="client-footer" style={{visibility: pagination ? undefined : "hidden"}} aria-hidden={!pagination}>
            <Tooltip title={!screens.md && pagination ? `${rows.length} 条 / 共 ${pagination.total} 条` : undefined} trigger={["hover", "click"]}>
                <Typography.Text type="secondary">{screens.md ? pagination ? `${rows.length} 条 / 共 ${pagination.total} 条` : "— 条 / 共 — 条" : `${pagination?.total ?? "—"} 条`}</Typography.Text>
            </Tooltip>
            <ConfigProvider input={{variant: "filled"}}>
                <Pagination className="client-pagination" size="small" current={pagination?.page ?? 1} pageSize={pagination?.pageSize ?? query.pageSize} total={pagination?.total ?? 0}
                    disabled={!pagination || loading || saving || Boolean(error)} showLessItems simple={!screens.md} showQuickJumper={Boolean(screens.md)} responsive
                    locale={screens.md ? undefined : {items_per_page: "/页"}}
                    showSizeChanger={{showSearch: false, variant: "filled", getPopupContainer: (trigger) => trigger.ownerDocument.body}}
                    pageSizeOptions={[10, 20, 50, 100, 200, 500, 1000]} onChange={(page, pageSize) => {
                        setQuery((current) => ({...current, page: pageSize === pagination?.pageSize ? page : 1, pageSize}));
                    }}/>
            </ConfigProvider>
        </div>}
        <ProModal title={<Title>{displayedEditor?.kind === "foreign_key" ? displayedEditor.item ? "编辑外键" : "新建外键" : displayedEditor?.readonly ? "对象定义" : displayedEditor?.item ? "编辑对象" : displayedEditor?.kind === "routine" ? "新建对象" : `新建${labels[displayedEditor?.kind] || "对象"}`}</Title>}
            onCancel={close} onOk={displayedEditor?.readonly ? close : form.submit} okText={displayedEditor?.readonly ? "确认" : "保存"} modalProps={{
                open: Boolean(editor) && features[kindFeatures[displayedEditor?.kind]] !== false,
                width: displayedEditor?.kind === "foreign_key" ? (theme?.isCompactTheme?.() ? 480 : 520)
                    : displayedEditor?.readonly ? (theme?.isCompactTheme?.() ? 480 : 520)
                    : displayedEditor?.kind === "view" && !displayedEditor?.item ? (theme?.isCompactTheme?.() ? 500 : 560) : (theme?.isCompactTheme?.() ? 540 : 600),
                style: {maxWidth: "calc(100vw - 24px)"},
                cancelText: "取消", confirmLoading: saving, closable: !saving, keyboard: !saving, mask: {closable: !saving}, destroyOnHidden: true,
                okButtonProps: {disabled: initializing}, cancelButtonProps: displayedEditor?.readonly ? {style: {display: "none"}} : undefined,
            }}>
            <ConfigProvider theme={{components: {
                Input: {controlHeight: token.controlHeight},
                Select: {controlHeight: token.controlHeight},
            }}}>
            {initializing && <div className={styles.loading}><Spin/></div>}
            {/* 引用字段由子组件加载，首次准备完成前只隐藏表单，保留挂载以完成请求。 */}
            {!reading && <Form form={form} layout="vertical" variant="filled" disabled={saving} onFinish={save} style={referencesLoading ? {display: "none"} : undefined}>
                {displayedEditor?.kind === "foreign_key" ? <Form.Item name="keys" validateTrigger={false} rules={[{validator: async (_, keys) => { validateForeignKeys(keys); if (keys?.length !== 1) throw new Error("请设置一个外键"); }}]}>
                    <SqlForeignKeyFields record={record} table={displayedEditor?.table} columns={keyColumns.map((item: any) => item.name)} tables={tables} schemas={schemas} schema={displayedEditor?.schema} maxItems={1} disabled={saving || reading} active={editor || false}
                        onReady={() => {
                            if (live.current && editor && editorSequence.current === editor.request) setReferencesLoading(false);
                        }}
                        onError={(reason: any) => {
                            if (!live.current || !editor || editorSequence.current !== editor.request) return;
                            editorSequence.current++;
                            setEditor(undefined);
                            message.error({key: `mysql-object-definition-${record.id}`, content: reason?.message || "获取引用字段失败"});
                        }}/>
                </Form.Item> : displayedEditor?.readonly ? <div className={styles.objectDefinition}>
                    <div className="object-definition-toolbar">
                        <div className="object-definition-context">
                            <span title={displayedEditor.item?.schema || displayedEditor.schema}>{displayedEditor.item?.schema || displayedEditor.schema}</span>
                            <span aria-hidden="true">/</span>
                            <span title={displayedEditor.item?.name}>{displayedEditor.item?.name}</span>
                        </div>
                        <Button color="default" variant="filled" size="small" autoInsertSpace={false} icon={<Icon type="CopyOutlined"/>} disabled={!definition}
                            onClick={async () => {
                                if (await copyTextToClipboard(definition)) message.success("复制成功");
                                else message.error("复制失败，请手动复制");
                            }}>复制</Button>
                    </div>
                    <pre aria-label="SQL 定义">{definition || "暂无定义"}</pre>
                </div> : <>
                    {!displayedEditor?.item && <>
                        <Form.Item label="对象名称" required htmlFor="name">
                            <Space.Compact block>
                                <Form.Item name="name" noStyle rules={[{required: true, whitespace: true, message: "请填写对象名称"}]}>
                                    <Input placeholder={`请输入${objectLabel}名称`}/>
                                </Form.Item>
                                <Button color="default" variant="filled" onClick={template}>生成模板</Button>
                            </Space.Compact>
                        </Form.Item>
                        {displayedEditor?.kind === "routine" && <Form.Item name="object_type" label="对象类型" rules={[{required: true, message: "请选择对象类型"}]}>
                            <Select placeholder="请选择对象类型" options={routineTypes}/></Form.Item>}
                        {displayedEditor?.kind === "trigger" && <Form.Item name="table" label="所属表" rules={[{required: true, whitespace: true, message: "请选择所属数据表"}]}>
                            <Select placeholder="请选择所属数据表" options={tables.filter((item: any) => !String(item.type).includes("VIEW")).map((item: any) => ({value: item.name, label: item.name}))}/></Form.Item>}
                    </>}
                    <Form.Item label="SQL 定义" required>
                        <Form.Item name="definition" noStyle rules={[{required: true, whitespace: true, message: "请输入 SQL 定义"}]}>
                            {AceEditor ? <AceEditor mode="sql" className={styles.codeEditor} readOnly={saving || reading} placeholder={definitionPlaceholder}
                                height={160} fontSize={12} theme={theme?.isDarkTheme?.() ? "tomorrow_night" : "textmate"} showPrintMargin={false} wrapEnabled/> : <Input.TextArea className={styles.codeEditor} readOnly={saving || reading} placeholder={definitionPlaceholder} style={{height: 160, resize: "none"}}/>}
                        </Form.Item>
                    </Form.Item>
                </>}
            </Form>}
            </ConfigProvider>
        </ProModal>
    </>;
};
