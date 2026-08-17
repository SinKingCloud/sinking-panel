import React, {
    forwardRef,
    memo,
    useCallback,
    useEffect,
    useImperativeHandle,
    useRef,
    useState,
} from "react";
import {App, Button, Progress} from "antd";
import {Icon, Title, useTheme} from "sinking-antd";
import {cancelSystemTask, getSystemTask} from "@/service/api/system";
import type {SystemTaskRecord} from "@/service/api/system";
import {readFileStorage, updateFileStorage} from "../hooks/file-storage";
import useStyles from "./task-center.styles";

interface TrackedTask extends SystemTaskRecord {
    title: string;
    missingCount?: number;
    unknown?: boolean;
}

export interface FileTaskCenterRef {
    track: (id: string, title: string) => void;
}

interface FileTaskCenterProps {
    onSettled: () => void;
}

interface DragState {
    pointerId: number;
    startX: number;
    startY: number;
    left: number;
    top: number;
    width: number;
    height: number;
}

const terminalStatuses = new Set([2, 3, 4]);
const statusMeta: Record<number, {label: string; className: string}> = {
    0: {label: "排队中", className: "running"},
    1: {label: "处理中", className: "running"},
    2: {label: "已完成", className: "completed"},
    3: {label: "失败", className: "failed"},
    4: {label: "已取消", className: "canceled"},
};

const isTerminalStatus = (status: unknown) => terminalStatuses.has(Number(status));

const createPendingTask = (id: string, title: string): TrackedTask => ({
    id,
    title,
    name: title,
    status: 0,
    progress: 0,
    message: "等待任务开始",
    data: undefined,
    start_time: 0,
    end_time: 0,
    create_time: Date.now(),
    update_time: Date.now(),
});

const restoreTasks = (): TrackedTask[] => {
    const stored = readFileStorage().tasks;
    if (!Array.isArray(stored)) {
        return [];
    }
    const unique = new Map<string, TrackedTask>();
    stored.forEach((item) => {
        if (!item || typeof item.id !== "string" || !item.id || typeof item.title !== "string") {
            return;
        }
        unique.set(item.id, createPendingTask(item.id, item.title));
    });
    return Array.from(unique.values());
};

const persistTasks = (tasks: TrackedTask[]) => {
    const pending = tasks
        .filter((task) => !isTerminalStatus(task.status))
        .map(({id, title}) => ({id, title}));
    updateFileStorage((current) => ({
        ...current,
        tasks: pending.length > 0 ? pending : undefined,
    }));
};

