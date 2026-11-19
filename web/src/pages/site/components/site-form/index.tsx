import {forwardRef, memo, useCallback, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Button, Col, Collapse, Descriptions, Form, Grid, Input, InputNumber, Radio, Row, Select, Steps, Switch} from "antd";
import type {ModalProps} from "antd";
import {createStyles} from "antd-style";
import {ProModal, ProModalRef, Title, useTheme} from "sinking-antd";
import FileSelect from "@/pages/components/file-select";
import {createSite} from "@/service/api/site";
import {clearEnumCache} from "@/utils/enum";
import {parseProxyTarget} from "../../proxy";

const siteTypes = [
    {label: "静态网站", value: 0, description: "已有 HTML 网页或前端打包后的文件"},
    {label: "反向代理", value: 1, description: "已有服务正在运行，希望通过域名访问"},
    {label: "通用网站", value: 2, description: "运行本机程序，由面板负责启动和保活"},
    {label: "PHP 网站", value: 3, description: "部署 PHP 项目，需已有 PHP-FPM 环境"},
];

const useStyles = createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    modal: css`
        .ant-modal { max-width: calc(100vw - 24px); }
        .ant-modal-body { padding-top: ${compact ? 12 : 16}px; }
    `,
    steps: css`
        margin-bottom: ${compact ? 20 : 24}px;
    `,
    form: css`
        .ant-form-item { margin-bottom: ${compact ? 16 : 20}px; }
        .ant-input-number { width: 100%; }
        textarea.ant-input { resize: vertical; }
    `,
    typeSelect: css`
        display: flex;
        flex-direction: column;
        gap: ${compact ? 6 : 8}px;
        width: 100%;
        .ant-radio-wrapper {
            align-items: center;
            margin: 0;
            padding: ${compact ? "10px 12px" : "12px 14px"};
            border-radius: ${token.borderRadius}px;
            background: ${token.colorFillQuaternary};
            transition: background-color ${token.motionDurationMid};
        }
        .ant-radio-wrapper:hover { background: ${token.colorFillTertiary}; }
        .ant-radio-wrapper-checked, .ant-radio-wrapper-checked:hover { background: ${token.colorPrimaryBg}; }
        .ant-radio-wrapper > span:last-child { min-width: 0; padding-inline-end: 0; }
        .site-type-title { display: block; font-weight: 500; color: ${token.colorText}; }
        .site-type-description { display: block; margin-top: 3px; color: ${token.colorTextTertiary}; font-size: 12px; line-height: 1.5; }
    `,
    more: css`
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};
        .ant-collapse-item { border: 0; }
        .ant-collapse-header {
            align-items: center !important;
            padding: ${token.paddingSM}px !important;
            font-size: ${token.fontSize}px;
        }
        && .ant-collapse-body {
            padding: ${token.paddingXS}px ${token.paddingSM}px ${token.paddingSM}px;
        }
        && .ant-form-item { margin-bottom: 0; }
    `,
    review: css`
        margin-bottom: ${compact ? 16 : 20}px;
        padding: ${compact ? "12px 14px" : "14px 16px"};
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};
        .ant-descriptions-item-label { white-space: nowrap; }
        .ant-descriptions-item-content { min-width: 0; overflow-wrap: anywhere; }
        .site-review-domains { max-height: 72px; overflow-y: auto; white-space: pre-line; }
    `,
    enabled: css`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        padding: 4px 0;
    `,
}));

const splitValues = (value: any) => Array.from(new Set(
    (Array.isArray(value) ? value : String(value || "").split(/[\s,，;；]+/))
        .map((item) => String(item || "").trim())
        .filter(Boolean),
));

