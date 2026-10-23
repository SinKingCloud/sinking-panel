import {App, AutoComplete, Button, Collapse, ConfigProvider, DatePicker, Input, InputNumber, Segmented, Select} from "antd";
import {useMemo, useState} from "react";
import dayjs from "dayjs";
import {ProModal, Title, useTheme} from "sinking-antd";
import {CloseOutlined, DownOutlined, FilterOutlined, OrderedListOutlined, PlusOutlined, ReloadOutlined} from "@ant-design/icons";
import {mongoQuery, mongoTypes, mongoValue, newMongoField} from "./field-data";
import useFilterStyles from "./filter-styles";
import useStyles from "../shared/styles/nosql-editor";
import usePanelStyles from "./styles";
import {useDatabaseClient} from "../shared/context";
import {copyTextToClipboard} from "../shared/clipboard";
import {parseMongoShell, shellFind} from "./shell";

let sequence = 0;
export const newMongoOrder = () => ({id: ++sequence, field: "", direction: 1});

export const MongoOrderEditor = ({value = [], onChange, indexes = false, projection = false, disabled = false, fieldNames = []}: any) => {
    const {features = {}} = useDatabaseClient();
    const {styles} = useStyles();
    const indexFeatures: Record<string, string> = {text: "text_indexes", hashed: "hashed_indexes", "2dsphere": "sphere_indexes"};
    const directions = [{value: 1, label: "升序"}, {value: -1, label: "降序"},
        ...(indexes ? [{value: "text", label: "文本搜索"}, {value: "hashed", label: "哈希"}, {value: "2dsphere", label: "地理位置"}, {value: "2d", label: "平面坐标"}].filter((item) => features[indexFeatures[item.value]] !== false) : [])];
    const update = (id: number, change: any) => onChange(value.map((row: any) => row.id === id ? {...row, ...change} : row));
    return <div className={styles.editor}>
        {indexes && value.length > 0 && <div className="mongo-order-heading"><span>字段名称</span>{!projection && <span>索引方式</span>}</div>}
        {value.map((row: any, index: number) => <div className="mongo-order-row" key={row.id}>
            {indexes && <span className="nosql-row-number">{index + 1}</span>}
            <AutoComplete variant="filled" className="nosql-cell" aria-label={`字段 ${index + 1}`} placeholder="选择或填写字段名称" value={row.field}
                options={fieldNames.map((value: string) => ({value}))} filterOption={(input, option) => String(option?.value).toLowerCase().includes(input.toLowerCase())}
                disabled={disabled} onChange={(field) => update(row.id, {field})}/>
            {!projection && <Select variant="filled" className="mongo-order-direction" placeholder={indexes ? "请选择索引方式" : "请选择顺序"} aria-label={`字段 ${index + 1} 排序`} value={row.direction}
                options={directions} disabled={disabled} onChange={(direction) => update(row.id, {direction})}/>}
            <Button type="text" className="nosql-remove" aria-label={`删除字段 ${index + 1}`} disabled={disabled} icon={<CloseOutlined/>}
                onClick={() => onChange(value.filter((item: any) => item.id !== row.id))}/>
        </div>)}
        <Button type={indexes ? "dashed" : undefined} color={indexes ? undefined : "default"} variant={indexes ? undefined : "filled"}
            icon={indexes ? undefined : <PlusOutlined/>} className="nosql-add" disabled={disabled} onClick={() => onChange([...value, newMongoOrder()])}>添加字段</Button>
    </div>;
};

const operators = [
    ["$eq", "等于"], ["$ne", "不等于"], ["contains", "包含文本"], ["starts", "开头是"], ["ends", "结尾是"],
    ["$gt", "大于"], ["$gte", "大于等于"], ["$lt", "小于"], ["$lte", "小于等于"],
    ["$in", "是其中之一"], ["$nin", "不是其中之一"], ["exists", "字段存在"], ["missing", "字段不存在"],
    ["null", "是空值"], ["notNull", "不是空值"], ["$regex", "正则匹配"],
].map(([value, label]) => ({value, label}));
const noValue = ["exists", "missing", "null", "notNull"];
const textOnly = ["contains", "starts", "ends", "$regex"];
const scalarTypes = mongoTypes.filter(({value}) => !["object", "array", "null"].includes(value));
const comparisonTypes = mongoTypes.filter(({value}) => value !== "null");

export const newMongoFilter = () => ({id: ++sequence, field: "", operator: "$eq", node: newMongoField()});
export const emptyMongoQuery = () => ({filters: [] as any[], logic: "all", sort: [] as any[], projection: [] as any[], projectionMode: 1, query: undefined as any, shell: ""});

