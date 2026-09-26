import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {
    App,
    Button,
    Checkbox,
    Collapse,
    ConfigProvider,
    Empty,
    Form,
    Grid,
    Input,
    InputNumber,
    Segmented,
    Select,
    Spin,
    Switch,
} from "antd";
import {createStyles} from "antd-style";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import AceEditor from "@/components/ace-editor";
import FilePicker from "@/pages/components/file-picker";
import Log from "@/pages/components/task-log";
import {getEnum as getSystemEnum, getHTTPLog, getHTTPService, updateHTTPService} from "@/service/api/system";

const acePath = `${defaultSettings?.basePath || "/"}ace`;

const durationFields: any[] = [
    ["read_timeout", "读取请求超时", "限制读取完整请求的最长时间，0 表示不限制。"],
    ["read_header_timeout", "请求头超时", "防止客户端长时间占用连接，0 使用默认值 10 秒。"],
    ["write_timeout", "响应写入超时", "限制完整响应的写入时间，0 表示不限制。"],
    ["idle_timeout", "空闲连接超时", "连接复用期间允许保持空闲的时间，0 使用默认值 300 秒。"],
    ["grace_period", "优雅关闭时间", "停止或重启时等待现有请求完成，0 使用默认值 10 秒。"],
];

const pathFields: any[] = [
    ["data_path", "运行数据目录", "directory", "保存 HTTP 内核运行数据，留空使用 temp/server/data。"],
    ["cache_path", "响应缓存目录", "directory", "保存网站响应缓存，留空使用 temp/server/cache。"],
    ["log_path", "HTTP 运行日志", "file", "记录服务启动和运行异常；留空使用 temp/server/server.log，填写 - 可关闭。"],
    ["waf_log_path", "WAF 日志", "file", "汇总防火墙检测与拦截事件，留空使用 temp/server/waf.log。"],
    ["config_path", "配置快照文件", "file", "保存实际运行配置用于排障，留空使用 temp/server/config.json。"],
];

const navigationItems = [
    {key: "service", icon: "CloudServerOutlined", label: "服务管理"},
    {key: "basic", icon: "DashboardOutlined", label: "常用设置"},
    {key: "network", icon: "GlobalOutlined", label: "监听管理"},
    {key: "runtime", icon: "ControlOutlined", label: "运行参数"},
    {key: "certificate", icon: "SafetyCertificateOutlined", label: "证书申请"},
    {key: "pages", icon: "FileTextOutlined", label: "默认页面"},
];

const pageItems = [
    {key: "initial_page", label: "初始页面", description: "静态或 PHP 网站尚未部署首页时显示", status: 200},
    {key: "not_found_page", label: "404 页面", description: "站点内资源不存在时显示", status: 404},
    {key: "site_not_found_page", label: "网站不存在", description: "访问的域名没有绑定网站时显示", status: 404},
    {key: "site_disabled_page", label: "网站停用", description: "访问已停用的网站时显示", status: 503},
];

const serviceOperations = [
    {key: "start", label: "启动", icon: "CaretRightOutlined", title: "启动 HTTP 服务", content: "确定启动 HTTP 服务吗？", okText: "启动"},
    {key: "stop", label: "停止", icon: "PauseOutlined", title: "停止 HTTP 服务", content: "停止后所有网站将暂时无法访问，确定继续吗？", okText: "停止", danger: true},
    {key: "restart", label: "重启", icon: "ReloadOutlined", title: "重启 HTTP 服务", content: "重启期间网站会短暂中断，确定继续吗？", okText: "重启"},
    {key: "sync", label: "同步", icon: "SyncOutlined", title: "同步 HTTP 配置", content: "确定根据当前网站配置重新同步 HTTP 服务吗？", okText: "同步"},
];

const httpStatusOptions = [
    ["1xx 信息响应", [
        [100, "Continue"], [101, "Switching Protocols"], [102, "Processing"],
        [103, "Early Hints"], [104, "Upload Resumption Supported"],
    ]],
    ["2xx 成功响应", [
        [200, "OK"], [201, "Created"], [202, "Accepted"], [203, "Non-Authoritative Information"],
        [204, "No Content"], [205, "Reset Content"], [206, "Partial Content"], [207, "Multi-Status"],
        [208, "Already Reported"], [226, "IM Used"],
    ]],
    ["3xx 重定向", [
        [300, "Multiple Choices"], [301, "Moved Permanently"], [302, "Found"], [303, "See Other"],
        [304, "Not Modified"], [305, "Use Proxy"], [306, "Unused"], [307, "Temporary Redirect"],
        [308, "Permanent Redirect"],
    ]],
    ["4xx 客户端错误", [
        [400, "Bad Request"], [401, "Unauthorized"], [402, "Payment Required"], [403, "Forbidden"],
        [404, "Not Found"], [405, "Method Not Allowed"], [406, "Not Acceptable"],
        [407, "Proxy Authentication Required"], [408, "Request Timeout"], [409, "Conflict"], [410, "Gone"],
        [411, "Length Required"], [412, "Precondition Failed"], [413, "Content Too Large"], [414, "URI Too Long"],
        [415, "Unsupported Media Type"], [416, "Range Not Satisfiable"], [417, "Expectation Failed"],
        [418, "Unused"], [421, "Misdirected Request"], [422, "Unprocessable Content"], [423, "Locked"],
        [424, "Failed Dependency"], [425, "Too Early"], [426, "Upgrade Required"],
        [428, "Precondition Required"], [429, "Too Many Requests"], [431, "Request Header Fields Too Large"],
        [451, "Unavailable For Legal Reasons"],
    ]],
    ["5xx 服务端错误", [
        [500, "Internal Server Error"], [501, "Not Implemented"], [502, "Bad Gateway"],
        [503, "Service Unavailable"], [504, "Gateway Timeout"], [505, "HTTP Version Not Supported"],
        [506, "Variant Also Negotiates"], [507, "Insufficient Storage"], [508, "Loop Detected"],
        [510, "Not Extended (Obsoleted)"], [511, "Network Authentication Required"],
    ]],
].map(([label, codes]: any) => ({
    label,
    options: codes.map(([value, text]: any) => ({value, label: `${value} ${text}`})),
}));

const normalizeListenAddresses = (values: any) => (Array.isArray(values) ? values : []).map((value) => {
    const address = String(value || "").trim();
    return /^\d+$/.test(address) ? `:${address.replace(/^0+(?=\d)/, "")}` : address;
});

