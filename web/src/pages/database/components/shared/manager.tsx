import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, Drawer, Dropdown, Empty, Grid, Input, Modal, Space, Spin, Tooltip, Typography} from "antd";
import {Icon, useTheme} from "sinking-antd";
import type {ComponentType} from "react";
import DatabasePanel from "./panel";
import DatabaseForm from "./database-form";
import Detail from "./database-detail";
import useStyles from "./manager-styles";
import useDatabaseStyles from "./styles/client";
import type {DatabaseManagerProps, DatabaseVersion} from "./entities";

interface ManagerProps extends DatabaseManagerProps {
    type: number;
    panel: ComponentType<any>;
    editForm?: ComponentType<any>;
    engine?: any;
}

const namesByType: Record<number, string> = {3: "Postgres", 4: "MongoDB", 5: "SQL Server", 6: "SQLite", 7: "Elasticsearch"};
const systemDatabases: Record<number, string[]> = {3: ["postgres", "template0", "template1"], 4: ["admin", "local", "config"], 5: ["master", "model", "msdb", "tempdb"]};
const DefaultActionDropdown = (props: any) => <Dropdown trigger={["click"]} {...props}/>;

const Workspace = ({record, requests, components, type, panel: Panel, editForm: EditForm, engine, open = true, onCancel, onConfigure, refreshKey = 0, initialTab, className, style}: ManagerProps) => {
    const ActionDropdown = components?.ActionDropdown || DefaultActionDropdown;
    const [version, setVersion] = useState<DatabaseVersion>();
    const engineName = namesByType[type] || "数据库";
    const objectLabel = type === 7 ? "索引" : "数据库";
    const mutable = [3, 4, 5, 7].includes(type);
    const hasCatalog = type !== 6;
    // 宿主重新渲染时可传入新的映射对象；仅连接标识或 refreshKey 改变才重新连接。
    const requestMethods = useRef(requests);
    requestMethods.current = requests;
    const request = useCallback(async (operation: string, body: Record<string, any>) => {
        const execute = requestMethods.current[operation];
        if (!execute) throw new Error("当前数据库不支持此操作");
        const response = await execute({body});
        if (!response) throw new Error("数据库请求失败");
        return response;
    }, []);
    const {message, modal} = App.useApp();
    const {styles: databaseStyles} = useDatabaseStyles();
    const screens = Grid.useBreakpoint();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const [names, setNames] = useState<string[]>([]);
    const [loadedKey, setLoadedKey] = useState("");
    const [selected, setSelected] = useState("");
    const [search, setSearch] = useState("");
    const [loading, setLoading] = useState(false);
    const [loadingMore, setLoadingMore] = useState(false);
    const [deleting, setDeleting] = useState(false);
    const [showCatalog, setShowCatalog] = useState(false);
    const [showSystem, setShowSystem] = useState(false);
    const [order, setOrder] = useState("asc");
    const [fullscreen, setFullscreen] = useState(false);
    const [entered, setEntered] = useState(false);
    const [editor, setEditor] = useState<any>();
    const [detail, setDetail] = useState<any>();
    const [revision, setRevision] = useState(0);
    const sequence = useRef(0);
    const catalogList = useRef<HTMLDivElement>(null);
    const catalogKeyword = useRef("");
    const catalogSystem = useRef(false);
    const catalogOrder = useRef("asc");
    const catalogPending = useRef(false);
    const catalogFailed = useRef(false);
    const catalogPage = useRef({page: 0, pageSize: 20, total: 0});
    const searchTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const connectedKey = useRef("");
    const currentTab = useRef<{database: string; tab: string} | undefined>(undefined);
    const live = useRef(true);
    const pendingDelete = useRef(false);
    const confirmation = useRef<any>(undefined);
    const lastRecord = useRef(record);
    if (record) lastRecord.current = record;
    const instance = record || lastRecord.current;
    const recordKey = record ? String(record.id) : "";
    const connectionKey = JSON.stringify([instance ? String(instance.id) : "", refreshKey]);
    const openedKey = useRef("");
    const instanceKey = useRef(recordKey);
    instanceKey.current = recordKey;
    const protectedName = (name: string) => type === 7 ? name.startsWith(".") : systemDatabases[type]?.includes(name.toLowerCase());
    const target = useMemo(() => (selected || type === 7) && instance && loadedKey === connectionKey
        ? {...instance, type, name: selected, database_name: selected, server_name: instance.name} : undefined,
    [instance, type, selected, loadedKey, connectionKey]);
    const changeTab = useCallback((tab: string) => { currentTab.current = {database: selected, tab}; }, [selected]);
    const load = useCallback(async (preferred?: string, page = 1, resetSelection = false) => {
        if (!record?.id || page > 1 && catalogPending.current) return;
        if (page === 1) { clearTimeout(searchTimer.current); searchTimer.current = undefined; }
        const requestId = ++sequence.current;
        const keyword = catalogKeyword.current;
        const connect = connectedKey.current !== connectionKey;
        catalogPending.current = true;
        setLoading(page === 1); setLoadingMore(page > 1);
        if (page === 1 && catalogList.current) catalogList.current.scrollTop = 0;
        try {
            // 目录和版本一并加载，切换数据库时复用版本信息。
            const [response, versionResponse] = await Promise.all([
                request(type === 7 ? "indices" : "databases", {id: record.id, keyword, page, page_size: catalogPage.current.pageSize, order_by_type: catalogOrder.current,
                    ...([3, 4, 5].includes(type) ? {show_system: catalogSystem.current} : {})}),
                connect ? request("version", {id: record.id}) : undefined,
            ]);
            if (requestId !== sequence.current) return;
            if (response.code !== 200) throw new Error(response.message || "获取目录失败");
            const data = response.data;
            const versionData = versionResponse?.data;
            if (connect && versionResponse?.code !== 200) throw new Error(versionResponse?.message || "获取版本失败");
            if (!Array.isArray(data?.list)) throw new Error("数据库列表数据无效，请刷新后重试");
            const values = (type === 7 ? data.list.map((item: any) => item.name) : data.list).filter((name: any) => typeof name === "string");
            if (!Number.isInteger(data.page) || data.page < 1 || data.page > page || !Number.isInteger(data.page_size) || data.page_size < 1
                || !Number.isSafeInteger(data.total) || data.total < 0 || values.length !== data.list.length) throw new Error("数据库列表分页数据无效，请重试");
            catalogPage.current = {page: data.page, pageSize: data.page_size, total: data.total};
            if (connect) {
                if (!versionData || typeof versionData.version !== "string" || !versionData.version.trim() || !versionData.features || typeof versionData.features !== "object" || Array.isArray(versionData.features)) {
                    throw new Error("数据库版本信息不完整，请重新连接");
                }
                setVersion(versionData);
                connectedKey.current = connectionKey;
            }
            catalogFailed.current = false;
            setNames((current) => page === 1 ? values : [...new Set([...current, ...values])]); setLoadedKey(connectionKey);
            // 搜索或刷新第一页不切换已打开的数据库，后续页也不改变右侧工作区。
            if (page === 1) setSelected((current) => {
                const next = resetSelection ? "" : preferred || current;
                if (next && (keyword || data.total > values.length || values.includes(next))) return next;
                return keyword ? "" : values[0] || "";
            });
        } catch (reason: any) {
            if (requestId !== sequence.current) return;
            catalogFailed.current = true;
            if (connect) onCancel();
            message.error(reason.message || "获取数据库列表失败");
        } finally {
            if (requestId === sequence.current) { catalogPending.current = false; setLoading(false); setLoadingMore(false); }
        }
    }, [record?.id, connectionKey, request, type]);
    useEffect(() => {
        if (!record) return;
        live.current = true;
        const sameInstance = openedKey.current === recordKey;
        openedKey.current = recordKey;
        connectedKey.current = ""; catalogPending.current = false; catalogFailed.current = false; catalogPage.current = {page: 0, pageSize: 20, total: 0};
        clearTimeout(searchTimer.current); searchTimer.current = undefined; setLoadingMore(false);
        setNames([]); setLoadedKey(""); setVersion(undefined); setEditor(undefined); setDetail(undefined); setShowCatalog(false); setRevision(0); setDeleting(false);
        if (!sameInstance) { setSelected(""); setSearch(""); catalogKeyword.current = ""; catalogSystem.current = false; setShowSystem(false); catalogOrder.current = "asc"; setOrder("asc"); currentTab.current = undefined; setFullscreen(false); }
        confirmation.current?.destroy(); pendingDelete.current = false;
        void load();
        return () => { live.current = false; sequence.current++; clearTimeout(searchTimer.current); confirmation.current?.destroy(); };
    }, [recordKey, load]);
    const clear = () => {
        // 配置连接只临时隐藏；真正关闭后才清理内容和版本，避免退出动画变形。
        if (instanceKey.current) return;
        lastRecord.current = undefined; openedKey.current = ""; connectedKey.current = ""; currentTab.current = undefined;
        setNames([]); setSelected(""); setLoadedKey(""); setSearch(""); setVersion(undefined);
        setLoading(false); setLoadingMore(false);
        setEditor(undefined); setDetail(undefined); setShowCatalog(false); setFullscreen(false); setEntered(false);
        catalogSystem.current = false; setShowSystem(false);
        catalogOrder.current = "asc"; setOrder("asc");
    };
    const loadNext = useCallback(() => {
        const {page, pageSize, total} = catalogPage.current;
        if (!catalogPending.current && !deleting && page > 0 && page * pageSize < total) void load(undefined, page + 1);
    }, [deleting, load]);
    useEffect(() => {
        const element = catalogList.current;
        if (!open || !element) return;
        // 补满可视区；失败后等待主动滚动或刷新，避免观察器持续重试。
        const fill = () => { if (!catalogFailed.current && element.clientHeight > 0 && element.scrollHeight <= element.clientHeight + 1) loadNext(); };
        const observer = new ResizeObserver(fill);
        observer.observe(element); fill();
        return () => observer.disconnect();
    }, [open, entered, names, loading, loadingMore, showCatalog, screens.md, loadNext]);
    const searchCatalog = (value: string) => {
        setSearch(value);
        const keyword = value.trim();
        if (keyword === catalogKeyword.current) return;
        catalogKeyword.current = keyword; sequence.current++; catalogPending.current = false;
        catalogPage.current = {page: 0, pageSize: 20, total: 0};
        setNames([]); setLoading(true); setLoadingMore(false);
        clearTimeout(searchTimer.current);
        if (keyword) searchTimer.current = setTimeout(() => { searchTimer.current = undefined; void load(); }, 250);
        else { searchTimer.current = undefined; void load(); }
    };
    const submitSearch = () => { if (searchTimer.current) { clearTimeout(searchTimer.current); searchTimer.current = undefined; void load(); } };
    const sortCatalog = (value: string) => {
        if (value === catalogOrder.current) return;
        catalogOrder.current = value; setOrder(value);
        catalogPage.current = {page: 0, pageSize: 20, total: 0};
        setNames([]); void load();
    };
    const toggleSystem = () => {
        catalogSystem.current = !catalogSystem.current; setShowSystem(catalogSystem.current);
        sequence.current++; catalogPending.current = false;
        clearTimeout(searchTimer.current); searchTimer.current = undefined;
        catalogKeyword.current = ""; catalogPage.current = {page: 0, pageSize: 20, total: 0};
        setSearch(""); setNames([]); setSelected(""); setLoadingMore(false); currentTab.current = undefined;
        void load(undefined, 1, true);
    };
    const choose = (name: string) => { setSelected(name); setShowCatalog(false); setDetail(undefined); };
    const remove = (name: string) => {
        if (!record || protectedName(name) || pendingDelete.current) return;
        const key = recordKey;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: `删除${objectLabel}“${name}”？`, content: `将删除${objectLabel}及全部数据，此操作不可撤销。`, okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                if (pendingDelete.current || !live.current || instanceKey.current !== key) return;
                pendingDelete.current = true; setDeleting(true);
                try {
                    const response = await request(type === 7 ? "drop_index" : "drop_database", {id: record.id, names: [name]});
                    if (response.code !== 200) throw new Error(response.message || "删除失败");
                    if (!live.current || instanceKey.current !== key) return;
                    message.success("删除成功"); setNames((values) => values.filter((value) => value !== name));
                    setSelected((value) => value === name ? "" : value); setDetail(undefined); await load();
                } catch (reason: any) { if (live.current && instanceKey.current === key) message.error(reason.message || "删除数据库失败"); }
                finally { if (live.current && instanceKey.current === key) { pendingDelete.current = false; setDeleting(false); } }
            }});
    };
    const mobile = screens.md === undefined ? typeof window !== "undefined" && window.matchMedia("(max-width: 767px)").matches : !screens.md;
    const closeCatalog = () => setShowCatalog(false);
    const connecting = loadedKey !== connectionKey;
    const catalog = <aside className="database-catalog" aria-label={`${objectLabel}目录`}>
        <div className="database-catalog-toolbar">
            <Input className="database-catalog-search" variant="filled" prefix={<Icon type="SearchOutlined"/>} placeholder={`搜索${objectLabel}`} aria-label={`搜索${objectLabel}`} value={search}
                onChange={(event) => searchCatalog(event.target.value)} onPressEnter={submitSearch} onBlur={submitSearch} allowClear/>
            <div className="database-catalog-actions">
                <Space size={4}>
                    {mutable && <Button color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} aria-label={`创建${objectLabel}`} disabled={deleting}
                        onClick={() => { closeCatalog(); setEditor({}); }}/>}
                    {!deleting && [3, 4, 5].includes(type) && <ActionDropdown placement="bottomLeft" arrow={false} menu={{items: [
                        {key: "system", label: "显示系统库", icon: <Icon type={showSystem ? "CheckSquareOutlined" : "BorderOutlined"}/>, onClick: toggleSystem},
                    ]}}><Button color={showSystem ? "primary" : "default"} variant="filled" icon={<Icon type="FilterOutlined"/>} aria-label="数据库目录筛选"/></ActionDropdown>}
                    {!deleting && <ActionDropdown placement="bottomLeft" arrow={false} menu={{selectable: true, selectedKeys: [order], onClick: ({key}) => sortCatalog(key), items: [
                        {key: "asc", label: "名称升序", icon: <Icon type="SortAscendingOutlined"/>},
                        {key: "desc", label: "名称降序", icon: <Icon type="SortDescendingOutlined"/>},
                    ]}}><Button color="default" variant="filled" icon={<Icon type={order === "asc" ? "SortAscendingOutlined" : "SortDescendingOutlined"}/>}
                        aria-label={`${objectLabel}排序`}/></ActionDropdown>}
                </Space>
                <Button color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} aria-label="刷新数据库列表" loading={loading} disabled={deleting} onClick={() => void load()}/>
            </div>
        </div>
        <div ref={catalogList} className="database-catalog-list" role="list" onScroll={(event) => {
            const {scrollHeight, scrollTop, clientHeight} = event.currentTarget;
            if (scrollHeight - scrollTop - clientHeight <= 48) loadNext();
        }}>
            {loading && !names.length ? <div className="database-catalog-state"><Spin/></div> : <>
                {names.map((name) => <div className={`database-catalog-item${selected === name ? " active" : ""}`} key={name} role="listitem">
                    <button className="database-catalog-select" type="button" title={name} aria-current={selected === name ? "true" : undefined} disabled={deleting} onClick={() => choose(name)}>
                        <Icon type="DatabaseOutlined"/><span>{name}</span>
                    </button>
                    {!deleting && <ActionDropdown menu={{items: [
                        {key: "info", label: "详情", onClick: () => { closeCatalog(); setDetail({...instance, name, database_name: name, server_name: instance?.name}); }},
                        ...(([3, 5].includes(type) || EditForm) && !protectedName(name) ? [{key: "edit", label: "修改", onClick: () => { closeCatalog(); setEditor({name}); }}] : []),
                        ...(mutable && !protectedName(name) ? [{key: "delete", label: "删除", danger: true, onClick: () => { closeCatalog(); remove(name); }}] : []),
                    ]}}><Button type="text" size="small" icon={<Icon type="MoreOutlined"/>} aria-label={`数据库操作 ${name}`}/></ActionDropdown>}
                </div>)}
                {!names.length && <div className="database-catalog-state"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
                </div>}
                {loadingMore && <div className="database-catalog-more" role="status"><Spin size="small"/></div>}
            </>}
        </div>
        {version && <footer className={styles.clientInstance}>
            <Typography.Text className="client-instance-name" ellipsis={{tooltip: instance?.name}}>{instance?.name}</Typography.Text>
            <Typography.Text className="client-instance-version" ellipsis={{tooltip: `${version.name || engineName} ${version.version}`}}>{version.name || engineName} {version.version}</Typography.Text>
        </footer>}
    </aside>;
    return <>
        <Modal open={Boolean(record) && open} onCancel={onCancel} afterClose={clear} footer={null} destroyOnHidden={!record} closable={false}
            afterOpenChange={(visible) => { if (visible) setEntered(true); }}
            title={`${engineName} 管理`} width={fullscreen ? "100vw" : compact ? 1120 : 1200}
            className={[styles.workspace, styles.clientWorkspace, fullscreen && "is-fullscreen", className].filter(Boolean).join(" ")}
            style={{top: fullscreen ? 0 : mobile ? 24 : 100, paddingBottom: fullscreen ? 0 : mobile ? 24 : 100,
                maxWidth: fullscreen ? "100vw" : "calc(100vw - 24px)", ...style}}
            focusable={{focusTriggerAfterClose: false}} mask={{closable: false}}>
            <div className="client-window">
                <header className="client-window-header">
                    <div className="client-window-title">
                        {hasCatalog && mobile && <Tooltip title="数据库目录"><Button className="client-catalog-toggle" color="default" variant={showCatalog ? "filled" : "text"}
                            icon={<Icon type="MenuOutlined"/>} aria-label={`${objectLabel}目录`} aria-expanded={showCatalog} disabled={connecting}
                            onClick={() => setShowCatalog((value) => !value)}/></Tooltip>}
                        <span className={databaseStyles.databaseIcon} data-type={type} aria-hidden="true"/>
                        <span className="client-window-label">{engineName} 管理</span>
                    </div>
                    <div className="client-window-actions">
                        {onConfigure && <Tooltip title="连接设置"><Button type="text" icon={<Icon type="SettingOutlined"/>} aria-label="连接设置" onClick={onConfigure}/></Tooltip>}
                        {!mobile && <Tooltip title={fullscreen ? "退出全屏" : "全屏"}><Button type="text" icon={<Icon type={fullscreen ? "FullscreenExitOutlined" : "FullscreenOutlined"}/>}
                            aria-label={fullscreen ? "退出全屏" : "全屏"} onClick={() => setFullscreen((value) => !value)}/></Tooltip>}
                        <Tooltip title="关闭"><Button type="text" icon={<Icon type="CloseOutlined"/>} aria-label={`关闭 ${engineName} 管理`} onClick={onCancel}/></Tooltip>
                    </div>
                </header>
                {/* 请求立即执行，入场动画结束后再挂载工作区，避免快速响应在缩放中途引发重排。 */}
                {connecting || !entered ? <div className="client-connection-state" role="status" aria-label="正在连接数据库"><Spin/></div> : <div className={`database-browser${hasCatalog ? "" : " is-single-database"}`}>
                    {hasCatalog && !mobile && catalog}
                    <section className="database-browser-content">
                        {target ? <DatabasePanel type={type} panel={Panel} engine={engine} request={request} version={version} components={components} key={`${instance!.id}:${selected}:${revision}`} record={target}
                            initialTab={currentTab.current?.database === selected ? currentTab.current.tab : initialTab || instance!.tab}
                            onTabChange={changeTab}/>
                            : <div className="database-browser-empty">{loading ? <Spin/> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择数据库"/>}</div>}
                    </section>
                    {hasCatalog && mobile && <Drawer closable={false} placement="left" size="min(300px, calc(100% - 40px))" open={Boolean(record) && open && showCatalog}
                        getContainer={false} push={false} rootStyle={{position: "absolute"}} rootClassName={styles.clientCatalogDrawer}
                        onClose={closeCatalog} destroyOnHidden>{catalog}</Drawer>}
                </div>}
            </div>
        </Modal>
        <DatabaseForm request={request} type={type} open={Boolean(record) && Boolean(editor) && !(EditForm && editor?.name !== undefined)} instance={instance} name={EditForm ? undefined : editor?.name} onCancel={() => setEditor(undefined)} onSuccess={(change: any) => {
            setSearch(""); catalogKeyword.current = ""; setRevision((value) => value + 1); setShowCatalog(false); void load(change.name);
        }}/>
        {EditForm && <EditForm request={request} instance={instance} version={version} open={Boolean(record) && editor?.name !== undefined}
            name={editor?.name} onCancel={() => setEditor(undefined)}/>}
        <Detail request={request} engineName={engineName} type={type} record={record ? detail : undefined} onCancel={() => setDetail(undefined)}/>
    </>;
};

export default (props: ManagerProps) => {
    const screens = Grid.useBreakpoint();
    return <ConfigProvider modal={{style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24}}}><App>
        <Workspace {...props}/>
    </App></ConfigProvider>;
};
