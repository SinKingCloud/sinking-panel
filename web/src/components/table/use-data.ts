import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App} from "antd";
import type {TableProps, TableSort} from "./types";

export default function useTableData<RecordType extends object>(props: TableProps<RecordType>) {
    const {message} = App.useApp();
    const latest = useRef(props);
    latest.current = props;
    const [query, setQuery] = useState(() => ({
        page: props.defaultPage || 1,
        pageSize: props.defaultPageSize || 10,
        keyword: props.toolbar ? props.toolbar.search?.value || "" : "",
        sort: props.defaultSort || {} as TableSort,
    }));
    const [result, setResult] = useState({data: [] as RecordType[], total: 0});
    const [loading, setLoading] = useState(Boolean(props.request && !props.manualRequest && props.requestEnabled !== false));
    const [revision, setRevision] = useState(0);
    const requestVersion = useRef(0);
    const handledRevision = useRef(0);
    const pagination = props.pagination || {};
    const page = pagination.page ?? query.page;
    const pageSize = pagination.pageSize ?? query.pageSize;
    const hasRequest = Boolean(props.request);
    const paramsKey = hasRequest ? JSON.stringify(props.params || {}) : "";
    const sortKey = hasRequest ? JSON.stringify(query.sort) : "";
    const previousParamsKey = useRef(paramsKey);

    useEffect(() => {
        if (!hasRequest || props.requestEnabled === false) {
            handledRevision.current = revision;
            setLoading(false);
            return;
        }
        if (props.manualRequest && handledRevision.current === revision) {
            setLoading(false);
            return;
        }
        handledRevision.current = revision;
        if (previousParamsKey.current !== paramsKey) {
            previousParamsKey.current = paramsKey;
            if (page !== 1) {
                setQuery((value) => ({...value, page: 1}));
                if (latest.current.pagination) latest.current.pagination.onChange?.(1, pageSize);
                if (props.manualRequest) setRevision((value) => value + 1);
                return;
            }
        }
        const version = ++requestVersion.current;
        let active = true;
        const current = {params: latest.current.params, page, pageSize, keyword: query.keyword, sort: query.sort};
        setLoading(true);
        void (async () => {
            try {
                const response = await latest.current.request!({
                    ...current.params, page: current.page, pageSize: current.pageSize, keyword: current.keyword,
                }, current.sort);
                if (!active || version !== requestVersion.current) return;
                if (response?.success === false || !Array.isArray(response?.data)) {
                    throw new Error("获取列表失败");
                }
                const total = Math.max(0, Number(response.total) || 0);
                const lastPage = Math.max(1, Math.ceil(total / current.pageSize));
                if (current.page > lastPage) {
                    setQuery((value) => ({...value, page: lastPage}));
                    if (latest.current.pagination) latest.current.pagination.onChange?.(lastPage, current.pageSize);
                    if (props.manualRequest) setRevision((value) => value + 1);
                    return;
                }
                setResult({data: response.data, total});
                latest.current.onLoad?.(response.data, {...response, total});
            } catch (reason) {
                if (!active || version !== requestVersion.current) return;
                const error = reason instanceof Error ? reason : new Error("获取列表失败");
                if (latest.current.onRequestError) latest.current.onRequestError(error);
                else message.error(error.message);
            } finally {
                if (active && version === requestVersion.current) setLoading(false);
            }
        })();
        return () => { active = false; };
    }, [hasRequest, props.manualRequest, props.requestEnabled, paramsKey, page, pageSize, query.keyword, sortKey, revision, message]);

    const reload = useCallback(() => {
        requestVersion.current += 1;
        if (latest.current.request) setRevision((value) => value + 1);
        else void latest.current.onReload?.();
    }, []);
    const changePage = useCallback((nextPage: number, nextSize: number) => {
        const config = latest.current.pagination || {};
        const resolvedPage = nextSize === pageSize ? nextPage : 1;
        if (config.page === undefined || config.pageSize === undefined) {
            setQuery((value) => {
                const page = config.page === undefined ? resolvedPage : value.page;
                const pageSize = config.pageSize === undefined ? nextSize : value.pageSize;
                return page === value.page && pageSize === value.pageSize ? value : {...value, page, pageSize};
            });
        }
        config.onChange?.(resolvedPage, nextSize);
    }, [pageSize]);
    const search = useCallback((keyword: string) => {
        const config = latest.current.pagination;
        if (latest.current.request) {
            const nextKeyword = keyword.trim();
            setQuery((value) => value.page === 1 && value.keyword === nextKeyword ? value : {...value, page: 1, keyword: nextKeyword});
        } else if (config !== false && config?.page === undefined) {
            setQuery((value) => value.page === 1 ? value : {...value, page: 1});
        }
        if (latest.current.request && config && config.page !== undefined) config.onChange?.(1, config.pageSize || pageSize);
    }, [pageSize]);
    const changeSort = useCallback((field: string, order?: "ascend" | "descend") => {
        const config = latest.current.pagination;
        if (latest.current.request) {
            const sort = field && order ? {[field]: order} : latest.current.defaultSort || {};
            setQuery((value) => value.page === 1 && JSON.stringify(value.sort) === JSON.stringify(sort) ? value : {...value, page: 1, sort});
        } else if (config !== false && config?.page === undefined) {
            setQuery((value) => value.page === 1 ? value : {...value, page: 1});
        }
        latest.current.onSortChange?.(field, order);
        if (latest.current.request && config && config.page !== undefined) config.onChange?.(1, config.pageSize || pageSize);
    }, [pageSize]);
    const reset = useCallback(() => {
        const resetSize = latest.current.defaultPageSize || 10;
        const config = latest.current.pagination || undefined;
        if (latest.current.request) {
            setQuery({page: 1, pageSize: resetSize, keyword: "", sort: latest.current.defaultSort || {}});
        } else if (latest.current.pagination !== false && (config?.page === undefined || config.pageSize === undefined)) {
            setQuery((value) => {
                const page = config?.page === undefined ? 1 : value.page;
                const pageSize = config?.pageSize === undefined ? resetSize : value.pageSize;
                return page === value.page && pageSize === value.pageSize ? value : {...value, page, pageSize};
            });
        }
        if (config?.onChange && (page !== 1 || pageSize !== resetSize)) {
            config.onChange(1, resetSize);
            // 受控列表换页已由父级发起请求，不能再用旧页码刷新。
            if (!latest.current.request) return;
        }
        reload();
    }, [page, pageSize, reload]);

    const source = props.dataSource ?? result.data;
    const localPagination = !hasRequest && props.pagination !== false && pagination.total === undefined && !pagination.onChange;
    const resolvedPage = localPagination ? Math.min(page, Math.max(1, Math.ceil(source.length / pageSize))) : page;
    const data = useMemo(() => localPagination ? source.slice((resolvedPage - 1) * pageSize, resolvedPage * pageSize) : source,
        [source, localPagination, resolvedPage, pageSize]);
    return {
        data,
        loading: props.loading ?? loading,
        total: pagination.total ?? (hasRequest ? result.total : source.length),
        page: resolvedPage, pageSize, reload, changePage, search, changeSort, reset,
    };
}
