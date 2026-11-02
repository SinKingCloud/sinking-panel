import React, {memo, useCallback, useMemo} from "react";
import {Button, Empty, Input, Pagination, Table, Tooltip, Typography} from "antd";
import type {InputRef, TableProps, TableRef} from "antd";
import {Icon} from "sinking-antd";
import {
    comparableFilePath as comparablePath,
    formatFileSize,
    formatFileTime,
    getFileIconType,
    joinFilePath,
} from "@/pages/file/utils";

type FilePickerBrowserProps = {
    className: string;
    compact: boolean;
    mode: string;
    path: string;
    loadedPath: string;
    selectedFile?: any;
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
    onSelectFile: (file: any) => void;
    onSelectionChange: (paths: string[]) => void;
    onConfirm: (file: any) => void;
    onPageChange: (page: number) => void;
    onReload: () => void;
};

const FilePickerBrowser = ({
    className,
    compact,
    mode,
    path,
    loadedPath,
    selectedFile,
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
    onSelectFile,
    onSelectionChange,
    onConfirm,
    onPageChange,
    onReload,
}: FilePickerBrowserProps) => {
    const selectedKeys = useMemo(() => new Set((multiple ? selectedPaths : selectedFile ? [selectedFile.path] : []).map(comparablePath)), [multiple, selectedFile, selectedPaths]);
    const availablePaths = useMemo(() => {
        const result = new Map(selectedPaths.map((item) => [comparablePath(item), item]));
        if (!showingDisks) items.filter((item) => !item.is_dir).forEach((item) => {
            const value = joinFilePath(path, item.name);
            result.set(comparablePath(value), value);
        });
        return result;
    }, [items, path, selectedPaths, showingDisks]);
    const columns = useMemo<TableProps<any>["columns"]>(() => [
        {
            title: "名称",
            dataIndex: "name",
            render: (name: string, record) => {
                const recordPath = showingDisks ? record.path : joinFilePath(path, name);
                const selected = !showingDisks && !record.is_dir
                    && selectedKeys.has(comparablePath(recordPath));
                return (
                    <div className="file-picker-name-cell">
                        <Icon type={getFileIconType(name, showingDisks || record.is_dir)} className="file-picker-icon"/>
                        <span className="file-picker-name" title={name}>{name}</span>
                        {!multiple && !showingDisks && !record.is_dir && mode === "file" ? (
                            <span className="file-picker-row-action">
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
    ], [compact, mode, multiple, path, selectedKeys, showingDisks]);

    const getRowClassName = useCallback((record: any) => {
        if (loading) {
            return "file-picker-passive";
        }
        const recordPath = showingDisks ? record.path : joinFilePath(path, record.name);
        const selected = !showingDisks && !record.is_dir
            && selectedKeys.has(comparablePath(recordPath));
        if (selected) {
            return "file-picker-selectable file-picker-selected";
        }
        if (showingDisks || record.is_dir) {
            return "file-picker-directory";
        }
        return mode === "file" ? "file-picker-selectable" : "file-picker-passive";
    }, [loading, mode, path, selectedKeys, showingDisks]);

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
        if (mode !== "file") {
            return {title: record.name};
        }
        const selected = selectedKeys.has(comparablePath(recordPath));
        const nextFile = {name: record.name, path: recordPath};
        return {
            role: "button",
            tabIndex: 0,
            "aria-label": `选择文件 ${record.name}`,
            "aria-pressed": selected,
            title: `选择 ${record.name}`,
            onClick: () => onSelectFile(nextFile),
            onDoubleClick: multiple ? undefined : () => onConfirm(nextFile),
            onKeyDown: (event) => {
                if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    onSelectFile(nextFile);
                }
            },
        };
    }, [loading, mode, multiple, onConfirm, onNavigate, onSelectFile, path, selectedKeys, showingDisks]);

    const emptyText = error ? (
        <div className="file-picker-error">
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={error}/>
            <Button size="small" onClick={onReload}>重试</Button>
        </div>
    ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>;

    const statusPath = showingDisks ? undefined : mode === "directory" ? path : selectedFile?.path;
    const statusLabel = multiple ? "选择结果" : showingDisks ? "本地存储" : mode === "directory" ? "当前目录" : "选择结果";
    const statusText = multiple ? `已选择 ${selectedPaths.length} 个文件` : showingDisks ? "请选择磁盘或输入路径" : statusPath || (mode === "directory" ? "正在读取磁盘..." : "未选择");
    const statusIcon = showingDisks || mode === "directory"
        ? "FolderOpenOutlined"
        : selectedFile ? getFileIconType(selectedFile.name) : "FileOutlined";
    const rangeStart = total > 0 ? (page - 1) * pageSize + 1 : 0;
    const rangeEnd = Math.min(page * pageSize, total);
    const countText = loading && !loadedPath
        ? "加载中"
        : !showingDisks && total > pageSize ? `${rangeStart}-${rangeEnd} / ${total} 项` : `${total} 项`;

    return (
        <div className={className}>
            <div className="file-picker-toolbar">
                <Tooltip title="上一级">
                    <Button
                        type="text"
                        className="file-picker-up"
                        disabled={parentPath === undefined || loading}
                        aria-label="返回上一级"
                        icon={<Icon type="ArrowLeftOutlined"/>}
                        onClick={() => parentPath !== undefined && onNavigate(parentPath)}/>
                </Tooltip>
                {editingPath ? (
                    <Input
                        ref={pathInputRef}
                        className="file-picker-path-input"
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
                    <ol ref={pathRef} className="file-picker-path" aria-label={showingDisks ? "当前路径：/，磁盘列表" : `当前路径：${path || "正在读取"}`}>
                        {pathBreadcrumbs.length > 0 ? pathBreadcrumbs.map((item, index) => {
                            const current = index === pathBreadcrumbs.length - 1;
                            return (
                                <li key={item.path}>
                                    {index > 0 && (
                                        <Icon
                                            type="RightOutlined"
                                            className="file-picker-path-separator"
                                            aria-hidden="true"/>
                                    )}
                                    {current ? (
                                        <button
                                            ref={currentPathRef}
                                            className="file-picker-path-current"
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
                                            className="file-picker-path-button"
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
                            <li><span className="file-picker-path-current">正在读取磁盘...</span></li>
                        )}
                    </ol>
                )}
            </div>

            <Table<any>
                ref={tableRef}
                className="file-picker-table"
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
                    selectedRowKeys: [...selectedKeys],
                    preserveSelectedRowKeys: true,
                    getCheckboxProps: (record) => ({disabled: loading || showingDisks || record.is_dir, "aria-label": `选择文件 ${record.name}`}),
                    onChange: (keys) => onSelectionChange(keys.flatMap((key) => {
                        const selected = availablePaths.get(String(key));
                        return selected ? [selected] : [];
                    })),
                    onCell: () => ({onClick: (event) => event.stopPropagation(), onKeyDown: (event) => event.stopPropagation()}),
                } : undefined}
                rowClassName={getRowClassName}
                onRow={getRowProps}/>

            <div className="file-picker-status">
                <div className="file-picker-status-main">
                    <Icon type={statusIcon} className="file-picker-status-icon"/>
                    <span className="file-picker-status-label">{statusLabel}</span>
                    <Typography.Text
                        className="file-picker-status-value"
                        ellipsis={{tooltip: statusPath}}>
                        {statusText}
                    </Typography.Text>
                    {multiple && selectedPaths.length > 0 && <Button size="small" type="text" onClick={() => onSelectionChange([])}>清空</Button>}
                </div>
                <div className="file-picker-status-meta">
                    <span className="file-picker-count">{countText}</span>
                    {!showingDisks && total > pageSize && (
                        <Pagination
                            className="file-picker-pagination"
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

export default memo(FilePickerBrowser);
