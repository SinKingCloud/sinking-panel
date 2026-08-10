import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, Card} from "antd";
import {Body, useTheme} from "sinking-antd";
import useEnum from "@/utils/enum";
import {deleteServer} from "@/service/api/server";
import Scripts from "./components/scripts";
import Form from "./components/form";
import type {FormRef, FormSuccessResult} from "./components/form";
import Sidebar from "./components/sidebar";
import Terminal from "./components/terminal";
import type {ConnectionStatus, TerminalRef} from "./components/terminal";
import useHeight from "./hooks/height";
import usePanels from "./hooks/panels";
import useServers, {localServer} from "./hooks/servers";
import type {ServerRecord} from "./hooks/servers";
import useStyles from "./styles";

const fallbackAuthTypeData = {
    "0": "密码验证",
    "1": "证书验证",
};

interface TerminalSession {
    server: ServerRecord;
    resetKey: number;
}

type TerminalSessions = Record<string, TerminalSession>;
type ConnectionStatuses = Record<string, ConnectionStatus>;

export default (): React.ReactNode => {
    const theme = useTheme();
    const isCompactMode = theme?.isCompactTheme?.() || false;
    const isDarkMode = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const styleProps = useMemo(() => ({isCompactMode, isDarkMode}), [isCompactMode, isDarkMode]);
    const {styles} = useStyles(styleProps);
    const {message, modal} = App.useApp();
    const [enumData, , enumError] = useEnum("server");
    const list = useServers();
    const formRef = useRef<FormRef>({} as FormRef);
    const terminalRefs = useRef(new Map<number, TerminalRef>());
    const terminalRefCallbacks = useRef(new Map<number, (value: TerminalRef | null) => void>());
    const {pageRef, pageStyle} = useHeight(isCompactMode);
    const [selectedId, setSelectedId] = useState(localServer.id);
    const [sessions, setSessions] = useState<TerminalSessions>(() => ({
        [String(localServer.id)]: {server: localServer, resetKey: 0},
    }));
    const [connectionStatuses, setConnectionStatuses] = useState<ConnectionStatuses>(() => ({
        [String(localServer.id)]: "idle",
    }));
    const {
        serverCollapsed,
        scriptsCollapsed,
        setServerCollapsed,
        setScriptsCollapsed,
    } = usePanels();
    const authTypeData = useMemo(() => ({
        ...fallbackAuthTypeData,
        ...(enumData?.auth_type || {}),
    }), [enumData?.auth_type]);

    useEffect(() => {
        if (enumError) {
            message.warning("认证方式加载失败，已使用默认选项");
        }
    }, [enumError, message]);

    useEffect(() => {
        const records = new Map(list.servers.map((server) => [server.id, server]));
        if (!list.localRefreshing) {
            records.set(list.localServer.id, list.localServer);
        }
        setSessions((current) => {
            let changed = false;
            const next = {...current};
            Object.entries(current).forEach(([key, session]) => {
                const server = records.get(session.server.id);
                if (server && server !== session.server) {
                    next[key] = {...session, server};
                    changed = true;
                }
            });
            return changed ? next : current;
        });
    }, [list.localRefreshing, list.localServer, list.servers]);

    const changeSessionStatus = useCallback((serverId: number, status: ConnectionStatus) => {
        const key = String(serverId);
        setConnectionStatuses((current) => {
            if (current[key] === status) {
                return current;
            }
            const next = {...current, [key]: status};
            return next;
        });
    }, []);

    const getTerminalRef = useCallback((serverId: number) => {
        const current = terminalRefCallbacks.current.get(serverId);
        if (current) {
            return current;
        }
        const callback = (value: TerminalRef | null) => {
            if (value) {
                terminalRefs.current.set(serverId, value);
            } else {
                terminalRefs.current.delete(serverId);
            }
        };
        terminalRefCallbacks.current.set(serverId, callback);
        return callback;
    }, []);

    const openCreate = useCallback(() => formRef.current?.open(), []);
    const openEdit = useCallback((server: ServerRecord) => {
        if (server.id === 0 && list.localRefreshing) {
            message.info("本机连接信息正在加载");
            return;
        }
        formRef.current?.open(server);
    }, [list.localRefreshing, message]);
    const selectServer = useCallback((server: ServerRecord) => {
        const key = String(server.id);
        setSessions((current) => {
            const existing = current[key];
            if (existing) {
                return current;
            }
            return {
                ...current,
                [key]: {
                    server,
                    resetKey: 0,
                },
            };
        });
        setConnectionStatuses((current) => key in current ? current : {...current, [key]: "idle"});
        setSelectedId(server.id);
    }, []);
    const connectServer = useCallback((server: ServerRecord) => {
        selectServer(server);
        const status = connectionStatuses[String(server.id)] || "idle";
        if (status === "connected" || status === "connecting") {
            return;
        }
        window.requestAnimationFrame(() => terminalRefs.current.get(server.id)?.connect());
    }, [connectionStatuses, selectServer]);
    const handleFormSuccess = useCallback(({serverId, server, reconnectRequired}: FormSuccessResult) => {
        if (serverId !== undefined) {
            const key = String(serverId);
            setSessions((current) => {
                const session = current[key];
                if (!session) {
                    return current;
                }
                return {
                    ...current,
                    [key]: {
                        ...session,
                        server: server || session.server,
                        resetKey: session.resetKey + (reconnectRequired ? 1 : 0),
                    },
                };
            });
            if (reconnectRequired) {
                changeSessionStatus(serverId, "idle");
            }
        }
        list.reload();
    }, [changeSessionStatus, list.reload]);
    const insertCommand = useCallback((command: string) => {
        if (!terminalRefs.current.get(selectedId)?.insertCommand(command)) {
            message.warning("请先连接终端");
        }
    }, [message, selectedId]);

    const remove = useCallback(async (server: ServerRecord) => {
        const key = `server-delete-${server.id}`;
        message.loading({key, content: "正在删除终端...", duration: 0});
        try {
            const response = await deleteServer({body: {ids: [server.id]}});
            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || "删除终端失败");
                return;
            }
            const serverKey = String(server.id);
            setSessions((current) => {
                if (!current[serverKey]) {
                    return current;
                }
                const next = {...current};
                delete next[serverKey];
                return next;
            });
            setConnectionStatuses((current) => {
                if (!(serverKey in current)) {
                    return current;
                }
                const next = {...current};
                delete next[serverKey];
                return next;
            });
            terminalRefs.current.delete(server.id);
            terminalRefCallbacks.current.delete(server.id);
            if (selectedId === server.id) {
                setSelectedId(list.localServer.id);
            }
            message.success(response.message || "删除成功");
            list.reload();
        } finally {
            message.destroy(key);
        }
    }, [list.localServer.id, list.reload, message, selectedId]);

    const confirmDelete = useCallback((server: ServerRecord) => {
        modal.confirm({
            title: "删除终端",
            content: `确定删除终端“${server.name || server.ip}”吗？`,
            okText: "删除",
            cancelText: "取消",
            okButtonProps: {danger: true, type: "default"},
            mask: {closable: true},
            onOk: () => remove(server),
        } as any);
    }, [modal, remove]);
    return (
        <Body space={false}>
            <div ref={pageRef} className={styles.page} style={pageStyle}>
                <Card className={styles.workspace} variant="borderless">
                    <main className={[
                        styles.terminalLayout,
                        serverCollapsed ? "server-collapsed" : "",
                        scriptsCollapsed ? "scripts-collapsed" : "",
                    ].filter(Boolean).join(" ")}>
                        <Sidebar
                            styles={styles}
                            collapsed={serverCollapsed}
                            servers={list.servers}
                            loading={list.loading || list.localRefreshing}
                            loadingMore={list.loadingMore}
                            loaded={list.loaded}
                            hasMore={list.hasMore}
                            localLoading={list.localRefreshing && !list.localLoaded}
                            localError={list.localError}
                            keyword={list.keyword}
                            selectedId={selectedId}
                            connectionStatuses={connectionStatuses}
                            authTypeData={authTypeData}
                            onSelect={selectServer}
                            onConnect={connectServer}
                            onEdit={openEdit}
                            onDelete={confirmDelete}
                            onReload={list.reload}
                            onLoadMore={list.loadMore}
                            onKeywordChange={list.setKeyword}
                            onCreate={openCreate}
                            onCollapsedChange={setServerCollapsed}/>
                        {Object.values(sessions).map((session) => {
                            const status = connectionStatuses[String(session.server.id)] || "idle";
                            const sessionActive = status === "connected" || status === "connecting";
                            return (
                                <Terminal
                                    key={String(session.server.id)}
                                    ref={getTerminalRef(session.server.id)}
                                    styles={styles}
                                    server={session.server}
                                    active={session.server.id === selectedId}
                                    initializing={session.server.id === 0 && !sessionActive &&
                                        list.localRefreshing && !list.localLoaded}
                                    unavailable={session.server.id === 0 && !sessionActive && !list.localRefreshing &&
                                        (!list.localLoaded || list.localError)}
                                    resetKey={session.resetKey}
                                    compact={isCompactMode}
                                    onStatusChange={changeSessionStatus}/>
                            );
                        })}
                        <Scripts
                            styles={styles}
                            collapsed={scriptsCollapsed}
                            onCollapsedChange={setScriptsCollapsed}
                            onInsert={insertCommand}/>
                    </main>
                </Card>
            </div>
            <Form ref={formRef} authTypeData={authTypeData} onSuccess={handleFormSuccess}/>
        </Body>
    );
};
