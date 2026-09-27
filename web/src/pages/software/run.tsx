import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Button, Col, Collapse, Form, Grid, Input, InputNumber, Row, Select, Spin, Switch} from "antd";
import {createStyles} from "antd-style";
import {ProModal, Title, useTheme} from "sinking-antd";
import FilePicker from "@/pages/components/file-picker";
import {getContainerImage, getContainerImages, getContainerList, runContainer} from "@/service/api/container";

export interface RunContainerRef {
    open: (image?: any) => void;
}

interface RunValues {
    id: string;
    name?: string;
    image_id: string;
    mounts?: {source: string; destination: string; read_only?: boolean}[];
    command?: string;
    env?: string;
    working_dir?: string;
    user?: string;
    memory?: number;
    cpu_quota?: number;
    cpu_period?: number;
    pids_limit?: number;
    auto_restart: boolean;
    read_only?: boolean;
}

const useStyles = createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    modal: css`
        .ant-modal { max-width: calc(100vw - 24px); }
        .ant-modal-body { padding-top: ${compact ? 12 : 16}px; overflow-x: hidden; }
    `,
    form: css`
        .ant-form-item { margin-bottom: ${compact ? 14 : 18}px; }
        .ant-form-item-extra { margin-top: 4px; min-height: 0; font-size: 12px; line-height: 1.6; color: ${token.colorTextTertiary}; }
        .ant-input-number { width: 100%; }
        .ant-collapse-header { padding-inline: 0 !important; }
        .ant-collapse-body { padding-inline: 0 !important; }
    `,
    hint: css`
        margin: 0 0 ${compact ? 14 : 18}px;
        font-size: 12px;
        line-height: 1.6;
        color: ${token.colorTextTertiary};
        overflow-wrap: anywhere;
    `,
    mount: css`
        margin-bottom: 10px;
        padding: ${compact ? 10 : 12}px;
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorFillQuaternary};
        .ant-form-item { margin-bottom: 10px; }
    `,
    line: css`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        font-size: ${token.fontSize}px;
    `,
    option: css`
        display: flex;
        align-items: center;
        gap: 6px;
    `,
}));

