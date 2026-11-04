import React, {forwardRef, memo, useCallback, useContext, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, Form, Grid, Input, Select, Space, Tooltip} from "antd";
import type {ButtonProps, InputProps, InputRef, TableRef} from "antd";
import DisabledContext from "antd/es/config-provider/DisabledContext";
import useVariant from "antd/es/form/hooks/useVariants";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {createFile, getFileDisks, getFileInfo, getFileList} from "@/service/api/file";
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
import FileSelectBrowser from "./browser";
import useStyles from "@/pages/components/ui/path-select/styles";

export interface FileSelectProps extends Omit<InputProps, "value" | "defaultValue" | "onChange" | "onSelect" | "multiple"> {
    buttonProps?: ButtonProps;
    initialPath?: string;
    mode?: "directory" | "file";
    multiple?: boolean;
    value?: string[];
    defaultValue?: string[];
    onChange?: (paths: string[]) => void;
    onSelect?: (paths: string[]) => boolean | void;
}

// 单选、多选均使用路径数组，单选输入框仅显示第一个路径。
const FileSelect = forwardRef<InputRef, FileSelectProps>((props, ref) => {
    const {
        buttonProps,
        defaultValue = [],
        disabled,
        initialPath,
        mode = "directory",
        multiple = false,
        onChange,
        onSelect,
        placeholder,
        size,
        value,
        ...inputProps
    } = props;
    const theme = useTheme();
    const {message} = App.useApp();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const contextDisabled = useContext(DisabledContext);
    const isDisabled = disabled ?? contextDisabled;
    const multi = mode === "file" && multiple;
    const controlled = Object.prototype.hasOwnProperty.call(props, "value");
    const [localValue, setLocalValue] = useState<string[]>(defaultValue);
    const values = controlled ? value || [] : localValue;
    // 与 Input 一样继承 Form 和 ConfigProvider 的外观，按钮可通过 buttonProps 单独覆盖。
    const [variant] = useVariant("input", inputProps.variant, inputProps.bordered);
    const requestRef = useRef(0);
    const selectionRequestRef = useRef(0);
    const directoryRequestRef = useRef(0);
    const directoryOpenRef = useRef(false);
    const directorySubmittingRef = useRef(false);
    const [directoryForm] = Form.useForm();
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
    const [directoryOpen, setDirectoryOpen] = useState(false);
    const [directoryPath, setDirectoryPath] = useState("");
    const [directorySubmitting, setDirectorySubmitting] = useState(false);

    useEffect(() => () => {
        requestRef.current += 1;
        selectionRequestRef.current += 1;
        directoryRequestRef.current += 1;
        directoryOpenRef.current = false;
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
        const input = values[0] || "";
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
        setSelectedPaths(multi ? values.slice() : []);
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
    }, [initialPath, isDisabled, mode, multi, values]);

    const close = useCallback(() => {
        requestRef.current += 1;
        selectionRequestRef.current += 1;
        directoryRequestRef.current += 1;
        directoryOpenRef.current = false;
        directorySubmittingRef.current = false;
        setDirectoryOpen(false);
        setDirectorySubmitting(false);
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
        const normalized = (multi ? selectedPaths : [selectedPath]).map(normalizeFilePath);
        if (onSelect) {
            if (onSelect(normalized) === false) {
                return;
            }
        } else {
            onChange?.(normalized);
        }
        if (!controlled) setLocalValue(normalized);
        close();
    }, [close, controlled, error, isDisabled, loadedPath, loading, message, mode, multi, onChange, onSelect, path, selectedFile, selectedPaths, showingDisks]);

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
        const nextValue = [...unique.values()];
        if (!controlled) setLocalValue(nextValue);
        onChange?.(nextValue);
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

    const openDirectoryForm = useCallback(() => {
        if (isDisabled || showingDisks || browserLoading || browserError || !loadedPath
            || comparablePath(loadedPath) !== comparablePath(path)) return;
        directoryRequestRef.current += 1;
        directoryOpenRef.current = true;
        setDirectoryPath(loadedPath);
        setDirectoryOpen(true);
    }, [browserError, browserLoading, isDisabled, loadedPath, path, showingDisks]);

    const closeDirectoryForm = useCallback(() => {
        directoryRequestRef.current += 1;
        directoryOpenRef.current = false;
        directorySubmittingRef.current = false;
        setDirectoryOpen(false);
        setDirectorySubmitting(false);
    }, []);

    const createDirectory = async ({name}: {name: string}) => {
        if (isDisabled || !directoryOpenRef.current || directorySubmittingRef.current || !directoryPath) return;
        const request = directoryRequestRef.current;
        directorySubmittingRef.current = true;
        setDirectorySubmitting(true);
        try {
            const response = await createFile({body: {path: directoryPath, name: `${name.trim()}/`}});
            if (directoryRequestRef.current !== request) return;
            if (response?.code !== 200) {
                message.error(response?.message || "新建文件夹失败");
                return;
            }
            message.success(response?.message || "创建成功");
            closeDirectoryForm();
            setLoading(true);
            setReload((current) => current + 1);
        } catch (error: any) {
            if (directoryRequestRef.current === request) message.error(error?.message || "新建文件夹失败");
        } finally {
            if (directoryRequestRef.current === request) {
                directorySubmittingRef.current = false;
                setDirectorySubmitting(false);
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
                    value={values[0] || ""}
                    disabled={isDisabled}
                    placeholder={placeholder || (mode === "directory" ? "请输入或选择目录" : "请输入或选择文件")}
                    onChange={(event) => changePaths(event.target.value ? [event.target.value] : [])}/>}
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
                        <FileSelectBrowser
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
                            onCreateDirectory={openDirectoryForm}
                            onReload={reloadList}/>
                    </Form>
                </ProModal>
            </ConfigProvider>
            <ProModal
                title={<Title>新建文件夹</Title>}
                width="340px"
                okText="创建"
                onOk={directoryForm.submit}
                onCancel={closeDirectoryForm}
                modalProps={{
                    open: directoryOpen && open && !isDisabled,
                    zIndex: 2000,
                    style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                    cancelText: "取消",
                    confirmLoading: directorySubmitting,
                    focusable: {focusTriggerAfterClose: true},
                    mask: {closable: true},
                    styles: {body: {paddingTop: compact ? 10 : 15}},
                    afterClose: () => {
                        if (directoryOpenRef.current) return;
                        directoryForm.resetFields();
                        setDirectoryPath("");
                    },
                }}>
                <Form form={directoryForm} layout="vertical" variant="filled" preserve={false} requiredMark={false} onFinish={createDirectory}>
                    <Form.Item name="name" rules={[
                        {required: true, whitespace: true, message: "请输入文件夹名称"},
                        {max: 255, message: "名称不能超过255个字符"},
                        {validator: async (_, value) => {
                            const name = String(value || "").trim();
                            if (name === "." || name === ".." || /[\\/\0]/.test(name)) {
                                throw new Error("名称不能包含路径分隔符");
                            }
                        }},
                    ]}>
                        <Input aria-label="请输入文件夹名称" placeholder="请输入文件夹名称" maxLength={255} disabled={directorySubmitting}/>
                    </Form.Item>
                </Form>
            </ProModal>
        </>
    );
});

FileSelect.displayName = "FileSelect";

export default memo(FileSelect);
