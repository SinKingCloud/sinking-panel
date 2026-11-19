import {forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Button, Form, Grid} from "antd";
import {ModalTable, ProModal, Title, useTheme} from "sinking-antd";
import type {ModalTableProps, ModalTableRef} from "sinking-antd";
import {useOutletContext} from "umi";
import type {LayoutContext} from "@/layouts/context";
import ActionDropdown from "@/pages/components/ui/action-dropdown";
import SiteSelect from "@/pages/components/site-select";
import Log, {LogRef} from "@/pages/components/task-log";
import {formatFileSize} from "@/pages/file/utils";
import {getParams} from "@/utils/page";
import {filePreviewUrl} from "@/service/api/file";
import {
    createSiteBackup,
    deleteSiteBackups,
    downloadSiteBackup,
    getSiteBackupLog,
    clearSiteBackupLog,
    getSiteBackups,
    restoreSiteBackup,
} from "@/service/api/site";
import {systemTask} from "@/service/api/system";
import BackupSettings, {BackupSettingsRef} from "./settings";

const requestBackupLog = (params: API.RequestParams) => {
    const {id: _id, action, ...body} = params.body || {};
    return (action === "clear" ? clearSiteBackupLog : getSiteBackupLog)({...params, body});
};

export interface BackupManagerRef {
    open: (site?: any) => void;
}

const BackupForm = ({open, onClose, afterClose, onAccepted}: any) => {
    const {message} = App.useApp();
    const screens = Grid.useBreakpoint();
    const compact = Boolean(useTheme()?.isCompactTheme?.());
    const [form] = Form.useForm();
    const pending = useRef(false);
    const mounted = useRef(true);
    const [saving, setSaving] = useState(false);
    useEffect(() => {
        mounted.current = true;
        return () => { mounted.current = false; };
    }, []);

    const submit = async (values: any) => {
        if (pending.current) return;
        pending.current = true;
        setSaving(true);
        try {
            const response = await createSiteBackup({body: {site_ids: values.site_ids}});
            if (!mounted.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "提交任务失败");
            const taskId = response?.data?.task_id;
            if (!taskId) throw new Error("未获取到后台任务，请重试");
            onAccepted(String(taskId));
            onClose();
            message.success("备份任务已提交");
        } catch (error) {
            if (mounted.current) message.error(error instanceof Error ? error.message : "提交任务失败");
        } finally {
            pending.current = false;
            if (mounted.current) setSaving(false);
        }
    };

    return <ProModal title={<Title>备份网站</Title>} width={compact ? 350 : 380}
                     onOk={form.submit} onCancel={() => { if (!pending.current) onClose(); }} modalProps={{
        open,
        zIndex: 1100,
        afterClose,
        okText: "备份",
        cancelText: "取消",
        confirmLoading: saving,
        cancelButtonProps: {disabled: saving},
        closable: !saving,
        keyboard: !saving,
        mask: {closable: !saving},
        style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"},
    } as any}>
        <Form form={form} layout="vertical" variant="filled" disabled={saving} onFinish={submit} initialValues={{site_ids: []}}>
            <Form.Item name="site_ids" label="选择网站"
                       rules={[{required: true, type: "array", min: 1, max: 1000, message: "请选择需要备份的网站",
                           defaultField: {type: "number", min: 1}}]}>
                <SiteSelect multiple requireRoot placeholder="请选择网站" zIndex={1160}/>
            </Form.Item>
        </Form>
    </ProModal>;
};

