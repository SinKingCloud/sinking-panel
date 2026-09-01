import React, {useCallback, useLayoutEffect, useMemo, useRef, useState} from "react";
import {
    Button,
    Card,
    Empty,
    Input,
    Pagination,
    Table as AntTable,
    Tooltip,
} from "antd";
import type {ButtonProps, MenuProps, PaginationProps, TableProps} from "antd";
import {Icon, useTheme} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import HeroGraphic from "@/pages/components/hero-graphic";
import useStyles from "./styles";

const useTableStyles = () => {
    const theme = useTheme();
    return useStyles({
        compact: Boolean(theme?.isCompactTheme?.()),
        dark: Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.()),
    });
};

export interface PageTablePaginationProps {
    page: number;
    pageSize: number;
    total: number;
    disabled?: boolean;
    unit?: string;
    pageSizeOptions?: number[];
    onChange: (page: number, pageSize: number) => void;
}

export interface PageTableProps {
    header: React.ReactNode;
    children: React.ReactNode;
    empty?: boolean;
    emptyContent?: React.ReactNode;
    pagination?: PageTablePaginationProps;
    ariaLabel?: string;
}

const TablePagination = ({
    page,
    pageSize,
    total,
    disabled,
    unit = "项",
    pageSizeOptions = [10, 20, 50, 100],
    onChange,
}: PageTablePaginationProps) => {
    const {styles} = useTableStyles();
    if (total <= 0) {
        return null;
    }

    return (
        <Card className={styles.paginationCard} variant="borderless">
            <Pagination
                className={styles.pagination}
                size="small"
                current={page}
                pageSize={pageSize}
                total={total}
                disabled={disabled}
                align="center"
                responsive
                showQuickJumper
                showLessItems
                showSizeChanger={{
                    showSearch: false,
                    variant: "filled",
                    size: "small",
                    className: styles.pageSizeSelect,
                    classNames: {popup: {root: styles.pageSizeDropdown}},
                } as PaginationProps["showSizeChanger"]}
                pageSizeOptions={pageSizeOptions}
                showTotal={(nextTotal, range) => `第 ${range[0]}-${range[1]} ${unit} / 共 ${nextTotal} ${unit}`}
                onChange={onChange}/>
        </Card>
    );
};

const PageTable = ({
    header,
    children,
    empty = false,
    emptyContent,
    pagination,
    ariaLabel,
}: PageTableProps) => {
    const {styles} = useTableStyles();
    return (
        <div className={styles.page}>
            <section className={styles.workspace} aria-label={ariaLabel}>
                {header}
                <div className={styles.dataPanel}>
                    {empty ? (
                        <div className={styles.state}>
                            {emptyContent || <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}
                        </div>
                    ) : children}
                </div>
            </section>
            {pagination && <TablePagination {...pagination}/>}
        </div>
    );
};

export interface TableHeroAction {
    label: React.ReactNode;
    ariaLabel?: string;
    icon?: string;
    suffixIcon?: string;
    disabled?: boolean;
    menu?: MenuProps;
    onClick?: () => void;
}

export interface TableHeroProps {
    title: React.ReactNode;
    eyebrow: React.ReactNode;
    action?: TableHeroAction;
}

export const TableHero = React.memo(({title, eyebrow, action}: TableHeroProps) => {
    const {styles} = useTableStyles();
    const button = action ? (
        <button
            className="create-button"
            type="button"
            aria-label={action.ariaLabel}
            disabled={action.disabled}
            onClick={action.menu ? undefined : action.onClick}>
            {action.icon && <Icon type={action.icon}/>}
            <span>{action.label}</span>
            {action.suffixIcon && <Icon type={action.suffixIcon} className="suffix-icon"/>}
        </button>
    ) : null;

    return (
        <section className={styles.hero}>
            <div className="hero-copy">
                <div className="eyebrow"><span className="status-dot"/>{eyebrow}</div>
                <h1>{title}</h1>
            </div>
            <div className="hero-visual"><HeroGraphic variant="task"/></div>
            {action?.menu && button ? (
                <Dropdown
                    trigger={["click"]}
                    placement="bottomRight"
                    classNames={{root: styles.toolbarDropdown}}
                    menu={action.menu}>
                    {button}
                </Dropdown>
            ) : button}
        </section>
    );
});

export interface TableSearchProps {
    value: string;
    placeholder: string;
    ariaLabel?: string;
    maxLength?: number;
    allowClear?: boolean;
    onChange: (value: string) => void;
    onSearch?: (value: string) => void;
}

export interface TableRefreshProps {
    loading?: boolean;
    disabled?: boolean;
    ariaLabel?: string;
    tooltip?: string;
    onClick: () => void;
}

export interface TableToolbarProps {
    search?: TableSearchProps;
    refresh?: TableRefreshProps;
    children?: React.ReactNode;
}

export const TableToolbar = React.memo(({search, refresh, children}: TableToolbarProps) => {
    const {styles} = useTableStyles();
    return (
        <div className={styles.commandBar}>
            {search && (
                <Input
                    className={styles.searchBox}
                    value={search.value}
                    aria-label={search.ariaLabel}
                    allowClear={search.allowClear !== false}
                    maxLength={search.maxLength}
                    prefix={<Icon type="SearchOutlined"/>}
                    placeholder={search.placeholder}
                    onChange={(event) => search.onChange(event.target.value)}
                    onPressEnter={search.onSearch
                        ? (event) => search.onSearch?.(event.currentTarget.value)
                        : undefined}/>
            )}
            <div className="command-actions">
                <div className="command-actions-scroll">{children}</div>
                {refresh && (
                    <Tooltip title={refresh.tooltip || "刷新列表"}>
                        <Button
                            className={styles.toolbarIconButton}
                            type="text"
                            aria-label={refresh.ariaLabel || "刷新列表"}
                            loading={refresh.loading}
                            disabled={refresh.disabled}
                            icon={<Icon type="ReloadOutlined"/>}
                            onClick={refresh.onClick}/>
                    </Tooltip>
                )}
            </div>
        </div>
    );
});

