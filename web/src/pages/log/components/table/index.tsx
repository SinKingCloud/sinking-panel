import React, {useMemo} from "react";
import {Tooltip, Typography} from "antd";
import type {DataTableProps} from "@/components/table";

const CopyText = ({value}: {value: any}) => {
    const text = String(value || "-");
    return (
        <Typography.Text className="log-copy" copyable={text === "-" ? false : {text}}>
            {text}
        </Typography.Text>
    );
};

const useLogTable = ({
    className,
    logs,
    loading,
    typeData,
    sort,
    order,
    onSortChange,
}: any): DataTableProps<any> => {
    const columns = useMemo<any[]>(() => [
        {
            title: "操作IP",
            dataIndex: "ip",
            key: "ip",
            width: 150,
            render: (value: any) => <CopyText value={value}/>,
        },
        {
            title: "IP归属地",
            dataIndex: "location",
            key: "location",
            width: 220,
            render: (value: any) => <CopyText value={value}/>,
        },
        {
            title: "操作类型",
            dataIndex: "type",
            key: "type",
            width: 100,
            render: (value: any) => (
                <span className={`log-type is-${value}`}>
                    <i/>
                    {typeData[String(value)] || "未知类型"}
                </span>
            ),
        },
        {
            title: "操作标题",
            dataIndex: "title",
            key: "title",
            width: 150,
            render: (value: any) => {
                const text = String(value || "-");
                return <Tooltip title={text}><span className="log-text">{text}</span></Tooltip>;
            },
        },
        {
            title: "操作内容",
            dataIndex: "content",
            key: "content",
            width: 260,
            render: (value: any) => {
                const text = String(value || "-");
                return <Tooltip title={text}><span className="log-text">{text}</span></Tooltip>;
            },
        },
        {
            title: "操作时间",
            dataIndex: "create_time",
            key: "create_time",
            width: 170,
            sorter: true,
            sortOrder: sort === "create_time" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value: any) => {
                const text = String(value || "-");
                return <Tooltip title={text}><span className="log-time">{text}</span></Tooltip>;
            },
        },
    ], [order, sort, typeData]);

    return {
        className,
        columns,
        dataSource: logs,
        loading,
        onSortChange,
        rowKey: (record) => String(record.id),
    };
};

export default useLogTable;
