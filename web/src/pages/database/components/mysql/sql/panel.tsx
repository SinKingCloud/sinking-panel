import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, Empty, Form, Grid, Pagination, Select, Space, Spin, Splitter, Tabs, Tooltip, Typography} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {ResultTable, VirtualTable} from "../client-common";
import {useMysqlClient} from "../context";
import useStyles from "../component-styles/client";
import {SqlField, SqlFilters} from "./fields";
import {encodeSqlFields, encodeSqlFilters, encodeSqlValue, sqlFieldValue} from "./values";
import SqlSchemaEditor from "./schema-editor";
import SqlObjects from "./objects";
import SqlTransfer from "../server-transfer";
import SqlQueryEditor from "./query-editor";
import SqlDefinition from "./definition";
import {copyTextToClipboard} from "../clipboard";
import usePanelStyles from "../component-styles/sql-panel";
import type {SqlPanelProps} from "./entities";

export default ({record, initialTab, engine, section = "browse", hidden = false, onTargetChange, onBrowse, reloadKey = 0, tableValue}: SqlPanelProps) => {
    const {api, ActionDropdown, features = {}} = useMysqlClient();
    const {message, modal} = App.useApp();
    const compact = useTheme()?.isCompactTheme?.();
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles();
    const {styles: panelStyles} = usePanelStyles();
    const live = useRef(true);
    const catalogRequest = useRef(0);
    const selectedRef = useRef("");
    const rowRequest = useRef(0);
    const metadataRequest = useRef(0);
    const controlledSelection = useRef<string>();
    const reportedSelection = useRef<string>();
    const metadataKey = useRef("");
    const rowKey = useRef("");
    const busyRef = useRef(false);
    const confirmations = useRef(new Set<any>());
    const schema = record.database_name || record.name;
    const schemas = useMemo(() => [schema], [schema]);
    const [tables, setTables] = useState<any[]>([]);
    const [selected, setSelected] = useState("");
    const [catalogLoading, setCatalogLoading] = useState(false);
    const [catalogError, setCatalogError] = useState("");
    const [metadataLoading, setMetadataLoading] = useState(false);
    const [metadataError, setMetadataError] = useState("");
    const [columns, setColumns] = useState<any[]>([]);
    const [indexes, setIndexes] = useState<any[]>([]);
    const [selectedObject, setObject] = useState(["view", "routine", "trigger"].includes(initialTab || "") ? initialTab! : "view");
    const [selectedTab, setTab] = useState(["objects", "view", "routine", "trigger"].includes(initialTab || "") ? "data"
        : initialTab === "structure" ? "columns"
        : ["accounts", "maintenance", "query", "browse"].includes(initialTab || "") ? "data" : initialTab || "data");
    const objectFeatures: Record<string, string> = {view: "views", routine: "routines", trigger: "triggers", foreign_key: "foreign_keys"};
    const tab = features[objectFeatures[selectedTab]] === false ? "data" : selectedTab;
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(20);
    const [sorting, setSorting] = useState({target: "", field: "", order: ""});
    const rowTarget = JSON.stringify([record.id, schema, selected]);
    const sort = sorting.target === rowTarget ? sorting.field : "";
    const order = sorting.target === rowTarget ? sorting.order : "";
    const [result, setResult] = useState<any>();
    const [loading, setLoading] = useState(false);
    const [refresh, setRefresh] = useState(0);
    const [catalogVersion, setCatalogVersion] = useState(0);
    const [tableListVersion, setTableListVersion] = useState(0);
    const [detailVersion, setDetailVersion] = useState(0);
    const [filterFields, setFilterFields] = useState<any[]>([]);
    const [filters, setFilters] = useState("");
    const [appliedFilters, setAppliedFilters] = useState<any[]>([]);
    const [filterLogic, setFilterLogic] = useState("and");
    const [appliedFilterLogic, setAppliedFilterLogic] = useState("and");
    // SQL 草稿由编辑器本地渲染，输入不触发整个数据面板更新；切换功能后仍可恢复。
    const sql = useRef(engine?.query || "SELECT 1;");
    const changeSql = useCallback((value: string) => { sql.current = value; }, []);
    const [filterOpen, setFilterOpen] = useState(false);
    const [queryResult, setQueryResult] = useState<any>();
    const [queryLoading, setQueryLoading] = useState(false);
    const [editor, setEditor] = useState<any>();
    const [displayedEditor, setDisplayedEditor] = useState<any>();
    const [saving, setSaving] = useState(false);
    const [schemaEditor, setSchemaEditor] = useState<any>();
    const [detailToolbar, setDetailToolbar] = useState<HTMLDivElement | null>(null);
    const [schemaBusy, setSchemaBusy] = useState("");
    const [queryExecution, setQueryExecution] = useState<any>();
    const schemaPending = useRef(false);
    const [form] = Form.useForm();
    selectedRef.current = selected;
    const currentTable = tables.find((item) => item.name === selected);
    const isView = String(currentTable?.type).toUpperCase().includes("VIEW");
    // 每批数据只定位一次主键/唯一键列和截断行，渲染操作菜单时直接取用。
    const rowKeys = useMemo(() => {
        if (isView || !Array.isArray(result?.columns) || !result?.columns?.length) return [];
        const names = result?.key_columns ?? (result?.primary_keys?.length ? result?.primary_keys : columns.filter((column) => column.primary_key).map((column) => column.name));
        if (!Array.isArray(names)) return [];
        const positions = names.map((name: string) => result?.columns?.indexOf(name));
        if (!names.length || positions.some((index: number) => index < 0)) return [];
        const truncated = new Set((Array.isArray(result?.truncated_cells) ? result?.truncated_cells : []).map((cell: any) => Array.isArray(cell) ? cell[0] : cell?.row));
        return (Array.isArray(result?.rows) ? result?.rows : []).map((row: any[], index: number) => {
            if (!Array.isArray(row) || truncated.has(index) || positions.some((position: number) => {
                const value = row[position];
                return value == null || (typeof value === "object" && !(value.encoding === "base64" && typeof value.data === "string" && Object.keys(value).length === 2));
            })) return undefined;
            return Object.fromEntries(names.map((name: string, index: number) => [name, row[positions[index]]]));
        });
    }, [result, columns, isView]);

    useEffect(() => {
        live.current = true;
        return () => { live.current = false; confirmations.current.forEach((item) => item.destroy()); };
    }, []);

    useEffect(() => {
        // 独立嵌入时也隔离实例/数据库切换，防止旧表结果回写到新库。
        catalogRequest.current++; rowRequest.current++; metadataRequest.current++;
        selectedRef.current = "";
        setTables([]); setSelected(""); setColumns([]); setIndexes([]); setResult(undefined);
        setPage(1); setFilters(""); setFilterFields([]);
    }, [record.id, schema]);

    useEffect(() => {
        const request = ++catalogRequest.current;
        setCatalogLoading(true);
        setCatalogError("");
        void api.tables(record, {schema}).then((data) => {
            if (!live.current || request !== catalogRequest.current) return;
            const values = Array.isArray(data) ? data.filter((item: any) => typeof item?.name === "string") : [];
            setTables(values);
            if (!values.some((item) => item.name === selectedRef.current)) {
                setSelected(values[0]?.name || ""); setPage(1); setFilters(""); setFilterFields([]);
            }
        }).catch((reason) => {
            if (!live.current || request !== catalogRequest.current) return;
            setCatalogError(reason?.message);
            message.error({key: `mysql-catalog-${record.id}`, content: reason?.message});
        })
            .finally(() => { if (live.current && request === catalogRequest.current) setCatalogLoading(false); });
    }, [record.id, schema, catalogVersion, tableListVersion, reloadKey]);

    useEffect(() => {
        const request = ++metadataRequest.current;
        const key = JSON.stringify([record.id, schema, selected]);
        if (metadataKey.current !== key) { metadataKey.current = key; setColumns([]); setIndexes([]); }
        setMetadataError("");
        if (!selected || selected !== selectedRef.current || hidden || section !== "browse" || !["data", "columns", "indexes", "foreign_key"].includes(tab)) { setMetadataLoading(false); return; }
        setMetadataLoading(true);
        void Promise.all([api.columns(record, {schema, table: selected}), tab === "indexes" ? api.indexes(record, {schema, table: selected}) : Promise.resolve(undefined)])
            .then(([columnData, indexData]) => {
                if (!live.current || request !== metadataRequest.current) return;
                setColumns(Array.isArray(columnData) ? columnData.filter((item: any) => typeof item?.name === "string") : []);
                if (indexData !== undefined) setIndexes(Array.isArray(indexData) ? indexData.filter((item: any) => typeof item?.name === "string") : []);
            }).catch((reason) => {
                if (!live.current || request !== metadataRequest.current) return;
                setMetadataError(reason?.message);
                message.error({key: `mysql-metadata-${record.id}`, content: reason?.message});
            })
            .finally(() => { if (live.current && request === metadataRequest.current) setMetadataLoading(false); });
    }, [record.id, schema, selected, catalogVersion, detailVersion, reloadKey, hidden, section, tab]);

    useEffect(() => {
        const request = ++rowRequest.current;
        if (rowKey.current !== rowTarget) {
            rowKey.current = rowTarget; setResult(undefined);
            // 更换数据表时清除排序，避免把旧表字段带入新表请求。
            setSorting((previous) => previous.field ? {target: "", field: "", order: ""} : previous);
        }
        setLoading(false);
        if (!selected || selected !== selectedRef.current || tab !== "data" || hidden || section !== "browse") return;
        setLoading(true);
        void api.rows(record, {schema, table: selected, page, page_size: pageSize, ...(filters ? {filters, filter_logic: appliedFilterLogic} : {}), ...(sort ? {order_by: sort, order} : {})})
            .then((data) => { if (live.current && request === rowRequest.current) setResult(data); })
            .catch((reason) => {
                if (!live.current || request !== rowRequest.current) return;
                message.error({key: `mysql-rows-${record.id}`, content: reason?.message});
            })
            .finally(() => { if (live.current && request === rowRequest.current) setLoading(false); });
    }, [record.id, schema, selected, rowTarget, page, pageSize, filters, appliedFilterLogic, sort, order, refresh, reloadKey, tab, hidden, section]);

    useEffect(() => { setFilterOpen(false); }, [hidden, section, tab, schema, selected]);

    const selectTable = (name: string) => {
        setFilterOpen(false);
        if (name !== selectedRef.current) { rowRequest.current += 1; metadataRequest.current += 1; setResult(undefined); setColumns([]); setIndexes([]); }
        selectedRef.current = name;
        setSelected(name); setPage(1); setFilters(""); setFilterFields([]);
    };
    const reload = () => setRefresh((value) => value + 1);
    const reloadCatalog = (change?: any) => {
        if (!live.current) return;
        if (change?.table) selectTable(change.table);
        if (["update_column", "drop_column"].includes(change?.mode)) setSorting({target: "", field: "", order: ""});
        setSchemaEditor(undefined); setFilterFields([]); setFilters(""); setPage(1);
        setCatalogVersion((value) => value + 1); reload();
    };
    const deleteObject = (operation: "drop_column" | "drop_index", item: any) => {
        if (schemaPending.current || !selected) return;
        const target = {schema, table: selected};
        const label = operation === "drop_column" ? "字段" : "索引";
        const name = item.name;
        const confirmation = modal.confirm({afterClose: () => confirmations.current.delete(confirmation), title: `删除${label}“${name}”？`,
            content: operation === "drop_column" ? "字段及其全部数据将永久删除，此操作不可撤销。" : "删除后可能影响查询性能或唯一性约束。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true}, onOk: async () => {
                if (schemaPending.current) return;
                schemaPending.current = true; setSchemaBusy(`${operation}:${name}`);
                try {
                    await (operation === "drop_column" ? api.drop_column : api.drop_index)(record, {...target,
                        ...(operation === "drop_column" ? {column_name: name} : {name})});
                    if (live.current) { message.success("删除成功"); reloadCatalog({mode: operation}); }
                } catch (reason: any) { if (live.current) message.error(reason?.message); throw reason; }
                finally { schemaPending.current = false; if (live.current) setSchemaBusy(""); }
            }});
        confirmations.current.add(confirmation);
        // 启用 AntD Promise 模式：失败保留弹窗重试，避免再次抛出未处理的拒绝。
        void Promise.resolve(confirmation);
    };
    const edit = (row?: any[], index?: number) => {
        if (isView) return;
        const keys = row ? rowKeys[index!] : undefined;
        if (row && !keys) return;
        if (!columns.length) return;
        const fields = columns.map((column) => sqlFieldValue(column, row?.[result?.columns?.indexOf(column.name)], Boolean(row)));
        const next = {keys, schema, table: selected, columns, original: row ? fields : undefined};
        setDisplayedEditor(next); setEditor(next);
        form.setFieldValue("fields", fields);
    };
    const save = async ({fields}: any) => {
        if (busyRef.current || !editor) return;
        let changes: string;
        try {
            changes = encodeSqlFields(editor.columns, fields, editor.original);
            if (!changes && editor.keys) { message.info("没有修改内容"); return; }
            // 新增时允许全部采用数据库默认值，更新仍要求实际修改内容。
            if (!changes) changes = "{}";
        } catch (reason: any) { message.error(reason?.message); return; }
        busyRef.current = true; setSaving(true);
        try {
            await (editor.keys ? api.update : api.insert)(record, {schema: editor.schema, table: editor.table,
                values: changes, ...(editor.keys ? {keys: JSON.stringify(editor.keys)} : {})});
            if (!live.current) return;
            message.success("保存成功"); setEditor(undefined); reload();
        } catch (reason: any) { if (live.current) message.error(reason?.message); }
        finally { busyRef.current = false; if (live.current) setSaving(false); }
    };
    const remove = (row: any[], index: number) => {
        const keys = rowKeys[index];
        if (!keys) return;
        const confirmation = modal.confirm({afterClose: () => confirmations.current.delete(confirmation), title: "删除这条数据？", content: "将删除数据库中的这条记录，此操作不可撤销。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                try {
                    await api.delete(record, {schema, table: selected, keys: JSON.stringify(keys)});
                    if (live.current) { message.success("删除成功"); reload(); }
                } catch (reason: any) { if (live.current) message.error(reason?.message); throw reason; }
            }});
        confirmations.current.add(confirmation);
        // 启用 AntD Promise 模式：失败保留弹窗重试，避免再次抛出未处理的拒绝。
        void Promise.resolve(confirmation);
    };
    const runQuery = async (sqlOverride?: string) => {
        const query = (sqlOverride ?? sql.current).trim();
        if (!query || busyRef.current) return;
        busyRef.current = true; setQueryLoading(true); setQueryResult(undefined); setQueryExecution(undefined);
        try {
            const data = await api.execute(record, {query, limit: 500});
            if (live.current) {
                const operation = data?.columns?.length ? "query" : "execute";
                setQueryResult(data); setQueryExecution({time: Date.now(), operation, statement: query, success: true, elapsed_ms: data?.elapsed_ms});
                if (operation === "execute" || data?.affected > 0) { setCatalogVersion((value) => value + 1); reload(); }
            }
        } catch (reason: any) {
            if (live.current) {
                setQueryExecution({time: Date.now(), operation: "execute", statement: query, success: false, error: reason?.message || "SQL 执行失败，请重试"});
            }
        }
        finally { busyRef.current = false; if (live.current) setQueryLoading(false); }
    };

    const queryOutput = queryExecution?.error || (queryExecution?.success && queryResult && !queryResult?.columns?.length
        ? ["执行成功", `影响行数：${queryResult?.affected ?? 0}`, `执行耗时：${queryResult?.elapsed_ms ?? 0} ms`, "", queryExecution.statement].join("\n") : "");
    const structureDisabled = isView || metadataLoading || Boolean(metadataError) || Boolean(schemaBusy);

    const objectKinds = (engine?.objects || []).filter((item) => features[objectFeatures[item.key]] !== false).map((item) => ({...item,
        label: item.key === "routine" && features.procedures === false ? "函数" : item.key === "routine" && features.functions === false ? "存储过程"
            : item.label}));
    const selectedKind = section === "objects"
        ? objectKinds.find((item) => item.key === selectedObject) || objectKinds[0]
        : objectKinds.find((item) => item.key === tab);
    const advancedTabs = [...(!isView ? [{key: "indexes", label: "索引管理"}] : []),
        ...(!isView && features.foreign_keys !== false ? [{key: "foreign_key", label: "外键关系"}] : []), {key: "ddl", label: "SQL 定义"}];
    const objectTabs = [{key: "data", label: "数据"}, {key: "columns", label: "字段"},
        ...advancedTabs.map((item) => ({...item, label: item.key === "indexes" ? "索引" : item.key === "foreign_key" ? "外键" : "定义"}))];
    useEffect(() => {
        const changedTable = controlledSelection.current !== tableValue && tableValue !== reportedSelection.current;
        controlledSelection.current = tableValue;
        // 外部选择先应用到工具区，再通知目录快照，避免迟到目录结果覆盖新选择。
        if (changedTable && tableValue !== undefined && tableValue !== selected) {
            selectTable(tableValue); setTab("data"); return;
        }
        reportedSelection.current = selected;
        onTargetChange?.({schema, table: selected, isView, schemas, tables, loading: catalogLoading, error: catalogError});
    }, [schema, selected, isView, schemas, tables, catalogLoading, catalogError, onTargetChange, tableValue]);
    const dataToolbar = <div className="client-toolbar">
            {(!isView) && <Button size="small" color="default" variant="filled" autoInsertSpace={false} icon={<Icon type="PlusOutlined"/>} disabled={loading || metadataLoading || !columns.length || isView} onClick={() => edit()}>新增</Button>}
            <Button size="small" color={filters ? "primary" : "default"} variant="filled" icon={<Icon type="FilterOutlined"/>}
                disabled={metadataLoading || !columns.length || Boolean(metadataError)} aria-haspopup="dialog" onClick={() => {
                    setFilterFields(filters ? appliedFilters : [{column: columns[0]?.name || "", operator: "eq", value: ""}]);
                    setFilterLogic(filters ? appliedFilterLogic : "and");
                    setFilterOpen(true);
                }}>筛选{filters ? ` · ${appliedFilters.length}` : ""}</Button>
            <div className="client-toolbar-spacer"/>
        </div>;
    const columnToolbar = <div className="client-toolbar">
        {(!isView) && <Button size="small" color="default" variant="filled" autoInsertSpace={false} icon={<Icon type="PlusOutlined"/>} disabled={structureDisabled} onClick={() => setSchemaEditor({mode: "add_column"})}>新增</Button>}
        </div>;
    const indexToolbar = <div className="client-toolbar">
        <Button size="small" color="default" variant="filled" autoInsertSpace={false} icon={<Icon type="PlusOutlined"/>} disabled={structureDisabled || !columns.length} onClick={() => setSchemaEditor({mode: "create_index"})}>新增</Button>
        </div>;
    const rowSummary = `${result?.rows?.length || 0} 条${result?.total != null ? ` / 共 ${result?.total} 条` : ""}${result?.elapsed_ms != null ? ` · ${result?.elapsed_ms} ms` : ""}${result?.truncated ? " · 数据已截断" : ""}`;
    const dataPanel = <>
        <ResultTable result={result} loading={loading} height={320} autoHeight rowOffset={((result?.page || page) - 1) * (result?.page_size || pageSize)}
            columnMetadata={columns} actionWidth={84} actionAlign="center"
            sort={sort} order={order} onSortChange={(field: string, direction: string) => {
                rowRequest.current += 1; setPage(1);
                setSorting({target: rowTarget, field, order: direction});
            }}
            action={!rowKeys.some(Boolean) ? undefined : (row: any[], index: number) => {
                const keys = rowKeys[index];
                if (!keys) return null;
                return <ActionDropdown key={JSON.stringify(keys)} disabled={loading} menu={{items: [
                    {key: "edit", label: "编辑", onClick: () => edit(row, index)},
                    {key: "delete", label: "删除", danger: true, onClick: () => remove(row, index)},
                ]}}><Button color="default" variant="filled" size="small" autoInsertSpace={false} disabled={loading}>操作</Button></ActionDropdown>;
            }}/>
        <div className="client-footer"><Tooltip title={screens.md ? undefined : rowSummary} trigger={["hover", "click"]}>
            <Typography.Text type="secondary">{screens.md ? rowSummary
                : `${result?.total ?? result?.rows?.length ?? 0} 条${result?.elapsed_ms != null ? ` · ${result?.elapsed_ms}ms` : ""}${result?.truncated ? " · 已截断" : ""}`}</Typography.Text>
            </Tooltip>
            <ConfigProvider input={{variant: "filled"}}>
                <Pagination className="client-pagination" size="small" current={result?.page || page} pageSize={result?.page_size || pageSize} total={result?.total || 0}
                    disabled={loading || !result} showLessItems simple={!screens.md} showQuickJumper={Boolean(screens.md)} responsive
                    locale={screens.md ? undefined : {items_per_page: "/页"}}
                    showSizeChanger={{showSearch: false, variant: "filled", getPopupContainer: (trigger) => trigger.ownerDocument.body}}
                    pageSizeOptions={[10, 20, 50, 100, 200, 500, 1000]} onChange={(nextPage, nextSize) => {
                        setPage(nextSize === (result?.page_size || pageSize) ? nextPage : 1); setPageSize(nextSize); reload();
                    }}/>
            </ConfigProvider></div>
    </>;
    const columnPanel = <>
        <VirtualTable className="mysql-fill-table" rowKey="name" size="small" loading={metadataLoading} pagination={false} dataSource={columns}
            scroll={{x: 880}}
            locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
            columns={[{title: "字段", dataIndex: "name", width: 160, ellipsis: true}, {title: "类型", dataIndex: "type", width: 140},
                {title: "允许空值", dataIndex: "nullable", width: 84, render: (value) => value ? "是" : "否"},
                {title: "主键", dataIndex: "primary_key", width: 72, render: (value) => value ? "是" : "—"},
                {title: "默认值", dataIndex: "default", width: 160, ellipsis: true, render: (value) => value == null ? "—" : String(value)},
                {title: "备注", dataIndex: "comment", width: 180, ellipsis: true, render: (value: string) => value || "—"},
                ...(!isView ? [{title: "操作", key: "action", width: 84, align: "center" as const, fixed: "right" as const, render: (_: any, column: any) => <ActionDropdown disabled={structureDisabled} menu={{items: [
                    ...(!column.generated ? [{key: "edit", label: "编辑", onClick: () => setSchemaEditor({mode: "update_column", column})}] : []),
                    {key: "delete", label: "删除", danger: true, onClick: () => deleteObject("drop_column", column)},
                ]}}><Button color="default" variant="filled" size="small" autoInsertSpace={false} disabled={structureDisabled}>操作</Button></ActionDropdown>}] : [])]}/>
    </>;
    const indexPanel = <>
        <VirtualTable className="mysql-fill-table" rowKey="name" size="small" loading={metadataLoading} pagination={false} dataSource={indexes}
            scroll={{x: 740}}
            locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
            columns={[{title: "索引", dataIndex: "name", width: 160, ellipsis: true}, {title: "字段", dataIndex: "columns", render: (value) => value?.join(", ") || "—"},
                {title: "唯一", dataIndex: "unique", width: 72, render: (value) => value ? "是" : "否"}, {title: "主键", dataIndex: "primary", width: 72, render: (value) => value ? "是" : "—"},
                ...(features.index_comments !== false ? [{title: "备注", dataIndex: "comment", width: 180, ellipsis: true, render: (value: string) => value || "—"}] : []),
                ...(!isView ? [{title: "操作", key: "action", width: 84, align: "center" as const, fixed: "right" as const, render: (_: any, index: any) => <ActionDropdown disabled={structureDisabled} menu={{items: [
                    ...(features.index_comments !== false && index.comment_supported !== false ? [{key: "comment", label: "备注", onClick: () => setSchemaEditor({mode: "update_index", index})}] : []),
                    {key: "delete", label: "删除", danger: true, onClick: () => deleteObject("drop_index", index)},
                ]}}><Button color="default" variant="filled" size="small" autoInsertSpace={false} disabled={structureDisabled}>操作</Button></ActionDropdown>}] : [])]}/>
    </>;
    const refreshTable = <Button className="mysql-toolbar-refresh" size="small" color="default" variant="filled" loading={loading || metadataLoading} aria-label="刷新数据表" icon={<Icon type="ReloadOutlined"/>}
        onClick={() => { if (tab === "data") reload(); else setDetailVersion((value) => value + 1); }}/>;
    const schemaObjects = selectedKind && <SqlObjects key={JSON.stringify([record.id, record.database_name || record.name, schema, selectedKind.key])} record={record} schema={schema} table={selected} columns={columns} tables={tables} schemas={schemas}
        kind={selectedKind.key} onChange={reloadCatalog} reloadKey={`${reloadKey}:${catalogVersion}`}
        onBrowse={(name: string) => { selectTable(name); setTab("data"); onBrowse?.(); }}/>
    const emptyTable = <div className="client-empty">{catalogLoading ? <Spin/> : <>
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
        </>}</div>;
    const browseTable = selected ? <>
        <div className="sql-detail-navigation">
            <Tabs className="sql-detail-tabs" size="small" activeKey={objectTabs.some((item) => item.key === tab) ? tab : "data"} onChange={setTab} items={objectTabs}/>
            {!selectedKind && <div className="mysql-detail-actions">
                {tab === "columns" ? columnToolbar : tab === "indexes" ? indexToolbar : !["foreign_key", "ddl"].includes(tab) ? dataToolbar : null}
                {["foreign_key", "ddl"].includes(tab) && <div className="client-toolbar" ref={setDetailToolbar}/>}
                {refreshTable}
            </div>}
        </div>
        <div className="sql-detail-body">
            {tab === "columns" ? columnPanel : tab === "indexes" ? indexPanel
                : tab === "foreign_key" ? <SqlObjects key={`${schema}:${selected}:foreign_key`} record={record} schema={schema} table={selected} columns={columns} tables={tables} schemas={schemas} kind="foreign_key" tableOnly onChange={reloadCatalog} reloadKey={`${reloadKey}:${catalogVersion}:${detailVersion}`}
                    toolbarContainer={detailToolbar}/>
                    : tab === "ddl" ? <SqlDefinition record={record} schema={schema} table={selected} version={`${catalogVersion}:${reloadKey}:${detailVersion}`}
                        toolbarContainer={detailToolbar}/>
                        : dataPanel}
        </div>
    </> : emptyTable;

    return <div className={`${panelStyles.panel} sql-workbench`} style={hidden ? {display: "none"} : undefined}>
        <div className="sql-stage sql-stage-wide">
            <section className="client-content sql-content">
                {/* 导入导出保留任务状态；只释放隐藏功能区内的重表格和 SQL 编辑器。 */}
                <div className="mysql-browser-header" style={section !== "browse" ? {display: "none"} : undefined}>
                    <div className="mysql-table-toolbar">
                        <div className="mysql-browse-target">
                            <Select className="mysql-table-picker" aria-label="选择数据表或视图" size="small" variant="borderless" showSearch optionFilterProp="label"
                                prefix={<Icon type={isView ? "EyeOutlined" : "TableOutlined"}/>}
                                value={selected || undefined} placeholder="选择数据表或视图" loading={catalogLoading} disabled={catalogLoading || !tables.length}
                                options={tables.map((item) => ({value: item.name, label: `${item.name}${String(item.type).toUpperCase().includes("VIEW") ? "（视图）" : ""}`}))}
                                notFoundContent={<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="无匹配数据表"/>}
                                onChange={(name) => { selectTable(name); setTab("data"); }} onSelect={() => setTab("data")}/>
                        </div>
                        <div className="mysql-table-tools">
                            <div className="mysql-table-operations">
                                <Button color="default" variant="filled" size="small" aria-label="新建数据表" icon={<Icon type="PlusOutlined"/>}
                                    disabled={Boolean(schemaBusy)} onClick={() => setSchemaEditor({mode: "create_table", schema, table: ""})}>新建</Button>
                                <Button color="default" variant="filled" size="small" icon={<Icon type="EditOutlined"/>}
                                    disabled={Boolean(schemaBusy)} onClick={() => {
                                        if (!selected) { message.info("请先选择数据表或视图"); return; }
                                        setSchemaEditor({mode: "rename_table", schema, table: selected});
                                    }}>重命名</Button>
                                <SqlTransfer record={record} schema={schema} table={selected || undefined}
                                    disabled={Boolean(schemaBusy)} onChange={() => { setCatalogVersion((value) => value + 1); reload(); }}/>
                            </div>
                            <Button color="default" variant="filled" size="small" icon={<Icon type="ReloadOutlined"/>} aria-label="刷新数据表列表"
                                loading={catalogLoading} disabled={Boolean(schemaBusy)} onClick={() => setTableListVersion((value) => value + 1)}/>
                        </div>
                    </div>
                </div>
                {!hidden && (section === "objects" ? <>
                    <div className="sql-detail-navigation"><Tabs className="sql-detail-tabs" size="small" activeKey={selectedKind?.key}
                        onChange={setObject} items={objectKinds.map(({key, label}) => ({key, label}))}/></div>
                    <div className="sql-detail-body">{schemaObjects}</div>
                </> : section === "query" ? <Splitter className="client-query" orientation="vertical"
                    classNames={{dragger: {default: "client-query-resizer"}}} styles={{dragger: {default: {height: 16, top: 2}}}}>
                    <Splitter.Panel className="client-query-editor" defaultSize={compact ? 180 : 200} min={compact ? 180 : 200}>
                        <SqlQueryEditor initialValue={sql.current} onChange={changeSql} onRun={runQuery} loading={queryLoading} record={record} schema={schema} table={selected} columns={columns} tables={tables}
                            completionKey={`${reloadKey}:${catalogVersion}:${tableListVersion}:${detailVersion}`}
                            execution={queryExecution}/>
                    </Splitter.Panel>
                    <Splitter.Panel className="client-query-results" min={220}>
                        {(queryResult || queryExecution?.error) && <>
                            <div className="client-result-status"><Space size={10} wrap><Typography.Text>{queryExecution?.error ? "执行失败" : "执行结果"}</Typography.Text>
                                {Boolean(queryResult?.columns?.length) && <Typography.Text type="secondary">{`${queryResult?.rows?.length || 0} 条记录 · ${queryResult?.elapsed_ms || 0} ms`}</Typography.Text>}</Space>
                                {(Boolean(queryResult?.rows?.length) || queryOutput) && <Button className="client-result-actions" color="default" variant="filled" size="small" autoInsertSpace={false}
                                    icon={<Icon type="CopyOutlined"/>} onClick={async () => {
                                        // 按表格复制，保留重复列名，并转义单元格内的换行、制表符和引号。
                                        const cell = (value: any) => {
                                            const text = value == null ? "NULL" : typeof value === "object" ? JSON.stringify(value) : String(value);
                                            return /[\t\r\n"]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
                                        };
                                        const content = queryOutput || [Array.isArray(queryResult?.columns) ? queryResult?.columns : [], ...(Array.isArray(queryResult?.rows) ? queryResult?.rows : [])].map((row: any[]) => (Array.isArray(row) ? row : []).map(cell).join("\t")).join("\n");
                                        if (await copyTextToClipboard(content)) message.success("复制成功");
                                        else message.error("复制失败，请手动复制");
                                    }}>复制</Button>}
                            </div>
                            {queryOutput && <pre className={`client-query-output${queryExecution?.error ? " client-query-error" : ""}`}
                                role={queryExecution?.error ? "alert" : "status"}>{queryOutput}</pre>}
                            {Boolean(queryResult?.rows?.length) && <ResultTable result={queryResult} autoHeight/>}
                            {(queryResult?.truncated || queryResult?.has_more) && <Typography.Text type="secondary" className="client-hint">当前查询结果不完整，请使用表数据的完整导出。</Typography.Text>}
                        </>}
                    </Splitter.Panel>
                </Splitter>
                    : browseTable)}
            </section>
        </div>
        <ProModal title={<Title>筛选数据</Title>} width={560} okText="确认" onCancel={() => setFilterOpen(false)}
            onOk={() => {
                try {
                    setFilters(encodeSqlFilters(filterFields, columns)); setAppliedFilters(filterFields); setAppliedFilterLogic(filterLogic);
                    setPage(1); reload(); setFilterOpen(false);
                }
                catch (reason: any) { message.error(reason?.message); }
            }} modalProps={{
                open: filterOpen && !hidden && section === "browse" && tab === "data", className: styles.editor,
                cancelText: "取消", mask: {closable: true}, destroyOnHidden: true,
                styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}}, style: {maxWidth: "calc(100vw - 24px)"},
                okButtonProps: {disabled: loading || metadataLoading},
            }}>
            <SqlFilters columns={columns} value={filterFields} onChange={setFilterFields} disabled={loading || metadataLoading}
                logic={filterLogic} onLogicChange={setFilterLogic} resetDisabled={!(filterFields.length || filters)} onReset={() => {
                    setAppliedFilters([]); setAppliedFilterLogic("and"); setFilters(""); setPage(1); reload(); setFilterOpen(false);
                }}/>
        </ProModal>
        <ProModal title={<Title>{displayedEditor?.keys ? "编辑数据" : "添加数据"}</Title>}
            onCancel={() => { if (!saving) setEditor(undefined); }} onOk={form.submit} okText="保存" modalProps={{
                open: Boolean(editor), className: styles.editor, width: compact ? 480 : 540, confirmLoading: saving, cancelText: "取消",
                mask: {closable: !saving}, closable: !saving, keyboard: !saving, forceRender: true,
                afterClose: () => { setDisplayedEditor(undefined); form.resetFields(); },
                styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}}, style: {maxWidth: "calc(100vw - 24px)"},
            }}>
            <ConfigProvider theme={{token: {fontSize: compact ? 12 : 13, fontSizeSM: compact ? 11 : 12, controlHeight: compact ? 28 : 32}, components: {
                Input: {controlHeight: compact ? 28 : 32, inputFontSize: compact ? 12 : 13},
                InputNumber: {controlHeight: compact ? 28 : 32, inputFontSize: compact ? 12 : 13},
                DatePicker: {controlHeight: compact ? 28 : 32, inputFontSize: compact ? 12 : 13},
                Button: {controlHeight: compact ? 28 : 32, contentFontSize: compact ? 12 : 13},
                Select: {controlHeight: compact ? 28 : 32, controlHeightSM: compact ? 24 : 28, optionFontSize: compact ? 12 : 13},
                Form: {labelFontSize: compact ? 12 : 13},
            }}}>
            <Form form={form} layout="vertical" variant="filled" onFinish={save} disabled={saving}>
                {displayedEditor?.columns.map((column: any, index: number) => <Form.Item key={column.name} name={["fields", index]}
                    hidden={!displayedEditor.keys && column.generated}
                    label={<span title={[column.type, column.comment].filter(Boolean).join(" · ")} style={{overflowWrap: "anywhere"}}>{column.name}</span>}
                    rules={[{validator: async (_, field) => {
                        if (!field || field.mode === "default" || column.generated) return;
                        const original = displayedEditor.original?.[index];
                        if (original && field.mode === original.mode && field.value === original.value && field.kind === original.kind) return;
                        if (field.mode === "null") { if (!column.nullable) throw new Error("此字段不允许 NULL"); return; }
                        encodeSqlValue(field.kind, field.value);
                    }}]}>
                    <SqlField column={column} editing={Boolean(displayedEditor.keys)} disabled={saving}/>
                </Form.Item>)}
            </Form>
            </ConfigProvider>
        </ProModal>
        <SqlSchemaEditor record={record} action={schemaEditor} schema={schemaEditor?.schema ?? schema} table={schemaEditor?.table ?? selected} columns={columns} indexes={indexes} tables={tables} schemas={schemas}
            onCancel={() => setSchemaEditor(undefined)} onSuccess={reloadCatalog}/>
    </div>;
};
