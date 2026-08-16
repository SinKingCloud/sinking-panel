import React, {
    forwardRef,
    memo,
    useCallback,
    useEffect,
    useImperativeHandle,
    useMemo,
    useRef,
    useState,
} from "react";
import {App, Button, ConfigProvider, Empty, Grid, message as antdMessage, Spin, Tooltip} from "antd";
import {Icon, ProModal, useTheme} from "sinking-antd";
import {history} from "umi";
import defaultSettings from "@/../config/defaultSettings";
import AceEditor from "@/components/ace-editor";
import {deleteFile, renameFile} from "@/service/api/file";
import useFileEditorDocument from "../hooks/editor-document";
import useFileEditorPreferences from "../hooks/editor-preferences";
import useFileEditorTree from "../hooks/editor-tree";
import type {FileCreateMode, FileFormMode, FileFormResult} from "../types";
import {
    isFilePathWithin,
    joinFilePath,
    normalizeFilePath,
    parentFilePath,
} from "../utils";
import FileEditorSettings from "./editor-settings";
import FileEditorTabs from "./editor-tabs";
import FileEditorTree from "./editor-tree";
import type {FileEditorTreeNode} from "./editor.utils";
import {getEditorMode} from "./editor.utils";
import FileForm from "./form";
import type {FileFormRef} from "./form";
import FilePermissions from "./permissions";
import type {FilePermissionsRef} from "./permissions";
import {copyTextToClipboard} from "./properties.utils";
import useStyles from "./editor.styles";

const acePath = `${defaultSettings?.basePath || "/"}ace`;

interface FileEditorSession {
    generation: number;
}

interface FormContext {
    generation: number;
    mode: FileFormMode;
    parentPath: string;
    isDirectory?: boolean;
}

interface EditorViewState {
    cursor?: {row: number; column: number};
    scrollTop?: number;
    scrollLeft?: number;
}

interface Destroyable {
    destroy: () => void;
}

interface FileEditorProps {
    roots: string[];
    onMutation: () => void;
}

export interface FileEditorRef {
    open: (path: string, name?: string) => void;
    close: (force?: boolean) => void;
}

const isMobileViewport = () => (
    typeof window !== "undefined" && window.matchMedia("(max-width: 720px)").matches
);

