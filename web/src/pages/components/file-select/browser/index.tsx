import React, {memo, useCallback, useMemo} from "react";
import {Button, Checkbox, Empty, Input, Pagination, Table, Tooltip, Typography} from "antd";
import type {InputRef, TableProps, TableRef} from "antd";
import {Icon} from "sinking-antd";
import {
    buildFileBreadcrumbs,
    comparableFilePath as comparablePath,
    formatFileSize,
    formatFileTime,
    getFileIconType,
    isAbsoluteFilePath,
    joinFilePath,
} from "@/pages/file/utils";

type FileSelectBrowserProps = {
    className: string;
    compact: boolean;
    type?: "file" | "directory";
    path: string;
    loadedPath: string;
    multiple: boolean;
    selectedPaths: string[];
    showingDisks: boolean;
    pathBreadcrumbs: any[];
    parentPath?: string;
    editingPath: boolean;
    pathDraft: string;
    loading: boolean;
    error: string;
    items: any[];
    page: number;
    pageSize: number;
    total: number;
    pathRef: React.RefObject<HTMLOListElement | null>;
    pathInputRef: React.RefObject<InputRef | null>;
    currentPathRef: React.RefObject<HTMLButtonElement | null>;
    tableRef: React.RefObject<TableRef | null>;
    onNavigate: (path: string) => void;
    onOpenPathEditor: () => void;
    onCancelPathEditor: (restoreFocus?: boolean) => void;
    onSubmitPathEditor: () => void;
    onPathDraftChange: (path: string) => void;
    onSelectionChange: (paths: string[], checked: boolean) => void;
    onConfirm: (path: string) => void;
    onPageChange: (page: number) => void;
    onCreateDirectory: () => void;
    onReload: () => void;
};