const listenPortValidator = (_: any, values: any) => {
    const invalid = normalizeListenAddresses(values).find((address: string) => {
        const matched = address.match(/^:(\d+)$/);
        const port = Number(matched?.[1] || 0);
        return matched && (port < 1 || port > 65535);
    });
    return invalid ? Promise.reject(new Error("监听端口范围为 1-65535")) : Promise.resolve();
};

const disclosureTheme: any = {
    token: {fontSize: 12, fontSizeSM: 12},
    components: {
        Button: {contentFontSize: 12, contentFontSizeSM: 12},
        Input: {inputFontSize: 12, inputFontSizeSM: 12},
        Select: {optionFontSize: 12},
    },
};

const toSeconds = (value: any) => {
    const duration = Number(value || 0);
    return Number.isFinite(duration) ? duration / 1_000_000_000 : 0;
};

const toNanoseconds = (value: any) => {
    const seconds = Number(value || 0);
    return Number.isFinite(seconds) ? Math.round(seconds * 1_000_000_000) : 0;
};

const headersToForm = (value: any) => Object.entries(value && typeof value === "object" ? value : {})
    .flatMap(([name, values]: any) => (Array.isArray(values) ? values : [values])
        .map((item: any) => ({name, value: String(item || "")})));

const headersToConfig = (value: any) => (Array.isArray(value) ? value : []).reduce((result: any, item: any) => {
    const name = String(item?.name || "").trim();
    const content = String(item?.value || "").trim();
    if (name && content) {
        (result[name] ||= []).push(content);
    }
    return result;
}, {});

const SectionHeading = ({title, description, styles}: any) => (
    <div className={styles.sectionHeading}>
        <h3>{title}</h3>
        {description && <p>{description}</p>}
    </div>
);

const SettingField = ({styles, title, description, wide, className = "", children, ...props}: any) => (
    <Form.Item
        {...props}
        className={`${styles.settingItem} setting-field ${wide ? "setting-wide" : ""} ${className}`.trim()}
        label={title}
        extra={description}
        colon={false}>
        {children}
    </Form.Item>
);

const SettingRow = ({styles, title, description, wide, children}: any) => (
    <div className={`${styles.settingRow} setting-row ${wide ? "setting-wide" : ""}`}>
        <div className="field-title">{title}</div>
        <div className="setting-control">{children}</div>
        {description && <div className="field-description">{description}</div>}
    </div>
);

const DisclosureLabel = ({title, description}: any) => (
    <span className="disclosure-label">
        <span>{title}</span>
        {description && <small>{description}</small>}
    </span>
);

const HeaderEditor = ({name, styles}: any) => (
    <Form.List name={[name, "headers"]}>
        {(fields, {add, remove}) => (
            <div className={styles.headerEditor}>
                {fields.map(({key, name: field}) => (
                    <div className="header-row" key={key}>
                        <Form.Item name={[field, "name"]} rules={[{required: true, whitespace: true, message: "请输入响应头名称"}]}>
                            <Input placeholder="名称" autoComplete="off"/>
                        </Form.Item>
                        <Form.Item name={[field, "value"]} rules={[{required: true, whitespace: true, message: "请输入响应头内容"}]}>
                            <Input placeholder="内容" autoComplete="off"/>
                        </Form.Item>
                        <Button
                            className="header-remove"
                            color="danger"
                            variant="filled"
                            style={{borderColor: "transparent"}}
                            aria-label="删除响应头"
                            icon={<Icon type="CloseOutlined"/>}
                            onClick={() => remove(field)}/>
                    </div>
                ))}
                <Button color="default" variant="filled" style={{borderColor: "transparent"}} icon={<Icon type="PlusOutlined"/>} onClick={() => add({name: "", value: ""})}>
                    添加响应头
                </Button>
            </div>
        )}
    </Form.List>
);

