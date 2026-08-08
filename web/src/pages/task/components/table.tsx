import React, {useMemo, useState} from "react";
import {Button, Dropdown, Table as AntTable, Tooltip} from "antd";
import {ago} from "@/utils/time";
import {describeTaskSchedule} from "../utils";

const formatRunTime = (value: any) => {
    const time = String(value || "");
    return !time || time.startsWith("0001-01-01") ? "尚未运行" : time;
};

const Table = ({
    className,
    tasks,
    loading,
    typeData,
    statusData,
    operating,
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
                    <button type="button" onClick={() => onEdit(record)}>
                        {record.name || "未命名任务"}
                    </button>
                </div>
            ),
        },
        {
            title: "类型",
            dataIndex: "type",
            key: "type",
            width: 90,
            className: "type-cell",
            render: (value: any) => (
                <span className="task-type">{typeData[String(value)] || "-"}</span>
            ),
        },
        {
            title: "执行周期",
            dataIndex: "spec",
            key: "spec",
            width: 220,
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
            width: 120,
            className: "runtime-cell",
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
    ], [onAction, onEdit, onLog, openRowId, operating, statusData, typeData]);

    return (
        <AntTable
            className={className}
            columns={columns}
            dataSource={tasks}
            loading={loading}
            pagination={false}
            rowKey={(record) => String(record.id)}
            rowClassName={(record) => openRowId === String(record.id) ? "action-menu-open" : ""}
            scroll={{x: 790}}
            tableLayout="fixed"
            size="middle"
        />
    );
};

export default React.memo(Table);