// 筛选与聚合共用条件编辑器；这里仅编辑草稿，不会发送查询。
export const MongoFilterEditor = ({value, onChange, disabled = false, fieldNames = [], sampleDocuments = []}: any) => {
    const {styles} = useFilterStyles();
    const filters = value.filters || [];
    const fields = useMemo(() => {
        const names = new Set<string>(fieldNames);
        const types = new Map<string, string>();
        // 仅识别当前值，不把样本文档或大数组展开为字段节点树。
        const readType = (item: any): string => {
            if (item == null) return "null";
            if (Array.isArray(item)) return "array";
            if (typeof item === "number") return Number.isInteger(item) && item >= -2147483648 && item <= 2147483647 ? "int32" : "double";
            if (typeof item !== "object") return typeof item === "boolean" ? "boolean" : "string";
            const keys = Object.keys(item);
            if (keys.length === 1) {
                const key = keys[0];
                const bsonType = ({$oid: "objectId", $numberInt: "int32", $numberLong: "int64", $numberDouble: "double", $numberDecimal: "decimal"} as any)[key];
                if (bsonType && typeof item[key] === "string") return bsonType;
                if (key === "$date") {
                    const date = new Date(typeof item.$date === "object" ? Number(item.$date?.$numberLong) : item.$date);
                    if (!Number.isNaN(date.getTime())) return "date";
                }
            }
            if (keys.some((key) => ["$binary", "$regularExpression", "$timestamp", "$minKey", "$maxKey", "$undefined", "$code", "$dbPointer", "$symbol", "$date"].includes(key))) return "raw";
            return "object";
        };
        const visit = (document: any, prefix = "", depth = 0) => {
            if (!document || typeof document !== "object" || Array.isArray(document) || depth > 4) return;
            for (const [key, item] of Object.entries(document)) {
                const name = prefix ? `${prefix}.${key}` : key;
                names.add(name);
                const type = readType(item);
                if (!["null", "array", "object", "raw"].includes(type) && !types.has(name)) types.set(name, type);
                if (type === "array" && !types.has(name)) {
                    for (const element of (item as any[]).slice(0, 8)) {
                        const elementType = readType(element);
                        if (!["null", "array", "object", "raw"].includes(elementType)) { types.set(name, elementType); break; }
                    }
                }
                if (type === "object") visit(item, name, depth + 1);
            }
        };
        sampleDocuments.slice(0, 50).forEach((document: any) => visit(document));
        return {options: Array.from(names, (name) => ({value: name})), types};
    }, [fieldNames, sampleDocuments]);
    const update = (id: number, change: any) => onChange({...value, filters: filters.map((row: any) => row.id === id ? {...row, ...change} : row)});
    return <div className={styles.editor}>
        <div className="mongo-filter-toolbar">
            <Select variant="filled" aria-label="条件关系" value={value.logic || "all"} disabled={disabled}
                options={[{value: "all", label: "同时满足以下条件"}, {value: "any", label: "满足以下任一条件"}]}
                onChange={(logic) => onChange({...value, logic})}/>
            {filters.length > 0 && <Button color="default" variant="filled" disabled={disabled} onClick={() => onChange({...value, filters: []})}>清空</Button>}
        </div>
        <div className="mongo-filter-rows">
            {filters.map((row: any, index: number) => {
                const list = ["$in", "$nin"].includes(row.operator);
                const hidden = noValue.includes(row.operator);
                const node = row.node || newMongoField();
                const type = list ? node.itemType || node.children?.[0]?.type || "string" : node.type;
                const changeType = (nextType: string) => update(row.id, {node: list ? {...newMongoField("array"), itemType: nextType}
                    : {...newMongoField(nextType), value: nextType === "boolean" ? true : ""}});
                let input;
                if (list) input = <Select variant="filled" mode={type === "boolean" ? "multiple" : "tags"} aria-label={`条件 ${index + 1} 的值`} placeholder={type === "boolean" ? "选择一个或多个值" : "输入一个值后按回车，可添加多个"}
                    value={(node.children || []).map((child: any) => String(child.value ?? ""))} open={type === "boolean" ? undefined : false} disabled={disabled}
                    options={type === "boolean" ? [{value: "true", label: "是（true）"}, {value: "false", label: "否（false）"}] : undefined}
                    onChange={(values) => update(row.id, {node: {...node, children: values.map((item: string) => ({...newMongoField(type), value: type === "boolean" ? item === "true" : item}))}})}/>;
                else if (type === "boolean") input = <Select variant="filled" aria-label={`条件 ${index + 1} 的值`} placeholder="请选择是或否" value={String(node.value)} disabled={disabled}
                    options={[{value: "true", label: "是（true）"}, {value: "false", label: "否（false）"}]} onChange={(next) => update(row.id, {node: {...node, value: next === "true"}})}/>;
                else if (type === "date") input = <DatePicker variant="filled" aria-label={`条件 ${index + 1} 的值`} placeholder="选择日期和时间" showTime format="YYYY-MM-DD HH:mm:ss.SSS"
                    value={node.value ? dayjs(node.value) : null} disabled={disabled} onChange={(next) => update(row.id, {node: {...node, value: next?.toISOString() || ""}})}/>;
                else if (["int32", "int64", "double", "decimal"].includes(type)) input = <InputNumber variant="filled" stringMode controls={false}
                    aria-label={`条件 ${index + 1} 的值`} placeholder="输入数字" value={node.value === "" ? null : node.value ?? null} disabled={disabled}
                    onChange={(next) => update(row.id, {node: {...node, value: next == null ? "" : String(next)}})}/>;
                else if (["object", "array", "raw"].includes(type)) input = <Input.TextArea variant="filled" aria-label={`条件 ${index + 1} 的值`} placeholder="输入 JSON 内容"
                    autoSize={{minRows: 2, maxRows: 5}} value={node.json ?? JSON.stringify(mongoValue(node))} disabled={disabled}
                    onChange={(event) => update(row.id, {node: {...node, json: event.target.value}})}/>;
                else input = <Input variant="filled" aria-label={`条件 ${index + 1} 的值`} placeholder={type === "objectId" ? "输入 24 位 ObjectId" : "输入要查找的内容"}
                    value={node.value} disabled={disabled} onChange={(event) => update(row.id, {node: {...node, value: event.target.value}})}/>;
                return <div className="mongo-filter-row" key={row.id}>
                    <div className="mongo-filter-main">
                        <AutoComplete variant="filled" aria-label={`条件 ${index + 1} 的字段`} placeholder="选择或输入字段" value={row.field}
                            options={fields.options} filterOption={(input, option) => String(option?.value).toLowerCase().includes(input.toLowerCase())}
                            disabled={disabled} onChange={(field) => {
                                const inferred = fields.types.get(field);
                                update(row.id, {field, ...(inferred && inferred !== type && !textOnly.includes(row.operator) && !hidden
                                    ? {node: list ? {...newMongoField("array"), itemType: inferred} : {...newMongoField(inferred), value: inferred === "boolean" ? true : ""}} : {})});
                            }}/>
                        <Select variant="filled" aria-label={`条件 ${index + 1} 的比较方式`} placeholder="选择条件" value={row.operator} options={operators} disabled={disabled}
                            onChange={(operator) => {
                                const inferred = fields.types.get(row.field) || "string";
                                update(row.id, {operator, node: ["$in", "$nin"].includes(operator)
                                    ? {...newMongoField("array"), itemType: list ? type : inferred}
                                    : textOnly.includes(operator) ? {...newMongoField(), value: typeof node.value === "string" ? node.value : ""}
                                        : list || noValue.includes(row.operator) ? {...newMongoField(inferred), value: inferred === "boolean" ? true : ""} : node});
                            }}/>
                        <Button color="default" variant="filled" disabled={disabled} aria-label={`删除条件 ${index + 1}`} icon={<CloseOutlined/>}
                            onClick={() => onChange({...value, filters: filters.filter((item: any) => item.id !== row.id)})}/>
                    </div>
                    {!hidden && <div className="mongo-filter-value">
                        <div className="mongo-filter-input">{input}</div>
                        {!textOnly.includes(row.operator) && <Select variant="filled" aria-label={`条件 ${index + 1} 的值类型`} placeholder="值类型" value={type} disabled={disabled}
                            options={[...(list ? scalarTypes : comparisonTypes), ...(!(list ? scalarTypes : comparisonTypes).some((item) => item.value === type) ? [{value: type, label: "JSON"}] : [])]}
                            onChange={changeType}/>}
                    </div>}
                </div>;
            })}
        </div>
        {!filters.length && <div className="mongo-filter-empty">未设置条件，将显示全部文档</div>}
        <Button type="dashed" className="mongo-filter-add" icon={<PlusOutlined/>} disabled={disabled}
            onClick={() => onChange({...value, filters: [...filters, newMongoFilter()]})}>添加条件</Button>
    </div>;
};

