import {useEffect, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, Empty, Form, Grid, Input, InputNumber, Table, Typography, theme as antdTheme} from "antd";
import {CopyOutlined, PlusOutlined, SearchOutlined} from "@ant-design/icons";
import {ProModal, Title, useTheme} from "sinking-antd";
import {useRedisClient} from "./context";
import {copyTextToClipboard} from "../shared/clipboard";
import useStyles from "./value-styles";

let sequence = 0;
export const redisRows = (kind: string, value: any): any => kind === "string" ? String(value ?? "")
    : (kind === "hash" ? Object.entries(value || {}).map(([key, content]) => ({key, value: content}))
        : (value || []).map((content: any) => kind === "zset" ? {...content} : {value: content}))
        .map((row: any) => ({...row, id: ++sequence}));

export const redisContent = (kind: string, value: any) => {
    if (kind === "string") return value ?? "";
    if (!Array.isArray(value) || !value.length) throw new Error("请至少添加一条内容");
    const seen = new Set<string>();
    if (kind === "hash") {
        const result: Record<string, string> = Object.create(null);
        value.forEach((row: any) => {
            if (seen.has(row.key ?? "")) throw new Error("字段名称不能重复");
            seen.add(row.key ?? ""); result[row.key ?? ""] = row.value ?? "";
        });
        return JSON.stringify(result);
    }
    if (kind === "zset") return JSON.stringify(value.map((row: any) => {
        if (row.score === null || row.score === undefined || row.score === "" || !Number.isFinite(Number(row.score))) throw new Error("请填写有效的分值");
        if (seen.has(row.member ?? "")) throw new Error("成员不能重复");
        seen.add(row.member ?? "");
        return {member: row.member ?? "", score: Number(row.score)};
    }));
    const values = value.map((row: any) => row.value ?? "");
    if (kind === "set" && new Set(values).size !== values.length) throw new Error("集合成员不能重复");
    return JSON.stringify(values);
};

