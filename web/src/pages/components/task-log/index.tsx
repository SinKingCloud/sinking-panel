import {forwardRef, memo, useImperativeHandle, useRef} from "react";
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
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorBgContainer};
    `,
    panelHeader: css`
        height: 42px;
        padding: 0 9px;
        display: flex;
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
        align-items: center;
        justify-content: flex-end;
        gap: 2px;

        .ant-btn {
            width: 30px;
            height: 30px;
            padding: 0;
            border: 0;
            border-radius: ${token.borderRadiusSM}px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            background: transparent;
            color: ${token.colorTextSecondary} !important;
            box-shadow: none;
            transition: background-color .16s ease, color .16s ease;
        }

        .ant-btn .anticon {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            line-height: 1;
        }

        .ant-btn:hover {
            color: ${token.colorText} !important;
            background: ${token.colorFillQuaternary};
        }

        .ant-btn-dangerous:hover {
            color: ${token.colorError} !important;
            background: ${token.colorErrorBg};
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
    const log = useLog(request, body);

    useImperativeHandle(ref, () => ({
        open: (record: any, requestBody?: Record<string, any>) => {
            if (log.open(record, requestBody)) {
                modalRef.current?.show();
            }
        },
    }), [log.open]);

    return (
        <ProModal
            ref={modalRef}
            title={<Title>{title}</Title>}
            width={900}
            modalProps={{
                rootClassName: styles.modal,
                footer: null,
                zIndex: 1200,
                style: {top: 100, paddingBottom: 100},
                mask: {closable: true},
                afterClose: log.reset,
            } as any}>
            <div className={styles.panel}>
                <div className={styles.panelHeader}>
                    <span className={styles.panelTitle}>{log.fileName || "日志文件"}</span>
                    <div className={styles.actions}>
                        <Tooltip title="刷新日志">
                            <Button
                                type="text"
                                aria-label="刷新日志"
                                icon={<Icon type="ReloadOutlined"/>}
                                loading={log.loading}
                                disabled={log.clearing}
                                onClick={log.refresh}/>
                        </Tooltip>
                        {showClear && (
                            <Tooltip title="清理日志">
                                <Button
                                    type="text"
                                    danger
                                    aria-label="清理日志"
                                    icon={<Icon type="DeleteOutlined"/>}
                                    loading={log.clearing}
                                    onClick={log.clear}/>
                            </Tooltip>
                        )}
                    </div>
                </div>
                <div ref={log.consoleRef} className={styles.console} onScroll={log.handleScroll}>
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
