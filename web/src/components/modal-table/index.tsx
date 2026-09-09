import React, {forwardRef, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Grid, Modal, Spin} from "antd";
import type {ModalProps} from "antd";
import {createStyles} from "antd-style";
import {Title, useTheme} from "sinking-antd";
import Table from "@/components/table";
import type {TableProps, TableRef} from "@/components/table";

export type * from "@/components/table";

/** 工具栏、筛选和请求参数复用 Table API，参见本目录 README。 */
export interface ModalTableProps<RecordType extends object = any> extends Omit<TableProps<RecordType>, "title"> {
    title?: React.ReactNode;
    width?: ModalProps["width"];
    open?: boolean;
    onOpenChange?: (open: boolean) => void;
    onCancel?: ModalProps["onCancel"];
    modalProps?: Omit<ModalProps, "open" | "children">;
    /** modal：整个弹窗随页面滚动；body：只滚动表格区域。 */
    scrollMode?: "modal" | "body";
    minHeight?: number | string;
    bodyHeight?: number | string;
}

export interface ModalTableRef<RecordType extends object = any> extends TableRef<RecordType> {
    open: () => void;
    show: () => void;
    close: () => void;
    hide: () => void;
}

const useStyles = createStyles(({css, token}: any) => ({
    modal: css`
        .ant-modal { max-width: calc(100vw - 24px); outline: none; }
        .ant-modal-body { padding: 0; }
        .ui-table-root {
            min-width: 0;
            max-width: none;
            min-height: var(--modal-table-min-height);
            display: flex;
            flex-direction: column;
        }
        .ui-table-workspace {
            min-width: 0;
            display: flex;
            flex: 1 0 auto;
            flex-direction: column;
            overflow: visible;
            border-radius: 0;
            background: transparent;
            box-shadow: none;
        }
        .ui-table-workspace > * { flex: none; }
        .ui-table-workspace > .ui-table-data {
            min-width: 0;
            padding: 0;
            display: flex;
            flex: 1 0 auto;
            flex-direction: column;
        }
        .ui-table-data > [role="status"] { flex: 1; }
        .ui-table-data > .ant-table-wrapper { flex: none; }
        /* 固定列定位变化不参与排序表头的过渡动画。 */
        .ui-table-data .ant-table-cell:is(.ant-table-cell-fix, .ant-table-cell-fix-left, .ant-table-cell-fix-right) {
            transition-property: background-color, color, border-color, box-shadow;
        }
        .modal-table-loading { align-items: center; justify-content: center; }
        .ui-table-content-bar { padding-inline: 0; }
        .ui-table-toolbar {
            min-height: 0;
            padding: 4px 0 12px;
            background: transparent;
        }
        .ui-table-toolbar > .ant-input-affix-wrapper { width: 240px; flex: 0 1 240px; }
        .ui-table-pagination { flex: none; }
        .ui-table-pagination > .ant-card { border: 0; background: ${token.colorFillQuaternary}; box-shadow: none; }
        .ui-table-pagination .ant-pagination { flex-wrap: wrap; gap: 6px 0; }

        .modal-table-scroll-body {
            height: var(--modal-table-body-height);
        }
        .modal-table-scroll-body .ui-table-workspace { min-height: 0; flex: 1 1 auto; }
        .modal-table-scroll-body .ui-table-workspace > .ui-table-data {
            min-height: 0;
            flex: 1 1 auto;
            overflow: auto;
            overscroll-behavior: contain;
            scrollbar-width: thin;
        }
        @container data-table-workspace (max-width: 680px) {
            .ui-table-toolbar > .ant-input-affix-wrapper { width: 100%; flex-basis: 100%; }
        }
    `,
}));

const ModalTableContent = <RecordType extends object = any>({tableRef, ...props}: TableProps<RecordType> & {
    tableRef: React.RefObject<TableRef<RecordType> | null>;
}) => {
    const {message} = App.useApp();
    const reload = Boolean(!props.request && props.onReload && !props.manualRequest && props.requestEnabled !== false);
    const [loadState, setLoadState] = useState<"pending" | "settled" | "ready">("pending");
    const loading = typeof props.loading === "boolean" ? props.loading : Boolean(props.loading?.spinning);
    const waiting = reload && loadState !== "ready";

    useEffect(() => {
        if (loadState === "settled" && !loading) setLoadState("ready");
    }, [loadState, loading]);

    useEffect(() => {
        if (!reload) return;
        let active = true;
        void (async () => {
            try {
                await props.onReload?.();
            } catch (reason) {
                if (!active) return;
                const error = reason instanceof Error ? reason : new Error("获取列表失败");
                if (props.onRequestError) props.onRequestError(error);
                else message.error(error.message);
            } finally {
                if (active) setLoadState((state) => state === "ready" ? state : "settled");
            }
        })();
        return () => { active = false; };
    }, [reload, message]);

    return <Table<RecordType>
        {...props}
        ref={tableRef}
        dataSource={waiting ? [] : props.dataSource}
        loading={waiting ? true : props.loading}
        empty={waiting ? true : props.empty}
        pagination={waiting && props.pagination !== false ? {...props.pagination, total: 0} : props.pagination}/>;
};

