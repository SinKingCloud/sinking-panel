import {useCallback, useEffect, useRef, useState} from "react";
import {App} from "antd";
import {getFilePreviewKind} from "@/pages/components/file-preview/media";
import type {FilePreviewKind} from "@/pages/components/file-preview/media";
import {getFileInfo, updateFile} from "@/service/api/file";
import {
    contentByteSize,
    editorReadPageSize,
    editorTreeKey,
    maxEditorContentSize,
} from "../components/editor/utils";

export interface FileEditorTab {
    key: string;
    generation: number;
    path: string;
    name: string;
    previewKind?: FilePreviewKind;
    contentLength: number;
    version: string;
    loading: boolean;
    saving: boolean;
    dirty: boolean;
    error: string;
}

export type FileEditorDocument = FileEditorTab;

export interface UseFileEditorDocumentOptions {
    onMutation: () => void;
    message: ReturnType<typeof App.useApp>["message"];
}

interface FileEditorTabRuntime {
    generation: number;
    loadToken: number;
    saveToken: number;
    path: string;
    name: string;
    previewKind?: FilePreviewKind;
    content: string;
    baseline: string;
    version: string;
    loading: boolean;
    saving: boolean;
    dirty: boolean;
}

const getErrorMessage = (error: unknown, fallback: string) => (
    error instanceof Error && error.message ? error.message : fallback
);

const createTab = (key: string, runtime: FileEditorTabRuntime): FileEditorTab => ({
    key,
    generation: runtime.generation,
    path: runtime.path,
    name: runtime.name,
    previewKind: runtime.previewKind,
    contentLength: runtime.content.length,
    version: runtime.version,
    loading: runtime.loading,
    saving: runtime.saving,
    dirty: runtime.dirty,
    error: "",
});

