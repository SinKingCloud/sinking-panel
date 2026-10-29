import React, {useCallback, useEffect, useMemo, useState} from "react";
import {App, Button, Tooltip} from "antd";
import type {MenuProps, TableColumnsType} from "antd";
import {Icon} from "sinking-antd";
import {isFilePreviewable} from "@/pages/components/file-preview";
import ActionDropdown from "@/pages/components/ui/action-dropdown";
import DirectorySize from "../directory-size";
import {copyTextToClipboard} from "../properties/utils";
import useStyles from "./styles";
import {
    formatFileMode,
    formatFileSize,
    formatFileTime,
    getFileIconType,
    isExtractableFile,
    joinFilePath,
} from "../../utils";

const FileContextMenu = React.createContext<any>(undefined);

const ContextMenuRow = React.forwardRef<HTMLTableRowElement, any>((rowProps, ref) => {
    const context = React.useContext(FileContextMenu);
    const recordPath = String(rowProps["data-row-key"] || "");
    const record = context?.recordsByPath.get(recordPath);
    if (!context || !record) {
        return <tr {...rowProps} ref={ref}/>;
    }
    const disabled = context.actionsDisabled || Boolean(context.operatingPaths?.has(recordPath));
    return (
        <ActionDropdown
            disabled={context.actionsDisabled}
            open={context.openMenu?.source === "context" && context.openMenu.path === recordPath}
            onOpenChange={(open) => context.changeMenuOpen(open, recordPath, "context")}
            classNames={{root: context.fileMenuClassName}}
            menu={{items: context.getMenuItems(record, recordPath, disabled)}}
            trigger={["contextMenu"]}
            placement="bottomLeft"
            arrow={false}>
            <tr {...rowProps} ref={ref}/>
        </ActionDropdown>
    );
});

ContextMenuRow.displayName = "ContextMenuRow";

const tableComponents = {body: {row: ContextMenuRow}};

const useFileTable = ({
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
    onPermissions,
    onCopy,
    onMove,
    onProperties,
    onOperation,
    onDelete,
}: any) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const [openMenu, setOpenMenu] = useState<any>();
    const openRowPath = openMenu?.path || "";
    const changeMenuOpen = useCallback((open: boolean, recordPath: string, source: any) => {
        setOpenMenu((current: any) => {
            if (open) {
                return {path: recordPath, source};
            }
            return current?.path === recordPath && current.source === source ? undefined : current;
        });
    }, []);
    const previewFiles = useMemo<any[]>(() => items
        .filter((record: any) => !record.is_dir && isFilePreviewable(record.name))
        .map((record: any) => ({
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
                {key: "edit", label: "编辑", onClick: () => onEdit(recordPath, record.name)},
                ...(previewable ? [
                    {key: "preview", label: "预览", onClick: () => onPreview(previewFiles, recordPath)},
                ] : []),
                {key: "download", label: "下载", onClick: () => onDownload(record)},
            ] : [{key: "open", label: "打开", onClick: () => onOpen(record)}]),
            {type: "divider" as const},
            {key: "copy", label: "复制", onClick: () => onCopy(record)},
            {key: "move", label: "移动", onClick: () => onMove(record)},
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
            {key: "rename", label: "重命名", onClick: () => onRename(record)},
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
        items.map((record: any) => [joinFilePath(path, record.name), record]),
    ), [items, path]);
    useEffect(() => {
        setOpenMenu((current: any) => !actionsDisabled && current && recordsByPath.has(current.path)
            ? current
            : undefined);
    }, [actionsDisabled, recordsByPath]);
    const contextMenuValue = useMemo<any>(() => ({
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
            render: (value, record) => {
                const recordPath = joinFilePath(path, record.name);
                const rowDisabled = actionsDisabled || Boolean(operatingPaths?.has(recordPath));
                return (
                    <button
                        className={styles.fileMetaButton}
                        type="button"
                        disabled={rowDisabled}
                        aria-label={`修改权限 ${record.name}`}
                        onClick={() => {
                            if (!rowDisabled) {
                                onPermissions(record);
                            }
                        }}>
                        <span className={`${styles.fileMeta} file-meta`}>{formatFileMode(value)}</span>
                    </button>
                );
            },
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
                    <ActionDropdown
                        open={openMenu?.source === "action" && openMenu.path === recordPath}
                        onOpenChange={(open) => changeMenuOpen(open, recordPath, "action")}
                        classNames={{root: styles.fileMenu}}
                        menu={{items: getMenuItems(
                            record,
                            recordPath,
                            Boolean(operatingPaths?.has(recordPath)),
                        )}}>
                        <Button
                            size="small"
                            disabled={actionsDisabled || operatingPaths?.has(recordPath)}>
                            操作
                        </Button>
                    </ActionDropdown>
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
        onPermissions,
        onProperties,
        openMenu,
        operatingPaths,
        order,
        path,
        previewFiles,
        sort,
        styles,
    ]);

    const changeSort = useCallback((field: string, value?: "ascend" | "descend") => {
        const supported = field === "name" || field === "update_time";
        onSortChange(supported ? field : undefined, value);
    }, [onSortChange]);
    const rowKey = useCallback((record: any) => joinFilePath(path, record.name), [path]);
    const selectedRowKeys = useMemo(() => Array.from<string>(selectedPaths), [selectedPaths]);

    return {
        columns,
        dataSource: items,
        loading,
        onSortChange: changeSort,
        components: tableComponents,
        rowSelection: {
            selectedRowKeys,
            columnWidth: 44,
            fixed: true,
            onChange: (keys: React.Key[]) => onSelectionChange(keys.map(String)),
            getCheckboxProps: (record: any) => ({
                disabled: selectionDisabled || Boolean(operatingPaths?.has(joinFilePath(path, record.name))),
            }),
        },
        rowKey,
        rowClassName: (record: any) => openRowPath === joinFilePath(path, record.name) ? "action-menu-open" : "",
        tableRender: (table: React.ReactNode) => (
            <FileContextMenu.Provider value={contextMenuValue}>{table}</FileContextMenu.Provider>
        ),
    };
};

export default useFileTable;
