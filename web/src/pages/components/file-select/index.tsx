import React, {forwardRef, memo, useCallback, useContext, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, Form, Grid, Input, Select, Space, Tooltip} from "antd";
import type {ButtonProps, InputProps, InputRef, TableRef} from "antd";
import DisabledContext from "antd/es/config-provider/DisabledContext";
import useVariant from "antd/es/form/hooks/useVariants";
import {useZIndex} from "antd/es/_util/hooks/useZIndex";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {createFile, getFileDisks, getFileInfo, getFileList} from "@/service/api/file";
import {readFileStorage, updateFileStorage} from "@/pages/file/hooks/file-storage";
import FileUpload from "@/pages/file/components/upload";
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
import FileSelectBrowser from "./browser";
import useFileSelection from "./selection";
import useStyles from "@/pages/components/ui/path-select/styles";

export interface FileSelectProps extends Omit<InputProps, "value" | "defaultValue" | "onChange" | "onSelect" | "multiple" | "type"> {
    buttonProps?: ButtonProps;
    initialPath?: string;
    zIndex?: number;
    /** 不传时允许选择文件和目录。 */
    type?: "file" | "directory";
    /** 设置后仅允许选择指定后缀的文件，如 ["png", ".tar.gz"]。 */
    extensions?: readonly string[];
    renderTrigger?: (open: () => void) => React.ReactNode;
    /** 文件、目录及混选均支持多选。 */
    multiple?: boolean;
    value?: string[];
    defaultValue?: string[];
    onChange?: (paths: string[]) => void;
    /** 弹窗确认前校验，返回 false 保持弹窗；通过后触发 onChange。 */
    onSelect?: (paths: string[]) => boolean | void;
}

