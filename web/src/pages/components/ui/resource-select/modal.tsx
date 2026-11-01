import {useEffect, useMemo, useRef, useState} from "react";
import type {Key} from "react";
import {App, Button, Grid} from "antd";
import {ModalTable, Title, useTheme} from "sinking-antd";
import type {ModalTableProps} from "sinking-antd";
import type {ResourceSelectDialogProps} from ".";

export interface ResourceSelectModalProps extends ResourceSelectDialogProps {
    title: string;
    placeholder: string;
    columns: ModalTableProps<any>["columns"];
    request: NonNullable<ModalTableProps<any>["request"]>;
    multiple?: boolean;
    disabledRecord?: (record: any) => boolean;
    toolbarActions?: Exclude<ModalTableProps<any>["toolbar"], boolean | undefined>["actions"];
    zIndex?: number;
}

// 选择弹窗只请求当前页；多选保留跨页记录，确认后再写回表单。
export default function ResourceSelectModal({open, onCancel, value, onSelect, title, placeholder, columns, request,
    multiple = false, disabledRecord, toolbarActions, zIndex = 1160}: ResourceSelectModalProps) {
    const {message} = App.useApp();
    const compact = Boolean(useTheme()?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const [selected, setSelected] = useState<Key[]>(value.map(String));
    const records = useRef(new Map<string, any>());
    const currentValue = useRef(value);
    currentValue.current = value;

    useEffect(() => {
        if (!open) return;
        setSelected(currentValue.current.map(String));
        records.current.clear();
    }, [open]);

    const tableColumns = useMemo(() => multiple ? columns : [...(columns || []), {
        title: "操作",
        key: "action",
        width: 82,
        align: "center" as const,
        className: "action-cell",
        render: (_: unknown, record: any) => <Button size="small" variant="filled" autoInsertSpace={false}
            color={value.some((id) => String(id) === String(record.id)) ? "primary" : "default"}
            disabled={disabledRecord?.(record)}
            onClick={() => onSelect([record])}>{value.some((id) => String(id) === String(record.id)) ? "已选" : "选择"}</Button>,
    }], [columns, disabledRecord, multiple, onSelect, value]);

    return <ModalTable<any>
        open={open}
        onCancel={onCancel}
        title={<Title>{title}</Title>}
        width={compact ? 760 : 800}
        minHeight={compact ? 440 : 480}
        scrollMode="modal"
        rowKey={(record) => String(record.id)}
        columns={tableColumns}
        scroll={{x: 520}}
        request={request}
        requestEnabled={open}
        defaultPageSize={10}
        onLoad={(rows) => {
            // 仅保留已选记录及当前页，浏览大量分页时不会累计全部资源。
            const retained = new Map<string, any>();
            selected.forEach((id) => {
                const record = records.current.get(String(id));
                if (record) retained.set(String(id), record);
            });
            rows.forEach((record) => { if (record?.id !== undefined) retained.set(String(record.id), record); });
            records.current = retained;
        }}
        onRequestError={(error) => message.error(error.message)}
        toolbar={{search: {placeholder, debounce: 300}, refresh: true, actions: toolbarActions}}
        rowSelection={multiple ? {
            selectedRowKeys: selected,
            preserveSelectedRowKeys: true,
            getCheckboxProps: (record) => ({disabled: disabledRecord?.(record)}),
            onChange: (keys, rows) => {
                rows.forEach((record) => { if (record?.id !== undefined) records.current.set(String(record.id), record); });
                setSelected(keys);
            },
        } : false}
        modalProps={{
            zIndex,
            destroyOnHidden: true,
            mask: {closable: true},
            footer: multiple ? <>
                <Button onClick={onCancel}>取消</Button>
                <Button type="primary" onClick={() => onSelect(selected.map((id) => records.current.get(String(id)) || {id}))}>确认</Button>
            </> : null,
            style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"},
        } as any}/>
}
