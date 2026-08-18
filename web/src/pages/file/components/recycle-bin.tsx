import React, {
    forwardRef,
    memo,
    useCallback,
    useEffect,
    useImperativeHandle,
    useMemo,
    useRef,
    useState,
} from "react";
import {App, Button, Card, Flex, Grid, Space, Spin, theme, Typography} from "antd";
import type {ModalProps} from "antd";
import {Icon, ProModal, ProTable, Title, useTheme} from "sinking-antd";
import type {ProColumns, ProModalRef, ProTableProps, ProTableRef} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import {
    clearRecycle,
    deleteRecycleItem,
    getRecycleCount,
    getRecycleList,
    restoreRecycleItem,
} from "@/service/api/file";
import {getData} from "@/utils/page";
import {formatFileSize, formatFileTime} from "../utils";

export interface FileRecycleBinRef {
    open: () => void;
}

interface FileRecycleBinProps {
    onMutation: () => void;
}

type RecycleAction = "restore" | "delete";

interface RecycleSummaryProps {
    compact: boolean;
    count: any;
    fontSize: number;
    onClear: () => void;
}

interface RecycleRowActionsProps {
    id: string;
    name: string;
    onDelete: (names: string[], name?: string) => void;
    onRestore: (names: string[], name?: string) => void;
}

const recycleSort = {update_time: "descend"};

const normalizeNames = (names: string[]) => Array.from(new Set((names || []).filter(Boolean)));

const RecycleSummary = memo(({
    compact,
    count,
    fontSize,
    onClear,
}: RecycleSummaryProps) => (
    <Card
        size="small"
        variant="borderless"
        style={{marginTop: 6}}
        styles={{body: {padding: "10px 15px"}}}>
        <Flex
            align="center"
            justify="space-between"
            gap={8}
            style={{minWidth: 0}}>
            <div style={{minWidth: 0, flex: 1, overflowX: "auto", whiteSpace: "nowrap"}}>
                <Space size={compact ? 12 : 14} wrap={false}>
                    <Space size={4}>
                        <Typography.Text type="secondary" style={{fontSize}}>
                            占用空间
                        </Typography.Text>
                        <Typography.Text style={{fontSize}}>
                            {formatFileSize(count.size)}
                        </Typography.Text>
                    </Space>
                    <Space size={4}>
                        <Typography.Text type="secondary" style={{fontSize}}>
                            文件
                        </Typography.Text>
                        <Typography.Text style={{fontSize}}>{count.file}</Typography.Text>
                    </Space>
                    <Space size={4}>
                        <Typography.Text type="secondary" style={{fontSize}}>
                            文件夹
                        </Typography.Text>
                        <Typography.Text style={{fontSize}}>{count.dir}</Typography.Text>
                    </Space>
                </Space>
            </div>
            <Button
                type="text"
                size="small"
                danger
                style={{fontSize}}
                disabled={count.file + count.dir === 0}
                onClick={onClear}>
                清空
            </Button>
        </Flex>
    </Card>
));

RecycleSummary.displayName = "RecycleSummary";

const RecycleRowActions = memo(({
    id,
    name,
    onDelete,
    onRestore,
}: RecycleRowActionsProps) => {
    const restore = useCallback(() => onRestore([id], name), [id, name, onRestore]);
    const remove = useCallback(() => onDelete([id], name), [id, name, onDelete]);
    const menu = useMemo(() => ({
        items: [
            {key: "restore", label: "恢复", onClick: restore},
            {key: "delete", label: "删除", danger: true, onClick: remove},
        ],
    }), [remove, restore]);

    return (
        <Dropdown
            menu={menu}
            trigger={["click"]}
            placement="bottom"
            arrow>
            <Button size="small">操作</Button>
        </Dropdown>
    );
});

RecycleRowActions.displayName = "RecycleRowActions";

