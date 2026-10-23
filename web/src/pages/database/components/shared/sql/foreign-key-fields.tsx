import {useLayoutEffect, useRef, useState} from "react";
import {Button, ConfigProvider, Form, Input, Select} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {createStyles} from "antd-style";
import {useDatabaseClient} from "../context";

const useStyles = createStyles(({css, token}) => ({
    root: css`
        --foreign-key-font-size: 12px;
        --foreign-key-control-height: 28px;
        display: flex; flex-direction: column; gap: 10px;
        &[data-compact] { --foreign-key-font-size: 11px; --foreign-key-control-height: 24px; gap: 8px; }
        .foreign-key-row { min-width: 0; padding: 10px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; }
        .foreign-key-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 8px; }
        .foreign-key-title { display: inline-flex; align-items: center; gap: 6px; color: ${token.colorTextSecondary}; font-size: var(--foreign-key-font-size); }
        .foreign-key-title > .anticon { color: ${token.colorTextTertiary}; }
        .foreign-key-heading .ant-btn { width: var(--foreign-key-control-height); height: var(--foreign-key-control-height); padding: 0; }
        .foreign-key-fields { display: flex; flex-direction: column; gap: 10px; }
        .ant-form-item { min-width: 0; margin-bottom: 0; }
        .ant-form-item-label { padding-bottom: 4px; }
        && :is(.ant-form-item-label > label, .ant-input, .ant-select-selector, .ant-select-selection-item, .ant-select-selection-placeholder, .ant-btn) { font-size: var(--foreign-key-font-size); }
        .foreign-key-add { width: 100%; height: var(--foreign-key-control-height); }
        &[data-compact] .foreign-key-row { padding: 8px; }
        &[data-compact] .foreign-key-fields { gap: 8px; }
    `,
    fields: css`
        .foreign-key-tables { display: grid; grid-template-columns: minmax(0, 1fr) 16px minmax(0, 1fr) var(--foreign-key-control-height, ${token.controlHeight}px); column-gap: ${token.marginXS}px; }
        .foreign-key-tables > .ant-form-item { min-width: 0; margin-bottom: 0; }
        .foreign-key-tables > .ant-form-item:last-child { grid-column: 3 / -1; }
        .foreign-key-mappings { --foreign-key-control-height: 28px; display: flex; flex-direction: column; gap: ${token.marginXS}px; }
        &[data-compact] .foreign-key-mappings { --foreign-key-control-height: 24px; }
        .foreign-key-mapping { display: grid; grid-template-columns: minmax(0, 1fr) 16px minmax(0, 1fr) var(--foreign-key-control-height, ${token.controlHeight}px); align-items: center; gap: ${token.marginXS}px; }
        .foreign-key-mapping > .ant-select { min-width: 0; }
        .foreign-key-mapping > .anticon { color: ${token.colorTextQuaternary}; font-size: var(--foreign-key-font-size, ${token.fontSizeSM}px); }
        .foreign-key-mapping > .ant-btn { padding-inline: 0; color: ${token.colorTextSecondary}; }
        .foreign-key-mapping-heading { display: grid; grid-template-columns: minmax(0, 1fr) 16px minmax(0, 1fr) var(--foreign-key-control-height, ${token.controlHeight}px); gap: ${token.marginXS}px; color: ${token.colorTextSecondary}; font-size: var(--foreign-key-font-size, ${token.fontSizeSM}px); }
        .foreign-key-mapping-heading > span:nth-child(2) { grid-column: 3; }
        .foreign-key-add { width: 100%; }
        .foreign-key-rules { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: ${token.marginSM}px; }
        .foreign-key-rules > .ant-form-item { min-width: 0; }
        @media (max-width: 399px) {
            .foreign-key-tables, .foreign-key-mapping, .foreign-key-mapping-heading { grid-template-columns: minmax(0, 1fr) 12px minmax(0, 1fr) var(--foreign-key-control-height, ${token.controlHeight}px); column-gap: ${token.marginXXS}px; }
        }
    `,
}));

export function validateForeignKeys(keys: any[] = []) {
    if (!Array.isArray(keys) || keys.length > 32) throw new Error("外键最多设置 32 项");
    const names = new Set<string>();
    keys.forEach((key, index) => {
        const prefix = `外键 ${index + 1}：`;
        const name = String(key.name || "").trim();
        if (!name) throw new Error(prefix + "请输入名称");
        if (names.has(name.toLowerCase())) throw new Error(prefix + "名称重复");
        names.add(name.toLowerCase());
        if (!key.columns?.length || key.columns.some((column: string) => !column)) throw new Error(prefix + "请选择当前表字段");
        if (!key.reference_table) throw new Error(prefix + "请选择引用表");
        if (!key.reference_columns?.length || key.reference_columns.some((column: string) => !column)) throw new Error(prefix + "请选择引用字段");
        if (key.columns.length !== key.reference_columns.length) throw new Error(prefix + "两侧字段数量应一致");
        if (new Set(key.columns).size !== key.columns.length || new Set(key.reference_columns).size !== key.reference_columns.length) throw new Error(prefix + "字段不能重复");
    });
}

