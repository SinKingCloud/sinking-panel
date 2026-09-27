import {forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {Alert, App, Button, ConfigProvider, Descriptions, Empty, Form, Grid, Segmented, Spin, Tag, Tooltip, Typography} from "antd";
import {ModalTable, ProModal, Title, useTheme} from "sinking-antd";
import type {ModalTableProps, ModalTableRef} from "sinking-antd";
import ActionDropdown from "@/pages/components/action-dropdown";
import FilePicker from "@/pages/components/file-picker";
import {getContainerImage, getContainerImages, importContainerImage, removeContainerImage} from "@/service/api/container";

import {getParams} from "@/utils/page";
import useStyles from "./styles";

const ImagesManager = forwardRef<any, any>(({onRun}, ref) => {
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles, cx} = useStyles();
    const tableRef = useRef<ModalTableRef<any> | null>(null);
    const confirmRef = useRef<{destroy: () => void} | null>(null);
    const mountedRef = useRef(true);
    const busyRef = useRef(new Set<string>());
    const importingRef = useRef(false);
    const detailRequestRef = useRef(0);
    const [form] = Form.useForm<{image_path: string}>();
    const [open, setOpen] = useState(false);
    const [session, setSession] = useState(0);
    const [busy, setBusy] = useState(new Set<string>());
    const [importOpen, setImportOpen] = useState(false);
    const [importing, setImporting] = useState(false);
    const [detailId, setDetailId] = useState("");
    const [detailTab, setDetailTab] = useState<"info" | "config" | "env">("info");
    const [detail, setDetail] = useState<any>(null);
    const [detailLoading, setDetailLoading] = useState(false);
    const [detailError, setDetailError] = useState("");

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            detailRequestRef.current += 1;
            confirmRef.current?.destroy();
        };
    }, []);

    const showDetail = useCallback(async (id: string) => {
        const requestId = ++detailRequestRef.current;
        setDetailId(id);
        setDetailTab("info");
        setDetail(null);
        setDetailError("");
        setDetailLoading(true);
        try {
            const response = await getContainerImage({body: {id}});
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "获取镜像详情失败");
            if (mountedRef.current && detailRequestRef.current === requestId) setDetail(response.data);
        } catch (error) {
            if (mountedRef.current && detailRequestRef.current === requestId) {
                setDetailError(error instanceof Error ? error.message : "获取镜像详情失败");
            }
        } finally {
            if (mountedRef.current && detailRequestRef.current === requestId) setDetailLoading(false);
        }
    }, []);

    useImperativeHandle(ref, () => ({
        open: () => {
            if (busyRef.current.size || importingRef.current) return;
            confirmRef.current?.destroy();
            setSession((value) => value + 1);
            setOpen(true);
        },
        openDetail: showDetail,
    }), [showDetail]);

    const importImage = async (values: {image_path: string}) => {
        if (importingRef.current) return;
        importingRef.current = true;
        setImporting(true);
        try {
            const response = await importContainerImage({body: {image_path: values.image_path.trim()}});
            if (!mountedRef.current) return;
            if (response?.code !== 200) throw new Error(response?.message || "导入镜像失败");
            setImportOpen(false);
            tableRef.current?.reload();
            message.success(response.message || "导入成功");
        } catch (error) {
            if (mountedRef.current) message.error(error instanceof Error ? error.message : "导入镜像失败");
        } finally {
            importingRef.current = false;
            if (mountedRef.current) setImporting(false);
        }
    };

    const confirmRemove = useCallback((records: any[]) => {
        const targets = records.filter((image) => !busyRef.current.has(image.id));
        if (!targets.length) return;
        confirmRef.current?.destroy();
        const confirmation = modal.confirm({
            title: targets.length === 1 ? "删除镜像" : `删除 ${targets.length} 个镜像`,
            content: targets.length === 1
                ? `确定删除镜像“${targets[0].name || targets[0].id}”吗？被容器使用的镜像无法删除。`
                : "确定删除所选镜像吗？被容器使用的镜像会保留，并显示失败原因。",
            zIndex: 1150,
            okText: "删除",
            cancelText: "取消",
            okButtonProps: {danger: true},
            onOk: async () => {
                if (!mountedRef.current) return;
                const pending = targets.filter((image) => !busyRef.current.has(image.id));
                if (!pending.length) return;
                pending.forEach((image) => busyRef.current.add(image.id));
                setBusy(new Set(busyRef.current));
                try {
                    const response = await removeContainerImage({body: {ids: pending.map((image) => image.id)}});
                    if (!mountedRef.current) return;
                    if (response?.code !== 200) throw new Error(response?.message || "删除镜像失败");
                    const removed = new Set(pending.map((image) => image.id));
                    tableRef.current?.setSelectedRowKeys((tableRef.current?.getSelectedRowKeys() || [])
                        .filter((key) => !removed.has(String(key))));
                    message.success(`删除成功${pending.length > 1 ? ` ${pending.length} 项` : ""}`);
                } catch (error) {
                    if (!mountedRef.current) return;
                    const reason = error instanceof Error ? error.message : "删除镜像失败";
                    if (pending.length === 1) message.error(reason);
                    else modal.warning({
                        title: "镜像删除失败",
                        content: <div style={{maxHeight: 280, overflow: "auto", whiteSpace: "pre-wrap", overflowWrap: "anywhere"}}>{reason}</div>,
                        zIndex: 1150,
                        okText: "知道了",
                    });
                } finally {
                    pending.forEach((image) => busyRef.current.delete(image.id));
                    if (mountedRef.current) {
                        setBusy(new Set(busyRef.current));
                        // 部分镜像删除失败时也刷新，移除已成功删除的数据和选中项。
                        tableRef.current?.reload();
                    }
                }
            },
            afterClose: () => {
                if (confirmRef.current === confirmation) confirmRef.current = null;
            },
        });
        confirmRef.current = confirmation;
    }, [message, modal]);

    const columns = useMemo<ModalTableProps<any>["columns"]>(() => [
        {
            title: "镜像名称", dataIndex: "name", key: "name", width: 205, sorter: true,
            render: (name: string, image: any) => <button type="button" className="software-name"
                title={name || image.id} onClick={() => void showDetail(image.id)}>{name || "未命名镜像"}</button>,
        },
        {
            title: "标签", key: "tags", width: 210,
            render: (_, image) => <Tooltip title={image.tags?.join("、") || "无标签"}>
                <span className="software-text">{image.tags?.join("、") || "—"}</span>
            </Tooltip>,
        },
        {
            title: "ID", dataIndex: "id", key: "id", width: 140, sorter: true,
            render: (id: string) => <Tooltip title={id}><span className="software-text software-muted">{id.replace(/^sha256:/, "").slice(0, 12)}</span></Tooltip>,
        },
        {
            title: "创建时间", dataIndex: "create_time", key: "create_time", width: 165, sorter: true,
            render: (value: string) => <span className="software-time">{value || "—"}</span>,
        },
        {
            title: "操作", key: "action", width: 82, fixed: "right", className: "action-cell",
            render: (_, image) => <ActionDropdown disabled={busy.has(image.id)}
                menu={{items: [
                    {key: "info", label: "查看详情", onClick: () => void showDetail(image.id)},
                    {key: "run", label: "创建容器", disabled: busy.size > 0 || importing, onClick: () => {
                        if (busyRef.current.size || importingRef.current) return;
                        setOpen(false);
                        onRun(image);
                    }},
                    {key: "remove", label: "删除镜像", danger: true, onClick: () => confirmRemove([image])},
                ]}}>
                <Button size="small" loading={busy.has(image.id)} aria-label={`操作镜像 ${image.name || image.id}`}>操作</Button>
            </ActionDropdown>,
        },
    ], [busy, confirmRemove, importing, onRun, showDetail]);

    return <>
        <ModalTable<any>
            key={session}
            ref={tableRef}
            title={<Title>镜像管理</Title>}
            open={open}
            onOpenChange={(visible) => {
                if (!visible && (busyRef.current.size || importingRef.current)) return;
                if (!visible) confirmRef.current?.destroy();
                setOpen(visible);
            }}
            width={compact ? 800 : 910}
            minHeight={compact ? 420 : 460}
            scrollMode="modal"
            rowKey="id"
            columns={columns}
            className={styles.table}
            scroll={{x: 840}}
            requestEnabled={open}
            request={async (params, sort) => {
                const response = await getContainerImages({body: getParams(params, sort)});
                if (response?.code !== 200) throw new Error(response?.message || "获取镜像列表失败");
                return {data: response.data?.list ?? [], total: response.data?.total ?? 0, success: true};
            }}
            defaultPageSize={20}
            search={{
                placeholder: "搜索镜像", maxLength: 100,
                fields: [
                    {value: "keyword", label: "全部", placeholder: "搜索名称、ID 或标签"},
                    {value: "name", label: "名称", placeholder: "搜索镜像名称", maxLength: 255},
                    {value: "id", label: "ID", placeholder: "输入完整镜像 ID", maxLength: 255},
                    {value: "tag", label: "标签", placeholder: "输入完整镜像标签", maxLength: 255},
                ],
            }}
            filters={[
                {type: "date", name: "create_time", label: "创建时间筛选"},
            ]}
            toolbar={{
                refresh: {ariaLabel: "刷新镜像列表", disabled: importing || busy.size > 0},
                actions: [{key: "import", type: "button", label: "导入", icon: "ImportOutlined", disabled: importing || busy.size > 0,
                    onClick: () => { form.resetFields(); setImportOpen(true); }}],
            }}
            rowSelection={{
                columnWidth: 44,
                clearDisabled: busy.size > 0,
                getCheckboxProps: (image) => ({disabled: busy.has(image.id)}),
                actions: (_, rows) => [{key: "remove", type: "button", label: "删除", danger: true,
                    disabled: busy.size > 0, onClick: () => confirmRemove(rows)}],
            }}
            modalProps={{
                zIndex: 1050,
                closable: !importing && busy.size === 0,
                keyboard: !importing && busy.size === 0,
                mask: {closable: !importing && busy.size === 0},
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"},
            }}
        />
        <ProModal title={<Title>导入镜像</Title>} width={350} onOk={form.submit}
            onCancel={() => { if (!importingRef.current) setImportOpen(false); }}
            modalProps={{
                open: importOpen, zIndex: 1100, okText: "导入", cancelText: "取消", confirmLoading: importing,
                cancelButtonProps: {disabled: importing}, closable: !importing, keyboard: !importing,
                mask: {closable: !importing}, forceRender: true, style: {maxWidth: "calc(100vw - 24px)"},
            } as any}>
            <ConfigProvider input={{variant: "filled"}}>
                <Form form={form} layout="vertical" onFinish={importImage} disabled={importing}>
                    <Form.Item name="image_path" label="镜像路径"
                        rules={[{required: true, whitespace: true, message: "请输入或选择镜像路径"}, {
                            validator: (_, value) => !value || String(value).trim().startsWith("/")
                                ? Promise.resolve() : Promise.reject(new Error("请输入服务器上的绝对路径")),
                        }]}>
                        <FilePicker mode="file" placeholder="请输入或选择镜像路径" autoComplete="off"
                            disabled={importing} buttonProps={{color: "default", variant: "filled"}}/>
                    </Form.Item>
                </Form>
            </ConfigProvider>
        </ProModal>
        <ProModal title={<Title>镜像详情</Title>} width={compact ? 500 : 560}
            onCancel={() => { detailRequestRef.current += 1; setDetailId(""); setDetail(null); }}
            modalProps={{open: Boolean(detailId), zIndex: 1100, destroyOnHidden: true, footer: null, mask: {closable: true},
                style: {top: screens.md ? 80 : 20, paddingBottom: 40, maxWidth: "calc(100vw - 24px)"}} as any}>
            <div className={cx(styles.detail, styles.imageDetail)}>
                {detailLoading ? <div className="detail-loading"><Spin/></div>
                    : detailError ? <Alert type="error" showIcon title={detailError}
                        action={<Button size="small" color="default" variant="filled" onClick={() => void showDetail(detailId)}>重试</Button>}/>
                        : detail && <>
                            <Segmented<"info" | "config" | "env"> className="detail-switch" aria-label="镜像详情内容"
                                size="middle" value={detailTab} onChange={setDetailTab}
                                options={[{value: "info", label: "基本信息"}, {value: "config", label: "运行配置"}, {value: "env", label: "环境变量"}]}/>
                            <div className="detail-content">
                                {detailTab === "info" ? <dl className="detail-overview">
                                    <div>
                                        <dt>镜像名称</dt>
                                        <dd>{detail.name || "未命名镜像"}</dd>
                                    </div>
                                    <div>
                                        <dt>创建时间</dt>
                                        <dd>{detail.create_time || "—"}</dd>
                                    </div>
                                    <div className="detail-overview-wide">
                                        <dt>镜像标签</dt>
                                        <dd className="detail-tags">{detail.tags?.length
                                            ? detail.tags.map((tag: string) => <Tag key={tag} variant="filled">{tag}</Tag>)
                                            : "无标签"}</dd>
                                    </div>
                                    <div className="detail-overview-wide">
                                        <dt>镜像 ID</dt>
                                        <dd><Typography.Text copyable={{text: detail.id}}>{detail.id}</Typography.Text></dd>
                                    </div>
                                </dl> : detailTab === "config" ? <div className="detail-config">
                                    <Descriptions column={{xs: 1, sm: 2}} layout={screens.sm ? "vertical" : "horizontal"}
                                        size="small" colon={false}
                                        styles={screens.sm ? undefined : {label: {flex: "0 0 84px", whiteSpace: "nowrap"}, content: {minWidth: 0}}}
                                        items={[
                                        {key: "user", label: "运行用户", children: detail.user || "root"},
                                        {key: "working_dir", label: "工作目录", children: detail.working_dir || "/"},
                                        {key: "ports", label: "声明端口", children: detail.exposed_ports?.join("、") || "未声明"},
                                        {key: "volumes", label: "数据卷", children: detail.volumes?.join("\n") || "未声明"},
                                    ]}/>
                                    {[
                                        {key: "entrypoint", label: "入口程序", args: detail.entrypoint},
                                        {key: "command", label: "默认命令", args: detail.command},
                                    ].map(({key, label, args}) => {
                                        const value = args?.map((arg: string) => JSON.stringify(arg)).join(" ");
                                        return <section className="detail-field detail-command" key={key}>
                                            <span className="detail-label">{label}</span>
                                            <div className="detail-code">
                                                <Typography.Text copyable={value ? {text: value} : false}>{value || "未设置"}</Typography.Text>
                                            </div>
                                        </section>;
                                    })}
                                </div> : detail.env?.length
                                    ? <dl className="detail-config">{detail.env.map((line: string, index: number) => {
                                        const separator = line.indexOf("=");
                                        const name = separator < 0 ? line : line.slice(0, separator);
                                        const value = separator < 0 ? "" : line.slice(separator + 1);
                                        return <div className="detail-field detail-code" key={index}>
                                            <dt className="detail-label">{name}</dt>
                                            <dd style={{margin: 0, minWidth: 0}}>
                                                <Typography.Text copyable={{text: line, tooltips: ["复制", "已复制"]}}>{value || <span className="detail-label">空值</span>}</Typography.Text>
                                            </dd>
                                        </div>;
                                    })}</dl>
                                    : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>}
                            </div>
                        </>}
            </div>
        </ProModal>
    </>;
});

ImagesManager.displayName = "ImagesManager";

export default ImagesManager;