const useFileEditorDocument = ({onMutation, message}: UseFileEditorDocumentOptions) => {
    const generationRef = useRef(0);
    const runtimesRef = useRef(new Map<string, FileEditorTabRuntime>());
    const tabsRef = useRef<FileEditorTab[]>([]);
    const activeKeyRef = useRef<string | undefined>(undefined);
    const onMutationRef = useRef(onMutation);
    const contentUpdateFrameRef = useRef<number | undefined>(undefined);
    const pendingContentUpdatesRef = useRef(new Map<string, {
        generation: number;
        contentLength: number;
        dirty: boolean;
    }>());
    const [tabs, setTabs] = useState<FileEditorTab[]>([]);
    const [activeKey, setActiveKey] = useState<string>();

    useEffect(() => {
        onMutationRef.current = onMutation;
    }, [onMutation]);

    useEffect(() => () => {
        if (contentUpdateFrameRef.current !== undefined) {
            window.cancelAnimationFrame(contentUpdateFrameRef.current);
        }
        pendingContentUpdatesRef.current.clear();
        runtimesRef.current.clear();
        tabsRef.current = [];
        activeKeyRef.current = undefined;
    }, []);

    const commitTabs = useCallback((next: FileEditorTab[]) => {
        tabsRef.current = next;
        setTabs(next);
    }, []);

    const updateTab = useCallback((
        key: string,
        generation: number,
        update: (tab: FileEditorTab) => FileEditorTab,
    ) => {
        let changed = false;
        const next = tabsRef.current.map((tab) => {
            if (tab.key !== key || tab.generation !== generation) {
                return tab;
            }
            changed = true;
            return update(tab);
        });
        if (changed) {
            commitTabs(next);
        }
        return changed;
    }, [commitTabs]);

    const flushContentUpdates = useCallback(() => {
        contentUpdateFrameRef.current = undefined;
        const pending = pendingContentUpdatesRef.current;
        if (pending.size === 0) {
            return;
        }
        pendingContentUpdatesRef.current = new Map();
        let next = tabsRef.current;
        pending.forEach((update, key) => {
            next = next.map((tab) => tab.key === key && tab.generation === update.generation
                ? {
                    ...tab,
                    contentLength: update.contentLength,
                    dirty: update.dirty,
                }
                : tab);
        });
        commitTabs(next);
    }, [commitTabs]);

    const load = useCallback(async (
        key: string,
        generation: number,
        loadToken: number,
        path: string,
    ) => {
        let cursor = 0;
        let fileVersion = "";
        const chunks: string[] = [];
        let loadedBytes = 0;

        const getCurrentRuntime = () => {
            const current = runtimesRef.current.get(key);
            return current?.generation === generation && current.loadToken === loadToken
                ? current
                : undefined;
        };

        try {
            while (true) {
                const response = await getFileInfo({body: {
                    path,
                    read: true,
                    cursor,
                    page_size: editorReadPageSize,
                    version: fileVersion,
                }});
                if (!getCurrentRuntime()) {
                    return;
                }
                if (response?.code !== 200 || !response.data) {
                    throw new Error(response?.message || "读取文件内容失败");
                }

                const data = response.data;
                if (data.is_dir) {
                    throw new Error("目录不支持内容编辑");
                }
                if (Number(data.size) > maxEditorContentSize) {
                    throw new Error("文件超过 8 MiB，无法在线编辑");
                }

                const responseVersion = String(data.version || "");
                if (!fileVersion) {
                    if (!responseVersion) {
                        throw new Error("无法获取文件版本，请重新加载");
                    }
                    fileVersion = responseVersion;
                } else if (responseVersion !== fileVersion) {
                    throw new Error("文件内容已发生变化，请重新加载");
                }

                const nextCursor = Number(data.next_cursor);
                const chunk = String(data.content || "");
                if (!Number.isSafeInteger(nextCursor) || nextCursor < cursor) {
                    throw new Error("文件游标异常，请重新加载");
                }
                loadedBytes += contentByteSize(chunk);
                if (nextCursor > maxEditorContentSize || loadedBytes > maxEditorContentSize) {
                    throw new Error("文件超过 8 MiB，无法在线编辑");
                }
                chunks.push(chunk);

                if (data.eof) {
                    break;
                }
                if (nextCursor === cursor) {
                    throw new Error("文件读取未能继续，请重新加载");
                }
                cursor = nextCursor;
            }

            const runtime = getCurrentRuntime();
            if (!runtime) {
                return;
            }
            const value = chunks.join("");
            runtime.content = value;
            runtime.baseline = value;
            runtime.version = fileVersion;
            runtime.loading = false;
            runtime.dirty = false;
            updateTab(key, generation, (tab) => ({
                ...tab,
                contentLength: value.length,
                version: fileVersion,
                loading: false,
                dirty: false,
                error: "",
            }));
        } catch (reason: unknown) {
            const runtime = getCurrentRuntime();
            if (!runtime) {
                return;
            }
            runtime.content = "";
            runtime.baseline = "";
            runtime.version = "";
            runtime.loading = false;
            runtime.dirty = false;
            updateTab(key, generation, (tab) => ({
                ...tab,
                contentLength: 0,
                version: "",
                loading: false,
                dirty: false,
                error: getErrorMessage(reason, "读取文件内容失败"),
            }));
        }
    }, [updateTab]);

    const activate = useCallback((key: string) => {
        if (!runtimesRef.current.has(key) || activeKeyRef.current === key) {
            return;
        }
        activeKeyRef.current = key;
        setActiveKey(key);
    }, []);

    const open = useCallback((path: string, name: string) => {
        const key = editorTreeKey(path);
        if (runtimesRef.current.has(key)) {
            activeKeyRef.current = key;
            setActiveKey(key);
            return key;
        }

        const generation = ++generationRef.current;
        const previewKind = getFilePreviewKind(name);
        const runtime: FileEditorTabRuntime = {
            generation,
            loadToken: 1,
            saveToken: 0,
            path,
            name,
            previewKind,
            content: "",
            baseline: "",
            version: "",
            loading: !previewKind,
            saving: false,
            dirty: false,
        };
        runtimesRef.current.set(key, runtime);
        commitTabs([...tabsRef.current, createTab(key, runtime)]);
        activeKeyRef.current = key;
        setActiveKey(key);
        if (!previewKind) {
            void load(key, generation, runtime.loadToken, path);
        }
        return key;
    }, [commitTabs, load]);

    const close = useCallback((key: string) => {
        const runtime = runtimesRef.current.get(key);
        if (!runtime || runtime.saving) {
            return activeKeyRef.current;
        }

        const currentTabs = tabsRef.current;
        const index = currentTabs.findIndex((tab) => tab.key === key);
        if (index < 0) {
            return activeKeyRef.current;
        }

        runtimesRef.current.delete(key);
        const nextTabs = currentTabs.filter((tab) => tab.key !== key);
        commitTabs(nextTabs);
        if (activeKeyRef.current !== key) {
            return activeKeyRef.current;
        }

        const nextActiveKey = nextTabs[index]?.key || nextTabs[index - 1]?.key;
        activeKeyRef.current = nextActiveKey;
        setActiveKey(nextActiveKey);
        return nextActiveKey;
    }, [commitTabs]);

    const reload = useCallback((key = activeKeyRef.current) => {
        if (!key) {
            return false;
        }
        const previous = runtimesRef.current.get(key);
        if (!previous || previous.saving) {
            return false;
        }

        const generation = ++generationRef.current;
        const runtime: FileEditorTabRuntime = {
            generation,
            loadToken: previous.loadToken + 1,
            saveToken: previous.saveToken + 1,
            path: previous.path,
            name: previous.name,
            previewKind: previous.previewKind,
            content: "",
            baseline: "",
            version: "",
            loading: !previous.previewKind,
            saving: false,
            dirty: false,
        };
        runtimesRef.current.set(key, runtime);
        updateTab(key, previous.generation, (tab) => ({
            ...tab,
            generation,
            previewKind: previous.previewKind,
            contentLength: 0,
            version: "",
            loading: !previous.previewKind,
            saving: false,
            dirty: false,
            error: "",
        }));
        if (!runtime.previewKind) {
            void load(key, generation, runtime.loadToken, runtime.path);
        }
        return true;
    }, [load, updateTab]);

    const changeContent = useCallback((keyOrValue: string, nextValue?: string) => {
        const key = nextValue === undefined ? activeKeyRef.current : keyOrValue;
        const value = nextValue === undefined ? keyOrValue : nextValue;
        if (!key) {
            return;
        }
        const runtime = runtimesRef.current.get(key);
        if (!runtime || runtime.loading || runtime.previewKind) {
            return;
        }
        runtime.content = value;
        runtime.dirty = value !== runtime.baseline;
        pendingContentUpdatesRef.current.set(key, {
            generation: runtime.generation,
            contentLength: value.length,
            dirty: runtime.dirty,
        });
        if (contentUpdateFrameRef.current === undefined) {
            contentUpdateFrameRef.current = window.requestAnimationFrame(flushContentUpdates);
        }
    }, [flushContentUpdates]);

    const save = useCallback(async (key = activeKeyRef.current) => {
        if (!key) {
            return false;
        }
        const runtime = runtimesRef.current.get(key);
        if (!runtime || runtime.loading || runtime.saving || runtime.previewKind) {
            return false;
        }
        flushContentUpdates();
        if (!runtime.version) {
            message.error("文件内容尚未加载完成");
            return false;
        }

        const submittedContent = runtime.content;
        if (contentByteSize(submittedContent) > maxEditorContentSize) {
            message.error("文件超过 8 MiB，无法在线保存");
            return false;
        }

        const generation = runtime.generation;
        const saveToken = ++runtime.saveToken;
        const submittedVersion = runtime.version;
        const path = runtime.path;
        runtime.saving = true;
        updateTab(key, generation, (tab) => ({...tab, saving: true}));

        const getCurrentRuntime = () => {
            const current = runtimesRef.current.get(key);
            return current?.generation === generation && current.saveToken === saveToken
                ? current
                : undefined;
        };

        try {
            const response = await updateFile({body: {
                path,
                content: submittedContent,
                version: submittedVersion,
            }});
            const current = getCurrentRuntime();
            if (!current) {
                return false;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "保存文件失败");
                return false;
            }

            const nextVersion = String(response.data?.version || "");
            current.baseline = submittedContent;
            current.dirty = current.content !== submittedContent;
            current.version = nextVersion;
            updateTab(key, generation, (tab) => ({
                ...tab,
                version: nextVersion,
                dirty: current.dirty,
                error: nextVersion ? "" : "文件已保存，但未返回新版本，请重新加载后继续编辑",
            }));
            message.success(response.message || "保存成功");
            onMutationRef.current();
            return true;
        } catch (reason: unknown) {
            if (getCurrentRuntime()) {
                message.error(getErrorMessage(reason, "保存文件失败"));
            }
            return false;
        } finally {
            const current = getCurrentRuntime();
            if (current) {
                current.saving = false;
                updateTab(key, generation, (tab) => ({...tab, saving: false}));
            }
        }
    }, [flushContentUpdates, message, updateTab]);

    const reset = useCallback(() => {
        if (contentUpdateFrameRef.current !== undefined) {
            window.cancelAnimationFrame(contentUpdateFrameRef.current);
            contentUpdateFrameRef.current = undefined;
        }
        pendingContentUpdatesRef.current.clear();
        runtimesRef.current.clear();
        tabsRef.current = [];
        activeKeyRef.current = undefined;
        setTabs([]);
        setActiveKey(undefined);
    }, []);

    const hasUnsavedChanges = useCallback((key?: string) => {
        if (key) {
            return Boolean(runtimesRef.current.get(key)?.dirty);
        }
        return Array.from(runtimesRef.current.values()).some((runtime) => runtime.dirty);
    }, []);

    const isSaving = useCallback((key?: string) => {
        if (key) {
            return Boolean(runtimesRef.current.get(key)?.saving);
        }
        return Array.from(runtimesRef.current.values()).some((runtime) => runtime.saving);
    }, []);

    const getContent = useCallback((key = activeKeyRef.current) => (
        key ? runtimesRef.current.get(key)?.content || "" : ""
    ), []);

    const getTab = useCallback((key: string) => (
        tabsRef.current.find((tab) => tab.key === key)
    ), []);

    const activeTab = tabs.find((tab) => tab.key === activeKey);

    return {
        tabs,
        activeKey,
        activeTab,
        open,
        activate,
        close,
        reload,
        save,
        changeContent,
        reset,
        hasUnsavedChanges,
        isSaving,
        getContent,
        getTab,
    };
};

export default useFileEditorDocument;
