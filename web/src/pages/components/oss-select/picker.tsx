import {forwardRef, useCallback, useContext, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {App, Button, Checkbox, ConfigProvider, Empty, Form, Grid, Input, Select, Space, Table, Tooltip, Typography} from "antd";
import type {ButtonProps, InputProps, InputRef, TableProps, TableRef} from "antd";
import DisabledContext from "antd/es/config-provider/DisabledContext";
import useVariant from "antd/es/form/hooks/useVariants";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {getOssFiles, ossFile} from "@/service/api/oss";
import {readFileStorage, updateFileStorage} from "@/pages/file/hooks/file-storage";
import {formatFileSize, formatFileTime, getFileIconType} from "@/pages/file/utils";
import useStyles from "@/pages/components/ui/path-select/styles";

/** 单选、多选均使用对象键数组；目录模式下空字符串表示桶根目录。 */
export interface OssPathPickerProps extends Omit<InputProps, "value" | "defaultValue" | "onChange" | "onSelect"> {
    mode: "file" | "directory";
    /** 已保存的 OSS 连接 ID；未选择连接时禁用浏览按钮。 */
    ossId?: string | number;
    initialPath?: string;
    zIndex?: number;
    buttonProps?: ButtonProps;
    multiple?: boolean;
    value?: string[];
    defaultValue?: string[];
    onChange?: (value: string[]) => void;
    onSelect?: (value: string[], records: any[]) => void | false;
}

const directoryPrefix = (key: string) => key && !key.endsWith("/") ? `${key}/` : key;
const parentPrefix = (key: string) => key.slice(0, key.lastIndexOf("/", key.endsWith("/") ? key.length - 2 : key.length) + 1);
const displayPath = (key: string) => `/${key}`;

/** 对象路径保留原始字符；分页只取当前游标，不读取整个桶来统计数量。 */
const OssPathPicker = forwardRef<InputRef, OssPathPickerProps>((props, ref) => {
    const {
        mode, ossId, value, defaultValue, multiple, onChange, onSelect, initialPath, zIndex, buttonProps,
        disabled, size, className, style, placeholder, ...inputProps
    } = props;
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles, cx} = useStyles({compact});
    const contextDisabled = useContext(DisabledContext);
    const isDisabled = disabled ?? contextDisabled;
    const [variant] = useVariant("input", inputProps.variant, inputProps.bordered);
    const multi = Boolean(multiple);
    const selectionKeys = useCallback((keys: string[]) => {
        if (!multi) return keys.slice(0, 1);
        const unique = new Set(mode === "directory" ? keys.map(directoryPrefix) : keys);
        if (mode !== "directory") return [...unique];
        // 保留父目录即可，空键表示整个桶；按目录层级匹配，避免误包含相似前缀。
        return [...unique].filter((key) => {
            for (let parent = parentPrefix(key); parent !== key; parent = parentPrefix(parent)) {
                if (unique.has(parent)) return false;
                if (!parent) break;
            }
            return true;
        });
    }, [mode, multi]);
    const selectionItems = useCallback((items: any[]) => {
        const available = new Map(items.map((item) => [mode === "directory" ? directoryPrefix(item.key) : item.key, item]));
        return selectionKeys([...available.keys()]).map((key) => ({...available.get(key), key}));
    }, [mode, selectionKeys]);
    const [localValue, setLocalValue] = useState<string[]>(defaultValue ?? []);
    const currentValue = Object.prototype.hasOwnProperty.call(props, "value") ? value : localValue;
    const values = useMemo(() => selectionKeys(currentValue ?? []), [currentValue, selectionKeys]);
    const singleValue = values[0];
    const [keyDraft, setKeyDraft] = useState("");
    const validId = Number(ossId) > 0;
    const title = mode === "directory" ? "选择目录" : "选择文件";
    const [open, setOpen] = useState(false);
    const [query, setQuery] = useState({prefix: "", cursor: 0, tokens: [""]});
    const [rows, setRows] = useState<any[]>([]);
    const [nextToken, setNextToken] = useState("");
    const [loadedPath, setLoadedPath] = useState<string>();
    const [selected, setSelected] = useState<any>();
    const [selectedItems, setSelectedItems] = useState<any[]>([]);
    const selectedKeys = useMemo(() => new Set(selectedItems.map((item) => item.key)), [selectedItems]);
    const checkedKeys = useMemo(() => new Set([...selectedKeys, ...rows.filter((item) => multi && mode === "directory"
        && item.directory && [...selectedKeys].some((key) => item.key.startsWith(key))).map((item) => item.key)]), [mode, multi, rows, selectedKeys]);
    const selectableRows = useMemo(() => rows.filter((item) => Boolean(item.directory) === (mode === "directory")), [mode, rows]);
    const partialKeys = useMemo(() => new Set(mode === "directory" ? selectableRows.filter((item) => !checkedKeys.has(item.key)
        && [...selectedKeys].some((key) => key.startsWith(item.key))).map((item) => item.key) : []), [checkedKeys, mode, selectableRows, selectedKeys]);
    const allChecked = selectableRows.length > 0 && selectableRows.every((item) => checkedKeys.has(item.key));
    const someChecked = !allChecked && selectableRows.some((item) => checkedKeys.has(item.key) || partialKeys.has(item.key));
    const [selecting, setSelecting] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");
    const [reload, setReload] = useState(0);
    const [editing, setEditing] = useState(false);
    const [draft, setDraft] = useState("");
    const [creating, setCreating] = useState(false);
    const [saving, setSaving] = useState(false);
    const [folderForm] = Form.useForm();
    const request = useRef(0);
    const folderRequest = useRef(0);
    const folderPending = useRef(false);
    const selectionRequest = useRef(0);
    const selectionPending = useRef(false);
    const selectionDirectories = useRef(new Map<string, any[]>());
    const selectionScope = useRef({ossId: String(ossId ?? ""), mode, multi});
    const breadcrumbRef = useRef<HTMLOListElement>(null);
    const pathInput = useRef<InputRef>(null);
    const table = useRef<TableRef>(null);
    const {prefix, cursor, tokens} = query;
    const directoryCovered = multi && mode === "directory" && [...selectedKeys].some((key) => prefix.startsWith(key));
    const loaded = loadedPath !== undefined && loadedPath === prefix && !loading && !error;
    const breadcrumbs = useMemo(() => {
        const parts = [{key: "", label: "/"}];
        let start = 0;
        for (let index = 0; index < prefix.length; index++) {
            if (prefix[index] !== "/") continue;
            parts.push({key: prefix.slice(0, index + 1), label: prefix.slice(start, index) || "/"});
            start = index + 1;
        }
        return parts;
    }, [prefix]);

    const change = (keys: string[]) => {
        const next = selectionKeys(keys);
        setLocalValue(next);
        onChange?.(next);
    };
    const addKey = () => {
        if (keyDraft) {
            const key = keyDraft.startsWith("/") ? keyDraft.slice(1) : keyDraft;
            if (mode === "directory" || key) change([...values, key]);
        }
        setKeyDraft("");
    };
    const updateSelection = async (items: any[], checked: boolean) => {
        if (selectionPending.current || isDisabled) return;
        const current = ++selectionRequest.current;
        selectionPending.current = true;
        setSelecting(true);
        try {
            let next = selectedItems;
            const cache = selectionDirectories.current;
            if (loaded && cursor === 0 && !nextToken) cache.set(prefix, rows);
            // 反选继承的目录时，只列举目标分支的各层兄弟目录，不扫描整棵树。
            const children = async (directory: string) => {
                if (cache.has(directory)) return cache.get(directory)!;
                const result: any[] = [];
                const cursors = new Set<string>();
                let token = "";
                do {
                    if (selectionRequest.current !== current) throw new Error("选择已取消");
                    const response = await getOssFiles({body: {id: ossId, prefix: directory, delimiter: "/", cursor_id: token, page_size: 1000}});
                    if (selectionRequest.current !== current) throw new Error("选择已取消");
                    if (response?.code !== 200) throw new Error(response?.message || "读取目录失败");
                    const data = response?.data;
                    if (!Array.isArray(data?.list)) throw new Error("目录数据不完整，请刷新后重试");
                    result.push(...data.list.filter((item: any) => typeof item?.key === "string"));
                    if (!data.has_more) break;
                    token = String(data.next_cursor_id || "");
                    if (!token || cursors.has(token)) throw new Error("目录分页异常，请刷新后重试");
                    cursors.add(token);
                } while (true);
                cache.set(directory, result);
                return result;
            };
            if (checked) next = selectionItems([...next, ...items]);
            else for (const item of selectionItems(items)) {
                if (mode !== "directory") {
                    next = next.filter((entry) => entry.key !== item.key);
                    continue;
                }
                const ancestor = next.find((entry) => item.key.startsWith(entry.key));
                next = next.filter((entry) => !entry.key.startsWith(item.key) && entry !== ancestor);
                if (!ancestor || ancestor.key === item.key) continue;
                let branch = ancestor.key;
                while (branch !== item.key) {
                    const child = item.key.slice(0, item.key.indexOf("/", branch.length) + 1);
                    const siblings = (await children(branch)).filter((entry) => entry.directory && entry.key !== branch
                        && parentPrefix(entry.key) === branch);
                    if (!child || !siblings.some((entry) => entry.key === child)) throw new Error("目录已变化，请刷新后重试");
                    next = [...next, ...siblings.filter((entry) => entry.key !== child)];
                    branch = child;
                }
            }
            if (checked && mode === "directory") {
                // 已完整读取过的层级可自下而上收起，重新勾全子目录后恢复父目录全选。
                for (const [directory, entries] of [...cache].sort(([left], [right]) => right.length - left.length)) {
                    const children = entries.filter((entry) => entry.directory && entry.key !== directory && parentPrefix(entry.key) === directory);
                    if (!children.length || next.some((entry) => directory.startsWith(entry.key))) continue;
                    if (children.every((child) => next.some((entry) => child.key.startsWith(entry.key)))) {
                        next = [...next.filter((entry) => !entry.key.startsWith(directory)), {key: directory, directory: true}];
                    }
                }
            }
            if (selectionRequest.current === current) setSelectedItems(selectionItems(next));
        } catch (reason: any) {
            if (selectionRequest.current === current) message.error(reason?.message || "更新选择失败");
        } finally {
            if (selectionRequest.current === current) {
                selectionPending.current = false;
                setSelecting(false);
            }
        }
    };
    const selectItem = (item: any) => {
        if (multi) void updateSelection([item], !checkedKeys.has(item.key));
        else setSelected(item);
    };
    const close = useCallback(() => {
        request.current++;
        folderRequest.current++;
        selectionRequest.current++;
        selectionPending.current = false;
        selectionDirectories.current.clear();
        setSelecting(false);
        setOpen(false);
        setCreating(false);
        setEditing(false);
    }, []);
    useEffect(() => {
        const previous = selectionScope.current;
        const next = {ossId: String(ossId ?? ""), mode, multi};
        if (previous.ossId !== next.ossId || previous.mode !== mode || previous.multi !== multi) {
            setLocalValue([]);
            setKeyDraft("");
        }
        selectionScope.current = next;
        close();
        setSelectedItems([]);
        return () => { request.current++; folderRequest.current++; selectionRequest.current++; };
    }, [ossId, mode, multi, close]);
    useEffect(() => { if (isDisabled) close(); }, [isDisabled, close]);

    const navigate = useCallback((key: string) => {
        if (selectionPending.current) return;
        request.current++;
        setLoading(true);
        setLoadedPath(undefined);
        setRows([]);
        setError("");
        setSelected(undefined);
        setEditing(false);
        setQuery({prefix: directoryPrefix(key), cursor: 0, tokens: [""]});
    }, []);
    const show = () => {
        if (isDisabled || !validId) return;
        let directory = initialPath;
        const firstValue = values[0];
        if (directory === undefined && firstValue !== undefined) {
            directory = mode === "file" ? parentPrefix(firstValue) : firstValue;
        }
        if (directory === undefined) {
            try { directory = readFileStorage()?.ossPickerPaths?.[String(ossId)]?.[mode]; } catch { /* 路径记忆不可用时仍可浏览。 */ }
        }
        navigate(typeof directory === "string" ? directory : "");
        setSelected(mode === "file" && singleValue ? {key: singleValue} : undefined);
        setSelectedItems(multi ? values.map((key) => ({key, directory: mode === "directory"})) : []);
        setOpen(true);
    };

    useEffect(() => {
        if (!open || !validId) return;
        const current = ++request.current;
        setLoading(true);
        setError("");
        setNextToken("");
        void getOssFiles({body: {id: ossId, prefix, delimiter: "/", cursor_id: tokens[cursor], page_size: 50}})
            .then((response) => {
                if (request.current !== current) return;
                if (response?.code !== 200) throw new Error(response?.message || "读取文件列表失败");
                const data = response?.data;
                setRows(Array.isArray(data?.list) ? data?.list.filter((item: any) => typeof item?.key === "string") : []);
                if (cursor === 0 && !data?.has_more && Array.isArray(data?.list)) {
                    selectionDirectories.current.set(prefix, data.list.filter((item: any) => typeof item?.key === "string"));
                }
                setNextToken(data?.has_more ? String(data?.next_cursor_id || "") : "");
                setLoadedPath(prefix);
                try {
                    if (readFileStorage()?.ossPickerPaths?.[String(ossId)]?.[mode] !== prefix) {
                        updateFileStorage((current: any) => ({...current, ossPickerPaths: {
                            ...current?.ossPickerPaths,
                            [String(ossId)]: {...current?.ossPickerPaths?.[String(ossId)], [mode]: prefix},
                        }}));
                    }
                } catch { /* 保存路径失败不影响当前选择。 */ }
            }).catch((reason: any) => {
                if (request.current !== current) return;
                setRows([]);
                setLoadedPath(undefined);
                setError(reason?.message || "读取文件列表失败");
            }).finally(() => { if (request.current === current) setLoading(false); });
        return () => { if (request.current === current) request.current++; };
    }, [open, validId, ossId, prefix, cursor, tokens, reload, mode]);

    useLayoutEffect(() => {
        if (editing) { pathInput.current?.focus(); pathInput.current?.select(); }
        else if (breadcrumbRef.current) breadcrumbRef.current.scrollLeft = breadcrumbRef.current.scrollWidth;
    }, [editing, breadcrumbs, open]);

    const refresh = () => {
        if (selectionPending.current) return;
        for (const directory of selectionDirectories.current.keys()) {
            if (directory.startsWith(prefix)) selectionDirectories.current.delete(directory);
        }
        request.current++;
        setLoading(true);
        setQuery((current) => ({...current, cursor: 0, tokens: [""]}));
        setReload((current) => current + 1);
    };
    const confirm = (record = selected) => {
        if (isDisabled || !validId || selectionPending.current) return;
        if (multi) {
            const items = selectionItems(selectedItems);
            const keys = items.map((item) => item.key);
            if (onSelect?.(keys, items) === false) return;
            change(keys);
        } else {
            if (!loaded || (mode === "file" && !record?.key)) return;
            const key = mode === "directory" ? prefix : record.key;
            if (onSelect?.([key], [mode === "directory" ? {key, directory: true} : record]) === false) return;
            change([key]);
        }
        close();
    };
    const columns = useMemo<TableProps<any>["columns"]>(() => [
        {title: "名称", dataIndex: "key", render: (key: string, record: any) => {
            const name = key.slice(prefix.length).replace(/\/$/, "") || "/";
            return <div className="path-select-name-cell">
                <Icon type={getFileIconType(name, record?.directory)} className="path-select-icon"/>
                <span className="path-select-name" title={name}>{name}</span>
                {!multi && mode === "file" && !record?.directory && <span className="path-select-row-action">
                    {selected?.key === key && <Icon type="CheckOutlined"/>}
                </span>}
            </div>;
        }},
        {title: "大小", dataIndex: "size", width: compact ? 80 : 96, render: (size: number, record: any) => record?.directory ? "-" : formatFileSize(size)},
        {title: "修改时间", dataIndex: "last_modified", width: compact ? 132 : 154,
            render: (value: string) => value?.startsWith("0001-") ? "-" : formatFileTime(value)},
    ], [compact, mode, multi, prefix, selected]);

    const openFolder = () => {
        if (!loaded || saving || selectionPending.current) return;
        folderRequest.current++;
        folderForm.resetFields();
        setCreating(true);
    };
    const createFolder = async ({name}: {name: string}) => {
        if (!open || !creating || folderPending.current || selectionPending.current || isDisabled || !validId) return;
        const current = ++folderRequest.current;
        folderPending.current = true;
        setSaving(true);
        try {
            const response = await ossFile({body: {id: ossId, action: "mkdir", key: `${prefix}${name}/`}});
            if (folderRequest.current !== current) return;
            if (response?.code !== 200) throw new Error(response?.message || "新建文件夹失败");
            selectionDirectories.current.delete(prefix);
            setCreating(false);
            message.success("新建成功");
            setQuery((current) => ({...current, cursor: 0, tokens: [""]}));
            setReload((current) => current + 1);
        } catch (reason: any) {
            if (folderRequest.current === current) message.error(reason?.message || "新建文件夹失败");
        } finally {
            folderPending.current = false;
            setSaving(false);
        }
    };

    return <>
        <Space.Compact block size={size} className={cx(styles.field, className)} style={style}>
            {multi ? <Select {...inputProps as any} ref={ref as any} size={size} disabled={isDisabled} variant={variant}
                mode="multiple" value={values} options={values.map((key) => ({value: key, label: displayPath(key)}))}
                allowClear open={false} suffixIcon={null} maxTagCount="responsive"
                searchValue={keyDraft} onSearch={setKeyDraft}
                placeholder={placeholder || (mode === "directory" ? "请输入或选择目录" : "请输入或选择文件")}
                onInputKeyDown={(event) => {
                    if (event.key === "Enter" && !event.nativeEvent.isComposing) {
                        event.preventDefault(); event.stopPropagation(); addKey();
                    }
                }}
                onBlur={(event) => { addKey(); inputProps.onBlur?.(event as any); }}
                onClear={() => setKeyDraft("")} onChange={change}/>
                : <Input {...inputProps} ref={ref} size={size} disabled={isDisabled} value={singleValue === undefined ? "" : displayPath(singleValue)}
                placeholder={placeholder || (mode === "directory" ? "请输入或选择 OSS 目录" : "请输入或选择 OSS 文件")}
                onChange={(event) => change(event.target.value ? [event.target.value.startsWith("/") ? event.target.value.slice(1) : event.target.value] : [])}/>}
            <Tooltip title={title}><Button color={buttonProps?.type || buttonProps?.danger ? undefined : "default"}
                variant={buttonProps?.type || buttonProps?.danger ? undefined : variant === "filled" ? "filled" : variant === "outlined" ? "outlined" : "text"}
                {...buttonProps} size={size} disabled={isDisabled || !validId} aria-label={title}
                icon={<Icon type={mode === "directory" ? "FolderOpenOutlined" : "FileSearchOutlined"}/>} onClick={show}/></Tooltip>
        </Space.Compact>
        <ConfigProvider componentSize="middle" variant="outlined">
            <ProModal title={<Title>{title}</Title>} width="640px" okText="选择" onOk={() => confirm()} onCancel={close}
                modalProps={{open: open && !isDisabled, zIndex, rootClassName: styles.modal, cancelText: "取消",
                    style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                    okButtonProps: {loading: selecting, disabled: selecting || (!multi && (!loaded || (mode === "file" && !selected?.key)))},
                    focusable: {focusTriggerAfterClose: false}, mask: {closable: true},
                    styles: {body: {paddingTop: compact ? 10 : 15}}}}>
                <Form component={false} size={compact ? "small" : "middle"} variant="outlined">
                    <div className={styles.browser}>
                        <div className="path-select-toolbar">
                            <Tooltip title="上一级"><Button type="text" className="path-select-up" aria-label="返回上一级"
                                disabled={!prefix || loading || selecting} icon={<Icon type="ArrowLeftOutlined"/>} onClick={() => navigate(parentPrefix(prefix))}/></Tooltip>
                            {editing ? <Input ref={pathInput} value={draft} className="path-select-path-input" placeholder="/ 表示存储桶根目录"
                                onChange={(event) => setDraft(event.target.value)} onBlur={() => setEditing(false)}
                                onKeyDown={(event) => {
                                    if (event.nativeEvent.isComposing || event.nativeEvent.keyCode === 229) return;
                                    if (event.key === "Escape") { event.stopPropagation(); setEditing(false); }
                                    if (event.key === "Enter") { event.preventDefault(); navigate(draft.startsWith("/") ? draft.slice(1) : draft); }
                                }}/>
                                : <ol ref={breadcrumbRef} className="path-select-path" aria-label="当前 OSS 目录">
                                    {breadcrumbs.map((item, index) => <li key={item.key}>
                                        {index > 0 && <Icon type="RightOutlined" className="path-select-path-separator"/>}
                                        <button type="button" disabled={loading || selecting}
                                            className={index === breadcrumbs.length - 1 ? "path-select-path-current" : "path-select-path-button"}
                                            aria-current={index === breadcrumbs.length - 1 ? "page" : undefined}
                                            title={displayPath(item.key)} onClick={() => {
                                                if (index < breadcrumbs.length - 1) navigate(item.key);
                                                else { setDraft(displayPath(prefix)); setEditing(true); }
                                            }}><span>{item.label}</span></button>
                                    </li>)}
                                </ol>}
                            <div className="path-select-actions">
                                <Button type="text" className="path-select-up" aria-label="新建文件夹"
                                    disabled={!loaded || saving || selecting} icon={<Icon type="PlusOutlined"/>} onClick={openFolder}/>
                                <Button type="text" className="path-select-up" aria-label="刷新"
                                    loading={loading} disabled={selecting} icon={<Icon type="ReloadOutlined"/>} onClick={refresh}/>
                            </div>
                        </div>
                        <Table<any> ref={table} className="path-select-table" size="small" rowKey="key" rowHoverable={false} tableLayout="auto" loading={loading}
                            columns={columns} dataSource={rows} pagination={false}
                            scroll={{x: "max-content", y: compact ? "clamp(180px, 34dvh, 238px)" : "clamp(200px, 38dvh, 278px)"}}
                            locale={{emptyText: error ? <div className="path-select-error">
                                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={error}/><Button size="small" onClick={refresh}>重试</Button>
                            </div> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
                            rowSelection={multi ? {
                                selectedRowKeys: [...checkedKeys],
                                preserveSelectedRowKeys: true,
                                columnTitle: mode === "directory" ? <Checkbox aria-label="选择当前页" checked={allChecked} indeterminate={someChecked}
                                    disabled={loading || selecting || selectableRows.length === 0}
                                    onChange={(event) => { void updateSelection(selectableRows, event.target.checked); }}/> : undefined,
                                getCheckboxProps: (item) => ({disabled: loading || selecting || Boolean(item.directory) !== (mode === "directory"),
                                    indeterminate: partialKeys.has(item.key),
                                }),
                                onSelect: (item, checked) => { void updateSelection([item], checked); },
                                onSelectAll: (checked, _, changed) => { void updateSelection(changed, checked); },
                                onSelectMultiple: (checked, _, changed) => { void updateSelection(changed, checked); },
                                onCell: () => ({onClick: (event) => event.stopPropagation(), onKeyDown: (event) => event.stopPropagation()}),
                            } : undefined}
                            rowClassName={(item) => [item?.directory ? "path-select-directory" : "", item?.directory || mode === "file" ? "path-select-selectable" : "",
                                multi ? checkedKeys.has(item.key) ? "path-select-selected" : "" : !item?.directory && selected?.key === item.key ? "path-select-selected" : ""].filter(Boolean).join(" ")}
                            onRow={(item) => loading || selecting ? {} : item?.directory ? {
                                tabIndex: 0, role: "button", "aria-label": `进入 ${item.key}`,
                                onClick: () => navigate(item.key),
                                onKeyDown: (event) => {
                                    if (event.key === "Enter" || event.key === " ") { event.preventDefault(); navigate(item.key); }
                                },
                            } : mode === "file" ? {
                                tabIndex: 0, role: "button", "aria-label": `选择 ${item.key}`,
                                "aria-pressed": multi ? selectedKeys.has(item.key) : selected?.key === item.key,
                                onClick: () => selectItem(item), onDoubleClick: multi ? undefined : () => confirm(item),
                                onKeyDown: (event) => {
                                    if (event.key === "Enter" || event.key === " ") { event.preventDefault(); selectItem(item); }
                                },
                            } : {}}/>
                        <div className="path-select-status">
                            <div className="path-select-status-main"><Icon className="path-select-status-icon" type={mode === "directory" ? "FolderOpenOutlined" : "FileOutlined"}/>
                                <span className="path-select-status-label">{!multi && mode === "directory" ? "当前目录" : "选择结果"}</span>
                                <Typography.Text className="path-select-status-value" ellipsis={{tooltip: !multi && (mode === "directory" ? displayPath(prefix) : selected?.key)}}>
                                    {multi ? `已选择 ${selectedItems.length} 个${mode === "directory" ? "目录" : "文件"}`
                                        : mode === "directory" ? displayPath(prefix) : selected?.key ? displayPath(selected.key) : "未选择"}
                                </Typography.Text>
                                {multi && mode === "directory" && <Checkbox checked={directoryCovered} disabled={!loaded || selecting}
                                    indeterminate={!directoryCovered && [...selectedKeys].some((key) => key.startsWith(prefix))}
                                    onChange={(event) => { void updateSelection([{key: prefix, directory: true}], event.target.checked); }}>当前目录</Checkbox>}
                            </div>
                            <div className="path-select-status-meta">
                                <span className="path-select-count">{loading && loadedPath === undefined ? "加载中" : `${rows.length} 项`}</span>
                                {(cursor > 0 || nextToken) && <div className="path-select-pagination">
                                    <Button type="text" className="path-select-page-button" aria-label="上一页" icon={<Icon type="LeftOutlined"/>} disabled={loading || selecting || cursor === 0}
                                        onClick={() => {
                                            request.current++;
                                            setLoading(true);
                                            setQuery((current) => ({...current, cursor: current.cursor - 1}));
                                            table.current?.scrollTo({top: 0});
                                        }}/>
                                    <span className="path-select-page-number">{cursor + 1}</span>
                                    <Button type="text" className="path-select-page-button" aria-label="下一页" icon={<Icon type="RightOutlined"/>} disabled={loading || selecting || !nextToken}
                                        onClick={() => {
                                            request.current++;
                                            setLoading(true);
                                            setQuery((current) => ({...current, cursor: current.cursor + 1,
                                                tokens: [...current.tokens.slice(0, current.cursor + 1), nextToken]}));
                                            table.current?.scrollTo({top: 0});
                                        }}/>
                                </div>}
                            </div>
                        </div>
                    </div>
                </Form>
            </ProModal>
        </ConfigProvider>
        <ProModal title={<Title>新建文件夹</Title>} width="340px" onOk={folderForm.submit} onCancel={() => {
            if (!folderPending.current) { folderRequest.current++; setCreating(false); }
        }} modalProps={{open: creating && open, zIndex: Math.max(2000, (zIndex || 0) + 10), okText: "创建", cancelText: "取消",
            style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
            styles: {body: {paddingTop: compact ? 10 : 15}},
            confirmLoading: saving, closable: !saving, keyboard: !saving, mask: {closable: !saving}, cancelButtonProps: {disabled: saving},
            forceRender: true, focusable: {focusTriggerAfterClose: true}}}>
            <Form form={folderForm} layout="vertical" variant="filled" preserve={false} requiredMark={false} onFinish={createFolder} disabled={saving}>
                <Form.Item name="name" rules={[{required: true, whitespace: true, message: "请输入文件夹名称"},
                    {max: 255, message: "名称不能超过255个字符"}, {validator: async (_, name) => {
                        if (name === "." || name === ".." || /[\\/\0\r\n]/.test(String(name || ""))) throw new Error("请输入单个文件夹名称");
                    }}]}><Input aria-label="请输入文件夹名称" placeholder="请输入文件夹名称" maxLength={255}/></Form.Item>
            </Form>
        </ProModal>
    </>;
});

OssPathPicker.displayName = "OssPathPicker";

export default OssPathPicker;
