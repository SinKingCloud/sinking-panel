import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import type {ComponentType} from "react";
import {App, Button, Col, Form, Input, Row, Select, Spin} from "antd";
import {createStyles} from "antd-style";
import {ModalTable, ProModal, Title, useTheme} from "sinking-antd";
import type {ModalTableRef} from "sinking-antd";
import ActionDropdown from "@/pages/components/ui/action-dropdown";
import TypeManager from "@/pages/components/type-manager";
import type {TypeManagerRef} from "@/pages/components/type-manager";
import {createScript, deleteScript, getScriptInfo, getScriptList, updateScript} from "@/service/api/script";
import {getTypeList} from "@/service/api/type";
import {getParams} from "@/utils/page";

interface ScriptToolsProps {
    module?: string;
    instanceId?: number | string;
    value: string | (() => string);
    onUse: (script: string) => void;
    disabled?: boolean;
    mode?: string;
    Editor?: ComponentType<any>;
}

const useStyles = createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    table: css`
        &.modal-table-loading { min-height: ${compact ? 200 : 240}px; }
        && .ui-table-data > [role="status"] { min-height: ${compact ? 140 : 160}px; }
        .script-time { color: ${token.colorTextSecondary}; font-variant-numeric: tabular-nums; white-space: nowrap; }
    `,
    editor: css`border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; overflow: hidden;`,
}));

