import {createElement, useCallback, useEffect, useRef, useState} from "react";
import {App, Checkbox} from "antd";
import {deleteSite, disableSite, enableSite, getSiteList} from "@/service/api/site";
import {clearEnumCache} from "@/utils/enum";

const actions: any = {
    enable: {
        request: enableSite,
        title: "启用网站",
        content: (name: string) => `确定启用网站“${name}”吗？`,
        okText: "启用",
        success: "网站已启用",
        filter: (record: any) => Number(record?.status) !== 0,
    },
    disable: {
        request: disableSite,
        title: "停用网站",
        content: (name: string) => `确定停用网站“${name}”吗？停用后将无法继续访问。`,
        okText: "停用",
        success: "网站已停用",
        filter: (record: any) => Number(record?.status) === 0,
    },
    delete: {
        request: deleteSite,
        title: "删除网站",
        content: (name: string) => `确定删除网站“${name}”吗？删除后无法恢复。`,
        okText: "删除",
        success: "网站已删除",
        filter: () => true,
        danger: true,
    },
};

const useList = () => {
    const {message, modal} = App.useApp();
    const requestRef = useRef(0);
    const operatingRef = useRef(new Set<string>());
    const [sites, setSites] = useState<any[]>([]);
    const [total, setTotal] = useState(0);
    const [keyword, setKeyword] = useState("");
    const [query, setQuery] = useState({
        page: 1,
        pageSize: 10,
        keyword: "",
        typeId: "0",
        type: "",
        status: "",
        sort: "id",
        order: "desc",
    });
    const [loading, setLoading] = useState(true);
    const [initialized, setInitialized] = useState(false);
    const [reloadKey, setReloadKey] = useState(0);
    const [operating, setOperating] = useState(() => new Set<string>());
    const {page, pageSize, typeId, type, status, sort, order} = query;

    useEffect(() => {
        const requestId = ++requestRef.current;
        let active = true;
        let pageAdjusted = false;
        setLoading(true);

        void getSiteList({
            body: {
                page,
                page_size: pageSize,
                order_by_field: sort,
                order_by_type: order,
                keyword: query.keyword,
                type_id: typeId,
                type,
                status,
            },
        }).then((response) => {
            if (!active || requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "获取网站列表失败");
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
            setSites(list);
            setTotal(nextTotal);
        }).catch(() => {
            if (active && requestRef.current === requestId) {
                message.error("获取网站列表失败");
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
    }, [message, order, page, pageSize, query.keyword, reloadKey, sort, status, type, typeId]);

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

    const changeTypeId = useCallback((value: string) => {
        setQuery((current) => ({...current, page: 1, typeId: value}));
    }, []);

    const changeStatus = useCallback((value: string) => {
        setQuery((current) => ({...current, page: 1, status: value}));
    }, []);

    const changeSort = useCallback((field: string, value: string) => {
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

    const execute = useCallback(async (action: string, record: any, body: any = {}) => {
        const option = actions[action];
        const id = record?.id;
        const rowKey = String(id ?? "");
        if (!option || id === undefined || id === null || !option.filter(record)) {
            message.warning("当前网站状态不支持此操作");
            return;
        }
        if (operatingRef.current.has(rowKey)) {
            return;
        }

        operatingRef.current.add(rowKey);
        setOperating(new Set(operatingRef.current));
        try {
            const response = await option.request({body: {id, ...body}});
            if (response?.code !== 200) {
                message.error(response?.message || "操作失败");
                if (action === "delete") {
                    reload();
                }
                return;
            }
            clearEnumCache("site");
            message.success(response?.message || option.success);
            reload();
        } catch {
            message.error("操作失败");
        } finally {
            operatingRef.current.delete(rowKey);
            setOperating(new Set(operatingRef.current));
        }
    }, [message, reload]);

    const confirmAction = useCallback((action: string, record: any) => {
        const option = actions[action];
        if (!option) {
            return;
        }
        let deleteRoot = false;
        modal.confirm({
            title: option.title,
            content: action === "delete" ? createElement(
                "div",
                null,
                createElement("div", null, option.content(record?.name || "-")),
                createElement(
                    "div",
                    {style: {marginTop: 12}},
                    createElement(Checkbox, {
                        onChange: (event) => {
                            deleteRoot = event.target.checked;
                        },
                    }, "同时删除网站根目录"),
                ),
            ) : option.content(record?.name || "-"),
            okText: option.okText,
            cancelText: "取消",
            okButtonProps: option.danger ? {danger: true, type: "default"} : undefined,
            mask: {closable: true},
            onOk: () => {
                if (action !== "delete") {
                    return execute(action, record);
                }
                modal.confirm({
                    title: "再次确认",
                    content: deleteRoot ? "确定删除网站及根目录吗？" : "确定删除网站吗？",
                    okText: "删除",
                    cancelText: "取消",
                    okButtonProps: {danger: true, type: "default"},
                    mask: {closable: true},
                    onOk: () => execute(action, record, {delete_root: deleteRoot}),
                } as any);
            },
        } as any);
    }, [execute, modal]);

    return {
        sites,
        total,
        page,
        pageSize,
        keyword,
        typeId,
        type,
        status,
        sort,
        order,
        loading,
        initialized,
        operating,
        reload,
        changeKeyword,
        search,
        changeTypeId,
        changeType,
        changeStatus,
        changeSort,
        changePage,
        confirmAction,
    };
};

export default useList;
