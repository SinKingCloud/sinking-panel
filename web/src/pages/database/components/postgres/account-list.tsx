import AccountsLayout, {PermissionSummary} from "../shared/accounts-layout";

export default (props: any) => <AccountsLayout {...props} scope={props.schema}
    renderPermissions={(role: any) => {
        const values = new Set(String(role.roles || "").split(/[,\s]+/).filter(Boolean));
        const readable = values.has("SELECT");
        const writable = ["INSERT","UPDATE","DELETE","TRUNCATE"].some(value => values.has(value));
        const label = values.has("SUPERUSER") ? "超级用户"
            : values.has("OWNER") ? "对象管理"
            : readable && writable ? "读写数据" : readable ? "只读访问" : writable ? "可写数据"
            : values.has("CREATE") ? "可建对象"
            : ["TRIGGER", "REFERENCES"].some(value => values.has(value)) ? "对象权限"
            : values.has("CREATEDB") ? "可建数据库" : values.has("CREATEROLE") ? "可管角色"
            : values.has("USAGE") ? "可访问架构"
            : values.has("TEMPORARY") ? "可建临时表" : values.has("CONNECT") ? "仅可连接"
            : values.has("LOGIN") ? "可登录" : "无数据权限";
        const names: Record<string, string> = {
            SUPERUSER: "超级用户", CREATEDB: "创建数据库", CREATEROLE: "管理角色", LOGIN: "允许登录",
            CONNECT: "连接数据库", TEMPORARY: "创建临时表", USAGE: "访问架构", OWNER: "管理所拥有的对象",
            CREATE: "创建对象", SELECT: "读取数据", INSERT: "新增数据", UPDATE: "修改数据", DELETE: "删除数据",
            TRUNCATE: "清空数据", REFERENCES: "引用字段", TRIGGER: "创建触发器",
        };
        const detail = `数据库：${props.database}；架构：${props.schema || "public"}。${values.size
            ? `已具备：${[...values].map(value => names[value] || value).join("、")}。`
            : "未检测到当前范围的权限。"}${values.has("SUPERUSER") ? "超级用户可管理所有对象。"
            : "对象权限可能仅覆盖部分对象或字段。"}`;
        return <PermissionSummary label={label} detail={detail}/>;
    }}/>;
