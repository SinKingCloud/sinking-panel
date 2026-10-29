import {createContext, forwardRef, memo, useCallback, useContext, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {
    App,
    AutoComplete,
    Button,
    Checkbox,
    Collapse,
    Empty,
    Form,
    Grid,
    Input,
    InputNumber,
    Segmented,
    Select,
    Spin,
    Switch,
    Tooltip,
} from "antd";
import {createStyles} from "antd-style";
import dayjs from "dayjs";
import {Icon, ProModal, ProModalRef, Title, useTheme} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import AceEditor from "@/components/ace-editor";
import FilePicker from "@/pages/components/file-picker";
import Log from "@/pages/components/task-log";
import ProcessControls from "../process-controls";
import ActionDropdown from "@/pages/components/ui/action-dropdown";
import {
    comparableFilePath,
    isAbsoluteFilePath,
    isFilePathWithin,
    joinFilePath,
    normalizeFilePath,
} from "@/pages/file/utils";
import {getCertList} from "@/service/api/cert";
import {clearEnumCache} from "@/utils/enum";
import {parseProxyTarget} from "../../proxy";
import {
    getSiteInfo,
    getSiteLog,
    siteCache,
    siteCompression,
    siteDomain,
    siteFastCGI,
    siteHeaders,
    siteProcess,
    siteProxy,
    siteRateLimit,
    siteRedirects,
    siteRoutes,
    siteSSL,
    siteStatic,
    siteTrafficLimit,
    siteWAF,
    updateSite,
} from "@/service/api/site";

const acePath = `${defaultSettings?.basePath || "/"}ace`;
const NS = 1_000_000_000;
const HTTP_METHOD_OPTIONS = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"].map((value) => ({label: value, value}));
const RATE_KEY_OPTIONS = [
    {value: "{http.vars.client_ip}", label: "真实访客 IP", description: "已按可信代理规则解析，适合大多数网站"},
    {value: "{http.request.remote.host}", label: "连接来源 IP", description: "直接连接到服务器的来源地址"},
    {value: "{http.request.host}", label: "访问域名", description: "按请求使用的域名分别统计"},
    {value: "{http.request.uri.path}", label: "请求路径", description: "按每个访问路径分别统计"},
    {value: "{http.request.method}", label: "请求方法", description: "按 GET、POST 等请求方法分别统计"},
];
const WAF_TARGET_OPTION_GROUPS = [
    {label: "请求信息", options: [
        {label: "访问地址", value: "REQUEST_URI", description: "完整访问路径", placeholder: "例如 /admin", operator: "contains"},
        {label: "请求文件路径", value: "REQUEST_FILENAME", description: "服务器解析后的请求文件路径", placeholder: "例如 /www/site/index.php", operator: "contains"},
        {label: "请求文件名", value: "REQUEST_BASENAME", description: "请求路径末尾的文件名", placeholder: "例如 index.php", operator: "equals"},
        {label: "请求方法", value: "REQUEST_METHOD", description: "GET、POST 等请求方法", placeholder: "例如 POST", operator: "equals"},
        {label: "请求协议", value: "REQUEST_PROTOCOL", description: "当前请求使用的协议版本", placeholder: "例如 HTTP/2.0", operator: "equals"},
        {label: "访客 IP", value: "REMOTE_ADDR", description: "当前请求的来源 IP", placeholder: "例如 192.168.1.0/24", operator: "ip"},
    ]},
    {label: "参数与正文", options: [
        {label: "查询字符串", value: "QUERY_STRING", description: "问号后的原始参数", placeholder: "例如 debug=true", operator: "contains"},
        {label: "全部参数值", value: "ARGS", description: "GET 和 POST 参数内容", placeholder: "例如 union select", operator: "contains"},
        {label: "全部参数名", value: "ARGS_NAMES", description: "GET 和 POST 参数名称", placeholder: "例如 redirect_url", operator: "equals"},
        {label: "GET 参数值", value: "ARGS_GET", description: "URL 查询参数内容", placeholder: "例如 debug=true", operator: "contains"},
        {label: "GET 参数名", value: "ARGS_GET_NAMES", description: "URL 查询参数名称", placeholder: "例如 redirect_url", operator: "equals"},
        {label: "POST 参数值", value: "ARGS_POST", description: "表单提交参数内容", placeholder: "例如 union select", operator: "contains"},
        {label: "POST 参数名", value: "ARGS_POST_NAMES", description: "表单提交参数名称", placeholder: "例如 command", operator: "equals"},
        {label: "请求参数大小", value: "ARGS_COMBINED_SIZE", description: "全部参数名称与内容的总字节数", placeholder: "例如 1048576", operator: "greater_than"},
        {label: "请求正文", value: "REQUEST_BODY", description: "未经解析的请求正文", placeholder: "例如 malicious payload", operator: "contains"},
        {label: "上传临时文件", value: "FILES", description: "上传文件在服务器中的临时路径", placeholder: "例如 /tmp/upload", operator: "contains"},
        {label: "上传文件名", value: "FILES_NAMES", description: "上传文件的原始名称", placeholder: "例如 .php", operator: "ends_with"},
    ]},
    {label: "请求头与 Cookie", options: [
        {label: "全部请求头", value: "REQUEST_HEADERS", description: "全部请求头的内容", placeholder: "例如 suspicious-value", operator: "contains"},
        {label: "请求头名称", value: "REQUEST_HEADERS_NAMES", description: "全部请求头的名称", placeholder: "例如 X-Debug", operator: "equals"},
        {label: "请求域名", value: "REQUEST_HEADERS:Host", description: "Host 请求头", placeholder: "例如 example.com", operator: "equals"},
        {label: "内容类型", value: "REQUEST_HEADERS:Content-Type", description: "Content-Type 请求头", placeholder: "例如 application/json", operator: "contains"},
        {label: "浏览器标识", value: "REQUEST_HEADERS:User-Agent", description: "User-Agent 请求头", placeholder: "例如 sqlmap", operator: "contains"},
        {label: "来源页面", value: "REQUEST_HEADERS:Referer", description: "Referer 请求头", placeholder: "例如 example.com", operator: "contains"},
        {label: "跨域来源", value: "REQUEST_HEADERS:Origin", description: "Origin 请求头", placeholder: "例如 https://example.com", operator: "equals"},
        {label: "认证信息", value: "REQUEST_HEADERS:Authorization", description: "Authorization 请求头", placeholder: "例如 Bearer", operator: "starts_with"},
        {label: "转发来源 IP", value: "REQUEST_HEADERS:X-Forwarded-For", description: "X-Forwarded-For 请求头", placeholder: "例如 192.168.1.1", operator: "contains"},
        {label: "Cookie 内容", value: "REQUEST_COOKIES", description: "全部 Cookie 的值", placeholder: "例如 admin=true", operator: "contains"},
        {label: "Cookie 名称", value: "REQUEST_COOKIES_NAMES", description: "全部 Cookie 的名称", placeholder: "例如 session_id", operator: "equals"},
    ]},
];
const WAF_OPERATOR_OPTIONS = [
    {label: "包含", value: "contains"},
    {label: "完全等于", value: "equals"},
    {label: "开头是", value: "starts_with"},
    {label: "结尾是", value: "ends_with"},
    {label: "正则表达式", value: "regex"},
    {label: "IP 或网段", value: "ip"},
    {label: "存在", value: "exists"},
    {label: "内容不为空", value: "not_empty"},
    {label: "数值大于", value: "greater_than"},
    {label: "数值小于", value: "less_than"},
];
const WAF_NEGATED_OPERATOR_LABELS = new Map([
    ["contains", "不包含"],
    ["equals", "不等于"],
    ["starts_with", "不以此开头"],
    ["ends_with", "不以此结尾"],
    ["regex", "不匹配正则"],
    ["ip", "不属于此网段"],
    ["exists", "不存在"],
    ["not_empty", "内容为空"],
    ["greater_than", "小于或等于"],
    ["less_than", "大于或等于"],
]);
const WAF_TARGET_OPTIONS = WAF_TARGET_OPTION_GROUPS.flatMap((group) => group.options).map((target) => {
    const numeric = target.value === "ARGS_COMBINED_SIZE";
    const operators = WAF_OPERATOR_OPTIONS.filter((item) => {
        if (item.value === "ip") return target.value === "REMOTE_ADDR";
        if (["greater_than", "less_than"].includes(item.value)) return numeric;
        return !numeric || item.value === "equals";
    });
    return {
        ...target,
        operators,
        options: operators.flatMap((item) => [
            item,
            {label: WAF_NEGATED_OPERATOR_LABELS.get(item.value), value: `not:${item.value}`},
        ]),
    };
});
const WAF_VALUELESS_OPERATORS = new Set(["exists", "not_empty"]);
const WAF_CASE_OPERATORS = new Set(["contains", "equals", "starts_with", "ends_with", "regex"]);
const WAF_PROTECTION_PRESETS: Record<string, Record<string, number>> = {
    loose: {paranoia_level: 1, detection_paranoia_level: 1, inbound_anomaly_score_threshold: 10, outbound_anomaly_score_threshold: 8},
    recommended: {paranoia_level: 1, detection_paranoia_level: 1, inbound_anomaly_score_threshold: 5, outbound_anomaly_score_threshold: 4},
    strict: {paranoia_level: 2, detection_paranoia_level: 2, inbound_anomaly_score_threshold: 5, outbound_anomaly_score_threshold: 4},
};
const normalizeWAFCondition = (condition: any) => ({
    target: String(condition?.target || "REQUEST_URI"),
    operator: String(condition?.operator || "contains"),
    value: String(condition?.value || ""),
    negated: condition?.negated ? "not" : "match",
    ignore_case: condition?.ignore_case == null ? true : Boolean(condition.ignore_case),
});
const normalizeWAFGroup = (group: any) => ({
    match: group?.match === "any" ? "any" : "all",
    conditions: (Array.isArray(group?.conditions) ? group.conditions : []).map(normalizeWAFCondition),
});
const SettingsFormContext = createContext<any>(undefined);
const splitValues = (value: any) => Array.from(new Set(
    (Array.isArray(value) ? value : String(value || "").split(/[\n,]/))
        .map((item) => String(item || "").trim())
        .filter(Boolean),
));
const seconds = (value: any) => Math.max(0, Number(value || 0)) / NS;
const duration = (value: any) => Math.max(0, Number(value || 0)) * NS;
const upstreamsToForm = (items: any) => (Array.isArray(items) ? items : []).map((item) => item?.dial).filter(Boolean);
const upstreamsToBody = (items: any) => splitValues(items).map((dial) => ({dial, max_requests: 0}));

const headersToRows = (add: any, set: any, remove: any) => [
    ...Object.entries(add || {}).flatMap(([name, values]: any) => (Array.isArray(values) ? values : [values])
        .map((value) => ({operation: "add", name, value: String(value || "")}))),
    ...Object.entries(set || {}).flatMap(([name, values]: any) => (Array.isArray(values) ? values : [values])
        .map((value) => ({operation: "set", name, value: String(value || "")}))),
    ...(Array.isArray(remove) ? remove : []).map((name) => ({operation: "delete", name: String(name || ""), value: ""})),
];

const rowsToHeaders = (rows: any) => (Array.isArray(rows) ? rows : []).reduce((result: any, item: any) => {
    const operation = ["add", "set", "delete"].includes(item?.operation) ? item.operation : "set";
    const name = String(item?.name || "").trim();
    const value = String(item?.value || "").trim();
    if (!name) return result;
    if (operation === "delete") {
        if (!result.delete.includes(name)) result.delete.push(name);
    } else if (value) {
        (result[operation][name] ||= []).push(value);
    }
    return result;
}, {add: {}, set: {}, delete: []});

const envToLines = (value: any) => Object.entries(value || {}).map(([key, item]) => `${key}=${item}`).join("\n");
const linesToEnv = (value: any) => String(value || "").split("\n").reduce((result: any, line) => {
    const index = line.indexOf("=");
    if (index > 0) {
        result[line.slice(0, index).trim()] = line.slice(index + 1).trim();
    }
    return result;
}, {});

const useStyles = createStyles(({css, token}: any, props: any = {}) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-container {
            height: ${props?.compact ? 680 : 700}px;
            min-height: ${props?.compact ? 680 : 700}px;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .ant-modal-body {
            min-height: 0;
            padding: 0 !important;
            display: flex;
            flex: 1 1 auto;
            overflow: hidden;
        }

        .ant-modal-body > div {
            width: 100%;
            min-height: 0;
            flex: 1 1 auto;
            overflow: hidden;
        }

        .ant-modal-header, .ant-modal-footer {
            flex: none;
        }

        .ant-modal-footer {
            margin-top: 0;
            padding-top: ${props?.compact ? 10 : 12}px;
        }

    `,
    layout: css`
        box-sizing: border-box;
        width: 100%;
        height: auto;
        min-width: 0;
        min-height: 0;
        display: grid;
        flex: 1 1 auto;
        grid-template-columns: ${props?.compact ? 166 : 178}px minmax(0, 1fr);
        grid-template-rows: minmax(0, 1fr);
        overflow: hidden;

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
        margin: ${props?.compact ? 6 : 8}px 0;
        padding: ${props?.compact ? 3 : 4}px 8px;
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
            gap: 2px;
        }

        .navigation-item {
            width: 100%;
            min-width: 0;
            min-height: 40px;
            padding: 0 11px;
            display: flex;
            align-items: center;
            gap: 8px;
            flex: none;
            overflow: hidden;
            border: 0;
            border-radius: ${token.borderRadius}px;
            outline: 0;
            background: transparent;
            color: ${token.colorTextSecondary};
            cursor: pointer;
            font: inherit;
            font-size: 13px;
            font-weight: 400;
            text-align: left;
            transition: background-color .16s ease, color .16s ease;
        }

        .navigation-item:hover { background: ${token.colorFillTertiary}; }

        .navigation-item:focus-visible {
            box-shadow: 0 0 0 2px ${token.colorPrimaryBg};
        }

        .navigation-item[data-active="true"] {
            background: color-mix(in srgb, ${token.colorPrimaryBg} 86%, ${token.colorBgContainer});
            color: ${token.colorTextHeading};
            font-weight: 600;
        }

        .navigation-dot {
            width: 7px;
            height: 7px;
            flex: none;
            border-radius: 50%;
            background: color-mix(in srgb, var(--accent), ${token.colorTextTertiary} 8%);
            opacity: .62;
        }

        .navigation-item[data-active="true"] .navigation-dot {
            background: var(--accent);
            box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent), transparent 78%);
            opacity: 1;
        }

        .navigation-label {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @media (max-width: 760px) {
            min-width: 0;
            max-height: none;
            margin: 0;
            padding: 5px 0;
            align-self: stretch;
            overflow: hidden;
            border-right: 0;
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
                min-height: 34px;
                padding: 0 9px;
                gap: 6px;
                background: ${token.colorFillQuaternary};
            }
        }
    `,
    content: css`
        box-sizing: border-box;
        min-width: 0;
        min-height: 0;
        height: auto;
        padding: ${props?.compact ? "15px 18px 18px" : "18px 22px 22px"};
        overflow: auto;
        overscroll-behavior: contain;
        scrollbar-color: ${token.colorBorderSecondary} transparent;
        scrollbar-width: thin;
        -webkit-overflow-scrolling: touch;
        touch-action: pan-y;

        .ant-collapse-ghost {
            margin-top: 2px;
            border: 0 !important;
            background: transparent !important;
        }

        .ant-collapse-ghost > .ant-collapse-item {
            border: 0 !important;
        }

        .ant-collapse-ghost > .ant-collapse-item > .ant-collapse-header {
            min-height: 0;
            padding: ${props?.compact ? "4px 0" : "5px 0"} !important;
            align-items: center !important;
            border-radius: ${token.borderRadiusSM}px !important;
            outline: 0 !important;
            background: transparent !important;
            color: ${token.colorTextSecondary} !important;
            font-size: 13px;
            font-weight: 500;
            box-shadow: none !important;
        }

        .ant-collapse-ghost > .ant-collapse-item-active > .ant-collapse-header {
            color: ${token.colorText} !important;
        }

        .ant-collapse-ghost .ant-collapse-expand-icon {
            color: ${token.colorTextTertiary};
        }

        .ant-collapse-ghost .ant-collapse-item-active .ant-collapse-expand-icon {
            color: ${token.colorPrimary};
        }

        .ant-collapse-ghost .ant-collapse-body {
            margin: 1px 0 ${props?.compact ? 3 : 4}px;
            padding: ${props?.compact ? "5px 8px" : "7px 10px"} !important;
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorFillQuaternary};
        }

        @media (max-width: 760px) {
            min-height: 0;
            height: auto;
            padding: ${props?.compact ? "12px 12px 16px" : "15px 12px 18px"};
        }
    `,
    contentInner: css`
        width: 100%;
        max-width: 800px;
        margin: 0 auto;
    `,
    sectionHeader: css`
        margin-bottom: ${props?.compact ? 10 : 13}px;
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 12px;

        h3 { margin: 0; color: ${token.colorTextHeading}; font-size: ${props?.compact ? 14 : 15}px; font-weight: 600; line-height: 22px; }
        p { margin: 2px 0 0; color: ${token.colorTextTertiary}; font-size: 12px; line-height: 18px; }

        @media (max-width: 575px) {
            margin-bottom: ${props?.compact ? 10 : 12}px;
        }
    `,
    settingAction: css`
        &.ant-btn {
            width: auto;
            min-width: 0;
            height: ${props?.compact ? 24 : 26}px;
            padding: 0 ${props?.compact ? 7 : 8}px;
            gap: 4px;
            font-size: 12px;
            line-height: 1;
            white-space: nowrap;
        }

        &.ant-btn .anticon {
            font-size: 12px;
        }
    `,
    ruleAction: css`
        &.ant-btn {
            min-width: 0;
            height: 24px;
            padding: 0 6px;
            font-size: 12px;
            line-height: 1;
        }

        &.ant-btn.ant-btn-icon-only {
            width: 24px;
            padding: 0;
        }

        &.ant-btn .anticon {
            font-size: 12px;
        }
    `,
    settingsForm: css`
        .ant-form-item {
            margin: 0;
            padding: ${props?.compact ? "4px 0" : "6px 0"};
        }

        .ant-form-item-label {
            padding: 0 0 ${props?.compact ? 5 : 6}px;
            text-align: left;
            white-space: normal;
        }

        .ant-form-item-label > label {
            height: auto;
            color: ${token.colorText};
            font-size: 13px;
            font-weight: 500;
            line-height: 20px;
            white-space: normal;
        }

        .ant-form-item-control {
            min-width: 0;
            width: 100%;
            max-width: none;
        }

        .ant-form-item-control-input {
            width: 100%;
            max-width: none;
        }

        .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            width: 100%;
            max-width: ${props?.compact ? 280 : 300}px;
        }

        .ant-form-item-control-input-content > :where(
            textarea.ant-input,
            .ant-select-multiple
        ),
        .setting-wide .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            max-width: ${props?.compact ? 380 : 420}px;
        }

        .setting-full .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            max-width: none;
        }

        :where(input, textarea)::placeholder,
        .ant-select-placeholder {
            font-size: ${props?.compact ? 12 : 13}px !important;
        }

        .ant-form-item-extra {
            margin-top: ${props?.compact ? 4 : 5}px;
            color: ${token.colorTextTertiary};
            font-size: 12px;
            line-height: 18px;
        }

        .ant-form-item-explain {
            margin-top: 4px;
            font-size: 11px;
        }

        .ant-collapse-body .ant-form-item {
            padding-top: ${props?.compact ? 2 : 3}px;
            padding-bottom: ${props?.compact ? 2 : 3}px;
        }

        .ant-collapse-body .ant-form-item-label {
            padding-bottom: 3px;
        }

        .ant-collapse-body .ant-form-item-label > label {
            font-size: 12px;
            line-height: 18px;
        }

        .ant-collapse-body .ant-form-item-extra {
            margin-top: 3px;
            font-size: 11px;
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
            .ant-select-placeholder
        ),
        .ant-collapse-body :where(input, textarea)::placeholder {
            font-size: 12px !important;
        }

    `,
    fieldGrid: css`
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr));
        align-items: start;
        gap: ${props?.compact ? 4 : 6}px 16px;
        min-width: 0;

        &.setting-grid-numeric {
            grid-template-columns: repeat(auto-fit, minmax(min(100%, 148px), 1fr));
        }

        > .ant-form-item {
            min-width: 0;
        }

        .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select
        ) {
            max-width: none;
        }
    `,
    formToolbar: css`
        min-height: 32px;
        margin: -4px 0 ${props?.compact ? 10 : 14}px;
        display: flex;
        justify-content: flex-end;
    `,
    toggleRow: css`
        min-height: 0;
        margin: 0;
        padding: ${props?.compact ? "7px 0" : "9px 0"};
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 20px;

        .toggle-title { color: ${token.colorText}; font-size: 13px; font-weight: 500; line-height: 20px; }
        .toggle-description { margin-top: 2px; color: ${token.colorTextTertiary}; font-size: 12px; line-height: 18px; }
        .ant-form-item { margin: 0; flex: none; }
    `,
    fieldHeading: css`
        margin: ${props?.compact ? "4px 0 10px" : "6px 0 12px"};
        color: ${token.colorText};
        font-size: 13px;
        font-weight: 500;
    `,
    httpsTable: css`
        min-width: 0;
        max-width: 100%;
        margin-bottom: 12px;

        .https-table-scroll {
            width: 100%;
            min-width: 0;
            max-width: 100%;
            overflow-x: auto;
            overscroll-behavior-inline: contain;
            scrollbar-width: thin;
            scrollbar-color: ${token.colorBorderSecondary} transparent;
            -webkit-overflow-scrolling: touch;
        }

        .https-table-scroll .domain-list-heading, .https-table-scroll .domain-list { min-width: 440px; }
        .https-table-scroll .domain-list-heading, .https-table-scroll .domain-row {
            grid-template-columns: minmax(140px, 1fr) minmax(150px, 1fr) 96px;
        }

        .https-table-scroll .domain-row { align-items: center; }
        .https-domain { min-width: 0; min-height: ${token.controlHeight}px; display: flex; align-items: center; gap: 8px; }
        .https-domain .ant-checkbox-wrapper { flex: none; }
        .domain-list-heading .https-domain { min-height: 0; }
        .https-certificate { min-width: 0; }
        .https-certificate-status { color: ${token.colorTextSecondary}; font-size: 12px; line-height: 20px; font-variant-numeric: tabular-nums; }
        .https-certificate-status[data-expired="true"] { color: ${token.colorError}; }
        .https-certificate-status[data-expiring="true"] { color: ${token.colorWarning}; }
        .https-operations { min-width: 0; text-align: right; }
        .https-table-scroll .domain-actions { min-height: ${token.controlHeight}px; align-items: center; }
        .https-certificate-empty { color: ${token.colorTextQuaternary}; }
    `,
    domainEditor: css`
        width: 100%;
        min-width: 0;

        .domain-composer {
            min-width: 0;
            margin-bottom: ${props?.compact ? 18 : 20}px;
        }

        .domain-composer-heading { margin-bottom: 8px; }
        .domain-composer-heading h4 { margin: 0; color: ${token.colorText}; font-size: 13px; font-weight: 500; line-height: 20px; }

        .domain-composer .ant-form-item { min-width: 0; margin: 0; padding: 0; }
        .domain-composer .ant-form-item-control-input { min-height: 0; }
        .domain-composer .ant-form-item-control-input-content > :where(.ant-input, .ant-input-affix-wrapper) {
            width: 100%;
            max-width: none;
            min-height: ${props?.compact ? 28 : 32}px;
            font-size: 13px;
        }

        .domain-composer .ant-form-item-control-input-content > textarea.ant-input {
            box-sizing: border-box;
            width: 100%;
            max-width: none;
            height: ${props?.compact ? 78 : 80}px;
            min-height: ${props?.compact ? 78 : 80}px;
            padding: 8px 10px;
            line-height: 20px;
            resize: vertical;
        }

        .domain-composer-footer { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px 12px; margin-top: 8px; }
        .domain-batch-status { flex: 1 1 160px; min-width: 0; color: ${token.colorTextTertiary}; font-size: 12px; line-height: 18px; overflow-wrap: anywhere; }
        .domain-batch-status strong { color: ${token.colorTextSecondary}; font-weight: 500; font-variant-numeric: tabular-nums; }
        .domain-composer-actions { display: flex; align-items: center; justify-content: flex-end; flex-wrap: nowrap; gap: 6px; margin-left: auto; }
        .domain-composer-actions > .ant-btn { flex: none; }

        .domain-toolbar { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 4px 12px; margin-bottom: 8px; }
        .domain-list-title { display: inline-flex; align-items: baseline; flex-wrap: wrap; gap: 8px; color: ${token.colorText}; font-size: 13px; font-weight: 500; line-height: 20px; }

        .domain-list-container { width: 100%; min-width: 0; box-sizing: border-box; overflow: hidden; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusSM}px; }
        .domain-list { width: 100%; min-width: 0; border-top: 1px solid ${token.colorBorderSecondary}; }
        .domain-list-heading, .domain-row {
            box-sizing: border-box;
            min-width: 0;
            display: grid;
            grid-template-columns: minmax(0, 1fr) 48px;
            align-items: center;
            gap: ${props?.compact ? 10 : 12}px;
            padding: ${props?.compact ? "7px 10px" : "8px 12px"};
        }

        .domain-list-heading { background: ${token.colorFillQuaternary}; color: ${token.colorTextTertiary}; font-size: 12px; line-height: 18px; }
        .domain-list-heading > :last-child { text-align: right; }
        .domain-row { min-height: ${props?.compact ? 44 : 48}px; }
        .domain-row + .domain-row { border-top: 1px solid ${token.colorBorderSecondary}; }
        .domain-main { min-width: 0; }
        .domain-value { display: block; min-width: 0; color: ${token.colorText}; font-size: 13px; line-height: 20px; overflow-wrap: anywhere; }

        .domain-actions { display: flex; align-items: center; justify-content: flex-end; flex-wrap: nowrap; gap: 4px; min-width: 0; }
        .domain-actions > .ant-btn { flex: none; }

        .domain-empty { padding: 20px 12px; text-align: center; }
        .domain-empty > .anticon { margin-bottom: 7px; color: ${token.colorTextQuaternary}; font-size: 20px; }
        .domain-empty-title { color: ${token.colorTextSecondary}; font-size: 13px; }
        .domain-empty p { margin: 4px 0 0; color: ${token.colorTextTertiary}; font-size: 12px; line-height: 18px; }

        .domain-errors { color: ${token.colorError}; font-size: 12px; line-height: 18px; }
        .domain-errors:not(:empty) { margin: 8px 0; }

        @media (max-width: 575px) {
            .domain-list-heading, .domain-row { padding-inline: 8px; column-gap: 8px; }
        }
    `,
    redirectRules: css`
        display: grid;
        gap: ${props?.compact ? 9 : 11}px;

        .redirect-toolbar {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 10px;
            color: ${token.colorTextTertiary};
            font-size: 12px;
        }

        .redirect-card {
            min-width: 0;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadius}px;
        }

        .redirect-heading {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: ${props?.compact ? "8px 9px" : "10px 12px"};
            border-radius: ${token.borderRadius}px;
            background: ${token.colorFillQuaternary};
        }

        .redirect-summary {
            min-width: 0;
            flex: 1;
            padding: 0;
            display: flex;
            align-items: center;
            gap: 8px;
            border: 0;
            background: transparent;
            color: ${token.colorText};
            font: inherit;
            text-align: left;
            cursor: pointer;
        }

        .redirect-summary:focus-visible { outline: 2px solid ${token.colorPrimary}; outline-offset: 3px; }
        .redirect-summary > .anticon { color: ${token.colorTextTertiary}; font-size: 11px; }
        .redirect-summary-text { min-width: 0; flex: 1; }
        .redirect-title { display: flex; align-items: center; gap: 7px; font-size: 13px; font-weight: 500; }
        .redirect-title > span:first-child { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .redirect-status { flex: none; color: ${token.colorTextTertiary}; font-size: 11px; font-weight: 400; }
        .redirect-destination { display: block; margin-top: 2px; overflow: hidden; color: ${token.colorTextTertiary}; font-size: 11px; line-height: 17px; text-overflow: ellipsis; white-space: nowrap; }
        .redirect-controls { flex: none; display: flex; align-items: center; gap: 5px; }
        .redirect-body { padding: ${props?.compact ? "6px 12px 10px" : "8px 14px 12px"}; }
        .redirect-body[hidden] { display: none; }
        .redirect-scope { max-width: 300px; }
        .redirect-scope .ant-segmented-item-label { min-height: 28px; padding: 0 10px; font-size: 12px; line-height: 28px; }
        .redirect-example { margin: 8px 0; padding: 9px 11px; border-radius: ${token.borderRadiusSM}px; background: ${token.colorFillQuaternary}; }
        .redirect-example-title { margin-bottom: 5px; color: ${token.colorTextSecondary}; font-size: 12px; font-weight: 500; }
        .redirect-example-row { display: grid; grid-template-columns: 38px minmax(0, 1fr); align-items: baseline; gap: 6px; color: ${token.colorTextTertiary}; font-size: 11px; line-height: 19px; }
        .redirect-example-row code { color: ${token.colorTextSecondary}; font: inherit; overflow-wrap: anywhere; }
        .redirect-example-note { margin-top: 5px; color: ${token.colorTextTertiary}; font-size: 11px; line-height: 17px; }
        .redirect-empty { padding: 22px 12px; text-align: center; }
        .redirect-empty .ant-empty { margin: 0 0 12px; }

        @media (max-width: 575px) {
            .redirect-heading { padding: 9px; gap: 5px; }
            .redirect-body { padding: 6px 10px 10px; }
            .redirect-toolbar { flex-wrap: wrap; }
            .redirect-title { font-size: 12px; }
        }
    `,
    routeRules: css`
        display: grid;
        gap: ${props?.compact ? 9 : 11}px;

        .route-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 10px; color: ${token.colorTextTertiary}; font-size: 12px; }
        .route-card { min-width: 0; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; scroll-margin-block: 8px; }
        .route-heading { display: flex; align-items: center; gap: 8px; padding: ${props?.compact ? "8px 9px" : "10px 12px"}; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; }
        .route-summary { min-width: 0; flex: 1; display: flex; align-items: center; gap: 8px; padding: 0; border: 0; background: transparent; color: ${token.colorText}; font: inherit; text-align: left; cursor: pointer; }
        .route-summary:focus-visible { outline: 2px solid ${token.colorPrimary}; outline-offset: 3px; }
        .route-summary > .anticon { flex: none; color: ${token.colorTextTertiary}; font-size: 11px; }
        .route-summary-text { min-width: 0; flex: 1; }
        .route-title { display: flex; align-items: center; gap: 7px; font-size: 13px; font-weight: 500; }
        .route-title > span:first-child { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .route-kind { flex: none; color: ${token.colorTextTertiary}; font-size: 11px; font-weight: 400; }
        .route-destination { display: block; margin-top: 2px; overflow: hidden; color: ${token.colorTextTertiary}; font-size: 11px; line-height: 17px; text-overflow: ellipsis; white-space: nowrap; }
        .route-body { padding: ${props?.compact ? "6px 12px 10px" : "8px 14px 12px"}; }
        .route-body[hidden] { display: none; }
        .route-empty { padding: 20px 12px; text-align: center; }
        .route-empty .ant-empty { margin: 0 0 12px; }
        .route-hint { margin: 4px 0 8px; color: ${token.colorTextTertiary}; font-size: 12px; line-height: 20px; }

        @media (max-width: 575px) {
            .route-heading { padding: 9px; gap: 5px; }
            .route-body { padding: 6px 10px 10px; }
            .route-toolbar { flex-wrap: wrap; }
            .route-title { font-size: 12px; }
        }
    `,
    code: css`
        min-width: 0;
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;

        .ace_editor {
            border-radius: ${token.borderRadius}px !important;
        }
    `,
    wafMode: css`
        max-width: min(100%, 300px);

        .ant-segmented-item-label {
            min-height: 28px;
            padding: 0 10px;
            font-size: 12px;
            line-height: 28px;
        }
    `,
    variableGuide: css`
        width: 100%;
        max-width: 560px;
        margin: 4px 0 2px;
        padding: ${props?.compact ? "8px 9px" : "9px 10px"};
        display: grid;
        gap: 5px;
        border-radius: ${token.borderRadiusSM}px;
        background: ${token.colorFillQuaternary};

        .variable-guide-title {
            margin-bottom: 1px;
            color: ${token.colorTextSecondary};
            font-size: 12px;
            font-weight: 500;
        }

        .variable-guide-row {
            min-width: 0;
            display: grid;
            grid-template-columns: 146px minmax(0, 1fr);
            gap: 10px;
            color: ${token.colorTextTertiary};
            font-size: 11px;
            line-height: 17px;
        }

        code {
            overflow: hidden;
            color: ${token.colorTextSecondary};
            font: inherit;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @media (max-width: 575px) {
            max-width: none;
            .variable-guide-row { grid-template-columns: minmax(0, 1fr); gap: 0; }
        }
    `,
    wafRules: css`
        margin-top: 2px;
        container-type: inline-size;

        .waf-rules-header {
            min-height: 32px;
            margin-bottom: 8px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
        }
        .waf-rules-title {
            color: ${token.colorTextHeading};
            font-size: 12px;
            font-weight: 500;
        }

        .waf-rule-list {
            display: grid;
            gap: ${props?.compact ? 7 : 9}px;
        }

        .waf-rule-entry {
            min-width: 0;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadius}px;
            scroll-margin-block: 8px;
        }

        .waf-rule-summary {
            box-sizing: border-box;
            min-height: 40px;
            padding: 4px 6px 4px 9px;
            display: grid;
            grid-template-columns: minmax(0, 1fr) auto;
            align-items: center;
            gap: 10px;
            background: ${token.colorFillQuaternary};
            border-radius: ${token.borderRadius}px;
        }

        .waf-rule-select {
            min-width: 0;
            height: 32px;
            padding: 0;
            display: flex;
            align-items: center;
            gap: 7px;
            overflow: hidden;
            border: 0;
            outline: 0;
            background: transparent;
            color: inherit;
            font: inherit;
            text-align: left;
            cursor: pointer;
        }

        .waf-rule-select:focus-visible {
            border-radius: ${token.borderRadiusSM}px;
            box-shadow: 0 0 0 2px ${token.colorPrimaryBorder};
        }

        .waf-rule-select > .anticon {
            flex: none;
            color: ${token.colorTextTertiary};
            font-size: 11px;
        }

        .waf-rule-summary-text { min-width: 0; flex: 1; }
        .waf-rule-title { display: flex; align-items: center; gap: 7px; line-height: 18px; }

        .waf-rule-name {
            min-width: 0;
            overflow: hidden;
            color: ${token.colorText};
            font-size: 12px;
            font-weight: 500;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .waf-rule-meta {
            display: block;
            overflow: hidden;
            color: ${token.colorTextTertiary};
            font-size: 11px;
            line-height: 14px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .waf-rule-actions {
            height: 32px;
            display: flex;
            align-items: center;
            gap: 3px;
            white-space: nowrap;
        }

        .waf-rule-action {
            flex: none;
            color: ${token.colorTextTertiary};
            font-size: 11px;
            line-height: 18px;
            white-space: nowrap;
        }

        .waf-rule-switch {
            flex: none;
            margin: 0 2px;
        }

        .waf-rule-editor {
            padding: ${props?.compact ? "9px 7px 10px" : "11px 9px 12px"};
        }

        .waf-rule-body[hidden] { display: none; }

        .waf-rule-name-row {
            max-width: 430px;
            display: grid;
            grid-template-columns: 64px minmax(0, 320px);
            align-items: start;
            gap: 9px;
            color: ${token.colorTextSecondary};
            font-size: 12px;
            line-height: 22px;
        }

        .waf-rule-name-row > span {
            height: 24px;
            padding-top: 0;
            display: flex;
            align-items: center;
        }
        .waf-rule-name-row .ant-form-item {
            height: 24px;
            min-height: 24px;
            margin-bottom: 0;
            padding: 0;
            position: relative;
        }

        .waf-rule-name-row .ant-form-item-control-input { min-height: 24px; }
        .waf-rule-name-row .ant-input {
            font-size: 12px;
            line-height: 22px;
        }
        .waf-rule-name-row .ant-input::placeholder { font-size: 12px !important; }

        .waf-rule-sentence,
        .waf-situation-sentence {
            min-height: 24px;
            display: flex;
            align-items: center;
            flex-wrap: wrap;
            gap: 6px;
            color: ${token.colorTextSecondary};
            font-size: 12px;
            line-height: 22px;
        }

        .waf-rule-sentence .ant-select,
        .waf-situation-sentence .ant-select {
            font-size: 12px;
            line-height: 22px;
        }

        .waf-rule-sentence .ant-select-content,
        .waf-situation-sentence .ant-select-content {
            font-size: 12px !important;
        }

        .waf-rule-sentence .ant-select { min-width: 104px; }
        .waf-rule-name-row,
        .waf-rule-action-row { margin-bottom: ${props?.compact ? 14 : 16}px; }
        .waf-rule-match { margin: 0 0 ${props?.compact ? 6 : 8}px; }
        .waf-situation-list { margin: 6px 0 10px; }
        .waf-rule-action-row:has(+ .waf-rule-match) { margin-bottom: ${props?.compact ? 18 : 20}px; }
        .waf-rule-action-row:has(+ .waf-situation-list) { margin-bottom: ${props?.compact ? 6 : 8}px; }
        .waf-rule-action-row + .waf-situation-list { margin-top: 0; }

        @container (min-width: 521px) {
            .waf-rule-action-row {
                display: grid;
                grid-template-columns: 64px 112px;
                gap: 9px;
            }

            .waf-rule-action-row .ant-select {
                width: 112px;
                min-width: 0;
            }
        }

        .waf-situation {
            min-width: 0;
            padding: 3px 0;
        }

        .waf-situation[data-multiple="true"] {
            padding-left: 10px;
            border-left: 1px solid ${token.colorBorderSecondary};
        }

        .waf-situation-heading {
            min-height: 28px;
            margin-bottom: ${props?.compact ? 6 : 8}px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
        }

        .waf-situation-sentence strong {
            color: ${token.colorText};
            font-size: 12px;
            font-weight: 500;
        }

        .waf-situation-sentence .ant-select { min-width: 100px; }

        .waf-group-join {
            height: 20px;
            padding-left: 12px;
            display: flex;
            align-items: center;
            color: ${token.colorTextSecondary};
            font-size: 11px;
        }

        .waf-condition-list { margin: 1px 0 3px; }

        .waf-condition {
            min-width: 0;
            padding: 3px 0 11px;
        }

        .waf-condition-row {
            min-width: 0;
            display: grid;
            grid-template-columns: 32px minmax(0, .85fr) minmax(0, .85fr) minmax(0, 1.5fr) 54px;
            align-items: start;
            gap: 6px;
        }

        .waf-condition-row .ant-form-item {
            min-width: 0;
            margin-bottom: 0;
            padding: 0;
        }

        .waf-condition-value {
            height: 24px;
            min-height: 24px;
            position: relative;
        }

        .waf-condition-row .ant-form-item-control-input {
            min-height: 24px;
        }

        .waf-condition-row :where(.ant-select, .ant-input) {
            width: 100%;
            max-width: none;
            font-size: 12px;
            line-height: 22px;
        }

        .waf-condition-row .ant-select-content,
        .waf-condition-row .ant-input::placeholder {
            font-size: 12px !important;
        }

        .waf-condition-join {
            height: 24px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: ${token.colorTextSecondary};
            font-size: 12px;
            line-height: 22px;
            white-space: nowrap;
        }

        .waf-condition-actions {
            height: 24px;
            display: flex;
            align-items: center;
            justify-content: flex-end;
            gap: 1px;
        }

        .waf-case-button {
            width: 25px;
            min-width: 25px;
            padding: 0;
            font-size: 12px;
        }

        .waf-case-button.is-active {
            font-weight: 600;
        }

        .waf-rules-empty {
            padding: ${props?.compact ? "16px 0" : "19px 0"};
            color: ${token.colorTextTertiary};
            font-size: 12px;
            text-align: center;
        }

        .waf-rules-empty .ant-empty { margin: 0 0 10px; }
        .waf-rules-empty .ant-empty-description { font-size: 12px; }

        .waf-rule-name-row .ant-form-item-additional,
        .waf-condition-value .ant-form-item-additional {
            height: 14px;
            position: absolute;
            top: 24px;
            right: 0;
            left: 0;
        }

        .waf-rule-name-row .ant-form-item-explain,
        .waf-condition-value .ant-form-item-explain {
            height: 14px;
            margin-top: 0;
            overflow: hidden;
        }

        .waf-rule-name-row .ant-form-item-explain-error,
        .waf-condition-value .ant-form-item-explain-error {
            height: 14px;
            margin-top: 0;
            overflow: hidden;
            font-size: 11px;
            line-height: 14px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @container (max-width: 492px) {
            .waf-rule-summary { grid-template-columns: minmax(0, 1fr) auto; gap: 5px; }
            .waf-rule-actions { gap: 1px; }
            .waf-rule-name-row { grid-template-columns: minmax(0, 1fr); gap: 2px; }
            .waf-rule-name-row > span { padding-top: 0; }
            .waf-rule-name-row .ant-form-item { width: 100%; }
            .waf-situation[data-multiple="true"] { padding-left: 8px; }
            .waf-condition-row {
                grid-template-columns: 30px minmax(0, 1fr) minmax(0, 1fr) 68px;
            }
            .waf-condition-value {
                grid-row: 2;
                grid-column: 2 / 5;
            }
            .waf-condition-actions {
                grid-row: 1;
                grid-column: 4;
            }
            .waf-condition-actions .ant-btn {
                width: 32px;
                min-width: 32px;
            }
        }

        @container (max-width: 380px) {
            .waf-rule-editor { padding-inline: 6px; }
            .waf-situation[data-multiple="true"] { padding-left: 6px; }
            .waf-condition-row {
                grid-template-columns: 28px minmax(0, 1fr) 68px;
            }
            .waf-condition-target {
                grid-row: 1;
                grid-column: 2;
            }
            .waf-condition-operator {
                grid-row: 2;
                grid-column: 2 / 4;
            }
            .waf-condition-value {
                grid-row: 3;
                grid-column: 2 / 4;
            }
            .waf-condition-actions {
                grid-row: 1;
                grid-column: 3;
            }
        }
    `,
    rule: css`
        margin-bottom: 10px;
        padding: ${props?.compact ? 11 : 13}px;
        border: 1px solid transparent;
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};

        .waf-rule-heading {
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            flex-wrap: wrap;
            gap: 10px;
        }

        .waf-rule-heading > .waf-rule-index,
        .waf-rule-heading > .ant-space {
            flex: none;
            white-space: nowrap;
        }

        .waf-rule-heading > .ant-space { margin-left: auto; }
        .waf-rule-index { color: ${token.colorTextSecondary}; font-size: 12px; }
        .ant-form-item:last-child { margin-bottom: 0; }
    `,
    headerPane: css`
        --header-control-height: ${props?.compact ? 26 : 28}px;
        min-width: 0;

        .header-switcher {
            width: 100%;
            max-width: 320px;
            margin-bottom: ${props?.compact ? 9 : 11}px;
        }

        .header-summary {
            margin: 0 0 ${props?.compact ? 10 : 12}px;
            color: ${token.colorTextTertiary};
            font-size: 12px;
            line-height: 18px;
        }

        .header-rule-list {
            display: flex;
            flex-direction: column;
            gap: ${props?.compact ? 8 : 10}px;
        }

        .header-columns,
        .header-rule {
            display: grid;
            grid-template-columns: 94px minmax(100px, .85fr) minmax(140px, 1.15fr) var(--header-control-height);
            gap: 8px;
        }

        .header-columns {
            margin-bottom: 6px;
            padding: 0 0 0 1px;
            color: ${token.colorTextSecondary};
            font-size: 12px;
            line-height: 18px;
        }

        .header-rule {
            min-width: 0;
            align-items: start;
        }

        .header-rule .ant-form-item { min-width: 0; margin: 0; padding: 0; }
        .header-rule .ant-form-item-control-input { min-height: var(--header-control-height); }
        .header-rule :where(.ant-input, .ant-select) {
            width: 100%;
            height: var(--header-control-height);
            max-width: none !important;
            font-size: 12px;
        }
        .header-rule .ant-input { padding-top: 0; padding-bottom: 0; }
        .header-rule input::placeholder,
        .header-rule .ant-select-placeholder { font-size: 12px !important; }
        .header-rule .ant-form-item-explain { margin-top: 3px; font-size: 11px; line-height: 16px; }

        .header-rule-remove.ant-btn {
            width: var(--header-control-height);
            height: var(--header-control-height);
            min-width: 0;
            padding: 0;
            align-self: start;
            font-size: 12px;
        }
        .header-rule-remove.ant-btn .anticon { font-size: 12px; }

        .header-rule-empty {
            padding: ${props?.compact ? "16px 12px" : "20px 14px"};
            border-radius: ${token.borderRadius}px;
            background: ${token.colorFillQuaternary};
            color: ${token.colorTextTertiary};
            font-size: 12px;
            text-align: center;
        }

        .header-rule-toolbar {
            margin-top: 8px;
        }

        @media (max-width: 620px) {
            .header-switcher { max-width: none; }
            .header-columns { display: none; }
            .header-rule {
                grid-template-columns: 94px minmax(0, 1fr) var(--header-control-height);
                gap: 6px 8px;
            }
            .header-rule-list { gap: 14px; }
            .header-rule .header-value {
                grid-column: 1 / 3;
                grid-row: 2;
            }
            .header-rule-remove {
                grid-column: 3;
                grid-row: 1;
            }
        }
    `,
    empty: css`height: 360px; display: flex; align-items: center; justify-content: center;`,
    logGrid: css`
        display: flex;
        flex-direction: column;
        gap: ${props?.compact ? 8 : 10}px;
    `,
    logItem: css`
        box-sizing: border-box;
        min-height: ${props?.compact ? 56 : 62}px;
        padding: ${props?.compact ? "10px 12px" : "12px 14px"};
        display: flex;
        align-items: center;
        gap: 12px;
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};
        transition: background-color .16s ease;

        .log-icon {
            display: flex;
            width: ${props?.compact ? 28 : 30}px;
            height: ${props?.compact ? 28 : 30}px;
            flex: none;
            align-items: center;
            justify-content: center;
            border-radius: ${token.borderRadiusSM}px;
            background: ${token.colorFillTertiary};
            color: ${token.colorTextSecondary};
            font-size: 14px;
        }

        .log-main { min-width: 0; flex: 1; }
        .log-name { color: ${token.colorText}; font-size: 12px; font-weight: 500; line-height: 18px; }
        .log-description {
            margin-top: 3px;
            color: ${token.colorTextTertiary};
            font-size: 12px;
            line-height: 18px;
        }

        .log-open.ant-btn { flex: none; }

        @media (max-width: 575px) {
            gap: 9px;
            padding: 10px;
            .log-icon { width: 26px; height: 26px; font-size: 13px; }
        }
    `,
}));

const SectionHeader = ({title, description, extra, styles}: any) => (
    <div className={styles.sectionHeader}>
        <div><h3>{title}</h3>{description && <p>{description}</p>}</div>
        {extra}
    </div>
);

const SettingForm = ({styles, className = "", children, ...props}: any) => (
    <Form
        {...props}
        layout="vertical"
        variant="filled"
        className={`${styles.settingsForm} ${className}`.trim()}>
        {children}
    </Form>
);

const SettingField = ({wide, full, className = "", children, ...props}: any) => (
    <Form.Item
        {...props}
        className={`setting-field ${wide ? "setting-wide" : ""} ${full ? "setting-full" : ""} ${className}`.trim()}>
        {children}
    </Form.Item>
);

const SettingSwitch = ({name, title, description, styles, ...props}: any) => (
    <div className={styles.toggleRow}>
        <div>
            <div className="toggle-title">{title}</div>
            {description && <div className="toggle-description">{description}</div>}
        </div>
        <Form.Item name={name} valuePropName="checked" noStyle><Switch {...props}/></Form.Item>
    </div>
);

const HeaderRules = ({name: listName, label, styles}: any) => (
    <Form.List name={listName}>
        {(fields, {add, remove}) => <>
            <div className="header-summary">
                {listName === "request_rules" ? "修改发往网站或上游服务的请求头。" : "修改返回给浏览器的响应头。"}
                名称不区分大小写。
            </div>
            {fields.length > 0 && <div className="header-columns" aria-hidden="true"><span>处理方式</span><span>名称</span><span>内容</span></div>}
            <div className="header-rule-list">
                {fields.length === 0 && <div className="header-rule-empty">暂无数据</div>}
                {fields.map(({key, name}) => <div className="header-rule" key={key}>
                    <Form.Item name={[name, "operation"]} rules={[{required: true, message: "请选择操作"}]}>
                        <Select
                            size="small"
                            aria-label={`${label}处理方式`}
                            placeholder="处理方式"
                            options={[
                                {label: "设置 / 覆盖", value: "set", title: "设置内容，已有同名头时替换原值"},
                                {label: "追加内容", value: "add", title: "保留原有内容，再追加一个值"},
                                {label: "移除此头", value: "delete", title: "移除这个名称对应的全部内容"},
                            ]}/>
                    </Form.Item>
                    <Form.Item
                        name={[name, "name"]}
                        rules={[{required: true, whitespace: true, message: `请输入${label}名称`}]}
                        normalize={(value) => String(value || "").trimStart()}>
                        <Input size="small" aria-label={`${label}名称`} placeholder={listName === "request_rules" ? "例如 X-Real-IP" : "例如 Cache-Control"} autoComplete="off"/>
                    </Form.Item>
                    <Form.Item noStyle shouldUpdate={(before, after) => before?.[listName]?.[name]?.operation !== after?.[listName]?.[name]?.operation}>
                        {({getFieldValue}) => {
                            const deleting = getFieldValue([listName, name, "operation"]) === "delete";
                            return <Form.Item
                                className="header-value"
                                name={[name, "value"]}
                                dependencies={[[listName, name, "operation"]]}
                                rules={[{
                                    validator: async (_, value) => {
                                        if (!deleting && !String(value || "").trim()) throw new Error(`请输入${label}内容`);
                                    },
                                }]}>
                                <Input
                                    size="small"
                                    aria-label={`${label}内容`}
                                    disabled={deleting}
                                    placeholder={deleting ? "无需填写内容" : listName === "request_rules" ? "输入需要传递的内容" : "例如 no-cache"}
                                    autoComplete="off"/>
                            </Form.Item>;
                        }}
                    </Form.Item>
                    <Button
                        color="danger"
                        variant="filled"
                        style={{borderColor: "transparent"}}
                        className="header-rule-remove"
                        title="删除规则"
                        aria-label={`删除${label}规则 ${name + 1}`}
                        icon={<Icon type="CloseOutlined"/>}
                        onClick={() => remove(name)}/>
                </div>)}
            </div>
            <div className="header-rule-toolbar">
                <Button
                    color="default"
                    variant="filled"
                    style={{borderColor: "transparent"}}
                    className={styles.settingAction}
                    size="small"
                    icon={<Icon type="PlusOutlined"/>}
                    onClick={() => add({operation: "set", name: "", value: ""})}>
                    添加{label}
                </Button>
            </div>
        </>}
    </Form.List>
);

const WAFConditions = ({ruleName, groupName, fields, add, remove, styles}: any) => {
    const form = Form.useFormInstance();
    const groupMatch = Form.useWatch(["rules", ruleName, "groups", groupName, "match"], {form, preserve: true}) || "all";
    const conditionPath = (conditionName: number, field: string) => ["rules", ruleName, "groups", groupName, "conditions", conditionName, field];
    return <>
        <div className="waf-condition-list">
            {fields.map(({key, name}: any, index: number) => <div className="waf-condition" key={key}>
                <Form.Item noStyle shouldUpdate={(before, after) => {
                    const previous = before?.rules?.[ruleName]?.groups?.[groupName]?.conditions?.[name];
                    const current = after?.rules?.[ruleName]?.groups?.[groupName]?.conditions?.[name];
                    return previous?.target !== current?.target || previous?.operator !== current?.operator
                        || previous?.negated !== current?.negated || previous?.ignore_case !== current?.ignore_case;
                }}>{({getFieldValue}) => {
                    const target = String(getFieldValue(conditionPath(name, "target")) || "REQUEST_URI");
                    const operator = String(getFieldValue(conditionPath(name, "operator")) || "contains");
                    const negated = getFieldValue(conditionPath(name, "negated")) === "not";
                    const ignoreCase = Boolean(getFieldValue(conditionPath(name, "ignore_case")));
                    const targetOption = WAF_TARGET_OPTIONS.find((item) => item.value === target) || WAF_TARGET_OPTIONS[0];
                    const numeric = target === "ARGS_COMBINED_SIZE";
                    const valueless = WAF_VALUELESS_OPERATORS.has(operator);
                    return <div className="waf-condition-row">
                        <span className="waf-condition-join">{index === 0 ? "如果" : groupMatch === "any" ? "或者" : "并且"}</span>
                        <Form.Item
                            className="waf-condition-target"
                            name={[name, "target"]}
                            rules={[{required: true, message: "请选择检查位置"}]}>
                            <Select
                                size="small"
                                showSearch
                                optionFilterProp="label"
                                title={targetOption.description}
                                placeholder="检查位置"
                                options={WAF_TARGET_OPTION_GROUPS}
                                onChange={(value) => {
                                    const option = WAF_TARGET_OPTIONS.find((item) => item.value === value) || WAF_TARGET_OPTIONS[0];
                                    const current = form.getFieldValue(conditionPath(name, "operator"));
                                    if (!current || !option.operators.some((item) => item.value === current)) {
                                        form.setFieldValue(conditionPath(name, "operator"), option.operator || "contains");
                                        form.setFieldValue(conditionPath(name, "negated"), "match");
                                    }
                                }}/>
                        </Form.Item>
                        <Form.Item name={[name, "operator"]} hidden rules={[{required: true}]}><Input/></Form.Item>
                        <Form.Item name={[name, "negated"]} hidden><Input/></Form.Item>
                        <Select
                            className="waf-condition-operator"
                            size="small"
                            value={negated ? `not:${operator}` : operator}
                            placeholder="匹配方式"
                            options={targetOption.options}
                            onChange={(value) => {
                                const negative = String(value).startsWith("not:");
                                form.setFieldValue(conditionPath(name, "operator"), negative ? String(value).slice(4) : value);
                                form.setFieldValue(conditionPath(name, "negated"), negative ? "not" : "match");
                            }}/>
                        <Form.Item
                            className="waf-condition-value"
                            name={[name, "value"]}
                            rules={valueless ? [] : [
                                {required: true, whitespace: true, message: "请输入匹配内容"},
                                ...(numeric ? [{pattern: /^\d+$/, message: "请输入非负整数"}] : []),
                            ]}>
                            <Input
                                size="small"
                                disabled={valueless}
                                inputMode={numeric ? "numeric" : undefined}
                                maxLength={2048}
                                placeholder={valueless ? "无需填写内容" : numeric ? "请输入比较整数" : targetOption.placeholder}/>
                        </Form.Item>
                        <Form.Item name={[name, "ignore_case"]} valuePropName="checked" hidden><Switch/></Form.Item>
                        <div className="waf-condition-actions">
                            {WAF_CASE_OPERATORS.has(operator) && !numeric && <Tooltip title={ignoreCase ? "已忽略英文大小写" : "忽略英文大小写"}>
                                <Button
                                    color={ignoreCase ? "primary" : "default"}
                                    variant="text"
                                    className={`${styles.ruleAction} waf-case-button ${ignoreCase ? "is-active" : ""}`}
                                    size="small"
                                    aria-label="忽略英文大小写"
                                    aria-pressed={ignoreCase}
                                    onClick={() => form.setFieldValue(conditionPath(name, "ignore_case"), !ignoreCase)}>
                                    Aa
                                </Button>
                            </Tooltip>}
                            <Tooltip title="删除条件">
                                <Button
                                    color="danger"
                                    variant="text"
                                    className={styles.ruleAction}
                                    size="small"
                                    aria-label={`删除条件 ${index + 1}`}
                                    icon={<Icon type="CloseOutlined"/>}
                                    disabled={fields.length <= 1}
                                    onClick={() => remove(name)}/>
                            </Tooltip>
                        </div>
                    </div>;
                }}</Form.Item>
            </div>)}
        </div>
        <Button
            color="default"
            variant="filled"
            style={{borderColor: "transparent"}}
            className={styles.settingAction}
            size="small"
            icon={<Icon type="PlusOutlined"/>}
            onClick={() => add({target: "REQUEST_URI", operator: "contains", value: "", negated: "match", ignore_case: true})}>
            添加条件
        </Button>
    </>;
};

const WAFGroupJoin = ({ruleName}: any) => {
    const form = Form.useFormInstance();
    const match = Form.useWatch(["rules", ruleName, "match"], {form, preserve: true}) || "any";
    return <div className="waf-group-join">{match === "all" ? "并且" : "或者"}</div>;
};

const WAFMatchSelect = ({path, label, options}: any) => {
    const form = Form.useFormInstance();
    const value = Form.useWatch(path, {form, preserve: true}) || options[0]?.value;
    return <Select
        size="small"
        value={value}
        aria-label={label}
        options={options}
        onChange={(next) => form.setFieldValue(path, next)}/>;
};

const WAFRuleEditor = ({field, styles}: any) => {
    const ruleName = field.name;
    return <div className="waf-rule-editor">
        <div className="waf-rule-name-row">
            <span>规则名称</span>
            <Form.Item
                name={[ruleName, "name"]}
                rules={[{required: true, whitespace: true, message: "请输入规则名称"}]}>
                <Input size="small" placeholder="例如：拦截后台扫描" maxLength={80}/>
            </Form.Item>
        </div>
        <div className="waf-rule-sentence waf-rule-action-row">
            <span>命中后</span>
            <Form.Item name={[ruleName, "action"]} noStyle>
                <Select
                    size="small"
                    aria-label="命中后的操作"
                    options={[{label: "拦截访问", value: "block"}, {label: "仅记日志", value: "log"}]}/>
            </Form.Item>
        </div>
        <Form.List
            name={[ruleName, "groups"]}
            rules={[{validator: async (_, groups) => {
                if (!Array.isArray(groups) || groups.length === 0) throw new Error("请至少添加一种情况");
            }}]}>
            {(groupFields, {add: addGroup, remove: removeGroup}, {errors: groupErrors}) => <>
                {groupFields.length > 1 && <div className="waf-rule-sentence waf-rule-match">
                    <span>符合</span>
                    <WAFMatchSelect
                        path={["rules", ruleName, "match"]}
                        label="情况关系"
                        options={[{label: "任一情况", value: "any"}, {label: "所有情况", value: "all"}]}/>
                    <span>即可命中</span>
                </div>}
                <div className="waf-situation-list">
                    {groupFields.map(({key, name: groupName}, groupIndex) => <div className="waf-situation-entry" key={key}>
                        {groupIndex > 0 && <WAFGroupJoin ruleName={ruleName}/>}
                        <div className="waf-situation" data-multiple={groupFields.length > 1}>
                            <Form.List
                                name={[groupName, "conditions"]}
                                rules={[{validator: async (_, conditions) => {
                                    if (!Array.isArray(conditions) || conditions.length === 0) throw new Error("每种情况至少需要一个条件");
                                }}]}>
                                {(conditionFields, conditionActions, {errors}) => <>
                                    {(groupFields.length > 1 || conditionFields.length > 1) && <div className="waf-situation-heading">
                                        <div className="waf-situation-sentence">
                                            {groupFields.length > 1 && <strong>情况 {groupIndex + 1}</strong>}
                                            {conditionFields.length > 1 && <>
                                                <span>满足</span>
                                                <WAFMatchSelect
                                                    path={["rules", ruleName, "groups", groupName, "match"]}
                                                    label={`情况 ${groupIndex + 1} 的条件关系`}
                                                    options={[{label: "全部条件", value: "all"}, {label: "任一条件", value: "any"}]}/>
                                            </>}
                                        </div>
                                        {groupFields.length > 1 && <Button
                                            color="danger"
                                            variant="text"
                                            className={styles.ruleAction}
                                            size="small"
                                            onClick={() => removeGroup(groupName)}>
                                            删除
                                        </Button>}
                                    </div>}
                                    <WAFConditions
                                        ruleName={ruleName}
                                        groupName={groupName}
                                        fields={conditionFields}
                                        add={conditionActions.add}
                                        remove={conditionActions.remove}
                                        styles={styles}/>
                                    <Form.ErrorList errors={errors}/>
                                </>}
                            </Form.List>
                        </div>
                    </div>)}
                </div>
                <Form.ErrorList errors={groupErrors}/>
                <Button
                    color="default"
                    variant="filled"
                    style={{borderColor: "transparent"}}
                    className={styles.settingAction}
                    size="small"
                    icon={<Icon type="PlusOutlined"/>}
                    onClick={() => addGroup({
                        match: "all",
                        conditions: [{target: "REQUEST_URI", operator: "contains", value: "", negated: "match", ignore_case: true}],
                    })}>
                    添加情况
                </Button>
            </>}
        </Form.List>
    </div>;
};

const WAFRuleBuilder = forwardRef(({fields, add, remove, errors, styles}: any, ref: any) => {
    const form = Form.useFormInstance();
    const [activeKey, setActiveKey] = useState<any>();
    const openAdded = useRef(false);
    const scrollToRule = useRef<any>();
    const listRef = useRef<HTMLDivElement>(null);
    const signature = fields.map((field: any) => String(field.key)).join("|");
    const activeField = fields.find((field: any) => String(field.key) === String(activeKey));

    useEffect(() => {
        if (fields.length === 0) {
            if (activeKey !== undefined) setActiveKey(undefined);
            return;
        }
        if (openAdded.current) {
            openAdded.current = false;
            scrollToRule.current = fields[fields.length - 1].key;
            setActiveKey(fields[fields.length - 1].key);
            return;
        }
        if (activeKey === undefined || (activeKey !== null && !fields.some((field: any) => String(field.key) === String(activeKey)))) {
            setActiveKey(fields[0].key);
        }
    }, [activeKey, signature]);

    useEffect(() => {
        if (activeField && scrollToRule.current === activeField.key) {
            scrollToRule.current = undefined;
            listRef.current?.querySelector(`[data-rule-key="${activeField.key}"]`)?.scrollIntoView({block: "nearest", inline: "nearest"});
        }
    }, [activeKey, signature]);

    useImperativeHandle(ref, () => ({
        reveal(errorFields: any[]) {
            const error = errorFields.find((item) => item.name?.[0] === "rules" && typeof item.name?.[1] === "number");
            const field = fields.find((item: any) => item.name === error?.name[1]);
            if (!field) return;
            setActiveKey(field.key);
            listRef.current?.querySelector(`[data-rule-key="${field.key}"]`)?.scrollIntoView({block: "nearest", inline: "nearest"});
        },
    }), [fields]);

    const addRule = (source?: any) => {
        const names = new Set((form.getFieldValue("rules") || []).map((item: any) => String(item?.name || "").trim()));
        const base = source ? `${source.name || "防护规则"} 副本` : "防护规则";
        let number = 1;
        let name = source ? base : `${base} ${number}`;
        while (names.has(name)) name = `${base} ${++number}`;
        openAdded.current = true;
        add({
            ...(source || {}),
            id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
            name,
            enabled: source?.enabled !== false,
            match: source?.match || "any",
            action: source?.action || "block",
            groups: source?.groups?.map((group: any) => ({
                ...group,
                conditions: (group.conditions || []).map((condition: any) => ({...condition})),
            })) || [{
                match: "all",
                conditions: [{target: "REQUEST_URI", operator: "contains", value: "", negated: "match", ignore_case: true}],
            }],
        });
    };
    const addButton = <Button color="default" variant="filled" style={{borderColor: "transparent"}} className={styles.settingAction} size="small" icon={<Icon type="PlusOutlined"/>} onClick={() => addRule()}>添加规则</Button>;

    return <>
        <div className="waf-rules-header">
            <div className="waf-rules-title">自定义防护规则</div>
            {fields.length > 0 && addButton}
        </div>
        {fields.length === 0 ? <div className="waf-rules-empty">
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
            {addButton}
        </div> : <>
            <div className="waf-rule-list" ref={listRef}>
                {fields.map((field: any, index: number) => {
                    const active = field.key === activeField?.key;
                    return <div className="waf-rule-entry" key={field.key} data-rule-key={field.key}>
                        <div className="waf-rule-summary" data-active={active}>
                            <Form.Item noStyle shouldUpdate={(before, after) => before?.rules?.[field.name] !== after?.rules?.[field.name]}>
                                {({getFieldValue}) => {
                                    const rule = getFieldValue(["rules", field.name]) || {};
                                    const groups = Array.isArray(rule.groups) ? rule.groups : [];
                                    const conditions = groups.reduce((count: number, group: any) => count + (Array.isArray(group?.conditions) ? group.conditions.length : 0), 0);
                                    const relation = groups.length > 1
                                        ? `符合${rule.match === "all" ? "所有" : "任一"}情况`
                                        : `满足${groups[0]?.match === "any" ? "任一" : "全部"}条件`;
                                    return <>
                                        <button className="waf-rule-select" type="button" aria-expanded={active} aria-label={`${active ? "收起" : "展开"}防护规则 ${index + 1}`} onClick={() => setActiveKey(active ? null : field.key)}>
                                            <Icon type={active ? "DownOutlined" : "RightOutlined"}/>
                                            <span className="waf-rule-summary-text">
                                                <span className="waf-rule-title">
                                                    <span className="waf-rule-name">{String(rule.name || `未命名规则 ${index + 1}`)}</span>
                                                    <span className="waf-rule-action">{rule.enabled === false ? "已停用" : rule.action === "log" ? "命中后记录" : "命中后拦截"}</span>
                                                </span>
                                                <span className="waf-rule-meta">{relation} · {groups.length > 1 ? `${groups.length} 种情况 · ` : ""}{conditions} 个条件</span>
                                            </span>
                                        </button>
                                        <div className="waf-rule-actions">
                                            <Form.Item name={[field.name, "enabled"]} valuePropName="checked" noStyle>
                                                <Switch
                                                    className="waf-rule-switch"
                                                    size="small"
                                                    aria-label={`启用规则 ${index + 1}`}/>
                                            </Form.Item>
                                            <ActionDropdown menu={{
                                                items: [
                                                    {key: "copy", label: "复制规则"},
                                                    {key: "delete", label: "删除规则", danger: true},
                                                ],
                                                onClick: ({key}) => {
                                                    if (key === "copy") addRule(rule);
                                                    if (key === "delete") {
                                                        const next = fields[index + 1] || fields[index - 1];
                                                        if (active) setActiveKey(next?.key ?? null);
                                                        remove(field.name);
                                                    }
                                                },
                                            }}>
                                                <Button color="default" variant="text" className={styles.ruleAction} size="small" aria-label={`防护规则 ${index + 1} 操作`} icon={<Icon type="MoreOutlined"/>}/>
                                            </ActionDropdown>
                                        </div>
                                    </>;
                                }}
                            </Form.Item>
                        </div>
                        <div className="waf-rule-body" hidden={!active}><WAFRuleEditor field={field} styles={styles}/></div>
                    </div>;
                })}
            </div>
        </>}
        <Form.ErrorList errors={errors}/>
    </>;
});

const LoadingForm = ({loading, children}: any) => (
    <Spin spinning={loading} size="medium">
        <div style={loading ? {minHeight: 380, visibility: "hidden"} : undefined}>{children}</div>
    </Spin>
);

const useSettingsFormFooter = (loading: boolean, saving: boolean) => {
    const context = useContext(SettingsFormContext);
    useEffect(() => {
        context?.update?.(context.formId, {loading, saving});
    }, [context, loading, saving]);
    useEffect(() => () => context?.update?.(context.formId), [context]);
    return context?.formId;
};

const SettingsFooter = memo(forwardRef(({formId, onClose}: any, ref: any) => {
    const [state, setState] = useState({loading: true, saving: false});
    useImperativeHandle(ref, () => ({
        update: (next?: any) => setState(next || {loading: true, saving: false}),
    }), []);
    if (state.loading) return null;
    return (
        <>
            <Button onClick={onClose}>关闭</Button>
            <Button
                type="primary"
                htmlType="submit"
                form={formId}
                loading={state.saving}>
                保存
            </Button>
        </>
    );
}));

SettingsFooter.displayName = "SettingsFooter";

const ConfigForm = memo(({site, request, update, title, description, styles, normalize, serialize, children, extra, onFinishFailed, hideHeader = false}: any): any => {
    const {message} = App.useApp();
    const [form] = Form.useForm();
    const requestRef = useRef(0);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const formId = useSettingsFormFooter(loading, saving);

    const load = useCallback(async () => {
        const requestId = ++requestRef.current;
        setLoading(true);
        try {
            const response = await request({body: {action: "get", id: site.id}});
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code === 200) {
                form.resetFields();
                form.setFieldsValue(normalize ? normalize(response?.data || {}) : (response?.data || {}));
            } else {
                message.error(response?.message || `获取${title}失败`);
            }
        } catch {
            if (requestRef.current === requestId) {
                message.error(`获取${title}失败`);
            }
        } finally {
            if (requestRef.current === requestId) {
                setLoading(false);
            }
        }
    }, [form, message, normalize, request, site.id, title]);

    useEffect(() => {
        void load();
        return () => { requestRef.current += 1; };
    }, [load]);

    const submit = useCallback(async (values: any) => {
        const requestId = ++requestRef.current;
        setSaving(true);
        let reload = false;
        try {
            const response = await update({body: {
                action: "set",
                id: site.id,
                config: serialize ? serialize(values, form) : values,
            }});
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code === 200) {
                message.success(response?.message || "保存成功");
                reload = true;
            } else {
                message.error(response?.message || "保存失败");
            }
        } catch {
            if (requestRef.current === requestId) {
                message.error("保存失败");
            }
        } finally {
            if (requestRef.current === requestId) {
                setSaving(false);
            }
        }
        if (reload && requestRef.current === requestId) {
            void load();
        }
    }, [form, load, message, serialize, site.id, update]);

    return (
        <LoadingForm loading={loading}>
            {!hideHeader && <SectionHeader title={title} description={description} extra={extra?.({reload: load, form})} styles={styles}/>}
            {hideHeader && extra && <div className={styles.formToolbar}>{extra({reload: load, form})}</div>}
            <SettingForm styles={styles} id={formId} form={form} onFinish={submit} onFinishFailed={onFinishFailed}>
                {children}
            </SettingForm>
        </LoadingForm>
    );
});

const RunPathPicker = ({root, value = "", onChange, ...props}: any) => {
    const {message} = App.useApp();
    const initialPath = useMemo(() => {
        const websiteRoot = String(root || "").trim();
        return isAbsoluteFilePath(websiteRoot) ? normalizeFilePath(websiteRoot) : "";
    }, [root]);
    const select = useCallback((selected: string) => {
        if (!initialPath) {
            message.warning("请先选择网站目录");
            return false;
        }
        const selectedPath = normalizeFilePath(selected);
        if (!isFilePathWithin(selectedPath, initialPath)) {
            message.warning("运行目录必须位于网站目录内");
            return false;
        }
        const prefix = initialPath.endsWith("/") ? initialPath : `${initialPath}/`;
        const relative = comparableFilePath(selectedPath) === comparableFilePath(initialPath)
            ? ""
            : selectedPath.slice(prefix.length);
        onChange?.(relative);
        return true;
    }, [initialPath, message, onChange]);
    const change = useCallback((next: string) => {
        const input = String(next || "");
        if (isAbsoluteFilePath(input)) {
            select(input);
            return;
        }
        onChange?.(input);
    }, [onChange, select]);
    return (
        <FilePicker
            buttonProps={{color: "default", variant: "filled", style: {borderColor: "transparent"}}}
            {...props}
            mode="directory"
            value={value}
            initialPath={initialPath}
            onChange={change}
            onSelect={select}/>
    );
};

const BasicPane = ({site, styles, typeItems, onChange, hideHeader = false}: any) => {
    const {message} = App.useApp();
    const [form] = Form.useForm();
    const root = Form.useWatch("root", form);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const categoryOptions = useMemo(() => [
        {label: "全部分类", value: 0},
        ...(typeItems || []).map((item: any) => ({label: String(item.name), value: Number(item.id)})),
    ], [typeItems]);
    const formId = useSettingsFormFooter(loading, saving);
    useEffect(() => {
        let active = true;
        setLoading(true);
        void getSiteInfo({body: {id: site.id}}).then((response) => {
            if (!active) return;
            setLoading(false);
            if (response?.code === 200) form.setFieldsValue({type_id: 0, ...(response?.data || site)});
            else message.error(response?.message || "获取网站信息失败");
        });
        return () => { active = false; };
    }, [form, message, site]);
    const submit = async (values: any) => {
        setSaving(true);
        const body: any = {
            id: site.id,
            name: String(values.name || "").trim(),
            type_id: String(values.type_id || 0),
        };
        if (Number(site.type) !== 1) {
            body.root = String(values.root || "").trim();
            body.run_path = String(values.run_path || "").trim();
        }
        const response = await updateSite({body});
        setSaving(false);
        if (response?.code === 200) {
            clearEnumCache("site");
            message.success(response?.message || "保存成功");
            onChange?.({...site, ...body, type_id: Number(body.type_id)});
        } else message.error(response?.message || "保存失败");
    };
    return (
        <LoadingForm loading={loading}>
            {!hideHeader && <SectionHeader title="基础设置" description="修改网站名称与站点目录" styles={styles}/>}
            <SettingForm styles={styles} id={formId} form={form} onFinish={submit}>
                <SettingField name="name" label="网站名称" rules={[{required: true, whitespace: true, message: "请输入网站名称"}]}><Input placeholder="请输入网站名称" maxLength={100}/></SettingField>
                <SettingField name="type_id" label="网站分类">
                    <Select
                        placeholder="全部分类"
                        options={categoryOptions}/>
                </SettingField>
                {Number(site.type) !== 1 && (
                    <>
                        <SettingField className="site-file-picker" name="root" label="网站目录" extra="网站文件保存位置" rules={[{required: true, message: "请选择网站目录"}]}><FilePicker buttonProps={{color: "default", variant: "filled", style: {borderColor: "transparent"}}} mode="directory" placeholder="请输入或选择网站目录"/></SettingField>
                        <SettingField
                            className="site-file-picker"
                            name="run_path"
                            label="运行目录"
                            dependencies={["root"]}
                            extra="选择网站目录内的子目录，留空则直接使用网站目录"
                            rules={[{
                                validator: async (_, value) => {
                                    const runPath = String(value || "").trim();
                                    const websiteRoot = String(form.getFieldValue("root") || "").trim();
                                    if (!runPath || !isAbsoluteFilePath(websiteRoot)) return;
                                    if (isAbsoluteFilePath(runPath) || !isFilePathWithin(
                                        normalizeFilePath(joinFilePath(websiteRoot, runPath)),
                                        normalizeFilePath(websiteRoot),
                                    )) throw new Error("运行目录必须位于网站目录内");
                                },
                            }]}>
                            <RunPathPicker root={root} placeholder="例如 public，可留空"/>
                        </SettingField>
                    </>
                )}
            </SettingForm>
        </LoadingForm>
    );
};

const DomainPane = ({site, styles, onChange, hideHeader = false}: any) => {
    const {message} = App.useApp();
    const domainOptional = Number(site.type) === 2;
    const [form] = Form.useForm();
    const requestRef = useRef(0);
    const [loading, setLoading] = useState(true);
    const [loadError, setLoadError] = useState(false);
    const [saving, setSaving] = useState(false);
    const [input, setInput] = useState("");
    const [inputError, setInputError] = useState("");
    const formId = useSettingsFormFooter(loading || loadError, saving);
    const load = useCallback(async () => {
        const requestId = ++requestRef.current;
        setLoading(true);
        setLoadError(false);
        try {
            const response = await siteDomain({body: {action: "get", id: site.id}});
            if (requestRef.current !== requestId) return;
            if (response?.code === 200) {
                const records = Array.isArray(response?.data) ? response?.data.filter(Boolean) : [];
                form.resetFields();
                form.setFieldsValue({domains: records.map((item: any) => String(item.domain || "").trim()).filter(Boolean)});
                setInput("");
                setInputError("");
            } else {
                setLoadError(true);
                message.error(response?.message || "获取域名失败");
            }
        } catch {
            if (requestRef.current === requestId) {
                setLoadError(true);
                message.error("获取域名失败");
            }
        } finally {
            if (requestRef.current === requestId) setLoading(false);
        }
    }, [form, message, site.id]);
    useEffect(() => {
        void load();
        return () => { requestRef.current += 1; };
    }, [load]);

    const parseDomains = (text: string) => {
        const values = Array.from(new Set(text.split(/[\s,，;；]+/)
            .filter(Boolean)
            .map((value) => value.toLowerCase().replace(/\.$/, "").replace(/^\[([^\]]+)\]$/, "$1"))));
        for (const value of values) {
            if (!value) throw new Error("请输入有效的域名或 IP");
            if (value.length > 253) throw new Error("域名不能超过 253 个字符");
            if (/[\/@?#\[\]\\]/.test(value) || (value.includes("*") && !/^\*\.[^*]+$/.test(value))) {
                throw new Error(`「${value}」格式不正确，请只填写域名或 IP，不要包含协议和路径`);
            }
            if ((value.match(/:/g) || []).length === 1) throw new Error("域名不需要填写端口");
        }
        return values;
    };
    const getDomainBatch = (text: string, current: string[]) => {
        try {
            const incoming = parseDomains(text);
            const existing = new Set(current.flatMap((value) => {
                try { return parseDomains(value); } catch { return []; }
            }));
            const additions = incoming.filter((value) => !existing.has(value));
            if (current.length + additions.length > 1000) throw new Error("每个网站最多绑定 1000 个域名");
            return {
                additions,
                skipped: text.split(/[\s,，;；]+/).filter(Boolean).length - additions.length,
                error: "",
            };
        } catch (error) {
            return {additions: [] as string[], skipped: 0, error: (error as Error).message};
        }
    };
    const validateDomains = async (_: any, values: string[] = []) => {
        if (!domainOptional && !values.length && !input.trim()) throw new Error("请至少添加一个域名");
        if (values.length > 1000) throw new Error("每个网站最多绑定 1000 个域名");
        const unique = new Set<string>();
        for (const value of values) {
            const parsed = parseDomains(value || "");
            if (parsed.length !== 1) throw new Error("每项请填写一个域名或 IP");
            if (unique.has(parsed[0])) throw new Error(`域名「${parsed[0]}」重复，请删除重复行`);
            unique.add(parsed[0]);
        }
    };
    const addDomains = () => {
        if (saving) return;
        try {
            const current: string[] = form.getFieldValue("domains") || [];
            const batch = getDomainBatch(input, current);
            if (batch.error) throw new Error(batch.error);
            if (!batch.additions.length) throw new Error(batch.skipped ? "这些域名已经在列表中" : "请输入要添加的域名");
            form.setFieldsValue({domains: [...current, ...batch.additions]});
            setInput("");
            setInputError("");
            message.info({key: formId + "-save-reminder", content: "域名已添加到列表，请点击“保存”生效"});
        } catch (error) {
            setInputError((error as Error).message);
        }
    };
    const submit = async () => {
        if (saving || loading || loadError) return;
        let values: string[];
        try {
            // 只读行未挂载输入字段，需从表单存储读取完整域名列表。
            const current: string[] = form.getFieldValue("domains") || [];
            values = Array.from(new Set([...current.flatMap(parseDomains), ...parseDomains(input)]));
            if (!domainOptional && !values.length) throw new Error("请至少添加一个域名");
            if (values.length > 1000) throw new Error("每个网站最多绑定 1000 个域名");
        } catch (error) {
            setInputError((error as Error).message);
            return;
        }
        setInputError("");
        const requestId = ++requestRef.current;
        setSaving(true);
        let reload = false;
        try {
            const response = await siteDomain({body: {action: "set", id: site.id, domains: values}});
            if (requestRef.current !== requestId) return;
            if (response?.code === 200) {
                message.success({key: formId + "-save-reminder", content: response?.message || "保存成功"});
                onChange?.();
                reload = true;
            } else message.error(response?.message || "保存失败");
        } catch {
            if (requestRef.current === requestId) message.error("保存失败");
        } finally {
            if (requestRef.current === requestId) setSaving(false);
        }
        if (reload && requestRef.current === requestId) void load();
    };
    if (loadError) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="域名加载失败，请重试">
        <Button color="default" variant="filled" style={{borderColor: "transparent"}} className={styles.settingAction} icon={<Icon type="ReloadOutlined"/>} onClick={() => void load()}>重新加载</Button>
    </Empty>;
    return <LoadingForm loading={loading}>
        {!hideHeader && <SectionHeader title="域名管理" description={domainOptional ? "设置项目的访问域名，也可不绑定域名独立运行。" : "管理网站的访问域名，支持通配符域名和 IP。"} styles={styles}/>}
        <SettingForm styles={styles} id={formId} form={form} onFinish={submit} disabled={saving} scrollToFirstError={{block: "nearest", focus: true}}>
            <div className={styles.domainEditor}>
                <Form.List name="domains" rules={[{validator: validateDomains}]}>
                    {(fields, {remove}, {errors}) => {
                        const batch = getDomainBatch(input, form.getFieldValue("domains") || []);
                        return <>
                            <div className="domain-composer">
                                <div className="domain-composer-heading">
                                    <h4>添加域名</h4>
                                </div>
                                <Form.Item validateStatus={inputError ? "error" : undefined} help={inputError || undefined}>
                                    <Input.TextArea
                                        aria-label="批量添加域名"
                                        aria-describedby={formId + "-domain-batch-status"}
                                        placeholder={"example.com\nwww.example.com\n*.example.com"}
                                        rows={3}
                                        autoComplete="off"
                                        autoCapitalize="none"
                                        spellCheck={false}
                                        value={input}
                                        onChange={(event) => {setInput(event.target.value); setInputError("");}}/>
                                </Form.Item>
                                <div className="domain-composer-footer">
                                    <span className="domain-batch-status" id={formId + "-domain-batch-status"} aria-live="polite">
                                        {!input.trim() ? "换行、空格、逗号或分号分隔" : batch.error ? "请检查输入格式和域名数量。" : <>
                                            待添加 <strong>{batch.additions.length}</strong> 个域名
                                            {batch.skipped > 0 && `，已忽略 ${batch.skipped} 个重复项`}
                                        </>}
                                    </span>
                                    <div className="domain-composer-actions">
                                        <Button color="default" variant="filled" style={{borderColor: "transparent"}} htmlType="button" className={styles.settingAction} disabled={saving || !input.trim()} onClick={addDomains}>添加</Button>
                                    </div>
                                </div>
                            </div>
                            <div className="domain-toolbar">
                                <div className="domain-list-title">域名列表</div>
                            </div>
                            <div className="domain-errors" role="alert"><Form.ErrorList errors={errors}/></div>
                            <div className="domain-list-container">
                                <div className="domain-list-heading" aria-hidden="true"><span>域名</span><span>操作</span></div>
                                <div className="domain-list" role="list" aria-label="域名列表">
                                    {fields.map(({key, name}) => {
                                        const domain = String(form.getFieldValue(["domains", name]) || "");
                                        return <div className="domain-row" role="listitem" key={key}>
                                            <div className="domain-main">
                                                <span className="domain-value">{domain}</span>
                                            </div>
                                            <div className="domain-actions">
                                                <Button color="danger" variant="filled" style={{borderColor: "transparent"}} htmlType="button" className={styles.settingAction}
                                                    disabled={saving || fields.length <= 1}
                                                    title={fields.length <= 1 ? "至少保留一个域名" : undefined}
                                                    aria-label={"删除域名 " + domain}
                                                    onClick={() => {
                                                        if (saving || (form.getFieldValue("domains") || []).length <= 1) return;
                                                        remove(name);
                                                        message.info({key: formId + "-save-reminder", content: "域名已从列表移除，请点击“保存”生效"});
                                                    }}>删除</Button>
                                            </div>
                                        </div>;
                                    })}
                                    {!fields.length && <div className="domain-empty">
                                        <Icon type="GlobalOutlined"/>
                                        <div className="domain-empty-title">暂未添加域名</div>
                                        <p>{domainOptional ? "项目可独立运行，需要域名访问时再添加。" : "在上方输入域名，可一次添加多条。"}</p>
                                    </div>}
                                </div>
                            </div>
                        </>;
                    }}
                </Form.List>
            </div>
        </SettingForm>
    </LoadingForm>;
};

const normalizeCertDomain = (value: string) => value.trim().toLowerCase().replace(/\.+$/, "");
const certCoversDomain = (certificateDomain: string, domain: string) => {
    if (certificateDomain === domain) return true;
    if (!certificateDomain.startsWith("*.") || domain.startsWith("*.") || domain.includes(":") || /^\d+(?:\.\d+){3}$/.test(domain)) return false;
    const suffix = certificateDomain.slice(1);
    if (!domain.endsWith(suffix)) return false;
    const label = domain.slice(0, -suffix.length);
    return Boolean(label) && !label.includes(".");
};

const SSLPane = ({site, styles, onManageCert, onObtainCert, onDeployCert, onRenewCert, settingsSessionRef, certVersion = 0, hideHeader = false}: any) => {
    const {message} = App.useApp();
    const [form] = Form.useForm();
    const bindings = Form.useWatch("bindings", {form, preserve: true}) || {};
    const [now, setNow] = useState(() => Date.now());
    const requestRef = useRef(0);
    const certRequestRef = useRef(0);
    const obtainRequestRef = useRef(0);
    const [loading, setLoading] = useState(true);
    const [loadError, setLoadError] = useState(false);
    const [certLoading, setCertLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const formId = useSettingsFormFooter(loading || loadError, saving);
    const [domains, setDomains] = useState<any[]>([]);
    const [selectedDomainIds, setSelectedDomainIds] = useState<Set<string>>(new Set());
    const [certs, setCerts] = useState<any[]>([]);
    const selectedDomains = useMemo(() => domains.filter((item) => selectedDomainIds.has(String(item.id))), [domains, selectedDomainIds]);
    const certMap = useMemo(() => {
        const records = new Map<number, any>();
        for (const certificate of certs) {
            const id = Number(certificate.id);
            if (!Number.isNaN(id) && !records.has(id)) records.set(id, certificate);
        }
        return records;
    }, [certs]);
    useEffect(() => {
        const timer = window.setInterval(() => setNow(Date.now()), 60000);
        return () => window.clearInterval(timer);
    }, []);
    const load = useCallback(async () => {
        const requestId = ++requestRef.current;
        setLoading(true);
        setLoadError(false);
        try {
            const response = await siteSSL({body: {action: "get", id: site.id}});
            if (requestRef.current !== requestId) return;
            if (response?.code !== 200) {
                setLoadError(true);
                message.error(response?.message || "获取 HTTPS 配置失败");
                return;
            }
            const nextDomains = Array.isArray(response?.data?.domains) ? response?.data?.domains.filter(Boolean) : [];
            const config = response?.data?.config || {};
            setDomains(nextDomains);
            setSelectedDomainIds(new Set());
            form.setFieldsValue({
                ...config,
                min_version: config.min_version || undefined,
                max_version: config.max_version || undefined,
                bindings: Object.fromEntries(nextDomains.map((item: any) => [String(item.id), Number(item.cert_id || 0) || undefined])),
            });
        } catch {
            if (requestRef.current === requestId) {
                setLoadError(true);
                message.error("获取 HTTPS 配置失败");
            }
        } finally {
            if (requestRef.current === requestId) setLoading(false);
        }
    }, [form, message, site.id]);
    useEffect(() => {
        void load();
        return () => {requestRef.current += 1;};
    }, [load]);
    useEffect(() => {
        const requestId = ++certRequestRef.current;
        setCertLoading(true);
        void getCertList({body: {page: 1, page_size: 1000, order_by_field: "id", order_by_type: "desc"}}).then((response) => {
            if (certRequestRef.current !== requestId) return;
            if (response?.code === 200) setCerts(Array.isArray(response?.data?.list) ? response?.data?.list.filter(Boolean) : []);
            else message.error(response?.message || "获取证书列表失败");
        }).catch(() => {
            if (certRequestRef.current === requestId) message.error("获取证书列表失败");
        }).finally(() => {
            if (certRequestRef.current === requestId) setCertLoading(false);
        });
        return () => {certRequestRef.current += 1;};
    }, [certVersion, message]);
    const openCertificate = (selection: any, action: "obtain" | "deploy" | "renew") => {
        if (saving || loading || loadError) return;
        if (action === "obtain" && !onObtainCert || action === "deploy" && !onDeployCert || action === "renew" && !onRenewCert) return;
        const selected = Array.isArray(selection) ? selection : [selection];
        const targets = [...new Set<string>(selected.map((item) => normalizeCertDomain(String(item.domain || ""))).filter(Boolean))];
        if (!targets.length) return;
        const requestId = requestRef.current;
        const obtainRequestId = ++obtainRequestRef.current;
        const settingsSession = settingsSessionRef.current;
        const matchPriorities = new Map<string, number>();
        const onSelected = (certificate: any) => {
            if (requestRef.current !== requestId || obtainRequestRef.current !== obtainRequestId || settingsSessionRef.current !== settingsSession || !certificate?.id) return;
            let certificateDomains: unknown = certificate.domains;
            if (typeof certificateDomains === "string") {
                try { certificateDomains = JSON.parse(certificateDomains); } catch { certificateDomains = []; }
            }
            const coveredDomains = Array.isArray(certificateDomains)
                ? certificateDomains.filter((value): value is string => typeof value === "string").map(normalizeCertDomain)
                : [];
            certRequestRef.current += 1;
            setCertLoading(false);
            setCerts((current) => [certificate, ...current.filter((item) => String(item.id) !== String(certificate.id))]);
            let changed = false;
            selected.forEach((item) => {
                const domain = normalizeCertDomain(String(item.domain || ""));
                const priority = coveredDomains.includes(domain) ? 2 : coveredDomains.some((value) => certCoversDomain(value, domain)) ? 1 : 0;
                const id = String(item.id);
                if (!priority || priority < (matchPriorities.get(id) || 0)) return;
                matchPriorities.set(id, priority);
                if (Number(form.getFieldValue(["bindings", id])) !== Number(certificate.id)) changed = true;
                form.setFieldValue(["bindings", id], Number(certificate.id));
            });
            setNow(Date.now());
            if (changed) message.info({key: formId + "-save-reminder", content: "证书已关联，请点击“保存”生效"});
        };
        if (action === "deploy") onDeployCert(targets[0], onSelected);
        else if (action === "renew") {
            const id = Number(form.getFieldValue(["bindings", String(selected[0].id)]));
            if (id) onRenewCert(certMap.get(id) || {id}, onSelected, targets[0]);
        } else onObtainCert(Array.isArray(selection) ? targets : targets[0], onSelected);
    };
    const submit = async () => {
        if (saving || loading || loadError) return;
        // 折叠的连接设置可能尚未挂载，从表单存储保留完整 TLS 配置。
        const values = form.getFieldsValue(true);
        const requestId = ++requestRef.current;
        setSaving(true);
        try {
            const response = await siteSSL({body: {
                action: "set",
                id: site.id,
                config: {
                    redirect_http: Boolean(values.redirect_http),
                    min_version: values.min_version || "",
                    max_version: values.max_version || "",
                },
                bindings: domains.map((item) => ({domain_id: item.id, cert_id: Number(values.bindings?.[String(item.id)] || 0)})),
            }});
            if (requestRef.current !== requestId) return;
            if (response?.code === 200) {
                message.success(response?.message || "保存成功");
                setSaving(false);
                void load();
            } else message.error(response?.message || "保存失败");
        } catch {
            if (requestRef.current === requestId) message.error("保存失败");
        } finally {
            if (requestRef.current === requestId) setSaving(false);
        }
    };
    if (loadError) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="HTTPS 配置加载失败，请重试">
        <Button color="default" variant="filled" style={{borderColor: "transparent"}} className={styles.settingAction} icon={<Icon type="ReloadOutlined"/>} onClick={() => void load()}>重新加载</Button>
    </Empty>;
    return (
        <LoadingForm loading={loading}>
            {!hideHeader && (
                <SectionHeader
                    title="HTTPS 证书"
                    description="查看证书部署状态，申请证书后保存生效"
                    styles={styles}
                    extra={<Button color="default" variant="filled" style={{borderColor: "transparent"}} htmlType="button" className={styles.settingAction} size="small" onClick={onManageCert}>证书管理</Button>}/>
            )}
            <SettingForm styles={styles} id={formId} form={form} onFinish={submit} disabled={saving}>
                <div className={`${styles.domainEditor} ${styles.httpsTable}`}>
                    <div className="domain-toolbar">
                        <div className="domain-list-title">域名证书</div>
                        <Button color="default" variant="filled" style={{borderColor: "transparent", visibility: selectedDomains.length ? "visible" : "hidden"}} htmlType="button" className={styles.settingAction}
                            aria-hidden={!selectedDomains.length}
                            disabled={!selectedDomains.length || saving || loading || !onObtainCert}
                            onClick={() => openCertificate(selectedDomains, "obtain")}>批量申请</Button>
                    </div>
                    <div className="domain-list-container">
                        <div className="https-table-scroll">
                            <div className="domain-list-heading">
                                <span className="https-domain">
                                    <Checkbox aria-label="全选域名" disabled={saving || loading || !domains.length}
                                        checked={domains.length > 0 && selectedDomains.length === domains.length}
                                        indeterminate={selectedDomains.length > 0 && selectedDomains.length < domains.length}
                                        onChange={(event) => setSelectedDomainIds(new Set(event.target.checked ? domains.map((item) => String(item.id)) : []))}/>
                                    <span>域名</span>
                                </span><span>证书</span><span>操作</span>
                            </div>
                            <div className="domain-list" role="list" aria-label="域名证书列表">
                                {domains.map((item) => {
                                    const domain = String(item.domain || "");
                                    const certId = Number(bindings[String(item.id)] || 0);
                                    const pending = certId !== Number(item.cert_id || 0);
                                    const certificate = certMap.get(certId);
                                    const expiry = certificate?.expire_time ? dayjs(String(certificate.expire_time)) : null;
                                    const remaining = expiry?.isValid() ? expiry.valueOf() - now : null;
                                    const expired = remaining !== null && remaining <= 0;
                                    const days = remaining === null ? 0 : Math.floor(remaining / 86400000);
                                    const hours = remaining === null ? 0 : Math.floor(remaining / 3600000) % 24;
                                    const expiryText = remaining === null ? certLoading ? "正在加载有效期" : "有效期未知"
                                        : expired ? "已过期"
                                            : days > 0 ? `剩余 ${days} 天${hours > 0 ? ` ${hours} 小时` : ""}`
                                                : hours > 0 ? `剩余 ${hours} 小时`
                                                    : `剩余 ${Math.max(1, Math.ceil(remaining / 60000))} 分钟`;
                                    return <div className="domain-row" role="listitem" key={item.id}>
                                        <div className="https-domain">
                                            <Checkbox aria-label={"选择域名 " + domain} disabled={saving || loading}
                                                checked={selectedDomainIds.has(String(item.id))}
                                                onChange={(event) => setSelectedDomainIds((current) => {
                                                    const next = new Set(current);
                                                    if (event.target.checked) next.add(String(item.id));
                                                    else next.delete(String(item.id));
                                                    return next;
                                                })}/>
                                            <span className="domain-value">{domain}</span>
                                        </div>
                                        <div className="https-certificate">
                                            {certId ? (
                                                <span className="https-certificate-status"
                                                    data-expired={expired || undefined}
                                                    data-expiring={!expired && remaining !== null && remaining <= 86400000 || undefined}
                                                    title={`${pending ? "待部署" : "已部署"}${expiry?.isValid() ? ` · ${String(certificate.expire_time).replace("T", " ").slice(0, 16)} 到期` : ""}`}>
                                                    {expiryText}
                                                </span>
                                            ) : <span className="https-certificate-empty">—</span>}
                                        </div>
                                        <div className="https-operations">
                                            <div className="domain-actions">
                                                <Button color="default" variant="filled" style={{borderColor: "transparent"}} htmlType="button" className={styles.settingAction}
                                                    disabled={saving || (certId ? certLoading || !onRenewCert : !onDeployCert)}
                                                    aria-label={"为 " + domain + (certId ? " 续签证书" : " 部署证书")}
                                                    onClick={() => openCertificate(item, certId ? "renew" : "deploy")}>{certId ? "续签" : "部署"}</Button>
                                                {certId ? (
                                                    <Button color="danger" variant="filled" style={{borderColor: "transparent"}} htmlType="button" className={styles.settingAction}
                                                        disabled={saving}
                                                        aria-label={"关闭 " + domain + " 的 HTTPS"}
                                                        onClick={() => form.setFieldValue(["bindings", String(item.id)], undefined)}>关闭</Button>
                                                ) : null}
                                            </div>
                                        </div>
                                    </div>;
                                })}
                                {!domains.length && <div className="domain-empty">
                                    <div className="domain-empty-title">暂无域名</div>
                                    <p>请先在域名管理中添加域名。</p>
                                </div>}
                            </div>
                        </div>
                    </div>
                </div>
                <SettingSwitch name="redirect_http" title="自动跳转 HTTPS" description="开启后，已部署证书的域名会将 HTTP 访问跳转到 HTTPS" styles={styles}/>
                <Collapse ghost expandIconPlacement="end" items={[{key: "advanced", label: "连接设置", children: <>
                    <Form.Item name="min_version" label="最低 TLS 版本"><Select allowClear placeholder="自动选择最低版本" options={[{label: "TLS 1.2", value: "tls1.2"}, {label: "TLS 1.3", value: "tls1.3"}]}/></Form.Item>
                    <Form.Item name="max_version" label="最高 TLS 版本"><Select allowClear placeholder="自动选择最高版本" options={[{label: "TLS 1.2", value: "tls1.2"}, {label: "TLS 1.3", value: "tls1.3"}]}/></Form.Item>
                </>}]}/>
            </SettingForm>
        </LoadingForm>
    );
};

const StaticPane = ({site, styles}: any) => (
    <ConfigForm
        site={site} request={siteStatic} update={siteStatic} title="静态网站设置" description="默认配置即可直接托管 HTML 和静态资源" styles={styles}
        normalize={(data: any) => ({
            index: Array.isArray(data.index) ? data.index : ["index.html", "index.htm"],
            browse: Boolean(data.browse),
            try_files: Array.isArray(data.try_files) ? data.try_files : [],
            hide: Array.isArray(data.hide) ? data.hide : [],
            precompressed: Array.isArray(data.precompressed) ? data.precompressed : ["zstd", "gzip"],
        })}
        serialize={(data: any) => ({
            index: splitValues(data.index),
            browse: Boolean(data.browse),
            try_files: data.browse ? [] : splitValues(data.try_files),
            hide: splitValues(data.hide),
            precompressed: splitValues(data.precompressed),
        })}>
        <Form.Item name="index" label="默认首页" extra="访问目录时按顺序查找这些首页文件">
            <Select mode="tags" tokenSeparators={[",", " "]} placeholder="index.html, index.htm"/>
        </Form.Item>
        <SettingSwitch
            name="browse"
            title="目录浏览"
            description="开启后访问目录显示文件列表，不再自动打开首页文件"
            styles={styles}/>
        <Collapse ghost expandIconPlacement="end" items={[{
            key: "advanced",
            label: "高级设置",
            children: <>
                <Form.Item noStyle shouldUpdate={(before, after) => before.browse !== after.browse}>{({getFieldValue}) => (
                    <Form.Item name="try_files" label="文件匹配顺序" extra={getFieldValue("browse") ? "开启目录浏览时自动留空，保证子目录可以正常进入" : "留空时按请求路径直接查找文件"}>
                        <Select disabled={Boolean(getFieldValue("browse"))} mode="tags" tokenSeparators={[",", " "]} placeholder="例如 {path}, /index.html"/>
                    </Form.Item>
                )}</Form.Item>
                <Form.Item name="hide" label="禁止访问的文件"><Select mode="tags" tokenSeparators={[",", " "]} placeholder="例如 .git, .env"/></Form.Item>
                <Form.Item name="precompressed" label="预压缩文件"><Select mode="multiple" placeholder="请选择预压缩格式" options={[{label: "Zstandard", value: "zstd"}, {label: "Gzip", value: "gzip"}]}/></Form.Item>
            </>,
        }]}/>
    </ConfigForm>
);

const ProxyPane = ({site, styles}: any) => {
    const configRef = useRef<any>({});
    const normalize = useCallback((data: any) => {
        const upstreams = upstreamsToForm(data.upstreams);
        const dial = String(upstreams[0] || "127.0.0.1:8080");
        const unix = /^unix(?:\+h2c)?\//.test(dial);
        const versions = Array.isArray(data.versions) ? data.versions : [];
        const requestSet = data.headers?.request_set || {};
        const hostKey = Object.keys(requestSet).find((key) => key.toLowerCase() === "host");
        const hostValue = hostKey ? requestSet[hostKey] : "";
        const sendHost = Array.isArray(hostValue) ? hostValue[0] : hostValue;
        const values = {
            ...data,
            versions,
            policy: data.policy || "random",
            retries: Number(data.retries || 0),
            target: unix
                ? (dial.startsWith("unix+h2c/") || versions.includes("h2c") ? dial.replace(/^unix(?:\+h2c)?\//, "unix+h2c/") : dial.replace(/^unix\//, ""))
                : (/^https?:\/\//i.test(dial) ? dial : `${data.scheme || "http"}://${dial}`),
            send_host: sendHost === "{http.request.host}" ? "$http_host" : String(sendHost || ""),
            backup_upstreams: upstreams.slice(1),
            try_duration: seconds(data.try_duration),
            try_interval: seconds(data.try_interval),
            dial_timeout: seconds(data.dial_timeout),
            read_timeout: seconds(data.read_timeout),
            write_timeout: seconds(data.write_timeout),
            health_interval: seconds(data.health_interval),
            health_timeout: seconds(data.health_timeout),
            tls_insecure_skip_verify: data.tls_insecure_skip_verify == null ? true : Boolean(data.tls_insecure_skip_verify),
        };
        configRef.current = values;
        return values;
    }, []);
    const serialize = useCallback((values: any) => {
        const data = {...configRef.current, ...values};
        const target = parseProxyTarget(data.target);
        const unix = target.dial.startsWith("unix/");
        const upstreams = [target, ...splitValues(data.backup_upstreams).map((value) => parseProxyTarget(value, target.scheme))];
        const limits = new Map<string, number>();
        for (const item of data.upstreams || []) {
            try {
                limits.set(parseProxyTarget(item.dial, data.scheme || "http").dial, Number(item.max_requests || 0));
            } catch {
                // 原地址无法识别时允许用户修正，新地址使用默认限制。
            }
        }
        const sendHost = String(data.send_host || "").trim();
        const hostname = target.hostname;
        const automaticHost = unix || hostname === "localhost" || hostname.includes(":") || /^\d{1,3}(?:\.\d{1,3}){3}$/.test(hostname)
            ? "{http.request.host}"
            : hostname;
        return {
            upstreams: splitValues(upstreams.map((item) => item.dial)).map((dial) => ({dial, max_requests: limits.get(dial) || 0})),
            scheme: target.scheme,
            versions: upstreams.some((item) => item.h2c) ? ["h2c"] : (data.versions || []).filter((version: string) =>
                (version !== "h2c" || !unix && target.scheme === "http") && (version !== "3" || target.scheme === "https")),
            policy: data.policy || "random",
            retries: Number(data.retries || 0),
            try_duration: duration(data.try_duration),
            try_interval: duration(data.try_interval),
            dial_timeout: duration(data.dial_timeout),
            read_timeout: duration(data.read_timeout),
            write_timeout: duration(data.write_timeout),
            tls_server_name: target.scheme === "https" ? String(data.tls_server_name || "").trim() : "",
            tls_insecure_skip_verify: target.scheme === "https" && Boolean(data.tls_insecure_skip_verify),
            health_uri: String(data.health_uri || "").trim(),
            health_interval: duration(data.health_interval),
            health_timeout: duration(data.health_timeout),
            health_status: Number(data.health_status || 0),
            headers: {
                ...data.headers,
                request_set: {
                    ...Object.fromEntries(Object.entries(data.headers?.request_set || {}).filter(([name]) => name.toLowerCase() !== "host")),
                    Host: [sendHost ? (sendHost === "$http_host" ? "{http.request.host}" : sendHost) : automaticHost],
                },
            },
        };
    }, []);
    return (
        <ConfigForm
            site={site} request={siteProxy} update={siteProxy} title="反向代理设置" description="将网站请求转发到另一个 HTTP 服务或 UNIX 套接字" styles={styles}
            normalize={normalize} serialize={serialize}>
            <Form.Item name="target" label="目标地址" extra="自动识别 HTTP、HTTPS 和 UNIX 套接字；网址不填协议时使用 HTTP。"
                rules={[{required: true, whitespace: true, message: "请输入目标地址"}, {validator: async (_, value) => {
                    if (String(value || "").trim()) parseProxyTarget(value);
                }}]}>
                <Input placeholder="http://127.0.0.1:15700 或 /tmp/panel.sock"/>
            </Form.Item>
            <Form.Item name="send_host" label="发送域名（Host）" extra="留空使用目标域名；代理本机 IP 时可填写 $http_host 透传访问域名">
                <Input placeholder="自动使用目标域名"/>
            </Form.Item>
            <Collapse ghost expandIconPlacement="end" items={[{
                key: "advanced",
                label: "高级设置",
                children: <>
                    <Form.Item name="backup_upstreams" label="备用上游" dependencies={["target"]} rules={[({getFieldValue}) => ({validator: async (_, value) => {
                        const values = splitValues(value);
                        if (!values.length) return;
                        const target = parseProxyTarget(getFieldValue("target"));
                        if (values.some((address) => parseProxyTarget(address, target.scheme).scheme !== target.scheme)) {
                            throw new Error("备用上游必须与目标地址使用相同协议");
                        }
                    }})]}><Select mode="tags" tokenSeparators={[",", " "]} placeholder="可添加相同协议的备用地址"/></Form.Item>
                    <div className={styles.fieldGrid}>
                        <Form.Item name="policy" label="负载均衡"><Select placeholder="请选择负载均衡策略" options={[{label: "随机", value: "random"}, {label: "轮询", value: "round_robin"}, {label: "最少连接", value: "least_conn"}, {label: "客户端 IP", value: "client_ip_hash"}, {label: "URI", value: "uri_hash"}]}/></Form.Item>
                        <Form.Item name="retries" label="重试次数"><InputNumber placeholder="请输入重试次数" min={0} precision={0}/></Form.Item>
                    </div>
                    <div className={`${styles.fieldGrid} setting-grid-numeric`}>
                        <Form.Item name="dial_timeout" label="连接超时（秒）"><InputNumber placeholder="连接超时" min={0}/></Form.Item>
                        <Form.Item name="read_timeout" label="读取超时（秒）"><InputNumber placeholder="读取超时" min={0}/></Form.Item>
                        <Form.Item name="write_timeout" label="写入超时（秒）"><InputNumber placeholder="写入超时" min={0}/></Form.Item>
                    </div>
                    <div className={styles.fieldGrid}>
                        <Form.Item name="tls_server_name" label="上游 TLS SNI"><Input placeholder="默认使用目标域名"/></Form.Item>
                        <Form.Item name="tls_insecure_skip_verify" label="忽略上游证书校验" valuePropName="checked"><Switch/></Form.Item>
                    </div>
                    <SettingField full name="health_uri" label="健康检查路径"><Input placeholder="例如 /health"/></SettingField>
                    <div className={`${styles.fieldGrid} setting-grid-numeric`}>
                        <Form.Item name="health_interval" label="间隔（秒）"><InputNumber placeholder="检查间隔" min={0}/></Form.Item>
                        <Form.Item name="health_timeout" label="超时（秒）"><InputNumber placeholder="检查超时" min={0}/></Form.Item>
                        <Form.Item name="health_status" label="状态码"><InputNumber placeholder="期望状态码" min={100} max={599} precision={0}/></Form.Item>
                    </div>
                </>,
            }]}/>
        </ConfigForm>
    );
};