// 单选、多选均使用路径数组，单选输入框仅显示第一个路径。
const FileSelect = forwardRef<InputRef, FileSelectProps>((props, ref) => {
    const {
        buttonProps,
        defaultValue = [],
        disabled,
        initialPath,
        zIndex,
        type: selectionType,
        extensions,
        renderTrigger,
        multiple = false,
        onChange,
        onSelect,
        placeholder,
        size,
        value,
        ...inputProps
    } = props;
    const suffixes = useMemo(() => selectionType === "directory" ? [] : [...new Set(
        (extensions || []).map((ext) => ext.trim().toLowerCase().replace(/^\.+/, "")).filter(Boolean).map((ext) => "." + ext),
    )], [extensions, selectionType]);
    const type = selectionType ?? (suffixes.length ? "file" : undefined);
    const acceptsFile = useCallback((path: string) => {
        if (!suffixes.length) return true;
        const name = path.slice(Math.max(path.lastIndexOf("/"), path.lastIndexOf("\\")) + 1).toLowerCase();
        return suffixes.some((suffix) => name.length > suffix.length && name.endsWith(suffix));
    }, [suffixes]);
    const theme = useTheme();
    const [, pickerZIndex] = useZIndex("Modal", zIndex);
    const {message} = App.useApp();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const contextDisabled = useContext(DisabledContext);
    const isDisabled = disabled ?? contextDisabled;
    const controlled = Object.prototype.hasOwnProperty.call(props, "value");
    const [localValue, setLocalValue] = useState<string[]>(defaultValue);
    const values = controlled ? value || [] : localValue;
    // 与 Input 一样继承 Form 和 ConfigProvider 的外观，按钮可通过 buttonProps 单独覆盖。
    const [variant] = useVariant("input", inputProps.variant, inputProps.bordered);
    const requestRef = useRef(0);
    const pickerSessionRef = useRef(0);
    const visiblePathRef = useRef("");
    const uploadRef = useRef<{open: () => void}>(null);
    const selectionRequestRef = useRef(0);
    const creationRequestRef = useRef(0);
    const creationOpenRef = useRef(false);
    const creationSubmittingRef = useRef(false);
    const creationFormRef = useRef<React.ComponentRef<typeof Form>>(null);
    const [creationForm] = Form.useForm();
    const restoredPathRef = useRef("");
    const pathRef = useRef<HTMLOListElement>(null);
    const pathInputRef = useRef<InputRef>(null);
    const currentPathRef = useRef<HTMLButtonElement>(null);
    const tableRef = useRef<TableRef>(null);
    const [open, setOpen] = useState(false);
    const [path, setPath] = useState("");
    const [loadedPath, setLoadedPath] = useState("");
    const {
        paths: selectedPaths,
        pending: selectionPending,
        select: selectPaths,
        compact: compactPaths,
        reset: resetSelection,
        cancel: cancelSelection,
        remember: rememberDirectory,
        forget: forgetDirectory,
    } = useFileSelection({type, multiple, disabled: isDisabled, onError: (error) => message.error(error)});
    const [searchValue, setSearchValue] = useState("");
    const [inputDraft, setInputDraft] = useState<string | null>(null);
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
    const [creationOpen, setCreationOpen] = useState(false);
    const [creationPath, setCreationPath] = useState("");
    const [creationType, setCreationType] = useState<"directory" | "file">("directory");
    const [creationSubmitting, setCreationSubmitting] = useState(false);
    const [uploadSession, setUploadSession] = useState<{path: string; session: number} | null>(null);
    visiblePathRef.current = path;
    const actionZIndex = pickerZIndex + 100;
    const creationLabel = creationType === "directory" ? "文件夹" : "文件";

    useEffect(() => () => {
        requestRef.current += 1;
        pickerSessionRef.current += 1;
        selectionRequestRef.current += 1;
        creationRequestRef.current += 1;
        creationOpenRef.current = false;
    }, []);

    const selectionLabel = type === "directory" ? "目录" : type === "file" ? "文件" : "文件或目录";
    const storageKey = type || "all";
    const title = `选择${selectionLabel}`;
    const showingDisks = !path || (path === "/" && !disks.includes("/"))
        || (disks.some(isWindowsFilePath) && !isWindowsFilePath(path));
    const browserLoading = selectionPending || (showingDisks ? disksLoading : loading);
    const browserError = showingDisks ? disksError : error;
    const diskItems = useMemo(() => disks.map((path) => ({path, name: getFileDiskName(path), is_dir: true})), [disks]);
    const {breadcrumbs: pathBreadcrumbs, parentPath} = useMemo(
        () => getFileNavigation(showingDisks ? "" : path, disks), [disks, path, showingDisks],
    );

    const navigate = useCallback((nextPath: string) => {
        if (selectionPending) return;
        const targetPath = nextPath ? normalizeFilePath(nextPath) : "";
        visiblePathRef.current = targetPath;
        requestRef.current += 1;
        selectionRequestRef.current += 1;
        setLoading(Boolean(targetPath));
        setEditingPath(false);
        setPathDraft(targetPath);
        setPath(targetPath);
        setLoadedPath("");
        if (!multiple) resetSelection();
        setItems([]);
        setTotal(0);
        setPage(1);
        setError("");
    }, [multiple, resetSelection, selectionPending]);

    const show = useCallback(() => {
        if (isDisabled) return;
        const input = values[0] || "";
        const absolute = isAbsoluteFilePath(input) ? normalizeFilePath(input) : "";
        const preferred = String(initialPath || "").trim();
        const preferredPath = isAbsoluteFilePath(preferred) ? normalizeFilePath(preferred) : "";
        let initialDirectory = preferredPath || (type !== "directory" && absolute ? parentFilePath(absolute) : absolute);
        restoredPathRef.current = "";
        if (!initialDirectory) {
            try {
                const saved = readFileStorage()?.pickerPaths?.[storageKey];
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
        pickerSessionRef.current += 1;
        visiblePathRef.current = initialDirectory;
        setPath(initialDirectory);
        setLoadedPath("");
        resetSelection(multiple ? values.filter(acceptsFile) : []);
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
        if (!multiple && type !== "directory" && absolute && initialName) {
            void getFileInfo({body: {path: absolute}}).then((response) => {
                if (selectionRequestRef.current === selectionRequest
                    && response?.code === 200 && response?.data && (type === undefined || !response?.data?.is_dir)
                    && (response.data.is_dir || acceptsFile(absolute))) {
                    if (response.data.is_dir && !preferredPath) navigate(absolute);
                    void selectPaths([absolute], true);
                }
            }).catch(() => undefined);
        }
    }, [acceptsFile, initialPath, isDisabled, type, multiple, navigate, resetSelection, selectPaths, storageKey, values]);

    const close = useCallback(() => {
        requestRef.current += 1;
        pickerSessionRef.current += 1;
        selectionRequestRef.current += 1;
        cancelSelection();
        creationRequestRef.current += 1;
        creationOpenRef.current = false;
        creationSubmittingRef.current = false;
        setCreationOpen(false);
        setCreationSubmitting(false);
        // 卸载上传队列会中止当前传输；下一次打开选择器使用新的会话。
        setUploadSession(null);
        setLoading(false);
        setEditingPath(false);
        setOpen(false);
    }, [cancelSelection]);

    const validExtensions = useCallback((paths: string[]) => {
        if (!suffixes.length || paths.every(acceptsFile)) return true;
        message.warning(`请选择以下格式的文件：${suffixes.join("、")}`);
        return false;
    }, [acceptsFile, message, suffixes]);

    const confirm = useCallback((filePath?: string) => {
        if (isDisabled || selectionPending) return;
        if (!multiple && (showingDisks || loading || error || !loadedPath || comparablePath(loadedPath) !== comparablePath(path))) return;
        const selectedPath = filePath || selectedPaths[0] || (type !== "file" ? loadedPath : undefined);
        if (!multiple && !selectedPath) return;
        const paths = multiple ? selectedPaths : [selectedPath!];
        if (paths.some((item) => !isAbsoluteFilePath(item))) {
            message.warning(`请输入${selectionLabel}的绝对路径`);
            return;
        }
        const normalized = compactPaths(paths).map(normalizeFilePath);
        if (!validExtensions(normalized) || onSelect?.(normalized) === false) return;
        onChange?.(normalized);
        if (!controlled) setLocalValue(normalized);
        close();
    }, [close, compactPaths, controlled, error, isDisabled, loadedPath, loading, message, type, multiple, onChange, onSelect, path, selectedPaths, selectionLabel, selectionPending, showingDisks, validExtensions]);

    useEffect(() => { close(); }, [close, isDisabled, multiple, type]);

    const changeSelection = useCallback((paths: string[], checked: boolean) => {
        selectionRequestRef.current += 1;
        void selectPaths(paths, checked);
    }, [selectPaths]);

    // 手输路径按原样加入，避免 tags 模式自动删去文件名末尾的空格。
    const changePaths = (paths: string[]) => {
        const nextValue = compactPaths(paths);
        if (!validExtensions(nextValue)) return;
        if (!controlled) setLocalValue(nextValue);
        onChange?.(nextValue);
    };
    const addPath = () => {
        if (searchValue) changePaths([...values, searchValue]);
        setSearchValue("");
    };
    const commitInput = () => {
        if (inputDraft === null) return;
        changePaths(inputDraft ? [inputDraft] : []);
        setInputDraft(null);
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
        if (selectionPending) return;
        if (showingDisks) setDiskReload((current) => current + 1);
        else {
            forgetDirectory(path);
            setReload((current) => current + 1);
        }
    }, [forgetDirectory, path, selectionPending, showingDisks]);

    const refreshDirectory = useCallback((targetPath: string) => {
        forgetDirectory(targetPath);
        if (comparablePath(visiblePathRef.current) !== comparablePath(targetPath)) return;
        requestRef.current += 1;
        setLoading(true);
        setReload((current) => current + 1);
    }, [forgetDirectory]);

    const openUpload = useCallback(() => {
        if (isDisabled || showingDisks || browserLoading || browserError || !loadedPath
            || comparablePath(loadedPath) !== comparablePath(path)) return;
        // 仅在打开上传窗口时更新目标，浏览目录变化不会修改已打开窗口的目标。
        setUploadSession({
            path: loadedPath,
            session: pickerSessionRef.current,
        });
    }, [browserError, browserLoading, isDisabled, loadedPath, path, showingDisks]);

    useEffect(() => {
        if (uploadSession && open) uploadRef.current?.open();
    }, [open, uploadSession]);

    const openCreationForm = useCallback((type: "directory" | "file") => {
        if (isDisabled || showingDisks || browserLoading || browserError || !loadedPath
            || comparablePath(loadedPath) !== comparablePath(path)) return;
        creationRequestRef.current += 1;
        creationOpenRef.current = true;
        creationFormRef.current?.resetFields();
        setCreationType(type);
        setCreationPath(loadedPath);
        setCreationOpen(true);
    }, [browserError, browserLoading, isDisabled, loadedPath, path, showingDisks]);

    const closeCreationForm = useCallback(() => {
        creationRequestRef.current += 1;
        creationOpenRef.current = false;
        creationSubmittingRef.current = false;
        setCreationOpen(false);
        setCreationSubmitting(false);
    }, []);

    const createEntry = async ({name}: {name: string}) => {
        if (isDisabled || !creationOpenRef.current || creationSubmittingRef.current || !creationPath) return;
        const request = creationRequestRef.current;
        creationSubmittingRef.current = true;
        setCreationSubmitting(true);
        try {
            const response = await createFile({body: {
                path: creationPath,
                name: `${name.trim()}${creationType === "directory" ? "/" : ""}`,
            }});
            if (creationRequestRef.current !== request) return;
            if (response?.code !== 200) {
                message.error(response?.message || `新建${creationLabel}失败`);
                return;
            }
            message.success(response?.message || "创建成功");
            closeCreationForm();
            refreshDirectory(creationPath);
        } catch (error: any) {
            if (creationRequestRef.current === request) message.error(error?.message || `新建${creationLabel}失败`);
        } finally {
            if (creationRequestRef.current === request) {
                creationSubmittingRef.current = false;
                setCreationSubmitting(false);
            }
        }
    };

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
            if (page === 1 && nextItems.length === nextTotal) rememberDirectory(path, nextItems);
            restoredPathRef.current = "";
            try {
                if (readFileStorage()?.pickerPaths?.[storageKey] !== path) {
                    updateFileStorage((current) => ({
                        ...current,
                        pickerPaths: {...current.pickerPaths, [storageKey]: path},
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
    }, [open, page, path, reload, rememberDirectory, showingDisks, storageKey]);

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
            {renderTrigger ? renderTrigger(show) : <Space.Compact className={styles.field} block>
                {multiple ? <Select
                    {...inputProps as any}
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
                    placeholder={placeholder || `请输入或选择${selectionLabel}`}
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
                    value={inputDraft ?? (values[0] || "")}
                    disabled={isDisabled}
                    placeholder={placeholder || `请输入或选择${selectionLabel}`}
                    onChange={(event) => {
                        const next = event.target.value;
                        if (suffixes.length && next) setInputDraft(next);
                        else {
                            setInputDraft(null);
                            changePaths(next ? [next] : []);
                        }
                    }}
                    onBlur={(event) => { commitInput(); inputProps.onBlur?.(event); }}
                    onPressEnter={(event) => { if (!event.nativeEvent.isComposing) commitInput(); inputProps.onPressEnter?.(event); }}/>}
                <Tooltip title={title}>
                    <Button
                        color={buttonProps?.type || buttonProps?.danger ? undefined : "default"}
                        variant={buttonProps?.type || buttonProps?.danger ? undefined : variant === "filled" ? "filled" : variant === "outlined" ? "outlined" : "text"}
                        {...buttonProps}
                        size={size}
                        disabled={isDisabled}
                        aria-label={title}
                        icon={<Icon type={type === "directory" ? "FolderOpenOutlined" : "FileSearchOutlined"}/>}
                        onClick={show}/>
                </Tooltip>
            </Space.Compact>}

            <ConfigProvider componentSize="middle" variant="outlined">
                <ProModal
                    title={<Title>{title}</Title>}
                    width="640px"
                    okText="选择"
                    onOk={() => confirm()}
                    onCancel={close}
                    modalProps={{
                        open: open && !isDisabled,
                        zIndex,
                        rootClassName: styles.modal,
                        style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                        cancelText: "取消",
                        okButtonProps: {
                            disabled: selectionPending || !multiple && (showingDisks || loading || Boolean(error) || !loadedPath
                                || comparablePath(loadedPath) !== comparablePath(path) || (type === "file" && !selectedPaths.length)),
                        },
                        focusable: {focusTriggerAfterClose: false},
                        mask: {closable: true},
                        styles: {body: {paddingTop: compact ? 10 : 15}},
                    }}>
                    <Form component={false} size={compact ? "small" : "middle"} variant="outlined">
                        <FileSelectBrowser
                            className={styles.browser}
                            compact={compact}
                            type={type}
                            path={path}
                            loadedPath={loadedPath}
                            multiple={multiple}
                            acceptsFile={acceptsFile}
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
                            onSelectionChange={changeSelection}
                            onConfirm={confirm}
                            onPageChange={changePage}
                            onCreate={openCreationForm}
                            onUpload={openUpload}
                            onReload={reloadList}/>
                    </Form>
                </ProModal>
            </ConfigProvider>
            <ProModal
                title={<Title>新建{creationLabel}</Title>}
                width="340px"
                okText="创建"
                onOk={creationForm.submit}
                onCancel={closeCreationForm}
                modalProps={{
                    open: creationOpen && open && !isDisabled,
                    zIndex: actionZIndex,
                    style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                    cancelText: "取消",
                    confirmLoading: creationSubmitting,
                    focusable: {focusTriggerAfterClose: true},
                    mask: {closable: true},
                    styles: {body: {paddingTop: compact ? 10 : 15}},
                    afterClose: () => {
                        if (creationOpenRef.current) return;
                        creationForm.resetFields();
                        setCreationPath("");
                    },
                }}>
                <Form ref={creationFormRef} form={creationForm} layout="vertical" variant="filled" preserve={false} requiredMark={false} onFinish={createEntry}>
                    <Form.Item name="name" rules={[
                        {required: true, whitespace: true, message: `请输入${creationLabel}名称`},
                        {max: 255, message: "名称不能超过255个字符"},
                        {validator: async (_, value) => {
                            const name = String(value || "").trim();
                            if (name === "." || name === ".." || /[\\/\0]/.test(name)) {
                                throw new Error("名称不能包含路径分隔符");
                            }
                        }},
                    ]}>
                        <Input aria-label={`请输入${creationLabel}名称`} placeholder={`请输入${creationLabel}名称`} maxLength={255} disabled={creationSubmitting}/>
                    </Form.Item>
                </Form>
            </ProModal>
            {uploadSession && open && !isDisabled && <FileUpload
                key={uploadSession.session}
                ref={uploadRef}
                path={uploadSession.path}
                zIndex={actionZIndex}
                onUploaded={(targetPath: string) => {
                    if (pickerSessionRef.current !== uploadSession.session) return;
                    refreshDirectory(targetPath);
                }}/>}
        </>
    );
});

FileSelect.displayName = "FileSelect";

export default memo(FileSelect);
