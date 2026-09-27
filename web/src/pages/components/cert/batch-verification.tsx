import {forwardRef, useEffect, useImperativeHandle, useRef, useState} from "react";
import {Button, Form, Grid, Input, Typography} from "antd";
import {createStyles} from "antd-style";
import dayjs from "dayjs";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import type {ProModalRef} from "sinking-antd";

export interface BatchVerificationRef {
    show: () => void;
    hide: () => void;
}

export interface BatchVerificationProps {
    session: {domain: string; data: any; error?: string} | null;
    busy: boolean;
    onVerify: () => void;
    onRefresh: () => void;
    onSkip: () => void;
    afterClose?: () => void;
}

const useStyles = createStyles(({css, token}: any, props: any = {}) => ({
    modal: css`
        .ant-modal { max-width: calc(100vw - 24px); outline: none; }
        .ant-modal-body { min-width: 0; padding: 0 !important; overflow: hidden; }
    `,
    content: css`
        box-sizing: border-box;
        width: 100%;
        min-width: 0;
        max-height: calc(100dvh - 290px);
        padding: 4px;
        overflow-x: hidden;
        overflow-y: auto;
        scrollbar-gutter: stable;
        overscroll-behavior: contain;
        @supports not (max-height: 100dvh) { max-height: calc(100vh - 290px); }
        @media (max-width: 767px) {
            max-height: calc(100dvh - 164px);
            @supports not (max-height: 100dvh) { max-height: calc(100vh - 164px); }
        }
    `,
    summary: css`
        margin-bottom: ${props?.compact ? 14 : 18}px;
        .verification-tools { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; }
        .verification-expiry { color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; white-space: nowrap; }
        .verification-actions { display: flex; align-items: center; gap: 6px; }
        .verification-actions .ant-btn { height: 22px; padding-inline: 6px; font-size: ${token.fontSizeSM}px; }
        .verification-error { margin-top: 10px; color: ${token.colorError}; font-size: ${token.fontSizeSM}px; line-height: 1.6; overflow-wrap: anywhere; }
    `,
    record: css`
        min-width: 0;
        & + & { margin-top: ${props?.compact ? 18 : 22}px; padding-top: ${props?.compact ? 18 : 22}px; border-top: 1px solid ${token.colorBorderSecondary}; }
        .verification-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: ${props?.compact ? 12 : 16}px; }
        .verification-heading > span { min-width: 0; color: ${token.colorText}; font-size: ${token.fontSize}px; }
        .verification-heading > .ant-btn { flex: none; height: 22px; padding-inline: 6px; gap: 4px; font-size: ${token.fontSizeSM}px; }
        .ant-form-item { min-width: 0; margin-bottom: ${props?.compact ? 12 : 16}px; }
        .ant-form-item:last-child { margin-bottom: 0; }
        .ant-form-item-label { padding-bottom: ${props?.compact ? 4 : 6}px; }
        .verification-type { margin-inline-start: 6px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; font-weight: 400; }
        .verification-field { position: relative; min-width: 0; }
        textarea.ant-input { max-width: 100%; padding-inline-end: 32px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: ${token.fontSizeSM}px; line-height: 1.65; overflow-wrap: anywhere; resize: none; }
        .verification-copy { position: absolute; top: ${props?.compact ? 2 : 4}px; inset-inline-end: ${props?.compact ? 4 : 6}px; font-size: ${token.fontSizeSM}px; line-height: 22px; }
        .verification-copy .ant-typography-copy { display: inline-flex; align-items: center; justify-content: center; width: 22px; height: 22px; margin-inline-start: 0; }
    `,
    hint: css`
        color: ${token.colorTextSecondary};
        font-size: ${token.fontSizeSM}px;
        line-height: 1.6;
    `,
}));

