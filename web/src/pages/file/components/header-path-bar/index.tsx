import React, {useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {Button, Input} from "antd";
import type {InputRef, MenuProps} from "antd";
import {Icon} from "sinking-antd";
import {TableAction} from "@/pages/components/table";
import {buildFileBreadcrumbs, normalizeFilePath, parentFilePath} from "../../utils";
import type {HeaderStyles} from "../header/types";

interface HeaderPathBarProps {
    path: string;
    disks: string[];
    clipboardCount: number;
    clipboardMode: "copy" | "move";
    pasting: boolean;
    pasteDisabled: boolean;
    directoryActionsDisabled: boolean;
    selectedCount: number;
    editorMinimized: boolean;
    selectionClipboardDisabled: boolean;
    selectionOperationDisabled: boolean;
    selectionClearDisabled: boolean;
    onNavigate: (path: string) => void;
    onPaste: () => void;
    onCopySelected: () => void;
    onMoveSelected: () => void;
    onCompressSelected: () => void;
    onDeleteSelected: () => void;
    onClearSelection: () => void;
    onRestoreEditor: () => void;
    styles: Pick<
        HeaderStyles,
        "pathSection" | "pathBar"
    >;
}

const HeaderPathBar = ({
    path,
    disks,
    clipboardCount,
    clipboardMode,
    pasting,
    pasteDisabled,
    directoryActionsDisabled,
    selectedCount,
    editorMinimized,
    selectionClipboardDisabled,
    selectionOperationDisabled,
    selectionClearDisabled,
    onNavigate,
    onPaste,
    onCopySelected,
    onMoveSelected,
    onCompressSelected,
    onDeleteSelected,
    onClearSelection,
    onRestoreEditor,
    styles,
}: HeaderPathBarProps) => {
    const breadcrumbRef = useRef<HTMLOListElement>(null);
    const pathInputRef = useRef<InputRef>(null);
    const currentPathRef = useRef<HTMLButtonElement>(null);
    const [editingPath, setEditingPath] = useState(false);
    const [pathDraft, setPathDraft] = useState(path);
    const normalizedPath = normalizeFilePath(path);
    const rootPath = useMemo(() => {
        return [...disks]
            .map((item) => normalizeFilePath(item))
            .sort((left, right) => right.length - left.length)
            .find((item) => normalizedPath === item || normalizedPath.startsWith(`${item.replace(/\/$/, "")}/`)) || "/";
    }, [disks, normalizedPath]);
    const breadcrumbs = useMemo(() => {
        const all = buildFileBreadcrumbs(path);
        const start = all.findIndex((item) => normalizeFilePath(item.path) === rootPath);
        if (start <= 0) {
            return all;
        }
        return [{label: rootPath, path: rootPath}, ...all.slice(start + 1)];
    }, [path, rootPath]);
    const parentPath = normalizedPath === rootPath ? rootPath : parentFilePath(path);
    const atRoot = normalizedPath === rootPath;
    const batchMenuItems = useMemo<MenuProps["items"]>(() => [
        {
            key: "copy",
            label: "复制",
            icon: <Icon type="CopyOutlined"/>,
            disabled: selectionClipboardDisabled,
        },
        {
            key: "move",
            label: "移动",
            icon: <Icon type="ScissorOutlined"/>,
            disabled: selectionClipboardDisabled,
        },
        {
            key: "compress",
            label: "压缩",
            icon: <Icon type="FileZipOutlined"/>,
            disabled: selectionOperationDisabled,
        },
        {type: "divider"},
        {
            key: "delete",
            label: "删除",
            icon: <Icon type="DeleteOutlined"/>,
            danger: true,
            disabled: selectionOperationDisabled,
        },
        {
            key: "clear",
            label: "取消选择",
            icon: <Icon type="CloseOutlined"/>,
            disabled: selectionClearDisabled,
        },
    ], [selectionClearDisabled, selectionClipboardDisabled, selectionOperationDisabled]);
    const handleBatchAction = useCallback<NonNullable<MenuProps["onClick"]>>(({key}) => {
        if (key === "copy") {
            onCopySelected();
        } else if (key === "move") {
            onMoveSelected();
        } else if (key === "compress") {
            onCompressSelected();
        } else if (key === "delete") {
            onDeleteSelected();
        } else if (key === "clear") {
            onClearSelection();
        }
    }, [onClearSelection, onCompressSelected, onCopySelected, onDeleteSelected, onMoveSelected]);

    useEffect(() => {
        setEditingPath(false);
        setPathDraft(normalizedPath);
    }, [normalizedPath]);

    useLayoutEffect(() => {
        if (!editingPath) {
            return;
        }
        pathInputRef.current?.focus();
        pathInputRef.current?.select();
    }, [editingPath]);

    useLayoutEffect(() => {
        if (editingPath) {
            return;
        }
        if (breadcrumbRef.current) {
            breadcrumbRef.current.scrollLeft = breadcrumbRef.current.scrollWidth;
        }
    }, [breadcrumbs, editingPath]);

    const openPathEditor = useCallback(() => {
        if (directoryActionsDisabled) {
            return;
        }
        setPathDraft(normalizedPath);
        setEditingPath(true);
    }, [directoryActionsDisabled, normalizedPath]);

    const cancelPathEditor = useCallback((restoreFocus = false) => {
        setPathDraft(normalizedPath);
        setEditingPath(false);
        if (restoreFocus) {
            window.requestAnimationFrame(() => currentPathRef.current?.focus());
        }
    }, [normalizedPath]);

    const submitPathEditor = useCallback(() => {
        const value = pathDraft.trim();
        if (!value) {
            cancelPathEditor(true);
            return;
        }
        const nextPath = normalizeFilePath(value);
        if (nextPath === normalizedPath) {
            cancelPathEditor(true);
            return;
        }
        setEditingPath(false);
        onNavigate(nextPath);
    }, [cancelPathEditor, normalizedPath, onNavigate, pathDraft]);

    useEffect(() => {
        if (directoryActionsDisabled && editingPath) {
            cancelPathEditor();
        }
    }, [cancelPathEditor, directoryActionsDisabled, editingPath]);

    return (
        <div className={styles.pathSection}>
            <nav className={styles.pathBar} aria-label={`当前路径：${normalizedPath}`}>
                <Button
                    className="path-back"
                    type="text"
                    disabled={atRoot}
                    aria-label="返回上一级目录"
                    icon={<Icon type="LeftOutlined"/>}
                    onClick={() => onNavigate(parentPath)}/>
                {editingPath ? (
                    <Input
                        ref={pathInputRef}
                        className="path-input"
                        value={pathDraft}
                        aria-label="编辑当前路径"
                        onChange={(event) => setPathDraft(event.target.value)}
                        onBlur={() => cancelPathEditor()}
                        onKeyDown={(event) => {
                            if (event.nativeEvent.isComposing || event.nativeEvent.keyCode === 229) {
                                return;
                            }
                            if (event.key === "Enter") {
                                event.preventDefault();
                                submitPathEditor();
                            } else if (event.key === "Escape") {
                                event.preventDefault();
                                cancelPathEditor(true);
                            }
                        }}/>
                ) : (
                    <ol ref={breadcrumbRef} className="path-breadcrumb">
                        {breadcrumbs.map((item, index) => {
                            const current = index === breadcrumbs.length - 1;
                            return (
                                <li key={item.path}>
                                    {index > 0 && (
                                        <Icon
                                            type="RightOutlined"
                                            className="path-separator"
                                            aria-hidden="true"/>
                                    )}
                                    {current ? (
                                        <button
                                            ref={currentPathRef}
                                            className="path-current"
                                            type="button"
                                            disabled={directoryActionsDisabled}
                                            aria-current="page"
                                            aria-label={`编辑路径 ${item.path}`}
                                            title="点击编辑路径"
                                            onClick={openPathEditor}>
                                            <span>{item.label}</span>
                                        </button>
                                    ) : (
                                        <button
                                            className="path-segment"
                                            type="button"
                                            title={item.path}
                                            aria-label={`前往 ${item.path}`}
                                            onClick={() => onNavigate(item.path)}>
                                            <span>{item.label}</span>
                                        </button>
                                    )}
                                </li>
                            );
                        })}
                    </ol>
                )}
                {(editorMinimized || selectedCount > 0 || clipboardCount > 0) && (
                    <div className="path-actions" role="group" aria-label="当前目录操作">
                        {editorMinimized && (
                            <TableAction
                                className="path-editor"
                                label="编辑器"
                                icon="EditOutlined"
                                aria-label="恢复文件编辑器"
                                onClick={onRestoreEditor}/>
                        )}
                        {selectedCount > 0 && (
                            <TableAction
                                className="path-batch"
                                label={`已选 ${selectedCount} 项`}
                                icon="CheckSquareOutlined"
                                suffixIcon="DownOutlined"
                                aria-label={`批量操作，已选 ${selectedCount} 项`}
                                menu={{items: batchMenuItems, onClick: handleBatchAction}}
                                menuClassName="batch-dropdown"/>
                        )}
                        {clipboardCount > 0 && (
                            <TableAction
                                className="path-paste"
                                label={`${clipboardMode === "move" ? "移动" : "粘贴"} ${clipboardCount} 项`}
                                aria-label={clipboardMode === "move"
                                    ? `将 ${clipboardCount} 项移动到当前目录`
                                    : `粘贴 ${clipboardCount} 项`}
                                disabled={pasteDisabled}
                                loading={pasting}
                                icon={clipboardMode === "move" ? "ScissorOutlined" : "SnippetsOutlined"}
                                onClick={onPaste}/>
                        )}
                    </div>
                )}
            </nav>
        </div>
    );
};

export default React.memo(HeaderPathBar);
