import React, {forwardRef, useCallback, useImperativeHandle, useRef, useState} from "react";
import {App, Form, Grid, Input} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import {updateFile} from "@/service/api/file";
import {formatFileMode} from "../../utils";

const FilePermissions = forwardRef<any, any>(({onSuccess}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const [form] = Form.useForm<any>();
    const generationRef = useRef(0);
    const submittingRef = useRef(false);
    const [state, setState] = useState<any>();
    const [open, setOpen] = useState(false);
    const [layered, setLayered] = useState(false);
    const [submitting, setSubmitting] = useState(false);

    const close = useCallback(() => {
        generationRef.current += 1;
        submittingRef.current = false;
        setSubmitting(false);
        setOpen(false);
    }, []);

    useImperativeHandle(ref, () => ({
        open: (path: string, record: any, nested = false) => {
            setLayered(nested);
            const generation = ++generationRef.current;
            submittingRef.current = false;
            setSubmitting(false);
            setState({generation, path});
            form.setFieldsValue({permissions: formatFileMode(record.mode)});
            setOpen(true);
            window.requestAnimationFrame(() => form.focusField("permissions"));
        },
        close,
    }), [close, form]);

    const submit = async ({permissions}: any) => {
        if (!open || !state || submittingRef.current) {
            return;
        }
        const generation = state.generation;
        submittingRef.current = true;
        setSubmitting(true);
        try {
            const response = await updateFile({body: {path: state.path, permissions}});
            const current = generationRef.current === generation;
            if (!response) {
                return;
            }
            if (response?.code !== 200) {
                if (current) {
                    message.error(response?.message || "修改权限失败");
                }
                return;
            }
            onSuccess(state.path, permissions);
            if (current) {
                message.success(response?.message || "权限修改成功");
                close();
            }
        } finally {
            if (generationRef.current === generation) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    };

    return (
        <ProModal
            title={<Title>文件权限</Title>}
            width="340px"
            okText="保存"
            onOk={() => form.submit()}
            onCancel={close}
            modalProps={{
                open,
                afterClose: () => {
                    if (state?.generation === generationRef.current) return;
                    setState(undefined);
                    form.resetFields();
                },
                zIndex: layered ? 2000 : undefined,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                cancelText: "取消",
                confirmLoading: submitting,
                cancelButtonProps: {disabled: false},
                closable: true,
                keyboard: true,
                focusable: {focusTriggerAfterClose: layered},
                mask: {closable: true},
                styles: {body: {paddingTop: compact ? 10 : 15}},
            }}>
            <Form<any>
                form={form}
                layout="vertical"
                variant="filled"
                preserve={false}
                requiredMark={false}
                onFinish={submit}>
                <Form.Item
                    name="permissions"
                    rules={[
                        {required: true, message: "请输入权限"},
                        {pattern: /^(?:0?[0-7]{3})$/, message: "请输入 000 到 0777 的八进制权限"},
                    ]}>
                    <Input maxLength={4} aria-label="权限" placeholder="请输入权限，例如 0644"/>
                </Form.Item>
            </Form>
        </ProModal>
    );
});

FilePermissions.displayName = "FilePermissions";

export default React.memo(FilePermissions);
