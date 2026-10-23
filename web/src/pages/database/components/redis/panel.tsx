import {useEffect, useRef, useState} from "react";
import {createPortal} from "react-dom";
import type {Key} from "react";
import {App, Button, Checkbox, ConfigProvider, Empty, Form, Grid, Input, InputNumber, Segmented, Select, Space, Spin, Tooltip, Typography, theme as antdTheme} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import {CheckSquareOutlined, LeftOutlined, EditOutlined, MoreOutlined, PlusOutlined, ReloadOutlined, RightOutlined, SaveOutlined, SettingOutlined} from "@ant-design/icons";
import {useRedisClient} from "./context";
import Accounts from "./accounts";
import RedisStatus from "./status";
import RedisCommand from "./command";
import Maintenance from "./maintenance";
import RedisValueEditor, {redisContent, redisRows} from "./value-editor";
import useStyles from "./styles";

const templates: Record<string, any> = {string: "", hash: {}, list: [], set: [], zset: []};
const kinds: Record<string, string> = {string: "文本", hash: "字段集合", list: "列表", set: "去重集合", zset: "排序集合", stream: "数据流", none: "已不存在"};
const hints: Record<string, string> = {string: "保存一段文本，例如缓存、配置或 JSON。", hash: "保存多组字段和内容，例如用户资料。", list: "按顺序保存内容，允许重复。", set: "保存不重复的内容。", zset: "为内容设置分值，按分值排序。"};
const ttlText = (value?: number) => value === -1 ? "永久" : value === -2 ? "已过期" : value === undefined ? "—"
    : value >= 86400 ? `${Math.floor(value / 86400)} 天 ${Math.floor(value % 86400 / 3600)} 小时`
        : value >= 3600 ? `${Math.floor(value / 3600)} 小时 ${Math.floor(value % 3600 / 60)} 分钟`
            : value >= 60 ? `${Math.floor(value / 60)} 分 ${value % 60} 秒` : `${value} 秒`;

