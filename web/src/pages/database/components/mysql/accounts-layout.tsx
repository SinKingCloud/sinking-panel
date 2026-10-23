import {useEffect, useMemo, useRef, useState} from "react";
import {Button, ConfigProvider, Empty, Grid, Input, Pagination, Tooltip, Typography} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import useStyles from "./account-styles";
import {canChangeAccess, canChangePassword, VirtualTable} from "./client-common";

const accountKey = (account: any) => JSON.stringify([account.name, account.host || ""]);
const privilegeNames: Record<string, string> = {
    SELECT: "读取数据", INSERT: "新增数据", UPDATE: "修改数据", DELETE: "删除数据", CREATE: "创建对象", ALTER: "修改结构",
    INDEX: "管理索引", DROP: "删除对象", EXECUTE: "执行程序", REFERENCES: "引用外键", TRIGGER: "管理触发器",
    "CREATE VIEW": "创建视图", "SHOW VIEW": "查看视图", "CREATE ROUTINE": "创建程序", "ALTER ROUTINE": "修改程序",
    "CREATE TEMPORARY TABLES": "创建临时表", "LOCK TABLES": "锁定表", "GRANT OPTION": "转授权限", USAGE: "仅可连接",
};
const dataPrivileges = new Set(["SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "INDEX", "DROP", "EXECUTE", "REFERENCES", "TRIGGER", "CREATE VIEW", "SHOW VIEW", "CREATE ROUTINE", "ALTER ROUTINE", "CREATE TEMPORARY TABLES", "LOCK TABLES", "ALL", "ALL PRIVILEGES"]);
export const permissionText = (privileges: string[] = []) => privileges.map((value) => privilegeNames[value] || value).join("、");
const permissionLabel = (privileges: string[] = []) => {
    if (privileges.some((value) => ["ALL", "ALL PRIVILEGES"].includes(value))) return "全部权限";
    const reads = privileges.includes("SELECT");
    const writes = privileges.some((value) => ["INSERT", "UPDATE", "DELETE"].includes(value));
    const structure = privileges.some((value) => ["CREATE", "ALTER", "INDEX", "DROP", "CREATE VIEW", "CREATE ROUTINE", "ALTER ROUTINE", "TRIGGER"].includes(value));
    if (structure) return reads || writes ? "数据与结构" : "结构操作";
    if (writes) return reads ? "读写数据" : "写入权限";
    if (reads && privileges.includes("EXECUTE")) return "查询与执行";
    if (reads) return privileges.every((value) => ["SELECT", "SHOW VIEW", "GRANT OPTION"].includes(value)) ? "只读访问" : "读取与其他权限";
    return permissionText(privileges) || "对象权限";
};

// 列表按库归纳，明细保留授权层级，避免把单表权限误认为整库权限。
function permissionGroups(account: any) {
    const groups: any[] = [];
    const global = account.global_privileges || [];
    if (global.some((value: string) => dataPrivileges.has(value))) {
        groups.push({name: "所有数据库", label: `${permissionLabel(global)}${account.restrictions?.length ? " · 含撤销限制" : ""}`, global: true});
    } else if (global.some((value: string) => value !== "USAGE")) {
        groups.push({name: "实例权限", label: "查看具体权限", global: true});
    }
    const databases = new Map<string, any>();
    for (const grant of account.grants || []) {
        const key = JSON.stringify([grant.database, grant.pattern]);
        let group = databases.get(key);
        if (!group) {
            group = {name: grant.database, pattern: grant.pattern, privileges: [], objects: 0};
            databases.set(key, group);
        }
        if (grant.scope === "database") group.privileges.push(...grant.privileges);
        else group.objects++;
    }
    for (const group of databases.values()) {
        groups.push({...group, label: group.privileges.length
            ? `${permissionLabel(group.privileges)}${group.objects ? " · 含对象授权" : ""}` : `${group.objects} 项对象授权`});
    }
    if (account.granted_roles?.length) groups.push({name: "角色授权", label: `${account.granted_roles.length} 个角色`, role: true});
    return groups;
}

type MySQLAccountsLayoutProps = {
    rows: any[];
    loading: boolean;
    busy: boolean;
    protectedReason: string;
    createDisabled?: boolean;
    sourceColumn?: any;
    ActionDropdown: any;
    onCreate: () => void;
    onPassword: (account?: any) => void;
    onPermissions: (account: any) => void;
    onRemove: (account: any) => void;
    onAccess?: (account: any) => void;
    onRefresh: () => void;
    keyword: string;
    pagination: {page: number; pageSize: number; total: number};
    onSearch: (value: string) => void;
    onPageChange: (page: number, pageSize: number) => void;
};

export default function MySQLAccountsLayout(props: MySQLAccountsLayoutProps) {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const [keyword, setKeyword] = useState(props.keyword);
    const [openRowKey, setOpenRowKey] = useState("");
    const [previewKey, setPreviewKey] = useState("");
    const [detail, setDetail] = useState<any>();
    const lastDetail = useRef<any>(undefined);
    useEffect(() => { if (detail) lastDetail.current = detail; }, [detail]);
    const currentDetail = detail || lastDetail.current;
    const permissionCache = useMemo(() => new WeakMap<object, ReturnType<typeof permissionGroups>>(), [props.rows]);
    const widths = {name: 180, source: typeof props.sourceColumn?.width === "number" ? props.sourceColumn.width : 170, permissions: 340, actions: 82};
    const tableWidth = widths.name + widths.permissions + (props.sourceColumn ? widths.source : 0) + widths.actions;
    const actionsDisabled = props.busy || props.loading;
    const Actions = props.ActionDropdown;
    const search = () => {
        const value = keyword.trim();
        setKeyword(value);
        props.onSearch(value);
    };

    useEffect(() => setKeyword(props.keyword), [props.keyword]);
    useEffect(() => { setOpenRowKey(""); setPreviewKey(""); }, [props.keyword, props.busy, props.loading, props.pagination.page, props.pagination.pageSize, detail]);

    return <>
        <div className={`${styles.panel} sql-workbench`}>
            <div className="mysql-account-toolbar">
                <Input className="mysql-account-search" variant="filled" size="small" allowClear
                    aria-label="搜索账户名称或来源" placeholder="搜索账户或来源"
                    prefix={<Icon type="SearchOutlined"/>} value={keyword}
                    onChange={(event) => {
                        const value = event.target.value; setKeyword(value); if (!value) props.onSearch("");
                    }} onPressEnter={search} onBlur={search}/>
                <div className="mysql-account-tools">
                    <Button color="default" variant="filled" size="small" icon={<Icon type="PlusOutlined"/>}
                        disabled={actionsDisabled || props.createDisabled} onClick={props.onCreate}>新增</Button>
                    <Button color="default" variant="filled" size="small" aria-label="刷新账户列表" icon={<Icon type="ReloadOutlined"/>}
                        loading={props.loading} disabled={props.busy} onClick={props.onRefresh}/>
                </div>
            </div>
            <VirtualTable className="mysql-account-table mysql-fill-table" size="small" loading={props.loading}
                rowKey={accountKey} dataSource={props.rows} scroll={{x: tableWidth, scrollToFirstRowOnChange: false}}
                rowClassName={(account) => openRowKey === accountKey(account) ? "action-menu-open" : ""}
                pagination={false}
                locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description={props.keyword ? "没有匹配的账户" : "暂无数据"}/>}}
                columns={[
                    {title: "账户名称", dataIndex: "name", width: widths.name, render: (name: string, account: any) =>
                        <div className="mysql-account-name">
                            <Tooltip title={`${name || "匿名账户"}${account.host ? `@${account.host}` : ""}`}><span>{name || "匿名账户"}</span></Tooltip>
                            {account.protected && <Tooltip title={canChangePassword(account) || canChangeAccess(account) ? props.protectedReason : "此账户不允许修改。"}>
                                <Icon type="LockOutlined" className="mysql-account-protected"/>
                            </Tooltip>}
                        </div>},
                    ...(props.sourceColumn ? [{...props.sourceColumn, width: widths.source}] : []),
                    {title: "授权数据库", key: "permissions", width: widths.permissions, render: (_: any, account: any) => {
                        let groups = permissionCache.get(account);
                        if (!groups) { groups = permissionGroups(account); permissionCache.set(account, groups); }
                        if (!groups.length) return <span className="mysql-account-no-grant">未查询到数据库授权</span>;
                        const key = accountKey(account);
                        return <Tooltip trigger="hover" destroyOnHidden
                            open={!detail && previewKey === key}
                            onOpenChange={(open) => setPreviewKey((current) => open ? key : current === key ? "" : current)}
                            title={<div className={styles.grantsTooltip}>
                                {groups.map((group, index) => <div key={index}>
                                    {group.name}{group.pattern && "（数据库名模式）"}：{group.label}{group.role && "，权限未展开"}
                                </div>)}
                            </div>}>
                            <button className="mysql-account-grants" type="button" onClick={() => { setPreviewKey(""); setDetail(account); }}>
                                {groups.map((group) => `${group.name}${group.pattern ? "（模式）" : ""}`).join("、")}
                            </button>
                        </Tooltip>;
                    }},
                    {title: "操作", key: "operation", width: widths.actions, align: "center" as const, fixed: "right" as const, render: (_: any, account: any) => actionsDisabled ? null :
                        <Actions open={openRowKey === accountKey(account)}
                            onOpenChange={(open: boolean) => setOpenRowKey(open ? accountKey(account) : "")}
                            menu={{items: [
                                {key: "view", label: "详情", onClick: () => setDetail(account)},
                                ...(!account.protected ? [
                                    {key: "permissions", label: "权限", onClick: () => props.onPermissions(account)},
                                ] : []),
                                ...(canChangePassword(account) ? [
                                    {key: "password", label: "密码", onClick: () => props.onPassword(account)},
                                ] : []),
                                ...(props.onAccess && canChangeAccess(account) ? [
                                    {key: "access", label: "来源", onClick: () => props.onAccess?.(account)},
                                ] : []),
                                ...(!account.protected ? [
                                    {type: "divider"},
                                    {key: "delete", label: "删除", danger: true, onClick: () => props.onRemove(account)},
                                ] : []),
                            ]}}>
                            <Button color="default" variant="filled" size="small" autoInsertSpace={false}>操作</Button>
                        </Actions>},
                ]}/>
            <div className="client-footer">
                <Tooltip title={screens.md ? undefined : `${props.rows.length} 个账户 / 共 ${props.pagination.total} 个`} trigger={["hover", "click"]}>
                    <Typography.Text type="secondary">{screens.md ? `${props.rows.length} 个账户 / 共 ${props.pagination.total} 个` : `${props.pagination.total} 个账户`}</Typography.Text>
                </Tooltip>
                <ConfigProvider input={{variant: "filled"}}>
                    <Pagination className="client-pagination" size="small" current={props.pagination.page} pageSize={props.pagination.pageSize} total={props.pagination.total}
                        disabled={actionsDisabled} showLessItems simple={!screens.md} showQuickJumper={Boolean(screens.md)} responsive
                        locale={screens.md ? undefined : {items_per_page: "/页"}}
                        showSizeChanger={{showSearch: false, variant: "filled", getPopupContainer: (trigger) => trigger.ownerDocument.body}}
                        pageSizeOptions={[10, 20, 50, 100, 200, 500, 1000]} onChange={props.onPageChange}/>
                </ConfigProvider>
            </div>
        </div>
        <ProModal title={<Title>账户权限</Title>} onOk={() => setDetail(undefined)} onCancel={() => setDetail(undefined)} okText="确认" okType="primary" modalProps={{
            open: Boolean(detail), width: compact ? 440 : 480, className: styles.permissions,
            style: {maxWidth: "calc(100vw - 24px)"}, destroyOnHidden: true,
            cancelButtonProps: {style: {display: "none"}}, mask: {closable: true},
        }}>
            {(currentDetail?.global_privileges || []).filter((value: string) => value !== "USAGE").length > 0 && <div className="mysql-permission-section">
                <div className="mysql-permission-heading"><span>全局授权</span><small>{currentDetail?.restrictions?.length ? "部分数据库存在撤销限制" : "数据类权限适用于所有数据库"}</small></div>
                <p>{permissionText(currentDetail?.global_privileges)}</p>
            </div>}
            {(currentDetail?.grants || []).map((grant: any, index: number) => <div className="mysql-permission-section" key={`grant-${index}`}>
                <div className="mysql-permission-heading"><span>{grant.database}{grant.pattern && "（数据库名模式）"}</span>
                    <small>{grant.scope === "database" ? "整个数据库" : grant.scope === "table" ? "指定数据表" : grant.scope === "column" ? "指定字段" : "存储程序"}</small></div>
                {grant.object && <div className="mysql-permission-object">{grant.object}</div>}
                <p>{permissionText(grant.privileges)}</p>
            </div>)}
            {currentDetail?.granted_roles?.length > 0 && <div className="mysql-permission-section">
                <div className="mysql-permission-heading"><span>角色授权</span><small>角色权限未展开</small></div>
                <p>{currentDetail.granted_roles.join("、")}</p>
            </div>}
            {currentDetail?.restrictions?.map((restriction: any, index: number) => <div className="mysql-permission-section" key={`restriction-${index}`}>
                <div className="mysql-permission-heading"><span>{restriction.database}</span><small>已撤销的全局权限</small></div>
                <p>{permissionText(restriction.privileges)}</p>
            </div>)}
            {currentDetail && !permissionGroups(currentDetail).length && !currentDetail.restrictions?.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="未查询到数据库授权"/>}
        </ProModal>
    </>;
}