const FileRecycleBin = memo(forwardRef<FileRecycleBinRef, FileRecycleBinProps>(({onMutation}, ref) => {
    const {message, modal} = App.useApp();
    const {token} = theme.useToken();
    const screens = Grid.useBreakpoint();
    const appTheme = useTheme();
    const compact = Boolean(appTheme?.isCompactTheme?.());
    const summaryFontSize = compact ? 11 : token.fontSizeSM;
    const modalBackground = appTheme?.isDarkMode?.() || appTheme?.isDarkTheme?.()
        ? token.colorBgElevated
        : `color-mix(in srgb, ${token.colorBgLayout} 60%, ${token.colorBgContainer})`;
    const modalRef = useRef<ProModalRef | null>(null);
    const tableRef = useRef<ProTableRef | null>(null);
    const countRequestRef = useRef(0);
    const countRef = useRef<any | undefined>(undefined);
    const [loading, setLoading] = useState(false);
    const [count, setCount] = useState<any>();

    useEffect(() => () => {
        countRequestRef.current += 1;
        countRef.current = undefined;
    }, []);

    const loadCount = useCallback(async (initial = false) => {
        const requestId = ++countRequestRef.current;
        if (initial) {
            countRef.current = undefined;
            setCount(undefined);
            setLoading(true);
            modalRef.current?.show();
        }
        const response = await getRecycleCount();
        if (requestId !== countRequestRef.current) {
            return;
        }
        if (response?.code !== 200 || !response.data) {
            if (response) {
                message.error(response.message || "获取回收站信息失败");
            }
            setLoading(false);
            if (!countRef.current) {
                modalRef.current?.hide();
            }
            return;
        }
        countRef.current = response.data;
        setCount(response.data);
        setLoading(false);
    }, [message]);

    const open = useCallback(() => {
        void loadCount(true);
    }, [loadCount]);

    useImperativeHandle(ref, () => ({
        open,
    }), [open]);

    const refreshAfterMutation = useCallback(() => {
        tableRef.current?.clearSelectedRows?.();
        tableRef.current?.refreshTableData?.();
        void loadCount();
        onMutation();
    }, [loadCount, onMutation]);

    const mutate = useCallback((names: string[], action: RecycleAction) => {
        const nextNames = normalizeNames(names);
        if (nextNames.length === 0) {
            message.warning("请选择回收站文件");
            return;
        }
        const request = action === "restore" ? restoreRecycleItem : deleteRecycleItem;
        return request({
            body: {names: nextNames},
            onSuccess: (response) => {
                message.success(response?.message || (action === "restore" ? "恢复成功" : "彻底删除成功"));
                refreshAfterMutation();
            },
            onFail: (response) => {
                message.error(response?.message || (action === "restore" ? "恢复失败" : "彻底删除失败"));
                refreshAfterMutation();
            },
        });
    }, [message, refreshAfterMutation]);

    const confirmRestore = useCallback((names: string[], name = "") => {
        const nextNames = normalizeNames(names);
        if (nextNames.length === 0) {
            message.warning("请选择回收站文件");
            return;
        }
        modal.confirm({
            title: "恢复文件",
            content: nextNames.length === 1 && name
                ? `确定恢复“${name}”吗？`
                : `确定恢复选中的 ${nextNames.length} 项吗？`,
            okText: "恢复",
            cancelText: "取消",
            mask: {closable: true},
            onOk: () => mutate(nextNames, "restore"),
        });
    }, [message, modal, mutate]);

    const confirmDelete = useCallback((names: string[], name = "") => {
        const nextNames = normalizeNames(names);
        if (nextNames.length === 0) {
            message.warning("请选择回收站文件");
            return;
        }
        modal.confirm({
            title: "彻底删除",
            content: nextNames.length === 1 && name
                ? `确定彻底删除“${name}”吗？此操作无法恢复。`
                : `确定彻底删除选中的 ${nextNames.length} 项吗？此操作无法恢复。`,
            okText: "删除",
            cancelText: "取消",
            okButtonProps: {danger: true},
            mask: {closable: true},
            onOk: () => mutate(nextNames, "delete"),
        });
    }, [message, modal, mutate]);

    const confirmClear = useCallback(() => modal.confirm({
        title: "清空回收站",
        content: "回收站中的全部文件将被彻底删除，此操作无法恢复。",
        okText: "清空",
        cancelText: "取消",
        okButtonProps: {danger: true},
        mask: {closable: true},
        onOk: () => clearRecycle({
            onSuccess: (response) => {
                message.success(response?.message || "回收站已清空");
                countRequestRef.current++;
                const emptyCount = {size: 0, file: 0, dir: 0};
                countRef.current = emptyCount;
                setCount(emptyCount);
                tableRef.current?.clearSelectedRows?.();
                tableRef.current?.refreshTableData?.();
                onMutation();
            },
            onFail: (response) => message.error(response?.message || "清空回收站失败"),
        }),
    }), [message, modal, onMutation]);

    const columns = useMemo<ProColumns<any>[]>(() => [
        {
            title: "名称",
            dataIndex: "name",
            width: 180,
            ellipsis: true,
            hideInSearch: true,
        },
        {
            title: "原路径",
            dataIndex: "path",
            width: 270,
            ellipsis: true,
            hideInSearch: true,
        },
        {
            title: "类型",
            dataIndex: "is_dir",
            width: 90,
            hideInSearch: true,
            render: (_: boolean, record: any) => (
                <Space size={6}>
                    <Icon type={record.is_dir ? "FolderOutlined" : "FileOutlined"}/>
                    {record.is_dir ? "文件夹" : "文件"}
                </Space>
            ),
        },
        {
            title: "大小",
            dataIndex: "size",
            width: 90,
            align: "right",
            hideInSearch: true,
            render: (size: number) => size > 0 ? formatFileSize(size) : "-",
        },
        {
            title: "删除时间",
            dataIndex: "update_time",
            width: 155,
            hideInSearch: true,
            render: (value: string) => formatFileTime(value),
        },
        {
            title: "操作",
            key: "option",
            width: 80,
            fixed: "right",
            hideInSearch: true,
            render: (_: unknown, record: any) => (
                <RecycleRowActions
                    id={record.id}
                    name={record.name}
                    onDelete={confirmDelete}
                    onRestore={confirmRestore}/>
            ),
        },
    ], [confirmDelete, confirmRestore]);

    const restoreSelected = useCallback(() => {
        confirmRestore((tableRef.current?.getSelectedRowKeys?.() || []).map(String));
    }, [confirmRestore]);

    const deleteSelected = useCallback(() => {
        confirmDelete((tableRef.current?.getSelectedRowKeys?.() || []).map(String));
    }, [confirmDelete]);

    const rightExtra = useMemo(() => [
        <Button
            key="restore"
            type="primary"
            ghost
            onClick={restoreSelected}>
            批量恢复
        </Button>,
        <Button
            key="delete"
            type="primary"
            danger
            ghost
            onClick={deleteSelected}>
            批量删除
        </Button>,
    ], [deleteSelected, restoreSelected]);

    const rowSelection = useMemo(() => ({
        fixed: true,
        rightExtra,
    }), [rightExtra]);

    const request = useCallback<NonNullable<ProTableProps["request"]>>(
        (params) => getData(params, recycleSort, getRecycleList),
        [],
    );

    const afterClose = useCallback(() => {
        countRequestRef.current++;
        countRef.current = undefined;
        setLoading(false);
        setCount(undefined);
        tableRef.current?.clearSelectedRows?.();
    }, []);

    const modalProps = useMemo<ModalProps>(() => ({
        footer: null,
        style: {
            top: screens.md ? 100 : 24,
            paddingBottom: screens.md ? 100 : 24,
        },
        styles: {
            container: {background: modalBackground},
            header: {background: modalBackground},
            body: {background: modalBackground, minHeight: 220},
        },
        mask: {closable: true},
        destroyOnHidden: false,
        afterClose,
    }), [afterClose, modalBackground, screens.md]);

    const modalTitle = useMemo(() => <Title>回收站</Title>, []);
    return (
        <ProModal
            ref={modalRef}
            title={modalTitle}
            width={910}
            modalProps={modalProps}>
            {loading || !count ? (
                <Flex
                    align="center"
                    justify="center"
                    role="status"
                    aria-label="正在加载回收站"
                    style={{height: screens.md ? 420 : "55dvh"}}>
                    <Spin size="large"/>
                </Flex>
            ) : (
                <Flex vertical gap={12}>
                    <RecycleSummary
                        compact={compact}
                        count={count}
                        fontSize={summaryFontSize}
                        onClear={confirmClear}/>
                    <ProTable
                        ref={tableRef}
                        rowKey="id"
                        search={false}
                        columns={columns}
                        defaultPage={1}
                        defaultPageSize={10}
                        rowSelection={rowSelection}
                        request={request}/>
                </Flex>
            )}
        </ProModal>
    );
}));

FileRecycleBin.displayName = "FileRecycleBin";

export default FileRecycleBin;