const BatchVerification = forwardRef<BatchVerificationRef, BatchVerificationProps>(({session, busy, onVerify, onRefresh, onSkip, afterClose}, ref) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const modalRef = useRef<ProModalRef | null>(null);
    const [content, setContent] = useState<BatchVerificationProps["session"]>(null);
    const displayedSession = session || content;

    useEffect(() => {
        if (session) setContent(session);
    }, [session]);

    useImperativeHandle(ref, () => ({
        show: () => modalRef.current?.show(),
        hide: () => modalRef.current?.hide(),
    }), []);

    return <ProModal
        ref={modalRef}
        title={<Title>验证信息</Title>}
        width={compact ? 380 : 400}
        onOk={() => {if (session && !busy) onVerify();}}
        onCancel={() => {if (!busy) modalRef.current?.hide();}}
        modalProps={{
            rootClassName: styles.modal,
            zIndex: 1050,
            footer: [
                <Button key="back" disabled={busy} onClick={() => modalRef.current?.hide()}>返回</Button>,
                <Button key="verify" type="primary" loading={busy} disabled={!session} onClick={onVerify}>验证</Button>,
            ],
            forceRender: true,
            closable: !busy,
            keyboard: !busy,
            mask: {closable: !busy},
            style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
            focusable: {focusTriggerAfterClose: false},
            afterClose: () => {
                if (!session) setContent(null);
                afterClose?.();
            },
        } as any}>
        {displayedSession && <div className={styles.content} aria-busy={busy}>
            <div className={styles.summary}>
                <div className="verification-tools">
                    <span className="verification-expiry">有效至 {dayjs.unix(displayedSession.data.expires_at).format("MM-DD HH:mm")}</span>
                    <div className="verification-actions">
                        <Button size="small" color="default" variant="filled" disabled={busy || !session} onClick={onRefresh}>重新获取</Button>
                        <Button size="small" color="default" variant="filled" disabled={busy || !session} onClick={onSkip}>跳过此项</Button>
                    </div>
                </div>
                {displayedSession.error && <div className="verification-error" role="alert">{displayedSession.error}</div>}
            </div>
            {displayedSession.data.challenges.length === 0 && <div className={styles.hint}>无需添加验证信息，可直接验证。</div>}
            {displayedSession.data.challenges.map((record: any, index: number) => <section className={styles.record} key={record.id || index}>
                {record.type === "HTTP" && <div className="verification-heading">
                    <span>部署验证文件{displayedSession.data.challenges.length > 1 ? ` ${index + 1}` : ""}</span>
                    <Button size="small" color="default" variant="filled" icon={<Icon type="DownloadOutlined"/>} onClick={() => {
                        const url = URL.createObjectURL(new Blob([record.value], {type: "application/octet-stream"}));
                        const anchor = document.createElement("a");
                        anchor.href = url;
                        anchor.download = record.name;
                        anchor.click();
                        window.setTimeout(() => URL.revokeObjectURL(url), 1000);
                    }}>下载文件</Button>
                </div>}
                <Form component="div" layout="vertical" variant="filled">
                    {(record.type === "HTTP" ? [
                        {label: "文件名", value: record.name},
                        {label: "部署路径", value: record.path || ""},
                        {label: "文件内容", value: record.value},
                    ] : [
                        {label: "记录名称", value: record.name},
                        {label: "记录值", value: record.value},
                    ]).map(({label, value}) => <Form.Item key={label} label={record.type !== "HTTP" && label === "记录名称" ? <span>
                        {label}<span className="verification-type">TXT{displayedSession.data.challenges.length > 1 ? ` ${index + 1}` : ""}</span>
                    </span> : label}>
                        <div className="verification-field">
                            <Input.TextArea value={value} readOnly variant="filled" autoSize={{minRows: 1}} spellCheck={false} aria-label={label}/>
                            <Typography.Text className="verification-copy" copyable={{text: value, tooltips: [`复制${label}`, "已复制"]}}/>
                        </div>
                    </Form.Item>)}
                </Form>
            </section>)}
        </div>}
    </ProModal>;
});

BatchVerification.displayName = "BatchVerification";

export default BatchVerification;
