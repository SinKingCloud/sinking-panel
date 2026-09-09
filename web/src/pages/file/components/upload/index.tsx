import React, {forwardRef, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Button, Grid, Progress, theme, Upload} from "antd";
import type {UploadProps} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {formatFileSize, getFileIconType} from "../../utils";
import {uploadFile} from "./transfer";
import useStyles from "./styles";

type UploadStatus = "waiting" | "checking" | "uploading" | "merging" | "done" | "error" | "canceled";

interface UploadItem {
    id: string;
    file: File;
    path: string;
    name: string;
    percent: number;
    status: UploadStatus;
    error?: string;
}

type UploadRequestOptions = Parameters<NonNullable<UploadProps["customRequest"]>>[0];
type UploadCallbacks = Pick<UploadRequestOptions, "onSuccess" | "onError">;

const getUploadStatusText = (item: UploadItem) => {
    if (item.status === "waiting") return "等待中";
    if (item.status === "checking") return "检查中";
    if (item.status === "uploading") return `上传中 ${Math.floor(item.percent)}%`;
    if (item.status === "merging") return "合并中";
    if (item.status === "done") return "已完成";
    if (item.status === "canceled") return "已暂停";
    return "失败";
};

const activeUploadStatuses = new Set<UploadStatus>(["waiting", "checking", "uploading", "merging"]);

export interface FileUploadRef {
    open: () => void;
}

interface FileUploadProps {
    path: string;
    selectionDisabled?: boolean;
    onUploaded: () => void;
    onUploadingChange?: (uploading: boolean) => void;
}

