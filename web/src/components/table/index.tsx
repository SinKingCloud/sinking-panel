import React, {forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {Empty, Spin} from "antd";
import {DataTable, TableHero, TablePagination, TableToolbar, useTableStyles} from "./parts";
import {TableCommands, TableSelectionAction} from "./commands";
import useTableData from "./use-data";
import useTableSelection from "./use-selection";
import type {TableProps, TableRef} from "./types";

export type * from "./types";
export {DataTable, TableHero, TableToolbar, TableAction, TableFilter, TablePagination} from "./parts";
export type {DataTableProps, TableHeroProps, TableActionProps, TableFilterProps, TableSearchProps, TableRefreshProps, PageTablePaginationProps} from "./parts";
export {TableDateFilter} from "./commands";

const Table = <RecordType extends object = any>(props: TableProps<RecordType>, ref: React.ForwardedRef<TableRef<RecordType>>) => {
    const {
        hero, toolbar, contentBar, rowSelection, pagination, ariaLabel, rootClassName = "", empty, emptyContent,
        tableRender, request, params, defaultPage, defaultPageSize, defaultSort, manualRequest, requestEnabled,
        onLoad, onRequestError, onReload, onSortChange, dataSource, loading, rowKey = "id", ...tableProps
    } = props;
    const {styles} = useTableStyles();
    const rootRef = useRef<HTMLDivElement>(null);
    const data = useTableData(props);
    const selection = useTableSelection(data.data, rowKey, rowSelection);
    const selectedKeys = selection.keys;
    const scrollToTop = useCallback(() => {
        const panel = rootRef.current?.querySelector<HTMLElement>(".ui-table-data");
        panel?.scrollTo({top: 0});
        const modal = rootRef.current?.closest(".ant-modal-wrap");
        if (modal) modal.scrollTo({top: 0});
        else rootRef.current?.scrollIntoView({block: "start", inline: "nearest"});
    }, []);
    const searchConfig = toolbar ? toolbar.search : undefined;
    const [keyword, setKeyword] = useState("");
    const searchValue = searchConfig?.value ?? keyword;
    const submittedSearch = useRef(searchValue);
    const searchTimer = useRef<number | undefined>(undefined);
    const searchCallback = useRef((value: string) => {});
    searchCallback.current = (value) => {
        window.clearTimeout(searchTimer.current);
        submittedSearch.current = value;
        data.search(value);
        searchConfig?.onSearch?.(value);
    };
    useEffect(() => {
        if (!searchConfig || searchConfig.debounce === undefined || requestEnabled === false || submittedSearch.current === searchValue) return;
        searchTimer.current = window.setTimeout(() => searchCallback.current(searchValue), searchConfig.debounce);
        return () => window.clearTimeout(searchTimer.current);
    }, [searchValue, searchConfig?.debounce, requestEnabled]);

    useImperativeHandle(ref, () => ({
        reload: data.reload,
        refreshTableData: data.reload,
        resetTableData: () => {
            window.clearTimeout(searchTimer.current);
            submittedSearch.current = "";
            if (searchConfig?.value === undefined) setKeyword("");
            searchConfig?.onChange?.("");
            selection.clear();
            data.reset();
            scrollToTop();
        },
        getTableData: () => data.data,
        getSelectedRowKeys: selection.getKeys,
        getSelectedRows: selection.getRows,
        setSelectedRowKeys: selection.setKeys,
        clearSelectedRows: selection.clear,
        allSelectedRow: selection.selectAll,
        invertSelectedRow: selection.invert,
        scrollToTop,
    }));

    const selectionAction = <TableSelectionAction count={selectedKeys.length} items={selection.actions}
        onClear={selection.clear} clearDisabled={selection.clearDisabled}/>;
    const content = contentBar ? contentBar.content : undefined;
    const hasContent = content !== undefined && content !== null && content !== false;
    const hasContentActions = Boolean(contentBar && contentBar.actions?.some((item) => !item.hidden));
    const toolbarActions = [...(toolbar ? toolbar.actions || [] : []), ...(!hasContent && contentBar ? contentBar.actions || [] : [])];
    const showToolbar = Boolean(toolbar || toolbarActions.length || (!hasContent && selectedKeys.length));
    const refresh = toolbar && toolbar.refresh;
    const showEmpty = empty ?? data.data.length === 0;
    const spinning = typeof data.loading === "boolean" ? data.loading : Boolean(data.loading?.spinning);
    const table = showEmpty ? <div className={styles.state} role="status">
        {spinning ? <Spin/> : emptyContent ?? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}
    </div> : <DataTable<RecordType> {...tableProps} dataSource={data.data} loading={data.loading}
                            rowKey={rowKey} rowSelection={selection.rowSelection} onSortChange={data.changeSort}/>;

    return (
        <div ref={rootRef} className={`${styles.page} ui-table-root ${rootClassName}`}>
            <section className={`${styles.workspace} ui-table-workspace`} aria-label={ariaLabel}>
                {hero && <TableHero {...hero}/>}
                {showToolbar && <TableToolbar
                    left={toolbar ? toolbar.left : undefined}
                    search={searchConfig ? {
                        ...searchConfig,
                        value: searchValue,
                        placeholder: searchConfig.placeholder || "搜索",
                        onChange: (value) => {
                            if (searchConfig.value === undefined) setKeyword(value);
                            searchConfig.onChange?.(value);
                            if (!value && searchConfig.debounce === undefined) {
                                submittedSearch.current = "";
                                data.search("");
                            }
                        },
                        onSearch: (value) => {
                            searchCallback.current(value);
                            scrollToTop();
                        },
                    } : undefined}
                    refresh={refresh === true ? {loading: spinning, onClick: data.reload} : refresh || undefined}
                    extra={!hasContent && selectedKeys.length > 0 ? selectionAction : undefined}>
                    <TableCommands items={toolbarActions}/>
                </TableToolbar>}
                {hasContent && contentBar && <div className={`${styles.contentBar} ui-table-content-bar`} role="group" aria-label={contentBar.ariaLabel || "表格内容与操作"}>
                    <div className="table-bar-content">{content}</div>
                    {(selectedKeys.length > 0 || hasContentActions) && <div className="table-bar-actions">
                        {hasContentActions && <div className="table-bar-actions-scroll"><TableCommands items={contentBar.actions}/></div>}
                        {selectionAction}
                    </div>}
                </div>}
                <div className={`${styles.dataPanel} ui-table-data`}>
                    {tableRender ? tableRender(table) : table}
                </div>
            </section>
            {pagination !== false && data.total > 0 && <nav className="ui-table-pagination" aria-label="表格分页">
                <TablePagination {...pagination} page={data.page} pageSize={data.pageSize} total={data.total}
                                 disabled={pagination?.disabled ?? spinning}
                                 onChange={(page, size) => {data.changePage(page, size); scrollToTop();}}/>
            </nav>}
        </div>
    );
};

export default forwardRef(Table) as <RecordType extends object = any>(props: TableProps<RecordType> & React.RefAttributes<TableRef<RecordType>>) => React.ReactElement;
