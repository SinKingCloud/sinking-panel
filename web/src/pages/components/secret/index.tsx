import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Button, Empty, Grid, Tooltip} from "antd";
import {createStyles} from "antd-style";
import {Title, useTheme} from "sinking-antd";
import ModalTable from "@/components/modal-table";
import type {ModalTableRef, TableRequestParams} from "@/components/modal-table";
import Dropdown from "@/pages/components/stable-dropdown";
import {deleteSecret, getSecretList} from "@/service/api/secret";
import type {SecretRecord} from "@/service/api/secret";
import {getEnum as getSystemEnum} from "@/service/api/system";
import SecretForm from "./form";
import type {SecretChange, SecretFormRef} from "./form";

export type {SecretChange} from "./form";

export interface SecretManagerRef {
    open: () => void;
    close: () => void;
    refresh: () => void;
    openCreate: () => void;
}

interface SecretManagerProps {
    zIndex?: number;
    onChange?: (change: SecretChange) => void;
}

const useStyles = createStyles(({css, token}: any) => ({
    table: css`
        .secret-name,
        .secret-provider,
        .secret-time {
            display: block;
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            line-height: 20px;
        }
        .secret-name {
            padding: 0;
            border: 0;
            background: transparent;
            color: inherit;
            font: inherit;
            text-align: left;
            cursor: pointer;
        }
        .secret-name:hover,
        .secret-name:focus-visible { color: ${token.colorPrimary}; }
        .secret-time { color: ${token.colorTextSecondary}; font-variant-numeric: tabular-nums; }
    `,
}));