const FileEditor = forwardRef(function FileEditor(
    {roots, onMutation}: FileEditorProps,
    ref: React.ForwardedRef<FileEditorRef>,
) {
    const {modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact, dark});
    const formRef = useRef<FileFormRef | null>(null);
    const permissionsRef = useRef<FilePermissionsRef | null>(null);
    const workspaceRef = useRef<HTMLDivElement | null>(null);
    const aceRef = useRef<any>(null);
    const sessionRef = useRef<FileEditorSession | undefined>(undefined);
    const sessionGenerationRef = useRef(0);
    const formContextRef = useRef<FormContext | undefined>(undefined);
    const discardConfirmRef = useRef<Destroyable | undefined>(undefined);
    const deletingPathsRef = useRef(new Set<string>());
    const saveCommandRef = useRef<() => void>(() => undefined);
    const viewStateRef = useRef(new Map<string, EditorViewState>());
    const [session, setSession] = useState<FileEditorSession>();
    const [treeCollapsed, setTreeCollapsed] = useState(false);
    const [aceError, setAceError] = useState<{key: string; message: string}>();
    const [aceAttempt, setAceAttempt] = useState(0);
    const [fullscreen, setFullscreen] = useState(false);

    const getWorkspacePopupContainer = useCallback(() => {
        const workspace = workspaceRef.current;
        return workspace && document.fullscreenElement === workspace
            ? workspace
            : document.body;
    }, []);
    const getWorkspaceMessageContainer = useCallback(() => (
        workspaceRef.current || document.body
    ), []);
    const [message, messageContextHolder] = antdMessage.useMessage({
        getContainer: getWorkspaceMessageContainer,
    });
    const tree = useFileEditorTree({roots, message});
    const files = useFileEditorDocument({onMutation, message});
    const {preferences, setPreferences, resolveTheme} = useFileEditorPreferences();
    const hasDirtyTabs = useMemo(() => files.tabs.some((tab) => tab.dirty), [files.tabs]);

    const captureViewState = useCallback((key = files.activeKey) => {
        const editor = aceRef.current;
        if (!key || !editor) {
            return;
        }
        viewStateRef.current.set(key, {
            cursor: editor.getCursorPosition?.(),
            scrollTop: editor.session?.getScrollTop?.(),
            scrollLeft: editor.session?.getScrollLeft?.(),
        });
    }, [files.activeKey]);

    const exitWorkspaceFullscreen = useCallback(async () => {
        if (document.fullscreenElement === workspaceRef.current) {
            await document.exitFullscreen();
        }
    }, []);

    const destroyDiscardConfirm = useCallback(() => {
        discardConfirmRef.current?.destroy();
        discardConfirmRef.current = undefined;
    }, []);

    const closeNow = useCallback(() => {
        sessionGenerationRef.current += 1;
        destroyDiscardConfirm();
        formContextRef.current = undefined;
        formRef.current?.close();
        permissionsRef.current?.close();
        if (document.fullscreenElement === workspaceRef.current) {
            void document.exitFullscreen().catch(() => undefined);
        }
        tree.close();
        files.reset();
        viewStateRef.current.clear();
        aceRef.current = null;
        sessionRef.current = undefined;
        setTreeCollapsed(false);
        setAceError(undefined);
        setFullscreen(false);
        setSession(undefined);
    }, [destroyDiscardConfirm, files.reset, tree.close]);

    const runAfterDiscard = useCallback((
        action: () => void,
        key?: string,
        content = "存在尚未保存的文件，继续操作将丢失这些内容。",
    ) => {
        if (files.isSaving(key)) {
            message.info("文件正在保存，请稍候");
            return;
        }
        if (!files.hasUnsavedChanges(key)) {
            action();
            return;
        }

        const showConfirm = () => {
            destroyDiscardConfirm();
            let instance: Destroyable | undefined;
            instance = modal.confirm({
                title: "放弃未保存的修改？",
                content,
                okText: "放弃修改",
                cancelText: "继续编辑",
                okButtonProps: {danger: true},
                onOk: () => {
                    if (discardConfirmRef.current === instance) {
                        discardConfirmRef.current = undefined;
                    }
                    action();
                },
                afterClose: () => {
                    if (discardConfirmRef.current === instance) {
                        discardConfirmRef.current = undefined;
                    }
                },
            });
            discardConfirmRef.current = instance;
        };

        if (document.fullscreenElement === workspaceRef.current) {
            void document.exitFullscreen().then(showConfirm).catch(showConfirm);
            return;
        }
        showConfirm();
    }, [destroyDiscardConfirm, files.hasUnsavedChanges, files.isSaving, message, modal]);

    useEffect(() => {
        if (!session || !hasDirtyTabs) {
            return;
        }
        let active = true;
        const filePathname = history.location.pathname;
        let unblock: () => void = () => undefined;
        let unlisten: () => void = () => undefined;
        const register = (): void => {
            unblock = history.block((transition) => {
                if (transition.location.pathname === filePathname) {
                    unblock();
                    unlisten();
                    unlisten = history.listen(() => {
                        unlisten();
                        unlisten = () => undefined;
                        if (active && files.hasUnsavedChanges()) {
                            register();
                        }
                    });
                    transition.retry();
                    return;
                }
                runAfterDiscard(() => {
                    active = false;
                    unlisten();
                    unblock();
                    transition.retry();
                });
            });
        };
        register();
        return () => {
            active = false;
            unlisten();
            unblock();
        };
    }, [files.hasUnsavedChanges, hasDirtyTabs, runAfterDiscard, session]);

    const initialize = useCallback((path: string, name?: string) => {
        const normalizedPath = normalizeFilePath(path);
        const hasFile = name !== undefined;
        let currentSession = sessionRef.current;
        const firstOpen = !currentSession;
        if (!currentSession) {
            currentSession = {generation: ++sessionGenerationRef.current};
            sessionRef.current = currentSession;
            setSession(currentSession);
        }
        captureViewState();
        if (firstOpen) {
            const directoryPath = hasFile ? parentFilePath(normalizedPath) : normalizedPath;
            tree.initialize(directoryPath, hasFile ? normalizedPath : undefined);
            setTreeCollapsed(isMobileViewport());
        }
        if (hasFile) {
            files.open(normalizedPath, String(name));
        }
    }, [captureViewState, files.open, tree.initialize]);

    const requestOpen = useCallback((path: string, name?: string) => {
        initialize(path, name);
    }, [initialize]);

    const requestClose = useCallback((force = false) => {
        if (force) {
            closeNow();
            return;
        }
        runAfterDiscard(closeNow);
    }, [closeNow, runAfterDiscard]);

    useImperativeHandle(ref, () => ({
        open: requestOpen,
        close: requestClose,
    }), [requestClose, requestOpen]);

    useEffect(() => () => {
        discardConfirmRef.current?.destroy();
        if (document.fullscreenElement === workspaceRef.current) {
            void document.exitFullscreen().catch(() => undefined);
        }
    }, []);

    useEffect(() => {
        const handleFullscreenChange = () => {
            setFullscreen(document.fullscreenElement === workspaceRef.current);
            window.requestAnimationFrame(() => aceRef.current?.resize?.());
        };
        document.addEventListener("fullscreenchange", handleFullscreenChange);
        return () => document.removeEventListener("fullscreenchange", handleFullscreenChange);
    }, []);

    useEffect(() => {
        aceRef.current = null;
        setAceError(undefined);
        setAceAttempt(0);
    }, [files.activeKey, files.activeTab?.generation]);

    const selectTreeNode = useCallback((node: FileEditorTreeNode) => {
        if (node.kind === "more") {
            tree.loadMore(node);
            return;
        }
        if (node.isDirectory) {
            tree.selectDirectory(node);
            return;
        }
        captureViewState();
        tree.selectFile(node);
        files.open(node.path, node.name);
        if (isMobileViewport()) {
            setTreeCollapsed(true);
        }
    }, [captureViewState, files.open, tree.loadMore, tree.selectDirectory, tree.selectFile]);

    const activateTab = useCallback((key: string) => {
        if (key === files.activeKey) {
            return;
        }
        captureViewState();
        const tab = files.getTab(key);
        files.activate(key);
        if (tab) {
            tree.selectPath(tab.path);
        }
    }, [captureViewState, files.activate, files.activeKey, files.getTab, tree.selectPath]);

    const closeTab = useCallback((key: string) => {
        const tab = files.getTab(key);
        if (!tab) {
            return;
        }
        runAfterDiscard(() => {
            if (key === files.activeKey) {
                captureViewState(key);
                aceRef.current = null;
            }
            viewStateRef.current.delete(key);
            const nextKey = files.close(key);
            const nextTab = nextKey ? files.getTab(nextKey) : undefined;
            if (nextTab) {
                tree.selectPath(nextTab.path);
            } else if (tree.targetDirectory) {
                tree.selectPath(tree.targetDirectory, true);
            }
        }, key, `“${tab.name}”尚未保存，关闭标签将丢失修改。`);
    }, [captureViewState, files.activeKey, files.close, files.getTab, runAfterDiscard, tree.selectPath, tree.targetDirectory]);

    const create = useCallback((mode: FileCreateMode, parentPath = tree.targetDirectory) => {
        if (!session) {
            return;
        }
        if (!parentPath) {
            message.info("请先选择目录");
            return;
        }
        const openForm = () => {
            formContextRef.current = {
                generation: session.generation,
                mode,
                parentPath,
            };
            formRef.current?.open(mode, parentPath, undefined, true);
        };
        if (document.fullscreenElement === workspaceRef.current) {
            void exitWorkspaceFullscreen().then(openForm).catch(() => message.error("退出全屏失败"));
            return;
        }
        openForm();
    }, [exitWorkspaceFullscreen, message, session, tree.targetDirectory]);

    const openPermissions = useCallback((node: FileEditorTreeNode) => {
        if (!session || !node.record) {
            return;
        }
        const open = () => permissionsRef.current?.open(node.path, node.record, true);
        if (document.fullscreenElement === workspaceRef.current) {
            void exitWorkspaceFullscreen().then(open).catch(() => message.error("退出全屏失败"));
            return;
        }
        open();
    }, [exitWorkspaceFullscreen, message, session]);

    const renameTreeNode = useCallback(async (node: FileEditorTreeNode, name: string) => {
        const value = name.trim();
        if (!session || !node.record || node.kind !== "entry") {
            return false;
        }
        if (!value) {
            message.error("请输入名称");
            return false;
        }
        if (value.length > 255) {
            message.error("名称不能超过255个字符");
            return false;
        }
        if (value === "." || value === ".." || /[\\/\0]/.test(value)) {
            message.error("名称不能包含路径分隔符");
            return false;
        }
        if (value === node.name) {
            return true;
        }
        const hasOpenTabs = files.tabs.some((tab) => (
            isFilePathWithin(tab.path, node.path)
        ));
        if (hasOpenTabs) {
            message.info(node.isDirectory
                ? "请先保存并关闭该目录下已打开的文件"
                : "请先保存并关闭该文件的标签");
            return false;
        }
        try {
            const response = await renameFile({body: {path: node.path, name: value}});
            if (!response) {
                return false;
            }
            if (response.code !== 200) {
                message.error(response.message || "重命名失败");
                return false;
            }
            message.success(response.message || "重命名成功");
            onMutation();
            if (sessionGenerationRef.current !== session.generation) {
                return true;
            }
            const targetPath = joinFilePath(node.parentPath, value);
            tree.selectPath(targetPath, node.isDirectory);
            await tree.revealCreated(node.parentPath, targetPath, node.isDirectory, true);
            return true;
        } catch (reason: unknown) {
            message.error(reason instanceof Error && reason.message ? reason.message : "重命名失败");
            return false;
        }
    }, [files.tabs, message, onMutation, session, tree.revealCreated, tree.selectPath]);

    const deleteTreeNode = useCallback((node: FileEditorTreeNode) => {
        if (!session || !node.record) {
            return;
        }
        const targetPath = node.path;
        if (files.tabs.some((tab) => isFilePathWithin(tab.path, targetPath))) {
            message.info(node.isDirectory
                ? "请先保存并关闭该目录下已打开的文件"
                : "请先保存并关闭该文件的标签");
            return;
        }
        modal.confirm({
            title: "删除",
            content: `确定将“${node.name}”移至回收站吗？`,
            okText: "删除",
            cancelText: "取消",
            okButtonProps: {danger: true},
            onOk: async () => {
                if (deletingPathsRef.current.has(targetPath)) {
                    return;
                }
                deletingPathsRef.current.add(targetPath);
                try {
                    const response = await deleteFile({body: {paths: [targetPath], recycle: true}});
                    if (!response) {
                        return;
                    }
                    if (response.code !== 200) {
                        message.error(response.message || "删除失败");
                        return;
                    }
                    message.success(response.message || "已移至回收站");
                    tree.selectPath(node.parentPath, true);
                    await tree.refresh(node.parentPath);
                    onMutation();
                } finally {
                    deletingPathsRef.current.delete(targetPath);
                }
            },
        });
    }, [files.tabs, message, modal, onMutation, session, tree.refresh, tree.selectPath]);

    const handlePermissionsSuccess = useCallback((targetPath: string) => {
        void tree.refresh(parentFilePath(targetPath));
        onMutation();
    }, [onMutation, tree.refresh]);

    const handleCreateSuccess = useCallback((result: FileFormResult) => {
        const context = formContextRef.current;
        formContextRef.current = undefined;
        onMutation();
        if (!context || context.generation !== sessionGenerationRef.current || result.mode !== context.mode) {
            return;
        }
        if (result.mode === "rename") {
            if (!result.sourcePath || !result.targetPath) {
                return;
            }
            tree.selectPath(result.targetPath, Boolean(context.isDirectory));
            void tree.revealCreated(
                context.parentPath,
                result.targetPath,
                Boolean(context.isDirectory),
                true,
            );
            return;
        }
        const targetPath = joinFilePath(context.parentPath, result.name);
        const isDirectory = result.mode === "directory";
        tree.selectPath(targetPath, isDirectory);
        void tree.revealCreated(context.parentPath, targetPath, isDirectory, true);
        if (!isDirectory) {
            captureViewState();
            files.open(targetPath, result.name);
            if (isMobileViewport()) {
                setTreeCollapsed(true);
            }
        }
    }, [captureViewState, files.open, onMutation, tree.revealCreated, tree.selectPath]);

    const refreshTree = useCallback(() => {
        void tree.refresh();
    }, [tree.refresh]);

    const copyTreeValue = useCallback(async (value: string, label: "名称" | "路径") => {
        if (await copyTextToClipboard(value)) {
            message.success(`${label}已复制`);
        } else {
            message.error(`复制${label}失败`);
        }
    }, [message]);

    const toggleTree = useCallback(() => {
        setTreeCollapsed((current) => !current);
    }, []);

    const saveCurrent = useCallback(() => {
        if (!files.activeKey) {
            return;
        }
        void files.save(files.activeKey);
    }, [files.activeKey, files.save]);
    saveCommandRef.current = saveCurrent;

    const reloadCurrent = useCallback(() => {
        const tab = files.activeTab;
        if (!tab) {
            return;
        }
        runAfterDiscard(() => {
            captureViewState(tab.key);
            files.reload(tab.key);
        }, tab.key, `“${tab.name}”尚未保存，重新加载将丢失修改。`);
    }, [captureViewState, files.activeTab, files.reload, runAfterDiscard]);

    const commands = useMemo(() => [{
        name: "saveFile",
        bindKey: {win: "Ctrl-S", mac: "Command-S"},
        exec: () => saveCommandRef.current(),
    }], []);

    const handleAceLoad = useCallback((instance: any) => {
        const key = files.activeKey;
        aceRef.current = instance;
        setAceError(undefined);
        instance.textInput?.getElement?.()?.setAttribute("aria-label", "文件内容");
        const viewState = key ? viewStateRef.current.get(key) : undefined;
        if (viewState?.cursor) {
            instance.moveCursorToPosition?.(viewState.cursor);
        }
        if (typeof viewState?.scrollTop === "number") {
            instance.session?.setScrollTop?.(viewState.scrollTop);
        }
        if (typeof viewState?.scrollLeft === "number") {
            instance.session?.setScrollLeft?.(viewState.scrollLeft);
        }
        window.requestAnimationFrame(() => instance.resize?.());
    }, [files.activeKey]);

    const handleAceError = useCallback((reason: Error) => {
        const key = files.activeKey;
        aceRef.current = null;
        if (key) {
            setAceError({key, message: reason?.message || "编辑器资源加载失败"});
        }
    }, [files.activeKey]);

    const retryAce = useCallback(() => {
        aceRef.current = null;
        setAceError(undefined);
        setAceAttempt((current) => current + 1);
    }, []);

    const toggleFullscreen = useCallback(async () => {
        const workspace = workspaceRef.current;
        if (!workspace) {
            return;
        }
        try {
            if (document.fullscreenElement === workspace) {
                await document.exitFullscreen();
            } else {
                await workspace.requestFullscreen();
            }
        } catch {
            message.error(fullscreen ? "退出全屏失败" : "进入全屏失败");
        }
    }, [fullscreen, message]);

    const afterOpenChange = useCallback((open: boolean) => {
        if (open) {
            window.requestAnimationFrame(() => aceRef.current?.resize?.());
        }
    }, []);

    const activeTab = files.activeTab;
    const activeMode = getEditorMode(activeTab?.name || "");
    const currentAceError = aceError && activeTab && aceError.key === activeTab.key
        ? aceError.message
        : "";
    const displayError = activeTab?.error || currentAceError;
    const editorStatus = displayError
        ? "加载失败"
        : activeTab?.loading
            ? "正在读取"
            : activeTab?.saving ? "正在保存" : activeTab?.dirty ? "未保存" : "已保存";
    const actionsDisabled = Boolean(activeTab?.loading || activeTab?.saving);
    const anyFileSaving = files.isSaving();
    const treeToggleIcon = treeCollapsed ? "MenuUnfoldOutlined" : "MenuFoldOutlined";
    const fullscreenIcon = fullscreen ? "FullscreenExitOutlined" : "FullscreenOutlined";
    const resolvedTheme = resolveTheme(dark);
    const body = session ? (
        <ConfigProvider getPopupContainer={getWorkspacePopupContainer}>
            <div
                ref={workspaceRef}
                className={`${styles.workspace} ${treeCollapsed ? "tree-collapsed" : ""}`}>
                {messageContextHolder}
                <FileEditorTree
                    treeData={tree.treeData}
                    expandedKeys={tree.expandedKeys}
                    selectedKeys={tree.selectedKeys}
                    loadedKeys={tree.loadedKeys}
                    loadingPaths={tree.loadingPaths}
                    initializing={tree.initializing}
                    locateToken={session.generation}
                    targetDirectory={tree.targetDirectory}
                    disabled={false}
                    tooltipsDisabled={fullscreen}
                    onExpand={tree.setExpandedKeys}
                    onSelect={selectTreeNode}
                    onLoadData={tree.loadData}
                    onRefresh={refreshTree}
                    onCreate={create}
                    onRename={renameTreeNode}
                    onPermissions={openPermissions}
                    onDelete={deleteTreeNode}
                    onCopy={copyTreeValue}
                    getPopupContainer={getWorkspacePopupContainer}
                    menuClassName={styles.treeMenu}/>

                <section className="file-editor-pane" aria-label="文件编辑区">
                    <div className="file-editor-toolbar">
                        <Tooltip
                            title={treeCollapsed ? "展开目录" : "收起目录"}
                            open={fullscreen ? false : undefined}
                            getPopupContainer={getWorkspacePopupContainer}>
                            <Button
                                type="text"
                                aria-label={treeCollapsed ? "展开目录树" : "收起目录树"}
                                icon={<Icon type={treeToggleIcon}/>}
                                onClick={toggleTree}/>
                        </Tooltip>
                        <FileEditorTabs
                            tabs={files.tabs}
                            activeKey={files.activeKey}
                            tooltipsDisabled={fullscreen}
                            getPopupContainer={getWorkspacePopupContainer}
                            onActivate={activateTab}
                            onClose={closeTab}/>
                        <div className="file-editor-toolbar-actions">
                            <Tooltip
                                title="重新加载"
                                open={fullscreen ? false : undefined}
                                getPopupContainer={getWorkspacePopupContainer}>
                                <Button
                                    className="file-editor-reload"
                                    type="text"
                                    disabled={!activeTab || activeTab.saving}
                                    aria-label="重新加载文件"
                                    icon={<Icon type="ReloadOutlined"/>}
                                    onClick={reloadCurrent}/>
                            </Tooltip>
                            <Tooltip
                                title="保存（Ctrl/Command + S）"
                                open={fullscreen ? false : undefined}
                                getPopupContainer={getWorkspacePopupContainer}>
                                <Button
                                    className={`file-editor-save ${activeTab?.dirty ? "is-dirty" : ""}`}
                                    type="text"
                                    disabled={!activeTab || !activeTab.version || actionsDisabled || Boolean(displayError)}
                                    loading={activeTab?.saving}
                                    aria-label="保存文件"
                                    icon={<Icon type="SaveOutlined"/>}
                                    onClick={saveCurrent}/>
                            </Tooltip>
                            <FileEditorSettings
                                key={fullscreen ? "fullscreen" : "windowed"}
                                value={preferences}
                                popupClassName={styles.settingsPopup}
                                onChange={setPreferences}/>
                            <Tooltip
                                title={fullscreen ? "退出全屏" : "全屏"}
                                open={fullscreen ? false : undefined}
                                getPopupContainer={getWorkspacePopupContainer}>
                                <Button
                                    className="file-editor-fullscreen"
                                    type="text"
                                    aria-label={fullscreen ? "退出全屏" : "全屏"}
                                    aria-pressed={fullscreen}
                                    icon={<Icon type={fullscreenIcon}/>}
                                    onClick={() => void toggleFullscreen()}/>
                            </Tooltip>
                            <span className="file-editor-toolbar-divider" aria-hidden/>
                            <Tooltip
                                title={anyFileSaving ? "文件保存中" : "关闭编辑器"}
                                open={fullscreen ? false : undefined}
                                getPopupContainer={getWorkspacePopupContainer}>
                                <span className="file-editor-close-wrap">
                                    <Button
                                        className="file-editor-close"
                                        type="text"
                                        disabled={anyFileSaving}
                                        aria-label="关闭文件编辑器"
                                        icon={<Icon type="CloseOutlined"/>}
                                        onClick={() => requestClose()}/>
                                </span>
                            </Tooltip>
                        </div>
                    </div>

                    <div id="file-editor-canvas" className="file-editor-canvas">
                        {activeTab && !activeTab.loading && !displayError && (
                            <AceEditor
                                key={`${activeTab.key}:${activeTab.generation}:${aceAttempt}`}
                                className="file-editor-ace"
                                defaultValue={files.getContent(activeTab.key)}
                                mode={activeMode}
                                theme={resolvedTheme}
                                width="100%"
                                height="100%"
                                fontSize={preferences.fontSize}
                                tabSize={preferences.tabSize}
                                readOnly={actionsDisabled}
                                showPrintMargin={false}
                                showLineNumbers={preferences.showLineNumbers}
                                wrapEnabled={preferences.wrapEnabled}
                                acePath={acePath}
                                commands={commands}
                                onChange={(value) => files.changeContent(activeTab.key, value)}
                                onLoad={handleAceLoad}
                                onError={handleAceError}
                                loadingContent={(
                                    <div className="file-editor-loading" role="status" aria-label="正在加载编辑器">
                                        <Spin size="default"/>
                                    </div>
                                )}/>
                        )}
                        {activeTab?.loading && (
                            <div className="file-editor-loading" role="status" aria-label="正在读取文件内容">
                                <Spin size="default"/>
                            </div>
                        )}
                        {!activeTab && (
                            <div className="file-editor-empty">
                                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择文件"/>
                            </div>
                        )}
                        {activeTab && displayError && (
                            <div className="file-editor-error">
                                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={displayError}>
                                    <Button
                                        className="file-editor-error-retry"
                                        type="text"
                                        aria-label={currentAceError ? "重新加载编辑器" : "重新读取文件"}
                                        icon={<Icon type="ReloadOutlined"/>}
                                        onClick={currentAceError ? retryAce : reloadCurrent}>
                                        {currentAceError ? "重新加载编辑器" : "重新读取文件"}
                                    </Button>
                                </Empty>
                            </div>
                        )}
                    </div>

                    <div className="file-editor-status" aria-live="polite">
                        <span
                            className="file-editor-status-path"
                            title={fullscreen ? undefined : activeTab?.path}>
                            {activeTab?.path || "请选择文件"}
                        </span>
                        {activeTab && (
                            <span className="file-editor-status-meta">
                                <span className="file-editor-status-detail">
                                    {`${activeMode.toUpperCase()} · UTF-8 · ${activeTab.contentLength.toLocaleString("zh-CN")} 字符`}
                                </span>
                                <span className="file-editor-status-state">{editorStatus}</span>
                            </span>
                        )}
                    </div>
                </section>
            </div>
        </ConfigProvider>
    ) : null;

    return (
        <>
            <ProModal
                title={<span>文件编辑器</span>}
                width="min(1480px, calc(100vw - 24px))"
                onCancel={() => requestClose()}
                modalProps={{
                    rootClassName: styles.modal,
                    open: Boolean(session),
                    footer: null,
                    closable: false,
                    keyboard: !fullscreen && !anyFileSaving,
                    mask: {closable: !anyFileSaving},
                    style: {
                        top: screens.md ? 16 : 8,
                        paddingBottom: screens.md ? 16 : 8,
                    },
                    styles: {
                        container: {padding: 0},
                        body: {padding: 0},
                    },
                    afterOpenChange,
                }}>
                {body}
            </ProModal>
            <FileForm ref={formRef} onSuccess={handleCreateSuccess}/>
            <FilePermissions ref={permissionsRef} onSuccess={handlePermissionsSuccess}/>
        </>
    );
});

FileEditor.displayName = "FileEditor";

export default memo(FileEditor);
