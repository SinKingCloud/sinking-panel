import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Button, Empty, Input, Spin, Tooltip, Tree} from "antd";
import type {InputRef, MenuProps, TreeProps} from "antd";
import {Icon} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import {getFileIconType} from "../utils";
import type {FileEditorTreeNode} from "./editor.utils";

type TreeSelectHandler = NonNullable<TreeProps<FileEditorTreeNode>["onSelect"]>;
type TreeExpandHandler = NonNullable<TreeProps<FileEditorTreeNode>["onExpand"]>;
type TreeSelectArgs = Parameters<TreeSelectHandler>;
type TreeExpandArgs = Parameters<TreeExpandHandler>;

interface InlineRenameInputProps {
    node: FileEditorTreeNode;
    onCancel: () => void;
    onSubmit: (node: FileEditorTreeNode, name: string) => Promise<boolean>;
}

const InlineRenameInput = ({node, onCancel, onSubmit}: InlineRenameInputProps) => {
    const inputRef = useRef<InputRef>(null);
    const mountedRef = useRef(true);
    const submittingRef = useRef(false);
    const [value, setValue] = useState(node.name);
    const [submitting, setSubmitting] = useState(false);

    useEffect(() => {
        const frame = window.requestAnimationFrame(() => {
            inputRef.current?.focus();
            inputRef.current?.select();
        });
        return () => window.cancelAnimationFrame(frame);
    }, []);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
        };
    }, []);

    const submit = useCallback(async () => {
        if (submittingRef.current) {
            return;
        }
        if (value === node.name) {
            onCancel();
            return;
        }
        submittingRef.current = true;
        setSubmitting(true);
        let success = false;
        try {
            success = await onSubmit(node, value);
        } catch {
            success = false;
        } finally {
            submittingRef.current = false;
        }
        if (success) {
            onCancel();
            return;
        }
        if (!mountedRef.current) {
            return;
        }
        setSubmitting(false);
        window.requestAnimationFrame(() => inputRef.current?.focus({cursor: "all"}));
    }, [node, onCancel, onSubmit, value]);

    const stopPropagation = (event: React.SyntheticEvent) => event.stopPropagation();

    return (
        <Input
            ref={inputRef}
            className="file-editor-tree-rename"
            size="small"
            maxLength={255}
            value={value}
            readOnly={submitting}
            aria-label={`重命名 ${node.name}`}
            suffix={(
                <span className="file-editor-tree-rename-status" aria-hidden>
                    {submitting && <Icon className="anticon-spin" type="LoadingOutlined"/>}
                </span>
            )}
            onChange={(event) => setValue(event.target.value)}
            onMouseDown={stopPropagation}
            onClick={stopPropagation}
            onBlur={() => void submit()}
            onKeyDown={(event) => {
                event.stopPropagation();
                if (event.nativeEvent.isComposing || event.keyCode === 229) {
                    return;
                }
                if (event.key === "Enter") {
                    event.preventDefault();
                    void submit();
                } else if (event.key === "Escape" && !submittingRef.current) {
                    event.preventDefault();
                    onCancel();
                }
            }}/>
    );
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
    tooltipsDisabled: boolean;
    onExpand: (keys: string[]) => void;
    onSelect: (node: FileEditorTreeNode) => void;
    onLoadData: NonNullable<TreeProps<FileEditorTreeNode>["loadData"]>;
    onRefresh: () => void;
    onCreate: (mode: any, parentPath?: string) => void;
    onRename: (node: FileEditorTreeNode, name: string) => Promise<boolean>;
    onPermissions: (node: FileEditorTreeNode) => void;
    onDelete: (node: FileEditorTreeNode) => void;
    onCopy: (value: string, label: "名称" | "路径") => Promise<void>;
    getPopupContainer: (triggerNode: HTMLElement) => HTMLElement;
    menuClassName: string;
}

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
}: FileEditorTreeProps) => {
    const treeBodyRef = useRef<HTMLDivElement | null>(null);
    const initialLocateDoneRef = useRef(false);
    const locateTokenRef = useRef<number | undefined>(undefined);
    const expandedKeysRef = useRef(expandedKeys);
    expandedKeysRef.current = expandedKeys;
    const [editingNode, setEditingNode] = useState<FileEditorTreeNode>();
    const directoryName = useMemo(() => {
        const normalized = targetDirectory.replace(/[\\/]+$/, "");
        const parts = normalized.split(/[\\/]/).filter(Boolean);
        return parts[parts.length - 1] || targetDirectory || "目录";
    }, [targetDirectory]);
    const beginRename = useCallback((node: FileEditorTreeNode) => {
        if (!disabled && !editingNode && node.kind === "entry" && node.record) {
            setEditingNode(node);
        }
    }, [disabled, editingNode]);

    const cancelRename = useCallback(() => setEditingNode(undefined), []);
    const toggleDirectory = useCallback((event: React.MouseEvent, node: FileEditorTreeNode) => {
        if (editingNode?.key === node.key) {
            return;
        }
        event.preventDefault();
        event.stopPropagation();
        if (node.kind !== "entry" || !node.isDirectory || disabled || editingNode) {
            return;
        }
        const currentKeys = expandedKeysRef.current;
        onExpand(currentKeys.includes(node.key)
            ? currentKeys.filter((key) => key !== node.key)
            : [...currentKeys, node.key]);
    }, [disabled, editingNode, onExpand]);

    const titleRender = useCallback((node: FileEditorTreeNode) => {
        const iconType = node.kind === "more"
            ? "EllipsisOutlined"
            : getFileIconType(node.name, node.isDirectory);
        const iconClassName = node.isDirectory ? "is-directory" : "";
        const title = (
            <span
                className={`file-editor-tree-node ${node.kind === "more" ? "is-more" : ""}`}
                onDoubleClick={(event) => toggleDirectory(event, node)}>
                {loadingPaths.has(node.key) && node.kind === "entry" ? (
                    <Icon className="anticon-spin" type="LoadingOutlined"/>
                ) : (
                    <Icon className={iconClassName} type={iconType}/>
                )}
                {editingNode?.key === node.key ? (
                    <InlineRenameInput
                        node={editingNode}
                        onCancel={cancelRename}
                        onSubmit={onRename}/>
                ) : (
                    <span
                        className="file-editor-tree-name"
                        title={tooltipsDisabled ? undefined : node.title}>
                        {node.title}
                    </span>
                )}
            </span>
        );
        if (node.kind !== "entry") {
            return title;
        }
        const isRoot = !node.record;
        const menuItems: MenuProps["items"] = isRoot ? [
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
        ];
        return (
            <span className="file-editor-tree-node-shell">
                {title}
                <Dropdown
                    key={`${node.key}-${tooltipsDisabled ? "fullscreen" : "window"}`}
                    trigger={["click"]}
                    autoAdjustOverflow={{adjustX: true, adjustY: true}}
                    arrow={{pointAtCenter: true}}
                    classNames={{root: menuClassName}}
                    getPopupContainer={getPopupContainer}
                    menu={{
                        items: menuItems,
                        onClick: (info) => {
                            if (info.key === "create-directory") {
                                onCreate("directory", node.path);
                            } else if (info.key === "create-file") {
                                onCreate("file", node.path);
                            } else if (info.key === "rename") {
                                beginRename(node);
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
                    <Button
                        className="file-editor-tree-node-more"
                        type="text"
                        size="small"
                        disabled={disabled}
                        aria-label={`管理 ${node.name}`}
                        aria-haspopup="menu"
                        icon={<Icon type="MoreOutlined"/>}/>
                </Dropdown>
            </span>
        );
    }, [beginRename, cancelRename, disabled, editingNode, getPopupContainer, loadingPaths, menuClassName, onCopy, onCreate, onDelete, onPermissions, onRename, toggleDirectory, tooltipsDisabled]);

    useEffect(() => {
        if (locateTokenRef.current !== locateToken) {
            locateTokenRef.current = locateToken;
            initialLocateDoneRef.current = false;
        }
        const selectedKey = selectedKeys[0];
        if (initializing || initialLocateDoneRef.current || !selectedKey || treeData.length === 0) {
            return;
        }
        let locateFrame = 0;
        const renderFrame = window.requestAnimationFrame(() => {
            locateFrame = window.requestAnimationFrame(() => {
                const selectedNode = treeBodyRef.current
                    ?.querySelector<HTMLElement>(".ant-tree-node-selected");
                const treeBody = treeBodyRef.current;
                if (selectedNode && treeBody) {
                    const nodeRect = selectedNode.getBoundingClientRect();
                    const bodyRect = treeBody.getBoundingClientRect();
                    if (nodeRect.top < bodyRect.top) {
                        treeBody.scrollTop += nodeRect.top - bodyRect.top;
                    } else if (nodeRect.bottom > bodyRect.bottom) {
                        treeBody.scrollTop += nodeRect.bottom - bodyRect.bottom;
                    }
                    initialLocateDoneRef.current = true;
                }
            });
        });
        return () => {
            window.cancelAnimationFrame(renderFrame);
            if (locateFrame) {
                window.cancelAnimationFrame(locateFrame);
            }
        };
    }, [expandedKeys, initializing, locateToken, selectedKeys, treeData]);

    const select = useCallback((_: TreeSelectArgs[0], info: TreeSelectArgs[1]) => {
        if (editingNode) {
            return;
        }
        onSelect(info.node);
    }, [editingNode, onSelect]);

    const expand = useCallback((keys: TreeExpandArgs[0]) => {
        onExpand(keys.map(String));
    }, [onExpand]);

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
                            onClick={onRefresh}/>
                    </Tooltip>
                </div>
            </div>
            <div ref={treeBodyRef} className="file-editor-tree-body">
                {initializing ? (
                    <div className="file-editor-tree-loading" role="status" aria-label="正在加载目录">
                        <Spin/>
                    </div>
                ) : treeData.length > 0 ? (
                    <Tree<FileEditorTreeNode>
                        blockNode
                        showIcon={false}
                        disabled={disabled}
                        treeData={treeData}
                        expandedKeys={expandedKeys}
                        selectedKeys={selectedKeys}
                        loadedKeys={loadedKeys}
                        loadData={onLoadData}
                        titleRender={titleRender}
                        onExpand={expand}
                        onSelect={select}/>
                ) : (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无目录"/>
                )}
            </div>
        </aside>
    );
};

export default React.memo(FileEditorTree);
