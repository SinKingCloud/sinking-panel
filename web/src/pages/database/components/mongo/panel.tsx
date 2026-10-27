import {useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Checkbox, ConfigProvider, Empty, Form, Grid, Input, InputNumber, Pagination, Select, Segmented, Spin, Tabs, Tooltip, Typography, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {DownOutlined, PlusOutlined, ReloadOutlined} from "@ant-design/icons";
import {VirtualTable} from "../shared/client-common";
import {useDatabaseClient} from "../shared/context";
import useStyles from "./styles";
import Accounts from "./accounts";
import Maintenance from "./maintenance";
import MongoFieldsEditor from "./fields-editor";
import MongoQueryEditor, {emptyMongoQuery, MongoOrderEditor, newMongoOrder} from "./query-editor";
import {mongoDocument, mongoFields, mongoJSON, mongoOrderedFields, mongoQuery, newMongoField} from "./field-data";
import ServerTransfer from "../shared/server-transfer";
import MongoObjectEditor from "./object-editor";
import MongoAggregate from "./aggregate";


export default ({record, initialTab, onTabChange}: any) => {
    const {operate, ActionDropdown, Editor, features = {}} = useDatabaseClient();
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const {token} = antdTheme.useToken();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const screens = Grid.useBreakpoint();
    const live = useRef(true);
    const collectionRequest = useRef(0);
    const documentRequest = useRef(0);
    const indexRequest = useRef(0);
    const editRequest = useRef(0);
    const collectionRef = useRef("");
    const busyRef = useRef(false);
    const confirmations = useRef(new Set<any>());
    const [collections, setCollections] = useState<any[]>([]);
    const [collection, setCollection] = useState("");
    const [catalogLoading, setCatalogLoading] = useState(false);
    const [catalogVersion, setCatalogVersion] = useState(0);
    const [indexes, setIndexes] = useState<any[]>([]);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(20);
    const [version, setVersion] = useState(0);
    const [result, setResult] = useState<any>();
    const [sampleDocuments, setSampleDocuments] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);
    const [queryDraft, setQueryDraft] = useState(emptyMongoQuery);
    const [conditions, setConditions] = useState<any>({filter: "{}", sort: "{}", projection: "{}"});
    const [appliedQueryCount, setAppliedQueryCount] = useState(0);
    const [indexLoading, setIndexLoading] = useState(false);
    const [selectedTab, setSelectedTab] = useState(["command", "query"].includes(initialTab) ? "aggregate" : initialTab || "documents");
    const [aggregateOpened, setAggregateOpened] = useState(selectedTab === "aggregate");
    const setTab = (value: string) => { if (value === "aggregate") setAggregateOpened(true); setSelectedTab(value); onTabChange?.(value); };
    const tab = selectedTab === "accounts" && features.accounts === false ? "documents" : selectedTab;
    const [action, setAction] = useState<any>();
    const lastAction = useRef<any>(undefined);
    useEffect(() => { if (action) lastAction.current = action; }, [action]);
    const displayAction = action || lastAction.current;
    const [saving, setSaving] = useState(false);
    const [documentMode, setDocumentMode] = useState("fields");
    const [documentFallback, setDocumentFallback] = useState(false);
    const [form] = Form.useForm();
    const indexExpiry = Form.useWatch("index_expiry", form) || "none";
    // 每批文档生成稳定行标识，不依赖 Table 的 index 参数。
    const documentRows = useMemo(() => {
        const occurrences = new Map<string, number>();
        return (Array.isArray(result?.documents) ? result?.documents : []).map((document: any, index: number) => {
            const identity = document?._id === undefined ? `row:${index}` : `id:${JSON.stringify(document?._id)}`;
            const occurrence = occurrences.get(identity) || 0;
            occurrences.set(identity, occurrence + 1);
            return {key: JSON.stringify([identity, occurrence]), document};
        });
    }, [result?.documents]);
    const documentFields = useMemo(() => Array.from(new Set<string>(documentRows.flatMap(({document}: any) => Object.keys(document || {})))), [documentRows]);
    const sampleFields = useMemo(() => Array.from(new Set<string>(sampleDocuments.flatMap((document) => Object.keys(document || {})))), [sampleDocuments]);
    const [objectAction, setObjectAction] = useState<any>();
    collectionRef.current = collection;

    useEffect(() => { live.current = true; return () => { live.current = false; confirmations.current.forEach((item) => item.destroy()); confirmations.current.clear(); }; }, []);
    useEffect(() => {
        const request = ++collectionRequest.current;
        setCatalogLoading(true);
        void operate(record, "collections").then((data) => {
            if (!live.current || request !== collectionRequest.current) return;
            const values = Array.isArray(data) ? data.filter((item) => typeof item?.name === "string") : [];
            setCollections(values);
            const next = values.some((item) => item.name === collectionRef.current) ? collectionRef.current : values[0]?.name || "";
            if (next !== collectionRef.current) choose(next);
        }).catch((reason) => { if (live.current && request === collectionRequest.current) message.error(reason?.message || "操作失败"); })
            .finally(() => { if (live.current && request === collectionRequest.current) setCatalogLoading(false); });
    }, [record.id, record.name, catalogVersion, operate]);
    useEffect(() => {
        const request = ++documentRequest.current;
        setLoading(false);
        if (!collection) { setResult(undefined); return; }
        if (tab !== "documents") return;
        setLoading(true);
        void operate(record, "documents", {collection, ...conditions, page, page_size: pageSize}).then((data) => {
            if (live.current && request === documentRequest.current) {
                setResult({...data, documents: Array.isArray(data?.documents) ? data?.documents.filter(Boolean) : []});
                // 留下少量完整样本，筛选为空或隐藏字段后仍可选择字段并识别类型。
                if (Array.isArray(data?.documents) && data?.documents?.length && conditions.projection === "{}") setSampleDocuments(data?.documents?.slice(0, 20));
                if (data?.truncated) message.warning("本页数据较大，请减少每页条数或筛选显示字段后继续浏览");
            }
        }).catch((reason) => { if (live.current && request === documentRequest.current) { setResult(undefined); message.error(reason?.message || "操作失败"); } })
            .finally(() => { if (live.current && request === documentRequest.current) setLoading(false); });
    }, [record.id, record.name, collection, conditions, page, pageSize, version, tab, operate]);
    useEffect(() => {
        const request = ++indexRequest.current;
        setIndexLoading(false);
        if (tab !== "indexes") return;
        if (!collection || collections.some((item) => item.name === collection && item.type === "view")) { setIndexes([]); return; }
        setIndexLoading(true);
        void operate(record, "indexes", {collection}).then((data) => {
            if (live.current && request === indexRequest.current) setIndexes(Array.isArray(data) ? data.filter(Boolean) : []);
        }).catch((reason) => { if (live.current && request === indexRequest.current) { setIndexes([]); message.error(reason?.message || "操作失败"); } })
            .finally(() => { if (live.current && request === indexRequest.current) setIndexLoading(false); });
    }, [record.id, record.name, collection, version, collections, tab, operate]);

    const reload = () => setVersion((value) => value + 1);
    const choose = (name: string) => {
        if (name !== collectionRef.current) {
            documentRequest.current += 1; indexRequest.current += 1; editRequest.current += 1;
            setResult(undefined); setSampleDocuments([]); setIndexes([]);
            setConditions({filter: "{}", sort: "{}", projection: "{}"}); setQueryDraft(emptyMongoQuery()); setAppliedQueryCount(0);
        }
        collectionRef.current = name;
        setCollection(name); setPage(1);
        if (tab === "indexes" && collections.some((item) => item.name === name && item.type === "view")) setTab("documents");
    };
    const applyFilters = (draft = queryDraft) => {
        try {
            const query = draft.query || mongoQuery(draft.filters, draft.sort, draft.projection, draft.projectionMode, draft.logic);
            setConditions(query);
            setAppliedQueryCount(draft.query ? Object.values(query).filter((value) => value !== "{}").length : draft.filters.length + draft.sort.length + draft.projection.length);
            setPage(1);
            return true;
        } catch (reason: any) { message.error(reason?.message || "操作失败"); return false; }
    };
    const showAction = async (operation: string, document?: any) => {
        const request = ++editRequest.current;
        const targetCollection = collection;
        const initialValues: any = {};
        let partial = false;
        if (["update", "view"].includes(operation) && document?._id !== undefined && !(operation === "view" && collections.some((item) => item.name === collection && item.type === "view"))) {
            try {
                // 列表可能使用字段投影，替换前重新读取完整文档，避免删除未显示的字段。
                const response = await operate(record, "documents", {collection: targetCollection, filter: JSON.stringify({_id: document._id}), page: 1, page_size: 1});
                if (!live.current || request !== editRequest.current || targetCollection !== collectionRef.current) return;
                if (!response?.documents?.length || (operation === "update" && response?.truncated)) throw new Error(response?.truncated ? "文档内容不完整，无法编辑" : "文档已不存在，请刷新列表");
                partial = Boolean(response?.truncated);
                initialValues.content = mongoFields(response?.documents?.[0], true);
            } catch (reason: any) {
                if (live.current && request === editRequest.current) message.error(reason?.message || "操作失败");
                return;
            }
        } else if (operation === "view") { initialValues.content = mongoFields(document); partial = Boolean(result?.truncated) || conditions.projection !== "{}"; }
        else if (operation === "insert") initialValues.content = [newMongoField()];
        else if (operation === "create_index") Object.assign(initialValues, {name: "", fields: [newMongoOrder()], unique: false, sparse: false, index_expiry: "none"});
        setDocumentMode("fields");
        setAction({operation, collection, partial, initialValues, filter: document?._id === undefined ? undefined : {_id: document._id}});
    };
    const readDocument = (text: string) => {
        const document = mongoJSON(text || "");
        // JSON 编辑同样保留原文档标识，转换为字段时继续锁定 _id。
        if (displayAction?.operation === "update") document._id = displayAction.filter._id;
        return document;
    };
    const changeDocumentMode = (mode: string) => {
        if (mode === documentMode) return;
        try {
            if (mode === "json") {
                // 新增的一行空白输入不属于文档内容，允许直接切换后粘贴 JSON。
                const fields = (form.getFieldValue("content") || []).filter((field: any) => field.existing || field.key || field.type !== "string" || field.value);
                form.setFieldValue("document", JSON.stringify(mongoDocument(fields), null, 2));
            } else {
                const fields = mongoFields(readDocument(form.getFieldValue("document")), displayAction?.operation === "update");
                form.setFieldValue("content", fields);
            }
            setDocumentMode(mode);
        } catch (reason: any) { message.error(reason?.message || "操作失败"); }
    };
    const submit = async (values: any) => {
        if (busyRef.current || !action || action.operation === "view") return;
        let body: any = {collection: action.collection};
        try {
            if (action.operation === "create_collection") body.collection = values.name.trim();
            else if (action.operation === "create_index") {
                const indexFeatures: Record<string, string> = {text: "text_indexes", hashed: "hashed_indexes", "2dsphere": "sphere_indexes"};
                if (values.fields.some((field: any) => features[indexFeatures[field.direction]] === false)) throw new Error("当前版本不支持所选索引方式");
                const keys = mongoOrderedFields(values.fields, "索引");
                const options: any = {unique: Boolean(values.unique), sparse: Boolean(values.sparse)};
                if (values.index_expiry === "ttl") {
                    if (features.ttl_indexes === false) throw new Error("当前版本不支持自动过期索引");
                    if (values.fields.length !== 1) throw new Error("自动过期索引只能选择一个日期字段");
                    options.expire_after_seconds = values.ttl;
                }
                body = {...body, name: values.name?.trim() || "", index: `{"keys":${keys},${JSON.stringify(options).slice(1)}`};
            } else {
                const document = documentMode === "json" ? readDocument(values.document) : mongoDocument(values.content);
                if (action.operation === "insert") body.document = JSON.stringify(document);
                else {
                    // 编辑期间固定主键；即使表单状态改变，也不能修改文档标识。
                    document._id = action.filter._id;
                    body = {...body, filter: JSON.stringify(action.filter), update: JSON.stringify(document), many: false, upsert: false};
                }
            }
        } catch (reason: any) { message.error(reason?.message || "操作失败"); return; }
        busyRef.current = true; setSaving(true);
        try {
            const response = await operate(record, action.operation, body);
            if (action.operation === "update" && response?.matched === 0) throw new Error("文档已被删除或修改，请重新读取后再保存");
            if (!live.current) return;
            message.success("保存成功"); setAction(undefined); reload();
            if (action.operation === "create_collection") { setCatalogVersion((value) => value + 1); choose(body.collection); }
        } catch (reason: any) { if (live.current) message.error(reason?.message || "操作失败"); }
        finally { busyRef.current = false; if (live.current) setSaving(false); }
    };
    const confirmDelete = (operation: string, body: any, title: string) => {
        const confirmation = modal.confirm({title, content: "将直接删除数据库中的内容，此操作不可撤销。", okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            afterClose: () => { confirmations.current.delete(confirmation); }, onOk: async () => {
                try { await operate(record, operation, body); if (live.current) { message.success("删除成功"); reload(); if (operation === "drop_collection") { choose(""); setCatalogVersion((value) => value + 1); } } }
                catch (reason: any) { if (live.current) message.error(reason?.message || "操作失败"); }
            }});
        confirmations.current.add(confirmation);
        void Promise.resolve(confirmation);
    };
    const displayValue = (value: any): string => {
        if (value === null) return "空值";
        if (value === undefined) return "—";
        if (typeof value === "boolean") return value ? "是" : "否";
        if (typeof value !== "object") return String(value);
        if (Object.keys(value).length === 1) {
            for (const key of ["$oid", "$numberInt", "$numberLong", "$numberDouble", "$numberDecimal"]) if (typeof value[key] === "string") return value[key];
        }
        return JSON.stringify(value);
    };
    const documents = () => {
        const rows = documentRows;
        const fields = documentFields;
        const sort = JSON.parse(conditions.sort || "{}");
        return <div className="manager-table"><VirtualTable size="small" rowKey="key" pagination={false} scroll={{x: Math.max(600, fields.length * 170 + 126)}}
            loading={loading} dataSource={rows}
            locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
            onChange={(_: any, __: any, sorter: any, extra: any) => {
                if (extra.action !== "sort") return;
                const sort = sorter.order ? [{...newMongoOrder(), field: sorter.columnKey, direction: sorter.order === "ascend" ? 1 : -1}] : [];
                const order = mongoOrderedFields(sort, "排序");
                setQueryDraft((current) => ({...current, sort, ...(current.query ? {query: {...current.query, sort: order}, shell: ""} : {})}));
                setConditions((current: any) => ({...current, sort: order}));
                setAppliedQueryCount(queryDraft.query ? Object.values({...queryDraft.query, sort: order}).filter((value) => value !== "{}").length : queryDraft.filters.length + sort.length + queryDraft.projection.length);
                setPage(1);
            }}
            columns={[...fields.map((name) => ({title: name, key: name, width: name === "_id" ? 220 : 170, ellipsis: true, sorter: true,
                sortOrder: Number(sort[name]?.$numberInt ?? sort[name]?.$numberLong ?? sort[name]) === 1 ? "ascend" : Number(sort[name]?.$numberInt ?? sort[name]?.$numberLong ?? sort[name]) === -1 ? "descend" : null,
                render: (_: any, {document}: any) => {
                    const text = displayValue(document?.[name]);
                    return <Tooltip title={text.length < 3000 ? text : undefined}><span style={{color: document?.[name] === null ? "var(--ant-color-text-quaternary)" : undefined}}>{text}</span></Tooltip>;
                }})), {title: "操作", key: "action", fixed: "right" as const, align: "center" as const, width: 76, render: (_: any, {document}: any) => {
                    const readonly = currentCollection?.type === "view" || document?._id === undefined;
                    return <ActionDropdown menu={{items: [
                        {key: "view", label: "详情", onClick: () => void showAction("view", document)},
                        ...(!readonly ? [{key: "edit", label: "编辑", onClick: () => void showAction("update", document)},
                            {key: "delete", label: "删除", danger: true, onClick: () => confirmDelete("delete", {collection, filter: JSON.stringify({_id: document._id}), many: false}, "删除这条文档？")}] : []),
                    ]}}><Button color="default" variant="filled" size="small">操作</Button></ActionDropdown>;
                }}]}/></div>;
    };
    const currentCollection = collections.find((item) => item.name === collection);
    const browsing = ["documents", "indexes"].includes(tab);

    return <div className={styles.panel}>
        <div className="manager-header"><Segmented className="manager-navigation" value={browsing ? "documents" : tab} options={[
            {value: "documents", label: "数据浏览"}, {value: "aggregate", label: "聚合查询"},
            ...(features.accounts !== false ? [{value: "accounts", label: "账户权限"}] : []), {value: "maintenance", label: "数据维护"},
        ]} onChange={(value) => setTab(String(value))}/></div>
        <div className="manager-body">
            {browsing && <div className="manager-stage">
                <div className="manager-toolbar mongo-main-toolbar"><div className="manager-tools">
                    <Select className="manager-picker" size="small" variant="borderless" prefix={<Icon type={currentCollection?.type === "view" ? "EyeOutlined" : "TableOutlined"}/>} showSearch optionFilterProp="label" value={collection || undefined} placeholder="选择集合或视图" onChange={choose} loading={catalogLoading}
                        options={collections.map((item) => ({value: item.name, label: item.name}))}/>
                    <div className="mongo-toolbar-actions">
                        {features.views !== false ? <ActionDropdown arrow={false} menu={{items: [
                            {key: "collection", label: "新建集合", onClick: () => void showAction("create_collection")},
                            {key: "view", label: "新建视图", onClick: () => setObjectAction({operation: "create_view", source: collection || undefined})},
                        ]}}><Button className="client-toolbar-dropdown" color="default" variant="filled" icon={<PlusOutlined/>}>新建<DownOutlined className="client-toolbar-arrow"/></Button></ActionDropdown>
                            : <Button color="default" variant="filled" icon={<PlusOutlined/>} onClick={() => void showAction("create_collection")}>新建</Button>}
                        <ServerTransfer dropdown record={record} collection={collection} readOnly={currentCollection?.type === "view"} disabled={!collection} onChange={reload}/>
                        {currentCollection?.type === "view" && features.views !== false && <ActionDropdown arrow={false} menu={{items: [
                            {key: "definition", label: "查看定义", onClick: () => setObjectAction({operation: "view_definition", collection})},
                            {key: "edit", label: "编辑视图", onClick: () => setObjectAction({operation: "update_view", collection})},
                        ]}}><Button className="client-toolbar-dropdown" color="default" variant="filled">视图<DownOutlined className="client-toolbar-arrow"/></Button></ActionDropdown>}
                    </div>
                </div><Button className="manager-refresh" color="default" variant="filled" icon={<ReloadOutlined/>} loading={catalogLoading} aria-label="刷新集合列表" onClick={() => setCatalogVersion((value) => value + 1)}/></div>
                {collection ? <>
                    <Tabs className="manager-tabs" activeKey={tab} onChange={setTab} items={[{key: "documents", label: "数据"}, ...(currentCollection?.type === "view" ? [] : [{key: "indexes", label: "索引"}])]}
                        tabBarExtraContent={<div className="manager-tools">
                            {currentCollection?.type !== "view" && <Button color="default" variant="filled" icon={<PlusOutlined/>} onClick={() => void showAction(tab === "indexes" ? "create_index" : "insert")}>新增</Button>}
                            {tab === "documents" && <MongoQueryEditor collection={collection} databaseName={record.database_name || record.name} value={queryDraft} appliedCount={appliedQueryCount} onChange={setQueryDraft} onApply={applyFilters} sampleDocuments={sampleDocuments} fieldNames={sampleFields}/>}
                            <Button color="default" variant="filled" icon={<ReloadOutlined/>} loading={tab === "indexes" ? indexLoading : loading} aria-label="刷新数据" onClick={reload}/>
                        </div>}/>
                    {tab === "documents" ? <>{documents()}
                        <div className="manager-footer"><span className="manager-statistics">共 {result?.total || 0} 条{result?.elapsed_ms !== undefined ? ` · ${result?.elapsed_ms} ms` : ""}</span>
                            <Pagination size="small" current={result?.page || page} pageSize={pageSize} total={Number(result?.total || 0)} showSizeChanger={Boolean(screens.md)} showQuickJumper={Boolean(screens.md)} simple={!screens.md}
                                pageSizeOptions={[20, 50, 100, 500, 1000]} onChange={(next, size) => { setPage(size === pageSize ? next : 1); setPageSize(size); }}/></div>
                    </> : <div className="manager-table"><VirtualTable rowKey="name" size="small" dataSource={indexes} loading={indexLoading} pagination={false} scroll={{x: 766}}
                        locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
                        columns={[
                            {title: "索引名称", dataIndex: "name", width: 200, ellipsis: true},
                            {title: "字段", key: "keys", width: 250, ellipsis: true, render: (_: any, item: any) => Object.entries(item.keys || {}).map(([name, order]) => `${name} (${displayValue(order)})`).join(", ")},
                            {title: "类型", key: "type", width: 120, render: (_: any, item: any) => item.name === "_id_" ? "主键索引" : item.unique ? "唯一索引" : "普通索引"},
                            {title: "过期时间", dataIndex: "expire_after_seconds", width: 120, render: (value: any) => value === undefined ? "—" : `${value} 秒`},
                            {title: "操作", key: "action", width: 76, align: "center", fixed: "right", render: (_: any, item: any) => item.name === "_id_" ? "—" : <ActionDropdown menu={{items: [{key: "delete", label: "删除", danger: true, onClick: () => confirmDelete("drop_index", {collection, name: item.name}, `删除索引“${item.name}”？`)}]}}><Button color="default" variant="filled" size="small">操作</Button></ActionDropdown>},
                        ]}/></div>}
                </> : <div className="manager-empty"><Spin spinning={catalogLoading}><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/></Spin></div>}
            </div>}
            {aggregateOpened && <MongoAggregate record={record} collection={collection} active={tab === "aggregate"} onChange={reload}/>}
            {tab === "accounts" && <Accounts record={record} initialAction={record.accountAction}/>}
            {tab === "maintenance" && <div className="manager-stage"><div className="manager-toolbar mongo-main-toolbar"><Select className="manager-picker" size="small" variant="borderless" prefix={<Icon type="TableOutlined"/>} value={collection || undefined} placeholder="选择要维护的集合" showSearch optionFilterProp="label"
                options={collections.filter((item) => item.type !== "view").map((item) => ({value: item.name, label: item.name}))} onChange={choose}/></div>
                <Maintenance key={collection} record={record} collection={currentCollection?.type === "view" ? undefined : collection} onChange={(operation: string) => { reload(); if (operation === "drop") { choose(""); setCatalogVersion((value) => value + 1); } }}/>
            </div>}
        </div>
        <ProModal title={<Title>{({insert: "添加文档", update: "编辑文档", view: "查看文档", create_collection: "新建集合", create_index: "添加索引"} as any)[displayAction?.operation]}</Title>}
            onCancel={() => { if (!saving) setAction(undefined); }} onOk={() => displayAction?.operation === "view" ? setAction(undefined) : form.submit()} okText={displayAction?.operation === "view" ? "确认" : "保存"}
            modalProps={{open: Boolean(action), rootClassName: styles.modal,
                width: displayAction?.operation === "create_collection" ? (compact ? 340 : 360) : displayAction?.operation === "create_index" ? (compact ? 440 : 480) : (compact ? 480 : 540), zIndex: 1140,
                styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}}, style: {maxWidth: "calc(100vw - 24px)"}, confirmLoading: saving, cancelText: "取消",
                closable: !saving, keyboard: !saving, mask: {closable: !saving}, cancelButtonProps: {disabled: saving, style: displayAction?.operation === "view" ? {display: "none"} : undefined}, destroyOnHidden: true}}>
            {displayAction?.partial && <Typography.Text type="secondary" className="nosql-readonly">当前文档内容不完整，仅可查看。</Typography.Text>}
            <Form form={form} key={`${displayAction?.operation}:${JSON.stringify(displayAction?.filter || {})}`} initialValues={displayAction?.initialValues} clearOnDestroy layout="vertical" variant="filled" onFinish={submit} disabled={saving || displayAction?.operation === "view"}>
                {(displayAction?.operation === "create_collection" || displayAction?.operation === "create_index") && <ConfigProvider theme={{components: {Input: {controlHeight: token.controlHeight}}}}><Form.Item name="name"
                    label={displayAction?.operation === "create_collection" ? "集合名称" : "索引名称"}
                    rules={displayAction?.operation === "create_collection" ? [{required: true, whitespace: true, message: "请输入集合名称"}] : []}><Input placeholder={displayAction?.operation === "create_collection" ? "例如 users" : "留空自动生成"}/></Form.Item></ConfigProvider>}
                {displayAction?.operation === "create_index" ? <>
                    <Form.Item name="fields" label="索引字段" rules={[{required: true, type: "array", min: 1, message: "请至少添加一个索引字段"}, {validator: async (_, value) => {
                        if (!value?.length) return;
                        mongoOrderedFields(value, "索引");
                    }}]}><MongoOrderEditor indexes disabled={saving} fieldNames={Object.keys(result?.documents?.[0] || {})}/></Form.Item>
                    <div className="mongo-index-options">
                        <Form.Item name="unique" valuePropName="checked"><Checkbox>禁止字段值重复</Checkbox></Form.Item>
                        <Form.Item name="sparse" valuePropName="checked"><Checkbox>忽略缺少该字段的文档</Checkbox></Form.Item>
                    </div>
                    {features.ttl_indexes !== false && <ConfigProvider theme={{components: {Select: {controlHeight: token.controlHeight}}}}>
                        <Form.Item name="index_expiry" label="自动清理文档"><Select placeholder="请选择清理方式" options={[{value: "none", label: "不自动清理"}, {value: "ttl", label: "按日期字段自动过期"}]}/></Form.Item>
                    </ConfigProvider>}
                    {features.ttl_indexes !== false && indexExpiry === "ttl" && <Form.Item name="ttl" label="日期经过多少秒后过期" extra="仅支持一个日期字段；填 0 表示到达字段中的日期即过期。" rules={[{required: true, message: "请输入过期时间"}]}>
                        <InputNumber min={0} max={2147483647} precision={0} style={{width: "100%"}} placeholder="请输入过期秒数，例如 3600"/>
                    </Form.Item>}
                </> : displayAction?.operation !== "create_collection" && <ConfigProvider componentSize="middle" theme={{components: {
                    Input: {controlHeight: token.controlHeight}, Select: {controlHeight: token.controlHeight},
                    InputNumber: {controlHeight: token.controlHeight}, DatePicker: {controlHeight: token.controlHeight},
                    Segmented: {borderRadiusSM: token.borderRadius, borderRadiusXS: token.borderRadiusSM},
                }}}>
                    <div className="mongo-document-toolbar">
                        <Segmented size="small" shape="default" value={documentMode} disabled={saving} options={[{value: "fields", label: "字段编辑"}, {value: "json", label: "JSON 编辑"}]} onChange={(value) => changeDocumentMode(String(value))}/>
                        {documentMode === "json" && displayAction?.operation !== "view" && <Button color="default" variant="filled" size="small" disabled={saving} onClick={() => {
                            try { form.setFieldValue("document", JSON.stringify(readDocument(form.getFieldValue("document")), null, 2)); }
                            catch (reason: any) { message.error(reason?.message || "操作失败"); }
                        }}>美化</Button>}
                    </div>
                    {documentMode === "fields" ? <Form.Item key="content" name="content" validateTrigger="onBlur" rules={[{validator: async (_, value) => { mongoDocument(value || []); }}]}>
                        <MongoFieldsEditor disabled={saving} readOnly={displayAction?.operation === "view"}/>
                    </Form.Item> : <Form.Item key="document" name="document" validateTrigger="onBlur" rules={[{validator: async (_, value) => { readDocument(value); }}]}>
                        {Editor && !documentFallback ? <Editor className="mongo-document-json" mode="json" height={compact ? 220 : 240} fontSize={compact ? 12 : 13} tabSize={2}
                            theme={theme?.isDarkTheme?.() ? "tomorrow_night" : "textmate"} placeholder={'请输入 JSON 文档，例如 {"name": "张三"}'}
                            showPrintMargin={false} wrapEnabled readOnly={saving || displayAction?.operation === "view"} onError={() => setDocumentFallback(true)}/>
                            : <Input.TextArea className="mongo-document-json" rows={9} readOnly={displayAction?.operation === "view"} placeholder={'请输入 JSON 文档，例如 {"name": "张三"}'}/>}
                    </Form.Item>}
                </ConfigProvider>}
            </Form>
        </ProModal>
        <MongoObjectEditor record={record} collections={collections} action={objectAction} onClose={() => setObjectAction(undefined)}
            onChange={(name: string) => { choose(name); setCatalogVersion((value) => value + 1); reload(); }}/>
    </div>;
};
