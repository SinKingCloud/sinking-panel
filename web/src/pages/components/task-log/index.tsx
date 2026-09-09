import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {Button, Empty, Spin, Tooltip} from "antd";
import {createStyles} from "antd-style";
import {Icon, ProModal, ProModalRef, Title} from "sinking-antd";
import useLog, {logLayout} from "./hooks";

const useStyles = createStyles(({css, token}: any) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
        }
    `,
    panel: css`
        position: relative;
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorBgContainer};

        &:fullscreen {
            width: 100%;
            height: 100%;
            padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left);
            box-sizing: border-box;
            display: flex;
            flex-direction: column;
            border: 0;
            border-radius: 0;

            .task-log-console {
                min-height: 0;
                height: auto;
                flex: 1;
            }
        }
    `,
    panelHeader: css`
        height: 42px;
        padding: 0 9px;
        display: flex;
        flex: none;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        box-sizing: border-box;
        background: ${token.colorBgContainer};
        box-shadow: inset 0 -1px 0 ${token.colorSplit};
    `,
    panelTitle: css`
        min-width: 0;
        padding-left: 5px;
        overflow: hidden;
        color: ${token.colorTextTertiary};
        font-size: ${token.fontSizeSM}px;
        font-weight: 500;
        line-height: 20px;
        text-overflow: ellipsis;
        white-space: nowrap;
    `,
    console: css`
        position: relative;
        height: min(480px, calc(100vh - 280px));
        min-height: 300px;
        overflow: auto;
        scrollbar-gutter: stable;
        background: ${token.colorBgContainer};

        &::-webkit-scrollbar {
            width: 8px !important;
            height: 8px !important;
        }

        &::-webkit-scrollbar-track {
            background: transparent;
        }

        &::-webkit-scrollbar-thumb {
            border: 2px solid ${token.colorBgContainer};
            border-radius: 8px;
            background: ${token.colorFillSecondary};
        }

        &::-webkit-scrollbar-thumb:hover {
            background: ${token.colorFill};
        }

        @media (max-width: 560px) {
            height: min(440px, calc(100vh - 250px));
            min-height: 260px;
        }
    `,
    content: css`
        width: max-content;
        min-width: 100%;
        min-height: 100%;
        margin: 0;
        box-sizing: border-box;
        color: ${token.colorTextSecondary};
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
        font-size: 12px;
        line-height: ${logLayout.lineHeight}px;
        white-space: pre;
        overflow-wrap: normal;
        word-break: normal;
        tab-size: 4;
    `,
    virtualContent: css`
        position: relative;
        width: max-content;
        min-width: 100%;
        box-sizing: border-box;
        color: ${token.colorTextSecondary};
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
        font-size: 12px;
        line-height: ${logLayout.lineHeight}px;
        tab-size: 4;
    `,
    virtualLine: css`
        position: absolute;
        left: 0;
        width: max-content;
        min-width: 100%;
        height: ${logLayout.lineHeight}px;
        color: inherit;
        font: inherit;
        line-height: ${logLayout.lineHeight}px;
        white-space: pre;
        overflow-wrap: normal;
        word-break: normal;
    `,
    actions: css`
        display: flex;
        flex: none;
        align-items: center;
        justify-content: flex-end;
        gap: 2px;

        .ant-btn {
            width: 30px;
            height: 30px;
            padding: 0;
            display: inline-flex;
            align-items: center;
            justify-content: center;
        }

        .ant-btn .anticon {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            line-height: 1;
        }

    `,
    state: css`
        min-width: 100%;
        min-height: 100%;
        box-sizing: border-box;
        display: flex;
        align-items: center;
        justify-content: center;

        .ant-empty-description {
            color: ${token.colorTextTertiary};
        }
    `,
}));

export interface LogRef {
    open: (record: any, body?: Record<string, any>) => void;
}

interface LogProps {
    request?: (params: API.RequestParams) => Promise<any>;
    showClear?: boolean;
    title?: React.ReactNode;
    body?: Record<string, any>;
}

const Log = forwardRef<LogRef, LogProps>(({request, showClear = true, title = "任务日志", body = {}}, ref):any => {
    const {styles} = useStyles();
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const panelRef = useRef<HTMLDivElement | null>(null);
    const openRef = useRef(false);
    const [fullscreen, setFullscreen] = useState(false);
    const [fullscreenPending, setFullscreenPending] = useState(false);
    const log = useLog(request, body, panelRef);

    const exitFullscreen = useCallback(() => {
        const panel = panelRef.current;
        if (panel && document.fullscreenElement === panel && typeof document.exitFullscreen === "function") {
            void document.exitFullscreen().catch(() => undefined);
        }
        setFullscreen(false);
    }, []);

    const setPanel = useCallback((panel: HTMLDivElement | null) => {
        if (!panel) {
            const current = panelRef.current;
            if (current && document.fullscreenElement === current && typeof document.exitFullscreen === "function") {
                void document.exitFullscreen().catch(() => undefined);
            }
        }
        panelRef.current = panel;
    }, []);

    const getPopupContainer = useCallback(() => {
        const panel = panelRef.current;
        return panel && document.fullscreenElement === panel ? panel : document.body;
    }, []);

    useEffect(() => {
        const handleFullscreen = () => {
            setFullscreen(Boolean(panelRef.current && document.fullscreenElement === panelRef.current));
            log.measureViewport();
        };
        document.addEventListener("fullscreenchange", handleFullscreen);
        return () => document.removeEventListener("fullscreenchange", handleFullscreen);
    }, [log.measureViewport]);

    const toggleFullscreen = useCallback(async () => {
        const panel = panelRef.current;
        if (!panel || fullscreenPending) return;
        if (!document.fullscreenEnabled || typeof panel.requestFullscreen !== "function" || typeof document.exitFullscreen !== "function") {
            log.message.info("当前浏览器不支持全屏");
            return;
        }
        const exiting = document.fullscreenElement === panel;
        setFullscreenPending(true);
        try {
            if (exiting) await document.exitFullscreen();
            else await panel.requestFullscreen();
            if ((!openRef.current || panelRef.current !== panel) && document.fullscreenElement === panel) {
                await document.exitFullscreen();
            }
        } catch {
            if (openRef.current && panelRef.current === panel) log.message.error(exiting ? "退出全屏失败" : "进入全屏失败");
        } finally {
            setFullscreenPending(false);
        }
    }, [fullscreenPending, log.message]);

    useImperativeHandle(ref, () => ({
        open: (record: any, requestBody?: Record<string, any>) => {
            if (log.open(record, requestBody)) {
                openRef.current = true;
                modalRef.current?.show();
            }
        },
    }), [log.open]);

    return (
        <ProModal
            ref={modalRef}
            title={<Title>{title}</Title>}
            width={900}
            afterOpenChange={(open) => {
                if (!open) {
                    openRef.current = false;
                    exitFullscreen();
                    log.pause();
                }
            }}
            modalProps={{
                rootClassName: styles.modal,
                footer: null,
                keyboard: !fullscreen,
                zIndex: 1200,
                style: {top: 100, paddingBottom: 100},
                mask: {closable: true},
                afterOpenChange: (open: boolean) => {
                    if (open) log.measureViewport();
                },
                afterClose: log.reset,
            } as any}>
            <div ref={setPanel} className={styles.panel}>
                {log.messageHolder}
                <div className={styles.panelHeader}>
                    <span className={styles.panelTitle}>{log.fileName || "日志文件"}</span>
                    <div className={styles.actions}>
                        <Tooltip title={fullscreen ? "退出全屏" : "全屏"} placement={fullscreen ? "bottom" : "top"} getPopupContainer={getPopupContainer}>
                            <Button
                                color="default"
                                variant="text"
                                aria-label={fullscreen ? "退出全屏" : "全屏"}
                                aria-pressed={fullscreen}
                                icon={<Icon type={fullscreen ? "FullscreenExitOutlined" : "FullscreenOutlined"}/>}
                                disabled={fullscreenPending}
                                onClick={() => void toggleFullscreen()}/>
                        </Tooltip>
                        <Tooltip title="刷新日志" placement={fullscreen ? "bottom" : "top"} getPopupContainer={getPopupContainer}>
                            <Button
                                color="default"
                                variant="text"
                                aria-label="刷新日志"
                                icon={<Icon type="ReloadOutlined"/>}
                                loading={log.loading}
                                disabled={log.clearing}
                                onClick={log.refresh}/>
                        </Tooltip>
                        {showClear && (
                            <Tooltip title="清理日志" placement={fullscreen ? "bottom" : "top"} getPopupContainer={getPopupContainer}>
                                <Button
                                    color="danger"
                                    variant="text"
                                    aria-label="清理日志"
                                    icon={<Icon type="DeleteOutlined"/>}
                                    loading={log.clearing}
                                    onClick={log.clear}/>
                            </Tooltip>
                        )}
                    </div>
                </div>
                <div ref={log.setConsole} className={`${styles.console} task-log-console`} onScroll={log.handleScroll}>
                    {log.logs.length > 0 ? (
                        log.virtual ? (
                            <div className={styles.virtualContent} style={{height: log.virtualHeight}}>
                                {log.logs.slice(log.virtualStart, log.virtualEnd).map((line, index) => {
                                    const lineIndex = log.virtualStart + index;
                                    return (
                                        <div
                                            className={styles.virtualLine}
                                            key={lineIndex}
                                            style={{top: lineIndex * logLayout.lineHeight}}>
                                            {line}
                                        </div>
                                    );
                                })}
                            </div>
                        ) : (
                            <pre className={styles.content}>{log.logs.join("\n")}</pre>
                        )
                    ) : log.loading ? (
                        <div className={styles.state}><Spin/></div>
                    ) : (
                        <div className={styles.state}>
                            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
                        </div>
                    )}
                </div>
            </div>
        </ProModal>
    );
});

Log.displayName = "Log";

export default memo(Log);
