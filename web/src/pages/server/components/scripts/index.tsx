import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Empty, Input, Spin, Tooltip} from "antd";
import {Icon, Title} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import ActionDropdown from "@/pages/components/action-dropdown";
import TypeManager from "@/pages/components/type-manager";
import type {TypeManagerRef} from "@/pages/components/type-manager";

import {deleteScript, getScriptInfo} from "@/service/api/script";
import {getTypeList} from "@/service/api/type";
import useScripts from "../../hooks/scripts";
import ScriptForm from "../script-form";

const manageTypeKey = "__manage_script_types__";

const fallbackCopyText = (value: string) => {
    const activeElement = document.activeElement as HTMLElement | null;
    let textarea: HTMLTextAreaElement | undefined;
    try {
        textarea = document.createElement("textarea");
        textarea.value = value;
        textarea.setAttribute("readonly", "");
        textarea.style.position = "fixed";
        textarea.style.top = "0";
        textarea.style.left = "0";
        textarea.style.opacity = "0";
        textarea.style.pointerEvents = "none";
        document.body.appendChild(textarea);
        textarea.focus({preventScroll: true});
        textarea.select();
        textarea.setSelectionRange(0, value.length);
        return document.execCommand("copy");
    } catch {
        return false;
    } finally {
        textarea?.remove();
        try {
            activeElement?.focus({preventScroll: true});
        } catch {
            // 原控件可能已经卸载，不影响复制结果。
        }
    }
};

