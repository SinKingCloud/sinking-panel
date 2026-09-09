import {useEffect, useMemo, useRef, useState} from "react";
import type {Key} from "react";
import type {TableProps as AntTableProps} from "antd";
import type {DataTableProps} from "./parts";
import type {TableProps} from "./types";

export default function useTableSelection<RecordType extends object>(
    data: readonly RecordType[],
    rowKey: NonNullable<DataTableProps<RecordType>["rowKey"]>,
    config: TableProps<RecordType>["rowSelection"],
) {
    const selection = config ? (config === true ? {} : config) : undefined;
    const enabled = Boolean(selection);
    const controlled = selection?.selectedRowKeys !== undefined;
    const preserve = Boolean(selection?.preserveSelectedRowKeys);
    const [internalKeys, setInternalKeys] = useState<Key[]>(() => selection?.defaultSelectedRowKeys || []);
    const selectedRecords = useRef<Map<Key, RecordType> | undefined>(undefined);
    const records = useMemo(() => {
        if (!enabled) return undefined;
        const indexed = new Map<Key, RecordType>();
        data.forEach((row, index) => indexed.set(
            typeof rowKey === "function" ? rowKey(row, index) : (row as Record<PropertyKey, Key>)[rowKey], row,
        ));
        return indexed;
    }, [data, rowKey, enabled]);
    const requestedKeys = selection?.selectedRowKeys ?? internalKeys;
    const keys = useMemo(() => {
        if (!enabled) return [];
        if (controlled || preserve) return requestedKeys;
        const available = requestedKeys.filter((key) => records!.has(key));
        return available.length === requestedKeys.length ? requestedKeys : available;
    }, [enabled, controlled, preserve, requestedKeys, records]);

    useEffect(() => {
        if (!records || controlled || preserve) return;
        setInternalKeys((current) => {
            const available = current.filter((key) => records.has(key));
            return available.length === current.length ? current : available;
        });
    }, [records, controlled, preserve]);

    const rows: RecordType[] = [];
    const retained = preserve && keys.length ? new Map<Key, RecordType>() : undefined;
    keys.forEach((key) => {
        const row = records?.get(key) ?? (preserve ? selectedRecords.current?.get(key) : undefined);
        if (row !== undefined) {
            rows.push(row);
            retained?.set(key, row);
        }
    });
    selectedRecords.current = retained;
    const snapshot = useRef({keys, rows});
    snapshot.current = {keys, rows};

    type SelectionChange = NonNullable<NonNullable<AntTableProps<RecordType>["rowSelection"]>["onChange"]>;
    const change = (nextKeys: Key[], info: Parameters<SelectionChange>[2], suppliedRows?: RecordType[]) => {
        if (!selection) return;
        const nextRows: RecordType[] = [];
        const nextRecords = preserve ? new Map<Key, RecordType>() : undefined;
        let suppliedRecords: Map<Key, RecordType> | undefined;
        // 受控父组件可能拒绝变更，在下一次渲染前保留其当前选中的行。
        if (controlled && nextRecords) selection.selectedRowKeys?.forEach((key) => {
            const row = records?.get(key) ?? selectedRecords.current?.get(key);
            if (row !== undefined) nextRecords.set(key, row);
        });
        nextKeys.forEach((key) => {
            let row = records?.get(key) ?? selectedRecords.current?.get(key);
            if (row === undefined && suppliedRows?.length) {
                suppliedRecords ??= new Map(suppliedRows.filter((item) => item !== undefined).map((item) => [
                    typeof rowKey === "function" ? rowKey(item) : (item as Record<PropertyKey, Key>)[rowKey], item,
                ]));
                row = suppliedRecords.get(key);
            }
            if (row !== undefined) {
                nextRows.push(row);
                nextRecords?.set(key, row);
            }
        });
        const next = [...nextKeys];
        selectedRecords.current = nextRecords;
        snapshot.current = {keys: next, rows: nextRows};
        if (!controlled) setInternalKeys(next);
        selection.onChange?.(next, nextRows, info);
    };
    const setKeys = (next: Key[]) => change(next, {type: "multiple"});
    const clear = () => {
        if (!selection || selection.clearDisabled) return;
        selection.onClear?.();
        change([], {type: "none"});
    };
    const selectAll = () => {
        if (!selection) return;
        const next = new Set(snapshot.current.keys);
        records?.forEach((row, key) => {
            if (!selection.getCheckboxProps?.(row).disabled) next.add(key);
        });
        change([...next], {type: "all"});
    };
    const invert = () => {
        if (!selection) return;
        const next = new Set(snapshot.current.keys);
        records?.forEach((row, key) => {
            if (selection.getCheckboxProps?.(row).disabled) return;
            if (next.has(key)) next.delete(key);
            else next.add(key);
        });
        change([...next], {type: "invert"});
    };

    let rowSelection: AntTableProps<RecordType>["rowSelection"];
    if (selection) {
        const {actions, onClear, clearDisabled, ...rest} = selection;
        rowSelection = {...rest, selectedRowKeys: keys, onChange: (next, nextRows, info) => change(next, info, nextRows)};
    }
    return {
        rowSelection,
        keys,
        rows,
        actions: typeof selection?.actions === "function" ? selection.actions(keys, rows) : selection?.actions || [],
        clearDisabled: selection?.clearDisabled,
        clear,
        setKeys,
        selectAll,
        invert,
        getKeys: () => [...snapshot.current.keys],
        getRows: () => [...snapshot.current.rows],
    };
}
