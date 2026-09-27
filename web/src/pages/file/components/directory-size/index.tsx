import React from "react";
import {Spin, Tooltip} from "antd";
import {formatFileSize} from "../../utils";

const DirectorySize = ({className = "", path, count, onCount}: any) => {
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
