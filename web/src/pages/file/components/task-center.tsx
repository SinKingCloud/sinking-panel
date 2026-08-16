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
import {Icon, useTheme} from "sinking-antd";
import {cancelSystemTask, getSystemTask} from "@/service/api/file";
import type {SystemTaskRecord} from "@/service/api/file";
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

const terminalStatuses = new Set([2, 3, 4]);
const statusMeta: Record<number, {label: string; className: string}> = {
    0: {label: "排队中", className: "running"},
    1: {label: "处理中", className: "running"},
    2: {label: "已完成", className: "completed"},
    3: {label: "失败", className: "failed"},
    4: {label: "已取消", className: "canceled"},
};

const FileTaskCenter = forwardRef<FileTaskCenterRef, FileTaskCenterProps>(({onSettled}, ref) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const {message} = App.useApp();
    const mountedRef = useRef(true);
    const taskRef = useRef<TrackedTask | undefined>(undefined);
    const dismissTimerRef = useRef<number | undefined>(undefined);
    const notifiedIdRef = useRef("");
    const [task, setTask] = useState<TrackedTask>();
    const active = Boolean(task && !terminalStatuses.has(task.status));

    const commitTask = useCallback((next?: TrackedTask) => {
        taskRef.current = next;
        if (mountedRef.current) {
            setTask(next);
        }
    }, []);

    const clearDismissTimer = useCallback(() => {
        if (dismissTimerRef.current !== undefined) {
            window.clearTimeout(dismissTimerRef.current);
            dismissTimerRef.current = undefined;
        }
    }, []);

    const scheduleDismiss = useCallback((id: string) => {
        clearDismissTimer();
        dismissTimerRef.current = window.setTimeout(() => {
            dismissTimerRef.current = undefined;
            if (taskRef.current?.id === id) {
                commitTask(undefined);
            }
        }, 2000);
    }, [clearDismissTimer, commitTask]);

    useImperativeHandle(ref, () => ({
        track: (id, title) => {
            if (!id) {
                return;
            }
            clearDismissTimer();
            notifiedIdRef.current = "";
            const current = taskRef.current;
            if (current?.id === id) {
                commitTask({...current, title});
                return;
            }
            commitTask({
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
        },
    }), [clearDismissTimer, commitTask]);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            clearDismissTimer();
        };
    }, [clearDismissTimer]);

    useEffect(() => {
        if (!active) {
            return;
        }
        let stopped = false;
        let timer = 0;

        const poll = async () => {
            const current = taskRef.current;
            if (stopped || !current || terminalStatuses.has(current.status)) {
                return;
            }
            try {
                const response = await getSystemTask({body: {id: current.id}});
                if (stopped || taskRef.current?.id !== current.id) {
                    return;
                }
                let next = current;
                if (response?.code === 200 && response.data) {
                    next = {...response.data, title: current.title, missingCount: 0};
                } else if (response && String(response.message || "").includes("任务不存在")) {
                    const missingCount = (current.missingCount || 0) + 1;
                    next = missingCount >= 3
                        ? {...current, status: 4, unknown: true, message: "任务状态已结束，结果未知", missingCount}
                        : {...current, missingCount};
                }
                commitTask(next);
                if (terminalStatuses.has(next.status) && notifiedIdRef.current !== next.id) {
                    notifiedIdRef.current = next.id;
                    if (next.status === 2) {
                        message.success(`${next.name || next.title}完成`);
                    } else if (next.status === 3) {
                        message.error(next.message || `${next.name || next.title}失败`);
                    }
                    scheduleDismiss(next.id);
                    onSettled();
                    return;
                }
            } finally {
                if (!stopped && taskRef.current && !terminalStatuses.has(taskRef.current.status)) {
                    timer = window.setTimeout(poll, 700);
                }
            }
        };

        void poll();
        return () => {
            stopped = true;
            window.clearTimeout(timer);
        };
    }, [active, commitTask, message, onSettled, scheduleDismiss]);

    const cancel = useCallback(async () => {
        const current = taskRef.current;
        if (!current) {
            return;
        }
        const response = await cancelSystemTask({body: {id: current.id}});
        if (response && response.code !== 200) {
            message.error(response.message || "取消任务失败");
        }
    }, [message]);

    if (!task) {
        return null;
    }

    const meta = task.unknown
        ? {label: "状态未知", className: "canceled"}
        : statusMeta[task.status] || statusMeta[0];
    const rawProgress = Number(task.progress);
    const progress = Number.isFinite(rawProgress)
        ? Math.max(0, Math.min(100, rawProgress))
        : 0;
    const terminal = terminalStatuses.has(task.status);
    const displayName = task.name || task.title;

    return (
        <section
            className={`${styles.root} file-task-progress-${meta.className}`}
            role="status"
            aria-live="polite"
            aria-label="文件任务进度">
            <div className="file-task-progress-header">
                <span className="file-task-progress-status-dot" aria-hidden="true"/>
                <span className="file-task-progress-name" title={displayName}>{displayName}</span>
                {!terminal && (
                    <span className="file-task-progress-value">
                        {`${Math.floor(progress)}%`}
                    </span>
                )}
                {!terminal && (
                    <Button
                        type="text"
                        className="file-task-progress-cancel"
                        aria-label={`取消${displayName}`}
                        title="取消任务"
                        icon={<Icon type="CloseOutlined"/>}
                        onClick={() => void cancel()}/>
                )}
            </div>
            <div className="file-task-progress-body">
                <div className={`file-task-progress-state ${meta.className}`}>{meta.label}</div>
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
            </div>
        </section>
    );
});

FileTaskCenter.displayName = "FileTaskCenter";

export default memo(FileTaskCenter);
