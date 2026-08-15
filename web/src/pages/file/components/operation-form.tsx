import React, {forwardRef, memo, useCallback, useImperativeHandle, useRef, useState} from "react";
import {App, Form as AntForm, Grid} from "antd";
import type {FormInstance} from "antd";
import {createStyles} from "antd-style";
import {ProModal, Title} from "sinking-antd";
import {
    compressFile,
    extractFile,
    remoteDownloadFile,
} from "@/service/api/file";
import type {FileOperationContext, FileOperationMode} from "../types";
import {joinFilePath, normalizeFilePath} from "../utils";
import {
    CompressOperationFields,
    ExtractOperationFields,
    RemoteDownloadOperationFields,
} from "./operation-fields";
import type {FileOperationFormValues} from "./operation-fields";
import {defaultArchiveName, getRemoteFileName} from "./operation-form.utils";

export interface FileOperationRef {
    open: (mode: FileOperationMode, context: FileOperationContext) => void;
    close: () => void;
}

interface FileOperationProps {
    onTask: (taskId: string, title: string) => void;
    onSuccess?: (context: FileOperationContext) => void;
}

interface OperationState {
    generation: number;
    mode: FileOperationMode;
    context: FileOperationContext;
}

interface PendingConfirm {
    destroy: () => void;
    resolve: (confirmed: boolean) => void;
}

const modeText: Record<FileOperationMode, {title: string; action: string}> = {
    compress: {title: "压缩文件", action: "压缩"},
    extract: {title: "解压文件", action: "解压"},
    "remote-download": {title: "远程下载", action: "下载"},
};

