import {useEffect, useRef, useState} from "react";
import {App, Form, Input} from "antd";
import {useTheme} from "sinking-antd";
import AccountsLayout, {AccountEditor, ConnectionDetails, PermissionFields, PermissionSummary} from "../shared/accounts-layout";
import {useDatabaseClient} from "../shared/context";
import DatabaseSelect from "../shared/database-select";

const profiles = [
    {value: "readonly", label: "只读访问", description: "读取当前数据库中的表和视图。", privileges: ["SELECT"]},
    {value: "readwrite", label: "读写数据", description: "读取、新增、修改和删除数据，不允许改动表结构。", privileges: ["SELECT", "INSERT", "UPDATE", "DELETE"]},
    {value: "develop", label: "应用开发", description: "读写数据、执行存储程序，可创建表和视图；不授予实例管理权限。", privileges: ["SELECT", "INSERT", "UPDATE", "DELETE", "EXECUTE", "CREATE TABLE", "CREATE VIEW"]},
];
const permissionOptions = [
    {value: "SELECT", label: "读取数据"},
    {value: "INSERT", label: "新增数据"},
    {value: "UPDATE", label: "修改数据"},
    {value: "DELETE", label: "删除数据"},
    {value: "EXECUTE", label: "执行存储程序"},
    {value: "CREATE TABLE", label: "创建表"},
    {value: "CREATE VIEW", label: "创建视图"},
];
const protectedReason = "系统管理员、当前登录或具有较高管理权限的用户禁止删除和调整权限，密码可按登录权限修改。";

