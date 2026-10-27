import {useEffect, useRef, useState} from "react";
import {App, Empty, Select, Spin} from "antd";
import {useMysqlClient} from "./context";
import {canChangePassword} from "./client-common";

type AccountOption = {value: string; label: string; current: boolean};
type AccountSelectProps = {record: any; database: string; value?: string; onChange: (value: string, account?: AccountOption) => void};
const pageSize = 20;
const accountLabel = (name: string, host: string) => `${name || "匿名账户"}${host ? `@${host}` : ""}`;
const selectedLabel = (value: string) => {
    try {
        const account = JSON.parse(value);
        if (Array.isArray(account) && account.length === 2 && account.every((part) => typeof part === "string")) {
            return accountLabel(account[0], account[1]);
        }
    } catch {}
    return value;
};

export default function MySQLAccountSelect({record, database, value, onChange}: AccountSelectProps) {
    const {api} = useMysqlClient();
    const {message} = App.useApp();
    const [keyword, setKeyword] = useState("");
    const [options, setOptions] = useState<AccountOption[]>([]);
    const [loading, setLoading] = useState(false);
    const [open, setOpen] = useState(false);
    const live = useRef(false);
    const opened = useRef(false);
    const sequence = useRef(0);
    const pendingPage = useRef(0);
    const failedPage = useRef(0);
    const query = useRef("");
    const pagination = useRef({page: 0, pageSize, total: 0});
    const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const scope = JSON.stringify([record.id, record.database_name || record.name, database]);
    const currentScope = useRef(scope);
    currentScope.current = scope;

    const cancel = () => {
        sequence.current++;
        pendingPage.current = 0;
        clearTimeout(timer.current); timer.current = undefined;
    };
    const reset = () => {
        cancel(); failedPage.current = 0;
        pagination.current = {page: 0, pageSize, total: 0};
        setOptions([]); setLoading(false);
    };
    const load = async (page: number) => {
        if (!live.current || !opened.current || pendingPage.current) return;
        const request = ++sequence.current;
        const current = () => live.current && opened.current && currentScope.current === scope && request === sequence.current;
        pendingPage.current = page; failedPage.current = 0; setLoading(true);
        try {
            const data = await api.users(record, {action: "list", database, keyword: query.current, page, page_size: pagination.current.pageSize});
            if (!current()) return;
            if (!Array.isArray(data?.list) || !Number.isSafeInteger(data?.total) || data?.total < 0
                || !Number.isInteger(data?.page) || data?.page < 1 || data?.page > page
                || !Number.isInteger(data?.page_size) || data?.page_size < 1 || data?.page > Math.max(1, Math.ceil(data?.total / data?.page_size))
                || (data?.page < page && data?.page * data?.page_size < data?.total)
                || data?.list?.length > data?.page_size || data?.list?.length > data?.total
                || !data?.list?.every((account: any) => typeof account?.name === "string" && typeof account.host === "string")) {
                throw new Error("账户列表分页数据无效，请重试");
            }
            const next: AccountOption[] = data?.list?.filter(canChangePassword).map((account: any) => ({
                value: JSON.stringify([account.name, account.host]),
                label: `${accountLabel(account.name, account.host)}${account.current ? "（当前连接）" : ""}`,
                current: Boolean(account.current),
            }));
            setOptions((previous) => Array.from(new Map([...(page === 1 ? [] : previous), ...next].map((option) => [option.value, option])).values()));
            pagination.current = {page: data?.page, pageSize: data?.page_size, total: data?.total};
        } catch (reason: any) {
            if (!current()) return;
            failedPage.current = page;
            opened.current = false; setOpen(false); cancel(); setLoading(false);
            message.error({key: `mysql-account-select-${record.id}`, content: reason?.message || "加载账户失败"});
        } finally {
            if (current()) { pendingPage.current = 0; setLoading(false); }
        }
    };

    useEffect(() => {
        live.current = true;
        return () => { live.current = false; cancel(); };
    }, []);
    useEffect(() => {
        query.current = ""; setKeyword(""); reset();
        if (opened.current) void load(1);
    }, [scope]);
    useEffect(() => {
        const current = pagination.current;
        // 过滤不可修改账户后补足首屏，避免空下拉无法滚动到后续页。
        if (open && !loading && !failedPage.current && options.length < pageSize
            && current.page > 0 && current.page * current.pageSize < current.total) void load(current.page + 1);
    }, [open, loading, options.length]);

    return <Select<string, AccountOption> value={value} onChange={(selected, option) => onChange(selected, Array.isArray(option) ? option[0] : option)} placeholder="选择要修改密码的账户" aria-label="选择要修改密码的账户"
        showSearch filterOption={false} searchValue={keyword} options={options} loading={loading} open={open}
        labelRender={({value: selected}) => selectedLabel(String(selected))}
        notFoundContent={loading ? <div style={{padding: 12, textAlign: "center"}}><Spin size="small"/></div>
            : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={query.current ? "没有匹配的账户" : "暂无数据"}/>}
        popupRender={(menu) => <>{menu}{options.length > 0 && loading &&
            <div role="status" style={{padding: 8, textAlign: "center"}}><Spin size="small"/></div>}</>}
        onSearch={(text) => {
            setKeyword(text);
            const next = text.trim();
            if (query.current === next) return;
            query.current = next; reset();
            if (!opened.current) return;
            setLoading(true);
            timer.current = setTimeout(() => { timer.current = undefined; void load(1); }, 250);
        }}
        onOpenChange={(open) => {
            opened.current = open; setOpen(open);
            if (!open) {
                const interrupted = pendingPage.current;
                cancel(); if (interrupted) failedPage.current = interrupted;
                setLoading(false);
            } else if (failedPage.current || !pagination.current.page) void load(failedPage.current || 1);
        }}
        onPopupScroll={(event) => {
            const {scrollTop, scrollHeight, clientHeight} = event.currentTarget;
            const current = pagination.current;
            if (!failedPage.current && current.page > 0 && current.page * current.pageSize < current.total
                && scrollHeight - scrollTop - clientHeight <= 24) void load(current.page + 1);
        }}/>
}
