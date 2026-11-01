import {useCallback, useEffect, useRef, useState} from "react";
import {App} from "antd";
import {ossFile} from "@/service/api/oss";

type ClipboardMode = "copy" | "move";
type ClipboardItem = {key: string; name: string};
type Clipboard = {id: number; revision: number; mode: ClipboardMode; items: ClipboardItem[]};

interface ClipboardProps {
    id: number;
    revision: number;
    active: boolean;
    path: string;
    disabled: boolean;
    selectedRecords: any[];
    onChange: () => void;
    onBusyChange: (busy: boolean) => void;
}

// 复制或移动先记住源文件，前往目标目录后确认执行；对象路径始终原样传递。
const useOssClipboard = (props: ClipboardProps) => {
    const {id, revision, active, path, disabled, selectedRecords} = props;
    const {message, modal} = App.useApp();
    const [clipboard, setClipboard] = useState<Clipboard>();
    const [pasting, setPasting] = useState(false);
    const live = useRef(true);
    const running = useRef(false);
    const latest = useRef(props);
    const confirmation = useRef<ReturnType<typeof modal.confirm> | null>(null);
    latest.current = props;
    const source = clipboard?.id === id && clipboard.revision === revision ? clipboard : undefined;

    const cancelPending = useCallback(() => {
        const instance = confirmation.current;
        confirmation.current = null;
        instance?.destroy();
    }, []);

    useEffect(() => {
        cancelPending();
    }, [active, path, id, revision, cancelPending]);
    useEffect(() => { setClipboard(undefined); }, [id, revision]);
    useEffect(() => {
        live.current = true;
        return () => {
            live.current = false;
            cancelPending();
        };
    }, [cancelPending]);

    const setRecords = useCallback((records: readonly any[], mode: ClipboardMode) => {
        if (!active || disabled || running.current) return;
        const skipped = records.filter((item) => item.directory).length;
        const items = Array.from(new Map(records.filter((item) => !item.directory && typeof item.key === "string" && item.key)
            .map((item) => [item.key, {key: item.key, name: item.key.slice(item.key.lastIndexOf("/") + 1)}] as const)).values());
        if (!items.length) {
            if (skipped) {
                cancelPending();
                setClipboard(undefined);
                message.info("已跳过目录，请选择文件进行复制或移动");
            }
            return;
        }
        cancelPending();
        setClipboard({id, revision, mode, items});
        const text = mode === "move" ? `已选择移动 ${items.length} 项，请前往目标目录` : `已复制 ${items.length} 项，可前往目标目录粘贴`;
        message.success(text + (skipped ? `；已跳过 ${skipped} 个目录` : ""));
    }, [active, disabled, id, revision, cancelPending, message]);

    const copyRecords = useCallback((records: readonly any[]) => setRecords(records, "copy"), [setRecords]);
    const moveRecords = useCallback((records: readonly any[]) => setRecords(records, "move"), [setRecords]);
    const copySelected = useCallback(() => copyRecords(selectedRecords), [copyRecords, selectedRecords]);
    const moveSelected = useCallback(() => moveRecords(selectedRecords), [moveRecords, selectedRecords]);

    const paste = useCallback(() => {
        if (!active || disabled || running.current || !source?.items.length) return;
        let remaining = source.items.filter((item) => item.key !== path + item.name);
        if (!remaining.length) {
            message.warning(source.mode === "move" ? "不能移动到原目录" : "不能粘贴到原目录");
            return;
        }
        cancelPending();
        const operationName = source.mode === "move" ? "移动" : "复制";
        const total = remaining.length;
        const skipped = source.items.length - total;
        const content = () => `将 ${remaining.length} 个文件${source.mode === "move" ? "移动" : "复制"}到“/${path}”，同名文件会被覆盖。${skipped ? `已跳过原目录中的 ${skipped} 项。` : ""}`;
        let operation: Promise<void> | undefined;
        const instance = modal.confirm({
            title: source.mode === "move" ? "确认移动" : "确认粘贴",
            content: content(), zIndex: 1040, okText: source.mode === "move" ? "移动" : "粘贴", cancelText: "取消",
            mask: {closable: true},
            onCancel: () => { if (confirmation.current === instance) confirmation.current = null; },
            onOk: () => {
                if (running.current) return operation;
                const current = latest.current;
                if (!live.current || confirmation.current !== instance || !current.active || current.path !== path || current.id !== id || current.revision !== revision) return;
                if (current.disabled) {
                    message.info("请等待当前文件操作完成");
                    return Promise.reject(new Error("当前文件操作尚未完成"));
                }
                running.current = true;
                setPasting(true);
                current.onBusyChange(true);
                instance.update({keyboard: false, mask: {closable: false}, cancelButtonProps: {disabled: true}});
                operation = (async () => {
                    try {
                        // 顺序提交，成功项立即移出队列，部分失败时只重试剩余文件。
                        while (remaining.length) {
                            if (!live.current || latest.current.id !== id || latest.current.revision !== revision) throw new Error("存储配置已变更，请重新选择文件");
                            const item = remaining[0];
                            const response = await ossFile({body: {id, action: source.mode, key: item.key, destination: path + item.name}});
                            if (response?.code !== 200) throw new Error(response?.message || `${operationName}失败`);
                            remaining = remaining.slice(1);
                            if (live.current) setClipboard((value) => {
                                if (!value || value.id !== id || value.revision !== revision || value.mode !== source.mode) return value;
                                const items = value.items.filter((entry) => entry.key !== item.key);
                                return items.length ? {...value, items} : undefined;
                            });
                        }
                        if (confirmation.current === instance) confirmation.current = null;
                        if (live.current) message.success(`${operationName}完成，共 ${total} 项`);
                    } catch (reason: any) {
                        const error = `${operationName}“${remaining[0]?.name || "文件"}”失败：${reason?.message || "请重试"}`;
                        if (live.current) {
                            if (confirmation.current === instance) instance.update({content: `${error}。${total > remaining.length ? `已完成 ${total - remaining.length} 项。` : ""}${content()}`, okText: "重试"});
                            message.error(error);
                        }
                        throw reason;
                    } finally {
                        running.current = false;
                        latest.current.onBusyChange(false);
                        if (live.current) {
                            setPasting(false);
                            latest.current.onChange();
                            if (confirmation.current === instance) instance.update({keyboard: true, mask: {closable: true}, cancelButtonProps: {disabled: false}});
                        }
                        operation = undefined;
                    }
                })();
                return operation;
            },
        });
        confirmation.current = instance;
    }, [active, disabled, source, path, id, revision, cancelPending, message, modal]);

    useEffect(() => {
        if (!active) return;
        const shortcut = (event: KeyboardEvent) => {
            const target = event.target;
            if (disabled || running.current || confirmation.current || event.repeat || event.altKey || event.shiftKey || (!event.ctrlKey && !event.metaKey)
                || !(target instanceof HTMLElement) || !target.closest(".oss-files-modal")
                || target.isContentEditable || target.closest("input, textarea, select, [role='textbox'], [role='combobox'], [contenteditable='true']")
                || window.getSelection()?.toString()) return;
            const key = event.key.toLowerCase();
            if ((key === "c" || key === "x") && selectedRecords.length) {
                event.preventDefault();
                if (key === "x") moveSelected();
                else copySelected();
            } else if (key === "v" && source?.items.length) {
                event.preventDefault();
                paste();
            }
        };
        window.addEventListener("keydown", shortcut);
        return () => window.removeEventListener("keydown", shortcut);
    }, [active, disabled, selectedRecords.length, source, copySelected, moveSelected, paste]);

    return {count: source?.items.length || 0, mode: source?.mode || "copy", pasting, copyRecords, moveRecords, copySelected, moveSelected, paste, cancelPending};
};

export default useOssClipboard;