const FileSelectBrowser = ({
    className,
    compact,
    type,
    path,
    loadedPath,
    multiple,
    selectedPaths,
    showingDisks,
    pathBreadcrumbs,
    parentPath,
    editingPath,
    pathDraft,
    loading,
    error,
    items,
    page,
    pageSize,
    total,
    pathRef,
    pathInputRef,
    currentPathRef,
    tableRef,
    onNavigate,
    onOpenPathEditor,
    onCancelPathEditor,
    onSubmitPathEditor,
    onPathDraftChange,
    onSelectionChange,
    onConfirm,
    onPageChange,
    onCreateDirectory,
    onReload,
}: FileSelectBrowserProps) => {
    const allowFiles = type !== "directory";
    const allowDirectories = type !== "file";
    const selectedKeys = useMemo(() => new Set(selectedPaths.filter(isAbsoluteFilePath).map(comparablePath)), [selectedPaths]);
    const coveredKeys = useMemo(() => {
        const result = new Set<string>();
        if (multiple && allowDirectories) {
            for (const item of items) {
                const value = showingDisks ? item.path : joinFilePath(path, item.name);
                if (!allowFiles && !showingDisks && !item.is_dir) continue;
                if (buildFileBreadcrumbs(value).slice(0, -1).some((parent) => selectedKeys.has(comparablePath(parent.path)))) {
                    result.add(comparablePath(value));
                }
            }
        }
        return result;
    }, [allowDirectories, allowFiles, items, multiple, path, selectedKeys, showingDisks]);
    const checkedKeys = useMemo(() => new Set([...selectedKeys, ...coveredKeys]), [coveredKeys, selectedKeys]);
    const partialKeys = useMemo(() => {
        const result = new Set<string>();
        if (multiple && allowDirectories) {
            for (const selected of selectedKeys) {
                buildFileBreadcrumbs(selected).slice(0, -1).forEach((parent) => result.add(comparablePath(parent.path)));
            }
        }
        return result;
    }, [allowDirectories, multiple, selectedKeys]);
    const selectablePaths = useMemo(() => items
        .filter((item) => showingDisks || item.is_dir ? allowDirectories : allowFiles)
        .map((item) => showingDisks ? item.path : joinFilePath(path, item.name)), [allowDirectories, allowFiles, items, path, showingDisks]);
    const allChecked = selectablePaths.length > 0 && selectablePaths.every((value) => checkedKeys.has(comparablePath(value)));
    const someChecked = selectablePaths.some((value) => checkedKeys.has(comparablePath(value)) || partialKeys.has(comparablePath(value)));
    const columns = useMemo<TableProps<any>["columns"]>(() => [
        {
            title: "名称",
            dataIndex: "name",
            render: (name: string, record) => {
                const recordPath = showingDisks ? record.path : joinFilePath(path, name);
                const selected = ((showingDisks || record.is_dir) ? allowDirectories : allowFiles)
                    && selectedKeys.has(comparablePath(recordPath));
                return (
                    <div className="path-select-name-cell">
                        <Icon type={getFileIconType(name, showingDisks || record.is_dir)} className="path-select-icon"/>
                        <span className="path-select-name" title={name}>{name}</span>
                        {!multiple && ((showingDisks || record.is_dir) ? allowDirectories : allowFiles) ? (
                            <span className="path-select-row-action">
                                {selected && <Icon type="CheckOutlined"/>}
                            </span>
                        ) : null}
                    </div>
                );
            },
        },
        ...(showingDisks ? [{
            title: "路径",
            dataIndex: "path",
            ellipsis: true,
        }] : [{
            title: "大小",
            dataIndex: "size",
            width: compact ? 80 : 96,
            render: (size: number, record) => record.is_dir ? "-" : formatFileSize(size),
        },
        {
            title: "修改时间",
            dataIndex: "update_time",
            width: compact ? 132 : 154,
            render: formatFileTime,
        }]),
    ], [allowDirectories, allowFiles, compact, multiple, path, selectedKeys, showingDisks]);

    const getRowClassName = useCallback((record: any) => {
        if (loading) {
            return "path-select-passive";
        }
        const recordPath = showingDisks ? record.path : joinFilePath(path, record.name);
        const selected = ((showingDisks || record.is_dir) ? allowDirectories : allowFiles)
            && checkedKeys.has(comparablePath(recordPath));
        if (selected) {
            return "path-select-selectable path-select-selected";
        }
        if (showingDisks || record.is_dir) {
            return "path-select-directory";
        }
        return allowFiles ? "path-select-selectable" : "path-select-passive";
    }, [allowDirectories, allowFiles, checkedKeys, loading, path, showingDisks]);

    const getRowProps = useCallback<NonNullable<TableProps<any>["onRow"]>>((record) => {
        if (loading) {
            return {"aria-disabled": true};
        }
        const recordPath = showingDisks ? record.path : joinFilePath(path, record.name);
        if (showingDisks || record.is_dir) {
            return {
                role: "button",
                tabIndex: 0,
                "aria-label": showingDisks ? `打开磁盘 ${record.path}` : `打开目录 ${record.name}`,
                title: `进入 ${record.name}`,
                onClick: () => onNavigate(recordPath),
                onKeyDown: (event) => {
                    if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault();
                        onNavigate(recordPath);
                    }
                },
            };
        }
        if (!allowFiles) {
            return {title: record.name};
        }
        const selected = checkedKeys.has(comparablePath(recordPath));
        return {
            role: "button",
            tabIndex: 0,
            "aria-label": `选择文件 ${record.name}`,
            "aria-pressed": selected,
            title: `选择 ${record.name}`,
            onClick: () => onSelectionChange([recordPath], !multiple || !selected),
            onDoubleClick: multiple ? undefined : () => onConfirm(recordPath),
            onKeyDown: (event) => {
                if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    onSelectionChange([recordPath], !multiple || !selected);
                }
            },
        };
    }, [allowFiles, checkedKeys, loading, multiple, onConfirm, onNavigate, onSelectionChange, path, showingDisks]);

    const emptyText = error ? (
        <div className="path-select-error">
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={error}/>
            <Button size="small" onClick={onReload}>重试</Button>
        </div>
    ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>;

    const selectedPath = selectedPaths[0];
    const selectingDirectory = allowDirectories && (!selectedPath || type === "directory" || comparablePath(selectedPath) === comparablePath(path));
    const statusPath = showingDisks ? undefined : selectedPath || (selectingDirectory ? path : undefined);
    const statusLabel = multiple ? "选择结果" : showingDisks ? "本地存储" : selectingDirectory ? "当前目录" : "选择结果";
    const statusText = multiple ? `已选择 ${selectedPaths.length} ${type === "file" ? "个文件" : type === "directory" ? "个目录" : "项"}` : showingDisks ? "请选择磁盘或输入路径" : statusPath || (selectingDirectory ? "正在读取磁盘..." : "未选择");
    const statusIcon = showingDisks || selectingDirectory
        ? "FolderOpenOutlined"
        : selectedPath ? getFileIconType(selectedPath) : "FileOutlined";
    const rangeStart = total > 0 ? (page - 1) * pageSize + 1 : 0;
    const rangeEnd = Math.min(page * pageSize, total);
    const countText = loading && !loadedPath
        ? "加载中"
        : !showingDisks && total > pageSize ? `${rangeStart}-${rangeEnd} / ${total} 项` : `${total} 项`;

    return (
        <div className={className}>
            <div className="path-select-toolbar">
                <Tooltip title="上一级">
                    <Button
                        type="text"
                        className="path-select-up"
                        disabled={parentPath === undefined || loading}
                        aria-label="返回上一级"
                        icon={<Icon type="ArrowLeftOutlined"/>}
                        onClick={() => parentPath !== undefined && onNavigate(parentPath)}/>
                </Tooltip>
                {editingPath ? (
                    <Input
                        ref={pathInputRef}
                        className="path-select-path-input"
                        value={pathDraft}
                        placeholder="请输入绝对路径"
                        aria-label="编辑当前路径"
                        onChange={(event) => onPathDraftChange(event.target.value)}
                        onBlur={() => onCancelPathEditor()}
                        onKeyDown={(event) => {
                            if (event.nativeEvent.isComposing || event.nativeEvent.keyCode === 229) {
                                return;
                            }
                            if (event.key === "Enter") {
                                event.preventDefault();
                                onSubmitPathEditor();
                            } else if (event.key === "Escape") {
                                event.preventDefault();
                                event.stopPropagation();
                                onCancelPathEditor(true);
                            }
                        }}/>
                ) : (
                    <ol ref={pathRef} className="path-select-path" aria-label={showingDisks ? "当前路径：/，磁盘列表" : `当前路径：${path || "正在读取"}`}>
                        {pathBreadcrumbs.length > 0 ? pathBreadcrumbs.map((item, index) => {
                            const current = index === pathBreadcrumbs.length - 1;
                            return (
                                <li key={item.path}>
                                    {index > 0 && (
                                        <Icon
                                            type="RightOutlined"
                                            className="path-select-path-separator"
                                            aria-hidden="true"/>
                                    )}
                                    {current ? (
                                        <button
                                            ref={currentPathRef}
                                            className="path-select-path-current"
                                            type="button"
                                            title={showingDisks ? "输入路径" : "点击编辑路径"}
                                            disabled={loading}
                                            aria-current="page"
                                            aria-label={showingDisks ? "输入路径" : `编辑路径 ${item.path}`}
                                            onClick={onOpenPathEditor}>
                                            <span>{item.label}</span>
                                        </button>
                                    ) : (
                                        <button
                                            className="path-select-path-button"
                                            type="button"
                                            title={item.path || "磁盘列表"}
                                            disabled={loading}
                                            aria-label={item.path ? `前往 ${item.path}` : "返回磁盘列表"}
                                            onClick={() => onNavigate(item.path)}>
                                            <span>{item.label}</span>
                                        </button>
                                    )}
                                </li>
                            );
                        }) : (
                            <li><span className="path-select-path-current">正在读取磁盘...</span></li>
                        )}
                    </ol>
                )}
                <div className="path-select-actions">
                    <Button
                        type="text"
                        className="path-select-up"
                        disabled={showingDisks || loading || Boolean(error) || !loadedPath || comparablePath(loadedPath) !== comparablePath(path)}
                        aria-label="新建文件夹"
                        icon={<Icon type="PlusOutlined"/>}
                        onClick={onCreateDirectory}/>
                    <Button
                        type="text"
                        className="path-select-up"
                        loading={loading}
                        aria-label="刷新"
                        icon={<Icon type="ReloadOutlined"/>}
                        onClick={onReload}/>
                </div>
            </div>

            <Table<any>
                ref={tableRef}
                className="path-select-table"
                rowKey={(record) => comparablePath(showingDisks ? record.path : joinFilePath(path, record.name))}
                size="small"
                rowHoverable={false}
                tableLayout="auto"
                loading={loading}
                columns={columns}
                dataSource={items}
                locale={{emptyText}}
                scroll={{
                    x: "max-content",
                    y: compact
                        ? "clamp(180px, 34dvh, 238px)"
                        : "clamp(200px, 38dvh, 278px)",
                }}
                pagination={false}
                rowSelection={multiple ? {
                    type: "checkbox",
                    selectedRowKeys: [...checkedKeys],
                    preserveSelectedRowKeys: true,
                    columnTitle: <Checkbox aria-label="选择当前页" checked={allChecked}
                        indeterminate={!allChecked && someChecked} disabled={loading || !selectablePaths.length}
                        onChange={(event) => onSelectionChange(selectablePaths, event.target.checked)}/>,
                    getCheckboxProps: (record) => ({
                        disabled: loading || ((showingDisks || record.is_dir) ? !allowDirectories : !allowFiles),
                        indeterminate: !checkedKeys.has(comparablePath(showingDisks ? record.path : joinFilePath(path, record.name)))
                            && partialKeys.has(comparablePath(showingDisks ? record.path : joinFilePath(path, record.name))),
                        "aria-label": `选择${record.is_dir ? "目录" : "文件"} ${record.name}`,
                    }),
                    onSelect: (record, checked) => onSelectionChange([showingDisks ? record.path : joinFilePath(path, record.name)], checked),
                    onSelectMultiple: (checked, _, records) => onSelectionChange(records.map((record) => showingDisks ? record.path : joinFilePath(path, record.name)), checked),
                    onCell: () => ({onClick: (event) => event.stopPropagation(), onKeyDown: (event) => event.stopPropagation()}),
                } : undefined}
                rowClassName={getRowClassName}
                onRow={getRowProps}/>

            <div className="path-select-status">
                <div className="path-select-status-main">
                    <Icon type={statusIcon} className="path-select-status-icon"/>
                    <span className="path-select-status-label">{statusLabel}</span>
                    <Typography.Text
                        className="path-select-status-value"
                        ellipsis={{tooltip: statusPath}}>
                        {statusText}
                    </Typography.Text>
                </div>
                <div className="path-select-status-meta">
                    <span className="path-select-count">{countText}</span>
                    {!showingDisks && total > pageSize && (
                        <Pagination
                            className="path-select-pagination"
                            size="small"
                            simple={{readOnly: true}}
                            current={page}
                            pageSize={pageSize}
                            total={total}
                            disabled={loading}
                            showSizeChanger={false}
                            showTitle={false}
                            onChange={onPageChange}/>
                    )}
                </div>
            </div>
        </div>
    );
};

export default memo(FileSelectBrowser);
