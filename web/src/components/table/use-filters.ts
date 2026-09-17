import {useEffect, useState} from "react";
import type {Dayjs} from "dayjs";
import type {TableFilterConfig, TableToolbarItem} from "./types";

type FilterState = Record<string, {type: TableFilterConfig["type"]; value?: string | [Dayjs, Dayjs] | null}>;

export default function useTableFilters(filters?: TableFilterConfig[]) {
    const [values, setValues] = useState<FilterState>(() => Object.fromEntries((filters || []).map((filter) => [
        filter.name, {type: filter.type, value: filter.defaultValue ?? (filter.type === "select" ? undefined : null)},
    ])));

    useEffect(() => {
        setValues((current) => {
            const next: FilterState = {};
            let changed = Object.keys(current).length !== (filters?.length || 0);
            filters?.forEach((filter) => {
                if (current[filter.name]?.type === filter.type) {
                    next[filter.name] = current[filter.name];
                } else {
                    changed = true;
                    next[filter.name] = {type: filter.type, value: filter.defaultValue ?? (filter.type === "select" ? undefined : null)};
                }
            });
            return changed ? next : current;
        });
    }, [filters]);

    const params: Record<string, unknown> = {};
    const items: TableToolbarItem[] = [];
    filters?.forEach((filter) => {
        const stored = values[filter.name]?.type === filter.type ? values[filter.name].value : undefined;
        if (filter.type === "select") {
            const options = filter.options ?? Object.entries(filter.valueEnum || {}).map(([value, label]) => ({value, label}));
            const value = filter.value ?? (stored as string | undefined) ?? filter.defaultValue
                ?? (filter.allLabel === false ? options[0]?.value ?? "" : "");
            Object.assign(params, filter.transform ? filter.transform(value) : {[filter.name]: value});
            items.push({
                type: "filter", key: `filter:${filter.name}`, hidden: filter.hidden,
                value, icon: filter.icon || "FilterOutlined", ariaLabel: filter.label, disabled: filter.disabled,
                tooltip: filter.tooltip,
                options: filter.allLabel !== false && !options.some((option) => option.value === "")
                    ? [{value: "", label: filter.allLabel ?? "全部"}, ...options] : options,
                onChange: (value) => {
                    if (filter.value === undefined) setValues((current) => ({...current, [filter.name]: {type: "select", value}}));
                    filter.onChange?.(value);
                },
            });
        } else {
            const value = filter.value !== undefined ? filter.value
                : stored !== undefined ? stored as [Dayjs, Dayjs] | null : filter.defaultValue ?? null;
            const [start, end] = filter.fields || [`${filter.name}_start`, `${filter.name}_end`];
            Object.assign(params, filter.transform ? filter.transform(value) : {
                [start]: value?.[0].startOf("day").format("YYYY-MM-DD HH:mm:ss") || "",
                [end]: value?.[1].endOf("day").format("YYYY-MM-DD HH:mm:ss") || "",
            });
            items.push({
                type: "date", key: `filter:${filter.name}`, hidden: filter.hidden,
                value, ariaLabel: filter.label, disabled: filter.disabled, presets: filter.presets,
                tooltip: filter.tooltip,
                onChange: (value) => {
                    if (filter.value === undefined) setValues((current) => ({...current, [filter.name]: {type: "date", value}}));
                    filter.onChange?.(value);
                },
            });
        }
    });

    return {
        items,
        params,
        reset: () => {
            const next: FilterState = {};
            filters?.forEach((filter) => {
                if (filter.type === "select") {
                    const value = filter.defaultValue ?? (filter.allLabel === false
                        ? (filter.options ? filter.options[0]?.value : Object.keys(filter.valueEnum || {})[0]) ?? "" : "");
                    next[filter.name] = {type: "select", value: filter.defaultValue};
                    filter.onChange?.(value);
                } else {
                    const value = filter.defaultValue ?? null;
                    next[filter.name] = {type: "date", value};
                    filter.onChange?.(value);
                }
            });
            setValues(next);
        },
    };
}
