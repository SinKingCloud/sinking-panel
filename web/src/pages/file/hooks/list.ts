import {useCallback, useEffect, useRef, useState} from "react";
import {getFileList} from "@/service/api/file";

interface ListQuery {
    path: string;
    keyword: string;
    page: number;
    pageSize: number;
    sort?: any;
    order?: any;
    generation: number;
}

const defaultQuery: ListQuery = {
    path: "/",
    keyword: "",
    page: 1,
    pageSize: 10,
    generation: 0,
};

const normalizePath = (value: string) => value || "/";

const normalizeOrder = (value?: string): any => (
    value === "asc" || value === "ascend" ? "asc" : "desc"
);

const useFileList = (initialPath = "/") => {
    const requestRef = useRef(0);
    const committedPathRef = useRef(normalizePath(initialPath));
    const keywordRef = useRef("");
    const [query, setQuery] = useState<ListQuery>(() => ({
        ...defaultQuery,
        path: normalizePath(initialPath),
    }));
    const [committedPath, setCommittedPath] = useState(committedPathRef.current);
    const [items, setItems] = useState<any[]>([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(true);
    const [initialized, setInitialized] = useState(false);
    const [loaded, setLoaded] = useState(false);
    const [error, setError] = useState("");
    const {path, keyword, page, pageSize, sort, order, generation} = query;

    useEffect(() => {
        const requestId = ++requestRef.current;
        let active = true;
        let pageAdjusted = false;
        const current = () => active && requestRef.current === requestId;
        const fail = (message: string) => {
            if (!current()) {
                return;
            }
            setError(message);
            setItems([]);
            setTotal(0);
            setLoaded(false);
            if (path !== committedPathRef.current) {
                committedPathRef.current = path;
                setCommittedPath(path);
            }
        };

        setLoading(true);
        setError("");

        void getFileList({
            body: {
                path,
                keyword: keyword || undefined,
                page,
                page_size: pageSize,
                ...(sort && order ? {
                    order_by_field: sort,
                    order_by_type: order,
                } : {}),
            },
        }).then((response) => {
            if (!current()) {
                return;
            }
            if (response?.code !== 200) {
                fail(response?.message || "获取文件列表失败");
                return;
            }

            const data = response.data;
            const list = Array.isArray(data?.list) ? data.list : [];
            const responseTotal = Number(data?.total);
            const nextTotal = Number.isFinite(responseTotal) && responseTotal > 0 ? responseTotal : 0;
            const lastPage = Math.max(1, Math.ceil(nextTotal / pageSize));
            if (page > lastPage) {
                pageAdjusted = true;
                setTotal(nextTotal);
                setQuery((value) => ({...value, page: lastPage}));
                return;
            }

            committedPathRef.current = path;
            setCommittedPath(path);
            setItems(list);
            setTotal(nextTotal);
            setLoaded(true);
        }).catch(() => {
            fail("获取文件列表失败");
        }).finally(() => {
            if (current() && !pageAdjusted) {
                setLoading(false);
                setInitialized(true);
            }
        });

        return () => {
            active = false;
            if (requestRef.current === requestId) {
                requestRef.current += 1;
            }
        };
    }, [generation, keyword, order, page, pageSize, path, sort]);

    const navigate = useCallback((nextPath: string) => {
        const value = normalizePath(nextPath);
        requestRef.current += 1;
        keywordRef.current = "";
        setLoading(true);
        setError("");
        setQuery((current) => ({
            ...current,
            path: value,
            keyword: "",
            page: 1,
            generation: current.generation + 1,
        }));
    }, []);

    const reload = useCallback(() => {
        requestRef.current += 1;
        setQuery((current) => ({...current, generation: current.generation + 1}));
    }, []);

    const changePage = useCallback((nextPage: number, nextPageSize: number = pageSize) => {
        requestRef.current += 1;
        setLoading(true);
        setError("");
        const size = Math.min(1000, Math.max(1, nextPageSize));
        setQuery((current) => ({
            ...current,
            page: size === current.pageSize ? Math.max(1, nextPage) : 1,
            pageSize: size,
            generation: current.generation + 1,
        }));
    }, [pageSize]);

    const changeSort = useCallback((field?: any, value?: string) => {
        requestRef.current += 1;
        setLoading(true);
        setError("");
        const nextSort = field && value ? field : undefined;
        setQuery((current) => ({
            ...current,
            page: 1,
            sort: nextSort,
            order: nextSort ? normalizeOrder(value) : undefined,
            generation: current.generation + 1,
        }));
    }, []);

    const search = useCallback((value: string) => {
        const nextKeyword = value.trim();
        if (keywordRef.current === nextKeyword) {
            return;
        }
        keywordRef.current = nextKeyword;
        requestRef.current += 1;
        setLoading(true);
        setError("");
        setQuery((current) => ({
            ...current,
            keyword: nextKeyword,
            page: 1,
            generation: current.generation + 1,
        }));
    }, []);

    return {
        path: committedPath,
        requestedPath: path,
        keyword,
        page,
        pageSize,
        total,
        sort,
        order,
        items,
        loading,
        initialized,
        navigating: path !== committedPath,
        loaded,
        error,
        initialError: !loaded ? error : "",
        navigate,
        reload,
        changePage,
        changeSort,
        search,
    };
};

export default useFileList;
