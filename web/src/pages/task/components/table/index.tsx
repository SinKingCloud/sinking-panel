import React, {useMemo, useState} from "react";
import {Button, Tooltip} from "antd";
import Dropdown from "@/pages/components/stable-dropdown";
import {DataTable} from "@/pages/components/table";
import {ago} from "@/utils/time";
import {describeTaskSchedule} from "../../utils";

const formatRunTime = (value: any) => {
    const time = String(value || "");
    return !time || time.startsWith("0001-01-01") ? "尚未运行" : time;
};

const Table = ({
    className,
    tasks,
    loading,
    typeData,
    execTypeData,
    statusData,
    sort,
    order,
    operating,
    onSortChange,
    onAction,
    onEdit,
    onLog,
}: any) => {
    const [openRowId, setOpenRowId] = useState("");

    const columns = useMemo<any[]>(() => [
        {
            title: "任务名称",
            dataIndex: "name",
            key: "name",
            width: 180,
            className: "task-name-cell",
            render: (_: any, record: any) => (
                <div className="task-name">
                    <button type="button" onClick={() => onLog(record)}>
                        {record.name || "未命名任务"}
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
                const label = Number(value) === 0 ? "全部分类" : (typeData[String(value)] || "未知分类");
                return (
                    <Tooltip title={label}>
                        <span className="task-category">{label}</span>
                    </Tooltip>
                );
            },
        },
        {
            title: "执行方式",
            dataIndex: "exec_type",
            key: "exec_type",
            width: 100,
            className: "exec-type-cell",
            sorter: true,
            sortOrder: sort === "exec_type" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value: any) => (
                <span className="task-exec-type">{execTypeData[String(value)] || "-"}</span>
            ),
        },
        {
            title: "执行周期",
            dataIndex: "spec",
            key: "spec",
            width: 150,
            className: "schedule-cell",
            render: (value: any) => (
                <span className="schedule-value">{describeTaskSchedule(value)}</span>
            ),
        },
        {
            title: "状态",
            dataIndex: "status",
            key: "status",
            width: 80,
            className: "status-cell",
            sorter: true,
            sortOrder: sort === "status" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value: any) => {
                const running = Number(value) === 0;
                return (
                    <span className={`task-state ${running ? "running" : "paused"}`}>
                        <i/>{statusData[String(value)] || (running ? "运行" : "暂停")}
                    </span>
                );
            },
        },
        {
            title: "最近执行",
            dataIndex: "run_time",
            key: "run_time",
            width: 110,
            className: "runtime-cell",
            sorter: true,
            sortOrder: sort === "run_time" ? (order === "asc" ? "ascend" : "descend") : null,
            render: (value: any) => {
                const runTime = formatRunTime(value);
                return (
                    <Tooltip title={runTime}>
                        <span className="runtime-value">
                            {runTime === "尚未运行" ? runTime : (ago(runTime) || runTime)}
                        </span>
                    </Tooltip>
                );
            },
        },
        {
            title: "操作",
            key: "action",
            width: 80,
            fixed: "right",
            className: "action-cell",
            render: (_: any, record: any) => {
                const id = String(record.id);
                const running = Number(record.status) === 0;
                const items = [
                    {key: "edit", label: "编辑", onClick: () => onEdit(record)},
                    {key: "log", label: "日志", onClick: () => onLog(record)},
                    {key: "run", label: "执行", onClick: () => onAction("run", record)},
                    running
                        ? {key: "stop", label: "暂停", onClick: () => onAction("stop", record)}
                        : {key: "restore", label: "恢复", onClick: () => onAction("restore", record)},
                    {type: "divider" as const},
                    {key: "delete", label: "删除", danger: true, onClick: () => onAction("delete", record)},
                ];
                return (
                    <Dropdown
                        open={openRowId === id}
                        onOpenChange={(open) => setOpenRowId(open ? id : "")}
                        menu={{items}}
                        trigger={["click"]}
                        placement="bottom"
                        arrow>
                        <Button size="small" loading={operating.has(id)}>操作</Button>
                    </Dropdown>
                );
            },
        },
    ], [execTypeData, onAction, onEdit, onLog, openRowId, operating, order, sort, statusData, typeData]);

    return (
        <DataTable
            className={className}
            columns={columns}
            dataSource={tasks}
            loading={loading}
            onSortChange={onSortChange}
            rowKey={(record) => String(record.id)}
            rowClassName={(record) => openRowId === String(record.id) ? "action-menu-open" : ""}/>
    );
};

export default React.memo(Table);
