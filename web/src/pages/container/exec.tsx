import {forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {Alert, Button, Col, Collapse, Form, Grid, Input, InputNumber, Row, Segmented, Spin, Tag, Typography} from "antd";
import {createStyles} from "antd-style";
import {ProModal, Title, useTheme} from "sinking-antd";
import {execContainer, getContainerStatus} from "@/service/api/container";

import useContainerStyles from "./styles";

const useStyles = createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    modal: css`
        .ant-modal { max-width: calc(100vw - 24px); }
        .ant-modal-body { padding-top: ${token.paddingSM}px; overflow-x: hidden; }
    `,
    form: css`
        .ant-form-item { margin-bottom: ${compact ? token.marginSM : token.margin}px; }
        .ant-form-item-label > label { color: ${token.colorTextSecondary}; font-size: 12px; }
        .exec-validated-field .ant-form-item-additional { min-height: 14px; }
        .exec-validated-field .ant-form-item-explain { font-size: 11px; line-height: 14px; }
        .exec-command-field .ant-form-item-label > label { width: 100%; }
        .exec-command-heading { display: flex; align-items: baseline; justify-content: space-between; flex-wrap: wrap; gap: 2px 12px; width: 100%; }
        .exec-command-hint { color: ${token.colorTextTertiary}; font-size: 11px; font-weight: 400; line-height: 18px; }
        .ant-input-number { width: 100%; }
        .ant-input, .ant-input-number, .ant-collapse { font-size: 12px; }
        textarea.ant-input { font-family: ${token.fontFamilyCode}; resize: none; }
        .exec-mode { display: flex; width: fit-content; max-width: 100%; margin: 0 auto ${token.marginSM}px; font-size: 12px; }
        .exec-mode .ant-form-item { margin: 0; }
        .exec-mode .ant-segmented { font-size: 12px; }
        .exec-input-panel .ant-form-item:last-child { margin-bottom: 0; }
        .exec-advanced-fields { display: grid; gap: ${compact ? 8 : 12}px; }
        .exec-advanced-fields .ant-form-item { margin-bottom: 0; }
        @media (max-width: 575px) {
            .exec-command-shortcut { display: none; }
        }
    `,
    advanced: css`
        && {
            margin-top: ${compact ? 12 : 16}px;
            border: 1px solid ${token.colorBorderSecondary};
            border-radius: ${token.borderRadiusLG}px;
        }
        && > .ant-collapse-item > .ant-collapse-header {
            color: ${token.colorTextSecondary};
            background: ${token.colorFillTertiary};
            border-radius: ${token.borderRadiusLG - 1}px;
        }
        && > .ant-collapse-item-active > .ant-collapse-header { border-end-start-radius: 0; border-end-end-radius: 0; }
        .ant-collapse-title { min-width: 0; }
        .exec-advanced-heading { display: flex; align-items: center; flex-wrap: wrap; gap: 2px 12px; }
        .exec-advanced-title { font-size: 12px; line-height: 20px; color: ${token.colorTextSecondary}; }
        .exec-advanced-summary { margin-inline-start: auto; font-size: 11px; line-height: 18px; color: ${token.colorTextTertiary}; }
        .exec-option-group + .exec-option-group { padding-top: ${compact ? 10 : 12}px; border-top: 1px solid ${token.colorBorderSecondary}; }
        .exec-option-heading { margin: 0 0 ${compact ? 8 : 10}px; color: ${token.colorTextTertiary}; font-size: 11px; font-weight: 400; line-height: 18px; }
        .exec-option-fields { display: grid; gap: ${compact ? 8 : 12}px; }
        .ant-input-number-suffix { color: ${token.colorTextTertiary}; font-size: 11px; }
        .ant-collapse-arrow { font-size: 10px; }
        @media (max-width: 575px) {
            .exec-advanced-summary { flex-basis: 100%; margin-inline-start: 0; }
        }
    `,
    result: css`
        display: grid;
        gap: ${token.marginXS}px;
        margin-top: ${token.marginSM}px;
        min-width: 0;
        .exec-result-heading { display: flex; align-items: center; justify-content: space-between; gap: ${token.marginXS}px; flex-wrap: wrap; min-height: 24px; }
        .exec-result-status { display: flex; align-items: center; gap: ${token.marginXS}px; flex-wrap: wrap; }
        .ant-tag { margin: 0; font-size: 11px; line-height: 18px; }
        .exec-result-code, .exec-copy { color: ${token.colorTextSecondary}; font-size: 11px; }
        .exec-output {
            height: ${compact ? 100 : 120}px;
            min-width: 0;
            margin: 0;
            padding: ${token.paddingSM}px;
            box-sizing: border-box;
            overflow-y: auto;
            overflow-x: hidden;
            border-radius: ${token.borderRadius}px;
            background: ${token.colorFillQuaternary};
            color: ${token.colorText};
            font-family: ${token.fontFamilyCode};
            font-size: 12px;
            line-height: 18px;
            white-space: pre-wrap;
            overflow-wrap: anywhere;
        }
        .exec-output-placeholder { display: flex; align-items: center; justify-content: center; gap: ${token.marginXS}px; color: ${token.colorTextTertiary}; font-family: inherit; text-align: center; white-space: normal; }
    `,
}));

