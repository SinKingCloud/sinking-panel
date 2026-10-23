import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Drawer, Empty, Grid, Input, Modal, Space, Spin, Tooltip, Typography} from "antd";
import {Icon, useTheme} from "sinking-antd";
import Content from "./content";
import DatabaseForm from "./database-form";
import Detail from "./database-detail";
import useStyles from "./workspace-styles";
import useDatabaseStyles from "./component-styles/client";
import {useMysqlClient} from "./context";
import type {DatabaseVersion, MysqlManagerProps} from "./entities";

export default ({record, open = true, onCancel, onConfigure, refreshKey = 0, initialTab, className, style, version, setVersion}: Omit<MysqlManagerProps, "requests" | "components"> & {version?: DatabaseVersion; setVersion: (version?: DatabaseVersion) => void}) => {
    const {api, ActionDropdown} = useMysqlClient();
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
    const catalogPage = useRef({page: 0, pageSize: 20, total: 0});
    const searchTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const connectedKey = useRef("");
    const mysqlTab = useRef<{database: string; tab: string} | undefined>(undefined);
    const live = useRef(true);
    const pendingDelete = useRef(false);
    const confirmation = useRef<any>();
    const lastRecord = useRef(record);
    if (record) lastRecord.current = record;
    const instance = record || lastRecord.current;
    const recordKey = record ? String(record.id) : "";
    const connectionKey = JSON.stringify([instance ? String(instance.id) : "", refreshKey]);
    const openedKey = useRef("");
    const instanceKey = useRef(recordKey);
    instanceKey.current = recordKey;
    const protectedName = (name: string) => ["mysql", "information_schema", "performance_schema", "sys"].includes(name.toLowerCase());
    const target = useMemo(() => selected && instance && loadedKey === connectionKey
        ? {...instance, type: 2, name: selected, database_name: selected, server_name: instance.name} : undefined,
    [instance, selected, loadedKey, connectionKey]);
    const changeTab = useCallback((tab: string) => { mysqlTab.current = {database: selected, tab}; }, [selected]);
    const load = useCallback(async (preferred?: string, page = 1, resetSelection = false) => {
        if (!record?.id || page > 1 && catalogPending.current) return;
        if (page === 1) { clearTimeout(searchTimer.current); searchTimer.current = undefined; }
        const request = ++sequence.current;
        const keyword = catalogKeyword.current;
        const connect = connectedKey.current !== connectionKey;
        catalogPending.current = true;
        setLoading(page === 1); setLoadingMore(page > 1);
        if (page === 1 && catalogList.current) catalogList.current.scrollTop = 0;
        try {
            // MySQL 工作区统一等待连接和版本信息，子面板复用版本，不再逐库重复请求。
            const [data, versionData] = await Promise.all([
                api.databases(record, {keyword, page, page_size: catalogPage.current.pageSize, show_system: catalogSystem.current, order_by_type: catalogOrder.current}),
                connect ? api.version(record) : undefined,
            ]);
            if (request !== sequence.current) return;
            if (!Array.isArray(data?.list)) throw new Error("数据库列表数据无效，请刷新后重试");
            const values = data.list.filter((name: any) => typeof name === "string");
            if (!Number.isInteger(data.page) || data.page < 1 || data.page > page || !Number.isInteger(data.page_size) || data.page_size < 1
                || !Number.isSafeInteger(data.total) || data.total < 0 || values.length !== data.list.length) throw new Error("数据库列表分页数据无效，请重试");
            catalogPage.current = {page: data.page, pageSize: data.page_size, total: data.total};
            if (connect) {
                if (!versionData || typeof versionData.version !== "string" || !versionData.version.trim() || !versionData.features || typeof versionData.features !== "object" || Array.isArray(versionData.features)) {
                    throw new Error("MySQL 版本信息不完整，请重新连接");
                }
                setVersion(versionData);
                connectedKey.current = connectionKey;
            }
            setNames((current) => page === 1 ? values : [...new Set([...current, ...values])]); setLoadedKey(connectionKey);
            // 搜索或刷新第一页不切换已打开的数据库，后续页也不改变右侧工作区。
            if (page === 1) setSelected((current) => {
                const next = resetSelection ? "" : preferred || current;
                if (next && (keyword || data.total > values.length || values.includes(next))) return next;
                return keyword ? "" : values[0] || "";
            });
        } catch (reason: any) {
            if (request !== sequence.current) return;
            sequence.current++; onCancel();
            message.error(reason.message || "获取数据库列表失败");
        } finally {
            if (request === sequence.current) { catalogPending.current = false; setLoading(false); setLoadingMore(false); }
        }
    }, [record?.id, connectionKey, api]);
    useEffect(() => {
        if (!record) return;
        live.current = true;
        const sameInstance = openedKey.current === recordKey;
        openedKey.current = recordKey;
        connectedKey.current = ""; catalogPending.current = false; catalogPage.current = {page: 0, pageSize: 20, total: 0};
        clearTimeout(searchTimer.current); searchTimer.current = undefined; setLoadingMore(false);
        setNames([]); setLoadedKey(""); setVersion(undefined); setEditor(undefined); setDetail(undefined); setShowCatalog(false); setRevision(0); setDeleting(false);
        if (!sameInstance) { setSelected(""); setSearch(""); catalogKeyword.current = ""; catalogSystem.current = false; setShowSystem(false); catalogOrder.current = "asc"; setOrder("asc"); mysqlTab.current = undefined; setFullscreen(false); }
        confirmation.current?.destroy(); pendingDelete.current = false;
        void load();
        return () => { live.current = false; sequence.current++; clearTimeout(searchTimer.current); confirmation.current?.destroy(); };
    }, [recordKey, load]);
    const clear = () => {
        // 配置连接只临时隐藏；真正关闭后才清理内容和版本，避免退出动画变形。
        if (instanceKey.current) return;
        lastRecord.current = undefined; openedKey.current = ""; connectedKey.current = ""; mysqlTab.current = undefined;
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
        // 大屏或展开目录时，若一页未填满可视区，继续读取直到可以滚动。
        const fill = () => { if (element.clientHeight > 0 && element.scrollHeight <= element.clientHeight + 1) loadNext(); };
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
        setSearch(""); setNames([]); setSelected(""); setLoadingMore(false); mysqlTab.current = undefined;
        void load(undefined, 1, true);
    };
    const choose = (name: string) => { setSelected(name); setShowCatalog(false); setDetail(undefined); };
    const remove = (name: string) => {
        if (!record || protectedName(name) || pendingDelete.current) return;
        const key = recordKey;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: `删除数据库“${name}”？`, content: "将删除数据库及全部数据，此操作不可撤销。", okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                if (pendingDelete.current || !live.current || instanceKey.current !== key) return;
                pendingDelete.current = true; setDeleting(true);
                try {
                    await api.drop_database(record, {names: [name]});
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
    const catalog = <aside className="database-catalog" aria-label="数据库目录">
        <div className="database-catalog-toolbar">
            <Input className="database-catalog-search" variant="filled" prefix={<Icon type="SearchOutlined"/>} placeholder="搜索数据库" aria-label="搜索数据库" value={search}
                onChange={(event) => searchCatalog(event.target.value)} onPressEnter={submitSearch} onBlur={submitSearch} allowClear/>
            <div className="database-catalog-actions">
                <Space size={4}>
                    <Button color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} aria-label="创建数据库" disabled={deleting}
                        onClick={() => { closeCatalog(); setEditor({}); }}/>
                    {!deleting && <ActionDropdown placement="bottomLeft" arrow={false} menu={{items: [
                        {key: "system", label: "显示系统库", icon: <Icon type={showSystem ? "CheckSquareOutlined" : "BorderOutlined"}/>, onClick: toggleSystem},
                    ]}}><Button color={showSystem ? "primary" : "default"} variant="filled" icon={<Icon type="FilterOutlined"/>} aria-label="数据库目录筛选"/></ActionDropdown>}
                    {!deleting && <ActionDropdown placement="bottomLeft" arrow={false} menu={{selectable: true, selectedKeys: [order], onClick: ({key}) => sortCatalog(key), items: [
                        {key: "asc", label: "名称升序", icon: <Icon type="SortAscendingOutlined"/>},
                        {key: "desc", label: "名称降序", icon: <Icon type="SortDescendingOutlined"/>},
                    ]}}><Button color="default" variant="filled" icon={<Icon type={order === "asc" ? "SortAscendingOutlined" : "SortDescendingOutlined"}/>}
                        aria-label="数据库排序"/></ActionDropdown>}
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
                        ...(!protectedName(name) ? [{key: "edit", label: "修改", onClick: () => { closeCatalog(); setEditor({name}); }}] : []),
                        ...(!protectedName(name) ? [{key: "delete", label: "删除", danger: true, onClick: () => { closeCatalog(); remove(name); }}] : []),
                    ]}}><Button type="text" size="small" icon={<Icon type="MoreOutlined"/>} aria-label={`数据库操作 ${name}`}/></ActionDropdown>}
                </div>)}
                {!names.length && <div className="database-catalog-state"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
                </div>}
                {loadingMore && <div className="database-catalog-more" role="status"><Spin size="small"/></div>}
            </>}
        </div>
        {version && <footer className={styles.mysqlInstance}>
            <Typography.Text className="mysql-instance-name" ellipsis={{tooltip: instance?.name}}>{instance?.name}</Typography.Text>
            <Typography.Text className="mysql-instance-version" ellipsis={{tooltip: `${version.name || "MySQL"} ${version.version}`}}>{version.name || "MySQL"} {version.version}</Typography.Text>
        </footer>}
    </aside>;
    return <>
        <Modal open={Boolean(record) && open} onCancel={onCancel} afterClose={clear} footer={null} destroyOnHidden={!record} closable={false}
            afterOpenChange={(visible) => { if (visible) setEntered(true); }}
            title="MySQL 管理" width={fullscreen ? "100vw" : compact ? 1120 : 1200}
            className={[styles.workspace, styles.mysqlWorkspace, fullscreen && "is-fullscreen", className].filter(Boolean).join(" ")}
            style={{top: fullscreen ? 0 : mobile ? 24 : 100, paddingBottom: fullscreen ? 0 : mobile ? 24 : 100,
                maxWidth: fullscreen ? "100vw" : "calc(100vw - 24px)", ...style}}
            focusable={{focusTriggerAfterClose: false}} mask={{closable: false}}>
            <div className="mysql-window">
                <header className="mysql-window-header">
                    <div className="mysql-window-title">
                        {mobile && <Tooltip title="数据库目录"><Button className="mysql-catalog-toggle" color="default" variant={showCatalog ? "filled" : "text"}
                            icon={<Icon type="MenuOutlined"/>} aria-label="数据库目录" aria-expanded={showCatalog} disabled={connecting}
                            onClick={() => setShowCatalog((value) => !value)}/></Tooltip>}
                        <span className={databaseStyles.databaseIcon} data-type="2" aria-hidden="true"/>
                        <span className="mysql-window-label">MySQL 管理</span>
                    </div>
                    <div className="mysql-window-actions">
                        {onConfigure && <Tooltip title="连接设置"><Button type="text" icon={<Icon type="SettingOutlined"/>} aria-label="连接设置" onClick={onConfigure}/></Tooltip>}
                        {!mobile && <Tooltip title={fullscreen ? "退出全屏" : "全屏"}><Button type="text" icon={<Icon type={fullscreen ? "FullscreenExitOutlined" : "FullscreenOutlined"}/>}
                            aria-label={fullscreen ? "退出全屏" : "全屏"} onClick={() => setFullscreen((value) => !value)}/></Tooltip>}
                        <Tooltip title="关闭"><Button type="text" icon={<Icon type="CloseOutlined"/>} aria-label="关闭 MySQL 管理" onClick={onCancel}/></Tooltip>
                    </div>
                </header>
                {/* 请求立即执行，入场动画结束后再挂载工作区，避免快速响应在缩放中途引发重排。 */}
                {connecting || !entered ? <div className="mysql-connection-state" role="status" aria-label="正在连接数据库"><Spin/></div> : <div className="database-browser">
                    {!mobile && catalog}
                    <section className="database-browser-content">
                        {target ? <Content key={`${instance!.id}:${selected}:${revision}`} record={target}
                            initialTab={mysqlTab.current?.database === selected ? mysqlTab.current.tab : initialTab || instance!.tab}
                            onTabChange={changeTab}/>
                            : <div className="database-browser-empty">{loading ? <Spin/> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择数据库"/>}</div>}
                    </section>
                    {mobile && <Drawer closable={false} placement="left" size="min(300px, calc(100% - 40px))" open={Boolean(record) && open && showCatalog}
                        getContainer={false} push={false} rootStyle={{position: "absolute"}} rootClassName={styles.mysqlCatalogDrawer}
                        onClose={closeCatalog} destroyOnHidden>{catalog}</Drawer>}
                </div>}
            </div>
        </Modal>
        <DatabaseForm open={Boolean(record) && Boolean(editor)} instance={instance} name={editor?.name} onCancel={() => setEditor(undefined)} onSuccess={(change: any) => {
            setSearch(""); catalogKeyword.current = ""; setRevision((value) => value + 1); setShowCatalog(false); void load(change.name);
        }}/>
        <Detail record={record ? detail : undefined} onCancel={() => setDetail(undefined)}/>
    </>;
};
