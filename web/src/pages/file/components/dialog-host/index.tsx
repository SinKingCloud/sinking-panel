import React, {
    forwardRef,
    memo,
    useCallback,
    useEffect,
    useImperativeHandle,
    useRef,
} from "react";
import {App} from "antd";
import FilePreview from "@/pages/components/file-preview";
import {joinFilePath} from "../../utils";
import FileForm from "../form";
import type {FileFormRef} from "../form";
import FileEditor from "../editor";
import type {FileEditorRef} from "../editor";
import FileOperationForm from "../operation-form";
import type {FileOperationRef} from "../operation-form";
import FilePermissions from "../permissions";
import type {FilePermissionsRef} from "../permissions";
import FileProperties from "../properties";
import type {FilePropertiesRef} from "../properties";
import FileRecycleBin from "../recycle-bin";
import type {FileRecycleBinRef} from "../recycle-bin";
import FileTaskCenter from "../task-center";
import type {FileTaskCenterRef} from "../task-center";
import FileUpload from "../upload";
import type {FileUploadRef} from "../upload";
import FileTerminal from "../file-terminal";

export interface FileDialogHostRef {
    openCreate: (mode: any) => void;
    openEditor: (path?: string, name?: string) => void;
    restoreEditor: () => void;
    openRename: (record: any) => void;
    openPermissions: (record: any) => void;
    openProperties: (record: any) => void;
    openPreview: (files: readonly any[], active: string) => void;
    openOperation: (mode: any, record?: any) => void;
    openOperationMany: (
        mode: any,
        records: any[],
        clearSelectionOnSuccess?: boolean,
    ) => void;
    openUpload: () => void;
    openTerminal: (path?: string) => void;
    openRecycle: () => void;
    trackTask: (taskId: string, title: string) => void;
    closePathBound: () => void;
}

export interface FileDialogHostProps {
    path: string;
    roots: string[];
    loaded: boolean;
    loading: boolean;
    navigating: boolean;
    uploading: boolean;
    onReload: () => void;
    onUploadingChange: (uploading: boolean) => void;
    onOperationCompleted: (paths: string[]) => void;
    onEditorMinimizedChange: (minimized: boolean) => void;
}

