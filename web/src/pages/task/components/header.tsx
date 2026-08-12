import React from "react";
import {Input} from "antd";
import {Icon} from "sinking-antd";
import Dropdown from "@/components/stable-dropdown";
import HeroGraphic from "@/components/hero-graphic";

const manageTypeKey = "__manage_types__";

const Header = ({
    styles,
    keyword,
    typeId,
    execType,
    status,
    typeData,
    typeItems,
    execTypeData,
    statusData,
    onKeywordChange,
    onSearch,
    onTypeIdChange,
    onManageTypes,
    onExecTypeChange,
    onStatusChange,
    onCreate,
}: any) => {
    const typeOptions = React.useMemo(() => [
        {label: "全部分类", value: "0"},
        ...(typeItems
            ? typeItems.map((item: any) => ({value: String(item.id), label: String(item.name)}))
            : Object.entries(typeData || {})
                .filter(([value]) => value !== "0")
                .map(([value, label]) => ({value, label: String(label)}))),
    ], [typeData, typeItems]);
    const typeMenuItems = React.useMemo(() => [
        ...typeOptions.map(({label, value}) => ({key: value, label})),
        {type: "divider" as const},
        {
            key: manageTypeKey,
            label: "分类管理",
            icon: <Icon type="SettingOutlined"/>,
        },
    ], [typeOptions]);
    const execTypeOptions = React.useMemo(() => [
        {label: "全部方式", value: ""},
        ...Object.entries(execTypeData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [execTypeData]);
    const statusOptions = React.useMemo(() => [
        {label: "全部状态", value: ""},
        ...Object.entries(statusData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [statusData]);
    const activeType = typeOptions.find((item) => item.value === typeId)?.label || typeOptions[0].label;
    const activeExecType = execTypeOptions.find((item) => item.value === execType)?.label || execTypeOptions[0].label;
    const activeStatus = statusOptions.find((item) => item.value === status)?.label || statusOptions[0].label;

    return (
        <>
            <section className={styles.hero}>
                <div className="hero-copy">
                    <div className="eyebrow"><span className="status-dot"/>TASK SCHEDULER</div>
                    <h1>计划任务</h1>
                </div>
                <div className="hero-visual"><HeroGraphic variant="task"/></div>
                <button className="create-button" type="button" onClick={onCreate}>
                    <span>添加任务</span>
                </button>
            </section>

            <div className={styles.commandBar}>
                <Input
                    className={styles.searchBox}
                    value={keyword}
                    allowClear
                    prefix={<Icon type="SearchOutlined"/>}
                    placeholder="搜索任务名称"
                    onChange={(event) => onKeywordChange(event.target.value)}
                    onPressEnter={(event) => onSearch(event.currentTarget.value)}
                />
                <div className="command-actions">
                    <Dropdown
                        trigger={["click"]}
                        placement="bottomRight"
                        classNames={{root: styles.toolbarDropdown}}
                        menu={{
                            selectable: true,
                            selectedKeys: [typeId || "0"],
                            items: typeMenuItems,
                            onClick: ({key}) => {
                                if (key === manageTypeKey) {
                                    onManageTypes();
                                    return;
                                }
                                onTypeIdChange(key);
                            },
                        }}>
                        <button className={`${styles.toolbarTrigger} category-trigger`} type="button" aria-label="任务分类筛选">
                            <Icon type="FolderOutlined" className="marker"/>
                            <span className="value" title={activeType}>{activeType}</span>
                            <Icon type="DownOutlined" className="arrow"/>
                        </button>
                    </Dropdown>
                    <Dropdown
                        trigger={["click"]}
                        placement="bottomRight"
                        classNames={{root: styles.toolbarDropdown}}
                        menu={{
                            selectable: true,
                            selectedKeys: [execType || "all"],
                            items: execTypeOptions.map(({label, value}) => ({key: value || "all", label})),
                            onClick: ({key}) => onExecTypeChange(key === "all" ? "" : key),
                        }}>
                        <button className={`${styles.toolbarTrigger} exec-type-trigger`} type="button" aria-label="执行方式筛选">
                            <Icon type="CodeOutlined" className="marker"/>
                            <span className="value" title={activeExecType}>{activeExecType}</span>
                            <Icon type="DownOutlined" className="arrow"/>
                        </button>
                    </Dropdown>
                    <Dropdown
                        trigger={["click"]}
                        placement="bottomRight"
                        classNames={{root: styles.toolbarDropdown}}
                        menu={{
                            selectable: true,
                            selectedKeys: [status || "all"],
                            items: statusOptions.map(({label, value}) => ({key: value || "all", label})),
                            onClick: ({key}) => onStatusChange(key === "all" ? "" : key),
                        }}>
                        <button className={`${styles.toolbarTrigger} status-trigger`} type="button" aria-label="任务状态筛选">
                            <Icon type="FlagOutlined" className="marker"/>
                            <span className="value" title={activeStatus}>{activeStatus}</span>
                            <Icon type="DownOutlined" className="arrow"/>
                        </button>
                    </Dropdown>
                </div>
            </div>
        </>
    );
};

export default React.memo(Header);