export default ({record, initialTab, onStateChange, onTabChange, catalogContainer, onCatalogClose}: any) => {
    const {api, ActionDropdown, features} = useRedisClient();
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const {token} = antdTheme.useToken();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const screens = Grid.useBreakpoint();
    const mobile = screens.md === undefined ? typeof window !== "undefined" && window.matchMedia("(max-width: 767px)").matches : !screens.md;
    const live = useRef(true);
    const listRequest = useRef(0);
    const valueRequest = useRef(0);
    const pending = useRef(false);
    const confirmation = useRef<any>(undefined);
    const [tab, setTab] = useState(["query", "command", "tools"].includes(initialTab) ? "command" : ["accounts", "info", "maintenance"].includes(initialTab) && !(initialTab === "accounts" && features.accounts === false) ? initialTab : "value");
    const [match, setMatch] = useState("");
    const [searchMode, setSearchMode] = useState("contains");
    const [pattern, setPattern] = useState("*");
    const [cursor, setCursor] = useState("0");
    const [history, setHistory] = useState<string[]>([]);
    const [list, setList] = useState<any>({list: []});
    const [checked, setChecked] = useState<Key[]>([]);
    const [selected, setSelected] = useState<string>();
    const [loading, setLoading] = useState(false);
    const [valueLoading, setValueLoading] = useState(false);
    const [valueError, setValueError] = useState("");
    const [detail, setDetail] = useState<any>();
    const [draft, setDraft] = useState<any>("");
    const originalDraft = useRef<any>("");
    const [dirty, setDirty] = useState(false);
    const [selecting, setSelecting] = useState(false);
    const closeCatalog = () => onCatalogClose?.();
    const [position, setPosition] = useState<any>({start: 0, cursor: ""});
    const [valueHistory, setValueHistory] = useState<any[]>([]);
    const [version, setVersion] = useState(0);
    const [valueVersion, setValueVersion] = useState(0);
    const [action, setAction] = useState<any>();
    const [displayAction, setDisplayAction] = useState<any>();
    const [saving, setSaving] = useState(false);
    const [form] = Form.useForm();
    const contentRef = useRef<HTMLDivElement>(null);
    const [contentHeight, setContentHeight] = useState(260);
    const dirtyRef = useRef(false);
    dirtyRef.current = dirty;
    const kind = Form.useWatch("type", form) || "string";
    const expiry = Form.useWatch("expiry", form) || "forever";
    const editable = Boolean(features.lua !== false && detail && detail.type in templates && !detail.has_more && !detail.truncated && !detail.encoding && !position.start && !position.cursor);

    useEffect(() => { onTabChange?.(tab); }, [tab, onTabChange]);
    useEffect(() => { onStateChange?.({dirty, busy: saving}); }, [dirty, saving, onStateChange]);
    useEffect(() => () => onStateChange?.({dirty: false, busy: false}), [onStateChange]);
    useEffect(() => {
        const node = contentRef.current;
        if (!node) return;
        const observer = new ResizeObserver(([entry]) => {
            if (entry.contentRect.height > 0) setContentHeight(Math.max(120, Math.floor(entry.contentRect.height)));
        });
        observer.observe(node);
        return () => observer.disconnect();
    }, [tab, selected]);
    useEffect(() => { live.current = true; return () => { live.current = false; confirmation.current?.destroy(); }; }, []);
    useEffect(() => { if (tab === "accounts" && features.accounts === false) setTab("value"); }, [tab, features.accounts]);
    useEffect(() => {
        const request = ++listRequest.current;
        setLoading(true); setChecked([]);
        void api.keys(record, {match: pattern, cursor, page_size: 20}).then((data) => {
            if (!live.current || request !== listRequest.current) return;
            // SCAN 在键变化期间可能返回重复项，导航按真实键名保持唯一。
            const rows = new Map<string, any>();
            for (const item of data?.list || []) rows.set(item.key, item);
            setList({...data, list: [...rows.values()]});
        }).catch((reason) => { if (live.current && request === listRequest.current) { setList({list: []}); message.error(reason.message); } })
            .finally(() => { if (live.current && request === listRequest.current) setLoading(false); });
    }, [record.id, pattern, cursor, version]);
    useEffect(() => {
        const request = ++valueRequest.current;
        setDetail(undefined); setDraft(""); setDirty(false); setValueError(""); setValueLoading(false);
        if (selected === undefined || tab !== "value") return;
        setValueLoading(true);
        void api.value(record, {key: selected, page_size: 50, ...position}).then((data) => {
            if (!live.current || request !== valueRequest.current) return;
            originalDraft.current = data?.type in templates ? redisRows(data.type, data.value) : data?.value;
            setDetail(data); setDraft(originalDraft.current);
        }).catch((reason) => { if (live.current && request === valueRequest.current) { setValueError(reason.message); message.error(reason.message); } })
            .finally(() => { if (live.current && request === valueRequest.current) setValueLoading(false); });
    }, [record.id, selected, position, valueVersion, tab]);
    useEffect(() => {
        if (tab !== "value" || selected !== undefined || loading) return;
        if (!list.list?.length) return;
        setSelected(list.list[0].key); setPosition({start: 0, cursor: ""}); setValueHistory([]);
    }, [tab, selected, loading, list]);

    const navigate = (run: () => void) => {
        if (pending.current) return;
        if (!dirty) { run(); return; }
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: "放弃未保存的修改？", content: "键的内容还没有保存。", okText: "放弃修改", cancelText: "继续编辑", okButtonProps: {danger: true}, onOk: () => {
            if (detail?.type in templates) setDraft(originalDraft.current);
            setDirty(false); dirtyRef.current = false; run();
        }});
    };
    const reload = () => { setVersion((value) => value + 1); setValueVersion((value) => value + 1); };
    const refreshBackground = () => {
        setVersion((value) => value + 1);
        // 命令执行后保留正在编辑的内容，避免后台刷新覆盖草稿。
        if (!dirtyRef.current) setValueVersion((value) => value + 1);
    };
    const select = (key: string) => {
        valueRequest.current++; setDetail(undefined); setSelected(key); setTab("value"); closeCatalog();
        setPosition({start: 0, cursor: ""}); setValueHistory([]); setDirty(false);
    };
    const search = (value: string, mode = searchMode, force = false) => {
        const keyword = mode === "exact" ? value : value.trim();
        const literal = keyword.replace(/[\\*?\[\]]/g, "\\$&");
        const next = !keyword ? "*" : mode === "contains" ? `*${literal}*` : mode === "exact" ? literal : keyword;
        if (next === pattern && !force) return;
        navigate(() => { setPattern(next); setCursor("0"); setHistory([]); setList({list: []}); setSelected(undefined); setVersion((value) => value + 1); });
    };
    const showAction = (operation: string, item = detail) => {
        if (operation === "save" && features.lua === false) return;
        form.resetFields();
        form.setFieldsValue(operation === "save" ? {key: "", type: "string", value: "", expiry: "forever", ttl: 3600}
            : operation === "rename" ? {new_key: item?.key ?? selected} : {expiry: item?.ttl >= 0 ? "seconds" : "forever", ttl: item?.ttl >= 0 ? item.ttl : 3600});
        const next = {operation, key: item?.key ?? selected};
        setAction(next); setDisplayAction(next);
    };
    const submit = async (values: any) => {
        if (pending.current || !action || (action.operation === "save" && features.lua === false)) return;
        pending.current = true; setSaving(true);
        const current = action;
        try {
            const body = current.operation === "save" ? {key: values.key, type: values.type, value: redisContent(values.type, values.value), ttl: values.expiry === "forever" ? -1 : values.ttl}
                : current.operation === "expire" ? {key: current.key, ttl: values.expiry === "forever" ? -1 : values.ttl} : {key: current.key, new_key: values.new_key};
            await api[current.operation as "save" | "expire" | "rename"](record, body);
            if (!live.current) return;
            message.success(current.operation === "save" ? "创建成功" : "保存成功"); setAction(undefined); reload();
            if (current.operation === "save") select(values.key);
            if (current.operation === "rename" && selected === current.key) select(values.new_key);
            if (current.operation === "expire" && values.expiry === "seconds" && values.ttl === 0) {
                setList((data: any) => ({...data, list: (data.list || []).filter((item: any) => item.key !== current.key)}));
                if (selected === current.key) setSelected(undefined);
            }
        } catch (reason: any) { if (live.current) message.error(reason.message); }
        finally { pending.current = false; if (live.current) setSaving(false); }
    };
    const saveValue = async () => {
        if (!editable || !dirty || pending.current) return;
        let value: string;
        try { value = redisContent(detail.type, draft); } catch (reason: any) { message.error(reason.message); return; }
        pending.current = true; setSaving(true);
        try {
            await api.save(record, {key: selected, type: detail.type, value});
            if (live.current) { message.success("保存成功"); setDirty(false); reload(); }
        } catch (reason: any) { if (live.current) message.error(reason.message); }
        finally { pending.current = false; if (live.current) setSaving(false); }
    };
    const remove = (keys: string[]) => {
        if (!keys.length || pending.current) return;
        confirmation.current = modal.confirm({title: keys.length > 1 ? `删除选中的 ${keys.length} 个键？` : "删除这个键？",
            content: keys.length > 1 ? "删除后无法恢复。" : <Typography.Text style={{overflowWrap: "anywhere"}}>{keys[0] || '""'}</Typography.Text>,
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true}, onOk: async () => {
                if (pending.current) return;
                pending.current = true; setSaving(true);
                try {
                    await api.delete(record, {keys: JSON.stringify(keys)});
                    if (live.current) { setList((data: any) => ({...data, list: (data.list || []).filter((item: any) => !keys.includes(item.key))}));
                        if (selected !== undefined && keys.includes(selected)) { valueRequest.current++; setSelected(undefined); setDirty(false); }
                        setChecked([]); if (selected !== undefined && keys.includes(selected)) reload(); else refreshBackground(); message.success("删除成功"); }
                } catch (reason: any) { if (live.current) message.error(reason.message); }
                finally { pending.current = false; if (live.current) setSaving(false); }
            }});
    };
    const keyMenu = (item: any) => saving ? [] : [
        {key: "rename", label: "重命名", onClick: () => navigate(() => showAction("rename", item))},
        {key: "expire", label: "有效期", onClick: () => navigate(() => showAction("expire", item))},
        {type: "divider" as const},
        {key: "delete", label: "删除键", danger: true, onClick: () => remove([item.key])},
    ];
    const detailView = <div className="redis-detail">
        <header className="redis-detail-header">
            <div className="redis-detail-heading"><div className="redis-detail-name">
                <Typography.Text className="redis-current-key" ellipsis={{tooltip: selected || '\"\"'}}>{selected || '\"\"'}</Typography.Text>
                <Typography.Text className="redis-key-copy" copyable={{text: selected}}/>
                {dirty && <Tooltip title="有未保存的修改"><span className="redis-dirty-dot"/></Tooltip>}
            </div><div className="redis-key-meta">{detail && <>
                <Tooltip title={kinds[detail.type] || detail.type}><span className="redis-type">{detail.type}</span></Tooltip>
                <span>{detail.size?.toLocaleString() ?? "—"} {detail.type === "string" ? "字节" : detail.type === "hash" ? "个字段" : "个成员"}</span>
                <Tooltip title={detail.ttl >= 0 ? `${detail.ttl} 秒后自动删除，点击修改` : "点击修改有效期"}>
                    <Button type="text" disabled={saving || valueLoading || detail.type === "none"} onClick={() => navigate(() => showAction("expire"))}>
                        {ttlText(detail.ttl)}<EditOutlined/>
                    </Button>
                </Tooltip>
            </>}</div></div>
            <div className="redis-detail-actions redis-toolbar">
                <Button color="default" variant="filled" icon={<ReloadOutlined/>} aria-label="刷新内容" loading={valueLoading} disabled={saving}
                    onClick={() => navigate(() => setValueVersion((value) => value + 1))}/>
                {!saving && !valueLoading && <ActionDropdown menu={{items: keyMenu({key: selected, ttl: detail?.ttl})}}>
                    <Button color="default" variant="filled" icon={<SettingOutlined/>}>操作</Button>
                </ActionDropdown>}
            </div>
        </header>
        <div className="redis-detail-content" ref={contentRef}><Spin spinning={valueLoading} classNames={{root: "redis-value-spin", container: "redis-value-spin-content"}}>
            {valueError ? <div className="redis-content-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/></div>
                : detail?.type === "none" ? <div className="redis-content-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="这个键已过期或被删除"/>
                    <Button color="default" variant="filled" onClick={() => { setList((data: any) => ({...data, list: (data.list || []).filter((item: any) => item.key !== selected)})); setSelected(undefined); reload(); }}>刷新列表</Button></div>
                    : detail && <>
                        {!editable && <div className="redis-readonly-note">{detail.encoding ? `二进制内容 · ${detail.encoding.toUpperCase()} 编码，只读显示`
                            : detail.has_more || detail.truncated || position.start || position.cursor ? "当前内容分段显示，仅供查看" : features.lua === false ? "当前版本不支持安全保存键值，仅供查看" : "此类型仅支持查看"}</div>}
                        <RedisValueEditor key={`${selected}:${position.start}:${position.cursor}`} kind={detail.type} value={draft} start={position.start} fill bodyHeight={contentHeight}
                            onChange={(value: any) => {
                                const original = originalDraft.current;
                                const changed = Array.isArray(value) ? value.length !== original.length || value.some((row: any, index: number) => {
                                    const previous = original[index];
                                    return row.key !== previous.key || row.value !== previous.value || row.member !== previous.member || row.score !== previous.score;
                                }) : value !== original;
                                setDraft(value); setDirty(changed); dirtyRef.current = changed;
                            }} readOnly={!editable} disabled={saving}/>
                    </>}
        </Spin></div>
        {(detail?.has_more || valueHistory.length > 0) && <div className="redis-member-pager redis-toolbar"><span>第 {valueHistory.length + 1} 段内容</span><Space size={8}>
            <Button color="default" variant="filled" disabled={!valueHistory.length || valueLoading || saving} onClick={() => navigate(() => { setPosition(valueHistory[valueHistory.length - 1]); setValueHistory((items) => items.slice(0, -1)); })}>上一段</Button>
            <Button color="default" variant="filled" disabled={!detail?.has_more || valueLoading || saving || detail?.type === "string"} onClick={() => navigate(() => { setValueHistory((items) => [...items, position]);
                setPosition(["hash", "set", "stream"].includes(detail.type) ? {start: 0, cursor: String(detail.cursor)} : {start: position.start + 50, cursor: ""}); })}>下一段</Button>
        </Space></div>}
        <footer className="redis-detail-footer redis-toolbar">
            <span className={`redis-save-state${dirty ? " is-dirty" : ""}`} role="status">{dirty ? "有未保存的修改" : editable ? "未修改" : "只读内容"}</span>
            <Space size={8}>
                <Button color="default" variant="filled" disabled={!dirty || saving} onClick={() => {
                    setDraft(originalDraft.current); setDirty(false); dirtyRef.current = false;
                }}>重置</Button>
                <Tooltip title="Ctrl / ⌘ + S"><Button type="primary" icon={<SaveOutlined/>} loading={saving}
                    disabled={!editable || !dirty || valueLoading} onClick={() => void saveValue()}>保存</Button></Tooltip>
            </Space>
        </footer>
    </div>;
    const catalog = <aside className="redis-catalog">
                <div className="redis-browser-toolbar redis-toolbar">
                    <div className="redis-search">
                        <ActionDropdown placement="bottomLeft" menu={{selectable: true, selectedKeys: [searchMode],
                            items: [{key: "contains", label: "包含"}, {key: "exact", label: "精确"}, {key: "pattern", label: "通配符"}],
                            onClick: ({key}: {key: string}) => { setSearchMode(key); if (match !== "") search(match, key); }}}>
                            <Button className="redis-search-mode" color="default" variant="text" aria-label="键名搜索方式">
                                {searchMode === "contains" ? "包含" : searchMode === "exact" ? "精确" : "通配符"}
                            </Button>
                        </ActionDropdown>
                        <Input variant="borderless" placeholder={searchMode === "pattern" ? "例如 user:*" : "搜索键名"} value={match} allowClear
                            onChange={(event) => { setMatch(event.target.value); if (!event.target.value) search(""); }}
                            onPressEnter={() => search(match, searchMode, true)} onBlur={() => search(match)}/>
                    </div>
                </div>
                <div className="redis-catalog-actions redis-toolbar"><Space size={4}>
                    <Button color={selecting ? "primary" : "default"} variant="filled" icon={<CheckSquareOutlined/>}
                        disabled={saving} aria-label="多选键" aria-pressed={selecting} onClick={() => setSelecting(true)}/>
                    {features.lua !== false && <Button color="default" variant="filled" icon={<PlusOutlined/>} disabled={saving}
                        aria-label="新建键" onClick={() => navigate(() => showAction("save"))}/>}
                </Space><Button color="default" variant="filled" icon={<ReloadOutlined/>} loading={loading} disabled={saving} aria-label="刷新列表" onClick={() => navigate(reload)}/></div>
                {selecting && <div className="redis-catalog-heading redis-toolbar">
                    <Checkbox aria-label="选择本页所有键" disabled={saving || loading || !list.list?.length}
                        checked={Boolean(list.list?.length) && checked.length === list.list.length} indeterminate={checked.length > 0 && checked.length < (list.list?.length || 0)}
                        onChange={(event) => setChecked(event.target.checked ? list.list.map((item: any) => item.key) : [])}/>
                        <span>已选 {checked.length} 项</span>
                    <Button type="text" disabled={saving} onClick={() => { setSelecting(false); setChecked([]); }}>取消多选</Button>
                </div>}
                <div className="redis-catalog-scroll"><Spin spinning={loading} classNames={{root: "redis-catalog-loading", container: "redis-catalog-items"}}>
                    {(list.list || []).map((item: any) => <ActionDropdown key={item.key} trigger={["contextMenu"]} placement="bottomLeft" arrow={false} disabled={saving || loading}
                        menu={{items: keyMenu(item)}}><div className={`redis-key-row${selected === item.key ? " is-selected" : ""}${selecting ? " is-selecting" : ""}`}>
                        {selecting && <Checkbox aria-label={`选择 ${item.key}`} checked={checked.includes(item.key)} disabled={saving || loading}
                            onChange={(event) => setChecked((items) => event.target.checked ? items.includes(item.key) ? items : [...items, item.key] : items.filter((key) => key !== item.key))}/>}
                        <button type="button" className="redis-key-item" title={item.key} disabled={saving || loading}
                            aria-current={selected === item.key ? "true" : undefined} onClick={() => {
                                if (selecting) setChecked((items) => items.includes(item.key) ? items.filter((key) => key !== item.key) : [...items, item.key]);
                                else if (selected !== item.key) navigate(() => select(item.key));
                                else { setTab("value"); closeCatalog(); }
                    }}><small>{item.type}</small><span>{item.key || '\"\"'}</span></button>
                        {!saving && !loading && <ActionDropdown menu={{items: keyMenu(item)}}><Button className="redis-key-menu" type="text" icon={<MoreOutlined/>}
                            aria-label={`操作 ${item.key}`}/></ActionDropdown>}
                    </div></ActionDropdown>)}
                    {!list.list?.length && !loading && <div className="redis-catalog-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE}
                            description="暂无数据"/></div>}
                </Spin></div>
                <footer className="redis-catalog-footer redis-toolbar">
                    {selecting && checked.length ? <Button color="danger" variant="text" disabled={saving || loading} onClick={() => remove(checked.map(String))}>删除所选</Button>
                        : <span>{list.list?.length || 0} 个键</span>}
                    <div className="redis-catalog-pager"><Button type="text" icon={<LeftOutlined/>} aria-label="上一页键列表" disabled={!history.length || loading || saving}
                        onClick={() => navigate(() => { setCursor(history[history.length - 1]); setHistory((items) => items.slice(0, -1)); setList({list: []}); setSelected(undefined); })}/>
                        <span>{history.length + 1}</span>
                        <Button type="text" icon={<RightOutlined/>} aria-label="下一页键列表" disabled={!list.has_more || loading || saving}
                            onClick={() => navigate(() => { setHistory((items) => [...items, cursor]); setCursor(String(list.cursor)); setList({list: []}); setSelected(undefined); })}/></div>
                </footer>
                {list.unsupported_binary_keys > 0 && <div className="redis-catalog-note">已跳过 {list.unsupported_binary_keys} 个二进制键</div>}
            </aside>;
    const browser = <div className="redis-browser">
        <div className="redis-workbench">
            {!mobile && catalog}
            <section className="redis-inspector">{selected !== undefined ? detailView : <div className="redis-inspector-empty">
                {loading ? <Spin/> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}
            </div>}</section>
        </div>
    </div>;

    return <div className={styles.panel} onKeyDown={(event) => {
        if (!action && tab === "value" && (event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "s") { event.preventDefault(); void saveValue(); }
    }}>
        <div className="redis-navigation">
            <Segmented aria-label="Redis 管理功能" value={tab} onChange={(next) => navigate(() => setTab(next))} options={[
                {value: "value", label: "数据浏览"}, {value: "command", label: "执行命令"},
                ...(features.accounts === false ? [] : [{value: "accounts", label: "账户管理"}]),
                {value: "info", label: "实例信息"}, {value: "maintenance", label: "数据维护"},
            ]}/>
        </div>
        {mobile && catalogContainer && createPortal(<div className={`${styles.panel} redis-mobile-catalog`}>{catalog}</div>, catalogContainer)}
        <div className="redis-tab-body">
            {tab === "value" && browser}
            <div className="redis-section redis-tool-page" hidden={tab !== "info"}><RedisStatus record={record} active={tab === "info"}/></div>
            {tab === "accounts" && features.accounts !== false && <div className="redis-section redis-tool-page"><Accounts record={record} initialAction={record.accountAction}/></div>}
            <div className="redis-section redis-tool-page redis-command-page" hidden={tab !== "command"}><RedisCommand record={record} onChange={refreshBackground}/></div>
            {tab === "maintenance" && <div className="redis-section redis-tool-page"><Maintenance record={record} onChange={(operation: string) => {
                if (operation === "flush") { valueRequest.current++; setCursor("0"); setHistory([]); setSelected(undefined); setDetail(undefined); setList({list: []}); setDirty(false); reload(); }
            }}/></div>}
        </div>
        <ProModal title={<Title>{displayAction?.operation === "save" ? "新建键" : displayAction?.operation === "rename" ? "重命名" : "设置有效期"}</Title>}
            onCancel={() => { if (!pending.current) setAction(undefined); }} onOk={form.submit} okText={displayAction?.operation === "save" ? "创建" : "保存"}
            modalProps={{open: Boolean(action), rootClassName: styles.modal,
                width: displayAction?.operation === "save" ? (compact ? 480 : 540) : displayAction?.operation === "rename" ? (compact ? 320 : 340) : (compact ? 320 : 360),
                style: {maxWidth: "calc(100vw - 24px)"}, styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}},
                confirmLoading: saving, cancelText: "取消", cancelButtonProps: {disabled: saving},
                closable: !saving, keyboard: !saving, mask: {closable: !saving}, forceRender: true,
                afterClose: () => { setDisplayAction(undefined); form.resetFields(); }}}>
            <Form form={form} layout="vertical" variant="filled" onFinish={(values) => {
                if (!action) return;
                if (action?.operation !== "save") { void submit(values); return; }
                confirmation.current = modal.confirm({title: "确认创建这个键？", content: "如果键名已存在，原有内容将被替换。", okText: "创建", cancelText: "取消", onOk: () => submit(values)});
            }} disabled={saving}>
                {displayAction && <>{displayAction.operation === "save" ? <>
                    <ConfigProvider theme={{components: {Input: {controlHeight: token.controlHeight}, Select: {controlHeight: token.controlHeight}}}}>
                    <div className="redis-create-fields"><Form.Item name="key" label="键名" rules={[{required: true, message: "请输入键名"}]}><Input placeholder="例如 user:1001"/></Form.Item>
                        <Form.Item name="type" label="数据类型" rules={[{required: true, message: "请选择数据类型"}]}><Select placeholder="请选择数据类型" options={Object.keys(templates).map((value) => ({value, label: `${kinds[value]} · ${value}`}))}
                            onChange={(value) => form.setFieldValue("value", redisRows(value, templates[value]))}/></Form.Item></div>
                    </ConfigProvider>
                    <Form.Item name="value" label={kind === "string" ? "文本内容" : kind === "hash" ? "字段与内容" : "成员内容"} extra={hints[kind]} validateTrigger="onBlur"
                        rules={[{validator: async (_, value) => { redisContent(kind, value); }}]}><RedisValueEditor kind={kind} embedded disabled={saving}/></Form.Item>
                </> : <>
                    {displayAction?.operation === "rename" && <ConfigProvider theme={{components: {Input: {controlHeight: token.controlHeight}}}}>
                        <Form.Item name="new_key" label="新键名" rules={[{required: true, message: "请输入新键名"}]}><Input placeholder="请输入新键名"/></Form.Item>
                    </ConfigProvider>}
                </>}
                {displayAction?.operation !== "rename" && <ConfigProvider theme={{components: {Select: {controlHeight: token.controlHeight}}}}>
                    <Form.Item name="expiry" label="有效期" rules={[{required: true, message: "请选择有效期"}]}>
                        <Select placeholder="请选择有效期" options={[{value: "forever", label: "永久保存"}, {value: "seconds", label: "到时删除"}]}/>
                    </Form.Item>
                    {expiry === "seconds" && <Form.Item name="ttl" label="多少秒后删除"
                        rules={[{required: true, message: "请输入有效期"}, {type: "number", min: displayAction?.operation === "expire" ? 0 : 1, message: displayAction?.operation === "expire" ? "有效期不能小于 0" : "有效期至少为 1 秒"}]}>
                        <InputNumber min={displayAction?.operation === "expire" ? 0 : 1} precision={0} style={{width: "100%"}} placeholder="例如 3600" suffix="秒"/></Form.Item>}
                </ConfigProvider>}</>}
            </Form>
        </ProModal>
    </div>;
};
