import React, {useCallback, useEffect, useMemo, useState} from "react";
import {App, Button, Table as AntTable, Tooltip} from "antd";
import type {MenuProps, TableColumnsType, TableProps} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {isFilePreviewable} from "@/pages/components/file-preview";
import Dropdown from "@/pages/components/stable-dropdown";
import type {DirectoryCountMap} from "../hooks/directory-counts";
import DirectorySize from "./directory-size";
import {copyTextToClipboard} from "./properties.utils";
import useStyles from "./table.styles";
import {
    formatFileMode,
    formatFileSize,
    formatFileTime,
    getFileIconType,
    isExtractableFile,
    joinFilePath,
} from "../utils";

export interface FileTableProps {
    path: string;
    items: any[];
    loading: boolean;
    actionsDisabled?: boolean;
    sort?: any;
    order?: any;
    directoryCounts?: DirectoryCountMap;
    onCountDirectory: (path: string) => void;
    operatingPaths?: ReadonlySet<string>;
    selectedPaths: ReadonlySet<string>;
    selectionDisabled: boolean;
    onSelectionChange: (paths: string[]) => void;
    onSortChange: (field?: any, order?: "ascend" | "descend") => void;
    onOpen: (record: any) => void;
    onEdit: (path: string, name: string) => void;
    onPreview: (files: readonly any[], active: string) => void;
    onDownload: (record: any) => void;
    onRename: (record: any) => void;
    onCopy: (record: any) => void;
    onMove: (record: any) => void;
    onProperties: (record: any) => void;
    onOperation: (mode: "compress" | "extract", record: any) => void;
    onDelete: (record: any) => void;
}

type FileMenuSource = "action" | "context";

interface FileMenuState {
    path: string;
    source: FileMenuSource;
}

type FileTableRowProps = React.HTMLAttributes<HTMLTableRowElement> & {
    "data-row-key"?: React.Key;
};

interface FileContextMenuValue {
    recordsByPath: ReadonlyMap<string, any>;
    actionsDisabled: boolean;
    operatingPaths?: ReadonlySet<string>;
    openMenu?: FileMenuState;
    fileMenuClassName: string;
    getMenuItems: (record: any, recordPath: string, disabled?: boolean) => MenuProps["items"];
    changeMenuOpen: (open: boolean, recordPath: string, source: FileMenuSource) => void;
}

const FileContextMenu = React.createContext<FileContextMenuValue | undefined>(undefined);

const ContextMenuRow = React.forwardRef<HTMLTableRowElement, FileTableRowProps>((rowProps, ref) => {
    const context = React.useContext(FileContextMenu);
    const recordPath = String(rowProps["data-row-key"] || "");
    const record = context?.recordsByPath.get(recordPath);
    if (!context || !record) {
        return <tr {...rowProps} ref={ref}/>;
    }
    const disabled = context.actionsDisabled || Boolean(context.operatingPaths?.has(recordPath));
    return (
        <Dropdown
            disabled={context.actionsDisabled}
            open={context.openMenu?.source === "context" && context.openMenu.path === recordPath}
            onOpenChange={(open) => context.changeMenuOpen(open, recordPath, "context")}
            classNames={{root: context.fileMenuClassName}}
            menu={{items: context.getMenuItems(record, recordPath, disabled)}}
            trigger={["contextMenu"]}>
            <tr {...rowProps} ref={ref}/>
        </Dropdown>
    );
});

ContextMenuRow.displayName = "ContextMenuRow";

const tableComponents = {body: {row: ContextMenuRow}};

