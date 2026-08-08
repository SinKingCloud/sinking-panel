import {forwardRef, memo, useCallback, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Col, Form as AntForm, Input, Row, Select, Spin} from "antd";
import {createStyles} from "antd-style";
import {ProModal, ProModalRef, Title} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import AceEditor from "@/components/ace-editor";
import {createTask, getTaskInfo, updateTask} from "@/service/api/task";
import {parseRequest, stringifyRequest} from "../utils";
import Request from "./request";
import Schedule from "./schedule";

const acePath = `${defaultSettings?.basePath || "/"}ace`;
const statusOptions = [
    {label: "运行", value: 0},
    {label: "暂停", value: 1},
];

const useStyles = createStyles(({css, token}: any) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }
    `,
    editor: css`
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;

        .ace_editor {
            border-radius: ${token.borderRadius}px !important;
        }
    `,
    loading: css`
        height: 380px;
        display: flex;
        align-items: center;
        justify-content: center;
    `,
}));

export interface FormRef {
    open: (record?: any) => void;
}

const Form = forwardRef<FormRef, {typeData?: any; onSuccess?: () => void}>(({
    typeData,
    onSuccess,
}, ref) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const requestRef = useRef(0);
    const [form] = AntForm.useForm();
    const [taskId, setTaskId] = useState<any>();
    const [taskType, setTaskType] = useState(0);
    const [active, setActive] = useState(false);
    const [infoLoading, setInfoLoading] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const editing = taskId !== undefined && taskId !== null;
    const typeOptions = useMemo(() => {
        return Object.entries(typeData || {}).map(([value, label]) => ({
            label: String(label),
            value: Number(value),
        }));
    }, [typeData]);

    const reset = useCallback(() => {
        requestRef.current += 1;
        form.resetFields();
        setTaskId(undefined);
        setTaskType(0);
        setActive(false);
        setInfoLoading(false);
        setSubmitting(false);
    }, [form]);

    const openCreate = useCallback(() => {
        const type = Number(typeOptions[0]?.value);
        if (!Number.isFinite(type)) {
            message.error("任务类型尚未加载");
            return;
        }
        requestRef.current += 1;
        setTaskId(undefined);
        setTaskType(type);
        setActive(true);
        setInfoLoading(false);
        form.resetFields();
        form.setFieldsValue({
            name: "",
            type,
            spec: "0 */5 * * * *",
            script: "",
            request: parseRequest({method: "GET"}),
            status: 0,
        });
        modalRef.current?.show();
    }, [form, message, typeOptions]);

    const openEdit = useCallback((record: any) => {
        if (record?.id === undefined || record?.id === null) {
            return;
        }
        if (typeOptions.length === 0) {
            message.error("任务类型尚未加载");
            return;
        }
        const requestId = ++requestRef.current;
        setTaskId(record.id);
        setTaskType(Number(record.type || 0));
        setActive(true);
        setInfoLoading(true);
        form.resetFields();
        modalRef.current?.show();

        void getTaskInfo({body: {id: record.id}}).then((response) => {
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                modalRef.current?.hide();
                message.error(response?.message || "获取任务详情失败");
                return;
            }
            const data: any = response?.data || {};
            const type = Number(data.type || 0);
            setTaskType(type);
            form.setFieldsValue({
                name: data.name,
                type,
                spec: data.spec,
                script: type === 0 ? String(data.script || "") : "",
                request: type === 1 ? parseRequest(data.script) : undefined,
                status: Number(data.status),
            });
        }).catch(() => {
            if (requestRef.current === requestId) {
                modalRef.current?.hide();
                message.error("获取任务详情失败");
            }
        }).finally(() => {
            if (requestRef.current === requestId) {
                setInfoLoading(false);
            }
        });
    }, [form, message, typeOptions.length]);

    useImperativeHandle(ref, () => ({
        open: (record?: any) => record ? openEdit(record) : openCreate(),
    }), [openCreate, openEdit]);

    const submit = useCallback(async (values: any) => {
        const requestId = ++requestRef.current;
        const type = Number(values.type);
        let script = String(values.script || "");
        if (type === 1) {
            try {
                script = stringifyRequest(values.request);
            } catch (error: any) {
                message.error(error?.message || "请求配置不合法");
                return;
            }
        }
        const body: any = {
            name: String(values.name || "").trim(),
            type,
            spec: String(values.spec || "").trim(),
            script,
            status: Number(values.status),
        };
        if (editing) {
            body.ids = [taskId];
        }

        setSubmitting(true);
        try {
            const response = await (editing ? updateTask : createTask)({body});
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "操作失败");
                return;
            }
            modalRef.current?.hide();
            message.success(response?.message || "操作成功");
            onSuccess?.();
        } catch {
            if (requestRef.current === requestId) {
                message.error("操作失败");
            }
        } finally {
            if (requestRef.current === requestId) {
                setSubmitting(false);
            }
        }
    }, [editing, message, onSuccess, taskId]);

    return (
        <ProModal
            ref={modalRef}
            title={<Title>{editing ? "编辑任务" : "添加任务"}</Title>}
            onOk={form.submit}
            width={680}
            modalProps={{
                rootClassName: styles.modal,
                confirmLoading: submitting || infoLoading,
                forceRender: true,
                okText: editing ? "保存" : "创建",
                cancelText: "取消",
                footer: infoLoading ? null : undefined,
                style: {top: 100, paddingBottom: 100},
                mask: {closable: true},
                afterClose: reset,
            } as any}>
            <AntForm form={form} layout="vertical" onFinish={submit}>
                {active && (infoLoading ? (
                    <div className={styles.loading}>
                        <Spin description="加载任务详情..."/>
                    </div>
                ) : (
                    <>
                        <Row gutter={[16, 0]}>
                            <Col xs={24} sm={12}>
                                <AntForm.Item
                                    name="name"
                                    label="任务名称"
                                    rules={[{required: true, whitespace: true, message: "请输入任务名称"}]}>
                                    <Input placeholder="请输入任务名称"/>
                                </AntForm.Item>
                            </Col>
                            <Col xs={24} sm={6}>
                                <AntForm.Item
                                    name="type"
                                    label="任务类型"
                                    rules={[{required: true, message: "请选择任务类型"}]}>
                                    <Select
                                        options={typeOptions}
                                        onChange={(value) => {
                                            const type = Number(value);
                                            setTaskType(type);
                                            form.setFieldsValue({
                                                script: "",
                                                request: type === 1 ? parseRequest({method: "GET"}) : undefined,
                                            });
                                        }}/>
                                </AntForm.Item>
                            </Col>
                            <Col xs={24} sm={6}>
                                <AntForm.Item
                                    name="status"
                                    label="任务状态"
                                    rules={[{required: true, message: "请选择任务状态"}]}>
                                    <Select options={statusOptions}/>
                                </AntForm.Item>
                            </Col>
                        </Row>
                        <AntForm.Item
                            name="spec"
                            label="执行周期"
                            rules={[{required: true, whitespace: true, message: "请设置执行周期"}]}>
                            <Schedule/>
                        </AntForm.Item>
                        {taskType === 0 ? (
                            <AntForm.Item
                                name="script"
                                label="任务内容"
                                rules={[{required: true, whitespace: true, message: "请输入任务内容"}]}>
                                <AceEditor
                                    mode="sh"
                                    width="100%"
                                    height={280}
                                    fontSize={14}
                                    showPrintMargin={false}
                                    wrapEnabled
                                    acePath={acePath}
                                    className={styles.editor}/>
                            </AntForm.Item>
                        ) : taskType === 1 ? (
                            <Request/>
                        ) : (
                            <AntForm.Item
                                name="script"
                                label="任务内容"
                                rules={[{required: true, whitespace: true, message: "请输入任务内容"}]}>
                                <Input placeholder="请输入任务内容" maxLength={2000}/>
                            </AntForm.Item>
                        )}
                    </>
                ))}
            </AntForm>
        </ProModal>
    );
});

Form.displayName = "Form";

export default memo(Form);
