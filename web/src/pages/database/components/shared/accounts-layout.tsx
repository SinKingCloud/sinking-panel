import {useEffect, useMemo, useRef, useState} from "react";
import type {ReactNode} from "react";
import {Alert, App, Button, Collapse, ConfigProvider, Empty, Form, Grid, Input, Pagination, Select, Spin, Tooltip, Typography, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import useStyles from "./styles/account";
import {VirtualTable} from "./client-common";
import PermissionSelect from "./permission-select";

export interface PermissionProfile {
    value: string;
    label: string;
    description: string;
    privileges: string[];
}

export interface AccountLayoutProps {
    rows: any[];
    pagination?: {page: number; page_size: number; total: number};
    onQueryChange?: (query: {keyword?: string; page?: number; page_size?: number}) => void;
    extra?: ReactNode;
    loading: boolean;
    busy: boolean;
    error?: string;
    warning?: string;
    scope: ReactNode;
    protectedReason: string;
    databaseSelect?: ReactNode;
    createDisabled?: boolean;
    renderPermissions: (account: any) => ReactNode;
    sourceColumn?: any;
    ActionDropdown: any;
    onCreate: () => void;
    onPassword: (account?: any) => void;
    onPermissions: (account: any) => void;
    onRemove: (account: any) => void;
    onAccess?: (account: any) => void;
    onRefresh: () => void;
    onDismissWarning: () => void;
}

// 只提供统一列表布局；请求、账户模型和权限方案由各数据库组件处理。
export default (props: AccountLayoutProps) => {
    const appearance = useTheme();
    const {styles} = useStyles({compact: Boolean(appearance?.isCompactTheme?.())});
    const {message} = App.useApp();
    const screens = Grid.useBreakpoint();
    useEffect(() => { if (props.error) message.error(props.error); }, [props.error]);
    const [keyword, setKeyword] = useState("");
    const rows = useMemo(() => props.onQueryChange ? props.rows : props.rows.filter((row) => `${row.name} ${row.host || ""}`.toLowerCase().includes(keyword.trim().toLowerCase())), [props.rows, keyword, props.onQueryChange]);
    const Actions = props.ActionDropdown;
    return <div className={styles.panel}>
        <div className="account-toolbar">
            <Input className="account-search" variant="filled" allowClear prefix={<Icon type="SearchOutlined"/>}
                placeholder="搜索账户" value={keyword} onChange={(event) => { setKeyword(event.target.value); if (!event.target.value) props.onQueryChange?.({keyword:"",page:1}); }} onPressEnter={() => props.onQueryChange?.({keyword:keyword.trim(),page:1})} onBlur={() => props.onQueryChange?.({keyword:keyword.trim(),page:1})}/>
            {props.databaseSelect}
            {props.extra}<div className="account-tools">
                <Button size="small" color="default" variant="filled" icon={<Icon type="PlusOutlined"/>} disabled={props.busy || props.createDisabled} onClick={props.onCreate}>添加账户</Button>
                <Button size="small" color="default" variant="filled" loading={props.loading} disabled={props.busy} onClick={props.onRefresh} aria-label="刷新账户" icon={<Icon type="ReloadOutlined"/>}/>
            </div>
        </div>
        {props.warning && <Alert className="account-notice" type="warning" showIcon title={props.warning} closable onClose={props.onDismissWarning}/>}
        <VirtualTable className="sql-fill-table" size="small" loading={props.loading} rowKey={(row: any) => JSON.stringify([row.name, row.host || "", row.login_name || ""])}
            scroll={{x: props.sourceColumn ? 680 : 540}} dataSource={rows}
            pagination={props.onQueryChange ? false : {pageSize: 20, simple: !screens.md, showSizeChanger: false, hideOnSinglePage: true}}
            locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}
            columns={[
                {title: "账户名称", dataIndex: "name", width: 170, render: (value: any, account: any) => <div className="account-name">
                    <span>{value}</span>{account.protected && <Tooltip title={props.protectedReason}><Icon type="LockOutlined" className="account-protected"/></Tooltip>}
                </div>},
                ...(props.sourceColumn ? [props.sourceColumn] : []),
                {title: "数据权限", key: "permissions", render: (_: any, account: any) => props.renderPermissions(account)},
                {title: "操作", key: "operation", width: 84, align:"center", fixed:"right", render: (_: any, account: any) => {
                    const items = [
                        ...(account.password_editable ?? !account.protected ? [{key: "password", label: "密码", onClick: () => props.onPassword(account)}] : []),
                        ...(!account.protected ? [
                            {key: "permissions", label: "权限", onClick: () => props.onPermissions(account)},
                            ...(props.onAccess ? [{key: "access", label: "来源", onClick: () => props.onAccess?.(account)}] : []),
                            {type: "divider" as const}, {key: "delete", label: "删除", danger: true, onClick: () => props.onRemove(account)},
                        ] : []),
                    ];
                    return items.length ? <Actions disabled={props.busy} menu={{items}}><Button color="default" variant="filled" size="small" autoInsertSpace={false} disabled={props.busy}>操作</Button></Actions> : null;
                }},
            ]}/>
        {props.onQueryChange && <div className="client-footer" style={{visibility: props.pagination ? undefined : "hidden"}}>
            <Typography.Text type="secondary">{screens.md ? `${rows.length} 条 / 共 ${props.pagination?.total ?? 0} 条` : `${props.pagination?.total ?? 0} 条`}</Typography.Text>
            <ConfigProvider input={{variant:"filled"}}><Pagination className="client-pagination" size="small" current={props.pagination?.page || 1} pageSize={props.pagination?.page_size || 20} total={props.pagination?.total || 0}
                simple={!screens.md} showQuickJumper={Boolean(screens.md)} showLessItems disabled={props.loading || props.busy}
                showSizeChanger={{variant:"filled",showSearch:false}} pageSizeOptions={[10,20,50,100,200,500,1000]}
                onChange={(page,page_size) => props.onQueryChange?.({page:page_size===props.pagination?.page_size?page:1,page_size})}/></ConfigProvider>
        </div>}

    </div>;
};