export default ({value, appliedCount = 0, onChange, onApply, disabled, collection, databaseName, fieldNames = [], sampleDocuments = []}: any) => {
    const {message, modal} = App.useApp();
    const {Editor} = useDatabaseClient();
    const {styles} = useFilterStyles();
    const [open, setOpen] = useState(false);
    const [draft, setDraft] = useState(emptyMongoQuery);
    const [editing, setEditing] = useState("visual");
    const [source, setSource] = useState("");
    const [preview, setPreview] = useState(false);
    const [fallback, setFallback] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles: panelStyles} = usePanelStyles({compact});
    const optionFontSize = compact ? 11 : 12;
    const optionHeight = compact ? 24 : 28;
    const optionTheme = {
        token: {fontSize: optionFontSize, fontSizeSM: optionFontSize, controlHeight: optionHeight, controlHeightSM: optionHeight},
        components: {
            Input: {inputFontSize: optionFontSize, inputFontSizeSM: optionFontSize, controlHeight: optionHeight, controlHeightSM: optionHeight},
            Select: {optionFontSize: optionFontSize, optionHeight, controlHeight: optionHeight, controlHeightSM: optionHeight},
            Button: {contentFontSize: optionFontSize, contentFontSizeSM: optionFontSize, controlHeight: optionHeight, controlHeightSM: optionHeight},
        },
    };
    const visualQuery = () => mongoQuery(draft.filters, draft.sort, draft.projection, draft.projectionMode, draft.logic);
    const readShell = () => {
        const parsed = parseMongoShell(source);
        if (parsed.kind !== "find" || !parsed.query) throw new Error("这里支持 find 筛选，其他 mongosh 命令请在聚合查询中执行");
        if (parsed.collection !== collection || (parsed.body.database_name && parsed.body.database_name !== databaseName)) throw new Error("请筛选当前数据库中的当前集合");
        if (parsed.query.skip || parsed.query.limit || Object.keys(JSON.parse(parsed.body.command)).some((key) => !["find", "filter", "sort", "projection", "skip", "limit"].includes(key)))
            throw new Error("这里使用表格分页，带分页或其他选项的语句请在聚合查询中执行");
        return {filter: parsed.query.filter, sort: parsed.query.sort, projection: parsed.query.projection};
    };
    const apply = async (next: any) => {
        try {
            if (!next.query) {
                next = {...next, filters: next.filters.filter((row: any) => row.field?.trim()),
                    sort: next.sort.filter((row: any) => row.field?.trim()), projection: next.projection.filter((row: any) => row.field?.trim())};
                mongoQuery(next.filters, next.sort, next.projection, next.projectionMode, next.logic);
            }
            setSubmitting(true);
            if (await onApply(next) === false) return;
            onChange(next);
            setOpen(false);
        } catch (error: any) { message.error(error?.message || "筛选条件不正确"); }
        finally { setSubmitting(false); }
    };
    const changeEditor = (next: string) => {
        if (next === editing) return;
        if (next === "shell") {
            try { setSource(shellFind(collection, visualQuery())); }
            catch (error: any) { message.error(error.message); return; }
        } else {
            try {
                const query = readShell();
                // 两边按同一套 BSON 规则比较，格式化或 ObjectId 大小写变化不应阻止切回。
                const visual = parseMongoShell(shellFind(collection, visualQuery())).query!;
                if (["filter", "sort", "projection"].some((key) => query[key as keyof typeof query] !== visual[key as keyof typeof visual])) {
                    message.info("语句已自定义，请继续使用 mongosh；重置后可重新配置可视化条件");
                    return;
                }
            } catch {
                // 错误语句无法转换时，仍允许返回并继续编辑原可视化条件。
            }
            setDraft({...draft, query: undefined, shell: ""});
        }
        setEditing(next); setPreview(false);
    };
    const blocked = disabled || submitting;
    return <>
        <Button color="default" variant="filled" icon={<FilterOutlined/>} disabled={disabled} onClick={() => {
            const next = JSON.parse(JSON.stringify({...emptyMongoQuery(), ...value}));
            setDraft(next); setEditing(next.query ? "shell" : "visual"); setPreview(false);
            setSource(next.shell || (next.query ? shellFind(collection, next.query) : ""));
            setOpen(true);
        }}>筛选</Button>
        {appliedCount > 0 && <Button color="default" variant="filled" icon={<ReloadOutlined/>} aria-label="重置筛选" disabled={blocked}
            onClick={() => void apply(emptyMongoQuery())}/>}
        <ProModal title={<Title>筛选文档</Title>} onCancel={() => { if (!submitting) setOpen(false); }} onOk={() => {
            try { void apply(editing === "shell" ? {...draft, query: readShell(), shell: source} : {...draft, query: undefined, shell: ""}); }
            catch (error: any) { message.error(error.message); }
        }} okText="确认"
            modalProps={{open, rootClassName: panelStyles.modal, className: styles.modal, width: compact ? 520 : 560, style: {maxWidth: "calc(100vw - 24px)"}, zIndex: 1140,
                cancelText: "取消", destroyOnHidden: true, confirmLoading: submitting, styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}},
                okButtonProps: {disabled}, mask: {closable: !submitting}}}>
            <div className="mongo-filter-mode">
                <Segmented value={editing} disabled={blocked} options={[{value: "visual", label: "可视化"}, {value: "shell", label: "mongosh"}]} onChange={(next) => changeEditor(String(next))}/>
                <div>
                    {editing === "visual" && <Button color="default" variant="filled" disabled={blocked} onClick={() => {
                        try { if (!preview) setSource(shellFind(collection, visualQuery())); setPreview(!preview); }
                        catch (error: any) { message.error(error.message); }
                    }}>{preview ? "返回" : "预览"}</Button>}
                    {(editing === "shell" || preview) && <Button color="default" variant="filled" onClick={() => void copyTextToClipboard(source).then((ok) => ok ? message.success("复制成功") : message.error("复制失败"))}>复制</Button>}
                    {editing === "shell" && <Button color="default" variant="filled" disabled={blocked} onClick={() => modal.confirm({title: "重置筛选条件？", content: "将清空当前语句，重新配置可视化条件。", okText: "重置", cancelText: "取消", zIndex: 1160,
                        onOk: () => { setDraft(emptyMongoQuery()); setEditing("visual"); setPreview(false); setSource(""); }})}>重置</Button>}
                </div>
            </div>
            {editing === "shell" || preview ? <div className="mongo-filter-shell">
                {Editor && !fallback ? <Editor value={source} mode="javascript" height={compact ? "180px" : "200px"} fontSize={compact ? 12 : 13}
                    theme={theme?.isDarkTheme?.() ? "tomorrow_night" : "textmate"} readOnly={preview || blocked} showPrintMargin={false} wrapEnabled
                    placeholder={`db.getCollection(${JSON.stringify(collection || "collection")}).find({})`} onChange={setSource} onError={() => setFallback(true)}/>
                    : <Input.TextArea value={source} readOnly={preview || blocked} spellCheck={false} rows={7} placeholder="db.collection.find({})" onChange={(event) => setSource(event.target.value)}/>}
            </div> : <>
                <MongoFilterEditor value={draft} disabled={blocked} fieldNames={fieldNames} sampleDocuments={sampleDocuments} onChange={setDraft}/>
                <Collapse className="mongo-filter-options" ghost expandIconPlacement="end" expandIcon={({isActive}) => <DownOutlined rotate={isActive ? 180 : 0}/>} items={[{
                    key: "options", label: <span className="mongo-query-options-label"><OrderedListOutlined/>排序与显示字段</span>,
                    children: <ConfigProvider componentSize="small" theme={optionTheme}><div className="mongo-query-options-content" data-compact={compact || undefined}>
                        <div className="mongo-query-option-card"><div className="mongo-query-option-heading">排序方式</div>
                            <MongoOrderEditor value={draft.sort} disabled={blocked} fieldNames={fieldNames} onChange={(sort: any[]) => setDraft({...draft, sort})}/></div>
                        <div className="mongo-query-option-card"><div className="mongo-query-option-heading">显示字段</div>
                            <div className="mongo-projection-settings">
                                <label><span>显示方式</span><Select variant="filled" aria-label="字段显示方式" placeholder="选择显示方式" value={draft.projectionMode} disabled={blocked}
                                    options={[{value: 1, label: "只显示"}, {value: 0, label: "不显示"}]} onChange={(projectionMode) => setDraft({...draft, projectionMode})}/>
                                </label>
                                <label><span>字段名称</span><Select variant="filled" mode="tags" allowClear aria-label="选择显示字段" placeholder="选择或输入字段，留空显示全部" disabled={blocked}
                                    options={fieldNames.map((field: string) => ({value: field}))} value={draft.projection.map((row: any) => row.field)}
                                    onChange={(fields) => setDraft({...draft, projection: fields.map((field: string) => ({...newMongoOrder(), field}))})}/>
                                </label>
                            </div></div>
                    </div></ConfigProvider>,
                }]}/>
            </>}
        </ProModal>
    </>;
};
