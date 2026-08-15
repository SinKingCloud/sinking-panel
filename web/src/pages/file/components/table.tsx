import React, {useCallback, useEffect, useMemo, useState} from "react";
import {Button, Table as AntTable, Tooltip} from "antd";
import type {MenuProps, TableColumnsType, TableProps} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {isFilePreviewable} from "@/pages/components/file-preview";
import type {FilePreviewItem} from "@/pages/components/file-preview";
import Dropdown from "@/pages/components/stable-dropdown";
import type {FileOrderField, FileOrderType, FileRecord} from "@/service/api/file";
import type {DirectoryCountMap} from "../hooks/directory-counts";
import DirectorySize from "./directory-size";
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
    items: FileRecord[];
    loading: boolean;
    actionsDisabled?: boolean;
    sort?: FileOrderField;
    order?: FileOrderType;
    directoryCounts?: DirectoryCountMap;
    onCountDirectory: (path: string) => void;
    operatingPaths?: ReadonlySet<string>;
    selectedPaths: ReadonlySet<string>;
    selectionDisabled: boolean;
    onSelectionChange: (paths: string[]) => void;
    onSortChange: (field?: FileOrderField, order?: "ascend" | "descend") => void;
    onOpen: (record: FileRecord) => void;
    onPreview: (files: readonly FilePreviewItem[], active: string) => void;
    onDownload: (record: FileRecord) => void;
    onCopy: (record: FileRecord) => void;
    onMove: (record: FileRecord) => void;
    onProperties: (record: FileRecord) => void;
    onOperation: (mode: "compress" | "extract", record: FileRecord) => void;
    onDelete: (record: FileRecord) => void;
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
    recordsByPath: ReadonlyMap<string, FileRecord>;
    actionsDisabled: boolean;
    operatingPaths?: ReadonlySet<string>;
    openMenu?: FileMenuState;
    fileMenuClassName: string;
    getMenuItems: (record: FileRecord, recordPath: string, disabled?: boolean) => MenuProps["items"];
    changeMenuOpen: (open: boolean, recordPath: string, source: FileMenuSource) => void;
}

const FileContextMenu = React.createContext<FileContextMenuValue | undefined>(undefined);

const ContextMenuRow = (rowProps: FileTableRowProps) => {
    const context = React.useContext(FileContextMenu);
    const recordPath = String(rowProps["data-row-key"] || "");
    const record = context?.recordsByPath.get(recordPath);
    if (!context || !record) {
        return <tr {...rowProps}/>;
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
            <tr {...rowProps}/>
        </Dropdown>
    );
};

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
    onPreview,
    onDownload,
    onCopy,
    onMove,
    onProperties,
    onOperation,
    onDelete,
}: FileTableProps) => {
    const theme = useTheme();
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
    const previewFiles = useMemo<FilePreviewItem[]>(() => items
        .filter((record) => !record.is_dir && isFilePreviewable(record.name))
        .map((record) => ({
            name: record.name,
            path: joinFilePath(path, record.name),
            size: record.size,
        })), [items, path]);
    const getMenuItems = useCallback((
        record: FileRecord,
        recordPath: string,
        disabled = false,
    ): MenuProps["items"] => {
        const archive = !record.is_dir && isExtractableFile(record.name);
        const previewable = !record.is_dir && isFilePreviewable(record.name);
        const items: MenuProps["items"] = [
            ...(!record.is_dir ? [
                ...(previewable ? [
                    {key: "preview", label: "预览", onClick: () => onPreview(previewFiles, recordPath)},
                ] : []),
                {key: "download", label: "下载", onClick: () => onDownload(record)},
            ] : [{key: "open", label: "打开", onClick: () => onOpen(record)}]),
            {type: "divider" as const},
            {key: "copy", label: "复制", onClick: () => onCopy(record)},
            {key: "move", label: "移动", onClick: () => onMove(record)},
            {key: "properties", label: "属性", onClick: () => onProperties(record)},
            {key: "compress", label: "压缩", onClick: () => onOperation("compress", record)},
            ...(archive ? [{key: "extract", label: "解压", onClick: () => onOperation("extract", record)}] : []),
            {type: "divider" as const},
            {
                key: "delete",
                label: "删除",
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
        onCopy,
        onDelete,
        onDownload,
        onMove,
        onOpen,
        onOperation,
        onPreview,
        onProperties,
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

    const columns = useMemo<TableColumnsType<FileRecord>>(() => [
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
                    : previewable ? `预览文件 ${record.name}` : `查看文件属性 ${record.name}`;
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
                            else onProperties(record);
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

    const change = useCallback<NonNullable<TableProps<FileRecord>["onChange"]>>((_, __, sorter, extra) => {
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
            <AntTable<FileRecord>
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
