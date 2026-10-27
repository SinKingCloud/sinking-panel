import {useEffect, useRef, useState} from "react";
import {Alert, Button, Collapse, ConfigProvider, Form, Select, Spin, theme as antdTheme} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import useStyles from "./component-styles/account";
import {canChangePassword} from "./client-common";
import DatabaseSelect from "./database-select";

interface PermissionProfile {
    value: string;
    label: string;
    description: string;
    privileges: string[];
}

export function AccountEditor({editor, form, busy, loading, initializing, rows, onSelect, onCancel, onFinish, children, repair, submitDisabled, accountSelect, accountLabel = "账户"}: any) {
    const {styles} = useStyles();
    const {token} = antdTheme.useToken();
    const theme = useTheme();
    const lastContent = useRef<any>(undefined);
    const [visible, setVisible] = useState(false);
    useEffect(() => {
        if (editor) lastContent.current = {editor, children, repair, accountSelect, initializing};
    });
    useEffect(() => { if (editor) setVisible(true); }, [editor]);
    // 退出动画仍展示最后一次打开的表单，实际开关继续使用 editor。
    const content = editor ? {editor, children, repair, accountSelect, initializing} : visible ? lastContent.current : undefined;
    const current = content?.editor;
    return <ProModal title={<Title>{current?.action === "create" ? `添加${accountLabel}` : current?.action === "password" ? "修改密码" : current?.action === "access" ? "访问来源" : `${accountLabel}权限`}</Title>}
        onCancel={onCancel} onOk={form.submit} modalProps={{
            open: Boolean(editor), className: styles.editor,
            width: current?.action === "password" ? (theme?.isCompactTheme?.() ? 340 : 380)
                : current?.action === "access" ? (theme?.isCompactTheme?.() ? 360 : 400) : (theme?.isCompactTheme?.() ? 440 : 480),
            style: {maxWidth: "calc(100vw - 24px)"},
            confirmLoading: busy, okText: "保存", cancelText: "取消",
            okButtonProps: {disabled: content?.initializing || submitDisabled || current?.selectAccount && !content?.accountSelect && (loading || !rows.some(canChangePassword))},
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
                    {content?.accountSelect || <Select loading={loading} placeholder={`选择要修改密码的${accountLabel}`} options={rows.filter(canChangePassword).map((account: any) => ({value: JSON.stringify([account.name, account.host || ""]),
                        label: `${account.name}${account.host ? `@${account.host}` : ""}`}))} onChange={onSelect}/>}
                </Form.Item>}
                {content?.children}
            </>}
        </Form>
        </ConfigProvider>
    </ProModal>;
}

export function PermissionFields({profiles, options, value = [], onChange, disabled, allowEmpty = false, showDescription = true, id}: {
    profiles: PermissionProfile[];
    options: {value: string; label: string}[];
    value?: string[];
    onChange?: (value: string[]) => void;
    disabled?: boolean;
    allowEmpty?: boolean;
    showDescription?: boolean;
    id?: string;
}) {
    const {styles} = useStyles();
    const selected = profiles.find((item) => item.privileges.length === value.length && item.privileges.every((privilege) => value.includes(privilege)));
    return <div className="account-permission-fields">
        <Select id={id} className="account-permission-select" mode="multiple" value={value} onChange={onChange} disabled={disabled}
            aria-label="权限方案" options={options} showSearch={false} maxTagCount={0}
            placeholder={allowEmpty ? "不授予权限" : "选择权限"} maxTagPlaceholder={() => selected?.label || `自定义 · ${value.length} 项`}
            tagRender={({label}) => <span className="account-permission-summary">{label}</span>}
            popupMatchSelectWidth={280} listHeight={240} classNames={{popup: {root: styles.permissionPopup}}}
            popupRender={(menu) => <>
                <div className="account-permission-presets" onMouseDown={(event) => event.preventDefault()}>
                    {profiles.map((item) => <Button key={item.value} size="small" color={selected?.value === item.value ? "primary" : "default"}
                        variant="filled" disabled={disabled || item.privileges.some((privilege) => !options.some((option) => option.value === privilege))}
                        onClick={() => onChange?.([...item.privileges])}>{item.label}</Button>)}
                </div>
                {menu}
            </>}/>
        {showDescription && <div className="account-permission-description">{allowEmpty && !value.length ? "移除此数据库的直接权限。"
            : selected?.description || "按需要选择具体权限。"}</div>}
    </div>;
}

