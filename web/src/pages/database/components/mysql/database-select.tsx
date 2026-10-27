import {useEffect, useRef, useState} from "react";
import {App, Empty, Select, Spin} from "antd";
import type {SelectProps} from "antd";
import {useMysqlClient} from "./context";
import type {DatabasePanelRecord} from "./entities";

interface DatabaseSelectProps extends Omit<SelectProps<string>, "options" | "loading" | "showSearch" | "filterOption" | "searchValue" | "onSearch" | "open" | "onOpenChange" | "onPopupScroll"> {
    record: DatabasePanelRecord;
    disabledValues?: string[];
    disableSystem?: boolean;
}

// 授权与跨库引用共用目录分页接口；系统库限制仅用于授权。
export default function MySQLDatabaseSelect({record, disabled, disabledValues = [], disableSystem = true, ...props}: DatabaseSelectProps) {
    const {api} = useMysqlClient();
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
    const scope = useRef(record.id);
    scope.current = record.id;
    const available = useRef(!disabled);
    available.current = !disabled;

    const cancel = () => {
        sequence.current++; pending.current = 0;
        clearTimeout(timer.current); timer.current = undefined;
    };
    const load = async (page: number) => {
        if (!live.current || !opened.current || !available.current || pending.current) return;
        const request = ++sequence.current;
        const current = () => live.current && opened.current && available.current && scope.current === record.id && request === sequence.current;
        pending.current = page; retry.current = 0; setLoading(true);
        try {
            const data = await api.databases(record, {keyword: query.current, page, page_size: 20, show_system: true});
            if (!current()) return;
            if (!Array.isArray(data?.list) || !data?.list?.every((name: any) => typeof name === "string")
                || !Number.isSafeInteger(data?.total) || data?.total < 0 || !Number.isInteger(data?.page) || data?.page < 1
                || data?.page > page || !Number.isInteger(data?.page_size) || data?.page_size < 1
                || (data?.page < page && data?.page * data?.page_size < data?.total)
                || data?.list?.length > data?.page_size || data?.list?.length > data?.total) throw new Error("数据库列表返回的分页数据不完整");
            const next = (data?.list as string[]).map((name) => ({value: name, label: name}));
            setOptions((previous) => Array.from(new Map([...(data?.page === 1 ? [] : previous), ...next].map((option) => [option.value, option])).values()));
            pagination.current = {page: data?.page, pageSize: data?.page_size, total: data?.total};
        } catch (reason: any) {
            if (!current()) return;
            retry.current = page;
            opened.current = false; setOpen(false); cancel(); setLoading(false);
            message.error({key: `mysql-databases-${record.id}`, content: reason?.message || "加载数据库失败"});
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
    }, [record.id]);
    useEffect(() => {
        if (!disabled) return;
        const interrupted = pending.current;
        opened.current = false; setOpen(false); cancel();
        if (interrupted) retry.current = interrupted;
        setLoading(false);
    }, [disabled]);

    const visibleOptions = options.filter((option) => !disabledValues.includes(option.value)
        && (!disableSystem || !["mysql", "information_schema", "performance_schema", "sys"].includes(option.value.toLowerCase())));
    useEffect(() => {
        const current = pagination.current;
        // 隐藏系统库和已选数据库后，仍可继续读取后面的可用选项。
        if (open && !disabled && !loading && !retry.current && visibleOptions.length < 20
            && current.page > 0 && current.page * current.pageSize < current.total) void load(current.page + 1);
    }, [open, disabled, loading, visibleOptions.length, options.length]);

    return <Select<string> placeholder="选择要授权的数据库" aria-label="权限数据库" {...props} disabled={disabled}
        showSearch virtual filterOption={false} searchValue={keyword} open={open && !disabled}
        options={visibleOptions} loading={loading}
        notFoundContent={loading ? <div style={{padding: 12, textAlign: "center"}}><Spin size="small"/></div>
            : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}
        popupRender={(menu) => <>{menu}{visibleOptions.length > 0 && loading && <div role="status" style={{padding: 8, textAlign: "center"}}>
            <Spin size="small"/></div>}</>}
        onSearch={(text) => {
            setKeyword(text);
            const next = text.trim();
            if (next === query.current) return;
            cancel(); query.current = next; retry.current = 0;
            pagination.current = {page: 0, pageSize: 20, total: 0};
            setOptions([]); setLoading(opened.current);
            if (opened.current) timer.current = setTimeout(() => { timer.current = undefined; void load(1); }, 250);
        }}
        onOpenChange={(open) => {
            if (open && !available.current) return;
            opened.current = open; setOpen(open);
            if (open) {
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
