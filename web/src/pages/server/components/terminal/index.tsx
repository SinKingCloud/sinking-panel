import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Button, Spin, Tooltip} from "antd";
import {Icon} from "sinking-antd";
import TerminalView, {TerminalRef as TerminalViewRef} from "@/pages/components/terminal";
import defaultSettings from "@/../config/defaultSettings";
import {getHeaders} from "@/utils/auth";
import type {ServerRecord} from "../../hooks/servers";

export type ConnectionStatus = "idle" | "connecting" | "connected" | "disconnected" | "error";

const socketDecoder = new TextDecoder();

interface ConnectionFailure {
    code: string;
    message: string;
}

interface TerminalProps {
    styles: any;
    server: ServerRecord;
    active: boolean;
    initializing: boolean;
    unavailable: boolean;
    resetKey: number;
    compact: boolean;
    showHeader?: boolean;
    terminalBackground?: string;
    terminalAccent?: string;
    onStatusChange?: (serverId: number, status: ConnectionStatus) => void;
    onEdit?: (server: ServerRecord) => void;
}

export interface TerminalRef {
    clear: () => void;
    insertCommand: (command: string) => boolean;
    connect: () => void;
}

const statusText: Record<ConnectionStatus, string> = {
    idle: "未连接",
    connecting: "连接中",
    connected: "已连接",
    disconnected: "已断开",
    error: "连接失败",
};

const createSocketUrl = (server: ServerRecord, cols: number, rows: number) => {
    const gateway = new URL(defaultSettings?.gateway || window.location.origin, window.location.origin);
    gateway.protocol = gateway.protocol === "https:" || gateway.protocol === "wss:" ? "wss:" : "ws:";
    gateway.hash = "";
    gateway.search = "";
    const pathname = gateway.pathname.replace(/\/$/, "");
    gateway.pathname = `${pathname}/server/ssh`.replace(/\/{2,}/g, "/");
    gateway.searchParams.set("id", String(server.id));
    gateway.searchParams.set("width", String(cols));
    gateway.searchParams.set("height", String(rows));
    return gateway.toString();
};

const createSocketProtocol = (headers: Record<string, string>) => btoa(JSON.stringify(headers))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");

