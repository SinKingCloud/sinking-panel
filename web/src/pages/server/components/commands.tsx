import React, {useMemo, useState} from "react";
import {Button, Collapse, Tooltip} from "antd";
import {Icon, Title} from "sinking-antd";

interface CommandItem {
    name: string;
    command: string;
}

interface CommandGroup {
    key: string;
    title: string;
    icon: string;
    items: CommandItem[];
}

interface CommandsProps {
    styles: any;
    collapsed: boolean;
    onCollapsedChange: (collapsed: boolean) => void;
    onInsert: (command: string) => void;
}

const groups: CommandGroup[] = [
    {
        key: "system",
        title: "系统信息",
        icon: "DesktopOutlined",
        items: [
            {name: "系统概览", command: "uname -a"},
            {name: "资源占用", command: "top -bn1 | head -n 15"},
            {name: "磁盘空间", command: "df -h"},
        ],
    },
    {
        key: "service",
        title: "服务管理",
        icon: "AppstoreOutlined",
        items: [
            {name: "运行服务", command: "systemctl --type=service --state=running"},
            {name: "最近日志", command: "journalctl -n 50 --no-pager"},
            {name: "容器列表", command: "docker ps"},
        ],
    },
    {
        key: "network",
        title: "网络诊断",
        icon: "GlobalOutlined",
        items: [
            {name: "网络地址", command: "ip addr"},
            {name: "监听端口", command: "ss -lntp"},
            {name: "路由信息", command: "ip route"},
        ],
    },
];

const Commands = ({styles, collapsed, onCollapsedChange, onInsert}: CommandsProps) => {
    const [activeKeys, setActiveKeys] = useState<string[]>(["system"]);
    const items = useMemo(() => groups.map((group) => ({
        key: group.key,
        label: (
            <span className="command-group-title">
                <Icon type={group.icon}/>
                <span>{group.title}</span>
            </span>
        ),
        children: (
            <div className="command-list">
                {group.items.map((item) => (
                    <div className="command-item" key={item.command}>
                        <button type="button" className="command-copy" onClick={() => onInsert(item.command)}>
                            <span className="command-name">{item.name}</span>
                            <code title={item.command}>{item.command}</code>
                        </button>
                        <Tooltip title="填入终端">
                            <Button
                                type="text"
                                aria-label={`填入${item.name}`}
                                icon={<Icon type="EnterOutlined"/>}
                                onClick={() => onInsert(item.command)}/>
                        </Tooltip>
                    </div>
                ))}
            </div>
        ),
    })), [onInsert]);

    return (
        <aside className={`${styles.commandPane} ${collapsed ? "collapsed" : ""}`}>
            <header className={styles.commandHeader}>
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
            <div className={styles.commandBody}>
                <Collapse
                    ghost
                    activeKey={activeKeys}
                    expandIconPlacement="end"
                    onChange={(keys) => setActiveKeys((Array.isArray(keys) ? keys : [keys]).map(String))}
                    items={items}/>
            </div>
            <div className={styles.commandRail} aria-hidden={!collapsed}>
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

export default React.memo(Commands);
