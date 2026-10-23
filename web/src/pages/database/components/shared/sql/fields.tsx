import {Button, ConfigProvider, DatePicker, Input, InputNumber, Select, Space} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {createStyles} from "antd-style";
import dayjs from "dayjs";
import {sqlColumnKind} from "./values";

const useStyles = createStyles(({css, token}) => ({
    field: css`
        display: flex; align-items: flex-start; gap: 8px; min-width: 0;
        .sql-field-control { flex: 1; min-width: 0; }
        .sql-field-null { flex: none; padding-inline: 10px; }
    `,
    datePopup: css`
        @media (max-width: 575px) {
            max-width: calc(100vw - 24px);
            .ant-picker-panel-container { max-height: calc(100dvh - 24px); overflow-y: auto; }
            .ant-picker-datetime-panel { flex-direction: column; }
            .ant-picker-datetime-panel .ant-picker-time-panel { border-inline-start: 0; border-top: 1px solid ${token.colorBorderSecondary}; }
            .ant-picker-datetime-panel .ant-picker-time-panel-column { height: 144px; flex: 1; }
        }
    `,
    filters: css`
        --sql-filter-font-size: 13px;
        min-width: 0;
        padding-block: ${token.paddingXS}px;
        color: ${token.colorText};
        font-size: var(--sql-filter-font-size);
        &[data-compact] { --sql-filter-font-size: 12px; }
        && .ant-input, && .ant-select, && .sql-filter-controls .ant-btn { font-size: var(--sql-filter-font-size); }
        .sql-filter-mode { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: ${token.margin}px; }
        .sql-filter-mode > .ant-select { width: 224px; max-width: 100%; min-width: 0; }
        .sql-filter-reset { flex: none; }
        .sql-filter-reset:not(:disabled) { color: ${token.colorTextSecondary}; }
        .sql-filter-columns, .sql-filter-row { display: grid; grid-template-columns: minmax(0, 1fr) 108px minmax(0, 1.25fr) 28px; gap: 8px; }
        .sql-filter-columns { margin-bottom: 8px; color: ${token.colorTextSecondary}; font-size: 12px; }
        .sql-filter-columns > span { padding-inline: 10px; }
        .sql-filter-row { align-items: start; margin-bottom: 10px; }
        .sql-filter-row:last-child { margin-bottom: 0; }
        .sql-filter-row > * { min-width: 0; }
        .sql-filter-row > .ant-select { width: 100%; }
        .sql-filter-remove { width: 28px; padding-inline: 0; color: ${token.colorTextTertiary}; align-self: start; }
        .sql-filter-empty { padding: 18px 12px; text-align: center; color: ${token.colorTextTertiary}; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; font-size: ${token.fontSizeSM}px; }
        .sql-filter-actions { margin-top: ${token.marginSM}px; }
        @media (max-width: 575px) {
            .sql-filter-mode > .ant-select { flex: 1; width: 0; }
            .sql-filter-columns { display: none; }
            .sql-filter-row { grid-template-columns: minmax(0, 1fr) minmax(94px, .75fr) 28px; gap: 8px; padding-bottom: 12px; margin-bottom: 12px; border-bottom: 1px solid ${token.colorBorderSecondary}; }
            .sql-filter-row:last-child { padding-bottom: 0; border-bottom: 0; }
            .sql-filter-value { grid-column: 1 / 3; grid-row: 2; }
            .sql-filter-remove { grid-column: 3; grid-row: 1; }
        }
    `,
}));

