import type React from "react";
import type {AffixProps, TableProps as AntTableProps} from "antd";
import type {Dayjs} from "dayjs";
import type {
    DataTableProps, PageTablePaginationProps, TableActionProps, TableFilterProps,
    TableHeroProps, TableRefreshProps, TableSearchProps,
} from "./parts";

export interface TableDateFilterProps {
    value: [Dayjs, Dayjs] | null;
    onChange: (value: [Dayjs, Dayjs] | null) => void;
    ariaLabel?: string;
    disabled?: boolean;
    tooltip?: React.ReactNode;
    presets?: {value: string; label: React.ReactNode; days: number}[];
}

type CommandBase = {key: React.Key; hidden?: boolean; tooltip?: React.ReactNode};
export type TableToolbarItem = CommandBase & (
    | ({type: "button"} & TableActionProps)
    | ({type: "filter"} & TableFilterProps)
    | ({type: "date"} & TableDateFilterProps)
    | ({type: "icon"; icon: string; ariaLabel: string} & Omit<TableActionProps, "label">)
    | {type: "custom"; render: React.ReactNode}
);
export type TableActionItem = TableToolbarItem;

export interface TableSearchConfig extends Omit<TableSearchProps, "value" | "onChange" | "placeholder"> {
    value?: string;
    defaultValue?: string;
    /** 单输入框模式的请求字段，默认 keyword；组合模式使用选中的 fields.value。 */
    name?: string;
    /** 组合搜索默认字段，省略时选择 fields 第一项。 */
    defaultField?: string;
    placeholder?: string;
    onChange?: (value: string) => void;
    /** 设置后输入自动搜索；省略时通过回车或清空搜索。 */
    debounce?: number;
}

export interface TableSelectFilterConfig {
    type: "select";
    name: string;
    label: string;
    icon?: string;
    hidden?: boolean;
    disabled?: boolean;
    tooltip?: React.ReactNode;
    options?: TableFilterProps["options"];
    valueEnum?: Record<string, React.ReactNode>;
    /** 默认提供“全部”选项，false 则不添加。 */
    allLabel?: React.ReactNode | false;
    defaultValue?: string;
    value?: string;
    onChange?: (value: string) => void;
    transform?: (value: string) => Record<string, unknown>;
}

export interface TableDateFilterConfig extends Omit<TableDateFilterProps, "value" | "onChange" | "ariaLabel"> {
    type: "date";
    name: string;
    label?: string;
    hidden?: boolean;
    /** 默认输出 name_start / name_end，包含所选日期的全天。 */
    fields?: [string, string];
    defaultValue?: [Dayjs, Dayjs] | null;
    value?: [Dayjs, Dayjs] | null;
    onChange?: (value: [Dayjs, Dayjs] | null) => void;
    transform?: (value: [Dayjs, Dayjs] | null) => Record<string, unknown>;
}

export type TableFilterConfig = TableSelectFilterConfig | TableDateFilterConfig;

export interface TableToolbarConfig {
    left?: React.ReactNode;
    search?: TableSearchConfig;
    /** Table 和 ModalTable 共用，支持日期、下拉筛选、按钮、图标和自定义内容。 */
    actions?: TableToolbarItem[];
    /** 请求模式默认显示；省略 onClick / loading 时使用组件内部状态。 */
    refresh?: (Omit<TableRefreshProps, "onClick"> & {onClick?: () => void}) | boolean;
}

export interface TableContentBarProps {
    content?: React.ReactNode;
    actions?: TableToolbarItem[];
    ariaLabel?: string;
}

export type TableRowSelection<RecordType extends object = any> = NonNullable<AntTableProps<RecordType>["rowSelection"]> & {
    actions?: TableToolbarItem[] | ((keys: React.Key[], rows: RecordType[]) => TableToolbarItem[]);
    onClear?: () => void;
    clearDisabled?: boolean;
};

export interface TableRequestParams extends Record<string, unknown> {
    current: number;
    pageSize: number;
}
export type TableSort = Record<string, "ascend" | "descend">;
export interface TableRequestResult<RecordType> {
    data: RecordType[];
    total: number;
    success?: boolean;
}

export interface TableProps<RecordType extends object = any> extends Omit<DataTableProps<RecordType>, "rowSelection"> {
    hero?: TableHeroProps | false;
    toolbar?: TableToolbarConfig | false;
    /** true 使用默认搜索框；配置对象可自定义字段、提示和防抖。 */
    search?: TableSearchConfig | boolean;
    /** 自动维护筛选值，变化后回到第一页并请求。 */
    filters?: TableFilterConfig[];
    contentBar?: TableContentBarProps | false;
    rowSelection?: TableRowSelection<RecordType> | boolean;
    pagination?: Partial<PageTablePaginationProps> | false;
    /** 分页切换后是否滚动到表格顶部，默认保留当前滚动位置。 */
    scrollToTopOnPageChange?: boolean;
    /** 开启分页固钉，或传入 Ant Design Affix 配置；默认关闭。 */
    paginationAffix?: boolean | Omit<AffixProps, "children">;
    ariaLabel?: string;
    rootClassName?: string;
    /** 外层容器样式，宽高和页面布局由调用方控制。 */
    rootStyle?: React.CSSProperties;
    empty?: boolean;
    emptyContent?: React.ReactNode;
    /** 仅用于包裹表格的业务上下文，例如文件右键菜单 Provider。 */
    tableRender?: (table: React.ReactNode) => React.ReactNode;
    /** 与 ProTable 一致，接收 current / pageSize 及查询条件，返回 {data, total, success?}。 */
    request?: (params: TableRequestParams, sort: TableSort) => Promise<TableRequestResult<RecordType>>;
    params?: Record<string, unknown>;
    defaultPage?: number;
    defaultPageSize?: number;
    defaultSort?: TableSort;
    manualRequest?: boolean;
    /** false 时暂停所有请求并忽略未完成的响应，适用于隐藏的面板。 */
    requestEnabled?: boolean;
    onLoad?: (data: RecordType[], result: TableRequestResult<RecordType>) => void;
    onRequestError?: (error: Error) => void;
    /** 受控 dataSource 模式下供 ref.reload / refresh:true 调用。 */
    onReload?: () => void | Promise<unknown>;
}

export interface TableRef<RecordType extends object = any> {
    reload: () => void;
    refreshTableData: () => void;
    resetTableData: () => void;
    getTableData: () => readonly RecordType[];
    getSelectedRowKeys: () => React.Key[];
    getSelectedRows: () => RecordType[];
    setSelectedRowKeys: (keys: React.Key[]) => void;
    clearSelectedRows: () => void;
    allSelectedRow: () => void;
    invertSelectedRow: () => void;
    scrollToTop: () => void;
}