const useStyles = createStyles(({css}) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
        }

        @media (max-width: 520px) {
            .ant-modal {
                margin: 0 auto;
            }
        }
    `,
}));

const FileOperationForm = forwardRef<FileOperationRef, FileOperationProps>(({onTask, onSuccess}, ref) => {
    const {message, modal} = App.useApp();
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles();
    const formRef = useRef<FormInstance<FileOperationFormValues> | null>(null);
    const generationRef = useRef(0);
    const submittingGenerationRef = useRef<number | undefined>(undefined);
    const confirmRef = useRef<PendingConfirm | undefined>(undefined);
    const [operation, setOperation] = useState<OperationState>();
    const [submitting, setSubmitting] = useState(false);

    const cancelPendingConfirm = useCallback(() => {
        const pending = confirmRef.current;
        confirmRef.current = undefined;
        pending?.resolve(false);
        pending?.destroy();
    }, []);

    const close = useCallback(() => {
        generationRef.current += 1;
        cancelPendingConfirm();
        setOperation(undefined);
        setSubmitting(false);
    }, [cancelPendingConfirm]);

    const open = useCallback((mode: FileOperationMode, context: FileOperationContext) => {
        const records = Array.isArray(context.records) ? [...context.records] : [];
        if (mode === "compress" && records.length === 0) {
            message.warning(`请选择需要${modeText[mode].action}的文件或目录`);
            return;
        }
        if (mode === "extract" && records.length !== 1) {
            message.warning("解压操作只能选择一个文件");
            return;
        }
        if (mode === "extract" && records[0]?.is_dir) {
            message.warning("请选择需要解压的文件");
            return;
        }
        const generation = ++generationRef.current;
        cancelPendingConfirm();
        setOperation({
            generation,
            mode,
            context: {
                path: context.path,
                records,
                clearSelectionOnSuccess: context.clearSelectionOnSuccess,
            },
        });
        setSubmitting(false);
    }, [cancelPendingConfirm, message]);

    useImperativeHandle(ref, () => ({open, close}), [close, open]);

    const confirmExtract = useCallback(() => {
        cancelPendingConfirm();
        return new Promise<boolean>((resolve) => {
            let settled = false;
            const finish = (confirmed: boolean) => {
                if (settled) {
                    return;
                }
                settled = true;
                if (confirmRef.current?.resolve === finish) {
                    confirmRef.current = undefined;
                }
                resolve(confirmed);
            };
            const instance = modal.confirm({
                title: "确认解压",
                content: "压缩包内的同名文件会覆盖目标目录中的现有文件，请确认目标目录已做好备份。",
                okText: "解压",
                cancelText: "取消",
                mask: {closable: true},
                onOk: () => finish(true),
                onCancel: () => finish(false),
                afterClose: () => finish(false),
            });
            confirmRef.current = {
                destroy: instance.destroy,
                resolve: finish,
            };
        });
    }, [cancelPendingConfirm, modal]);

    const submit = useCallback(async (values: FileOperationFormValues) => {
        if (!operation || submittingGenerationRef.current === operation.generation) {
            return;
        }
        const generation = operation.generation;
        const mode = operation.mode;
        const records = operation.context.records || [];
        submittingGenerationRef.current = generation;
        setSubmitting(true);

        try {
            let response: API.Response<string> | undefined;
            let taskTitle = modeText[mode].title;

            if (mode === "compress") {
                const format = values.format || "zip";
                if (format === "gz" && (records.length !== 1 || records[0].is_dir)) {
                    message.error("GZIP 仅支持压缩单个普通文件");
                    return;
                }
                const name = String(values.name ?? "").trim();
                response = await compressFile({
                    body: {
                        paths: records.map((record) => joinFilePath(operation.context.path, record.name)),
                        dir: normalizeFilePath(String(values.dir ?? "").trim()),
                        format,
                        name,
                    },
                });
                taskTitle = records.length === 1 ? `压缩 ${records[0].name}` : `压缩 ${records.length} 项`;
            } else if (mode === "extract") {
                const record = records[0];
                if (!await confirmExtract() || generationRef.current !== generation) {
                    return;
                }
                response = await extractFile({
                    body: {
                        path: joinFilePath(operation.context.path, record.name),
                        dir: normalizeFilePath(String(values.dir ?? "").trim()),
                    },
                });
                taskTitle = `解压 ${record.name}`;
            } else {
                const name = String(values.name ?? "").trim();
                const url = String(values.url ?? "").trim();
                const targetPath = normalizeFilePath(String(values.path ?? "").trim());
                const targetName = name || getRemoteFileName(url);
                response = await remoteDownloadFile({
                    body: {
                        url,
                        path: targetPath,
                        ...(targetName ? {name: targetName} : {}),
                    },
                });
                taskTitle = targetName ? `下载 ${targetName}` : "远程下载";
            }

            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || `${modeText[mode].action}任务创建失败`);
                return;
            }
            const taskId = typeof response.data === "string" ? response.data : "";
            if (!taskId) {
                message.error("服务器未返回任务ID");
                return;
            }

            onTask(taskId, taskTitle);
            if (generationRef.current === generation) {
                onSuccess?.(operation.context);
            }
            message.success(response.message || `${modeText[mode].action}任务已创建`);
            if (generationRef.current === generation) {
                close();
            }
        } catch (error) {
            message.error(error instanceof Error && error.message
                ? error.message
                : `${modeText[mode].action}任务创建失败`);
        } finally {
            if (submittingGenerationRef.current === generation) {
                submittingGenerationRef.current = undefined;
            }
            if (generationRef.current === generation) {
                setSubmitting(false);
            }
        }
    }, [close, confirmExtract, message, onSuccess, onTask, operation]);

    const records = operation?.context.records || [];
    const initialValues: FileOperationFormValues | undefined = operation?.mode === "compress"
        ? {dir: operation.context.path, name: defaultArchiveName(records), format: "zip"}
        : operation?.mode === "extract"
            ? {dir: operation.context.path}
            : operation?.mode === "remote-download"
                ? {url: "", name: "", path: operation.context.path}
                : undefined;
    const modalWidth = operation?.mode === "remote-download" ? "460px" : "380px";

    return (
        <ProModal
            title={<Title>{operation ? modeText[operation.mode].title : "文件操作"}</Title>}
            width={modalWidth}
            okText={operation ? modeText[operation.mode].action : "确定"}
            onOk={() => formRef.current?.submit()}
            onCancel={close}
            modalProps={{
                open: Boolean(operation),
                forceRender: true,
                rootClassName: styles.modal,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                cancelText: "取消",
                confirmLoading: submitting,
                cancelButtonProps: {disabled: false},
                focusable: {focusTriggerAfterClose: false},
                mask: {closable: true},
            }}>
            <AntForm<FileOperationFormValues>
                key={operation?.generation ?? 0}
                ref={formRef}
                initialValues={initialValues}
                layout="vertical"
                preserve={false}
                onFinish={submit}>
                {operation && (
                    <>
                        {operation.mode === "compress" && <CompressOperationFields/>}
                        {operation.mode === "extract" && <ExtractOperationFields/>}
                        {operation.mode === "remote-download" && <RemoteDownloadOperationFields/>}
                    </>
                )}
            </AntForm>
        </ProModal>
    );
});

FileOperationForm.displayName = "FileOperationForm";

export default memo(FileOperationForm);
