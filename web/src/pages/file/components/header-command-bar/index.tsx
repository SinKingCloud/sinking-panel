import React, {useCallback, useMemo} from "react";
import type {MenuProps} from "antd";
import {Icon} from "sinking-antd";
import {TableAction, TableToolbar} from "@/pages/components/table";

interface HeaderCommandBarProps {
    keyword: string;
    uploading: boolean;
    refreshing: boolean;
    onKeywordChange: (value: string) => void;
    onCreate: (type: any) => void;
    onUpload: () => void;
    onRemoteDownload: () => void;
    onOpenTerminal: () => void;
    onOpenRecycle: () => void;
    onRefresh: () => void;
}

const HeaderCommandBar = ({
    keyword,
    uploading,
    refreshing,
    onKeywordChange,
    onCreate,
    onUpload,
    onRemoteDownload,
    onOpenTerminal,
    onOpenRecycle,
    onRefresh,
}: HeaderCommandBarProps) => {
    const handleCreate = useCallback<NonNullable<MenuProps["onClick"]>>(({key}) => {
        onCreate(key);
    }, [onCreate]);
    const createMenu = useMemo<MenuProps>(() => ({
        items: [
            {key: "directory", label: "新建文件夹", icon: <Icon type="FolderAddOutlined"/>},
            {key: "file", label: "新建空文件", icon: <Icon type="FileAddOutlined"/>},
        ],
        onClick: handleCreate,
    }), [handleCreate]);

    return (
        <TableToolbar
            search={{
                value: keyword,
                ariaLabel: "搜索当前目录",
                placeholder: "搜索当前目录",
                onChange: onKeywordChange,
            }}
            refresh={{loading: refreshing, ariaLabel: "刷新文件列表", onClick: onRefresh}}>
            <TableAction
                label="新建"
                icon="PlusOutlined"
                suffixIcon="DownOutlined"
                aria-label="新建文件或文件夹"
                menu={createMenu}/>
            <TableAction
                label={uploading ? "查看上传" : "上传文件"}
                icon={uploading ? "LoadingOutlined" : "UploadOutlined"}
                aria-label={uploading ? "查看上传" : "上传文件"}
                onClick={onUpload}/>
            <TableAction
                label="远程下载"
                icon="CloudDownloadOutlined"
                aria-label="远程下载"
                onClick={onRemoteDownload}/>
            <TableAction
                label="终端"
                icon="CodeOutlined"
                aria-label="打开终端"
                onClick={onOpenTerminal}/>
            <TableAction
                label="回收站"
                icon="DeleteOutlined"
                aria-label="回收站"
                onClick={onOpenRecycle}/>
        </TableToolbar>
    );
};

export default React.memo(HeaderCommandBar);
