import {createContext, forwardRef, memo, useCallback, useContext, useEffect, useMemo, useRef, useState} from "react";
import type {Key} from "react";
import {App, Button, Empty, Form, Grid, Input, Select, Space, Tooltip} from "antd";
import type {MenuProps, TableColumnsType} from "antd";
import {Icon, ModalTable, ProModal, Title, useTheme} from "sinking-antd";
import dayjs from "dayjs";
import {getOssFiles, getOssInfo, ossFile, ossSign, uploadOssFile} from "@/service/api/oss";
import {formatSize} from "@/utils/string";
import ActionDropdown from "@/pages/components/ui/action-dropdown";
import {getFileIconType} from "../../utils";
import useTableStyles from "../table/styles";
import FileUpload from "../upload";
import {copyTextToClipboard} from "../properties/utils";
import {scheduleFileUpload} from "../upload/transfer";
import OssPathBar from "./path-bar";
import ServerTransfer from "./server-transfer";
import FileInfo from "./info";
import useOssClipboard from "./clipboard";
import useStyles from "./styles";

const FileContextMenu = createContext<any>(undefined);

const ContextMenuRow = forwardRef<HTMLTableRowElement, any>((rowProps, ref) => {
    const context = useContext(FileContextMenu);
    const key = String(rowProps["data-row-key"] || "");
    const record = context?.recordsByKey.get(key);
    if (!context || !record) return <tr {...rowProps} ref={ref}/>;
    return <ActionDropdown disabled={context.disabled}
        open={context.openMenu?.source === "context" && context.openMenu.key === key}
        onOpenChange={(open) => context.changeMenuOpen(open, key, "context")}
        classNames={{root: context.fileMenuClassName}}
        menu={{items: context.getMenuItems(record), onClick: () => context.changeMenuOpen(false, key, "context")}}
        trigger={["contextMenu"]} placement="bottomLeft" arrow={false}>
        <tr {...rowProps} ref={ref}/>
    </ActionDropdown>;
});

ContextMenuRow.displayName = "ContextMenuRow";
const tableComponents = {body: {row: ContextMenuRow}};

interface FilesProps {
    id: number;
    active: boolean;
    revision: number;
    onCancel: () => void;
    onEdit: (id: number) => void;
    onBusyChange: (id: number, busy: boolean) => void;
}

