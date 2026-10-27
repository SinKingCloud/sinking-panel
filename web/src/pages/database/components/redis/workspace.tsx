import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Drawer, Empty, Grid, Input, Modal, Select, Spin, Tooltip, Typography} from "antd";
import {Icon, useTheme} from "sinking-antd";
import Content from "./content";
import {useRedisClient} from "./context";
import useStyles from "./workspace-styles";
import type {RedisManagerProps, RedisVersion} from "./entities";

export default ({record, open = true, onCancel, onConfigure, refreshKey = 0, initialTab, className, style, version, setVersion}: Omit<RedisManagerProps, "requests" | "components"> & {
    version?: RedisVersion; setVersion: (value?: RedisVersion) => void;
}) => {
    const {api} = useRedisClient();
    const {modal, message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const mobile = screens.md === undefined ? typeof window !== "undefined" && window.matchMedia("(max-width: 767px)").matches : !screens.md;
    const {styles} = useStyles({compact});
    const lastRecord = useRef(record);
    if (record) lastRecord.current = record;
    const instance = record || lastRecord.current;
    const recordKey = record ? String(record.id) : "";
    const connectionKey = JSON.stringify([instance?.id, refreshKey]);
    const instanceKey = useRef(recordKey);
    instanceKey.current = recordKey;
    const openedKey = useRef("");
    const activeTab = useRef<{database: string; tab: string} | undefined>(undefined);
    const sequence = useRef(0);
    const catalogList = useRef<HTMLDivElement>(null);
    const catalogKeyword = useRef("");
    const catalogPending = useRef(false);
    const catalogFailed = useRef(false);
    const catalogPage = useRef({page: 0, pageSize: 20, total: 0});
    const searchTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const editState = useRef({dirty: false, busy: false});
    const confirmation = useRef<any>(undefined);
    const [names, setNames] = useState<string[]>([]);
    const [selected, setSelected] = useState("");
    const [search, setSearch] = useState("");
    const [loadedKey, setLoadedKey] = useState("");
    const loaded = useRef("");
    loaded.current = loadedKey;
    const [loading, setLoading] = useState(false);
    const [loadingMore, setLoadingMore] = useState(false);
    const [fullscreen, setFullscreen] = useState(false);
    const [entered, setEntered] = useState(false);
    const [busy, setBusy] = useState(false);
    const [catalogOpen, setCatalogOpen] = useState(false);
    const [catalogContainer, setCatalogContainer] = useState<HTMLDivElement | null>(null);
    const closeCatalog = useCallback(() => setCatalogOpen(false), []);
    const load = useCallback(async (page = 1) => {
        if (!record?.id || page > 1 && catalogPending.current) return;
        if (page === 1) { clearTimeout(searchTimer.current); searchTimer.current = undefined; }
        const request = ++sequence.current;
        const keyword = catalogKeyword.current;
        const connect = loaded.current !== connectionKey;
        catalogPending.current = true;
        catalogFailed.current = false;
        setLoading(page === 1); setLoadingMore(page > 1);
        if (page === 1 && catalogList.current) catalogList.current.scrollTop = 0;
        try {
            const [data, server] = await Promise.all([api.databases(record, {keyword, page, page_size: catalogPage.current.pageSize}), connect ? api.version(record).catch((reason: any) => {
                // INFO 被 ACL 限制不代表连接失败，也不能据此推断为旧版本。
                if (/\bNOPERM\b/i.test(reason?.message || "")) return undefined;
                throw reason;
            }) : undefined]);
            if (request !== sequence.current) return;
            if (!Array.isArray(data?.list) || data?.list?.some((name: any) => typeof name !== "string")
                || !Number.isInteger(data?.page) || data?.page < 1 || data?.page > page || !Number.isInteger(data?.page_size) || data?.page_size < 1
                || !Number.isSafeInteger(data?.total) || data?.total < 0) throw new Error("数据库列表分页数据无效");
            if (server !== undefined && (!server?.version || !server?.features)) throw new Error("Redis 版本信息不完整");
            catalogPage.current = {page: data?.page, pageSize: data?.page_size, total: data?.total};
            setNames((current) => page === 1 ? data?.list : [...new Set<string>([...current, ...data?.list])]);
            if (connect) setVersion(server || {name: "Redis", version: "未知", features: {accounts: false}});
            loaded.current = connectionKey; setLoadedKey(connectionKey);
            // 搜索及后续页只更新目录，不切换正在编辑的逻辑库。
            if (page === 1) setSelected((current) => current && (keyword || data?.total > data?.list?.length || data?.list?.includes(current)) ? current : data?.list?.[0] || "");
        } catch (reason: any) {
            if (request !== sequence.current) return;
            catalogFailed.current = true;
            message.error(reason?.message || "获取数据库列表失败");
            if (connect) onCancel();
        } finally { if (request === sequence.current) { catalogPending.current = false; setLoading(false); setLoadingMore(false); } }
    }, [record?.id, connectionKey, api]);
    useEffect(() => {
        if (!record) return;
        const sameInstance = openedKey.current === recordKey;
        openedKey.current = recordKey;
        catalogPage.current = {page: 0, pageSize: 20, total: 0}; loaded.current = "";
        setNames([]); setLoadedKey(""); setVersion(undefined); setBusy(false); setCatalogOpen(false);
        if (!sameInstance) { catalogKeyword.current = ""; setSelected(""); setSearch(""); setFullscreen(false); activeTab.current = undefined; }
        editState.current = {dirty: false, busy: false};
        void load();
        return () => { sequence.current++; catalogPending.current = false; clearTimeout(searchTimer.current); searchTimer.current = undefined; confirmation.current?.destroy(); };
    }, [recordKey, load]);
    const loadNext = useCallback(() => {
        const {page, pageSize, total} = catalogPage.current;
        if (!catalogPending.current && !busy && page > 0 && page * pageSize < total) void load(page + 1);
    }, [busy, load]);
    useEffect(() => {
        const element = catalogList.current;
        if (!open || !element) return;
        const fill = () => { if (!catalogFailed.current && element.clientHeight > 0 && element.scrollHeight <= element.clientHeight + 1) loadNext(); };
        const observer = new ResizeObserver(fill);
        observer.observe(element); fill();
        return () => observer.disconnect();
    }, [open, entered, names, loading, loadingMore, mobile, loadNext]);
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
    const onStateChange = useCallback((state: {dirty: boolean; busy: boolean}) => { editState.current = state; setBusy(state.busy); }, []);
    const changeTab = useCallback((tab: string) => { activeTab.current = {database: selected, tab}; }, [selected]);
    const navigate = (run: () => void) => {
        if (editState.current.busy) { message.info("正在保存，请稍候"); return; }
        if (!editState.current.dirty) { run(); return; }
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: "放弃未保存的修改？", content: "键的内容还没有保存。", okText: "放弃修改", cancelText: "继续编辑",
            okButtonProps: {danger: true}, onOk: () => { if (!editState.current.busy) run(); }});
    };
    const target = useMemo(() => selected && instance && loadedKey === connectionKey
        ? {...instance, name: selected, database_name: selected, server_name: instance.name} : undefined, [selected, instance, loadedKey, connectionKey]);
    const clear = () => {
        if (instanceKey.current) return;
        lastRecord.current = undefined; openedKey.current = ""; activeTab.current = undefined;
        catalogKeyword.current = ""; catalogPage.current = {page: 0, pageSize: 20, total: 0};
        setNames([]); setSelected(""); setSearch(""); setLoadedKey(""); setVersion(undefined); setLoading(false);
        setLoadingMore(false); setCatalogOpen(false); setFullscreen(false); setEntered(false); editState.current = {dirty: false, busy: false};
    };
    const footer = version && <footer className="redis-instance-footer">
        <Typography.Text className="redis-instance-name" ellipsis={{tooltip: instance?.name}}>{instance?.name}</Typography.Text>
        <Typography.Text className="redis-instance-version" ellipsis={{tooltip: `Redis ${version.version}`}}>Redis {version.version}</Typography.Text>
    </footer>;
    const connecting = loadedKey !== connectionKey;
    const catalog = <aside className="redis-database-catalog" aria-label="数据库目录">
        <div className="redis-database-toolbar">
            <Input variant="filled" prefix={<Icon type="SearchOutlined"/>} placeholder="搜索数据库" aria-label="搜索数据库" value={search}
                onChange={(event) => searchCatalog(event.target.value)} onPressEnter={submitSearch} onBlur={submitSearch} allowClear/>
            <Button color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} loading={loading} disabled={busy}
                aria-label="刷新数据库列表" onClick={() => void load()}/>
        </div>
        <div ref={catalogList} className="redis-database-list" onScroll={(event) => {
            const {scrollHeight, scrollTop, clientHeight} = event.currentTarget;
            if (scrollHeight - scrollTop - clientHeight <= 48) loadNext();
        }}>{loading && !names.length ? <div className="redis-catalog-empty"><Spin/></div> : <>{names.map((name) => <button type="button" key={name}
            className={`redis-database-item${selected === name ? " is-selected" : ""}`} disabled={busy} aria-current={selected === name ? "true" : undefined}
            onClick={() => { if (name !== selected) navigate(() => setSelected(name)); }}><Icon type="DatabaseOutlined"/><span>DB {name}</span></button>)}
            {!names.length && <div className="redis-catalog-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/></div>}
        </>}{loadingMore && <div className="redis-catalog-more"><Spin size="small"/></div>}
        </div>
        {footer}
    </aside>;
    return <Modal open={Boolean(record) && open} onCancel={() => navigate(onCancel)} footer={null} title="Redis 管理" closable={false} destroyOnHidden={!record}
        afterOpenChange={(visible) => { if (visible) setEntered(true); }} afterClose={clear}
        mask={{closable: false}} focusable={{focusTriggerAfterClose: false}} className={[styles.modal, fullscreen && "is-fullscreen", className].filter(Boolean).join(" ")}
        width={fullscreen ? "100vw" : compact ? 1120 : 1200}
        style={{top: fullscreen ? 0 : mobile ? 24 : 100, paddingBottom: fullscreen ? 0 : mobile ? 24 : 100,
            maxWidth: fullscreen ? "100vw" : "calc(100vw - 24px)", ...style}}>
        <div className={styles.workspace}>
            <header className="redis-window-header">
                <div className="redis-window-title">
                    {mobile && <Tooltip title="数据库目录"><Button className="redis-catalog-toggle" color="default" variant={catalogOpen ? "filled" : "text"}
                        icon={<Icon type="MenuOutlined"/>} aria-label="数据库目录" disabled={connecting} onClick={() => setCatalogOpen((value) => !value)}/></Tooltip>}
                    <span className={styles.icon} aria-hidden="true"/><span>Redis 管理</span>
                </div>
                <div className="redis-window-actions">
                    {onConfigure && <Tooltip title="连接设置"><Button type="text" icon={<Icon type="SettingOutlined"/>} disabled={busy} aria-label="连接设置" onClick={() => navigate(onConfigure)}/></Tooltip>}
                    {!mobile && <Tooltip title={fullscreen ? "退出全屏" : "全屏"}><Button type="text" icon={<Icon type={fullscreen ? "FullscreenExitOutlined" : "FullscreenOutlined"}/>}
                        aria-label={fullscreen ? "退出全屏" : "全屏"} onClick={() => setFullscreen((value) => !value)}/></Tooltip>}
                    <Tooltip title="关闭"><Button type="text" icon={<Icon type="CloseOutlined"/>} disabled={busy} aria-label="关闭 Redis 管理" onClick={() => navigate(onCancel)}/></Tooltip>
                </div>
            </header>
            {connecting || !entered ? <div className="redis-connection-state"><Spin/></div> : <div className="redis-window-body">
                {!mobile && catalog}
                <div className="redis-window-main">{target ? <Content key={`${instance!.id}:${selected}`} record={target}
                    initialTab={activeTab.current?.database === selected ? activeTab.current.tab : initialTab || instance!.tab}
                    onTabChange={changeTab} onStateChange={onStateChange} catalogContainer={mobile ? catalogContainer : undefined}
                    onCatalogClose={closeCatalog}/>
                    : <div className="redis-connection-state"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/></div>}
                </div>
                {mobile && <Drawer open={Boolean(record) && open && catalogOpen} placement="left" size="min(340px, calc(100% - 40px))" closable={false}
                    push={false} getContainer={false} forceRender rootStyle={{position: "absolute"}} classNames={{body: styles.catalogBody}} onClose={closeCatalog}>
                    <div className="redis-mobile-database"><Select virtual aria-label="选择数据库" placeholder="请选择数据库" variant="filled" value={selected || undefined} loading={loading}
                        labelRender={({value}) => `DB ${value}`} showSearch filterOption={false} searchValue={search} onSearch={searchCatalog}
                        onBlur={submitSearch} onOpenChange={(visible) => { if (!visible && catalogKeyword.current) searchCatalog(""); }}
                        onPopupScroll={(event) => {
                            const {scrollHeight, scrollTop, clientHeight} = event.currentTarget;
                            if (scrollHeight - scrollTop - clientHeight <= 48) loadNext();
                        }} notFoundContent={loading ? <Spin size="small"/> : undefined}
                        disabled={busy} options={names.map((name) => ({value: name, label: `DB ${name}`}))} onChange={(name) => navigate(() => setSelected(name))}/>
                        <Button color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} loading={loading} disabled={busy}
                            aria-label="刷新数据库" onClick={() => void load()}/>
                    </div>
                    <div className="redis-mobile-key-host" ref={setCatalogContainer}/>
                    {footer}
                </Drawer>}
            </div>}
        </div>
    </Modal>;
};