export default ({kind, value, onChange, disabled = false, readOnly = false, parameter = false, start = 0, embedded = false, bodyHeight, fill = false}: any) => {
    const {Editor: AceEditor, ActionDropdown} = useRedisClient();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());
    const mobile = Grid.useBreakpoint().sm === false;
    const {styles} = useStyles({compact});
    const {message} = App.useApp();
    const {token} = antdTheme.useToken();
    const [fallback, setFallback] = useState(false);
    const [keyword, setKeyword] = useState("");
    const [editing, setEditing] = useState<any>();
    const editingRef = useRef<any>();
    const [expanded, setExpanded] = useState<any>();
    const lastExpanded = useRef<any>();
    if (expanded) lastExpanded.current = expanded;
    const [error, setError] = useState("");
    const tableRef = useRef<any>();
    const tableArea = useRef<HTMLDivElement>(null);
    const [tableHeight, setTableHeight] = useState(240);
    const rows = Array.isArray(value) ? value : [];
    const rowsRef = useRef(rows);
    rowsRef.current = rows;
    const indexes = useMemo(() => new Map(rows.map((row: any, index: number) => [row.id, start + index + 1])), [value, start]);
    const filtered = useMemo(() => {
        const search = keyword.trim().toLowerCase();
        return search ? rows.filter((row: any) => row.id === editing?.id || [row.key, row.value, row.member, row.score].some((item) => String(item ?? "").toLowerCase().includes(search))) : rows;
    }, [value, keyword, editing?.id]);
    const update = (id: number, values: any) => onChange?.(rowsRef.current.map((row: any) => row.id === id ? {...row, ...values} : row));
    const copy = async (text: string) => {
        if (await copyTextToClipboard(text)) message.success("复制成功");
        else message.error("复制失败，请手动复制");
    };
    const invalid = (row: any) => kind === "zset" && (row.score === null || row.score === undefined || row.score === "" || !Number.isFinite(Number(row.score)))
        ? "请填写有效的分值"
        : kind === "hash" && rowsRef.current.some((item: any) => item.id !== row.id && (item.key ?? "") === (row.key ?? "")) ? "字段名称不能重复"
            : ["set", "zset"].includes(kind) && rowsRef.current.some((item: any) => item.id !== row.id && (kind === "set" ? item.value ?? "" : item.member ?? "") === (kind === "set" ? row.value ?? "" : row.member ?? "")) ? "成员不能重复" : "";
    const finish = (cancel = false) => {
        const cell = editingRef.current;
        if (!cell) return;
        editingRef.current = undefined; setEditing(undefined);
        const row = rowsRef.current.find((item: any) => item.id === cell.id);
        if (!row) return;
        const reason = !cancel && invalid(row);
        if (cancel || reason) update(cell.id, {[cell.field]: cell.original});
        if (reason) message.error(reason);
    };
    const expand = (row?: any) => {
        finish(); setError("");
        const next = row ? {...row, isNew: false} : {id: ++sequence, key: "", value: "", member: "", score: 0, isNew: true};
        setExpanded(next); lastExpanded.current = next;
    };
    useEffect(() => {
        const node = tableArea.current;
        if (!node) return;
        const measure = () => {
            const height = node.clientHeight;
            if (!height) return;
            const header = node.querySelector<HTMLElement>(".ant-table-thead")?.offsetHeight || (compact ? 32 : 36);
            setTableHeight(Math.max(80, Math.floor(height - header - 2)));
        };
        measure();
        const observer = new ResizeObserver(measure);
        observer.observe(node);
        return () => observer.disconnect();
    }, [kind, fill, embedded, parameter, compact, bodyHeight]);
    const height = bodyHeight ?? (parameter ? 180 : embedded ? 200 : 300);
    if (!embedded && !parameter) {
        const collection = ["hash", "list", "set", "zset"].includes(kind);
        const text = collection ? "" : kind === "string" ? String(value ?? "") : JSON.stringify(value, null, 2) ?? "";
        const writable = !disabled && !readOnly;
        const displayed = expanded || lastExpanded.current;
        const memberField = kind === "zset" ? "member" : "value";
        const cell = (field: string, row: any) => {
            const content = row[field];
            const active = editing?.id === row.id && editing?.field === field;
            const props = {autoFocus: true, disabled, onBlur: () => finish(), onKeyDown: (event: any) => {
                if (event.key === "Escape") { event.preventDefault(); event.stopPropagation(); finish(true); }
                else if (event.key === "Enter" && (field !== memberField || event.ctrlKey || event.metaKey)) { event.preventDefault(); event.stopPropagation(); finish(); }
            }};
            return <div className={`redis-value-inline${field === "score" ? " is-number" : ""}`}>{active ? field === "score"
                ? <InputNumber {...props} variant="filled" controls={false} value={content} placeholder="分值" aria-label="编辑分值" onChange={(score) => update(row.id, {score})}/>
                : field === "key" ? <Input {...props} variant="filled" value={content ?? ""} placeholder="字段名称" aria-label="编辑字段名称" onChange={(event) => update(row.id, {[field]: event.target.value})}/>
                    : <Input.TextArea {...props} variant="filled" value={content ?? ""} placeholder="输入内容" aria-label="编辑内容" spellCheck={false} onChange={(event) => update(row.id, {[field]: event.target.value})}/>
                : <button type="button" className="redis-value-display" disabled={disabled} aria-label={`${writable ? "编辑" : "查看"}${field === "key" ? "字段名称" : field === "score" ? "分值" : "内容"}`} onClick={() => {
                    if (!writable) { expand(row); return; }
                    finish(); const next = {id: row.id, field, original: row[field]}; editingRef.current = next; setEditing(next);
                }}><span>{content === "" || content === undefined || content === null ? <em>空文本</em> : String(content)}</span></button>}</div>;
        };
        const columns: any[] = [{title: "序号", key: "number", width: mobile ? 40 : 52, align: "center", render: (_: any, row: any) => <span className="redis-value-number">{indexes.get(row.id)}</span>}];
        if (kind === "hash") columns.push({title: "字段", dataIndex: "key", width: mobile ? 88 : 160, render: (_: any, row: any) => cell("key", row)});
        columns.push({title: "内容", dataIndex: memberField, width: mobile ? ["hash", "zset"].includes(kind) ? 128 : 192 : kind === "hash" ? 260 : 300, render: (_: any, row: any) => cell(memberField, row)});
        if (kind === "zset") columns.push({title: "分值", dataIndex: "score", width: mobile ? 64 : 100, align: "right", render: (_: any, row: any) => cell("score", row)});
        columns.push({title: "操作", key: "operation", width: 80, align: "center", render: (_: any, row: any) => <ActionDropdown menu={{items: [
            {key: "expand", label: "展开", onClick: () => expand(row)},
            {key: "copy", label: "复制", onClick: () => void copy(String(row[memberField] ?? ""))},
            ...(writable ? [{type: "divider" as const}, {key: "delete", label: "删除", danger: true, onClick: () => onChange?.(rowsRef.current.filter((item: any) => item.id !== row.id))}] : []),
        ]}}><Button color="default" variant="filled" size="small" autoInsertSpace={false} disabled={disabled}>操作</Button></ActionDropdown>});
        return <div className={`${styles.editor} redis-value-browser${fill ? " redis-value-fill" : ""}`}>
            <div className="redis-value-toolbar redis-toolbar">
                {collection ? <Input className="redis-value-search" variant="filled" prefix={<SearchOutlined/>} placeholder="搜索当前页内容" aria-label="搜索当前页内容" allowClear value={keyword}
                    onChange={(event) => { finish(); setKeyword(event.target.value); }}/>
                    : <span className="redis-value-label">{kind === "string" ? "内容" : "内容预览"}</span>}
                <div className="redis-value-tools">
                    {collection && writable && <Button color="default" variant="filled" size="small" icon={<PlusOutlined/>} onClick={() => expand()}>新增</Button>}
                    {!collection && <Button color="default" variant="filled" size="small" icon={<CopyOutlined/>} onClick={() => void copy(text)}>复制</Button>}
                </div>
            </div>
            {collection ? <div className="redis-value-table-area" ref={tableArea} style={fill ? undefined : {height}}>
                <Table ref={tableRef} className="redis-value-table" size="small" rowKey="id" pagination={false} dataSource={filtered} columns={columns} tableLayout="fixed" virtual
                    rowClassName={(_, index) => index === filtered.length - 1 ? "client-last-row" : ""}
                    scroll={{x: mobile ? kind === "hash" ? 336 : 312 : kind === "hash" ? 552 : kind === "zset" ? 532 : 432, y: tableHeight, scrollToFirstRowOnChange: false}}
                    locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}}/>
            </div> : <div className="redis-value-code" style={fill ? undefined : {height}}>
                {AceEditor && !fallback ? <AceEditor value={text} mode={kind === "string" ? "text" : "json"} height="100%" fontSize={12} theme={appearance?.isDarkTheme?.() ? "tomorrow_night" : "textmate"}
                    placeholder="请输入键的内容" readOnly={!writable || kind !== "string"} showPrintMargin={false} wrapEnabled enableSnippets={false}
                    enableBasicAutocompletion={false} enableLiveAutocompletion={false} onChange={(next: string) => { if (writable && kind === "string") onChange?.(next); }} onError={() => setFallback(true)}/>
                    : <Input.TextArea className="redis-value-text" value={text} placeholder="请输入键的内容" aria-label="键内容" spellCheck={false} readOnly={!writable || kind !== "string"}
                        onChange={(event) => { if (writable && kind === "string") onChange?.(event.target.value); }}/>}</div>}
            <ProModal title={<Title>{displayed?.isNew ? kind === "hash" ? "新增字段" : "新增成员" : writable ? "编辑内容" : "查看内容"}</Title>} width={compact ? 440 : 480}
                okText={writable ? "确认" : "关闭"} onCancel={() => setExpanded(undefined)} onOk={() => {
                    if (!expanded || !writable) { setExpanded(undefined); return; }
                    const reason = invalid(expanded);
                    if (reason) { setError(reason); return; }
                    const {isNew, ...row} = expanded;
                    if (kind === "zset") row.score = Number(row.score);
                    onChange?.(isNew ? [...rowsRef.current, row] : rowsRef.current.map((item: any) => item.id === row.id ? row : item));
                    setKeyword(""); setExpanded(undefined);
                    if (isNew) requestAnimationFrame(() => tableRef.current?.scrollTo({key: row.id}));
                }} modalProps={{open: Boolean(expanded), mask: {closable: true}, destroyOnHidden: true,
                    style: {maxWidth: "calc(100vw - 24px)"}, cancelButtonProps: writable ? undefined : {style: {display: "none"}},
                    afterClose: () => { lastExpanded.current = undefined; setError(""); },
                    modalRender: (node) => <div onKeyDown={(event) => { if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "s") { event.preventDefault(); event.stopPropagation(); } }}>{node}</div>,
                }}>
                    <ConfigProvider theme={{components: {Input: {controlHeight: token.controlHeight}, InputNumber: {controlHeight: token.controlHeight}}}}>
                    <Form layout="vertical" variant="filled" disabled={disabled}>
                        {kind === "hash" && <Form.Item label="字段名称"><Input readOnly={!writable} value={displayed?.key ?? ""} placeholder="请输入字段名称" onChange={(event) => { if (writable) { setExpanded({...expanded, key: event.target.value}); setError(""); } }}/></Form.Item>}
                        {kind === "zset" && <Form.Item label="分值" required><InputNumber readOnly={!writable} style={{width: "100%"}} value={displayed?.score} placeholder="请输入分值" onChange={(score) => { if (writable) { setExpanded({...expanded, score}); setError(""); } }}/></Form.Item>}
                        <Form.Item label="内容" style={{marginBottom: 0}} validateStatus={error ? "error" : undefined} help={error || undefined}>
                            <Input.TextArea readOnly={!writable} value={displayed?.[memberField] ?? ""} placeholder="请输入内容" spellCheck={false} style={{height: compact ? 200 : 220, resize: "vertical"}}
                                onChange={(event) => { if (writable) { setExpanded({...expanded, [memberField]: event.target.value}); setError(""); } }}/>
                        </Form.Item>
                    </Form>
                    </ConfigProvider>
            </ProModal>
        </div>;
    }
    if (kind === "string" || !["hash", "list", "set", "zset"].includes(kind)) {
        const text = kind === "string" ? String(value ?? "") : JSON.stringify(value, null, 2) ?? "";
        return <div className={`${styles.editor}${fill ? " redis-value-fill" : ""}${embedded ? " redis-value-embedded" : ""}`}>
        <Input.TextArea className="redis-value-text" variant="filled" value={text} aria-label="键内容" placeholder="请输入键的内容"
            style={fill ? {resize: "none"} : {height, resize: "vertical"}} readOnly={readOnly || kind !== "string"} disabled={disabled} spellCheck={false}
            onChange={(event) => onChange?.(event.target.value)}/>
        <div className="redis-value-footer"><span>{readOnly ? "只读" : ""}</span><Typography.Text className="redis-copy-action" copyable={{text, icon: ["复制", "已复制"]}}/></div>
    </div>;
    }

    const content = (value: any) => <Typography.Paragraph className="redis-value-preview" copyable={{text: String(value ?? "")}}
        ellipsis={{rows: 3, expandable: "collapsible", symbol: (expanded) => expanded ? "收起" : "展开"}}>
        {value === "" || value === undefined || value === null ? <Typography.Text type="secondary">空文本</Typography.Text> : String(value)}
    </Typography.Paragraph>;
    const columns: any[] = [{title: "序号", key: "index", width: 54, align: "center", render: (_: any, __: any, index: number) =>
        <span className="redis-value-index">{start + index + 1}</span>}];
    if (kind === "hash") columns.push({title: "字段", dataIndex: "key", width: 170, render: (key: string, row: any, index: number) => readOnly ? content(key)
        : <Input variant="filled" value={key} disabled={disabled} aria-label={`第 ${start + index + 1} 条字段`} placeholder="字段名称"
            onChange={(event) => update(row.id, {key: event.target.value})}/>});
    columns.push({title: parameter ? "参数" : "内容", dataIndex: kind === "zset" ? "member" : "value", render: (value: any, row: any, index: number) => readOnly ? content(value)
        : <div className="redis-value-cell"><Input.TextArea variant="filled" value={value} disabled={disabled} spellCheck={false}
            autoSize={{minRows: 1, maxRows: parameter ? 3 : 5}} aria-label={`第 ${start + index + 1} 条${parameter ? "参数" : "内容"}`}
            placeholder={parameter ? "输入参数" : "输入内容"} onChange={(event) => update(row.id, {[kind === "zset" ? "member" : "value"]: event.target.value})}/>
            <Typography.Text className="redis-value-copy" copyable={{text: String(value ?? "")}}/>
        </div>});
    if (kind === "zset") columns.push({title: "分值", dataIndex: "score", width: 112, align: "right", render: (score: any, row: any, index: number) => readOnly
        ? <span className="redis-value-score">{score}</span>
        : <InputNumber variant="filled" value={score} disabled={disabled} controls={false} placeholder="分值" aria-label={`第 ${start + index + 1} 条分值`}
            onChange={(score) => update(row.id, {score})}/>});
    if (!readOnly) columns.push({title: "操作", key: "operation", width: 66, align: "center", render: (_: any, row: any, index: number) =>
        <Button type="link" danger size="small" disabled={disabled} aria-label={`删除第 ${start + index + 1} 条`}
            onClick={() => onChange?.(rows.filter((item: any) => item.id !== row.id))}>删除</Button>});

    return <div className={`${styles.editor}${parameter ? " redis-value-parameters" : ""}${embedded ? " redis-value-embedded" : ""}`}>
        {embedded ? <div className="redis-value-entries">
            {rows.map((row: any, index: number) => <div className="redis-value-entry" key={row.id}>
                <div className="redis-value-entry-heading"><span>第 {start + index + 1} {kind === "hash" ? "个字段" : "个成员"}</span>
                    {!readOnly && <Button type="text" danger size="small" disabled={disabled} aria-label={`删除第 ${start + index + 1} 条`}
                        onClick={() => onChange?.(rows.filter((item: any) => item.id !== row.id))}>删除</Button>}</div>
                {columns.filter((column) => column.dataIndex).map((column) => <div className="redis-value-entry-field" key={column.dataIndex}>
                    <span>{column.title}</span>{column.render(row[column.dataIndex], row, index)}
                </div>)}
            </div>)}
            {!rows.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无内容"/>}
        </div> : parameter && !rows.length ? <Typography.Text type="secondary">此命令不带参数</Typography.Text> : <Table size="small" rowKey="id" pagination={false}
            dataSource={rows} columns={columns} tableLayout="fixed" virtual={!embedded} showHeader={!parameter}
            rowClassName={(_, index) => index === rows.length - 1 ? "client-last-row" : ""}
            scroll={{x: kind === "hash" ? 540 : kind === "zset" ? 470 : parameter ? 280 : 340, y: embedded ? undefined : height, scrollToFirstRowOnChange: false}}
            locale={{emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无内容"/>}}/>}
        {!readOnly && <div className="redis-value-footer"><Button color="default" variant="filled" disabled={disabled}
            onClick={() => onChange?.([...rows, {id: ++sequence, key: "", value: "", member: "", score: 0}])}>
            添加{kind === "hash" ? "字段" : parameter ? "参数" : "成员"}</Button>
            {!parameter && <span>共 {rows.length} {kind === "hash" ? "个字段" : "个成员"}</span>}
        </div>}
    </div>;
};
