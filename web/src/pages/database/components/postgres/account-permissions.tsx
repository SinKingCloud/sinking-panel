import {useEffect, useRef, useState} from "react";
import {Button, ConfigProvider, Form, Select} from "antd";
import {createStyles} from "antd-style";
import {Icon} from "sinking-antd";
import DatabaseSelect from "../shared/database-select";
import PermissionSelect from "../shared/permission-select";
import {useDatabaseClient} from "../shared/context";

const profiles = [
    {value: "readonly", label: "只读访问", privileges: ["USAGE", "SELECT"]},
    {value: "readwrite", label: "读写数据", privileges: ["USAGE", "SELECT", "INSERT", "UPDATE", "DELETE"]},
    {value: "manage", label: "数据库开发", privileges: ["USAGE", "SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER", "CREATE"]},
];
const labels: Record<string, string> = {
    CONNECT: "连接数据库", TEMPORARY: "创建临时表", USAGE: "访问架构与序列", CREATE: "创建对象",
    SELECT: "读取数据", INSERT: "新增数据", UPDATE: "修改数据", DELETE: "删除数据",
    TRUNCATE: "清空数据表", REFERENCES: "引用外键", TRIGGER: "创建触发器",
};
const useStyles = createStyles(({css, token}) => ({
    list: css`
        display: flex; flex-direction: column; gap: ${token.marginSM}px; min-width: 0;
        .postgres-permission-group { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; padding: ${token.paddingSM}px; min-width: 0; }
        .postgres-permission-heading { display: grid; grid-template-columns: minmax(0, 1fr) ${token.controlHeight}px; align-items: end; gap: ${token.marginXS}px; margin-bottom: ${token.margin}px; }
        .postgres-permission-heading > .ant-form-item { min-width: 0; margin-bottom: 0; }
        .postgres-permission-heading > .ant-btn, .postgres-schema-row > .ant-btn { width: ${token.controlHeight}px; height: ${token.controlHeight}px; padding: 0; flex: none; color: ${token.colorTextSecondary}; }
        .postgres-schema-heading, .postgres-schema-row { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1.25fr) ${token.controlHeight}px; gap: ${token.marginXS}px; align-items: start; }
        .postgres-schema-heading { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; margin-bottom: ${token.marginXS}px; }
        .postgres-schema-row { margin-bottom: ${token.marginXS}px; }
        .postgres-schema-row > .ant-form-item { margin-bottom: 0; min-width: 0; }
        .postgres-schema-row .ant-select { width: 100%; min-width: 0; }
        .postgres-schema-add { margin-top: ${token.marginXXS}px; }
        .postgres-permission-group > .ant-form-item { min-width: 0; }
        @media (max-width: 399px) {
            .postgres-schema-heading, .postgres-schema-row { grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr) ${token.controlHeight}px; gap: ${token.marginXXS}px; }
        }
    `,
}));

