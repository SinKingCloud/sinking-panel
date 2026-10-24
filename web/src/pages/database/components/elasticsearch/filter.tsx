import {useLayoutEffect, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, DatePicker, Form, Input, InputNumber, Segmented, Select} from "antd";
import dayjs from "dayjs";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import type {DatabaseVersion} from "../shared/entities";
import {useDatabaseClient} from "../shared/context";
import {copyTextToClipboard} from "../shared/clipboard";
import useModalStyles from "./styles";
import useStyles from "./filter-styles";
import {buildFilterQuery, getFilterFields, getFilterOperators, parseFilterQuery} from "./filter-query";

export default ({open, query, mapping, fields, version, onCancel, onApply}: {
    open: boolean; query: string; mapping: Record<string, any>; fields: string[]; version?: DatabaseVersion;
    onCancel: () => void; onApply: (query: string) => void;
}) => {
    const {message, modal} = App.useApp();
    const {Editor: AceEditor} = useDatabaseClient();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const {styles: modalStyles} = useModalStyles({compact});
    const candidates = useMemo(() => getFilterFields(mapping, fields), [mapping, fields]);
    const selectable = useMemo(() => candidates.filter((item) => !item.ambiguous), [candidates]);
    const candidateMap = useMemo(() => new Map(candidates.map((item) => [item.name, item])), [candidates]);
    const options = useMemo(() => ({legacy: version?.features?.exists_query === false
        || version?.features?.exists_query === undefined && Boolean(version?.version) && Number(version?.version?.split(".")[0]) < 2}), [version]);
    const [mode, setMode] = useState("visual");
    const [visual, setVisual] = useState<any>({logic: "all", conditions: []});
    const [dsl, setDsl] = useState("{}");
    const [preview, setPreview] = useState(false);
    const [changed, setChanged] = useState(false);
    const [fallback, setFallback] = useState(false);
    const confirmation = useRef<any>(undefined);
    const sequence = useRef(0);

    // 字段映射异步返回时只更新选项，不覆盖用户已经填写的条件。
    useLayoutEffect(() => {
        sequence.current++;
        if (!open) return;
        const source = query.trim() || "{}";
        const parsed = parseFilterQuery(source, candidates, options);
        setDsl(source);
        setVisual(parsed || {logic: "all", conditions: []});
        setMode(parsed ? "visual" : "dsl");
        setPreview(false);
        setChanged(false);
        return () => { sequence.current++; confirmation.current?.destroy(); };
    }, [open, query]);

    const output = useMemo(() => {
        try {
            if (mode === "visual" && changed) return {text: JSON.stringify(buildFilterQuery(visual, candidates, options), null, 2), error: ""};
            let value: any;
            try { value = JSON.parse(dsl); } catch { throw new Error("请输入有效的 JSON 查询对象"); }
            if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("请输入有效的 JSON 查询对象");
            return {text: dsl, error: ""};
        } catch (reason: any) { return {text: "", error: reason.message || "请补全筛选条件"}; }
    }, [mode, visual, candidates, options, dsl, changed]);

    const update = (index: number, values: any) => {
        setChanged(true);
        setVisual((current: any) => ({...current, conditions: current.conditions.map((item: any, position: number) => position === index ? {...item, ...values} : item)}));
    };
    const switchMode = (next: string) => {
        if (next === mode) return;
        if (next === "dsl") {
            if (output.error) { message.error(output.error); return; }
            setDsl(changed ? output.text : dsl);
            setMode(next);
            return;
        }
        const parsed = parseFilterQuery(dsl, candidates, options);
        if (parsed) { setVisual(parsed); setMode(next); setChanged(false); return; }
        const current = sequence.current;
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({title: "切换到可视化筛选？", content: "当前 DSL 包含可视化暂不支持的内容，切换将重置当前编辑的条件。", okText: "重置并切换", cancelText: "保留 DSL", zIndex: 1160,
            onOk: () => { if (current !== sequence.current) return; setVisual({logic: "all", conditions: []}); setChanged(true); setMode("visual"); },
        });
    };
    const submit = () => {
        if (output.error) { setPreview(true); message.error(output.error); return; }
        onApply(mode === "visual" && !changed ? dsl : output.text);
    };

    return <ProModal title={<Title>筛选文档</Title>} onCancel={onCancel} onOk={submit} modalProps={{
        open, width: compact ? 520 : 560, rootClassName: modalStyles.modal, zIndex: 1140,
        styles: {wrapper: {paddingInline: 12}, body: {maxHeight: "none", overflow: "visible"}},
        style: {maxWidth: "calc(100vw - 24px)"}, okText: "确认", cancelText: "取消", mask: {closable: true},
        focusable: {focusTriggerAfterClose: false},
    }}>
        <Form layout="vertical" variant="filled" className={styles.form} onFinish={submit}>
            <div className="es-filter-heading">
                <Segmented value={mode} options={[{value: "visual", label: "可视化"}, {value: "dsl", label: "DSL"}]} onChange={(value) => switchMode(String(value))}/>
                <Button color="default" variant="filled" onClick={() => { setVisual({logic: "all", conditions: []}); setDsl("{}"); setChanged(true); }}>重置</Button>
            </div>
            {mode === "visual" ? <>
                <Select className="es-filter-logic" value={visual.logic} aria-label="条件组合方式" placeholder="请选择条件组合方式" variant="filled"
                    options={[{value: "all", label: "同时满足以下条件"}, {value: "any", label: "满足以下任一条件"}]}
                    onChange={(logic) => { setChanged(true); setVisual((current: any) => ({...current, logic})); }}/>
                <ConfigProvider theme={{token: {fontSize: compact ? 12 : 13, controlHeight: compact ? 28 : 32}}}>
                    {visual.conditions.length > 0 && <div className="es-filter-labels" aria-hidden="true"><span>字段</span><span>条件</span><span>筛选值</span></div>}
                    {visual.conditions.map((condition: any, index: number) => {
                        const selected = candidateMap.get(condition.field);
                        const operators = getFilterOperators(selected);
                        const input = (name: "value" | "valueTo") => {
                            const value = condition[name];
                            const label = `${name === "valueTo" ? "结束值" : "筛选值"} ${index + 1}`;
                            const placeholder = condition.operator === "between" ? name === "value" ? "起始值" : "结束值" : "请输入筛选值";
                            if (selected?.type === "boolean") return <Select variant="filled" value={value === "" || value == null ? undefined : String(value)} aria-label={label} placeholder="请选择真假"
                                options={[{value: "true", label: "是（true）"}, {value: "false", label: "否（false）"}]} onChange={(next) => update(index, {[name]: next === "true"})}/>;
                            if (["byte", "short", "integer", "long", "unsigned_long", "half_float", "float", "double", "scaled_float", "token_count"].includes(selected?.type || "")) return <InputNumber variant="filled" stringMode value={value ?? null} aria-label={label} placeholder={placeholder}
                                controls={false} onChange={(next) => update(index, {[name]: next ?? ""})}/>;
                            if (["date", "date_nanos"].includes(selected?.type || "")) return <div className="es-filter-date">
                                <Input variant="filled" value={value ?? ""} aria-label={label} placeholder={condition.operator === "between" ? placeholder : "日期或 now-1d"} onChange={(event) => update(index, {[name]: event.target.value})}/>
                                <DatePicker variant="filled" showTime allowClear={false} inputReadOnly aria-label={`选择日期 ${label}`} placeholder="选择日期"
                                    classNames={{popup: {root: styles.datePopup}}}
                                    styles={{input: {width: 0, minWidth: 0, flex: "none", padding: 0, opacity: 0}, suffix: {marginInlineStart: 0, flex: 1, justifyContent: "center"}}}
                                    value={value && dayjs(value).isValid() ? dayjs(value) : null}
                                    onChange={(next) => { if (next) update(index, {[name]: next.toISOString()}); }}/>
                            </div>;
                            return <Input variant="filled" value={value ?? ""} aria-label={label} placeholder={placeholder} onChange={(event) => update(index, {[name]: event.target.value})}/>;
                        };
                        return <div className="es-filter-row" key={index}>
                            {selectable.length ? <Select variant="filled" showSearch optionFilterProp="label" value={condition.field || undefined} aria-label={`筛选字段 ${index + 1}`} placeholder="选择字段"
                                popupMatchSelectWidth={false} styles={{popup: {root: {maxWidth: "min(460px, calc(100vw - 24px))"}}}}
                                options={selectable.map((item) => ({value: item.name, label: item.name}))}
                                onChange={(field) => { const available = getFilterOperators(candidateMap.get(field)); update(index, {field, operator: available[0]?.value || "match", value: "", valueTo: ""}); }}/>
                                : <Input variant="filled" value={condition.field || ""} aria-label={`筛选字段 ${index + 1}`} placeholder="请输入字段名称"
                                    onChange={(event) => update(index, {field: event.target.value})}/>}
                            <Select variant="filled" value={condition.operator || undefined} aria-label={`筛选条件 ${index + 1}`} placeholder="选择条件" options={operators}
                                onChange={(operator) => update(index, {operator})}/>
                            <div className="es-filter-value">{["exists", "missing"].includes(condition.operator) ? <Input variant="filled" disabled placeholder="无需填写值"/>
                                : <>{input("value")}{condition.operator === "between" && input("valueTo")}</>}</div>
                            <Button className="es-filter-remove" color="default" variant="filled" aria-label={`删除条件 ${index + 1}`} icon={<Icon type="CloseOutlined"/>}
                                onClick={() => { setChanged(true); setVisual((current: any) => ({...current, conditions: current.conditions.filter((_: any, position: number) => position !== index)})); }}/>
                        </div>;
                    })}
                </ConfigProvider>
                {!visual.conditions.length && <div className="es-filter-empty">添加条件，筛选需要的文档</div>}
                <ConfigProvider wave={{disabled: true}}><Button block type="dashed" icon={<Icon type="PlusOutlined"/>} disabled={visual.conditions.length >= 64}
                    onClick={() => { const first = selectable[0]; setChanged(true); setVisual((current: any) => ({...current, conditions: [...current.conditions, {field: first?.name || "", operator: getFilterOperators(first)[0]?.value || "match", value: ""}]})); }}>添加条件</Button></ConfigProvider>
                {!selectable.length && <div className="es-filter-hint">请输入字段名称，或切换 DSL 编写查询。</div>}
            </> : <Form.Item className="es-filter-dsl" label={<div className="es-filter-dsl-heading"><span>查询 DSL</span>
                <Button size="small" type="text" onClick={() => {
                    if (output.error) { message.error(output.error); return; }
                    setDsl(JSON.stringify(JSON.parse(dsl), null, 2));
                    setChanged(true);
                }}>美化</Button>
            </div>}>
                <div className="es-filter-dsl-editor">
                    {AceEditor && !fallback ? <AceEditor mode="json" value={dsl} onChange={(value) => { setDsl(value); setChanged(true); }}
                        height="100%" fontSize={compact ? 12 : 13} tabSize={2} theme={appearance?.isDarkTheme?.() ? "tomorrow_night" : "textmate"}
                        placeholder={'例如 {"query":{"match":{"title":"关键词"}}}'} showPrintMargin={false} wrapEnabled enableSnippets={false} onError={() => setFallback(true)}/>
                        : <Input.TextArea className="es-filter-dsl-fallback" value={dsl} onChange={(event) => { setDsl(event.target.value); setChanged(true); }} spellCheck={false} placeholder={'例如 {"query":{"match":{"title":"关键词"}}}'}/>}
                </div>
            </Form.Item>}
            {mode === "visual" && <div className="es-filter-preview">
                <div className="es-filter-preview-heading">
                    <button type="button" className="es-filter-preview-toggle" aria-expanded={preview} onClick={() => setPreview((current) => !current)}>预览 DSL</button>
                    <Button size="small" type="text" disabled={Boolean(output.error)} onClick={async () => {
                        try { if (await copyTextToClipboard(output.text)) message.success("复制成功"); else message.error("复制失败"); }
                        catch { message.error("复制失败"); }
                    }}>复制</Button>
                </div>
                {preview && (output.error ? <div className="es-filter-preview-error" role="status">{output.error}</div> : <pre>{output.text}</pre>)}
            </div>}
        </Form>
    </ProModal>;
};
