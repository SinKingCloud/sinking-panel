import {startTransition, useCallback, useEffect, useLayoutEffect, useRef, useState} from "react";
import {App} from "antd";
import {getTaskLog} from "@/service/api/task";

const pageSize = 300;
const maxDisplayLines = 10000;
const maxCachedLines = maxDisplayLines * 5;
const pollInterval = 1000;

export const logLayout = {
    lineHeight: 20,
    overscan: 24,
    paddingHorizontal: 18,
    paddingTop: 16,
    paddingBottom: 20,
    viewportHeight: 520,
};

const toCursor = (value: any) => {
    const cursor = Number(value);
    return Number.isFinite(cursor) && cursor >= 0 ? cursor : 0;
};

const useLog = () => {
    const {message, modal} = App.useApp();
    const consoleRef = useRef<HTMLDivElement>(null);
    const taskRef = useRef<any>({});
    const controlRef = useRef<any>({
        requestId: 0,
        requesting: false,
        historyLoading: false,
        cursor: 0,
        startCursor: 0,
        lineCount: 0,
        stickToBottom: true,
        historyAnchor: null,
        scrollFrame: 0,
    });
    const [taskId, setTaskId] = useState<any>();
    const [fileName, setFileName] = useState("");
    const [logs, setLogs] = useState<string[]>([]);
    const hasPreviousRef = useRef(false);
    const [loading, setLoading] = useState(false);
    const [clearing, setClearing] = useState(false);
    const [scrollTop, setScrollTop] = useState(0);

    const virtual = logs.length > maxDisplayLines;
    const virtualStart = Math.max(
        0,
        Math.floor((scrollTop - logLayout.paddingTop) / logLayout.lineHeight) - logLayout.overscan,
    );
    const virtualEnd = Math.min(
        logs.length,
        Math.ceil((scrollTop + logLayout.viewportHeight - logLayout.paddingTop) / logLayout.lineHeight)
        + logLayout.overscan,
    );
    const virtualHeight = logLayout.paddingTop + logs.length * logLayout.lineHeight + logLayout.paddingBottom;

    const resetPosition = useCallback(() => {
        const control = controlRef.current;
        control.cursor = 0;
        control.startCursor = 0;
        control.lineCount = 0;
        control.stickToBottom = true;
        control.historyAnchor = null;
        if (control.scrollFrame) {
            cancelAnimationFrame(control.scrollFrame);
            control.scrollFrame = 0;
        }
        setScrollTop(0);
    }, []);

    const requestLogs = useCallback(async (body: any, silent = false) => {
        const taskId = body?.id;
        const control = controlRef.current;
        if (taskId === undefined || taskId === null || (silent && control.requesting)) {
            return null;
        }

        const requestId = ++control.requestId;
        control.requesting = true;
        if (!silent) {
            setLoading(true);
        }
        try {
            const response = await getTaskLog({body});
            if (control.requestId !== requestId || String(taskRef.current?.id) !== String(taskId)) {
                return null;
            }
            if (response?.code !== 200) {
                if (!silent) {
                    message.error(response?.message || "获取任务日志失败");
                }
                return null;
            }
            return (response?.data || {}) as any;
        } catch {
            if (control.requestId === requestId && !silent) {
                message.error("获取任务日志失败");
            }
            return null;
        } finally {
            if (control.requestId === requestId) {
                control.requesting = false;
                setLoading(false);
            }
        }
    }, [message]);

    const loadLatest = useCallback(async (taskId: any, silent = false) => {
        const data = await requestLogs({
            id: taskId,
            page_size: pageSize,
            ...(silent ? {after: 0} : {}),
        }, silent);
        if (!data) {
            return;
        }

        const nextLogs = Array.isArray(data.lines) ? data.lines : [];
        resetPosition();
        const control = controlRef.current;
        control.cursor = toCursor(data.cursor);
        control.startCursor = toCursor(data.start_cursor);
        control.lineCount = nextLogs.length;
        const previous = Boolean(data.has_previous);
        hasPreviousRef.current = previous;
        const commit = () => {
            setFileName(String(data.file_name || ""));
            setLogs(nextLogs);
        };
        if (silent) {
            startTransition(commit);
        } else {
            commit();
        }
    }, [requestLogs, resetPosition]);

    const loadFollow = useCallback(async (taskId: any) => {
        const control = controlRef.current;
        const data = await requestLogs({
            id: taskId,
            after: control.cursor,
            page_size: pageSize,
        }, true);
        if (!data) {
            return;
        }
        if (data.end) {
            await loadLatest(taskId, true);
            return;
        }

        const nextCursor = toCursor(data.cursor);
        const nextLogs = Array.isArray(data.lines) ? data.lines : [];
        if (nextLogs.length === 0) {
            control.cursor = nextCursor;
            return;
        }
        if (control.lineCount + nextLogs.length > maxCachedLines) {
            await loadLatest(taskId, true);
            return;
        }
        control.cursor = nextCursor;
        control.lineCount += nextLogs.length;
        startTransition(() => setLogs((current) => [...current, ...nextLogs]));
    }, [loadLatest, requestLogs]);

    const loadHistory = useCallback(async () => {
        const control = controlRef.current;
        const taskId = taskRef.current?.id;
        if (!taskId || !hasPreviousRef.current || control.historyLoading || control.requesting) {
            return;
        }

        const container = consoleRef.current;
        if (container) {
            control.historyAnchor = {
                scrollTop: container.scrollTop,
                scrollHeight: container.scrollHeight,
            };
        }
        control.historyLoading = true;
        try {
            const data = await requestLogs({
                id: taskId,
                before: control.startCursor,
                page_size: pageSize,
            }, true);
            if (!data) {
                control.historyAnchor = null;
                return;
            }

            const olderLogs = Array.isArray(data.lines) ? data.lines : [];
            const previous = Boolean(data.has_previous);
            control.startCursor = toCursor(data.start_cursor);
            if (olderLogs.length === 0) {
                control.historyAnchor = null;
                hasPreviousRef.current = previous;
                return;
            }
            if (control.lineCount + olderLogs.length > maxCachedLines) {
                control.historyAnchor = null;
                hasPreviousRef.current = false;
                return;
            }

            control.lineCount += olderLogs.length;
            setLogs((current) => [...olderLogs, ...current]);
            hasPreviousRef.current = previous;
        } finally {
            control.historyLoading = false;
        }
    }, [requestLogs]);

    const handleScroll = useCallback(() => {
        const container = consoleRef.current;
        if (!container) {
            return;
        }
        const control = controlRef.current;
        control.stickToBottom = container.scrollHeight - container.scrollTop - container.clientHeight <= 24;
        if (virtual && !control.scrollFrame) {
            control.scrollFrame = requestAnimationFrame(() => {
                control.scrollFrame = 0;
                setScrollTop(container.scrollTop);
            });
        }
        if (container.scrollTop <= 64) {
            void loadHistory();
        }
    }, [loadHistory, virtual]);

    useEffect(() => {
        if (taskId === undefined || taskId === null) {
            return undefined;
        }

        let stopped = false;
        let timer = 0;
        const poll = () => {
            timer = window.setTimeout(async () => {
                if (stopped) {
                    return;
                }
                if (document.visibilityState === "visible") {
                    await loadFollow(taskId);
                }
                if (!stopped) {
                    poll();
                }
            }, pollInterval);
        };
        const handleVisibility = () => {
            if (document.visibilityState === "visible") {
                void loadFollow(taskId);
            }
        };
        poll();
        document.addEventListener("visibilitychange", handleVisibility);
        return () => {
            stopped = true;
            window.clearTimeout(timer);
            document.removeEventListener("visibilitychange", handleVisibility);
        };
    }, [loadFollow, taskId]);

    useLayoutEffect(() => {
        const container = consoleRef.current;
        const control = controlRef.current;
        if (!container) {
            return;
        }
        if (control.historyAnchor) {
            const anchor = control.historyAnchor;
            container.scrollTop = anchor.scrollTop + container.scrollHeight - anchor.scrollHeight;
            control.historyAnchor = null;
            return;
        }
        if (logs.length > 0 && control.stickToBottom) {
            container.scrollTop = container.scrollHeight;
        }
    }, [logs]);

    useEffect(() => () => {
        const frame = controlRef.current.scrollFrame;
        if (frame) {
            cancelAnimationFrame(frame);
        }
    }, []);

    const open = useCallback((record: any) => {
        if (record?.id === undefined || record?.id === null) {
            return false;
        }
        const control = controlRef.current;
        control.requestId += 1;
        control.requesting = false;
        control.historyLoading = false;
        resetPosition();
        taskRef.current = record;
        hasPreviousRef.current = false;
        setTaskId(record.id);
        setFileName("");
        setLogs([]);
        setLoading(false);
        setClearing(false);
        void loadLatest(record.id);
        return true;
    }, [loadLatest, resetPosition]);

    const reset = useCallback(() => {
        const control = controlRef.current;
        control.requestId += 1;
        control.requesting = false;
        control.historyLoading = false;
        resetPosition();
        taskRef.current = {};
        hasPreviousRef.current = false;
        setTaskId(undefined);
        setFileName("");
        setLogs([]);
        setLoading(false);
        setClearing(false);
    }, [resetPosition]);

    const refresh = useCallback(() => {
        const taskId = taskRef.current?.id;
        if (taskId !== undefined && taskId !== null) {
            void loadLatest(taskId);
        }
    }, [loadLatest]);

    const clear = useCallback(() => {
        const currentTask = taskRef.current;
        if (currentTask?.id === undefined || currentTask?.id === null || clearing) {
            return;
        }
        modal.confirm({
            title: "清理任务日志",
            content: `确定清理任务“${currentTask.name || "-"}”的全部日志吗？清理后无法恢复。`,
            okText: "清理",
            cancelText: "取消",
            okButtonProps: {danger: true},
            mask: {closable: true},
            onOk: async () => {
                const taskId = currentTask.id;
                const control = controlRef.current;
                const requestId = ++control.requestId;
                control.requesting = true;
                setLoading(false);
                setClearing(true);
                const loadingKey = `task-log-clear-${String(taskId)}`;
                message.loading({key: loadingKey, content: "正在清理日志...", duration: 0});
                try {
                    const response = await getTaskLog({body: {id: taskId, action: "clear"}});
                    if (control.requestId !== requestId || String(taskRef.current?.id) !== String(taskId)) {
                        return;
                    }
                    if (response?.code !== 200) {
                        message.error(response?.message || "清理日志失败");
                        return;
                    }
                    resetPosition();
                    setLogs([]);
                    hasPreviousRef.current = false;
                    message.success(response?.message || "清理成功");
                } catch {
                    if (control.requestId === requestId) {
                        message.error("清理日志失败");
                    }
                } finally {
                    message.destroy(loadingKey);
                    if (control.requestId === requestId) {
                        control.requesting = false;
                        setClearing(false);
                    }
                }
            },
        } as any);
    }, [clearing, message, modal, resetPosition]);

    return {
        consoleRef,
        fileName,
        logs,
        loading,
        clearing,
        virtual,
        virtualStart,
        virtualEnd,
        virtualHeight,
        open,
        reset,
        refresh,
        clear,
        handleScroll,
    };
};

export default useLog;
