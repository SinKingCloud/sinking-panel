import {useEffect, useRef, useState} from "react";
import {App, Empty, Select, Spin} from "antd";
import type {SelectProps} from "antd";
import {useDatabaseClient} from "./context";
import type {DatabasePanelRecord} from "./entities";

interface DatabaseSelectProps extends Omit<SelectProps<string>, "options" | "loading" | "showSearch" | "filterOption" | "searchValue" | "onSearch" | "open" | "onOpenChange" | "onPopupScroll"> {
    record: DatabasePanelRecord;
}

// 每个选择器独立按需读取数据库，避免筛选与授权表单相互影响。
export default function DatabaseSelect({record, ...props}: DatabaseSelectProps) {
    const {operate} = useDatabaseClient();
    const {message} = App.useApp();
    const [keyword, setKeyword] = useState("");
    const [options, setOptions] = useState<{value: string; label: string}[]>([]);
    const [loading, setLoading] = useState(false);
    const [open, setOpen] = useState(false);
    const live = useRef(false);
    const opened = useRef(false);
    const sequence = useRef(0);
    const pending = useRef(0);
    const query = useRef("");
    const pagination = useRef({page: 0, pageSize: 20, total: 0});
    const retry = useRef(0);
    const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const scopeKey = JSON.stringify([record.id, record.database_name || record.name]);
    const scope = useRef(scopeKey);
    scope.current = scopeKey;

    const cancel = () => {
        sequence.current++; pending.current = 0;
        clearTimeout(timer.current); timer.current = undefined;
    };
    const load = async (page: number) => {
        if (!live.current || !opened.current || pending.current) return;
        const request = ++sequence.current;
        const current = () => live.current && opened.current && scope.current === scopeKey && request === sequence.current;
        pending.current = page; retry.current = 0; setLoading(true);
        try {
            const data = await operate(record, "users", {action: "databases", keyword: query.current, page, page_size: 20});
            if (!current()) return;
            if (!Array.isArray(data?.list) || !data.list.every((name: any) => typeof name === "string")
                || !Number.isSafeInteger(data.total) || data.total < 0 || !Number.isInteger(data.page) || data.page < 1
                || data.page > page || !Number.isInteger(data.page_size) || data.page_size < 1
                || data.list.length > data.page_size || data.list.length > data.total) throw new Error("数据库列表返回的分页数据不完整");
            const next = (data.list as string[]).map((name) => ({value: name, label: name}));
            setOptions((previous) => Array.from(new Map([...(data.page === 1 ? [] : previous), ...next].map((option) => [option.value, option])).values()));
            pagination.current = {page: data.page, pageSize: data.page_size, total: data.total};
        } catch (reason: any) {
            if (!current()) return;
            retry.current = page;
            opened.current = false; setOpen(false); cancel(); setLoading(false);
            message.error(reason?.message || "加载数据库失败");
        } finally {
            if (current()) { pending.current = 0; setLoading(false); }
        }
    };
    useEffect(() => {
        live.current = true;
        return () => { live.current = false; cancel(); };
    }, []);
    useEffect(() => {
        cancel(); query.current = ""; retry.current = 0;
        pagination.current = {page: 0, pageSize: 20, total: 0};
        setOptions([]); setKeyword(""); setLoading(false);
        if (opened.current) void load(1);
    }, [scopeKey]);

    return <Select<string> {...props} virtual showSearch filterOption={false} searchValue={keyword} open={open}
        options={options} loading={loading}
        notFoundContent={loading ? <div style={{padding: 12, textAlign: "center"}}><Spin size="small"/></div>
            : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}
        popupRender={(menu) => <>{menu}{options.length > 0 && loading && <div role="status" style={{padding: 8, textAlign: "center"}}><Spin size="small"/></div>}</>}
        onSearch={(text) => {
            setKeyword(text);
            const next = text.trim();
            if (next === query.current) return;
            cancel(); query.current = next; retry.current = 0;
            pagination.current = {page: 0, pageSize: 20, total: 0};
            setOptions([]); setLoading(opened.current);
            if (opened.current) timer.current = setTimeout(() => { timer.current = undefined; void load(1); }, 250);
        }}
        onOpenChange={(next) => {
            opened.current = next; setOpen(next);
            if (next) {
                if (retry.current || !pagination.current.page) void load(retry.current || 1);
            } else {
                const interrupted = pending.current;
                cancel(); if (interrupted) retry.current = interrupted;
                setLoading(false);
            }
        }}
        onPopupScroll={(event) => {
            const {scrollTop, scrollHeight, clientHeight} = event.currentTarget;
            const current = pagination.current;
            if (!retry.current && current.page > 0 && current.page * current.pageSize < current.total
                && scrollHeight - scrollTop - clientHeight <= 24) void load(current.page + 1);
        }}/>
}