const Scripts = ({styles, collapsed, connected, selectedId, onCollapsedChange, onInsert}: any) => {
    const {message, modal} = App.useApp();
    const instanceModule = selectedId === undefined || selectedId === null ? undefined : `ssh_${selectedId}`;
    const list = useScripts(instanceModule || "ssh");
    const formRef = useRef<any>({} as any);
    const typeManagerRef = useRef<TypeManagerRef>({} as TypeManagerRef);
    const typeRequestRef = useRef(0);
    const insertRequestRef = useRef(0);
    const copyRequestRef = useRef(0);
    const copyModalRef = useRef<{destroy: () => void} | null>(null);
    const [typeItems, setTypeItems] = useState<any[]>([]);
    const [insertingId, setInsertingId] = useState<number>();
    const [copyingId, setCopyingId] = useState<number>();
    const [deletingId, setDeletingId] = useState<number>();
    const typeMap = useMemo(() => Object.fromEntries(
        typeItems.map((item) => [String(item.id), item.name]),
    ), [typeItems]);
    const typeOptions = useMemo(() => [
        {label: "全部分类", value: "0"},
        ...typeItems.map((item) => ({label: item.name, value: String(item.id)})),
    ], [typeItems]);
    const typeMenuItems = useMemo(() => [
        ...typeOptions.map((item) => ({key: item.value, label: item.label})),
        {type: "divider" as const},
        {key: manageTypeKey, label: "分类管理", icon: <Icon type="SettingOutlined"/>},
    ], [typeOptions]);
    const activeTypeName = typeOptions.find((item) => item.value === list.typeId)?.label || "全部分类";

    const loadTypes = useCallback(async () => {
        const requestId = ++typeRequestRef.current;
        const response = await getTypeList({body: {module: instanceModule || "ssh", order_by_field: "sort", order_by_type: "asc"}});
        if (typeRequestRef.current !== requestId || !response) {
            return;
        }
        if (response.code !== 200) {
            message.error(response.message || "获取脚本分类失败");
            return;
        }
        const items = Array.isArray(response.data?.list) ? response.data.list : [];
        setTypeItems(items);
        return items;
    }, [instanceModule, message]);

    useEffect(() => {
        void loadTypes();
        return () => {
            typeRequestRef.current += 1;
            insertRequestRef.current += 1;
            copyRequestRef.current += 1;
            copyModalRef.current?.destroy();
        };
    }, [loadTypes]);

    useEffect(() => {
        insertRequestRef.current += 1;
        setInsertingId(undefined);
    }, [selectedId]);

    const openCreate = useCallback(() => {
        const defaultTypeId = list.typeId === "0" ? 0 : Number(list.typeId);
        formRef.current?.open(undefined, Number.isFinite(defaultTypeId) ? defaultTypeId : 0);
    }, [list.typeId]);

    const openEdit = useCallback((record: any) => {
        formRef.current?.open(record);
    }, []);

    const insert = useCallback(async (record: any) => {
        if (!connected) {
            message.warning("请先连接终端");
            return;
        }
        const targetServerId = selectedId;
        const requestId = ++insertRequestRef.current;
        setInsertingId(record.id);
        try {
            const response = await getScriptInfo({body: {id: record.id}});
            if (insertRequestRef.current !== requestId || !response) {
                return;
            }
            if (response.code !== 200 || !response.data) {
                message.error(response.message || "获取脚本详情失败");
                return;
            }
            const script = String(response.data.script || "");
            if (!script) {
                message.warning("脚本内容为空");
                return;
            }
            if (!onInsert(script, targetServerId)) {
                message.warning("请先连接终端");
            }
        } finally {
            if (insertRequestRef.current === requestId) {
                setInsertingId(undefined);
            }
        }
    }, [connected, message, onInsert, selectedId]);

    const copy = useCallback(async (record: any) => {
        const requestId = ++copyRequestRef.current;
        copyModalRef.current?.destroy();
        setCopyingId(record.id);
        try {
            const detail = getScriptInfo({body: {id: record.id}});
            let pendingCopy: Promise<boolean> | undefined;
            if (window.isSecureContext && navigator.clipboard?.write && typeof ClipboardItem !== "undefined") {
                // 先在点击时发起复制，再等待接口正文，保留 Safari 的用户操作授权。
                const content = detail.then((response) => {
                    if (copyRequestRef.current !== requestId || response?.code !== 200 || !response.data?.script) {
                        throw new Error("脚本内容不可用");
                    }
                    return new Blob([String(response.data.script)], {type: "text/plain"});
                });
                // 不支持 Promise 内容的浏览器可能同步抛错，接口失败时也需要处理内容 Promise。
                void content.catch(() => undefined);
                try {
                    pendingCopy = navigator.clipboard.write([new ClipboardItem({"text/plain": content})])
                        .then(() => true, () => false);
                } catch {
                    // 继续尝试普通文本复制。
                }
            }
            const response = await detail;
            if (copyRequestRef.current !== requestId || !response) {
                return;
            }
            if (response.code !== 200 || !response.data) {
                message.error(response.message || "获取脚本详情失败");
                return;
            }
            const script = String(response.data.script ?? "");
            if (!script) {
                message.warning("脚本内容为空");
                return;
            }
            let copied = pendingCopy ? await pendingCopy : false;
            if (copyRequestRef.current !== requestId) {
                return;
            }
            if (!copied && window.isSecureContext && navigator.clipboard?.writeText) {
                copied = await navigator.clipboard.writeText(script).then(() => true, () => false);
            }
            if (copyRequestRef.current !== requestId) {
                return;
            }
            if (!copied) {
                copied = fallbackCopyText(script);
            }
            if (copied) {
                message.success("脚本已复制");
            } else {
                // HTTP 等环境需要重新点击，才能取得复制授权；复用本次已获取的正文。
                const dialog = modal.confirm({
                    title: "复制脚本",
                    content: "脚本已加载，点击“复制”完成操作。",
                    okText: "复制",
                    cancelText: "取消",
                    onOk: () => {
                        if (copyRequestRef.current !== requestId) {
                            return;
                        }
                        if (!fallbackCopyText(script)) {
                            message.error("复制脚本失败，请检查浏览器权限");
                            return Promise.reject(new Error("复制失败"));
                        }
                        message.success("脚本已复制");
                    },
                    afterClose: () => {
                        if (copyModalRef.current === dialog) {
                            copyModalRef.current = null;
                        }
                    },
                });
                copyModalRef.current = dialog;
                void Promise.resolve(dialog);
            }
        } catch {
            if (copyRequestRef.current === requestId) {
                message.error("复制脚本失败");
            }
        } finally {
            if (copyRequestRef.current === requestId) {
                setCopyingId(undefined);
            }
        }
    }, [message, modal]);

    const remove = useCallback((record: any) => {
        modal.confirm({
            title: "删除脚本",
            content: `确定删除脚本“${record.name}”吗？`,
            okText: "删除",
            cancelText: "取消",
            okButtonProps: {danger: true, type: "default"},
            mask: {closable: true},
            onOk: async () => {
                setDeletingId(record.id);
                try {
                    const response = await deleteScript({body: {ids: [record.id]}});
                    if (!response) {
                        return;
                    }
                    if (response.code !== 200) {
                        message.error(response.message || "删除脚本失败");
                        return;
                    }
                    message.success(response.message || "删除成功");
                    list.reload();
                } finally {
                    setDeletingId(undefined);
                }
            },
        } as any);
    }, [list.reload, message, modal]);

    const handleTypesChange = useCallback((items: any[]) => {
        typeRequestRef.current += 1;
        setTypeItems(items);
        if (list.typeId !== "0" && !items.some((item) => String(item.id) === list.typeId)) {
            list.setTypeId("0");
        }
    }, [list.setTypeId, list.typeId]);

    const handleScroll = useCallback((event: React.UIEvent<HTMLDivElement>) => {
        const target = event.currentTarget;
        const horizontal = target.scrollWidth > target.clientWidth + 1;
        const remaining = horizontal
            ? target.scrollWidth - target.scrollLeft - target.clientWidth
            : target.scrollHeight - target.scrollTop - target.clientHeight;
        if (remaining < 160) {
            void list.loadMore();
        }
    }, [list.loadMore]);

    return (
        <>
            <aside className={`${styles.scriptPane} ${collapsed ? "collapsed" : ""}`} aria-label="常用脚本">
                <header className={styles.scriptHeader}>
                    <div className="panel-heading">
                        <Title size="small">常用脚本</Title>
                    </div>
                    <Button
                        className="collapse-trigger"
                        type="text"
                        aria-label={collapsed ? "展开常用脚本" : "收起常用脚本"}
                        aria-controls="terminal-script-panel"
                        aria-expanded={!collapsed}
                        icon={<Icon type={collapsed ? "MenuFoldOutlined" : "MenuUnfoldOutlined"}/>}
                        onClick={() => onCollapsedChange(!collapsed)}/>
                </header>

                <div id="terminal-script-panel" className={styles.scriptBody}>
                    <div className="script-toolbar">
                        <Input
                            variant="filled"
                            value={list.keyword}
                            allowClear
                            aria-label="搜索常用脚本"
                            prefix={<Icon type="SearchOutlined"/>}
                            placeholder="搜索名称或内容"
                            onChange={(event) => list.setKeyword(event.target.value)}/>
                        <Dropdown
                            classNames={{root: styles.scriptDropdown}}
                            trigger={["click"]}
                            placement="bottomRight"
                            menu={{
                                selectable: true,
                                selectedKeys: [list.typeId],
                                items: typeMenuItems,
                                onClick: ({key}) => {
                                    if (key === manageTypeKey) {
                                        typeManagerRef.current?.open();
                                        return;
                                    }
                                    list.setTypeId(key);
                                },
                            }}>
                            <Button
                                className="script-type-trigger"
                                color="default"
                                variant="filled"
                                style={{borderColor: "transparent"}}
                                aria-label={`筛选脚本分类，当前${activeTypeName}`}>
                                <Icon className="marker" type="FolderOutlined"/>
                                <span className="value" title={activeTypeName}>{activeTypeName}</span>
                                <Icon className="arrow" type="DownOutlined"/>
                            </Button>
                        </Dropdown>
                    </div>

                    <div
                        className="script-scroll"
                        aria-busy={list.loading || list.loadingMore}
                        onScroll={handleScroll}>
                        <button
                            className={`${styles.serverAdd} mobile`}
                            type="button"
                            aria-label="添加脚本"
                            onClick={openCreate}>
                            <Icon type="PlusOutlined"/>
                            <span>添加脚本</span>
                        </button>
                        {list.loading && list.items.length === 0 ? (
                            <div className="script-state"><Spin size="small"/></div>
                        ) : list.error && list.items.length === 0 ? (
                            <div className="script-state error-state">
                                <span>脚本加载失败</span>
                                <Button
                                    color="primary"
                                    variant="text"
                                    aria-label="重新加载脚本"
                                    icon={<Icon type="ReloadOutlined"/>}
                                    onClick={list.reload}/>
                            </div>
                        ) : list.items.length === 0 ? (
                            <div className="script-state">
                                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
                            </div>
                        ) : (
                            <div className="script-list">
                                {list.items.map((record) => {
                                    const typeName = [
                                        record.module === "ssh" ? "全局脚本" : "实例脚本",
                                        record.type_id ? typeMap[String(record.type_id)] || "未知分类" : "",
                                    ].filter(Boolean).join(" · ");
                                    const copying = copyingId === record.id;
                                    return (
                                        <div
                                            className={`script-item ${copying ? "copying" : ""}`}
                                            key={record.id}
                                            aria-busy={copying}>
                                            <button
                                                className="script-copy"
                                                type="button"
                                                aria-label={`双击复制脚本 ${record.name}`}
                                                onDoubleClick={() => void copy(record)}>
                                                <span className="script-copy-content">
                                                    <span className="script-name" title={record.name}>{record.name}</span>
                                                    <span className="script-type" title={typeName}>{typeName}</span>
                                                </span>
                                            </button>
                                            <Tooltip title="填入终端">
                                                <Button
                                                    className="script-terminal"
                                                    color="default"
                                                    variant="text"
                                                    loading={insertingId === record.id}
                                                    aria-label={`填入终端 ${record.name}`}
                                                    icon={<Icon type="EnterOutlined"/>}
                                                    onClick={() => void insert(record)}/>
                                            </Tooltip>
                                            <ActionDropdown
                                                placement="bottomRight"
                                                menu={{
                                                    items: [
                                                        {key: "copy", label: "复制"},
                                                        {key: "edit", label: "编辑"},
                                                        {type: "divider"},
                                                        {key: "delete", label: "删除", danger: true},
                                                    ],
                                                    onClick: ({key}) => {
                                                        if (key === "copy") {
                                                            void copy(record);
                                                        } else if (key === "edit") {
                                                            openEdit(record);
                                                        } else if (key === "delete") {
                                                            remove(record);
                                                        }
                                                    },
                                                }}>
                                                <Button
                                                    className="script-action"
                                                    color="default"
                                                    variant="text"
                                                    loading={deletingId === record.id}
                                                    aria-label={`管理脚本 ${record.name}`}
                                                    icon={<Icon type="MoreOutlined"/>}/>
                                            </ActionDropdown>
                                            {copying && (
                                                <div className="script-copying" role="status" aria-label={`正在复制脚本 ${record.name}`}>
                                                    <Spin size="small"/>
                                                </div>
                                            )}
                                        </div>
                                    );
                                })}
                                {list.loadingMore && (
                                    <div className="script-load-more"><Spin size="small"/></div>
                                )}
                                {list.hasMore && !list.loadingMore && (
                                    <div className="script-load-more">
                                        <Tooltip title="加载更多脚本">
                                            <Button
                                                type="text"
                                                aria-label="加载更多脚本"
                                                icon={<Icon type="DownOutlined"/>}
                                                onClick={() => void list.loadMore()}/>
                                        </Tooltip>
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </div>

                <div
                    className={styles.scriptRail}
                    aria-hidden={!collapsed}
                    aria-busy={list.loading || list.loadingMore}
                    onScroll={handleScroll}>
                    {list.items.map((record) => (
                        <Tooltip key={`rail-script-${record.id}`} placement="left" title={record.name}>
                            <Button
                                className="rail-action"
                                color="default"
                                variant="text"
                                aria-label={`复制脚本 ${record.name}`}
                                onClick={() => void copy(record)}>
                                <Icon type={copyingId === record.id ? "LoadingOutlined" : "CodeOutlined"}/>
                            </Button>
                        </Tooltip>
                    ))}
                    {list.loading && list.items.length === 0 && (
                        <Icon className="rail-loading" type="LoadingOutlined"/>
                    )}
                    {list.hasMore && !list.loadingMore && (
                        <Tooltip placement="left" title="加载更多脚本">
                            <Button
                                className="rail-action rail-load-more"
                                color="default"
                                variant="text"
                                aria-label="加载更多脚本"
                                onClick={() => void list.loadMore()}>
                                <Icon type="DownOutlined"/>
                            </Button>
                        </Tooltip>
                    )}
                    {list.loadingMore && <Icon className="rail-loading" type="LoadingOutlined"/>}
                </div>

                <footer className={`${styles.serverFooter} script-footer`}>
                    <button className={styles.serverAdd} type="button" aria-label="添加脚本" onClick={openCreate}>
                        <Icon type="PlusOutlined"/>
                        <span className="add-label">添加脚本</span>
                    </button>
                </footer>
            </aside>

            <ScriptForm ref={formRef} instanceModule={instanceModule} onSuccess={list.reload}/>
            <TypeManager
                ref={typeManagerRef}
                module={instanceModule || "ssh"}
                scopes={instanceModule ? [
                    {label: "全局分类", value: "ssh"},
                    {label: "实例分类", value: instanceModule},
                ] : undefined}
                onChange={handleTypesChange}
                onMutation={list.reload}/>
        </>
    );
};

export default React.memo(Scripts);