// 按模块及可选实例复用脚本库，终端、计划任务和数据库管理器共用脚本与分类接口。
export default function ScriptTools({module = "script_database", instanceId, value, onUse, disabled, mode = "sql", Editor}: ScriptToolsProps) {
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const [open, setOpen] = useState(false);
    const [typeId, setTypeId] = useState<string>();
    const [types, setTypes] = useState<{id: number; name: string; module: string}[]>([]);
    const [typesLoading, setTypesLoading] = useState(false);
    const [editing, setEditing] = useState<any>();
    const [formOpen, setFormOpen] = useState(false);
    const [formActive, setFormActive] = useState(false);
    const [loading, setLoading] = useState(false);
    const [busy, setBusy] = useState(false);
    const [form] = Form.useForm();
    const table = useRef<ModalTableRef<any> | null>(null);
    const categories = useRef<TypeManagerRef | null>(null);
    const sequence = useRef(0);
    const typeRequest = useRef(0);
    const scope = instanceId === undefined || instanceId === null ? module : `${module}_${instanceId}`;
    const scopes = useMemo(() => [{value: module, label: "全局脚本"}, ...(scope === module ? [] : [{value: scope, label: "当前实例"}])], [module, scope]);
    const formModule = Form.useWatch("module", form) || scope;
    const formTypeId = Form.useWatch("type_id", form);
    const formTypes = useMemo(() => formModule === module ? types.filter((item) => item.module === module) : types, [formModule, module, types]);
    const typeNames = useMemo(() => Object.fromEntries(types.map(({id, name}) => [id, name])), [types]);
    const params = useMemo(() => ({module: scope, type_id: typeId}), [scope, typeId]);
    const handleTypesChange = useCallback((items: typeof types) => {
        typeRequest.current++;
        const nextItems = (Array.isArray(items) ? items : []).filter((item) => item?.id);
        setTypes(nextItems);
        setTypesLoading(false);
        setTypeId((current) => current && !nextItems.some((item) => String(item?.id) === current) ? undefined : current);
    }, []);

    useEffect(() => () => { sequence.current++; }, []);
    useEffect(() => { sequence.current++; setOpen(false); setFormOpen(false); setTypeId(undefined); setTypes([]); }, [scope]);
    useEffect(() => {
        if (!open) return;
        let active = true;
        const requestId = ++typeRequest.current;
        setTypesLoading(true);
        void getTypeList({body: {module: scope, order_by_field: "sort", order_by_type: "asc"}}).then((response) => {
            if (!active || typeRequest.current !== requestId) return;
            if (response?.code !== 200) throw new Error(response?.message || "获取脚本分类失败");
            handleTypesChange(Array.isArray(response?.data?.list) ? response?.data?.list : []);
        }).catch((error) => {
            if (active) message.error(error?.message || "获取脚本分类失败");
        }).finally(() => { if (active) setTypesLoading(false); });
        return () => { active = false; };
    }, [scope, open, message, handleTypesChange]);

    useEffect(() => {
        if (!formOpen || editing?.id || typesLoading || !formTypeId) return;
        if (!formTypes.some((item) => item.id === formTypeId)) form.setFieldValue("type_id", 0);
    }, [editing?.id, form, formOpen, formTypeId, formTypes, typesLoading]);

    const request = useCallback(async (query: any, sort: any) => {
        const response = await getScriptList({body: getParams({...query}, sort)});
        if (response?.code !== 200) throw new Error(response?.message || "获取脚本失败");
        return {data: Array.isArray(response?.data?.list) ? response?.data?.list.filter((item: any) => item?.id) : [], total: response?.data?.total || 0, success: true};
    }, []);

    const edit = async (item?: any) => {
        let script = "";
        if (!item) {
            try {
                script = typeof value === "function" ? value() : value;
            } catch (error) {
                message.error((error as Error)?.message || "读取脚本内容失败");
                return;
            }
        }
        const id = ++sequence.current;
        setBusy(false); setLoading(Boolean(item)); setFormActive(true); setFormOpen(true);
        setEditing(item);
        form.resetFields();
        if (!item) {
            form.setFieldsValue({name: "", module: scope, type_id: typeId ? Number(typeId) : 0, script});
            return;
        }
        try {
            const response = await getScriptInfo({body: {id: item.id}});
            if (id !== sequence.current) return;
            if (response?.code !== 200 || !response?.data) throw new Error(response?.message || "获取脚本失败");
            const detail = response?.data;
            setEditing(detail);
            form.setFieldsValue({name: detail?.name, script: detail?.script});
        } catch (error) {
            if (id === sequence.current) { setFormOpen(false); message.error((error as Error)?.message || "操作失败"); }
        } finally { if (id === sequence.current) setLoading(false); }
    };

    const use = async (item: any) => {
        const id = ++sequence.current;
        try {
            const response = await getScriptInfo({body: {id: item.id}});
            if (id !== sequence.current) return;
            if (response?.code !== 200 || !response?.data) throw new Error(response?.message || "获取脚本失败");
            const script = String(response?.data?.script || "");
            onUse(script);
            setOpen(false);
        } catch (error) { if (id === sequence.current) message.error((error as Error)?.message || "操作失败"); }
    };

    const remove = (item: any) => {
        const confirmation = modal.confirm({title: "删除脚本", content: `确定删除“${item.name}”吗？`, zIndex: 1250, okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                const response = await deleteScript({body: {ids: [item.id]}});
                if (response?.code !== 200) { message.error(response?.message || "删除失败"); throw new Error("删除失败"); }
                table.current?.reload();
            }});
        void Promise.resolve(confirmation);
    };

    const submit = async (values: any) => {
        if (busy) return;
        setBusy(true);
        const id = sequence.current;
        try {
            const body = {name: values.name.trim(), script: values.script};
            const response = editing?.id ? await updateScript({body: {...body, ids: [editing.id]}}) : await createScript({body: {...body, module: values.module, type_id: values.type_id}});
            if (id !== sequence.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "保存脚本失败");
            setFormOpen(false); table.current?.reload(); message.success("保存成功");
        } catch (error) { if (id === sequence.current) message.error((error as Error)?.message || "操作失败"); }
        finally { if (id === sequence.current) setBusy(false); }
    };

    return <>
        <Button size="small" color="default" variant="filled" onClick={() => setOpen(true)}>常用脚本</Button>
        <ModalTable ref={table} open={open} title={<Title>常用脚本</Title>} width={compact ? 740 : 780}
            rootClassName={styles.table} scrollMode="modal" rowKey="id" scroll={{x: 694}}
            onCancel={() => { sequence.current++; setOpen(false); }} request={request} requestEnabled={open} params={params}
            defaultPageSize={10} pagination={{pageSizeOptions: [10, 20, 50, 100]}}
            onRequestError={(error) => message.error(error?.message || "获取脚本失败")}
            search={{placeholder: "搜索名称或内容", maxLength: 100}}
            toolbar={{refresh: true, actions: [
                {key: "type", type: "filter", value: typeId || "", options: [{label: "全部分类", value: ""}, ...types.map(({id, name}) => ({label: name, value: String(id)}))],
                    onChange: (next: any) => setTypeId(next || undefined), icon: "FolderOutlined", ariaLabel: "脚本分类", matchWidth: true,
                    command: {label: "分类管理", icon: "SettingOutlined", onClick: () => categories.current?.open()}},
                {key: "add", type: "button", label: "添加", icon: "PlusOutlined", onClick: () => void edit()},
            ]}}
            columns={[
                {title: "脚本名称", dataIndex: "name", width: 190, ellipsis: true},
                {title: "分类", dataIndex: "type_id", width: 110, ellipsis: true, render: (id: number) => id ? typeNames[id] || "—" : "全部分类"},
                {title: "保存范围", dataIndex: "module", width: 90, render: (value: string) => value === module ? "全局脚本" : "实例脚本"},
                {title: "更新时间", dataIndex: "update_time", width: 160, render: (value: string) => <span className="script-time">{value ? value.slice(0, 16) : "—"}</span>},
                {title: "操作", key: "action", width: 90, align: "center", className: "action-cell", render: (_: unknown, item: any) =>
                    <ActionDropdown menu={{items: [
                        {key: "use", label: "填入", disabled, onClick: () => void use(item)},
                        {key: "edit", label: "编辑", onClick: () => void edit(item)},
                        {key: "delete", label: "删除", danger: true, onClick: () => remove(item)},
                    ]}}><Button size="small" autoInsertSpace={false}>操作</Button></ActionDropdown>
                },
            ]}
            modalProps={{zIndex: 1200, destroyOnHidden: true, style: {maxWidth: "calc(100vw - 24px)", top: "clamp(12px, 7vh, 72px)", paddingBottom: 24}}}/>
        <ProModal title={<Title>{editing?.id ? "编辑脚本" : "添加脚本"}</Title>} width={compact ? 540 : 580} onOk={form.submit}
            onCancel={() => { sequence.current++; setFormOpen(false); }}
            modalProps={{open: formOpen, zIndex: 1250, confirmLoading: busy || loading, okText: "保存", cancelText: "取消", mask: {closable: true},
                footer: loading ? null : undefined,
                afterClose: () => setFormActive(false),
                style: {maxWidth: "calc(100vw - 24px)", top: "clamp(12px, 7vh, 72px)", paddingBottom: 24}} as any}>
            <Form form={form} layout="vertical" variant="filled" onFinish={submit}>
                {formActive && (loading ? <div style={{minHeight: 220, display: "grid", placeItems: "center"}}><Spin/></div> : <>
                    <Form.Item name="name" label="脚本名称" rules={[{required: true, whitespace: true, message: "请输入脚本名称"}, {max: 100, message: "最多 100 个字符"}]}>
                        <Input placeholder="请输入脚本名称" maxLength={100}/>
                    </Form.Item>
                    {!editing?.id && <Row gutter={16}>
                        <Col xs={24} sm={12}>
                            <Form.Item name="module" label="保存范围" rules={[{required: true, message: "请选择保存范围"}]}>
                                <Select placeholder="请选择保存范围" options={scopes}/>
                            </Form.Item>
                        </Col>
                        <Col xs={24} sm={12}>
                            <Form.Item name="type_id" label="脚本分类" rules={[{required: true, message: "请选择脚本分类"}]}>
                                <Select placeholder="请选择脚本分类" loading={typesLoading} options={[{value: 0, label: "全部分类"}, ...formTypes.map(({id, name}) => ({value: id, label: name}))]}/>
                            </Form.Item>
                        </Col>
                    </Row>}
                    <Form.Item name="script" label="脚本内容" rules={[{required: true, whitespace: true, message: "请输入脚本内容"}, {max: 131072, message: "脚本内容过长"}]}>
                        {Editor ? <Editor className={styles.editor} mode={mode} height={compact ? "180px" : "200px"} width="100%" fontSize={12}
                            theme={theme?.isDarkTheme?.() ? "tomorrow_night" : "textmate"} placeholder="请输入脚本内容" showPrintMargin={false} wrapEnabled/>
                            : <Input.TextArea placeholder="请输入脚本内容" autoSize={{minRows: 7, maxRows: 12}}/>}
                    </Form.Item>
                </>)}
            </Form>
        </ProModal>
        <TypeManager key={scope} ref={categories} module={scope} scopes={scope === module ? undefined : [{label: "全局分类", value: module}, {label: "实例分类", value: scope}]} zIndex={1300}
            onChange={handleTypesChange} onMutation={() => table.current?.reload()}/>
    </>;
}
