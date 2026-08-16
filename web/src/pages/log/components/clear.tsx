import React, {forwardRef, useImperativeHandle, useRef, useState} from "react";
import {App, Form, InputNumber} from "antd";
import {ProModal, ProModalRef, Title} from "sinking-antd";
import {getLog} from "@/service/api/system";

export interface ClearRef {
    open: () => void;
}

interface ClearProps {
    onSuccess?: () => void;
}

const Clear = forwardRef<ClearRef, ClearProps>(({onSuccess}, ref) => {
    const {message} = App.useApp();
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const [form] = Form.useForm<{day: number}>();
    const [submitting, setSubmitting] = useState(false);

    useImperativeHandle(ref, () => ({
        open: () => {
            form.resetFields();
            form.setFieldsValue({day: 30});
            modalRef.current?.show();
        },
    }), [form]);

    const submit = async (values: {day: number}) => {
        if (submitting) {
            return;
        }
        setSubmitting(true);
        try {
            const response = await getLog({body: {action: "clear", day: String(values.day)}});
            if (response?.code !== 200) {
                message.error(response?.message || "清理失败");
                return;
            }
            modalRef.current?.hide();
            onSuccess?.();
            message.success(response.message || "清理成功");
        } catch {
            message.error("清理失败");
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <ProModal
            ref={modalRef}
            title={<Title>清理操作日志</Title>}
            width={320}
            onOk={form.submit}
            modalProps={{
                confirmLoading: submitting,
                okText: "提交",
                cancelText: "取消",
                forceRender: true,
                style: {top: 100, paddingBottom: 100},
                mask: {closable: true},
                afterClose: () => {
                    form.resetFields();
                    setSubmitting(false);
                },
            } as any}
        >
            <Form form={form} layout="vertical" onFinish={submit}>
                <Form.Item
                    name="day"
                    label="保留天数"
                    rules={[{required: true, message: "请输入保留天数"}]}
                >
                    <InputNumber
                        min={1}
                        precision={0}
                        step={1}
                        style={{width: "100%"}}
                        placeholder="请输入保留天数"
                    />
                </Form.Item>
            </Form>
        </ProModal>
    );
});

Clear.displayName = "Clear";

export default Clear;