const SiteForm = forwardRef(({onSuccess, typeItems}: any, ref): any => {
    const {message} = App.useApp();
    const theme = useTheme();
    const screens = Grid.useBreakpoint();
    const compact = Boolean(theme?.isCompactTheme?.());
    const styleProps = useMemo(() => ({compact}), [compact]);
    const {styles} = useStyles(styleProps);
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const requestRef = useRef(0);
    const submittingRef = useRef(false);
    const [form] = Form.useForm();
    const [siteType, setSiteType] = useState(0);
    const [step, setStep] = useState(0);
    const [review, setReview] = useState<any>({});
    const [advanced, setAdvanced] = useState<string[]>([]);
    const [submitting, setSubmitting] = useState(false);
    const categoryOptions = useMemo(() => [
        {label: "全部分类", value: 0},
        ...(typeItems || []).map((item: any) => ({label: String(item.name), value: Number(item.id)})),
    ], [typeItems]);

    const initialValues = useMemo(() => ({
        type: 0,
        type_id: 0,
        domains: "",
        remark: "",
        root: "",
        target: "",
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
        setStep(0);
        setReview({});
        setAdvanced([]);
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
        if (step !== 2 || submittingRef.current) {
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
                onSuccess?.(response?.data);
            } else {
                message.error(response?.message || "创建失败");
            }
        } finally {
            if (requestRef.current === requestId) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    }, [message, onSuccess, siteType, step]);

    const next = useCallback(async () => {
        if (submittingRef.current) return;
        if (step === 0) {
            setStep(1);
            return;
        }
        const requestId = ++requestRef.current;
        try {
            const values = await form.validateFields(step === 1 ? [
                "domains",
                ...(siteType === 1 ? ["target"] : siteType === 2 ? ["project_file", "project_name", "project_port"] : siteType === 3 ? ["fastcgi_address"] : []),
            ] : undefined);
            if (requestRef.current !== requestId) return;
            if (step === 1) {
                setReview(form.getFieldsValue(true));
                setStep(2);
            } else {
                await submit(values);
            }
        } catch (error: any) {
            if (requestRef.current === requestId && error?.errorFields?.[0]) {
                setStep(1);
                window.requestAnimationFrame(() => {
                    if (requestRef.current === requestId) form.scrollToField(error?.errorFields?.[0]?.name, {block: "nearest", focus: true});
                });
            }
        }
    }, [form, siteType, step, submit]);

    const footer: ModalProps["footer"] = step > 0 ? (_, {OkBtn}) => <>
        <Button disabled={submitting} onClick={() => {
            requestRef.current += 1;
            setStep(step - 1);
        }}>上一步</Button>
        <OkBtn/>
    </> : undefined;

    const domainField = (
        <Form.Item
            name="domains"
            label={siteType === 2 ? "绑定域名（可选）" : "绑定域名"}
            required={siteType !== 2}
            rules={[
                {validator: async (_, value) => {
                    const domains = splitValues(value);
                    if (siteType !== 2 && !domains.length) throw new Error("请至少输入一个域名");
                    if (domains.length > 1000) throw new Error("每个网站最多绑定 1000 个域名");
                    if (domains.some((domain) => /[\/?#]/.test(domain) || (domain.match(/:/g) || []).length === 1)) {
                        throw new Error("只需填写域名或 IP，不要包含协议、端口和路径");
                    }
                }},
            ]}>
            <Input.TextArea rows={2} placeholder={"例如 example.com\n多个域名用换行或逗号分隔"}/>
        </Form.Item>
    );

    return (
        <ProModal
            ref={modalRef}
            title={<Title>添加网站</Title>}
            width={520}
            okText={step === 2 ? "创建网站" : "下一步"}
            onOk={next}
            onCancel={close}
            modalProps={{
                rootClassName: styles.modal,
                closable: true,
                keyboard: true,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                mask: {closable: true},
                afterClose: reset,
                okButtonProps: {loading: submitting},
                footer,
            } as any}>
            <Steps className={styles.steps} size="small" current={step} responsive={false}
                titlePlacement={screens.sm ? "horizontal" : "vertical"}
                items={[{title: "选择类型"}, {title: "网站配置"}, {title: "确认创建"}]}/>
            <Form className={styles.form} form={form} layout="vertical" variant="filled" disabled={submitting}
                initialValues={initialValues} onFinish={next}
                onSubmitCapture={(event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    void next();
                }}>
                <div>
                    <div hidden={step !== 0}>
                        <Form.Item name="type">
                            <Radio.Group className={styles.typeSelect} aria-label="网站类型" onChange={(event) => setSiteType(Number(event.target.value))}>
                                {siteTypes.map((item) => <Radio key={item.value} value={item.value}>
                                    <span className="site-type-title">{item.label}</span>
                                    <span className="site-type-description">{item.description}</span>
                                </Radio>)}
                            </Radio.Group>
                        </Form.Item>
                    </div>
                    <div hidden={step !== 1}>
                        {siteType !== 2 && domainField}
                        {siteType === 1 && <Form.Item
                            name="target"
                            label="目标服务地址"
                            rules={[
                                {required: true, whitespace: true, message: "请输入目标服务地址"},
                                {validator: async (_, value) => {
                                    if (String(value || "").trim()) parseProxyTarget(value);
                                }},
                            ]}>
                            <Input placeholder="例如 http://127.0.0.1:8080"/>
                        </Form.Item>}
                        {siteType === 2 && <>
                            <Form.Item name="project_file" label="执行文件"
                                getValueProps={(value) => ({value: value ? [value] : []})}
                                getValueFromEvent={(paths: string[]) => paths[0] || ""}
                                rules={[{required: true, whitespace: true, message: "请选择执行文件"}]}>
                                <FileSelect type="file" placeholder="输入路径，或点击右侧选择文件"
                                    buttonProps={{color: "default", variant: "filled"}}/>
                            </Form.Item>
                            <Row gutter={[16, 0]}>
                                <Col xs={24} sm={14}>
                                    <Form.Item name="project_name" label="项目名称" rules={[{required: true, whitespace: true, message: "请输入项目名称"}]}>
                                        <Input maxLength={100} placeholder="例如 API 服务"/>
                                    </Form.Item>
                                </Col>
                                <Col xs={24} sm={10}>
                                    <Form.Item name="project_port" label="程序监听端口"
                                        rules={[{required: true, message: "请输入程序监听端口"}]}>
                                        <InputNumber min={1} max={65535} precision={0} placeholder="例如 8080"/>
                                    </Form.Item>
                                </Col>
                            </Row>
                            {domainField}
                        </>}
                        {(siteType === 0 || siteType === 3) && <>
                            <Form.Item name="root" label="网站目录"
                                getValueProps={(value) => ({value: value ? [value] : []})}
                                getValueFromEvent={(paths: string[]) => paths[0] || ""}>
                                <FileSelect type="directory" placeholder="选择已有目录，留空自动创建"
                                    buttonProps={{color: "default", variant: "filled"}}/>
                            </Form.Item>
                            {siteType === 3 && <Form.Item name="fastcgi_address" label="PHP-FPM 地址"
                                rules={[{required: true, whitespace: true, message: "请输入 PHP-FPM 地址"}]}>
                                <Input placeholder="127.0.0.1:9000 或 /run/php/php-fpm.sock"/>
                            </Form.Item>}
                        </>}
                        {(siteType === 1 || siteType === 2) && <Collapse size="small" bordered={false} className={styles.more}
                            activeKey={advanced} onChange={(keys) => setAdvanced(Array.isArray(keys) ? keys : [keys])}
                            items={[{
                                key: "advanced",
                                label: "更多选项",
                                forceRender: true,
                                children: siteType === 2 ? <Form.Item name="arguments" label="启动参数">
                                    <Input placeholder="例如 --config ./config.yaml"/>
                                </Form.Item> : <Form.Item name="host" label="请求域名（Host）">
                                    <Input placeholder="默认自动处理"/>
                                </Form.Item>,
                            }]}/>}
                    </div>
                    <div hidden={step !== 2}>
                        <div className={styles.review}>
                            <Descriptions column={1} size="small" colon={false} items={[
                                {key: "type", label: "网站类型", children: siteTypes[siteType].label},
                                ...(siteType === 2 ? [{key: "name", label: "项目名称", children: review.project_name}] : []),
                                {key: "domains", label: "绑定域名", children: <div className="site-review-domains">{splitValues(review.domains).join("\n") || "不绑定，独立运行"}</div>},
                                {key: "target", label: siteType === 1 ? "转发到" : siteType === 2 ? "执行文件" : "网站目录",
                                    children: siteType === 1 ? review.target : siteType === 2 ? review.project_file : String(review.root || "").trim() || "按首个域名自动创建"},
                                ...(siteType === 2 ? [{key: "port", label: "监听端口", children: review.project_port}] : []),
                                ...(siteType === 3 ? [{key: "php", label: "PHP-FPM", children: review.fastcgi_address}] : []),
                            ]}/>
                        </div>
                        <Row gutter={[16, 0]}>
                            {siteType !== 2 && <Col xs={24} sm={14}>
                                <Form.Item name="remark" label="网站名称">
                                    <Input maxLength={100} placeholder={splitValues(review.domains)[0] || "留空使用首个域名"}/>
                                </Form.Item>
                            </Col>}
                            <Col xs={24} sm={siteType === 2 ? 24 : 10}>
                                <Form.Item name="type_id" label="网站分类">
                                    <Select options={categoryOptions} showSearch={false}/>
                                </Form.Item>
                            </Col>
                        </Row>
                        <div className={styles.enabled}>
                            <label htmlFor="create-site-enabled">创建后启用</label>
                            <Form.Item name="enabled" valuePropName="checked" noStyle><Switch id="create-site-enabled" aria-label="创建后启用"/></Form.Item>
                        </div>
                    </div>
                </div>
            </Form>
        </ProModal>
    );
});

SiteForm.displayName = "SiteForm";

export default memo(SiteForm);
