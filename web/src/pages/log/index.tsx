import React, {useRef} from "react";
import {Button} from "antd";
import {Body, ProTable, Title} from "sinking-antd";
import type {ProTableRef} from "sinking-antd";
import {getLog} from "@/service/api/system";
import useEnum from "@/utils/enum";
import {dateRangeTransform, getData} from "@/utils/page";
import Clear, {ClearRef} from "./components/clear";

const color: Record<string, string> = {
    "0": "green",
    "1": "geekblue",
    "2": "red",
    "3": "warning",
    "4": "blue",
};

export default (): React.ReactNode => {
    const [enumData, enumLoading] = useEnum("log");
    const tableRef = useRef<ProTableRef | null>(null);
    const clearRef = useRef<ClearRef | null>(null);

    const columns: any[] = [
        {
            title: "ID",
            dataIndex: "id",
            tip: "日志ID",
            width: 180,
            sorter: true,
            hideInTable: true,
            hideInSearch: true,
            copyable: true,
            ellipsis: true,
        },
        {
            title: "操作IP",
            dataIndex: "ip",
            tip: "请求来源IP地址",
            width: 120,
            valueType: "text",
            copyable: true,
            ellipsis: true,
        },
        {
            title: "IP归属地",
            dataIndex: "location",
            tip: "请求来源IP归属地",
            width: 200,
            valueType: "text",
            copyable: true,
            ellipsis: true,
        },
        {
            title: "操作类型",
            dataIndex: "type",
            tip: "操作事件类型",
            width: 100,
            valueEnum: Object.fromEntries(Object.entries(enumData?.type || {}).map(([key, value]) => {
                return [key, {text: value, color: color[key]}];
            })),
            ellipsis: true,
        },
        {
            title: "操作标题",
            dataIndex: "title",
            tip: "操作事件标题",
            width: 120,
            valueType: "text",
            ellipsis: true,
            hideInSearch: true,
        },
        {
            title: "操作内容",
            dataIndex: "content",
            tip: "操作事件详细内容",
            width: 160,
            valueType: "text",
            ellipsis: true,
            hideInSearch: true,
        },
        {
            title: "操作时间",
            dataIndex: "create_time",
            tip: "日志创建时间",
            width: 100,
            valueType: "dateRange",
            sorter: true,
            transform: dateRangeTransform("create_time"),
            ellipsis: true,
        },
    ];

    return (
        <Body loading={enumLoading}>
            <ProTable
                ref={tableRef}
                extraRefreshBtn
                title={<Title>操作日志</Title>}
                extra={(
                    <Button
                        type="primary"
                        aria-label="清理操作日志"
                        onClick={() => clearRef.current?.open()}
                    >
                        清理
                    </Button>
                )}
                rowKey="id"
                columns={columns}
                defaultPage={1}
                defaultPageSize={10}
                request={(params, sort) => getData(params, sort, getLog)}
            />
            <Clear ref={clearRef} onSuccess={() => tableRef.current?.refreshTableData?.()}/>
        </Body>
    );
};
