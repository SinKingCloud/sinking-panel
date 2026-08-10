import React, {useMemo, useState} from "react";
import {Button, Collapse, Tooltip} from "antd";
import {Icon, Title} from "sinking-antd";

interface ScriptItem {
    name: string;
    script: string;
}

interface ScriptGroup {
    key: string;
    title: string;
    icon: string;
    items: ScriptItem[];
}

interface ScriptsProps {
    styles: any;
    collapsed: boolean;
    onCollapsedChange: (collapsed: boolean) => void;
    onInsert: (script: string) => void;
}

const groups: ScriptGroup[] = [
    {
        key: "system",
        title: "系统信息",
        icon: "DesktopOutlined",
        items: [
            {name: "系统概览", script: "uname -a"},
            {name: "资源占用", script: "top -bn1 | head -n 15"},
            {name: "磁盘空间", script: "df -h"},
        ],
    },
    {
        key: "service",
        title: "服务管理",
        icon: "AppstoreOutlined",
        items: [
            {name: "运行服务", script: "systemctl --type=service --state=running"},
            {name: "最近日志", script: "journalctl -n 50 --no-pager"},
            {name: "容器列表", script: "docker ps"},
        ],
    },
    {
        key: "network",
        title: "网络诊断",
        icon: "GlobalOutlined",
        items: [
            {name: "网络地址", script: "ip addr"},
            {name: "监听端口", script: "ss -lntp"},
            {name: "路由信息", script: "ip route"},
        ],
    },
];

const Scripts = ({styles, collapsed, onCollapsedChange, onInsert}: ScriptsProps) => {
    const [activeKeys, setActiveKeys] = useState<string[]>(["system"]);
    const items = useMemo(() => groups.map((group) => ({
        key: group.key,
        label: (
            <span className="script-group-title">
                <Icon type={group.icon}/>
                <span>{group.title}</span>
            </span>
        ),
        children: (
            <div className="script-list">
                {group.items.map((item) => (
                    <div className="script-item" key={item.script}>
                        <button type="button" className="script-copy" onClick={() => onInsert(item.script)}>
                            <span className="script-name">{item.name}</span>
                            <code title={item.script}>{item.script}</code>
                        </button>
                        <Tooltip title="填入终端">
                            <Button
                                type="text"
                                aria-label={`填入${item.name}`}
                                icon={<Icon type="EnterOutlined"/>}
                                onClick={() => onInsert(item.script)}/>
                        </Tooltip>
                    </div>
                ))}
            </div>
        ),
    })), [onInsert]);

    return (
        <aside className={`${styles.scriptPane} ${collapsed ? "collapsed" : ""}`}>
            <header className={styles.scriptHeader}>
                <div className="panel-heading">
                    <Title size="small">操作面板</Title>
                </div>
                <Button
                    className="collapse-trigger"
                    type="text"
                    aria-label={collapsed ? "展开操作面板" : "收起操作面板"}
                    icon={<Icon type={collapsed ? "MenuFoldOutlined" : "MenuUnfoldOutlined"}/>}
                    onClick={() => onCollapsedChange(!collapsed)}/>
            </header>
            <div className={styles.scriptBody}>
                <Collapse
                    ghost
                    activeKey={activeKeys}
                    expandIconPlacement="end"
                    onChange={(keys) => setActiveKeys((Array.isArray(keys) ? keys : [keys]).map(String))}
                    items={items}/>
            </div>
            <div className={styles.scriptRail} aria-hidden={!collapsed}>
                {groups.map((group) => (
                    <Tooltip
                        key={`rail-${group.key}`}
                        placement="left"
                        title={`${group.title}：${group.items.map((item) => item.name).join("、")}`}>
                        <button
                            className="rail-action"
                            type="button"
                            aria-label={`展开${group.title}`}
                            onClick={() => {
                                setActiveKeys((current) => current.includes(group.key)
                                    ? current
                                    : [...current, group.key]);
                                onCollapsedChange(false);
                            }}>
                            <Icon type={group.icon}/>
                        </button>
                    </Tooltip>
                ))}
            </div>
        </aside>
    );
};

export default React.memo(Scripts);
