const address = [
    {name: "host", label: "主机地址", type: "text", placeholder: "IP 地址或域名"},
    {name: "port", label: "端口", type: "number", min: 1, max: 65535},
];
const user = {name: "user", label: "用户名", type: "text", placeholder: "输入数据库用户名"};
const password = {name: "password", label: "密码", type: "password", placeholder: "未设置可留空"};
const socket = {name: "socket", label: "Socket 路径", type: "path", advanced: true, wide: true};
const database = {name: "database", label: "默认数据库", type: "text", placeholder: "可选", advanced: true, wide: true};
const tls = {
    name: "tls", label: "TLS 连接", type: "select", advanced: true, wide: true,
    valueAliases: {"": "false", disable: "false", require: "skip-verify"},
    options: [
        {value: "false", label: "关闭"},
        {value: "true", label: "开启并验证证书"}, {value: "skip-verify", label: "仅加密，不验证证书"},
    ],
};
const certificates = [
    {name: "ssl_ca", label: "CA 证书"},
    {name: "ssl_cert", label: "客户端证书"},
    {name: "ssl_key", label: "客户端私钥"},
].map((field) => ({
    ...field, type: "path", advanced: true, wide: true, placeholder: "可选，选择服务器上的文件",
    // 已保存的证书仍可查看和清空，切换连接方式不会丢失配置。
    visible: (config: any) => ["true", "skip-verify", "require", "preferred"].includes(config?.tls)
        || Boolean(config?.ssl_ca || config?.ssl_cert || config?.ssl_key),
}));

// 配置值由后端各数据库的 Config 返回，页面只定义输入控件与布局。
const connectionFields: Record<number, any[]> = {
    1: [
        ...address,
        {...password, wide: true},
        {...user, label: "用户名（可选）", placeholder: "留空使用默认用户", advanced: true, wide: true},
        {name: "db", label: "默认数据库", type: "number", min: 0, max: 2147483647, advanced: true, wide: true},
        socket,
        tls,
        ...certificates,
    ],
    2: [
        ...address, user, password, database, socket,
        {...tls, options: [{value: "preferred", label: "优先加密"}, ...tls.options],
            description: (config: any) => config?.tls === "preferred"
                ? (config?.ssl_ca || config?.ssl_cert || config?.ssl_key
                    ? "已指定证书，将校验证书且仅使用加密连接。"
                    : "不校验证书；服务端不支持加密时使用普通连接。")
                : undefined},
        ...certificates,
    ],
    3: [
        ...address, user, password, {...database, placeholder: "留空使用 postgres"},
        {name: "sslmode", label: "TLS 连接", type: "select", advanced: true, wide: true, valueAliases: {"": "prefer"}, options: [
            {value: "disable", label: "关闭"},
            {value: "prefer", label: "优先加密"}, {value: "require", label: "要求加密"},
            {value: "verify-ca", label: "验证 CA"}, {value: "verify-full", label: "验证 CA 和主机名"},
            {value: "allow", label: "按需加密"},
        ]},
        ...certificates.map((field) => ({
            ...field, name: field.name === "ssl_ca" ? "sslrootcert" : field.name.replace("_", ""),
            visible: (config: any) => ["allow", "prefer", "require", "verify-ca", "verify-full"].includes(config?.sslmode || "prefer")
                || Boolean(config?.sslrootcert || config?.sslcert || config?.sslkey),
        })),
    ],
    4: [
        ...address,
        {...user, label: "用户名（可选）", placeholder: "无需认证可留空"},
        password,
        {name: "auth_source", label: "认证数据库", type: "text", placeholder: "留空使用 admin", advanced: true, wide: true},
        database,
        {name: "replica_set", label: "副本集名称", type: "text", placeholder: "可选", advanced: true, wide: true},
        {name: "direct_connection", label: "直连节点", type: "boolean", advanced: true, wide: true},
        tls,
        ...certificates,
    ],
    5: [
        ...address, user, password, {...database, placeholder: "留空使用账户默认数据库"},
        {name: "encrypt", label: "TLS 连接", type: "select", advanced: true, wide: true, valueAliases: {"": "true"}, options: [
            {value: "true", label: "开启"},
            {value: "strict", label: "严格加密"}, {value: "false", label: "按服务端协商"},
            {value: "disable", label: "关闭"},
        ]},
        {name: "trust_server_certificate", label: "证书校验", type: "select", advanced: true, wide: true,
            valueAliases: {"": "false"},
            visible: (config: any) => !["disable", "strict"].includes(config?.encrypt),
            options: [{value: "false", label: "验证服务器证书"}, {value: "true", label: "跳过证书验证"}]},
    ],
    6: [
        {name: "path", label: "数据库文件", type: "path", required: true, wide: true},
        {name: "read_only", label: "只读连接", type: "boolean", wide: true},
    ],
    7: [
        ...address,
        {...user, label: "用户名（可选）", placeholder: "无需认证可留空"},
        password,
        {name: "api_key", label: "API Key", type: "password", placeholder: "可选，填写后优先使用", advanced: true, wide: true},
        {...tls, advanced: false, options: [
            {value: "false", label: "HTTP（不加密）"},
            {value: "true", label: "HTTPS（验证证书）"}, {value: "skip-verify", label: "HTTPS（不验证证书）"},
        ]},
        ...certificates.map((field) => ({
            ...field,
            placeholder: field.name === "ssl_ca" ? "自签名证书请选择 http_ca.crt" : field.placeholder,
        })),
    ],
};

export default connectionFields;
