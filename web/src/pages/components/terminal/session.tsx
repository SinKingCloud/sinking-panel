import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {createPortal} from "react-dom";
import {App, Button, ConfigProvider, Spin, Tooltip, theme as antdTheme} from "antd";
import {Icon} from "sinking-antd";
import TerminalView, {type TerminalRef as TerminalViewRef} from "./index";
import useStyles from "./styles";
import defaultSettings from "@/../config/defaultSettings";
import {getHeaders} from "@/utils/auth";

export type ConnectionStatus = "idle" | "connecting" | "connected" | "disconnected" | "error";


export interface TerminalProps {
    path: string;
    params: Record<string, string | number>;
    binaryControl?: boolean;
    acknowledge?: boolean;
    title?: string;
    endpoint?: string;
    active?: boolean;
    loading?: boolean;
    disabledReason?: string;
    disabledIcon?: string;
    resetKey?: string | number;
    compact?: boolean;
    styles?: any;
    showHeader?: boolean;
    toolbarContainer?: HTMLElement | null;
    showStatus?: boolean;
    showClear?: boolean;
    showReconnect?: boolean;
    showConnect?: boolean; // 根据连接状态显示连接或断开。
    showFullscreen?: boolean;
    showClose?: boolean;
    autoConnect?: boolean;
    terminalBackground?: string;
    terminalAccent?: string;
    errorAction?: {label: string; codes: string[]; onClick: () => void};
    onStatusChange?: (status: ConnectionStatus) => void;
    onClose?: () => void;
}

export interface TerminalRef {
    clear: () => void;
    insertCommand: (command: string) => boolean;
    connect: () => void;
    reconnect: () => void;
}

const statusText: Record<ConnectionStatus, string> = {
    idle: "未连接", connecting: "连接中", connected: "已连接", disconnected: "已断开", error: "连接失败",
};
const decoder = new TextDecoder();