const BackupManager = forwardRef<BackupManagerRef>((_, ref) => {
    const {message, modal} = App.useApp();
    const layout = useOutletContext<LayoutContext | undefined>();
    const screens = Grid.useBreakpoint();
    const compact = Boolean(useTheme()?.isCompactTheme?.());
    const tableRef = useRef<ModalTableRef<any> | null>(null);
    const settingsRef = useRef<BackupSettingsRef | null>(null);
    const logRef = useRef<LogRef | null>(null);
    const confirmRef = useRef<{destroy: () => void} | null>(null);
    const busyRef = useRef(false);
    const mounted = useRef(true);
    const tracked = useRef(new Map<string, number>());
    const [open, setOpen] = useState(false);
    const [session, setSession] = useState(0);
    const [site, setSite] = useState<any>();
    const [formMounted, setFormMounted] = useState(false);
    const [formOpen, setFormOpen] = useState(false);
    const [busy, setBusy] = useState(false);
    const [taskCount, setTaskCount] = useState(0);
    const siteId = Number(site?.id) || 0;
    const singleSite = siteId > 0;

    useEffect(() => {
        mounted.current = true;
        return () => { mounted.current = false; confirmRef.current?.destroy(); };
    }, []);
    useImperativeHandle(ref, () => ({open: (record) => {
        if (busyRef.current) return;
        setSite(record);
        setSession((value) => value + 1);
        setOpen(true);
    }}), []);

    // 仅在备份窗口可见时跟踪本窗口提交的任务，完成后刷新列表。
    useEffect(() => {
        if (!open || !taskCount) return;
        let stopped = false;
        let timer = 0;
        const poll = async () => {
            if (!document.hidden) {
                const results = await Promise.all([...tracked.current.keys()].map(async (id) => {
                    try { return {id, response: await systemTask({body: {id, action: "info"}})}; }
                    catch { return {id, response: undefined}; }
                }));
                if (stopped) return;
                let completed = false;
                for (const {id, response} of results) {
                    if (!tracked.current.has(id)) continue;
                    const status = Number(response?.data?.status);
                    if (response?.code === 200 && [0, 1, 2, 3, 4].includes(status)) {
                        tracked.current.set(id, 0);
                        if (status < 2) continue;
                        tracked.current.delete(id);
                        completed = true;
                        if (status === 3) message.error(response?.data?.message || "备份任务执行失败，请查看后台任务日志");
                    } else if (response?.code !== 200 && response?.message === "任务不存在") {
                        const attempts = (tracked.current.get(id) || 0) + 1;
                        if (attempts >= 3) { tracked.current.delete(id); completed = true; }
                        else tracked.current.set(id, attempts);
                    }
                }
                if (completed) tableRef.current?.reload();
                setTaskCount(tracked.current.size);
            }
            if (tracked.current.size) timer = window.setTimeout(poll, 3000);
        };
        void poll();
        return () => { stopped = true; window.clearTimeout(timer); };
    }, [message, open, taskCount]);

    const accepted = useCallback((id: string) => {
        tracked.current.set(id, 0);
        setTaskCount(tracked.current.size);
        layout?.task?.open(id);
    }, [layout?.task]);
    const confirmAction = useCallback((action: "create" | "restore" | "delete" | "download", records: any[]) => {
        if (busyRef.current || !records.length) return;
        const creating = action === "create";
        if (creating && !String(records[0]?.root || "").trim()) {
            message.warning("该网站未设置网站目录，无法备份");
            return;
        }
        confirmRef.current?.destroy();
        const restoring = action === "restore";
        const deleting = action === "delete";
        confirmRef.current = modal.confirm({
            title: creating ? "备份网站" : restoring ? "恢复网站" : deleting ? "删除备份" : "下载备份",
            content: creating ? `确定备份网站“${records[0]?.name || "当前网站"}”吗？` : restoring
                ? `确定将“${records[0]?.name || "此备份"}”恢复到原网站“${records[0]?.site_name || records[0]?.data?.site_name || `网站 #${records[0]?.site_id}`}”吗？将替换网站目录中的文件，网站数据库不会恢复。` : deleting
                ? `确定删除${records.length > 1 ? `所选 ${records.length} 份备份` : `“${records[0]?.name || "此备份"}”`}吗？本地和云端副本会一并删除，无法恢复。`
                : `确定下载“${records[0]?.name || "此备份"}”吗？`,
            zIndex: 1140,
            okText: creating ? "备份" : restoring ? "恢复" : deleting ? "删除" : "下载",
            cancelText: "取消",
            okButtonProps: {danger: deleting || restoring},
            mask: {closable: true},
            onOk: async () => {
                if (busyRef.current || !mounted.current) return;
                busyRef.current = true;
                setBusy(true);
                try {
                    const response = await (creating ? createSiteBackup : restoring ? restoreSiteBackup : deleting ? deleteSiteBackups : downloadSiteBackup)({body: creating
                        ? {site_ids: [records[0]?.id]} : deleting ? {ids: records.map((record) => record?.id)} : {id: records[0]?.id}});
                    if (!mounted.current) return;
                    if (response?.code !== 200) throw new Error(response?.message || "操作失败");
                    if (creating || restoring) {
                        const taskId = response?.data?.task_id;
                        if (!taskId) throw new Error("未获取到后台任务，请重试");
                        accepted(String(taskId));
                        message.success(response?.message || (restoring ? "恢复任务已提交" : "备份任务已提交"));
                    } else if (deleting) {
                        tableRef.current?.clearSelectedRows();
                        tableRef.current?.reload();
                        message.success(response?.message || "备份已删除");
                    } else {
                        const sign = response?.data;
                        if (typeof sign !== "string" || !sign) throw new Error("获取下载签名失败");
                        const anchor = document.createElement("a");
                        anchor.href = /^https?:\/\//i.test(sign) ? sign : filePreviewUrl(sign);
                        anchor.download = records[0]?.name || "backup.tar.gz";
                        anchor.rel = "noopener noreferrer";
                        anchor.referrerPolicy = "no-referrer";
                        document.body.appendChild(anchor);
                        anchor.click();
                        anchor.remove();
                    }
                } catch (error) {
                    if (mounted.current) {
                        message.error(error instanceof Error ? error.message : "操作失败");
                        if (deleting) tableRef.current?.reload();
                    }
                } finally {
                    busyRef.current = false;
                    if (mounted.current) setBusy(false);
                }
            },
        });
    }, [accepted, message, modal]);

    const columns = useMemo<ModalTableProps<any>["columns"]>(() => [
        {title: "备份名称", dataIndex: "name", width: 240, ellipsis: true},
        ...(!singleSite ? [{title: "网站名称", dataIndex: "site_name", width: 170, ellipsis: true,
            render: (value: any, record: any) => value || record?.data?.site_name || "网站已删除"}] : []),
        {title: "大小", dataIndex: "size", width: 110, render: (value) => formatFileSize(value)},
        {title: "存储位置", key: "storage", width: 125, render: (_, record) =>
            [record?.data?.local_path ? "本地" : "", record?.data?.oss_path ? "云端" : ""].filter(Boolean).join("、") || "—"},
        {title: "备份时间", dataIndex: "create_time", width: 170, sorter: true, ellipsis: true},
        {title: "操作", key: "action", width: 82, fixed: "right", align: "center", className: "action-cell",
            render: (_, record) => <ActionDropdown disabled={busy} menu={{items: [
                {key: "download", label: "下载", onClick: () => confirmAction("download", [record])},
                {key: "restore", label: "恢复", onClick: () => confirmAction("restore", [record])},
                {key: "delete", label: "删除", danger: true, onClick: () => confirmAction("delete", [record])},
            ]}}><Button size="small" disabled={busy}>操作</Button></ActionDropdown>},
    ], [busy, confirmAction, singleSite]);
    const request = useCallback<NonNullable<ModalTableProps<any>["request"]>>(async (params, sort) => {
        const response = await getSiteBackups({body: {...getParams({...params}, sort), site_id: siteId}});
        if (response?.code !== 200) throw new Error(response?.message || "获取备份列表失败");
        const rows = Array.isArray(response?.data?.list) ? response.data.list.filter(Boolean) : [];
        return {success: true, total: Number(response?.data?.total || 0), data: rows.map((row: any) => {
            let data = {};
            try { data = typeof row?.data === "string" ? JSON.parse(row.data) : row?.data || {}; } catch { /* 单条元信息异常不影响列表。 */ }
            return {...row, data};
        })};
    }, [siteId]);

    return <>
        <ModalTable<any> key={session} ref={tableRef} open={open} onCancel={() => {
            if (busyRef.current) return;
            confirmRef.current?.destroy();
            setOpen(false);
        }} title={<Title>备份管理</Title>} width={compact ? 800 : 910} minHeight={compact ? 440 : 480}
                        scrollMode="modal" columns={columns} rowKey="id" scroll={{x: singleSite ? 770 : 940}}
                        request={request} requestEnabled={open} defaultPageSize={10}
                        onRequestError={(error) => message.error(error.message)}
                        search={{maxLength: 255, debounce: 300, ...(singleSite ? {name: "name", placeholder: "搜索备份名称"} : {fields: [
                            {value: "keyword", label: "全部", placeholder: "搜索备份名称或网站名称"},
                            {value: "name", label: "备份名称", placeholder: "搜索备份名称"},
                            {value: "site_name", label: "网站名称", placeholder: "搜索网站名称"},
                        ]})}}
                        filters={[{type: "date", name: "create_time", label: "备份时间筛选"}]}
                        toolbar={{refresh: true, actions: [
                            {key: "create", type: "button", label: "备份", icon: "PlusOutlined",
                                onClick: () => {
                                    if (singleSite) confirmAction("create", [site]);
                                    else { setFormMounted(true); setFormOpen(true); }
                                }},
                            ...(!singleSite ? [
                                {key: "settings", type: "button" as const, label: "设置", icon: "SettingOutlined", onClick: () => settingsRef.current?.open()},
                                {key: "log", type: "button" as const, label: "日志", icon: "FileTextOutlined",
                                    onClick: () => logRef.current?.open({id: "site-auto-backup", name: "网站自动备份"})},
                            ] : []),
                        ]}}
                        rowSelection={{columnWidth: 44, clearDisabled: busy,
                            getCheckboxProps: () => ({disabled: busy}),
                            actions: (_, records) => [{key: "delete", type: "button", label: "删除", danger: true,
                                disabled: busy, onClick: () => confirmAction("delete", records)}]}}
                        modalProps={{zIndex: 1010, destroyOnHidden: true,
                            mask: {closable: !busy}, keyboard: !busy, closable: !busy,
                            style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"}}}/>
        {!singleSite && <>
            <BackupSettings ref={settingsRef}/>
            <Log ref={logRef} request={requestBackupLog} title="备份日志"/>
        </>}
        {formMounted && <BackupForm open={formOpen} onClose={() => setFormOpen(false)}
                                    afterClose={() => setFormMounted(false)} onAccepted={accepted}/>}
    </>;
});

BackupManager.displayName = "BackupManager";
export default BackupManager;