const RunContainer = forwardRef<RunContainerRef, {onChange?: () => void}>(({onChange}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const [form] = Form.useForm<RunValues>();
    const sessionRef = useRef(0);
    const listRequestRef = useRef(0);
    const imageRequestRef = useRef(0);
    const submittingRef = useRef(false);
    const searchTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const [open, setOpen] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [images, setImages] = useState<any[]>([]);
    const [selectedImage, setSelectedImage] = useState<any>();
    const [image, setImage] = useState<any>();
    const [imageLoading, setImageLoading] = useState(false);
    const [listLoading, setListLoading] = useState(false);
    const [listError, setListError] = useState("");
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(false);
    const [keyword, setKeyword] = useState("");
    const [advanced, setAdvanced] = useState<string[]>([]);

    const loadImages = useCallback(async (nextPage: number, search: string) => {
        const session = sessionRef.current;
        const requestId = ++listRequestRef.current;
        setListLoading(true);
        setListError("");
        try {
            const response = await getContainerImages({body: {page: nextPage, page_size: 20, keyword: search}});
            if (sessionRef.current !== session || listRequestRef.current !== requestId) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "获取镜像失败");
            const list = Array.isArray(response.data.list) ? response.data.list : [];
            setImages((current) => Array.from(new Map(
                [...(nextPage > 1 ? current : []), ...list].map((item) => [item.id, item]),
            ).values()));
            setPage(nextPage);
            setHasMore(nextPage * 20 < Number(response.data.total || 0));
        } catch (error) {
            if (sessionRef.current === session && listRequestRef.current === requestId) {
                setListError(error instanceof Error ? error.message : "获取镜像失败");
            }
        } finally {
            if (sessionRef.current === session && listRequestRef.current === requestId) setListLoading(false);
        }
    }, []);

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
        open: (selected) => {
            sessionRef.current += 1;
            listRequestRef.current += 1;
            imageRequestRef.current += 1;
            clearTimeout(searchTimerRef.current);
            submittingRef.current = false;
            form.resetFields();
            form.setFieldsValue({image_id: selected?.id, mounts: [], auto_restart: true});
            setSubmitting(false);
            setImages([]);
            setSelectedImage(selected);
            setImage(undefined);
            setImageLoading(false);
            setPage(1);
            setHasMore(false);
            setKeyword("");
            setAdvanced([]);
            setOpen(true);
            void loadImages(1, "");
            if (selected) void selectImage(selected.id);
        },
    }), [form, loadImages, selectImage]);

    useEffect(() => () => {
        sessionRef.current += 1;
        clearTimeout(searchTimerRef.current);
    }, []);

    const close = useCallback(() => {
        if (submittingRef.current) return;
        sessionRef.current += 1;
        clearTimeout(searchTimerRef.current);
        setOpen(false);
    }, []);

    const submit = useCallback(async (values: RunValues) => {
        if (submittingRef.current) return;
        const session = sessionRef.current;
        submittingRef.current = true;
        setSubmitting(true);
        let submitted = false;
        try {
            const existing = await getContainerList({body: {id: values.id.trim(), page: 1, page_size: 1}});
            if (sessionRef.current !== session) return;
            if (existing?.code !== 200 || !existing.data) throw new Error(existing?.message || "检查实例标识失败");
            if (existing.data.list?.length) throw new Error("容器标识已存在，请更换");
            submitted = true;
            const response = await runContainer({body: {
                id: values.id.trim(),
                name: values.name?.trim() || "",
                image_id: values.image_id,
                mounts: (values.mounts || []).map((mount) => ({
                    source: mount.source.trim(), destination: mount.destination.trim(), read_only: Boolean(mount.read_only),
                })),
                // 参数按行原样传递，保留参数内的空格；不做 shell 拆词或执行。
                command: (values.command || "").split(/\r?\n/).filter((line) => line.length > 0),
                env: (values.env || "").split(/\r?\n/).filter((line) => line.length > 0),
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
            if (response?.code !== 200) throw new Error(response?.message || "创建软件失败");
            message.success(response.message || "创建软件成功");
            submittingRef.current = false;
            setSubmitting(false);
            sessionRef.current += 1;
            setOpen(false);
        } catch (error) {
            if (sessionRef.current === session) message.error(error instanceof Error ? error.message : "创建软件失败");
        } finally {
            if (sessionRef.current === session) {
                submittingRef.current = false;
                setSubmitting(false);
            }
            // 启动失败也可能生成实例；列表刷新失败不影响创建结果。
            if (submitted) void Promise.resolve().then(() => onChange?.()).catch(() => undefined);
        }
    }, [message, onChange]);

    const imageOptions = Array.from(new Map([...images, ...(selectedImage ? [selectedImage] : [])].map((item) => [item.id, item])).values())
        .map((item) => ({value: item.id, label: item.name || item.tags?.[0] || item.id}));

    return <ProModal title={<Title>添加软件</Title>} width={660} onOk={() => form.submit()} onCancel={close}
        okText="创建" modalProps={{
            open, forceRender: true, rootClassName: styles.modal,
            style: {top: screens.md ? 64 : 24, paddingBottom: screens.md ? 64 : 24},
            closable: !submitting, keyboard: !submitting, mask: {closable: !submitting},
            confirmLoading: submitting, cancelText: "取消", cancelButtonProps: {disabled: submitting},
            okButtonProps: {disabled: imageLoading},
        } as any}>
        <Form form={form} layout="vertical" variant="filled" className={styles.form} disabled={submitting}
            initialValues={{auto_restart: true, mounts: []}} onFinish={submit}>
            <Form.Item name="image_id" label="运行镜像" rules={[{required: true, message: "请选择已导入的镜像"}]}>
                <Select showSearch filterOption={false} options={imageOptions} loading={listLoading}
                    placeholder="选择已导入的镜像" onSelect={(value) => {
                        setSelectedImage(images.find((item) => item.id === value) || (selectedImage?.id === value ? selectedImage : undefined));
                        void selectImage(value);
                    }}
                    onSearch={(value) => {
                        setKeyword(value);
                        setPage(1);
                        setHasMore(false);
                        setImages([]);
                        setListLoading(true);
                        listRequestRef.current += 1;
                        clearTimeout(searchTimerRef.current);
                        searchTimerRef.current = setTimeout(() => void loadImages(1, value), 250);
                    }}
                    notFoundContent={listLoading ? <Spin size="small"/> : "暂无镜像，请先导入"}
                    popupRender={(menu) => <>
                        {menu}
                        {(hasMore || listError) && <Button block type="text" size="small" loading={listLoading}
                            onMouseDown={(event) => event.preventDefault()}
                            onClick={() => void loadImages(listError && !images.length ? 1 : page + 1, keyword)}>
                            {listError ? "加载失败，重试" : "加载更多"}
                        </Button>}
                    </>}/>
            </Form.Item>
            <div className={styles.hint}>
                使用宿主机网络，软件直接占用宿主机端口。
                {Boolean(image?.exposed_ports?.length) && <>镜像端口：{image!.exposed_ports!.join("、")}。</>}
            </div>
            <Row gutter={16}>
                <Col xs={24} sm={12}><Form.Item name="id" label="实例标识" rules={[
                    {required: true, whitespace: true, message: "请输入实例标识"},
                    {pattern: /^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$/, message: "以字母或数字开头，可用点、横线和下划线，最多 64 位"},
                ]}><Input placeholder="例如 redis" maxLength={64}/></Form.Item></Col>
                <Col xs={24} sm={12}><Form.Item name="name" label="软件名称"><Input placeholder="留空使用镜像名称" maxLength={100}/></Form.Item></Col>
            </Row>
            <Form.List name="mounts">{(fields, {add, remove}) => <>
                <div className={styles.line} style={{marginBottom: 8}}>
                    <span>持久化目录</span>
                    <Button color="default" variant="filled" size="small" disabled={submitting}
                        onClick={() => add({source: "", destination: "", read_only: false})}>添加</Button>
                </div>
                <div className={styles.hint}>
                    挂载已有目录以保留数据；宿主目录需先在文件管理中创建。
                </div>
                {fields.map(({key, name}) => <div className={styles.mount} key={key}>
                    <Form.Item name={[name, "source"]} label="宿主目录" rules={[{required: true, whitespace: true, message: "请选择宿主目录"}]}>
                        <FilePicker placeholder="选择已有目录，也可输入文件路径"
                            disabled={submitting} buttonProps={{color: "default", variant: "filled"}}/>
                    </Form.Item>
                    <Form.Item name={[name, "destination"]} label="容器路径" rules={[
                        {required: true, message: "请输入容器路径"}, {pattern: /^\//, message: "请输入以 / 开头的绝对路径"},
                    ]}><Input placeholder="例如 /data"/></Form.Item>
                    <div className={styles.line}>
                        <div className={styles.option}><Form.Item name={[name, "read_only"]} valuePropName="checked" noStyle><Switch size="small"/></Form.Item><span>只读</span></div>
                        <Button color="danger" variant="text" size="small" onClick={() => remove(name)}>移除</Button>
                    </div>
                </div>)}
            </>}</Form.List>
            <Collapse ghost activeKey={advanced} onChange={(keys) => setAdvanced(Array.isArray(keys) ? keys : [keys])}
                items={[{key: "advanced", label: "高级设置", children: <>
                    <Form.Item name="command" label="启动参数" extra="每行一个参数，无需添加引号；留空使用镜像默认命令。">
                        <Input.TextArea rows={3} placeholder={image?.command?.join("\n") || "使用镜像默认命令"}/>
                    </Form.Item>
                    <Form.Item name="env" label="环境变量" extra="每行一个 KEY=VALUE，同名变量会覆盖镜像默认值。" rules={[{
                        validator: async (_, value) => {
                            if (String(value || "").split(/\r?\n/).some((line) => line.length > 0 && (line.indexOf("=") < 1 || line.includes("\0")))) {
                                throw new Error("环境变量请按 KEY=VALUE 填写");
                            }
                        },
                    }]}><Input.TextArea rows={3} placeholder="例如 MYSQL_ROOT_PASSWORD=你的密码"/></Form.Item>
                    <Row gutter={16}>
                        <Col xs={24} sm={12}><Form.Item name="working_dir" label="工作目录" rules={[{pattern: /^\//, message: "请输入容器内绝对路径"}]}>
                            <Input placeholder={image?.working_dir || "使用镜像默认目录"}/>
                        </Form.Item></Col>
                        <Col xs={24} sm={12}><Form.Item name="user" label="运行用户"><Input placeholder={image?.user || "使用镜像默认用户"}/></Form.Item></Col>
                        <Col xs={24} sm={12}><Form.Item name="memory" label="内存上限（MB）" extra="0 表示不限制"><InputNumber min={0} precision={0} placeholder="0"/></Form.Item></Col>
                        <Col xs={24} sm={12}><Form.Item name="pids_limit" label="进程数上限" extra="0 使用容器管理器默认值"><InputNumber min={0} precision={0} placeholder="0"/></Form.Item></Col>
                        <Col xs={24} sm={12}><Form.Item name="cpu_quota" label="CPU 配额（微秒）" extra="每个周期可使用的 CPU 时间，0 不限制"><InputNumber min={0} precision={0} placeholder="0"/></Form.Item></Col>
                        <Col xs={24} sm={12}><Form.Item name="cpu_period" label="CPU 周期（微秒）" extra="填写配额时，默认周期为 100000"><InputNumber min={0} precision={0} placeholder="100000"/></Form.Item></Col>
                    </Row>
                    <Form.Item name="read_only" label="只读根文件系统" valuePropName="checked" extra="启用后程序只能向可写的挂载目录写入数据。"><Switch size="small"/></Form.Item>
                </>}]} />
            <div className={styles.line} style={{marginTop: 12}}>
                <div>自动启动<div className={styles.hint} style={{marginBottom: 0}}>随面板启动，并在异常退出后重启；手动停止后暂停。</div></div>
                <Form.Item name="auto_restart" valuePropName="checked" noStyle><Switch size="small"/></Form.Item>
            </div>
        </Form>
    </ProModal>;
});

RunContainer.displayName = "RunContainer";

export default memo(RunContainer);
