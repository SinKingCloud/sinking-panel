import {forwardRef, lazy, memo, Suspense, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Button, Grid, Tooltip, Typography} from "antd";
import {ModalTable, Title, useTheme} from "sinking-antd";
import type {ModalTableProps, ModalTableRef} from "sinking-antd";
import ActionDropdown from "@/pages/components/ui/action-dropdown";
import {getContainerList, removeContainer, restartContainer, startContainer, stopContainer} from "@/service/api/container";

import useEnum from "@/utils/enum";
import {getParams} from "@/utils/page";
import RunContainer from "./run";

import ContainerDetail from "./detail";

import ContainerExec from "./exec";

import ImagesManager from "./images";

import useStyles from "./styles";

const ContainerTerminal = lazy(() => import("@/pages/components/container-terminal"));
const emptyStatusData: Record<string, string> = {};
const operations = {
    start: {
        label: "启动容器", okText: "启动", request: startContainer,
        content: "确定按已保存的配置启动容器吗？",
    },
    stop: {
        label: "停止容器", okText: "停止", request: stopContainer,
        content: "停止后服务将不可用，自动恢复也会暂停，容器数据会保留。确定停止吗？",
    },
    restart: {
        label: "重启容器", okText: "重启", request: restartContainer,
        content: "重启会暂时中断服务，并使用已保存的配置重新启动。确定重启吗？",
    },
    remove: {
        label: "删除容器", okText: "删除", request: removeContainer,
        content: "将停止容器并删除容器内的写入内容、日志及运行记录。挂载到服务器的数据目录和镜像会保留。",
    },
};

