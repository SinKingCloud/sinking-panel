import {useCallback, useEffect, useRef, useState} from "react";
import {App} from "antd";
import {getLog} from "@/service/api/system";

const useList = () => {
    const {message} = App.useApp();
    const requestRef = useRef(0);
    const [logs, setLogs] = useState<any[]>([]);
    const [total, setTotal] = useState(0);
    const [keyword, setKeyword] = useState("");
    const [dateRange, setDateRange] = useState<any>(null);
    const [query, setQuery] = useState({
        page: 1,
        pageSize: 10,
        keyword: "",
        type: "",
        createTimeStart: "",
        createTimeEnd: "",
        sort: "id",
        order: "desc",
    });
    const [loading, setLoading] = useState(true);
    const [initialized, setInitialized] = useState(false);
    const [reloadKey, setReloadKey] = useState(0);
    const {
        page,
        pageSize,
        type,
        createTimeStart,
        createTimeEnd,
        sort,
        order,
    } = query;

    useEffect(() => {
        const requestId = ++requestRef.current;
        let active = true;
        let pageAdjusted = false;
        setLoading(true);

        void getLog({
            body: {
                page,
                page_size: pageSize,
                order_by_field: sort,
                order_by_type: order,
                keyword: query.keyword,
                type,
                create_time_start: createTimeStart,
                create_time_end: createTimeEnd,
            },
        }).then((response) => {
            if (!active || requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "获取操作日志失败");
                return;
            }

            const data: any = response?.data || {};
            const list = Array.isArray(data.list) ? data.list : [];
            const nextTotal = Number(data.total || 0);
            if (list.length === 0 && nextTotal > 0 && page > 1) {
                pageAdjusted = true;
                setQuery((current) => ({
                    ...current,
                    page: Math.max(1, Math.min(page - 1, Math.ceil(nextTotal / pageSize))),
                }));
                return;
            }
            setLogs(list);
            setTotal(nextTotal);
        }).catch(() => {
            if (active && requestRef.current === requestId) {
                message.error("获取操作日志失败");
            }
        }).finally(() => {
            if (active && requestRef.current === requestId && !pageAdjusted) {
                setLoading(false);
                setInitialized(true);
            }
        });

        return () => {
            active = false;
        };
    }, [
        createTimeEnd,
        createTimeStart,
        message,
        order,
        page,
        pageSize,
        query.keyword,
        reloadKey,
        sort,
        type,
    ]);

    const reload = useCallback(() => {
        requestRef.current += 1;
        setReloadKey((value) => value + 1);
    }, []);

    const changeKeyword = useCallback((value: string) => {
        setKeyword(value);
        if (!value) {
            setQuery((current) => ({...current, page: 1, keyword: ""}));
        }
    }, []);

    const search = useCallback((value: string) => {
        setKeyword(value);
        setQuery((current) => ({...current, page: 1, keyword: value.trim()}));
    }, []);

    const changeType = useCallback((value: string) => {
        setQuery((current) => ({...current, page: 1, type: value}));
    }, []);

    const changeDateRange = useCallback((value: any) => {
        setDateRange(value);
        setQuery((current) => ({
            ...current,
            page: 1,
            createTimeStart: value?.[0]?.format?.("YYYY-MM-DD 00:00:00") || "",
            createTimeEnd: value?.[1]?.format?.("YYYY-MM-DD 23:59:59") || "",
        }));
    }, []);

    const changeSort = useCallback((field: string, value?: string) => {
        setQuery((current) => ({
            ...current,
            page: 1,
            sort: value ? field : "id",
            order: value === "ascend" ? "asc" : "desc",
        }));
    }, []);

    const changePage = useCallback((nextPage: number, nextPageSize: number) => {
        setQuery((current) => ({
            ...current,
            page: nextPageSize === current.pageSize ? nextPage : 1,
            pageSize: nextPageSize,
        }));
    }, []);

    return {
        logs,
        total,
        keyword,
        type,
        dateRange,
        page,
        pageSize,
        sort,
        order,
        loading,
        initialized,
        reload,
        changeKeyword,
        search,
        changeType,
        changeDateRange,
        changeSort,
        changePage,
    };
};

export default useList;
