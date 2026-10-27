import {useEffect, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, Empty, Form, Input, Select, Table, Tooltip, theme as antdTheme} from "antd";
import {DownOutlined, LockOutlined, PlusOutlined, ReloadOutlined, SearchOutlined} from "@ant-design/icons";
import {ProModal, Title, useTheme} from "sinking-antd";
import {useRedisClient} from "./context";
import useStyles from "./account-styles";

const permissionOptions = [{value: "read", label: "读取数据"}, {value: "write", label: "修改数据"}];
const protectedReason = "默认账户、当前连接或具有实例管理权限的账户受保护";

// Redis ACL 属于实例级权限；逻辑库之间不能使用 SQL 式的独立授权。
export default ({record}: any) => {
    const {api, ActionDropdown} = useRedisClient();
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const {token} = antdTheme.useToken();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const [form] = Form.useForm();
    const [rows, setRows] = useState<any[]>([]);
    const [keyword, setKeyword] = useState("");
    const [loading, setLoading] = useState(false);
    const [busy, setBusy] = useState(false);
    const [version, setVersion] = useState(0);
    const [editor, setEditor] = useState<any>();
    const [displayEditor, setDisplayEditor] = useState<any>();
    const live = useRef(true);
    const pending = useRef(false);
    const listSequence = useRef(0);
    const confirmation = useRef<any>(undefined);
    const tableContainer = useRef<HTMLDivElement>(null);
    const [tableHeight, setTableHeight] = useState(320);
    useEffect(() => {
        live.current = true;
        const node = tableContainer.current;
        const observer = new ResizeObserver(([entry]) => {
            if (entry.contentRect.height > 0) setTableHeight(Math.max(80, Math.floor(entry.contentRect.height - 40)));
        });
        if (node) observer.observe(node);
        return () => { live.current = false; listSequence.current++; confirmation.current?.destroy(); observer.disconnect(); };
    }, []);
    useEffect(() => {
        const request = ++listSequence.current;
        setLoading(true);
        void api.users(record, {action: "list"}).then((data) => {
            if (live.current && request === listSequence.current) setRows(Array.isArray(data?.list) ? data?.list.filter(Boolean) : []);
        }).catch((reason) => { if (live.current && request === listSequence.current) { setRows([]); message.error(reason?.message || "操作失败"); } })
            .finally(() => { if (live.current && request === listSequence.current) setLoading(false); });
    }, [record.id, version]);

    const permissions = (account: any) => {
        let read = false, write = false;
        for (const rule of String(account?.roles || "").split(/\s+/)) {
            if (rule === "+@all" || rule === "-@all") read = write = rule[0] === "+";
            if (rule === "+@read" || rule === "-@read") read = rule[0] === "+";
            if (rule === "+@write" || rule === "-@write") write = rule[0] === "+";
        }
        return {label: account?.protected ? "系统权限" : read && write ? "读写权限" : read ? "只读权限" : "自定义", values: [read && "read", write && "write"].filter(Boolean)};
    };
    const open = (action: string, account?: any) => {
        if (pending.current || account?.protected) return;
        const privileges = account ? permissions(account).values : ["read", "write"];
        form.resetFields();
        form.setFieldsValue({name: account?.name, privileges});
        const next = {action, account};
        setEditor(next); setDisplayEditor(next);
    };
    const save = async (values: any) => {
        if (pending.current || !editor) return;
        pending.current = true; setBusy(true);
        try {
            const data = await api.users(record, {action: editor.action, name: editor.account?.name || values.name,
                ...(editor.action !== "permissions" ? {password: values.password} : {}),
                ...(editor.action !== "password" ? {privileges: values.privileges} : {})});
            if (!live.current) return;
            if (data?.warning) message.warning(data?.warning); else message.success("保存成功");
            setVersion((value) => value + 1); setEditor(undefined);
        } catch (reason: any) { if (live.current) message.error(reason?.message || "操作失败"); }
        finally { pending.current = false; if (live.current) setBusy(false); }
    };
    const remove = (account: any) => {
        if (pending.current || account.protected) return;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: `删除账户“${account.name}”？`, content: "删除后，使用这个账户的应用将无法连接 Redis。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true}, onOk: async () => {
                if (pending.current || !live.current) return;
                pending.current = true; setBusy(true);
                try {
                    const data = await api.users(record, {action: "delete", name: account.name});
                    if (!live.current) return;
                    if (data?.warning) message.warning(data?.warning); else message.success("删除成功");
                    setVersion((value) => value + 1);
                } catch (reason: any) { if (live.current) message.error(reason?.message || "操作失败"); }
                finally { pending.current = false; if (live.current) setBusy(false); }
            }});
    };
    const visible = useMemo(() => rows.filter((account) => String(account.name).toLowerCase().includes(keyword.trim().toLowerCase())), [rows, keyword]);
    return <div className={styles.panel}>
        <div className="redis-account-toolbar redis-toolbar">
            <Input className="redis-account-search" variant="filled" allowClear value={keyword} onChange={(event) => setKeyword(event.target.value)}
                placeholder="搜索账户" aria-label="搜索账户" prefix={<SearchOutlined/>}/>
            <Button color="default" variant="filled" icon={<PlusOutlined/>} disabled={busy} onClick={() => open("create")}>新增</Button>
            <Button className="redis-account-refresh" color="default" variant="filled" icon={<ReloadOutlined/>}
                loading={loading} disabled={busy} aria-label="刷新账户" onClick={() => setVersion((value) => value + 1)}/>
        </div>
        <div className="redis-account-table" ref={tableContainer}>
            <Table<any> aria-label="Redis 账户管理" rowKey="name" size="small" loading={loading} dataSource={visible} virtual
                rowClassName={(_, index) => index === visible.length - 1 ? "client-last-row" : ""}
                scroll={{x: 540, y: tableHeight, scrollToFirstRowOnChange: false}} pagination={false} tableLayout="fixed"
                locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}} columns={[
                    {title: "账户名称", dataIndex: "name", width: 240, ellipsis: true, render: (name, account) => <div className="redis-account-name"><span title={name}>{name}</span>
                        {account.protected && <Tooltip title={protectedReason}><LockOutlined className="redis-account-lock"/></Tooltip>}</div>},
                    {title: "数据权限", key: "permissions", width: 210, ellipsis: true, render: (_, account) => <Tooltip title={account.roles || undefined}>{permissions(account).label}</Tooltip>},
                    {title: "操作", key: "operation", width: 90, fixed: "right", align: "center", render: (_, account) => account.protected
                        ? <Tooltip title={protectedReason}><span className="redis-account-protected">受保护</span></Tooltip>
                        : busy ? null : <ActionDropdown menu={{items: [{key: "permissions", label: "权限", onClick: () => open("permissions", account)},
                            {key: "password", label: "密码", onClick: () => open("password", account)},
                            {type: "divider"}, {key: "delete", label: "删除", danger: true, onClick: () => remove(account)}]}}>
                            <Button color="default" variant="filled" size="small">操作<DownOutlined/></Button>
                        </ActionDropdown>},
                ]}/>
        </div>
        <div className="redis-account-scope"><span>权限作用于整个实例</span><span>共 {visible.length} 条</span></div>
        <ProModal title={<Title>{displayEditor?.action === "create" ? "添加账户" : displayEditor?.action === "password" ? "修改密码" : "配置权限"}</Title>}
            onCancel={() => { if (!pending.current) setEditor(undefined); }} onOk={form.submit} okText="保存"
            modalProps={{open: Boolean(editor), rootClassName: styles.editor,
                width: displayEditor?.action === "password" ? (compact ? 340 : 380) : displayEditor?.action === "permissions" ? (compact ? 360 : 400) : (compact ? 440 : 480),
                style: {maxWidth: "calc(100vw - 24px)"}, styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}},
                confirmLoading: busy, cancelText: "取消", cancelButtonProps: {disabled: busy},
                mask: {closable: !busy}, keyboard: !busy, closable: !busy, forceRender: true,
                afterClose: () => { setDisplayEditor(undefined); form.resetFields(); }}}>
            <ConfigProvider theme={{components: {Input: {controlHeight: token.controlHeight}, Select: {controlHeight: token.controlHeight}}}}>
            <Form form={form} layout="vertical" variant="filled" disabled={busy} onFinish={save}>
                {displayEditor?.action === "create" && <Form.Item name="name" label="账户名称" rules={[{required: true, whitespace: true, message: "请输入账户名称"}, {pattern: /^[^\s\x00-\x1f\x7f]+$/, message: "名称不能包含空白或控制字符"}]}>
                    <Input maxLength={128} autoComplete="off" placeholder="例如 app_user"/>
                </Form.Item>}
                {(displayEditor?.action === "create" || displayEditor?.action === "password") && <Form.Item name="password" label={displayEditor?.action === "password" ? "新密码" : "登录密码"}
                    rules={[{required: true, message: "请输入密码"}, {min: 8, max: 256, message: "密码长度为 8–256 个字符"}]}>
                    <Input.Password autoComplete="new-password" maxLength={256} placeholder="请输入至少 8 个字符的密码"/>
                </Form.Item>}
                {(displayEditor?.action === "create" || displayEditor?.action === "permissions") &&
                    <Form.Item name="privileges" label="数据权限" rules={[{required: true, type: "array", min: 1, message: "请至少选择一项权限"}]}>
                        <Select mode="multiple" showSearch={false} options={permissionOptions} placeholder="请选择数据权限"/>
                    </Form.Item>}
            </Form>
            </ConfigProvider>
        </ProModal>
    </div>;
};
