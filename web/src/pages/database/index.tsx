import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Empty, Tag} from "antd";
import {Body, Icon, PageTable} from "sinking-antd";
import type {PageTableRef} from "sinking-antd";
import ActionDropdown from "@/pages/components/action-dropdown";
import TypeManager from "@/pages/components/type-manager";
import type {TypeManagerRef} from "@/pages/components/type-manager";
import {deleteDatabase, getDatabaseList} from "@/service/api/database";
import {getTypeList} from "@/service/api/type";
import {getEnum} from "@/service/api/system";
import {getParams} from "@/utils/page";
import InstanceCreate from "./instance-create";
import InstanceForm from "./instance-form";
import Workspace from "./workspace";
import useStyles from "./components/shared/styles/client";

export default () => {
    const {message, modal} = App.useApp();
    const {styles} = useStyles();
    const tableRef = useRef<PageTableRef<any> | null>(null);
    const categoryRef = useRef<TypeManagerRef | null>(null);
    const confirmRef = useRef<{destroy: () => void} | null>(null);
    const live = useRef(true);
    const catalogRequest = useRef(0);
    const deletingRef = useRef(false);
    const [types, setTypes] = useState<Record<string, string>>({});
    const [categories, setCategories] = useState<any[]>([]);
    const [ready, setReady] = useState(false);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [listError, setListError] = useState("");
    const [keyword, setKeyword] = useState("");
    const [typeFilter, setTypeFilter] = useState("");
    const [categoryFilter, setCategoryFilter] = useState("");
    const [createOpen, setCreateOpen] = useState(false);
    const [formOpen, setFormOpen] = useState(false);
    const [record, setRecord] = useState<any>();
    const [workspace, setWorkspace] = useState<any>();
    const [workspaceRefresh, setWorkspaceRefresh] = useState(0);
    const [deleting, setDeleting] = useState(false);
    const categoryNames = useMemo(() => Object.fromEntries(categories.map((item) => [item.id, item.name])), [categories]);
    const reload = () => tableRef.current?.reload();
    const loadCatalog = useCallback(async () => {
        const request = ++catalogRequest.current;
        setLoading(true); setError("");
        try {
            const [enums, groups] = await Promise.all([getEnum({body: {name: "database"}}).catch(() => undefined), getTypeList({body: {module: "database", order_by_field: "sort", order_by_type: "asc"}})]);
            if (!live.current || request !== catalogRequest.current) return;
            if (groups?.code !== 200) throw new Error(groups?.message || "获取分类失败");
            setTypes(enums?.code === 200 ? enums?.data?.kind || {} : {}); setCategories(Array.isArray(groups?.data?.list) ? groups?.data?.list.filter((item: any) => item?.id != null) : []); setReady(true);
        } catch (reason: any) { if (live.current && request === catalogRequest.current) setError(reason?.message || "数据加载失败"); }
        finally { if (live.current && request === catalogRequest.current) setLoading(false); }
    }, []);
    useEffect(() => { live.current = true; void loadCatalog(); return () => { live.current = false; catalogRequest.current++; confirmRef.current?.destroy(); }; }, [loadCatalog]);
    const configure = (item: any) => { setRecord(item); setFormOpen(true); };
    const manage = (item: any) => { if (!item.configured) configure(item); else setWorkspace(item); };
    const remove = (items: any[]) => {
        if (!items.length || deletingRef.current) return;
        confirmRef.current?.destroy();
        confirmRef.current = modal.confirm({title: items.length === 1 ? `删除实例“${items[0].name}”？` : `删除 ${items.length} 个实例？`,
            content: "仅删除面板中的连接信息，远端数据库及数据不受影响。", okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                if (deletingRef.current) return;
                deletingRef.current = true; setDeleting(true);
                try {
                    const response = await deleteDatabase({body: {ids: items.map((item) => item.id)}});
                    if (!live.current) return;
                    if (response?.code !== 200) throw new Error(response?.message || "删除实例失败");
                    message.success("删除成功"); tableRef.current?.clearSelectedRows();
                    setWorkspace((value: any) => items.some((item) => String(item.id) === String(value?.id)) ? undefined : value); reload();
                } catch (reason: any) { if (live.current) message.error(reason?.message || "删除实例失败"); }
                finally { deletingRef.current = false; if (live.current) setDeleting(false); }
            }});
    };
    return <Body loading={loading && !ready}>
        <PageTable<any> ref={tableRef} ariaLabel="数据管理" rootClassName={styles.table} rowKey="id"
            defaultPageSize={20} paginationAffix={{offsetBottom: 15}} scroll={{x: 880}} requestEnabled={ready}
            params={{type_id: categoryFilter}}
            request={async (params, sort) => {
                const response = await getDatabaseList({body: getParams({...params}, sort)});
                if (response?.code !== 200 || !Array.isArray(response?.data?.list)) throw new Error(response?.message || "获取实例列表失败");
                return {data: response?.data?.list, total: response?.data?.total ?? 0, success: true};
            }} onLoad={() => setListError("")} onRequestError={(reason) => { setListError(reason?.message); message.error(reason?.message); }}
            emptyContent={error || listError ? <div className="database-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="数据加载失败"/>
                <Button color="default" variant="filled" onClick={() => { void loadCatalog(); reload(); }}>重新加载</Button></div> : undefined}
            hero={{eyebrow: "DATABASE CLIENT", title: "数据管理", action: {label: "添加实例", onClick: () => setCreateOpen(true)}}}
            columns={[
                {title: "实例名称", dataIndex: "name", key: "name", width: 240, sorter: true, render: (value: string, item: any) => <div className="database-name-cell">
                    {item.type >= 1 && item.type <= 7 ? <span className={styles.databaseIcon} data-type={item.type} aria-hidden="true"/> : <Icon type="DatabaseOutlined"/>}
                    <button type="button" className="database-name" title={value} onClick={() => manage(item)}>{value}</button></div>},
                {title: "类型", dataIndex: "type", key: "type", width: 130, render: (value: number) => types[value] || value},
                {title: "分类", dataIndex: "type_id", key: "type_id", width: 140, render: (value: number) => Number(value) > 0 ? categoryNames[value] || "未知分类" : "全部分类"},
                {title: "连接配置", dataIndex: "configured", key: "configured", width: 120, render: (value: boolean) => <Tag variant="filled" color={value ? "success" : "default"}>{value ? "已配置" : "待配置"}</Tag>},
                {title: "添加时间", dataIndex: "create_time", key: "create_time", width: 170, sorter: true, render: (value: string) => <span className="database-time">{value}</span>},
                {title: "操作", key: "action", width: 82, fixed: "right", className: "action-cell", render: (_, item) => <ActionDropdown disabled={deleting} menu={{items: [
                        {key: "manage", label: "管理", onClick: () => manage(item)},
                        {key: "configure", label: "修改", onClick: () => configure(item)},
                        {key: "delete", label: "删除", danger: true, onClick: () => remove([item])},
                    ]}}><Button size="small" disabled={deleting} aria-label={`操作 ${item.name}`}>操作</Button></ActionDropdown>},
            ]}
            search={{value: keyword, onChange: setKeyword, placeholder: "搜索实例名称", maxLength: 100}}
            filters={[
                {type: "select", name: "type", label: "数据库类型", icon: "DatabaseOutlined", allLabel: "全部类型", valueEnum: types, value: typeFilter, disabled: deleting,
                    onChange: (value) => { setTypeFilter(value); tableRef.current?.clearSelectedRows(); }},
            ]}
            toolbar={{refresh: {disabled: deleting || loading, onClick: () => { reload(); void loadCatalog(); }}, actions: [
                {key: "category", type: "filter", value: categoryFilter, icon: "FolderOutlined", ariaLabel: "数据库分类筛选", matchWidth: true, disabled: deleting,
                    options: [{value: "", label: "全部分类"}, ...categories.map((item) => ({value: String(item.id), label: item.name}))],
                    command: {label: "分类管理", icon: "SettingOutlined", onClick: () => categoryRef.current?.open()},
                    onChange: (value) => { setCategoryFilter(value); tableRef.current?.clearSelectedRows(); }},
            ]}}
            rowSelection={{columnWidth: 44, getCheckboxProps: () => ({disabled: deleting}), clearDisabled: deleting,
                actions: (_, rows) => [{key: "delete", type: "button", label: "批量删除", danger: true, disabled: deleting, onClick: () => remove(rows)}]}}/>
        <InstanceCreate open={createOpen} types={types} categories={categories} defaultType={Number(typeFilter) || undefined} defaultTypeId={categoryFilter}
            onCancel={() => setCreateOpen(false)} onSuccess={reload}
            onCreated={(item: any) => { setCreateOpen(false); configure(item); }}/>
        <InstanceForm open={formOpen} record={record} categories={categories}
            onCancel={() => setFormOpen(false)} onSuccess={(change: any) => {
                reload();
                if (change.action === "update" && String(workspace?.id) === String(change.id)) setWorkspaceRefresh((value) => value + 1);
                if (change.action === "update") setWorkspace((value: any) => {
                    if (String(value?.id) !== String(change.id)) return value;
                    return {...value, name: change.name, type_id: change.type_id};
                });
            }}/>
        <TypeManager ref={categoryRef} module="database" onChange={(items) => {
            catalogRequest.current++; setLoading(false); setCategories(items);
            if (categoryFilter && !items.some((item) => String(item.id) === categoryFilter)) setCategoryFilter("");
        }} onMutation={reload}/>
        <Workspace record={workspace} open={!formOpen} refreshKey={workspaceRefresh} onCancel={() => setWorkspace(undefined)} onConfigure={() => configure(workspace)}/>
    </Body>;
};
