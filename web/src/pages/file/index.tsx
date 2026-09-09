import React, {useCallback, useLayoutEffect, useRef, useState} from "react";
import {App, Button, Empty} from "antd";
import {Body, Icon} from "sinking-antd";
import Table from "@/components/table";
import FileDialogHost from "./components/dialog-host";
import type {FileDialogHostRef} from "./components/dialog-host";
import FilePathBar from "./components/path-bar";
import useFileTable from "./components/table";
import useFileClipboard from "./hooks/clipboard";
import useFileDeletion from "./hooks/deletion";
import useDirectoryCounts from "./hooks/directory-counts";
import useFileDownload from "./hooks/download";
import useFileNavigation from "./hooks/navigation";
import useFileOperationLock from "./hooks/operation-lock";
import useFileSelection from "./hooks/selection";
import {joinFilePath, normalizeFilePath} from "./utils";

export default (): React.ReactNode => {
    const {message} = App.useApp();
    const dialogHostRef = useRef<FileDialogHostRef | null>(null);
    const cancelPendingNavigationRef = useRef<() => void>(() => undefined);
    const [uploading, setUploading] = useState(false);
    const [editorMinimized, setEditorMinimized] = useState(false);

    const cancelPendingNavigation = useCallback(() => {
        cancelPendingNavigationRef.current();
    }, []);
    const closePathBoundDialogs = useCallback(() => {
        dialogHostRef.current?.closePathBound();
    }, []);
    const {
        list,
        disks,
        keyword,
        setKeyword,
        navigate,
        navigationVersionRef,
    } = useFileNavigation({
        onBeforeNavigate: cancelPendingNavigation,
        onExternalPathChange: closePathBoundDialogs,
    });
    const operationLock = useFileOperationLock();
    const selection = useFileSelection({
        path: list.path,
        items: list.items,
        operatingPaths: operationLock.paths,
    });
    const {counts: directoryCounts, countDirectory} = useDirectoryCounts({
        path: list.path,
        items: list.items,
        enabled: list.loaded,
    });
    const download = useFileDownload(list.path);
    const trackTask = useCallback((taskId: string, title: string) => {
        dialogHostRef.current?.trackTask(taskId, title);
    }, []);
    const clipboard = useFileClipboard({
        path: list.path,
        loaded: list.loaded,
        loading: list.loading,
        navigating: list.navigating,
        selectedRecords: selection.selectedRecords,
        hasBusySelection: selection.hasBusySelection,
        trackTask,
    });
    const deletion = useFileDeletion({
        path: list.path,
        selectedRecords: selection.selectedRecords,
        navigationVersionRef,
        operationLock,
        reload: list.reload,
        removeSelection: selection.remove,
    });

    useLayoutEffect(() => {
        cancelPendingNavigationRef.current = () => {
            clipboard.cancelPending();
            deletion.closeConfirm();
        };
        return () => {
            cancelPendingNavigationRef.current = () => undefined;
        };
    }, [clipboard.cancelPending, deletion.closeConfirm]);

    const openDirectory = useCallback((record: any) => {
        navigate(joinFilePath(list.path, record.name));
    }, [list.path, navigate]);
    const openCreate = useCallback((mode: any) => {
        dialogHostRef.current?.openCreate(mode);
    }, []);
    const openEditor = useCallback((path?: string, name?: string) => {
        dialogHostRef.current?.openEditor(path, name);
    }, []);
    const restoreEditor = useCallback(() => {
        dialogHostRef.current?.restoreEditor();
    }, []);
    const openRename = useCallback((record: any) => {
        dialogHostRef.current?.openRename(record);
    }, []);
    const openPermissions = useCallback((record: any) => {
        dialogHostRef.current?.openPermissions(record);
    }, []);
    const openProperties = useCallback((record: any) => {
        dialogHostRef.current?.openProperties(record);
    }, []);
    const openPreview = useCallback((files: readonly any[], active: string) => {
        dialogHostRef.current?.openPreview(files, active);
    }, []);
    const openOperation = useCallback((mode: any, record?: any) => {
        dialogHostRef.current?.openOperation(mode, record);
    }, []);
    const openRemoteDownload = useCallback(() => {
        dialogHostRef.current?.openOperation("remote-download");
    }, []);
    const openTerminal = useCallback(() => {
        dialogHostRef.current?.openTerminal(list.path);
    }, [list.path]);
    const openRecycle = useCallback(() => {
        dialogHostRef.current?.openRecycle();
    }, []);
    const openUpload = useCallback(() => {
        dialogHostRef.current?.openUpload();
    }, []);
    const copyRecord = useCallback((record: any) => {
        clipboard.copyRecords([record]);
    }, [clipboard.copyRecords]);
    const moveRecord = useCallback((record: any) => {
        clipboard.moveRecords([record]);
    }, [clipboard.moveRecords]);
    const paste = useCallback(() => {
        void clipboard.paste();
    }, [clipboard.paste]);
    const openSelectedCompress = useCallback(() => {
        if (selection.selectedRecords.length === 0) {
            return;
        }
        if (selection.hasBusySelection) {
            message.warning("选中的文件正在执行其他操作");
            return;
        }
        dialogHostRef.current?.openOperationMany(
            "compress",
            [...selection.selectedRecords],
            true,
        );
    }, [message, selection.hasBusySelection, selection.selectedRecords]);
    const removeCompletedSelection = useCallback((paths: string[]) => {
        selection.remove(paths);
    }, [selection.remove]);

    const fileTable = useFileTable({
        path: list.path,
        items: list.items,
        loading: list.loading,
        actionsDisabled: list.loading || list.navigating,
        sort: list.sort,
        order: list.order,
        directoryCounts,
        onCountDirectory: countDirectory,
        operatingPaths: operationLock.paths,
        selectedPaths: selection.selectedPaths,
        selectionDisabled: list.loading || list.navigating,
        onSelectionChange: selection.change,
        onSortChange: list.changeSort,
        onOpen: openDirectory,
        onEdit: openEditor,
        onPreview: openPreview,
        onDownload: download,
        onRename: openRename,
        onPermissions: openPermissions,
        onCopy: copyRecord,
        onMove: moveRecord,
        onProperties: openProperties,
        onOperation: openOperation,
        onDelete: deletion.confirmRecord,
    });
    const selectionOperationDisabled = list.loading || list.navigating || selection.hasBusySelection;
    const selectionClipboardDisabled = selectionOperationDisabled || clipboard.pasting;

    return (
        <Body>
            <Table
                {...fileTable}
                ariaLabel="文件管理"
                hero={{
                    title: "文件管理",
                    eyebrow: "FILE MANAGER",
                    action: disks.length > 1 ? {
                        label: "切换磁盘",
                        ariaLabel: "切换磁盘",
                        icon: "DatabaseOutlined",
                        suffixIcon: "SwapOutlined",
                        menu: {
                            items: disks.map((disk, index) => ({key: String(index), label: disk})),
                            onClick: ({key}) => {
                                const target = disks[Number(key)];
                                if (target) navigate(normalizeFilePath(target));
                            },
                        },
                    } : undefined,
                }}
                toolbar={{
                    search: {
                        value: keyword,
                        ariaLabel: "搜索当前目录",
                        placeholder: "搜索当前目录",
                        onChange: setKeyword,
                    },
                    refresh: {loading: list.loading, ariaLabel: "刷新文件列表", onClick: list.reload},
                    actions: [
                        {
                            key: "create",
                            type: "button",
                            label: "新建",
                            icon: "PlusOutlined",
                            suffixIcon: "DownOutlined",
                            "aria-label": "新建文件或文件夹",
                            menu: {
                                items: [
                                    {key: "directory", label: "新建文件夹", icon: <Icon type="FolderAddOutlined"/>},
                                    {key: "file", label: "新建空文件", icon: <Icon type="FileAddOutlined"/>},
                                ],
                                onClick: ({key}) => openCreate(key),
                            },
                        },
                        {
                            key: "upload",
                            type: "button",
                            label: uploading ? "查看上传" : "上传文件",
                            icon: uploading ? "LoadingOutlined" : "UploadOutlined",
                            "aria-label": uploading ? "查看上传" : "上传文件",
                            onClick: openUpload,
                        },
                        {
                            key: "download", type: "button", label: "远程下载", icon: "CloudDownloadOutlined",
                            "aria-label": "远程下载", onClick: openRemoteDownload,
                        },
                        {
                            key: "terminal", type: "button", label: "终端", icon: "CodeOutlined",
                            "aria-label": "打开终端", onClick: openTerminal,
                        },
                        {
                            key: "recycle", type: "button", label: "回收站", icon: "DeleteOutlined",
                            "aria-label": "回收站", onClick: openRecycle,
                        },
                    ],
                }}
                contentBar={{
                    content: <FilePathBar
                        path={list.path}
                        roots={disks}
                        disabled={!list.loaded || list.navigating}
                        onNavigate={navigate}/>,
                    ariaLabel: "当前目录操作",
                    actions: [
                        ...(editorMinimized ? [{
                            key: "editor", type: "button" as const, label: "编辑器", icon: "EditOutlined",
                            "aria-label": "恢复文件编辑器", onClick: restoreEditor,
                        }] : []),
                        ...(clipboard.count > 0 ? [{
                            key: "paste",
                            type: "button" as const,
                            label: `${clipboard.mode === "move" ? "移动" : "粘贴"} ${clipboard.count} 项`,
                            "aria-label": clipboard.mode === "move"
                                ? `将 ${clipboard.count} 项移动到当前目录`
                                : `粘贴 ${clipboard.count} 项`,
                            disabled: !list.loaded || list.navigating,
                            loading: clipboard.pasting,
                            icon: clipboard.mode === "move" ? "ScissorOutlined" : "SnippetsOutlined",
                            onClick: paste,
                        }] : []),
                    ],
                }}
                rowSelection={{
                    ...fileTable.rowSelection,
                    onClear: selection.clear,
                    clearDisabled: list.loading || list.navigating,
                    actions: [
                        {
                            key: "copy", type: "button", label: "复制", icon: "CopyOutlined",
                            disabled: selectionClipboardDisabled, onClick: clipboard.copySelected,
                        },
                        {
                            key: "move", type: "button", label: "移动", icon: "ScissorOutlined",
                            disabled: selectionClipboardDisabled, onClick: clipboard.moveSelected,
                        },
                        {
                            key: "compress", type: "button", label: "压缩", icon: "FileZipOutlined",
                            disabled: selectionOperationDisabled, onClick: openSelectedCompress,
                        },
                        {
                            key: "delete", type: "button", label: "删除", icon: "DeleteOutlined",
                            danger: true, disabled: selectionOperationDisabled, onClick: deletion.confirmSelection,
                        },
                    ],
                }}
                empty={list.initialized && list.items.length === 0}
                emptyContent={list.initialError ? (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="无法读取当前目录">
                        <Button
                            size="small"
                            aria-label="重新加载文件列表"
                            icon={<Icon type="ReloadOutlined"/>}
                            onClick={list.reload}>
                            重新加载
                        </Button>
                    </Empty>
                ) : undefined}
                pagination={{
                    page: list.page,
                    pageSize: list.pageSize,
                    total: list.total,
                    disabled: list.navigating,
                    onChange: list.changePage,
                }}/>

            <FileDialogHost
                ref={dialogHostRef}
                path={list.path}
                roots={disks}
                loaded={list.loaded}
                loading={list.loading}
                navigating={list.navigating}
                uploading={uploading}
                onReload={list.reload}
                onUploadingChange={setUploading}
                onEditorMinimizedChange={setEditorMinimized}
                onOperationCompleted={removeCompletedSelection}/>
        </Body>
    );
};