export function DatabasePermissionFields({record, privileges, profiles, disabled, required = false}: any) {
    const form = Form.useFormInstance();
    const grants = Form.useWatch("grants", form) || [];
    const labels: Record<string, string> = {
        SELECT: "读取数据", INSERT: "新增数据", UPDATE: "修改数据", DELETE: "删除数据", CREATE: "创建表", ALTER: "修改表结构",
        INDEX: "管理索引", DROP: "删除表", EXECUTE: "执行存储程序", REFERENCES: "引用外键", TRIGGER: "管理触发器",
        "CREATE VIEW": "创建视图", "SHOW VIEW": "查看视图", "CREATE ROUTINE": "创建程序", "ALTER ROUTINE": "修改程序",
        "CREATE TEMPORARY TABLES": "创建临时表", "LOCK TABLES": "锁定表", EVENT: "管理事件", "DELETE HISTORY": "删除历史数据",
    };
    const options = (privileges || []).map((value: string) => ({value, label: labels[value] || value}));
    const permissionProfiles = [...profiles, {value: "all", label: "全部权限", description: "授予此数据库支持配置的全部权限。", privileges: privileges || []}];
    return <Form.List name="grants" rules={required ? [{validator: async (_, grants) => {
        if (!grants?.length) throw new Error("请至少添加一个数据库");
    }}] : undefined}>
        {(fields, {add, remove}, {errors}) => <div className="account-permission-list">
            {fields.length > 0 && <div className="account-permission-heading" aria-hidden="true"><span>数据库</span><span>权限方案</span></div>}
            {fields.map((field) => <div className="account-permission-item" key={field.key}>
                <Form.Item className="account-permission-database" name={[field.name, "database"]} rules={[
                    {required: true, whitespace: true, message: "请选择数据库"},
                    {validator: (_, value) => (form.getFieldValue("grants") || []).some((grant: any, index: number) =>
                        index !== field.name && value && grant?.database === value)
                        ? Promise.reject(new Error("该数据库已添加")) : Promise.resolve()},
                ]}>
                    <DatabaseSelect record={record} disabled={disabled}
                        disabledValues={grants.filter((_: any, index: number) => index !== field.name).map((grant: any) => grant?.database).filter(Boolean)}/>
                </Form.Item>
                <Form.Item name={[field.name, "privileges"]} getValueProps={(value) => ({value: value || []})}
                    rules={required ? [{required: true, type: "array", min: 1, message: "请至少选择一项权限"}] : undefined}>
                    <PermissionFields profiles={permissionProfiles} options={options} disabled={disabled} allowEmpty={!required} showDescription={false}/>
                </Form.Item>
                <Button color="default" variant="filled" disabled={disabled} icon={<Icon type="CloseOutlined"/>}
                    aria-label="移除数据库" title="移除数据库" onClick={() => remove(field.name)}/>
            </div>)}
            <ConfigProvider wave={{disabled: true}}>
                <Button block type="dashed" disabled={disabled || fields.length >= 1000} icon={<Icon type="PlusOutlined"/>}
                    onClick={() => add({database: undefined, privileges: [...(profiles.find((item: PermissionProfile) => item.value === "manage"
                        && item.privileges.every((privilege) => privileges.includes(privilege)))?.privileges || [])]})}>添加数据库</Button>
            </ConfigProvider>
            <Form.ErrorList errors={errors}/>
        </div>}
    </Form.List>;
}

export function ConnectionDetails({data, loading, description}: any) {
    return <Collapse className="account-network" ghost items={[{key: "connection", label: "连接与访问说明", children: <>
        {loading ? <Spin size="small"/> : Array.isArray(data?.items) && data?.items?.length ? <dl className="account-network-list">
            {data?.items?.map((item: any, index: number) => <div key={`${item?.label}:${index}`} style={{display: "contents"}}><dt>{item?.label}</dt><dd>{item?.value || "未设置"}</dd></div>)}
        </dl> : null}
        <div className="account-network-note">{data?.note || description}</div>
    </>}]} />;
}
