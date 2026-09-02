import React, {useCallback, useLayoutEffect, useRef, useState} from "react";
import {App, Button, Empty} from "antd";
import {Body, Icon} from "sinking-antd";
import PageTable from "@/pages/components/table";
import FileDialogHost from "./components/dialog-host";
import type {FileDialogHostRef} from "./components/dialog-host";
import Header from "./components/header";
import FileTable from "./components/table";
import useFileClipboard from "./hooks/clipboard";
import useFileDeletion from "./hooks/deletion";
import useDirectoryCounts from "./hooks/directory-counts";
import useFileDownload from "./hooks/download";
import useFileNavigation from "./hooks/navigation";
import useFileOperationLock from "./hooks/operation-lock";
import useFileSelection from "./hooks/selection";
import {joinFilePath} from "./utils";

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

    return (
        <Body>
            <PageTable
                ariaLabel="文件管理"
                header={(
                    <Header
                        path={list.path}
                        disks={disks}
                        keyword={keyword}
                        uploading={uploading}
                        refreshing={list.loading}
                        clipboardCount={clipboard.count}
                        clipboardMode={clipboard.mode}
                        pasting={clipboard.pasting}
                        pasteDisabled={!list.loaded || list.navigating}
                        directoryActionsDisabled={!list.loaded || list.navigating}
                        selectedCount={selection.selectedRecords.length}
                        editorMinimized={editorMinimized}
                        selectionClipboardDisabled={list.loading
                            || list.navigating
                            || selection.hasBusySelection
                            || clipboard.pasting}
                        selectionOperationDisabled={list.loading
                            || list.navigating
                            || selection.hasBusySelection}
                        selectionClearDisabled={list.loading || list.navigating}
                        onKeywordChange={setKeyword}
                        onNavigate={navigate}
                        onCreate={openCreate}
                        onUpload={openUpload}
                        onPaste={paste}
                        onCopySelected={clipboard.copySelected}
                        onMoveSelected={clipboard.moveSelected}
                        onCompressSelected={openSelectedCompress}
                        onDeleteSelected={deletion.confirmSelection}
                        onClearSelection={selection.clear}
                        onRestoreEditor={restoreEditor}
                        onRemoteDownload={openRemoteDownload}
                        onOpenTerminal={openTerminal}
                        onOpenRecycle={openRecycle}
                        onRefresh={list.reload}/>
                )}
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
                }}>
                <FileTable
                    path={list.path}
                    items={list.items}
                    loading={list.loading}
                    actionsDisabled={list.loading || list.navigating}
                    sort={list.sort}
                    order={list.order}
                    directoryCounts={directoryCounts}
                    onCountDirectory={countDirectory}
                    operatingPaths={operationLock.paths}
                    selectedPaths={selection.selectedPaths}
                    selectionDisabled={list.loading || list.navigating}
                    onSelectionChange={selection.change}
                    onSortChange={list.changeSort}
                    onOpen={openDirectory}
                    onEdit={openEditor}
                    onPreview={openPreview}
                    onDownload={download}
                    onRename={openRename}
                    onPermissions={openPermissions}
                    onCopy={copyRecord}
                    onMove={moveRecord}
                    onProperties={openProperties}
                    onOperation={openOperation}
                    onDelete={deletion.confirmRecord}/>
            </PageTable>

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
