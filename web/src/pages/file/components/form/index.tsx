import React, {forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {App, Form as AntForm, Grid, Input} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import {createFile, renameFile} from "@/service/api/file";
import {joinFilePath, parentFilePath} from "../../utils";

interface EditorState {
    generation: number;
    mode: any;
    path: string;
    targetPath?: string;
    record?: any;
}

export interface FileFormRef {
    open: (mode: any, path: string, record?: any, layered?: boolean) => void;
    openRename: (path: string, record: any, layered?: boolean) => void;
    close: () => void;
}

interface FileFormProps {
    onSuccess: (result: any) => void;
}

interface FormValues {
    name: string;
}

const FileForm = forwardRef<FileFormRef, FileFormProps>(({onSuccess}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const [form] = AntForm.useForm<FormValues>();
    const generationRef = useRef(0);
    const submittingRef = useRef(false);
    const [editor, setEditor] = useState<EditorState>();
    const [layered, setLayered] = useState(false);
    const [submitting, setSubmitting] = useState(false);

    useEffect(() => () => {
        generationRef.current += 1;
        submittingRef.current = false;
    }, []);

    const close = useCallback(() => {
        generationRef.current += 1;
        submittingRef.current = false;
        setSubmitting(false);
        setEditor(undefined);
        form.resetFields();
    }, [form]);

    useImperativeHandle(ref, () => ({
        open: (mode, path, record, nested = false) => {
            if (submittingRef.current) {
                return;
            }
            setLayered(nested);
            const generation = ++generationRef.current;
            setEditor({generation, mode, path, record});
            form.setFieldsValue({name: mode === "rename" ? record?.name || "" : ""});
            window.requestAnimationFrame(() => form.focusField("name"));
        },
        openRename: (path, record, nested = false) => {
            if (submittingRef.current) {
                return;
            }
            setLayered(nested);
            const generation = ++generationRef.current;
            setEditor({generation, mode: "rename", path, targetPath: path, record});
            form.setFieldsValue({name: record.name});
            window.requestAnimationFrame(() => form.focusField("name"));
        },
        close,
    }), [close, form]);

    const submit = async ({name}: FormValues) => {
        if (!editor || submittingRef.current) {
            return;
        }
        const generation = editor.generation;
        const value = name.trim();
        if (editor.mode === "rename" && value === editor.record?.name) {
            close();
            return;
        }
        submittingRef.current = true;
        setSubmitting(true);
        try {
            const sourcePath = editor.mode === "rename" && editor.record
                ? editor.targetPath || joinFilePath(editor.path, editor.record.name)
                : undefined;
            const response = editor.mode === "rename" && editor.record
                ? await renameFile({
                    body: {
                        path: sourcePath || "",
                        name: value,
                    },
                })
                : await createFile({
                    body: {
                        path: editor.path,
                        name: editor.mode === "directory" ? `${value}/` : value,
                    },
                });
            if (generationRef.current !== generation) {
                return;
            }
            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || (editor.mode === "rename" ? "重命名失败" : "创建失败"));
                return;
            }
            message.success(response.message || (editor.mode === "rename" ? "重命名成功" : "创建成功"));
            onSuccess({
                mode: editor.mode,
                name: value,
                sourcePath,
                targetPath: sourcePath ? joinFilePath(parentFilePath(sourcePath), value) : undefined,
            });
            if (generationRef.current === generation) {
                close();
            }
        } finally {
            if (generationRef.current === generation) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    };

    const title = editor?.mode === "rename"
        ? "重命名"
        : editor?.mode === "directory" ? "新建文件夹" : "新建空文件";
    const placeholder = editor?.mode === "rename"
        ? "请输入新名称"
        : editor?.mode === "directory" ? "请输入文件夹名称" : "请输入文件名称";

    return (
        <ProModal
            title={<Title>{title}</Title>}
            width="340px"
            okText={editor?.mode === "rename" ? "保存" : "创建"}
            onOk={() => form.submit()}
            onCancel={() => close()}
            modalProps={{
                open: Boolean(editor),
                forceRender: true,
                zIndex: layered ? 2000 : undefined,
                closable: true,
                keyboard: true,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                cancelText: "取消",
                confirmLoading: submitting,
                focusable: {focusTriggerAfterClose: layered},
                mask: {closable: true},
                cancelButtonProps: {disabled: false},
                styles: {body: {paddingTop: compact ? 10 : 15}},
            }}>
            <AntForm<FormValues>
                form={form}
                layout="vertical"
                preserve={false}
                requiredMark={false}
                onFinish={submit}>
                <AntForm.Item
                    name="name"
                    rules={[
                        {required: true, whitespace: true, message: "请输入名称"},
                        {max: 255, message: "名称不能超过255个字符"},
                        {
                            validator: async (_, value) => {
                                const name = String(value || "").trim();
                                if (name === "." || name === ".." || /[\\/\0]/.test(name)) {
                                    throw new Error("名称不能包含路径分隔符");
                                }
                            },
                        },
                    ]}>
                    <Input maxLength={255} aria-label={placeholder} placeholder={placeholder}/>
                </AntForm.Item>
            </AntForm>
        </ProModal>
    );
});

FileForm.displayName = "FileForm";

export default React.memo(FileForm);
