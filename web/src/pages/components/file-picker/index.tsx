import React, {memo, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {App, Button, Grid, Input, Space, Tooltip} from "antd";
import type {InputRef, TableRef} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {getFileDisks, getFileInfo, getFileList} from "@/service/api/file";
import type {FileRecord} from "@/service/api/file";
import {
    buildFileBreadcrumbs,
    comparableFilePath as comparablePath,
    isAbsoluteFilePath,
    isFilePathWithin as isPathWithin,
    normalizeFilePath,
    parentFilePath,
} from "@/pages/file/utils";
import FilePickerBrowser from "./browser";
import useStyles from "./styles";
import type {FilePickerProps, SelectedFile} from "./types";

export type {FilePickerMode, FilePickerProps} from "./types";

const FilePicker = ({
    disabled,
    mode = "directory",
    onChange,
    placeholder,
    value = "",
    ...inputProps
}: FilePickerProps) => {
    const theme = useTheme();
    const {message} = App.useApp();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const requestRef = useRef(0);
    const selectionRequestRef = useRef(0);
    const pathRef = useRef<HTMLOListElement>(null);
    const pathInputRef = useRef<InputRef>(null);
    const currentPathRef = useRef<HTMLButtonElement>(null);
    const tableRef = useRef<TableRef>(null);
    const [open, setOpen] = useState(false);
    const [path, setPath] = useState("");
    const [loadedPath, setLoadedPath] = useState("");
    const [selectedFile, setSelectedFile] = useState<SelectedFile>();
    const [disks, setDisks] = useState<string[]>([]);
    const [items, setItems] = useState<FileRecord[]>([]);
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
    const diskRoot = useMemo(() => {
        if (!path) {
            return disks[0] || "/";
        }
        const matched = disks
            .filter((disk) => isPathWithin(path, disk))
            .sort((left, right) => comparablePath(right).length - comparablePath(left).length);
        return matched[0] || buildFileBreadcrumbs(path)[0]?.path || "/";
    }, [disks, path]);
    const selectedDisk = disks.find((disk) => comparablePath(disk) === comparablePath(diskRoot));
    const pathBreadcrumbs = useMemo(() => {
        if (!path) {
            return [];
        }
        const all = buildFileBreadcrumbs(path);
        const start = all.findIndex((item) => comparablePath(item.path) === comparablePath(diskRoot));
        if (start <= 0) {
            return all;
        }
        return [{label: diskRoot, path: diskRoot}, ...all.slice(start + 1)];
    }, [diskRoot, path]);
    const parentPath = path ? parentFilePath(path) : diskRoot;
    const canGoUp = Boolean(path) && comparablePath(path) !== comparablePath(diskRoot);

    const navigate = useCallback((nextPath: string) => {
        const targetPath = normalizeFilePath(nextPath);
        requestRef.current += 1;
        selectionRequestRef.current += 1;
        setLoading(true);
        setEditingPath(false);
        setPathDraft(targetPath);
        setPath(targetPath);
        setLoadedPath("");
        setSelectedFile(undefined);
        setPage(1);
        setError("");
    }, []);

    const show = useCallback(() => {
        const input = String(value || "").trim();
        const absolute = isAbsoluteFilePath(input) ? normalizeFilePath(input) : "";
        const initialPath = mode === "file" && absolute ? parentFilePath(absolute) : absolute;
        const breadcrumbs = absolute ? buildFileBreadcrumbs(absolute) : [];
        const initialName = breadcrumbs.length > 1 ? breadcrumbs[breadcrumbs.length - 1].label : "";
        const selectionRequest = ++selectionRequestRef.current;
        requestRef.current += 1;
        setPath(initialPath);
        setLoadedPath("");
        setSelectedFile(undefined);
        setDisks([]);
        setItems([]);
        setTotal(0);
        setPage(1);
        setError("");
        setLoading(true);
        setEditingPath(false);
        setPathDraft(initialPath);
        setOpen(true);
        if (mode === "file" && absolute && initialName) {
            void getFileInfo({body: {path: absolute}}).then((response) => {
                if (selectionRequestRef.current === selectionRequest
                    && response?.code === 200 && response.data && !response.data.is_dir) {
                    setSelectedFile({name: response.data.name || initialName, path: absolute});
                }
            }).catch(() => undefined);
        }
    }, [mode, value]);

    const close = useCallback(() => {
        requestRef.current += 1;
        selectionRequestRef.current += 1;
        setLoading(false);
        setEditingPath(false);
        setPathDraft(path);
        setOpen(false);
    }, [path]);

    const confirm = useCallback((file = selectedFile) => {
        const selectedPath = mode === "directory" ? loadedPath : file?.path;
        if (!selectedPath) {
            return;
        }
        onChange?.(normalizeFilePath(selectedPath));
        close();
    }, [close, loadedPath, mode, onChange, selectedFile]);

    const selectFile = useCallback((file: SelectedFile) => {
        selectionRequestRef.current += 1;
        setSelectedFile(file);
    }, []);

    const changePage = useCallback((nextPage: number) => {
        setLoading(true);
        setPage(nextPage);
        window.requestAnimationFrame(() => {
            tableRef.current?.scrollTo({top: 0});
        });
    }, []);

    const openPathEditor = useCallback(() => {
        if (loading || !path) {
            return;
        }
        setPathDraft(path);
        setEditingPath(true);
    }, [loading, path]);

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
        const allowedRoots = Array.from(new Set([...disks, diskRoot]));
        if (!allowedRoots.some((root) => isPathWithin(nextPath, root))) {
            message.warning("路径不在可用磁盘中");
            return;
        }
        if (comparablePath(nextPath) === comparablePath(path)) {
            cancelPathEditor(true);
            return;
        }
        navigate(nextPath);
    }, [cancelPathEditor, diskRoot, disks, message, navigate, path, pathDraft]);

    const reloadList = useCallback(() => {
        setReload((current) => current + 1);
    }, []);

    useEffect(() => {
        if (!open) {
            return;
        }
        let active = true;
        void getFileDisks().then((response) => {
            if (!active) {
                return;
            }
            if (response?.code !== 200 || !Array.isArray(response.data)) {
                setPath((current) => current || "/");
                return;
            }
            const nextDisks = Array.from(new Set(response.data.map(normalizeFilePath)));
            setDisks(nextDisks);
            setPath((current) => current || nextDisks[0] || "/");
        }).catch(() => {
            if (active) {
                setPath((current) => current || "/");
            }
        });
        return () => {
            active = false;
        };
    }, [open]);

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
        if (loading && editingPath) {
            setEditingPath(false);
            setPathDraft(path);
        }
    }, [editingPath, loading, path]);

    useEffect(() => {
        if (!open || !path) {
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
            const data = response.data;
            const nextItems = Array.isArray(data?.list) ? data.list : [];
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
    }, [mode, open, page, path, reload]);

    return (
        <>
            <Space.Compact className={styles.field} block>
                <Input
                    {...inputProps}
                    value={value}
                    disabled={disabled}
                    placeholder={placeholder || (mode === "directory" ? "请输入或选择目录" : "请输入或选择文件")}
                    onChange={(event) => onChange?.(event.target.value)}/>
                <Tooltip title={title}>
                    <Button
                        disabled={disabled}
                        aria-label={title}
                        icon={<Icon type={mode === "directory" ? "FolderOpenOutlined" : "FileSearchOutlined"}/>}
                        onClick={show}/>
                </Tooltip>
            </Space.Compact>

            <ProModal
                title={<Title>{title}</Title>}
                width="640px"
                okText="选择"
                onOk={() => confirm()}
                onCancel={close}
                modalProps={{
                    open,
                    rootClassName: styles.modal,
                    style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                    cancelText: "取消",
                    okButtonProps: {
                        disabled: loading || Boolean(error) || (mode === "directory"
                            ? !loadedPath || comparablePath(loadedPath) !== comparablePath(path)
                            : !selectedFile),
                    },
                    focusable: {focusTriggerAfterClose: false},
                    mask: {closable: true},
                    styles: {body: {paddingTop: compact ? 10 : 15}},
                }}>
                <FilePickerBrowser
                    className={styles.browser}
                    compact={compact}
                    mode={mode}
                    path={path}
                    loadedPath={loadedPath}
                    selectedFile={selectedFile}
                    disks={disks}
                    selectedDisk={selectedDisk}
                    pathBreadcrumbs={pathBreadcrumbs}
                    parentPath={parentPath}
                    diskRoot={diskRoot}
                    canGoUp={canGoUp}
                    editingPath={editingPath}
                    pathDraft={pathDraft}
                    loading={loading}
                    error={error}
                    items={items}
                    page={page}
                    pageSize={50}
                    total={total}
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
                    onConfirm={confirm}
                    onPageChange={changePage}
                    onReload={reloadList}/>
            </ProModal>
        </>
    );
};

export default memo(FilePicker);