const ModalTable = <RecordType extends object = any>({
    title, width, open: controlledOpen, onOpenChange, onCancel, modalProps = {}, scrollMode = "modal",
    minHeight, bodyHeight, rootClassName = "", requestEnabled, ...tableProps
}: ModalTableProps<RecordType>, ref: React.ForwardedRef<ModalTableRef<RecordType>>) => {
    const screens = Grid.useBreakpoint();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles();
    const [internalOpen, setInternalOpen] = useState(false);
    const [opened, setOpened] = useState(false);
    const open = controlledOpen ?? internalOpen;
    const tableRef = useRef<TableRef<RecordType>>(null);
    const afterOpenChange = useRef(modalProps.afterOpenChange);
    afterOpenChange.current = modalProps.afterOpenChange;

    useEffect(() => {
        if (!open) setOpened(false);
        else if (opened) afterOpenChange.current?.(true);
    }, [open, opened]);

    const changeOpen = (next: boolean) => {
        setInternalOpen(next);
        onOpenChange?.(next);
    };
    useImperativeHandle(ref, () => ({
        open: () => changeOpen(true), show: () => changeOpen(true),
        close: () => changeOpen(false), hide: () => changeOpen(false),
        reload: () => tableRef.current?.reload(),
        refreshTableData: () => tableRef.current?.reload(),
        resetTableData: () => tableRef.current?.resetTableData(),
        getTableData: () => tableRef.current?.getTableData() || [],
        getSelectedRowKeys: () => tableRef.current?.getSelectedRowKeys() || [],
        getSelectedRows: () => tableRef.current?.getSelectedRows() || [],
        setSelectedRowKeys: (keys) => tableRef.current?.setSelectedRowKeys(keys),
        clearSelectedRows: () => tableRef.current?.clearSelectedRows(),
        allSelectedRow: () => tableRef.current?.allSelectedRow(),
        invertSelectedRow: () => tableRef.current?.invertSelectedRow(),
        scrollToTop: () => tableRef.current?.scrollToTop(),
    }));
    const cssSize = (value: number | string) => typeof value === "number" ? `${value}px` : value;

    return <Modal
        {...modalProps}
        open={open}
        destroyOnHidden={modalProps.destroyOnHidden ?? true}
        title={typeof title === "string" ? <Title>{title}</Title> : title}
        width={width ?? (compact ? 760 : 800)}
        footer={modalProps.footer ?? null}
        mask={modalProps.mask ?? {closable: true}}
        focusable={modalProps.focusable ?? {focusTriggerAfterClose: false}}
        rootClassName={`${styles.modal} ${modalProps.rootClassName || ""}`}
        style={{top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, ...modalProps.style}}
        onCancel={(event) => {
            changeOpen(false);
            onCancel?.(event);
            modalProps.onCancel?.(event);
        }}
        afterOpenChange={(visible) => {
            setOpened(visible && open);
            if (!visible) afterOpenChange.current?.(false);
        }}>
        <div style={{
            "--modal-table-min-height": cssSize(minHeight ?? (compact ? 440 : 480)),
            "--modal-table-body-height": cssSize(bodyHeight ?? (screens.md ? "calc(100dvh - 280px)" : "calc(100dvh - 148px)")),
        } as React.CSSProperties}>
            {open && opened ? <ModalTableContent<RecordType>
                {...tableProps}
                tableRef={tableRef}
                rootClassName={`modal-table-scroll-${scrollMode} ${rootClassName}`}
                requestEnabled={requestEnabled}/>
                : <div className={`ui-table-root modal-table-loading modal-table-scroll-${scrollMode}`} role="status">
                    <Spin/>
                </div>}
        </div>
    </Modal>;
};

export default forwardRef(ModalTable) as <RecordType extends object = any>(props: ModalTableProps<RecordType> & React.RefAttributes<ModalTableRef<RecordType>>) => React.ReactElement;