const FileTable = ({
    path,
    items,
    loading,
    actionsDisabled = false,
    sort,
    order,
    directoryCounts,
    onCountDirectory,
    operatingPaths,
    selectedPaths,
    selectionDisabled,
    onSelectionChange,
    onSortChange,
    onOpen,
    onEdit,
    onPreview,
    onDownload,
    onRename,
    onCopy,
    onMove,
    onProperties,
    onOperation,
    onDelete,
}: FileTableProps) => {
    const theme = useTheme();
    const {message} = App.useApp();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles} = useStyles({compact, dark});
    const [openMenu, setOpenMenu] = useState<FileMenuState>();
    const openRowPath = openMenu?.path || "";
    const changeMenuOpen = useCallback((open: boolean, recordPath: string, source: FileMenuSource) => {
        setOpenMenu((current) => {
            if (open) {
                return {path: recordPath, source};
            }
            return current?.path === recordPath && current.source === source ? undefined : current;
        });
    }, []);
    const previewFiles = useMemo<any[]>(() => items
        .filter((record) => !record.is_dir && isFilePreviewable(record.name))
        .map((record) => ({
            name: record.name,
            path: joinFilePath(path, record.name),
            size: record.size,
        })), [items, path]);
    const copyValue = useCallback(async (value: string, label: string) => {
        if (await copyTextToClipboard(value)) {
            message.success(`${label}已复制`);
        } else {
            message.error(`${label}复制失败`);
        }
    }, [message]);
    const getMenuItems = useCallback((
        record: any,
        recordPath: string,
        disabled = false,
    ): MenuProps["items"] => {
        const archive = !record.is_dir && isExtractableFile(record.name);
        const previewable = !record.is_dir && isFilePreviewable(record.name);
        const items: MenuProps["items"] = [
            ...(!record.is_dir ? [
                {key: "edit", label: "编辑文件", onClick: () => onEdit(recordPath, record.name)},
                ...(previewable ? [
                    {key: "preview", label: "预览文件", onClick: () => onPreview(previewFiles, recordPath)},
                ] : []),
                {key: "download", label: "下载文件", onClick: () => onDownload(record)},
            ] : [{key: "open", label: "打开目录", onClick: () => onOpen(record)}]),
            {type: "divider" as const},
            {key: "copy", label: "复制文件", onClick: () => onCopy(record)},
            {key: "move", label: "移动文件", onClick: () => onMove(record)},
            {
                key: "copy-name",
                label: "复制名称",
                onClick: () => void copyValue(record.name, "名称"),
            },
            {
                key: "copy-path",
                label: "复制路径",
                onClick: () => void copyValue(recordPath, "路径"),
            },
            {key: "rename", label: "重新命名", onClick: () => onRename(record)},
            {key: "properties", label: "文件属性", onClick: () => onProperties(record)},
            {key: "compress", label: "压缩文件", onClick: () => onOperation("compress", record)},
            ...(archive ? [{key: "extract", label: "解压文件", onClick: () => onOperation("extract", record)}] : []),
            {type: "divider" as const},
            {
                key: "delete",
                label: "删除文件",
                danger: true,
                onClick: () => onDelete(record),
            },
        ];
        if (!disabled && !actionsDisabled) {
            return items;
        }
        return items.map((item) => item && item.type !== "divider" ? {...item, disabled: true} : item);
    }, [
        actionsDisabled,
        copyValue,
        onCopy,
        onDelete,
        onDownload,
        onEdit,
        onMove,
        onOpen,
        onOperation,
        onPreview,
        onProperties,
        onRename,
        previewFiles,
    ]);

    const recordsByPath = useMemo(() => new Map(
        items.map((record) => [joinFilePath(path, record.name), record]),
    ), [items, path]);
    useEffect(() => {
        setOpenMenu((current) => !actionsDisabled && current && recordsByPath.has(current.path)
            ? current
            : undefined);
    }, [actionsDisabled, recordsByPath]);
    const contextMenuValue = useMemo<FileContextMenuValue>(() => ({
        recordsByPath,
        actionsDisabled,
        operatingPaths,
        openMenu,
        fileMenuClassName: styles.fileMenu,
        getMenuItems,
        changeMenuOpen,
    }), [
        actionsDisabled,
        changeMenuOpen,
        getMenuItems,
        openMenu,
        operatingPaths,
        recordsByPath,
        styles.fileMenu,
    ]);

    const columns = useMemo<TableColumnsType<any>>(() => [
        {
            title: "名称",
            dataIndex: "name",
            key: "name",
            width: 300,
            sorter: true,
            sortOrder: sort === "name" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (_, record) => {
                const recordPath = joinFilePath(path, record.name);
                const rowDisabled = actionsDisabled || Boolean(operatingPaths?.has(recordPath));
                const previewable = !record.is_dir && isFilePreviewable(record.name);
                const label = record.is_dir
                    ? `打开目录 ${record.name}`
                    : previewable ? `预览文件 ${record.name}` : `编辑文件 ${record.name}`;
                return (
                    <button
                        className={styles.fileNameButton}
                        type="button"
                        disabled={rowDisabled}
                        aria-label={label}
                        onClick={() => {
                            if (rowDisabled) return;
                            if (record.is_dir) onOpen(record);
                            else if (previewable) onPreview(previewFiles, recordPath);
                            else onEdit(recordPath, record.name);
                        }}>
                        <span className={styles.fileNameContent}>
                            <Icon
                                className={`${styles.fileIcon} ${record.is_dir ? "folder" : ""}`}
                                type={getFileIconType(record.name, record.is_dir)}/>
                            <Tooltip title={record.name}>
                                <span className={`${styles.fileName} file-name`}>{record.name}</span>
                            </Tooltip>
                        </span>
                    </button>
                );
            },
        },
        {
            title: "大小",
            dataIndex: "size",
            key: "size",
            width: 130,
            render: (value, record) => {
                const recordPath = joinFilePath(path, record.name);
                return record.is_dir ? (
                    <DirectorySize
                        className={`${styles.fileMeta} ${styles.directorySize}`}
                        path={recordPath}
                        count={directoryCounts?.get(recordPath)}
                        onCount={onCountDirectory}/>
                ) : <span className={styles.fileMeta}>{formatFileSize(value)}</span>;
            },
        },
        {
            title: "权限",
            dataIndex: "mode",
            key: "mode",
            width: 100,
            render: (value) => <span className={styles.fileMeta}>{formatFileMode(value)}</span>,
        },
        {
            title: "修改时间",
            dataIndex: "update_time",
            key: "update_time",
            width: 190,
            sorter: true,
            sortOrder: sort === "update_time" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value) => <span className={styles.fileMeta}>{formatFileTime(value)}</span>,
        },
        {
            title: "操作",
            key: "action",
            width: 80,
            fixed: "right",
            className: "action-cell",
            render: (_, record) => {
                const recordPath = joinFilePath(path, record.name);
                return (
                    <Dropdown
                        open={openMenu?.source === "action" && openMenu.path === recordPath}
                        onOpenChange={(open) => changeMenuOpen(open, recordPath, "action")}
                        classNames={{root: styles.fileMenu}}
                        menu={{items: getMenuItems(
                            record,
                            recordPath,
                            Boolean(operatingPaths?.has(recordPath)),
                        )}}
                        trigger={["click"]}
                        placement="bottom"
                        arrow>
                        <Button
                            size="small"
                            disabled={actionsDisabled || operatingPaths?.has(recordPath)}>
                            操作
                        </Button>
                    </Dropdown>
                );
            },
        },
    ], [
        changeMenuOpen,
        actionsDisabled,
        directoryCounts,
        getMenuItems,
        onCountDirectory,
        onEdit,
        onOpen,
        onPreview,
        onProperties,
        openMenu,
        operatingPaths,
        order,
        path,
        previewFiles,
        sort,
        styles,
    ]);

    const change = useCallback<NonNullable<TableProps<any>["onChange"]>>((_, __, sorter, extra) => {
        if (extra?.action !== "sort") {
            return;
        }
        const current = Array.isArray(sorter) ? sorter[0] : sorter;
        const field = current?.field || current?.columnKey;
        const supported = field === "name" || field === "update_time";
        onSortChange(supported ? field : undefined, current?.order || undefined);
    }, [onSortChange]);

    return (
        <FileContextMenu.Provider value={contextMenuValue}>
            <AntTable<any>
                className={styles.fileTable}
                columns={columns}
                dataSource={items}
                loading={loading}
                pagination={false}
                onChange={change}
                components={tableComponents}
                rowSelection={{
                    selectedRowKeys: Array.from(selectedPaths),
                    columnWidth: 44,
                    fixed: true,
                    onChange: (keys) => onSelectionChange(keys.map(String)),
                    getCheckboxProps: (record) => ({
                        disabled: selectionDisabled || Boolean(operatingPaths?.has(joinFilePath(path, record.name))),
                    }),
                }}
                rowKey={(record) => joinFilePath(path, record.name)}
                rowClassName={(record) => openRowPath === joinFilePath(path, record.name) ? "action-menu-open" : ""}
                scroll={{x: "max-content"}}
                showSorterTooltip={false}
                tableLayout="fixed"
                size="middle"/>
        </FileContextMenu.Provider>
    );
};

export default React.memo(FileTable);
