import React, {useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {Button, Empty, Spin, Tooltip} from "antd";
import {Icon, useTheme} from "sinking-antd";
import {filePreviewUrl, getFileSign} from "@/service/api/file";
import useFullscreen from "@/utils/fullscreen";
import useStyles from "../styles";

export type FilePreviewKind = "image" | "video" | "audio";

const previewKinds = new Map<string, FilePreviewKind>([
    ["bmp", "image"],
    ["gif", "image"],
    ["jpeg", "image"],
    ["jpg", "image"],
    ["png", "image"],
    ["svg", "image"],
    ["webp", "image"],
    ["mp4", "video"],
    ["mov", "video"],
    ["webm", "video"],
    ["mp3", "audio"],
]);

const getExtension = (name?: string) => {
    const value = String(name || "").toLowerCase();
    const index = value.lastIndexOf(".");
    return index >= 0 && index < value.length - 1 ? value.slice(index + 1) : "";
};

export const getFilePreviewKind = (name?: string) => previewKinds.get(getExtension(name));

export const isFilePreviewable = (name?: string) => Boolean(getFilePreviewKind(name));

interface FilePreviewPaneProps {
    file: {
        name?: string;
        path?: string;
    };
    embedded?: boolean;
    active?: boolean;
    showFullscreen?: boolean;
    controls?: (getPopupContainer: (trigger: HTMLElement) => HTMLElement) => React.ReactNode;
}

const FilePreviewPane = ({
    file,
    embedded = false,
    active = true,
    showFullscreen = true,
    controls,
}: FilePreviewPaneProps) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const requestRef = useRef(0);
    const activeUrlRef = useRef("");
    const stageRef = useRef<HTMLDivElement | null>(null);
    const controlsTimerRef = useRef<number | undefined>(undefined);
    const [url, setUrl] = useState("");
    const [loading, setLoading] = useState(true);
    const [mediaLoading, setMediaLoading] = useState(true);
    const [error, setError] = useState("");
    const [loadedKey, setLoadedKey] = useState("");
    const [reload, setReload] = useState(0);
    const {fullscreen, pageFullscreen, pending: fullscreenPending, toggle: toggleFullscreen} = useFullscreen(stageRef, {active});
    const [controlsVisible, setControlsVisible] = useState(true);
    const kind = getFilePreviewKind(file?.name);
    const path = String(file?.path || "");
    const resourceKey = `${path}\0${kind || ""}`;
    const currentResource = loadedKey === resourceKey;
    const previewUrl = currentResource ? url : "";
    const previewError = currentResource ? error : "";
    const previewLoading = !currentResource || loading;
    const previewMediaLoading = currentResource && mediaLoading;

    const clearControlsTimer = useCallback(() => {
        if (controlsTimerRef.current !== undefined) {
            window.clearTimeout(controlsTimerRef.current);
            controlsTimerRef.current = undefined;
        }
    }, []);

    const scheduleControlsHide = useCallback(() => {
        clearControlsTimer();
        if (
            (!showFullscreen && !controls)
            || typeof window === "undefined"
            || !window.matchMedia("(hover: hover) and (pointer: fine)").matches
        ) {
            return;
        }
        controlsTimerRef.current = window.setTimeout(() => {
            controlsTimerRef.current = undefined;
            setControlsVisible(false);
        }, 1800);
    }, [clearControlsTimer, controls, showFullscreen]);

    const revealControls = useCallback(() => {
        setControlsVisible(true);
        scheduleControlsHide();
    }, [scheduleControlsHide]);

    useEffect(() => {
        clearControlsTimer();
        setControlsVisible(true);
        scheduleControlsHide();
        return clearControlsTimer;
    }, [clearControlsTimer, fullscreen, scheduleControlsHide]);

    useLayoutEffect(() => {
        return () => {
            requestRef.current += 1;
            activeUrlRef.current = "";
            clearControlsTimer();
        };
    }, [clearControlsTimer]);

    useEffect(() => {
        if (!active) return;
        setLoadedKey(resourceKey);
        if (!path || !kind) {
            requestRef.current += 1;
            activeUrlRef.current = "";
            setUrl("");
            setLoading(false);
            setMediaLoading(false);
            setError("该文件暂不支持预览");
            return;
        }
        const requestId = ++requestRef.current;
        activeUrlRef.current = "";
        setUrl("");
        setError("");
        setLoading(true);
        setMediaLoading(true);
        void getFileSign({body: {path, download: false}}).then((response) => {
            if (requestRef.current !== requestId) {
                return;
            }
            const key = typeof response?.data === "string" ? response?.data : "";
            if (response?.code !== 200 || !/^[a-f0-9]{32}$/.test(key)) {
                setError(response?.message || "获取文件预览地址失败");
                setMediaLoading(false);
                return;
            }
            try {
                const nextUrl = filePreviewUrl(key);
                activeUrlRef.current = nextUrl;
                setUrl(nextUrl);
            } catch {
                setError("生成文件预览地址失败");
                setMediaLoading(false);
            }
        }).catch(() => {
            if (requestRef.current === requestId) {
                setError("获取文件预览地址失败");
                setMediaLoading(false);
            }
        }).finally(() => {
            if (requestRef.current === requestId) {
                setLoading(false);
            }
        });
        return () => {
            if (requestRef.current === requestId) {
                requestRef.current += 1;
            }
        };
    }, [active, kind, path, reload, resourceKey]);

    const retry = useCallback(() => {
        requestRef.current += 1;
        activeUrlRef.current = "";
        setUrl("");
        setError("");
        setLoading(true);
        setMediaLoading(true);
        setReload((value) => value + 1);
    }, []);

    const getPopupContainer = useCallback((trigger: HTMLElement) => (
        stageRef.current || trigger.parentElement || document.body
    ), []);

    const media = useMemo(() => {
        if (!kind || !previewUrl || previewError) {
            return null;
        }
        const sourceUrl = previewUrl;
        const ready = () => {
            if (activeUrlRef.current === sourceUrl) {
                setMediaLoading(false);
            }
        };
        const failed = () => {
            if (activeUrlRef.current !== sourceUrl) {
                return;
            }
            setMediaLoading(false);
            setError("文件加载失败或浏览器不支持该媒体格式");
        };
        if (kind === "image") {
            return (
                <img
                    key={url}
                    className="file-preview-media file-preview-image"
                    src={previewUrl}
                    alt={file?.name || "文件预览"}
                    referrerPolicy="no-referrer"
                    onLoad={ready}
                    onError={failed}/>
            );
        }
        if (kind === "video") {
            return (
                <video
                    key={url}
                    className="file-preview-media file-preview-video"
                    src={previewUrl}
                    controls
                    playsInline
                    preload="metadata"
                    aria-label={file?.name || "视频预览"}
                    onLoadedMetadata={ready}
                    onError={failed}/>
            );
        }
        if (kind === "audio") {
            return (
                <div className="file-preview-audio">
                    <Icon className="file-preview-audio-icon" type="AudioOutlined" aria-hidden/>
                    <audio
                        key={url}
                        src={previewUrl}
                        controls
                        preload="metadata"
                        aria-label={file?.name || "音频预览"}
                        onLoadedMetadata={ready}
                        onError={failed}/>
                </div>
            );
        }
        return null;
    }, [file?.name, kind, previewError, previewUrl, url]);

    return (
        <div className={styles.root} style={embedded ? {width: "100%", height: "100%"} : undefined}>
            <div
                ref={stageRef}
                className={`file-preview-stage ${embedded ? "is-embedded" : ""} ${pageFullscreen ? "panel-page-fullscreen" : ""} ${controlsVisible ? "" : "file-preview-controls-hidden"}`}
                aria-label={`${file?.name || "文件"}预览`}
                aria-busy={previewLoading || previewMediaLoading}
                onPointerMove={(event) => event.pointerType === "mouse" && revealControls()}
                onPointerDown={(event) => event.pointerType === "mouse" && revealControls()}
                onPointerEnter={(event) => event.pointerType === "mouse" && revealControls()}
                onFocusCapture={revealControls}>
                {media}
                {(previewLoading || (previewUrl && previewMediaLoading && !previewError)) && (
                    <div className="file-preview-loading" role="status" aria-label="正在加载文件预览">
                        <Spin size={embedded ? "medium" : "large"}/>
                    </div>
                )}
                {!previewLoading && previewError && (
                    <div className="file-preview-error" role="alert">
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={previewError}>
                            <Button size="small" icon={<Icon type="ReloadOutlined"/>} onClick={retry}>重新加载</Button>
                        </Empty>
                    </div>
                )}
                {showFullscreen && (
                    <Tooltip title={fullscreen ? "退出全屏" : "全屏"} placement="left" getPopupContainer={getPopupContainer}>
                        <Button
                            className="file-preview-fullscreen"
                            color="default"
                            variant="outlined"
                            shape="circle"
                            aria-label={fullscreen ? "退出全屏" : "全屏"}
                            aria-pressed={fullscreen}
                            disabled={fullscreenPending}
                            icon={<Icon type={fullscreen ? "FullscreenExitOutlined" : "FullscreenOutlined"}/>}
                            onClick={() => void toggleFullscreen()}/>
                    </Tooltip>
                )}
                {controls?.(getPopupContainer)}
            </div>
        </div>
    );
};

export default React.memo(FilePreviewPane);