export interface TableFilterOption {
    value: string;
    label: React.ReactNode;
}

export interface TableFilterCommand {
    label: React.ReactNode;
    icon?: string;
    onClick: () => void;
}

export interface TableFilterProps {
    value: string;
    options: TableFilterOption[];
    icon: string;
    ariaLabel: string;
    title?: string;
    disabled?: boolean;
    matchWidth?: boolean;
    command?: TableFilterCommand;
    onChange: (value: string) => void;
}

export const TableFilter = React.memo(({
    value,
    options,
    icon,
    ariaLabel,
    title,
    disabled,
    matchWidth,
    command,
    onChange,
}: TableFilterProps) => {
    const {styles} = useTableStyles();
    const triggerRef = useRef<HTMLButtonElement>(null);
    const [width, setWidth] = useState(0);
    const commandKey = "__table_filter_command__";
    const active = options.find((item) => item.value === value) || options[0];
    const items = useMemo<MenuProps["items"]>(() => [
        ...options.map((item) => ({key: item.value, label: item.label})),
        ...(command ? [
            {type: "divider" as const},
            {
                key: commandKey,
                label: command.label,
                icon: command.icon ? <Icon type={command.icon}/> : undefined,
            },
        ] : []),
    ], [command, options]);

    useLayoutEffect(() => {
        if (!matchWidth) {
            return;
        }
        const update = () => setWidth(triggerRef.current?.offsetWidth || 0);
        update();
        if (typeof ResizeObserver === "undefined") {
            window.addEventListener("resize", update);
            return () => window.removeEventListener("resize", update);
        }
        const observer = new ResizeObserver(update);
        if (triggerRef.current) observer.observe(triggerRef.current);
        return () => observer.disconnect();
    }, [active?.label, matchWidth, options]);

    return (
        <Dropdown
            disabled={disabled}
            trigger={["click"]}
            placement="bottomRight"
            classNames={{root: styles.toolbarDropdown}}
            styles={{root: matchWidth && width ? {minWidth: width} : undefined}}
            menu={{
                selectable: true,
                selectedKeys: [value],
                items,
                onClick: ({key}) => key === commandKey ? command?.onClick() : onChange(key),
            }}>
            <button
                ref={triggerRef}
                className={`${styles.toolbarTrigger} ${matchWidth ? "match-width" : ""}`}
                type="button"
                disabled={disabled}
                aria-label={ariaLabel}>
                <Icon type={icon} className="marker"/>
                {matchWidth ? (
                    <span className="value-stack" title={title || String(active?.label || "")}>
                        {options.map((item) => (
                            <span className="value measure" aria-hidden="true" key={item.value}>{item.label}</span>
                        ))}
                        <span className="value active">{active?.label}</span>
                    </span>
                ) : (
                    <span className="value" title={title || String(active?.label || "")}>{active?.label}</span>
                )}
                <Icon type="DownOutlined" className="arrow"/>
            </button>
        </Dropdown>
    );
});

export interface TableActionProps extends Omit<ButtonProps, "icon" | "children" | "type"> {
    label: React.ReactNode;
    icon?: string;
    suffixIcon?: string;
    menu?: MenuProps;
    menuClassName?: string;
    placement?: "bottom" | "bottomLeft" | "bottomRight" | "top" | "topLeft" | "topRight";
}

export const TableAction = React.memo(({
    label,
    icon,
    suffixIcon,
    menu,
    menuClassName = "",
    placement = "bottomRight",
    className = "",
    ...buttonProps
}: TableActionProps) => {
    const {styles} = useTableStyles();
    const button = (
        <Button
            {...buttonProps}
            type="text"
            className={`${styles.toolbarAction} ${className}`}
            icon={icon ? <Icon type={icon}/> : undefined}>
            <span className="value">{label}</span>
            {suffixIcon && <Icon type={suffixIcon} className="arrow"/>}
        </Button>
    );
    if (!menu) {
        return button;
    }
    return (
        <Dropdown
            trigger={["click"]}
            placement={placement}
            classNames={{root: `${styles.toolbarDropdown} ${menuClassName}`}}
            menu={menu}>
            {button}
        </Dropdown>
    );
});

export type DataTableProps<RecordType extends object = any> = Omit<TableProps<RecordType>, "pagination"> & {
    onSortChange?: (field: string, order?: "ascend" | "descend") => void;
};

export const DataTable = <RecordType extends object = any>({
    className = "",
    scroll,
    showSorterTooltip = false,
    tableLayout = "fixed",
    size = "middle",
    onChange,
    onSortChange,
    ...props
}: DataTableProps<RecordType>) => {
    const {styles} = useTableStyles();
    const change = useCallback<NonNullable<TableProps<RecordType>["onChange"]>>((pagination, filters, sorter, extra) => {
        onChange?.(pagination, filters, sorter, extra);
        if (extra.action !== "sort" || !onSortChange) {
            return;
        }
        const current = Array.isArray(sorter) ? sorter[0] : sorter;
        const field = current?.field || current?.columnKey;
        onSortChange(typeof field === "string" ? field : "", current?.order || undefined);
    }, [onChange, onSortChange]);

    return (
        <AntTable<RecordType>
            {...props}
            className={`${styles.table} ${className}`}
            pagination={false}
            scroll={{x: "max-content", ...scroll}}
            showSorterTooltip={showSorterTooltip}
            tableLayout={tableLayout}
            size={size}
            onChange={change}/>
    );
};

export default React.memo(PageTable);
