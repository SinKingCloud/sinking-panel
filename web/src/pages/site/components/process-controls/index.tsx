import {memo, useCallback, useEffect, useRef, useState} from "react";
import {App, Button, Spin} from "antd";
import {createStyles} from "antd-style";
import {Icon, useTheme} from "sinking-antd";
import {controlSiteProcess, getSiteProcessStatus} from "@/service/api/site";

type ProcessAction = "start" | "stop" | "restart";
type ProcessState = "starting" | "running" | "stopping" | "restarting" | "stopped" | "failed";

interface ProcessStatus {
    state: ProcessState;
    pid: number;
    restart_count: number;
    max_retries: number;
    retry_exhausted: boolean;
    error?: string;
}

interface ProcessControlsProps {
    siteId: number;
    disabled?: boolean;
}

const processStates: Record<ProcessState, {label: string; color: string}> = {
    starting: {label: "正在启动", color: "processing"},
    running: {label: "运行中", color: "success"},
    stopping: {label: "正在停止", color: "warning"},
    restarting: {label: "等待重试", color: "processing"},
    stopped: {label: "已停止", color: "default"},
    failed: {label: "运行失败", color: "error"},
};

const useStyles = createStyles(({css, token}: any, {compact}: {compact: boolean}) => ({
    toolbar: css`
        min-width: 0;
        margin-bottom: ${compact ? 14 : 18}px;
        padding: ${compact ? "10px 12px" : "12px 14px"};
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorFillQuaternary};
    `,
    row: css`
        display: flex;
        align-items: center;
        justify-content: space-between;
        flex-wrap: wrap;
        gap: 10px 16px;
    `,
    status: css`
        min-width: 0;
        display: flex;
        align-items: center;
        gap: 12px;

        .status-label {
            color: ${token.colorTextTertiary};
            font-size: 12px;
            white-space: nowrap;
        }

        .status-value {
            --status-color: ${token.colorTextQuaternary};
            display: inline-flex;
            align-items: center;
            gap: 8px;
            min-height: 24px;
            color: ${token.colorText};
            font-size: 13px;
            font-weight: 500;
            white-space: nowrap;
        }

        .status-value[data-tone="success"] { --status-color: ${token.colorSuccess}; }
        .status-value[data-tone="processing"] { --status-color: ${token.colorPrimary}; }
        .status-value[data-tone="warning"] { --status-color: ${token.colorWarning}; }
        .status-value[data-tone="error"] { --status-color: ${token.colorError}; }

        .status-dot {
            width: 6px;
            height: 6px;
            flex: none;
            margin: 0 2px;
            border-radius: 50%;
            background: var(--status-color);
            box-shadow: 0 0 0 3px color-mix(in srgb, var(--status-color) 12%, transparent);
        }

        .ant-spin {
            display: inline-flex;
        }

        .ant-spin-dot {
            font-size: 12px;
        }
    `,
    actions: css`
        display: flex;
        flex-wrap: wrap;
        gap: ${compact ? 5 : 6}px;

        .service-action.ant-btn {
            width: auto;
            min-width: max-content;
            height: ${compact ? 28 : 32}px;
            padding: 0 ${compact ? 8 : 10}px;
            gap: 5px;
            font-size: ${compact ? 11 : 12}px;
        }
    `,
    metadata: css`
        display: flex;
        align-items: center;
        gap: 8px 20px;
        flex-wrap: wrap;
        margin: ${compact ? 8 : 10}px 0 0;
        padding: 0;
        font-size: 12px;
        line-height: 20px;

        > div {
            display: flex;
            align-items: baseline;
            gap: 8px;
            min-width: 0;
        }

        dt { color: ${token.colorTextTertiary}; }

        dd {
            margin: 0;
            color: ${token.colorTextSecondary};
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
            font-variant-numeric: tabular-nums;
        }

        .retry-limit {
            color: ${token.colorTextTertiary};
        }
    `,
    error: css`
        display: flex;
        align-items: flex-start;
        gap: 7px;
        margin-top: 10px;
        padding-top: 10px;
        border-top: 1px solid ${token.colorSplit};
        color: ${token.colorError};
        font-size: 12px;
        line-height: 1.6;

        > .anticon { flex: none; margin-top: 3px; }
        > span:last-child { min-width: 0; overflow-wrap: anywhere; white-space: pre-wrap; }
    `,
}));

