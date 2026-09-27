import React, {memo, useCallback, useEffect, useRef, useState} from "react";
import {useTheme} from "sinking-antd";
import {getServerInfo} from "@/service/api/server";
import TerminalModal, {terminalBackground} from "@/pages/components/terminal/modal";
import ServerTerminal, {type ConnectionStatus, type TerminalRef} from "@/pages/server/components/terminal";
import ServerForm from "@/pages/server/components/form";
import useServerStyles from "@/pages/server/styles";
import { localServer } from "@/pages/server/hooks/servers";

const authTypeData: Record<string, string> = {};

const shellQuote = (value: string) => `'${value.replace(/'/g, "'\\''")}'`;

const normalizeLocalServer = (data: any) => ({
    ...localServer,
    ip: String(data?.ip || ""),
    port: Number(data?.port) || 22,
    user: String(data?.user || ""),
    auth_type: Number(data?.auth_type) || 0,
    name: String(data?.name || "本机终端"),
    searchText: "",
});

const FileTerminal = ({path, open, onClose}: any) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles: serverStyles} = useServerStyles({
        isCompactMode: compact,
        isDarkMode: dark,
        terminalBackground,
    });
    const terminalRef = useRef<TerminalRef | null>(null);
    const formRef = useRef<any>({} as any);
    const pathRef = useRef(path);
    const requestRef = useRef(0);
    const [server, setServer] = useState(localServer);
    const [loading, setLoading] = useState(false);
    const [toolbarContainer, setToolbarContainer] = useState<HTMLDivElement | null>(null);
    const [error, setError] = useState(false);
    const [status, setStatus] = useState<ConnectionStatus>("idle");
    const [resetKey, setResetKey] = useState(0);
    const serverLoadedRef = useRef(false);
    const enteredPathRef = useRef("");

    pathRef.current = path;

    useEffect(() => {
        if (!path) {
            return;
        }

        // Server information belongs to the terminal session, not to the
        // working directory. Fetch it only when the session is first opened.
        if (serverLoadedRef.current) {
            return;
        }
        const requestId = ++requestRef.current;
        setLoading(true);
        setError(false);
        setStatus("idle");
        void getServerInfo({body: {id: 0}}).then((response) => {
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code === 200 && response.data) {
                setServer(normalizeLocalServer(response.data));
                serverLoadedRef.current = true;
                setLoading(false);
                return;
            }
            setError(true);
        }).catch(() => {
            if (requestRef.current === requestId) {
                setError(true);
            }
        }).finally(() => {
            if (requestRef.current === requestId) {
                setLoading(false);
            }
        });

        return () => {
            if (requestRef.current === requestId) {
                requestRef.current += 1;
            }
        };
    }, [path]);

    const connect = useCallback(() => {
        window.setTimeout(() => terminalRef.current?.connect(), 0);
    }, []);

    useEffect(() => {
        if (!open || !path || loading || error) {
            return;
        }
        const frame = window.requestAnimationFrame(connect);
        return () => window.cancelAnimationFrame(frame);
    }, [connect, error, loading, open, path, resetKey, server]);

    const clearAndEnterDirectory = useCallback(() => {
        const currentPath = pathRef.current;
        if (!currentPath || status !== "connected" || enteredPathRef.current) {
            return;
        }
        terminalRef.current?.clear();
        const sent = terminalRef.current?.insertCommand(
            `cd ${shellQuote(currentPath)}; printf '\\033[2J\\033[3J\\033[H'\n`,
        );
        if (sent) {
            enteredPathRef.current = currentPath;
        }
    }, [status]);

    useEffect(() => {
        if (!open || status !== "connected" || !path) {
            return;
        }
        const timer = window.setTimeout(clearAndEnterDirectory, 0);
        return () => window.clearTimeout(timer);
    }, [clearAndEnterDirectory, open, status]);

    const handleStatusChange = useCallback((_serverId: number, nextStatus: ConnectionStatus) => {
        if (nextStatus !== "connected") {
            // A later reconnect must enter the current directory again.
            enteredPathRef.current = "";
        }
        setStatus(nextStatus);
    }, []);

    const openEdit = useCallback((record: any) => {
        formRef.current?.open(record);
    }, []);

    const handleFormSuccess = useCallback(({server: nextServer, reconnectRequired}: any) => {
        if (nextServer) {
            setServer(nextServer);
        }
        if (reconnectRequired) {
            enteredPathRef.current = "";
            setResetKey((value) => value + 1);
        }
    }, []);

    return (
        <>
            <TerminalModal
                title="本机终端"
                open={open && Boolean(path)}
                onClose={onClose}
                closable={loading || !toolbarContainer}
                toolbarRef={setToolbarContainer}
                loading={Boolean(path) && loading}>
                {path && <ServerTerminal
                    ref={terminalRef}
                    styles={serverStyles}
                    server={server}
                    active={open}
                    initializing={false}
                    unavailable={error}
                    resetKey={resetKey}
                    compact={compact}
                    showHeader={false}
                    toolbarContainer={toolbarContainer}
                    showStatus={false}
                    showConnect={false}
                    showClear
                    showReconnect
                    showFullscreen
                    showClose
                    onClose={onClose}
                    terminalBackground={terminalBackground}
                    terminalAccent={dark ? "#c3ccd6" : undefined}
                    onEdit={openEdit}
                    onStatusChange={handleStatusChange}/>}
            </TerminalModal>
            <ServerForm
                ref={formRef}
                authTypeData={authTypeData}
                layered
                onSuccess={handleFormSuccess}/>
        </>
    );
};

export default memo(FileTerminal);
