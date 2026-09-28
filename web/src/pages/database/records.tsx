import {useCallback, useEffect, useRef, useState} from "react";
import {App, Button, Empty, Form, Input, Select, Space, Spin, Typography} from "antd";
import {PageTable, ProModal, Title} from "sinking-antd";
import type {PageTableProps, PageTableRef} from "sinking-antd";
import {databaseDatabases, databaseUsers} from "@/service/api/database";
import {formatSize} from "@/utils/string";
import useStyles from "./styles";

const systemDatabases: Record<string, string[]> = {
    mysql: ["mysql", "sys", "information_schema", "performance_schema"],
    pgsql: ["postgres", "template0", "template1"],
    mongo: ["admin", "local", "config"],
};

export default ({kind, mode, ready, administrator, onBusy}: {
    kind: string; mode: "databases" | "users"; ready: boolean; administrator: string; onBusy: (value: boolean) => void;
}) => {
    const {message, modal} = App.useApp();
    const {styles} = useStyles();
    const [form] = Form.useForm();
    const [editor, setEditor] = useState<{action: "create" | "password"; record?: any}>();
    const [databases, setDatabases] = useState<{value: string; label: string}[]>([]);
    const [loadingDatabases, setLoadingDatabases] = useState(false);
    const [busy, setBusy] = useState(false);
    const tableRef = useRef<PageTableRef<any>>(null);
    const mounted = useRef(true);
    const pending = useRef(false);
    const confirmation = useRef<{destroy: () => void} | null>(null);
    const editingUser = mode === "users";
    const protectedUser = (record: any) => record.protected || record.name === administrator || record.name === "root"
        || (kind === "mysql" && record.name.startsWith("mysql.")) || (kind === "pgsql" && record.name.startsWith("pg_"));

    useEffect(() => {
        mounted.current = true;
        return () => { mounted.current = false; confirmation.current?.destroy(); };
    }, []);

    useEffect(() => {
        if (!editor || !editingUser || editor.action !== "create") return;
        let disposed = false;
        setLoadingDatabases(true);
        void databaseDatabases(kind, {body: {action: "list"}}).then((response) => {
            if (disposed) return;
            if (response?.code !== 200) throw new Error(response?.message || "获取数据库列表失败");
            setDatabases((response.data?.list || []).filter((record: any) => !systemDatabases[kind].includes(record.name))
                .map((record: any) => ({value: record.name, label: record.name})));
        }).catch((reason) => {
            if (!disposed) message.error(reason instanceof Error ? reason.message : "获取数据库列表失败");
        }).finally(() => { if (!disposed) setLoadingDatabases(false); });
        return () => { disposed = true; };
    }, [editor, editingUser, kind, message]);

    const request = useCallback(async (params: {current: number; pageSize: number; [key: string]: unknown}) => {
        const response = mode === "users"
            ? await databaseUsers(kind, {body: {action: "list"}})
            : await databaseDatabases(kind, {body: {action: "list"}});
        if (response?.code !== 200) throw new Error(response?.message || "获取列表失败");
        const keyword = String(params.keyword || "").trim().toLowerCase();
        const rows = (response.data?.list || []).filter((record: any) => !keyword || record.name.toLowerCase().includes(keyword));
        return {data: rows.slice((params.current - 1) * params.pageSize, params.current * params.pageSize), total: rows.length, success: true};
    }, [kind, mode]);

    const open = (action: "create" | "password", record?: any) => {
        form.resetFields();
        form.setFieldsValue({host: "%", ...record, password: ""});
        setDatabases([]);
        setEditor({action, record});
    };

    const execute = async (task: () => Promise<any>) => {
        if (pending.current || !mounted.current) return false;
        pending.current = true;
        setBusy(true);
        onBusy(true);
        try {
            const response = await task();
            if (!mounted.current) return false;
            if (response?.code !== 200) throw new Error(response?.message || "操作失败");
            message.success("操作成功");
            tableRef.current?.reload();
            return true;
        } catch (reason) {
            if (mounted.current) message.error(reason instanceof Error ? reason.message : "操作失败");
            return false;
        } finally {
            pending.current = false;
            onBusy(false);
            if (mounted.current) setBusy(false);
        }
    };

    const save = async (values: any) => {
        if (!editor) return;
        const success = await execute(() => editingUser
            ? databaseUsers(kind, {body: {...editor.record, ...values, action: editor.action}})
            : databaseDatabases(kind, {body: {action: "create", name: values.name.trim()}}));
        if (success && mounted.current) setEditor(undefined);
    };

    const remove = (record: any) => {
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({
            title: `删除${editingUser ? "账户" : "数据库"}“${record.name}”`,
            content: editingUser ? "删除后，该账户将无法继续连接数据库。" : "此操作会永久删除数据库及其中的数据，无法恢复。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                const success = await execute(() => editingUser
                    ? databaseUsers(kind, {body: {...record, action: "delete"}})
                    : databaseDatabases(kind, {body: {action: "delete", name: record.name}}));
                if (!success) throw new Error("删除失败");
            },
        });
    };

    const columns: PageTableProps<any>["columns"] = editingUser ? [
        {title: "账户", dataIndex: "name", width: 160, ellipsis: true},
        ...(kind === "mysql" ? [{title: "来源主机", dataIndex: "host", width: 150, ellipsis: true}] : []),
        ...(kind !== "mysql" ? [{title: "数据库", dataIndex: "database", width: 160, ellipsis: true, render: (value: string) => value || "—"}] : []),
        {title: "权限", dataIndex: "roles", width: 200, ellipsis: true, render: (value: string, record) => value
            || (protectedUser(record) ? "管理员" : record.database ? "数据库读写" : "业务账户")},
        {title: "操作", key: "actions", width: 154, fixed: "right", render: (_, record) => protectedUser(record)
            ? <Typography.Text type="secondary">系统账户</Typography.Text>
            : <Space size={0}><Button variant="link" color="primary" disabled={busy} onClick={() => open("password", record)}>改密</Button>
                <Button variant="link" color="danger" disabled={busy} onClick={() => remove(record)}>删除</Button></Space>},
    ] : [
        {title: "数据库", dataIndex: "name", width: 210, ellipsis: true},
        {title: "占用空间", dataIndex: "size", width: 130, render: (value?: number) => value === undefined ? "—" : formatSize(value)},
        ...(kind === "pgsql" ? [{title: "所属账户", dataIndex: "owner", width: 160, ellipsis: true}] : []),
        {title: "操作", key: "actions", width: 104, fixed: "right", render: (_, record) => (record.protected || systemDatabases[kind].includes(record.name))
            ? <Typography.Text type="secondary">系统库</Typography.Text>
            : <Button variant="link" color="danger" disabled={busy} onClick={() => remove(record)}>删除</Button>},
    ];

    if (!ready) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="服务就绪后可管理数据库和账户"/>;

    return <>
        <PageTable<any> ref={tableRef} hero={false} rowKey={(record) => `${record.name}:${record.host || ""}:${record.database || ""}`}
            columns={columns} request={request} defaultPageSize={20} scroll={{x: editingUser ? 690 : 520}}
            search={{placeholder: editingUser ? "搜索账户" : "搜索数据库", maxLength: 100}}
            toolbar={{actions: [{key: "create", type: "button", label: editingUser ? "添加账户" : "添加数据库", disabled: busy, onClick: () => open("create")}], refresh: {disabled: busy}}}/>
        <ProModal title={<Title>{editor?.action === "password" ? "修改密码" : editingUser ? "添加账户" : "添加数据库"}</Title>}
            width={430} onCancel={() => { if (!busy) setEditor(undefined); }} modalProps={{
                open: Boolean(editor), destroyOnHidden: true, mask: {closable: !busy}, closable: !busy,
                confirmLoading: busy, okText: editor?.action === "password" ? "保存" : "添加", cancelText: "取消", onOk: () => form.submit(),
                style: {top: 100, maxWidth: "calc(100vw - 24px)"},
            } as any}>
            <Form form={form} className={styles.form} layout="vertical" variant="filled" onFinish={save} disabled={busy}>
                <Form.Item name="name" label={editingUser ? "账户名称" : "数据库名称"} rules={[
                    {required: true, whitespace: true, message: "请输入名称"},
                    {pattern: /^[_A-Za-z][_A-Za-z0-9-]{0,62}$/, message: "以字母或下划线开头，仅支持字母、数字、下划线和短横线"},
                    {max: editingUser && kind === "mysql" ? 32 : 63, message: `最多 ${editingUser && kind === "mysql" ? 32 : 63} 个字符`},
                ]}>
                    <Input disabled={busy || editor?.action === "password"} maxLength={editingUser && kind === "mysql" ? 32 : 63} autoComplete="off"/>
                </Form.Item>
                {editingUser && <>
                    <Form.Item name="password" label="密码" rules={[
                        {required: true, message: "请输入密码"},
                        {min: 8, max: 256, message: "密码长度为 8–256 个字符"},
                        {pattern: /^[^\r\n\x00]*$/, message: "密码不能包含换行或空字符"},
                    ]}>
                        <Input.Password maxLength={256} autoComplete="new-password"/>
                    </Form.Item>
                    {editor?.action === "create" && <>
                        <Form.Item name="database" label="授权数据库" rules={[{required: true, message: "请选择数据库"}]}>
                            <Select loading={loadingDatabases} options={databases} placeholder="请选择数据库" notFoundContent={loadingDatabases ? <Spin/> : "请先添加数据库"}/>
                        </Form.Item>
                        {kind === "mysql" && <Form.Item name="host" label="来源主机" tooltip="localhost 仅本机，% 允许任意来源，也可指定 IP" rules={[{required: true, whitespace: true, message: "请输入来源主机"}]}>
                            <Input placeholder="localhost、IP 或 %"/>
                        </Form.Item>}
                    </>}
                </>}
            </Form>
        </ProModal>
    </>;
};