const ForeignKeyRow = ({record, table, columns, tables = [], schemas = [], schema, value, disabled, active, onChange, onRemove, onError, onReady, index, single, controlHeight}: any) => {
    const {operate} = useDatabaseClient();
    const [targets, setTargets] = useState<any[]>([]);
    const [targetColumns, setTargetColumns] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);
    const request = useRef(0);
    const referenceSchema = value.reference_schema || schema;
    const showReferenceSchema = Number(record.type) !== 6 && (schemas.length > 1 || Boolean(referenceSchema && referenceSchema !== schema));
    const referenceTable = value.reference_table;
    useLayoutEffect(() => {
        const sequence = ++request.current;
        if (!active) return;
        setTargetColumns([]);
        if (!record?.id) { setTargets(tables); setLoading(false); onReady?.(); return; }
        setLoading(true);
        void Promise.all([
            referenceSchema === schema ? Promise.resolve(tables) : operate(record, "tables", {schema: referenceSchema}),
            referenceTable ? operate(record, "columns", {schema: referenceSchema, table: referenceTable}) : Promise.resolve([]),
        ]).then(([items, fields]) => {
            if (request.current !== sequence) return;
            setTargets(Array.isArray(items) ? items : []);
            setTargetColumns(Array.isArray(fields) ? fields : []);
            setLoading(false);
            onReady?.();
        }).catch((reason) => {
            if (request.current !== sequence) return;
            request.current++;
            onError?.(reason);
        });
        return () => { request.current++; };
    }, [record?.id, referenceSchema, referenceTable, schema, tables, active]);
    const update = (changes: any) => onChange({...value, ...changes});
    const mappingCount = Math.max(value.columns?.length || 0, value.reference_columns?.length || 0, 1);
    const changeMapping = (position: number, key: "columns" | "reference_columns", selected: string) => {
        const items = Array.from({length: mappingCount}, (_, index) => value[key]?.[index] || "");
        items[position] = selected;
        update({[key]: items});
    };
    const referenceActions = [
        {value: "NO ACTION", label: "禁止操作（NO ACTION）"},
        {value: "RESTRICT", label: "限制操作（RESTRICT）"},
        {value: "CASCADE", label: "同步处理（CASCADE）"},
        {value: "SET NULL", label: "设为空值（SET NULL）"},
        {value: "SET DEFAULT", label: "设为默认值（SET DEFAULT）"},
    ].filter((item) => !(Number(record.type) === 5 && item.value === "RESTRICT"));
    return <div className={single ? undefined : "foreign-key-row"}>
        {!single && <div className="foreign-key-heading"><span className="foreign-key-title"><Icon type="LinkOutlined"/>外键 {index + 1}</span>
            {onRemove && <Button type="text" danger disabled={disabled} aria-label={`删除外键 ${index + 1}`} icon={<Icon type="DeleteOutlined"/>} onClick={onRemove}/>}</div>}
        <div className="foreign-key-fields">
            <Form.Item label="外键名称" required><Input variant="filled" value={value.name} disabled={disabled} placeholder="例如 fk_orders_user"
                onChange={(event) => update({name: event.target.value})}/></Form.Item>
            {showReferenceSchema && <Form.Item label="引用架构" required><Select variant="filled" value={referenceSchema} disabled={disabled || schemas.length < 2}
                placeholder="请选择引用架构" options={schemas.map((name: string) => ({value: name, label: name}))}
                onChange={(reference_schema) => update({reference_schema, reference_table: undefined, reference_columns: []})}/></Form.Item>}
            <Form.Item label="字段对应" required>
                <ConfigProvider theme={single ? {components: {
                    Input: {controlHeight}, Select: {controlHeight}, Button: {controlHeight},
                }} : undefined}>
                <div className="foreign-key-mappings">
                    <div className="foreign-key-tables">
                        <Form.Item label="当前数据表"><Input variant="filled" value={table} readOnly aria-label="当前数据表" placeholder="先填写数据表名称"/></Form.Item>
                        <Form.Item label="引用数据表" required><Select variant="filled" value={referenceTable} disabled={disabled} loading={loading} placeholder="请选择引用数据表"
                            showSearch optionFilterProp="label" options={targets.filter((item) => !String(item.type).includes("VIEW")).map((item) => ({value: item.name, label: item.name}))}
                            onChange={(reference_table) => update({reference_table, reference_columns: []})}/></Form.Item>
                    </div>
                    <div className="foreign-key-mapping-heading"><span>当前表字段</span><span>引用表字段</span></div>
                    {Array.from({length: mappingCount}, (_, position) => <div key={position} className="foreign-key-mapping">
                        <Select variant="filled" value={value.columns?.[position] || undefined} disabled={disabled} showSearch optionFilterProp="label"
                            aria-label={`当前表字段 ${position + 1}`} placeholder="选择字段" options={columns.filter((name: string) => !value.columns?.includes(name) || value.columns[position] === name)
                                .map((name: string) => ({value: name, label: name}))}
                            onChange={(selected) => changeMapping(position, "columns", selected)}/>
                        <Icon type="ArrowRightOutlined"/>
                        <Select variant="filled" value={value.reference_columns?.[position] || undefined} disabled={disabled || !referenceTable || loading} loading={loading}
                            aria-label={`引用表字段 ${position + 1}`} placeholder={referenceTable ? "选择字段" : "先选择数据表"} showSearch optionFilterProp="label"
                            options={targetColumns.filter((item) => !value.reference_columns?.includes(item.name) || value.reference_columns[position] === item.name)
                                .map((item) => ({value: item.name, label: item.name}))}
                            onChange={(selected) => changeMapping(position, "reference_columns", selected)}/>
                        <Button color="default" variant="filled" disabled={disabled || mappingCount === 1} aria-label={`移除字段对应 ${position + 1}`}
                            icon={<Icon type="CloseOutlined"/>} onClick={() => update({
                                columns: Array.from({length: mappingCount}, (_, index) => value.columns?.[index] || "").filter((_, index) => index !== position),
                                reference_columns: Array.from({length: mappingCount}, (_, index) => value.reference_columns?.[index] || "").filter((_, index) => index !== position),
                            })}/>
                    </div>)}
                    <Button className="foreign-key-add" type="dashed" disabled={disabled || mappingCount >= columns.length} icon={<Icon type="PlusOutlined"/>}
                        onClick={() => update({columns: [...Array.from({length: mappingCount}, (_, index) => value.columns?.[index] || ""), ""],
                            reference_columns: [...Array.from({length: mappingCount}, (_, index) => value.reference_columns?.[index] || ""), ""]})}>添加字段对应</Button>
                </div>
                </ConfigProvider>
            </Form.Item>
            <div className="foreign-key-rules">
                <Form.Item label="关联记录删除时"><Select variant="filled" placeholder="请选择删除规则" value={value.on_delete || "NO ACTION"} disabled={disabled}
                    options={referenceActions} onChange={(on_delete) => update({on_delete})}/></Form.Item>
                <Form.Item label="关联字段更新时"><Select variant="filled" placeholder="请选择更新规则" value={value.on_update || "NO ACTION"} disabled={disabled}
                    options={referenceActions} onChange={(on_update) => update({on_update})}/></Form.Item>
            </div>
        </div>
    </div>;
};

