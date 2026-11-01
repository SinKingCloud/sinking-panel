import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Button} from "antd";
import type {MenuProps} from "antd";
import {PageTable} from "sinking-antd";
import type {PageTableProps, PageTableRef} from "sinking-antd";
import ActionDropdown from "@/pages/components/ui/action-dropdown";
import {getOssList} from "@/service/api/oss";
import useEnum from "@/utils/enum";
import useStyles from "./styles";

interface StorageProps {
    active: boolean;
    revision: number;
    hero: NonNullable<PageTableProps["hero"]>;
    removing?: number;
    onOpen: (record: any) => void;
    onCreate: () => void;
    onEdit: (id: number) => void;
    onDelete: (record: any) => void;
}

const StorageContextMenu = React.createContext<any>(undefined);

const ContextMenuRow = React.forwardRef<HTMLTableRowElement, any>((rowProps, ref) => {
    const context = React.useContext(StorageContextMenu);
    const id = String(rowProps["data-row-key"] || "");
    const record = context?.recordsById.get(id);
    if (!record) return <tr {...rowProps} ref={ref}/>;
    return <ActionDropdown disabled={context.isDisabled(record)}
        open={context.openMenu?.source === "context" && context.openMenu.id === id}
        onOpenChange={(open) => context.changeMenuOpen(open, id, "context")}
        menu={{items: context.getMenuItems(record)}} trigger={["contextMenu"]} placement="bottomLeft" arrow={false}>
        <tr {...rowProps} ref={ref}/>
    </ActionDropdown>;
});

ContextMenuRow.displayName = "ContextMenuRow";

const tableComponents = {body: {row: ContextMenuRow}};

// 云端存储按页加载，切换视图时保留搜索条件和页码。
export default function Storage({active, revision, hero, removing, onOpen, onCreate, onEdit, onDelete}: StorageProps) {
    const {styles} = useStyles();
    const [enumData] = useEnum("secret");
    const table = useRef<PageTableRef<any>>(null);
    const previous = useRef({active, revision});
    const [records, setRecords] = useState<any[]>([]);
    const [openMenu, setOpenMenu] = useState<{id: string; source: "button" | "context"}>();
    const recordsById = useMemo(() => new Map(records.map((record) => [String(record.id), record])), [records]);

    const request = useCallback<NonNullable<PageTableProps<any>["request"]>>(async ({current, pageSize, keyword}) => {
        // 翻页或刷新时停用旧行，只有 PageTable 接受的响应才能恢复操作。
        setRecords([]);
        setOpenMenu(undefined);
        const response = await getOssList({body: {page: current, page_size: pageSize, keyword}});
        if (response?.code !== 200) throw new Error(response?.message || "获取存储列表失败");
        return {data: response?.data?.list || [], total: Number(response?.data?.total || 0), success: true};
    }, []);

    useEffect(() => {
        // 恢复显示时 PageTable 自行加载；仅可见期间的配置变更需要主动刷新。
        if (active && previous.current.active && previous.current.revision !== revision) table.current?.reload();
        previous.current = {active, revision};
    }, [active, revision]);
    useEffect(() => {
        setOpenMenu((current) => active && current && recordsById.has(current.id) && current.id !== String(removing)
            ? current : undefined);
    }, [active, recordsById, removing]);

    const isDisabled = useCallback((record: any) => !active || removing === record.id || !recordsById.has(String(record.id)),
        [active, recordsById, removing]);
    const changeMenuOpen = useCallback((open: boolean, id: string, source: "button" | "context") => {
        setOpenMenu((current) => open ? {id, source} : current?.id === id && current.source === source ? undefined : current);
    }, []);
    const getMenuItems = useCallback((record: any): MenuProps["items"] => {
        const disabled = isDisabled(record);
        return [
            {key: "open", label: "管理", disabled, onClick: () => onOpen(record)},
            {key: "edit", label: "修改", disabled, onClick: () => onEdit(record.id)},
            {key: "delete", label: "删除", danger: true, disabled, onClick: () => onDelete(record)},
        ];
    }, [isDisabled, onDelete, onEdit, onOpen]);
    const contextMenu = useMemo(() => ({recordsById, openMenu, isDisabled, changeMenuOpen, getMenuItems}),
        [recordsById, openMenu, isDisabled, changeMenuOpen, getMenuItems]);

    const columns = useMemo<PageTableProps<any>["columns"]>(() => [
        {title: "存储名称", dataIndex: "name", width: 230, ellipsis: true,
            render: (name: string, record: any) => <button type="button" className="storage-name" title={name}
                disabled={isDisabled(record)} onClick={() => onOpen(record)}>{name}</button>},
        {title: "密钥名称", dataIndex: "secret_name", width: 190, ellipsis: true,
            render: (name: string) => name || "—"},
        {title: "厂商", dataIndex: "provider", width: 130, ellipsis: true,
            render: (provider: number | null) => provider == null ? "—" : enumData?.provider?.[provider] || "未知厂商"},
        {title: "更新时间", dataIndex: "update_time", width: 180, ellipsis: true},
        {title: "操作", key: "action", width: 90, fixed: "right", className: "action-cell",
            render: (_: unknown, record: any) => <ActionDropdown
                open={openMenu?.source === "button" && openMenu.id === String(record.id)}
                onOpenChange={(open) => changeMenuOpen(open, String(record.id), "button")}
                disabled={isDisabled(record)} menu={{items: getMenuItems(record)}}>
                <Button size="small" disabled={isDisabled(record)} loading={removing === record.id}>操作</Button>
            </ActionDropdown>},
    ], [changeMenuOpen, enumData, getMenuItems, isDisabled, onOpen, openMenu, removing]);

    return <PageTable ref={table} hero={hero} ariaLabel="云端存储" className={styles.table}
        rowKey="id" columns={columns} rowSelection={false} scroll={{x: 820}}
        components={tableComponents} tableRender={(children) => <StorageContextMenu.Provider value={contextMenu}>{children}</StorageContextMenu.Provider>}
        rowClassName={(record) => openMenu?.id === String(record.id) ? "action-menu-open" : ""}
        request={request} requestEnabled={active} onLoad={setRecords} pagination={{unit: "个"}}
        toolbar={{
            search: {placeholder: "搜索存储名称", debounce: 300},
            refresh: true,
            actions: [{key: "create", type: "button", label: "添加存储", icon: "PlusOutlined", onClick: onCreate}],
        }}/>;
}
