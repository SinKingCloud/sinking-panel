import React, {useState} from "react";
import {Button, DatePicker, Dropdown, Tooltip} from "antd";
import type {MenuProps} from "antd";
import dayjs from "dayjs";
import {Icon} from "sinking-antd";
import {TableAction, TableFilter, useTableStyles} from "./parts";
import type {TableDateFilterProps, TableToolbarItem} from "./types";

const datePresets = [
    {value: "all", label: "全部时间", days: 0},
    {value: "day", label: "最近一天", days: 1},
    {value: "three-days", label: "最近三天", days: 3},
    {value: "week", label: "最近一周", days: 7},
    {value: "month", label: "最近一月", days: 30},
    {value: "custom", label: "自定义", days: 0},
];

const datePopupAlign = {overflow: {adjustX: true, adjustY: true, shiftX: true}};

export const TableDateFilter = ({value, onChange, ariaLabel = "时间筛选", disabled, presets = datePresets}: TableDateFilterProps) => {
    const {styles} = useTableStyles();
    const [open, setOpen] = useState(false);
    const today = dayjs();
    const preset = !value ? "all" : presets.find((item) => item.days > 0
        && value[1].isSame(today, "day")
        && value[0].isSame(today.subtract(item.days - 1, "day"), "day"))?.value || "custom";
    const title = value?.[0] && value?.[1]
        ? `${value[0].format("YYYY-MM-DD")} 至 ${value[1].format("YYYY-MM-DD")}`
        : "全部时间";

    return (
        <div className={styles.dateFilter}>
            <TableFilter
                value={preset}
                options={presets}
                disabled={disabled}
                icon="CalendarOutlined"
                ariaLabel={`${ariaLabel}：${title}`}
                title={title}
                onChange={(next) => {
                    if (next === "custom") {
                        setOpen(true);
                        return;
                    }
                    const days = presets.find((item) => item.value === next)?.days;
                    const end = dayjs();
                    onChange(days ? [end.subtract(days - 1, "day"), end] : null);
                }}/>
            <DatePicker.RangePicker
                className={styles.datePickerAnchor}
                classNames={{popup: {root: styles.datePickerPopup}}}
                value={value}
                open={open}
                disabled={disabled}
                tabIndex={-1}
                allowClear
                inputReadOnly
                placement="bottomRight"
                popupAlign={datePopupAlign}
                format="YYYY-MM-DD"
                placeholder={["开始日期", "结束日期"]}
                onOpenChange={setOpen}
                onChange={(range) => {
                    if (range && (!range[0] || !range[1])) return;
                    onChange(range ? [range[0]!, range[1]!] : null);
                    setOpen(false);
                }}/>
        </div>
    );
};

const TableIconCommand = ({item}: {item: Omit<Extract<TableToolbarItem, {type: "icon"}>, "key" | "hidden">}) => {
    const {styles} = useTableStyles();
    const {type, icon, ariaLabel, tooltip, suffixIcon, menu, menuClassName = "", placement = "bottomRight", ...props} = item;
    const button = <Button {...props} type="text" aria-label={ariaLabel}
        className={`${styles.toolbarIconButton} ${props.className || ""}`}
        icon={<Icon type={icon}/>}>{suffixIcon && <Icon type={suffixIcon}/>}</Button>;
    return <Tooltip title={tooltip || ariaLabel}>
        {menu ? <Dropdown trigger={["click"]} placement={placement}
            classNames={{root: `${styles.toolbarDropdown} ${menuClassName}`}} menu={menu}>{button}</Dropdown> : button}
    </Tooltip>;
};

export const TableCommand = ({item}: {item: TableToolbarItem}) => {
    const {key: _, hidden, ...command} = item;
    if (hidden) return null;
    switch (command.type) {
        case "button": {
            const {type, ...props} = command;
            return <TableAction {...props}/>;
        }
        case "filter": {
            const {type, ...props} = command;
            return <TableFilter {...props}/>;
        }
        case "date": {
            const {type, ...props} = command;
            return <TableDateFilter {...props}/>;
        }
        case "icon": return <TableIconCommand item={command}/>;
        case "custom": return <>{command.render}</>;
    }
};

export const TableCommands = ({items = []}: {items?: TableToolbarItem[]}) => (
    <>{items.map((item) => <TableCommand key={item.key} item={item}/>)}</>
);

export const TableSelectionAction = ({count, items, onClear, clearDisabled}: {
    count: number;
    items: TableToolbarItem[];
    onClear: () => void;
    clearDisabled?: boolean;
}) => {
    if (!count) return null;
    const menuItems: MenuProps["items"] = items.filter((item) => !item.hidden).map((item) => {
        if (item.type !== "button") return {key: item.key, label: <TableCommand item={item}/>};
        return {
            key: item.key,
            label: item.label,
            icon: item.icon ? <Icon type={item.icon}/> : undefined,
            danger: item.danger,
            disabled: item.disabled || Boolean(item.loading),
            children: item.menu?.items,
            onClick: ({domEvent}) => item.onClick?.(domEvent as React.MouseEvent<HTMLElement>),
        };
    });
    if (menuItems.length) menuItems.push({type: "divider"});
    menuItems.push({key: "__clear_selection__", label: "取消选择", icon: <Icon type="CloseOutlined"/>, disabled: clearDisabled, onClick: onClear});
    return <TableAction
        className="table-selection-action"
        label={`已选 ${count} 项`}
        icon="CheckSquareOutlined"
        suffixIcon="DownOutlined"
        aria-label={`批量操作，已选 ${count} 项`}
        menu={{
            items: menuItems,
            onClick: (info) => {
                const parent = items.find((item) => item.type === "button" && item.menu && info.keyPath.includes(String(item.key)));
                if (parent?.type === "button") parent.menu?.onClick?.(info);
            },
        }}/>;
};
