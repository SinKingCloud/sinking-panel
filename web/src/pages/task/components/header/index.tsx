import React from "react";
import type {TableProps} from "@/components/table";

const useTaskHeader = ({
    keyword,
    typeId,
    execType,
    status,
    typeData,
    typeItems,
    execTypeData,
    statusData,
    refreshing,
    onKeywordChange,
    onSearch,
    onTypeIdChange,
    onManageTypes,
    onExecTypeChange,
    onStatusChange,
    onRefresh,
    onCreate,
}: any): Pick<TableProps, "hero" | "toolbar"> => {
    const typeOptions = React.useMemo(() => [
        {label: "全部分类", value: "0"},
        ...(typeItems
            ? typeItems.map((item: any) => ({value: String(item.id), label: String(item.name)}))
            : Object.entries(typeData || {})
                .filter(([value]) => value !== "0")
                .map(([value, label]) => ({value, label: String(label)}))),
    ], [typeData, typeItems]);
    const execTypeOptions = React.useMemo(() => [
        {label: "全部方式", value: ""},
        ...Object.entries(execTypeData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [execTypeData]);
    const statusOptions = React.useMemo(() => [
        {label: "全部状态", value: ""},
        ...Object.entries(statusData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [statusData]);
    const activeType = typeOptions.find((item) => item.value === typeId)?.label || typeOptions[0].label;
    const activeExecType = execTypeOptions.find((item) => item.value === execType)?.label || execTypeOptions[0].label;
    const activeStatus = statusOptions.find((item) => item.value === status)?.label || statusOptions[0].label;

    return {
        hero: {
            eyebrow: "TASK SCHEDULER",
            title: "计划任务",
            action: {label: "添加任务", onClick: onCreate},
        },
        toolbar: {
            search: {
                value: keyword,
                placeholder: "搜索任务名称",
                onChange: onKeywordChange,
                onSearch,
            },
            refresh: {loading: refreshing, ariaLabel: "刷新计划任务列表", onClick: onRefresh},
            actions: [
                {
                    key: "category",
                    type: "filter",
                    value: typeId || "0",
                    options: typeOptions,
                    icon: "FolderOutlined",
                    ariaLabel: "任务分类筛选",
                    title: activeType,
                    command: {label: "分类管理", icon: "SettingOutlined", onClick: onManageTypes},
                    onChange: onTypeIdChange,
                },
                {
                    key: "exec-type",
                    type: "filter",
                    value: execType || "all",
                    options: execTypeOptions.map(({label, value}) => ({label, value: value || "all"})),
                    icon: "CodeOutlined",
                    ariaLabel: "执行方式筛选",
                    title: activeExecType,
                    onChange: (value) => onExecTypeChange(value === "all" ? "" : value),
                },
                {
                    key: "status",
                    type: "filter",
                    value: status || "all",
                    options: statusOptions.map(({label, value}) => ({label, value: value || "all"})),
                    icon: "FlagOutlined",
                    ariaLabel: "任务状态筛选",
                    title: activeStatus,
                    onChange: (value) => onStatusChange(value === "all" ? "" : value),
                },
            ],
        },
    };
};

export default useTaskHeader;
