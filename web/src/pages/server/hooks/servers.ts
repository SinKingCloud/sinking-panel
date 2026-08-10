import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App} from "antd";
import {getServerInfo, getServerList} from "@/service/api/server";

export interface ServerRecord {
    id: number;
    ip: string;
    port: number;
    user: string;
    auth_type: number;
    name: string;
    create_time?: string;
    update_time?: string;
    credential_configured?: boolean;
    searchText: string;
}

const normalizeServer = (item: any): ServerRecord => {
    const id = Number(item?.id || 0);
    const ip = String(item?.ip || "");
    const port = Number(item?.port || 22);
    const user = String(item?.user || "");
    const name = String(item?.name || (id === 0 ? "本机终端" : ""));
    return {
        id,
        ip,
        port,
        user,
        auth_type: Number(item?.auth_type || 0),
        name,
        create_time: item?.create_time ? String(item.create_time) : undefined,
        update_time: item?.update_time ? String(item.update_time) : undefined,
        credential_configured: Boolean(item?.credential_configured),
        searchText: [name, ip, String(port), `${ip}:${port}`].join("\n").toLocaleLowerCase(),
    };
};

export const localServer = normalizeServer({
    id: 0,
    ip: "127.0.0.1",
    port: 22,
    user: "",
    auth_type: 0,
    name: "本机终端",
    credential_configured: false,
});

const useServers = () => {
    const {message} = App.useApp();
    const listRequestRef = useRef(0);
    const localRequestRef = useRef(0);
    const [remoteServers, setRemoteServers] = useState<ServerRecord[]>([]);
    const [local, setLocal] = useState<ServerRecord>(localServer);
    const [localLoaded, setLocalLoaded] = useState(false);
    const [localRefreshing, setLocalRefreshing] = useState(true);
    const [localError, setLocalError] = useState(false);
    const [loading, setLoading] = useState(true);
    const [loadingMore, setLoadingMore] = useState(false);
    const [loaded, setLoaded] = useState(false);
    const [hasMore, setHasMore] = useState(false);
    const [nextCursor, setNextCursor] = useState("");
    const [keyword, setKeyword] = useState("");
    const [queryKeyword, setQueryKeyword] = useState("");
    const [reloadKey, setReloadKey] = useState(0);
    const loadingMoreRef = useRef(false);

    useEffect(() => {
        const timer = window.setTimeout(() => {
            setQueryKeyword(keyword.trim());
        }, 300);
        return () => window.clearTimeout(timer);
    }, [keyword]);

    useEffect(() => {
        const requestId = ++listRequestRef.current;
        let active = true;
        const current = () => active && listRequestRef.current === requestId;
        setLoading(true);
        setLoadingMore(false);
        loadingMoreRef.current = false;
        setHasMore(false);
        setNextCursor("");

        void getServerList({
            body: {
                page_size: 100,
                order_by_field: "id",
                order_by_type: "desc",
                keyword: queryKeyword,
            },
        }).then((response) => {
            if (!current() || !response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response?.message || "获取终端列表失败");
                return;
            }
            const list = Array.isArray((response?.data as any)?.list)
                ? (response?.data as any).list
                : [];
            setRemoteServers(list.map(normalizeServer));
            setHasMore(Boolean((response?.data as any)?.has_more));
            setNextCursor(String((response?.data as any)?.next_cursor_id || ""));
        }).catch(() => {
            if (current()) {
                message.error("获取终端列表失败");
            }
        }).finally(() => {
            if (current()) {
                setLoading(false);
                setLoaded(true);
            }
        });

        return () => {
            active = false;
            if (listRequestRef.current === requestId) {
                listRequestRef.current += 1;
                loadingMoreRef.current = false;
            }
        };
    }, [message, queryKeyword, reloadKey]);

    useEffect(() => {
        const requestId = ++localRequestRef.current;
        let active = true;
        const current = () => active && localRequestRef.current === requestId;
        setLocalRefreshing(true);
        setLocalError(false);

        void getServerInfo({body: {id: 0}}).then((response) => {
            if (!current()) {
                return;
            }
            if (response?.code === 200 && response.data) {
                setLocal(normalizeServer(response.data));
                setLocalLoaded(true);
                setLocalError(false);
                return;
            }
            setLocalError(true);
            if (response) {
                message.error(response.message || "获取本机连接失败");
            }
        }).catch(() => {
            if (current()) {
                setLocalError(true);
                message.error("获取本机连接失败");
            }
        }).finally(() => {
            if (current()) {
                setLocalRefreshing(false);
            }
        });

        return () => {
            active = false;
            if (localRequestRef.current === requestId) {
                localRequestRef.current += 1;
            }
        };
    }, [message, reloadKey]);

    const reload = useCallback(() => {
        listRequestRef.current += 1;
        localRequestRef.current += 1;
        setReloadKey((value) => value + 1);
    }, []);

    const loadMore = useCallback(async () => {
        if (loadingMoreRef.current || loading || !hasMore || !nextCursor) {
            return;
        }
        const requestId = listRequestRef.current;
        loadingMoreRef.current = true;
        setLoadingMore(true);
        try {
            const response = await getServerList({
                body: {
                    page_size: 100,
                    order_by_field: "id",
                    order_by_type: "desc",
                    cursor_id: nextCursor,
                    keyword: queryKeyword,
                },
            });
            if (listRequestRef.current !== requestId) {
                return;
            }
            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || "加载更多终端失败");
                return;
            }
            const list = Array.isArray((response.data as any)?.list)
                ? (response.data as any).list.map(normalizeServer)
                : [];
            setRemoteServers((current) => {
                const ids = new Set(current.map((server) => server.id));
                return [...current, ...list.filter((server: ServerRecord) => !ids.has(server.id))];
            });
            setHasMore(Boolean((response.data as any)?.has_more));
            setNextCursor(String((response.data as any)?.next_cursor_id || ""));
        } finally {
            if (listRequestRef.current === requestId) {
                loadingMoreRef.current = false;
                setLoadingMore(false);
            }
        }
    }, [hasMore, loading, message, nextCursor, queryKeyword]);

    const servers = useMemo(() => {
        const search = queryKeyword.toLocaleLowerCase();
        return search && !local.searchText.includes(search)
            ? remoteServers
            : [local, ...remoteServers];
    }, [local, queryKeyword, remoteServers]);

    return {
        servers,
        keyword,
        loading,
        loadingMore,
        loaded,
        hasMore,
        localLoaded,
        localRefreshing,
        localError,
        localServer: local,
        setKeyword,
        loadMore,
        reload,
    };
};

export default useServers;