const useStyles = createStyles(({css, token}: any, props: any = {}) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-container {
            height: ${props?.compact ? 480 : 520}px;
            min-height: ${props?.compact ? 480 : 520}px;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .ant-modal-header, .ant-modal-footer {
            flex: none;
        }

        .ant-modal-body {
            min-height: 0;
            padding: 0 !important;
            display: flex;
            flex: 1 1 auto;
            overflow: hidden;
        }

        .ant-modal-body > div,
        .ant-modal-body > div > .ant-spin,
        .ant-modal-body > div > .ant-spin > .ant-spin-container,
        .ant-modal-body > div > .ant-spin-nested-loading,
        .ant-modal-body > div > .ant-spin-nested-loading > .ant-spin-container {
            width: 100%;
            min-height: 0;
            display: flex;
            flex: 1 1 auto;
            flex-direction: column;
            overflow: hidden;
        }

        .ant-modal-footer {
            margin-top: 0;
            padding-top: ${props?.compact ? 10 : 12}px;
        }
    `,
    workbench: css`
        box-sizing: border-box;
        width: 100%;
        height: auto;
        min-width: 0;
        min-height: 0;
        display: grid;
        flex: 1 1 auto;
        grid-template-columns: ${props?.compact ? 156 : 166}px minmax(0, 1fr);
        grid-template-rows: minmax(0, 1fr);
        overflow: hidden;

        &[data-section="service"] {
            min-height: ${props?.compact ? 250 : 280}px;
        }

        @media (max-width: 760px) {
            grid-template-columns: minmax(0, 1fr);
            grid-template-rows: auto minmax(0, 1fr);
        }
    `,
    navigation: css`
        box-sizing: border-box;
        min-width: 0;
        min-height: 0;
        max-height: calc(100% - ${props?.compact ? 20 : 24}px);
        margin: ${props?.compact ? 10 : 12}px 0;
        padding: ${props?.compact ? 6 : 7}px;
        display: flex;
        align-self: start;
        overflow: hidden;
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};

        .navigation-scroll {
            min-width: 0;
            min-height: 0;
            flex: 1;
            overflow-x: hidden;
            overflow-y: auto;
            scrollbar-color: ${token.colorBorderSecondary} transparent;
            scrollbar-width: thin;
        }

        .navigation-grid {
            min-width: 0;
            display: flex;
            flex-direction: column;
            gap: 3px;
        }

        .navigation-item {
            width: 100%;
            min-width: 0;
            min-height: ${props?.compact ? 34 : 38}px;
            padding: 0 10px;
            display: flex;
            align-items: center;
            gap: 8px;
            flex: none;
            overflow: hidden;
            border: 0;
            border-radius: ${token.borderRadius}px;
            outline: 0;
            appearance: none;
            background: transparent;
            color: ${token.colorTextSecondary};
            cursor: pointer;
            font: inherit;
            font-size: ${props?.compact ? 12 : 13}px;
            text-align: left;
            transition: background-color .16s ease, color .16s ease;
        }

        .navigation-item:hover {
            background: ${token.colorFillTertiary};
        }

        .navigation-item:focus-visible {
            box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
        }

        .navigation-item[data-active="true"] {
            background: color-mix(in srgb, ${token.colorPrimaryBg} 84%, ${token.colorBgContainer});
            color: ${token.colorTextHeading};
            font-weight: 600;
        }

        .navigation-item .anticon {
            width: 15px;
            flex: none;
            color: ${token.colorTextTertiary};
            font-size: 15px;
            line-height: 1;
        }

        .navigation-item[data-active="true"] .anticon {
            color: ${token.colorPrimary};
        }

        .navigation-label {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @media (max-width: 760px) {
            width: 100%;
            max-height: none;
            margin: 0;
            padding: 5px 0;
            align-self: stretch;
            overflow: hidden;
            border-radius: 0;
            background: transparent;

            .navigation-scroll {
                width: 100%;
                max-width: 100%;
                flex: 1 1 auto;
                overflow-x: scroll;
                overflow-y: hidden;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
                -webkit-overflow-scrolling: touch;
                touch-action: pan-x;
            }

            .navigation-scroll::-webkit-scrollbar { display: none; }

            .navigation-grid {
                width: max-content;
                min-width: max-content;
                flex-direction: row;
                gap: 6px;
            }

            .navigation-item {
                width: auto;
                min-width: max-content;
                min-height: ${props?.compact ? 30 : 34}px;
                padding: 0 9px;
                gap: 6px;
                background: ${token.colorFillQuaternary};
            }
        }
    `,
    mainPane: css`
        min-width: 0;
        min-height: 0;
        height: auto;
        display: flex;
        overflow: hidden;
        container-name: http-service-pane;
        container-type: inline-size;
    `,
    content: css`
        box-sizing: border-box;
        min-width: 0;
        min-height: 0;
        height: auto;
        padding: ${props?.compact ? "15px 18px 18px" : "18px 22px 22px"};
        flex: 1 1 auto;
        overflow: auto;
        overscroll-behavior: contain;
        scrollbar-color: ${token.colorBorderSecondary} transparent;
        scrollbar-width: thin;
        -webkit-overflow-scrolling: touch;
        touch-action: pan-y;

        .ant-input-number, .ant-select {
            width: 100%;
        }

        .field-title {
            color: ${token.colorText};
            font-size: 13px;
            font-weight: 500;
        }

        .field-description {
            color: ${token.colorTextTertiary};
            font-size: 12px;
            font-weight: 400;
            line-height: 18px;
        }

        .setting-field :where(input, textarea)::placeholder,
        .setting-field .ant-select-placeholder {
            font-size: ${props?.compact ? 12 : 13}px !important;
        }

        @media (max-width: 760px) {
            padding: ${props?.compact ? "12px 12px 16px" : "15px 12px 18px"};
        }
    `,
    contentInner: css`
        width: 100%;
        max-width: 700px;
        margin: 0 auto;
    `,
    section: css`
        min-width: 0;
        padding-bottom: 2px;
    `,
    sectionHeading: css`
        margin-bottom: ${props?.compact ? 10 : 13}px;

        h3 {
            margin: 0;
            color: ${token.colorTextHeading};
            font-size: ${props?.compact ? 14 : 15}px;
            font-weight: 600;
            line-height: 22px;
        }

        p {
            margin: 2px 0 0;
            color: ${token.colorTextTertiary};
            font-size: 12px;
            line-height: 18px;
        }
    `,
    settingItem: css`
        && {
            margin: 0;
            padding: ${props?.compact ? "4px 0" : "6px 0"};
        }

        && .ant-form-item-row {
            display: block;
        }

        && .ant-form-item-label {
            padding: 0 0 ${props?.compact ? 5 : 6}px;
            text-align: left;
            white-space: normal;
        }

        && .ant-form-item-label > label {
            height: auto;
            color: ${token.colorText};
            font-size: 13px;
            font-weight: 500;
            line-height: 20px;
            white-space: normal;
        }

        && .ant-form-item-label > label::after {
            display: none;
        }

        && .ant-form-item-control {
            min-width: 0;
            width: 100%;
            max-width: none;
        }

        && .ant-form-item-control-input {
            width: 100%;
            max-width: none;
        }

        && .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            width: 100%;
            max-width: ${props?.compact ? 240 : 260}px;
        }

        &&.setting-wide .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            max-width: ${props?.compact ? 300 : 320}px;
        }

        && .ant-form-item-explain {
            margin-top: 4px;
            font-size: 11px;
        }

        && .ant-form-item-extra {
            margin-top: ${props?.compact ? 4 : 5}px;
            color: ${token.colorTextTertiary};
            font-size: 12px;
            line-height: 18px;
        }
    `,
    settingRow: css`
        min-width: 0;
        padding: ${props?.compact ? "7px 0" : "9px 0"};

        .field-title {
            margin-bottom: ${props?.compact ? 5 : 6}px;
            line-height: 20px;
        }

        .setting-control {
            min-width: 0;
            width: 100%;
            max-width: none;
        }

        .field-description {
            margin-top: ${props?.compact ? 4 : 5}px;
        }
    `,
    disclosure: css`
        margin-top: 2px;
        border: 0 !important;
        background: transparent !important;

        .ant-collapse-item {
            border: 0 !important;
        }

        .ant-collapse-header {
            min-height: 0;
            padding: ${props?.compact ? "4px 0" : "5px 0"} !important;
            align-items: center !important;
            color: ${token.colorTextSecondary} !important;
            border-radius: ${token.borderRadiusSM}px !important;
            transition: color .16s ease, background-color .16s ease !important;
        }

        .ant-collapse-header:hover,
        .ant-collapse-header:focus,
        .ant-collapse-header:focus-visible {
            outline: 0 !important;
            background: transparent !important;
            box-shadow: none !important;
        }

        .ant-collapse-item-active > .ant-collapse-header {
            color: ${token.colorText} !important;
        }

        .ant-collapse-expand-icon {
            color: ${token.colorTextTertiary};
        }

        .ant-collapse-item-active .ant-collapse-expand-icon {
            color: ${token.colorPrimary};
        }

        .ant-collapse-body {
            margin: 1px 0 ${props?.compact ? 3 : 4}px;
            padding: ${props?.compact ? "4px 7px" : "6px 9px"} !important;
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorFillQuaternary};
        }

        .ant-collapse-body .ant-form-item-label {
            padding-bottom: 3px !important;
        }

        .ant-collapse-body .ant-form-item-label > label,
        .ant-collapse-body .field-title {
            font-size: 12px !important;
            line-height: 18px;
        }

        .ant-collapse-body .ant-form-item-extra,
        .ant-collapse-body .field-description {
            margin-top: 3px !important;
            font-size: 11px !important;
            line-height: 16px;
        }

        .ant-collapse-body :where(
            .ant-input,
            .ant-input-number-input,
            .ant-input-prefix,
            .ant-input-suffix,
            .ant-btn,
            .ant-select,
            .ant-select-selection-search-input,
            .ant-select-selection-item,
            .ant-select-selection-placeholder,
            .ant-select-placeholder
        ) {
            font-size: 12px !important;
        }

        .ant-collapse-body :where(input, textarea)::placeholder,
        .ant-collapse-body .ant-select-selection-search-input::placeholder,
        .ant-collapse-body .setting-field .ant-select-placeholder {
            font-size: 12px !important;
        }

        .ant-collapse-body :where(.ant-btn, .ant-input-prefix, .ant-input-suffix, .ant-select-arrow) .anticon,
        .ant-collapse-body > .anticon {
            font-size: 12px !important;
        }

        .ant-collapse-body > .setting-field,
        .ant-collapse-body > .setting-row {
            padding-top: ${props?.compact ? 2 : 3}px;
            padding-bottom: ${props?.compact ? 2 : 3}px;
        }

        .ant-collapse-body > .setting-field .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            max-width: ${props?.compact ? 230 : 250}px;
        }

        @media (max-width: 760px) {
            .ant-collapse-body > .setting-field .ant-form-item-control-input-content > :where(
                .ant-input,
                .ant-input-affix-wrapper,
                .ant-input-number,
                .ant-select,
                .ant-space-compact
            ) {
                max-width: none;
            }
        }

        .ant-collapse-body > .setting-field:first-child,
        .ant-collapse-body > .setting-row:first-child {
            padding-top: 0;
        }

        .ant-collapse-body > .setting-field:last-child,
        .ant-collapse-body > .setting-row:last-child {
            padding-bottom: 0;
        }

        .disclosure-label {
            min-width: 0;
            display: grid;
            gap: 2px;
            font-size: 13px;
            font-weight: 500;
            line-height: 18px;
        }

        .disclosure-label small {
            color: ${token.colorTextTertiary};
            font-size: 12px;
            font-weight: 400;
            line-height: 18px;
        }
    `,
    serviceActions: css`
        display: flex;
        flex-wrap: wrap;
        gap: ${props?.compact ? 5 : 6}px;

        .service-action.ant-btn {
            width: auto;
            min-width: max-content;
            height: ${props?.compact ? 28 : 32}px;
            padding: 0 ${props?.compact ? 8 : 10}px;
            gap: 5px;
            font-size: ${props?.compact ? 11 : 12}px;
        }
    `,
    serviceStatus: css`
        min-height: ${props?.compact ? 28 : 32}px;
        display: inline-flex;
        align-items: center;
        gap: 8px;
        color: ${token.colorTextSecondary};
        font-size: 13px;

        .status-dot {
            width: 7px;
            height: 7px;
            flex: none;
            border-radius: 50%;
            background: ${token.colorTextQuaternary};
            box-shadow: 0 0 0 3px color-mix(in srgb, ${token.colorTextQuaternary}, transparent 84%);
        }

        &[data-running="true"] {
            color: ${token.colorSuccess};
        }

        &[data-running="true"] .status-dot {
            background: ${token.colorSuccess};
            box-shadow: 0 0 0 3px color-mix(in srgb, ${token.colorSuccess}, transparent 84%);
        }
    `,
    infoNote: css`
        margin-top: ${props?.compact ? 10 : 12}px;
        padding: ${props?.compact ? "8px 9px" : "9px 10px"};
        display: flex;
        align-items: flex-start;
        gap: 7px;
        border-left: 2px solid ${token.colorPrimary};
        background: ${token.colorFillQuaternary};
        color: ${token.colorTextSecondary};
        font-size: 12px;
        line-height: 18px;

        .anticon {
            margin-top: 2px;
            flex: none;
            color: ${token.colorPrimary};
        }
    `,
    errorState: css`
        min-height: ${props?.compact ? 280 : 320}px;
        height: 100%;
        padding: 24px;
        display: flex;
        align-items: center;
        justify-content: center;

        .ant-empty-description {
            color: ${token.colorTextSecondary};
        }
    `,
    protocolGroup: css`
        display: flex;
        flex-wrap: wrap;
        gap: ${props?.compact ? 5 : 6}px;

        .ant-checkbox-wrapper {
            min-height: ${props?.compact ? 28 : 32}px;
            margin-inline-start: 0;
            padding: 0 ${props?.compact ? 8 : 10}px;
            display: inline-flex;
            align-items: center;
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorFillQuaternary};
            color: ${token.colorTextSecondary};
            font-size: 12px;
            transition: background-color .16s ease, color .16s ease;
        }

        .ant-checkbox-wrapper:hover {
            background: ${token.colorFillSecondary};
            color: ${token.colorText};
        }

        .ant-checkbox-wrapper-checked {
            background: ${token.colorPrimaryBg};
            color: ${token.colorText};
        }

        .ant-checkbox-wrapper-checked:hover {
            background: ${token.colorPrimaryBgHover};
        }
    `,
    pageSwitcher: css`
        width: 100%;
        margin-bottom: ${props?.compact ? 10 : 12}px;

        .ant-segmented-item-label {
            padding-inline: ${props?.compact ? 8 : 10}px;
            font-size: 12px;
        }

        @media (max-width: 760px) {
            overflow-x: auto;
            scrollbar-width: none;
            -webkit-overflow-scrolling: touch;

            &::-webkit-scrollbar { display: none; }

            &.ant-segmented-block .ant-segmented-group {
                width: max-content;
                min-width: 100%;
            }

            &.ant-segmented-block .ant-segmented-item {
                min-width: max-content;
                flex: 1 0 auto;
            }
        }
    `,
    pageEditor: css`
        min-width: 0;

        .editor-field {
            margin: ${props?.compact ? "10px 0" : "12px 0"};
        }

        .editor-label {
            margin-bottom: 7px;
            display: grid;
            gap: 2px;
        }

        .editor-label strong {
            color: ${token.colorText};
            font-size: 13px;
            font-weight: 500;
        }

        .editor-label span {
            color: ${token.colorTextTertiary};
            font-size: 12px;
            line-height: 18px;
        }

        .editor-field > .ant-form-item {
            margin: 0;
        }
    `,
    headerEditor: css`
        --header-control-height: 26px;
        display: grid;
        gap: 6px;

        .header-row {
            display: grid;
            grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) var(--header-control-height);
            gap: 6px;
            align-items: start;
        }

        .header-row .ant-form-item {
            min-width: 0;
            margin: 0;
            padding: 0;
        }

        .header-row .ant-form-item-control-input {
            min-height: var(--header-control-height);
        }

        .header-row .ant-input {
            height: var(--header-control-height);
            padding-top: 0;
            padding-bottom: 0;
        }

        .header-remove.ant-btn {
            width: var(--header-control-height);
            height: var(--header-control-height);
            min-width: 0;
            padding: 0;
            align-self: start;
        }

        @container http-service-pane (max-width: 360px) {
            .header-row {
                grid-template-columns: minmax(0, 1fr) var(--header-control-height);
            }

            .header-row .ant-form-item:nth-child(2) {
                grid-column: 1 / -1;
                grid-row: 2;
            }

            .header-remove {
                grid-column: 2;
                grid-row: 1;
            }
        }
    `,
    code: css`
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;

        .ace_editor {
            border-radius: ${token.borderRadius}px !important;
        }
    `,
}));

const HTTPService = forwardRef<any, any>((_, ref): any => {
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const modalRef = useRef<any>(null);
    const logRef = useRef<any>(null);
    const navigationRef = useRef<HTMLDivElement>(null);
    const requestRef = useRef(0);
    const [form] = Form.useForm();
    const [active, setActive] = useState(false);
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [running, setRunning] = useState(false);
    const [action, setAction] = useState("");
    const [loadError, setLoadError] = useState("");
    const [siteOptions, setSiteOptions] = useState<any[]>([]);
    const [section, setSection] = useState("service");
    const [pageSection, setPageSection] = useState("initial_page");
    const [networkPanels, setNetworkPanels] = useState<string[]>([]);
    const [runtimePanels, setRuntimePanels] = useState<string[]>([]);
    const [pagePanels, setPagePanels] = useState<string[]>([]);

    const fillConfig = useCallback((config: any) => {
        const values: any = {...(config || {})};
        durationFields.forEach(([name]) => {
            values[name] = toSeconds(config?.[name]);
        });
        pathFields.forEach(([name]) => {
            values[name] = String(config?.[name] ?? "");
        });
        values.http_listen = normalizeListenAddresses(config?.http_listen);
        values.https_listen = normalizeListenAddresses(config?.https_listen);
        const maxHeaderBytes = Number(config?.max_header_bytes || 0);
        values.max_header_bytes = Number.isFinite(maxHeaderBytes) ? maxHeaderBytes / 1024 : 0;
        pageItems.forEach(({key, status}) => {
            values[key] = {
                status: Number(config?.[key]?.status || status),
                body: String(config?.[key]?.body || ""),
                headers: headersToForm(config?.[key]?.headers),
            };
        });
        values.default_site = config?.default_site ? String(config.default_site) : undefined;
        form.setFieldsValue(values);
    }, [form]);

    const load = useCallback(async (showLoading = true, includeSiteOptions = false) => {
        const requestId = ++requestRef.current;
        if (showLoading) {
            setLoading(true);
        }
        setLoadError("");
        try {
            const [response, enumResponse] = await Promise.all([
                getHTTPService({body: {action: "get"}}),
                includeSiteOptions
                    ? getSystemEnum({body: {name: "site"}})
                    : Promise.resolve(null),
            ]);
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                const error = response?.message || "获取 HTTP 服务配置失败";
                setLoadError(error);
                message.error(error);
                return;
            }
            const data: any = response?.data || {};
            setRunning(Boolean(data.running));
            if (includeSiteOptions) {
                if (enumResponse?.code === 200) {
                    setSiteOptions(Object.entries(enumResponse?.data?.name || {}).map(([value, label]) => ({
                        label: String(label),
                        value: String(value),
                    })).sort((first, second) => first.label.localeCompare(second.label, "zh-CN")));
                } else {
                    setSiteOptions([]);
                    message.error(enumResponse?.message || "获取默认站点选项失败");
                }
            }
            fillConfig(data.config || {});
        } catch {
            if (requestRef.current === requestId) {
                setLoadError("无法连接服务器，请检查网络后重试");
                message.error("获取 HTTP 服务配置失败");
            }
        } finally {
            if (requestRef.current === requestId && showLoading) {
                setLoading(false);
            }
        }
    }, [fillConfig, message]);

    const open = useCallback(() => {
        requestRef.current += 1;
        setActive(true);
        setLoading(true);
        setAction("");
        setLoadError("");
        setSection("service");
        setPageSection("initial_page");
        setNetworkPanels([]);
        setRuntimePanels([]);
        setPagePanels([]);
        modalRef.current?.show();
    }, []);

    useEffect(() => {
        if (active) {
            form.resetFields();
            void load(true, true);
        }
    }, [active, form, load]);

    useEffect(() => {
        const element = navigationRef.current;
        if (!active || !element) return;
        let frame = requestAnimationFrame(() => {
            const item = element.querySelector(`[data-section="${section}"]`);
            item?.scrollIntoView({behavior: "smooth", block: "nearest", inline: "nearest"});
        });
        const handleWheel = (event: WheelEvent) => {
            const maxScrollLeft = Math.max(0, element.scrollWidth - element.clientWidth);
            if (maxScrollLeft <= 0) return;
            const delta = Math.abs(event.deltaX) > Math.abs(event.deltaY) ? event.deltaX : event.deltaY;
            const canScroll = delta < 0 ? element.scrollLeft > 0 : element.scrollLeft < maxScrollLeft;
            if (!canScroll) return;
            event.preventDefault();
            element.scrollLeft += delta;
        };
        const observer = new ResizeObserver(() => {
            cancelAnimationFrame(frame);
            frame = requestAnimationFrame(() => {
                const item = element.querySelector(`[data-section="${section}"]`);
                item?.scrollIntoView({behavior: "auto", block: "nearest", inline: "nearest"});
            });
        });
        element.addEventListener("wheel", handleWheel, {passive: false});
        observer.observe(element);
        return () => {
            cancelAnimationFrame(frame);
            observer.disconnect();
            element.removeEventListener("wheel", handleWheel);
        };
    }, [active, section]);

    useImperativeHandle(ref, () => ({
        open,
        refresh: () => active ? load(false, false) : Promise.resolve(),
    }), [active, load, open]);

    const runAction = useCallback(async (nextAction: string) => {
        if (action) {
            return;
        }
        const requestId = ++requestRef.current;
        setAction(nextAction);
        try {
            const response = await updateHTTPService({body: {action: nextAction}});
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "操作失败");
                return;
            }
            setRunning(Boolean(response?.data?.running));
            message.success(response?.message || "操作成功");
        } catch {
            if (requestRef.current === requestId) {
                message.error("操作失败");
            }
        } finally {
            if (requestRef.current === requestId) {
                setAction("");
            }
        }
    }, [action, message]);

    const submit = useCallback(async (submitted: any) => {
        const values = {...form.getFieldsValue(true), ...submitted};
        const config: any = {...values, default_site: String(values.default_site || "")};
        pageItems.forEach(({key, status}) => {
            config[key] = {
                status: Number(values[key]?.status || status),
                body: String(values[key]?.body || ""),
                headers: headersToConfig(values[key]?.headers),
            };
        });
        durationFields.forEach(([name]) => {
            config[name] = toNanoseconds(values[name]);
        });
        pathFields.forEach(([name]) => {
            config[name] = String(values[name] ?? "");
        });
        config.http_listen = normalizeListenAddresses(values.http_listen);
        config.https_listen = normalizeListenAddresses(values.https_listen);
        config.max_header_bytes = Math.round(Number(values.max_header_bytes || 0) * 1024);
        const requestId = ++requestRef.current;
        setSaving(true);
        try {
            const response = await updateHTTPService({body: {action: "set", config}});
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "保存 HTTP 服务配置失败");
                return;
            }
            setRunning(Boolean(response?.data?.running));
            message.success(response?.message || "保存成功");
        } catch {
            if (requestRef.current === requestId) {
                message.error("保存 HTTP 服务配置失败");
            }
        } finally {
            if (requestRef.current === requestId) {
                setSaving(false);
            }
        }
    }, [form, message]);

    const close = useCallback(() => {
        requestRef.current += 1;
        setLoading(false);
        setSaving(false);
        setAction("");
        modalRef.current?.hide();
    }, []);

    const protocolValidator = useCallback((_: any, values: any) => {
        const protocols = Array.isArray(values) ? values : [];
        if (protocols.length === 0) {
            return Promise.reject(new Error("请至少启用一种 HTTP 协议"));
        }
        if ((protocols.includes("h2") || protocols.includes("h2c")) && !protocols.includes("h1")) {
            return Promise.reject(new Error("启用 HTTP/2 或 h2c 时必须同时启用 HTTP/1.1"));
        }
        const httpListen = form.getFieldValue("http_listen");
        if (Array.isArray(httpListen) && httpListen.length > 0 && !protocols.some((item: any) => item === "h1" || item === "h2c")) {
            return Promise.reject(new Error("启用 HTTP 监听时至少需要 HTTP/1.1 或 HTTP/2 明文"));
        }
        return Promise.resolve();
    }, [form]);

    const httpListenValidator = useCallback((_: any, addresses: any) => {
        const protocols = form.getFieldValue("protocols");
        if (Array.isArray(addresses) && addresses.length > 0
            && (!Array.isArray(protocols)
                || !protocols.some((item: any) => item === "h1" || item === "h2c"))) {
            return Promise.reject(new Error("HTTP 监听至少需要启用 HTTP/1.1 或 HTTP/2 明文"));
        }
        return Promise.resolve();
    }, [form]);

    const serviceContent = section === "service" && (
        <div className={styles.section}>
            <SectionHeading styles={styles} title="服务管理" description="管理承载全部网站的 HTTP 服务，操作前会再次确认。"/>
            <SettingRow
                styles={styles}
                title="运行状态"
                description="HTTP 服务负责网站监听、转发、缓存和防火墙配置。">
                <span className={styles.serviceStatus} data-running={running}>
                    <i className="status-dot"/>
                    {running ? "运行中" : "已停止"}
                </span>
            </SettingRow>
            <SettingRow
                styles={styles}
                wide
                title="服务操作"
                description="重启会短暂中断访问；同步仅重新加载当前网站配置。">
                <div className={styles.serviceActions} role="toolbar" aria-label="HTTP 服务操作">
                    {serviceOperations.map((item) => (
                        <Button
                            key={item.key}
                            className="service-action"
                            color={item.key === "stop" ? "danger" : "default"}
                            variant="filled"
                            style={{borderColor: "transparent"}}
                            htmlType="button"
                            icon={<Icon type={item.icon}/>}
                            disabled={saving || Boolean(action) || (item.key === "start" ? running : item.key === "stop" ? !running : false)}
                            loading={action === item.key}
                            onClick={() => modal.confirm({
                                title: item.title,
                                content: item.content,
                                okText: item.okText,
                                cancelText: "取消",
                                okButtonProps: item.danger ? {danger: true, type: "default"} : undefined,
                                mask: {closable: true},
                                onOk: () => runAction(item.key),
                            } as any)}>
                            {item.label}
                        </Button>
                    ))}
                </div>
            </SettingRow>
            <SettingRow
                styles={styles}
                title="运行日志"
                description="查看服务启动、配置加载和运行异常记录。">
                <div className={styles.serviceActions}>
                    <Button
                        className="service-action"
                        color="default"
                        variant="filled"
                        style={{borderColor: "transparent"}}
                        htmlType="button"
                        icon={<Icon type="FileTextOutlined"/>}
                        disabled={Boolean(action)}
                        onClick={() => logRef.current?.open({id: "server", name: "HTTP 服务"})}>
                        日志
                    </Button>
                </div>
            </SettingRow>
        </div>
    );

    const basicContent = section === "basic" && (
        <div className={styles.section}>
            <SectionHeading styles={styles} title="常用设置" description="这些配置保存后会立即应用到当前 HTTP 服务。"/>
            <SettingField
                styles={styles}
                name="default_site"
                title="默认站点"
                description="域名未绑定网站时转发到这里，留空显示“网站不存在”页面。">
                <Select
                    allowClear
                    showSearch={false}
                    options={siteOptions}
                    placeholder="请选择默认站点，留空表示关闭"/>
            </SettingField>
            <SettingField
                styles={styles}
                name="protocols"
                title="启用协议"
                description="HTTP/1.1 兼容性最好；HTTP/2 和 HTTP/3 可提升并发传输效率。"
                rules={[{validator: protocolValidator}]}>
                <Checkbox.Group className={styles.protocolGroup} options={[
                    {label: "HTTP/1.1", value: "h1"},
                    {label: "HTTP/2", value: "h2"},
                    {label: "HTTP/2 明文", value: "h2c"},
                    {label: "HTTP/3", value: "h3"},
                ]}/>
            </SettingField>
            <SettingField
                styles={styles}
                name="log_level"
                title="日志级别"
                description="只控制 HTTP 服务运行日志；日常使用 INFO，DEBUG 仅用于排障。"
                rules={[{required: true, message: "请选择日志级别"}]}>
                <Select
                    options={["DEBUG", "INFO", "WARN", "ERROR", "PANIC", "FATAL"].map((value) => ({label: value, value}))}
                    placeholder="请选择日志级别"/>
            </SettingField>
        </div>
    );

    const networkContent = section === "network" && (
        <div className={styles.section}>
            <SectionHeading styles={styles} title="监听管理" description="配置网站对外提供服务的地址，以及经过 CDN 或反向代理时的客户端 IP 识别。"/>
            <SettingField
                styles={styles}
                wide
                name="http_listen"
                title="HTTP 监听地址"
                description="支持多个 IP:端口，例如 :80；删除全部地址会关闭 HTTP 监听。"
                normalize={normalizeListenAddresses}
                rules={[{validator: listenPortValidator}, {validator: httpListenValidator}]}>
                <Select mode="tags" tokenSeparators={[",", " "]} open={false} placeholder="请输入监听地址，例如 :80"/>
            </SettingField>
            <SettingField
                styles={styles}
                wide
                name="https_listen"
                title="HTTPS 监听地址"
                description="支持多个 IP:端口，例如 :443；网站需要绑定证书才能正常提供 HTTPS。"
                normalize={normalizeListenAddresses}
                rules={[{validator: listenPortValidator}]}>
                <Select mode="tags" tokenSeparators={[",", " "]} open={false} placeholder="请输入监听地址，例如 :443"/>
            </SettingField>
            <Collapse
                ghost
                className={styles.disclosure}
                expandIconPlacement="end"
                activeKey={networkPanels}
                onChange={(keys) => setNetworkPanels((Array.isArray(keys) ? keys : [keys]).map(String))}
                items={[{
                    key: "proxy",
                    label: <DisclosureLabel title="可信代理（高级）" description="仅在服务器前方使用 CDN、负载均衡或反向代理时配置"/>,
                    children: (
                        <ConfigProvider componentSize="small" theme={disclosureTheme}>
                            <SettingField
                                styles={styles}
                                wide
                                name="trusted_proxies"
                                title="可信代理"
                                description="只填写你能控制的代理出口 IP 或 CIDR，配置错误可能导致客户端 IP 被伪造。">
                                <Select mode="tags" tokenSeparators={[",", " "]} open={false} placeholder="请输入 IP 或 CIDR，例如 127.0.0.1/32"/>
                            </SettingField>
                            <SettingField
                                styles={styles}
                                wide
                                name="client_ip_headers"
                                title="客户端 IP 请求头"
                                description="按顺序读取代理写入的真实 IP 请求头，常用 X-Forwarded-For。">
                                <Select mode="tags" tokenSeparators={[",", " "]} open={false} placeholder="请输入请求头，例如 X-Forwarded-For"/>
                            </SettingField>
                            <SettingRow
                                styles={styles}
                                title="严格解析代理链"
                                description="从代理链右侧开始解析，只信任连续经过可信代理的地址。">
                                <Form.Item name="trusted_proxies_strict" valuePropName="checked" noStyle>
                                    <Switch/>
                                </Form.Item>
                            </SettingRow>
                        </ConfigProvider>
                    ),
                }]}/>
        </div>
    );

    const runtimeContent = section === "runtime" && (
        <div className={styles.section}>
            <SectionHeading styles={styles} title="运行参数" description="默认值适合大多数网站，仅在明确遇到超时或请求头限制时调整。"/>
            <SettingField
                styles={styles}
                name="max_header_bytes"
                title="请求头大小"
                description="单个请求允许的请求头上限，默认 1024 KB。"
                rules={[{type: "number", min: 0, message: "不能小于 0"}]}>
                <InputNumber min={0} precision={0} suffix="KB" placeholder="请输入请求头大小"/>
            </SettingField>
            <Collapse
                ghost
                className={styles.disclosure}
                expandIconPlacement="end"
                activeKey={runtimePanels}
                onChange={(keys) => setRuntimePanels((Array.isArray(keys) ? keys : [keys]).map(String))}
                items={[{
                    key: "timeouts",
                    label: <DisclosureLabel title="连接与超时" description="5 项 · 数值单位均为秒，0 的具体含义见各项说明"/>,
                    children: (
                        <ConfigProvider componentSize="small" theme={disclosureTheme}>
                            {durationFields.map(([name, label, description]) => (
                                <SettingField
                                    key={name}
                                    styles={styles}
                                    name={name}
                                    title={label}
                                    description={description}
                                    rules={[{type: "number", min: 0, transform: (value: any) => value === "" ? undefined : Number(value), message: "请输入不小于 0 的数字"}]}>
                                    <Input suffix="秒" inputMode="decimal" placeholder="请输入秒数" autoComplete="off"/>
                                </SettingField>
                            ))}
                        </ConfigProvider>
                    ),
                }, {
                    key: "paths",
                    label: <DisclosureLabel title="数据与日志路径" description="5 项 · 留空时自动使用系统默认路径"/>,
                    children: (
                        <>
                            {pathFields.map(([name, label, mode, description]) => (
                                <SettingField
                                    key={name}
                                    styles={styles}
                                    wide
                                    name={name}
                                    title={label}
                                    description={description}>
                                    <FilePicker
                                        mode={mode}
                                        size="small"
                                        buttonProps={{color: "default", variant: "filled", style: {borderColor: "transparent"}}}
                                        placeholder={`请输入或选择${label}，留空使用默认路径`}
                                        autoComplete="off"/>
                                    </SettingField>
                            ))}
                        </>
                    ),
                }]}/>
        </div>
    );

    const certificateContent = section === "certificate" && (
        <div className={styles.section}>
            <SectionHeading styles={styles} title="证书申请" description="这里仅配置全局 ACME HTTP-01 验证参数，证书申请和网站绑定仍在证书管理中完成。"/>
            <SettingField
                styles={styles}
                name="acme_email"
                title="账户邮箱"
                description="用于 ACME 账户注册和证书到期提醒，可留空。"
                rules={[{type: "email", message: "邮箱格式不正确"}]}>
                <Input placeholder="例如 admin@example.com，可留空" autoComplete="off"/>
            </SettingField>
            <SettingField
                styles={styles}
                name="http_challenge_host"
                title="验证监听地址"
                description="限制验证服务监听的本机地址，留空表示监听所有可用地址。">
                <Input placeholder="例如 0.0.0.0，留空监听所有地址" autoComplete="off"/>
            </SettingField>
            <SettingField
                styles={styles}
                name="http_challenge_port"
                title="验证端口"
                description="HTTP-01 验证使用的本机端口，默认 80，并需能被公网访问。"
                rules={[{type: "number", min: 1, max: 65535, message: "端口范围为 1-65535"}]}>
                <InputNumber min={1} max={65535} precision={0} placeholder="请输入验证端口" style={{width: "100%"}}/>
            </SettingField>
            <div className={styles.infoNote}>
                <Icon type="InfoCircleOutlined"/>
                ACME 账户与签发数据由系统保存在 data/acme；修改运行数据或缓存目录不会影响账户信息。
            </div>
        </div>
    );

    const pagesContent = section === "pages" && (
        <div className={styles.section}>
            <SectionHeading styles={styles} title="默认页面" description="自定义网站初始页面、资源不存在、域名未绑定和网站停用时的页面内容。"/>
            <Segmented
                block
                className={styles.pageSwitcher}
                value={pageSection}
                onChange={(value) => setPageSection(String(value))}
                options={pageItems.map((item) => ({label: item.label, value: item.key}))}/>
            {pageItems.filter((item) => pageSection === item.key).map((item) => (
                <div className={styles.pageEditor} key={item.key}>
                    <div className="editor-field">
                        <div className="editor-label">
                            <strong>页面内容</strong>
                            <span>支持完整 HTML；留空即可恢复系统内置页面。</span>
                        </div>
                        <Form.Item name={[item.key, "body"]}>
                            <AceEditor
                                mode="html"
                                theme={dark ? "one_dark" : "github"}
                                width="100%"
                                height={compact ? 220 : 260}
                                fontSize={12}
                                tabSize={4}
                                showPrintMargin={false}
                                showLineNumbers
                                wrapEnabled
                                acePath={acePath}
                                placeholder="留空即可恢复系统内置页面"
                                className={styles.code}/>
                        </Form.Item>
                    </div>
                    <SettingField
                        styles={styles}
                        name={[item.key, "status"]}
                        title="响应状态码"
                        description={`${item.description}，留空使用默认状态码 ${item.status}。`}>
                        <Select
                            allowClear
                            showSearch
                            optionFilterProp="label"
                            options={httpStatusOptions}
                            placeholder={`请选择响应状态码，留空使用 ${item.status}`}/>
                    </SettingField>
                    <Collapse
                        ghost
                        className={styles.disclosure}
                        expandIconPlacement="end"
                        activeKey={pagePanels}
                        onChange={(keys) => setPagePanels((Array.isArray(keys) ? keys : [keys]).map(String))}
                        items={[{
                            key: "headers",
                            label: <DisclosureLabel title="响应头（高级）" description="设置此页面额外返回的 HTTP 响应头"/>,
                            children: (
                                <ConfigProvider componentSize="small" theme={disclosureTheme}>
                                    <HeaderEditor name={item.key} styles={styles}/>
                                </ConfigProvider>
                            ),
                        }]}/>
                </div>
            ))}
        </div>
    );

    const contents: any = {
        service: serviceContent,
        basic: basicContent,
        network: networkContent,
        runtime: runtimeContent,
        certificate: certificateContent,
        pages: pagesContent,
    };

    return (
        <>
            <ProModal
                ref={modalRef}
                title={<Title>服务设置</Title>}
                width={compact ? "620px" : "680px"}
                onOk={form.submit}
                onCancel={close}
                modalProps={{
                    rootClassName: styles.modal,
                    confirmLoading: saving || loading,
                    okText: "保存",
                    cancelText: "关闭",
                    okButtonProps: section === "service" ? {style: {display: "none"}} : undefined,
                    footer: loading || Boolean(loadError) ? null : undefined,
                    forceRender: true,
                    closable: true,
                    keyboard: true,
                    mask: {closable: true},
                    focusable: {focusTriggerAfterClose: false},
                    style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                    afterClose: () => {
                        requestRef.current += 1;
                        setActive(false);
                        setLoading(false);
                        setSaving(false);
                        setAction("");
                        setLoadError("");
                        setSiteOptions([]);
                        setSection("service");
                        setPageSection("initial_page");
                        setNetworkPanels([]);
                        setRuntimePanels([]);
                        setPagePanels([]);
                    },
                } as any}>
                <div style={!active ? {display: "none"} : undefined}>
                    <Spin spinning={loading} size="medium">
                        {loadError ? (
                            <div className={styles.errorState}>
                                <Empty
                                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                                    description={loadError}>
                                    <Button type="primary" onClick={() => void load(true, true)}>重新加载</Button>
                                </Empty>
                            </div>
                        ) : (
                            <Form
                                form={form}
                                className={`http-service-workbench ${styles.workbench}`}
                                data-section={section}
                                layout="vertical"
                                variant="filled"
                                onFinish={submit}
                                onFinishFailed={({errorFields}) => {
                                    const first = errorFields[0];
                                    if (!first) return;
                                    const names = first.name.map(String);
                                    if (durationFields.some(([name]) => name === names[0])) {
                                        setRuntimePanels((current) => Array.from(new Set([...current, "timeouts"])));
                                    }
                                    if (names.includes("headers")) {
                                        setPagePanels((current) => Array.from(new Set([...current, "headers"])));
                                    }
                                    requestAnimationFrame(() => requestAnimationFrame(() => {
                                        form.scrollToField(first.name, {block: "center"});
                                    }));
                                }}
                                style={loading ? {visibility: "hidden"} : undefined}>
                                <aside className={styles.navigation} aria-label="服务设置分类">
                                    <div className="navigation-scroll" ref={navigationRef}>
                                        <div className="navigation-grid">
                                            {navigationItems.map((item) => (
                                                <button
                                                    type="button"
                                                    className="navigation-item"
                                                    key={item.key}
                                                    data-section={item.key}
                                                    data-active={section === item.key}
                                                    aria-current={section === item.key ? "page" : undefined}
                                                    onClick={() => setSection(item.key)}>
                                                    <Icon type={item.icon}/>
                                                    <span className="navigation-label">{item.label}</span>
                                                </button>
                                            ))}
                                        </div>
                                    </div>
                                </aside>
                                <div className={styles.mainPane}>
                                    <main className={styles.content}>
                                        <div className={styles.contentInner}>{contents[section]}</div>
                                    </main>
                                </div>
                            </Form>
                        )}
                    </Spin>
                </div>
            </ProModal>
            <Log ref={logRef} request={getHTTPLog} title="服务日志"/>
        </>
    );
});

HTTPService.displayName = "HTTPService";

export default memo(HTTPService);
