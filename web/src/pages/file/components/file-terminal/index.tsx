import React, {memo, useCallback, useEffect, useRef, useState} from "react";
import {Grid, Spin} from "antd";
import {createStyles} from "antd-style";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {getServerInfo} from "@/service/api/server";
import ServerTerminal, {
    type ConnectionStatus,
    type TerminalRef,
} from "@/pages/server/components/terminal";
import useServerStyles from "@/pages/server/styles";
import {localServer} from "@/pages/server/hooks/servers";

interface FileTerminalProps {
    path: string;
    open: boolean;
    onClose: () => void;
}

const terminalBackground = "rgb(15, 15, 15)";

const useStyles = createStyles(({css, token}: any, props: {compact?: boolean; dark?: boolean} = {}) => {
    const compact = Boolean(props.compact);
    const dark = Boolean(props.dark);
    const headerPadding = compact ? "9px 14px" : "10px 16px";
    const desktopHeight = compact ? "min(560px, calc(100dvh - 140px))" : "min(640px, calc(100dvh - 170px))";
    const mobileHeight = compact ? "calc(100dvh - 88px)" : "calc(100dvh - 100px)";

    return {
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
        }

        .ant-modal-container {
            padding: 0;
            overflow: hidden;
            background: ${token.colorBgContainer};
        }

        .ant-modal-header {
            margin: 0;
            padding: ${headerPadding};
            border-bottom: 1px solid ${token.colorSplit};
            background: ${token.colorBgContainer};
        }

        .ant-modal-title {
            min-width: 0;
        }

    `,
    titleBar: css`
        width: 100%;
        min-width: 0;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
    `,
    closeButton: css`
        width: 30px;
        height: 30px;
        flex: none;
        padding: 0;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        border: 0;
        border-radius: ${token.borderRadiusSM}px;
        background: transparent;
        color: ${token.colorTextTertiary};
        cursor: pointer;

        &:hover,
        &:focus-visible {
            background: ${token.colorFillTertiary};
            color: ${token.colorText};
        }

        &:focus-visible {
            outline: 2px solid ${token.colorPrimaryBg};
            outline-offset: 1px;
        }
    `,
    host: css`
        height: ${desktopHeight};
        padding: 0;
        min-width: 0;
        min-height: 0;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        background: ${terminalBackground};
        box-sizing: border-box;

        > * {
            min-height: 0;
            flex: 1;
            overflow: hidden;
            border-radius: 0;
        }

        ${dark ? `
        .ant-btn-primary {
            border-color: #6f7b88;
            background: #6f7b88;
            box-shadow: none;
        }

        .ant-btn-primary:hover,
        .ant-btn-primary:focus-visible {
            border-color: #84919e;
            background: #84919e;
        }
        ` : ""}

        @media (max-width: 767px) {
            height: ${mobileHeight};
            padding: 0;
        }
    `,
    loading: css`
        width: 100%;
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: ${terminalBackground};
    `,
    };
});

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

const FileTerminal = ({path, open, onClose}: FileTerminalProps) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact, dark});
    const {styles: serverStyles} = useServerStyles({
        isCompactMode: compact,
        isDarkMode: dark,
        terminalBackground,
    });
    const terminalRef = useRef<TerminalRef | null>(null);
    const pathRef = useRef(path);
    const requestRef = useRef(0);
    const [server, setServer] = useState(localServer);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(false);
    const [status, setStatus] = useState<ConnectionStatus>("idle");
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
    }, [connect, error, loading, open, path, server]);

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

    return (
        <ProModal
            title={(
                <div className={styles.titleBar}>
                    <Title size="small">本机终端</Title>
                    <button
                        className={styles.closeButton}
                        type="button"
                        aria-label="关闭本机终端"
                        onClick={onClose}>
                        <Icon type="CloseOutlined"/>
                    </button>
                </div>
            )}
            width={compact ? "min(960px, calc(100vw - 24px))" : "min(1040px, calc(100vw - 24px))"}
            onCancel={onClose}
            modalProps={{
                open: open && Boolean(path),
                rootClassName: styles.modal,
                destroyOnHidden: false,
                style: {
                    top: screens.md ? 100 : 24,
                    paddingBottom: screens.md ? 100 : 24,
                },
                closable: false,
                footer: null,
                mask: {closable: true},
                styles: {
                    container: {padding: 0},
                    body: {padding: 0},
                },
            }}>
            <div className={styles.host}>
                {!path ? null : loading ? (
                    <div className={styles.loading}>
                        <Spin size="small"/>
                    </div>
                ) : (
                    <ServerTerminal
                        ref={terminalRef}
                        styles={serverStyles}
                        server={server}
                        active
                        initializing={false}
                        unavailable={error}
                        resetKey={0}
                        compact={compact}
                        showHeader={false}
                        terminalBackground={terminalBackground}
                        terminalAccent={dark ? "#c3ccd6" : undefined}
                        onStatusChange={handleStatusChange}/>
                )}
            </div>
        </ProModal>
    );
};

export default memo(FileTerminal);
