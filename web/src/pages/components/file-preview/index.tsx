import React, {forwardRef, memo, useCallback, useImperativeHandle, useState} from "react";
import {App, Button, Grid, Tooltip} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import FilePreviewPane, {getFilePreviewKind, isFilePreviewable} from "./media";
import type {FilePreviewKind} from "./media";
import useStyles from "./styles";

export {getFilePreviewKind, isFilePreviewable} from "./media";

interface PreviewSession {
    files: any[];
    index: number;
}

const previewIcons: Record<FilePreviewKind, string> = {
    image: "FileImageOutlined",
    video: "VideoCameraOutlined",
    audio: "AudioOutlined",
};

const FilePreview = memo(forwardRef<any>((_, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const [session, setSession] = useState<PreviewSession>();
    const current = session?.files[session.index];
    const kind = current ? getFilePreviewKind(current.name) : undefined;
    const total = session?.files.length || 0;
    const index = session?.index || 0;

    const close = useCallback(() => {
        setSession(undefined);
    }, []);

    useImperativeHandle(ref, () => ({
        open: (files, active) => {
            const seen = new Set<string>();
            const available = files.reduce<any[]>((result, file) => {
                if (
                    file.isDirectory
                    || file.is_dir
                    || !file.path
                    || !isFilePreviewable(file.name)
                    || seen.has(file.path)
                ) {
                    return result;
                }
                seen.add(file.path);
                result.push({...file});
                return result;
            }, []);
            if (available.length === 0) {
                message.info("没有可预览的文件");
                return;
            }
            const activePath = typeof active === "number" ? files[active]?.path : active;
            const requestedIndex = available.findIndex((file) => file.path === activePath);
            setSession({
                files: available,
                index: requestedIndex >= 0 ? requestedIndex : 0,
            });
        },
        close,
    }), [close, message]);

    const changeIndex = useCallback((nextIndex: number) => {
        setSession((currentSession) => {
            if (
                !currentSession
                || nextIndex < 0
                || nextIndex >= currentSession.files.length
                || nextIndex === currentSession.index
            ) {
                return currentSession;
            }
            return {...currentSession, index: nextIndex};
        });
    }, []);

    return (
        <ProModal
            title={<Title>文件预览</Title>}
            width="900px"
            onCancel={close}
            modalProps={{
                open: Boolean(session),
                rootClassName: styles.root,
                footer: null,
                mask: {closable: true},
                destroyOnHidden: true,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
            }}>
            {current && kind && (
                <section className="file-preview" aria-label="文件预览">
                    <div className="file-preview-meta" aria-live="polite">
                        <span className="file-preview-meta-icon">
                            <Icon type={previewIcons[kind]} aria-hidden/>
                        </span>
                        <span className="file-preview-name" title={current.name}>{current.name}</span>
                        <span className="file-preview-position" aria-label={`第 ${index + 1} 项，共 ${total} 项`}>
                            {index + 1} / {total}
                        </span>
                    </div>
                    <FilePreviewPane
                        file={current}
                        controls={(getPopupContainer) => total > 1 ? (
                            <>
                                <Tooltip title="上一个" getPopupContainer={getPopupContainer}>
                                    <Button
                                        className="file-preview-nav is-previous"
                                        type="text"
                                        disabled={index <= 0}
                                        aria-label="预览上一个文件"
                                        icon={<Icon type="LeftOutlined"/>}
                                        onClick={() => changeIndex(index - 1)}/>
                                </Tooltip>
                                <Tooltip title="下一个" getPopupContainer={getPopupContainer}>
                                    <Button
                                        className="file-preview-nav is-next"
                                        type="text"
                                        disabled={index >= total - 1}
                                        aria-label="预览下一个文件"
                                        icon={<Icon type="RightOutlined"/>}
                                        onClick={() => changeIndex(index + 1)}/>
                                </Tooltip>
                            </>
                        ) : null}/>
                </section>
            )}
        </ProModal>
    );
}));

FilePreview.displayName = "FilePreview";

export default FilePreview;
