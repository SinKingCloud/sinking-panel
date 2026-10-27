import {useCallback, useEffect, useRef, useState} from "react";
import {App, Button, Empty, Form, Grid, Input, Tag} from "antd";
import {Body, Icon, PageTable, ProModal, Title} from "sinking-antd";
import type {PageTableRef} from "sinking-antd";
import ActionDropdown from "@/pages/components/action-dropdown";
import {uninstallSoftware, getSoftwareList, updateSoftware} from "@/service/api/software";
import {getParams} from "@/utils/page";
import useEnum from "@/utils/enum";
import ContainerManager from "./components/container";
import InstallSoftware from "./install";
import SoftwareWorkspace from "./workspace";
import useStyles from "./styles";

const processing = (record: any) => Number(record.status) === 4;

export default () => {
    const {message, modal} = App.useApp();
    const screens = Grid.useBreakpoint();
    const [enumData] = useEnum("software");
    const {styles} = useStyles();
    const containerRef = useRef<any>(null);
    const tableRef = useRef<PageTableRef<any>>(null);
    const confirmRef = useRef<{destroy: () => void} | null>(null);
    const active = useRef(true);
    const pending = useRef(false);
    const [form] = Form.useForm();
    const [install, setInstall] = useState(false);
    const [workspace, setWorkspace] = useState(false);
    const [edit, setEdit] = useState(false);
    const [record, setRecord] = useState<any>();
    const [editing, setEditing] = useState<any>();
    const [busy, setBusy] = useState(false);
    const [polling, setPolling] = useState(false);
    const [error, setError] = useState("");
    useEffect(() => { active.current = true; return () => { active.current = false; confirmRef.current?.destroy(); }; }, []);
    useEffect(() => {
        if (!polling) return;
        const timer = window.setInterval(() => { if (document.visibilityState === "visible") tableRef.current?.reload(); }, 3000);
        return () => window.clearInterval(timer);
    }, [polling]);
    const reload = useCallback(() => tableRef.current?.reload(), []);
    const show = (item: any) => {
        if (processing(item)) { message.info("软件任务正在执行，请稍候"); return; }
        if (Number(item.status) !== 2) { message.info("软件状态异常，请重新安装或重试卸载"); return; }
        setRecord(item); setWorkspace(true);
    };
    const remove = (item: any) => {
        if (pending.current) return;
        confirmRef.current?.destroy();
        confirmRef.current = modal.confirm({title: `卸载“${item.name}”？`, zIndex: 1150, okText: "卸载", cancelText: "取消", okButtonProps: {danger: true},
            content: "将卸载软件并删除插件数据，此操作无法撤销。",
            onOk: async () => {
                if (pending.current) return;
                pending.current = true; setBusy(true);
                try {
                    const response = await uninstallSoftware({body: {id: item.id}});
                    if (!active.current) return;
                    if (response?.code !== 200) throw new Error(response?.message || "卸载软件失败");
                    message.success("卸载成功"); reload();
                } catch (reason: any) { if (active.current) message.error(reason?.message || "卸载失败"); }
                finally { pending.current = false; if (active.current) setBusy(false); }
            }});
    };
    const rename = (item: any) => { setEditing(item); form.setFieldsValue({name: item.name}); setEdit(true); };
    const save = async ({name}: any) => {
        if (pending.current) return;
        pending.current = true; setBusy(true);
        try {
            const response = await updateSoftware({body: {id: editing.id, name: name.trim()}});
            if (!active.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "修改失败");
            message.success("保存成功"); setEdit(false); reload();
        } catch (reason: any) { if (active.current) message.error(reason?.message || "保存失败"); }
        finally { pending.current = false; if (active.current) setBusy(false); }
    };
    return <Body>
        <PageTable<any> ref={tableRef} ariaLabel="软件管理" rowKey="id" defaultPageSize={20} paginationAffix={{offsetBottom: 15}}
            scroll={{x: 860}} request={async (params, sort) => {
                const response = await getSoftwareList({body: getParams(params, sort)});
                if (response?.code !== 200 || !Array.isArray(response?.data?.list)) throw new Error(response?.message || "获取软件列表失败");
                return {data: response?.data?.list.filter(Boolean), total: response?.data?.total ?? 0, success: true};
            }} onLoad={(rows) => { setError(""); setPolling(rows.some(processing)); }}
            onRequestError={(reason) => { const text = reason?.message || "获取软件列表失败"; setError(text); message.error(text); }}
            emptyContent={error ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="数据加载失败"/> : undefined}
            search={{placeholder: "搜索软件名称或标识", maxLength: 100}}
            columns={[
                {title: "软件名称", dataIndex: "name", key: "name", width: 220, sorter: true, render: (value: string, item: any) => <div className={styles.name}>
                    <Icon type="AppstoreOutlined"/><button type="button" onClick={() => show(item)}>{value}</button>
                </div>},
                {title: "标识", dataIndex: "code", key: "code", width: 150, ellipsis: true},
                {title: "版本", dataIndex: "version", key: "version", width: 100, ellipsis: true, render: (value: string) => value || "—"},
                {title: "安装状态", dataIndex: "status", key: "status", width: 120, render: (value: number) => <Tag variant="filled"
                    color={Number(value) === 2 ? "success" : Number(value) === 4 ? "processing" : "default"}>{enumData.status?.[value] || value}</Tag>},
                {title: "更新时间", dataIndex: "update_time", key: "update_time", width: 170, sorter: true},
                {title: "操作", key: "action", width: 86, fixed: "right", align: "center", render: (_, item) => <ActionDropdown
                    disabled={busy || processing(item)} menu={{items: [
                        {key: "manage", label: "管理", onClick: () => show(item)},
                        {key: "rename", label: "修改", onClick: () => rename(item)},
                        {key: "delete", label: "卸载", danger: true, onClick: () => remove(item)},
                    ]}}><Button size="small" disabled={busy || processing(item)}>操作</Button></ActionDropdown>},
            ]}
            toolbar={{actions: [
                {key: "containers", type: "button", label: "容器管理", icon: "CodeSandboxOutlined", onClick: () => containerRef.current?.open()},
            ], refresh: true}}
            hero={{eyebrow: "SOFTWARE MANAGER", title: "软件管理", action: {label: "添加软件", onClick: () => setInstall(true)}}}/>
        <ContainerManager ref={containerRef}/>
        <InstallSoftware open={install} onCancel={() => setInstall(false)} onSuccess={() => {
            if (!active.current) return;
            setInstall(false); message.success("软件安装成功"); reload();
        }}/>
        <SoftwareWorkspace open={workspace} software={record} onCancel={() => setWorkspace(false)}/>
        <ProModal title={<Title>修改软件</Title>} onOk={form.submit} onCancel={() => { if (!busy) setEdit(false); }} modalProps={{
            open: edit, forceRender: true, width: 380, confirmLoading: busy, mask: {closable: !busy}, closable: !busy,
            okText: "保存", cancelText: "取消", style: {top: screens.md ? 90 : 24, maxWidth: "calc(100vw - 24px)"},
        } as any}>
            <Form form={form} layout="vertical" variant="filled" disabled={busy} onFinish={save}>
                <Form.Item name="name" label="软件名称" rules={[{required: true, whitespace: true, message: "请输入软件名称"}, {max: 100, message: "最多 100 个字符"}]}>
                    <Input placeholder="请输入软件名称" maxLength={100}/>
                </Form.Item>
            </Form>
        </ProModal>
    </Body>;
};