const FileTaskCenter = forwardRef<FileTaskCenterRef, FileTaskCenterProps>(({onSettled}, ref) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const {message} = App.useApp();
    const mountedRef = useRef(true);
    const tasksRef = useRef<TrackedTask[]>([]);
    const dismissTimersRef = useRef(new Map<string, number>());
    const notifiedIdsRef = useRef(new Set<string>());
    const [tasks, setTasks] = useState<TrackedTask[]>(restoreTasks);
    const [panelClosed, setPanelClosed] = useState(false);
    const panelRef = useRef<HTMLElement | null>(null);
    const dragRef = useRef<DragState | undefined>(undefined);
    const [dragging, setDragging] = useState(false);
    const [panelPosition, setPanelPosition] = useState<{left?: number; top?: number}>({});
    const active = tasks.some((task) => !isTerminalStatus(task.status));

    const commitTasks = useCallback((next: TrackedTask[], persist = false) => {
        tasksRef.current = next;
        if (persist) {
            persistTasks(next);
        }
        if (mountedRef.current) {
            setTasks(next);
        }
    }, []);

    const clearDismissTimer = useCallback((id?: string) => {
        if (id) {
            const timer = dismissTimersRef.current.get(id);
            if (timer !== undefined) {
                window.clearTimeout(timer);
                dismissTimersRef.current.delete(id);
            }
            return;
        }
        dismissTimersRef.current.forEach((timer) => window.clearTimeout(timer));
        dismissTimersRef.current.clear();
    }, []);

    const scheduleDismiss = useCallback((id: string) => {
        if (dismissTimersRef.current.has(id)) {
            return;
        }
        clearDismissTimer(id);
        const timer = window.setTimeout(() => {
            dismissTimersRef.current.delete(id);
            const next = tasksRef.current.filter((task) => task.id !== id);
            commitTasks(next, true);
        }, 2000);
        dismissTimersRef.current.set(id, timer);
    }, [clearDismissTimer, commitTasks]);

    useImperativeHandle(ref, () => ({
        track: (id, title) => {
            if (!id) {
                return;
            }
            setPanelClosed(false);
            clearDismissTimer(id);
            notifiedIdsRef.current.delete(id);
            const current = tasksRef.current;
            const existing = current.find((task) => task.id === id);
            const next = existing
                ? current.map((task) => task.id === id
                        ? isTerminalStatus(task.status)
                        ? createPendingTask(id, title)
                        : {...task, title, name: task.name || title}
                    : task)
                : [createPendingTask(id, title), ...current];
            commitTasks(next, true);
        },
    }), [clearDismissTimer, commitTasks]);

    useEffect(() => {
        mountedRef.current = true;
        tasksRef.current = tasks;
        return () => {
            mountedRef.current = false;
            clearDismissTimer();
        };
    }, [clearDismissTimer, tasks]);

    useEffect(() => {
        if (!active) {
            return;
        }
        let stopped = false;
        let timer = 0;

        const poll = async () => {
            const pending = tasksRef.current.filter((task) => !isTerminalStatus(task.status));
            if (stopped || pending.length === 0) {
                return;
            }
            try {
                const responses = await Promise.all(pending.map(async (task) => ({
                    task,
                    response: await getSystemTask({body: {id: task.id}}),
                })));
                if (stopped) {
                    return;
                }
                const next = tasksRef.current.slice();
                const settled: TrackedTask[] = [];
                responses.forEach(({task, response}) => {
                    const index = next.findIndex((item) => item.id === task.id);
                    if (index < 0) {
                        return;
                    }
                    const current = next[index];
                    let updated = current;
                    if (response?.code === 200 && response.data) {
                        updated = {
                            ...response.data,
                            status: Number(response.data.status) as TrackedTask["status"],
                            title: current.title,
                            missingCount: 0,
                        };
                    } else if (response && String(response.message || "").includes("任务不存在")) {
                        const missingCount = (current.missingCount || 0) + 1;
                        updated = missingCount >= 3
                            ? {...current, status: 4, unknown: true, message: "任务状态已结束，结果未知", missingCount}
                            : {...current, missingCount};
                    }
                    next[index] = updated;
                    if (!isTerminalStatus(current.status) && isTerminalStatus(updated.status)) {
                        settled.push(updated);
                    }
                });
                commitTasks(next, settled.length > 0);
                settled.forEach((task) => {
                    if (notifiedIdsRef.current.has(task.id)) {
                        return;
                    }
                    notifiedIdsRef.current.add(task.id);
                    if (task.status === 2) {
                        message.success(`${task.name || task.title}完成`);
                    } else if (task.status === 3) {
                        message.error(task.message || `${task.name || task.title}失败`);
                    }
                    scheduleDismiss(task.id);
                    onSettled();
                });
            } finally {
                if (!stopped && tasksRef.current.some((task) => !isTerminalStatus(task.status))) {
                    timer = window.setTimeout(poll, 700);
                }
            }
        };

        void poll();
        return () => {
            stopped = true;
            window.clearTimeout(timer);
        };
    }, [active, commitTasks, message, onSettled, scheduleDismiss]);

    useEffect(() => {
        tasks.forEach((task) => {
            if (isTerminalStatus(task.status)) {
                scheduleDismiss(task.id);
            }
        });
    }, [scheduleDismiss, tasks]);

    const cancel = useCallback(async (id: string) => {
        const response = await cancelSystemTask({body: {id}});
        if (response && response.code !== 200) {
            message.error(response.message || "取消任务失败");
        }
    }, [message]);

    const handleDragStart = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
        if (event.button !== 0 || (event.target as HTMLElement).closest("button")) {
            return;
        }
        const panel = panelRef.current;
        if (!panel) {
            return;
        }
        const rect = panel.getBoundingClientRect();
        dragRef.current = {
            pointerId: event.pointerId,
            startX: event.clientX,
            startY: event.clientY,
            left: rect.left,
            top: rect.top,
            width: rect.width,
            height: rect.height,
        };
        event.currentTarget.setPointerCapture(event.pointerId);
        setDragging(true);
    }, []);

    const handleDragMove = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
        const drag = dragRef.current;
        if (!drag || drag.pointerId !== event.pointerId) {
            return;
        }
        const left = Math.min(
            Math.max(8, drag.left + event.clientX - drag.startX),
            Math.max(8, window.innerWidth - drag.width - 8),
        );
        const top = Math.min(
            Math.max(8, drag.top + event.clientY - drag.startY),
            Math.max(8, window.innerHeight - drag.height - 8),
        );
        setPanelPosition({left, top});
    }, []);

    const handleDragEnd = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
        if (dragRef.current?.pointerId !== event.pointerId) {
            return;
        }
        dragRef.current = undefined;
        if (event.currentTarget.hasPointerCapture(event.pointerId)) {
            event.currentTarget.releasePointerCapture(event.pointerId);
        }
        setDragging(false);
    }, []);

    if (tasks.length === 0 || panelClosed) {
        return null;
    }

    return (
        <section
            ref={panelRef}
            className={styles.root}
            style={panelPosition.left === undefined ? undefined : {...panelPosition, transform: "none"}}
            role="status"
            aria-live="polite"
            aria-label="文件任务进度">
            <div
                className={`file-task-progress-panel-header ${dragging ? "is-dragging" : ""}`}
                onPointerDown={handleDragStart}
                onPointerMove={handleDragMove}
                onPointerUp={handleDragEnd}
                onPointerCancel={handleDragEnd}>
                <div className="file-task-progress-panel-title">
                    <Title size="small">文件任务</Title>
                </div>
                <Button
                    type="text"
                    className="file-task-progress-panel-close"
                    aria-label="关闭任务面板"
                    title="关闭任务面板"
                    icon={<Icon type="CloseOutlined"/>}
                    onClick={() => setPanelClosed(true)}/>
            </div>
            <div className="file-task-progress-list">
                {tasks.map((task) => {
                    const meta = task.unknown
                        ? {label: "状态未知", className: "canceled"}
                        : statusMeta[Number(task.status)] || statusMeta[0];
                    const rawProgress = Number(task.progress);
                    const progress = Number.isFinite(rawProgress)
                        ? Math.max(0, Math.min(100, rawProgress))
                        : 0;
                    const terminal = isTerminalStatus(task.status);
                    const displayName = task.name || task.title;
                    return (
                        <article
                            className={`file-task-progress-item file-task-progress-${meta.className}`}
                            key={task.id}>
                            <div className="file-task-progress-header">
                                <span className="file-task-progress-status-dot" aria-hidden="true"/>
                                <span className="file-task-progress-name" title={displayName}>{displayName}</span>
                                {!terminal && (
                                    <span className="file-task-progress-value">{Math.floor(progress)}%</span>
                                )}
                                {!terminal && (
                                    <Button
                                        type="text"
                                        className="file-task-progress-cancel"
                                        aria-label={`取消${displayName}`}
                                        title="取消任务"
                                        icon={<Icon type="CloseOutlined"/>}
                                        onClick={() => void cancel(task.id)}/>
                                )}
                            </div>
                            {meta.className !== "running" && (
                                <div className={`file-task-progress-state ${meta.className}`}>{meta.label}</div>
                            )}
                            {!terminal && (
                                <Progress
                                    className="file-task-progress-bar"
                                    percent={progress}
                                    size="small"
                                    showInfo={false}/>
                            )}
                            <div className="file-task-progress-message" title={task.message || undefined}>
                                {task.message || "等待处理"}
                            </div>
                        </article>
                    );
                })}
            </div>
        </section>
    );
});

FileTaskCenter.displayName = "FileTaskCenter";

export default memo(FileTaskCenter);
