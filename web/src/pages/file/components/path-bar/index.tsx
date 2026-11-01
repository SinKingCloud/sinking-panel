import React, {useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {Button, Input} from "antd";
import type {InputRef} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {getFileNavigation, normalizeFilePath} from "../../utils";
import useStyles from "./styles";

const FilePathBar = ({
    path,
    roots,
    showingDisks,
    disabled,
    onNavigate,
    onShowDisks,
}: {
    path: string;
    roots: string[];
    showingDisks: boolean;
    disabled: boolean;
    onNavigate: (path: string) => void;
    onShowDisks: () => void;
}) => {
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const breadcrumbRef = useRef<HTMLOListElement>(null);
    const pathInputRef = useRef<InputRef>(null);
    const currentPathRef = useRef<HTMLButtonElement>(null);
    const [editingPath, setEditingPath] = useState(false);
    const [pathDraft, setPathDraft] = useState(path);
    const navigationDisabled = disabled || !onNavigate;
    const normalizedPath = normalizeFilePath(path);
    const {breadcrumbs, parentPath} = useMemo(
        () => getFileNavigation(showingDisks ? "" : path, roots),
        [path, roots, showingDisks],
    );

    useEffect(() => {
        setEditingPath(false);
        setPathDraft(normalizedPath);
    }, [normalizedPath, showingDisks]);

    useLayoutEffect(() => {
        if (!editingPath) {
            return;
        }
        pathInputRef.current?.focus();
        pathInputRef.current?.select();
    }, [editingPath]);

    useLayoutEffect(() => {
        const breadcrumb = breadcrumbRef.current;
        if (editingPath || !breadcrumb) return;
        const revealCurrentPath = () => { breadcrumb.scrollLeft = breadcrumb.scrollWidth; };
        revealCurrentPath();
        if (typeof ResizeObserver === "undefined") return;
        const observer = new ResizeObserver(revealCurrentPath);
        observer.observe(breadcrumb);
        return () => observer.disconnect();
    }, [breadcrumbs, editingPath]);

    const openPathEditor = useCallback(() => {
        if (navigationDisabled) {
            return;
        }
        setPathDraft(normalizedPath);
        setEditingPath(true);
    }, [navigationDisabled, normalizedPath]);

    const cancelPathEditor = useCallback((restoreFocus = false) => {
        setPathDraft(normalizedPath);
        setEditingPath(false);
        if (restoreFocus) {
            window.requestAnimationFrame(() => currentPathRef.current?.focus());
        }
    }, [normalizedPath]);

    const submitPathEditor = useCallback(() => {
        if (navigationDisabled) {
            cancelPathEditor();
            return;
        }
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
        onNavigate?.(nextPath);
    }, [cancelPathEditor, navigationDisabled, normalizedPath, onNavigate, pathDraft]);

    useEffect(() => {
        if (navigationDisabled && editingPath) {
            cancelPathEditor();
        }
    }, [cancelPathEditor, navigationDisabled, editingPath]);

    return (
        <nav className={styles.pathBar} aria-label={showingDisks ? "当前路径：/，磁盘列表" : `当前路径：${normalizedPath}`}>
            <Button
                className="path-back"
                type="text"
                disabled={parentPath === undefined || navigationDisabled}
                aria-label="返回上一级目录"
                icon={<Icon type="LeftOutlined"/>}
                onClick={() => parentPath ? onNavigate(parentPath) : onShowDisks()}/>
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
                                        disabled={navigationDisabled || showingDisks}
                                        aria-current="page"
                                        aria-label={showingDisks ? "磁盘列表" : `编辑路径 ${item.path}`}
                                        title={showingDisks ? "磁盘列表" : "点击编辑路径"}
                                        onClick={openPathEditor}>
                                        <span>{item.label}</span>
                                    </button>
                                ) : (
                                    <button
                                        className="path-segment"
                                        type="button"
                                        disabled={navigationDisabled}
                                        title={item.path || "磁盘列表"}
                                        aria-label={item.path ? `前往 ${item.path}` : "返回磁盘列表"}
                                        onClick={() => item.path ? onNavigate(item.path) : onShowDisks()}>
                                        <span>{item.label}</span>
                                    </button>
                                )}
                            </li>
                        );
                    })}
                </ol>
            )}
        </nav>
    );
};

export default React.memo(FilePathBar);
