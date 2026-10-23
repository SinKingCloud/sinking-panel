import {Button, DatePicker, Input, Select} from "antd";
import {CloseOutlined, DownOutlined, PlusOutlined, RightOutlined} from "@ant-design/icons";
import {useState} from "react";
import dayjs from "dayjs";
import {mongoTypes, newMongoField} from "./field-data";
import useStyles from "./fields-editor-styles";

export const MongoValueInput = ({node, onChange, disabled = false, readOnly = false}: any) => {
    const {styles} = useStyles();
    const label = node.key ? `字段 ${node.key} 的内容` : "字段内容";
    let content;
    if (node.type === "object" || node.type === "array") return <MongoFieldsEditor value={node.children} array={node.type === "array"}
        disabled={disabled} readOnly={readOnly} onChange={(children: any[]) => onChange?.({...node, children})}/>;
    if (node.type === "null") content = <div className="mongo-document-null">空值 <span>NULL</span></div>;
    else if (node.type === "raw") content = <pre className="mongo-document-raw">{JSON.stringify(node.value, null, 2)}</pre>;
    else if (readOnly) content = <div className="mongo-document-read-value">{node.type === "boolean" ? node.value ? "是（true）" : "否（false）"
        : node.type === "date" ? dayjs(node.value).format("YYYY-MM-DD HH:mm:ss.SSS") : node.value === "" ? <span className="mongo-document-empty-value">空文本</span> : String(node.value ?? "")}</div>;
    else if (node.type === "boolean") content = <Select variant="filled" aria-label={label} placeholder="请选择是或否" value={String(node.value)} disabled={disabled}
        options={[{value: "true", label: "是（true）"}, {value: "false", label: "否（false）"}]} onChange={(value) => onChange?.({...node, value: value === "true"})}/>;
    else if (node.type === "date") content = <DatePicker variant="filled" aria-label={label} placeholder="请选择日期和时间" showTime format="YYYY-MM-DD HH:mm:ss.SSS" disabled={disabled}
        value={node.value && dayjs(node.value).isValid() ? dayjs(node.value) : null} onChange={(value) => onChange?.({...node, value: value?.toISOString() || ""})}/>;
    else if (node.type === "string") content = <Input.TextArea variant="filled" aria-label={label} autoSize={{minRows: 1, maxRows: 6}} value={node.value}
        disabled={disabled} placeholder="请输入内容，可留空" onChange={(event) => onChange?.({...node, value: event.target.value})}/>;
    else content = <Input variant="filled" aria-label={label} value={node.value} disabled={disabled}
        inputMode={node.type === "objectId" ? "text" : ["int32", "int64"].includes(node.type) ? "numeric" : "decimal"}
        placeholder={node.type === "objectId" ? "请输入 24 位十六进制标识" : ["int32", "int64"].includes(node.type) ? "请输入整数" : "请输入数字，可包含小数"}
        onChange={(event) => onChange?.({...node, value: event.target.value})}/>;
    return <div className={styles.value}>{content}</div>;
};

const Field = ({node, index, array, disabled, readOnly, update, remove}: any) => {
    const [expanded, setExpanded] = useState(false);
    const locked = disabled || node.locked;
    const nested = node.type === "object" || node.type === "array";
    const typeLabel = node.type === "raw" ? "原始类型" : mongoTypes.find((item) => item.value === node.type)?.label || node.type;
    return <div className={`mongo-document-field${readOnly ? " is-readonly" : ""}`}>
        <div className="mongo-document-field-header">
            <div className="mongo-document-field-name">
                {array ? <span className="mongo-document-element-name">元素 {index + 1}</span> : readOnly ? <span title={node.key}>{node.key || "空字段名"}</span>
                    : <><span className="mongo-document-field-number">{index + 1}</span><Input variant="filled" aria-label={`字段 ${index + 1} 名称`} placeholder="字段名称，例如 name" value={node.key} disabled={locked}
                        onChange={(event) => update({...node, key: event.target.value})}/></>}
            </div>
            {readOnly || node.type === "raw" ? <span className={`mongo-document-field-kind${readOnly ? "" : " mongo-document-field-type-label"}`}>{typeLabel}</span> : <Select variant="filled" className="mongo-document-field-type" placeholder="选择类型" aria-label={`字段 ${index + 1} 类型`} value={node.type} disabled={locked}
                options={mongoTypes}
                onChange={(type) => { update({...newMongoField(type, node.key), id: node.id, existing: node.existing}); setExpanded(type === "object" || type === "array"); }}/>}
            {!readOnly && !locked ? <Button type="text" className="mongo-document-field-remove" aria-label={`删除${array ? "元素" : "字段"} ${index + 1}`} icon={<CloseOutlined/>} onClick={remove}/>
                : !readOnly && <span className="mongo-document-field-lock">{node.locked ? "固定" : ""}</span>}
        </div>
        <div className="mongo-document-field-content">{nested ? <>
            <Button type="text" className="mongo-document-nested-toggle" aria-label={expanded ? "收起字段" : "展开字段"} aria-expanded={expanded}
                icon={expanded ? <DownOutlined/> : <RightOutlined/>} onClick={() => setExpanded(!expanded)}>
                <span>{expanded ? "收起" : "展开"}{node.type === "array" ? "元素" : "字段"}</span><span className="mongo-document-child-count">{node.children?.length || 0} 项</span>
            </Button>
            {expanded && <div className="mongo-document-nested"><MongoValueInput node={node} disabled={locked} readOnly={readOnly || node.locked} onChange={update}/></div>}
        </> : <MongoValueInput node={node} disabled={locked} readOnly={readOnly || node.locked} onChange={update}/>}</div>
    </div>;
};

const MongoFieldsEditor = ({value = [], onChange, disabled = false, readOnly = false, array = false}: any) => {
    const {styles} = useStyles();
    return <div className={styles.editor}>
        {value.map((node: any, index: number) => <Field key={node.id} node={node} index={index} array={array} disabled={disabled} readOnly={readOnly}
            update={(next: any) => onChange?.(value.map((field: any) => field.id === node.id ? next : field))}
            remove={() => onChange?.(value.filter((field: any) => field.id !== node.id))}/>)}
        {!value.length && <div className="mongo-document-editor-empty">{array ? "暂无元素" : "暂无字段"}</div>}
        {!readOnly && <Button className="mongo-document-field-add" type="dashed" icon={<PlusOutlined/>} disabled={disabled} onClick={() => onChange?.([...value, newMongoField()])}>
            {array ? "添加元素" : "添加字段"}</Button>}
    </div>;
};

export default MongoFieldsEditor;