// SSH 和容器共用连接生命周期、终端操作及状态界面，协议差异由调用方配置。
const TerminalSession = forwardRef<TerminalRef, TerminalProps>(({
    path, params, binaryControl = false, acknowledge = false, title = "终端", endpoint,
    active = true, loading = false, disabledReason, disabledIcon = "SettingOutlined", resetKey = 0, compact = false,
    styles: customStyles, showHeader = true, toolbarContainer, showStatus = true, showClear = true,
    showReconnect = false, showConnect = true, showFullscreen = true, autoConnect = false,
    terminalBackground, terminalAccent, errorAction, onStatusChange, onClose, showClose = Boolean(onClose),
}, ref) => {
    const {message, modal} = App.useApp();
    const {styles: defaultStyles} = useStyles({compact, terminalBackground});
    const styles = customStyles || defaultStyles;
    const terminalTheme = useMemo(() => ({
        algorithm: compact ? [antdTheme.darkAlgorithm, antdTheme.compactAlgorithm] : antdTheme.darkAlgorithm,
    }), [compact]);
    const screenRef = useRef<HTMLDivElement>(null);
    const terminalRef = useRef<TerminalViewRef>(null);
    const socketRef = useRef<WebSocket | undefined>(undefined);
    const generationRef = useRef(0);
    const mountedRef = useRef(false);
    const failureRef = useRef<any>(null);
    const timersRef = useRef<{timeout?: number; heartbeat?: number}>({});
    const [attempt, setAttempt] = useState<{key: string; number: number}>();
    const [socket, setSocket] = useState<WebSocket>();
    const [status, setStatus] = useState<ConnectionStatus>("idle");
    const [failure, setFailure] = useState<any>(null);
    const [fullscreen, setFullscreen] = useState(false);
    const [fullscreenPending, setFullscreenPending] = useState(false);
    const connectionKey = JSON.stringify([path, params, binaryControl, acknowledge, resetKey]);
    const optionsRef = useRef({path, params, binaryControl, acknowledge, connectionKey, active, loading, disabledReason, onStatusChange});
    optionsRef.current = {path, params, binaryControl, acknowledge, connectionKey, active, loading, disabledReason, onStatusChange};

    const changeStatus = useCallback((value: ConnectionStatus) => {
        if (!mountedRef.current) return;
        setStatus(value);
        optionsRef.current.onStatusChange?.(value);
    }, []);

    const changeFailure = useCallback((value: any) => {
        failureRef.current = value;
        if (mountedRef.current) setFailure(value);
    }, []);

    const closeSocket = useCallback((nextStatus?: ConnectionStatus) => {
        generationRef.current += 1;
        window.clearTimeout(timersRef.current.timeout);
        window.clearInterval(timersRef.current.heartbeat);
        timersRef.current = {};
        const connection = socketRef.current;
        socketRef.current = undefined;
        if (mountedRef.current) setSocket(undefined);
        if (connection) {
            connection.onopen = null;
            connection.onmessage = null;
            connection.onerror = null;
            connection.onclose = null;
            if (connection.readyState === WebSocket.CONNECTING || connection.readyState === WebSocket.OPEN) connection.close();
        }
        if (nextStatus) changeStatus(nextStatus);
    }, [changeStatus]);

    const connect = useCallback((force = false) => {
        const options = optionsRef.current;
        if (options.loading || options.disabledReason || !terminalRef.current) return;
        const connection = socketRef.current;
        if (!force && (connection?.readyState === WebSocket.CONNECTING || connection?.readyState === WebSocket.OPEN)) return;
        closeSocket();
        changeFailure(null);
        changeStatus("connecting");
        setAttempt({key: options.connectionKey, number: generationRef.current});
    }, [changeFailure, changeStatus, closeSocket]);

    useImperativeHandle(ref, () => ({
        clear: () => { terminalRef.current?.reset(); terminalRef.current?.clear(); },
        insertCommand: (command) => terminalRef.current?.paste(command) || false,
        connect: () => connect(),
        reconnect: () => connect(true),
    }), [connect]);

    useEffect(() => {
        mountedRef.current = true;
        return () => { mountedRef.current = false; closeSocket(); };
    }, [closeSocket]);

    useEffect(() => {
        closeSocket();
        changeFailure(null);
        changeStatus("idle");
    }, [connectionKey, changeFailure, changeStatus, closeSocket]);

    useEffect(() => {
        if (!attempt || attempt.key !== connectionKey) return;
        const terminal = terminalRef.current;
        if (!terminal) return;
        const options = optionsRef.current;
        const headers = getHeaders();
        if (!headers.token) {
            changeFailure({code: "authorization_failed", message: "登录状态已失效，请重新登录"});
            changeStatus("error");
            return;
        }
        const generation = ++generationRef.current;
        let connection: WebSocket;
        try {
            const gateway = new URL(defaultSettings?.gateway || window.location.origin, window.location.origin);
            gateway.protocol = gateway.protocol === "https:" || gateway.protocol === "wss:" ? "wss:" : "ws:";
            gateway.hash = "";
            gateway.search = "";
            gateway.pathname = `${gateway.pathname.replace(/\/$/, "")}/${options.path.replace(/^\//, "")}`.replace(/\/{2,}/g, "/");
            Object.entries(options.params).forEach(([key, value]) => gateway.searchParams.set(key, String(value)));
            const {cols, rows} = terminal.getSize();
            gateway.searchParams.set("width", String(cols));
            gateway.searchParams.set("height", String(rows));
            const protocol = btoa(JSON.stringify(headers)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
            connection = new WebSocket(gateway.toString(), protocol);
        } catch {
            changeFailure({code: "connection_failed", message: "无法创建终端连接，请稍后重试"});
            changeStatus("error");
            return;
        }
        socketRef.current = connection;
        connection.binaryType = "arraybuffer";
        let ready = false;
        const current = () => mountedRef.current && optionsRef.current.connectionKey === connectionKey
            && generationRef.current === generation && socketRef.current === connection;
        const fail = (code: string, text: string) => {
            if (!current()) return;
            changeFailure({code, message: text});
            closeSocket("error");
        };
        timersRef.current.timeout = window.setTimeout(() => fail("connection_timeout", "终端连接超时，请重试"), 30000);
        const handleControl = (control: any) => {
            if (control?.event === "ready") {
                ready = true;
                window.clearTimeout(timersRef.current.timeout);
                changeFailure(null);
                setSocket(connection);
                changeStatus("connected");
                terminalRef.current?.fit(optionsRef.current.active);
                return true;
            }
            if (control?.event === "error") {
                fail(String(control.code || "connection_failed"), String(control.content || "终端连接失败"));
                return true;
            }
            return control?.event === "pong";
        };
        const handleBinary = (value: ArrayBuffer) => {
            if (!current()) return;
            const payload = new Uint8Array(value);
            if (options.binaryControl && payload[0] === 123 && payload[payload.length - 1] === 125) {
                try { if (handleControl(JSON.parse(decoder.decode(payload)))) return; } catch { /* 普通输出交给 xterm。 */ }
            }
            terminalRef.current?.write(payload, options.acknowledge ? () => {
                if (current() && connection.readyState === WebSocket.OPEN) {
                    connection.send(JSON.stringify({event: "ack"}));
                }
            } : undefined);
        };
        connection.onmessage = ({data}) => {
            if (!current()) return;
            if (typeof data === "string") {
                try { if (handleControl(JSON.parse(data))) return; } catch { /* 兼容 SSH 的文本输出。 */ }
                if (options.binaryControl) terminalRef.current?.write(data);
            } else if (data instanceof ArrayBuffer) {
                handleBinary(data);
            } else if (data instanceof Blob) {
                void data.arrayBuffer().then(handleBinary).catch(() => fail("output_failed", "读取终端输出失败"));
            }
        };
        connection.onopen = () => {
            if (!current()) return;
            timersRef.current.heartbeat = window.setInterval(() => {
                if (current() && connection.readyState === WebSocket.OPEN) connection.send(JSON.stringify({event: "ping"}));
            }, 20000);
        };
        connection.onerror = () => fail("connection_failed", "终端连接失败，请检查面板网络");
        connection.onclose = () => {
            if (!current()) return;
            if (!ready) changeFailure(failureRef.current || {code: "connection_closed", message: "终端连接已关闭，请确认服务正在运行"});
            closeSocket(failureRef.current ? "error" : "disconnected");
        };
        return () => closeSocket();
    }, [attempt, connectionKey, changeFailure, changeStatus, closeSocket]);

    useEffect(() => {
        if (autoConnect && active && !loading && !disabledReason) connect();
    }, [active, autoConnect, connectionKey, connect, disabledReason, loading]);

    useEffect(() => {
        const screen = screenRef.current;
        const change = () => {
            const selected = document.fullscreenElement === screen;
            setFullscreen(selected);
            terminalRef.current?.fit(selected);
        };
        document.addEventListener("fullscreenchange", change);
        return () => {
            document.removeEventListener("fullscreenchange", change);
            if (screen && document.fullscreenElement === screen) void document.exitFullscreen().catch(() => undefined);
        };
    }, []);

    const toggleFullscreen = useCallback(async () => {
        const screen = screenRef.current;
        if (!screen || fullscreenPending) return;
        if (!document.fullscreenEnabled || !screen.requestFullscreen || !document.exitFullscreen) {
            message.info("当前浏览器不支持全屏");
            return;
        }
        setFullscreenPending(true);
        try {
            if (document.fullscreenElement === screen) await document.exitFullscreen();
            else await screen.requestFullscreen();
            if (!screen.isConnected && document.fullscreenElement === screen) await document.exitFullscreen();
        } catch {
            if (screen.isConnected) message.error("切换全屏失败");
        } finally {
            if (screen.isConnected) setFullscreenPending(false);
        }
    }, [fullscreenPending, message]);

    const close = async () => {
        // 文件终端关闭后保留会话，需要先退出全屏再隐藏弹窗。
        if (document.fullscreenElement === screenRef.current && document.exitFullscreen) {
            try {
                await document.exitFullscreen();
            } catch {
                message.error("退出全屏失败，请重试");
                return;
            }
        }
        onClose?.();
    };

    const canDisconnect = status === "connected" || status === "connecting";
    const blockedReason = canDisconnect ? "" : disabledReason;
    const connectionLabel = canDisconnect ? "断开连接" : disabledReason || "连接终端";
    const actions = [
        {key: "clear", visible: showClear, title: "清理", icon: "ClearOutlined", onClick: () => terminalRef.current?.clear()},
        {key: "reconnect", visible: showReconnect, title: "重连", icon: "ReloadOutlined", disabled: loading || Boolean(disabledReason), onClick: () => connect(true)},
        {key: "connect", visible: showConnect, title: connectionLabel, icon: canDisconnect ? "DisconnectOutlined" : "LinkOutlined", disabled: !canDisconnect && (loading || Boolean(disabledReason)), onClick: () => {
            if (!canDisconnect) {
                connect();
                return;
            }
            const generation = generationRef.current;
            modal.confirm({
                title: "断开连接",
                content: `确定断开终端“${title}”的连接吗？`,
                okText: "断开",
                cancelText: "取消",
                okButtonProps: {danger: true, type: "default"},
                mask: {closable: true},
                getContainer: document.fullscreenElement === screenRef.current ? screenRef.current || document.body : document.body,
                onOk: () => {
                    if (mountedRef.current && generationRef.current === generation) closeSocket("disconnected");
                },
            });
        }},
        {key: "fullscreen", visible: showFullscreen, title: fullscreen ? "退出全屏" : "全屏", icon: fullscreen ? "FullscreenExitOutlined" : "FullscreenOutlined", disabled: fullscreenPending, onClick: () => void toggleFullscreen()},
        {key: "close", visible: showClose && Boolean(onClose), title: "关闭", icon: "CloseOutlined", onClick: () => void close()},
    ].filter((action) => action.visible);
    const toolbar = <>
        {showStatus && <span className={`status-text ${status}`}>{statusText[status]}</span>}
        {actions.map((action) => <Tooltip key={action.key} title={action.title}>
            <Button type="text" aria-label={action.title} disabled={action.disabled} icon={<Icon type={action.icon}/>} onClick={action.onClick}/>
        </Tooltip>)}
    </>;

    return <section className={`${styles.consolePane} ${active ? "" : "session-hidden"}`} aria-hidden={!active}>
        {showHeader && <header className={styles.consoleHeader}>
            <div className="console-title">
                {showStatus && <span className={`status-dot ${status}`}/>}
                <div>
                    <div className="console-name">{title}</div>
                    {endpoint && <div className="console-endpoint">{endpoint}</div>}
                </div>
            </div>
            <div className="console-actions">
                {toolbar}
            </div>
        </header>}
        {!showHeader && toolbarContainer && createPortal(toolbar, toolbarContainer)}
        <ConfigProvider theme={terminalTheme}>
            <div className={styles.terminalBody}>
                <div className={styles.terminalScreen} ref={screenRef}>
                    <TerminalView key={`${connectionKey}:${attempt?.number || 0}`} ref={terminalRef} socket={socket} active={active} compact={compact} background={terminalBackground} accentColor={terminalAccent}/>
                    {(loading || blockedReason || status !== "connected") && <div className={`${styles.terminalOverlay} ${status}`}>
                        {loading ? <Button type="text" disabled>{disabledReason || "正在加载连接"}</Button>
                            : blockedReason ? <Button type="text" disabled icon={<Icon type={disabledIcon}/>}>{blockedReason}</Button>
                            : status === "connecting" ? <Spin size="small" description="正在连接..."/>
                            : status === "error" ? <div className="terminal-error">
                                <div className="terminal-error-title"><Icon type="WarningOutlined"/><span>连接失败</span></div>
                                <div className="terminal-error-message">{failure?.message || "终端连接失败，请稍后重试"}</div>
                                <div className="terminal-error-actions">
                                    {failure && errorAction?.codes.includes(failure.code) && <Button color="primary" variant="solid" onClick={() => {
                                        const open = () => { if (mountedRef.current) errorAction.onClick(); };
                                        if (document.fullscreenElement === screenRef.current && document.exitFullscreen) void document.exitFullscreen().then(open).catch(open);
                                        else open();
                                    }}>{errorAction.label}</Button>}
                                    <Button color="default" variant="filled" style={{borderColor: "transparent"}} onClick={() => connect()}>重新连接</Button>
                                </div>
                            </div> : <Button color="primary" variant="solid" icon={<Icon type="LinkOutlined"/>} onClick={() => connect()}>连接终端</Button>}
                    </div>}
                    {fullscreen && actions.length > 0 && <div className="fullscreen-actions">
                        {actions.map((action) => <Button key={action.key} type="text" title={action.title} aria-label={action.title} disabled={action.disabled} icon={<Icon type={action.icon}/>} onClick={action.onClick}/>)}
                    </div>}
                </div>
            </div>
        </ConfigProvider>
    </section>;
});

TerminalSession.displayName = "TerminalSession";
export default memo(TerminalSession);
