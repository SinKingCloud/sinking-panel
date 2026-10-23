import {useState} from "react";
import {Button, Empty, Segmented, Select, Spin} from "antd";
import {Icon} from "sinking-antd";
import SqlPanel from "./sql/panel";
import Accounts from "./accounts";
import Maintenance from "./maintenance";
import {useMysqlClient} from "./context";
import useStyles from "./styles";
import type {SqlHostProps, SqlTarget} from "./sql/entities";

const sections = [
    {value: "browse", label: "数据浏览"},
    {value: "objects", label: "对象管理"},
    {value: "query", label: "SQL 查询"},
    {value: "accounts", label: "账户权限"},
    {value: "maintenance", label: "数据维护"},
];

export default ({record, initialTab, engine, compact, onTabChange}: SqlHostProps & {compact: boolean; onTabChange?: (tab: string) => void}) => {
    const {features = {}} = useMysqlClient();
    const {styles} = useStyles({compact});
    const hasObjects = engine.objects.some(({key}) => features[`${key}s`] !== false);
    const [selectedSection, setSection] = useState(["objects", "view", "routine", "trigger"].includes(initialTab || "") ? "objects"
        : ["query", "accounts", "maintenance"].includes(initialTab || "") ? initialTab! : "browse");
    const section = selectedSection === "accounts" && features.accounts === false || selectedSection === "objects" && !hasObjects ? "browse" : selectedSection;
    const [target, setTarget] = useState<SqlTarget>({table: "", isView: false, schemas: [], tables: []});
    const [reload, setReload] = useState(0);
    const refresh = () => setReload((value) => value + 1);
    const changeSection = (value: string) => { setSection(value); onTabChange?.(value); };

    return <div className={styles.host}>
        <header className="mysql-header">
            <Segmented className="mysql-navigation" aria-label="MySQL 管理功能" value={section} onChange={changeSection}
                options={sections.filter((item) => (item.value !== "accounts" || features.accounts !== false) && (item.value !== "objects" || hasObjects))}/>
        </header>
        {/* 保持 SQL 面板挂载，切换功能区时保留 SQL、执行结果和数据筛选。 */}
        <SqlPanel record={record} initialTab={initialTab} engine={engine} section={section === "query" ? "query" : section === "objects" ? "objects" : "browse"}
            hidden={section !== "browse" && section !== "query" && section !== "objects"} onBrowse={() => changeSection("browse")}
            onTargetChange={setTarget} tableValue={target.table} reloadKey={reload}/>
        {section === "accounts" && <Accounts record={record} initialAction={record.accountAction}/>}
        {section === "maintenance" && <div className="mysql-maintenance">
            <div className="mysql-maintenance-target">
                <div className="mysql-browse-target">
                    <Select className="mysql-table-picker" aria-label="维护数据表" variant="filled" size="small" prefix={<Icon type="TableOutlined"/>}
                        value={target.table || undefined} placeholder="选择数据表或视图" showSearch optionFilterProp="label"
                        loading={target.loading} disabled={target.loading || Boolean(target.error) || !target.tables.length}
                        options={target.tables.map((item) => ({value: item.name, label: `${item.name}${String(item.type).toUpperCase().includes("VIEW") ? "（视图）" : ""}`}))}
                        onChange={(table) => setTarget((value) => ({...value, table, isView: String(value.tables.find((item) => item.name === table)?.type).toUpperCase().includes("VIEW")}))}/>
                </div>
                <Button color="default" variant="filled" size="small" icon={<Icon type="ReloadOutlined"/>} aria-label="刷新维护数据表" loading={target.loading} onClick={refresh}/>
            </div>
            {target.loading && !target.tables.length ? <div className="mysql-section-loading"><Spin/></div>
                : !target.table ? <div className="client-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/></div>
                : <Maintenance key={`${target.schema}:${target.table}`} record={record} schema={target.schema}
                table={target.table} isView={target.isView} storageEngine={target.tables.find((item) => item.name === target.table)?.engine}
                disabled={target.loading || Boolean(target.error)} onChange={refresh}/>}
        </div>}
    </div>;
};