const FileDialogHost = forwardRef<FileDialogHostRef, FileDialogHostProps>(({
    path,
    roots,
    loaded,
    loading,
    navigating,
    uploading,
    onReload,
    onUploadingChange,
    onOperationCompleted,
    onEditorMinimizedChange,
}, ref) => {
    const {message} = App.useApp();
    const formRef = useRef<FileFormRef | null>(null);
    const editorRef = useRef<FileEditorRef | null>(null);
    const propertiesRef = useRef<FilePropertiesRef | null>(null);
    const previewRef = useRef<any>(null);
    const permissionsRef = useRef<FilePermissionsRef | null>(null);
    const recycleBinRef = useRef<FileRecycleBinRef | null>(null);
    const operationRef = useRef<FileOperationRef | null>(null);
    const taskCenterRef = useRef<FileTaskCenterRef | null>(null);
    const uploadRef = useRef<FileUploadRef | null>(null);
    const [terminalPath, setTerminalPath] = React.useState("");
    const [terminalOpen, setTerminalOpen] = React.useState(false);
    const reloadTimerRef = useRef<number | undefined>(undefined);
    const onReloadRef = useRef(onReload);
    const previousPathRef = useRef(path);

    useEffect(() => {
        onReloadRef.current = onReload;
    }, [onReload]);

    const closePathBound = useCallback(() => {
        formRef.current?.close();
        operationRef.current?.close();
        permissionsRef.current?.close();
        propertiesRef.current?.close();
        previewRef.current?.close();
    }, []);

    useEffect(() => {
        if (previousPathRef.current !== path) {
            previousPathRef.current = path;
            closePathBound();
            // Keep the local terminal session alive. Its shell will receive
            // the new directory through FileTerminal without reconnecting.
            setTerminalPath((current) => current ? path : current);
        }
    }, [closePathBound, path]);

    useEffect(() => () => {
        if (reloadTimerRef.current !== undefined) {
            window.clearTimeout(reloadTimerRef.current);
        }
    }, []);

    const scheduleReload = useCallback(() => {
        if (reloadTimerRef.current !== undefined) {
            window.clearTimeout(reloadTimerRef.current);
        }
        reloadTimerRef.current = window.setTimeout(() => {
            reloadTimerRef.current = undefined;
            onReloadRef.current();
        }, 180);
    }, []);

    const openCreate = useCallback((mode: any) => {
        if (!loaded || navigating) {
            message.info("目录正在加载");
            return;
        }
        formRef.current?.open(mode, path);
    }, [loaded, message, navigating, path]);

    const openProperties = useCallback((record: any) => {
        propertiesRef.current?.open(joinFilePath(path, record.name), record);
    }, [path]);

    const openRename = useCallback((record: any) => {
        if (!loaded || navigating) {
            message.info("目录正在加载");
            return;
        }
        formRef.current?.openRename(joinFilePath(path, record.name), record);
    }, [loaded, message, navigating, path]);

    const openPermissions = useCallback((record: any) => {
        if (!loaded || navigating) {
            message.info("目录正在加载");
            return;
        }
        permissionsRef.current?.open(joinFilePath(path, record.name), record);
    }, [loaded, message, navigating, path]);

    const openEditor = useCallback((targetPath?: string, name?: string) => {
        if (!loaded || loading || navigating) {
            message.info("目录正在加载");
            return;
        }
        editorRef.current?.open(targetPath || path, name);
    }, [loaded, loading, message, navigating, path]);

    const restoreEditor = useCallback(() => {
        editorRef.current?.restore();
    }, []);

    const openPreview = useCallback((files: readonly any[], active: string) => {
        if (!loaded || loading || navigating) {
            message.info("目录正在加载");
            return;
        }
        previewRef.current?.open(files, active);
    }, [loaded, loading, message, navigating]);

    const openOperationContext = useCallback((
        mode: any,
        records?: any[],
        clearSelectionOnSuccess?: boolean,
    ) => {
        if (!loaded || navigating) {
            message.info("目录正在加载");
            return;
        }
        operationRef.current?.open(mode, {
            path,
            records,
            clearSelectionOnSuccess,
        });
    }, [loaded, message, navigating, path]);

    const openOperation = useCallback((mode: any, record?: any) => {
        openOperationContext(mode, record ? [record] : undefined);
    }, [openOperationContext]);

    const openOperationMany = useCallback((
        mode: any,
        records: any[],
        clearSelectionOnSuccess = false,
    ) => {
        openOperationContext(mode, records, clearSelectionOnSuccess);
    }, [openOperationContext]);

    const openUpload = useCallback(() => {
        if (uploading) {
            uploadRef.current?.open();
            return;
        }
        if (!loaded || loading) {
            message.info("目录正在加载");
            return;
        }
        uploadRef.current?.open();
    }, [loaded, loading, message, uploading]);

    const openTerminal = useCallback((targetPath = path) => {
        if (!loaded || loading || navigating) {
            message.info("目录正在加载");
            return;
        }
        setTerminalOpen(true);
        setTerminalPath(targetPath);
    }, [loaded, loading, message, navigating, path]);

    const openRecycle = useCallback(() => {
        recycleBinRef.current?.open();
    }, []);

    const trackTask = useCallback((taskId: string, title: string) => {
        taskCenterRef.current?.track(taskId, title);
    }, []);

    useImperativeHandle(ref, () => ({
        openCreate,
        openEditor,
        restoreEditor,
        openRename,
        openPermissions,
        openProperties,
        openPreview,
        openOperation,
        openOperationMany,
        openUpload,
        openTerminal,
        openRecycle,
        trackTask,
        closePathBound,
    }), [
        closePathBound,
        openCreate,
        openEditor,
        restoreEditor,
        openRename,
        openPermissions,
        openOperation,
        openOperationMany,
        openProperties,
        openPreview,
        openRecycle,
        openUpload,
        openTerminal,
        trackTask,
    ]);

    const renameFromProperties = useCallback((targetPath: string, record: any) => {
        formRef.current?.openRename(targetPath, record, true);
    }, []);

    const permissionsFromProperties = useCallback((targetPath: string, record: any) => {
        permissionsRef.current?.open(targetPath, record, true);
    }, []);

    const handleFormSuccess = useCallback((result: any) => {
        onReloadRef.current();
        if (result.mode === "rename" && result.sourcePath && result.targetPath) {
            propertiesRef.current?.updateRename(result.sourcePath, result.targetPath, result.name);
        }
    }, []);

    const handlePermissionsSuccess = useCallback((targetPath: string, permissions: string) => {
        onReloadRef.current();
        propertiesRef.current?.updatePermissions(targetPath, permissions);
    }, []);

    const handleOperationSuccess = useCallback((context: any) => {
        if (!context.clearSelectionOnSuccess) {
            return;
        }
        onOperationCompleted((context.records || []).map((record) => (
            joinFilePath(context.path, record.name)
        )));
    }, [onOperationCompleted]);

    const handleUploadingChange = useCallback((active: boolean) => {
        onUploadingChange(active);
    }, [onUploadingChange]);

    const handleRecycleMutation = useCallback(() => {
        onReloadRef.current();
    }, []);

    return (
        <>
            <FileForm ref={formRef} onSuccess={handleFormSuccess}/>
            <FileEditor
                ref={editorRef}
                roots={roots}
                onMutation={scheduleReload}
                onMinimizedChange={onEditorMinimizedChange}/>
            <FileProperties
                ref={propertiesRef}
                onRename={renameFromProperties}
                onPermissions={permissionsFromProperties}/>
            <FilePreview ref={previewRef}/>
            <FilePermissions ref={permissionsRef} onSuccess={handlePermissionsSuccess}/>
            <FileRecycleBin ref={recycleBinRef} onMutation={handleRecycleMutation}/>
            <FileOperationForm
                ref={operationRef}
                onTask={trackTask}
                onSuccess={handleOperationSuccess}/>
            <FileTaskCenter ref={taskCenterRef} onSettled={scheduleReload}/>
            <FileUpload
                ref={uploadRef}
                path={path}
                selectionDisabled={!loaded || loading}
                onUploaded={scheduleReload}
                onUploadingChange={handleUploadingChange}/>
            <FileTerminal
                path={terminalPath}
                open={terminalOpen}
                onClose={() => setTerminalOpen(false)}/>
        </>
    );
});

FileDialogHost.displayName = "FileDialogHost";

export default memo(FileDialogHost);