export function SqlValueInput({id, kind, value, onChange, disabled, label, multiline = false, placeholder, onClear, column}: any) {
    const {styles} = useStyles();
    if (column && kind === "number") {
        const decimal = /^(?:decimal|numeric)\s*\(\d+\s*,\s*(\d+)\)/i.exec(column.type);
        const scale = Math.min(65, Number(decimal?.[1] || 0));
        const canStep = value === "" || /^-?(?:\d+(?:\.\d*)?|\.\d+)$/.test(value);
        // 输入保留原文；只有主动增减才采用组件计算结果，避免科学计数法和未完成输入被改写。
        return <InputNumber<string> id={id} variant="filled" stringMode value={value || null} disabled={disabled} aria-label={label}
            style={{width: "100%"}} inputMode="decimal" placeholder={placeholder ?? "请输入数字"} step={scale ? `0.${"0".repeat(scale - 1)}1` : "1"}
            controls={canStep} keyboard={canStep} changeOnBlur={false} parser={(content) => content ?? ""} formatter={() => value ?? ""}
            onInput={(content) => { if (content === "" && onClear) onClear(); else onChange(content); }} onStep={(content) => onChange(String(content))}/>;
    }
    const dateType = /^(date|datetime2?|smalldatetime|timestamp(?:tz)?|time(?:tz)?|year)\b/i.exec(column?.type || "")?.[1].toLowerCase().replace(/^timetz$/, "time").replace(/^timestamptz$/, "timestamp").replace(/^(datetime2|smalldatetime)$/, "datetime");
    if (dateType) {
        const withTime = /datetime|timestamp/.test(dateType);
        const dateLabel = dateType === "time" ? "时间" : dateType === "year" ? "年份" : withTime ? "日期时间" : "日期";
        const format = dateType === "year" ? "YYYY" : dateType === "time" ? "HH:mm:ss" : withTime ? "YYYY-MM-DD HH:mm:ss" : "YYYY-MM-DD";
        const content = String(value || "");
        const parsed = dayjs(dateType === "time" ? `2000-01-01T${content}` : dateType === "year" ? `${content}-01-01` : content.replace(" ", "T"));
        const selected = parsed.isValid() && parsed.format(format) === content.slice(0, format.length) ? parsed : null;
        // 保留数据库返回的原始时间精度和偏移；选择器只负责快捷填入。
        return <Space.Compact block>
            <Input id={id} variant="filled" value={value} disabled={disabled} aria-label={label} placeholder={placeholder ?? `请选择或输入${dateLabel}`}
                style={{minWidth: 0, flex: 1}} allowClear={Boolean(onClear)} onClear={onClear}
                onChange={(event) => { if (event.type !== "click" || !onClear) onChange(event.target.value); }}/>
            <DatePicker variant="filled" picker={dateType === "time" ? "time" : dateType === "year" ? "year" : "date"}
                showTime={withTime} format={format} value={selected} inputReadOnly allowClear={false} disabled={disabled}
                aria-label={`${label} 选择${dateLabel}`} placeholder={`请选择${dateLabel}`}
                classNames={{popup: {root: styles.datePopup}}}
                style={{width: 32, flex: "none", paddingInline: 0}}
                styles={{input: {width: 0, minWidth: 0, flex: "none", padding: 0, opacity: 0}, suffix: {marginInlineStart: 0, flex: 1, justifyContent: "center"}}}
                onChange={(date) => {
                    if (!date) return;
                    const fraction = (withTime || dateType === "time") ? /\.\d{1,6}$/.exec(content)?.[0] || "" : "";
                    onChange(date.format(format) + fraction);
                }}/>
        </Space.Compact>;
    }
    if (kind === "boolean") return <Select id={id} variant="filled" aria-label={label} value={value || undefined} disabled={disabled} style={{width: "100%"}}
        placeholder={placeholder ?? "请选择是或否"} options={[{value: "true", label: "是（true）"}, {value: "false", label: "否（false）"}]} allowClear={Boolean(onClear)} onClear={onClear}
        onChange={(content) => { if (content !== undefined || !onClear) onChange(content); }}/>;
    if (kind === "null") return <Input id={id} variant="filled" value="NULL" placeholder="NULL" disabled aria-label={label}/>;
    if (multiline || kind === "binary") return <Input.TextArea id={id} variant="filled" value={value} disabled={disabled} aria-label={label}
        autoSize={{minRows: 2, maxRows: 5}} placeholder={placeholder ?? (kind === "binary" ? "请输入 Base64 内容" : `请输入${label || "内容"}`)}
        allowClear={Boolean(onClear)} onClear={onClear}
        onChange={(event) => { if (event.type !== "click" || !onClear) onChange(event.target.value); }} spellCheck={false}/>;
    return <Input id={id} variant="filled" value={value} disabled={disabled} aria-label={label} inputMode={kind === "number" ? "decimal" : undefined}
        placeholder={placeholder ?? (kind === "number" ? "请输入数字" : `请输入${label || "内容"}`)} allowClear={Boolean(onClear)} onClear={onClear}
        onChange={(event) => { if (event.type !== "click" || !onClear) onChange(event.target.value); }}/>;
}

export function SqlField({id, column, editing, value, onChange, disabled}: any) {
    const {styles} = useStyles();
    const isNull = value?.mode === "null";
    const multiline = /text|json|xml|clob/i.test(column.type) || (value?.value || "").includes("\n");
    const hasDefault = column.auto_increment || column.default != null || column.nullable;
    const placeholder = isNull ? "NULL" : value?.mode === "default"
        ? column.auto_increment ? "自动生成，无需填写" : column.default != null ? `留空使用默认值：${column.default === "" ? "空字符串" : column.default}` : undefined
        : undefined;
    if (column.generated) return <Input id={id} variant="filled" value={!editing ? "由数据库生成" : isNull ? "NULL" : value?.value} placeholder="由数据库自动生成" disabled aria-label={column.name}/>;
    return <div className={styles.field}>
        <div className="sql-field-control"><SqlValueInput id={id} column={column} kind={value?.kind} value={value?.mode === "value" ? value.value : ""}
            label={column.name} disabled={disabled || isNull} multiline={multiline} placeholder={placeholder}
            onClear={() => onChange({...value, mode: !editing && hasDefault ? "default" : column.nullable ? "null" : "value", value: ""})}
            onChange={(content: any) => onChange({...value, mode: "value", value: content})}/></div>
        {column.nullable && <Button className="sql-field-null" color={isNull ? "primary" : "default"} variant="filled" autoInsertSpace={false}
            disabled={disabled} aria-label={`${column.name} 设为 NULL`} aria-pressed={isNull}
            onClick={() => onChange({...value, mode: isNull ? value.previousMode ?? "value" : "null",
                previousMode: isNull ? undefined : value?.mode})}>NULL</Button>}
    </div>;
}