export function AccountEditor({editor, form, busy, loading, initializing, rows, onSelect, onCancel, onFinish, children, repair, submitDisabled, accountSelect, width, accountLabel = "账户"}: any) {
    const {styles} = useStyles();
    const {token} = antdTheme.useToken();
    const theme = useTheme();
    const lastContent = useRef<any>(undefined);
    const [visible, setVisible] = useState(false);
    useEffect(() => {
        if (editor) lastContent.current = {editor, children, repair, accountSelect, width, initializing};
    });
    useEffect(() => { if (editor) setVisible(true); }, [editor]);
    // 退出动画仍展示最后一次打开的表单，实际开关继续使用 editor。
    const content = editor ? {editor, children, repair, accountSelect, width, initializing} : visible ? lastContent.current : undefined;
    const current = content?.editor;
    return <ProModal title={<Title>{current?.action === "create" ? `添加${accountLabel}` : current?.action === "password" ? "修改密码" : current?.action === "access" ? "访问来源" : `${accountLabel}权限`}</Title>}
        onCancel={onCancel} onOk={form.submit} modalProps={{
            open: Boolean(editor), className: styles.editor,
            width: content?.width ?? (current?.action === "password" ? (theme?.isCompactTheme?.() ? 340 : 380)
                : current?.action === "access" ? (theme?.isCompactTheme?.() ? 360 : 400) : (theme?.isCompactTheme?.() ? 440 : 480)),
            style: {maxWidth: "calc(100vw - 24px)"},
            confirmLoading: busy, okText: "保存", cancelText: "取消",
            okButtonProps: {disabled: content?.initializing || submitDisabled || current?.selectAccount && !content?.accountSelect && (loading || !rows.some((row: any) => row.password_editable ?? !row.protected))},
            // 仅保留空 Form 连接实例；字段在关闭动画结束后释放。
            mask: {closable: !busy}, keyboard: !busy, closable: !busy, forceRender: true,
            afterClose: () => { lastContent.current = undefined; setVisible(false); },
        }}>
        <ConfigProvider theme={{components: {
            Input: {controlHeight: token.controlHeight},
            Select: {controlHeight: token.controlHeight},
        }}}>
        <Form form={form} layout="vertical" variant="filled" disabled={busy} onFinish={onFinish}>
            {content?.initializing ? <div className="account-loading"><Spin/></div> : <>
                {content?.repair && <Alert className="account-repair" type="warning" showIcon title={`${accountLabel}已创建，请确认权限后保存`}
                    description={content.repair}/>}
                {current?.selectAccount && <Form.Item name="account" label={`选择${accountLabel}`} rules={[{required: true, message: `请选择${accountLabel}`}]}>
                    {content?.accountSelect || <Select loading={loading} placeholder={`选择要修改密码的${accountLabel}`} options={rows.filter((account: any) => account.password_editable ?? !account.protected).map((account: any) => ({value: JSON.stringify([account.name, account.host || "", account.login_name || ""]),
                        label: `${account.name}${account.host ? `@${account.host}` : ""}${account.login_name && account.login_name !== account.name ? `（登录：${account.login_name}）` : ""}`}))} onChange={onSelect}/>}
                </Form.Item>}
                {content?.children}
            </>}
        </Form>
        </ConfigProvider>
    </ProModal>;
}

export function PermissionFields({profiles, options, disabled, allowEmpty = true}: {
    profiles: PermissionProfile[];
    options: {value: string; label: string}[];
    disabled?: boolean;
    allowEmpty?: boolean;
}) {
    return <Form.Item name="privileges" label="权限方案" getValueProps={(value) => ({value: value || []})}
        rules={allowEmpty ? undefined : [{required: true, type: "array", min: 1, message: "请至少选择一项权限"}]}>
        <PermissionSelect profiles={profiles} options={options} disabled={disabled} allowEmpty={allowEmpty}/>
    </Form.Item>;
}

export function ConnectionDetails({data, loading, error, description}: any) {
    return <Collapse className="account-network" ghost items={[{key: "connection", label: "连接与访问说明", children: <>
        {loading ? <Spin size="small"/> : Array.isArray(data?.items) && data?.items?.length ? <dl className="account-network-list">
            {data?.items?.map((item: any, index: number) => <div key={`${item?.label}:${index}`} style={{display: "contents"}}><dt>{item?.label}</dt><dd>{item?.value || "未设置"}</dd></div>)}
        </dl> : null}
        <div className="account-network-note">{data?.note || description}{error && <><br/>连接信息暂不可读：{error}</>}</div>
    </>}]} />;
}

export function PermissionSummary({label, detail, readOnly = false}: {label: string; detail?: string; readOnly?: boolean}) {
    return <Tooltip title={detail || undefined}><span className="account-permission">
        {readOnly && <Icon type="ReadOutlined" className="account-readonly"/>}<span>{label}</span>
    </span></Tooltip>;
}