const ContainerManager = forwardRef<any>((_, ref) => {
    const {message, modal} = App.useApp();
    const {styles} = useStyles();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const [open, setOpen] = useState(false);
    const [enumData, enumLoading] = useEnum(open ? "container" : "");
    const statusData = enumData.status || emptyStatusData;
    const tableRef = useRef<ModalTableRef<any> | null>(null);
    const runRef = useRef<any | null>(null);
    const detailRef = useRef<any | null>(null);
    const execRef = useRef<any | null>(null);
    const imagesRef = useRef<any | null>(null);
    const confirmRef = useRef<{destroy: () => void} | null>(null);
    const mountedRef = useRef(true);
    const busyRef = useRef(new Set<string>());
    const [busy, setBusy] = useState(new Set<string>());
    const [terminal, setTerminal] = useState<any>(null);
    const [terminalMounted, setTerminalMounted] = useState(false);
    const [runMounted, setRunMounted] = useState(false);
    const [execMounted, setExecMounted] = useState(false);
    const pendingRunRef = useRef<any>(undefined);
    const pendingExecRef = useRef<any>(undefined);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            confirmRef.current?.destroy();
        };
    }, []);

    useImperativeHandle(ref, () => ({open: () => {
        if (busyRef.current.size) return;
        confirmRef.current?.destroy();
        setOpen(true);
    }}), []);

    const reload = useCallback(() => tableRef.current?.reload(), []);
    const closeTerminal = useCallback(() => setTerminal(null), []);
    const openTerminal = useCallback((record: any) => {
        setTerminalMounted(true);
        setTerminal(record);
    }, []);
    const openRun = useCallback((image?: any) => {
        if (runRef.current) runRef.current.open(image);
        else {
            pendingRunRef.current = image;
            setRunMounted(true);
        }
    }, []);
    const openExec = useCallback((record: any) => {
        if (execRef.current) execRef.current.open(record);
        else {
            pendingExecRef.current = record;
            setExecMounted(true);
        }
    }, []);

    // 首次使用才挂载表单；等挂载检查结束再打开，避免 StrictMode 清理中断初始化请求。
    useEffect(() => {
        if (!runMounted) return;
        let disposed = false;
        queueMicrotask(() => {
            if (!disposed) runRef.current?.open(pendingRunRef.current);
        });
        return () => { disposed = true; };
    }, [runMounted]);
    useEffect(() => {
        if (!execMounted) return;
        let disposed = false;
        queueMicrotask(() => {
            if (!disposed) execRef.current?.open(pendingExecRef.current);
        });
        return () => { disposed = true; };
    }, [execMounted]);

    const operate = useCallback((action: keyof typeof operations, records: any[]) => {
        const targets = records.filter((record) => !busyRef.current.has(record.id)
            && !["starting", "stopping"].includes(record.status)
            && (action !== "start" || record.status !== "running")
            && (action !== "stop" || record.status !== "stopped" || (record.auto_restart && !record.manual_stop)));
        if (!targets.length) {
            message.info("所选容器当前无需此操作，或正在处理中");
            return;
        }
        const execute = async () => {
            if (!mountedRef.current) return;
            // 同一容器的操作互斥；启动、停止和重启依次执行，避免同时启动大量实例。
            const pending = targets.filter((record) => !busyRef.current.has(record.id));
            if (!pending.length) return;
            pending.forEach((record) => busyRef.current.add(record.id));
            setBusy(new Set(busyRef.current));
            const failures: {record: any; reason: string}[] = [];
            let succeeded = 0;
            try {
                if (action === "remove") {
                    try {
                        const response = await removeContainer({body: {ids: pending.map((record) => record.id)}});
                        if (!mountedRef.current) return;
                        if (response?.code !== 200) throw new Error(response?.message || "删除容器失败");
                        const removed = new Set(pending.map((record) => record.id));
                        tableRef.current?.setSelectedRowKeys((tableRef.current?.getSelectedRowKeys() || [])
                            .filter((key) => !removed.has(String(key))));
                        const skipped = records.length - pending.length;
                        message.success(`删除容器成功${pending.length > 1 ? ` ${pending.length} 项` : ""}${skipped ? `，跳过 ${skipped} 项` : ""}`);
                    } catch (error) {
                        if (!mountedRef.current) return;
                        const reason = error instanceof Error ? error.message : "删除容器失败";
                        if (pending.length === 1) message.error(reason);
                        else modal.warning({
                            title: "容器删除失败",
                            zIndex: 1100,
                            content: <div style={{maxHeight: 280, overflow: "auto", whiteSpace: "pre-wrap", overflowWrap: "anywhere"}}>{reason}</div>,
                            okText: "知道了",
                        });
                    }
                    return;
                }
                for (const record of pending) {
                    try {
                        const response = await operations[action].request({body: {id: record.id}});
                        if (response?.code !== 200) throw new Error(response?.message || "操作失败");
                        succeeded += 1;
                    } catch (error) {
                        failures.push({record, reason: error instanceof Error ? error.message : "操作失败"});
                    }
                }
                if (!mountedRef.current) return;
                const completed = new Set(pending.map((record) => record.id));
                failures.forEach(({record}) => completed.delete(record.id));
                tableRef.current?.setSelectedRowKeys((tableRef.current?.getSelectedRowKeys() || [])
                    .filter((key) => !completed.has(String(key))));
                reload();
                const skipped = records.length - pending.length;
                if (!failures.length) {
                    message.success(`${operations[action].label}成功${pending.length > 1 ? ` ${succeeded} 项` : ""}${skipped ? `，跳过 ${skipped} 项` : ""}`);
                } else if (records.length === 1) {
                    message.error(failures[0].reason);
                } else {
                    modal.warning({
                        title: `${operations[action].label}完成：成功 ${succeeded} 项，失败 ${failures.length} 项${skipped ? `，跳过 ${skipped} 项` : ""}`,
                        zIndex: 1100,
                        content: <div style={{maxHeight: 280, overflow: "auto"}}>
                            {failures.map(({record, reason}) => <Typography.Paragraph key={record.id}>
                                {record.name || record.id}：{reason}
                            </Typography.Paragraph>)}
                        </div>,
                        okText: "知道了",
                    });
                }
            } finally {
                pending.forEach((record) => busyRef.current.delete(record.id));
                if (mountedRef.current) {
                    setBusy(new Set(busyRef.current));
                    // 部分容器删除失败时也刷新，移除已成功删除的数据和选中项。
                    if (action === "remove") reload();
                }
            }
        };
        const operation = operations[action];
        confirmRef.current?.destroy();
        const confirmation = modal.confirm({
            title: targets.length === 1
                ? `${operation.okText}“${targets[0].name || targets[0].id}”`
                : `${operation.okText} ${targets.length} 个容器`,
            content: operation.content,
            zIndex: 1100,
            okText: operation.okText,
            cancelText: "取消",
            okButtonProps: {danger: action === "remove"},
            onOk: execute,
            afterClose: () => {
                if (confirmRef.current === confirmation) confirmRef.current = null;
            },
        });
        confirmRef.current = confirmation;
    }, [message, modal, reload]);

    const columns = useMemo<ModalTableProps<any>["columns"]>(() => [
        {
            title: "ID", dataIndex: "id", key: "id", width: 145, sorter: true,
            render: (value: string) => <Tooltip title={value}><span className="container-text container-muted">{value}</span></Tooltip>,
        },
        {
            title: "名称", dataIndex: "name", key: "name", width: 170, sorter: true,
            render: (name: string, record: any) => <button type="button" className="container-name"
                title={name || record.id} onClick={() => detailRef.current?.open(record)}>{name || record.id}</button>,
        },
        {
            title: "镜像", dataIndex: "image_name", key: "image_name", width: 200,
            render: (name: string, record: any) => <button type="button" className="container-name"
                aria-label={`查看镜像 ${name || record.image_id} 的详情`}
                onClick={() => void imagesRef.current?.openDetail(record.image_id)}>{name || record.image_id}</button>,
        },
        {
            title: "运行状态", dataIndex: "status", key: "status", width: 110,
            render: (value: string) => <span className={`container-state ${value}`}>{statusData[value] || "未知状态"}</span>,
        },
        {
            title: "自动恢复", key: "auto_restart", width: 110,
            render: (_, record) => <Tooltip title={record.auto_restart
                ? record.manual_stop ? "已手动停止，启动后恢复自动重启" : "跟随面板启动，退出后自动重启"
                : "容器退出后不自动重启"}>
                <span className={record.auto_restart && !record.manual_stop ? "" : "container-muted"}>
                    {record.auto_restart ? record.manual_stop ? "已暂停" : "已开启" : "未开启"}
                </span>
            </Tooltip>,
        },
        {
            title: "创建时间", dataIndex: "create_time", key: "create_time", width: 165, sorter: true,
            render: (value: string) => <span className="container-time">{value || "—"}</span>,
        },
        {
            title: "操作", key: "action", width: 82, fixed: "right", className: "action-cell",
            render: (_, record) => {
                const changing = ["starting", "stopping"].includes(record.status);
                return <ActionDropdown disabled={busy.has(record.id)}
                    menu={{items: [
                        {key: "detail", label: "容器详情", onClick: () => detailRef.current?.open(record)},
                        {key: "resources", label: "资源状态", onClick: () => detailRef.current?.open(record, "resources")},
                        {key: "terminal", label: "终端管理", disabled: record.status !== "running", onClick: () => openTerminal(record)},
                        {key: "log", label: "容器日志", onClick: () => detailRef.current?.open(record, "log")},
                        {key: "exec", label: "执行命令", disabled: record.status !== "running", onClick: () => openExec(record)},
                        {type: "divider"},
                        {key: "start", label: operations.start.label, disabled: changing || record.status === "running", onClick: () => operate("start", [record])},
                        {key: "stop", label: operations.stop.label, disabled: changing || (record.status === "stopped" && (!record.auto_restart || record.manual_stop)), onClick: () => operate("stop", [record])},
                        {key: "restart", label: operations.restart.label, disabled: changing, onClick: () => operate("restart", [record])},
                        {key: "remove", label: operations.remove.label, danger: true, disabled: changing, onClick: () => operate("remove", [record])},
                    ]}}>
                    <Button size="small" loading={busy.has(record.id)} aria-label={`操作 ${record.name || record.id}`}>操作</Button>
                </ActionDropdown>;
            },
        },
    ], [busy, openExec, openTerminal, operate, statusData]);

    return <>
        <ModalTable<any>
            ref={tableRef}
            title={<Title>容器管理</Title>}
            open={open}
            onOpenChange={(visible) => {
                if (!visible && busyRef.current.size) return;
                if (!visible) {
                    confirmRef.current?.destroy();
                    setTerminal(null);
                }
                setOpen(visible);
            }}
            width={compact ? 800 : 910}
            minHeight={compact ? 420 : 460}
            scrollMode="modal"
            className={styles.table}
            ariaLabel="容器管理"
            loading={enumLoading || undefined}
            rowKey="id"
            columns={columns}
            scroll={{x: 1050}}
            requestEnabled={open}
            request={async (params, sort) => {
                const response = await getContainerList({body: getParams(params, sort)});
                if (response?.code !== 200) throw new Error(response?.message || "获取容器列表失败");
                return {data: Array.isArray(response?.data?.list) ? response?.data?.list.filter(Boolean) : [], total: response?.data?.total ?? 0, success: true};
            }}
            defaultPageSize={20}
            search={{
                placeholder: "搜索容器", maxLength: 100,
                fields: [
                    {value: "keyword", label: "全部", placeholder: "搜索名称、ID 或镜像"},
                    {value: "name", label: "名称", placeholder: "搜索容器名称"},
                    {value: "id", label: "ID", placeholder: "输入容器 ID", maxLength: 64},
                    {value: "image_name", label: "镜像", placeholder: "搜索镜像名称", maxLength: 256},
                ],
            }}
            filters={[
                {type: "select", name: "status", label: "运行状态筛选", icon: "FlagOutlined", allLabel: "全部状态", valueEnum: statusData},
                {type: "date", name: "create_time", label: "创建时间筛选"},
            ]}
            toolbar={{
                refresh: {ariaLabel: "刷新容器列表", disabled: busy.size > 0},
                actions: [
                    {key: "create", type: "button", label: "创建容器", icon: "PlusOutlined", onClick: () => openRun()},
                    {key: "images", type: "button", label: "镜像管理", icon: "AppstoreOutlined", onClick: () => imagesRef.current?.open()},
                ],
            }}
            rowSelection={{
                columnWidth: 44,
                getCheckboxProps: (record) => ({disabled: busy.has(record.id)}),
                clearDisabled: busy.size > 0,
                actions: (_, rows) => (Object.keys(operations) as (keyof typeof operations)[]).map((action) => ({
                    key: action, type: "button", label: operations[action].label, danger: action === "remove",
                    disabled: busy.size > 0, onClick: () => operate(action, rows),
                })),
            }}
            modalProps={{
                zIndex: 1000,
                closable: busy.size === 0,
                keyboard: busy.size === 0,
                mask: {closable: busy.size === 0},
                style: {
                    top: screens.md ? 100 : 24,
                    paddingBottom: screens.md ? 100 : 24,
                    maxWidth: "calc(100vw - 24px)",
                },
            }}
        />
        {runMounted && <RunContainer ref={runRef} onChange={reload}/>}
        <ContainerDetail ref={detailRef}/>
        {execMounted && <ContainerExec ref={execRef} onChange={reload}/>}
        <ImagesManager ref={imagesRef} onRun={openRun}/>
        {terminalMounted && <Suspense fallback={null}>
            <ContainerTerminal id={terminal?.id || ""} open={Boolean(terminal)} onClose={closeTerminal}/>
        </Suspense>}
    </>;
});

ContainerManager.displayName = "ContainerManager";
export default memo(ContainerManager);
