import React from "react";
import {Spin, Tooltip} from "antd";
import type {DirectoryCountState} from "../hooks/directory-counts";
import {formatFileSize} from "../utils";

interface DirectorySizeProps {
    className?: string;
    path: string;
    count?: DirectoryCountState;
    onCount: (path: string) => void;
}

const DirectorySize = ({className = "", path, count, onCount}: DirectorySizeProps) => {
    const countDirectory = () => onCount(path);
    if (count?.status === "loading") {
        return (
            <span className={className} aria-label="正在统计目录大小">
                <Spin size="small"/>
            </span>
        );
    }
    if (count?.status === "success") {
        return <span className={className}>{formatFileSize(count.data.size)}</span>;
    }
    if (count?.status === "error") {
        return (
            <Tooltip title={count.message}>
                <button
                    className={`${className} calculate error`}
                    type="button"
                    aria-label="重新计算目录大小"
                    onClick={countDirectory}>
                    重新计算
                </button>
            </Tooltip>
        );
    }
    return (
        <button
            className={`${className} calculate`}
            type="button"
            aria-label="计算目录大小"
            onClick={countDirectory}>
            计算
        </button>
    );
};

export default React.memo(DirectorySize);