export function SqlFilters({columns, value, onChange, disabled, logic, onLogicChange, onReset, resetDisabled}: any) {
    const {styles} = useStyles();
    const compact = useTheme()?.isCompactTheme?.();
    const fontSize = compact ? 12 : 13;
    const controlHeight = compact ? 28 : 32;
    const update = (index: number, changes: any) => onChange(value.map((item: any, position: number) => position === index ? {...item, ...changes} : item));
    return <div className={styles.filters} data-compact={compact || undefined}>
        <ConfigProvider theme={{token: {fontSize, controlHeight}, components: {
            Input: {controlHeight, inputFontSize: fontSize},
            Select: {controlHeight, optionFontSize: fontSize},
            Button: {controlHeight, contentFontSize: fontSize},
        }}}>
            <div className="sql-filter-controls">
                <div className="sql-filter-mode">
                    <Select variant="filled" placeholder="请选择筛选方式" value={logic} onChange={onLogicChange} disabled={disabled} aria-label="筛选方式"
                        options={[{value: "and", label: "同时满足以下条件"}, {value: "or", label: "满足以下其中一个条件"}]}/>
                    <Button className="sql-filter-reset" color="default" variant="filled" disabled={disabled || resetDisabled} onClick={onReset}>重置</Button>
                </div>
                {value.length ? <>
                    <div className="sql-filter-columns" aria-hidden="true"><span>字段</span><span>条件</span><span>筛选值</span></div>
                    <div className="sql-filter-conditions">
                        {value.map((filter: any, index: number) => {
                            const column = columns.find((item: any) => item.name === filter.column);
                            const kind = filter.operator === "like" ? "text" : sqlColumnKind(column || {});
                            const empty = filter.operator === "null" || filter.operator === "not_null";
                            return <div className="sql-filter-row" key={index}>
                                <Select variant="filled" showSearch optionFilterProp="label" placeholder="选择字段" value={filter.column || undefined} disabled={disabled} aria-label={`筛选字段 ${index + 1}`}
                                    options={columns.map((item: any) => ({value: item.name, label: item.name}))}
                                    onChange={(name) => update(index, {column: name, value: filter.operator === "in" ? [] : ""})}/>
                                <Select variant="filled" placeholder="请选择比较方式" value={filter.operator} disabled={disabled} aria-label={`比较方式 ${index + 1}`}
                                    options={[["eq", "等于"], ["ne", "不等于"], ["gt", "大于"], ["gte", "大于等于"], ["lt", "小于"], ["lte", "小于等于"], ["like", "模糊匹配"], ["in", "属于以下值"], ["null", "是 NULL"], ["not_null", "不是 NULL"]].map(([value, label]) => ({value, label}))}
                                    onChange={(operator) => update(index, {operator, value: operator === "in" ? [] : filter.operator === "in" ? "" : filter.value})}/>
                                <div className="sql-filter-value">{empty ? <Input variant="filled" disabled placeholder="无需填写值"/>
                                    : filter.operator === "in" ? <Select variant="filled" mode="tags" open={false} maxTagCount="responsive" style={{width: "100%"}} value={filter.value} disabled={disabled}
                                        aria-label={`筛选值 ${index + 1}`} placeholder="输入值后按回车" onChange={(items) => update(index, {value: items})}/>
                                        : <SqlValueInput kind={kind} value={filter.value} label={`筛选值 ${index + 1}`} disabled={disabled}
                                            placeholder={kind === "number" ? "输入数字" : kind === "binary" ? "Base64 内容" : "输入筛选值"}
                                            onChange={(content: any) => update(index, {value: content})}/>}</div>
                                <Button className="sql-filter-remove" color="default" variant="filled" disabled={disabled} aria-label={`移除筛选 ${index + 1}`} icon={<Icon type="CloseOutlined"/>}
                                    onClick={() => onChange(value.filter((_: any, position: number) => position !== index))}/>
                            </div>;
                        })}
                    </div>
                </> : <div className="sql-filter-empty">添加条件，筛选需要的数据</div>}
            </div>
        </ConfigProvider>
        <div className="sql-filter-actions">
            <ConfigProvider wave={{disabled: true}}>
                <Button block type="dashed" icon={<Icon type="PlusOutlined"/>} disabled={disabled || !columns.length || value.length >= 64}
                    onClick={() => onChange([...value, {column: columns[0]?.name || "", operator: "eq", value: ""}])}>添加条件</Button>
            </ConfigProvider>
        </div>
    </div>;
}
