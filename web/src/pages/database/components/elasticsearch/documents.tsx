import {useCallback, useEffect, useRef, useState} from "react";
import {App, Button, ConfigProvider, Empty, Grid, Input, Pagination, Select, Tooltip, Typography} from "antd";
import {Icon} from "sinking-antd";
import {useDatabaseClient} from "../shared/context";
import {VirtualTable} from "../shared/client-common";
import Document from "./document";
import Filter from "./filter";

export default ({record}: any) => {
    const {message, modal} = App.useApp();
    const {operate, ActionDropdown, features, version} = useDatabaseClient();
    const screens = Grid.useBreakpoint();
    const [result, setResult] = useState<any>();
    const [editor, setEditor] = useState<any>();
    const [filterOpen, setFilterOpen] = useState(false);
    const [keyword, setKeyword] = useState("");
    const [field, setField] = useState("");
    const [query, setQuery] = useState("{}");
    const [sort, setSort] = useState<{field: string; order: string}>({field: "", order: ""});
    const [mapping, setMapping] = useState<Record<string, any>>({});
    const [filterMapping, setFilterMapping] = useState<Record<string, any>>({});
    const [columnFields, setColumnFields] = useState<{scope: string; fields: string[]}>({scope: "", fields: []});
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(20);
    const [loading, setLoading] = useState(false);
    const [resultPage, setResultPage] = useState(1);
    const [searchExpired, setSearchExpired] = useState(false);
    const [refresh, setRefresh] = useState(0);
    const [retry, setRetry] = useState(0);
    const [deleting, setDeleting] = useState(false);
    const sequence = useRef(0);
    const live = useRef(true);
    const pending = useRef(false);
    const confirmation = useRef<any>(undefined);
    const appliedSearch = useRef({keyword: "", field: ""});
    const session = useRef<any>();
    const navigationPending = useRef(false);
    const closeSearch = useCallback((current: any, searchId = current?.searchId) => {
        if (searchId && ["pit", "scroll"].includes(current?.mode)) {
            void operate(current.record, "close_search", {pagination: current.mode, search_id: searchId}).catch(() => {});
        }
    }, [operate]);
    const rows = result?.documents || [];
    const scope = `${record.id}:${record.name || ""}`;
    const fields = columnFields.scope === scope ? columnFields.fields : [];
    useEffect(() => {
        live.current = true;
        return () => {
            live.current = false; sequence.current++; confirmation.current?.destroy();
            if (session.current) { session.current.closed = true; closeSearch(session.current); session.current = undefined; }
        };
    }, [closeSearch]);
    useEffect(() => {
        let active = true;
        setMapping({});
        setFilterMapping({});
        if (record.name) void operate(record, "mapping").then((data) => {
            if (!active) return;
            const value = data?.[record.name]?.mappings || data?.mappings || data || {};
            setFilterMapping({mappings: value});
            const types = value.properties ? [value] : Object.values(value);
            setMapping(Object.assign({}, ...types.map((type: any) => type?.properties || {})));
        }).catch(() => {});
        return () => { active = false; };
    }, [record.id, record.name, operate]);
    useEffect(() => {
        const request = ++sequence.current;
        const requestQuery = sort.field ? JSON.stringify({...JSON.parse(query), sort: [{[sort.field]: sort.order}]}) : query;
        const key = JSON.stringify([scope, requestQuery, pageSize, refresh]);
        if (session.current?.key !== key) {
            const previous = session.current;
            if (previous) { previous.closed = true; closeSearch(previous); }
            session.current = {key, scope, record, mode: "offset", searchId: "", tried: false, closed: false, loaded: false,
                cursors: new Map([[1, ""]]), cursorBytes: 0, pages: new Map(), bytes: 0, lastPage: 0};
            setSearchExpired(false);
            if (previous?.scope !== scope) { setResult(undefined); setResultPage(1); }
            if (page !== 1) { setPage(1); return; }
        }
        const current = session.current;
        if (!record.name) { setResult(undefined); setLoading(false); navigationPending.current = false; return; }
        setLoading(true);
        navigationPending.current = true;
        let redirected = false;
        void (async () => {
            let data = current.mode === "scroll" ? current.pages.get(page)?.data : undefined;
            if (!data) {
                if (current.mode === "pit" && !current.cursors.has(page)) throw new Error("该页游标已释放，请返回首页重新浏览");
                if (current.mode === "scroll" && page !== current.lastPage + 1) throw new Error("该页已不在缓存中，请返回首页重新浏览");
                data = await operate(record, "documents", {query: requestQuery, page, page_size: pageSize,
                    pagination: current.mode, search_id: current.searchId,
                    ...(current.mode === "pit" ? {cursor: current.cursors.get(page)} : {})});
                // 查询关闭或被新筛选替代时，迟到的响应也必须释放刚创建的会话。
                const accept = (response: any) => {
                    if (["pit", "scroll"].includes(response?.pagination) && response?.search_id) {
                        current.mode = response?.pagination;
                        current.searchId = response?.search_id;
                    }
                    if (!live.current || request !== sequence.current || current.closed) {
                        if (current.closed || session.current !== current) closeSearch(current);
                        return false;
                    }
                    return true;
                };
                if (!accept(data)) return;
                // 小列表保留页码跳转；超出结果窗口后才建立快照，旧服务端不会循环重试。
                if (features.deep_pagination && current.mode === "offset" && !current.tried && (Number(data?.total) > 10000 || data?.total_relation === "gte")) {
                    current.tried = true;
                    data = await operate(record, "documents", {query: requestQuery, page: 1, page_size: pageSize, pagination: "cursor"});
                    if (!accept(data)) return;
                    if (page !== 1) { redirected = true; setPage(1); }
                }
                const loadedPage = redirected ? 1 : page;
                if (current.mode === "pit") {
                    // 保留最近的游标，不保留文档副本，也不解析可能包含 64 位整数的 sort JSON。
                    if (data?.has_more && data?.next_cursor) {
                        const number = loadedPage + 1;
                        current.cursorBytes += (data?.next_cursor?.length - (current.cursors.get(number)?.length || 0)) * 2;
                        current.cursors.set(number, data?.next_cursor);
                    }
                    while (current.cursors.size > 1000 || current.cursorBytes > 8 * 1024 * 1024) {
                        const oldest = [...current.cursors.keys()].find((value) => value !== 1);
                        if (oldest === undefined) break;
                        current.cursorBytes -= current.cursors.get(oldest).length * 2;
                        current.cursors.delete(oldest);
                    }
                } else if (current.mode === "scroll") {
                    current.lastPage = loadedPage;
                    const bytes = (JSON.stringify(data)?.length || 0) * 2;
                    while (current.pages.size && (current.pages.size >= 3 || current.bytes + bytes > 8 * 1024 * 1024)) {
                        const oldest = current.pages.keys().next().value;
                        current.bytes -= current.pages.get(oldest).bytes;
                        current.pages.delete(oldest);
                    }
                    if (bytes <= 8 * 1024 * 1024) { current.pages.set(loadedPage, {data, bytes}); current.bytes += bytes; }
                }
            }
            if (!live.current || request !== sequence.current || current.closed) return;
            // 排序和翻页只更新文档，不重排已有列；同一索引中新发现的字段追加在末尾。
            setColumnFields((value) => {
                const names = new Set(value.scope === scope ? value.fields : []);
                for (const item of Array.isArray(data?.documents) ? data?.documents : []) {
                    for (const name of Object.keys(item?._source || {})) names.add(name);
                }
                return value.scope === scope && names.size === value.fields.length ? value : {scope, fields: [...names]};
            });
            current.loaded = true;
            setResult({...data, documents: Array.isArray(data?.documents) ? data?.documents.filter(Boolean) : []});
            setResultPage(redirected ? 1 : page);
            if (data?.timed_out) message.warning("查询达到时间限制，当前显示已返回的结果");
        })().catch((reason) => {
            if (!live.current || request !== sequence.current || current.closed) return;
            if (!current.loaded) { setResult(undefined); setResultPage(1); }
            if (current.mode !== "offset" || current.tried) {
                // Scroll 请求可能已在服务端前进，失败时不能重试同一页而跳过数据。
                setSearchExpired(true);
                closeSearch(current);
                current.searchId = "";
                message.error(`${reason?.message || "读取文档失败"}，请刷新或返回首页重新加载`);
            } else message.error(reason?.message || "读取文档失败");
        }).finally(() => {
            if (live.current && request === sequence.current && !redirected) { setLoading(false); navigationPending.current = false; }
        });
    }, [record.id, record.name, operate, query, page, pageSize, refresh, sort, closeSearch, features.deep_pagination, retry]);
    const navigate = (next: number) => {
        if (navigationPending.current || searchExpired || next === resultPage) return;
        navigationPending.current = true;
        setLoading(true);
        if (next === page) setRetry((value) => value + 1);
        else setPage(next);
    };
    const reload = () => { setPage(1); setRefresh((value) => value + 1); };
    const search = (value = keyword, selectedField = field) => {
        const content = value.trim();
        // 输入未变时失焦不重置筛选；空输入切换字段也不触发查询。
        if (content === appliedSearch.current.keyword && (!content || selectedField === appliedSearch.current.field)) return;
        appliedSearch.current = {keyword: content, field: selectedField};
        const data = content ? {query: selectedField ? {match: {[selectedField]: content}} : {query_string: {query: content}}} : {};
        const next = JSON.stringify(data);
        if (next !== query) { setQuery(next); setPage(1); }
    };
    const remove = (item: any) => {
        if (pending.current) return;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: "删除文档？", content: `文档“${item._id}”将永久删除。`, okText: "删除", cancelText: "取消", okButtonProps: {danger: true}, onOk: async () => {
            if (pending.current || !live.current) return;
            pending.current = true; setDeleting(true);
            try {
                await operate(record, "delete", {document_id: item._id, document_type: item._type, routing: item._routing ?? item.fields?._routing?.[0] ?? ""});
                if (live.current) { message.success("删除成功"); reload(); }
            } catch (reason: any) { if (live.current) message.error(reason?.message || "删除失败"); }
            finally { pending.current = false; if (live.current) setDeleting(false); }
        }});
    };
    const sortableField = (name: string) => {
        const field = mapping[name];
        if (!field || field.doc_values === false || ["object", "nested", "binary"].includes(field.type)) return "";
        if (["text", "string"].includes(field.type)) return field.fields?.keyword ? `${name}.keyword` : field.index === "not_analyzed" ? name : "";
        return name;
    };
    if (!record.name) return <div className="manager-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择索引"/></div>;
    const cursorPaging = ["pit", "scroll"].includes(result?.pagination);
    const previousAvailable = resultPage > 1 && (result?.pagination === "pit" ? session.current?.cursors.has(resultPage - 1) : session.current?.pages.has(resultPage - 1));
    const nextAvailable = result?.has_more && (result?.pagination !== "scroll" || resultPage === session.current?.lastPage || session.current?.pages.has(resultPage + 1));
    const total = Math.min(Number(result?.total) || 0, 10000);
    const stats = `${result?.total ?? 0}${result?.total_relation === "gte" ? "+" : ""} 条${result?.elapsed_ms !== undefined ? ` · ${result?.elapsed_ms} ms` : ""}`;
    return <div className="manager-stage">
        <div className="manager-toolbar es-document-toolbar">
            <div className="es-search"><Select variant="borderless" styles={{root: {outline: "none", boxShadow: "none", border: 0}}}
                value={field} aria-label="搜索字段" placeholder="请选择搜索字段" options={[{value: "", label: "全部字段"}, ...[...new Set([...Object.keys(mapping), ...fields])].map((name) => ({value: name, label: name}))]}
                popupMatchSelectWidth={false} labelRender={({label}) => <span className="es-search-label">{label}</span>}
                onChange={(value) => { setField(value); if (keyword.trim()) search(keyword, value); }}/>
                <Input variant="borderless" styles={{root: {outline: "none", boxShadow: "none"}, input: {outline: "none", boxShadow: "none"}}}
                    placeholder="搜索文档内容" allowClear value={keyword} onChange={(event) => { setKeyword(event.target.value); if (!event.target.value) search(""); }} onPressEnter={() => search()} onBlur={() => search()}/>
            </div>
            <div className="manager-tools">
                <Button color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} disabled={deleting} onClick={() => setEditor({mode: "create"})}>新增</Button>
                <Button color="default" variant="filled" icon={<Icon type="FilterOutlined"/>} onClick={() => { setFilterOpen(true); }}>筛选</Button>
                <Button className="manager-refresh" color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} aria-label="刷新文档" loading={loading} disabled={deleting} onClick={reload}/>
            </div>
        </div>
        <VirtualTable className="manager-table" size="small" rowKey={(item: any) => JSON.stringify([item._index, item._type || "", item._id, item._routing ?? item.fields?._routing?.[0] ?? ""])} dataSource={rows} loading={loading} scroll={{x: Math.max(640, fields.length * 160 + 240)}} pagination={false}
            locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}} showSorterTooltip={{target: "sorter-icon"}}
            onChange={(_: any, __: any, value: any, extra: any) => {
                if (extra.action !== "sort") return;
                const next = Array.isArray(value) ? value[0] : value;
                setSort({field: next.order ? sortableField(String(next.field)) : "", order: next.order === "ascend" ? "asc" : "desc"}); setPage(1);
            }}
            columns={[
                {title: "文档 ID", dataIndex: "_id", width: 160, ellipsis: true, render: (value: string, item: any) => <button className="es-document-key" onClick={() => setEditor({mode: "view", item})} title={value}>{value}</button>},
                ...(features.typed_documents ? [{title: "文档类型", dataIndex: "_type", width: 120, ellipsis: true}] : []),
                ...fields.map((name) => ({title: <Typography.Text ellipsis={{tooltip: name}}>{name}</Typography.Text>, key: name, dataIndex: name, width: 160, ellipsis: true,
                    sorter: Boolean(sortableField(name)), sortOrder: sort.field && sort.field === sortableField(name) ? sort.order === "asc" ? "ascend" : "descend" : null,
                    render: (_: any, item: any) => { const value = item._source?.[name]; const text = value === undefined ? "—" : value === null ? "NULL" : typeof value === "object" ? JSON.stringify(value) : String(value);
                        return <Tooltip title={text.length < 3000 ? text : undefined}><span className={`manager-cell${value == null ? " manager-cell-null" : ""}`}>{text}</span></Tooltip>; }})),
                {title: "操作", key: "action", width: 76, align: "center", fixed: "right", render: (_: any, item: any) => deleting ? null : <ActionDropdown menu={{items: [
                    {key: "view", label: "详情", onClick: () => setEditor({mode: "view", item})},
                    {key: "edit", label: "编辑", onClick: () => setEditor({mode: "edit", item})},
                    {key: "delete", label: "删除", danger: true, onClick: () => remove(item)},
                ]}}><Button color="default" variant="filled" size="small">操作</Button></ActionDropdown>},
            ]}/>
        <div className="manager-footer"><Tooltip title={!cursorPaging && Number(result?.total) > 10000 ? `${stats}，可浏览前 10000 条，请筛选缩小范围` : stats}><Typography.Text>{stats}</Typography.Text></Tooltip>
            <ConfigProvider select={{variant: "filled"}}>{cursorPaging ? <div className="es-cursor-pagination">
                <Button size="small" type="text" icon={<Icon type="DoubleLeftOutlined"/>} aria-label="返回首页" disabled={loading || resultPage === 1 && !searchExpired} onClick={reload}/>
                <Button size="small" type="text" icon={<Icon type="LeftOutlined"/>} aria-label="上一页" disabled={loading || searchExpired || !previousAvailable} onClick={() => navigate(resultPage - 1)}/>
                <span>第 {resultPage} 页</span>
                <Button size="small" type="text" icon={<Icon type="RightOutlined"/>} aria-label="下一页" disabled={loading || searchExpired || !nextAvailable} onClick={() => navigate(resultPage + 1)}/>
                {screens.md && <Select size="small" value={pageSize} aria-label="每页条数" disabled={loading} options={[20,50,100,500].map((value) => ({value, label: `${value} 条/页`}))}
                    onChange={(value) => { setPage(1); setPageSize(value); }}/>}
            </div> : <Pagination size="small" current={resultPage} pageSize={pageSize} total={total} disabled={loading} showSizeChanger={Boolean(screens.md)} showQuickJumper={Boolean(screens.md)} simple={!screens.md}
                pageSizeOptions={[20,50,100,500]} locale={!screens.md ? {items_per_page: "/页"} : undefined}
                onChange={(next, size) => { if (size !== pageSize) { setPage(1); setPageSize(size); } else navigate(next); }}/>}</ConfigProvider>
        </div>
        <Filter open={filterOpen} query={query} mapping={filterMapping} fields={fields} version={version} onCancel={() => setFilterOpen(false)} onApply={(value) => {
            setQuery(value); setSort({field: "", order: ""}); setKeyword("");
            appliedSearch.current = {keyword: "", field};
            setPage(1); setFilterOpen(false); setRefresh((value) => value + 1);
        }}/>
        <Document record={record} editor={editor} onCancel={() => setEditor(undefined)} onSuccess={reload}/>
    </div>;
};
