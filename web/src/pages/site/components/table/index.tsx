import {useMemo, useState} from "react";
import {Button, Tooltip} from "antd";
import Dropdown from "@/pages/components/stable-dropdown";
import type {DataTableProps} from "@/components/table";
import {getFullPathByName, historyPush} from "@/utils/route";
import {ago} from "@/utils/time";

const useSiteTable = ({
    className,
    sites,
    loading,
    categoryData,
    typeData,
    statusData,
    sort,
    order,
    operating,
    onSortChange,
    onSettings,
    onAction,
}: any): DataTableProps<any> => {
    const [openRowId, setOpenRowId] = useState("");

    const columns = useMemo<any[]>(() => [
        {
            title: "网站名称",
            dataIndex: "name",
            key: "name",
            width: 190,
            className: "site-name-cell",
            sorter: true,
            sortOrder: sort === "name" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (_: any, record: any) => (
                <div className="site-name">
                    <button type="button" title={record.name || "未命名网站"} onClick={() => onSettings(record, "basic")}>
                        {record.name || "未命名网站"}
                    </button>
                </div>
            ),
        },
        {
            title: "分类",
            dataIndex: "type_id",
            key: "type_id",
            width: 110,
            className: "category-cell",
            sorter: true,
            sortOrder: sort === "type_id" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value: any) => {
                const label = Number(value) > 0 ? categoryData[String(value)] || "未知分类" : "全部分类";
                return <Tooltip title={label}><span className="site-category">{label}</span></Tooltip>;
            },
        },
        {
            title: "网站类型",
            dataIndex: "type",
            key: "type",
            width: 110,
            className: "type-cell",
            sorter: true,
            sortOrder: sort === "type" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value: any) => (
                <span className="site-type">{typeData[String(value)] || "未知类型"}</span>
            ),
        },
        {
            title: "网站目录",
            dataIndex: "root",
            key: "root",
            width: 220,
            className: "root-cell",
            render: (value: any) => {
                const root = String(value || "").trim();
                if (!root) {
                    return <span className="site-root">-</span>;
                }
                return (
                    <Tooltip title={root}>
                        <button
                            type="button"
                            className="site-root site-root-link"
                            onClick={() => historyPush(`${getFullPathByName("file")}?${new URLSearchParams({path: root})}`)}>
                            {root}
                        </button>
                    </Tooltip>
                );
            },
        },
        {
            title: "运行状态",
            dataIndex: "status",
            key: "status",
            width: 100,
            className: "status-cell",
            sorter: true,
            sortOrder: sort === "status" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value: any) => {
                const enabled = Number(value) === 0;
                return (
                    <span className={`site-state ${enabled ? "running" : "stopped"}`}>
                        <i/>{statusData[String(value)] || (enabled ? "已启用" : "已停用")}
                    </span>
                );
            },
        },
        {
            title: "更新时间",
            dataIndex: "update_time",
            key: "update_time",
            width: 130,
            className: "update-cell",
            sorter: true,
            sortOrder: sort === "update_time" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value: any) => {
                const time = String(value || "");
                return (
                    <Tooltip title={time || "-"}>
                        <span className="update-time">{time ? (ago(time) || time) : "-"}</span>
                    </Tooltip>
                );
            },
        },
        {
            title: "操作",
            key: "action",
            width: 82,
            fixed: "right",
            className: "action-cell",
            render: (_: any, record: any) => {
                const id = String(record.id);
                const enabled = Number(record.status) === 0;
                const items = [
                    {key: "settings", label: "网站设置", onClick: () => onSettings(record, "basic")},
                    {key: "domain", label: "域名管理", onClick: () => onSettings(record, "domain")},
                    {key: "ssl", label: "SSL设置", onClick: () => onSettings(record, "ssl")},
                    {key: "log", label: "网站日志", onClick: () => onSettings(record, "log")},
                    {type: "divider" as const},
                    enabled
                        ? {key: "disable", label: "停用网站", onClick: () => onAction("disable", record)}
                        : {key: "enable", label: "启用网站", onClick: () => onAction("enable", record)},
                    {key: "delete", label: "删除网站", danger: true, onClick: () => onAction("delete", record)},
                ];
                return (
                    <Dropdown
                        open={openRowId === id}
                        onOpenChange={(open) => setOpenRowId(open ? id : "")}
                        menu={{items}}
                        trigger={["click"]}
                        placement="bottom"
                        arrow>
                        <Button size="small" loading={operating?.has(id)}>操作</Button>
                    </Dropdown>
                );
            },
        },
    ], [categoryData, onAction, onSettings, openRowId, operating, order, sort, statusData, typeData]);

    return {
        className,
        columns,
        dataSource: sites,
        loading,
        onSortChange,
        rowKey: (record) => String(record.id),
        rowClassName: (record) => openRowId === String(record.id) ? "action-menu-open" : "",
    };
};

export default useSiteTable;
