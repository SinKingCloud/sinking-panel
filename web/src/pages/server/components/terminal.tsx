import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {Button, Spin, Tooltip} from "antd";
import {Icon} from "sinking-antd";
import TerminalView, {TerminalRef as TerminalViewRef} from "@/pages/components/terminal";
import defaultSettings from "@/../config/defaultSettings";
import {getHeaders} from "@/utils/auth";
import type {ServerRecord} from "../hooks/servers";

export type ConnectionStatus = "idle" | "connecting" | "connected" | "disconnected" | "error";

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
                                                         }, ref): any => {
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
    const mountedRef = useRef(false);
    const [socket, setSocket] = useState<WebSocket | undefined>(undefined);
    const [status, setStatus] = useState<ConnectionStatus>("idle");
    const [fullscreen, setFullscreen] = useState(false);
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
            return;
        }

        closeSocket();
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
            changeStatus("error");
            return;
        }
        nextSocket.binaryType = "arraybuffer";
        socketRef.current = nextSocket;
        failedRef.current = false;
        setSocket(nextSocket);

        nextSocket.onmessage = (event) => {
            if (generationRef.current !== generation || socketRef.current !== nextSocket) {
                return;
            }
            if (typeof event.data === "string") {
                terminalRef.current?.write(event.data);
                return;
            }
            if (event.data instanceof ArrayBuffer) {
                terminalRef.current?.write(new Uint8Array(event.data));
                return;
            }
            if (event.data instanceof Blob) {
                void event.data.arrayBuffer().then((value) => {
                    if (generationRef.current === generation && socketRef.current === nextSocket) {
                        terminalRef.current?.write(new Uint8Array(value));
                    }
                });
            }
        };

        nextSocket.onopen = () => {
            if (generationRef.current !== generation || socketRef.current !== nextSocket) {
                nextSocket.close();
                return;
            }
            changeStatus("connected");
            terminalRef.current?.fit(activeRef.current);
        };
        nextSocket.onerror = () => {
            if (generationRef.current === generation && socketRef.current === nextSocket) {
                failedRef.current = true;
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
            changeStatus(failedRef.current ? "error" : "disconnected");
        };
    }, [changeStatus, closeSocket]);

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
        terminalRef.current?.reset();
        changeStatus("idle");
    }, [
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
        try {
            if (document.fullscreenElement === screenRef.current) {
                await document.exitFullscreen();
            } else {
                await screenRef.current.requestFullscreen();
            }
        } catch {
            // Fullscreen can be blocked by browser or embedding policy.
        }
    }, []);

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
                            ) : (
                                <Button
                                    type="primary"
                                    icon={<Icon type={status === "error" ? "ReloadOutlined" : "LinkOutlined"}/>}
                                    onClick={connect}>
                                    {status === "error" ? "重新连接" : "连接终端"}
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
