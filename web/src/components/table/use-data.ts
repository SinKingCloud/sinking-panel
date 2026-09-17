import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App} from "antd";
import type {TableProps, TableSort} from "./types";

export default function useTableData<RecordType extends object>(props: TableProps<RecordType>) {
    const {message} = App.useApp();
    const latest = useRef(props);
    latest.current = props;
    const searchConfig = typeof props.search === "object" ? props.search : undefined;
    const fields = searchConfig?.fields;
    const defaultField = fields?.find((field) => field.value === searchConfig?.defaultField)?.value || fields?.[0]?.value || "";
    const hasSearch = Boolean(searchConfig);
    const [query, setQuery] = useState(() => ({
        page: props.defaultPage || 1,
        pageSize: props.defaultPageSize || 10,
        keyword: (searchConfig?.value ?? searchConfig?.defaultValue ?? "").trim(),
        field: defaultField,
        sort: props.defaultSort || {} as TableSort,
    }));
    const searchName = fields?.length
        ? fields.find((field) => field.value === (searchConfig?.field ?? query.field))?.value || defaultField
        : searchConfig?.name || "keyword";
    // 空搜索词切换字段不改变查询条件，也不打断正在加载的请求。
    const searchKey = hasSearch && query.keyword ? searchName : "";
    const [result, setResult] = useState({data: [] as RecordType[], total: 0});
    const [loading, setLoading] = useState(Boolean(props.request && !props.manualRequest && props.requestEnabled !== false));
    const [revision, setRevision] = useState(0);
    const requestVersion = useRef(0);
    const handledRevision = useRef(0);
    const adjustedPage = useRef<number | undefined>(undefined);
    const pagination = props.pagination || {};
    const page = pagination.page ?? query.page;
    const pageSize = pagination.pageSize ?? query.pageSize;
    const hasRequest = Boolean(props.request);
    const paramsKey = hasRequest ? JSON.stringify(props.params || {}) : "";
    const sortKey = hasRequest ? JSON.stringify(query.sort) : "";
    const previousParamsKey = useRef(paramsKey);
    const previousSearchKey = useRef(searchKey);

    useEffect(() => {
        if (!hasRequest || props.requestEnabled === false) {
            adjustedPage.current = undefined;
            handledRevision.current = revision;
            setLoading(false);
            return;
        }
        if (props.manualRequest && handledRevision.current === revision && adjustedPage.current !== page) {
            setLoading(false);
            return;
        }
        adjustedPage.current = undefined;
        handledRevision.current = revision;
        if (previousParamsKey.current !== paramsKey || previousSearchKey.current !== searchKey) {
            previousParamsKey.current = paramsKey;
            previousSearchKey.current = searchKey;
            if (page !== 1) {
                setQuery((value) => ({...value, page: 1}));
                if (latest.current.pagination) latest.current.pagination.onChange?.(1, pageSize);
                if (props.manualRequest) setRevision((value) => value + 1);
                return;
            }
        }
        const version = ++requestVersion.current;
        let active = true;
        let pageAdjusted = false;
        const request = latest.current.request!;
        const current = {params: latest.current.params, page, pageSize, keyword: query.keyword, sort: query.sort};
        setLoading(true);
        void (async () => {
            try {
                const response = await request({
                    ...current.params, current: current.page, pageSize: current.pageSize,
                    ...(searchKey ? {[searchKey]: current.keyword} : {}),
                }, current.sort);
                if (!active || version !== requestVersion.current) return;
                if (!response || response.success === false || (response.data != null && !Array.isArray(response.data))) {
                    throw new Error("获取列表失败");
                }
                const data = response.data ?? [];
                const total = Math.max(0, Number(response.total) || 0);
                setResult({data, total});
                latest.current.onLoad?.(data, {...response, data, total});
                const lastPage = Math.max(1, Math.ceil(total / current.pageSize));
                if (current.page > lastPage) {
                    const pagination = latest.current.pagination;
                    const controlled = Boolean(pagination && pagination.page !== undefined);
                    pageAdjusted = total > 0 && !controlled;
                    if (props.manualRequest && controlled) adjustedPage.current = lastPage;
                    setQuery((value) => ({...value, page: lastPage}));
                    if (pagination) pagination.onChange?.(lastPage, current.pageSize);
                    if (props.manualRequest && !controlled) setRevision((value) => value + 1);
                    return;
                }
            } catch (reason) {
                if (!active || version !== requestVersion.current) return;
                const error = reason instanceof Error ? reason : new Error("获取列表失败");
                if (latest.current.onRequestError) latest.current.onRequestError(error);
                else message.error(error.message);
            } finally {
                if (active && version === requestVersion.current && !pageAdjusted) setLoading(false);
            }
        })();
        return () => { active = false; };
    }, [hasRequest, props.manualRequest, props.requestEnabled, paramsKey, page, pageSize, query.keyword, searchKey, sortKey, revision, message]);

    const reload = useCallback(() => {
        adjustedPage.current = undefined;
        requestVersion.current += 1;
        if (latest.current.request) setRevision((value) => value + 1);
        else void latest.current.onReload?.();
    }, []);
    const changePage = useCallback((nextPage: number, nextSize: number) => {
        adjustedPage.current = undefined;
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
    const search = useCallback((keyword: string, field?: string) => {
        adjustedPage.current = undefined;
        const config = latest.current.pagination;
        const nextKeyword = keyword.trim();
        const fieldOnly = field !== undefined && !nextKeyword;
        if (latest.current.request) {
            setQuery((value) => {
                const nextField = field ?? value.field;
                const page = fieldOnly ? value.page : 1;
                return value.page === page && value.keyword === nextKeyword && value.field === nextField
                    ? value : {...value, page, keyword: nextKeyword, field: nextField};
            });
        } else {
            setQuery((value) => {
                const page = !fieldOnly && config !== false && config?.page === undefined ? 1 : value.page;
                const nextField = field ?? value.field;
                return page === value.page && nextField === value.field ? value : {...value, page, field: nextField};
            });
        }
        if (!fieldOnly && latest.current.request && config && config.page !== undefined) config.onChange?.(1, config.pageSize || pageSize);
    }, [pageSize]);
    const changeSort = useCallback((field: string, order?: "ascend" | "descend") => {
        adjustedPage.current = undefined;
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
        const search = typeof latest.current.search === "object" ? latest.current.search : undefined;
        const field = search?.fields?.find((field) => field.value === search.defaultField)?.value || search?.fields?.[0]?.value || "";
        if (latest.current.request) {
            const keyword = search?.defaultValue || "";
            setQuery({page: 1, pageSize: resetSize, keyword: keyword.trim(), field, sort: latest.current.defaultSort || {}});
        } else {
            setQuery((value) => {
                const page = latest.current.pagination !== false && config?.page === undefined ? 1 : value.page;
                const pageSize = latest.current.pagination !== false && config?.pageSize === undefined ? resetSize : value.pageSize;
                return page === value.page && pageSize === value.pageSize && field === value.field ? value : {...value, page, pageSize, field};
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
        page: resolvedPage, pageSize, sort: query.sort, searchName, reload, changePage, search, changeSort, reset,
    };
}
