import React from "react";
import {Dropdown, Input} from "antd";
import {Icon} from "sinking-antd";
import HeroGraphic from "@/components/hero-graphic";

const Header = ({
    styles,
    keyword,
    type,
    status,
    typeData,
    statusData,
    onKeywordChange,
    onSearch,
    onTypeChange,
    onStatusChange,
    onCreate,
}: any) => {
    const typeOptions = React.useMemo(() => [
        {label: "全部类型", value: ""},
        ...Object.entries(typeData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [typeData]);
    const statusOptions = React.useMemo(() => [
        {label: "全部状态", value: ""},
        ...Object.entries(statusData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [statusData]);
    const activeType = typeOptions.find((item) => item.value === type)?.label || typeOptions[0].label;
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
                            selectedKeys: [type || "all"],
                            items: typeOptions.map(({label, value}) => ({key: value || "all", label})),
                            onClick: ({key}) => onTypeChange(key === "all" ? "" : key),
                        }}>
                        <button className={`${styles.toolbarTrigger} type-trigger`} type="button" aria-label="任务类型筛选">
                            <Icon type="TagsOutlined" className="marker"/>
                            <span className="value">{activeType}</span>
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
                            <span className="value">{activeStatus}</span>
                            <Icon type="DownOutlined" className="arrow"/>
                        </button>
                    </Dropdown>
                </div>
            </div>
        </>
    );
};

export default React.memo(Header);
