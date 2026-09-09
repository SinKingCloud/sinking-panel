import React, {useRef} from "react";
import {Body} from "sinking-antd";
import Table from "@/components/table";
import useEnum from "@/utils/enum";
import useLogHeader from "./components/header";
import Clear, {ClearRef} from "./components/clear";
import useLogTable from "./components/table";
import useList from "./hooks/list";
import useStyles from "./styles";

const emptyEnum = {};

export default (): React.ReactNode => {
    const {styles} = useStyles();
    const [enumData, enumLoading] = useEnum("log");
    const clearRef = useRef<ClearRef | null>(null);
    const list = useList();
    const typeData = enumData?.type || emptyEnum;

    const header = useLogHeader({
        keyword: list.keyword,
        type: list.type,
        dateRange: list.dateRange,
        typeData,
        refreshing: list.loading,
        onKeywordChange: list.changeKeyword,
        onSearch: list.search,
        onTypeChange: list.changeType,
        onDateRangeChange: list.changeDateRange,
        onRefresh: list.reload,
        onClear: () => clearRef.current?.open(),
    });
    const table = useLogTable({
        className: styles.logTable,
        logs: list.logs,
        loading: list.loading,
        typeData,
        sort: list.sort,
        order: list.order,
        onSortChange: list.changeSort,
    });

    return (
        <Body loading={enumLoading}>
            <Table
                {...header}
                {...table}
                ariaLabel="操作日志"
                empty={list.initialized && list.logs.length === 0}
                pagination={{
                    page: list.page,
                    pageSize: list.pageSize,
                    total: list.total,
                    unit: "条",
                    onChange: list.changePage,
                }}/>
            <Clear ref={clearRef} onSuccess={list.reload}/>
        </Body>
    );
};