export default ({record, table, columns = [], tables = [], schemas = [], schema, value = [], onChange, onError, onReady, active = true, disabled, maxItems = 32}: any) => {
    const {styles} = useStyles();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const fontSize = compact ? 11 : 12;
    const controlHeight = compact ? 24 : 28;
    const single = maxItems === 1;
    const fields = <div className={single ? styles.fields : `${styles.fields} ${styles.root}`} data-compact={compact || undefined}>
        {value.map((key: any, index: number) => <ForeignKeyRow key={index} index={index} record={record} table={table} value={key} columns={columns} tables={tables} schemas={schemas} schema={schema}
            single={single} controlHeight={controlHeight} disabled={disabled} active={active} onError={onError} onReady={onReady} onChange={(key: any) => onChange(value.map((item: any, position: number) => position === index ? key : item))}
            onRemove={() => onChange(value.filter((_: any, position: number) => position !== index))}/>)}
        {!single && value.length < maxItems && <Button className="foreign-key-add" type="dashed" disabled={disabled} icon={<Icon type="PlusOutlined"/>}
            onClick={() => onChange([...value, {name: "", columns: [], reference_schema: schema, reference_columns: [], on_delete: "NO ACTION", on_update: "NO ACTION"}])}>添加外键</Button>}
    </div>;
    if (single) return <ConfigProvider wave={{disabled: true}}>{fields}</ConfigProvider>;
    return <ConfigProvider wave={{disabled: true}} componentSize="small" theme={{
        token: {fontSize, fontSizeSM: fontSize, controlHeight, controlHeightSM: controlHeight},
        components: {
            Input: {inputFontSize: fontSize, inputFontSizeSM: fontSize, controlHeight, controlHeightSM: controlHeight},
            Select: {optionFontSize: fontSize, optionHeight: controlHeight, controlHeight, controlHeightSM: controlHeight},
            Form: {labelFontSize: fontSize, controlHeight, controlHeightSM: controlHeight},
            Button: {contentFontSize: fontSize, contentFontSizeSM: fontSize, controlHeight, controlHeightSM: controlHeight},
        },
    }}>{fields}</ConfigProvider>;
};
