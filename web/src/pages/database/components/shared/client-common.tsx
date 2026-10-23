import {memo, useCallback, useLayoutEffect, useMemo, useRef, useState} from "react";
import {Empty, Table, Tooltip, Typography} from "antd";
import {Icon} from "sinking-antd";

export function jsonValue(text: string, kind?: "object" | "array") {
    let value: any;
    try { value = JSON.parse(text); } catch { throw new Error("请输入有效的 JSON"); }
    if (kind === "array" && !Array.isArray(value)) throw new Error("请输入 JSON 数组");
    if (kind === "object" && (!value || Array.isArray(value) || typeof value !== "object")) throw new Error("请输入 JSON 对象");
    return value;
}

const emptyValues: any[] = [];
const emptyLocale = {emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>};

// 外层分配可用空间，表格本身保持自然高度；长表只挂载可见行。
export function VirtualTable({height = 360, autoHeight = true, scroll, rowClassName, ...props}: any) {
    const container = useRef<HTMLDivElement>(null);
    const [bodyHeight, setBodyHeight] = useState(height ?? 360);
    const naturalHeight = height === null;
    useLayoutEffect(() => {
        const element = container.current;
        if (naturalHeight || !autoHeight || !element) return;
        const header = element.querySelector<HTMLElement>(".ant-table-header");
        const table = element.querySelector<HTMLElement>(".ant-table-container");
        const paginations = Array.from(element.querySelectorAll<HTMLElement>(".ant-table-pagination"));
        const size = (value: string) => parseFloat(value) || 0;
        const resize = () => {
            let availableHeight = size(getComputedStyle(element).height);
            if (!availableHeight) return;
            for (const pagination of paginations) {
                const style = getComputedStyle(pagination);
                availableHeight -= pagination.offsetHeight + size(style.marginTop) + size(style.marginBottom);
            }
            const headerHeight = header ? size(getComputedStyle(header).height) : 0;
            const tableStyle = table ? getComputedStyle(table) : undefined;
            const borderHeight = size(tableStyle?.borderTopWidth || "") + size(tableStyle?.borderBottomWidth || "");
            // 使用布局尺寸，不受 modal 缩放动画和分页自动外边距影响。
            const next = Math.max(1, Math.floor(availableHeight - headerHeight - borderHeight));
            setBodyHeight((previous: number) => previous === next ? previous : next);
        };
        const observer = new ResizeObserver(resize);
        observer.observe(element);
        if (header) observer.observe(header);
        paginations.forEach((item) => observer.observe(item));
        resize();
        return () => observer.disconnect();
    }, [autoHeight, naturalHeight, props.columns?.length, Boolean(props.dataSource?.length), Boolean(props.pagination)]);
    const tableScroll = useMemo(() => ({...scroll,
        x: typeof scroll?.x === "number" ? scroll.x : 600,
        ...(!naturalHeight ? {y: autoHeight ? bodyHeight : height} : {}),
    }), [scroll?.x, scroll?.scrollToFirstRowOnChange, naturalHeight, autoHeight, bodyHeight, height]);
    // 按真实数据标记末行，避免虚拟滚动的可见末行被误当成表格底部。
    const getRowClassName = useCallback((record: any, index: number, indent: number) => {
        const className = typeof rowClassName === "function" ? rowClassName(record, index, indent) : rowClassName;
        return [className, index === (props.dataSource?.length || 0) - 1 ? "client-last-row" : ""].filter(Boolean).join(" ");
    }, [rowClassName, props.dataSource?.length]);
    const table = <Table {...props} rowClassName={getRowClassName} virtual={!naturalHeight} scroll={tableScroll}/>;
    return autoHeight && !naturalHeight
        ? <div ref={container} style={{flex: "1 1 0", minWidth: 0, minHeight: 140, overflow: "hidden"}}>{table}</div>
        : table;
}

export const ResultTable = memo(function ResultTable({result, loading = false, action, actionWidth = 116, actionAlign, height = 360, autoHeight = false, rowOffset = 0, rowNumbers = true, columnMetadata, sort, order, onSortChange}: any) {
    const rows = useMemo(() => (result?.rows || emptyValues).map((values: any[], index: number) => ({index, values})), [result?.rows]);
    const comments = useMemo(() => new Map<string, string>((columnMetadata || emptyValues).map((column: any) => [column.name, column.comment])), [columnMetadata]);
    const fields = useMemo(() => [
        ...(rowNumbers && result?.columns?.length ? [{title: "#", key: "row-number", width: 52, render: (_: any, row: any) => <span className="client-row-number">{rowOffset + row.index + 1}</span>}] : []),
        ...(result?.columns || emptyValues).map((name: string, index: number) => ({
            title: <span className="client-column-title">
                <Typography.Text ellipsis={{tooltip: true}}>{name}</Typography.Text>
                {comments.get(name) && <Tooltip title={comments.get(name)}><span className="client-column-comment" tabIndex={0} aria-label={`${name} 备注：${comments.get(name)}`}
                    onClick={(event) => event.stopPropagation()} onKeyDown={(event) => { if (event.key === "Enter") event.stopPropagation(); }}>
                    <Icon type="QuestionCircleOutlined"/>
                </span></Tooltip>}
            </span>,
            key: `${name}-${index}`, width: 160,
            shouldCellUpdate: (next: any, previous: any) => next.values[index] !== previous.values[index],
            ...(onSortChange ? {dataIndex: name, sorter: true, sortOrder: sort === name ? (order === "asc" ? "ascend" as const : "descend" as const) : null} : {}),
            render: (_: any, row: any) => {
                const value = row.values[index];
                const text = value === null ? "NULL" : value === undefined ? "" : typeof value === "object" ? JSON.stringify(value) : String(value);
                return <Tooltip title={text.length < 3000 ? text : undefined}><span className={`client-cell${value === null ? " client-null" : ""}`}>{text}</span></Tooltip>;
            },
        })),
    ], [result?.columns, rowNumbers, rowOffset, comments, sort, order, Boolean(onSortChange)]);
    const columns = useMemo(() => action ? [...fields, {title: "操作", key: "action", width: actionWidth, align: actionAlign, fixed: "right" as const,
        render: (_: any, row: any) => action(row.values, row.index)}] : fields, [fields, action, actionWidth, actionAlign]);
    return <VirtualTable size="small" loading={loading} pagination={false} rowKey="index" height={height} autoHeight={autoHeight}
        showSorterTooltip={{target: "sorter-icon"}}
        onChange={onSortChange ? (_: any, __: any, sorter: any, extra: any) => {
            if (extra.action !== "sort") return;
            const value = Array.isArray(sorter) ? sorter[0] : sorter;
            onSortChange(value.order ? value.field : "", value.order === "ascend" ? "asc" : value.order === "descend" ? "desc" : "");
        } : undefined}
        locale={emptyLocale}
        scroll={{x: Math.max(600, (result?.columns?.length || 0) * 160 + (rowNumbers ? 52 : 0) + (action ? actionWidth : 0))}}
        dataSource={rows}
        columns={columns}/>;
});
