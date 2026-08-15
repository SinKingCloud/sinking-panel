import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Button, Empty, Input, Tooltip, Tree} from "antd";
import type {InputRef, MenuProps, TreeProps} from "antd";
import {Icon} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import type {FileCreateMode} from "../types";
import {getFileIconType} from "../utils";
import type {FileEditorTreeNode} from "./editor.utils";

type TreeSelectHandler = NonNullable<TreeProps<FileEditorTreeNode>["onSelect"]>;
type TreeExpandHandler = NonNullable<TreeProps<FileEditorTreeNode>["onExpand"]>;
type MenuClickHandler = NonNullable<MenuProps["onClick"]>;
type TreeSelectArgs = Parameters<TreeSelectHandler>;
type TreeExpandArgs = Parameters<TreeExpandHandler>;
type MenuClickArgs = Parameters<MenuClickHandler>;

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
            onDoubleClick={stopPropagation}
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
    targetDirectory: string;
    disabled: boolean;
    tooltipsDisabled: boolean;
    onExpand: (keys: string[]) => void;
    onSelect: (node: FileEditorTreeNode) => void;
    onLoadData: NonNullable<TreeProps<FileEditorTreeNode>["loadData"]>;
    onRefresh: () => void;
    onCreate: (mode: FileCreateMode) => void;
    onRename: (node: FileEditorTreeNode, name: string) => Promise<boolean>;
}

const FileEditorTree = ({
    treeData,
    expandedKeys,
    selectedKeys,
    loadedKeys,
    loadingPaths,
    targetDirectory,
    disabled,
    tooltipsDisabled,
    onExpand,
    onSelect,
    onLoadData,
    onRefresh,
    onCreate,
    onRename,
}: FileEditorTreeProps) => {
    const treeBodyRef = useRef<HTMLDivElement | null>(null);
    const locatedKeyRef = useRef<string | undefined>(undefined);
    const selectTimerRef = useRef<number | undefined>(undefined);
    const [editingNode, setEditingNode] = useState<FileEditorTreeNode>();
    const directoryName = useMemo(() => {
        const normalized = targetDirectory.replace(/[\\/]+$/, "");
        const parts = normalized.split(/[\\/]/).filter(Boolean);
        return parts[parts.length - 1] || targetDirectory || "目录";
    }, [targetDirectory]);
    const createItems: MenuProps["items"] = useMemo(() => [
        {key: "directory", icon: <Icon type="FolderAddOutlined"/>, label: "新建文件夹"},
        {key: "file", icon: <Icon type="FileAddOutlined"/>, label: "新建空文件"},
    ], []);

    const beginRename = useCallback((event: React.MouseEvent, node: FileEditorTreeNode) => {
        event.preventDefault();
        event.stopPropagation();
        if (selectTimerRef.current !== undefined) {
            window.clearTimeout(selectTimerRef.current);
            selectTimerRef.current = undefined;
        }
        if (!disabled && !editingNode && node.kind === "entry" && node.record) {
            setEditingNode(node);
        }
    }, [disabled, editingNode]);

    const cancelRename = useCallback(() => setEditingNode(undefined), []);

    const titleRender = useCallback((node: FileEditorTreeNode) => {
        const iconType = node.kind === "more"
            ? "EllipsisOutlined"
            : getFileIconType(node.name, node.isDirectory);
        const iconClassName = node.isDirectory ? "is-directory" : "";
        return (
            <span
                className={`file-editor-tree-node ${node.kind === "more" ? "is-more" : ""}`}
                onDoubleClick={(event) => beginRename(event, node)}>
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
    }, [beginRename, cancelRename, editingNode, loadingPaths, onRename, tooltipsDisabled]);

    useEffect(() => () => {
        if (selectTimerRef.current !== undefined) {
            window.clearTimeout(selectTimerRef.current);
        }
    }, []);

    useEffect(() => {
        const selectedKey = selectedKeys[0];
        if (!selectedKey || treeData.length === 0) {
            if (treeData.length === 0) {
                locatedKeyRef.current = undefined;
            }
            return;
        }
        if (
            locatedKeyRef.current === selectedKey &&
            treeBodyRef.current?.querySelector(".ant-tree-node-selected")
        ) {
            return;
        }
        let locateFrame = 0;
        const renderFrame = window.requestAnimationFrame(() => {
            locateFrame = window.requestAnimationFrame(() => {
                const selectedNode = treeBodyRef.current
                    ?.querySelector<HTMLElement>(".ant-tree-node-selected");
                if (selectedNode) {
                    selectedNode.scrollIntoView({block: "center", inline: "nearest"});
                    locatedKeyRef.current = selectedKey;
                }
            });
        });
        return () => {
            window.cancelAnimationFrame(renderFrame);
            if (locateFrame) {
                window.cancelAnimationFrame(locateFrame);
            }
        };
    }, [expandedKeys, selectedKeys, treeData]);

    const select = useCallback((_: TreeSelectArgs[0], info: TreeSelectArgs[1]) => {
        if (editingNode) {
            return;
        }
        if (selectTimerRef.current !== undefined) {
            window.clearTimeout(selectTimerRef.current);
        }
        selectTimerRef.current = window.setTimeout(() => {
            selectTimerRef.current = undefined;
            onSelect(info.node);
        }, 600);
    }, [editingNode, onSelect]);

    const expand = useCallback((keys: TreeExpandArgs[0]) => {
        onExpand(keys.map(String));
    }, [onExpand]);
    const createEntry = useCallback(({key}: MenuClickArgs[0]) => {
        onCreate(key === "directory" ? "directory" : "file");
    }, [onCreate]);

    return (
        <aside className="file-editor-sidebar" aria-label="文件目录">
            <div className="file-editor-tree-toolbar">
                <div className="file-editor-tree-target">
                    <Icon type="FolderOpenOutlined" aria-hidden/>
                    <span title={tooltipsDisabled ? undefined : targetDirectory}>{directoryName}</span>
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
                    <Dropdown
                        trigger={["click"]}
                        placement="bottomRight"
                        disabled={disabled || !targetDirectory}
                        menu={{items: createItems, onClick: createEntry}}>
                        <button
                            className="file-editor-create-trigger"
                            type="button"
                            disabled={disabled || !targetDirectory}
                            aria-label="在当前目录新建"
                            aria-haspopup="menu">
                            <Icon className="marker" type="PlusOutlined"/>
                            <span>新建</span>
                            <Icon className="arrow" type="DownOutlined"/>
                        </button>
                    </Dropdown>
                </div>
            </div>
            <div ref={treeBodyRef} className="file-editor-tree-body">
                {treeData.length > 0 ? (
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
