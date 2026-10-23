import {useEffect, useRef, useState} from "react";
import {App, Form, Input, Select, Tooltip} from "antd";
import {AccountEditor, ConnectionDetails, DatabasePermissionFields} from "./account-editor";
import AccountsLayout from "./accounts-layout";
import AccountSelect from "./account-select";
import {useMysqlClient} from "./context";
import {canChangeAccess, canChangePassword} from "./client-common";

const profiles = [
    {value: "readonly", label: "只读访问", description: "查看数据，不能新增、修改或删除数据。", privileges: ["SELECT"]},
    {value: "readwrite", label: "读写数据", description: "读取、新增、修改和删除数据，不允许改动表结构。", privileges: ["SELECT", "INSERT", "UPDATE", "DELETE"]},
    {value: "manage", label: "数据库管理", description: "支持数据增删改查、表结构和索引管理，以及执行存储程序。", privileges: ["SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "INDEX", "DROP", "EXECUTE"]},
];
const protectedReason = "此账户不允许删除或调整权限。";

// MySQL 的 host 是账户来源许可，不是数据库服务的监听地址。
function AccessFields() {
    const form = Form.useFormInstance();
    const source = Form.useWatch("source", form);
    return <>
        <Form.Item name="source" label="访问来源" rules={[{required: true, message: "请选择访问来源"}]}>
            <Select placeholder="请选择访问来源" options={[{value: "local", label: "本地访问"}, {value: "address", label: "指定IP访问"}, {value: "remote", label: "任何IP可访问"}]}/>
        </Form.Item>
        {source === "address" && <Form.Item name="source_host" label="IP 地址" rules={[{required: true, whitespace: true, message: "请输入允许访问的 IP 地址"}]}>
            <Input maxLength={255} placeholder="例如 192.168.1.10"/>
        </Form.Item>}
    </>;
}
function sourceHost(values: any) { return values.source === "remote" ? "%" : values.source === "address" ? values.source_host.trim() : "localhost"; }

export default ({record, initialAction}: any) => {
    const {api, ActionDropdown} = useMysqlClient();
    const {message, modal} = App.useApp();
    const [form] = Form.useForm();
    const [rows, setRows] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);
    const [busy, setBusy] = useState(false);
    const [version, setVersion] = useState(0);
    const [editor, setEditor] = useState<any>();
    const [repair, setRepair] = useState("");
    const [query, setQuery] = useState({keyword: "", page: 1, pageSize: 20});
    const [pagination, setPagination] = useState({total: 0, page: 1, pageSize: 20});
    const database = record.database_name || record.name;
    const [permissions, setPermissions] = useState<any>();
    const [permissionsLoading, setPermissionsLoading] = useState(false);
    const [connection, setConnection] = useState<any>();
    const [connectionLoading, setConnectionLoading] = useState(false);
    const live = useRef(true);
    const pending = useRef(false);
    const listSequence = useRef(0);
    const listScope = useRef("");
    const confirmation = useRef<any>(undefined);

    useEffect(() => {
        live.current = true;
        return () => { live.current = false; listSequence.current++; confirmation.current?.destroy(); };
    }, []);
    useEffect(() => {
        const request = ++listSequence.current;
        const key = String(record.id);
        if (listScope.current !== key) {
            listScope.current = key; setRows([]); setPagination({total: 0, page: 1, pageSize: query.pageSize});
        }
        setLoading(true);
        void api.users(record, {action: "list", keyword: query.keyword, page: query.page, page_size: query.pageSize}).then((data) => {
            if (!live.current || request !== listSequence.current) return;
            if (!Array.isArray(data?.list) || !Number.isSafeInteger(data.total) || data.total < 0
                || !Number.isInteger(data.page) || data.page < 1 || !Number.isInteger(data.page_size) || data.page_size < 1) {
                throw new Error("账户列表返回的分页数据不完整");
            }
            setRows(data.list); setPagination({total: data.total, page: data.page, pageSize: data.page_size});
        }).catch((reason) => {
            if (live.current && request === listSequence.current) {
                message.error({key: `mysql-account-list-${record.id}`, content: reason.message || "加载账户列表失败"});
            }
        })
            .finally(() => { if (live.current && request === listSequence.current) setLoading(false); });
        return () => { listSequence.current++; };
    }, [record.id, query, version]);
    useEffect(() => {
        if (initialAction === "password") open("password");
    }, [initialAction]);
    useEffect(() => {
        if (editor?.action !== "create" && editor?.action !== "access") return;
        let current = true;
        setConnectionLoading(true);
        void api.users(record, {action: "connection"}).then((data) => {
            if (live.current && current) setConnection(data || {});
        }).catch((reason) => {
            if (!live.current || !current) return;
            setEditor(undefined);
            message.error(reason.message || "加载连接信息失败");
        })
            .finally(() => { if (live.current && current) setConnectionLoading(false); });
        return () => { current = false; };
    }, [record.id, editor]);
    useEffect(() => {
        if (editor?.action !== "create" && editor?.action !== "permissions") return;
        if (editor.permissionDraft) {
            // 授权中断后保留全部草稿；重试时由后端重新读取实际权限并计算差量。
            setPermissions({privileges: editor.permissionDraft.privileges, grants: [],
                uncertain: editor.permissionDraft.grants.map((grant: any) => grant.database)});
            form.setFieldsValue({grants: editor.permissionDraft.grants});
            setPermissionsLoading(false);
            return;
        }
        let current = true;
        setPermissions(undefined); setPermissionsLoading(true);
        void api.users(record, {action: "permissions", name: editor.account?.name, host: editor.account?.host}).then((data) => {
            if (!live.current || !current) return;
            if (!Array.isArray(data?.privileges) || !data.privileges.every((value: any) => typeof value === "string")
                || !Array.isArray(data.grants) || !data.grants.every((grant: any) => typeof grant?.database === "string"
                    && Array.isArray(grant.privileges) && grant.privileges.every((value: any) => data.privileges.includes(value)))) {
                throw new Error("账户权限数据不完整");
            }
            setPermissions(data);
            form.setFieldsValue({grants: editor.action === "create"
                ? (form.getFieldValue("grants") || []).map((grant: any) => ({...grant,
                    privileges: grant.privileges.filter((value: string) => data.privileges.includes(value))}))
                : data.grants.map((grant: any) => ({...grant, privileges: [...grant.privileges]}))});
        }).catch((reason: any) => {
            if (!live.current || !current) return;
            setEditor(undefined);
            message.error(reason.message || "加载账户权限失败");
        }).finally(() => { if (live.current && current) setPermissionsLoading(false); });
        return () => { current = false; };
    }, [record.id, editor]);

    function open(action: string, account?: any) {
        if (pending.current || account && (action === "password" ? !canChangePassword(account)
            : action === "access" ? !canChangeAccess(account) : account.protected)) return;
        form.resetFields(); setRepair(""); setPermissions(undefined); setConnection(undefined);
        setPermissionsLoading(action === "create" || action === "permissions");
        setConnectionLoading(action === "create" || action === "access");
        form.setFieldsValue({name: account?.name, host: account?.host,
            source: account?.host === "%" ? "remote" : account?.host && account.host !== "localhost" ? "address" : "local", source_host: account?.host !== "%" ? account?.host : "",
            grants: action === "create" ? [{database: !["mysql", "information_schema", "performance_schema", "sys"].includes(database.toLowerCase()) ? database : undefined,
                privileges: [...profiles[2].privileges]}] : []});
        setEditor({action, account, selectAccount: action === "password" && !account, database});
    }
    const close = () => { if (!pending.current) { setEditor(undefined); } };
    const reload = () => {
        setQuery((current) => current.keyword !== query.keyword || current.pageSize !== query.pageSize || current.page !== query.page || current.page === pagination.page
            ? current : {...current, page: pagination.page});
        setVersion((value) => value + 1);
    };
    const showWarning = (data: any) => {
        if (data?.warning) message.warning(data.warning);
    };
    const save = async (values: any, confirmed = false) => {
        if (pending.current || !editor || (editor.action === "create" || editor.action === "access") && !connection) return;
        if ((editor.action === "create" || editor.action === "permissions") && (!permissions || permissionsLoading)) return;
        let grants: any[] | undefined;
        if (editor.action === "create") {
            grants = values.grants.map((grant: any) => ({database: grant.database, privileges: [...grant.privileges]}));
        } else if (editor.action === "permissions") {
            const previous = new Map<string, string[]>(permissions.grants.map((grant: any) => [grant.database, grant.privileges]));
            const uncertain = new Set<string>(permissions.uncertain || []);
            for (const database of uncertain) if (!previous.has(database)) previous.set(database, []);
            const next = new Map<string, string[]>((values.grants || []).map((grant: any) => [grant.database, grant.privileges || []]));
            grants = [];
            for (const [database, privileges] of next) {
                const original = previous.get(database) || [];
                if (uncertain.has(database) || privileges.length !== original.length || privileges.some((value) => !original.includes(value))) grants.push({database, privileges});
            }
            for (const [database, privileges] of previous) {
                if ((privileges.length || uncertain.has(database)) && !next.has(database)) grants.push({database, privileges: []});
            }
            if (!grants.length) { close(); return; }
        }
        if (editor.action === "access") {
            if (!canChangeAccess(editor.account)) return;
            const newHost = sourceHost(values);
            if (newHost.toLowerCase() === (editor.account.host || "%").toLowerCase()) { close(); return; }
            if (editor.account.name === "root" && !confirmed) {
                confirmation.current?.destroy();
                confirmation.current = modal.confirm({
                    title: "修改 root 访问来源？", okText: "确认修改", cancelText: "取消",
                    content: <>
                        <p style={{overflowWrap: "anywhere"}}>{`root@${editor.account.host} → root@${newHost}`}</p>
                        <span>新来源须允许面板所在服务器连接，否则后续将无法连接此实例。</span>
                    </>,
                    onOk: () => save(values, true),
                });
                return;
            }
        }
        pending.current = true; setBusy(true);
        const target = {database: editor.database};
        const account = {name: editor.account?.name || values.name, host: editor.action === "create" ? sourceHost(values) : editor.account?.host || values.host};
        let created = false;
        try {
            if (editor.action === "create") {
                const createdData = await api.users(record, {action: "create", ...target, ...account, password: values.password});
                created = true;
                // 创建成功后只重试授权；不再次发送 create，避免重复账户。
                if (live.current) showWarning(createdData);
                const grantedData = await api.users(record, {action: "update_permissions", ...account, grants});
                if (live.current) showWarning(grantedData);
            } else {
                const data = await api.users(record, {action: editor.action === "permissions" ? "update_permissions" : editor.action,
                    ...target, ...account, grants, new_host: editor.action === "access" ? sourceHost(values) : undefined, password: values.password});
                if (editor.action === "password" && data?.current && !data.connection_updated) {
                    if (live.current) {
                        message.warning({key: `mysql-account-password-${record.id}`, duration: 0,
                            content: data.warning || "密码已修改，但实例连接密码未保存，请在配置中填写新密码。"});
                        setEditor(undefined);
                    }
                    return;
                }
                if (live.current) showWarning(data);
            }
            message.success(editor.action === "create" ? "账户已创建并完成授权" : editor.action === "permissions" ? "权限已保存" : "操作成功");
            if (live.current) { reload(); setEditor(undefined); }
        } catch (reason: any) {
            if (created) {
                message.warning(`账户“${account.name}”已创建，权限设置失败，请通过“配置权限”补充授权。`);
                if (live.current) {
                    setRepair(reason.message); reload();
                    setPermissions(undefined); setPermissionsLoading(true);
                    setEditor({action: "permissions", account, database: editor.database,
                        permissionDraft: {privileges: [...permissions.privileges], grants}});
                    form.setFieldsValue({name: account.name, host: account.host});
                    form.setFieldValue("password", undefined);
                }
            } else {
                message.error(reason.message);
                // 多库授权可能已部分生效；保留表单供重试，同时刷新列表中的实际权限。
                if (editor.action === "permissions" && live.current) {
                    setPermissions((current: any) => ({...current, uncertain: Array.from(new Set([
                        ...(current.uncertain || []), ...(grants?.map((grant) => grant.database) || []),
                    ]))}));
                    reload();
                }
            }
        } finally { pending.current = false; if (live.current) setBusy(false); }
    };
    const remove = (account: any) => {
        if (pending.current || account.protected) return;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: `删除账户“${account.name}”？`, content: "删除后该账户将无法继续连接，已授予的权限也会移除。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                if (pending.current) throw new Error("请等待当前操作完成");
                pending.current = true; setBusy(true);
                try {
                    const data = await api.users(record, {action: "delete", name: account.name, host: account.host, database});
                    if (live.current) { showWarning(data); reload(); }
                    message.success("账户已删除");
                } catch (reason: any) { message.error(reason.message); throw reason; }
                finally { pending.current = false; if (live.current) setBusy(false); }
            }});
        // 启用 AntD Promise 模式：失败保留弹窗重试，避免再次抛出未处理的拒绝。
        void Promise.resolve(confirmation.current);
    };

    return <>
        <AccountsLayout rows={rows} loading={loading} busy={busy}
            keyword={query.keyword} pagination={pagination}
            onSearch={(keyword) => setQuery((current) => current.keyword === keyword ? current : {...current, keyword, page: 1})}
            onPageChange={(page, pageSize) => setQuery((current) => ({...current, page: pagination.pageSize === pageSize ? page : 1, pageSize}))}
            protectedReason={protectedReason} ActionDropdown={ActionDropdown}
            sourceColumn={{title: "访问来源", dataIndex: "host", width: 170, render: (host: string) => <Tooltip title={host === "%" ? "任何IP可访问（%）" : host === "localhost" ? "本地访问（localhost）" : `指定IP访问：${host}`}><span className="mysql-account-source">{host === "%" ? "任何IP可访问" : host === "localhost" ? "本地访问" : "指定IP访问"}</span></Tooltip>}}
            onCreate={() => open("create")} onPassword={(account) => open("password", account)} onPermissions={(account) => open("permissions", account)}
            onAccess={(account) => open("access", account)} onRemove={remove} onRefresh={() => reload()}/>
        <AccountEditor editor={editor} form={form} busy={busy} loading={loading} rows={rows} repair={repair} onCancel={close} onFinish={save}
            initializing={(editor?.action === "create" || editor?.action === "permissions") && (!permissions || permissionsLoading)
                || (editor?.action === "create" || editor?.action === "access") && !connection}
            submitDisabled={(editor?.action === "create" || editor?.action === "permissions") && (!permissions || permissionsLoading)}
            accountSelect={<AccountSelect record={record} database={editor?.database || database}
                onChange={(value, account) => {
                    const [name, host] = JSON.parse(value);
                    form.setFieldsValue({name, host});
                    setEditor((current: any) => ({...current, currentAccount: account?.current}));
                }}/>}>
            {editor?.selectAccount && <Form.Item name="host" hidden><Input/></Form.Item>}
            {editor?.selectAccount && <Form.Item name="name" hidden rules={[{required: true, message: "请选择账户"}]}><Input/></Form.Item>}
            {(editor?.action === "create" || editor?.action === "password") && <div className={editor.action === "create" ? "account-grid" : undefined}>
                {editor.action === "create" && <Form.Item name="name" label="账户名称" rules={[{required: true, whitespace: true, message: "请输入账户名称"}]}>
                    <Input maxLength={32} autoComplete="off" placeholder="例如 app_user"/>
                </Form.Item>}
                <Form.Item name="password" label="登录密码" rules={[{required: true, message: "请输入密码"}, {min: 8, max: 256, message: "密码长度为 8–256 个字符"}]}>
                    <Input.Password autoComplete="new-password" maxLength={256}
                        placeholder={editor.action === "password" ? "请输入新密码（8–256 个字符）" : "请输入登录密码"}/>
                </Form.Item>
            </div>}
            {editor?.action === "password" && (editor.account?.current || editor.currentAccount) &&
                <div className="account-hint">保存后会同步更新此实例的连接密码。</div>}
            {(editor?.action === "create" || editor?.action === "access") && <AccessFields/>}
            {editor?.action === "create" && permissions && <Form.Item required>
                <DatabasePermissionFields record={record} privileges={permissions.privileges} profiles={profiles} disabled={busy} required/>
            </Form.Item>}
            {editor?.action === "permissions" && permissions && <>
                <DatabasePermissionFields record={record} privileges={permissions.privileges} profiles={profiles} disabled={busy}/>
                {(permissions.retained || editor.account?.global_privileges?.some((value: string) => value !== "USAGE")
                    || editor.account?.granted_roles?.length > 0 || editor.account?.grants?.some((grant: any) => grant.scope !== "database")
                    || editor.account?.restrictions?.length > 0) &&
                    <div className="account-hint">这里只调整数据库的直接权限，其他授权保持不变。</div>}
            </>}
            {(editor?.action === "create" || editor?.action === "access") && <ConnectionDetails data={connection} loading={connectionLoading} description={"访问来源是账户允许登录的客户端地址，不会修改数据库监听地址或服务器防火墙。"}/>}
        </AccountEditor>
    </>;
};
