import React from "react";
import {Dropdown, Input} from "antd";
import {Icon} from "sinking-antd";

const Graphic = React.memo(() => (
    <svg className="scheduler-graphic" viewBox="0 0 430 132" aria-hidden="true" focusable="false">
        <path className="time-track" d="M14 84C76 84 87 49 143 49S211 98 270 77 350 47 416 62"/>
        <path className="time-track secondary" d="M18 105H410"/>
        <path className="time-ticks" d="M44 100V110M78 102V108M112 100V110M146 102V108M180 100V110M214 102V108M248 100V110M282 102V108M316 100V110M350 102V108M384 100V110"/>
        <circle className="track-node" cx="82" cy="67" r="5"/>
        <circle className="track-node secondary" cx="164" cy="57" r="4"/>
        <circle className="track-node" cx="350" cy="53" r="5"/>
        <g className="hero-clock" transform="translate(270 64)">
            <circle className="clock-orbit" r="48"/>
            <circle className="clock-face" r="31"/>
            <path className="clock-markers" d="M0-25V-21M25 0H21M0 25V21M-25 0H-21"/>
            <path className="clock-hands" d="M0 0V-15M0 0L13 7"/>
            <circle className="clock-center" r="3"/>
            <circle className="orbit-node" cx="-37" cy="-30" r="4"/>
            <circle className="orbit-node secondary" cx="43" cy="21" r="3.5"/>
        </g>
    </svg>
));

Graphic.displayName = "Graphic";

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
                <div className="hero-visual"><Graphic/></div>
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
