import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Button, Collapse, Dropdown, Form, Grid, Input, InputNumber, Select, Spin, Switch} from "antd";
import {createStyles} from "antd-style";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import FilePicker from "@/pages/components/file-picker";
import {getContainerImage, runContainer} from "@/service/api/container";
import {getEnum} from "@/service/api/system";

const useStyles = createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    modal: css`
        .ant-modal { max-width: calc(100vw - 24px); }
        .ant-modal-body { padding-top: ${token.paddingSM}px; overflow-x: hidden; }
    `,
    form: css`
        min-width: 0;
        .run-options-fields {
            font-size: 12px;
            .ant-form-item { margin-bottom: ${compact ? 12 : 16}px; }
            .ant-form-item-label { padding-bottom: 5px; }
            .ant-form-item-label > label { color: ${token.colorTextSecondary}; font-size: 12px; }
            .ant-form-item-label > label.ant-form-item-required::before { display: none; }
            .ant-form-item-extra, .ant-form-item-explain { min-height: 0; font-size: 11px; line-height: 18px; }
            .ant-form-item-extra { margin-top: 4px; color: ${token.colorTextTertiary}; }
            .ant-input, .ant-input-number, .ant-select, .ant-btn { font-size: 12px; }
            textarea.ant-input { font-family: ${token.fontFamilyCode}; resize: none; }
            .ant-input-number-suffix { color: ${token.colorTextTertiary}; font-size: 11px; }
        }
        .ant-input-number { width: 100%; }
        .run-muted { color: ${token.colorTextTertiary}; font-size: 11px; line-height: 18px; overflow-wrap: anywhere; }
        .run-image-status { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-height: 20px; }
        .run-image-error { color: ${token.colorError}; }
        .run-auto-start { display: flex; align-items: center; gap: 12px; margin-top: ${compact ? 12 : 16}px; padding: 4px ${compact ? 10 : 12}px; }
        .run-auto-copy { display: flex; flex: 1; flex-wrap: wrap; align-items: baseline; gap: 2px 12px; min-width: 0; }
        .run-auto-copy > label { color: ${token.colorText}; font-size: 12px; line-height: 20px; cursor: pointer; }
        .run-auto-description { color: ${token.colorTextTertiary}; font-size: 11px; line-height: 18px; }
        .run-auto-start > .ant-switch { flex-shrink: 0; }
        .run-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 8px; color: ${token.colorTextSecondary}; font-size: 11px; }
        .run-actions { display: flex; align-items: center; gap: 8px; }
        .run-section-heading { display: flex; align-items: baseline; justify-content: space-between; gap: 8px 12px; }
        .run-section-heading > span:first-child { flex-shrink: 0; }
        .run-variables { display: grid; gap: 8px; margin-bottom: 12px; }
        .run-variable, .run-variable-labels { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 40px; gap: 8px; align-items: start; }
        .run-variable-labels { color: ${token.colorTextSecondary}; font-size: 12px; }
        .run-variable .ant-form-item { margin-bottom: 0; }
        .run-remove { height: ${token.controlHeight}px; display: flex; align-items: center; justify-content: flex-end; }
        .run-mount { padding-bottom: 12px; }
        .run-mount + .run-mount { padding-top: 12px; border-top: 1px solid ${token.colorBorderSecondary}; }
        .run-fields { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
        .run-fields > .ant-form-item { min-width: 0; margin-bottom: 0; }
        .run-fields-wide { grid-column: 1 / -1; }
        .run-setting-group + .run-setting-group { margin-top: ${compact ? 14 : 16}px; padding-top: ${compact ? 14 : 16}px; border-top: 1px solid ${token.colorBorderSecondary}; }
        .run-setting-title { margin: 0 0 12px; font-size: 12px; color: ${token.colorTextSecondary}; }
        .run-field-label { display: flex; justify-content: space-between; align-items: baseline; gap: 8px; width: 100%; }
        .run-command .ant-form-item-label > label { width: 100%; }
        .run-access .ant-form-item { margin-bottom: 0; }
        .run-access .ant-form-item-row { display: flex; align-items: center; justify-content: space-between; }
        .run-access .ant-form-item-label { padding: 0; }
        .run-access .ant-form-item-control { flex: none; }
        @media (max-width: 575px) {
            .run-section-heading { flex-wrap: wrap; gap: 2px 8px; }
            .run-variable-labels { display: none; }
            .run-variable { grid-template-columns: minmax(0, 1fr) auto; }
            .run-variable > .ant-form-item:nth-child(2) { grid-column: 1; grid-row: 2; }
            .run-variable > .run-remove { grid-column: 2; grid-row: 1 / 3; align-self: center; }
            .run-variables { gap: 12px; }
            .run-mount .run-fields, .run-runtime-fields { grid-template-columns: minmax(0, 1fr); }
            .run-field-label { flex-wrap: wrap; gap: 0 8px; }
        }
    `,
    options: css`
        && { background: transparent; border: 0; }
        && > .ant-collapse-item { overflow: hidden; margin-bottom: 8px; border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadiusLG}px; }
        && > .ant-collapse-item:last-child { margin-bottom: 0; }
        && > .ant-collapse-item > .ant-collapse-header { align-items: center; padding: ${compact ? 10 : 12}px; background: ${token.colorFillQuaternary}; color: ${token.colorText}; font-size: 12px; }
        && > .ant-collapse-item > .ant-collapse-panel { border-top: 1px solid ${token.colorBorderSecondary}; }
        .ant-collapse-title { min-width: 0; }
        .ant-collapse-arrow { font-size: 10px; }
        .ant-collapse-body { padding: ${compact ? 12 : 16}px; }
        .ant-collapse-body > .ant-form-item:last-child { margin-bottom: 0; }
    `,
    mountPath: css`
        display: block;
        max-width: min(360px, calc(100vw - 64px));
        white-space: normal;
        overflow-wrap: anywhere;
        font-size: 12px;
    `,
}));

