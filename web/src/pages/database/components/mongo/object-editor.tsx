import {useEffect, useRef, useState} from "react";
import {App, Button, ConfigProvider, Form, Input, InputNumber, Select, Spin, theme as antdTheme} from "antd";
import {ArrowDownOutlined, ArrowUpOutlined, CloseOutlined, PlusOutlined} from "@ant-design/icons";
import {createStyles} from "antd-style";
import {ProModal, Title, useTheme} from "sinking-antd";
import {useDatabaseClient} from "../shared/context";
import {MongoValueInput} from "./fields-editor";
import {mongoFields, mongoTypes, mongoValue, newMongoField} from "./field-data";
import usePanelStyles from "./styles";

const stages = ["$match", "$project", "$sort", "$limit", "$skip", "$group", "$unwind", "$addFields", "$set", "$unset", "$lookup", "$facet", "$count", "$sample", "$replaceRoot", "$replaceWith", "$sortByCount", "$unionWith"];
const stageFeatures: Record<string, string> = {$lookup: "aggregate_32", $sample: "aggregate_32", $addFields: "aggregate_34", $facet: "aggregate_34", $count: "aggregate_34", $replaceRoot: "aggregate_34", $sortByCount: "aggregate_34", $set: "aggregate_42", $unset: "aggregate_42", $replaceWith: "aggregate_42", $unionWith: "aggregate_44"};
const stageLabels: Record<string, string> = {$match: "筛选文档", $project: "选择字段", $sort: "排序", $limit: "限制数量", $skip: "跳过数量",
    $group: "分组统计", $unwind: "展开数组", $addFields: "新增字段", $set: "设置字段", $unset: "移除字段", $lookup: "关联集合",
    $facet: "多路处理", $count: "统计数量", $sample: "随机抽样", $replaceRoot: "替换根内容", $replaceWith: "替换内容", $sortByCount: "分组排序", $unionWith: "合并集合"};
const stageType = (name: string) => ["$limit", "$skip"].includes(name) ? "int64"
    : ["$unwind", "$unset", "$count", "$sortByCount", "$unionWith"].includes(name) ? "string" : "object";

// 每行表示一个管道阶段，数值与 BSON 类型沿用文档编辑器的转换逻辑。
export function mongoViewPipeline(values: any[] = [], features: Record<string, boolean> = {}) {
    if (!Array.isArray(values) || values.length > 100) throw new Error("聚合管道最多配置 100 个阶段");
    return values.map((node, index) => {
        const name = String(node.key || "").trim();
        if (!/^\$[^\s\0]+$/.test(name)) throw new Error(`阶段 ${index + 1}：请输入以 $ 开头的阶段名称`);
        if (features[stageFeatures[name]] === false) throw new Error(`当前版本不支持 ${name} 阶段`);
        if (["$out", "$merge"].includes(name)) throw new Error("视图管道不支持写入集合的阶段");
        return {[name]: mongoValue({...node, key: name})};
    });
}

const useStyles = createStyles(({css, token}) => ({
    pipeline: css`
        display: flex; flex-direction: column; gap: 12px; min-width: 0;
        .mongo-stage { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; background: ${token.colorBgContainer}; overflow: hidden; }
        .mongo-stage-header { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 7px 12px; background: ${token.colorFillQuaternary}; }
        .mongo-stage-label { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .mongo-stage-actions { display: flex; flex: none; gap: 2px; }
        .mongo-stage-actions .ant-btn { width: 26px; height: 26px; padding: 0; color: ${token.colorTextTertiary}; }
        .mongo-stage-actions .ant-btn-dangerous:not(:disabled) { color: ${token.colorError}; }
        .mongo-stage-body { display: flex; flex-direction: column; gap: 12px; padding: 12px; min-width: 0; }
        .mongo-stage-field { display: flex; flex-direction: column; gap: 6px; min-width: 0; }
        .mongo-stage-field > span { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .mongo-stage-field > .ant-select, .mongo-stage-field > .ant-input-number { width: 100%; }
        .mongo-stage-field > .nosql-raw { margin: 0; padding: 10px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; font: ${token.fontSizeSM}px/1.6 ${token.fontFamilyCode}; white-space: pre-wrap; overflow-wrap: anywhere; }
        .mongo-stage-format { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 12px; }
        .mongo-stage-empty { color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; padding: 0 0 2px; }
        .mongo-stage-add { width: 100%; }
        .mongo-stage-sort { display: flex; flex-direction: column; gap: 8px; }
        .mongo-stage-sort-row { display: grid; grid-template-columns: minmax(0, 1fr) 104px 26px; gap: 8px; align-items: center; }
        .mongo-stage-sort-row > * { min-width: 0; }
        .mongo-stage-sort-row > .ant-btn { width: 26px; padding: 0; color: ${token.colorTextTertiary}; }
        @media (max-width: 575px) {
            .mongo-stage-header { padding-inline: 10px; }
            .mongo-stage-body { padding: 10px; }
            .mongo-stage-format { grid-template-columns: minmax(0, 1fr); }
            .mongo-stage-sort-row { grid-template-columns: minmax(0, 1fr) 80px 24px; gap: 6px; }
        }
    `,
}));

