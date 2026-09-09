import React from "react";
import type {TableProps} from "@/components/table";

const useLogHeader = ({
    keyword,
    type,
    dateRange,
    typeData,
    refreshing,
    onKeywordChange,
    onSearch,
    onTypeChange,
    onDateRangeChange,
    onRefresh,
    onClear,
}: any): Pick<TableProps, "hero" | "toolbar"> => {
    const typeOptions = React.useMemo(() => [
        {label: "全部类型", value: "all"},
        ...Object.entries(typeData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [typeData]);

    return {
        hero: {
            eyebrow: "OPERATION AUDIT",
            title: "操作日志",
            action: {label: "清理日志", onClick: onClear},
        },
        toolbar: {
            search: {
                value: keyword,
                placeholder: "搜索 IP、归属地、标题或内容",
                maxLength: 200,
                onChange: onKeywordChange,
                onSearch,
            },
            refresh: {loading: refreshing, ariaLabel: "刷新操作日志列表", onClick: onRefresh},
            actions: [
                {
                    key: "type",
                    type: "filter",
                    value: type || "all",
                    options: typeOptions,
                    icon: "TagsOutlined",
                    ariaLabel: "日志类型筛选",
                    onChange: (value) => onTypeChange(value === "all" ? "" : value),
                },
                {
                    key: "date",
                    type: "date",
                    value: dateRange,
                    ariaLabel: "操作时间筛选",
                    onChange: onDateRangeChange,
                },
            ],
        },
    };
};

export default useLogHeader;
