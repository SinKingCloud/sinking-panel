import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App} from "antd";
import {Body, Icon} from "sinking-antd";
import type {PageTableProps} from "sinking-antd";
import {history, useLocation} from "umi";
import {deleteOss} from "@/service/api/oss";
import Local from "./components/local";
import Oss from "./components/oss";
import OssForm from "./components/oss/form";
import Storage from "./components/storage";
import {readFileStorage, updateFileStorage} from "./hooks/file-storage";

export default () => {
    const {message, modal} = App.useApp();
    const location = useLocation();
    const locationRef = useRef(location);
    locationRef.current = location;
    const content = useRef<HTMLDivElement>(null);
    const live = useRef(true);
    const busy = useRef(new Set<number>());
    const confirmation = useRef<{destroy: () => void}>();
    const created = useRef<number>();
    const search = useMemo(() => new URLSearchParams(location.search), [location.search]);
    const cloud = useMemo(() => {
        if (search.has("storage")) return search.get("storage") === "cloud";
        if (search.has("path") || search.get("view") === "disks") return false;
        try { return readFileStorage().storage === "cloud"; }
        catch { return false; }
    }, [search]);
    const previous = useRef(cloud);
    const [selected, setSelected] = useState(0);
    const [visited, setVisited] = useState<number[]>([]);
    const [localVisited, setLocalVisited] = useState(!cloud);
    const [storageVisited, setStorageVisited] = useState(cloud);
    const [storageRevision, setStorageRevision] = useState(0);
    const [removing, setRemoving] = useState<number>();
    const [revisions, setRevisions] = useState<Record<number, number>>({});
    const [record, setRecord] = useState<{id: number}>();
    const [formOpen, setFormOpen] = useState(false);
    const panes = useMemo(() => !selected || visited.includes(selected) ? visited : [...visited, selected], [visited, selected]);

    useEffect(() => {
        live.current = true;
        return () => { live.current = false; confirmation.current?.destroy(); };
    }, []);
    useEffect(() => {
        try { updateFileStorage((current) => ({...current, storage: cloud ? "cloud" : "local"})); }
        catch { /* 浏览器禁用本地存储时仍可正常切换。 */ }
    }, [cloud]);
    useEffect(() => {
        if (search.has("storage") || search.has("path") || search.get("view") === "disks") return;
        // 菜单入口恢复上次视图，并写入当前历史记录，保证前进、后退时视图确定。
        const next = new URLSearchParams(search);
        next.set("storage", cloud ? "cloud" : "local");
        history.replace({pathname: location.pathname, search: `?${next}`});
    }, [cloud, location.pathname, search]);
    useEffect(() => {
        if (cloud) setStorageVisited(true);
        else { setLocalVisited(true); setSelected(0); }
        if (selected) setVisited((items) => items.includes(selected) ? items : [...items, selected]);
        if (previous.current !== cloud) {
            content.current?.querySelector<HTMLButtonElement>(`[data-storage="${cloud ? "cloud" : "local"}"] button[aria-label="切换存储"]`)?.focus({preventScroll: true});
            previous.current = cloud;
        }
    }, [cloud, selected]);

    // 只更新存储参数，本地目录及前进、后退记录继续由文件组件管理。
    const changeStorage = useCallback((cloud: boolean) => {
        setSelected(0);
        const current = locationRef.current;
        const search = new URLSearchParams(current.search);
        search.set("storage", cloud ? "cloud" : "local");
        const next = `?${search}`;
        if (next !== current.search) history.push({pathname: current.pathname, search: next});
    }, []);
    const changeBusy = useCallback((id: number, value: boolean) => {
        if (value) busy.current.add(id);
        else busy.current.delete(id);
    }, []);
    const configure = (id?: number) => {
        if (id && busy.current.has(id)) { message.warning("请等待当前存储的文件操作完成"); return; }
        created.current = undefined;
        setRecord(id ? {id} : undefined);
        setFormOpen(true);
    };
    const remove = ({id, name}: {id: number; name: string}) => {
        if (busy.current.has(id)) { message.warning("请等待当前存储的文件操作完成"); return; }
        confirmation.current?.destroy();
        confirmation.current = modal.confirm({
            title: `删除 ${name || "OSS 存储"}？`, content: "只删除面板中的连接配置，存储桶及云端文件不会删除。",
            okText: "删除", cancelText: "取消", okButtonProps: {danger: true},
            onOk: async () => {
                if (busy.current.has(id)) { message.warning("请等待当前存储的文件操作完成"); throw new Error("存储正在使用中"); }
                setRemoving(id);
                try {
                    const response = await deleteOss({body: {ids: [id]}});
                    if (!live.current) return;
                    if (response?.code !== 200) throw new Error(response?.message || "删除失败");
                    setSelected((current) => current === id ? 0 : current);
                    setVisited((items) => items.filter((item) => item !== id));
                    busy.current.delete(id);
                    setStorageRevision((value) => value + 1);
                    message.success("已删除 OSS 配置");
                } catch (reason: any) {
                    if (live.current) message.error(reason?.message || "删除失败");
                    throw reason;
                } finally { if (live.current) setRemoving(undefined); }
            },
        });
    };
    const saved = (id: number, editing: boolean) => {
        setStorageRevision((value) => value + 1);
        if (!editing) return;
        setRevisions((items) => ({...items, [id]: (items[id] || 0) + 1}));
        if (created.current === id) setSelected(id);
        created.current = undefined;
    };
    const hero: NonNullable<PageTableProps["hero"]> = {
        eyebrow: "FILE MANAGER", title: "文件管理",
        action: {
            label: "切换存储", ariaLabel: "切换存储", icon: cloud ? "CloudOutlined" : "FolderOutlined",
            menu: {
                selectedKeys: [cloud ? "cloud" : "local"],
                items: [
                    {key: "local", label: "本地存储", icon: <Icon type="FolderOutlined"/>},
                    {key: "cloud", label: "云端存储", icon: <Icon type="CloudOutlined"/>},
                ],
                onClick: ({key}) => changeStorage(key === "cloud"),
            },
        },
    };

    // 已访问的存储保留挂载，上传不中断；隐藏列表由各文件组件暂停查询。
    return <Body>
        <div ref={content}>
            {(storageVisited || cloud) && <div data-storage="cloud" hidden={!cloud}>
                <Storage active={cloud} revision={storageRevision} hero={hero} removing={removing}
                    onOpen={(item) => setSelected(item.id)} onCreate={() => configure()} onEdit={configure} onDelete={remove}/>
            </div>}
            {(localVisited || !cloud) && <div data-storage="local" hidden={cloud}>
                <Local active={!cloud} hero={hero}/>
            </div>}
        </div>
        {panes.map((id) => <Oss key={id} id={id} active={cloud && selected === id} revision={revisions[id] || 0}
            onCancel={() => setSelected(0)} onEdit={configure} onBusyChange={changeBusy}/>)}
        <OssForm open={formOpen} record={record} onCancel={() => setFormOpen(false)}
            onCreated={(id: number) => { created.current = id; setRecord({id}); }} onSuccess={saved}/>
    </Body>;
};
