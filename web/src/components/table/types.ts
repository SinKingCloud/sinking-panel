import type React from "react";
import type {TableProps as AntTableProps} from "antd";
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
    presets?: {value: string; label: React.ReactNode; days: number}[];
}

type CommandBase = {key: React.Key; hidden?: boolean};
export type TableToolbarItem = CommandBase & (
    | ({type: "button"} & TableActionProps)
    | ({type: "filter"} & TableFilterProps)
    | ({type: "date"} & TableDateFilterProps)
    | ({type: "icon"; icon: string; ariaLabel: string; tooltip?: React.ReactNode} & Omit<TableActionProps, "label">)
    | {type: "custom"; render: React.ReactNode}
);
export type TableActionItem = TableToolbarItem;

export interface TableToolbarConfig {
    left?: React.ReactNode;
    search?: Omit<TableSearchProps, "value" | "onChange" | "placeholder"> & {
        value?: string;
        placeholder?: string;
        onChange?: (value: string) => void;
        /** 设置后输入自动搜索；省略时通过回车或清空搜索。 */
        debounce?: number;
    };
    /** Table 和 ModalTable 共用，支持日期、下拉筛选、按钮、图标和自定义内容。 */
    actions?: TableToolbarItem[];
    refresh?: TableRefreshProps | boolean;
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
    page: number;
    pageSize: number;
    keyword: string;
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
    contentBar?: TableContentBarProps | false;
    rowSelection?: TableRowSelection<RecordType> | boolean;
    pagination?: Partial<PageTablePaginationProps> | false;
    /** 分页切换后是否滚动到表格顶部，默认保留当前滚动位置。 */
    scrollToTopOnPageChange?: boolean;
    /** 是否将页面底部分页栏固钉在视口底部，默认关闭。 */
    paginationAffix?: boolean;
    ariaLabel?: string;
    rootClassName?: string;
    empty?: boolean;
    emptyContent?: React.ReactNode;
    /** 仅用于包裹表格的业务上下文，例如文件右键菜单 Provider。 */
    tableRender?: (table: React.ReactNode) => React.ReactNode;
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