export default ({record, initialAction}: any) => {
    const {operate, ActionDropdown} = useDatabaseClient();
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const [form] = Form.useForm();
    const [query, setQuery] = useState({keyword:"",page:1,page_size:20});
    const [pagination, setPagination] = useState<any>();
    const changeQuery = (change: any) => setQuery(value => { const next={...value,...change}; return JSON.stringify(next)===JSON.stringify(value)?value:next; });
    const [rows, setRows] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const [warning, setWarning] = useState("");
    const [version, setVersion] = useState(0);
    const [editor, setEditor] = useState<any>();
    const [repair, setRepair] = useState("");
    const database = record.database_name || record.name;
    const targetDatabase = Form.useWatch("target_database", form) || editor?.database || database;
    const [connection, setConnection] = useState<any>();
    const [connectionError, setConnectionError] = useState("");
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
        const key = `${record.id}:${database}`;
        if (listScope.current !== key) { listScope.current = key; setRows([]); }
        setLoading(true); setError("");
        void operate(record, "users", {action: "list", database, ...query}).then((data) => {
            if (live.current && request === listSequence.current) { setRows(data?.list || []); setPagination(data); }
        }).catch((reason) => { if (live.current && request === listSequence.current) setError(reason.message); })
            .finally(() => { if (live.current && request === listSequence.current) setLoading(false); });
    }, [record.id, database, version, query]);
    useEffect(() => {
        if (initialAction === "password") open("password");
    }, [initialAction]);
    useEffect(() => {
        if (editor?.action !== "create") return;
        let current = true;
        setConnection(undefined); setConnectionError(""); setConnectionLoading(true);
        void operate(record, "users", {action: "connection", database: targetDatabase}).then((data) => {
            if (live.current && current) setConnection(data);
        }).catch((reason) => { if (live.current && current) setConnectionError(reason.message); })
            .finally(() => { if (live.current && current) setConnectionLoading(false); });
        return () => { current = false; };
    }, [record.id, editor?.action, targetDatabase]);

    function open(action: string, account?: any) {
        if (pending.current || account && (action === "password" ? !(account.password_editable ?? !account.protected) : account.protected)) return;
        form.resetFields(); setRepair("");
        const granted = account ? permissionOptions.filter(option => String(account.roles || "").toUpperCase().includes(`GRANT ${option.value}`)).map(option => option.value) : profiles[1].privileges;
        form.setFieldsValue({name: account?.name, host: account?.host, privileges: granted,
            target_database: database});
        setEditor({action, account, selectAccount: action === "password" && !account, database});
    }
    const close = () => { if (!pending.current) { setEditor(undefined); } };
    const showWarning = (data: any) => {
        if (data?.warning) { setWarning(data.warning); message.warning(data.warning); }
    };
    const save = async (values: any) => {
        if (pending.current || !editor) return;
        pending.current = true; setBusy(true);
        const target = {database: values.target_database || editor.database};
        const account = {name: editor.action === "password" ? editor.account?.login_name || editor.account?.name || values.name : editor.account?.name || values.name};
        try {
            if (editor.action === "create") {
                // 登录、数据库用户及权限在同一后端事务中完成。
                const data = await operate(record, "users", {action: "create", ...target, ...account, password: values.password, privileges: values.privileges});
                if (live.current) showWarning(data);
            } else {
                const data = await operate(record, "users", {action: editor.action === "permissions" ? repair ? "grant" : "permissions" : editor.action,
                    ...target, ...account, password: values.password, privileges: values.privileges});
                if (editor.action === "password" && data?.current && !data.connection_updated) {
                    const warning = data.warning || "密码已修改，但实例连接信息未保存，请在配置中填写新密码";
                    message.warning(warning);
                    if (live.current) { setWarning(warning); setEditor(undefined); }
                    return;
                }
                if (live.current) showWarning(data);
            }
            message.success(editor.action === "create" ? "账户已创建并完成授权" : "操作成功");
            if (live.current) { setVersion((value) => value + 1); setEditor(undefined); }
        } catch (reason: any) { message.error(reason.message);
        } finally { pending.current = false; if (live.current) setBusy(false); }
    };
    const remove = (account: any) => {
        if (pending.current || account.protected) return;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: `删除账户“${account.name}”？`, content: "删除实例登录和当前数据库用户后，该账户将无法登录。若登录还关联其他数据库，服务器会拒绝删除。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                if (pending.current) throw new Error("请等待当前操作完成");
                pending.current = true; setBusy(true);
                try {
                    const data = await operate(record, "users", {action: "delete", name: account.name, database});
                    if (live.current) { showWarning(data); setVersion((value) => value + 1); }
                    message.success("账户已删除");
                } catch (reason: any) { message.error(reason.message); throw reason; }
                finally { pending.current = false; if (live.current) setBusy(false); }
            }});
        void Promise.resolve(confirmation.current);
    };

    return <>
        <AccountsLayout pagination={pagination} onQueryChange={changeQuery} rows={rows} loading={loading} busy={busy} error={error} warning={warning} scope={`${database} · 添加账户时创建实例登录，并在所选数据库建立对应用户。`}
            protectedReason={protectedReason} ActionDropdown={ActionDropdown}
            renderPermissions={(account) => {
                const roles = String(account.roles || "").toUpperCase();
                const writable = /GRANT (?:INSERT|UPDATE|DELETE)|DB_DATAWRITER/.test(roles);
                const developer = /GRANT (?:EXECUTE|CREATE TABLE|CREATE VIEW)/.test(roles);
                const readable = /GRANT SELECT|DB_DATAREADER/.test(roles);
                return <PermissionSummary label={account.protected ? "受保护用户" : developer ? "应用开发" : writable ? "读写数据" : readable ? "只读访问" : "自定义或未授权"} readOnly={readable && !writable && !developer} detail={String(account.roles || "")}/>;
            }}

            onCreate={() => open("create")} onPassword={(account: any) => open("password", account)} onPermissions={(account: any) => open("permissions", account)}
            onRemove={remove} onRefresh={() => setVersion((value) => value + 1)} onDismissWarning={() => setWarning("")}/>
        <AccountEditor editor={editor} form={form} busy={busy} loading={loading} rows={rows} repair={repair} onCancel={close} onFinish={save}
            width={editor?.action === "permissions" ? (theme?.isCompactTheme?.() ? 360 : 400) : undefined}
            onSelect={(value: string) => { const [name, host, login] = JSON.parse(value); form.setFieldsValue({name: login || name, host}); }}>
            {editor?.selectAccount && <Form.Item name="name" hidden rules={[{required: true, message: "请选择账户"}]}><Input/></Form.Item>}
            {(editor?.action === "create" || editor?.action === "password") && <div className={editor.action === "create" ? "account-grid" : undefined}>
                {editor.action === "create" && <Form.Item name="name" label="账户名称" rules={[{required: true, whitespace: true, message: "请输入账户名称"}]}>
                    <Input maxLength={128} autoComplete="off" placeholder="例如 app_user"/>
                </Form.Item>}
                <Form.Item name="password" label="登录密码" rules={[{required: true, message: "请输入密码"}, {min: 8, max: 128, message: "密码长度为 8–128 个字符"}]}>
                    <Input.Password autoComplete="new-password" maxLength={128}
                        placeholder={editor.action === "password" ? "请输入新密码（8–128 个字符）" : "请输入登录密码"}/>
                </Form.Item>
            </div>}

            {(editor?.action === "create" || editor?.action === "permissions") && <div className={editor.action === "create" ? "account-grid" : undefined}>
                <Form.Item name="target_database" label="权限数据库" rules={[{required: true, message: "请选择数据库"}]}>
                    <DatabaseSelect record={record} placeholder="请选择权限数据库" disabled={busy}/>
                </Form.Item>

                <PermissionFields profiles={profiles} options={permissionOptions} disabled={busy} allowEmpty={editor?.action !== "create"}/>
            </div>}
            {editor?.action === "create" && <ConnectionDetails data={connection} error={connectionError} loading={connectionLoading} description={"远程连接由 SQL Server 的 TCP/IP 协议、监听端口、登录认证模式和服务器防火墙决定。"}/>}
        </AccountEditor>
    </>;
};
