import {useCallback, useEffect, useRef, useState} from "react";
import {App} from "antd";
import {getScriptList} from "@/service/api/script";
import type {ScriptRecord} from "@/service/api/script";

const pageSize = 50;

const useScripts = () => {
    const {message} = App.useApp();
    const requestRef = useRef(0);
    const loadingMoreRef = useRef(false);
    const queryRef = useRef("");
    const [items, setItems] = useState<ScriptRecord[]>([]);
    const [keyword, setKeyword] = useState("");
    const [queryKeyword, setQueryKeyword] = useState("");
    const [typeId, setTypeId] = useState("0");
    const [loading, setLoading] = useState(true);
    const [loadingMore, setLoadingMore] = useState(false);
    const [loaded, setLoaded] = useState(false);
    const [error, setError] = useState(false);
    const [hasMore, setHasMore] = useState(false);
    const [nextCursor, setNextCursor] = useState("");
    const [reloadKey, setReloadKey] = useState(0);

    useEffect(() => {
        const timer = window.setTimeout(() => setQueryKeyword(keyword.trim()), 300);
        return () => window.clearTimeout(timer);
    }, [keyword]);

    useEffect(() => {
        const requestId = ++requestRef.current;
        let active = true;
        const current = () => active && requestRef.current === requestId;
        const queryKey = `${typeId}\u0000${queryKeyword}`;
        if (queryRef.current !== queryKey) {
            queryRef.current = queryKey;
            setItems([]);
            setLoaded(false);
        }
        setLoading(true);
        setLoadingMore(false);
        loadingMoreRef.current = false;
        setError(false);
        setHasMore(false);
        setNextCursor("");

        void getScriptList({
            body: {
                page_size: pageSize,
                order_by_field: "id",
                order_by_type: "desc",
                keyword: queryKeyword || undefined,
                type_id: typeId === "0" ? undefined : typeId,
            },
        }).then((response) => {
            if (!current()) {
                return;
            }
            if (response?.code !== 200) {
                setError(true);
                if (response) {
                    message.error(response.message || "获取常用脚本失败");
                }
                return;
            }
            const list = Array.isArray(response.data?.list) ? response.data.list : [];
            const cursor = String(response.data?.next_cursor_id || "");
            setItems(list);
            setHasMore(Boolean(response.data?.has_more && cursor));
            setNextCursor(cursor);
            setLoaded(true);
        }).finally(() => {
            if (current()) {
                setLoading(false);
            }
        });

        return () => {
            active = false;
            if (requestRef.current === requestId) {
                requestRef.current += 1;
                loadingMoreRef.current = false;
            }
        };
    }, [message, queryKeyword, reloadKey, typeId]);

    const loadMore = useCallback(async () => {
        if (loadingMoreRef.current || loading || !hasMore || !nextCursor) {
            return;
        }
        const requestId = requestRef.current;
        loadingMoreRef.current = true;
        setLoadingMore(true);
        try {
            const response = await getScriptList({
                body: {
                    page_size: pageSize,
                    order_by_field: "id",
                    order_by_type: "desc",
                    cursor_id: nextCursor,
                    keyword: queryKeyword || undefined,
                    type_id: typeId === "0" ? undefined : typeId,
                },
            });
            if (requestRef.current !== requestId || !response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || "加载更多脚本失败");
                return;
            }
            const list = Array.isArray(response.data?.list) ? response.data.list : [];
            setItems((current) => {
                const ids = new Set(current.map((item) => item.id));
                return [...current, ...list.filter((item) => !ids.has(item.id))];
            });
            const cursor = String(response.data?.next_cursor_id || "");
            setHasMore(Boolean(response.data?.has_more && cursor));
            setNextCursor(cursor);
        } finally {
            if (requestRef.current === requestId) {
                loadingMoreRef.current = false;
                setLoadingMore(false);
            }
        }
    }, [hasMore, loading, message, nextCursor, queryKeyword, typeId]);

    const reload = useCallback(() => {
        requestRef.current += 1;
        loadingMoreRef.current = false;
        setReloadKey((value) => value + 1);
    }, []);

    return {
        items,
        keyword,
        typeId,
        loading,
        loadingMore,
        loaded,
        error,
        hasMore,
        setKeyword,
        setTypeId,
        loadMore,
        reload,
    };
};

export default useScripts;