export const MongoPipelineEditor = ({value = [], onChange, disabled}: any) => {
    const {features = {}} = useDatabaseClient();
    const {modal} = App.useApp();
    const {styles} = useStyles();
    const confirmation = useRef<ReturnType<typeof modal.confirm> | undefined>(undefined);
    useEffect(() => () => { confirmation.current?.destroy(); }, []);
    const update = (index: number, next: any) => onChange?.(value.map((node: any, position: number) => index === position ? next : node));
    const move = (index: number, offset: number) => {
        const next = [...value];
        [next[index], next[index + offset]] = [next[index + offset], next[index]];
        onChange?.(next);
    };
    const replace = (index: number, key: string, type: string, custom = false) => {
        const node = value[index];
        const next = {...newMongoField(type, key), id: node.id, custom};
        if (key === "$limit") next.value = "20";
        if (key === "$count") next.value = "count";
        const apply = () => update(index, next);
        // 更换处理方式时先确认，已有表达式和 BSON 内容不会被悄悄重置。
        if (node.existing || node.children?.length || (node.value !== "" && node.value !== "0" && node.value !== false)) {
            if (confirmation.current) return;
            confirmation.current = modal.confirm({title: "更换处理方式？", content: "当前步骤的内容将清空，其他步骤保持不变。", zIndex: 1160,
                okText: "确认", cancelText: "取消", onOk: apply, afterClose: () => { confirmation.current = undefined; }});
        } else apply();
    };
    return <div className={styles.pipeline}>
        {value.map((node: any, index: number) => {
            const custom = node.custom || !stages.includes(node.key);
            const formats: Record<string, {value: string; label: string}[]> = {
                $unwind: [{value: "string", label: "指定数组字段"}, {value: "object", label: "详细展开选项"}],
                $unset: [{value: "string", label: "单个字段"}, {value: "array", label: "多个字段"}],
                $unionWith: [{value: "string", label: "指定集合"}, {value: "object", label: "集合与处理步骤"}],
                $sortByCount: [{value: "string", label: "按字段统计"}, {value: "object", label: "按表达式统计"}],
                $replaceWith: [{value: "string", label: "引用字段"}, {value: "object", label: "文档表达式"}, {value: "array", label: "数组表达式"}],
            };
            const options = custom ? mongoTypes : formats[node.key];
            const choices = options && !options.some(({value}) => value === node.type) ? [...options, {value: node.type, label: "保留当前内容"}] : options;
            const sort = node.key === "$sort" && node.type === "object" && (node.children || []).every((item: any) => ["int32", "int64", "double"].includes(item.type) && ["1", "-1"].includes(String(item.value)));
            const numeric = ["$limit", "$skip"].includes(node.key) && ["int32", "int64"].includes(node.type);
            const text = node.type === "string" && ["$count", "$unwind", "$unset", "$sortByCount", "$unionWith", "$replaceWith"].includes(node.key);
            const label = ({$match: "匹配字段与条件", $project: "返回字段与表达式", $sort: "排序字段", $limit: "最多返回多少条", $skip: "跳过前多少条",
                $group: "分组字段与统计项", $count: "统计结果字段名", $unwind: "数组字段", $unset: "需要移除的字段", $unionWith: "集合名称", $sortByCount: "分组字段", $replaceWith: "替换内容"} as Record<string, string>)[node.key] || "处理内容";
            return <div className="mongo-stage" key={node.id}>
                <div className="mongo-stage-header"><span className="mongo-stage-label">步骤 {index + 1}</span>
                    {!disabled && <div className="mongo-stage-actions">
                        <Button type="text" aria-label={`上移步骤 ${index + 1}`} disabled={index === 0} icon={<ArrowUpOutlined/>} onClick={() => move(index, -1)}/>
                        <Button type="text" aria-label={`下移步骤 ${index + 1}`} disabled={index === value.length - 1} icon={<ArrowDownOutlined/>} onClick={() => move(index, 1)}/>
                        <Button type="text" danger aria-label={`删除步骤 ${index + 1}`} icon={<CloseOutlined/>} onClick={() => onChange?.(value.filter((_: any, position: number) => position !== index))}/>
                    </div>}
                </div>
                <div className="mongo-stage-body">
                    <div className={choices ? "mongo-stage-format" : undefined}>
                        <label className="mongo-stage-field"><span>处理方式</span><Select variant="filled" value={custom ? "custom" : node.key} disabled={disabled} showSearch optionFilterProp="label" placeholder="请选择处理方式"
                            options={[...stages.filter((name) => features[stageFeatures[name]] !== false || name === node.key).map((name) => ({value: name, label: stageLabels[name]})), {value: "custom", label: "自定义步骤"}]}
                            onChange={(key) => { if (key !== node.key && !(key === "custom" && custom)) replace(index, key === "custom" ? "" : key, key === "custom" ? "object" : stageType(key), key === "custom"); }}/></label>
                        {choices && <label className="mongo-stage-field"><span>内容形式</span><Select variant="filled" value={node.type} disabled={disabled || node.type === "raw"} placeholder="请选择内容形式" options={choices}
                            onChange={(type) => { if (type !== node.type) replace(index, node.key, type, custom); }}/></label>}
                    </div>
                    {custom && <label className="mongo-stage-field"><span>步骤名称</span><Input variant="filled" value={node.key} disabled={disabled} placeholder="例如 $match" onChange={(event) => update(index, {...node, key: event.target.value, custom: true})}/></label>}
                    <div className="mongo-stage-field"><span>{label}</span>
                        {sort ? <div className="mongo-stage-sort">{(node.children || []).map((field: any, position: number) => <div className="mongo-stage-sort-row" key={field.id}>
                            <Input variant="filled" value={field.key} disabled={disabled} aria-label={`排序字段 ${position + 1}`} placeholder="请输入字段名称" onChange={(event) => update(index, {...node, children: node.children.map((item: any) => item.id === field.id ? {...item, key: event.target.value} : item)})}/>
                            <Select variant="filled" value={String(field.value)} disabled={disabled} aria-label={`排序字段 ${position + 1} 顺序`} placeholder="选择顺序" options={[{value: "1", label: "升序"}, {value: "-1", label: "降序"}]}
                                onChange={(direction) => update(index, {...node, children: node.children.map((item: any) => item.id === field.id ? {...item, value: direction} : item)})}/>
                            <Button type="text" disabled={disabled} aria-label={`删除排序字段 ${position + 1}`} icon={<CloseOutlined/>} onClick={() => update(index, {...node, children: node.children.filter((item: any) => item.id !== field.id)})}/>
                        </div>)}{!disabled && <Button type="dashed" icon={<PlusOutlined/>} onClick={() => update(index, {...node, children: [...(node.children || []), {...newMongoField("int32"), value: "1"}]})}>添加排序字段</Button>}</div>
                            : numeric ? <InputNumber variant="filled" stringMode precision={0} value={node.value || undefined} min={node.key === "$limit" ? "1" : "0"} max={node.type === "int32" ? "2147483647" : "9223372036854775807"} disabled={disabled} placeholder={node.key === "$limit" ? "例如 20" : "例如 0"} onChange={(next) => update(index, {...node, value: next == null ? "" : String(next)})}/>
                                : text ? <Input variant="filled" value={node.value} disabled={disabled} placeholder={node.key === "$count" ? "例如 count" : ["$unwind", "$sortByCount", "$replaceWith"].includes(node.key) ? "例如 $items" : node.key === "$unionWith" ? "请输入集合名称" : "请输入字段名称"} onChange={(event) => update(index, {...node, value: event.target.value})}/>
                                    : <MongoValueInput node={node} disabled={disabled} onChange={(next: any) => update(index, next)}/>}</div>
                </div>
            </div>;
        })}
        {!value.length && <div className="mongo-stage-empty">不添加步骤时，保留源集合全部内容。</div>}
        {!disabled && <Button className="mongo-stage-add" type="dashed" icon={<PlusOutlined/>} disabled={value.length >= 100}
            onClick={() => onChange?.([...value, newMongoField("object", "$match")])}>添加步骤</Button>}
    </div>;
};

