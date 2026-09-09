import React, {useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {Button, Input} from "antd";
import type {InputRef} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {buildFileBreadcrumbs, normalizeFilePath, parentFilePath} from "../../utils";
import useStyles from "./styles";

export interface FilePathBarProps {
    path: string;
    roots?: string[];
    disabled?: boolean;
    onNavigate?: (path: string) => void;
}

const FilePathBar = ({
    path,
    roots,
    disabled,
    onNavigate,
}: FilePathBarProps) => {
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const breadcrumbRef = useRef<HTMLOListElement>(null);
    const pathInputRef = useRef<InputRef>(null);
    const currentPathRef = useRef<HTMLButtonElement>(null);
    const [editingPath, setEditingPath] = useState(false);
    const [pathDraft, setPathDraft] = useState(path);
    const navigationDisabled = disabled || !onNavigate;
    const normalizedPath = normalizeFilePath(path);
    const rootPath = useMemo(() => {
        return [...(roots || [])]
            .map((item) => normalizeFilePath(item))
            .sort((left, right) => right.length - left.length)
            .find((item) => normalizedPath === item || normalizedPath.startsWith(`${item.replace(/\/$/, "")}/`))
            || buildFileBreadcrumbs(normalizedPath)[0]?.path || "/";
    }, [roots, normalizedPath]);
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
        <nav className={styles.pathBar} aria-label={`当前路径：${normalizedPath}`}>
            <Button
                className="path-back"
                type="text"
                disabled={atRoot || navigationDisabled}
                aria-label="返回上一级目录"
                icon={<Icon type="LeftOutlined"/>}
                onClick={() => onNavigate?.(parentPath)}/>
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
                                        disabled={navigationDisabled}
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
                                        disabled={navigationDisabled}
                                        title={item.path}
                                        aria-label={`前往 ${item.path}`}
                                        onClick={() => onNavigate?.(item.path)}>
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
