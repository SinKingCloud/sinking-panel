import React, {forwardRef, useCallback, useImperativeHandle, useRef, useState} from "react";
import {App, Form, Grid, Input} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import {updateFile} from "@/service/api/file";
import {formatFileMode} from "../utils";

interface PermissionsState {
    generation: number;
    path: string;
}

interface PermissionsValues {
    permissions: string;
}

export interface FilePermissionsRef {
    open: (path: string, record: any, layered?: boolean) => void;
    close: () => void;
}

interface FilePermissionsProps {
    onSuccess: (path: string, permissions: string) => void;
}

const FilePermissions = forwardRef<FilePermissionsRef, FilePermissionsProps>(({onSuccess}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const [form] = Form.useForm<PermissionsValues>();
    const generationRef = useRef(0);
    const submittingRef = useRef(false);
    const [state, setState] = useState<PermissionsState>();
    const [layered, setLayered] = useState(false);
    const [submitting, setSubmitting] = useState(false);

    const close = useCallback(() => {
        setState(undefined);
        form.resetFields();
    }, [form]);

    useImperativeHandle(ref, () => ({
        open: (path, record, nested = false) => {
            setLayered(nested);
            const generation = ++generationRef.current;
            setState({generation, path});
            form.setFieldsValue({permissions: formatFileMode(record.mode)});
            window.requestAnimationFrame(() => form.focusField("permissions"));
        },
        close,
    }), [close, form]);

    const submit = async ({permissions}: PermissionsValues) => {
        if (!state || submittingRef.current) {
            return;
        }
        const generation = state.generation;
        submittingRef.current = true;
        setSubmitting(true);
        try {
            const response = await updateFile({body: {path: state.path, permissions}});
            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || "修改权限失败");
                return;
            }
            message.success(response.message || "权限修改成功");
            onSuccess(state.path, permissions);
            if (generationRef.current === generation) {
                close();
            }
        } finally {
            submittingRef.current = false;
            setSubmitting(false);
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
                open: Boolean(state),
                forceRender: true,
                zIndex: layered ? 2000 : undefined,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                cancelText: "取消",
                confirmLoading: submitting,
                focusable: {focusTriggerAfterClose: layered},
                mask: {closable: true},
                styles: {body: {paddingTop: compact ? 10 : 15}},
            }}>
            <Form<PermissionsValues>
                form={form}
                layout="vertical"
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