const SecretManager = forwardRef<SecretManagerRef, SecretManagerProps>(({zIndex = 1050, onChange}, ref) => {
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles();
    const tableRef = useRef<ModalTableRef<SecretRecord> | null>(null);
    const formRef = useRef<SecretFormRef | null>(null);
    const confirmRef = useRef<{destroy: () => void} | null>(null);
    const mountedRef = useRef(true);
    const openRef = useRef(false);
    const sessionRef = useRef(0);
    const enumRequestRef = useRef(0);
    const onChangeRef = useRef(onChange);
    onChangeRef.current = onChange;
    const [open, setOpen] = useState(false);
    const [session, setSession] = useState(0);
    const [providerFilter, setProviderFilter] = useState("");
    const [providerNames, setProviderNames] = useState<Record<string, string>>({});
    const [listError, setListError] = useState("");
    const [deletingIds, setDeletingIds] = useState<number[]>([]);
    const providerOptions = useMemo(() => [
        {value: "", label: "全部厂商"},
        ...Object.entries(providerNames).map(([value, label]) => ({value, label})),
    ], [providerNames]);

    const loadProviders = useCallback(async () => {
        const requestId = ++enumRequestRef.current;
        try {
            const response = await getSystemEnum({body: {name: "secret"}});
            if (!mountedRef.current || enumRequestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "获取密钥厂商失败");
                return;
            }
            setProviderNames(response.data?.provider || {});
        } catch {
            if (mountedRef.current && enumRequestRef.current === requestId) {
                message.error("获取密钥厂商失败");
            }
        }
    }, [message]);

    const close = useCallback(() => {
        sessionRef.current += 1;
        enumRequestRef.current += 1;
        openRef.current = false;
        formRef.current?.close();
        confirmRef.current?.destroy();
        confirmRef.current = null;
        setOpen(false);
        setProviderFilter("");
        setProviderNames({});
        setListError("");
    }, []);

    const show = useCallback(() => {
        formRef.current?.close();
        confirmRef.current?.destroy();
        confirmRef.current = null;
        sessionRef.current += 1;
        openRef.current = true;
        setSession(sessionRef.current);
        setProviderFilter("");
        setProviderNames({});
        setListError("");
        setOpen(true);
        void loadProviders();
    }, [loadProviders]);

    const refresh = useCallback(() => {
        if (openRef.current) {
            tableRef.current?.reload();
            void loadProviders();
        }
    }, [loadProviders]);

    const openCreate = useCallback(() => {
        formRef.current?.open();
    }, []);

    useImperativeHandle(ref, () => ({open: show, close, refresh, openCreate}), [close, openCreate, refresh, show]);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            openRef.current = false;
            sessionRef.current += 1;
            enumRequestRef.current += 1;
            confirmRef.current?.destroy();
        };
    }, []);

    const request = useCallback(async ({current, pageSize, keyword, provider}: TableRequestParams) => {
        const response = await getSecretList({body: {
            page: current,
            page_size: pageSize,
            keyword: typeof keyword === "string" ? keyword.trim() : "",
            provider,
            order_by_field: "id",
            order_by_type: "desc",
        }});
        if (response?.code !== 200 || !Array.isArray(response.data?.list)) {
            throw new Error(response?.message || "获取密钥列表失败");
        }
        const data: SecretRecord[] = response.data.list.map((item: SecretRecord) => ({
            id: Number(item.id),
            name: item.name,
            provider: Number(item.provider),
            create_time: item.create_time,
            update_time: item.update_time,
        }));
        return {data, total: Number(response.data.total || 0), success: true};
    }, []);

    const changed = useCallback((change: SecretChange) => {
        if (!mountedRef.current) {
            return;
        }
        if (openRef.current) {
            tableRef.current?.reload();
        }
        // 调用方的刷新失败不改变密钥操作已成功的结果。
        try {
            void Promise.resolve(onChangeRef.current?.(change)).catch(() => undefined);
        } catch {}
    }, []);

    const confirmDelete = useCallback((record: SecretRecord) => {
        const sessionId = sessionRef.current;
        confirmRef.current?.destroy();
        const confirmation = modal.confirm({
            title: "删除密钥",
            content: `确定删除密钥“${record.name}”吗？将解除证书关联并关闭自动续签。`,
            zIndex: zIndex + 30,
            okText: "删除",
            cancelText: "取消",
            okButtonProps: {danger: true},
            mask: {closable: true},
            onOk: async () => {
                setDeletingIds((ids) => [...ids, record.id]);
                try {
                    const response = await deleteSecret({body: {id: record.id}});
                    if (response?.code !== 200) {
                        throw new Error(response?.message || "删除密钥失败");
                    }
                    if (mountedRef.current && sessionRef.current === sessionId) {
                        message.success(response?.message || "删除成功");
                    }
                    changed({action: "delete", id: record.id});
                } catch (error) {
                    if (mountedRef.current && sessionRef.current === sessionId) {
                        message.error(error instanceof Error ? error.message : "删除密钥失败");
                    }
                    throw error;
                } finally {
                    if (mountedRef.current) {
                        setDeletingIds((ids) => ids.filter((id) => id !== record.id));
                    }
                }
            },
            afterClose: () => {
                if (confirmRef.current === confirmation) {
                    confirmRef.current = null;
                }
            },
        } as any);
        confirmRef.current = confirmation;
    }, [changed, message, modal, zIndex]);

    const columns = useMemo<any[]>(() => [
        {
            title: "密钥名称",
            dataIndex: "name",
            key: "name",
            width: 190,
            render: (name: string, record: SecretRecord) => <button
                type="button"
                className="secret-name"
                title={name}
                disabled={deletingIds.includes(record.id)}
                onClick={() => formRef.current?.open(record)}>{name}</button>,
        },
        {
            title: "厂商",
            dataIndex: "provider",
            key: "provider",
            width: 125,
            render: (provider: number) => <span className="secret-provider">
                {providerNames[provider] || "未知厂商"}
            </span>,
        },
        {
            title: "更新时间",
            dataIndex: "update_time",
            key: "update_time",
            width: 160,
            render: (value: string) => <Tooltip title={value}>
                <span className="secret-time">{value ? String(value).slice(0, 16) : "—"}</span>
            </Tooltip>,
        },
        {
            title: "操作",
            key: "action",
            width: 82,
            fixed: "right",
            className: "action-cell",
            render: (_: unknown, record: SecretRecord) => <Dropdown
                trigger={["click"]}
                placement="bottom"
                arrow
                disabled={deletingIds.includes(record.id)}
                menu={{
                    items: [{key: "edit", label: "编辑"}, {key: "delete", label: "删除", danger: true}],
                    onClick: ({key}) => key === "edit" ? formRef.current?.open(record) : confirmDelete(record),
                }}>
                <Button size="small" autoInsertSpace={false} disabled={deletingIds.includes(record.id)} aria-label={`操作密钥 ${record.name}`}>操作</Button>
            </Dropdown>,
        },
    ], [confirmDelete, deletingIds, providerNames]);

    return <>
        <ModalTable<SecretRecord>
            key={session}
            ref={tableRef}
            open={open}
            onCancel={close}
            title={<Title>密钥管理</Title>}
            width={compact ? 700 : 740}
            minHeight={compact ? 440 : 480}
            className={styles.table}
            scrollMode="modal"
            columns={columns}
            rowKey="id"
            scroll={{x: 557}}
            request={request}
            requestEnabled={open}
            params={{provider: providerFilter}}
            defaultPageSize={10}
            onLoad={() => setListError("")}
            onRequestError={(error) => {
                setListError(error.message);
                message.error(error.message);
            }}
            emptyContent={<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={listError || "暂无数据"}/>}
            toolbar={{
                search: {placeholder: "搜索密钥名称", ariaLabel: "搜索密钥", maxLength: 100, debounce: 300},
                refresh: true,
                actions: [
                    {
                        key: "provider",
                        type: "filter",
                        value: providerFilter,
                        options: providerOptions,
                        icon: "CloudOutlined",
                        ariaLabel: "筛选密钥厂商",
                        onChange: setProviderFilter,
                    },
                    {key: "create", type: "button", label: "添加", icon: "PlusOutlined", onClick: openCreate},
                ],
            }}
            pagination={{unit: "条"}}
            modalProps={{
                zIndex,
                destroyOnHidden: true,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"},
            }}/>
        <SecretForm ref={formRef} zIndex={zIndex + 20} onSuccess={changed}/>
    </>;
});

SecretManager.displayName = "SecretManager";

export default memo(SecretManager);
