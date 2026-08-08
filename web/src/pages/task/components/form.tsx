import {forwardRef, memo, useCallback, useImperativeHandle, useRef, useState} from "react";
import {App, Col, Form as AntForm, Input, Row, Select, Spin} from "antd";
import {createStyles} from "antd-style";
import {ProModal, ProModalRef, Title} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import AceEditor from "@/components/ace-editor";
import {createTask, getTaskInfo, updateTask} from "@/service/api/task";
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

const Form = forwardRef<FormRef, {onSuccess?: () => void}>(({onSuccess}, ref) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const requestRef = useRef(0);
    const [form] = AntForm.useForm();
    const [taskId, setTaskId] = useState<any>();
    const [active, setActive] = useState(false);
    const [infoLoading, setInfoLoading] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const editing = taskId !== undefined && taskId !== null;

    const reset = useCallback(() => {
        requestRef.current += 1;
        form.resetFields();
        setTaskId(undefined);
        setActive(false);
        setInfoLoading(false);
        setSubmitting(false);
    }, [form]);

    const openCreate = useCallback(() => {
        requestRef.current += 1;
        setTaskId(undefined);
        setActive(true);
        setInfoLoading(false);
        form.resetFields();
        form.setFieldsValue({
            name: "",
            spec: "0 */5 * * * *",
            script: "",
            status: 0,
        });
        modalRef.current?.show();
    }, [form]);

    const openEdit = useCallback((record: any) => {
        if (record?.id === undefined || record?.id === null) {
            return;
        }
        const requestId = ++requestRef.current;
        setTaskId(record.id);
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
            form.setFieldsValue({
                name: data.name,
                spec: data.spec,
                script: data.script || "",
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
    }, [form, message]);

    useImperativeHandle(ref, () => ({
        open: (record?: any) => record ? openEdit(record) : openCreate(),
    }), [openCreate, openEdit]);

    const submit = useCallback(async (values: any) => {
        const requestId = ++requestRef.current;
        const body: any = {
            name: String(values.name || "").trim(),
            type: 0,
            spec: String(values.spec || "").trim(),
            script: String(values.script || ""),
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
                            <Col xs={24} sm={16}>
                                <AntForm.Item
                                    name="name"
                                    label="任务名称"
                                    rules={[{required: true, whitespace: true, message: "请输入任务名称"}]}>
                                    <Input placeholder="请输入任务名称"/>
                                </AntForm.Item>
                            </Col>
                            <Col xs={24} sm={8}>
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
                    </>
                ))}
            </AntForm>
        </ProModal>
    );
});

Form.displayName = "Form";

export default memo(Form);
