import {forwardRef, memo, useCallback, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Col, Form, Grid, Input, InputNumber, Row, Segmented, Select, Switch} from "antd";
import {createStyles} from "antd-style";
import {ProModal, ProModalRef, Title, useTheme} from "sinking-antd";
import FilePicker from "@/pages/components/file-picker";
import {createSite} from "@/service/api/site";
import {clearEnumCache} from "@/utils/enum";
import {parseProxyTarget} from "../../proxy";

const siteTypes = [
    {label: "静态网站", value: 0},
    {label: "反向代理", value: 1},
    {label: "通用网站", value: 2},
    {label: "PHP 网站", value: 3},
];

const useStyles = createStyles(({css}) => ({
    typeSelect: css`
        @media (max-width: 480px) {
            .ant-segmented-item-label {
                padding-inline: 4px;
                font-size: 12px;
            }
        }
    `,
}));

const splitValues = (value: any) => Array.from(new Set(
    (Array.isArray(value) ? value : String(value || "").split(/[,\n\s]+/))
        .map((item) => String(item || "").trim())
        .filter(Boolean),
));

const SiteForm = forwardRef(({onSuccess, typeItems}: any, ref): any => {
    const {message} = App.useApp();
    const theme = useTheme();
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles();
    const compact = Boolean(theme?.isCompactTheme?.());
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const requestRef = useRef(0);
    const submittingRef = useRef(false);
    const [form] = Form.useForm();
    const [siteType, setSiteType] = useState(0);
    const [submitting, setSubmitting] = useState(false);
    const categoryOptions = useMemo(() => [
        {label: "全部分类", value: 0},
        ...(typeItems || []).map((item: any) => ({label: String(item.name), value: Number(item.id)})),
    ], [typeItems]);

    const initialValues = useMemo(() => ({
        type: 0,
        type_id: 0,
        domains: [],
        remark: "",
        root: "",
        target: "http://127.0.0.1:8080",
        host: "",
        project_name: "",
        project_file: "",
        project_port: 8080,
        arguments: "",
        enabled: true,
        fastcgi_address: "127.0.0.1:9000",
    }), []);

    const reset = useCallback(() => {
        requestRef.current += 1;
        submittingRef.current = false;
        form.resetFields();
        setSiteType(0);
        setSubmitting(false);
    }, [form]);

    const open = useCallback(() => {
        reset();
        form.setFieldsValue(initialValues);
        modalRef.current?.show();
    }, [form, initialValues, reset]);

    const close = useCallback(() => {
        requestRef.current += 1;
        submittingRef.current = false;
        setSubmitting(false);
        modalRef.current?.hide();
    }, []);

    useImperativeHandle(ref, () => ({open}), [open]);

    const submit = useCallback(async (values: any) => {
        if (submittingRef.current) {
            return;
        }
        const requestId = ++requestRef.current;
        submittingRef.current = true;
        setSubmitting(true);

        const domains = splitValues(values.domains);
        const remark = String(values.remark || "").trim();
        const body: any = {
            type: siteType,
            type_id: Number(values.type_id || 0),
            status: values.enabled === false ? 1 : 0,
            domains,
            name: siteType === 2 ? String(values.project_name || "").trim() : remark || domains[0],
        };

        if (siteType === 0) {
            body.root = String(values.root || "").trim();
            body.static = {
                index: ["index.html", "index.htm"],
                browse: false,
                try_files: [],
                hide: [".git", ".svn", ".env", ".env.*", "*.sql", "*.bak", "*.backup"],
                precompressed: ["zstd", "gzip"],
            };
        } else if (siteType === 1) {
            const target = parseProxyTarget(values.target);
            const host = String(values.host || "").trim();
            body.proxy = {
                upstreams: [{dial: target.dial, max_requests: 0}],
                transport: "http",
                scheme: target.scheme,
                ...(target.h2c ? {versions: ["h2c"]} : {}),
                policy: "random",
            };
            if (host) {
                body.proxy.headers = {
                    request_set: {
                        Host: [host === "$host" || host === "$http_host" ? "{http.request.host}" : host],
                    },
                };
            }
        } else if (siteType === 2) {
            const executable = String(values.project_file || "").trim();
            const argumentsValue = String(values.arguments || "").trim();
            const command = executable.includes(" ") ? `"${executable.replace(/"/g, "\\\"")}"` : executable;
            const separator = Math.max(executable.lastIndexOf("/"), executable.lastIndexOf("\\"));
            body.root = separator > 0 ? executable.slice(0, separator) : "";
            body.proxy = {
                upstreams: [{dial: `127.0.0.1:${Number(values.project_port)}`, max_requests: 0}],
                transport: "http",
                scheme: "http",
                policy: "random",
            };
            body.process = {
                command: `${command}${argumentsValue ? ` ${argumentsValue}` : ""}`,
                environment: {},
                max_retries: 3,
                restart_delay: 2_000_000_000,
                stop_timeout: 10_000_000_000,
            };
        } else {
            const address = String(values.fastcgi_address || "").trim();
            body.root = String(values.root || "").trim();
            body.fastcgi = {
                upstreams: [{dial: address.startsWith("/") ? `unix/${address}` : address, max_requests: 0}],
                transport: "fastcgi",
                policy: "random",
                split_path: [".php"],
                index: "index.php",
                try_files: [],
                capture_stderr: true,
            };
        }

        try {
            const response = await createSite({body});
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code === 200) {
                clearEnumCache("site");
                modalRef.current?.hide();
                message.success(response?.message || "创建成功");
                onSuccess?.(response.data);
            } else {
                message.error(response?.message || "创建失败");
            }
        } finally {
            if (requestRef.current === requestId) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    }, [message, onSuccess, siteType]);

    const domainField = (
        <Form.Item name="domains" label="绑定域名" extra={siteType === 2 ? "可留空，项目仍可独立运行；需要通过网站访问时再绑定域名。" : undefined} rules={[{required: siteType !== 2, message: "请至少输入一个域名"}]}>
            <Select mode="tags" tokenSeparators={[",", " "]} placeholder={siteType === 2 ? "可留空，或输入域名后按回车" : "输入域名后按回车，例如 example.com"}/>
        </Form.Item>
    );

    return (
        <ProModal
            ref={modalRef}
            title={<Title>添加网站</Title>}
            width={620}
            onOk={form.submit}
            onCancel={close}
            modalProps={{
                forceRender: true,
                confirmLoading: submitting,
                closable: true,
                keyboard: true,
                okText: "确定",
                cancelText: "取消",
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                mask: {closable: true},
                styles: {body: {
                    maxHeight: screens.md ? "calc(100dvh - 316px)" : "calc(100dvh - 190px)",
                    paddingTop: compact ? 10 : 16,
                    overflowX: "hidden",
                    overflowY: "auto",
                }},
                afterClose: reset,
            } as any}>
            <Form form={form} layout="vertical" initialValues={initialValues} onFinish={submit}>
                <Form.Item name="type" label="网站类型">
                    <Segmented
                        className={styles.typeSelect}
                        block
                        options={siteTypes}
                        onChange={(value) => {
                            const nextType = Number(value);
                            setSiteType(nextType);
                            if (nextType === 1) {
                                form.setFieldsValue({target: "http://127.0.0.1:8080"});
                            }
                        }}/>
                </Form.Item>
                <Form.Item name="type_id" label="网站分类">
                    <Select
                        placeholder="全部分类"
                        options={categoryOptions}/>
                </Form.Item>

                {siteType === 1 && (
                    <>
                        {domainField}
                        <Form.Item
                            name="target"
                            label="目标地址"
                            extra="自动识别 HTTP、HTTPS 和 UNIX 套接字；网址不填协议时使用 HTTP。"
                            rules={[
                                {required: true, whitespace: true, message: "请输入目标地址"},
                                {validator: async (_, value) => {
                                    if (String(value || "").trim()) parseProxyTarget(value);
                                }},
                            ]}>
                            <Input placeholder="http://127.0.0.1:8080 或 /tmp/panel.sock"/>
                        </Form.Item>
                        <Form.Item name="host" label="发送域名 (Host)">
                            <Input placeholder="留空自动使用目标地址的域名"/>
                        </Form.Item>
                        <Form.Item name="remark" label="备注">
                            <Input maxLength={100} placeholder="可选"/>
                        </Form.Item>
                    </>
                )}

                {siteType === 2 && (
                    <>
                        <Form.Item name="project_file" label="项目执行文件" rules={[{required: true, whitespace: true, message: "请选择项目执行文件"}]}>
                            <FilePicker mode="file" placeholder="请输入或选择执行文件"/>
                        </Form.Item>
                        <Row gutter={[16, 0]}>
                            <Col xs={24} sm={15}>
                                <Form.Item name="project_name" label="项目名称" rules={[{required: true, whitespace: true, message: "请输入项目名称"}]}>
                                    <Input maxLength={100} placeholder="例如 API 服务"/>
                                </Form.Item>
                            </Col>
                            <Col xs={24} sm={9}>
                                <Form.Item name="project_port" label="项目端口" rules={[{required: true, message: "请输入项目端口"}]}>
                                    <InputNumber min={1} max={65535} precision={0} placeholder="请输入项目端口" style={{width: "100%"}}/>
                                </Form.Item>
                            </Col>
                        </Row>
                        <Form.Item name="arguments" label="执行参数">
                            <Input placeholder="例如 --config ./config.yaml"/>
                        </Form.Item>
                        <Row gutter={[16, 0]} align="bottom">
                            <Col xs={24} sm={18}>{domainField}</Col>
                            <Col xs={24} sm={6}>
                                <Form.Item name="enabled" label="开机启动" valuePropName="checked"><Switch/></Form.Item>
                            </Col>
                        </Row>
                    </>
                )}

                {(siteType === 0 || siteType === 3) && (
                    <>
                        {domainField}
                        <Form.Item name="remark" label="备注">
                            <Input maxLength={100} placeholder="可选，留空使用首个域名"/>
                        </Form.Item>
                        <Form.Item name="root" label="根目录" extra="留空时自动创建 ./data/wwwroot/首个域名，例如 ./data/wwwroot/example.com">
                            <FilePicker mode="directory" placeholder="自动创建，或选择已有目录"/>
                        </Form.Item>
                        {siteType === 3 && (
                            <Form.Item name="fastcgi_address" label="PHP-FPM 地址" rules={[{required: true, whitespace: true, message: "请输入 PHP-FPM 地址"}]}>
                                <Input placeholder="127.0.0.1:9000 或 /run/php/php-fpm.sock"/>
                            </Form.Item>
                        )}
                    </>
                )}
            </Form>
        </ProModal>
    );
});

SiteForm.displayName = "SiteForm";

export default memo(SiteForm);