const RunContainer = forwardRef<any, {onChange?: () => void}>(({onChange}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const [form] = Form.useForm<any>();
    const sessionRef = useRef(0);
    const listRequestRef = useRef(0);
    const imageRequestRef = useRef(0);
    const submittingRef = useRef(false);
    const validationTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const [open, setOpen] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [images, setImages] = useState<any[]>([]);
    const [image, setImage] = useState<any>();
    const [imageLoading, setImageLoading] = useState(false);
    const [listLoading, setListLoading] = useState(false);
    const [listError, setListError] = useState("");
    const [expanded, setExpanded] = useState<string[]>([]);
    const mounts = Form.useWatch("mounts", form) || [];
    const variables = Form.useWatch("env", form) || [];

    const loadImages = useCallback(async () => {
        const session = sessionRef.current;
        const requestId = ++listRequestRef.current;
        setListLoading(true);
        setListError("");
        try {
            // 动态枚举每次打开重新获取，导入或删除镜像后立即生效。
            const response = await getEnum({body: {name: "container"}});
            if (sessionRef.current !== session || listRequestRef.current !== requestId) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "获取镜像失败");
            const names = response.data.image || {};
            setImages(Object.entries(names).map(([value, label]) => ({value, label: String(label), title: String(label)}))
                .sort((left, right) => left.label.localeCompare(right.label)));
            const selected = form.getFieldValue("image_id");
            if (selected && !Object.hasOwn(names, selected)) {
                form.setFieldValue("image_id", undefined);
                imageRequestRef.current += 1;
                setImage(undefined);
                setImageLoading(false);
            }
        } catch (error) {
            if (sessionRef.current === session && listRequestRef.current === requestId) {
                setListError(error instanceof Error ? error.message : "获取镜像失败");
            }
        } finally {
            if (sessionRef.current === session && listRequestRef.current === requestId) setListLoading(false);
        }
    }, [form]);

    const selectImage = useCallback(async (id: string) => {
        const session = sessionRef.current;
        const requestId = ++imageRequestRef.current;
        setImage(undefined);
        setImageLoading(true);
        try {
            const response = await getContainerImage({body: {id}});
            if (sessionRef.current !== session || imageRequestRef.current !== requestId) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "获取镜像详情失败");
            setImage(response.data);
        } catch (error) {
            if (sessionRef.current === session && imageRequestRef.current === requestId) {
                message.error(error instanceof Error ? error.message : "获取镜像详情失败");
            }
        } finally {
            if (sessionRef.current === session && imageRequestRef.current === requestId) setImageLoading(false);
        }
    }, [message]);

    useImperativeHandle(ref, () => ({
        open: (selected?: any) => {
            sessionRef.current += 1;
            listRequestRef.current += 1;
            imageRequestRef.current += 1;
            submittingRef.current = false;
            form.resetFields();
            form.setFieldsValue({image_id: selected?.id, mounts: [], env: [], auto_restart: true});
            setSubmitting(false);
            setImages(selected ? [{value: selected.id, label: selected.name || selected.id}] : []);
            setImage(undefined);
            setImageLoading(false);
            clearTimeout(validationTimerRef.current);
            setExpanded([]);
            setOpen(true);
            void loadImages();
            if (selected) void selectImage(selected.id);
        },
    }), [form, loadImages, selectImage]);

    useEffect(() => () => {
        sessionRef.current += 1;
        clearTimeout(validationTimerRef.current);
    }, []);

    const close = useCallback(() => {
        if (submittingRef.current) return;
        sessionRef.current += 1;
        clearTimeout(validationTimerRef.current);
        setOpen(false);
    }, []);

    const submit = useCallback(async (values: any) => {
        if (submittingRef.current || listLoading || imageLoading || listError) return;
        const session = sessionRef.current;
        submittingRef.current = true;
        clearTimeout(validationTimerRef.current);
        setSubmitting(true);
        try {
            const response = await runContainer({body: {
                id: values.id?.trim() || "",
                name: values.name?.trim() || "",
                image_id: values.image_id,
                mounts: (values.mounts || []).map((mount: any) => ({
                    source: mount.source.trim(), destination: mount.destination.trim(), read_only: Boolean(mount.read_only),
                })),
                // 参数按行原样传递，保留参数内的空格；不做 shell 拆词或执行。
                command: (values.command || "").split(/\r?\n/).filter((line: string) => line.length > 0),
                env: (values.env || []).map((item: any) => `${item.name.trim()}=${item.value || ""}`),
                working_dir: values.working_dir?.trim() || "",
                user: values.user?.trim() || "",
                resources: {
                    memory: Math.round(Number(values.memory || 0) * 1024 * 1024),
                    cpu_quota: Number(values.cpu_quota || 0),
                    cpu_period: Number(values.cpu_period || 0),
                    pids_limit: Number(values.pids_limit || 0),
                },
                auto_restart: Boolean(values.auto_restart),
                read_only: Boolean(values.read_only),
            }});
            if (sessionRef.current !== session) return;
            if (response?.code !== 200) throw new Error(response?.message || "创建容器失败");
            message.success(response.message || "创建容器成功");
            submittingRef.current = false;
            setSubmitting(false);
            sessionRef.current += 1;
            setOpen(false);
        } catch (error) {
            if (sessionRef.current === session) message.error(error instanceof Error ? error.message : "创建容器失败");
        } finally {
            if (sessionRef.current === session) {
                submittingRef.current = false;
                setSubmitting(false);
            }
            // 启动失败也可能生成实例；列表刷新失败不影响创建结果。
            void Promise.resolve().then(() => onChange?.()).catch(() => undefined);
        }
    }, [message, onChange, listLoading, imageLoading, listError]);

    return <ProModal title={<Title>添加容器</Title>} width={compact ? 560 : 600} onOk={() => form.submit()} onCancel={close}
        okText="创建" modalProps={{
            open, forceRender: true, rootClassName: styles.modal,
            style: {top: screens.md ? 80 : 20, paddingBottom: 40},
            closable: !submitting, keyboard: !submitting, mask: {closable: !submitting},
            confirmLoading: submitting, cancelText: "取消", cancelButtonProps: {disabled: submitting},
            okButtonProps: {disabled: imageLoading || listLoading || Boolean(listError)},
        } as any}>
        <Form form={form} layout="vertical" variant="filled" className={styles.form} disabled={submitting}
            initialValues={{auto_restart: true, mounts: [], env: []}} onFinish={submit}
            onFinishFailed={({errorFields}) => {
                const field = errorFields[0]?.name;
                if (!field) return;
                const groups = errorFields.map(({name}) => name[0] === "env" ? "env" : name[0] === "mounts" ? "mounts"
                    : ["image_id", "name"].includes(String(name[0])) ? "" : "advanced").filter(Boolean);
                setExpanded((current) => Array.from(new Set([...current, ...groups])));
                clearTimeout(validationTimerRef.current);
                // 等待折叠内容展开，再定位到错误项，避免焦点落在隐藏字段里。
                validationTimerRef.current = setTimeout(() => form.scrollToField(field, {block: "nearest", focus: true}), 300);
            }}>
            <Form.Item name="image_id" label="运行镜像" rules={[{required: true, message: "请选择要运行的镜像"}]}
                extra={listError ? <div className="run-image-status">
                    <span className="run-image-error">{listError}</span>
                    <Button size="small" color="default" variant="filled" loading={listLoading} onClick={() => void loadImages()}>重试</Button>
                </div> : undefined}>
                <Select showSearch={{optionFilterProp: "label"}} options={images} loading={listLoading || imageLoading} disabled={submitting || listLoading}
                    placeholder="搜索或选择镜像"
                    onChange={(value) => void selectImage(value)}
                    notFoundContent={listLoading ? <Spin size="small"/> : listError ? "镜像加载失败" : "暂无镜像"}/>
            </Form.Item>
            <Form.Item name="name" label="容器名称">
                <Input placeholder={image?.name || "例如：测试数据库（可不填）"} maxLength={100}/>
            </Form.Item>
            <Collapse className={`${styles.options} run-options-fields`} activeKey={expanded} onChange={(keys) => setExpanded(Array.isArray(keys) ? keys : [keys])}
                expandIconPlacement="end" items={[
                    {
                        key: "env", forceRender: true, collapsible: submitting ? "disabled" : undefined,
                        label: <div className="run-section-heading"><span>环境变量</span>
                            <span className="run-muted">{variables.length ? `已添加 ${variables.length} 项` : "设置软件密码等参数，按需填写"}</span></div>,
                        children: <Form.List name="env">{(fields, {add, remove}) => <>
                            {fields.length > 0 && <div className="run-variables">
                                <div className="run-variable-labels" aria-hidden="true"><span>变量名称</span><span>变量值</span></div>
                                {fields.map(({key, name}) => <div className="run-variable" key={key}>
                                <Form.Item name={[name, "name"]} rules={[
                                    {required: true, whitespace: true, message: "请输入变量名称"},
                                    {validator: async (_, value) => {
                                        if (/[=\0]/.test(value || "")) throw new Error("变量名称不能包含 = 或空字符");
                                    }},
                                ]}><Input aria-label={`变量 ${name + 1} 名称`} placeholder="变量名称" autoComplete="off" spellCheck={false}/></Form.Item>
                                <Form.Item name={[name, "value"]} rules={[{validator: async (_, value) => {
                                    if (String(value || "").includes("\0")) throw new Error("变量值不能包含空字符");
                                }}]}><Input aria-label={`变量 ${name + 1} 值`} placeholder="变量值" autoComplete="off" spellCheck={false}/></Form.Item>
                                <div className="run-remove"><Button color="danger" variant="text" size="small" disabled={submitting}
                                    aria-label={`移除变量 ${name + 1}`} onClick={() => remove(name)}>移除</Button></div>
                            </div>)}</div>}
                            <Button color="default" variant="filled" size="small" disabled={submitting} onClick={() => add({name: "", value: ""})}>添加变量</Button>
                        </>}</Form.List>,
                    },
                    {
                        key: "mounts", forceRender: true, collapsible: submitting ? "disabled" : undefined,
                        label: <div className="run-section-heading"><span>数据存储</span>
                            <span className="run-muted">{mounts.length ? `已添加 ${mounts.length} 项挂载` : "将容器数据保存在服务器上"}</span></div>,
                        children: <Form.List name="mounts">{(fields, {add, remove}) => <>
                            {fields.map(({key, name}, index) => <div className="run-mount" key={key}>
                                <div className="run-heading"><span>挂载 {index + 1}</span>
                                    <div className="run-actions">
                                        <span>只读</span>
                                        <Form.Item name={[name, "read_only"]} valuePropName="checked" noStyle><Switch size="small" aria-label={`挂载 ${index + 1} 只读`}/></Form.Item>
                                        <Button color="danger" variant="text" size="small" disabled={submitting} onClick={() => remove(name)}>移除</Button>
                                    </div>
                                </div>
                                <div className="run-fields">
                                    <Form.Item name={[name, "source"]} label="服务器路径"
                                        tooltip="选择服务器上已创建的数据目录或文件。"
                                        rules={[{required: true, whitespace: true, message: "请选择已创建的服务器路径"}]}>
                                        <FilePicker placeholder="选择目录或输入文件路径" disabled={submitting} buttonProps={{color: "default", variant: "filled"}}/>
                                    </Form.Item>
                                    <Form.Item name={[name, "destination"]} label="容器路径" tooltip="软件在容器内读写数据的位置，以镜像说明为准。" rules={[
                                        {required: true, message: "请输入软件在容器内使用的路径"}, {pattern: /^\//, message: "请输入以 / 开头的绝对路径"},
                                    ]}><Input placeholder={image?.volumes?.[0] || "例如 /data"}/></Form.Item>
                                </div>
                            </div>)}
                            {image?.volumes?.length ? <Dropdown trigger={["click"]} disabled={submitting} menu={{items: [
                                ...image.volumes.map((path: string) => ({
                                    key: path, label: <span className={styles.mountPath}>{path}</span>,
                                    disabled: mounts.some((mount: any) => mount?.destination === path),
                                    onClick: () => add({source: "", destination: path, read_only: false}),
                                })),
                                {type: "divider"},
                                {key: "custom", label: "自定义路径", onClick: () => add({source: "", destination: "", read_only: false})},
                            ]}}>
                                <Button color="default" variant="filled" size="small" disabled={submitting}>添加挂载<Icon type="DownOutlined"/></Button>
                            </Dropdown> : <Button color="default" variant="filled" size="small" disabled={submitting}
                                onClick={() => add({source: "", destination: "", read_only: false})}>添加挂载</Button>}
                        </>}</Form.List>,
                    },
                    {
                        key: "advanced", forceRender: true, collapsible: submitting ? "disabled" : undefined,
                        label: <div className="run-section-heading"><span>更多设置</span><span className="run-muted">启动参数、运行用户与资源限制</span></div>,
                        children: <>
                            <div className="run-setting-group">
                                <div className="run-setting-title">运行设置</div>
                                <div className="run-fields run-runtime-fields">
                                    <Form.Item name="id" label="容器标识" tooltip="留空自动生成；自定义时使用字母、数字、点、横线或下划线，最多 64 位。" rules={[
                                        {pattern: /^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$/, message: "以字母或数字开头，支持 . _ -，最多 64 位"},
                                    ]}><Input placeholder="自动生成" maxLength={64} autoComplete="off"/></Form.Item>
                                    <Form.Item name="user" label="运行用户"><Input placeholder={image?.user || "镜像默认用户"}/></Form.Item>
                                    <Form.Item name="working_dir" label="工作目录" className="run-fields-wide" rules={[{pattern: /^\//, message: "请输入容器内绝对路径"}]}>
                                        <Input placeholder={image?.working_dir || "镜像默认目录"}/>
                                    </Form.Item>
                                    <Form.Item name="command" className="run-fields-wide run-command"
                                        label={<span className="run-field-label"><span>启动参数</span><span className="run-muted">每行一个，留空使用镜像默认值</span></span>}
                                        tooltip="覆盖镜像默认参数，并接在镜像入口程序之后；不执行 Shell 命令。">
                                        <Input.TextArea rows={2} spellCheck={false} placeholder={image?.command?.join("\n") || "使用镜像默认命令"}/>
                                    </Form.Item>
                                </div>
                            </div>
                            <div className="run-setting-group">
                                <div className="run-setting-title">资源限制</div>
                                <div className="run-fields">
                                    <Form.Item name="memory" label="内存上限" tooltip="留空或 0 表示不限制内存。">
                                        <InputNumber min={0} precision={0} placeholder="不限制" suffix="MB"/>
                                    </Form.Item>
                                    <Form.Item name="pids_limit" label="进程数上限" tooltip="留空或 0 使用容器管理器默认值。">
                                        <InputNumber min={0} precision={0} placeholder="系统默认"/>
                                    </Form.Item>
                                    <Form.Item name="cpu_quota" label="CPU 配额" tooltip="每个周期可使用的 CPU 时间，留空或 0 表示不限制。">
                                        <InputNumber min={0} precision={0} placeholder="不限制" suffix="μs"/>
                                    </Form.Item>
                                    <Form.Item name="cpu_period" label="CPU 周期" tooltip="默认 100000 微秒；配额与周期相等时，最多使用 1 个 CPU 核心。">
                                        <InputNumber min={0} precision={0} placeholder="100000" suffix="μs"/>
                                    </Form.Item>
                                </div>
                            </div>
                            <div className="run-setting-group run-access">
                                <Form.Item name="read_only" label="只读文件系统" layout="horizontal" valuePropName="checked"
                                    tooltip="开启后，仅可写的挂载目录允许写入数据。"><Switch size="small" aria-label="只读文件系统"/></Form.Item>
                            </div>
                        </>,
                    },
                ]}/>
            <div className="run-auto-start">
                <div className="run-auto-copy">
                    <label htmlFor="run-auto-restart">自动恢复</label>
                    <span id="run-auto-restart-help" className="run-auto-description">异常退出后自动重启，手动停止后暂停</span>
                </div>
                <Form.Item name="auto_restart" valuePropName="checked" noStyle>
                    <Switch id="run-auto-restart" size="small" aria-label="自动恢复" aria-describedby="run-auto-restart-help"/>
                </Form.Item>
            </div>
        </Form>
    </ProModal>;
});

RunContainer.displayName = "RunContainer";

export default memo(RunContainer);
