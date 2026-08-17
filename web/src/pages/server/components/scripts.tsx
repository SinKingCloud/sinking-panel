import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, Button, Empty, Input, Spin, Tooltip} from "antd";
import {Icon, Title} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import TypeManager from "@/pages/components/type-manager";
import type {TypeManagerRef} from "@/pages/components/type-manager";
import {deleteScript, getScriptInfo} from "@/service/api/script";
import {getAllTypes} from "@/service/api/type";
import useScripts from "../hooks/scripts";
import ScriptForm from "./script-form";
import type {ScriptFormRef} from "./script-form";

const manageTypeKey = "__manage_script_types__";

const fallbackCopyText = (value: string) => {
    let textarea: HTMLTextAreaElement | undefined;
    try {
        textarea = document.createElement("textarea");
        textarea.value = value;
        textarea.setAttribute("readonly", "");
        textarea.style.position = "fixed";
        textarea.style.left = "-9999px";
        document.body.appendChild(textarea);
        textarea.select();
        return document.execCommand("copy");
    } catch {
        return false;
    } finally {
        textarea?.remove();
    }
};

interface ScriptsProps {
    styles: any;
    collapsed: boolean;
    connected: boolean;
    selectedId: number;
    onCollapsedChange: (collapsed: boolean) => void;
    onInsert: (script: string, serverId: number) => boolean;
}

const Scripts = ({styles, collapsed, connected, selectedId, onCollapsedChange, onInsert}: ScriptsProps) => {
    const {message, modal} = App.useApp();
    const list = useScripts();
    const formRef = useRef<ScriptFormRef>({} as ScriptFormRef);
    const typeManagerRef = useRef<TypeManagerRef>({} as TypeManagerRef);
    const typeRequestRef = useRef(0);
    const insertRequestRef = useRef(0);
    const copyRequestRef = useRef(0);
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
        const response = await getAllTypes("script");
        if (typeRequestRef.current !== requestId || !response) {
            return;
        }
        if (response.code !== 200) {
            message.error(response.message || "获取脚本分类失败");
            return;
        }
        setTypeItems(Array.isArray(response.data?.list) ? response.data.list : []);
    }, [message]);

    useEffect(() => {
        void loadTypes();
        return () => {
            typeRequestRef.current += 1;
            insertRequestRef.current += 1;
            copyRequestRef.current += 1;
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
        setCopyingId(record.id);
        const responsePromise = getScriptInfo({body: {id: record.id}});
        let clipboardWrite: Promise<boolean> | undefined;
        if (navigator.clipboard?.write && typeof ClipboardItem !== "undefined") {
            try {
                const content = responsePromise.then((response) => {
                    if (copyRequestRef.current !== requestId || response?.code !== 200 || !response.data) {
                        throw new Error("script detail unavailable");
                    }
                    return new Blob([String(response.data.script || "")], {type: "text/plain"});
                });
                clipboardWrite = navigator.clipboard.write([
                    new ClipboardItem({"text/plain": content}),
                ]).then(() => true, () => false);
            } catch {
                clipboardWrite = undefined;
            }
        }
        try {
            const response = await responsePromise;
            if (copyRequestRef.current !== requestId || !response) {
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
            let copied = clipboardWrite ? await clipboardWrite : false;
            if (!copied && navigator.clipboard?.writeText) {
                copied = await navigator.clipboard.writeText(script).then(() => true, () => false);
            }
            if (!copied) {
                copied = fallbackCopyText(script);
            }
            if (copied) {
                message.success("脚本已复制");
            } else {
                message.error("复制脚本失败");
            }
        } finally {
            if (copyRequestRef.current === requestId) {
                setCopyingId(undefined);
            }
        }
    }, [message]);

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
                            <button
                                className="script-type-trigger"
                                type="button"
                                aria-label={`筛选脚本分类，当前${activeTypeName}`}>
                                <Icon className="marker" type="FolderOutlined"/>
                                <span className="value" title={activeTypeName}>{activeTypeName}</span>
                                <Icon className="arrow" type="DownOutlined"/>
                            </button>
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
                                    type="text"
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
                                    const typeName = record.type_id === 0
                                        ? "全部分类"
                                        : typeMap[String(record.type_id)] || "未知分类";
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
                                                    type="text"
                                                    loading={insertingId === record.id}
                                                    aria-label={`填入终端 ${record.name}`}
                                                    icon={<Icon type="EnterOutlined"/>}
                                                    onClick={() => void insert(record)}/>
                                            </Tooltip>
                                            <Dropdown
                                                classNames={{root: styles.serverMenu}}
                                                trigger={["click"]}
                                                placement="bottomRight"
                                                arrow={{pointAtCenter: true}}
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
                                                    type="text"
                                                    loading={deletingId === record.id}
                                                    aria-label={`管理脚本 ${record.name}`}
                                                    icon={<Icon type="MoreOutlined"/>}/>
                                            </Dropdown>
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
                            <button
                                className="rail-action"
                                type="button"
                                aria-label={`复制脚本 ${record.name}`}
                                onClick={() => void copy(record)}>
                                <Icon type={copyingId === record.id ? "LoadingOutlined" : "CodeOutlined"}/>
                            </button>
                        </Tooltip>
                    ))}
                    {list.loading && list.items.length === 0 && (
                        <Icon className="rail-loading" type="LoadingOutlined"/>
                    )}
                    {list.hasMore && !list.loadingMore && (
                        <Tooltip placement="left" title="加载更多脚本">
                            <button
                                className="rail-action rail-load-more"
                                type="button"
                                aria-label="加载更多脚本"
                                onClick={() => void list.loadMore()}>
                                <Icon type="DownOutlined"/>
                            </button>
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

            <ScriptForm ref={formRef} typeItems={typeItems} onSuccess={list.reload}/>
            <TypeManager
                ref={typeManagerRef}
                module="script"
                onChange={handleTypesChange}
                onMutation={list.reload}/>
        </>
    );
};

export default React.memo(Scripts);
