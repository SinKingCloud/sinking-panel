import {useCallback, useEffect, useRef, useState} from "react";
import {App} from "antd";
import {
    deleteTask,
    getTaskList,
    restoreTask,
    runTask,
    stopTask,
} from "@/service/api/task";

const actions: any = {
    run: {
        request: runTask,
        title: "执行任务",
        content: (name: string) => `确定立即执行任务“${name}”吗？`,
        okText: "执行",
        success: "任务已提交执行",
        empty: "请选择要执行的任务",
        filter: () => true,
    },
    stop: {
        request: stopTask,
        title: "暂停任务",
        content: (name: string) => `确定暂停任务“${name}”吗？`,
        okText: "暂停",
        success: "任务已暂停",
        loading: "正在暂停任务...",
        empty: "请选择运行中的任务",
        filter: (record: any) => Number(record?.status) === 0,
    },
    restore: {
        request: restoreTask,
        title: "恢复任务",
        content: (name: string) => `确定恢复任务“${name}”吗？`,
        okText: "恢复",
        success: "任务已恢复",
        loading: "正在恢复任务...",
        empty: "请选择已暂停的任务",
        filter: (record: any) => Number(record?.status) === 1,
    },
    delete: {
        request: deleteTask,
        title: "删除任务",
        content: (name: string) => `确定删除任务“${name}”吗？`,
        okText: "确定",
        success: "删除成功",
        empty: "请选择要删除的任务",
        filter: () => true,
        danger: true,
    },
};

const useList = () => {
    const {message, modal} = App.useApp();
    const requestRef = useRef(0);
    const operatingRef = useRef(new Set<string>());
    const [tasks, setTasks] = useState<any[]>([]);
    const [total, setTotal] = useState(0);
    const [keyword, setKeyword] = useState("");
    const [query, setQuery] = useState({
        page: 1,
        pageSize: 10,
        name: "",
        typeId: "0",
        execType: "",
        status: "",
        sort: "id",
        order: "desc",
    });
    const [loading, setLoading] = useState(true);
    const [reloadKey, setReloadKey] = useState(0);
    const [operating, setOperating] = useState(() => new Set<string>());
    const {page, pageSize, name, typeId, execType, status, sort, order} = query;

    useEffect(() => {
        const requestId = ++requestRef.current;
        let active = true;
        let pageAdjusted = false;
        setLoading(true);

        void getTaskList({
            body: {
                page,
                page_size: pageSize,
                order_by_field: sort,
                order_by_type: order,
                name,
                type_id: typeId,
                exec_type: execType,
                status,
            },
        }).then((response) => {
            if (!active || requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "获取计划任务失败");
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
            setTasks(list);
            setTotal(nextTotal);
        }).catch(() => {
            if (active && requestRef.current === requestId) {
                message.error("获取计划任务失败");
            }
        }).finally(() => {
            if (active && requestRef.current === requestId && !pageAdjusted) {
                setLoading(false);
            }
        });

        return () => {
            active = false;
        };
    }, [execType, message, name, order, page, pageSize, reloadKey, sort, status, typeId]);

    const reload = useCallback(() => {
        requestRef.current += 1;
        setReloadKey((value) => value + 1);
    }, []);

    const changeKeyword = useCallback((value: string) => {
        setKeyword(value);
        if (!value) {
            setQuery((current) => ({...current, page: 1, name: ""}));
        }
    }, []);

    const search = useCallback((value: string) => {
        setQuery((current) => ({...current, page: 1, name: value.trim()}));
    }, []);

    const changeTypeId = useCallback((value: string) => {
        setQuery((current) => ({...current, page: 1, typeId: value}));
    }, []);

    const changeExecType = useCallback((value: string) => {
        setQuery((current) => ({...current, page: 1, execType: value}));
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

    const execute = useCallback(async (action: string, record: any) => {
        const option = actions[action];
        const id = record?.id;
        const rowKey = String(id ?? "");
        if (!option || id === undefined || id === null || !option.filter(record)) {
            message.warning(option?.empty || "操作参数不正确");
            return;
        }
        if (operatingRef.current.has(rowKey)) {
            return;
        }

        operatingRef.current.add(rowKey);
        setOperating(new Set(operatingRef.current));
        const loadingKey = `task-operation-${action}-${rowKey}`;
        if (option.loading) {
            message.loading({key: loadingKey, content: option.loading, duration: 0});
        }

        try {
            const response = await option.request({body: {ids: [id]}});
            if (response?.code !== 200) {
                message.error(response?.message || "操作失败");
                return;
            }
            message.success(response?.message || option.success);
            if (action === "stop" || action === "restore") {
                reload();
                return;
            }
            if (action === "delete") {
                reload();
            }
        } catch {
            message.error("操作失败");
        } finally {
            if (option.loading) {
                message.destroy(loadingKey);
            }
            operatingRef.current.delete(rowKey);
            setOperating(new Set(operatingRef.current));
        }
    }, [message, reload]);

    const confirmAction = useCallback((action: string, record: any) => {
        const option = actions[action];
        if (!option) {
            return;
        }
        modal.confirm({
            title: option.title,
            content: option.content(record?.name || "-"),
            okText: option.okText,
            cancelText: "取消",
            okButtonProps: option.danger ? {danger: true, type: "default"} : undefined,
            mask: {closable: true},
            onOk: () => execute(action, record),
        } as any);
    }, [execute, modal]);

    return {
        tasks,
        total,
        page,
        pageSize,
        keyword,
        typeId,
        execType,
        status,
        sort,
        order,
        loading,
        operating,
        reload,
        changeKeyword,
        search,
        changeTypeId,
        changeExecType,
        changeStatus,
        changeSort,
        changePage,
        confirmAction,
    };
};

export default useList;