export default ({record, collections = [], action, onClose, onChange}: any) => {
    const {operate, features = {}} = useDatabaseClient();
    const {message} = App.useApp();
    const theme = useTheme();
    const {token} = antdTheme.useToken();
    const {styles: panelStyles} = usePanelStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const [form] = Form.useForm();
    const live = useRef(true);
    const sequence = useRef(0);
    const pending = useRef(false);
    const [reading, setReading] = useState(false);
    const [saving, setSaving] = useState(false);
    const lastAction = useRef(action);
    const displayAction = action || lastAction.current;
    const readonly = displayAction?.operation === "view_definition";
    const supported = features.views !== false;
    useEffect(() => { live.current = true; return () => { live.current = false; sequence.current++; }; }, []);
    useEffect(() => {
        const request = ++sequence.current;
        if (!action) return;
        lastAction.current = action;
        if (!supported) { setReading(false); return; }
        form.resetFields();
        form.setFieldsValue({name: action.operation === "create_view" ? "" : action.collection, source: action.source, pipeline: []});
        if (action.operation === "create_view") { setReading(false); return; }
        setReading(true);
        void operate(record, "view_definition", {collection: action.collection}).then((data) => {
                if (!live.current || sequence.current !== request) return;
                if (!data || typeof data?.source !== "string" || !Array.isArray(data?.pipeline)) throw new Error("视图定义格式不正确");
                const pipeline = data?.pipeline?.map((stage: any, index: number) => {
                    if (!stage || Array.isArray(stage) || typeof stage !== "object" || Object.keys(stage).length !== 1) throw new Error(`阶段 ${index + 1} 格式不正确`);
                    return mongoFields(stage)[0];
                });
                form.setFieldsValue({name: data?.name || action.collection, source: data?.source, pipeline});
            }).catch((reason) => { if (live.current && sequence.current === request) { message.error(reason?.message || "读取视图失败"); onClose?.(); } })
            .finally(() => { if (live.current && sequence.current === request) setReading(false); });
    }, [record.id, action, supported]);
    const close = () => { if (!pending.current) { sequence.current++; onClose?.(); } };
    const save = async (values: any) => {
        if (!action || !supported || pending.current || reading || readonly) return;
        let pipeline: string;
        try { pipeline = JSON.stringify(mongoViewPipeline(values.pipeline, features)); } catch (reason: any) { message.error(reason?.message || "操作失败"); return; }
        const current = action;
        const request = sequence.current;
        const name = current.operation === "create_view" ? values.name.trim() : current.collection;
        if (current.operation === "create_view" && name === values.source) { message.error("视图名称不能与源集合相同"); return; }
        pending.current = true; setSaving(true);
        try {
            await operate(record, current.operation, {collection: name, pipeline, ...(current.operation === "create_view" ? {source: values.source} : {})});
            if (live.current) { message.success("保存成功"); if (sequence.current === request) onClose?.(); onChange?.(name); }
        } catch (reason: any) { if (live.current && sequence.current === request) message.error(reason?.message || "操作失败"); }
        finally { pending.current = false; if (live.current) setSaving(false); }
    };
    return <ProModal title={<Title>{readonly ? "视图定义" : displayAction?.operation === "update_view" ? "编辑视图" : "新建视图"}</Title>}
        onCancel={close} onOk={() => readonly ? close() : form.submit()} okText={readonly ? "确认" : "保存"}
        modalProps={{open: Boolean(action) && supported, rootClassName: panelStyles.modal,
            width: readonly ? (theme?.isCompactTheme?.() ? 480 : 520) : displayAction?.operation === "create_view" ? (theme?.isCompactTheme?.() ? 500 : 560) : (theme?.isCompactTheme?.() ? 540 : 600), zIndex: 1140,
            styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}}, style: {maxWidth: "calc(100vw - 24px)"}, cancelText: "取消", confirmLoading: saving, closable: !saving, keyboard: !saving, mask: {closable: !saving},
            okButtonProps: {disabled: reading}, cancelButtonProps: {disabled: saving, style: readonly ? {display: "none"} : undefined}, destroyOnHidden: true, forceRender: true}}>
        <Spin spinning={reading}><Form form={form} layout="vertical" variant="filled" disabled={saving || readonly} onFinish={save}>
                <ConfigProvider componentSize="middle" theme={{components: {Input: {controlHeight: token.controlHeight}, InputNumber: {controlHeight: token.controlHeight}, Select: {controlHeight: token.controlHeight}}}}>
                <Form.Item name="name" label="视图名称" rules={[{required: true, whitespace: true, max: 255, message: "请输入视图名称，最多 255 个字符"}]}><Input placeholder="请输入视图名称" disabled={saving || readonly || displayAction?.operation !== "create_view"}/></Form.Item>
                <Form.Item name="source" label="源集合" rules={[{required: true, message: "请选择源集合"}]}><Select placeholder="请选择源集合" showSearch optionFilterProp="label"
                    disabled={saving || readonly || reading || displayAction?.operation !== "create_view"} options={collections.map((item: any) => ({value: item.name, label: item.name}))}/></Form.Item>
                <Form.Item name="pipeline" label="处理步骤" rules={[{validator: async (_, values) => { mongoViewPipeline(values || [], features); }}]}>
                    <MongoPipelineEditor disabled={saving || readonly || reading}/>
                </Form.Item>
                </ConfigProvider>
            </Form></Spin>
    </ProModal>;
};
