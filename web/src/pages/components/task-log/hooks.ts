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
    viewportHeight: 520,
};

const toCursor = (value: any) => {
    const cursor = Number(value);
    return Number.isFinite(cursor) && cursor >= 0 ? cursor : 0;
};

const useLog = (
    request: (params: API.RequestParams) => Promise<any> = getTaskLog,
    body: Record<string, any> = {},
) => {
    const {message, modal} = App.useApp();
    const consoleRef = useRef<HTMLDivElement>(null);
    const taskRef = useRef<any>({});
    const defaultBodyRef = useRef(body);
    const requestBodyRef = useRef(body);
    defaultBodyRef.current = body;
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
        historyFrame: 0,
    });
    const loadHistoryRef = useRef<() => Promise<void>>(async () => undefined);
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
        Math.floor(scrollTop / logLayout.lineHeight) - logLayout.overscan,
    );
    const virtualEnd = Math.min(
        logs.length,
        Math.ceil((scrollTop + logLayout.viewportHeight) / logLayout.lineHeight)
        + logLayout.overscan,
    );
    const virtualHeight = logs.length * logLayout.lineHeight;

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
        if (control.historyFrame) {
            cancelAnimationFrame(control.historyFrame);
            control.historyFrame = 0;
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
            const response = await request({body: {...requestBodyRef.current, ...body}});
            if (control.requestId !== requestId || String(taskRef.current?.id) !== String(taskId)) {
                return null;
            }
            if (response?.code !== 200) {
                if (!silent) {
                    message.error(response?.message || "获取日志失败");
                }
                return null;
            }
            return (response?.data || {}) as any;
        } catch {
            if (control.requestId === requestId && !silent) {
                message.error("获取日志失败");
            }
            return null;
        } finally {
            if (control.requestId === requestId) {
                control.requesting = false;
                setLoading(false);
            }
        }
    }, [message, request]);

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
        const previousCursor = control.cursor;
        const nextCursor = toCursor(data.cursor);
        const nextLogs = Array.isArray(data.lines) ? data.lines : [];
        if (nextCursor < previousCursor) {
            await loadLatest(taskId, true);
            return;
        }
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
        let continueHistory = false;
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
            const currentStart = control.startCursor;
            const nextStart = toCursor(data.start_cursor);
            const previous = !data.end && Boolean(data.has_previous) && nextStart < currentStart;
            control.startCursor = nextStart;
            if (olderLogs.length === 0) {
                control.historyAnchor = null;
                hasPreviousRef.current = previous;
                continueHistory = previous;
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
            if (continueHistory && !control.historyFrame) {
                const continueLoading = () => {
                    control.historyFrame = 0;
                    const container = consoleRef.current;
                    if (!container || container.scrollTop > 64 || !hasPreviousRef.current) {
                        return;
                    }
                    if (control.requesting || control.historyLoading) {
                        control.historyFrame = requestAnimationFrame(continueLoading);
                        return;
                    }
                    void loadHistoryRef.current();
                };
                control.historyFrame = requestAnimationFrame(continueLoading);
            }
        }
    }, [requestLogs]);
    loadHistoryRef.current = loadHistory;

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
        const control = controlRef.current;
        if (control.scrollFrame) {
            cancelAnimationFrame(control.scrollFrame);
        }
        if (control.historyFrame) {
            cancelAnimationFrame(control.historyFrame);
        }
    }, []);

    const open = useCallback((record: any, requestBody?: Record<string, any>) => {
        if (record?.id === undefined || record?.id === null) {
            return false;
        }
        const control = controlRef.current;
        control.requestId += 1;
        control.requesting = false;
        control.historyLoading = false;
        resetPosition();
        taskRef.current = record;
        requestBodyRef.current = {...defaultBodyRef.current, ...(requestBody || {})};
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
        requestBodyRef.current = defaultBodyRef.current;
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
            title: "清理日志",
            content: `确定清理“${currentTask.name || "-"}”的全部日志吗？清理后无法恢复。`,
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
                    const response = await request({
                        body: {...requestBodyRef.current, id: taskId, action: "clear"},
                    });
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
    }, [clearing, message, modal, request, resetPosition]);

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