const FastCGIPane = ({site, styles}: any) => (
    <ConfigForm
        site={site} request={siteFastCGI} update={siteFastCGI} title="PHP 设置" description="选择 PHP-FPM 服务即可运行 PHP 网站" styles={styles}
        normalize={(data: any) => {
            const upstreams = upstreamsToForm(data.upstreams).map((item: any) => String(item).replace(/^unix(?:\+h2c)?\//, ""));
            return {
                policy: "random", index: "index.php", split_path: [".php"], try_files: [], capture_stderr: true,
                ...data,
                fpm_address: upstreams[0] || "127.0.0.1:9000",
                backup_upstreams: upstreams.slice(1),
                dial_timeout: seconds(data.dial_timeout), read_timeout: seconds(data.read_timeout), write_timeout: seconds(data.write_timeout),
                health_interval: seconds(data.health_interval), health_timeout: seconds(data.health_timeout), env_lines: envToLines(data.env),
            };
        }}
        serialize={(data: any) => {
            const upstreams = [data.fpm_address, ...splitValues(data.backup_upstreams)].map((item) => {
                const address = String(item || "").trim();
                return address.startsWith("/") ? `unix/${address}` : address;
            });
            const splitPath = splitValues(data.split_path);
            return {
                upstreams: upstreamsToBody(upstreams),
                policy: data.policy || "random", index: String(data.index || "index.php").trim(),
                split_path: splitPath.length ? splitPath : [".php"],
                try_files: splitValues(data.try_files), capture_stderr: Boolean(data.capture_stderr), resolve_root_symlink: Boolean(data.resolve_root_symlink),
                dial_timeout: duration(data.dial_timeout), read_timeout: duration(data.read_timeout), write_timeout: duration(data.write_timeout),
                health_uri: String(data.health_uri || "").trim(), health_interval: duration(data.health_interval), health_timeout: duration(data.health_timeout),
                health_status: Number(data.health_status || 0), env: linesToEnv(data.env_lines),
            };
        }}>
        <Form.Item name="fpm_address" label="PHP-FPM 地址" extra="可填写 TCP 地址或 UNIX 套接字" rules={[{required: true, whitespace: true, message: "请输入 PHP-FPM 地址"}]}>
            <Input placeholder="127.0.0.1:9000"/>
        </Form.Item>
        <Form.Item name="index" label="默认入口文件" rules={[{required: true, whitespace: true, message: "请输入入口文件"}]}><Input placeholder="index.php"/></Form.Item>
        <Collapse ghost expandIconPlacement="end" items={[{
            key: "advanced", label: "高级设置", children: <>
                <Form.Item name="backup_upstreams" label="备用 PHP-FPM"><Select mode="tags" tokenSeparators={[",", " "]} placeholder="输入备用 PHP-FPM 地址"/></Form.Item>
                <div className={styles.fieldGrid}>
                    <Form.Item name="split_path" label="脚本扩展名"><Select mode="tags" tokenSeparators={[",", " "]} placeholder="例如 .php"/></Form.Item>
                    <Form.Item name="policy" label="负载均衡"><Select placeholder="请选择负载均衡策略" options={[{label: "随机", value: "random"}, {label: "轮询", value: "round_robin"}, {label: "最少连接", value: "least_conn"}]}/></Form.Item>
                </div>
                <Form.Item name="try_files" label="文件匹配顺序"><Select mode="tags" tokenSeparators={[",", " "]} placeholder="留空使用默认规则"/></Form.Item>
                <div className={styles.fieldGrid}>
                    <Form.Item name="capture_stderr" label="记录 PHP 错误" valuePropName="checked"><Switch/></Form.Item>
                    <Form.Item name="resolve_root_symlink" label="解析目录软链接" valuePropName="checked"><Switch/></Form.Item>
                </div>
                <div className={`${styles.fieldGrid} setting-grid-numeric`}>
                    <Form.Item name="dial_timeout" label="连接超时（秒）"><InputNumber placeholder="连接超时" min={0}/></Form.Item>
                    <Form.Item name="read_timeout" label="读取超时（秒）"><InputNumber placeholder="读取超时" min={0}/></Form.Item>
                    <Form.Item name="write_timeout" label="写入超时（秒）"><InputNumber placeholder="写入超时" min={0}/></Form.Item>
                </div>
                <SettingField full name="health_uri" label="健康检查路径"><Input placeholder="例如 /status"/></SettingField>
                <div className={`${styles.fieldGrid} setting-grid-numeric`}>
                    <Form.Item name="health_interval" label="间隔（秒）"><InputNumber placeholder="检查间隔" min={0}/></Form.Item>
                    <Form.Item name="health_timeout" label="超时（秒）"><InputNumber placeholder="检查超时" min={0}/></Form.Item>
                    <Form.Item name="health_status" label="状态码"><InputNumber placeholder="期望状态码" min={100} max={599} precision={0}/></Form.Item>
                </div>
                <Form.Item name="env_lines" label="环境变量" extra="每行一个 KEY=VALUE"><Input.TextArea placeholder="例如 APP_ENV=production" autoSize={{minRows: 3, maxRows: 8}}/></Form.Item>
            </>,
        }]}/>
    </ConfigForm>
);

const GeneralPane = ({site, styles}: any) => {
    const {message} = App.useApp();
    const [form] = Form.useForm();
    const requestRef = useRef(0);
    const configRef = useRef<any>({});
    const logRef = useRef<any>(null);
    const [loading, setLoading] = useState(true);
    const [loadError, setLoadError] = useState(false);
    const [saving, setSaving] = useState(false);
    const formId = useSettingsFormFooter(loading || loadError, saving);
    const load = useCallback(async () => {
        const requestId = ++requestRef.current;
        setLoading(true);
        setLoadError(false);
        try {
            const [processResponse, proxyResponse] = await Promise.all([
                siteProcess({body: {action: "get", id: site.id}}),
                siteProxy({body: {action: "get", id: site.id}}),
            ]);
            if (requestRef.current !== requestId) return;
            if (processResponse?.code !== 200 || proxyResponse?.code !== 200) {
                setLoadError(true);
                message.error(processResponse?.code !== 200 ? (processResponse?.message || "获取运行配置失败") : (proxyResponse?.message || "获取本地代理失败"));
                return;
            }
            const process = processResponse?.data || {};
            const nextProxy = proxyResponse?.data || {};
            const dial = String(nextProxy.upstreams?.[0]?.dial || "127.0.0.1:8080");
            const match = dial.match(/:(\d+)$/);
            const requestSet = nextProxy.headers?.request_set || {};
            const hostKey = Object.keys(requestSet).find((key) => key.toLowerCase() === "host");
            const hostValue = hostKey ? requestSet[hostKey] : "";
            const sendHost = Array.isArray(hostValue) ? hostValue[0] : hostValue;
            const values = {
                command: process.command || "",
                port: Number(match?.[1] || 8080),
                environment_lines: envToLines(process.environment),
                max_retries: Number(process.max_retries || 3),
                restart_delay: seconds(process.restart_delay) || 2,
                stop_timeout: seconds(process.stop_timeout) || 10,
                send_host: sendHost === "{http.request.host}" ? "$http_host" : String(sendHost || ""),
                policy: nextProxy.policy || "random",
                retries: Number(nextProxy.retries || 0),
                try_duration: seconds(nextProxy.try_duration),
                try_interval: seconds(nextProxy.try_interval),
                dial_timeout: seconds(nextProxy.dial_timeout),
                read_timeout: seconds(nextProxy.read_timeout),
                write_timeout: seconds(nextProxy.write_timeout),
                response_header_timeout: seconds(nextProxy.response_header_timeout),
                flush_interval: Number(nextProxy.flush_interval || 0) / NS,
                stream_timeout: seconds(nextProxy.stream_timeout),
                stream_close_delay: seconds(nextProxy.stream_close_delay),
                health_uri: nextProxy.health_uri || "",
                health_interval: seconds(nextProxy.health_interval),
                health_timeout: seconds(nextProxy.health_timeout),
                health_status: Number(nextProxy.health_status || 0) || undefined,
            };
            configRef.current = {proxy: nextProxy, values};
            form.resetFields();
            form.setFieldsValue(values);
        } catch {
            if (requestRef.current === requestId) {
                setLoadError(true);
                message.error("获取项目设置失败");
            }
        } finally {
            if (requestRef.current === requestId) setLoading(false);
        }
    }, [form, message, site.id]);
    useEffect(() => {
        void load();
        return () => {requestRef.current += 1;};
    }, [load]);
    const submit = async () => {
        if (saving || loading || loadError) return;
        const requestId = ++requestRef.current;
        const values = {...configRef.current.values, ...form.getFieldsValue(true)};
        const proxy = configRef.current.proxy || {};
        const upstreams = Array.isArray(proxy.upstreams) ? proxy.upstreams : [];
        const upstream = upstreams[0] || {};
        const proxyConfig: any = {
            upstreams: [{
                max_requests: 0,
                ...upstream,
                dial: Number(values.port) === configRef.current.values.port && upstream.dial
                    ? upstream.dial : `127.0.0.1:${Number(values.port)}`,
            }, ...upstreams.slice(1)],
            policy: values.policy || "random",
            retries: Number(values.retries || 0),
            try_duration: duration(values.try_duration),
            try_interval: duration(values.try_interval),
            dial_timeout: duration(values.dial_timeout),
            read_timeout: duration(values.read_timeout),
            write_timeout: duration(values.write_timeout),
            response_header_timeout: duration(values.response_header_timeout),
            flush_interval: Number(values.flush_interval || 0) * NS,
            stream_timeout: duration(values.stream_timeout),
            stream_close_delay: duration(values.stream_close_delay),
            health_uri: String(values.health_uri || "").trim(),
            health_interval: duration(values.health_interval),
            health_timeout: duration(values.health_timeout),
            health_status: Number(values.health_status || 0),
        };
        const sendHost = String(values.send_host || "").trim();
        if (sendHost !== configRef.current.values.send_host) {
            proxyConfig.headers = {
                request_set: {
                    ...Object.fromEntries(Object.entries(proxy.headers?.request_set || {}).filter(([name]) => name.toLowerCase() !== "host")),
                    ...(sendHost ? {Host: [sendHost === "$http_host" || sendHost === "$host" ? "{http.request.host}" : sendHost]} : {}),
                },
            };
        }
        setSaving(true);
        let reload = false;
        let processSaved = false;
        try {
            const processResponse = await siteProcess({body: {action: "set", id: site.id, config: {
                command: String(values.command || "").trim(), environment: linesToEnv(values.environment_lines),
                max_retries: Number(values.max_retries),
                restart_delay: duration(values.restart_delay), stop_timeout: duration(values.stop_timeout),
            }}});
            if (requestRef.current !== requestId) return;
            if (processResponse?.code !== 200) {
                message.error(processResponse?.message || "保存运行配置失败");
                return;
            }
            processSaved = true;
            const proxyResponse = await siteProxy({body: {action: "set", id: site.id, config: proxyConfig}});
            if (requestRef.current !== requestId) return;
            if (proxyResponse?.code === 200) {
                message.success(proxyResponse?.message || "保存成功");
                reload = true;
            } else message.error(`运行配置已保存，本地代理保存失败：${proxyResponse?.message || "请重试"}`);
        } catch {
            if (requestRef.current === requestId) message.error(processSaved ? "运行配置已保存，本地代理保存失败，请重试" : "保存运行配置失败");
        } finally {
            if (requestRef.current === requestId) setSaving(false);
        }
        if (reload && requestRef.current === requestId) void load();
    };
    if (loadError) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="项目设置加载失败，请重试">
        <Button color="default" variant="filled" style={{borderColor: "transparent"}} className={styles.settingAction} icon={<Icon type="ReloadOutlined"/>} onClick={() => void load()}>重新加载</Button>
    </Empty>;
    return <LoadingForm loading={loading}>
        <SectionHeader title="项目设置" description="设置项目启动方式和本机代理，未绑定域名时也可独立运行" styles={styles}
            extra={<Button color="default" variant="filled" className={styles.settingAction} style={{flexShrink: 0, borderColor: "transparent"}} icon={<Icon type="CodeOutlined"/>} disabled={saving} onClick={() => logRef.current?.open(site, {type: "process"})}>进程日志</Button>}/>
        <ProcessControls key={site.id} siteId={site.id} disabled={saving || loading}/>
        <SettingForm styles={styles} id={formId} form={form} onFinish={submit} disabled={saving}>
            <SettingField wide name="command" label="启动命令" extra="请输入完整执行命令和所需参数" rules={[{required: true, whitespace: true, message: "请输入启动命令"}]}>
                <Input placeholder="例如 /www/app/server --port 8080"/>
            </SettingField>
            <SettingField name="port" label="项目端口" extra="项目必须监听本机的这个端口" rules={[{required: true, message: "请输入项目端口"}]}>
                <InputNumber placeholder="请输入项目端口" min={1} max={65535} precision={0} style={{width: "100%"}}/>
            </SettingField>
            <SettingField name="max_retries" label="失败重试次数" extra="首次失败后最多重试指定次数，耗尽后停止自动保活，网站和代理配置不变。可通过上方按钮启动、停止或重启项目；稳定运行 1 分钟后重置次数。" rules={[{required: true, type: "integer", min: 1, message: "请输入大于 0 的整数"}]}>
                <InputNumber placeholder="默认 3 次" min={1} precision={0} style={{width: "100%"}}/>
            </SettingField>
            <Collapse ghost expandIconPlacement="end" items={[{
                key: "process", label: "进程设置", children: <>
                    <Form.Item name="environment_lines" label="环境变量" extra="每行一个 KEY=VALUE"><Input.TextArea placeholder="例如 NODE_ENV=production" autoSize={{minRows: 3, maxRows: 8}}/></Form.Item>
                    <div className={styles.fieldGrid}>
                        <Form.Item name="restart_delay" label="异常重启等待（秒）"><InputNumber placeholder="请输入重启等待时间" min={0}/></Form.Item>
                        <Form.Item name="stop_timeout" label="停止等待（秒）"><InputNumber placeholder="请输入停止等待时间" min={1}/></Form.Item>
                    </div>
                </>,
            }, {
                key: "proxy", label: "本机反向代理", children: <>
                    <SettingField wide name="send_host" label="发送域名（Host）" extra="留空透传访问域名，也可填写指定域名或 $http_host。">
                        <Input placeholder="默认透传访问域名"/>
                    </SettingField>
                    <SettingField name="policy" label="负载均衡"><Select placeholder="请选择负载均衡策略" options={[{label: "随机", value: "random"}, {label: "轮询", value: "round_robin"}, {label: "最少连接", value: "least_conn"}, {label: "客户端 IP", value: "client_ip_hash"}, {label: "URI", value: "uri_hash"}]}/></SettingField>
                    <div className={`${styles.fieldGrid} setting-grid-numeric`}>
                        <Form.Item name="retries" label="请求重试次数" extra="上游请求失败时的重试次数，与进程保活次数独立。"><InputNumber placeholder="默认不重试" min={0} precision={0}/></Form.Item>
                        <Form.Item name="try_interval" label="重试间隔（秒）"><InputNumber placeholder="默认间隔" min={0}/></Form.Item>
                        <Form.Item name="try_duration" label="重试总时长（秒）"><InputNumber placeholder="0 按重试次数" min={0}/></Form.Item>
                    </div>
                    <div className={`${styles.fieldGrid} setting-grid-numeric`}>
                        <Form.Item name="dial_timeout" label="连接超时（秒）"><InputNumber placeholder="默认超时" min={0}/></Form.Item>
                        <Form.Item name="read_timeout" label="读取超时（秒）"><InputNumber placeholder="默认超时" min={0}/></Form.Item>
                        <Form.Item name="write_timeout" label="写入超时（秒）"><InputNumber placeholder="默认超时" min={0}/></Form.Item>
                    </div>
                    <SettingField name="response_header_timeout" label="响应头超时（秒）"><InputNumber placeholder="等待上游响应头的时间" min={0}/></SettingField>
                    <div className={`${styles.fieldGrid} setting-grid-numeric`}>
                        <Form.Item name="flush_interval" label="刷新间隔（秒）" extra="0 使用默认策略，负数表示立即刷新响应。"><InputNumber placeholder="默认刷新策略"/></Form.Item>
                        <Form.Item name="stream_timeout" label="流式超时（秒）" extra="流式连接的最长持续时间，0 为不限制。"><InputNumber placeholder="0 为不限制" min={0}/></Form.Item>
                        <Form.Item name="stream_close_delay" label="流式关闭延迟（秒）" extra="配置重载后，延迟关闭已有流式连接。"><InputNumber placeholder="默认立即关闭" min={0}/></Form.Item>
                    </div>
                    <SettingField full name="health_uri" label="健康检查路径" extra="留空关闭主动检查；配置后定期请求此路径，失败时暂停向该上游转发请求。" rules={[{validator: async (_, value) => {
                        const path = String(value || "").trim();
                        if (path && !path.startsWith("/")) throw new Error("健康检查路径须以 / 开头");
                    }}]}><Input placeholder="例如 /health"/></SettingField>
                    <div className={`${styles.fieldGrid} setting-grid-numeric`}>
                        <Form.Item name="health_interval" label="检查间隔（秒）"><InputNumber placeholder="默认间隔" min={0}/></Form.Item>
                        <Form.Item name="health_timeout" label="检查超时（秒）"><InputNumber placeholder="默认超时" min={0}/></Form.Item>
                        <Form.Item name="health_status" label="期望状态码"><InputNumber placeholder="默认 200" min={100} max={599} precision={0}/></Form.Item>
                    </div>
                </>,
            }]}/>
        </SettingForm>
        <Log ref={logRef} request={getSiteLog} title="进程日志"/>
    </LoadingForm>;
};

const WAFPane = ({site, styles, hideHeader = false}: any) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const rulesRef = useRef<any>(null);
    return <ConfigForm
        site={site}
        request={siteWAF}
        update={siteWAF}
        title="Web 防火墙"
        description="识别 SQL 注入、跨站脚本、恶意扫描等常见攻击"
        styles={styles}
        hideHeader={hideHeader}
        onFinishFailed={({errorFields}: any) => rulesRef.current?.reveal(errorFields)}
        normalize={(data: any) => {
            const defaultOWASP = {
                enabled: true,
                paranoia_level: 1,
                detection_paranoia_level: 1,
                inbound_anomaly_score_threshold: 5,
                outbound_anomaly_score_threshold: 4,
                sampling_percentage: 100,
            };
            const rawOWASP = data.owasp || {};
            const hasCustomProtection = (Array.isArray(data.rules) && data.rules.some((item: any) => item?.enabled && (
                Array.isArray(item?.groups) && item.groups.length
            )))
                || (Array.isArray(data.directives) && data.directives.some((item: any) => String(item || "").trim()));
            const hasOWASPConfig = Boolean(rawOWASP.enabled || rawOWASP.paranoia_level || rawOWASP.inbound_anomaly_score_threshold);
            const owasp = hasOWASPConfig || hasCustomProtection
                ? {...defaultOWASP, ...rawOWASP}
                : defaultOWASP;
            if (!owasp.enabled && !hasCustomProtection) {
                owasp.enabled = true;
            }
            const protection_level = Object.entries(WAF_PROTECTION_PRESETS).find(([, preset]) =>
                Object.entries(preset).every(([key, value]) => Number(owasp[key]) === value),
            )?.[0] || "custom";

            const rules = (Array.isArray(data.rules) ? data.rules : []).map((item: any) => ({
                id: String(item?.id || ""),
                name: String(item?.name || ""),
                enabled: Boolean(item?.enabled),
                match: item?.match === "all" ? "all" : "any",
                action: item?.action === "log" ? "log" : "block",
                groups: (Array.isArray(item?.groups) ? item.groups : []).map(normalizeWAFGroup),
            }));
            return {
                ...data,
                enabled: Boolean(data.enabled),
                mode: data.mode || "block",
                block_page: String(data.block_page || ""),
                audit_log: data.audit_log == null ? true : Boolean(data.audit_log),
                request_body_limit_mb: Number(data.request_body_limit || 16 * 1024 * 1024) / 1024 / 1024,
                owasp,
                protection_level,
                rules,
                directives_text: (Array.isArray(data.directives) ? data.directives : []).join("\n"),
            };
        }}
        serialize={(_values: any, form: any) => {
            // Include preserved fields when protection controls or expert settings are hidden.
            const data = form.getFieldsValue(true);
            const owasp = {...(data.owasp || {}), ...(WAF_PROTECTION_PRESETS[data.protection_level] || {})};
            owasp.detection_paranoia_level = Math.max(Number(owasp.paranoia_level || 1), Number(owasp.detection_paranoia_level || 0));
            const rules = (Array.isArray(data.rules) ? data.rules : []).map((item: any) => {
                return {
                    id: String(item.id || "").trim(),
                    name: String(item.name || "").trim(),
                    enabled: Boolean(item.enabled),
                    match: item.match === "all" ? "all" : "any",
                    action: item.action === "log" ? "log" : "block",
                    groups: (Array.isArray(item.groups) ? item.groups : []).map((group: any) => ({
                        match: group?.match === "any" ? "any" : "all",
                        conditions: (Array.isArray(group?.conditions) ? group.conditions : []).map((condition: any) => {
                            const operator = String(condition?.operator || "contains");
                            const target = String(condition?.target || "").trim();
                            return {
                                target,
                                operator,
                                value: WAF_VALUELESS_OPERATORS.has(operator) ? "" : String(condition?.value || "").trim(),
                                negated: condition?.negated === "not",
                                ignore_case: target !== "ARGS_COMBINED_SIZE" && WAF_CASE_OPERATORS.has(operator) && Boolean(condition?.ignore_case),
                            };
                        }),
                    })),
                };
            });

            return {
                enabled: Boolean(data.enabled),
                mode: data.mode || "block",
                block_page: String(data.block_page || ""),
                audit_log: Boolean(data.audit_log),
                request_body_limit: Math.round(Number(data.request_body_limit_mb || 0) * 1024 * 1024),
                owasp,
                directives: String(data.directives_text || "").split("\n").map((item) => item.trim()).filter(Boolean),
                rules,
            };
        }}>
        <SettingSwitch
            name="enabled"
            title="Web 防火墙"
            description="自动检查进入网站的请求，识别常见攻击行为"
            styles={styles}/>
        <Form.Item noStyle shouldUpdate={(before, after) => before.enabled !== after.enabled}>{({getFieldValue}) => getFieldValue("enabled") ? <>
            <SettingSwitch
                name={["owasp", "enabled"]}
                title="启用内置通用规则"
                description="覆盖常见注入、跨站脚本、协议异常和恶意文件请求"
                styles={styles}/>
            <Form.Item name="mode" label="防护模式" extra="建议先观察日志，确认业务正常后再拦截">
                <Segmented className={styles.wafMode} options={[{label: "拦截攻击", value: "block"}, {label: "仅观察", value: "detection"}]}/>
            </Form.Item>
            <Form.Item noStyle shouldUpdate={(before, after) => before.owasp?.enabled !== after.owasp?.enabled || before.protection_level !== after.protection_level}>{({getFieldValue, setFieldsValue}) => getFieldValue(["owasp", "enabled"]) ? <>
                <Form.Item name="protection_level" label="防护强度" extra="均衡模式兼顾安全性和兼容性">
                    <Segmented
                        className={styles.wafMode}
                        options={[{label: "基础", value: "loose"}, {label: "均衡", value: "recommended"}, {label: "严格", value: "strict"}, {label: "自定义", value: "custom"}]}
                        onChange={(value) => {
                            const preset = WAF_PROTECTION_PRESETS[String(value)];
                            if (preset) setFieldsValue({owasp: {...getFieldValue("owasp"), ...preset}});
                        }}/>
                </Form.Item>
                {getFieldValue("protection_level") === "custom" && <>
                    <Form.Item name={["owasp", "paranoia_level"]} label="规则敏感级别" extra="1 到 4 级，级别越高检查越严格，也越容易误拦截。" rules={[{required: true, type: "integer", min: 1, max: 4, message: "请输入 1 到 4 的整数"}]}>
                        <InputNumber placeholder="1 - 4" min={1} max={4} precision={0} style={{width: "100%"}}/>
                    </Form.Item>
                    <Form.Item name={["owasp", "inbound_anomaly_score_threshold"]} label="请求拦截阈值（分）" extra="访客请求命中规则会累加异常分数，达到或超过此值时拦截请求。越小越严格，均衡值为 5。" rules={[{required: true, type: "integer", min: 1, message: "请输入大于 0 的整数"}]}>
                        <InputNumber placeholder="均衡值 5" min={1} precision={0} style={{width: "100%"}}/>
                    </Form.Item>
                    <Form.Item name={["owasp", "outbound_anomaly_score_threshold"]} label="响应拦截阈值（分）" extra="网站返回的响应命中规则会累加异常分数，达到或超过此值时拦截响应。越小越严格，均衡值为 4。" rules={[{required: true, type: "integer", min: 1, message: "请输入大于 0 的整数"}]}>
                        <InputNumber placeholder="均衡值 4" min={1} precision={0} style={{width: "100%"}}/>
                    </Form.Item>
                    <div className="ant-form-item-extra">“仅观察”模式只记录命中情况，不执行拦截。</div>
                </>}
            </> : null}</Form.Item>
            <Form.Item name="request_body_limit_mb" label="请求内容检查上限（MB）" extra="超过此大小的请求按防护模式处理">
                <InputNumber placeholder="请输入检查上限" min={0} precision={0} style={{width: "100%", maxWidth: 360}}/>
            </Form.Item>
            <SettingSwitch
                name="audit_log"
                title="记录命中详情"
                description="保留命中规则、访客来源与处置结果，用于排查误拦截"
                styles={styles}/>

            <Collapse ghost expandIconPlacement="end" items={[{
                key: "block_page", label: "拦截页面", children: <Form.Item name="block_page" label="页面内容" extra="支持完整 HTML；留空使用系统默认页面。">
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
                        placeholder="留空使用系统默认页面"
                        className={styles.code}/>
                </Form.Item>,
            }]}/>

            <div className={styles.wafRules}>
                <Form.List
                    name="rules"
                    validateTrigger="onSubmit"
                    rules={[{validator: async (_, rules) => {
                        for (const [ruleIndex, rule] of (Array.isArray(rules) ? rules : []).entries()) {
                            if (!["any", "all"].includes(rule?.match) || !["block", "log"].includes(rule?.action)) {
                                throw new Error(`规则 ${ruleIndex + 1} 的执行方式无效`);
                            }
                            if (!Array.isArray(rule?.groups) || rule.groups.length === 0) throw new Error(`规则 ${ruleIndex + 1} 至少需要一种情况`);
                            for (const [groupIndex, group] of rule.groups.entries()) {
                                if (!["any", "all"].includes(group?.match)) throw new Error(`规则 ${ruleIndex + 1} 的情况关系无效`);
                                if (!Array.isArray(group?.conditions) || group.conditions.length === 0) {
                                    throw new Error(`规则 ${ruleIndex + 1} 的情况 ${groupIndex + 1} 至少需要一个条件`);
                                }
                                for (const condition of group.conditions) {
                                    const target = String(condition?.target || "");
                                    const operator = String(condition?.operator || "");
                                    if (!WAF_TARGET_OPTIONS.some((item) => item.value === target) || !WAF_OPERATOR_OPTIONS.some((item) => item.value === operator)) {
                                        throw new Error(`规则 ${ruleIndex + 1} 存在未完成的条件`);
                                    }
                                    const numeric = target === "ARGS_COMBINED_SIZE";
                                    if ((operator === "ip" && target !== "REMOTE_ADDR")
                                        || (["greater_than", "less_than"].includes(operator) && !numeric)
                                        || (numeric && !["equals", "greater_than", "less_than"].includes(operator))) {
                                        throw new Error(`规则 ${ruleIndex + 1} 的匹配方式不适用于检查位置`);
                                    }
                                }
                            }
                        }
                    }}]}>
                    {(fields, {add, remove}, {errors}) => <WAFRuleBuilder
                        ref={rulesRef}
                        fields={fields}
                        add={add}
                        remove={remove}
                        errors={errors}
                        styles={styles}/>}
                </Form.List>
            </div>

            <Collapse ghost expandIconPlacement="end" items={[{key: "expert", label: "专家设置", children: <>
                <Form.Item name="directives_text" label="底层 Coraza 指令" extra="每行一条原生 Coraza 配置；语法错误会导致网站配置无法生效，普通用户请保持为空"><Input.TextArea placeholder="普通用户无需填写" autoSize={{minRows: 3, maxRows: 10}}/></Form.Item>
            </>}]}/>
        </> : null}</Form.Item>
    </ConfigForm>;
};

const RatePane = ({site, styles, hideHeader = false}: any) => <ConfigForm site={site} request={siteRateLimit} update={siteRateLimit} title="访问频率" description="限制单个访客在指定时间内的请求次数" styles={styles} hideHeader={hideHeader}
    normalize={(data: any) => ({...data, enabled: Boolean(data.enabled), key: data.key || "{http.vars.client_ip}", window: seconds(data.window) || 60, max_events: Number(data.max_events || 100), paths: Array.isArray(data.paths) ? data.paths : [], methods: Array.isArray(data.methods) ? data.methods : [], ipv4_prefix: data.ipv4_prefix == null ? 32 : data.ipv4_prefix, ipv6_prefix: data.ipv6_prefix == null ? 128 : data.ipv6_prefix})}
    serialize={(data: any) => ({...data, methods: splitValues(data.methods).filter((value) => value !== "*"), paths: splitValues(data.paths), window: duration(data.window)})}>
    <SettingSwitch
        name="enabled"
        title="访问频率限制"
        description="默认按访客 IP 统计指定时间内的请求次数"
        styles={styles}/>
    <Form.Item noStyle shouldUpdate={(before, after) => before.enabled !== after.enabled}>{({getFieldValue}) => getFieldValue("enabled") ? <>
        <Form.Item name="window" label="统计周期（秒）"><InputNumber placeholder="请输入统计周期" min={1} style={{width: "100%"}}/></Form.Item>
        <Form.Item name="max_events" label="最多请求次数"><InputNumber placeholder="请输入最多请求次数" min={1} precision={0} style={{width: "100%"}}/></Form.Item>
        <Collapse ghost expandIconPlacement="end" items={[{key: "advanced", label: "高级设置", children: <>
            <Form.Item name="paths" label="限制路径"><Select mode="tags" tokenSeparators={[",", " "]} placeholder="留空表示整个网站"/></Form.Item>
            <Form.Item name="methods" label="请求方法"><Select mode="multiple" placeholder="留空表示全部方法" options={HTTP_METHOD_OPTIONS}/></Form.Item>
            <Form.Item name="key" label="统计对象" extra="可选择常用对象，也可以组合多个变量或直接输入固定前缀。">
                <AutoComplete
                    placeholder="请选择或输入统计对象"
                    options={RATE_KEY_OPTIONS.map((item) => ({value: item.value, label: `${item.label} · ${item.value}`}))}
                    filterOption={(input, option) => String(option?.label || "").toLowerCase().includes(input.toLowerCase())}/>
            </Form.Item>
            <div className={styles.variableGuide}>
                <div className="variable-guide-title">常用变量</div>
                {RATE_KEY_OPTIONS.map((item) => <div className="variable-guide-row" key={item.value}>
                    <code>{item.value}</code><span>{item.description}</span>
                </div>)}
                <div className="variable-guide-row"><code>{"{http.request.header.<名称>}"}</code><span>按指定请求头内容统计，例如 API Key</span></div>
                <div className="variable-guide-row"><code>{"{http.request.cookie.<名称>}"}</code><span>按指定 Cookie 内容统计</span></div>
            </div>
            <Form.Item name="ipv4_prefix" label="IPv4 前缀"><InputNumber placeholder="请输入 IPv4 前缀长度" min={0} max={32} precision={0} style={{width: "100%"}}/></Form.Item>
            <Form.Item name="ipv6_prefix" label="IPv6 前缀"><InputNumber placeholder="请输入 IPv6 前缀长度" min={0} max={128} precision={0} style={{width: "100%"}}/></Form.Item>
        </>}]}/>
    </> : null}</Form.Item>
</ConfigForm>;

const TrafficPane = ({site, styles, hideHeader = false}: any) => <ConfigForm site={site} request={siteTrafficLimit} update={siteTrafficLimit} title="并发与速度" description="限制站点并发量和单个请求的传输速度" styles={styles} hideHeader={hideHeader}
    normalize={(data: any) => {
        const empty = !data.enabled && !data.max_connections && !data.max_connections_per_ip && !data.rate_per_request;
        return {enabled: false, ...data, max_connections: empty ? 1000 : Number(data.max_connections || 0), max_connections_per_ip: empty ? 50 : Number(data.max_connections_per_ip || 0), rate_per_request_kb: Number(data.rate_per_request || 0) / 1024};
    }}
    serialize={(data: any) => ({enabled: Boolean(data.enabled), max_connections: Number(data.max_connections || 0), max_connections_per_ip: Number(data.max_connections_per_ip || 0), rate_per_request: Math.round(Number(data.rate_per_request_kb || 0) * 1024)})}>
    <SettingSwitch
        name="enabled"
        title="并发与速度限制"
        description="控制网站总并发、单个访客并发和响应传输速度"
        styles={styles}/>
    <Form.Item noStyle shouldUpdate={(before, after) => before.enabled !== after.enabled}>{({getFieldValue}) => getFieldValue("enabled") ? <>
        <Form.Item name="max_connections" label="网站最大并发" extra="0 表示不限"><InputNumber placeholder="请输入网站最大并发数" min={0} precision={0} style={{width: "100%"}}/></Form.Item>
        <Form.Item name="max_connections_per_ip" label="单 IP 最大并发" extra="0 表示不限"><InputNumber placeholder="请输入单 IP 最大并发数" min={0} precision={0} style={{width: "100%"}}/></Form.Item>
        <Form.Item name="rate_per_request_kb" label="单请求速度（KB/s）" extra="0 表示不限"><InputNumber placeholder="请输入单请求速度" min={0} precision={0} style={{width: "100%"}}/></Form.Item>
    </> : null}</Form.Item>
</ConfigForm>;

const CachePane = ({site, styles, hideHeader = false}: any) => {
    const {message, modal} = App.useApp();
    const clear = ({reload}: any) => <Button color="danger" variant="filled" style={{borderColor: "transparent"}} className={styles.settingAction} size="small" icon={<Icon type="ClearOutlined"/>} onClick={() => modal.confirm({title: "清理网站缓存", content: "确定清理这个网站的全部响应缓存吗？", okText: "清理", okButtonProps: {danger: true}, onOk: async () => {const response = await siteCache({body: {action: "clear", id: site.id}}); if (response?.code === 200) {message.success(response?.message || "清理成功"); reload();} else message.error(response?.message || "清理失败");}})}>清理缓存</Button>;
    return <ConfigForm site={site} request={siteCache} update={siteCache} title="响应缓存" description="缓存常用响应，减少磁盘或上游服务压力" styles={styles} hideHeader={hideHeader} extra={clear}
        normalize={(data: any) => ({...data, enabled: Boolean(data.enabled), ttl: seconds(data.ttl) || 120, stale: seconds(data.stale) || 30, methods: Array.isArray(data.methods) && data.methods.length ? data.methods : ["GET", "HEAD"], key_headers: Array.isArray(data.key_headers) ? data.key_headers : [], default_cache_control: data.default_cache_control || "public, max-age=120", max_size_mb: Number(data.max_size || 256 * 1024 * 1024) / 1024 / 1024})}
        serialize={(data: any) => ({enabled: Boolean(data.enabled), methods: splitValues(data.methods).filter((value) => value !== "*"), key_headers: splitValues(data.key_headers), ttl: duration(data.ttl), stale: duration(data.stale), max_size: Math.round(Number(data.max_size_mb || 0) * 1024 * 1024), default_cache_control: String(data.default_cache_control || "").trim()})}>
        <SettingSwitch
            name="enabled"
            title="响应缓存"
            description="缓存常用响应以降低磁盘或上游服务压力，动态网站请谨慎开启"
            styles={styles}/>
        <Form.Item noStyle shouldUpdate={(before, after) => before.enabled !== after.enabled}>{({getFieldValue}) => getFieldValue("enabled") ? <>
            <div className={styles.fieldGrid}>
                <Form.Item name="ttl" label="缓存时间（秒）"><InputNumber placeholder="请输入缓存时间" min={1}/></Form.Item>
                <Form.Item name="max_size_mb" label="容量上限（MB）"><InputNumber placeholder="请输入容量上限" min={1} precision={0}/></Form.Item>
            </div>
            <Collapse ghost expandIconPlacement="end" items={[{key: "advanced", label: "高级设置", children: <>
                <Form.Item name="stale" label="过期宽限（秒）"><InputNumber placeholder="请输入过期宽限时间" min={0} style={{width: "100%"}}/></Form.Item>
                <Form.Item name="methods" label="缓存请求方法"><Select mode="multiple" placeholder="请选择缓存请求方法" options={HTTP_METHOD_OPTIONS}/></Form.Item>
                <Form.Item name="key_headers" label="参与缓存键的请求头"><Select mode="tags" tokenSeparators={[",", " "]} placeholder="输入请求头名称后回车"/></Form.Item>
                <Form.Item name="default_cache_control" label="默认 Cache-Control"><Input placeholder="public, max-age=120"/></Form.Item>
            </>}]}/>
        </> : null}</Form.Item>
    </ConfigForm>;
};

const CompressionPane = ({site, styles, hideHeader = false}: any) => <ConfigForm site={site} request={siteCompression} update={siteCompression} title="响应压缩" description="自动压缩文本、脚本和样式文件" styles={styles} hideHeader={hideHeader}
    normalize={(data: any) => ({enabled: true, min_length: 1024, ...data, algorithms: Array.isArray(data.algorithms) && data.algorithms.length ? data.algorithms : ["zstd", "gzip"]})}>
    <SettingSwitch
        name="enabled"
        title="响应压缩"
        description="压缩文本、脚本和样式文件，推荐保持开启"
        styles={styles}/>
    <Form.Item noStyle shouldUpdate={(before, after) => before.enabled !== after.enabled}>{({getFieldValue}) => getFieldValue("enabled") ? <Collapse ghost expandIconPlacement="end" items={[{
        key: "advanced", label: "高级设置", children: <div className={styles.fieldGrid}>
            <Form.Item name="algorithms" label="压缩算法"><Select mode="multiple" placeholder="请选择压缩算法" options={[{label: "Zstandard", value: "zstd"}, {label: "Gzip", value: "gzip"}]}/></Form.Item>
            <Form.Item name="min_length" label="最小响应（字节）"><InputNumber placeholder="请输入最小响应字节数" min={0} precision={0}/></Form.Item>
        </div>,
    }]}/> : null}</Form.Item>
</ConfigForm>;

const HeaderPane = ({site, styles, hideHeader = false}: any) => {
    const [active, setActive] = useState<string | number>("request");
    const normalize = useCallback((data: any) => ({
        request_rules: headersToRows(data.request_add, data.request_set, data.request_delete),
        response_rules: headersToRows(data.response_add, data.response_set, data.response_delete),
    }), []);
    const serialize = useCallback((data: any) => {
        const request = rowsToHeaders(data.request_rules);
        const response = rowsToHeaders(data.response_rules);
        return {
            request_add: request.add,
            request_set: request.set,
            request_delete: request.delete,
            response_add: response.add,
            response_set: response.set,
            response_delete: response.delete,
        };
    }, []);
    return <ConfigForm
        site={site}
        request={siteHeaders}
        update={siteHeaders}
        title="请求与响应头"
        description="按规则追加、覆盖或删除请求头和响应头"
        styles={styles}
        hideHeader={hideHeader}
        normalize={normalize}
        serialize={serialize}
        onFinishFailed={({errorFields}: any) => {
            const listName = errorFields?.[0]?.name?.[0];
            if (listName === "request_rules" || listName === "response_rules") {
                setActive(listName === "request_rules" ? "request" : "response");
            }
        }}>
        <div className={styles.headerPane}>
            <Segmented
                className="header-switcher"
                block
                value={active}
                onChange={setActive}
                options={[{label: "请求头", value: "request"}, {label: "响应头", value: "response"}]}/>
            <div hidden={active !== "request"}><HeaderRules name="request_rules" label="请求头" styles={styles}/></div>
            <div hidden={active !== "response"}><HeaderRules name="response_rules" label="响应头" styles={styles}/></div>
        </div>
    </ConfigForm>;
};

const RedirectRule = ({field, index, count, move, remove, domainOptions, errors, styles}: any) => {
    const form = Form.useFormInstance();
    const path = ["redirects", field.name];
    const rule = Form.useWatch(path, form) || form.getFieldValue(path) || {};
    const [expanded, setExpanded] = useState(index === 0 || !rule.target);
    const headingRef = useRef<HTMLDivElement>(null);
    const domains = splitValues(rule.domains);
    const customPaths = rule.path_scope === "custom";
    const paths = customPaths ? splitValues(rule.paths) : [];
    const sourcePath = customPaths ? String(paths[0] || "/old/*") : "/news/123";
    const samplePath = sourcePath.replace(/\*/g, "example");
    const sourceDomain = String(domains[0] || domainOptions[0]?.value || "example.com").replace(/^\*\./, "www.");
    const target = String(rule.target || "").trim();
    const source = `${sourceDomain}${samplePath}?from=site`;
    const destination = target
        ? (rule.preserve_uri ? `${target.replace(/\/+$/, "")}${samplePath}?from=site` : target)
        : "填写目标地址后显示";
    const statusText = [301, 308].includes(Number(rule.status)) ? "永久跳转" : "临时跳转";
    const scopeText = `${domains.length ? domains.join("、") : "全部域名"} · ${customPaths ? paths.join("、") || "待选择路径" : "所有路径"}`;

    useEffect(() => {
        if (index > 0 && !rule.target) headingRef.current?.scrollIntoView({block: "start", inline: "nearest", behavior: "smooth"});
    }, []);

    useEffect(() => {
        if (errors.some((item: any) => item.name?.[1] === field.name)) {
            setExpanded(true);
            if (errors[0]?.name?.[1] === field.name) headingRef.current?.scrollIntoView({block: "start", inline: "nearest"});
        }
    }, [errors, field.name]);

    return <div className="redirect-card">
        <div className="redirect-heading" ref={headingRef}>
            <button type="button" className="redirect-summary" aria-expanded={expanded} aria-label={`${expanded ? "收起" : "展开"}重定向规则 ${index + 1}`} onClick={() => setExpanded(!expanded)}>
                <Icon type={expanded ? "DownOutlined" : "RightOutlined"}/>
                <span className="redirect-summary-text">
                    <span className="redirect-title">
                        <span>{rule.name || `重定向 ${index + 1}`}</span>
                        <span className="redirect-status">{rule.enabled === false ? "已停用" : statusText}</span>
                    </span>
                    <span className="redirect-destination" title={`${scopeText} → ${target || "待填写目标地址"}`}>
                        {scopeText} → {target || "待填写目标地址"}
                    </span>
                </span>
            </button>
            <div className="redirect-controls">
                <Form.Item name={[field.name, "enabled"]} valuePropName="checked" noStyle><Switch size="small" aria-label={`启用重定向规则 ${index + 1}`}/></Form.Item>
                <ActionDropdown menu={{
                    items: [
                        {key: "up", label: "上移规则", disabled: index === 0},
                        {key: "down", label: "下移规则", disabled: index === count - 1},
                        {type: "divider"},
                        {key: "delete", label: "删除规则", danger: true},
                    ],
                    onClick: ({key}) => {
                        if (key === "up") move(field.name, field.name - 1);
                        if (key === "down") move(field.name, field.name + 1);
                        if (key === "delete") remove(field.name);
                    },
                }}>
                    <Button color="default" variant="text" className={styles.ruleAction} size="small" aria-label={`重定向规则 ${index + 1} 操作`} icon={<Icon type="MoreOutlined"/>}/>
                </ActionDropdown>
            </div>
        </div>
        <div className="redirect-body" hidden={!expanded}>
            <Form.Item name={[field.name, "name"]} label="规则名称" rules={[
                {required: true, whitespace: true, message: "请输入规则名称"},
                ({getFieldValue}) => ({validator: async (_, value) => {
                    const name = String(value || "").trim().toLowerCase();
                    if (name && (getFieldValue("redirects") || []).some((item: any, other: number) => other !== field.name && String(item?.name || "").trim().toLowerCase() === name)) {
                        throw new Error("规则名称已存在，请换一个名称");
                    }
                }}),
            ]}>
                <Input placeholder="例如：旧域名跳转，方便以后辨认" maxLength={100}/>
            </Form.Item>
            <Form.Item name={[field.name, "path_scope"]} label="哪些访问需要跳转">
                <Segmented className="redirect-scope" options={[{label: "整个网站", value: "all"}, {label: "指定路径", value: "custom"}]}/>
            </Form.Item>
            {customPaths && <Form.Item
                name={[field.name, "paths"]}
                label="访问路径"
                extra="例如 /old 只匹配该地址；/old/* 匹配其下的页面。可添加多个，按回车确认。"
                rules={[{validator: async (_, value) => {
                    const paths = splitValues(value);
                    if (!paths.length) throw new Error("请至少填写一个访问路径");
                    if (paths.some((item) => !item.startsWith("/") || /[\r\n]/.test(item))) {
                        throw new Error("访问路径必须以 / 开头且不能包含换行");
                    }
                }}]}>
                <Select mode="tags" tokenSeparators={[",", " "]} placeholder="例如 /old 或 /old/*"/>
            </Form.Item>}
            <Form.Item name={[field.name, "domains"]} label="生效域名" extra="不选择则对本站全部域名生效，也会包含以后新增的域名。">
                <Select mode="multiple" showSearch={false} placeholder="本站全部域名" options={domainOptions}/>
            </Form.Item>
            <Form.Item
                name={[field.name, "target"]}
                label="跳转到哪里"
                className="setting-wide"
                extra="填写完整网址，或用 / 开头跳转到本站页面。"
                dependencies={[["redirects", field.name, "preserve_uri"]]}
                rules={[
                    {required: true, whitespace: true, message: "请填写跳转后的目标地址"},
                    ({getFieldValue}) => ({
                        validator: async (_, value) => {
                            const target = String(value || "").trim();
                            if (!target) return;
                            if (/[\r\n{}]/.test(target)) throw new Error("目标地址包含无效字符");
                            const preserveURI = Boolean(getFieldValue(["redirects", field.name, "preserve_uri"]));
                            if (/^\/(?!\/)/.test(target)) {
                                if (target.includes("\\")) throw new Error("站内路径不能包含反斜杠");
                                if (preserveURI && target === "/") throw new Error("跳转到首页时，请选择「只跳转到目标」");
                                if (preserveURI && /[?#]/.test(target)) throw new Error("目标已包含参数，请选择「只跳转到目标」");
                                return;
                            }
                            if (!/^https?:\/\/[^/?#\\\s]/i.test(target)) {
                                throw new Error("请输入完整网址（含 http:// 或 https://）或 / 开头的站内路径");
                            }
                            let url: URL;
                            try { url = new URL(target); } catch {
                                throw new Error("请输入完整网址（含 http:// 或 https://）或 / 开头的站内路径");
                            }
                            if (!["http:", "https:"].includes(url.protocol) || url.username || url.password) {
                                throw new Error("请输入 HTTP、HTTPS 网址或站内路径");
                            }
                            if (preserveURI && (url.search || url.hash)) {
                                throw new Error("目标已包含参数或锚点，请选择「只跳转到目标」");
                            }
                        },
                    }),
                ]}>
                <Input placeholder="https://new.example.com 或 /new-page"/>
            </Form.Item>
            <Form.Item
                name={[field.name, "preserve_uri"]}
                label="跳转后的路径"
                getValueProps={(value) => ({value: value ? "preserve" : "target"})}
                getValueFromEvent={(value) => value === "preserve"}>
                <Select placeholder="请选择路径处理方式" options={[
                    {label: "保留原路径和参数", value: "preserve"},
                    {label: "只跳转到目标", value: "target"},
                ]}/>
            </Form.Item>
            <div className="redirect-example" aria-live="polite">
                <div className="redirect-example-title">跳转示例</div>
                <div className="redirect-example-row"><span>访问</span><code>{source}</code></div>
                <div className="redirect-example-row"><span>跳转到</span><code>{destination}</code></div>
                <div className="redirect-example-note">{rule.preserve_uri
                    ? "在目标地址后追加完整原路径和参数，不会移除匹配的路径前缀。"
                    : "直接使用目标地址，不附加原路径和参数。"}</div>
            </div>
            <Collapse ghost expandIconPlacement="end" items={[{
                key: "advanced",
                label: `更多设置 · ${statusText}（${rule.status || 302}）`,
                forceRender: true,
                children: <>
                    <Form.Item name={[field.name, "status"]} label="跳转方式" extra={statusText === "永久跳转" ? "适合域名迁移、旧链接替换，浏览器可能记住跳转结果。" : "适合临时调整或测试，以后仍可改回原地址。"}>
                        <Select placeholder="请选择跳转方式" options={[
                            {label: "临时跳转（302）", value: 302},
                            {label: "永久跳转（301）", value: 301},
                            {label: "临时跳转，保留请求方法（307）", value: 307},
                            {label: "永久跳转，保留请求方法（308）", value: 308},
                        ]}/>
                    </Form.Item>
                </>,
            }]}/>
        </div>
    </div>;
};

const RedirectRules = ({fields, add, remove, move, domainOptions, errors, onChange, styles}: any) => {
    const form = Form.useFormInstance();
    const addRule = () => {
        const names = new Set((form.getFieldValue("redirects") || []).map((item: any) => String(item?.name || "").trim().toLowerCase()));
        let number = 1;
        while (names.has(`重定向 ${number}`)) number += 1;
        onChange();
        add({enabled: true, name: `重定向 ${number}`, domains: [], path_scope: "all", paths: [], target: "", status: 302, preserve_uri: true});
    };
    const addButton = <Button color="default" variant="filled" style={{borderColor: "transparent"}} className={styles.settingAction} size="small" icon={<Icon type="PlusOutlined"/>} onClick={addRule}>添加规则</Button>;
    return <div className={styles.redirectRules}>
        {fields.length ? <>
            <div className="redirect-toolbar"><span>从上到下匹配，命中第一条后跳转</span>{addButton}</div>
            {fields.map((field: any, index: number) => <RedirectRule
                key={field.key}
                field={field}
                index={index}
                count={fields.length}
                move={(from: number, to: number) => {onChange(); move(from, to);}}
                remove={(name: number) => {onChange(); remove(name);}}
                domainOptions={domainOptions}
                errors={errors}
                styles={styles}/>)}
        </> : <div className="redirect-empty">
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
            {addButton}
        </div>}
    </div>;
};

const RedirectPane = ({site, styles}: any) => {
    const {message} = App.useApp();
    const [domainOptions, setDomainOptions] = useState<any[]>([]);
    const [errors, setErrors] = useState<any[]>([]);
    const normalize = useCallback((data: any) => ({redirects: (Array.isArray(data) ? data : []).map((item: any) => ({
        ...item,
        enabled: item.enabled !== false,
        domains: Array.isArray(item.domains) ? item.domains : [],
        path_scope: !Array.isArray(item.paths) || !item.paths.length || item.paths.includes("/*") ? "all" : "custom",
        paths: Array.isArray(item.paths) ? item.paths : [],
        status: [301, 302, 307, 308].includes(Number(item.status)) ? Number(item.status) : 302,
        preserve_uri: item.preserve_uri !== false,
    }))}), []);
    const serialize = useCallback((data: any) => (Array.isArray(data.redirects) ? data.redirects : []).map((item: any) => ({
        enabled: Boolean(item.enabled),
        name: String(item.name || "").trim(),
        domains: splitValues(item.domains),
        paths: item.path_scope === "all" ? ["/*"] : splitValues(item.paths),
        target: String(item.target || "").trim(),
        status: Number(item.status || 302),
        preserve_uri: Boolean(item.preserve_uri),
    })), []);
    useEffect(() => {
        let active = true;
        setDomainOptions([]);
        void siteDomain({body: {action: "get", id: site.id}}).then((response) => {
            if (!active) return;
            if (response?.code === 200) {
                setDomainOptions((Array.isArray(response?.data) ? response?.data.filter(Boolean) : []).map((item: any) => ({
                    label: item.domain,
                    value: item.domain,
                })));
            } else {
                message.error(response?.message || "获取网站域名失败");
            }
        }).catch(() => {
            if (active) message.error("获取网站域名失败");
        });
        return () => {active = false;};
    }, [message, site.id]);

    return <ConfigForm
        site={site}
        request={siteRedirects}
        update={siteRedirects}
        title="重定向"
        description="将访客从旧域名或旧页面，自动引导到新的地址"
        styles={styles}
        normalize={normalize}
        serialize={serialize}
        onFinishFailed={({errorFields}: any) => setErrors(errorFields)}>
        <Form.List name="redirects">{(fields, {add, remove, move}) => <RedirectRules
            fields={fields}
            add={add}
            remove={remove}
            move={move}
            domainOptions={domainOptions}
            errors={errors}
            onChange={() => setErrors([])}
            styles={styles}/>}</Form.List>
    </ConfigForm>;
};

const RouteRule = ({field, index, count, move, remove, copy, errors, styles}: any) => {
    const form = Form.useFormInstance();
    const path = ["routes", field.name];
    const rule = Form.useWatch(path, {form, preserve: true}) || form.getFieldValue(path) || {};
    const [expanded, setExpanded] = useState(index === 0 || Boolean(rule._new));
    const [advanced, setAdvanced] = useState<string[]>([]);
    const headingRef = useRef<HTMLDivElement>(null);
    const kind = rule.kind || "proxy";
    const kindLabel = kind === "proxy" ? "反向代理" : kind === "response" ? "固定响应" : "静态文件";
    const paths = splitValues(rule.paths);
    const destination = kind === "proxy"
        ? splitValues(rule.proxy_upstreams).join("、") || "待填写目标服务"
        : kind === "response" ? `返回 ${rule.response_status || 200} · ${rule.response_body || "空白内容"}` : rule.root || "待选择文件目录";
    const summary = `${paths.length ? paths.join("、") : "全部路径"} → ${destination}`;

    useEffect(() => {
        if (rule._new) headingRef.current?.scrollIntoView({block: "nearest", inline: "nearest"});
    }, []);

    useEffect(() => {
        const ownErrors = errors.filter((item: any) => item.name?.[1] === field.name);
        if (!ownErrors.length) return;
        setExpanded(true);
        if (ownErrors.some((item: any) => ["methods", "strip_prefix", "rewrite", "index", "try_files"].includes(item.name?.[2]))) setAdvanced(["advanced"]);
        if (errors[0]?.name?.[1] === field.name) headingRef.current?.scrollIntoView({block: "nearest", inline: "nearest"});
    }, [errors, field.name]);

    return <div className="route-card">
        <div className="route-heading" ref={headingRef}>
            <button className="route-summary" type="button" aria-expanded={expanded} aria-label={`${expanded ? "收起" : "展开"}路由 ${index + 1}`} onClick={() => setExpanded(!expanded)}>
                <Icon type={expanded ? "DownOutlined" : "RightOutlined"}/>
                <span className="route-summary-text">
                    <span className="route-title"><span>{rule.name || `路由 ${index + 1}`}</span><span className="route-kind">{kindLabel}</span></span>
                    <span className="route-destination" title={summary}>{summary}</span>
                </span>
            </button>
            <ActionDropdown menu={{
                items: [
                    {key: "up", label: "上移规则", disabled: index === 0},
                    {key: "down", label: "下移规则", disabled: index === count - 1},
                    {key: "copy", label: "复制规则"},
                    {type: "divider"},
                    {key: "delete", label: "删除规则", danger: true},
                ],
                onClick: ({key}) => {
                    if (key === "up") move(field.name, field.name - 1);
                    if (key === "down") move(field.name, field.name + 1);
                    if (key === "copy") copy(rule);
                    if (key === "delete") remove(field.name);
                },
            }}>
                <Button color="default" variant="text" className={styles.ruleAction} size="small" aria-label={`路由 ${index + 1} 操作`} icon={<Icon type="MoreOutlined"/>}/>
            </ActionDropdown>
        </div>
        <div className="route-body" hidden={!expanded}>
            <Form.Item name={[field.name, "name"]} label="规则名称" rules={[{required: true, whitespace: true, message: "请输入规则名称"}]}>
                <Input placeholder="例如：接口转发、下载目录" maxLength={100}/>
            </Form.Item>
            <Form.Item name={[field.name, "paths"]} label="哪些访问需要处理" extra="例如 /api 只匹配该地址；/api/* 匹配其下的页面。留空表示全部路径。" dependencies={[["routes", field.name, "methods"]]} rules={[
                ({getFieldValue}) => ({validator: async (_, value) => {
                    const paths = splitValues(value);
                    if (paths.some((item) => !item.startsWith("/") || /[\r\n]/.test(item))) throw new Error("访问路径必须以 / 开头且不能包含换行");
                    if (!paths.length && !splitValues(getFieldValue(["routes", field.name, "methods"])).length && index < count - 1) throw new Error("匹配全部访问的规则请移到最后");
                }}),
            ]}>
                <Select mode="tags" tokenSeparators={[",", " "]} placeholder="例如 /api/*，按回车添加"/>
            </Form.Item>
            <Form.Item name={[field.name, "kind"]} label="如何处理访问">
                <Select showSearch={false} placeholder="请选择处理方式" options={[
                    {label: "转发到其他服务", value: "proxy"},
                    {label: "提供目录中的文件", value: "static"},
                    {label: "返回固定内容", value: "response"},
                ]}/>
            </Form.Item>
            {kind === "proxy" ? <>
                <Form.Item name={[field.name, "proxy_upstreams"]} label="目标服务" extra={rule.proxy?.transport === "fastcgi" ? "填写 PHP-FPM 地址或套接字路径，保留已有 FastCGI 处理配置。" : "自动识别 HTTP、HTTPS 和套接字。可填写多个地址，按回车添加，所有目标应使用相同协议。"} rules={[{validator: async (_, value) => {
                    const values = splitValues(value);
                    if (!values.length) throw new Error("请至少填写一个目标服务地址");
                    if (rule.proxy?.transport === "fastcgi") {
                        if (values.some((address) => /[\r\n\0]/.test(address) || address.includes("://"))) throw new Error("FastCGI 地址不能带 HTTP 或 HTTPS 协议");
                        return;
                    }
                    const first = parseProxyTarget(values[0]);
                    const targets = values.map((address) => parseProxyTarget(address, first.scheme));
                    if (targets.some((target) => target.scheme !== targets[0].scheme)) throw new Error("所有目标服务需要使用相同协议");
                }}]}>
                    <Select mode="tags" tokenSeparators={[",", " "]} placeholder={rule.proxy?.transport === "fastcgi" ? "例如 127.0.0.1:9000 或 /tmp/php.sock" : "例如 https://example.com 或 /tmp/app.sock"}/>
                </Form.Item>
            </> : kind === "response" ? <>
                <Form.Item className="setting-wide" name={[field.name, "response_body"]} label="返回内容" extra="可填写文本或 HTML，留空则只返回状态码。">
                    <Input.TextArea placeholder="例如：服务维护中，请稍后再试" autoSize={{minRows: 3, maxRows: 8}}/>
                </Form.Item>
                <Form.Item name={[field.name, "response_status"]} label="响应状态码" rules={[{type: "number", min: 100, max: 599, message: "请输入 100 至 599 的状态码"}]}>
                    <InputNumber placeholder="默认 200，可填写 100 至 599" min={100} max={599} precision={0} style={{width: "100%"}}/>
                </Form.Item>
            </> : <>
                <Form.Item className="site-file-picker" name={[field.name, "root"]} label="文件所在目录" rules={[{required: true, whitespace: true, message: "请选择文件所在目录"}]}>
                    <FilePicker buttonProps={{color: "default", variant: "filled", style: {borderColor: "transparent"}}} mode="directory" placeholder="请输入或选择文件目录"/>
                </Form.Item>
                <SettingSwitch name={[field.name, "browse"]} title="允许浏览目录" description="直接显示文件列表，不再自动打开首页文件" styles={styles}/>
            </>}
            <Collapse ghost activeKey={advanced} onChange={(keys) => setAdvanced(Array.isArray(keys) ? keys : [keys])} expandIconPlacement="end" items={[{
                key: "advanced",
                label: "更多设置",
                forceRender: true,
                children: <>
                    <Form.Item name={[field.name, "methods"]} label="请求方法" extra="通常无需选择，留空即可处理所有请求方法。">
                        <Select mode="multiple" showSearch={false} placeholder="全部请求方法" options={HTTP_METHOD_OPTIONS}/>
                    </Form.Item>
                    <Form.Item name={[field.name, "strip_prefix"]} label="移除路径前缀" extra="例如填写 /api，访问 /api/users 时会按 /users 处理。" rules={[{validator: async (_, value) => {
                        if (value && !String(value).trim().startsWith("/")) throw new Error("路径前缀必须以 / 开头");
                    }}]}>
                        <Input placeholder="可选，例如 /api"/>
                    </Form.Item>
                    <Form.Item name={[field.name, "rewrite"]} label="改写访问路径" extra="在移除前缀后执行。只改变内部处理路径，不会让浏览器跳转。">
                        <Input placeholder="可选，例如 /index.html 或 /v1{http.request.uri}"/>
                    </Form.Item>
                    {kind === "static" && <>
                        <Form.Item name={[field.name, "index"]} label="默认首页" extra="按填写顺序查找首页文件，可留空。">
                            <Select mode="tags" tokenSeparators={[",", " "]} placeholder="例如 index.html、index.htm"/>
                        </Form.Item>
                        <Form.Item name={[field.name, "try_files"]} label="文件查找顺序" extra="例如先查找请求文件，找不到时返回 /index.html，适合前端单页应用。">
                            <Select mode="tags" placeholder="例如 {http.request.uri.path}、/index.html"/>
                        </Form.Item>
                    </>}
                </>,
            }]}/>
        </div>
    </div>;
};

const RouteRules = forwardRef(({fields, add, remove, move, styles}: any, ref: any) => {
    const form = Form.useFormInstance();
    const [errors, setErrors] = useState<any[]>([]);
    useImperativeHandle(ref, () => ({reveal: setErrors, getValues: () => form.getFieldValue("routes")}), [form]);
    const addRule = (source?: any) => {
        const names = new Set((form.getFieldValue("routes") || []).map((item: any) => String(item?.name || "").trim()));
        const base = source ? `${source.name || "路由"} 副本` : "路由";
        let number = 1;
        let name = source ? base : `${base} ${number}`;
        while (names.has(name)) name = `${base} ${++number}`;
        setErrors([]);
        add(source ? {...JSON.parse(JSON.stringify(source)), name, _new: true} : {name, _new: true, methods: [], paths: ["/api/*"], kind: "proxy", response_status: 200});
    };
    const addButton = <Button color="default" variant="filled" style={{borderColor: "transparent"}} className={styles.settingAction} size="small" icon={<Icon type="PlusOutlined"/>} onClick={() => addRule()}>添加规则</Button>;
    return <div className={styles.routeRules}>
        {fields.length ? <>
            <div className="route-toolbar"><span>从上到下匹配，命中后按对应方式处理</span>{addButton}</div>
            {fields.map((field: any, index: number) => <RouteRule
                key={field.key}
                field={field}
                index={index}
                count={fields.length}
                errors={errors}
                move={(from: number, to: number) => {setErrors([]); move(from, to);}}
                remove={(name: number) => {setErrors([]); remove(name);}}
                copy={addRule}
                styles={styles}/>)}
        </> : <div className="route-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>{addButton}</div>}
    </div>;
});

const RoutePane = ({site, styles, hideHeader = false}: any) => {
    const rulesRef = useRef<any>(null);
    const normalize = useCallback((data: any) => ({
        routes: (Array.isArray(data) ? data : []).map((item) => ({
            ...item,
            kind: item.proxy ? "proxy" : item.response ? "response" : "static",
            proxy_upstreams: upstreamsToForm(item.proxy?.upstreams).map((address: any) => {
                const dial = String(address);
                if (item.proxy?.transport === "fastcgi") return dial;
                if (/^unix(?:\+h2c)?\//.test(dial)) return item.proxy?.versions?.includes("h2c") ? dial.replace(/^unix(?:\+h2c)?\//, "unix+h2c/") : dial.replace(/^unix\//, "");
                return /^https?:\/\//i.test(dial) || dial.startsWith("/") ? dial : `${item.proxy?.scheme || "http"}://${dial}`;
            }),
            response_status: item.response?.status || 200,
            response_body: item.response?.body || "",
        })),
    }), []);
    const serialize = useCallback((data: any) => (rulesRef.current?.getValues() || data.routes || []).map((item: any) => {
        const {kind, _new, proxy_upstreams, response_status, response_body, ...stored} = item;
        const result: any = {
            ...stored,
            name: String(item.name || "").trim(),
            paths: splitValues(item.paths),
            methods: splitValues(item.methods).filter((value) => value !== "*"),
            strip_prefix: String(item.strip_prefix || "").trim(),
            rewrite: item.rewrite || "",
            root: "", index: [], browse: false, try_files: [], hide: [], precompressed: [],
            proxy: null, response: null,
        };
        if (kind === "proxy") {
            const original = stored.proxy || {};
            const fastCGI = original.transport === "fastcgi";
            const addresses = splitValues(proxy_upstreams);
            const scheme = fastCGI ? original.scheme : parseProxyTarget(addresses[0]).scheme;
            const upstreams = addresses.map((address) => fastCGI
                ? {dial: address.startsWith("/") ? `unix/${address}` : address, scheme: original.scheme || "http", h2c: false}
                : parseProxyTarget(address, scheme));
            result.proxy = {
                ...original,
                upstreams: upstreams.map(({dial}) => {
                    const existing = (original.upstreams || []).find((upstream: any) => {
                        if (fastCGI) return upstream.dial === dial || upstream.dial === dial.replace(/^unix\//, "");
                        try { return parseProxyTarget(upstream.dial, original.scheme || "http").dial === dial; } catch { return upstream.dial === dial; }
                    });
                    return {max_requests: 0, ...existing, dial: existing?.dial || dial};
                }),
                transport: original.transport || "http",
                scheme,
                policy: original.policy || "random",
                ...(!fastCGI ? {
                    versions: upstreams.some((upstream) => upstream.h2c) ? ["h2c"] : (original.versions || []).filter((version: string) =>
                        (version !== "h2c" || scheme === "http" && !upstreams[0]?.dial.startsWith("unix/")) && (version !== "3" || scheme === "https")),
                    tls_server_name: scheme === "https" ? original.tls_server_name || "" : "",
                    tls_insecure_skip_verify: scheme === "https" ? original.tls_insecure_skip_verify : false,
                } : {}),
            };
        } else if (kind === "response") {
            result.response = {...stored.response, status: Number(response_status || 200), body: response_body || "", headers: stored.response?.headers || {}};
        } else {
            result.root = String(item.root || "").trim();
            result.index = splitValues(item.index);
            result.browse = Boolean(item.browse);
            result.try_files = splitValues(item.try_files);
            result.hide = splitValues(item.hide);
            result.precompressed = splitValues(item.precompressed);
        }
        return result;
    }), []);
    return <ConfigForm
        site={site}
        request={siteRoutes}
        update={siteRoutes}
        title="自定义路由"
        description="让不同访问路径转发到服务、读取文件或返回指定内容"
        styles={styles}
        hideHeader={hideHeader}
        normalize={normalize}
        serialize={serialize}
        onFinishFailed={({errorFields}: any) => rulesRef.current?.reveal(errorFields)}>
        <Form.List name="routes">{(fields, {add, remove, move}) => <RouteRules ref={rulesRef} fields={fields} add={add} remove={remove} move={move} styles={styles}/>}</Form.List>
    </ConfigForm>;
};

const LogPane = ({site, styles}: any) => {
    const logRef = useRef<any>({});
    const logs = [
        {type: "access", name: "访问日志", icon: "FileTextOutlined", hint: "访问请求、响应状态与耗时"},
        {type: "waf", name: "防火墙日志", icon: "SafetyCertificateOutlined", hint: "规则命中与拦截结果"},
    ];
    return <>
        <SectionHeader title="网站日志" description="选择日志查看内容，可在日志窗口中刷新或清理" styles={styles}/>
        <div className={styles.logGrid}>
            {logs.map((item) => <div className={styles.logItem} key={item.type}>
                <span className="log-icon" aria-hidden="true"><Icon type={item.icon}/></span>
                <div className="log-main">
                    <div className="log-name">{item.name}</div>
                    <div className="log-description">{item.hint}</div>
                </div>
                <Button
                    color="default"
                    variant="filled"
                    style={{borderColor: "transparent"}}
                    className={`${styles.settingAction} log-open`}
                    size="small"
                    aria-label={`查看${item.name}`}
                    onClick={() => logRef.current?.open(site, {type: item.type})}>
                    查看
                </Button>
            </div>)}
        </div>
        <Log ref={logRef} request={getSiteLog} title="网站日志"/>
    </>;
};

const Settings = forwardRef(({onSuccess, onManageCert, onObtainCert, onDeployCert, onRenewCert, certVersion = 0, typeItems}: any, ref: any): any => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const settingsSessionRef = useRef(0);
    const navigationRef = useRef<HTMLDivElement>(null);
    const contentRef = useRef<HTMLElement>(null);
    const footerRef = useRef<any>({});
    const [site, setSite] = useState<any>();
    const [active, setActive] = useState("basic");
    const formId = `site-settings-${String(site?.id || "none")}-${active}`;
    const formIdRef = useRef(formId);
    formIdRef.current = formId;
    const updateFooter = useCallback((source: string, state?: any) => {
        if (source !== formIdRef.current) return;
        footerRef.current?.update?.(state);
    }, []);
    const formContext = useMemo(() => ({formId, update: updateFooter}), [formId, updateFooter]);

    const open = useCallback((record: any, tab = "basic") => {
        settingsSessionRef.current += 1;
        const aliases: any = {common: "basic", safety: "waf", performance: "cache", advanced: "header"};
        const next = aliases[tab] || tab;
        const nextActive = [
            "basic", "domain", "ssl", "runtime", "redirect", "waf", "rate", "traffic",
            "cache", "compression", "log", "header", "route",
        ].includes(next) ? next : "basic";
        setSite(record);
        setActive(nextActive);
        modalRef.current?.show();
    }, []);
    useImperativeHandle(ref, () => ({open}), [open]);

    const sections = useMemo(() => {
        if (!site) return [];
        const runtimeLabels: any = {0: "静态设置", 1: "反向代理", 2: "项目设置", 3: "PHP 设置"};
        return [
            {key: "basic", label: "基础信息", color: "#1677ff"},
            {key: "domain", label: "域名管理", color: "#13a8a8"},
            {key: "ssl", label: "HTTPS", color: "#52a832"},
            {key: "runtime", label: runtimeLabels[Number(site.type)] || "运行设置", color: "#722ed1"},
            {key: "redirect", label: "重定向", color: "#389e0d"},
            {key: "waf", label: "Web 防火墙", color: "#d4380d"},
            {key: "rate", label: "访问频率", color: "#d46b08"},
            {key: "traffic", label: "并发限速", color: "#d4a106"},
            {key: "cache", label: "响应缓存", color: "#2f54c9"},
            {key: "compression", label: "响应压缩", color: "#08979c"},
            {key: "log", label: "网站日志", color: "#597ef7"},
            {key: "header", label: "请求与响应头", color: "#c41d7f"},
            {key: "route", label: "自定义路由", color: "#531dab"},
        ];
    }, [site]);

    useEffect(() => {
        const element = navigationRef.current;
        if (!element) return;
        let frame = 0;
        const revealActive = () => {
            cancelAnimationFrame(frame);
            frame = requestAnimationFrame(() => {
                const item = element.querySelector(`[data-section="${active}"]`);
                item?.scrollIntoView({behavior: "auto", block: "nearest", inline: "nearest"});
            });
        };
        const handleWheel = (event: WheelEvent) => {
            const maxScrollLeft = Math.max(0, element.scrollWidth - element.clientWidth);
            if (maxScrollLeft <= 0) return;
            const delta = Math.abs(event.deltaX) > Math.abs(event.deltaY) ? event.deltaX : event.deltaY;
            const canScroll = delta < 0 ? element.scrollLeft > 0 : element.scrollLeft < maxScrollLeft;
            if (!canScroll) return;
            event.preventDefault();
            element.scrollLeft += delta;
        };
        const observer = new ResizeObserver(revealActive);
        element.addEventListener("wheel", handleWheel, {passive: false});
        observer.observe(element);
        revealActive();
        return () => {
            cancelAnimationFrame(frame);
            observer.disconnect();
            element.removeEventListener("wheel", handleWheel);
        };
    }, [active, site]);

    useEffect(() => {
        contentRef.current?.scrollTo({top: 0, behavior: "auto"});
    }, [active, site?.id]);

    const content = useMemo(() => {
        if (!site) return null;
        const common = {site, styles, typeItems};
        switch (active) {
        case "basic": return <BasicPane {...common} onChange={(next: any) => {
                if (next) setSite(next);
                onSuccess?.();
            }}/>;
        case "domain": return <DomainPane {...common} onChange={onSuccess}/>;
        case "ssl": return <SSLPane key={site.id} {...common} settingsSessionRef={settingsSessionRef} certVersion={certVersion} onManageCert={onManageCert} onObtainCert={onObtainCert} onDeployCert={onDeployCert} onRenewCert={onRenewCert}/>;
        case "redirect": return <RedirectPane {...common}/>;
        case "runtime":
            if (Number(site.type) === 0) return <StaticPane {...common}/>;
            if (Number(site.type) === 3) return <FastCGIPane {...common}/>;
            if (Number(site.type) === 2) return <GeneralPane {...common}/>;
            return <ProxyPane {...common}/>;
        case "waf": return <WAFPane {...common}/>;
        case "rate": return <RatePane {...common}/>;
        case "traffic": return <TrafficPane {...common}/>;
        case "cache": return <CachePane {...common}/>;
        case "compression": return <CompressionPane {...common}/>;
        case "log": return <LogPane {...common}/>;
        case "header": return <HeaderPane {...common}/>;
        case "route": return <RoutePane {...common}/>;
        default: return null;
        }
    }, [active, certVersion, onManageCert, onObtainCert, onDeployCert, onRenewCert, onSuccess, site, styles, typeItems]);

    return (
        <ProModal
            ref={modalRef}
            title={<Title>网站设置</Title>}
            width={compact ? 840 : 800}
            afterOpenChange={(isOpen) => {if (!isOpen) settingsSessionRef.current += 1;}}
            modalProps={{
                rootClassName: styles.modal,
                footer: active === "log" ? null : (
                    <SettingsFooter
                        key={formId}
                        ref={footerRef}
                        formId={formId}
                        onClose={() => modalRef.current?.hide()}/>
                ),
                forceRender: true,
                closable: true,
                keyboard: true,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                mask: {closable: true},
                afterClose: () => {
                    setSite(undefined);
                    setActive("basic");
                },
            } as any}>
            <SettingsFormContext.Provider value={formContext}>
                <div className={styles.layout}>
                    <aside className={styles.navigation} aria-label="网站设置分类">
                        <div className="navigation-scroll" ref={navigationRef}>
                            <div className="navigation-grid">
                                {sections.map((item: any) => (
                                    <button
                                        type="button"
                                        className="navigation-item"
                                        key={item.key}
                                        data-section={item.key}
                                        data-active={active === item.key}
                                        aria-current={active === item.key ? "page" : undefined}
                                        style={{"--accent": item.color} as any}
                                        onClick={() => {
                                            if (active === item.key) return;
                                            setActive(item.key);
                                        }}>
                                        <span className="navigation-dot"/>
                                        <span className="navigation-label">{item.label}</span>
                                    </button>
                                ))}
                            </div>
                        </div>
                    </aside>
                    <main ref={contentRef} className={styles.content}><div className={styles.contentInner}>{content}</div></main>
                </div>
            </SettingsFormContext.Provider>
        </ProModal>
    );
});

Settings.displayName = "Settings";

export default memo(Settings);
