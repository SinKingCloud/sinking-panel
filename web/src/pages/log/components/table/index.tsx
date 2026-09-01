import React, {useMemo} from "react";
import {Tag, Tooltip, Typography} from "antd";
import {DataTable} from "@/pages/components/table";

const typeColors: Record<string, string> = {
    "0": "success",
    "1": "processing",
    "2": "error",
    "3": "warning",
    "4": "blue",
};

const CopyText = ({value}: {value: any}) => {
    const text = String(value || "-");
    return (
        <Typography.Text className="log-copy" copyable={text === "-" ? false : {text}} ellipsis={{tooltip: text}}>
            {text}
        </Typography.Text>
    );
};

const Table = ({
    className,
    logs,
    loading,
    typeData,
    sort,
    order,
    onSortChange,
}: any) => {
    const columns = useMemo<any[]>(() => [
        {
            title: "操作IP",
            dataIndex: "ip",
            key: "ip",
            width: 140,
            render: (value: any) => <CopyText value={value}/>,
        },
        {
            title: "IP归属地",
            dataIndex: "location",
            key: "location",
            width: 180,
            render: (value: any) => <CopyText value={value}/>,
        },
        {
            title: "操作类型",
            dataIndex: "type",
            key: "type",
            width: 100,
            render: (value: any) => (
                <Tag className="log-type" variant="filled" color={typeColors[String(value)]}>
                    {typeData[String(value)] || "未知类型"}
                </Tag>
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

    return (
        <DataTable
            className={className}
            columns={columns}
            dataSource={logs}
            loading={loading}
            onSortChange={onSortChange}
            rowKey={(record) => String(record.id)}/>
    );
};

export default React.memo(Table);
