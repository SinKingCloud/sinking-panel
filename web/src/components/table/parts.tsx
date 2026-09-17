import React, {useCallback, useMemo} from "react";
import {
    Button,
    Card,
    ConfigProvider,
    Dropdown,
    Input,
    Pagination,
    Select,
    Table as AntTable,
    Tooltip,
} from "antd";
import type {ButtonProps, MenuProps, PaginationProps, TableProps} from "antd";
import {Icon, useTheme} from "sinking-antd";
import HeroGraphic from "./hero-graphic";
import useStyles from "./styles";

export const useTableStyles = () => {
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
    affixed?: boolean;
    onChange: (page: number, pageSize: number) => void;
}

export const TablePagination = ({
    page,
    pageSize,
    total,
    disabled,
    unit = "条",
    pageSizeOptions = [10, 20, 50, 100],
    affixed = false,
    onChange,
}: PageTablePaginationProps) => {
    const {styles} = useTableStyles();
    if (total <= 0) {
        return null;
    }

    return (
        <Card className={`${styles.paginationCard} ${affixed ? "ui-table-pagination-affixed" : ""}`} variant="borderless">
            <ConfigProvider input={{variant: "filled"}}>
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
            </ConfigProvider>
        </Card>
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
    eyebrow?: React.ReactNode;
    action?: TableHeroAction;
    background?: boolean | React.CSSProperties;
    graphic?: false | "task" | "setting";
}

export const TableHero = React.memo(({title, eyebrow, action, background = true, graphic = "task"}: TableHeroProps) => {
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
        <section className={`${styles.hero} ${background === false ? "plain-background" : ""} ${action ? "" : "without-action"}`}
                 style={typeof background === "object" ? background : undefined}>
            <div className="hero-copy">
                {eyebrow && <div className="eyebrow"><span className="status-dot"/>{eyebrow}</div>}
                <h1>{title}</h1>
            </div>
            {background !== false && graphic !== false && <div className="hero-visual"><HeroGraphic variant={graphic}/></div>}
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
    className?: string;
    style?: React.CSSProperties;
    ariaLabel?: string;
    maxLength?: number;
    allowClear?: boolean;
    fields?: {value: string; label: React.ReactNode; placeholder?: string; maxLength?: number}[];
    field?: string;
    tooltip?: React.ReactNode;
    onChange: (value: string) => void;
    onFieldChange?: (field: string) => void;
    onSearch?: (value: string, field?: string) => void;
    onBlur?: (value: string) => void;
}

export interface TableRefreshProps {
    loading?: boolean;
    disabled?: boolean;
    ariaLabel?: string;
    tooltip?: React.ReactNode;
    onClick: () => void;
}

export interface TableToolbarProps {
    left?: React.ReactNode;
    search?: TableSearchProps;
    refresh?: TableRefreshProps;
    children?: React.ReactNode;
    extra?: React.ReactNode;
}

export const TableSearch = React.memo((search: TableSearchProps) => {
    const {styles} = useTableStyles();
    const field = search.fields?.find((item) => item.value === search.field) || search.fields?.[0];
    const className = `${styles.searchBox} ui-table-search ${search.className || ""}`;
    const onBlur: React.FocusEventHandler<HTMLElement> = (event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) search.onBlur?.(search.value);
    };
    const searchInput = <Input
        className={field ? "search-input" : className}
        style={field ? undefined : search.style}
        variant={field ? "borderless" : "filled"}
        styles={field ? {root: {outline: "none", boxShadow: "none"}, input: {outline: "none", boxShadow: "none"}} : undefined}
        value={search.value}
        aria-label={search.ariaLabel}
        allowClear={search.allowClear !== false}
        maxLength={field?.maxLength ?? search.maxLength}
        prefix={field ? undefined : <Icon type="SearchOutlined"/>}
        placeholder={field?.placeholder ?? search.placeholder}
        onChange={(event) => search.onChange(event.target.value)}
        onBlur={field ? undefined : onBlur}
        onPressEnter={search.onSearch
            ? (event) => search.onSearch?.(event.currentTarget.value, field?.value)
            : undefined}/>;
    const searchControl = field ? <div className={`${className} ${styles.searchGroup}`} style={search.style} onBlur={onBlur}>
        <Select
            className="search-field"
            classNames={{popup: {root: styles.pageSizeDropdown}}}
            styles={{root: {outline: "none", boxShadow: "none", border: 0}}}
            variant="borderless"
            aria-label="搜索字段"
            title=""
            value={field.value}
            showSearch={false}
            popupMatchSelectWidth={false}
            options={search.fields?.map(({value, label}) => ({value, label, title: ""}))}
            labelRender={({label}) => <span className="search-field-label">{label}</span>}
            onChange={search.onFieldChange}/>
        {searchInput}
    </div> : searchInput;
    return search.tooltip ? <Tooltip title={search.tooltip}>{searchControl}</Tooltip> : searchControl;
});

export const TableToolbar = React.memo(({left, search, refresh, children, extra}: TableToolbarProps) => {
    const {styles} = useTableStyles();
    const hasLeft = left !== undefined && left !== null && left !== false;
    const refreshButton = refresh && <Button
        className={styles.toolbarIconButton}
        type="text"
        aria-label={refresh.ariaLabel || "刷新列表"}
        loading={refresh.loading}
        disabled={refresh.disabled}
        icon={<Icon type="ReloadOutlined"/>}
        onClick={refresh.onClick}/>;
    return (
        <div className={`${styles.commandBar} ui-table-toolbar ${!search && !hasLeft ? "without-search" : ""} ${hasLeft && !search ? "with-content" : ""}`}>
            {hasLeft && <div className="command-content">{left}</div>}
            {search && <TableSearch {...search}/>}
            <div className="command-actions">
                {hasLeft && extra}
                <div className="command-actions-scroll">{children}</div>
                {!hasLeft && extra}
                {refresh?.tooltip ? <Tooltip title={refresh.tooltip}>{refreshButton}</Tooltip> : refreshButton}
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
    tooltip?: React.ReactNode;
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
    tooltip,
    disabled,
    matchWidth,
    command,
    onChange,
}: TableFilterProps) => {
    const {styles} = useTableStyles();
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

    const dropdown = (
        <Dropdown
            disabled={disabled}
            trigger={["click"]}
            placement="bottomRight"
            classNames={{root: styles.toolbarDropdown}}
            menu={{
                selectable: true,
                selectedKeys: [value],
                items,
                onClick: ({key}) => key === commandKey ? command?.onClick() : onChange(key),
            }}>
            <button
                className={`${styles.toolbarTrigger} ${matchWidth ? "match-width" : ""}`}
                type="button"
                disabled={disabled}
                aria-label={ariaLabel}>
                <Icon type={icon} className="marker"/>
                {matchWidth ? (
                    <span className="value-stack" title={title}>
                        {options.map((item) => (
                            <span className="value measure" aria-hidden="true" key={item.value}>{item.label}</span>
                        ))}
                        <span className="value active">{active?.label}</span>
                    </span>
                ) : (
                    <span className="value" title={title}>{active?.label}</span>
                )}
                <Icon type="DownOutlined" className="arrow"/>
            </button>
        </Dropdown>
    );
    return tooltip ? <Tooltip title={tooltip}>{dropdown}</Tooltip> : dropdown;
});

export interface TableActionProps extends Omit<ButtonProps, "icon" | "children" | "type"> {
    label: React.ReactNode;
    icon?: string;
    suffixIcon?: string;
    tooltip?: React.ReactNode;
    menu?: MenuProps;
    menuClassName?: string;
    placement?: "bottom" | "bottomLeft" | "bottomRight" | "top" | "topLeft" | "topRight";
}

export const TableAction = React.memo(({
    label,
    icon,
    suffixIcon,
    tooltip,
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
    const control = menu ? (
        <Dropdown
            trigger={["click"]}
            placement={placement}
            classNames={{root: `${styles.toolbarDropdown} ${menuClassName}`}}
            menu={menu}>
            {button}
        </Dropdown>
    ) : button;
    return tooltip ? <Tooltip title={tooltip}>{control}</Tooltip> : control;
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
        onSortChange(Array.isArray(field) ? field.join(".") : typeof field === "string" ? field : "", current?.order || undefined);
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
