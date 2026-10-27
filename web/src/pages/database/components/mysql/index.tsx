import {useMemo, useRef, useState} from "react";
import {App, ConfigProvider, Dropdown, Grid, Input} from "antd";
import {MysqlContext} from "./context";
import Workspace from "./workspace";
import type {DatabasePanelRecord, DatabaseVersion, MysqlApi, MysqlManagerProps, MysqlRequest} from "./entities";

export type {MysqlManagerProps, MysqlRequests, MysqlComponents, DatabasePanelRecord, DatabaseVersion} from "./entities";

const FilePicker = ({mode, buttonProps, onChange, ...props}: any) => <Input {...props} onChange={(event) => onChange?.(event.target.value)}/>;
const ActionDropdown = (props: any) => <Dropdown trigger={["click"]} {...props}/>;

// 独立使用时只需实例、接口方法及关闭回调，不依赖数据管理页面或 shared 组件。
export default ({requests, components, ...props}: MysqlManagerProps) => {
    const screens = Grid.useBreakpoint();
    const [version, setVersion] = useState<DatabaseVersion>();
    const requestMethods = useRef<Record<string, MysqlRequest>>(requests);
    requestMethods.current = requests;
    const operations = JSON.stringify(Object.keys(requests).sort());
    const api = useMemo(() => ({
        ...Object.fromEntries(Object.keys(requests).filter((name) => !name.includes("/")).map((name) => [name,
            async (record: DatabasePanelRecord, body: Record<string, any> = {}) => {
                const response = await requestMethods.current[name]({body: {...body, id: record.id, database_name: body.database_name ?? record.database_name}});
                if (response?.code !== 200) {
                    const error: any = new Error(response?.message || "数据库操作失败");
                    if (response) error.unconfirmed = false;
                    throw error;
                }
                return response?.data;
            },
        ])),
        task: (body: Record<string, any>) => requestMethods.current["system/task"]({body}),
    }) as MysqlApi, [operations]);
    const context = useMemo(() => ({api, version, features: version?.features || {},
        FilePicker: components?.FilePicker || FilePicker,
        ActionDropdown: components?.ActionDropdown || ActionDropdown,
        Editor: components?.Editor, ScriptTools: components?.ScriptTools,
    }), [api, version, components?.FilePicker, components?.ActionDropdown, components?.Editor, components?.ScriptTools]);

    return <MysqlContext.Provider value={context}><ConfigProvider modal={{style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24}}}><App>
        <Workspace {...props} version={version} setVersion={setVersion}/>
    </App></ConfigProvider></MysqlContext.Provider>;
};