const ContainerExec = forwardRef<any, {onChange?: () => void}>(({onChange}, ref) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const {styles: containerStyles} = useContainerStyles();
    const [form] = Form.useForm<any>();
    const mode = Form.useWatch("mode", form) || "shell";
    const executionTimeout = Form.useWatch("timeout", form);
    const outputLimit = Form.useWatch("max_output", form);
    const sessionRef = useRef(0);
    const statusRequestRef = useRef(0);
    const mountedRef = useRef(true);
    const pendingRef = useRef(new Set<string>());
    const [open, setOpen] = useState(false);
    const [target, setTarget] = useState<any>();
    const [pending, setPending] = useState(new Set<string>());
    const [status, setStatus] = useState<any>();
    const [statusLoading, setStatusLoading] = useState(false);
    const [statusError, setStatusError] = useState("");
    const [executeError, setExecuteError] = useState("");
    const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
    const [result, setResult] = useState<any>();
    const [advanced, setAdvanced] = useState<string[]>([]);
    const executing = Boolean(target && pending.has(target.id));

    const loadStatus = useCallback(async (record: any, session: number) => {
        const request = ++statusRequestRef.current;
        setStatusLoading(true);
        setStatusError("");
        setStatus(undefined);
        try {
            const response = await getContainerStatus({body: {id: record.id}});
            if (sessionRef.current !== session || statusRequestRef.current !== request) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "获取容器状态失败");
            setStatus(response.data);
        } catch (error) {
            if (sessionRef.current === session && statusRequestRef.current === request) {
                setStatusError(error instanceof Error ? error.message : "获取容器状态失败");
            }
        } finally {
            if (sessionRef.current === session && statusRequestRef.current === request) setStatusLoading(false);
        }
    }, []);

    useImperativeHandle(ref, () => ({open: (record: any) => {
        const session = ++sessionRef.current;
        form.resetFields();
        setTarget(record);
        setResult(undefined);
        setExecuteError("");
        setFieldErrors({});
        setAdvanced([]);
        setOpen(true);
        void loadStatus(record, session);
    }}), [form, loadStatus]);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            sessionRef.current += 1;
        };
    }, []);

    const close = useCallback(() => {
        sessionRef.current += 1;
        setOpen(false);
    }, []);

    const execute = async (values: any) => {
        if (!open || !target || statusLoading || statusError || status?.status !== "running" || pendingRef.current.has(target.id)) return;
        const session = sessionRef.current;
        const id = target.id;
        pendingRef.current.add(id);
        setPending(new Set(pendingRef.current));
        setResult(undefined);
        setExecuteError("");
        try {
            const response = await execContainer({body: {
                id,
                // Shell 接收完整命令；直接执行仍按行传参，保留参数内的空格。
                command: values.mode === "shell"
                    ? ["/bin/sh", "-c", values.script || ""]
                    : [(values.command || "").trim(), ...(values.args || "").split(/\r?\n/).filter((line: string) => line.length > 0)],
                env: (values.env || "").split(/\r?\n/).filter((line: string) => line.length > 0),
                working_dir: values.working_dir?.trim() || "",
                stdin: values.stdin || "",
                timeout: values.timeout,
                max_output: Number(values.max_output ?? 512) * 1024,
            }});
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "执行失败");
            if (mountedRef.current) onChange?.();
            if (sessionRef.current !== session) return;
            setResult(response.data);
        } catch (error) {
            if (sessionRef.current === session) setExecuteError(error instanceof Error ? error.message : "执行失败");
        } finally {
            // 关闭后仍保留执行锁，避免重开同一容器时发出重复命令。
            pendingRef.current.delete(id);
            if (mountedRef.current) setPending(new Set(pendingRef.current));
        }
    };

    const retryStatus = <Button size="small" color="default" variant="filled" disabled={executing}
        onClick={() => target && void loadStatus(target, sessionRef.current)}>重新检查</Button>;

    return <ProModal title={<Title>执行命令</Title>} width={compact ? 520 : 560}
        onOk={() => form.submit()} onCancel={close} okText="执行" modalProps={{
            open, forceRender: true, rootClassName: styles.modal,
            style: {top: screens.md ? 80 : 20, paddingBottom: 40},
            confirmLoading: executing, cancelText: "关闭", mask: {closable: true},
            okButtonProps: {disabled: statusLoading || Boolean(statusError) || status?.status !== "running"},
            afterOpenChange: (visible: boolean) => { if (visible) form.focusField(mode === "shell" ? "script" : "command"); },
        } as any}>
        <div className={containerStyles.detail} style={{minHeight: 0}}>
            {statusError && <Alert type="error" showIcon title={statusError} action={retryStatus}/>}
            {!statusLoading && !statusError && status && status.status !== "running" && <Alert type="warning" showIcon
                title="容器未运行，请启动后再执行命令。" action={retryStatus}/>}
            <Form form={form} layout="vertical" variant="filled" className={styles.form} onFinish={execute}
                disabled={executing} requiredMark={false} initialValues={{mode: "shell", timeout: 30, max_output: 512}}
                onFieldsChange={() => setFieldErrors(Object.fromEntries(form.getFieldsError().map(({name, errors}) => [name[0], errors[0]])))}
                onKeyDown={(event) => {
                    if ((event.ctrlKey || event.metaKey) && event.key === "Enter" && !event.nativeEvent.isComposing) {
                        event.preventDefault();
                        if (!executing) form.submit();
                    }
                }}
                onFinishFailed={({errorFields}) => {
                    if (errorFields.some(({name}) => name[0] === "timeout" || name[0] === "max_output")) setAdvanced(["more"]);
                }}>
                <div className="exec-mode"><Form.Item name="mode">
                    <Segmented aria-label="执行方式" disabled={executing}
                        options={[{value: "shell", label: "命令执行"}, {value: "direct", label: "程序执行"}]}/>
                </Form.Item></div>
                {/* 保留输入值，隐藏非当前模式；固定行数避免切换时重新测量高度。 */}
                <div className="exec-input-panel" hidden={mode !== "shell"}>
                    <Form.Item name="script" className="exec-validated-field exec-command-field"
                        label={<span className="exec-command-heading">
                            <span>命令内容</span>
                            <span className="exec-command-hint">支持管道<span className="exec-command-shortcut"> · Ctrl / ⌘ + Enter 执行</span></span>
                        </span>}
                        help={<span>{fieldErrors.script || "\u00a0"}</span>}
                        rules={mode === "shell" ? [{required: true, whitespace: true, message: "请输入命令内容"}] : []}>
                        <Input.TextArea rows={3} spellCheck={false} placeholder="例如：ls -al /data"/>
                    </Form.Item>
                </div>
                <div className="exec-input-panel" hidden={mode !== "direct"}>
                    <Form.Item name="command" className="exec-validated-field exec-command-field"
                        label={<span className="exec-command-heading">
                            <span>执行程序</span>
                            <span className="exec-command-hint">无需 Shell<span className="exec-command-shortcut"> · Ctrl / ⌘ + Enter 执行</span></span>
                        </span>}
                        help={<span>{fieldErrors.command || "\u00a0"}</span>}
                        rules={mode === "direct" ? [{required: true, whitespace: true, message: "请输入程序名称或路径"}] : []}>
                        <Input autoComplete="off" spellCheck={false} placeholder="例如：redis-cli 或 /usr/bin/redis-cli"/>
                    </Form.Item>
                    <Form.Item name="args" className="exec-command-field"
                        label={<span className="exec-command-heading">
                            <span>命令参数</span>
                            <span className="exec-command-hint">每行一个参数，空格原样保留</span>
                        </span>}>
                        <Input.TextArea rows={3} spellCheck={false} placeholder={"例如：\n-h\n127.0.0.1\nPING"}/>
                    </Form.Item>
                </div>
                <Collapse ghost size="small" className={styles.advanced} activeKey={advanced} onChange={setAdvanced} expandIconPlacement="end"
                    style={{marginTop: mode === "shell" ? 0 : undefined}}
                    styles={{
                        header: {padding: `${compact ? 8 : 10}px 12px`, alignItems: "center"},
                        icon: {height: 20},
                        body: {padding: `${compact ? 10 : 12}px 12px`},
                    }}
                    items={[{key: "more", label: <div className="exec-advanced-heading">
                        <span className="exec-advanced-title">高级选项</span>
                        <span className="exec-advanced-summary">
                            {executionTimeout == null ? "超时未设置" : `超时 ${executionTimeout} 秒`}
                            {" · "}{outputLimit == null ? "输出上限未设置" : `输出 ${outputLimit} KB`}
                        </span>
                    </div>, forceRender: true, children: <div className="exec-advanced-fields">
                        <section className="exec-option-group" aria-label="执行限制">
                            <h3 className="exec-option-heading">执行限制</h3>
                            <Row gutter={[12, compact ? 8 : 12]}>
                                <Col span={12}><Form.Item name="timeout" label="超时时间" className="exec-validated-field"
                                    help={<span>{fieldErrors.timeout || "\u00a0"}</span>} rules={[{required: true, message: "请输入超时时间"}]}>
                                    <InputNumber min={1} max={3600} precision={0} suffix="秒"/>
                                </Form.Item></Col>
                                <Col span={12}><Form.Item name="max_output" label="输出上限" className="exec-validated-field"
                                    help={<span>{fieldErrors.max_output || "\u00a0"}</span>} rules={[{required: true, message: "请输入输出上限"}]}>
                                    <InputNumber min={1} max={16384} precision={0} suffix="KB"/>
                                </Form.Item></Col>
                            </Row>
                        </section>
                        <section className="exec-option-group" aria-label="运行环境">
                            <h3 className="exec-option-heading">运行环境</h3>
                            <div className="exec-option-fields">
                                <Form.Item name="working_dir" label="工作目录"><Input placeholder="沿用容器工作目录"/></Form.Item>
                                <Row gutter={[12, compact ? 8 : 12]}>
                                    <Col xs={24} sm={12}><Form.Item name="env" label="环境变量">
                                        <Input.TextArea rows={2} spellCheck={false} placeholder="每行一个 KEY=VALUE"/>
                                    </Form.Item></Col>
                                    <Col xs={24} sm={12}><Form.Item name="stdin" label="标准输入">
                                        <Input.TextArea rows={2} spellCheck={false} placeholder="需要传给程序的输入内容"/>
                                    </Form.Item></Col>
                                </Row>
                            </div>
                        </section>
                    </div>}]}/>
            </Form>
            <section className={styles.result} aria-label="执行结果">
                <div className="exec-result-heading">
                    <div className="exec-result-status">
                        <span className="detail-label">执行结果</span>
                        {executing ? <Tag variant="filled" color="processing">执行中</Tag> : result ? <>
                            <Tag variant="filled" color={result.timed_out ? "warning" : result.exit_code === 0 ? "success" : "error"}>
                                {result.timed_out ? "执行超时" : result.exit_code === 0 ? "执行成功" : "执行失败"}
                            </Tag>
                            <span className="exec-result-code">退出码 {result.exit_code}</span>
                            {result.truncated && <Tag variant="filled" color="warning">输出已截断</Tag>}
                        </> : executeError ? <Tag variant="filled" color="error">执行失败</Tag> : null}
                    </div>
                    {result?.output && <Typography.Text className="exec-copy"
                        styles={{
                            root: {display: "inline-flex", alignItems: "center", paddingInlineEnd: 0, lineHeight: "18px", whiteSpace: "nowrap"},
                            actions: {display: "inline-flex", alignItems: "center"},
                            action: {position: "static", display: "inline-flex", alignItems: "center", marginInlineStart: 4},
                        }}
                        copyable={{text: result.output, tooltips: ["复制输出", "已复制"]}}>复制</Typography.Text>}
                </div>
                {result?.output || executeError ? <pre className="exec-output" tabIndex={0} aria-label="命令输出">
                    {result?.output || executeError}
                </pre> : <div className="exec-output exec-output-placeholder" role="status">
                    {executing && <Spin size="small"/>}
                    <span>{executing ? "执行中，关闭窗口不会终止命令" : result ? "命令已结束，没有输出" : "执行结果将在这里显示"}</span>
                </div>}
            </section>
        </div>
    </ProModal>;
});

ContainerExec.displayName = "ContainerExec";
export default ContainerExec;