const FileUpload = forwardRef<FileUploadRef, FileUploadProps>(({
    path,
    selectionDisabled = false,
    onUploaded,
    onUploadingChange,
}, ref) => {
    const {message} = App.useApp();
    const {token} = theme.useToken();
    const appTheme = useTheme();
    const compact = Boolean(appTheme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const screens = Grid.useBreakpoint();
    const transfersRef = useRef(new Map<string, ReturnType<typeof uploadFile>>());
    const pendingNamesRef = useRef(new Set<string>());
    const fileTargetsRef = useRef(new WeakMap<File, string>());
    const onUploadingChangeRef = useRef(onUploadingChange);
    const mountedRef = useRef(true);
    const [open, setOpen] = useState(false);
    const [items, setItems] = useState<UploadItem[]>([]);
    const activeCount = useMemo(
        () => items.filter((item) => activeUploadStatuses.has(item.status)).length,
        [items],
    );

    useImperativeHandle(ref, () => ({open: () => setOpen(true)}), []);

    useEffect(() => {
        onUploadingChangeRef.current = onUploadingChange;
    }, [onUploadingChange]);

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            transfersRef.current.forEach((transfer) => transfer.abort());
            transfersRef.current.clear();
            onUploadingChangeRef.current?.(false);
        };
    }, []);

    useEffect(() => {
        onUploadingChangeRef.current?.(activeCount > 0);
    }, [activeCount]);

    const update = (id: string, values: Partial<UploadItem>) => {
        if (!mountedRef.current) {
            return;
        }
        setItems((current) => current.map((item) => item.id === id ? {...item, ...values} : item));
    };

    const getPendingKey = (targetPath: string, name: string) => `${targetPath}\0${name}`;

    const start = (
        file: File,
        targetPath: string,
        reuseId?: string,
        callbacks?: UploadCallbacks,
    ) => {
        const id = reuseId || `${file.name}-${file.size}-${file.lastModified}-${Date.now()}-${Math.random()}`;
        setItems((current) => reuseId
                ? current.map((item) => item.id === id ? {...item, percent: 0, status: "waiting" as const, error: undefined} : item)
                : [
                    ...current.filter((item) => item.path !== targetPath || item.name !== file.name || activeUploadStatuses.has(item.status)),
                    {id, file, path: targetPath, name: file.name, percent: 0, status: "waiting" as const},
                ]);
        const transfer = uploadFile({
            path: targetPath,
            file,
            fileName: file.name,
            onProgress: ({percent, phase}) => {
                if (phase === "complete") {
                    return;
                }
                update(id, {
                    percent: phase === "uploading" ? Math.min(99, Math.floor(percent)) : Math.floor(percent),
                    status: phase === "checking" ? "checking" : phase === "merging" ? "merging" : "uploading",
                });
            },
        });
        transfersRef.current.set(id, transfer);
        void transfer.promise.then((response) => {
            if (response.code !== 200) {
                callbacks?.onError?.(new Error(response.message || "上传失败"));
                update(id, {status: "error", error: response.message || "上传失败"});
                if (mountedRef.current) message.error(`${file.name}：${response.message || "上传失败"}`);
                return;
            }
            callbacks?.onSuccess?.(response);
            update(id, {status: "done", percent: 100});
            if (mountedRef.current) onUploaded();
        }).catch((error: unknown) => {
            const uploadError = error instanceof Error ? error : new Error("上传失败");
            callbacks?.onError?.(uploadError);
            if (uploadError.name === "AbortError") {
                update(id, {status: "canceled", error: undefined});
                return;
            }
            const text = uploadError.message || "上传失败";
            update(id, {status: "error", error: text});
            if (mountedRef.current) message.error(`${file.name}：${text}`);
        }).finally(() => {
            transfersRef.current.delete(id);
            pendingNamesRef.current.delete(getPendingKey(targetPath, file.name));
            if (!mountedRef.current) {
                return;
            }
            setItems((current) => current.map((item) => item.id === id && activeUploadStatuses.has(item.status)
                    ? {...item, status: "error" as const, error: "上传未完成"}
                    : item));
        });
        return transfer;
    };

    const canUpload = (file: File, targetPath: string) => {
        const pendingKey = getPendingKey(targetPath, file.name);
        if (pendingNamesRef.current.has(pendingKey)) {
            message.warning(`上传队列中已存在“${file.name}”`);
            return false;
        }
        pendingNamesRef.current.add(pendingKey);
        return true;
    };

    const beforeUpload = (file: File) => {
        const targetPath = path;
        if (!canUpload(file, targetPath)) {
            return false;
        }
        fileTargetsRef.current.set(file, targetPath);
        return true;
    };

    const customRequest = (options: UploadRequestOptions) => {
        const file = options.file as File;
        const targetPath = fileTargetsRef.current.get(file) || path;
        fileTargetsRef.current.delete(file);
        const transfer = start(file, targetPath, undefined, {
            onSuccess: options.onSuccess,
            onError: options.onError,
        });
        return {abort: transfer.abort};
    };

    const cancel = (id: string) => transfersRef.current.get(id)?.abort();
    const retry = (item: UploadItem) => {
        if (transfersRef.current.has(item.id)) {
            return;
        }
        if (canUpload(item.file, item.path)) {
            start(item.file, item.path, item.id);
        }
    };
    const removeItem = (id: string) => setItems((current) => current.filter((item) => item.id !== id));
    const modalBodyHeight = items.length > 0
        ? undefined
        : (screens.md ? 300 : "42dvh");

    return (
        <ProModal
            title={<Title>上传文件</Title>}
            width="640px"
            onCancel={() => setOpen(false)}
            modalProps={{
                open,
                rootClassName: styles.fileUploadModal,
                footer: items.length > 0 ? (
                    <div className="file-upload-footer">
                        <div className="file-upload-footer-actions">
                            <Button onClick={() => setOpen(false)}>取消</Button>
                            <Upload
                                multiple
                                disabled={selectionDisabled}
                                showUploadList={false}
                                beforeUpload={beforeUpload}
                                customRequest={customRequest}>
                                <Button type="primary" disabled={selectionDisabled}>上传</Button>
                            </Upload>
                        </div>
                    </div>
                ) : null,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                styles: {
                    container: {background: token.colorBgContainer},
                    header: {background: token.colorBgContainer},
                    body: {
                        background: token.colorBgContainer,
                        height: modalBodyHeight,
                        overflow: "hidden",
                    },
                },
                mask: {closable: true},
                destroyOnHidden: false,
            }}>
            <div className="file-upload-picker" hidden={items.length > 0}>
                <Upload.Dragger
                    multiple
                    disabled={selectionDisabled}
                    showUploadList={false}
                    beforeUpload={beforeUpload}
                    customRequest={customRequest}>
                    <div className="ant-upload-drag-icon"><Icon type="InboxOutlined"/></div>
                    <div className="ant-upload-text">
                        {selectionDisabled ? "目录加载中" : "拖拽文件到此处，或点击选择"}
                    </div>
                    <div className="ant-upload-hint">
                        {selectionDisabled ? "请稍候" : "支持多文件上传，同名文件将直接覆盖"}
                    </div>
                </Upload.Dragger>
            </div>

            {items.length > 0 && (
                <section className="file-upload-tasks" aria-label="上传列表">
                    <div className="file-upload-list">
                        {items.map((item) => (
                            <article
                                className={`file-upload-item ${item.status}`}
                                key={item.id}
                                aria-busy={activeUploadStatuses.has(item.status)}>
                                <span className="upload-file-icon">
                                    <Icon type={getFileIconType(item.name, false)}/>
                                </span>
                                <div className="upload-file-content">
                                    <div className="upload-file-head">
                                        <span className="upload-file-name" title={item.name}>{item.name}</span>
                                        <div className="upload-file-control">
                                            {activeUploadStatuses.has(item.status) ? (
                                                <Button
                                                    color="default"
                                                    variant="text"
                                                    size="small"
                                                    aria-label={`暂停上传 ${item.name}`}
                                                    icon={<Icon type="PauseOutlined"/>}
                                                    onClick={() => cancel(item.id)}>
                                                    暂停
                                                </Button>
                                            ) : item.status === "error" || item.status === "canceled" ? (
                                                <>
                                                    <Button
                                                        color="default"
                                                        variant="text"
                                                        size="small"
                                                        aria-label={`${item.status === "canceled" ? "继续" : "重试"}上传 ${item.name}`}
                                                        icon={(
                                                            <Icon type={item.status === "canceled" ? "PlayCircleOutlined" : "ReloadOutlined"}/>
                                                        )}
                                                        onClick={() => retry(item)}>
                                                        {item.status === "canceled" ? "继续" : "重试"}
                                                    </Button>
                                                    <Button
                                                        color="default"
                                                        variant="text"
                                                        size="small"
                                                        aria-label={`移除上传记录 ${item.name}`}
                                                        onClick={() => removeItem(item.id)}>
                                                        移除
                                                    </Button>
                                                </>
                                            ) : (
                                                <Button
                                                    color="default"
                                                    variant="text"
                                                    size="small"
                                                    aria-label={`移除上传记录 ${item.name}`}
                                                    onClick={() => removeItem(item.id)}>
                                                    移除
                                                </Button>
                                            )}
                                        </div>
                                    </div>
                                    <div className="upload-file-meta">
                                        <span className={`upload-file-state ${item.status}`}>{getUploadStatusText(item)}</span>
                                        <span>{formatFileSize(item.file.size)}</span>
                                        {item.error && <span className="upload-file-error" title={item.error}>{item.error}</span>}
                                    </div>
                                    <Progress
                                        percent={item.percent}
                                        size="small"
                                        status={item.status === "error"
                                            ? "exception"
                                            : item.status === "done" ? "success" : activeUploadStatuses.has(item.status) && item.status !== "waiting" ? "active" : "normal"}
                                        showInfo={false}/>
                                </div>
                            </article>
                        ))}
                    </div>
                </section>
            )}
        </ProModal>
    );
});

FileUpload.displayName = "FileUpload";

export default React.memo(FileUpload);
