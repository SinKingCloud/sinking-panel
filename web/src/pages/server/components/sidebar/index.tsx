import React, {useCallback, useEffect, useMemo, useState} from "react";
import {Button, Empty, Input, Spin, Tooltip} from "antd";
import {Icon, Title} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import type {ConnectionStatus} from "../terminal";
import type {ServerRecord} from "../../hooks/servers";

interface SidebarProps {
    styles: any;
    collapsed: boolean;
    servers: ServerRecord[];
    loading: boolean;
    loadingMore: boolean;
    loaded: boolean;
    initializing: boolean;
    hasMore: boolean;
    localLoading: boolean;
    localError: boolean;
    keyword: string;
    selectedId: number;
    connectionStatuses: Record<string, ConnectionStatus>;
    authTypeData: Record<string, string>;
    onSelect: (server: ServerRecord) => void;
    onConnect: (server: ServerRecord) => void;
    onEdit: (server: ServerRecord) => void;
    onDelete: (server: ServerRecord) => void;
    onReload: () => void;
    onLoadMore: () => void;
    onKeywordChange: (keyword: string) => void;
    onCreate: () => void;
    onCollapsedChange: (collapsed: boolean) => void;
}

const renderBatchSize = 80;

const Sidebar = ({
    styles,
    collapsed,
    servers,
    loading,
    loadingMore,
    loaded,
    initializing,
    hasMore,
    localLoading,
    localError,
    keyword,
    selectedId,
    connectionStatuses,
    authTypeData,
    onSelect,
    onConnect,
    onEdit,
    onDelete,
    onReload,
    onLoadMore,
    onKeywordChange,
    onCreate,
    onCollapsedChange,
}: SidebarProps) => {
    const [visibleCount, setVisibleCount] = useState(renderBatchSize);
    const visibleItems = useMemo(
        () => servers.slice(0, visibleCount),
        [servers, visibleCount],
    );

    useEffect(() => {
        setVisibleCount(renderBatchSize);
    }, [keyword]);

    const loadVisibleItems = useCallback((event: React.UIEvent<HTMLDivElement>) => {
        const target = event.currentTarget;
        const horizontal = target.scrollWidth > target.clientWidth + 1;
        const remaining = horizontal
            ? target.scrollWidth - target.scrollLeft - target.clientWidth
            : target.scrollHeight - target.scrollTop - target.clientHeight;
        if (remaining < 180) {
            setVisibleCount((current) => Math.min(servers.length, current + renderBatchSize));
            if (visibleCount >= servers.length && hasMore && !loadingMore) {
                onLoadMore();
            }
        }
    }, [hasMore, loadingMore, onLoadMore, servers.length, visibleCount]);

    const renderServer = (server: ServerRecord) => {
        const selected = selectedId === server.id;
        const local = server.id === 0;
        const connectionStatus = connectionStatuses[String(server.id)] || "idle";
        const sessionActive = connectionStatus === "connected" || connectionStatus === "connecting";
        const pending = local && localLoading && !sessionActive;
        const failed = local && localError && !sessionActive;
        const disabled = pending;
        const status = pending ? "connecting" : failed ? "error" : connectionStatus;
        const configured = Boolean(server.ip && server.user);
        const endpoint = pending
            ? "正在加载本机连接..."
            : failed
                ? "本机连接加载失败"
                : configured
                    ? `${server.user}@${server.ip}:${server.port}`
                    : local ? "尚未配置本机 SSH" : `${server.ip}:${server.port}`;
        const authType = authTypeData[String(server.auth_type)] || "密码验证";
        const menuItems = [
            {
                key: "edit",
                label: "编辑",
                disabled: pending,
                onClick: () => onEdit(server),
            },
            ...(!local ? [
                {type: "divider" as const},
                    {
                        key: "delete",
                        label: "删除",
                        danger: true,
                        onClick: () => onDelete(server),
                },
            ] : []),
        ];

        return (
            <div
                key={String(server.id)}
                className={`${styles.serverItem} ${selected ? "selected" : ""}`}>
                <button
                    className="server-select"
                    type="button"
                    disabled={disabled}
                    title={`${server.name || endpoint} · ${authType}`}
                    onClick={() => onSelect(server)}
                    onDoubleClick={() => onConnect(server)}>
                    <span className={`connection-dot ${status}`} title={status === "connected" ? "已连接" : undefined}/>
                    <span className="server-icon">
                        <Icon type={local ? "DesktopOutlined" : "CloudServerOutlined"}/>
                    </span>
                    <span className="server-copy">
                        <span className="server-name" title={server.name || endpoint}>{server.name || endpoint}</span>
                        <span className="server-endpoint" title={endpoint}>{endpoint}</span>
                    </span>
                </button>
                <div className="server-actions">
                    <Dropdown
                        disabled={pending}
                        trigger={["click"]}
                        placement="bottom"
                        arrow={{pointAtCenter: true}}
                        classNames={{root: styles.serverMenu}}
                        menu={{items: menuItems}}>
                        <Button
                            className="server-action"
                            color="default"
                            variant="text"
                            disabled={pending}
                            aria-label={`${server.name || endpoint}的更多操作`}>
                            <Icon type="MoreOutlined"/>
                        </Button>
                    </Dropdown>
                </div>
            </div>
        );
    };

    return (
        <aside className={`${styles.serverPane} ${collapsed ? "collapsed" : ""}`} aria-label="终端列表">
            <header className={styles.serverHeader}>
                <div className="panel-heading">
                    <Title size="small">终端列表</Title>
                </div>
                <Button
                    className="collapse-trigger"
                    type="text"
                    aria-label={collapsed ? "展开终端列表" : "收起终端列表"}
                    aria-controls="terminal-server-panel"
                    aria-expanded={!collapsed}
                    icon={<Icon type={collapsed ? "MenuUnfoldOutlined" : "MenuFoldOutlined"}/>}
                    onClick={() => {
                        if (!collapsed) onKeywordChange("");
                        onCollapsedChange(!collapsed);
                    }}/>
            </header>
            <div className={styles.serverToolbar}>
                <Input
                    variant="filled"
                    allowClear
                    value={keyword}
                    aria-label="搜索终端"
                    prefix={<Icon type="SearchOutlined"/>}
                    placeholder="搜索终端"
                    onChange={(event) => onKeywordChange(event.target.value)}/>
                <Button
                    type="text"
                    disabled={loading}
                    aria-label="刷新终端列表"
                    icon={<Icon type="ReloadOutlined"/>}
                    onClick={onReload}/>
            </div>
            <div id="terminal-server-panel" className={styles.serverListArea} aria-busy={loading}>
                <div className={styles.serverList} onScroll={loadVisibleItems}>
                    {initializing ? (
                        <div className="server-initial-loading" role="status" aria-label="正在加载终端列表">
                            <Spin size="small"/>
                        </div>
                    ) : (
                        <>
                            <button className={`${styles.serverAdd} mobile`} type="button" onClick={onCreate}>
                                <Icon type="PlusOutlined"/>
                                <span>添加终端</span>
                            </button>
                            {visibleItems.map(renderServer)}
                            {loaded && servers.length === 0 && keyword.trim() && !loading && (
                                <div className="server-empty">
                                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
                                </div>
                            )}
                            {loaded && loadingMore && (
                                <div className="server-loading"><Spin size="small"/></div>
                            )}
                            {loaded && hasMore && !loadingMore && (
                                <div className="server-load-more">
                                    <Tooltip title="加载更多终端">
                                        <Button
                                            type="text"
                                            aria-label="加载更多终端"
                                            icon={<Icon type="DownOutlined"/>}
                                            onClick={onLoadMore}/>
                                    </Tooltip>
                                </div>
                            )}
                        </>
                    )}
                </div>
                {loading && !initializing && (
                    <div className={styles.serverLoadingOverlay} role="status" aria-label="正在加载终端列表">
                        <Spin size="small"/>
                    </div>
                )}
            </div>
            <footer className={styles.serverFooter}>
                <button className={styles.serverAdd} type="button" aria-label="添加终端" onClick={onCreate}>
                    <Icon type="PlusOutlined"/>
                    <span className="add-label">添加终端</span>
                </button>
            </footer>
        </aside>
    );
};

export default React.memo(Sidebar);