function DatabasePermissions({record, field, permissions, disabled, onRemove, onLoading, onError}: any) {
    const form = Form.useFormInstance();
    const {operate} = useDatabaseClient();
    const database = Form.useWatch(["grants", field.name, "database"], form);
    const [schemas, setSchemas] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);
    const loaded = useRef("");
    const pending = useRef("");
    const sequence = useRef(0);
    const active = useRef(true);
    const load = async () => {
        if (!database || loaded.current === database || pending.current === database) return;
        const request = ++sequence.current;
        pending.current = database;
        setLoading(true); onLoading(field.key, true);
        try {
            const data = await operate(record, "users", {action: "schemas", database});
            if (!active.current || request !== sequence.current) return;
            if (!Array.isArray(data) || !data?.every((name) => typeof name === "string")) throw new Error("架构列表数据不完整");
            const names = data?.filter((name) => !name.startsWith("pg_") && name !== "information_schema");
            setSchemas(names); loaded.current = database;
            // 只给新添加的数据库填默认架构，读取既有授权时不额外授予 public 权限。
            if (form.getFieldValue(["grants", field.name, "initialize"])) {
                const schema = names.includes("public") ? "public" : names[0];
                form.setFieldValue(["grants", field.name, "schemas"], schema
                    ? [{schema, privileges: profiles[1].privileges.filter((value) => permissions?.privileges?.includes(value))}] : []);
                form.setFieldValue(["grants", field.name, "initialize"], false);
            }
        } catch (reason: any) {
            if (active.current && request === sequence.current) onError(reason?.message || "加载数据库架构失败");
        } finally {
            if (active.current && request === sequence.current) { pending.current = ""; setLoading(false); onLoading(field.key, false); }
        }
    };
    useEffect(() => {
        active.current = true;
        return () => { active.current = false; sequence.current++; onLoading(field.key, false); };
    }, [field.key, field.name]);
    useEffect(() => {
        sequence.current++; loaded.current = ""; pending.current = ""; setSchemas([]); setLoading(false); onLoading(field.key, false);
        if (database && form.getFieldValue(["grants", field.name, "initialize"])) void load();
    }, [database, field.name]);
    const options = (permissions?.privileges || []).map((value: string) => ({value, label: labels[value] || value}));
    return <div className="postgres-permission-group">
        <div className="postgres-permission-heading">
        <Form.Item name={[field.name, "database"]} label={`数据库 ${field.name + 1}`} rules={[
            {required: true, message: "请选择数据库"},
            {validator: (_, value) => (form.getFieldValue("grants") || []).some((grant: any, index: number) => index !== field.name && value && grant.database === value)
                ? Promise.reject(new Error("该数据库已添加")) : Promise.resolve()},
        ]}>
            <DatabaseSelect record={record} placeholder="请选择数据库" disabled={disabled} onChange={() => {
                form.setFieldValue(["grants", field.name, "privileges"], ["CONNECT"]);
                form.setFieldValue(["grants", field.name, "schemas"], undefined);
                form.setFieldValue(["grants", field.name, "initialize"], true);
            }}/>
        </Form.Item>
            <Button color="default" variant="filled" disabled={disabled} icon={<Icon type="CloseOutlined"/>} aria-label="移除数据库" onClick={onRemove}/>
        </div>
        <Form.Item name={[field.name, "privileges"]} label="数据库级权限" getValueProps={(value) => ({value: value || []})}>
            <Select mode="multiple" showSearch={false} placeholder="请选择数据库级权限" disabled={disabled}
                options={(permissions?.database_privileges || []).map((value: string) => ({value, label: value === "CREATE" ? "创建架构" : labels[value] || value}))}/>
        </Form.Item>
        <Form.List name={[field.name, "schemas"]}>
            {(fields, {add, remove}) => <>
                {fields.length > 0 && <div className="postgres-schema-heading"><span>架构</span><span>架构内权限</span></div>}
                {fields.map((item) => <div className="postgres-schema-row" key={item.key}>
                    <Form.Item name={[item.name, "schema"]} rules={[
                        {required: true, message: "请选择架构"},
                        {validator: (_, value) => (form.getFieldValue(["grants", field.name, "schemas"]) || []).some((entry: any, index: number) => index !== item.name && value && entry.schema === value)
                            ? Promise.reject(new Error("该架构已添加")) : Promise.resolve()},
                    ]}>
                        <Select placeholder="选择架构" disabled={disabled || !database} loading={loading} showSearch optionFilterProp="label"
                            options={schemas.map((name) => ({value: name, label: name}))} onOpenChange={(open) => { if (open) void load(); }}/>
                    </Form.Item>
                    <Form.Item name={[item.name, "privileges"]} getValueProps={(value) => ({value: value || []})}>
                        <PermissionSelect profiles={profiles} options={options} disabled={disabled} allowEmpty/>
                    </Form.Item>
                    <Button color="default" variant="filled" disabled={disabled} icon={<Icon type="CloseOutlined"/>} aria-label="移除架构" onClick={() => remove(item.name)}/>
                </div>)}
                <Button className="postgres-schema-add" block type="dashed" disabled={disabled || !database || loading} loading={loading}
                    icon={<Icon type="PlusOutlined"/>} onClick={() => { add({schema: undefined, privileges: profiles[1].privileges.filter((value) => permissions?.privileges?.includes(value))}); void load(); }}>添加架构</Button>
            </>}
        </Form.List>
    </div>;
}

export default function DatabasePermissionFields({record, permissions, disabled, onLoading, onError}: any) {
    const {styles} = useStyles();
    return <ConfigProvider wave={{disabled: true}}><Form.List name="grants">
        {(fields, {add, remove}) => <div className={styles.list}>
            {fields.map((field) => <DatabasePermissions key={field.key} field={field} record={record} permissions={permissions}
                disabled={disabled} onRemove={() => remove(field.name)} onLoading={onLoading} onError={onError}/>)}
            <Button block type="dashed" disabled={disabled || fields.length >= 1000} icon={<Icon type="PlusOutlined"/>}
                onClick={() => add({database: undefined, privileges: ["CONNECT"], initialize: true})}>添加数据库</Button>
        </div>}
    </Form.List></ConfigProvider>;
}
