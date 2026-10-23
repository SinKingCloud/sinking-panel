import {useMemo, useRef, useState} from "react";
import {App, ConfigProvider, Dropdown, Grid} from "antd";
import {RedisContext} from "./context";
import Workspace from "./workspace";
import type {RedisApi, RedisManagerProps, RedisRequest, RedisRecord, RedisVersion} from "./entities";

export type {RedisManagerProps, RedisRequests, RedisRecord, RedisVersion} from "./entities";
const ActionDropdown = (props: any) => <Dropdown trigger={["click"]} {...props}/>;

// 完整管理器的请求及菜单由使用方注入，不依赖面板页面或固定接口地址。
export default ({requests, components, ...props}: RedisManagerProps) => {
    const screens = Grid.useBreakpoint();
    const [version, setVersion] = useState<RedisVersion>();
    const requestMethods = useRef<Record<string, RedisRequest>>(requests);
    requestMethods.current = requests;
    const operations = JSON.stringify(Object.keys(requests).sort());
    const api = useMemo(() => Object.fromEntries(Object.keys(requests).map((name) => [name,
        async (record: RedisRecord, body: Record<string, any> = {}) => {
            const response = await requestMethods.current[name]({body: {...body, id: record.id, database_name: body.database_name ?? record.database_name}});
            if (response?.code !== 200) throw new Error(response?.message || "数据库操作失败");
            return response.data;
        },
    ])) as RedisApi, [operations]);
    const context = useMemo(() => ({api, version, features: version?.features || {}, ActionDropdown: components?.ActionDropdown || ActionDropdown, Editor: components?.Editor}), [api, version, components?.ActionDropdown, components?.Editor]);
    return <RedisContext.Provider value={context}>
        <ConfigProvider modal={{style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24}}}><App>
            <Workspace {...props} version={version} setVersion={setVersion}/>
        </App></ConfigProvider>
    </RedisContext.Provider>;
};
