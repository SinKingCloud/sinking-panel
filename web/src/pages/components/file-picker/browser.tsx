import React, {memo, useCallback, useMemo} from "react";
import {Button, Empty, Input, Pagination, Select, Table, Tooltip, Typography} from "antd";
import type {InputRef, TableProps, TableRef} from "antd";
import {Icon} from "sinking-antd";
import {
    comparableFilePath as comparablePath,
    formatFileSize,
    formatFileTime,
    getFileIconType,
    isFilePathWithin as isPathWithin,
    joinFilePath,
} from "@/pages/file/utils";
import type {FileBreadcrumbItem} from "@/pages/file/utils";

interface FilePickerBrowserProps {
    className: string;
    compact: boolean;
    mode: string;
    path: string;
    loadedPath: string;
    selectedFile?: any;
    disks: string[];
    selectedDisk?: string;
    pathBreadcrumbs: FileBreadcrumbItem[];
    parentPath: string;
    diskRoot: string;
    canGoUp: boolean;
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
    onConfirm: (file: any) => void;
    onPageChange: (page: number) => void;
    onReload: () => void;
}

const FilePickerBrowser = ({
    className,
    compact,
    mode,
    path,
    loadedPath,
    selectedFile,
    disks,
    selectedDisk,
    pathBreadcrumbs,
    parentPath,
    diskRoot,
    canGoUp,
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
    onConfirm,
    onPageChange,
    onReload,
}: FilePickerBrowserProps) => {
    const columns = useMemo<TableProps<any>["columns"]>(() => [
        {
            title: "名称",
            dataIndex: "name",
            render: (name: string, record) => {
                const recordPath = joinFilePath(path, name);
                const selected = !record.is_dir
                    && comparablePath(selectedFile?.path || "") === comparablePath(recordPath);
                return (
                    <div className="file-picker-name-cell">
                        <Icon type={getFileIconType(name, record.is_dir)} className="file-picker-icon"/>
                        <span className="file-picker-name" title={name}>{name}</span>
                        {!record.is_dir && mode === "file" ? (
                            <span className="file-picker-row-action">
                                {selected && <Icon type="CheckOutlined"/>}
                            </span>
                        ) : null}
                    </div>
                );
            },
        },
        {
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
        },
    ], [compact, mode, path, selectedFile?.path]);

    const diskOptions = useMemo(
        () => disks.map((disk) => ({label: disk, value: disk})),
        [disks],
    );

    const getRowClassName = useCallback((record: any) => {
        if (loading) {
            return "file-picker-passive";
        }
        const recordPath = joinFilePath(path, record.name);
        const selected = !record.is_dir
            && comparablePath(selectedFile?.path || "") === comparablePath(recordPath);
        if (selected) {
            return "file-picker-selectable file-picker-selected";
        }
        if (record.is_dir) {
            return "file-picker-directory";
        }
        return mode === "file" ? "file-picker-selectable" : "file-picker-passive";
    }, [loading, mode, path, selectedFile?.path]);

    const getRowProps = useCallback<NonNullable<TableProps<any>["onRow"]>>((record) => {
        if (loading) {
            return {"aria-disabled": true};
        }
        const recordPath = joinFilePath(path, record.name);
        if (record.is_dir) {
            return {
                role: "button",
                tabIndex: 0,
                "aria-label": `打开目录 ${record.name}`,
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
        const selected = comparablePath(selectedFile?.path || "") === comparablePath(recordPath);
        const nextFile = {name: record.name, path: recordPath};
        return {
            role: "button",
            tabIndex: 0,
            "aria-label": `选择文件 ${record.name}`,
            "aria-pressed": selected,
            title: `选择 ${record.name}`,
            onClick: () => onSelectFile(nextFile),
            onDoubleClick: () => onConfirm(nextFile),
            onKeyDown: (event) => {
                if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    onSelectFile(nextFile);
                }
            },
        };
    }, [loading, mode, onConfirm, onNavigate, onSelectFile, path, selectedFile?.path]);

    const emptyText = error ? (
        <div className="file-picker-error">
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={error}/>
            <Button size="small" onClick={onReload}>重试</Button>
        </div>
    ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>;

    const statusPath = mode === "directory" ? path : selectedFile?.path;
    const statusLabel = mode === "directory" ? "当前目录" : "选择结果";
    const statusText = statusPath || (mode === "directory" ? "正在读取磁盘..." : "未选择");
    const statusIcon = mode === "directory"
        ? "FolderOpenOutlined"
        : selectedFile ? getFileIconType(selectedFile.name) : "FileOutlined";
    const rangeStart = total > 0 ? (page - 1) * pageSize + 1 : 0;
    const rangeEnd = Math.min(page * pageSize, total);
    const countText = loading && !loadedPath
        ? "加载中"
        : total > pageSize ? `${rangeStart}-${rangeEnd} / ${total} 项` : `${total} 项`;

    return (
        <div className={className}>
            <div className={`file-picker-toolbar${disks.length > 1 ? " file-picker-toolbar-with-disks" : ""}`}>
                <Tooltip title="上一级">
                    <Button
                        type="text"
                        className="file-picker-up"
                        disabled={!canGoUp || loading}
                        aria-label="返回上一级"
                        icon={<Icon type="ArrowLeftOutlined"/>}
                        onClick={() => onNavigate(isPathWithin(parentPath, diskRoot) ? parentPath : diskRoot)}/>
                </Tooltip>
                {editingPath ? (
                    <Input
                        ref={pathInputRef}
                        className="file-picker-path-input"
                        value={pathDraft}
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
                    <ol ref={pathRef} className="file-picker-path" aria-label={`当前路径：${path || "正在读取"}`}>
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
                                            title="点击编辑路径"
                                            disabled={loading}
                                            aria-current="page"
                                            aria-label={`编辑路径 ${item.path}`}
                                            onClick={onOpenPathEditor}>
                                            <span>{item.label}</span>
                                        </button>
                                    ) : (
                                        <button
                                            className="file-picker-path-button"
                                            type="button"
                                            title={item.path}
                                            disabled={loading}
                                            aria-label={`前往 ${item.path}`}
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
                {disks.length > 1 && (
                    <Select
                        className="file-picker-disks"
                        value={selectedDisk}
                        aria-label="切换磁盘"
                        placeholder="切换磁盘"
                        disabled={loading}
                        options={diskOptions}
                        onChange={onNavigate}/>
                )}
            </div>

            <Table<any>
                ref={tableRef}
                className="file-picker-table"
                rowKey="name"
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
                </div>
                <div className="file-picker-status-meta">
                    <span className="file-picker-count">{countText}</span>
                    {total > pageSize && (
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
