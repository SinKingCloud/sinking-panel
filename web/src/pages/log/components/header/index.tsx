import React from "react";
import {DatePicker} from "antd";
import dayjs from "dayjs";
import {TableFilter, TableHero, TableToolbar} from "@/pages/components/table";

const Header = ({
    dateFilterClassName,
    datePickerAnchorClassName,
    datePickerPopupClassName,
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
}: any) => {
    const typeOptions = React.useMemo(() => [
        {label: "全部类型", value: "all"},
        ...Object.entries(typeData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [typeData]);
    const dateOptions = React.useMemo(() => [
        {value: "all", label: "全部时间", days: 0},
        {value: "day", label: "最近一天", days: 1},
        {value: "three-days", label: "最近三天", days: 3},
        {value: "week", label: "最近一周", days: 7},
        {value: "month", label: "最近一月", days: 30},
        {value: "custom", label: "自定义", days: 0},
    ], []);
    const [dateOpen, setDateOpen] = React.useState(false);
    const [datePreset, setDatePreset] = React.useState("all");
    const dateTitle = dateRange?.[0] && dateRange?.[1]
        ? `${dateRange[0].format("YYYY-MM-DD")} 至 ${dateRange[1].format("YYYY-MM-DD")}`
        : "全部时间";

    return (
        <>
            <TableHero
                eyebrow="OPERATION AUDIT"
                title="操作日志"
                action={{label: "清理日志", onClick: onClear}}/>
            <TableToolbar
                search={{
                    value: keyword,
                    placeholder: "搜索 IP、归属地、标题或内容",
                    maxLength: 200,
                    onChange: onKeywordChange,
                    onSearch,
                }}
                refresh={{loading: refreshing, ariaLabel: "刷新操作日志列表", onClick: onRefresh}}>
                <TableFilter
                    value={type || "all"}
                    options={typeOptions}
                    icon="TagsOutlined"
                    ariaLabel="日志类型筛选"
                    onChange={(value) => onTypeChange(value === "all" ? "" : value)}/>
                <div className={dateFilterClassName}>
                    <TableFilter
                        value={datePreset}
                        options={dateOptions}
                        icon="CalendarOutlined"
                        ariaLabel={`时间筛选：${dateTitle}`}
                        title={dateTitle}
                        onChange={(value) => {
                            if (value === "custom") {
                                setDateOpen(true);
                                return;
                            }
                            const preset = dateOptions.find((item) => item.value === value);
                            setDatePreset(value);
                            if (!preset?.days) {
                                onDateRangeChange(null);
                                return;
                            }
                            const end = dayjs();
                            onDateRangeChange([end.subtract(preset.days - 1, "day"), end]);
                        }}/>
                    <DatePicker.RangePicker
                        className={datePickerAnchorClassName}
                        classNames={{popup: {root: datePickerPopupClassName}}}
                        value={dateRange}
                        open={dateOpen}
                        tabIndex={-1}
                        allowClear
                        inputReadOnly
                        placement="bottomRight"
                        format="YYYY-MM-DD"
                        placeholder={["开始日期", "结束日期"]}
                        onOpenChange={setDateOpen}
                        onChange={(value) => {
                            setDatePreset(value ? "custom" : "all");
                            onDateRangeChange(value);
                            setDateOpen(false);
                        }}/>
                </div>
            </TableToolbar>
        </>
    );
};

export default React.memo(Header);
