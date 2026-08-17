import React, {useCallback, useMemo} from "react";
import {Button, Input} from "antd";
import type {MenuProps} from "antd";
import {Icon} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import type {HeaderStyles} from "./header.types";

interface HeaderCommandBarProps {
    keyword: string;
    uploading: boolean;
    onKeywordChange: (value: string) => void;
    onCreate: (type: any) => void;
    onUpload: () => void;
    onRemoteDownload: () => void;
    onOpenRecycle: () => void;
    styles: Pick<
        HeaderStyles,
        "commandBar" | "searchBox" | "toolbarTrigger" | "toolbarAction" | "toolbarDropdown"
    >;
}

const HeaderCommandBar = ({
    keyword,
    uploading,
    onKeywordChange,
    onCreate,
    onUpload,
    onRemoteDownload,
    onOpenRecycle,
    styles,
}: HeaderCommandBarProps) => {
    const handleCreate = useCallback<NonNullable<MenuProps["onClick"]>>(({key}) => {
        onCreate(key as any);
    }, [onCreate]);
    const createMenu = useMemo<MenuProps>(() => ({
        items: [
            {key: "directory", label: "新建文件夹", icon: <Icon type="FolderAddOutlined"/>},
            {key: "file", label: "新建空文件", icon: <Icon type="FileAddOutlined"/>},
        ],
        onClick: handleCreate,
    }), [handleCreate]);

    return (
        <div className={styles.commandBar}>
            <Input
                className={`${styles.searchBox} file-search`}
                value={keyword}
                aria-label="搜索当前目录"
                allowClear
                prefix={<Icon type="SearchOutlined"/>}
                placeholder="搜索当前目录"
                onChange={(event) => onKeywordChange(event.target.value)}/>

            <div className="command-actions">
                <Dropdown
                    trigger={["click"]}
                    placement="bottomRight"
                    classNames={{root: styles.toolbarDropdown}}
                    menu={createMenu}>
                    <button
                        className={styles.toolbarTrigger}
                        type="button"
                        aria-label="新建文件或文件夹">
                        <Icon type="PlusOutlined" className="marker"/>
                        <span className="value">新建</span>
                        <Icon type="DownOutlined" className="arrow"/>
                    </button>
                </Dropdown>
                <Button
                    type="text"
                    className={styles.toolbarAction}
                    aria-label={uploading ? "查看上传" : "上传文件"}
                    icon={<Icon type={uploading ? "LoadingOutlined" : "UploadOutlined"}/>}
                    onClick={onUpload}>
                    <span className="value">{uploading ? "查看上传" : "上传文件"}</span>
                </Button>
                <Button
                    type="text"
                    className={styles.toolbarAction}
                    aria-label="远程下载"
                    icon={<Icon type="CloudDownloadOutlined"/>}
                    onClick={onRemoteDownload}>
                    <span className="value">远程下载</span>
                </Button>
                <Button
                    type="text"
                    className={styles.toolbarAction}
                    aria-label="回收站"
                    icon={<Icon type="DeleteOutlined"/>}
                    onClick={onOpenRecycle}>
                    <span className="value">回收站</span>
                </Button>
            </div>
        </div>
    );
};

export default React.memo(HeaderCommandBar);
