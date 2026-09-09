import React, {useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {Button, Empty, Spin, Tooltip, Tree} from "antd";
import type {MenuProps, TreeProps} from "antd";
import {Icon} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import {getFileIconType} from "../../utils";
import type {FileEditorTreeNode} from "../editor/utils";

type TreeSelectHandler = NonNullable<TreeProps<FileEditorTreeNode>["onSelect"]>;
type TreeExpandHandler = NonNullable<TreeProps<FileEditorTreeNode>["onExpand"]>;
type TreeSelectArgs = Parameters<TreeSelectHandler>;
type TreeExpandArgs = Parameters<TreeExpandHandler>;

const useStableEvent = <T extends (...args: any[]) => any>(callback: T): T => {
    const callbackRef = useRef(callback);
    callbackRef.current = callback;
    return useCallback(((...args: any[]) => callbackRef.current(...args)) as T, []);
};

export interface FileEditorTreeProps {
    treeData: FileEditorTreeNode[];
    expandedKeys: string[];
    selectedKeys: string[];
    loadedKeys: string[];
    loadingPaths: ReadonlySet<string>;
    initializing: boolean;
    locateToken: number;
    targetDirectory: string;
    disabled: boolean;
    layoutReady: boolean;
    tooltipsDisabled: boolean;
    onExpand: (keys: string[]) => void;
    onSelect: (node: FileEditorTreeNode) => void;
    onLoadData: NonNullable<TreeProps<FileEditorTreeNode>["loadData"]>;
    onRefresh: () => void;
    onCreate: (mode: any, parentPath?: string) => void;
    onRename: (node: FileEditorTreeNode) => void;
    onPermissions: (node: FileEditorTreeNode) => void;
    onDelete: (node: FileEditorTreeNode) => void;
    onCopy: (value: string, label: "名称" | "路径") => Promise<void>;
    getPopupContainer: (triggerNode: HTMLElement) => HTMLElement;
    menuClassName: string;
    itemHeight: number;
}

interface FileEditorTreeNodeTitleProps {
    node: FileEditorTreeNode;
    loading: boolean;
    disabled: boolean;
    tooltipsDisabled: boolean;
    onToggleDirectory: (event: React.MouseEvent, node: FileEditorTreeNode) => void;
    onRename: (node: FileEditorTreeNode) => void;
    onCreate: (mode: any, parentPath?: string) => void;
    onPermissions: (node: FileEditorTreeNode) => void;
    onDelete: (node: FileEditorTreeNode) => void;
    onCopy: (value: string, label: "名称" | "路径") => Promise<void>;
    menuOpen: boolean;
    onOpenMenu: (node: FileEditorTreeNode) => void;
    onCloseMenu: () => void;
    getPopupContainer: (triggerNode: HTMLElement) => HTMLElement;
    menuClassName: string;
}

const FileEditorTreeNodeTitle = React.memo(({
    node,
    loading,
    disabled,
    tooltipsDisabled,
    onToggleDirectory,
    onRename,
    onCreate,
    onPermissions,
    onDelete,
    onCopy,
    menuOpen,
    onOpenMenu,
    onCloseMenu,
    getPopupContainer,
    menuClassName,
}: FileEditorTreeNodeTitleProps) => {
    const iconType = node.kind === "more"
        ? "EllipsisOutlined"
        : getFileIconType(node.name, node.isDirectory);
    const iconClassName = node.isDirectory ? "is-directory" : "";
    const title = (
        <span
            className={`file-editor-tree-node ${node.kind === "more" ? "is-more" : ""}`}
            onDoubleClick={(event) => onToggleDirectory(event, node)}>
            {loading && node.kind === "entry" ? (
                <Icon className="anticon-spin" type="LoadingOutlined"/>
            ) : (
                <Icon className={iconClassName} type={iconType}/>
            )}
            <span
                className="file-editor-tree-name"
                title={tooltipsDisabled ? undefined : node.title}>
                {node.title}
            </span>
        </span>
    );
    if (node.kind !== "entry") {
        return title;
    }
    const isRoot = !node.record;
    const menuItems: MenuProps["items"] = menuOpen ? (isRoot ? [
        {key: "create-directory", icon: <Icon type="FolderAddOutlined"/>, label: "新建文件夹"},
        {key: "create-file", icon: <Icon type="FileAddOutlined"/>, label: "新建空文件"},
        {type: "divider"},
        {key: "copy-path", icon: <Icon type="LinkOutlined"/>, label: "复制路径"},
    ] : node.isDirectory ? [
        {key: "create-directory", icon: <Icon type="FolderAddOutlined"/>, label: "新建文件夹"},
        {key: "create-file", icon: <Icon type="FileAddOutlined"/>, label: "新建空文件"},
        {type: "divider"},
        {key: "rename", icon: <Icon type="EditOutlined"/>, label: "重命名"},
        {key: "permissions", icon: <Icon type="SafetyOutlined"/>, label: "修改权限"},
        {key: "copy-name", icon: <Icon type="CopyOutlined"/>, label: "复制名称"},
        {key: "copy-path", icon: <Icon type="LinkOutlined"/>, label: "复制路径"},
        {type: "divider"},
        {key: "delete", icon: <Icon type="DeleteOutlined"/>, label: "删除", danger: true},
    ] : [
        {key: "rename", icon: <Icon type="EditOutlined"/>, label: "重命名"},
        {key: "permissions", icon: <Icon type="SafetyOutlined"/>, label: "修改权限"},
        {key: "copy-name", icon: <Icon type="CopyOutlined"/>, label: "复制名称"},
        {key: "copy-path", icon: <Icon type="LinkOutlined"/>, label: "复制路径"},
        {type: "divider"},
        {key: "delete", icon: <Icon type="DeleteOutlined"/>, label: "删除", danger: true},
    ]) : [];
    const menuButton = (
        <Button
            className="file-editor-tree-node-more"
            type="text"
            size="small"
            disabled={disabled}
            aria-label={`管理 ${node.name}`}
            aria-haspopup="menu"
            aria-expanded={menuOpen}
            icon={<Icon type="MoreOutlined"/>}
            onMouseDown={(event) => event.stopPropagation()}
            onClick={(event) => {
                event.preventDefault();
                event.stopPropagation();
                if (menuOpen) {
                    onCloseMenu();
                } else {
                    onOpenMenu(node);
                }
            }}/>
    );
    return (
        <span className="file-editor-tree-node-shell">
            {title}
            {menuOpen ? (
                <Dropdown
                    open
                    destroyOnHidden
                    trigger={["click"]}
                    placement="bottom"
                    autoAdjustOverflow={{adjustX: true, adjustY: true}}
                    arrow={{pointAtCenter: true}}
                    classNames={{root: menuClassName}}
                    getPopupContainer={getPopupContainer}
                    onOpenChange={(open) => {
                        if (!open) {
                            onCloseMenu();
                        }
                    }}
                    menu={{
                        items: menuItems,
                        onClick: (info) => {
                            info.domEvent?.preventDefault();
                            info.domEvent?.stopPropagation();
                            onCloseMenu();
                            if (info.key === "create-directory") {
                                onCreate("directory", node.path);
                            } else if (info.key === "create-file") {
                                onCreate("file", node.path);
                            } else if (info.key === "rename") {
                                onRename(node);
                            } else if (info.key === "permissions") {
                                onPermissions(node);
                            } else if (info.key === "copy-name") {
                                void onCopy(node.name, "名称");
                            } else if (info.key === "copy-path") {
                                void onCopy(node.path, "路径");
                            } else if (info.key === "delete") {
                                onDelete(node);
                            }
                        },
                    }}>
                    {menuButton}
                </Dropdown>
            ) : menuButton}
        </span>
    );
});

const FileEditorTree = ({
    treeData,
    expandedKeys,
    selectedKeys,
    loadedKeys,
    loadingPaths,
    initializing,
    locateToken,
    targetDirectory,
    disabled,
    layoutReady,
    tooltipsDisabled,
    onExpand,
    onSelect,
    onLoadData,
    onRefresh,
    onCreate,
    onRename,
    onPermissions,
    onDelete,
    onCopy,
    getPopupContainer,
    menuClassName,
    itemHeight,
}: FileEditorTreeProps) => {
    const treeBodyRef = useRef<HTMLDivElement | null>(null);
    const treeRef = useRef<any>(null);
    const initialLocateDoneRef = useRef(false);
    const locateTokenRef = useRef<number | undefined>(undefined);
    const expandedKeysRef = useRef(expandedKeys);
    expandedKeysRef.current = expandedKeys;
    const [treeHeight, setTreeHeight] = useState(0);
    const [menuNodeKey, setMenuNodeKey] = useState<string>();
    const stableOnExpand = useStableEvent(onExpand);
    const stableOnSelect = useStableEvent(onSelect);
    const stableOnLoadData = useStableEvent(onLoadData);
    const stableOnRefresh = useStableEvent(onRefresh);
    const stableOnCreate = useStableEvent(onCreate);
    const stableOnRename = useStableEvent(onRename);
    const stableOnPermissions = useStableEvent(onPermissions);
    const stableOnDelete = useStableEvent(onDelete);
    const stableOnCopy = useStableEvent(onCopy);
    const stableGetPopupContainer = useStableEvent(getPopupContainer);
    const directoryName = useMemo(() => {
        const normalized = targetDirectory.replace(/[\\/]+$/, "");
        const parts = normalized.split(/[\\/]/).filter(Boolean);
        return parts[parts.length - 1] || targetDirectory || "目录";
    }, [targetDirectory]);
    const openNodeMenu = useCallback((node: FileEditorTreeNode) => setMenuNodeKey(node.key), []);
    const closeNodeMenu = useCallback(() => setMenuNodeKey(undefined), []);

    useLayoutEffect(() => {
        const body = treeBodyRef.current;
        if (!body) {
            return;
        }
        let measuredHeight = -1;
        const commitHeight = (height: number) => {
            const next = Math.max(0, Math.floor(height));
            if (next <= 0 || next === measuredHeight) {
                return;
            }
            measuredHeight = next;
            setTreeHeight((current) => current === next ? current : next);
        };
        const updateHeight = () => {
            const styles = window.getComputedStyle(body);
            const verticalPadding = (Number.parseFloat(styles.paddingTop) || 0)
                + (Number.parseFloat(styles.paddingBottom) || 0);
            commitHeight(body.clientHeight - verticalPadding);
        };
        updateHeight();
        const observer = typeof ResizeObserver === "function"
            ? new ResizeObserver((entries) => {
                const height = entries[0]?.contentRect.height;
                if (typeof height === "number") {
                    commitHeight(height);
                } else {
                    updateHeight();
                }
            })
            : undefined;
        observer?.observe(body);
        return () => observer?.disconnect();
    }, []);

    useLayoutEffect(() => {
        if (!layoutReady || treeHeight <= 0) {
            return;
        }
        let frame = 0;
        const syncScroll = () => {
            const holder = treeBodyRef.current?.querySelector<HTMLElement>(".ant-tree-list-holder");
            if (holder) {
                treeRef.current?.scrollTo?.(holder.scrollTop);
            }
        };
        frame = window.requestAnimationFrame(syncScroll);
        const settleTimer = window.setTimeout(syncScroll, 240);
        return () => {
            window.cancelAnimationFrame(frame);
            window.clearTimeout(settleTimer);
        };
    }, [expandedKeys, layoutReady, treeData, treeHeight]);

    const toggleDirectory = useCallback((event: React.MouseEvent, node: FileEditorTreeNode) => {
        event.preventDefault();
        event.stopPropagation();
        if (node.kind !== "entry" || !node.isDirectory || disabled) {
            return;
        }
        const currentKeys = expandedKeysRef.current;
        stableOnExpand(currentKeys.includes(node.key)
            ? currentKeys.filter((key) => key !== node.key)
            : [...currentKeys, node.key]);
    }, [disabled, stableOnExpand]);

    const titleRender = useCallback((node: FileEditorTreeNode) => {
        return <FileEditorTreeNodeTitle
            node={node}
            loading={loadingPaths.has(node.key)}
            disabled={disabled}
            tooltipsDisabled={tooltipsDisabled}
            onToggleDirectory={toggleDirectory}
            onRename={stableOnRename}
            onCreate={stableOnCreate}
            onPermissions={stableOnPermissions}
            onDelete={stableOnDelete}
            onCopy={stableOnCopy}
            menuOpen={menuNodeKey === node.key}
            onOpenMenu={openNodeMenu}
            onCloseMenu={closeNodeMenu}
            getPopupContainer={stableGetPopupContainer}
            menuClassName={menuClassName}/>;
    }, [closeNodeMenu, disabled, loadingPaths, menuClassName, menuNodeKey, openNodeMenu, stableGetPopupContainer, stableOnCopy, stableOnCreate, stableOnDelete, stableOnPermissions, stableOnRename, toggleDirectory, tooltipsDisabled]);

    useEffect(() => {
        if (locateTokenRef.current !== locateToken) {
            locateTokenRef.current = locateToken;
            initialLocateDoneRef.current = false;
        }
        const selectedKey = selectedKeys[0];
        const hasTree = treeData.length > 0;
        if (initializing || initialLocateDoneRef.current || !selectedKey || !hasTree) {
            return;
        }
        const renderFrame = window.requestAnimationFrame(() => {
            treeRef.current?.scrollTo?.({key: selectedKey, align: "auto"});
            initialLocateDoneRef.current = true;
        });
        return () => window.cancelAnimationFrame(renderFrame);
    }, [initializing, locateToken, selectedKeys[0], treeData.length]);

    const select = useCallback((_: TreeSelectArgs[0], info: TreeSelectArgs[1]) => {
        closeNodeMenu();
        stableOnSelect(info.node);
    }, [closeNodeMenu, stableOnSelect]);

    const expand = useCallback((keys: TreeExpandArgs[0]) => {
        closeNodeMenu();
        stableOnExpand(keys.map(String));
    }, [closeNodeMenu, stableOnExpand]);

    const handleTreeScroll = useCallback(() => {
        setMenuNodeKey(undefined);
    }, []);

    return (
        <aside className="file-editor-sidebar" aria-label="文件目录">
            <div className="file-editor-tree-toolbar">
                <div
                    className="file-editor-tree-target"
                    title={tooltipsDisabled ? undefined : `双击复制路径：${targetDirectory}`}
                    onDoubleClick={(event) => {
                        event.preventDefault();
                        event.stopPropagation();
                        void onCopy(targetDirectory, "路径");
                    }}>
                    <Icon type="FolderOpenOutlined" aria-hidden/>
                    <span>{directoryName}</span>
                </div>
                <div className="file-editor-tree-actions">
                    <Tooltip title="刷新目录" open={tooltipsDisabled ? false : undefined}>
                        <Button
                            type="text"
                            disabled={disabled || !targetDirectory}
                            aria-label="刷新当前目录"
                            icon={<Icon type="ReloadOutlined"/>}
                            onClick={stableOnRefresh}/>
                    </Tooltip>
                </div>
            </div>
            <div ref={treeBodyRef} className="file-editor-tree-body">
                {initializing || (treeData.length > 0 && treeHeight <= 0) ? (
                    <div className="file-editor-tree-loading" role="status" aria-label="正在加载目录">
                        <Spin/>
                    </div>
                ) : treeData.length > 0 ? (
                    <Tree<FileEditorTreeNode>
                        ref={treeRef}
                        blockNode
                        showIcon={false}
                        virtual
                        height={treeHeight}
                        itemHeight={itemHeight}
                        disabled={disabled}
                        treeData={treeData}
                        expandedKeys={expandedKeys}
                        selectedKeys={selectedKeys}
                        loadedKeys={loadedKeys}
                        loadData={stableOnLoadData}
                        titleRender={titleRender}
                        onExpand={expand}
                        onScroll={menuNodeKey ? handleTreeScroll : undefined}
                        onSelect={select}/>
                ) : (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
                )}
            </div>
        </aside>
    );
};

export default React.memo(FileEditorTree);
