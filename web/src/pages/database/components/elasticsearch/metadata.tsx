import {useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Empty, Select} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {useDatabaseClient} from "../shared/context";
import useStyles from "./styles";
import {VirtualTable} from "../shared/client-common";
import MappingForm from "./mapping-form";

const mappingTypes = ["text", "keyword", "long", "integer", "short", "byte", "double", "float", "boolean", "date", "ip", "object", "nested", "binary"];

export default ({record}: any) => {
    const {message, modal} = App.useApp();
    const {operate, version, features, ActionDropdown} = useDatabaseClient();
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const [data, setData] = useState<any>();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [editor, setEditor] = useState<any>();
    const [editorOpen, setEditorOpen] = useState(false);
    const [analyzers, setAnalyzers] = useState<any[]>([]);
    const [analyzersLoading, setAnalyzersLoading] = useState(false);
    const [refreshingFields, setRefreshingFields] = useState(false);
    const [raw, setRaw] = useState(false);
    const [refresh, setRefresh] = useState(0);
    const [documentType, setDocumentType] = useState("");
    const sequence = useRef(0);
    const live = useRef(true);
    const pending = useRef(false);
    const refreshPending = useRef(false);
    const confirmation = useRef<any>();
    const parts = (version?.version || "").split(".").map(Number);
    const legacyStrings = features.text_keyword_fields === false || (features.text_keyword_fields === undefined && Boolean(version?.version) && parts[0] < 5);
    const typedMapping = features.typed_mapping ?? features.typed_documents ?? (Boolean(version?.version) && parts[0] < 7);
    const multiFields = features.multi_fields ?? parts[0] >= 1;
    const canUpdateSearch = features.search_analyzer_update ?? parts[0] >= 2;
    const canRefreshFields = features.mapping_backfill ?? parts[0] >= 5;
    const types = legacyStrings ? ["string", ...mappingTypes.filter((type) => type !== "text" && type !== "keyword")] : mappingTypes;
    const mapping = data?.[record.name]?.mappings || data?.mappings || data || {};
    const documentTypes = typedMapping && !mapping.properties ? Object.keys(mapping).filter((name) => name !== "_default_" && mapping[name] && typeof mapping[name] === "object") : [];
    const selectedType = documentTypes.includes(documentType) ? documentType : documentTypes[0] || "";
    const properties = mapping.properties || mapping[selectedType]?.properties || {};
    const rows = useMemo(() => {
        const values: any[] = [];
        // 保存 properties/fields 的实际路径，嵌套对象与多字段不能按点号拆分后混用。
        const collect = (fields: any, prefix = "", parents: any[] = [], key = "properties") => {
            for (const [name, definition] of Object.entries<any>(fields || {})) {
                if (!definition || typeof definition !== "object") continue;
                const path = [...parents, {key, name, definition}];
                values.push({name: `${prefix}${name}`, type: definition.type || (definition.properties ? "object" : "—"), definition, path, isSubfield: key === "fields"});
                if (definition.properties) collect(definition.properties, `${prefix}${name}.`, path);
                if (definition.fields) collect(definition.fields, `${prefix}${name}.`, path, "fields");
            }
        };
        collect(properties);
        return values;
    }, [data, selectedType, record.name]);
    const rawContent = useMemo(() => JSON.stringify(data || {}, null, 2), [data]);
    useEffect(() => { live.current = true; return () => { live.current = false; sequence.current++; confirmation.current?.destroy(); }; }, []);
    useEffect(() => {
        setEditorOpen(false);
        setRaw(false);
    }, [record.id, record.name, selectedType]);
    useEffect(() => {
        const request = ++sequence.current;
        if (!record.name) { setData(undefined); setLoading(false); return; }
        setLoading(true);
        void operate(record, "mapping").then((data) => {
            if (!live.current || request !== sequence.current) return;
            setData(data);
        }).catch((reason) => { if (live.current && request === sequence.current) { setData(undefined); message.error(reason?.message || "读取索引配置失败"); } })
            .finally(() => { if (live.current && request === sequence.current) setLoading(false); });
    }, [record.id, record.name, operate, refresh]);
    useEffect(() => {
        if (!editorOpen) return;
        let active = true;
        void operate(record, "analyzers").then((data) => {
            if (active) setAnalyzers(Array.isArray(data) ? data : []);
        }).catch((reason) => {
            if (!active) return;
            setAnalyzers([]);
            setEditorOpen(false);
            message.error(reason?.message || "读取分析器失败");
        }).finally(() => { if (active) setAnalyzersLoading(false); });
        return () => { active = false; };
    }, [editorOpen, record.id, record.name, operate]);
    const openEditor = (mode: string, item?: any) => {
        if (pending.current) return;
        setAnalyzers([]);
        setAnalyzersLoading(true);
        setEditor({mode, item, properties});
        setEditorOpen(true);
    };
    const save = async (values: any) => {
        if (!editorOpen || pending.current || loading || analyzersLoading) return;
        pending.current = true; setSaving(true);
        const request = sequence.current;
        try {
            const creating = editor.mode === "create";
            const field: any = creating ? {type: values.type} : {...editor.item.definition};
            const textField = ["text", "string"].includes(field.type) && ![false, "no", "not_analyzed"].includes(field.index);
            if (creating && values.format?.trim() && field.type === "date") field.format = values.format.trim();
            if (creating && values.analyzer && textField) field.analyzer = values.analyzer;
            if (editor.mode !== "subfield" && values.search_analyzer && textField && (creating || canUpdateSearch)) field.search_analyzer = values.search_analyzer;
            const children = multiFields && !editor.item?.isSubfield && ["text", "keyword", "string"].includes(field.type) ? values.fields || [] : [];
            for (const child of children) {
                const definition: any = {type: child.type};
                if (["text", "string"].includes(child.type)) {
                    if (child.analyzer) definition.analyzer = child.analyzer;
                    if (child.search_analyzer) definition.search_analyzer = child.search_analyzer;
                }
                field.fields = {...field.fields, [child.name.trim()]: definition};
            }
            let changed: any;
            if (creating) changed = {properties: {[values.name.trim()]: field}};
            else {
                let definition = field;
                const path = editor.item.path;
                for (let index = path.length - 1; index >= 0; index--) {
                    const branch = {[path[index].name]: definition};
                    if (index === 0) changed = {[path[index].key]: branch};
                    else {
                        const {properties, fields, ...parent} = path[index - 1].definition;
                        definition = {...parent, [path[index].key]: branch};
                    }
                }
            }
            await operate(record, "update_mapping", {mapping: JSON.stringify(typedMapping && selectedType ? {[selectedType]: changed} : changed)});
            if (live.current && request === sequence.current) {
                message.success(children.length && !creating ? "保存成功，可更新文档使已有数据应用新增子字段" : "保存成功");
                setEditorOpen(false);
                setRefresh((value) => value + 1);
            }
        } catch (reason: any) { if (live.current && request === sequence.current) message.error(reason?.message || "保存失败"); }
        finally { pending.current = false; if (live.current) setSaving(false); }
    };
    const refreshFields = () => {
        if (refreshPending.current) return;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: "更新文档", content: "按最新索引结构重新处理当前索引的已有文档，使新增子字段可被搜索。任务将在后台执行；如配置了最终处理管道，也会一并执行。",
            okText: "确认", cancelText: "取消", onOk: async () => {
                if (refreshPending.current || !live.current) return;
                refreshPending.current = true;
                setRefreshingFields(true);
                try { await operate(record, "refresh_fields"); if (live.current) message.success("已加入后台任务，可在系统任务中查看进度"); }
                catch (reason: any) { if (live.current) message.error(reason?.message || "创建任务失败"); }
                finally { refreshPending.current = false; if (live.current) setRefreshingFields(false); }
            }});
    };
    if (!record.name) return <div className="manager-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择索引"/></div>;
    return <div className="manager-stage">
        <div className="manager-toolbar es-mapping-toolbar"><div className="manager-tools">
            {documentTypes.length > 0 && <Select className="manager-picker" size="small" variant="borderless" prefix={<Icon type="TableOutlined"/>} aria-label="文档类型" placeholder="请选择文档类型" value={selectedType} disabled={saving}
                options={documentTypes.map((value) => ({value, label: value}))} onChange={setDocumentType}/>}
            <Button color="default" variant="filled" icon={<Icon type="FileTextOutlined"/>} onClick={() => setRaw(true)} disabled={!data}>完整配置</Button>
        </div><div className="manager-tools es-mapping-actions">
            {canRefreshFields && <Button color="default" variant="filled" icon={<Icon type="SyncOutlined"/>} loading={refreshingFields} onClick={refreshFields}>更新文档</Button>}
            <Button color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} disabled={saving} onClick={() => openEditor("create")}>新增</Button>
            <Button color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} aria-label="刷新索引配置" loading={loading} disabled={saving} onClick={() => { if (!pending.current) setRefresh((value) => value + 1); }}/>
        </div></div>
        <div className="manager-table"><VirtualTable rowKey="name" size="small" scroll={{x: 850}} dataSource={rows} loading={loading} pagination={false}
            locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
            columns={[{title: "字段名称", dataIndex: "name", ellipsis: true, width: 220}, {title: "类型", dataIndex: "type", width: 110},
                {title: "索引", key: "index", width: 90, render: (_: any, item: any) => [false, "no"].includes(item.definition.index) ? "关闭" : "开启"},
                {title: "写入分析器", key: "analyzer", ellipsis: true, width: 160, render: (_: any, item: any) => ["text", "string"].includes(item.type) && item.definition.index !== "not_analyzed" ? item.definition.analyzer || "默认" : "—"},
                {title: "搜索分析器", key: "search_analyzer", ellipsis: true, width: 160, render: (_: any, item: any) => ["text", "string"].includes(item.type) && item.definition.index !== "not_analyzed" ? item.definition.search_analyzer || item.definition.analyzer || "默认" : "—"},
                {title: "操作", key: "action", width: 76, align: "center", fixed: "right", render: (_: any, item: any) => {
                    const editable = canUpdateSearch && ["text", "string"].includes(item.type) && ![false, "no", "not_analyzed"].includes(item.definition.index);
                    const extensible = multiFields && !item.isSubfield && ["text", "keyword", "string"].includes(item.type);
                    if ((!editable && !extensible) || saving) return null;
                    return <ActionDropdown menu={{items: [
                        ...(editable ? [{key: "edit", label: "编辑", onClick: () => openEditor("edit", item)}] : []),
                        ...(extensible ? [{key: "subfield", label: "扩展", onClick: () => openEditor("subfield", item)}] : []),
                    ]}}><Button color="default" variant="filled" size="small">操作</Button></ActionDropdown>;
                }},
            ]}/></div>
        <ProModal title={<Title>完整配置</Title>} onOk={() => setRaw(false)} onCancel={() => setRaw(false)} modalProps={{
            open: raw, width: theme?.isCompactTheme?.() ? 480 : 520, rootClassName: styles.modal,
            styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}}, style: {maxWidth: "calc(100vw - 24px)"},
            zIndex: 1140, cancelButtonProps: {style: {display: "none"}}, okText: "确认", destroyOnHidden: true, mask: {closable: true},
        }}><pre className="es-code">{rawContent}</pre></ProModal>
        <MappingForm open={editorOpen} editor={editor} types={types} legacyStrings={legacyStrings} analyzers={analyzers} loading={analyzersLoading} saving={saving}
            multiFields={multiFields} canUpdateSearch={canUpdateSearch} onCancel={() => { if (!pending.current) setEditorOpen(false); }} onSubmit={save}/>
    </div>;
};
