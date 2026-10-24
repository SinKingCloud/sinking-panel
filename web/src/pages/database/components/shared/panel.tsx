import {useEffect, useMemo, useState} from "react";
import type {ComponentType} from "react";
import {App, Button, ConfigProvider, Dropdown, Input, Spin, theme, Tooltip, Typography} from "antd";
import {useTheme} from "sinking-antd";
import {DatabaseContext} from "./context";
import type {DatabasePanelProps, DatabaseVersion} from "./entities";
import useStyles from "./styles/client";

const FilePicker = ({mode, buttonProps, onChange, ...props}: any) => <Input {...props} onChange={(event) => onChange?.(event.target.value)}/>;
const ActionDropdown = (props: any) => <Dropdown trigger={["click"]} {...props}/>;

// 为每个独立入口提供请求和完整样式，不依赖页面工作区或外部弹窗。
export default ({record: target, initialTab, onTabChange, request, version: suppliedVersion, components, className, style, type, panel: Panel, engine, panelProps}: DatabasePanelProps & {
    type: number;
    panel: ComponentType<any>;
    engine?: any;
    panelProps?: Record<string, any>;
}) => {
    const {styles} = useStyles();
    const {token} = theme.useToken();
    const appearance = useTheme();
    const compact = Boolean(appearance?.isCompactTheme?.());
    const fontSize = compact ? 12 : 13;
    const buttonFontSize = compact ? 11 : 12;
    const controlHeight = compact ? 24 : 28;
    const record = useMemo(() => ({...target, type}), [target, type]);
    const key = JSON.stringify([type, target.id, target.database_name || target.name]);
    const [retry, setRetry] = useState(0);
    const [server, setServer] = useState<{key: string; ready: boolean; loading: boolean; version?: DatabaseVersion; error?: string}>({key, ready: false, loading: true});
    useEffect(() => {
        if (suppliedVersion) return;
        let active = true;
        setServer((value) => value.key === key ? {...value, loading: true} : {key, ready: false, loading: true});
        void (async () => {
            try {
                const response = await request("version", {id: target.id, database_name: target.database_name || target.name});
                if (response?.code !== 200) throw new Error(response?.message || "获取数据库版本失败");
                const data = response.data;
                if (!data || typeof data.version !== "string" || !data.version.trim() || !data.features || typeof data.features !== "object" || Array.isArray(data.features)) {
                    throw new Error("数据库版本信息不完整");
                }
                if (active) setServer({key, ready: true, loading: false, version: data});
            } catch (reason: any) {
                // 权限或网络错误不代表旧版本，保留基础操作并允许重新获取。
                if (active) setServer((value) => ({...value, key, ready: true, loading: false, error: reason?.message || "获取数据库版本失败"}));
            }
        })();
        return () => { active = false; };
    }, [key, request, retry, suppliedVersion]);
    const current = suppliedVersion ? {key, ready: true, loading: false, version: suppliedVersion, error: undefined} : server.key === key ? server : undefined;
    const version = current?.version;
    const actions = useMemo(() => {
        const operate = async (target: DatabasePanelProps["record"], operation: string, body: Record<string, any> = {}) => {
            const response = await request(operation, {...body, id: target.id, database_name: body.database_name ?? target.database_name ?? target.name});
            if (response?.code !== 200) {
                const error: any = new Error(response?.message || "数据库操作失败");
                error.responseReceived = Boolean(response);
                throw error;
            }
            return response.data;
        };
        const operations = ["version", "databases", "database_info", "schemas", "tables", "columns", "indexes", "rows", "ddl", "query", "execute", "insert", "update", "delete",
            "create_table", "drop_table", "rename_table", "add_column", "update_column", "drop_column", "create_index", "update_index", "drop_index", "objects", "object_definition", "save_object", "drop_object",
            "create_foreign_key", "update_foreign_key", "drop_foreign_key", "users", "maintenance", "import", "export"];
        return {operate, api: {...Object.fromEntries(operations.map((operation) => [operation,
            (record: DatabasePanelProps["record"], body?: Record<string, any>) => operate(record, operation, body)])),
            task: (body: Record<string, any>) => request("system/task", body)}};
    }, [request]);
    const context = useMemo(() => ({
        ...actions, request, version, features: version?.features || {},
        FilePicker: components?.FilePicker || FilePicker,
        ActionDropdown: components?.ActionDropdown || ActionDropdown,
        Editor: components?.Editor, ScriptTools: components?.ScriptTools,
    }), [actions, request, version, components?.FilePicker, components?.ActionDropdown, components?.Editor, components?.ScriptTools]);
    return <DatabaseContext.Provider value={context}><ConfigProvider theme={{
        token: {fontSize, fontSizeSM: fontSize},
        components: {
            Button: {contentFontSize: token.fontSize, contentFontSizeSM: token.fontSize},
            Modal: {titleFontSize: token.fontSizeHeading5, titleLineHeight: token.lineHeightHeading5},
            Input: {controlHeight, inputFontSize: fontSize, inputFontSizeSM: fontSize},
            Select: {controlHeight, optionFontSize: fontSize},
            Table: {fontFamily: token.fontFamily, lineHeight: token.lineHeight, colorText: token.colorTextSecondary, headerColor: token.colorTextSecondary,
                cellFontSize: buttonFontSize, cellFontSizeMD: buttonFontSize, cellFontSizeSM: buttonFontSize,
                cellPaddingBlock: compact ? 6 : 8, cellPaddingBlockMD: compact ? 6 : 8, cellPaddingBlockSM: compact ? 6 : 8,
                cellPaddingInline: compact ? 8 : 10, cellPaddingInlineMD: compact ? 8 : 10, cellPaddingInlineSM: compact ? 8 : 10},
            Segmented: {controlHeight: compact ? 28 : 32, ...(!compact && {controlHeightSM: 28}), itemColor: token.colorTextSecondary},
        },
    }}>
        <App className={[styles.panel, type !== 1 && styles.workspace, className].filter(Boolean).join(" ")} style={style}>
            {!current?.ready ? <div className={styles.loading}><Spin/></div> : <>
                {current.error && <div className={styles.version}>
                    <Tooltip title={current.error}><Typography.Text type="secondary">版本获取失败</Typography.Text></Tooltip>
                    <Button type="text" size="small" loading={current.loading} onClick={() => setRetry((value) => value + 1)}>重试</Button>
                </div>}
                <Panel {...panelProps} key={key} record={record} initialTab={initialTab} onTabChange={onTabChange} engine={engine} compact={compact}/>
            </>}
        </App>
    </ConfigProvider></DatabaseContext.Provider>;
};
