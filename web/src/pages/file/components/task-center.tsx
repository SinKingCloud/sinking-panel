import React, {
    forwardRef,
    useCallback,
    useEffect,
    useImperativeHandle,
    useRef,
    useState,
} from "react";
import {App, Button, Drawer, Empty, Grid, Progress} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {
    cancelSystemTask,
    getSystemTask,
} from "@/service/api/file";
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
    0: {label: "等待中", className: "running"},
    1: {label: "进行中", className: "running"},
    2: {label: "已完成", className: "completed"},
    3: {label: "失败", className: "failed"},
    4: {label: "已取消", className: "canceled"},
};

const FileTaskCenter = forwardRef<FileTaskCenterRef, FileTaskCenterProps>(({onSettled}, ref) => {
    const theme = useTheme();
    const screens = Grid.useBreakpoint();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const {message} = App.useApp();
    const mountedRef = useRef(true);
    const tasksRef = useRef<TrackedTask[]>([]);
    const notifiedRef = useRef(new Set<string>());
    const [open, setOpen] = useState(false);
    const [tasks, setTasks] = useState<TrackedTask[]>([]);
    const hasActiveTasks = tasks.some((item) => !terminalStatuses.has(item.status));

    const commitTasks = useCallback((next: TrackedTask[]) => {
        tasksRef.current = next;
        if (mountedRef.current) {
            setTasks(next);
        }
    }, []);

    useImperativeHandle(ref, () => ({
        track: (id, title) => {
            if (!id) {
                return;
            }
            const current = tasksRef.current;
            const existing = current.find((item) => item.id === id);
            const next = existing
                ? current.map((item) => item.id === id ? {...item, title} : item)
                : [{
                    id,
                    title,
                    name: title,
                    status: 0 as const,
                    progress: 0,
                    message: "等待任务开始",
                    data: undefined,
                    start_time: 0,
                    end_time: 0,
                    create_time: Date.now(),
                    update_time: Date.now(),
                }, ...current];
            commitTasks(next);
            setOpen(true);
        },
    }), [commitTasks]);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
        };
    }, []);

    useEffect(() => {
        if (!hasActiveTasks) {
            return;
        }
        let stopped = false;
        let timer = 0;

        const poll = async () => {
            if (stopped) {
                return;
            }
            try {
                const current = tasksRef.current;
                const pending = current.filter((item) => !terminalStatuses.has(item.status));
                const responses = await Promise.all(pending.map(async (item) => ({
                    item,
                    response: await getSystemTask({body: {id: item.id}}),
                })));
                if (stopped) {
                    return;
                }
                const updates = new Map<string, TrackedTask>();
                responses.forEach(({item, response}) => {
                    if (response?.code === 200 && response.data) {
                        updates.set(item.id, {...response.data, title: item.title, missingCount: 0});
                        return;
                    }
                    if (!response || !String(response.message || "").includes("任务不存在")) {
                        updates.set(item.id, item);
                        return;
                    }
                    const missingCount = (item.missingCount || 0) + 1;
                    updates.set(item.id, missingCount >= 3
                        ? {...item, status: 4, unknown: true, message: "任务状态已结束，结果未知", missingCount}
                        : {...item, missingCount});
                });
                const next = tasksRef.current.map((item) => updates.get(item.id) || item);
                commitTasks(next);
                next.forEach((item) => {
                    if (!terminalStatuses.has(item.status) || notifiedRef.current.has(item.id)) {
                        return;
                    }
                    notifiedRef.current.add(item.id);
                    if (item.status === 2) {
                        message.success(`${item.title}完成`);
                    } else if (item.status === 3) {
                        message.error(item.message || `${item.title}失败`);
                    }
                    onSettled();
                });
            } finally {
                if (!stopped && tasksRef.current.some((item) => !terminalStatuses.has(item.status))) {
                    timer = window.setTimeout(poll, 700);
                }
            }
        };

        void poll();
        return () => {
            stopped = true;
            window.clearTimeout(timer);
        };
    }, [commitTasks, hasActiveTasks, message, onSettled]);

    const removeCompleted = () => commitTasks(tasksRef.current.filter((item) => !terminalStatuses.has(item.status)));
    const cancel = async (task: TrackedTask) => {
        const response = await cancelSystemTask({body: {id: task.id}});
        if (response && response.code !== 200) {
            message.error(response.message || "取消任务失败");
        }
    };

    return (
        <Drawer
            open={open}
            rootClassName={styles.root}
            title="文件任务"
            placement={screens.md ? "right" : "bottom"}
            size={screens.md ? 400 : "72dvh"}
            extra={tasks.some((item) => terminalStatuses.has(item.status)) ? (
                <Button type="text" size="small" onClick={removeCompleted}>清除记录</Button>
            ) : null}
            onClose={() => setOpen(false)}>
            {tasks.length === 0 ? (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
            ) : (
                <div className="file-task-list">
                    {tasks.map((task) => {
                        const meta = task.unknown
                            ? {label: "状态未知", className: "canceled"}
                            : statusMeta[task.status] || statusMeta[0];
                        const rawProgress = Number(task.progress);
                        const progress = Number.isFinite(rawProgress)
                            ? Math.max(0, Math.min(100, rawProgress))
                            : 0;
                        return (
                            <article className="file-task-item" key={task.id} aria-busy={!terminalStatuses.has(task.status)}>
                                <div className="file-task-head">
                                    <span className="file-task-title" title={task.title}>{task.title}</span>
                                    <span className={`file-task-state ${meta.className}`}>{meta.label}</span>
                                </div>
                                <div className="file-task-message" title={task.message}>{task.message || "等待处理"}</div>
                                <Progress
                                    className="file-task-progress"
                                    percent={progress}
                                    format={task.status === 1
                                        ? (value) => `${Math.floor(value || 0)}%`
                                        : undefined}
                                    size="small"
                                    status={task.status === 3 ? "exception" : task.status === 2 ? "success" : "active"}
                                    showInfo={task.status !== 0}/>
                                {task.status === 1 && (
                                    <div className="file-task-actions">
                                        <Button type="text" danger icon={<Icon type="CloseOutlined"/>} onClick={() => cancel(task)}>
                                            取消
                                        </Button>
                                    </div>
                                )}
                            </article>
                        );
                    })}
                </div>
            )}
        </Drawer>
    );
});

FileTaskCenter.displayName = "FileTaskCenter";

export default React.memo(FileTaskCenter);
