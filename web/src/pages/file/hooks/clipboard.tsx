import {useCallback, useEffect, useRef, useState} from "react";
import {App} from "antd";
import {copyFile, getFileInfo, moveFile} from "@/service/api/file";
import {comparableFilePath, joinFilePath} from "../utils";

interface PasteConfirmState {
    id: symbol;
    destroy: () => void;
    resolve: (value: boolean) => void;
}

export interface UseFileClipboardOptions {
    path: string;
    loaded: boolean;
    loading: boolean;
    navigating: boolean;
    selectedRecords: readonly any[];
    hasBusySelection: boolean;
    trackTask: (taskId: string, title: string) => void;
}

const emptyItems: any[] = [];

const ignoreShortcut = (target: EventTarget | null) => {
    if (!(target instanceof HTMLElement)) {
        return false;
    }
    return target.isContentEditable
        || Boolean(target.closest("input, textarea, select, [contenteditable='true'], [role='textbox'], .ace_editor"))
        || Boolean(target.closest(".ant-modal-root, .ant-drawer-root"));
};

const useFileClipboard = ({
    path,
    loaded,
    loading,
    navigating,
    selectedRecords,
    hasBusySelection,
    trackTask,
}: UseFileClipboardOptions) => {
    const {message, modal} = App.useApp();
    const [clipboard, setClipboard] = useState<any>();
    const [pasting, setPasting] = useState(false);
    const generationRef = useRef(0);
    const startingRef = useRef(false);
    const confirmRef = useRef<PasteConfirmState | undefined>(undefined);
    const items = clipboard?.items ?? emptyItems;
    const mode = clipboard?.mode ?? "copy";

    const closeConfirm = useCallback(() => {
        const current = confirmRef.current;
        confirmRef.current = undefined;
        if (!current) {
            return;
        }
        current.resolve(false);
        current.destroy();
    }, []);

    const cancelPending = useCallback(() => {
        generationRef.current += 1;
        closeConfirm();
    }, [closeConfirm]);

    const setRecords = useCallback((records: readonly any[], nextMode: any) => {
        if (records.length === 0) {
            return;
        }
        if (startingRef.current) {
            message.info("正在粘贴文件");
            return;
        }
        const entries = Array.from(new Map(records.map((record) => {
            const sourcePath = joinFilePath(path, record.name);
            return [sourcePath, {
                sourcePath,
                sourceDirectory: path,
                name: record.name,
                isDir: record.is_dir,
            } satisfies any];
        })).values());
        setClipboard({mode: nextMode, items: entries});
        message.success(nextMode === "move"
            ? `已选择移动 ${entries.length} 项，请前往目标目录`
            : `已复制 ${entries.length} 项，可前往目标目录粘贴`);
    }, [message, path]);

    const copyRecords = useCallback((records: readonly any[]) => {
        setRecords(records, "copy");
    }, [setRecords]);

    const moveRecords = useCallback((records: readonly any[]) => {
        setRecords(records, "move");
    }, [setRecords]);

    const copySelected = useCallback(() => copyRecords(selectedRecords), [copyRecords, selectedRecords]);
    const moveSelected = useCallback(() => moveRecords(selectedRecords), [moveRecords, selectedRecords]);

    const confirmOverwrite = useCallback((names: string[], currentMode: any) => new Promise<boolean>((resolve) => {
        const confirmId = Symbol("paste-overwrite");
        let settled = false;
        const finish = (value: boolean) => {
            if (settled) {
                return;
            }
            settled = true;
            if (confirmRef.current?.id === confirmId) {
                confirmRef.current = undefined;
            }
            resolve(value);
        };
        const previewNames = names.slice(0, 3).join("、");
        const moving = currentMode === "move";
        const instance = modal.confirm({
            title: moving ? "确认移动" : "确认粘贴",
            content: `目标目录已有 ${names.length} 个同名项${previewNames ? `（${previewNames}${names.length > 3 ? "等" : ""}）` : ""}，继续后会替换现有内容。`,
            okText: moving ? "移动" : "粘贴",
            cancelText: "取消",
            mask: {closable: true},
            onOk: () => finish(true),
            onCancel: () => finish(false),
        });
        confirmRef.current = {id: confirmId, destroy: instance.destroy, resolve: finish};
    }), [modal]);

    const paste = useCallback(async () => {
        if (startingRef.current || items.length === 0) {
            return;
        }
        if (!loaded || navigating) {
            message.info("目录正在加载");
            return;
        }

        const targetPath = path;
        const currentMode = mode;
        const targetKey = comparableFilePath(targetPath);
        const snapshot = items.filter((item) => {
            const sourceDirectoryKey = comparableFilePath(item.sourceDirectory);
            const sourceKey = comparableFilePath(item.sourcePath);
            if (sourceDirectoryKey === targetKey) {
                return false;
            }
            return !item.isDir || (targetKey !== sourceKey && !targetKey.startsWith(`${sourceKey}/`));
        });
        const skipped = items.length - snapshot.length;
        if (snapshot.length === 0) {
            message.warning(currentMode === "move"
                ? "不能移动到原目录或源目录内部"
                : "不能粘贴到原目录或源目录内部");
            return;
        }

        startingRef.current = true;
        setPasting(true);
        const generation = ++generationRef.current;
        try {
            const conflicts: string[] = [];
            for (const item of snapshot) {
                const response = await getFileInfo({body: {path: joinFilePath(targetPath, item.name)}});
                if (generationRef.current !== generation) {
                    return;
                }
                if (response?.code === 200) {
                    conflicts.push(item.name);
                    continue;
                }
                if (!response) {
                    return;
                }
                if (!String(response.message || "").includes("不存在")) {
                    message.error(response.message || "检查目标路径失败");
                    return;
                }
            }

            if (conflicts.length > 0 && !await confirmOverwrite(conflicts, currentMode)) {
                return;
            }
            if (generationRef.current !== generation) {
                return;
            }

            const transfer = currentMode === "move" ? moveFile : copyFile;
            const operationName = currentMode === "move" ? "移动" : "复制";
            const response = await transfer({
                body: {
                    source_paths: snapshot.map((item) => item.sourcePath),
                    target_path: targetPath,
                },
            });
            if (response?.code === 200 && typeof response.data === "string" && response.data) {
                trackTask(response.data, snapshot.length === 1
                    ? `${operationName} ${snapshot[0].name}`
                    : `${operationName} ${snapshot.length} 项`);
                if (skipped > 0) {
                    message.warning(`已创建${operationName}任务，跳过 ${skipped} 项`);
                } else {
                    message.success(`已创建${operationName}任务，共 ${snapshot.length} 项`);
                }
                const submittedPaths = new Set(snapshot.map((item) => item.sourcePath));
                setClipboard((current) => {
                    if (!current || current.mode !== currentMode) {
                        return current;
                    }
                    const remaining = current.items.filter((item) => !submittedPaths.has(item.sourcePath));
                    return remaining.length > 0 ? {...current, items: remaining} : undefined;
                });
            } else if (response) {
                message.error(response.message || `创建${operationName}任务失败`);
            }
        } finally {
            closeConfirm();
            startingRef.current = false;
            setPasting(false);
        }
    }, [closeConfirm, confirmOverwrite, items, loaded, message, mode, navigating, path, trackTask]);

    useEffect(() => {
        const handleShortcut = (event: KeyboardEvent) => {
            if (event.repeat || event.altKey || event.shiftKey || (!event.ctrlKey && !event.metaKey) || ignoreShortcut(event.target)) {
                return;
            }
            if (window.getSelection()?.toString()) {
                return;
            }
            const key = event.key.toLowerCase();
            if ((key === "c" || key === "x") && selectedRecords.length > 0 && !hasBusySelection && !loading && !navigating && !pasting) {
                event.preventDefault();
                if (key === "x") {
                    moveSelected();
                } else {
                    copySelected();
                }
                return;
            }
            if (key === "v" && items.length > 0 && loaded && !navigating && !pasting) {
                event.preventDefault();
                void paste();
            }
        };
        window.addEventListener("keydown", handleShortcut);
        return () => window.removeEventListener("keydown", handleShortcut);
    }, [copySelected, hasBusySelection, items.length, loaded, loading, moveSelected, navigating, paste, pasting, selectedRecords.length]);

    useEffect(() => () => cancelPending(), [cancelPending]);

    return {
        count: items.length,
        mode,
        pasting,
        copyRecords,
        moveRecords,
        copySelected,
        moveSelected,
        paste,
        cancelPending,
    };
};

export default useFileClipboard;