// 配置和文件列表共用一个弹窗；关闭时重置目录和分页，上传继续执行。
const Files = ({id, active, revision, onCancel, onEdit, onBusyChange}: FilesProps) => {
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const {styles: tableStyles} = useTableStyles();
    const [form] = Form.useForm();
    const alive = useRef(true);
    const visible = useRef(active);
    const request = useRef(0);
    const infoRequest = useRef(0);
    const loaded = useRef("");
    const listPrefix = useRef("");
    const configuration = useRef<number>();
    const dirty = useRef(true);
    const pending = useRef(false);
    const upload = useRef<any>(null);
    const uploading = useRef(false);
    const confirmation = useRef<ReturnType<typeof modal.confirm> | null>(null);
    const serverTransfer = useRef<any>(null);
    const busyChanged = useRef(onBusyChange);
    const editorVisible = useRef(false);
    const [record, setRecord] = useState<any>({id, name: "云端文件"});
    const [query, setQuery] = useState({prefix: "", keyword: "", tokens: [""], cursor: 0, pageSize: 10});
    const [searchValue, setSearchValue] = useState("");
    const [nextToken, setNextToken] = useState("");
    const [rows, setRows] = useState<any[]>([]);
    const [selected, setSelected] = useState<Key[]>([]);
    const [loading, setLoading] = useState(true);
    const [opened, setOpened] = useState(false);
    const [error, setError] = useState("");
    const [configError, setConfigError] = useState(false);
    const [busy, setBusy] = useState(false);
    const [serverBusy, setServerBusy] = useState(false);
    const [openMenu, setOpenMenu] = useState<{key: string; source: "context" | "action"}>();
    const [uploadingFiles, setUploadingFiles] = useState(false);
    const [editor, setEditor] = useState<{action: "mkdir" | "rename"; key?: string; prefix: string; title: string}>();
    const [editorOpen, setEditorOpen] = useState(false);
    const [info, setInfo] = useState<any>();
    const [infoLoading, setInfoLoading] = useState(false);
    const creatingDirectory = editor?.action === "mkdir";
    const renaming = editor?.action === "rename";
    const namePlaceholder = renaming ? "请输入新名称" : "请输入文件夹名称";
    const {prefix, keyword, cursor, tokens, pageSize} = query;
    // 打开动画及首次取数期间只显示加载状态，后续刷新保留列表和操作栏。
    const initialLoading = !opened || (loading && !loaded.current);
    const blocked = busy || uploadingFiles;
    const uploadDisabled = busy || loading || configuration.current !== revision;
    const disabled = blocked || loading || configuration.current !== revision;
    const queryKey = JSON.stringify([id, revision, prefix, keyword, tokens[cursor], pageSize]);
    const recordsByKey = useMemo(() => new Map(rows.map((item) => [item.key, item])), [rows]);
    const selectedRecords = useMemo(() => selected.map((key) => recordsByKey.get(String(key))).filter(Boolean), [selected, recordsByKey]);
    visible.current = active;
    busyChanged.current = onBusyChange;
    const changeMenuOpen = useCallback((open: boolean, key: string, source: "context" | "action") => {
        setOpenMenu((current) => open ? {key, source}
            : current?.key === key && current.source === source ? undefined : current);
    }, []);
    const closeEditor = useCallback(() => {
        editorVisible.current = false;
        setEditorOpen(false);
    }, []);
    const closeInfo = useCallback(() => {
        infoRequest.current++;
        setInfo(undefined); setInfoLoading(false);
    }, []);

    useEffect(() => {
        setOpenMenu((current) => active && !disabled && current && recordsByKey.has(current.key) ? current : undefined);
    }, [active, disabled, recordsByKey]);

    const load = useCallback(async () => {
        if (!visible.current) { dirty.current = true; return; }
        const current = ++request.current;
        setLoading(true); setError(""); setConfigError(false); setSelected([]); setNextToken("");
        try {
            // 配置按版本复用，更新配置只刷新内容，不重新挂载弹窗或传输组件。
            if (configuration.current !== revision) {
                const response = await getOssInfo({body: {id}});
                if (!alive.current || !visible.current || current !== request.current) return;
                if (response?.code !== 200 || !response.data?.config || Number(response.data.id) !== id) {
                    throw new Error(response?.message || "获取存储配置失败");
                }
                configuration.current = revision;
                setRecord(response.data);
            }
            const response = await getOssFiles({body: {id, prefix: prefix + keyword, delimiter: "/", cursor_id: tokens[cursor], page_size: pageSize}});
            if (!alive.current || !visible.current || current !== request.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "读取文件失败");
            const data = response.data;
            listPrefix.current = prefix;
            setRows(Array.isArray(data?.list) ? data.list : []);
            setNextToken(data?.has_more ? data.next_cursor_id || "" : "");
            loaded.current = queryKey;
            dirty.current = false;
        } catch (reason: any) {
            if (alive.current && visible.current && current === request.current) {
                setRows([]);
                setConfigError(configuration.current !== revision);
                setError(reason?.message || "读取文件失败");
            }
        } finally {
            if (alive.current && current === request.current) setLoading(false);
        }
    }, [id, revision, prefix, keyword, tokens, cursor, pageSize, queryKey]);

    useEffect(() => {
        if (active && (dirty.current || loaded.current !== queryKey)) void load();
    }, [active, load, queryKey]);
    useEffect(() => {
        if (active) return;
        request.current++;
        loaded.current = "";
        dirty.current = true;
        setQuery((current) => ({...current, prefix: "", keyword: "", tokens: [""], cursor: 0}));
        setSearchValue(""); setSelected([]); setNextToken(""); setError(""); setConfigError(false);
        setLoading(true); setOpened(false); setOpenMenu(undefined); closeEditor(); closeInfo();
        confirmation.current?.destroy();
    }, [active, closeEditor, closeInfo]);
    useEffect(() => { onBusyChange(id, blocked || serverBusy); }, [id, blocked, serverBusy, onBusyChange]);
    useEffect(() => {
        alive.current = true;
        return () => {
            alive.current = false;
            request.current++;
            infoRequest.current++;
            confirmation.current?.destroy();
            busyChanged.current(id, false);
        };
    }, [id]);

    // 写入后从当前目录的第一页重新读取，避免继续使用目录变更前的游标。
    const reload = useCallback(() => {
        if (!alive.current) return;
        dirty.current = true;
        request.current++;
        if (visible.current) setLoading(true);
        setQuery((current) => ({...current, tokens: [""], cursor: 0}));
    }, []);
    const changePasting = useCallback((value: boolean) => {
        pending.current = value;
        if (alive.current) setBusy(value);
    }, []);
    const clipboard = useOssClipboard({id, revision, active, path: prefix, disabled, selectedRecords,
        onChange: reload, onBusyChange: changePasting});
    const navigate = useCallback((value: string) => {
        dirty.current = true;
        request.current++;
        setLoading(true); setSelected([]); setNextToken(""); setOpenMenu(undefined); setSearchValue("");
        setQuery((current) => ({...current, prefix: value, keyword: "", tokens: [""], cursor: 0}));
    }, []);
    const search = useCallback((value: string) => {
        dirty.current = true;
        request.current++;
        setLoading(true); setSelected([]); setNextToken("");
        setQuery((current) => ({...current, keyword: value, tokens: [""], cursor: 0}));
    }, []);
    const invoke = useCallback(async (body: any) => {
        const response = await ossFile({body: {id: record.id, ...body}});
        if (response?.code !== 200) throw new Error(response?.message || "操作失败");
        if (body.action === "delete" && Array.isArray(response.data)) {
            const failed = response.data.filter((item: any) => item.error);
            if (failed.length) throw new Error(failed.map((item: any) => `${item.key}: ${item.error}`).join("；"));
        }
        return response.data;
    }, [record.id]);
    // 先打开详情弹窗再读取内容，独立校验请求，关闭后忽略旧响应。
    const details = useCallback(async (item: any) => {
        if (pending.current || uploading.current) return;
        const current = ++infoRequest.current;
        setInfo({key: item.key}); setInfoLoading(true);
        try {
            const data = await invoke({action: "info", key: item.key});
            if (!alive.current || !visible.current || current !== infoRequest.current) return;
            if (!data || data.key !== item.key) throw new Error("获取文件信息失败");
            setInfo(data);
        } catch (reason: any) {
            if (alive.current && visible.current && current === infoRequest.current) {
                closeInfo(); message.error(reason?.message || "获取文件信息失败");
            }
        } finally {
            if (alive.current && current === infoRequest.current) setInfoLoading(false);
        }
    }, [closeInfo, invoke, message]);
    // 下载和复制链接共用附件签名，复制时不打开浏览器窗口。
    const download = useCallback(async (item: any, action: "download" | "copy" = "download") => {
        if (pending.current || uploading.current) return;
        const tab = action === "download" ? window.open("about:blank", "_blank") : null;
        if (tab) tab.opener = null;
        pending.current = true; setBusy(true);
        try {
            const response = await ossSign({body: {id: record.id, key: item.key, method: "GET", expires: 86400, download: true}});
            if (response?.code !== 200) throw new Error(response?.message || "操作失败");
            const data = response.data;
            if (!data?.url) throw new Error("服务器未返回下载地址");
            if (Object.keys(data.headers || {}).some((key) => key.toLowerCase() !== "host")) throw new Error("下载链接需要额外请求头，请通过云厂商控制台下载");
            const url = new URL(data.url);
            if (!["http:", "https:"].includes(url.protocol)) throw new Error("下载地址不合法");
            if (action === "copy") {
                if (!await copyTextToClipboard(data.url)) throw new Error("复制下载链接失败，请重试");
                if (alive.current) message.success("下载链接已复制，有效期 24 小时");
            }
            else if (tab) tab.location.href = url.href;
            else { const link = document.createElement("a"); link.href = url.href; link.target = "_blank"; link.rel = "noopener noreferrer"; link.click(); }
        } catch (reason: any) { tab?.close(); if (alive.current) message.error(reason.message); }
        finally { pending.current = false; if (alive.current) setBusy(false); }
    }, [record.id, message]);
    const remove = useCallback((keys: string[]) => {
        if (pending.current || uploading.current || !keys.length) return;
        confirmation.current?.destroy();
        const instance = modal.confirm({
            title: `删除 ${keys.length} 项？`, content: keys.some((key) => key.endsWith("/"))
                ? "所选目录及其所有内容将一并删除，无法从面板回收站恢复。"
                : "此操作直接删除云端文件，无法从面板回收站恢复。",
            zIndex: 1040, okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onCancel: () => { if (confirmation.current === instance) confirmation.current = null; },
            onOk: async () => {
                if (!visible.current || pending.current || uploading.current) return;
                pending.current = true; setBusy(true);
                instance.update({keyboard: false, mask: {closable: false}, cancelButtonProps: {disabled: true}});
                try {
                    await invoke({action: "delete", keys});
                    if (confirmation.current === instance) confirmation.current = null;
                    if (alive.current) message.success("删除完成");
                }
                catch (reason: any) { if (alive.current) message.error(reason.message); throw reason; }
                finally {
                    pending.current = false;
                    if (alive.current) {
                        setBusy(false); reload();
                        if (confirmation.current === instance) instance.update({keyboard: true, mask: {closable: true}, cancelButtonProps: {disabled: false}});
                    }
                }
            },
        });
        confirmation.current = instance;
    }, [invoke, message, modal, reload]);
    const edit = useCallback((action: "mkdir" | "rename", item?: any) => {
        if (pending.current || uploading.current || !visible.current) return;
        const key = item?.key || "";
        form.resetFields();
        // 重命名只编辑最后一段名称，目录前缀交由后端原样保留。
        form.setFieldsValue({value: action === "mkdir" ? "" : key.slice(key.lastIndexOf("/") + 1)});
        setEditor({action, key: item?.key, prefix, title: action === "mkdir" ? "新建文件夹" : "重命名"});
        editorVisible.current = true;
        setEditorOpen(true);
        window.requestAnimationFrame(() => { if (editorVisible.current) form.focusField("value"); });
    }, [form, prefix]);
    const submit = useCallback(async (values: any) => {
        if (!editor || !editorVisible.current || pending.current || uploading.current || !visible.current) return;
        pending.current = true; setBusy(true);
        try {
            const body = editor.action === "mkdir"
                ? {action: "mkdir", key: `${editor.prefix}${values.value}/`}
                : {action: "rename", key: editor.key, name: values.value};
            await invoke(body);
            if (alive.current) { closeEditor(); message.success("操作完成"); reload(); }
        } catch (reason: any) { if (alive.current) message.error(reason.message); }
        finally { pending.current = false; if (alive.current) setBusy(false); }
    }, [editor, invoke, message, reload, closeEditor]);
    // 复用文件管理的上传列表与并发队列，每个文件保留选择时的目标目录。
    const uploadFile = useCallback(({path, file, onProgress}: any) => {
        const controller = new AbortController();
        let finalizing = false;
        const promise = scheduleFileUpload(async () => {
            onProgress({percent: 0, phase: "uploading"});
            const data = await uploadOssFile(id, path + file.name, file, controller.signal, (bytes, completing) => {
                finalizing = Boolean(completing);
                onProgress({percent: file.size ? bytes / file.size * 100 : 0,
                    phase: finalizing ? "merging" : "uploading", cancelable: !finalizing});
            });
            return {code: 200, data};
        }, controller.signal);
        return {promise, abort: () => { if (!finalizing) controller.abort(); }};
    }, [id]);
    const changeUploading = useCallback((value: boolean) => {
        uploading.current = value;
        if (alive.current) setUploadingFiles(value);
    }, []);

    const getMenuItems = useCallback((item: any): MenuProps["items"] => item.directory ? [
        {key: "open", label: "打开", onClick: () => navigate(item.key)},
        {type: "divider"},
        {key: "delete", label: "删除", danger: true, onClick: () => remove([item.key])},
    ] : [
        {key: "info", label: "文件信息", onClick: () => void details(item)},
        {key: "download", label: "下载到本机", onClick: () => void download(item)},
        {key: "server", label: "下载到服务器", onClick: () => serverTransfer.current?.open("download", item.key)},
        {key: "link", label: "复制下载链接", onClick: () => void download(item, "copy")},
        {type: "divider"},
        {key: "copy", label: "复制", onClick: () => clipboard.copyRecords([item])},
        {key: "move", label: "移动", onClick: () => clipboard.moveRecords([item])},
        {key: "rename", label: "重命名", onClick: () => edit("rename", item)},
        {key: "delete", label: "删除", danger: true, onClick: () => remove([item.key])},
    ], [navigate, details, download, edit, remove, clipboard.copyRecords, clipboard.moveRecords]);
    const contextMenuValue = useMemo(() => ({
        recordsByKey, disabled: !active || disabled, openMenu,
        fileMenuClassName: tableStyles.fileMenu, getMenuItems, changeMenuOpen,
    }), [recordsByKey, active, disabled, openMenu, tableStyles.fileMenu, getMenuItems, changeMenuOpen]);

    const columns = useMemo<TableColumnsType<any>>(() => [
        {title: "名称", dataIndex: "key", width: 300, render: (key, item) => {
            // 加载期间保留原列表及其目录前缀，避免旧文件名被新路径截断。
            const value = key.slice(listPrefix.current.length) || key;
            const name = item.directory && value.endsWith("/") ? value.slice(0, -1) || "/" : value;
            return <button type="button" className={tableStyles.fileNameButton} disabled={disabled}
                aria-label={`${item.directory ? "打开目录" : "查看文件信息"} ${name}`}
                onClick={() => item.directory ? navigate(key) : void details(item)}>
                <span className={tableStyles.fileNameContent}>
                    <Icon className={`${tableStyles.fileIcon} ${item.directory ? "folder" : ""}`} type={getFileIconType(name, item.directory)}/>
                    <Tooltip title={name}><span className={`${tableStyles.fileName} file-name`}>{name}</span></Tooltip>
                </span>
            </button>;
        }},
        {title: "大小", dataIndex: "size", width: 130, render: (value, item) => <span className={tableStyles.fileMeta}>{item.directory ? "—" : formatSize(value)}</span>},
        {title: "存储类型", dataIndex: "storage_class", width: 130, render: (value) => <span className={tableStyles.fileMeta}>{value || "—"}</span>},
        {title: "修改时间", dataIndex: "last_modified", width: 190, render: (value) => <span className={tableStyles.fileMeta}>{value && !value.startsWith("0001-") ? dayjs(value).format("YYYY-MM-DD HH:mm:ss") : "—"}</span>},
        {title: "操作", key: "action", width: 80, fixed: "right", className: "action-cell", render: (_, item) => (
            <ActionDropdown disabled={disabled} open={active && openMenu?.source === "action" && openMenu.key === item.key}
                onOpenChange={(open) => changeMenuOpen(open, item.key, "action")} classNames={{root: tableStyles.fileMenu}}
                menu={{items: getMenuItems(item), onClick: () => changeMenuOpen(false, item.key, "action")}}>
                <Button size="small" disabled={disabled}>操作</Button>
            </ActionDropdown>
        )},
    ], [tableStyles, disabled, navigate, details, active, openMenu, changeMenuOpen, getMenuItems]);

    return <>
        <ModalTable open={active} onCancel={onCancel} title={<Title>{record.name}</Title>}
            width={compact ? 980 : 1080} minHeight={compact ? 440 : 480} scrollMode="modal"
            manualRequest ariaLabel={`对象存储文件：${record.name}`} rowKey="key" columns={columns}
            modalProps={{className: "oss-files-modal", zIndex: 1010, style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"},
                afterOpenChange: setOpened,
                afterClose: () => { if (!visible.current) setRows([]); },
                footer: initialLoading ? null : <nav className={styles.pagination} aria-label="云端文件分页">
                    <div className="pagination-content">
                        <span className="pagination-count">本页 {rows.length} 项</span>
                        <div className="pagination-pages">
                            <Button type="text" size="small" icon={<Icon type="LeftOutlined"/>} aria-label="上一页" title="上一页" disabled={cursor === 0 || disabled}
                                onClick={() => { request.current++; setLoading(true); setQuery((current) => ({...current, cursor: current.cursor - 1})); }}/>
                            <span className="pagination-current" aria-current="page" aria-label={`第 ${cursor + 1} 页`}>{cursor + 1}</span>
                            <Button type="text" size="small" icon={<Icon type="RightOutlined"/>} aria-label="下一页" title="下一页" disabled={!nextToken || disabled}
                                onClick={() => { request.current++; setLoading(true); setQuery((current) => ({...current, tokens: [...current.tokens.slice(0, current.cursor + 1), nextToken], cursor: current.cursor + 1})); }}/>
                        </div>
                        <Select className={styles.pageSizeSelect} classNames={{popup: {root: styles.pageSizeDropdown}}}
                            size="small" variant="filled" value={pageSize} disabled={disabled} aria-label="分页容量"
                            showSearch={false} options={[10, 20, 50, 100].map((value) => ({value, label: `${value} 条 / 页`}))}
                            onChange={(value) => { dirty.current = true; request.current++; setLoading(true); setQuery((current) => ({...current, pageSize: value, tokens: [""], cursor: 0})); }}/>
                    </div>
                </nav>,
            }}
            dataSource={initialLoading ? [] : rows} loading={initialLoading || loading} pagination={false} scroll={{x: 830}} onReload={reload} requestEnabled={active}
            components={tableComponents}
            tableRender={(table) => <FileContextMenu.Provider value={contextMenuValue}>{table}</FileContextMenu.Provider>}
            rowClassName={(item) => openMenu?.key === item.key ? "action-menu-open" : ""}
            toolbar={initialLoading ? false : {
                search: {value: searchValue, ariaLabel: "按文件名前缀搜索", placeholder: "按文件名前缀搜索", onChange: setSearchValue, onSearch: search, debounce: 300},
                refresh: {loading, disabled: blocked, ariaLabel: "刷新文件列表", onClick: reload},
                actions: active ? [
                    {key: "create", type: "button", label: "新建文件夹", icon: "FolderAddOutlined", disabled, onClick: () => edit("mkdir")},
                    {key: "upload", type: "button", label: "上传文件", icon: "UploadOutlined", suffixIcon: "DownOutlined", disabled: uploadDisabled, menu: {items: [
                        {key: "local", label: "从本机上传", icon: <Icon type="UploadOutlined"/>, onClick: () => upload.current?.open()},
                        {key: "server", label: "从服务器上传", disabled, icon: <Icon type="CloudUploadOutlined"/>, onClick: () => serverTransfer.current?.open("upload", prefix)},
                    ]}},
                ] : [],
            }}
            contentBar={initialLoading ? false : {content: <OssPathBar path={prefix} disabled={disabled} onNavigate={navigate}/>, ariaLabel: "当前目录操作",
                actions: clipboard.count > 0 ? [{key: "paste", type: "button",
                    label: `${clipboard.mode === "move" ? "移动" : "粘贴"} ${clipboard.count} 项`,
                    "aria-label": clipboard.mode === "move" ? `将 ${clipboard.count} 项移动到当前目录` : `粘贴 ${clipboard.count} 项`,
                    icon: clipboard.mode === "move" ? "ScissorOutlined" : "SnippetsOutlined",
                    disabled, onClick: clipboard.paste}] : [],
            }}
            rowSelection={{selectedRowKeys: selected, columnWidth: 44, fixed: true, onChange: setSelected,
                onClear: () => setSelected([]), clearDisabled: disabled,
                getCheckboxProps: () => ({disabled}),
                actions: [
                    {key: "copy", type: "button", label: "复制", icon: "CopyOutlined", disabled, onClick: clipboard.copySelected},
                    {key: "move", type: "button", label: "移动", icon: "ScissorOutlined", disabled, onClick: clipboard.moveSelected},
                    {key: "delete", type: "button", label: "删除", icon: "DeleteOutlined", danger: true,
                        disabled, onClick: () => remove(selected.map(String))},
                ],
            }}
            emptyContent={error ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={configError ? error : "无法读取当前目录"}>
                <Space><Button size="small" icon={<Icon type="ReloadOutlined"/>} aria-label="重新加载文件列表" onClick={reload}>重新加载</Button>
                    {configError && <Button size="small" onClick={() => onEdit(id)}>修改配置</Button>}</Space>
            </Empty> : undefined}/>
        <FileUpload key={revision} ref={upload} path={prefix} active={active} zIndex={1140} request={uploadFile} resumable={false}
            selectionDisabled={uploadDisabled} onUploaded={reload} onUploadingChange={changeUploading}/>
        <ServerTransfer ref={serverTransfer} record={record} active={active} onChange={reload} onBusyChange={setServerBusy}/>
        <ProModal title={<Title>{editor?.title}</Title>} width="340px" okText={creatingDirectory ? "创建" : "保存"}
            onCancel={() => { if (!busy) closeEditor(); }} onOk={form.submit}
            modalProps={{open: active && editorOpen, zIndex: 1030, confirmLoading: busy, cancelText: "取消",
                afterClose: () => { if (!editorVisible.current) { setEditor(undefined); form.resetFields(); } },
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                styles: {body: {paddingTop: compact ? 10 : 15}},
                closable: !busy, keyboard: !busy, mask: {closable: !busy}, cancelButtonProps: {disabled: busy}, forceRender: true}}>
            <Form form={form} layout="vertical" variant="filled" disabled={busy} preserve={false}
                requiredMark={false} onFinish={submit}>
                <Form.Item name="value"
                    rules={[{required: true, whitespace: true, message: "请输入名称"},
                        {max: 255, message: "名称不能超过255个字符"},
                        {validator: async (_, value) => {
                            if (creatingDirectory && String(value).includes("/")) throw new Error("请输入单个文件夹名称");
                            if (renaming) {
                                const key = editor?.key || "";
                                if (value === "." || value === ".." || /[\\/\0\r\n]/.test(value)) throw new Error("请输入单个文件名称");
                                if (value === key.slice(key.lastIndexOf("/") + 1)) throw new Error("新名称不能与原名称相同");
                            }
                        }}]}><Input maxLength={255} aria-label={namePlaceholder} placeholder={namePlaceholder}/></Form.Item>
            </Form>
        </ProModal>
        <FileInfo info={info} loading={infoLoading} active={active} onCancel={closeInfo}/>
    </>;
};

export default memo(Files);
