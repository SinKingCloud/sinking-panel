import React, {
    forwardRef,
    memo,
    useCallback,
    useEffect,
    useImperativeHandle,
    useMemo,
    useRef,
    useState,
} from "react";
import {App, Button, Empty, Grid, Spin, Tooltip} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import {getFileSign} from "@/service/api/file";
import useStyles from "./styles";

interface PreviewSession {
    files: any[];
    index: number;
}

const previewKinds = new Map<string, string>([
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

const previewIcons: Record<string, string> = {
    image: "FileImageOutlined",
    video: "VideoCameraOutlined",
    audio: "AudioOutlined",
};

const getExtension = (name: string) => {
    const value = String(name || "").toLowerCase();
    const index = value.lastIndexOf(".");
    return index >= 0 && index < value.length - 1 ? value.slice(index + 1) : "";
};

export const getFilePreviewKind = (name: string) => previewKinds.get(getExtension(name));

export const isFilePreviewable = (name: string) => Boolean(getFilePreviewKind(name));

const apiUrl = (path: string) => {
    const origin = window.location.origin;
    const gateway = new URL(defaultSettings?.gateway || "/", origin);
    const gatewayPath = gateway.pathname.replace(/\/+$/, "");
    gateway.pathname = `${gatewayPath}/${path.replace(/^\/+/, "")}`.replace(/\/{2,}/g, "/");
    gateway.search = "";
    gateway.hash = "";
    return gateway.toString();
};

const buildPreviewUrl = (key: string) => {
    const url = new URL(apiUrl("/preview"));
    url.searchParams.set("key", key);
    return url.toString();
};

const FilePreview = memo(forwardRef<any>((_, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const requestRef = useRef(0);
    const activeUrlRef = useRef("");
    const stageRef = useRef<HTMLDivElement | null>(null);
    const [session, setSession] = useState<PreviewSession>();
    const [url, setUrl] = useState("");
    const [loading, setLoading] = useState(false);
    const [mediaLoading, setMediaLoading] = useState(false);
    const [error, setError] = useState("");
    const [reload, setReload] = useState(0);
    const [fullscreen, setFullscreen] = useState(false);
    const [controlsVisible, setControlsVisible] = useState(true);
    const controlsTimerRef = useRef<number | undefined>(undefined);

    const current = session?.files[session.index];
    const kind = current ? getFilePreviewKind(current.name) : undefined;

    const clearControlsTimer = useCallback(() => {
        if (controlsTimerRef.current !== undefined) {
            window.clearTimeout(controlsTimerRef.current);
            controlsTimerRef.current = undefined;
        }
    }, []);

    const scheduleControlsHide = useCallback(() => {
        clearControlsTimer();
        if (
            typeof window === "undefined" ||
            !window.matchMedia("(hover: hover) and (pointer: fine)").matches
        ) {
            return;
        }
        controlsTimerRef.current = window.setTimeout(() => {
            controlsTimerRef.current = undefined;
            setControlsVisible(false);
        }, 1800);
    }, [clearControlsTimer]);

    const revealControls = useCallback(() => {
        setControlsVisible(true);
        scheduleControlsHide();
    }, [scheduleControlsHide]);

    const handlePointerActivity = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
        if (event.pointerType === "mouse") {
            revealControls();
        }
    }, [revealControls]);

    const close = useCallback(() => {
        if (document.fullscreenElement === stageRef.current) {
            void document.exitFullscreen().catch(() => undefined);
        }
        requestRef.current += 1;
        activeUrlRef.current = "";
        setSession(undefined);
        setUrl("");
        setLoading(false);
        setMediaLoading(false);
        setError("");
        setReload(0);
        setFullscreen(false);
        clearControlsTimer();
        setControlsVisible(true);
    }, [clearControlsTimer]);

    useImperativeHandle(ref, () => ({
        open: (files, active) => {
            const seen = new Set<string>();
            const available = files.reduce<any[]>((result, file) => {
                if (
                    file.isDirectory ||
                    !file.path ||
                    !isFilePreviewable(file.name) ||
                    seen.has(file.path)
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
            const index = requestedIndex >= 0 && requestedIndex < available.length ? requestedIndex : 0;
            requestRef.current += 1;
            activeUrlRef.current = "";
            setUrl("");
            setError("");
            setLoading(true);
            setMediaLoading(true);
            setReload(0);
            setSession({files: available, index});
        },
        close,
    }), [close, message]);

    useEffect(() => () => {
        requestRef.current += 1;
        activeUrlRef.current = "";
        clearControlsTimer();
    }, [clearControlsTimer]);

    useEffect(() => {
        clearControlsTimer();
        setControlsVisible(true);
        if (session) {
            scheduleControlsHide();
        }
        return clearControlsTimer;
    }, [clearControlsTimer, fullscreen, scheduleControlsHide, session]);

    useEffect(() => {
        const handleFullscreenChange = () => {
            setFullscreen(document.fullscreenElement === stageRef.current);
        };
        document.addEventListener("fullscreenchange", handleFullscreenChange);
        return () => document.removeEventListener("fullscreenchange", handleFullscreenChange);
    }, []);

    useEffect(() => {
        if (!current) {
            return;
        }
        const requestId = ++requestRef.current;
        setError("");
        setMediaLoading(true);
        activeUrlRef.current = "";
        setUrl("");
        setLoading(true);
        void getFileSign({body: {path: current.path, download: false}}).then((response) => {
            if (requestRef.current !== requestId) {
                return;
            }
            const key = typeof response?.data === "string" ? response.data : "";
            if (response?.code !== 200 || !/^[a-f0-9]{32}$/.test(key)) {
                setError(response?.message || "获取文件预览地址失败");
                setMediaLoading(false);
                return;
            }
            try {
                const nextUrl = buildPreviewUrl(key);
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
    }, [current, reload]);

    const changeIndex = useCallback((nextIndex: number) => {
        if (!session || nextIndex < 0 || nextIndex >= session.files.length || nextIndex === session.index) {
            return;
        }
        requestRef.current += 1;
        activeUrlRef.current = "";
        setUrl("");
        setError("");
        setLoading(true);
        setMediaLoading(true);
        setSession({...session, index: nextIndex});
    }, [session]);

    const retry = useCallback(() => {
        if (!current) {
            return;
        }
        requestRef.current += 1;
        activeUrlRef.current = "";
        setUrl("");
        setError("");
        setLoading(true);
        setMediaLoading(true);
        setReload((value) => value + 1);
    }, [current]);

    const toggleFullscreen = useCallback(async () => {
        const stage = stageRef.current;
        if (!stage) {
            return;
        }
        try {
            if (document.fullscreenElement === stage) {
                await document.exitFullscreen();
            } else {
                await stage.requestFullscreen();
            }
        } catch {
            message.error(fullscreen ? "退出全屏失败" : "进入全屏失败");
        }
    }, [fullscreen, message]);

    const getStagePopupContainer = useCallback((trigger: HTMLElement) => (
        stageRef.current || trigger.parentElement || document.body
    ), []);

    const media = useMemo(() => {
        if (!current || !kind || !url || error) {
            return null;
        }
        const sourceUrl = url;
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
                    src={url}
                    alt={current.name}
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
                    src={url}
                    controls
                    playsInline
                    preload="metadata"
                    aria-label={current.name}
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
                        src={url}
                        controls
                        preload="metadata"
                        aria-label={current.name}
                        onLoadedMetadata={ready}
                        onError={failed}/>
                </div>
            );
        }
        return null;
    }, [current, error, kind, url]);

    const total = session?.files.length || 0;
    const index = session?.index || 0;

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
                    <div
                        ref={stageRef}
                        className={`file-preview-stage ${controlsVisible ? "" : "file-preview-controls-hidden"}`}
                        aria-busy={loading || mediaLoading}
                        onPointerMove={handlePointerActivity}
                        onPointerDown={handlePointerActivity}
                        onPointerEnter={handlePointerActivity}
                        onFocusCapture={revealControls}>
                        {media}
                        {(loading || (url && mediaLoading && !error)) && (
                            <div className="file-preview-loading" role="status" aria-label="正在加载文件预览">
                                <Spin size="large"/>
                            </div>
                        )}
                        {!loading && error && (
                            <div className="file-preview-error" role="alert">
                                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={error}>
                                    <Button
                                        size="small"
                                        icon={<Icon type="ReloadOutlined"/>}
                                        onClick={retry}>
                                        重新加载
                                    </Button>
                                </Empty>
                            </div>
                        )}
                        <Tooltip
                            title={fullscreen ? "退出全屏" : "全屏"}
                            placement="left"
                            getPopupContainer={getStagePopupContainer}>
                            <Button
                                className="file-preview-fullscreen"
                                type="text"
                                aria-label={fullscreen ? "退出全屏" : "全屏"}
                                aria-pressed={fullscreen}
                                icon={<Icon type={fullscreen
                                    ? "FullscreenExitOutlined"
                                    : "FullscreenOutlined"}/>}
                                onClick={() => void toggleFullscreen()}/>
                        </Tooltip>
                        {total > 1 && (
                            <>
                                <Tooltip title="上一个" getPopupContainer={getStagePopupContainer}>
                                    <Button
                                        className="file-preview-nav is-previous"
                                        type="text"
                                        disabled={index <= 0}
                                        aria-label="预览上一个文件"
                                        icon={<Icon type="LeftOutlined"/>}
                                        onClick={() => changeIndex(index - 1)}/>
                                </Tooltip>
                                <Tooltip title="下一个" getPopupContainer={getStagePopupContainer}>
                                    <Button
                                        className="file-preview-nav is-next"
                                        type="text"
                                        disabled={index >= total - 1}
                                        aria-label="预览下一个文件"
                                        icon={<Icon type="RightOutlined"/>}
                                        onClick={() => changeIndex(index + 1)}/>
                                </Tooltip>
                            </>
                        )}
                    </div>
                </section>
            )}
        </ProModal>
    );
}));

FilePreview.displayName = "FilePreview";

export default FilePreview;
