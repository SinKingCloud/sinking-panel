import React, {forwardRef, memo, useCallback, useContext, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, Form, Grid, Input, Select, Space, Tooltip} from "antd";
import type {InputRef, TableRef} from "antd";
import DisabledContext from "antd/es/config-provider/DisabledContext";
import useVariant from "antd/es/form/hooks/useVariants";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {getFileDisks, getFileInfo, getFileList} from "@/service/api/file";
import {readFileStorage, updateFileStorage} from "@/pages/file/hooks/file-storage";
import {
    buildFileBreadcrumbs,
    comparableFilePath as comparablePath,
    getFileDiskName,
    getFileNavigation,
    isAbsoluteFilePath,
    isWindowsFilePath,
    normalizeFilePath,
    parentFilePath,
} from "@/pages/file/utils";
import FilePickerBrowser from "./browser";
import useStyles from "./styles";

// FilePicker 默认返回单个路径；文件多选时使用 string[]，浏览目录和分页不会丢失选择。
const FilePicker = forwardRef<InputRef, any>(({
    buttonProps,
    disabled,
    initialPath,
    mode = "directory",
    multiple = false,
    onChange,
    onSelect,
    placeholder,
    size,
    value = "",
    ...inputProps
}, ref) => {
    const theme = useTheme();
    const {message} = App.useApp();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const contextDisabled = useContext(DisabledContext);
    const isDisabled = disabled ?? contextDisabled;
    const multi = mode === "file" && multiple;
    const values: string[] = multi && Array.isArray(value) ? value : [];
    // 与 Input 一样继承 Form 和 ConfigProvider 的外观，按钮可通过 buttonProps 单独覆盖。
    const [variant] = useVariant("input", inputProps.variant, inputProps.bordered);
    const requestRef = useRef(0);
    const selectionRequestRef = useRef(0);
    const restoredPathRef = useRef("");
    const pathRef = useRef<HTMLOListElement>(null);
    const pathInputRef = useRef<InputRef>(null);
    const currentPathRef = useRef<HTMLButtonElement>(null);
    const tableRef = useRef<TableRef>(null);
    const [open, setOpen] = useState(false);
    const [path, setPath] = useState("");
    const [loadedPath, setLoadedPath] = useState("");
    const [selectedFile, setSelectedFile] = useState<any>();
    const [selectedPaths, setSelectedPaths] = useState<string[]>([]);
    const [searchValue, setSearchValue] = useState("");
    const [disks, setDisks] = useState<string[]>([]);
    const [disksLoading, setDisksLoading] = useState(false);
    const [disksError, setDisksError] = useState("");
    const [diskReload, setDiskReload] = useState(0);
    const [items, setItems] = useState<any[]>([]);
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");
    const [reload, setReload] = useState(0);
    const [editingPath, setEditingPath] = useState(false);
    const [pathDraft, setPathDraft] = useState("");

    useEffect(() => () => {
        requestRef.current += 1;
        selectionRequestRef.current += 1;
    }, []);

    const title = mode === "directory" ? "选择目录" : "选择文件";
    const showingDisks = !path || (path === "/" && !disks.includes("/"))
        || (disks.some(isWindowsFilePath) && !isWindowsFilePath(path));
    const browserLoading = showingDisks ? disksLoading : loading;
    const browserError = showingDisks ? disksError : error;
    const diskItems = useMemo(() => disks.map((path) => ({path, name: getFileDiskName(path), is_dir: true})), [disks]);
    const {breadcrumbs: pathBreadcrumbs, parentPath} = useMemo(
        () => getFileNavigation(showingDisks ? "" : path, disks), [disks, path, showingDisks],
    );

    const navigate = useCallback((nextPath: string) => {
        const targetPath = nextPath ? normalizeFilePath(nextPath) : "";
        requestRef.current += 1;
        selectionRequestRef.current += 1;
        setLoading(Boolean(targetPath));
        setEditingPath(false);
        setPathDraft(targetPath);
        setPath(targetPath);
        setLoadedPath("");
        setSelectedFile(undefined);
        setItems([]);
        setTotal(0);
        setPage(1);
        setError("");
    }, []);

    const show = useCallback(() => {
        if (isDisabled) return;
        const input = multi ? String(value?.[0] || "") : String(value || "").trim();
        const absolute = isAbsoluteFilePath(input) ? normalizeFilePath(input) : "";
        const preferred = String(initialPath || "").trim();
        const preferredPath = isAbsoluteFilePath(preferred) ? normalizeFilePath(preferred) : "";
        let initialDirectory = preferredPath || (mode === "file" && absolute ? parentFilePath(absolute) : absolute);
        restoredPathRef.current = "";
        if (!initialDirectory) {
            try {
                const saved = readFileStorage()?.pickerPaths?.[mode];
                if (typeof saved === "string" && isAbsoluteFilePath(saved)) {
                    initialDirectory = normalizeFilePath(saved);
                    restoredPathRef.current = initialDirectory;
                }
            } catch {
                // 本地存储不可用时，仍可正常浏览和选择。
            }
        }
        const breadcrumbs = absolute ? buildFileBreadcrumbs(absolute) : [];
        const initialName = breadcrumbs.length > 1 ? breadcrumbs[breadcrumbs.length - 1].label : "";
        const selectionRequest = ++selectionRequestRef.current;
        requestRef.current += 1;
        setPath(initialDirectory);
        setLoadedPath("");
        setSelectedFile(undefined);
        setSelectedPaths(multi && Array.isArray(value) ? value.slice() : []);
        setDisks([]);
        setDisksLoading(true);
        setDisksError("");
        setItems([]);
        setTotal(0);
        setPage(1);
        setError("");
        setLoading(true);
        setEditingPath(false);
        setPathDraft(initialDirectory);
        setOpen(true);
        if (!multi && mode === "file" && absolute && initialName) {
            void getFileInfo({body: {path: absolute}}).then((response) => {
                if (selectionRequestRef.current === selectionRequest
                    && response?.code === 200 && response?.data && !response?.data?.is_dir) {
                    setSelectedFile({name: response?.data?.name || initialName, path: absolute});
                }
            }).catch(() => undefined);
        }
    }, [initialPath, isDisabled, mode, multi, value]);

    const close = useCallback(() => {
        requestRef.current += 1;
        selectionRequestRef.current += 1;
        setLoading(false);
        setEditingPath(false);
        setPathDraft(path);
        setOpen(false);
    }, [path]);

    const confirm = useCallback((file = selectedFile) => {
        if (isDisabled) return;
        if (!multi && (showingDisks || loading || error || !loadedPath || comparablePath(loadedPath) !== comparablePath(path))) {
            return;
        }
        const selectedPath = mode === "directory" ? loadedPath : file?.path;
        if (!multi && !selectedPath) {
            return;
        }
        if (multi && selectedPaths.some((item) => !isAbsoluteFilePath(item))) {
            message.warning("请输入文件的绝对路径");
            return;
        }
        const normalized = multi ? selectedPaths.map(normalizeFilePath) : normalizeFilePath(selectedPath);
        if (onSelect) {
            if (onSelect(normalized) === false) {
                return;
            }
        } else {
            onChange?.(normalized);
        }
        close();
    }, [close, error, isDisabled, loadedPath, loading, message, mode, multi, onChange, onSelect, path, selectedFile, selectedPaths, showingDisks]);

    const selectFile = useCallback((file: any) => {
        selectionRequestRef.current += 1;
        if (multi) {
            const key = comparablePath(file.path);
            setSelectedPaths((current) => current.some((item) => comparablePath(item) === key)
                ? current.filter((item) => comparablePath(item) !== key) : [...current, file.path]);
        } else setSelectedFile(file);
    }, [multi]);

    // 手输路径按原样加入，避免 tags 模式自动删去文件名末尾的空格。
    const changePaths = (paths: string[]) => {
        const unique = new Map(paths.map((item) => [isAbsoluteFilePath(item) ? comparablePath(item) : item, item]));
        onChange?.([...unique.values()]);
    };
    const addPath = () => {
        if (searchValue) changePaths([...values, searchValue]);
        setSearchValue("");
    };

    const changePage = useCallback((nextPage: number) => {
        setLoading(true);
        setPage(nextPage);
        window.requestAnimationFrame(() => {
            tableRef.current?.scrollTo({top: 0});
        });
    }, []);

    const openPathEditor = useCallback(() => {
        if (browserLoading) {
            return;
        }
        setPathDraft(path);
        setEditingPath(true);
    }, [browserLoading, path]);

    const cancelPathEditor = useCallback((restoreFocus = false) => {
        setPathDraft(path);
        setEditingPath(false);
        if (restoreFocus) {
            window.requestAnimationFrame(() => currentPathRef.current?.focus());
        }
    }, [path]);

    const submitPathEditor = useCallback(() => {
        const value = pathDraft.trim();
        if (!value) {
            cancelPathEditor(true);
            return;
        }
        if (!isAbsoluteFilePath(value)) {
            message.warning("请输入绝对路径");
            return;
        }
        const nextPath = normalizeFilePath(value);
        if (path && comparablePath(nextPath) === comparablePath(path)) {
            cancelPathEditor(true);
            return;
        }
        navigate(nextPath);
    }, [cancelPathEditor, message, navigate, path, pathDraft]);

    const reloadList = useCallback(() => {
        if (showingDisks) setDiskReload((current) => current + 1);
        else setReload((current) => current + 1);
    }, [showingDisks]);

    useEffect(() => {
        if (!open) {
            return;
        }
        let active = true;
        setDisksLoading(true);
        setDisksError("");
        void getFileDisks().then((response) => {
            if (!active) {
                return;
            }
            if (response?.code !== 200 || !Array.isArray(response?.data)) {
                setDisks([]);
                setDisksError(response?.message || "获取磁盘列表失败");
                return;
            }
            const nextDisks = Array.from(new Set(response?.data?.filter((item) => typeof item === "string" && item).map(normalizeFilePath)));
            setDisks(nextDisks);
            if (nextDisks.includes("/")) setPath((current) => current || "/");
        }).catch(() => {
            if (active) {
                setDisks([]);
                setDisksError("获取磁盘列表失败");
            }
        }).finally(() => {
            if (active) setDisksLoading(false);
        });
        return () => {
            active = false;
        };
    }, [diskReload, open]);

    useLayoutEffect(() => {
        if (open && !editingPath && pathRef.current) {
            pathRef.current.scrollLeft = pathRef.current.scrollWidth;
        }
    }, [editingPath, open, pathBreadcrumbs]);

    useEffect(() => {
        setEditingPath(false);
        setPathDraft(path);
    }, [path]);

    useLayoutEffect(() => {
        if (!editingPath) {
            return;
        }
        pathInputRef.current?.focus();
        pathInputRef.current?.select();
    }, [editingPath]);

    useEffect(() => {
        if (browserLoading && editingPath) {
            setEditingPath(false);
            setPathDraft(path);
        }
    }, [browserLoading, editingPath, path]);

    useEffect(() => {
        if (!open || showingDisks) {
            return;
        }
        const requestId = ++requestRef.current;
        setLoading(true);
        setError("");
        void getFileList({
            body: {
                path,
                page,
                page_size: 50,
                order_by_field: "name",
                order_by_type: "asc",
            },
        }).then((response) => {
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                setItems([]);
                setTotal(0);
                setLoadedPath("");
                setError(response?.message || "获取文件列表失败");
                return;
            }
            const data = response?.data;
            const nextItems = Array.isArray(data?.list) ? data?.list.filter((item: any) => item?.name) : [];
            const nextTotal = Math.max(0, Number(data?.total) || 0);
            const lastPage = Math.max(1, Math.ceil(nextTotal / 50));
            setTotal(nextTotal);
            if (page > lastPage) {
                setItems([]);
                setPage(lastPage);
                return;
            }
            setItems(nextItems);
            setLoadedPath(path);
            restoredPathRef.current = "";
            try {
                if (readFileStorage()?.pickerPaths?.[mode] !== path) {
                    updateFileStorage((current) => ({
                        ...current,
                        pickerPaths: {...current.pickerPaths, [mode]: path},
                    }));
                }
            } catch {
                // 记忆路径失败不影响文件选择。
            }
        }).catch(() => {
            if (requestRef.current === requestId) {
                setItems([]);
                setTotal(0);
                setLoadedPath("");
                setError("获取文件列表失败");
            }
        }).finally(() => {
            if (requestRef.current === requestId) {
                setLoading(false);
            }
        });
        return () => {
            if (requestRef.current === requestId) requestRef.current += 1;
        };
    }, [mode, open, page, path, reload, showingDisks]);

    useEffect(() => {
        if (!open || !error || !disks.length || restoredPathRef.current !== path) {
            return;
        }
        // 记忆目录失效时回到浏览入口，手动输入的错误路径仍保留错误提示。
        restoredPathRef.current = "";
        const fallback = disks.includes("/") ? "/" : "";
        if (!fallback || comparablePath(fallback) !== comparablePath(path)) {
            navigate(fallback);
        }
    }, [disks, error, navigate, open, path]);

    return (
        <>
            <Space.Compact className={styles.field} block>
                {multi ? <Select
                    {...inputProps}
                    ref={ref as any}
                    size={size}
                    mode="multiple"
                    value={values}
                    disabled={isDisabled}
                    variant={variant}
                    allowClear
                    open={false}
                    suffixIcon={null}
                    maxTagCount="responsive"
                    searchValue={searchValue}
                    onSearch={setSearchValue}
                    placeholder={placeholder || "请输入或选择文件"}
                    onInputKeyDown={(event) => {
                        if (event.key === "Enter" && !event.nativeEvent.isComposing) {
                            event.preventDefault(); event.stopPropagation(); addPath();
                        }
                    }}
                    onBlur={(event) => { addPath(); inputProps.onBlur?.(event); }}
                    onClear={() => setSearchValue("")}
                    onChange={changePaths}/> : <Input
                    ref={ref}
                    {...inputProps}
                    size={size}
                    value={value}
                    disabled={isDisabled}
                    placeholder={placeholder || (mode === "directory" ? "请输入或选择目录" : "请输入或选择文件")}
                    onChange={(event) => onChange?.(event.target.value)}/>}
                <Tooltip title={title}>
                    <Button
                        color={buttonProps?.type || buttonProps?.danger ? undefined : "default"}
                        variant={buttonProps?.type || buttonProps?.danger ? undefined : variant === "filled" ? "filled" : variant === "outlined" ? "outlined" : "text"}
                        {...buttonProps}
                        size={size}
                        disabled={isDisabled}
                        aria-label={title}
                        icon={<Icon type={mode === "directory" ? "FolderOpenOutlined" : "FileSearchOutlined"}/>}
                        onClick={show}/>
                </Tooltip>
            </Space.Compact>

            <ConfigProvider componentSize="middle" variant="outlined">
                <ProModal
                    title={<Title>{title}</Title>}
                    width="640px"
                    okText="选择"
                    onOk={() => confirm()}
                    onCancel={close}
                    modalProps={{
                        open: open && !isDisabled,
                        rootClassName: styles.modal,
                        style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                        cancelText: "取消",
                        okButtonProps: {
                            disabled: !multi && (showingDisks || loading || Boolean(error) || (mode === "directory"
                                ? !loadedPath || comparablePath(loadedPath) !== comparablePath(path)
                                : !selectedFile)),
                        },
                        focusable: {focusTriggerAfterClose: false},
                        mask: {closable: true},
                        styles: {body: {paddingTop: compact ? 10 : 15}},
                    }}>
                    <Form component={false} size={compact ? "small" : "middle"} variant="outlined">
                        <FilePickerBrowser
                            className={styles.browser}
                            compact={compact}
                            mode={mode}
                            path={path}
                            loadedPath={loadedPath}
                            selectedFile={selectedFile}
                            multiple={multi}
                            selectedPaths={selectedPaths}
                            showingDisks={showingDisks}
                            pathBreadcrumbs={pathBreadcrumbs}
                            parentPath={parentPath}
                            editingPath={editingPath}
                            pathDraft={pathDraft}
                            loading={browserLoading}
                            error={browserError}
                            items={showingDisks ? diskItems : items}
                            page={page}
                            pageSize={50}
                            total={showingDisks ? diskItems.length : total}
                            pathRef={pathRef}
                            pathInputRef={pathInputRef}
                            currentPathRef={currentPathRef}
                            tableRef={tableRef}
                            onNavigate={navigate}
                            onOpenPathEditor={openPathEditor}
                            onCancelPathEditor={cancelPathEditor}
                            onSubmitPathEditor={submitPathEditor}
                            onPathDraftChange={setPathDraft}
                            onSelectFile={selectFile}
                            onSelectionChange={setSelectedPaths}
                            onConfirm={confirm}
                            onPageChange={changePage}
                            onReload={reloadList}/>
                    </Form>
                </ProModal>
            </ConfigProvider>
        </>
    );
});

FilePicker.displayName = "FilePicker";

export default memo(FilePicker);