const ProcessControls = ({siteId, disabled = false}: ProcessControlsProps) => {
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const {message} = App.useApp();
    const siteRef = useRef<number | null>(null);
    const versionRef = useRef(0);
    const actionRef = useRef(false);
    const pendingRef = useRef<Promise<void>>(Promise.resolve());
    const [status, setStatus] = useState<ProcessStatus | null>(null);
    const [loading, setLoading] = useState(true);
    const [action, setAction] = useState<ProcessAction | null>(null);
    const [loadError, setLoadError] = useState("");

    const isCurrent = useCallback((version: number) => versionRef.current === version && siteRef.current !== null, []);

    // 状态查询与操作串行执行，操作开始时使之前的查询结果失效。
    const enqueue = useCallback((request: () => Promise<void>) => {
        const pending = pendingRef.current.then(request, request);
        pendingRef.current = pending.catch(() => undefined);
        return pending;
    }, []);

    const loadStatus = useCallback(async (version: number) => {
        if (!isCurrent(version)) {
            return;
        }
        try {
            const response = await getSiteProcessStatus({body: {id: siteRef.current}});
            if (!isCurrent(version)) {
                return;
            }
            if (response?.code === 200) {
                setStatus(response.data);
                setLoadError("");
            } else {
                setLoadError(response?.message || "获取进程状态失败");
            }
        } catch {
            if (isCurrent(version)) {
                setLoadError("获取进程状态失败");
            }
        } finally {
            if (isCurrent(version)) {
                setLoading(false);
            }
        }
    }, [isCurrent]);

    useEffect(() => {
        versionRef.current += 1;
        siteRef.current = siteId;
        actionRef.current = false;
        setStatus(null);
        setAction(null);
        setLoadError("");
        setLoading(true);
        return () => {
            versionRef.current += 1;
            siteRef.current = null;
        };
    }, [siteId]);

    useEffect(() => {
        if (action) {
            return;
        }
        const version = versionRef.current;
        let closed = false;
        let timer = 0;
        const poll = async () => {
            if (closed || !isCurrent(version)) {
                return;
            }
            if (document.visibilityState === "visible") {
                await enqueue(() => loadStatus(version));
            }
            if (!closed && isCurrent(version)) {
                timer = window.setTimeout(poll, 1500);
            }
        };
        void poll();
        return () => {
            closed = true;
            window.clearTimeout(timer);
        };
    }, [siteId, action, enqueue, isCurrent, loadStatus]);

    const control = async (nextAction: ProcessAction) => {
        if (!siteRef.current || actionRef.current || disabled) {
            return;
        }
        const id = siteRef.current;
        const version = ++versionRef.current;
        actionRef.current = true;
        setAction(nextAction);
        setLoadError("");
        try {
            await enqueue(async () => {
                if (!isCurrent(version)) {
                    return;
                }
                try {
                    const response = await controlSiteProcess({body: {id, action: nextAction}});
                    if (!isCurrent(version)) {
                        return;
                    }
                    if (response?.code === 200) {
                        setStatus(response.data);
                    } else {
                        message.error(response?.message || "进程操作失败");
                    }
                } catch {
                    if (isCurrent(version)) {
                        message.error("进程操作失败");
                    }
                }
            });
        } finally {
            if (isCurrent(version)) {
                actionRef.current = false;
                setAction(null);
            }
        }
    };

    const currentState = status ? processStates[status.state] : undefined;
    const running = status && ["starting", "running", "restarting"].includes(status.state);
    const stopping = status?.state === "stopping";
    const unavailable = disabled || !status || loading || Boolean(action);

    return <div className={styles.toolbar}>
        <div className={styles.row}>
            <div className={styles.status} aria-live="polite">
                <span className="status-label">运行状态</span>
                <span className="status-value" data-tone={status?.retry_exhausted ? "error" : currentState?.color}>
                    {loading ? <Spin size="small"/> : <span className="status-dot" aria-hidden="true"/>}
                    {loading ? "获取中" : status?.retry_exhausted ? "已停止重试" : currentState?.label || "状态未知"}
                </span>
            </div>
            <div className={styles.actions} role="group" aria-label="进程操作">
                <Button className="service-action" color="default" variant="filled" style={{borderColor: "transparent"}} htmlType="button" icon={<Icon type="CaretRightOutlined"/>} loading={action === "start"}
                    disabled={unavailable || Boolean(running) || stopping} onClick={() => void control("start")}>启动</Button>
                <Button className="service-action" color="danger" variant="filled" style={{borderColor: "transparent"}} htmlType="button" icon={<Icon type="PauseOutlined"/>} loading={action === "stop"}
                    disabled={unavailable || stopping || !running && !status?.pid} onClick={() => void control("stop")}>停止</Button>
                <Button className="service-action" color="default" variant="filled" style={{borderColor: "transparent"}} htmlType="button" icon={<Icon type="ReloadOutlined"/>} loading={action === "restart"}
                    disabled={unavailable || stopping} onClick={() => void control("restart")}>重启</Button>
            </div>
        </div>
        <dl className={styles.metadata}>
            <div><dt>PID</dt><dd>{status?.pid || "—"}</dd></div>
            <div><dt>重试次数</dt><dd>{status ? <>{status.restart_count}<span className="retry-limit"> / {status.max_retries}</span></> : "—"}</dd></div>
        </dl>
        {status?.error && <div className={styles.error}><Icon type="InfoCircleOutlined"/><span>{status.error}</span></div>}
        {loadError && <div className={styles.error} role="alert"><Icon type="InfoCircleOutlined"/><span>{loadError}</span></div>}
    </div>;
};

ProcessControls.displayName = "ProcessControls";

export default memo(ProcessControls);
