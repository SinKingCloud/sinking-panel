import {useEffect, useRef, useState} from "react";
import {App, Button, ConfigProvider, Empty, Form, Grid, Input, Pagination, Select, Spin, Tooltip, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {VirtualTable} from "../shared/client-common";
import {useDatabaseClient} from "../shared/context";
import DatabaseSelect from "../shared/database-select";
import useStyles from "./styles";

const permissions = [{value: "read", label: "只读访问"}, {value: "readWrite", label: "读写数据"}, {value: "dbAdmin", label: "数据库管理"}, {value: "userAdmin", label: "账户管理"}];

export default ({record}: any) => {
    const {operate, ActionDropdown} = useDatabaseClient();
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const {token} = antdTheme.useToken();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const screens = Grid.useBreakpoint();
    const [form] = Form.useForm();
    const [result, setResult] = useState<any>();
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [refresh, setRefresh] = useState(0);
    const [keyword, setKeyword] = useState("");
    const [search, setSearch] = useState("");
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(20);
    const database = record.database_name || record.name;
    const [editor, setEditor] = useState<any>();
    const lastEditor = useRef<any>(undefined);
    if (editor) lastEditor.current = editor;
    const displayEditor = editor || lastEditor.current;
    const live = useRef(true);
    const pending = useRef(false);
    const sequence = useRef(0);
    const editorSequence = useRef(0);
    const confirmation = useRef<any>(undefined);
    useEffect(() => { live.current = true; return () => { live.current = false; sequence.current++; editorSequence.current++; confirmation.current?.destroy(); }; }, []);
    useEffect(() => {
        const request = ++sequence.current;
        setLoading(true);
        void operate(record, "users", {action: "list", database, keyword, page, page_size: pageSize}).then((data) => {
            if (live.current && request === sequence.current) setResult({...data, list: Array.isArray(data?.list) ? data?.list.filter(Boolean) : []});
        }).catch((reason) => { if (live.current && request === sequence.current) { setResult(undefined); message.error(reason?.message || "读取账户失败"); } })
            .finally(() => { if (live.current && request === sequence.current) setLoading(false); });
    }, [record.id, record.name, database, keyword, page, pageSize, refresh, operate]);
    const close = () => {
        if (pending.current) return;
        editorSequence.current++;
        setEditor(undefined);
    };
    const open = async (action: string, account?: any) => {
        if (pending.current || account && (action === "password" ? !account.password_editable : account.protected)) return;
        const request = ++editorSequence.current;
        const target = account?.database || database;
        setEditor({action, account, database: target, request, loading: action === "permissions",
            values: {name: account?.name, target_database: database, grants: [{database, privileges: ["readWrite", "dbAdmin"]}]}});
        if (action !== "permissions") return;
        try {
            const data = await operate(record, "users", {action: "permissions", database: target, name: account.name});
            if (!live.current || request !== editorSequence.current) return;
            if (!Array.isArray(data?.grants) || !data?.grants?.every((grant: any) => typeof grant?.database === "string"
                && grant?.database && Array.isArray(grant?.privileges)
                && grant?.privileges?.every((value: any) => permissions.some((item) => item.value === value)))) {
                throw new Error("读取账户权限失败");
            }
            setEditor((current: any) => ({...current, loading: false, values: {grants: data?.grants}}));
        } catch (reason: any) {
            if (!live.current || request !== editorSequence.current) return;
            setEditor(undefined);
            message.error(reason?.message || "读取账户权限失败");
        }
    };
    const save = async (values: any) => {
        if (!editor || editor.loading || pending.current) return;
        pending.current = true; setSaving(true);
        try {
            const data = await operate(record, "users", {action: editor.action === "permissions" ? "update_permissions" : editor.action,
                database: editor.action === "create" ? values.target_database : editor.database,
                name: editor.account?.name || values.name, password: values.password,
                grants: ["create", "permissions"].includes(editor.action) ? values.grants || [] : undefined});
            if (!live.current) return;
            if (data?.warning) message.warning(data?.warning); else message.success(editor.action === "create" && values.target_database !== database ? `已在 ${values.target_database} 创建账户` : "保存成功");
            setEditor(undefined); setRefresh((value) => value + 1);
            if (editor.action === "create") setPage(1);
        } catch (reason: any) { if (live.current) message.error(reason?.message || "保存失败"); }
        finally { pending.current = false; if (live.current) setSaving(false); }
    };
    const applySearch = () => { const value = search.trim(); if (value !== keyword) { setKeyword(value); setPage(1); } };
    return <div className="manager-stage">
        <div className="manager-toolbar mongo-main-toolbar"><div className="manager-tools">
            <Input className="manager-search" variant="filled" prefix={<Icon type="SearchOutlined"/>} placeholder="搜索账户" allowClear value={search} onChange={(event) => { setSearch(event.target.value); if (!event.target.value) { setKeyword(""); setPage(1); } }} onPressEnter={applySearch} onBlur={applySearch}/>
            <div className="mongo-toolbar-actions">
                <Button color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} onClick={() => void open("create")}>新增</Button>
            </div>
        </div><Button className="manager-refresh" color="default" variant="filled" icon={<Icon type="ReloadOutlined"/>} aria-label="刷新账户" loading={loading} onClick={() => setRefresh((value) => value + 1)}/></div>
        <div className="manager-table"><VirtualTable rowKey={(account: any) => JSON.stringify([account.database, account.name])} size="small" scroll={{x: 676}} loading={loading} pagination={false} dataSource={result?.database === database ? result?.list || [] : []} locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
            columns={[
                {title: "账户名称", dataIndex: "name", width: 180, ellipsis: true},
                {title: "认证数据库", dataIndex: "database", width: 160, ellipsis: true},
                {title: "数据库权限", dataIndex: "roles", width: 260, ellipsis: true, render: (value: string) => <Tooltip title={value || undefined}><span>{value || "未授权"}</span></Tooltip>},
                {title: "操作", width: 76, fixed: "right", align: "center", key: "action", render: (_: any, account: any) => account.protected && !account.password_editable ? <span>—</span> : <ActionDropdown menu={{items: [
                    ...(!account.protected ? [{key: "permissions", label: "权限", onClick: () => void open("permissions", account)}] : []),
                    ...(account.password_editable ? [{key: "password", label: "密码", onClick: () => void open("password", account)}] : []),
                    ...(!account.protected ? [{key: "delete", label: "删除", danger: true, onClick: () => { confirmation.current = modal.confirm({title: `删除账户“${account.name}”？`, content: "该账户将无法再连接数据库。", okText: "删除", cancelText: "取消", okButtonProps: {danger: true}, onOk: async () => {
                        if (pending.current) return;
                        pending.current = true;
                        try { await operate(record, "users", {action: "delete", database: account.database, name: account.name}); if (live.current) { message.success("删除成功"); setRefresh((value) => value + 1); } }
                        catch (reason: any) { if (live.current) message.error(reason?.message || "删除失败"); }
                        finally { pending.current = false; }
                    }}); }}] : []),
                ]}}><Button color="default" variant="filled" size="small">操作</Button></ActionDropdown>},
            ]}/></div>
        <div className="manager-footer"><span className="manager-statistics">共 {result?.total || 0} 条</span><Pagination size="small" current={result?.page || page} pageSize={pageSize} total={result?.total || 0} simple={!screens.md} showSizeChanger={Boolean(screens.md)} showQuickJumper={Boolean(screens.md)} pageSizeOptions={[20, 50, 100, 500, 1000]} onChange={(value, size) => { setPage(size === pageSize ? value : 1); setPageSize(size); }}/></div>
        <ProModal title={<Title>{({create: "添加账户", password: "修改密码", permissions: "配置权限"} as any)[displayEditor?.action]}</Title>} onCancel={close} onOk={form.submit} modalProps={{open: Boolean(editor), width: displayEditor?.action === "create" ? (compact ? 380 : 420) : displayEditor?.action === "permissions" ? (compact ? 350 : 380) : (compact ? 340 : 380), rootClassName: styles.modal, zIndex: 1140,
            styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}}, style: {maxWidth: "calc(100vw - 24px)"},
            destroyOnHidden: true, okText: "保存", cancelText: "取消", confirmLoading: saving, okButtonProps: {disabled: Boolean(editor?.loading)}, cancelButtonProps: {disabled: saving}, closable: !saving, keyboard: !saving, mask: {closable: !saving}}}>
            <ConfigProvider wave={{disabled: true}} theme={{components: {Input: {controlHeight: token.controlHeight}, Select: {controlHeight: token.controlHeight}}}}>
            {displayEditor?.loading ? <div className="mongo-account-loading"><Spin/></div> : <Form form={form} key={displayEditor?.request} initialValues={displayEditor?.values} clearOnDestroy layout="vertical" variant="filled" onFinish={save} disabled={saving}>
                {["create", "password"].includes(displayEditor?.action) && <>
                {displayEditor?.action === "create" && <Form.Item name="name" label="账户名称" rules={[{required: true, whitespace: true, message: "请输入账户名称"}]}><Input autoComplete="off" maxLength={128} placeholder="请输入账户名称"/></Form.Item>}
                <Form.Item name="password" label={displayEditor?.action === "password" ? "新密码" : "登录密码"} rules={[{required: true, message: "请输入密码"}, {max: 1024, message: "密码不能超过 1024 个字符"}]}><Input.Password autoComplete="new-password" maxLength={1024} placeholder={displayEditor?.action === "password" ? "请输入新密码" : "请输入登录密码"}/></Form.Item>
                </>}
                {displayEditor?.action === "create" && <Form.Item name="target_database" label="认证数据库" rules={[{required: true, message: "请选择认证数据库"}]}><DatabaseSelect record={record} placeholder="请选择数据库"/></Form.Item>}
                {["create", "permissions"].includes(displayEditor?.action) && <Form.List name="grants" rules={displayEditor?.action === "create"
                    ? [{validator: (_, grants) => grants?.length ? Promise.resolve() : Promise.reject(new Error("请添加授权数据库"))}] : []}>
                    {(fields, {add, remove}, {errors}) => <div className="mongo-account-grants">
                        {fields.map((field) => <div className="mongo-account-grant" key={field.key}>
                            <div className="mongo-account-grant-heading">
                                <Form.Item name={[field.name, "database"]} label={`授权数据库 ${field.name + 1}`} rules={[
                                    {required: true, message: "请选择数据库"},
                                    {validator: (_, value) => (form.getFieldValue("grants") || []).some((grant: any, index: number) => index !== field.name && value && grant.database === value)
                                        ? Promise.reject(new Error("该数据库已添加")) : Promise.resolve()},
                                ]}><DatabaseSelect record={record} placeholder="请选择数据库" disabled={saving}/></Form.Item>
                                <Button color="default" variant="filled" icon={<Icon type="CloseOutlined"/>} aria-label="移除授权数据库" disabled={saving} onClick={() => remove(field.name)}/>
                            </div>
                            <Form.Item name={[field.name, "privileges"]} label="数据库权限" rules={[{type: "array", min: 1, required: true, message: "请选择数据库权限"}]}>
                                <Select mode="multiple" variant="filled" showSearch={false} options={permissions} placeholder="请选择数据库权限" disabled={saving}/>
                            </Form.Item>
                        </div>)}
                        <Button block type="dashed" icon={<Icon type="PlusOutlined"/>} disabled={saving || fields.length >= 1000}
                            onClick={() => add({privileges: ["readWrite", "dbAdmin"]})}>添加数据库</Button>
                        <Form.ErrorList errors={errors}/>
                    </div>}
                </Form.List>}
            </Form>}
            </ConfigProvider>
        </ProModal>
    </div>;
};
