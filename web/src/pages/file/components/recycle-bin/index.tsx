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
import {App, Button, Space, Tooltip} from "antd";
import type {TableColumnsType} from "antd";
import {Icon, Title, useTheme} from "sinking-antd";
import ModalTable from "@/components/modal-table";
import type {ModalTableRef} from "@/components/modal-table";
import Dropdown from "@/pages/components/stable-dropdown";
import {
    clearRecycle,
    deleteRecycleItem,
    getRecycleCount,
    getRecycleList,
    restoreRecycleItem,
} from "@/service/api/file";
import {getData} from "@/utils/page";
import {formatFileSize, formatFileTime} from "../../utils";
import useStyles from "./styles";

export interface FileRecycleBinRef {
    open: () => void;
}

interface FileRecycleBinProps {
    onMutation: () => void;
}

type RecycleAction = "restore" | "delete";

interface RecycleRowActionsProps {
    id: string;
    name: string;
    onDelete: (names: string[], name?: string) => void;
    onRestore: (names: string[], name?: string) => void;
}

const recycleSort = {update_time: "descend"};

const normalizeNames = (names: string[]) => Array.from(new Set((names || []).filter(Boolean)));

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
            <Button size="small" autoInsertSpace={false}>操作</Button>
        </Dropdown>
    );
});

RecycleRowActions.displayName = "RecycleRowActions";

const FileRecycleBin = memo(forwardRef<FileRecycleBinRef, FileRecycleBinProps>(({onMutation}, ref) => {
    const {message, modal} = App.useApp();
    const appTheme = useTheme();
    const compact = Boolean(appTheme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const tableRef = useRef<ModalTableRef | null>(null);
    const countRequestRef = useRef(0);
    const [loading, setLoading] = useState(false);
    const [count, setCount] = useState<any>();

    useEffect(() => () => {
        countRequestRef.current += 1;
    }, []);

    const loadCount = useCallback(async () => {
        const requestId = ++countRequestRef.current;
        setLoading(true);
        try {
            const response = await getRecycleCount();
            if (requestId !== countRequestRef.current) return;
            if (response?.code !== 200 || !response.data) {
                message.error(response?.message || "获取回收站信息失败");
                return;
            }
            setCount(response.data);
        } catch {
            if (requestId === countRequestRef.current) {
                message.error("获取回收站信息失败");
            }
        } finally {
            if (requestId === countRequestRef.current) setLoading(false);
        }
    }, [message]);

    const open = useCallback(() => {
        tableRef.current?.open();
        void loadCount();
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
                setCount(emptyCount);
                setLoading(false);
                tableRef.current?.clearSelectedRows?.();
                tableRef.current?.refreshTableData?.();
                onMutation();
            },
            onFail: (response) => message.error(response?.message || "清空回收站失败"),
        }),
    }), [message, modal, onMutation]);

    const columns = useMemo<TableColumnsType<any>>(() => [
        {
            title: "名称",
            dataIndex: "name",
            width: 180,
            ellipsis: true,
        },
        {
            title: "原路径",
            dataIndex: "path",
            width: 270,
            ellipsis: {showTitle: false},
            render: (path: string) => <Tooltip title={path} placement="topLeft"><span>{path}</span></Tooltip>,
        },
        {
            title: "类型",
            dataIndex: "is_dir",
            width: 100,
            render: (_: boolean, record: any) => (
                <Space size={6} style={{whiteSpace: "nowrap"}}>
                    <Icon type={record.is_dir ? "FolderOutlined" : "FileOutlined"}/>
                    {record.is_dir ? "文件夹" : "文件"}
                </Space>
            ),
        },
        {
            title: "大小",
            dataIndex: "size",
            width: 110,
            align: "right",
            render: (size: number) => (
                <span style={{whiteSpace: "nowrap"}}>{size > 0 ? formatFileSize(size) : "-"}</span>
            ),
        },
        {
            title: "删除时间",
            dataIndex: "update_time",
            width: 190,
            render: (value: string) => (
                <span style={{whiteSpace: "nowrap", fontVariantNumeric: "tabular-nums"}}>{formatFileTime(value)}</span>
            ),
        },
        {
            title: "操作",
            key: "option",
            width: 80,
            className: "action-cell",
            render: (_: unknown, record: any) => (
                <RecycleRowActions
                    id={record.id}
                    name={record.name}
                    onDelete={confirmDelete}
                    onRestore={confirmRestore}/>
            ),
        },
    ], [confirmDelete, confirmRestore]);

    const request = useCallback(
        ({page, pageSize}: {page: number; pageSize: number}) => getData({current: page, pageSize}, recycleSort, getRecycleList),
        [],
    );

    const close = useCallback(() => {
        countRequestRef.current++;
        setLoading(false);
        setCount(undefined);
        tableRef.current?.clearSelectedRows?.();
    }, []);

    return (
        <ModalTable
            ref={tableRef}
            title={<Title>回收站</Title>}
            width={910}
            minHeight={compact ? 440 : 480}
            scrollMode="modal"
            onCancel={close}
            rowKey="id"
            columns={columns}
            scroll={{x: 978}}
            defaultPageSize={10}
            request={request}
            onRequestError={() => message.error("获取回收站文件失败")}
            toolbar={{
                left: <div className={styles.summary} role="group" aria-label="回收站统计">
                    <span className="space-used">
                        <span>占用</span>
                        <strong>{count ? formatFileSize(count.size) : "—"}</strong>
                    </span>
                    <span className="entry-count"><strong>{count?.file ?? "—"}</strong>个文件</span>
                    <span className="entry-count"><strong>{count?.dir ?? "—"}</strong>个文件夹</span>
                </div>,
                actions: [{
                    key: "clear",
                    type: "button",
                    label: "清空",
                    icon: "DeleteOutlined",
                    danger: true,
                    disabled: !count || count.file + count.dir === 0,
                    onClick: confirmClear,
                }],
                refresh: {
                    loading,
                    onClick: () => {
                        void tableRef.current?.reload();
                        void loadCount();
                    },
                },
            }}
            rowSelection={{
                columnWidth: 48,
                actions: (keys) => [
                    {key: "restore", type: "button", label: "恢复", icon: "RollbackOutlined", onClick: () => confirmRestore(keys.map(String))},
                    {key: "delete", type: "button", label: "删除", icon: "DeleteOutlined", danger: true, onClick: () => confirmDelete(keys.map(String))},
                ],
            }}/>
    );
}));

FileRecycleBin.displayName = "FileRecycleBin";

export default FileRecycleBin;