const Terminal = forwardRef<TerminalRef, TerminalProps>(({
                                                             styles,
                                                             server,
                                                             active,
                                                             initializing,
                                                             unavailable,
                                                             resetKey,
                                                             compact,
                                                             showHeader = true,
                                                             terminalBackground,
                                                             terminalAccent,
                                                             onStatusChange,
                                                             onEdit,
                                                         }, ref): any => {
    const {message} = App.useApp();
    const screenRef = useRef<HTMLDivElement | any>(null);
    const terminalRef = useRef<TerminalViewRef | any>(null);
    const socketRef = useRef<WebSocket | undefined>(undefined);
    const serverRef = useRef(server);
    const activeRef = useRef(active);
    const initializingRef = useRef(initializing);
    const unavailableRef = useRef(unavailable);
    const statusCallbackRef = useRef(onStatusChange);
    const generationRef = useRef(0);
    const failedRef = useRef(false);
    const failureRef = useRef<ConnectionFailure | null>(null);
    const mountedRef = useRef(false);
    const [socket, setSocket] = useState<WebSocket | undefined>(undefined);
    const [status, setStatus] = useState<ConnectionStatus>("idle");
    const [fullscreen, setFullscreen] = useState(false);
    const [failure, setFailure] = useState<ConnectionFailure | null>(null);
    serverRef.current = server;
    activeRef.current = active;
    initializingRef.current = initializing;
    unavailableRef.current = unavailable;
    statusCallbackRef.current = onStatusChange;

    const changeStatus = useCallback((value: ConnectionStatus) => {
        if (!mountedRef.current) {
            return;
        }
        setStatus(value);
        statusCallbackRef.current?.(serverRef.current.id, value);
    }, []);

    const changeFailure = useCallback((value: ConnectionFailure | null) => {
        failureRef.current = value;
        if (mountedRef.current) {
            setFailure(value);
        }
    }, []);

    const closeSocket = useCallback((nextStatus?: ConnectionStatus) => {
        generationRef.current += 1;
        const currentSocket = socketRef.current;
        socketRef.current = undefined;
        failedRef.current = false;
        if (mountedRef.current) {
            setSocket(undefined);
        }
        if (currentSocket) {
            currentSocket.onopen = null;
            currentSocket.onmessage = null;
            currentSocket.onerror = null;
            currentSocket.onclose = null;
            if (currentSocket.readyState === WebSocket.CONNECTING || currentSocket.readyState === WebSocket.OPEN) {
                currentSocket.close();
            }
        }
        if (nextStatus) {
            changeStatus(nextStatus);
        }
    }, [changeStatus]);

    const insertCommand = useCallback((command: string) => {
        return terminalRef.current?.paste(command) || false;
    }, []);

    const openEditForm = useCallback(() => {
        const open = () => {
            if (mountedRef.current) {
                onEdit?.(serverRef.current);
            }
        };
        if (document.fullscreenElement === screenRef.current && typeof document.exitFullscreen === "function") {
            void document.exitFullscreen().then(open).catch(open);
            return;
        }
        open();
    }, [onEdit]);

    const connect = useCallback(() => {
        const currentServer = serverRef.current;
        const terminal = terminalRef.current;
        if (initializingRef.current || unavailableRef.current || !currentServer || !terminal ||
            (currentServer.id === 0 && (!currentServer.ip || !currentServer.user))) {
            closeSocket("idle");
            return;
        }
        const currentSocket = socketRef.current;
        if (currentSocket?.readyState === WebSocket.CONNECTING || currentSocket?.readyState === WebSocket.OPEN) {
            return;
        }
        const headers = getHeaders();
        if (!headers.token) {
            closeSocket("error");
            changeFailure({code: "authorization_failed", message: "登录状态已失效，请重新登录"});
            return;
        }

        closeSocket();
        changeFailure(null);
        terminal.reset();
        const {cols, rows} = terminal.getSize();
        changeStatus("connecting");
        const generation = ++generationRef.current;
        let nextSocket: WebSocket | any;
        try {
            nextSocket = new WebSocket(
                createSocketUrl(currentServer, cols, rows),
                createSocketProtocol(headers),
            );
        } catch {
            changeFailure({code: "connection_failed", message: "无法创建终端连接，请稍后重试"});
            changeStatus("error");
            return;
        }
        nextSocket.binaryType = "arraybuffer";
        socketRef.current = nextSocket;
        failedRef.current = false;
        setSocket(nextSocket);

        const handleControl = (control: any) => {
            if (control?.event === "ready") {
                failedRef.current = false;
                changeFailure(null);
                changeStatus("connected");
                terminalRef.current?.fit(activeRef.current);
                return true;
            }
            if (control?.event === "error") {
                const nextFailure = {
                    code: String(control.code || "connection_failed"),
                    message: String(control.content || "SSH连接失败"),
                };
                failedRef.current = true;
                changeFailure(nextFailure);
                changeStatus("error");
                return true;
            }
            return control?.event === "pong";
        };

        const handleBinary = (value: ArrayBuffer) => {
            const payload = new Uint8Array(value);
            if (payload[0] === 123 && payload[payload.length - 1] === 125) {
                try {
                    if (handleControl(JSON.parse(socketDecoder.decode(payload)))) {
                        return;
                    }
                } catch {
                    // 普通终端输出继续交给 xterm 处理。
                }
            }
            terminalRef.current?.write(payload);
        };

        nextSocket.onmessage = (event) => {
            if (generationRef.current !== generation || socketRef.current !== nextSocket) {
                return;
            }
            if (typeof event.data === "string") {
                try {
                    if (handleControl(JSON.parse(event.data))) {
                        return;
                    }
                } catch {
                    // 兼容旧服务端发送的文本输出。
                }
                terminalRef.current?.write(event.data);
                return;
            }
            if (event.data instanceof ArrayBuffer) {
                handleBinary(event.data);
                return;
            }
            if (event.data instanceof Blob) {
                void event.data.arrayBuffer().then((value) => {
                    if (generationRef.current === generation && socketRef.current === nextSocket) {
                        handleBinary(value);
                    }
                });
            }
        };

        nextSocket.onopen = () => {
            if (generationRef.current !== generation || socketRef.current !== nextSocket) {
                nextSocket.close();
                return;
            }
            terminalRef.current?.fit(activeRef.current);
        };
        nextSocket.onerror = () => {
            if (generationRef.current === generation && socketRef.current === nextSocket) {
                failedRef.current = true;
                if (!failureRef.current) {
                    changeFailure({code: "connection_failed", message: "终端连接失败，请检查面板网络"});
                }
                changeStatus("error");
            }
        };
        nextSocket.onclose = () => {
            if (generationRef.current !== generation || socketRef.current !== nextSocket) {
                return;
            }
            socketRef.current = undefined;
            if (mountedRef.current) {
                setSocket(undefined);
            }
            changeStatus(failedRef.current || failureRef.current ? "error" : "disconnected");
        };
    }, [changeFailure, changeStatus, closeSocket]);

    useImperativeHandle(ref, () => ({
        clear: () => {
            terminalRef.current?.reset();
            terminalRef.current?.clear();
        },
        insertCommand,
        connect,
    }), [connect, insertCommand]);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            closeSocket();
        };
    }, [closeSocket]);

    useEffect(() => {
        closeSocket();
        changeFailure(null);
        terminalRef.current?.reset();
        changeStatus("idle");
    }, [
        changeFailure,
        changeStatus,
        closeSocket,
        resetKey,
        server.auth_type,
        server.id,
        server.ip,
        server.port,
        server.user,
    ]);

    useEffect(() => {
        const handleFullscreen = () => {
            const active = document.fullscreenElement === screenRef.current;
            setFullscreen(active);
            terminalRef.current?.fit(active);
        };
        document.addEventListener("fullscreenchange", handleFullscreen);
        return () => document.removeEventListener("fullscreenchange", handleFullscreen);
    }, []);

    const toggleFullscreen = useCallback(async () => {
        if (!screenRef.current) {
            return;
        }
        if (
            !document.fullscreenEnabled
            || typeof screenRef.current.requestFullscreen !== "function"
            || typeof document.exitFullscreen !== "function"
        ) {
            message.info("当前浏览器不支持全屏");
            return;
        }
        try {
            if (document.fullscreenElement === screenRef.current) {
                await document.exitFullscreen();
            } else {
                await screenRef.current.requestFullscreen();
            }
        } catch {
            message.error(fullscreen ? "退出全屏失败" : "进入全屏失败");
        }
    }, [fullscreen, message]);

    const canDisconnect = status === "connected" || status === "connecting";
    const localUnavailable = server.id === 0 && unavailable && !canDisconnect;
    const endpoint = initializing
        ? "正在加载本机连接..."
        : localUnavailable
            ? "本机连接加载失败"
            : server.ip && server.user
                ? `${server.user}@${server.ip}:${server.port}`
                : server.id === 0 ? "尚未配置本机 SSH" : `${server.ip}:${server.port}`;
    const needsConfiguration = !localUnavailable && server.id === 0 &&
        (!server.ip || !server.user);
    const needsCredentialUpdate = failure?.code === "credential_invalid" ||
        failure?.code === "password_invalid" ||
        failure?.code === "private_key_invalid";

    return (
        <section
            className={`${styles.consolePane} ${active ? "" : "session-hidden"}`}
            aria-hidden={!active}>
            {showHeader && <header className={styles.consoleHeader}>
                <div className="console-title">
                    <span className={`status-dot ${status}`}/>
                    <div>
                        <div className="console-name">{server.name || "终端"}</div>
                        <div className="console-endpoint">{endpoint}</div>
                    </div>
                </div>
                <div className="console-actions">
                    <span className={`status-text ${status}`}>{statusText[status]}</span>
                    <Tooltip title="清空终端">
                        <Button
                            type="text"
                            aria-label="清空终端"
                            icon={<Icon type="ClearOutlined"/>}
                            onClick={() => terminalRef.current?.clear()}/>
                    </Tooltip>
                    <Tooltip
                        title={canDisconnect ? "断开连接" : initializing ? "正在加载本机连接" : localUnavailable ? "本机连接加载失败" : needsConfiguration ? "请先配置本机连接" : "连接终端"}>
                        <Button
                            type="text"
                            disabled={!canDisconnect && (initializing || localUnavailable || needsConfiguration)}
                            aria-label={canDisconnect ? "断开连接" : initializing ? "正在加载本机连接" : localUnavailable ? "本机连接加载失败" : needsConfiguration ? "请先配置本机连接" : "连接终端"}
                            icon={<Icon type={canDisconnect ? "DisconnectOutlined" : "LinkOutlined"}/>}
                            onClick={() => canDisconnect ? closeSocket("disconnected") : connect()}/>
                    </Tooltip>
                    <Tooltip title={fullscreen ? "退出全屏" : "全屏"}>
                        <Button
                            type="text"
                            aria-label={fullscreen ? "退出全屏" : "全屏"}
                            icon={<Icon type={fullscreen ? "FullscreenExitOutlined" : "FullscreenOutlined"}/>}
                            onClick={() => void toggleFullscreen()}/>
                    </Tooltip>
                </div>
            </header>}
            <div className={styles.terminalBody}>
                <div className={styles.terminalScreen} ref={screenRef}>
                    <TerminalView
                        ref={terminalRef}
                        socket={socket}
                        active={active}
                        compact={compact}
                        background={terminalBackground}
                        accentColor={terminalAccent}/>
                    {(initializing || localUnavailable || status !== "connected") && (
                        <div className={`${styles.terminalOverlay} ${status}`}>
                            {initializing ? (
                                <Button type="text" disabled>
                                    正在加载本机连接
                                </Button>
                            ) : localUnavailable ? (
                                <Button type="text" disabled icon={<Icon type="WarningOutlined"/>}>
                                    本机连接加载失败
                                </Button>
                            ) : status === "connecting" ? (
                                <Spin size="small" description="正在连接..."/>
                            ) : needsConfiguration ? (
                                <Button type="text" disabled icon={<Icon type="SettingOutlined"/>}>
                                    请先配置本机连接
                                </Button>
                            ) : status === "error" ? (
                                <div className="terminal-error">
                                    <div className="terminal-error-title">
                                        <Icon type="WarningOutlined"/>
                                        <span>连接失败</span>
                                    </div>
                                    <div className="terminal-error-message">
                                        {failure?.message || "终端连接失败，请稍后重试"}
                                    </div>
                                    <div className="terminal-error-actions">
                                        {needsCredentialUpdate && (
                                            <Button type="primary" onClick={openEditForm}>
                                                修改信息
                                            </Button>
                                        )}
                                        <Button onClick={connect}>重新连接</Button>
                                    </div>
                                </div>
                            ) : (
                                <Button
                                    type="primary"
                                    icon={<Icon type="LinkOutlined"/>}
                                    onClick={connect}>
                                    连接终端
                                </Button>
                            )}
                        </div>
                    )}
                    {fullscreen && (
                        <div className="fullscreen-actions">
                            <Button
                                type="text"
                                title="清空终端"
                                aria-label="清空终端"
                                icon={<Icon type="ClearOutlined"/>}
                                onClick={() => terminalRef.current?.clear()}/>
                            <Button
                                type="text"
                                disabled={!canDisconnect && (initializing || localUnavailable || needsConfiguration)}
                                title={canDisconnect ? "断开连接" : initializing ? "正在加载本机连接" : localUnavailable ? "本机连接加载失败" : needsConfiguration ? "请先配置本机连接" : "连接终端"}
                                aria-label={canDisconnect ? "断开连接" : initializing ? "正在加载本机连接" : localUnavailable ? "本机连接加载失败" : needsConfiguration ? "请先配置本机连接" : "连接终端"}
                                icon={<Icon type={canDisconnect ? "DisconnectOutlined" : "LinkOutlined"}/>}
                                onClick={() => canDisconnect ? closeSocket("disconnected") : connect()}/>
                            <Button
                                type="text"
                                title="退出全屏"
                                aria-label="退出全屏"
                                icon={<Icon type="FullscreenExitOutlined"/>}
                                onClick={() => void toggleFullscreen()}/>
                        </div>
                    )}
                </div>
            </div>
        </section>
    );
});

Terminal.displayName = "Terminal";

export default memo(Terminal);
