import {useCallback, useEffect, useRef, useState} from "react";
import {App, Form, Input} from "antd";
import {useTheme} from "sinking-antd";
import {AccountEditor, ConnectionDetails} from "../shared/accounts-layout";
import {useDatabaseClient} from "../shared/context";
import AccountList from "./account-list";
import DatabasePermissionFields from "./account-permissions";

const protectedReason = "系统角色、超级用户和当前连接角色受保护，禁止删除和修改权限。";

export default ({record, schema, initialAction}: any) => {
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
    const accountSchema = schema || "public";
    const [connection, setConnection] = useState<any>();
    const [connectionError, setConnectionError] = useState("");
    const [connectionLoading, setConnectionLoading] = useState(false);
    const [permissions, setPermissions] = useState<any>();
    const [permissionsLoading, setPermissionsLoading] = useState(false);
    const [schemaLoading, setSchemaLoading] = useState(false);
    const schemaRequests = useRef(new Set<number>());
    const live = useRef(true);
    const pending = useRef(false);
    const listSequence = useRef(0);
    const listScope = useRef("");
    const confirmation = useRef<any>(undefined);
    const onSchemaLoading = useCallback((key: number, loading: boolean) => {
        if (loading) schemaRequests.current.add(key); else schemaRequests.current.delete(key);
        if (live.current) setSchemaLoading(schemaRequests.current.size > 0);
    }, []);
    const onSchemaError = useCallback((error: string) => {
        if (live.current) { setEditor(undefined); message.error(error); }
    }, [message]);

    useEffect(() => {
        live.current = true;
        return () => { live.current = false; listSequence.current++; confirmation.current?.destroy(); };
    }, []);
    useEffect(() => {
        const request = ++listSequence.current;
        const key = `${record.id}:${database}:${accountSchema}`;
        if (listScope.current !== key) { listScope.current = key; setRows([]); }
        setLoading(true); setError("");
        void operate(record, "users", {action: "list", database, schema: accountSchema, ...query}).then((data) => {
            if (live.current && request === listSequence.current) { setRows(data?.list || []); setPagination(data); }
        }).catch((reason) => { if (live.current && request === listSequence.current) setError(reason.message); })
            .finally(() => { if (live.current && request === listSequence.current) setLoading(false); });
    }, [record.id, accountSchema, database, version, query]);
    useEffect(() => {
        if (initialAction === "password") open("password");
    }, [initialAction]);
    useEffect(() => {
        if (editor?.action !== "create") return;
        let current = true;
        setConnection(undefined); setConnectionError(""); setConnectionLoading(true);
        void operate(record, "users", {action: "connection", database}).then((data) => {
            if (live.current && current) setConnection(data);
        }).catch((reason) => { if (live.current && current) setConnectionError(reason.message); })
            .finally(() => { if (live.current && current) setConnectionLoading(false); });
        return () => { current = false; };
    }, [record.id, editor?.action, database]);
    useEffect(() => {
        if (editor?.action !== "create" && editor?.action !== "permissions") return;
        if (editor.permissionDraft) {
            setPermissions(editor.permissionDraft);
            setPermissionsLoading(false);
            return;
        }
        let current = true;
        setPermissions(undefined); setPermissionsLoading(true);
        void operate(record, "users", {action: "permissions_info", name: editor.account?.name}).then((data) => {
            if (!live.current || !current) return;
            if (!Array.isArray(data?.privileges) || !Array.isArray(data.database_privileges) || !Array.isArray(data.grants)
                || !data.grants.every((grant: any) => typeof grant.database === "string" && typeof grant.schema === "string"
                    && Array.isArray(grant.privileges) && grant.privileges.every((value: any) => (grant.schema ? data.privileges : data.database_privileges).includes(value)))) {
                throw new Error("角色权限数据不完整");
            }
            setPermissions(data);
            if (editor.action === "permissions") {
                const groups = new Map<string, any>();
                for (const grant of data.grants) {
                    if (!groups.has(grant.database)) groups.set(grant.database, {database: grant.database, privileges: [], schemas: []});
                    const group = groups.get(grant.database);
                    if (grant.schema) group.schemas.push({schema: grant.schema, privileges: [...grant.privileges]});
                    else group.privileges = [...grant.privileges];
                }
                form.setFieldValue("grants", [...groups.values()]);
            }
        }).catch((reason: any) => {
            if (live.current && current) { setEditor(undefined); message.error(reason.message || "加载角色权限失败"); }
        }).finally(() => { if (live.current && current) setPermissionsLoading(false); });
        return () => { current = false; };
    }, [record.id, editor]);

    function open(action: string, account?: any) {
        if (pending.current || account && (action === "password" ? !(account.password_editable ?? !account.protected) : account.protected)) return;
        form.resetFields(); setRepair(""); setPermissions(undefined);
        setPermissionsLoading(action === "create" || action === "permissions");
        schemaRequests.current.clear(); setSchemaLoading(false);
        form.setFieldsValue({name: account?.name, host: account?.host,
            grants: action === "create" ? [{database, privileges: ["CONNECT"], initialize: true}] : []});
        setEditor({action, account, selectAccount: action === "password" && !account, database, schema: accountSchema});
    }
    const close = () => { if (!pending.current) { setEditor(undefined); } };
    const showWarning = (data: any) => {
        if (data?.warning) { setWarning(data.warning); message.warning(data.warning); }
    };
    const save = async (values: any) => {
        if (pending.current || !editor) return;
        if ((editor.action === "create" || editor.action === "permissions") && (!permissions || permissionsLoading || schemaLoading)) return;
        let grants: any[] | undefined;
        if (editor.action === "create" || editor.action === "permissions") {
            const next = new Map<string, any>();
            for (const group of form.getFieldValue("grants") || []) {
                if (group.initialize || group.schemas === undefined) { message.info("请等待数据库架构加载完成"); return; }
                const scope = {database: group.database, schema: "", privileges: group.privileges || []};
                next.set(JSON.stringify([scope.database, scope.schema]), scope);
                for (const entry of group.schemas) {
                    const scope = {database: group.database, schema: entry.schema, privileges: entry.privileges || []};
                    next.set(JSON.stringify([scope.database, scope.schema]), scope);
                }
            }
            if (editor.action === "create") grants = [...next.values()].filter((grant) => grant.privileges.length);
            else {
                const previous = new Map<string, any>(permissions.grants.map((grant: any) => [JSON.stringify([grant.database, grant.schema]), grant]));
                const uncertain = new Set<string>(permissions.uncertain || []);
                for (const key of uncertain) if (!previous.has(key)) {
                    const [database, schema] = JSON.parse(key); previous.set(key, {database, schema, privileges: []});
                }
                grants = [];
                for (const [key, grant] of next) {
                    const original = previous.get(key)?.privileges || [];
                    if (uncertain.has(key) || original.length !== grant.privileges.length || grant.privileges.some((value: string) => !original.includes(value))) grants.push(grant);
                }
                for (const [key, grant] of previous) if (!next.has(key) && (grant.privileges.length || uncertain.has(key))) grants.push({...grant, privileges: []});
                if (!grants.length) { close(); return; }
            }
        }
        pending.current = true; setBusy(true);
        const target = {database: editor.database, schema: editor.schema};
        const account = {name: editor.account?.name || values.name};
        let created = false;
        try {
            if (editor.action === "create") {
                const createdData = await operate(record, "users", {action: "create", ...target, ...account, password: values.password});
                created = true;
                // 创建成功后只重试授权；不再次发送 create，避免重复账户。
                if (live.current) showWarning(createdData);
                const grantedData = await operate(record, "users", {action: "update_permissions", ...account, grants});
                if (live.current) showWarning(grantedData);
            } else {
                const data = await operate(record, "users", {action: editor.action === "permissions" ? "update_permissions" : editor.action,
                    ...target, ...account, password: values.password, grants});
                if (editor.action === "password" && data?.current && !data.connection_updated) {
                    if (live.current) {
                        message.warning({key: `postgres-account-password-${record.id}`, duration: 0,
                            content: data.warning || "密码已修改，但实例连接密码未保存，请在配置中填写新密码。"});
                        setEditor(undefined);
                    }
                    return;
                }
                if (live.current) showWarning(data);
            }
            message.success(editor.action === "create" ? "角色已创建并完成授权" : "操作成功");
            if (live.current) { setVersion((value) => value + 1); setEditor(undefined); }
        } catch (reason: any) {
            if (created) {
                message.warning(`角色“${account.name}”已创建，权限设置失败，请通过“配置权限”补充授权。`);
                if (live.current) {
                    setRepair(reason.message); setVersion((value) => value + 1);
                    setEditor({action: "permissions", account, ...target, permissionDraft: {...permissions, grants: [],
                        uncertain: (grants || []).map((grant) => JSON.stringify([grant.database, grant.schema]))}});
                    form.setFieldValue("name", account.name);
                    form.setFieldValue("password", undefined);
                }
            } else {
                message.error(reason.message);
                if (editor.action === "permissions" && live.current) {
                    setPermissions((current: any) => ({...current, uncertain: [...new Set([
                        ...(current.uncertain || []), ...(grants || []).map((grant) => JSON.stringify([grant.database, grant.schema])),
                    ])]}));
                    setVersion((value) => value + 1);
                }
            }
        } finally { pending.current = false; if (live.current) setBusy(false); }
    };
    const remove = (account: any) => {
        if (pending.current || account.protected) return;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: `删除角色“${account.name}”？`, content: "删除后使用该角色的应用将无法连接，已授予的权限也会移除。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                if (pending.current) throw new Error("请等待当前操作完成");
                pending.current = true; setBusy(true);
                try {
                    const data = await operate(record, "users", {action: "delete", name: account.name, database, schema: accountSchema});
                    if (live.current) { showWarning(data); setVersion((value) => value + 1); }
                    message.success("角色已删除");
                } catch (reason: any) { message.error(reason.message); throw reason; }
                finally { pending.current = false; if (live.current) setBusy(false); }
            }});
        void Promise.resolve(confirmation.current);
    };

    return <>
        <AccountList pagination={pagination} onQueryChange={changeQuery} rows={rows} loading={loading} busy={busy} error={error} warning={warning}
            protectedReason={protectedReason} ActionDropdown={ActionDropdown}
            database={database} schema={accountSchema}
            onCreate={() => open("create")} onPassword={(account: any) => open("password", account)} onPermissions={(account: any) => open("permissions", account)}
            onRemove={remove} onRefresh={() => setVersion((value) => value + 1)} onDismissWarning={() => setWarning("")}/>
        <AccountEditor accountLabel="角色" editor={editor} form={form} busy={busy} loading={loading} rows={rows} repair={repair} onCancel={close} onFinish={save}
            width={editor?.action === "permissions" ? (theme?.isCompactTheme?.() ? 440 : 480) : undefined}
            initializing={(editor?.action === "create" || editor?.action === "permissions") && (!permissions || permissionsLoading)}
            submitDisabled={(editor?.action === "create" || editor?.action === "permissions") && (!permissions || permissionsLoading || schemaLoading)}
            onSelect={(value: string) => { const [name, host] = JSON.parse(value); form.setFieldsValue({name, host}); }}>
            {editor?.selectAccount && <Form.Item name="name" hidden rules={[{required: true, message: "请选择角色"}]}><Input/></Form.Item>}
            {(editor?.action === "create" || editor?.action === "password") && <div className={editor.action === "create" ? "account-grid" : undefined}>
                {editor.action === "create" && <Form.Item name="name" label="角色名称" rules={[{required: true, whitespace: true, message: "请输入角色名称"}]}>
                    <Input maxLength={63} autoComplete="off" placeholder="例如 app_user"/>
                </Form.Item>}
                <Form.Item name="password" label="登录密码" rules={[{required: true, message: "请输入密码"}, {min: 8, max: 256, message: "密码长度为 8–256 个字符"}]}>
                    <Input.Password autoComplete="new-password" maxLength={256}
                        placeholder={editor.action === "password" ? "请输入新密码（8–256 个字符）" : "请输入登录密码"}/>
                </Form.Item>
            </div>}

            {(editor?.action === "create" || editor?.action === "permissions") && permissions && <>
                <DatabasePermissionFields record={record} permissions={permissions} disabled={busy} onLoading={onSchemaLoading} onError={onSchemaError}/>
                {permissions.retained && <div className="account-hint" style={{marginTop: 12, marginBottom: 0}}>
                    单独对象、继承、默认权限及其他角色授予的权限保持不变。
                </div>}
            </>}
            {editor?.action === "create" && <ConnectionDetails data={connection} error={connectionError} loading={connectionLoading} description={"是否允许远程连接由 listen_addresses、pg_hba.conf 和服务器防火墙决定，这里只管理登录角色和数据权限。"}/>}
        </AccountEditor>
    </>;
};
